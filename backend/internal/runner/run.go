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
	HeadSHA    string              `json:"headSha"`
	ProjectDir string              `json:"projectDir"`
	Stack      string              `json:"stack"`
	Profile    string              `json:"profile"`
	Operation  string              `json:"operation"` // "plan" or "apply"
	Destroy    bool                `json:"destroy"`
	PlanRunID        string              `json:"planRunId,omitempty"` // apply only: the plan run whose workspace to reuse
	TerraformVersion string              `json:"terraformVersion,omitempty"`
	Status     RunStatus           `json:"status"`
	Projects   []discovery.Project `json:"projects"`
	CreatedAt  time.Time           `json:"createdAt"`
	Config     *discovery.Config   `json:"-"` // parsed kiln.yaml, not exposed to API
	WorkDir    string              `json:"-"` // absolute path to project dir, not exposed to API
}
