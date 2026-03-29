package api

import (
	"encoding/json"
	"net/http"

	"github.com/kbntx/kiln/internal/auth"
)

type authHandlers struct {
	oauth    *auth.OAuthHandler
	sessions *auth.SessionStore
}

func (h *authHandlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	h.oauth.HandleLogin(w, r)
}

func (h *authHandlers) handleCallback(w http.ResponseWriter, r *http.Request) {
	h.oauth.HandleCallback(w, r)
}

func (h *authHandlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Clear(w)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (h *authHandlers) handleMe(w http.ResponseWriter, r *http.Request) {
	session, err := h.sessions.Get(r)
	if err != nil || session == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"login":  session.Login,
		"avatar": session.Avatar,
	})
}
