package auth

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

// UserKey is the context key used to store the authenticated session.
const UserKey contextKey = "user"

// RequireAuth returns middleware that checks for a valid session cookie.
// If the session is missing or invalid it responds with 401 JSON.
// Otherwise it stores the Session in the request context under UserKey.
func RequireAuth(sessions *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessions.Get(r)
			if err != nil || session == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "unauthorized",
				})
				return
			}
			ctx := context.WithValue(r.Context(), UserKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext extracts the Session from the request context.
// Returns nil if no session is present.
func UserFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(UserKey).(*Session)
	return s
}
