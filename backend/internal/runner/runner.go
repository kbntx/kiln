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

// StartDiscovery kicks off an async goroutine that clones the repo and discovers projects.
func (r *Runner) StartDiscovery(ctx context.Context, runID string) {
	go func() {
		run := r.store.Get(runID)
		if run == nil {
			return
		}

		// 1. Cloning
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

		// 2. Discovering
		r.setStatus(runID, RunStatusDiscovering)

		repoDir := workDir + "/repo"
		// Resolve symlinks so filepath.Walk can traverse the directory (needed for dev mode).
		if resolved, err := filepath.EvalSymlinks(repoDir); err == nil {
			repoDir = resolved
		}

		projects, err := discovery.DiscoverProjects(repoDir)
		if err != nil {
			r.fail(runID, fmt.Sprintf("discovery: %v", err))
			return
		}

		// Update run with discovered projects and workspace path.
		r.store.Update(runID, func(run *Run) {
			run.Projects = projects
			run.WorkDir = repoDir
		})

		// Emit projects event.
		data, err := json.Marshal(projects)
		if err != nil {
			log.Printf("runner: marshal projects: %v", err)
			return
		}
		r.broker.Publish(runID, stream.Event{Type: "projects", Data: string(data)})

		r.setStatus(runID, RunStatusSuccess)
	}()
}

// StartRun kicks off an async goroutine that runs the plan or apply operation.
func (r *Runner) StartRun(ctx context.Context, runID string) {
	go func() {
		run := r.store.Get(runID)
		if run == nil {
			return
		}

		r.setStatus(runID, RunStatusRunning)

		// Resolve projectDir to an absolute path using the workspace from the discovery run.
		projectDir := run.ProjectDir
		if workDir := r.store.FindWorkDir(run.Owner, run.Repo, run.PRNumber); workDir != "" {
			projectDir = filepath.Join(workDir, run.ProjectDir)
		}

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

		opts := engine.RunOptions{
			WorkDir: projectDir,
			Stack:   run.Stack,
		}

		// Init
		if err := eng.Init(ctx, opts); err != nil {
			r.fail(runID, fmt.Sprintf("init: %v", err))
			return
		}

		// Plan or Apply
		output := make(chan engine.LogLine, 64)
		errCh := make(chan error, 1)

		switch run.Operation {
		case "plan":
			go func() {
				errCh <- eng.Plan(ctx, opts, output)
			}()
		case "apply":
			go func() {
				errCh <- eng.Apply(ctx, opts, output)
			}()
		default:
			r.fail(runID, fmt.Sprintf("unknown operation: %s", run.Operation))
			return
		}

		// Forward log lines to broker.
		for line := range output {
			data, _ := json.Marshal(line)
			r.broker.Publish(runID, stream.Event{Type: "log", Data: string(data)})
		}

		if runErr := <-errCh; runErr != nil {
			r.fail(runID, fmt.Sprintf("%s failed: %v", run.Operation, runErr))
			return
		}

		r.setStatus(runID, RunStatusSuccess)
	}()
}

// setStatus updates the run status and publishes a status event.
func (r *Runner) setStatus(runID string, status RunStatus) {
	r.store.Update(runID, func(run *Run) {
		run.Status = status
	})
	r.broker.Publish(runID, stream.Event{Type: "status", Data: string(status)})
}

// fail sets the run to failed status and publishes the error.
func (r *Runner) fail(runID string, msg string) {
	log.Printf("runner: run %s failed: %s", runID, msg)
	r.setStatus(runID, RunStatusFailed)
}
