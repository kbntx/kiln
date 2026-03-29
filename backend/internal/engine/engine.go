package engine

import (
	"context"
	"time"
)

type RunOptions struct {
	WorkDir          string
	Stack            string
	EnvVars          map[string]string
	Destroy          bool
	PlanFile         string // path to saved plan file (used by apply)
	TerraformVersion string // if set, use this specific terraform version
}

type LogLine struct {
	Stream string    `json:"stream"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
}

type Engine interface {
	Init(ctx context.Context, opts RunOptions, output chan<- LogLine) error
	Plan(ctx context.Context, opts RunOptions, output chan<- LogLine) error
	Apply(ctx context.Context, opts RunOptions, output chan<- LogLine) error
	HasChanges(ctx context.Context, opts RunOptions) (bool, error)
	Name() string
}
