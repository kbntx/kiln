package engine

import (
	"context"
	"time"
)

type RunOptions struct {
	WorkDir string
	Stack   string
	EnvVars map[string]string
}

type LogLine struct {
	Stream string    `json:"stream"`
	Text   string    `json:"text"`
	Time   time.Time `json:"time"`
}

type Engine interface {
	Init(ctx context.Context, opts RunOptions) error
	Plan(ctx context.Context, opts RunOptions, output chan<- LogLine) error
	Apply(ctx context.Context, opts RunOptions, output chan<- LogLine) error
	Name() string
}
