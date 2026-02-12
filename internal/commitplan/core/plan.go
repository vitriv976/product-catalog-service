// Package core provides the core commitplan functionality for atomic transactions
package core

import (
	"context"
)

// Mutation represents a database mutation that can be applied atomically
type Mutation interface {
	Apply(ctx context.Context) error
}

// Plan represents a collection of mutations to be applied atomically
type Plan struct {
	mutations []Mutation
}

// NewPlan creates a new empty commit plan
func NewPlan() *Plan {
	return &Plan{
		mutations: make([]Mutation, 0),
	}
}

// Add adds a mutation to the plan
func (p *Plan) Add(mutation Mutation) {
	if mutation != nil {
		p.mutations = append(p.mutations, mutation)
	}
}

// Mutations returns all mutations in the plan
func (p *Plan) Mutations() []Mutation {
	return p.mutations
}

// IsEmpty returns true if the plan has no mutations
func (p *Plan) IsEmpty() bool {
	return len(p.mutations) == 0
}
