package config

import (
	"os"
	"testing"
)

func TestParseRepos(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"myorg/infra", 1, false},
		{"myorg/infra,myorg/platform", 2, false},
		{"bad-format", 0, true},
		{"/nope", 0, true},
	}
	for _, tt := range tests {
		repos, err := parseRepos(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseRepos(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if len(repos) != tt.want {
			t.Errorf("parseRepos(%q) got %d repos, want %d", tt.input, len(repos), tt.want)
		}
	}
}

func TestLoadDevMode(t *testing.T) {
	os.Setenv("DEV_MODE", "true")
	os.Setenv("REPOS", "org/repo1,org/repo2")
	defer os.Unsetenv("DEV_MODE")
	defer os.Unsetenv("REPOS")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.DevMode {
		t.Error("expected DevMode true")
	}
	if len(cfg.Repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(cfg.Repos))
	}
}
