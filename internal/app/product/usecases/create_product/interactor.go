// Package create_product implements the create product use case
package create_product

import (
	"context"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/commitplan/spanner"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
)

// Request encapsulates create product parameters
type Request struct {
	Name             string
	Description      string
	Category         string
	PriceNumerator   int64
	PriceDenominator int64
}

// Interactor implements the create product use case
type Interactor struct {
	productRepo contracts.ProductRepository
	outboxRepo  contracts.OutboxRepository
	committer   *committer.Committer
	clock       clock.Clock
}

// NewInteractor creates a new create product interactor
func NewInteractor(
	productRepo contracts.ProductRepository,
	outboxRepo contracts.OutboxRepository,
	committer *committer.Committer,
	clock clock.Clock,
) *Interactor {
	return &Interactor{
		productRepo: productRepo,
		outboxRepo:  outboxRepo,
		committer:   committer,
		clock:       clock,
	}
}

// Execute creates a new product following the Golden Mutation Pattern
func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	// 1. Create aggregate
	basePrice := domain.NewMoney(req.PriceNumerator, req.PriceDenominator)
	product, err := domain.NewProduct(
		req.Name,
		req.Description,
		req.Category,
		basePrice,
		it.clock.Now(),
	)
	if err != nil {
		return "", err
	}

	// 2. Domain validation (already done in constructor)

	// 3. Build commit plan
	plan := committer.NewPlan()

	// 4. Get mutations from repository
	if mut := it.productRepo.InsertMut(product); mut != nil {
		plan.Add(spanner.NewSpannerMutation(mut))
	}

	// 5. Add outbox events
	for _, event := range product.DomainEvents() {
		outboxEvent := repo.EnrichDomainEvent(event, product.ID())
		if outboxMut := it.outboxRepo.InsertMut(outboxEvent); outboxMut != nil {
			plan.Add(spanner.NewSpannerMutation(outboxMut))
		}
	}

	// 6. Apply plan (usecase applies, NOT handler!)
	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}

	return product.ID(), nil
}
