// Package contracts defines repository interfaces
package contracts

import (
	"context"
	"time"

	"cloud.google.com/go/spanner"
)

// ProductReadModel defines read operations for queries
type ProductReadModel interface {
	// GetByID retrieves a product with effective price
	GetByID(ctx context.Context, txn *spanner.ReadOnlyTransaction, id string) (*ProductDTO, error)

	// List retrieves products with filtering and pagination
	List(ctx context.Context, txn *spanner.ReadOnlyTransaction, req ListRequest) (*ListResponse, error)
}

// ProductDTO is a data transfer object for product queries
type ProductDTO struct {
	ID             string
	Name           string
	Description    string
	Category       string
	BasePrice      string
	EffectivePrice string
	DiscountPercent *int
	DiscountStartDate *time.Time
	DiscountEndDate *time.Time
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ListRequest encapsulates list query parameters
type ListRequest struct {
	Category string
	Status   string
	Limit    int
	Offset   int
}

// ListResponse encapsulates list query results
type ListResponse struct {
	Products []*ProductDTO
	Total    int
}
