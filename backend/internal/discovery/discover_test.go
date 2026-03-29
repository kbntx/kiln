package discovery

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine test file path")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "fake-infra")
}

func TestDiscoverProjects(t *testing.T) {
	root := testdataDir(t)

	projects, err := DiscoverProjects(root)
	if err != nil {
		t.Fatalf("DiscoverProjects returned error: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d: %+v", len(projects), projects)
	}

	// Projects are sorted by Dir, so "." (pulumi at root) comes before "terraform".
	pulumi := projects[0]
	tf := projects[1]

	// Pulumi project checks.
	if pulumi.Name != "fake-pulumi-project" {
		t.Errorf("pulumi project name = %q, want %q", pulumi.Name, "fake-pulumi-project")
	}
	if pulumi.Dir != "." {
		t.Errorf("pulumi project dir = %q, want %q", pulumi.Dir, ".")
	}
	if pulumi.Engine != EngineTypePulumi {
		t.Errorf("pulumi engine = %q, want %q", pulumi.Engine, EngineTypePulumi)
	}
	if len(pulumi.Stacks) != 2 {
		t.Fatalf("pulumi stacks count = %d, want 2", len(pulumi.Stacks))
	}
	if pulumi.Stacks[0] != "dev" || pulumi.Stacks[1] != "prod" {
		t.Errorf("pulumi stacks = %v, want [dev prod]", pulumi.Stacks)
	}

	// Terraform project checks.
	if tf.Name != "terraform" {
		t.Errorf("terraform project name = %q, want %q", tf.Name, "terraform")
	}
	if tf.Dir != "terraform" {
		t.Errorf("terraform project dir = %q, want %q", tf.Dir, "terraform")
	}
	if tf.Engine != EngineTypeTerraform {
		t.Errorf("terraform engine = %q, want %q", tf.Engine, EngineTypeTerraform)
	}
	if len(tf.Stacks) != 1 || tf.Stacks[0] != "default" {
		t.Errorf("terraform stacks = %v, want [default]", tf.Stacks)
	}
}
