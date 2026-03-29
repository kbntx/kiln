package api

import (
	"encoding/json"
	"net/http"

	"github.com/kbntx/kiln/internal/config"
)

type reposHandlers struct {
	repos []config.Repo
}

func (h *reposHandlers) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.repos)
}
