package discovery

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Profile defines a named set of environment variables injected into the engine subprocess.
type Profile struct {
	Env map[string]string `yaml:"env"`
}

// ProjectConfig represents a project entry in kiln.yaml.
type ProjectConfig struct {
	Name             string   `yaml:"name"`
	Dir              string   `yaml:"dir"`
	Engine           string   `yaml:"engine"`
	Stacks           []string `yaml:"stacks"`
	Profile          string   `yaml:"profile"`
	TerraformVersion string   `yaml:"terraform_version,omitempty"`
}

// Config represents the top-level kiln.yaml configuration.
type Config struct {
	Profiles         map[string]Profile `yaml:"profiles"`
	Projects         []ProjectConfig    `yaml:"projects"`
	TerraformVersion string             `yaml:"terraform_version,omitempty"` // global default
}

// ParseConfig parses a kiln.yaml file and returns the configuration.
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse kiln.yaml: %w", err)
	}

	if len(cfg.Projects) == 0 {
		return nil, fmt.Errorf("parse kiln.yaml: no projects defined")
	}

	for i, p := range cfg.Projects {
		if p.Name == "" {
			return nil, fmt.Errorf("parse kiln.yaml: project %d has no name", i)
		}
		if p.Engine == "" {
			return nil, fmt.Errorf("parse kiln.yaml: project %q has no engine", p.Name)
		}
		// TODO(pulumi): Add "pulumi" back when Pulumi support is implemented.
		if p.Engine != "terraform" {
			return nil, fmt.Errorf("parse kiln.yaml: project %q has unsupported engine %q (only terraform is supported)", p.Name, p.Engine)
		}
		if len(p.Stacks) == 0 {
			return nil, fmt.Errorf("parse kiln.yaml: project %q has no stacks", p.Name)
		}
	}

	// Validate that every terraform project has a version (global or per-project).
	for _, p := range cfg.Projects {
		if p.Engine == "terraform" && p.TerraformVersion == "" && cfg.TerraformVersion == "" {
			return nil, fmt.Errorf("parse kiln.yaml: project %q requires terraform_version (set per-project or globally)", p.Name)
		}
	}

	return &cfg, nil
}

// ToProjects converts config projects to the downstream Project type.
func (c *Config) ToProjects() []Project {
	projects := make([]Project, len(c.Projects))
	for i, p := range c.Projects {
		tfVersion := p.TerraformVersion
		if tfVersion == "" {
			tfVersion = c.TerraformVersion
		}
		projects[i] = Project{
			Name:             p.Name,
			Dir:              p.Dir,
			Engine:           EngineType(p.Engine),
			Stacks:           p.Stacks,
			Profile:          p.Profile,
			TerraformVersion: tfVersion,
		}
	}
	return projects
}

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// ResolveProfileEnv builds a slice of KEY=VALUE strings from a profile's env map.
// Values containing ${VAR} are resolved from the server's environment.
// Minimal system vars (PATH, HOME, TERM) are always included.
func ResolveProfileEnv(profile Profile) []string {
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"TERM=" + os.Getenv("TERM"),
	}

	for k, v := range profile.Env {
		resolved := envVarPattern.ReplaceAllStringFunc(v, func(match string) string {
			varName := envVarPattern.FindStringSubmatch(match)[1]
			return os.Getenv(varName)
		})
		env = append(env, k+"="+resolved)
	}

	return env
}
