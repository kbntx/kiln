package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kbntx/kiln/internal/runner"
	"github.com/kbntx/kiln/internal/stream"
)

type runsHandlers struct {
	runner *runner.Runner
	store  *runner.RunStore
	broker *stream.Broker
}

type createRunRequest struct {
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	PRNumber   int    `json:"prNumber"`
	PRBranch   string `json:"prBranch"`
	HeadSHA    string `json:"headSha"`
	ProjectDir string `json:"projectDir"`
	Stack      string `json:"stack"`
	Profile    string `json:"profile"`
	Operation  string `json:"operation"`
	Destroy    bool   `json:"destroy"`
	PlanRunID        string `json:"planRunId"`        // apply only: the plan run whose workspace to reuse
	TerraformVersion string `json:"terraformVersion"` // terraform version to use
}

func (h *runsHandlers) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	run := &runner.Run{
		Owner:      req.Owner,
		Repo:       req.Repo,
		PRNumber:   req.PRNumber,
		PRBranch:   req.PRBranch,
		HeadSHA:    req.HeadSHA,
		ProjectDir: req.ProjectDir,
		Stack:      req.Stack,
		Profile:    req.Profile,
		Operation:  req.Operation,
		Destroy:    req.Destroy,
		PlanRunID:        req.PlanRunID,
		TerraformVersion: req.TerraformVersion,
		Status:     runner.RunStatusPending,
	}
	h.store.Create(run)

	if req.Operation == "" {
		// Discovery-only run: fetch kiln.yaml via GitHub API.
		h.runner.StartDiscovery(context.Background(), run.ID)
	} else {
		// Full run: clone + plan or apply.
		// Attach the config from the prior discovery run so we can resolve profile env vars.
		if cfg := h.store.FindConfig(run.Owner, run.Repo, run.HeadSHA); cfg != nil {
			h.store.Update(run.ID, func(r *runner.Run) {
				r.Config = cfg
			})
		}
		h.runner.StartRun(context.Background(), run.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(run)
}

func (h *runsHandlers) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run := h.store.Get(id)
	if run == nil {
		http.Error(w, "run not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func (h *runsHandlers) handleStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run := h.store.Get(id)
	if run == nil {
		http.Error(w, "run not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	events, unsub := h.broker.Subscribe(id)
	defer unsub()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			stream.WriteSSE(w, ev.Type, ev.Data)
		}
	}
}
