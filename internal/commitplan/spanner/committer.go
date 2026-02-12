// Package spanner provides Spanner-specific commitplan implementation
package spanner

import (
	"context"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/commitplan/core"
)

// SpannerMutation wraps a Spanner mutation to implement core.Mutation interface
type SpannerMutation struct {
	mutation *spanner.Mutation
}

// NewSpannerMutation creates a new Spanner mutation wrapper
func NewSpannerMutation(m *spanner.Mutation) *SpannerMutation {
	return &SpannerMutation{mutation: m}
}

// Apply is a placeholder - actual application happens in Committer
func (sm *SpannerMutation) Apply(ctx context.Context) error {
	return nil
}

// Mutation returns the underlying Spanner mutation
func (sm *SpannerMutation) Mutation() *spanner.Mutation {
	return sm.mutation
}

// Committer handles atomic application of commit plans to Spanner
type Committer struct {
	client *spanner.Client
}

// NewCommitter creates a new Spanner committer
func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{client: client}
}

// Apply applies a commit plan atomically to Spanner
func (c *Committer) Apply(ctx context.Context, plan *core.Plan) error {
	if plan.IsEmpty() {
		return nil
	}

	mutations := make([]*spanner.Mutation, 0, len(plan.Mutations()))
	for _, mut := range plan.Mutations() {
		if sm, ok := mut.(*SpannerMutation); ok {
			mutations = append(mutations, sm.Mutation())
		}
	}

	if len(mutations) == 0 {
		return nil
	}

	_, err := c.client.Apply(ctx, mutations)
	return err
}
