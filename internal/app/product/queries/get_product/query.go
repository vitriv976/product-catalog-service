// Package get_product implements the get product query
package get_product

import (
	"context"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
	"product-catalog-service/internal/models/m_product"
	"product-catalog-service/internal/pkg/clock"
)

// Query implements the get product query
type Query struct {
	client            *spanner.Client
	pricingCalculator *services.PricingCalculator
	clock             clock.Clock
}

// NewQuery creates a new get product query
func NewQuery(client *spanner.Client, clock clock.Clock) *Query {
	return &Query{
		client:            client,
		pricingCalculator: services.NewPricingCalculator(),
		clock:             clock,
	}
}

// Execute retrieves a product by ID with effective price
func (q *Query) Execute(ctx context.Context, productID string) (*contracts.ProductDTO, error) {
	txn := q.client.ReadOnlyTransaction()
	defer txn.Close()

	model := m_product.NewModel()
	data, err := model.ReadByID(ctx, txn, productID)
	if err != nil {
		return nil, err
	}

	// Calculate effective price
	basePrice := domain.NewMoney(data.BasePriceNumerator, data.BasePriceDenominator)
	
	var discount *domain.Discount
	if data.DiscountPercent.Valid {
		discount, _ = domain.NewDiscount(
			int(data.DiscountPercent.Int64),
			data.DiscountStartDate.Time,
			data.DiscountEndDate.Time,
		)
	}

	effectivePrice := q.pricingCalculator.CalculateEffectivePrice(basePrice, discount, q.clock.Now())

	dto := &contracts.ProductDTO{
		ID:             data.ProductID,
		Name:           data.Name,
		Description:    data.Description,
		Category:       data.Category,
		BasePrice:      basePrice.String(),
		EffectivePrice: effectivePrice.String(),
		Status:         data.Status,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
	}

	// Add discount info if present and active
	if discount != nil && discount.IsActive(q.clock.Now()) {
		percent := discount.Percentage()
		startDate := discount.StartDate()
		endDate := discount.EndDate()
		dto.DiscountPercent = &percent
		dto.DiscountStartDate = &startDate
		dto.DiscountEndDate = &endDate
	}

	return dto, nil
}
