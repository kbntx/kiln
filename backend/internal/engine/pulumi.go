package engine

// TODO(pulumi): Uncomment when Pulumi support is implemented.
//
// import (
// 	"context"
// 	"fmt"
// 	"os/exec"
// )
//
// var _ Engine = (*PulumiEngine)(nil)
//
// type PulumiEngine struct{}
//
// func (p *PulumiEngine) Name() string {
// 	return "pulumi"
// }
//
// func (p *PulumiEngine) Init(ctx context.Context, opts RunOptions) error {
// 	env := buildEnv(opts.EnvVars)
//
// 	install := exec.CommandContext(ctx, "pulumi", "install")
// 	install.Dir = opts.WorkDir
// 	install.Env = env
// 	if out, err := install.CombinedOutput(); err != nil {
// 		return fmt.Errorf("pulumi install: %w\n%s", err, out)
// 	}
//
// 	if opts.Stack != "" {
// 		sel := exec.CommandContext(ctx, "pulumi", "stack", "select", opts.Stack)
// 		sel.Dir = opts.WorkDir
// 		sel.Env = env
// 		if err := sel.Run(); err != nil {
// 			init := exec.CommandContext(ctx, "pulumi", "stack", "init", opts.Stack)
// 			init.Dir = opts.WorkDir
// 			init.Env = env
// 			if out, err := init.CombinedOutput(); err != nil {
// 				return fmt.Errorf("pulumi stack init: %w\n%s", err, out)
// 			}
// 		}
// 	}
//
// 	return nil
// }
//
// func (p *PulumiEngine) Plan(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
// 	defer close(output)
// 	args := []string{"preview", "--non-interactive"}
// 	if opts.Stack != "" {
// 		args = append(args, "--stack", opts.Stack)
// 	}
// 	cmd := exec.CommandContext(ctx, "pulumi", args...)
// 	cmd.Dir = opts.WorkDir
// 	cmd.Env = buildEnv(opts.EnvVars)
// 	return pipeCommand(cmd, output)
// }
//
// func (p *PulumiEngine) Apply(ctx context.Context, opts RunOptions, output chan<- LogLine) error {
// 	defer close(output)
// 	args := []string{"up", "--non-interactive", "--yes"}
// 	if opts.Stack != "" {
// 		args = append(args, "--stack", opts.Stack)
// 	}
// 	cmd := exec.CommandContext(ctx, "pulumi", args...)
// 	cmd.Dir = opts.WorkDir
// 	cmd.Env = buildEnv(opts.EnvVars)
// 	return pipeCommand(cmd, output)
// }
