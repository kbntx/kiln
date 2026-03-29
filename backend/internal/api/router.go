package api

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/kbntx/kiln/internal/auth"
	"github.com/kbntx/kiln/internal/config"
	"github.com/kbntx/kiln/internal/github"
	"github.com/kbntx/kiln/internal/runner"
	"github.com/kbntx/kiln/internal/stream"
	"github.com/kbntx/kiln/static"
)

type Dependencies struct {
	Config   *config.Config
	Sessions *auth.SessionStore
	OAuth    *auth.OAuthHandler
	GHClient github.GitHubClient
	Runner   *runner.Runner
	Store    *runner.RunStore
	Broker   *stream.Broker
}

func NewRouter(deps *Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	authH := &authHandlers{oauth: deps.OAuth, sessions: deps.Sessions}
	reposH := &reposHandlers{repos: deps.Config.Repos}
	prsH := &prsHandlers{ghClient: deps.GHClient}
	runsH := &runsHandlers{runner: deps.Runner, store: deps.Store, broker: deps.Broker}

	// Auth routes (no auth middleware)
	r.Get("/auth/login", authH.handleLogin)
	r.Get("/auth/callback", authH.handleCallback)
	r.Get("/auth/logout", authH.handleLogout)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})
		r.Get("/me", authH.handleMe)

		// Protected
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(deps.Sessions))
			r.Get("/repos", reposH.handleList)
			r.Get("/repos/{owner}/{repo}/prs", prsH.handleList)
			r.Post("/runs", runsH.handleCreate)
			r.Get("/runs/{id}", runsH.handleGet)
			r.Get("/runs/{id}/stream", runsH.handleStream)
		})
	})

	// Serve embedded frontend
	frontendFS, err := fs.Sub(static.FrontendDist, "dist")
	if err != nil {
		slog.Error("failed to create frontend sub-fs", "error", err)
	}
	fileServer := http.FileServer(http.FS(frontendFS))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(frontendFS, path); err != nil {
			path = "index.html"
		}
		r.URL.Path = "/" + path
		fileServer.ServeHTTP(w, r)
	})

	return r
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
