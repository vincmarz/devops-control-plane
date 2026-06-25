package workflow

import "context"

type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Run(ctx context.Context, changeID string) error {
	// TODO: orchestrate GitLab -> Tekton -> Argo CD -> Evidence workflow.
	return nil
}
