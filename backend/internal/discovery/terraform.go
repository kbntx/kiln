package discovery

import (
	"os"
	"path/filepath"
	"strings"
)

// discoverTerraform walks rootDir looking for directories that contain *.tf
// files and returns a Project for each one found.
func discoverTerraform(rootDir string) ([]Project, error) {
	seen := make(map[string]bool)
	var projects []Project

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .terraform directories entirely.
		if info.IsDir() && info.Name() == ".terraform" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".tf") {
			return nil
		}

		dir := filepath.Dir(path)
		if seen[dir] {
			return nil
		}
		seen[dir] = true

		relDir, err := filepath.Rel(rootDir, dir)
		if err != nil {
			return err
		}

		name := filepath.Base(dir)

		stacks := []string{"default"}

		projects = append(projects, Project{
			Name:   name,
			Dir:    relDir,
			Engine: EngineTypeTerraform,
			Stacks: stacks,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}
