package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kbntx/kiln/internal/api"
	"github.com/kbntx/kiln/internal/auth"
	"github.com/kbntx/kiln/internal/config"
	"github.com/kbntx/kiln/internal/devmode"
	"github.com/kbntx/kiln/internal/engine"
	"github.com/kbntx/kiln/internal/git"
	"github.com/kbntx/kiln/internal/github"
	"github.com/kbntx/kiln/internal/runner"
	"github.com/kbntx/kiln/internal/stream"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	// Wire dependencies based on dev mode
	var ghClient github.GitHubClient
	var workspace git.WorkspaceManager
	var eng engine.Engine

	if cfg.DevMode {
		slog.Info("running in dev mode")
		ghClient = &devmode.MockGitHubClient{}
		workspace = devmode.NewMockWorkspace("testdata/fake-infra")
		eng = nil // auto-detect real engine per project (Terraform/Pulumi)
	} else {
		ghClient = github.NewRealClient(cfg.GitHubToken)
		workspace = git.NewRealWorkspace("/tmp/kiln", cfg.GitHubToken)
		eng = nil // auto-detect per run
	}

	sessions := auth.NewSessionStore(cfg.SessionSecret)
	oauth := auth.NewOAuthHandler(
		cfg.GitHubClientID,
		cfg.GitHubClientSecret,
		cfg.BaseURL,
		cfg.AllowedOrg,
		sessions,
		ghClient,
		cfg.DevMode,
	)

	broker := stream.NewBroker()
	store := runner.NewRunStore()
	rnr := runner.NewRunner(store, workspace, ghClient, eng, broker)

	deps := &api.Dependencies{
		Config:   cfg,
		Sessions: sessions,
		OAuth:    oauth,
		GHClient: ghClient,
		Runner:   rnr,
		Store:    store,
		Broker:   broker,
	}

	router := api.NewRouter(deps)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("kiln starting", "port", cfg.Port, "dev_mode", cfg.DevMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
