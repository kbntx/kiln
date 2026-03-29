package discovery

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// discoverPulumi walks rootDir looking for Pulumi.yaml files and returns
// a Project for each one found.
func discoverPulumi(rootDir string) ([]Project, error) {
	var projects []Project

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() != "Pulumi.yaml" {
			return nil
		}

		dir := filepath.Dir(path)
		relDir, err := filepath.Rel(rootDir, dir)
		if err != nil {
			return err
		}

		name, err := parsePulumiName(path)
		if err != nil {
			return err
		}

		stacks, err := findPulumiStacks(dir)
		if err != nil {
			return err
		}

		projects = append(projects, Project{
			Name:   name,
			Dir:    relDir,
			Engine: EngineTypePulumi,
			Stacks: stacks,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// parsePulumiName reads a Pulumi.yaml file and extracts the "name:" field.
func parsePulumiName(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "name:") {
			value := strings.TrimPrefix(line, "name:")
			return strings.TrimSpace(value), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return filepath.Base(filepath.Dir(path)), nil
}

// findPulumiStacks looks for Pulumi.*.yaml files in the given directory
// and returns the stack names extracted from the filenames.
func findPulumiStacks(dir string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "Pulumi.*.yaml"))
	if err != nil {
		return nil, err
	}

	var stacks []string
	for _, m := range matches {
		base := filepath.Base(m)
		// Pulumi.<stack>.yaml
		name := strings.TrimPrefix(base, "Pulumi.")
		name = strings.TrimSuffix(name, ".yaml")
		if name != "" {
			stacks = append(stacks, name)
		}
	}
	sort.Strings(stacks)
	return stacks, nil
}
