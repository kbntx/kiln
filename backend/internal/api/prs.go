package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kbntx/kiln/internal/github"
)

type prsHandlers struct {
	ghClient github.GitHubClient
}

func (h *prsHandlers) handleList(w http.ResponseWriter, r *http.Request) {
	owner := chi.URLParam(r, "owner")
	repo := chi.URLParam(r, "repo")

	prs, err := h.ghClient.ListOpenPRs(r.Context(), owner, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prs)
}
