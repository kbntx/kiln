package config

import (
	"fmt"
	"os"
	"strings"
)

type Repo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type Config struct {
	Port               string
	DevMode            bool
	DevRepoDir         string
	Repos              []Repo
	SessionSecret      string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubToken        string
	AllowedOrg         string
	BaseURL            string
	LogLevel           string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               envOr("PORT", "8080"),
		DevMode:            os.Getenv("DEV_MODE") == "true",
		DevRepoDir:         os.Getenv("DEV_REPO_DIR"),
		SessionSecret:      envOr("SESSION_SECRET", "dev-secret-change-in-prod"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		AllowedOrg:         os.Getenv("ALLOWED_ORG"),
		BaseURL:            envOr("BASE_URL", "http://localhost:8080"),
		LogLevel:           envOr("LOG_LEVEL", "info"),
	}

	repos, err := parseRepos(os.Getenv("REPOS"))
	if err != nil {
		return nil, err
	}
	cfg.Repos = repos

	if !cfg.DevMode {
		if cfg.GitHubClientID == "" || cfg.GitHubClientSecret == "" {
			return nil, fmt.Errorf("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET required when DEV_MODE is not true")
		}
		if cfg.AllowedOrg == "" {
			return nil, fmt.Errorf("ALLOWED_ORG required when DEV_MODE is not true")
		}
	}

	return cfg, nil
}

func parseRepos(raw string) ([]Repo, error) {
	if raw == "" {
		return nil, nil
	}
	var repos []Repo
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		parts := strings.SplitN(entry, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid repo format %q, expected owner/name", entry)
		}
		repos = append(repos, Repo{Owner: parts[0], Name: parts[1]})
	}
	return repos, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
