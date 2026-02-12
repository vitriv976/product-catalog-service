// Package deactivate_product implements the deactivate product use case
package deactivate_product

import (
	"context"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/repo"
	spannercp "product-catalog-service/internal/commitplan/spanner"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

// Request encapsulates deactivate product parameters
type Request struct {
	ProductID string
}

// Interactor implements the deactivate product use case
type Interactor struct {
	client      *spanner.Client
	productRepo contracts.ProductRepository
	outboxRepo  contracts.OutboxRepository
	committer   *committer.Committer
	clock       clock.Clock
}

// NewInteractor creates a new deactivate product interactor
func NewInteractor(
	client *spanner.Client,
	productRepo contracts.ProductRepository,
	outboxRepo contracts.OutboxRepository,
	committer *committer.Committer,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		client:      client,
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		committer:   committer,
		clock:       clock,
	}
}

// Execute deactivates a product following the Golden Mutation Pattern
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate
	txn := it.client.ReadOnlyTransaction()
	defer txn.Close()

	product, err := it.productRepo.FindByID(ctx, txn, req.ProductID)
	if err != nil {
		return err
	}

	// 2. Call domain method
	if err := product.Deactivate(it.clock.Now()); err != nil {
		return err
	}

	// 3. Build commit plan
	plan := committer.NewPlan()

	// 4. Get mutations from repository
	if mut := it.productRepo.UpdateMut(product); mut != nil {
		plan.Add(spannercp.NewSpannerMutation(mut))
	}

	// 5. Add outbox events
	for _, event := range product.DomainEvents() {
		outboxEvent := repo.EnrichDomainEvent(event, product.ID())
		if outboxMut := it.outboxRepo.InsertMut(outboxEvent); outboxMut != nil {
			plan.Add(spannercp.NewSpannerMutation(outboxMut))
		}
	}

	// 6. Apply plan
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}
