// Package committer provides typed CommitPlan wrapper
package committer

import (
	"context"

	"product-catalog-service/internal/commitplan/core"
	"product-catalog-service/internal/commitplan/spanner"
)

// Committer wraps the spanner committer for easier usage
type Committer struct {
	spannerCommitter *spanner.Committer
}

// New creates a new typed committer
func New(sc *spanner.Committer) *Committer {
	return &Committer{spannerCommitter: sc}
}

// Apply applies a commit plan atomically
func (c *Committer) Apply(ctx context.Context, plan *core.Plan) error {
	return c.spannerCommitter.Apply(ctx, plan)
}

// NewPlan creates a new commit plan
func NewPlan() *core.Plan {
	return core.NewPlan()
}
