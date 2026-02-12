// Package apply_discount implements the apply discount use case
package apply_discount

import (
	"context"
	"time"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/repo"
	spannercp "product-catalog-service/internal/commitplan/spanner"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

// Request encapsulates apply discount parameters
type Request struct {
	ProductID        string
	DiscountPercent  int
	DiscountStartDate time.Time
	DiscountEndDate   time.Time
}

// Interactor implements the apply discount use case
type Interactor struct {
	client      *spanner.Client
	productRepo contracts.ProductRepository
	outboxRepo  contracts.OutboxRepository
	committer   *committer.Committer
	clock       clock.Clock
}

// NewInteractor creates a new apply discount interactor
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

// Execute applies a discount to a product following the Golden Mutation Pattern
func (it *Interactor) Execute(ctx context.Context, req Request) error {
	// 1. Load aggregate
	txn := it.client.ReadOnlyTransaction()
	defer txn.Close()

	product, err := it.productRepo.FindByID(ctx, txn, req.ProductID)
	if err != nil {
		return err
	}

	// 2. Create discount value object
	discount, err := domain.NewDiscount(req.DiscountPercent, req.DiscountStartDate, req.DiscountEndDate)
	if err != nil {
		return err
	}

	// 3. Call domain method
	if err := product.ApplyDiscount(discount, it.clock.Now()); err != nil {
		return err
	}

	// 4. Build commit plan
	plan := committer.NewPlan()

	// 5. Get mutations from repository
	if mut := it.productRepo.UpdateMut(product); mut != nil {
		plan.Add(spannercp.NewSpannerMutation(mut))
	}

	// 6. Add outbox events
	for _, event := range product.DomainEvents() {
		outboxEvent := repo.EnrichDomainEvent(event, product.ID())
		if outboxMut := it.outboxRepo.InsertMut(outboxEvent); outboxMut != nil {
			plan.Add(spannercp.NewSpannerMutation(outboxMut))
		}
	}

	// 7. Apply plan
	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}
