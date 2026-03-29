package git

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// Compile-time interface check.
var _ WorkspaceManager = (*RealWorkspace)(nil)

// RealWorkspace manages temporary directories and performs real git clones.
type RealWorkspace struct {
	baseDir string
	token   string
	mu      sync.Mutex
	dirs    map[string]string
}

// NewRealWorkspace creates a RealWorkspace rooted at baseDir. The directory is
// created (with parents) if it does not already exist.
func NewRealWorkspace(baseDir, token string) *RealWorkspace {
	_ = os.MkdirAll(baseDir, 0o755)
	return &RealWorkspace{
		baseDir: baseDir,
		token:   token,
		dirs:    make(map[string]string),
	}
}

// Allocate creates a temporary directory under baseDir for the given run and
// returns its path.
func (w *RealWorkspace) Allocate(runID string) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if dir, ok := w.dirs[runID]; ok {
		return dir, nil
	}

	dir, err := os.MkdirTemp(w.baseDir, runID+"-")
	if err != nil {
		return "", fmt.Errorf("allocate workspace: %w", err)
	}
	w.dirs[runID] = dir
	return dir, nil
}

// CloneOrLink performs a shallow clone of the given branch into destDir/repo.
// If a token is configured it is injected into the HTTPS URL for authentication.
func (w *RealWorkspace) CloneOrLink(ctx context.Context, repoURL, branch, destDir string) error {
	cloneURL, err := w.authenticatedURL(repoURL)
	if err != nil {
		return fmt.Errorf("prepare clone url: %w", err)
	}

	target := filepath.Join(destDir, "repo")

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", branch, cloneURL, target)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// Release removes the workspace directory that was previously allocated for
// runID.
func (w *RealWorkspace) Release(runID string) error {
	w.mu.Lock()
	dir, ok := w.dirs[runID]
	if ok {
		delete(w.dirs, runID)
	}
	w.mu.Unlock()

	if !ok {
		return fmt.Errorf("release: unknown run %q", runID)
	}
	return os.RemoveAll(dir)
}

// authenticatedURL injects the configured token into an HTTPS GitHub URL. If no
// token is set the original URL is returned unchanged.
func (w *RealWorkspace) authenticatedURL(raw string) (string, error) {
	if w.token == "" {
		return raw, nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-access-token", w.token)
	return u.String(), nil
}
