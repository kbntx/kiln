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
	ProjectDir string `json:"projectDir"`
	Stack      string `json:"stack"`
	Operation  string `json:"operation"`
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
		ProjectDir: req.ProjectDir,
		Stack:      req.Stack,
		Operation:  req.Operation,
		Status:     runner.RunStatusPending,
	}
	h.store.Create(run)

	if req.Operation == "" {
		// Discovery-only run: clone + discover projects
		// Use background context — the request context is cancelled when the handler returns.
		h.runner.StartDiscovery(context.Background(), run.ID)
	} else {
		// Full run: plan or apply
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
