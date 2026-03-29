package engine

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// ensureTerraformVersion installs the requested terraform version via tfenv
// and returns the version string to set as TFENV_TERRAFORM_VERSION in the
// subprocess environment. This avoids global `tfenv use` which would affect
// other concurrent runs.
// If version is empty, returns "" (use whatever is currently active).
func ensureTerraformVersion(version string) (string, error) {
	if version == "" {
		return "", nil
	}

	slog.Info("ensuring terraform version", "version", version)

	// Set TFENV_TERRAFORM_VERSION so tfenv doesn't try to read its version file.
	env := append(os.Environ(), "TFENV_TERRAFORM_VERSION="+version)

	// Install the version (no-op if already installed).
	install := exec.Command("tfenv", "install", version)
	install.Env = env
	if out, err := install.CombinedOutput(); err != nil {
		return "", fmt.Errorf("tfenv install %s: %w\n%s", version, err, out)
	}

	return version, nil
}
