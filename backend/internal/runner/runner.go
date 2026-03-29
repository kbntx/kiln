package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"

	"github.com/kbntx/kiln/internal/discovery"
	"github.com/kbntx/kiln/internal/engine"
	"github.com/kbntx/kiln/internal/git"
	"github.com/kbntx/kiln/internal/github"
	"github.com/kbntx/kiln/internal/stream"
)

// Runner orchestrates discovery and plan/apply operations.
type Runner struct {
	store     *RunStore
	workspace git.WorkspaceManager
	ghClient  github.GitHubClient
	engine    engine.Engine
	broker    *stream.Broker
}

// NewRunner creates a new Runner.
func NewRunner(store *RunStore, ws git.WorkspaceManager, ghClient github.GitHubClient, eng engine.Engine, broker *stream.Broker) *Runner {
	return &Runner{
		store:     store,
		workspace: ws,
		ghClient:  ghClient,
		engine:    eng,
		broker:    broker,
	}
}

// StartDiscovery fetches kiln.yaml from the PR branch via the GitHub API and publishes
// the parsed projects. No cloning is performed.
func (r *Runner) StartDiscovery(ctx context.Context, runID string) {
	go func() {
		run := r.store.Get(runID)
		if run == nil {
			return
		}

		r.setStatus(runID, RunStatusDiscovering)

		// Fetch kiln.yaml at the exact commit SHA via GitHub API.
		data, err := r.ghClient.GetFileContent(ctx, run.Owner, run.Repo, run.HeadSHA, "kiln.yaml")
		if err != nil {
			r.fail(runID, fmt.Sprintf("fetch kiln.yaml: %v", err))
			return
		}

		cfg, err := discovery.ParseConfig(data)
		if err != nil {
			r.fail(runID, fmt.Sprintf("parse kiln.yaml: %v", err))
			return
		}

		projects := cfg.ToProjects()

		// Update run with parsed config and projects.
		r.store.Update(runID, func(run *Run) {
			run.Projects = projects
			run.Config = cfg
		})

		// Emit projects event.
		projData, err := json.Marshal(projects)
		if err != nil {
			log.Printf("runner: marshal projects: %v", err)
			return
		}
		r.broker.Publish(runID, stream.Event{Type: "projects", Data: string(projData)})

		r.setStatus(runID, RunStatusSuccess)
	}()
}

