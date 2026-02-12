// Package contracts defines repository interfaces
package contracts

import (
	"context"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/domain"
)

// ProductRepository defines operations for product persistence
type ProductRepository interface {
	// InsertMut returns a mutation to insert a new product
	InsertMut(p *domain.Product) *spanner.Mutation

	// UpdateMut returns a mutation to update a product (only changed fields)
	UpdateMut(p *domain.Product) *spanner.Mutation

	// FindByID retrieves a product by ID
	FindByID(ctx context.Context, txn *spanner.ReadOnlyTransaction, id string) (*domain.Product, error)
}

// OutboxRepository defines operations for outbox events
type OutboxRepository interface {
	// InsertMut returns a mutation to insert an outbox event
	InsertMut(event *OutboxEvent) *spanner.Mutation
}

// OutboxEvent represents an enriched domain event for the outbox
type OutboxEvent struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     string
	Status      string
	CreatedAt   interface{}
}
