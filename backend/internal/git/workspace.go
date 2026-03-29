package git

import "context"

type WorkspaceManager interface {
	Allocate(runID string) (string, error)
	CloneOrLink(ctx context.Context, url, branch, destDir string) error
	Release(runID string) error
}
