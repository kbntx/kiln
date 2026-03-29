package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var _ Engine = (*TerraformEngine)(nil)

type TerraformEngine struct{}

func (t *TerraformEngine) Name() string {
	return "terraform"
}

func (t *TerraformEngine) Init(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
	defer close(output)

	// Ensure the requested terraform version is installed via tfenv.
	if opts.TerraformVersion != "" {
		if _, err := ensureTerraformVersion(opts.TerraformVersion); err != nil {
			return err
		}
	}

	cmd := exec.CommandContext(ctx, "terraform", "init")
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnvWithTFVersion(opts.EnvVars, opts.TerraformVersion)
	return pipeCommand(cmd, output)
}

func (t *TerraformEngine) Plan(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
	defer close(output)
	planFile := opts.PlanFile
	if planFile == "" {
		planFile = "tf.plan"
	}
	args := []string{"plan", "-out=" + planFile}
	if opts.Destroy {
		args = append(args, "-destroy")
	}
	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnvWithTFVersion(opts.EnvVars, opts.TerraformVersion)
	return pipeCommand(cmd, output)
}

func (t *TerraformEngine) Apply(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
	defer close(output)
	planFile := opts.PlanFile
	if planFile == "" {
		planFile = "tf.plan"
	}
	// Apply the saved plan file — no -auto-approve needed since the plan is pre-approved.
	args := []string{"apply", planFile}
	cmd := exec.CommandContext(ctx, "terraform", args...)
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnvWithTFVersion(opts.EnvVars, opts.TerraformVersion)
	return pipeCommand(cmd, output)
}

// pipeCommand runs a command and sends stdout/stderr line by line to the output channel.
func pipeCommand(cmd *exec.Cmd, output chan<- LogLine) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	done := make(chan struct{}, 2)

	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			output <- LogLine{Stream: "stdout", Text: sc.Text(), Time: time.Now()}
		}
		done <- struct{}{}
	}()

	go func() {
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			output <- LogLine{Stream: "stderr", Text: sc.Text(), Time: time.Now()}
		}
		done <- struct{}{}
	}()

	// Wait for both scanners to finish.
	<-done
	<-done

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func (t *TerraformEngine) HasChanges(ctx context.Context, opts RunOptions) (bool, error) {
	planFile := opts.PlanFile
	if planFile == "" {
		planFile = "tf.plan"
	}

	cmd := exec.CommandContext(ctx, "terraform", "show", "-json", planFile)
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnvWithTFVersion(opts.EnvVars, opts.TerraformVersion)

	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("terraform show -json: %w", err)
	}

	var plan struct {
		ResourceChanges []struct {
			Change struct {
				Actions []string `json:"actions"`
			} `json:"change"`
		} `json:"resource_changes"`
	}
	if err := json.Unmarshal(out, &plan); err != nil {
		return false, fmt.Errorf("parse plan json: %w", err)
	}

	for _, rc := range plan.ResourceChanges {
		for _, action := range rc.Change.Actions {
			if action != "no-op" {
				return true, nil
			}
		}
	}
	return false, nil
}

// buildEnv converts a map of environment variables to a slice of KEY=VALUE strings.
// Returns nil (inherit parent env) if the map is empty.
func buildEnv(envVars map[string]string) []string {
	if len(envVars) == 0 {
		return nil
	}
	env := make([]string, 0, len(envVars))
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}
	return env
}

// buildEnvWithTFVersion builds the env slice and injects TFENV_TERRAFORM_VERSION
// so tfenv uses the correct version for this specific subprocess.
func buildEnvWithTFVersion(envVars map[string]string, tfVersion string) []string {
	env := buildEnv(envVars)
	if tfVersion != "" {
		if env == nil {
			env = os.Environ()
		}
		env = append(env, "TFENV_TERRAFORM_VERSION="+tfVersion)
	}
	return env
}
