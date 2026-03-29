package devmode

import (
	"context"
	"time"

	"github.com/kbntx/kiln/internal/engine"
)

var _ engine.Engine = (*MockEngine)(nil)

type MockEngine struct{}

func (m *MockEngine) Init(_ context.Context, _ engine.RunOptions) error {
	return nil
}

func (m *MockEngine) Plan(ctx context.Context, _ engine.RunOptions, output chan<- engine.LogLine) error {
	defer close(output)
	lines := []string{
		"Initializing the backend...",
		"",
		"Initializing provider plugins...",
		"- Finding hashicorp/aws versions matching \"~> 5.0\"...",
		"- Installing hashicorp/aws v5.46.0...",
		"- Installed hashicorp/aws v5.46.0 (signed by HashiCorp)",
		"",
		"Terraform has been successfully initialized!",
		"",
		"\033[1mTerraform used the selected providers to generate the following execution plan.\033[0m",
		"",
		"Resource actions are indicated with the following symbols:",
		"  \033[32m+\033[0m create",
		"  \033[33m~\033[0m update in-place",
		"",
		"Terraform will perform the following actions:",
		"",
		"  \033[1m# aws_s3_bucket.logs\033[0m will be created",
		"  \033[32m+\033[0m resource \"aws_s3_bucket\" \"logs\" {",
		"      + bucket        = \"kiln-logs-dev\"",
		"      + force_destroy = false",
		"    }",
		"",
		"  \033[1m# aws_iam_role.runner\033[0m will be updated in-place",
		"  \033[33m~\033[0m resource \"aws_iam_role\" \"runner\" {",
		"      ~ tags = {",
		"          + \"env\" = \"dev\"",
		"        }",
		"    }",
		"",
		"\033[1mPlan:\033[0m 1 to add, 1 to change, 0 to destroy.",
	}
	for _, text := range lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			output <- engine.LogLine{
				Stream: "stdout",
				Text:   text,
				Time:   time.Now(),
			}
			time.Sleep(80 * time.Millisecond)
		}
	}
	return nil
}

func (m *MockEngine) Apply(ctx context.Context, opts engine.RunOptions, output chan<- engine.LogLine) error {
	defer close(output)
	lines := []string{
		"\033[1maws_s3_bucket.logs: Creating...\033[0m",
		"\033[32maws_s3_bucket.logs: Creation complete after 2s [id=kiln-logs-dev]\033[0m",
		"",
		"\033[1maws_iam_role.runner: Modifying...\033[0m",
		"\033[33maws_iam_role.runner: Modifications complete after 1s [id=runner-role]\033[0m",
		"",
		"\033[1m\033[32mApply complete! Resources: 1 added, 1 changed, 0 destroyed.\033[0m",
	}
	for _, text := range lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			output <- engine.LogLine{
				Stream: "stdout",
				Text:   text,
				Time:   time.Now(),
			}
			time.Sleep(80 * time.Millisecond)
		}
	}
	return nil
}

func (m *MockEngine) Name() string {
	return "mock-terraform"
}
