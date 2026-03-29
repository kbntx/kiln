package runner

import (
	"time"

	"github.com/kbntx/kiln/internal/discovery"
)

// RunStatus represents the current state of a run.
type RunStatus string

const (
	RunStatusPending     RunStatus = "pending"
	RunStatusCloning     RunStatus = "cloning"
	RunStatusDiscovering RunStatus = "discovering"
	RunStatusRunning     RunStatus = "running"
	RunStatusSuccess     RunStatus = "success"
	RunStatusFailed      RunStatus = "failed"
)

// Run represents a single plan or apply execution.
type Run struct {
	ID         string              `json:"id"`
	Owner      string              `json:"owner"`
	Repo       string              `json:"repo"`
	PRNumber   int                 `json:"prNumber"`
	PRBranch   string              `json:"prBranch"`
	ProjectDir string              `json:"projectDir"`
	Stack      string              `json:"stack"`
	Operation  string              `json:"operation"` // "plan" or "apply"
	Status     RunStatus           `json:"status"`
	Projects   []discovery.Project `json:"projects"`
	CreatedAt  time.Time           `json:"createdAt"`
	WorkDir    string              `json:"-"` // absolute path to cloned repo root, not exposed to API
}
