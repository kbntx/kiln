package engine

import (
	"fmt"
	"os"
	"path/filepath"
)

// DetectEngine inspects a directory and returns the appropriate Engine
// based on the presence of Pulumi.yaml or *.tf files.
func DetectEngine(dir string) (Engine, error) {
	// Check for Pulumi.yaml first.
	if _, err := os.Stat(filepath.Join(dir, "Pulumi.yaml")); err == nil {
		return &PulumiEngine{}, nil
	}

	// Check for any .tf files.
	matches, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("detect engine: %w", err)
	}
	if len(matches) > 0 {
		return &TerraformEngine{}, nil
	}

	return nil, fmt.Errorf("detect engine: no Pulumi.yaml or *.tf files found in %s", dir)
}
