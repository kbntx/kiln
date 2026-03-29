package engine

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"
)

var _ Engine = (*TerraformEngine)(nil)

type TerraformEngine struct{}

func (t *TerraformEngine) Name() string {
	return "terraform"
}

func (t *TerraformEngine) Init(ctx context.Context, opts RunOptions) error {
	cmd := exec.CommandContext(ctx, "terraform", "init")
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnv(opts.EnvVars)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init: %w\n%s", err, out)
	}
	return nil
}

func (t *TerraformEngine) Plan(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
	defer close(output)
	cmd := exec.CommandContext(ctx, "terraform", "plan")
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnv(opts.EnvVars)
	return pipeCommand(cmd, output)
}

func (t *TerraformEngine) Apply(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
	defer close(output)
	cmd := exec.CommandContext(ctx, "terraform", "apply", "-auto-approve")
	cmd.Dir = opts.WorkDir
	cmd.Env = buildEnv(opts.EnvVars)
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