// StartRun kicks off an async goroutine that clones the repo and runs the plan or apply
// operation. For apply, it reuses the workspace from the prior plan run to use the saved plan file.
func (r *Runner) StartRun(ctx context.Context, runID string) {
	go func() {
		run := r.store.Get(runID)
		if run == nil {
			return
		}

		var projectDir string

		if run.Operation == "apply" && run.PlanRunID != "" {
			// Reuse workspace from the specific plan run — it contains the saved plan file.
			if planRun := r.store.Get(run.PlanRunID); planRun != nil && planRun.WorkDir != "" {
				projectDir = planRun.WorkDir
			}
		}

		if projectDir == "" {
			// Clone the repo (plan runs, or apply if no prior plan workspace found).
			r.setStatus(runID, RunStatusCloning)

			workDir, err := r.workspace.Allocate(runID)
			if err != nil {
				r.fail(runID, fmt.Sprintf("allocate workspace: %v", err))
				return
			}

			repoURL := fmt.Sprintf("https://github.com/%s/%s.git", run.Owner, run.Repo)
			if err := r.workspace.CloneOrLink(ctx, repoURL, run.PRBranch, workDir); err != nil {
				r.fail(runID, fmt.Sprintf("clone: %v", err))
				return
			}

			repoDir := workDir + "/repo"
			if resolved, err := filepath.EvalSymlinks(repoDir); err == nil {
				repoDir = resolved
			}

			projectDir = filepath.Join(repoDir, run.ProjectDir)
		}

		// Save workspace path so apply can reuse it for the saved plan file.
		r.store.Update(runID, func(run *Run) {
			run.WorkDir = projectDir
		})

		// Run the engine.
		r.setStatus(runID, RunStatusRunning)

		// Detect or use injected engine.
		eng := r.engine
		if eng == nil {
			var err error
			eng, err = engine.DetectEngine(projectDir)
			if err != nil {
				r.fail(runID, fmt.Sprintf("detect engine: %v", err))
				return
			}
		}

		// Resolve profile env vars.
		var envVars map[string]string
		if run.Config != nil && run.Profile != "" {
			if profile, ok := run.Config.Profiles[run.Profile]; ok {
				envSlice := discovery.ResolveProfileEnv(profile)
				envVars = make(map[string]string, len(envSlice))
				for _, e := range envSlice {
					for i := 0; i < len(e); i++ {
						if e[i] == '=' {
							envVars[e[:i]] = e[i+1:]
							break
						}
					}
				}
			}
		}

		// Resolve terraform version from the config (global or per-project).
		tfVersion := run.TerraformVersion
		if tfVersion == "" && run.Config != nil {
			for _, p := range run.Config.Projects {
				if p.Dir == run.ProjectDir {
					tfVersion = p.TerraformVersion
					break
				}
			}
			if tfVersion == "" {
				tfVersion = run.Config.TerraformVersion
			}
		}

		opts := engine.RunOptions{
			WorkDir:          projectDir,
			Stack:            run.Stack,
			EnvVars:          envVars,
			Destroy:          run.Destroy,
			TerraformVersion: tfVersion,
		}

		// Helper to stream a channel of log lines to the broker.
		streamLogs := func(ch <-chan engine.LogLine) {
			for line := range ch {
				data, _ := json.Marshal(line)
				r.broker.Publish(runID, stream.Event{Type: "log", Data: string(data)})
			}
		}

		// Helper to run an engine step async, stream its logs, and return the error.
		runStep := func(fn func(chan<- engine.LogLine) error) error {
			output := make(chan engine.LogLine, 64)
			errCh := make(chan error, 1)
			go func() { errCh <- fn(output) }()
			streamLogs(output)
			return <-errCh
		}

		// Init
		if err := runStep(func(out chan<- engine.LogLine) error { return eng.Init(ctx, opts, out) }); err != nil {
			r.fail(runID, fmt.Sprintf("init failed: %v", err))
			r.cleanupRun(runID)
			return
		}

		// Plan or Apply
		var stepErr error
		switch run.Operation {
		case "plan":
			stepErr = runStep(func(out chan<- engine.LogLine) error { return eng.Plan(ctx, opts, out) })
		case "apply":
			stepErr = runStep(func(out chan<- engine.LogLine) error { return eng.Apply(ctx, opts, out) })
		default:
			r.fail(runID, fmt.Sprintf("unknown operation: %s", run.Operation))
			return
		}

		if stepErr != nil {
			r.fail(runID, fmt.Sprintf("%s failed: %v", run.Operation, stepErr))
			r.cleanupRun(runID)
			return
		}

		// After a successful plan, check if there are actual changes.
		if run.Operation == "plan" {
			hasChanges, err := eng.HasChanges(ctx, opts)
			if err != nil {
				log.Printf("runner: has_changes check failed for %s: %v", runID, err)
				hasChanges = true // assume changes on error to be safe
			}
			if hasChanges {
				r.broker.Publish(runID, stream.Event{Type: "has_changes", Data: "true"})
			} else {
				r.broker.Publish(runID, stream.Event{Type: "has_changes", Data: "false"})
			}
		}

		r.setStatus(runID, RunStatusSuccess)

		// Clean up after success.
		r.cleanupDiscoveryRuns(run.Owner, run.Repo, run.PRNumber)
		if run.Operation == "apply" {
			// Apply done: clean up the specific plan run that was used.
			if run.PlanRunID != "" {
				r.cleanupRun(run.PlanRunID)
			}
			// Clean up this apply run too.
			r.cleanupRun(runID)
		}
		// Plan success: keep the workspace (apply needs the tfplan file).
	}()
}

// setStatus updates the run status and publishes a status event.
// For terminal statuses (success, failed), a "done" event is also sent
// so the client can close the EventSource before the server cleans up.
func (r *Runner) setStatus(runID string, status RunStatus) {
	r.store.Update(runID, func(run *Run) {
		run.Status = status
	})
	r.broker.Publish(runID, stream.Event{Type: "status", Data: string(status)})
	if status == RunStatusSuccess || status == RunStatusFailed {
		r.broker.Publish(runID, stream.Event{Type: "done", Data: ""})
	}
}

// fail sets the run to failed status and publishes the error message.
func (r *Runner) fail(runID string, msg string) {
	log.Printf("runner: run %s failed: %s", runID, msg)
	r.broker.Publish(runID, stream.Event{Type: "run_error", Data: msg})
	r.setStatus(runID, RunStatusFailed)
}

// cleanupRun releases workspace and cleans up the run from the store and broker.
func (r *Runner) cleanupRun(runID string) {
	if err := r.workspace.Release(runID); err != nil {
		log.Printf("runner: release workspace %s: %v", runID, err)
	}
	r.store.Delete(runID)
	r.broker.Cleanup(runID)
}

// cleanupAfterPlan cleans up discovery runs for this PR (no longer needed).
func (r *Runner) cleanupDiscoveryRuns(owner, repo string, prNumber int) {
	r.store.ForEach(func(run *Run) bool {
		if run.Owner == owner && run.Repo == repo && run.PRNumber == prNumber && run.Operation == "" {
			r.broker.Cleanup(run.ID)
			return true // mark for deletion
		}
		return false
	})
}

