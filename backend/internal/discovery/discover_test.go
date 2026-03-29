package discovery

import (
	"os"
	"testing"
)

func TestParseConfig(t *testing.T) {
	data := []byte(`
profiles:
  prod:
    env:
      AWS_PROFILE: prod-account
      AWS_REGION: eu-west-1
  dev:
    env:
      AWS_PROFILE: dev-account

terraform_version: "1.8.4"

projects:
  - name: networking
    dir: terraform
    engine: terraform
    stacks: [default]
    profile: prod
  - name: compute
    dir: terraform/compute
    engine: terraform
    stacks: [default]
    profile: dev
`)

	cfg, err := ParseConfig(data)
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	if len(cfg.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(cfg.Projects))
	}

	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(cfg.Profiles))
	}

	// Check first project.
	p := cfg.Projects[0]
	if p.Name != "networking" {
		t.Errorf("project 0 name = %q, want %q", p.Name, "networking")
	}
	if p.Dir != "terraform" {
		t.Errorf("project 0 dir = %q, want %q", p.Dir, "terraform")
	}
	if p.Engine != "terraform" {
		t.Errorf("project 0 engine = %q, want %q", p.Engine, "terraform")
	}
	if len(p.Stacks) != 1 || p.Stacks[0] != "default" {
		t.Errorf("project 0 stacks = %v, want [default]", p.Stacks)
	}
	if p.Profile != "prod" {
		t.Errorf("project 0 profile = %q, want %q", p.Profile, "prod")
	}

	// Check ToProjects conversion.
	projects := cfg.ToProjects()
	if len(projects) != 2 {
		t.Fatalf("ToProjects returned %d, want 2", len(projects))
	}
	if projects[0].Engine != EngineTypeTerraform {
		t.Errorf("ToProjects[0].Engine = %q, want %q", projects[0].Engine, EngineTypeTerraform)
	}

	// Check profiles.
	prod := cfg.Profiles["prod"]
	if prod.Env["AWS_PROFILE"] != "prod-account" {
		t.Errorf("prod profile AWS_PROFILE = %q, want %q", prod.Env["AWS_PROFILE"], "prod-account")
	}
}

func TestParseConfig_Validation(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"no projects", `profiles: {}`},
		{"no name", `projects: [{dir: ".", engine: terraform, stacks: [default]}]`},
		{"no engine", `projects: [{name: foo, dir: ".", stacks: [default]}]`},
		// TODO(pulumi): Update when Pulumi support is implemented.
		{"unsupported engine", `projects: [{name: foo, dir: ".", engine: pulumi, stacks: [default]}]`},
		{"no stacks", `projects: [{name: foo, dir: ".", engine: terraform}]`},
		{"no terraform version", `projects: [{name: foo, dir: ".", engine: terraform, stacks: [default]}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig([]byte(tt.data))
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestResolveProfileEnv(t *testing.T) {
	// Set a test env var.
	os.Setenv("TEST_KILN_VAR", "resolved-value")
	defer os.Unsetenv("TEST_KILN_VAR")

	profile := Profile{
		Env: map[string]string{
			"STATIC":  "hello",
			"DYNAMIC": "${TEST_KILN_VAR}",
			"MIXED":   "prefix-${TEST_KILN_VAR}-suffix",
		},
	}

	env := ResolveProfileEnv(profile)

	// Should have 3 system vars + 3 profile vars.
	if len(env) != 6 {
		t.Fatalf("expected 6 env vars, got %d: %v", len(env), env)
	}

	envMap := make(map[string]string)
	for _, e := range env {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}

	if envMap["STATIC"] != "hello" {
		t.Errorf("STATIC = %q, want %q", envMap["STATIC"], "hello")
	}
	if envMap["DYNAMIC"] != "resolved-value" {
		t.Errorf("DYNAMIC = %q, want %q", envMap["DYNAMIC"], "resolved-value")
	}
	if envMap["MIXED"] != "prefix-resolved-value-suffix" {
		t.Errorf("MIXED = %q, want %q", envMap["MIXED"], "prefix-resolved-value-suffix")
	}
}
