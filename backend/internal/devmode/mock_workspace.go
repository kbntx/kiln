package devmode

import (
	"context"
	"os"
	"path/filepath"

	"github.com/kbntx/kiln/internal/git"
)

var _ git.WorkspaceManager = (*MockWorkspace)(nil)

type MockWorkspace struct {
	testdataDir  string
	workspaceDir string
}

func NewMockWorkspace(testdataDir string) *MockWorkspace {
	if d := os.Getenv("DEV_REPO_DIR"); d != "" {
		testdataDir = d
	}
	abs, err := filepath.Abs(testdataDir)
	if err == nil {
		testdataDir = abs
	}

	wsDir := filepath.Join(".", ".kiln-workspace")
	absWs, err := filepath.Abs(wsDir)
	if err == nil {
		wsDir = absWs
	}

	return &MockWorkspace{
		testdataDir:  testdataDir,
		workspaceDir: wsDir,
	}
}

func (w *MockWorkspace) Allocate(runID string) (string, error) {
	dir := filepath.Join(w.workspaceDir, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func (w *MockWorkspace) CloneOrLink(_ context.Context, _, _, destDir string) error {
	return os.Symlink(w.testdataDir, filepath.Join(destDir, "repo"))
}

func (w *MockWorkspace) Release(runID string) error {
	return os.RemoveAll(filepath.Join(w.workspaceDir, runID))
}
