package discovery

import (
	"sort"
)

// EngineType represents the type of IaC engine used by a project.
type EngineType string

const (
	EngineTypePulumi    EngineType = "pulumi"
	EngineTypeTerraform EngineType = "terraform"
)

// Project represents a discovered IaC project.
type Project struct {
	Name   string     `json:"name"`
	Dir    string     `json:"dir"`
	Engine EngineType `json:"engine"`
	Stacks []string   `json:"stacks"`
}

// DiscoverProjects walks rootDir and returns all discovered IaC projects
// (both Pulumi and Terraform), sorted by Dir.
func DiscoverProjects(rootDir string) ([]Project, error) {
	var all []Project

	pulumi, err := discoverPulumi(rootDir)
	if err != nil {
		return nil, err
	}
	all = append(all, pulumi...)

	tf, err := discoverTerraform(rootDir)
	if err != nil {
		return nil, err
	}
	all = append(all, tf...)

	sort.Slice(all, func(i, j int) bool {
		return all[i].Dir < all[j].Dir
	})

	return all, nil
}
