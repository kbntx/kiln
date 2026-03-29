package discovery

// EngineType represents the type of IaC engine used by a project.
type EngineType string

const (
	EngineTypeTerraform EngineType = "terraform"
	// TODO(pulumi): Uncomment when Pulumi support is implemented.
	// EngineTypePulumi EngineType = "pulumi"
)

// Project represents an IaC project defined in kiln.yaml.
type Project struct {
	Name             string     `json:"name"`
	Dir              string     `json:"dir"`
	Engine           EngineType `json:"engine"`
	Stacks           []string   `json:"stacks"`
	Profile          string     `json:"profile"`
	TerraformVersion string     `json:"terraformVersion,omitempty"`
}
