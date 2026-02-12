// Package list_products implements the list products query
package list_products

import (
	"context"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/domain/services"
	"product-catalog-service/internal/pkg/clock"
)

// Query implements the list products query
type Query struct {
	client            *spanner.Client
	pricingCalculator *services.PricingCalculator
	clock             clock.Clock
}

// NewQuery creates a new list products query
func NewQuery(client *spanner.Client, clock clock.Clock) *Query {
	return &Query{
		client:            client,
		pricingCalculator: services.NewPricingCalculator(),
		clock:             clock,
	}
}

// Execute retrieves a list of products with filtering and pagination
func (q *Query) Execute(ctx context.Context, req contracts.ListRequest) (*contracts.ListResponse, error) {
	txn := q.client.ReadOnlyTransaction()
	defer txn.Close()

	// Build query
	sql := `SELECT product_id, name, description, category, base_price_numerator, base_price_denominator,
		discount_percent, discount_start_date, discount_end_date, status, created_at, updated_at
		FROM products
		WHERE 1=1`

	params := make(map[string]interface{})

	if req.Category != "" {
		sql += " AND category = @category"
		params["category"] = req.Category
	}

	if req.Status != "" {
		sql += " AND status = @status"
		params["status"] = req.Status
	}
	// When status is empty, list all products (including draft) - useful for testing

	sql += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		sql += " LIMIT @limit"
		params["limit"] = req.Limit
	} else {
		// Default limit
		sql += " LIMIT @limit"
		params["limit"] = 50
	}

	if req.Offset > 0 {
		sql += " OFFSET @offset"
		params["offset"] = req.Offset
	}

	stmt := spanner.Statement{
		SQL:    sql,
		Params: params,
	}

	iter := txn.Query(ctx, stmt)
	defer iter.Stop()

	var products []*contracts.ProductDTO
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var (
			productID            string
			name                 string
			description          string
			category             string
			basePriceNumerator   int64
			basePriceDenominator int64
			discountPercent      spanner.NullInt64
			discountStartDate    spanner.NullTime
			discountEndDate      spanner.NullTime
			status               string
			createdAt            spanner.NullTime
			updatedAt            spanner.NullTime
		)

		err = row.Columns(
			&productID,
			&name,
			&description,
			&category,
			&basePriceNumerator,
			&basePriceDenominator,
			&discountPercent,
			&discountStartDate,
			&discountEndDate,
			&status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Calculate effective price
		basePrice := domain.NewMoney(basePriceNumerator, basePriceDenominator)

		var discount *domain.Discount
		if discountPercent.Valid {
			discount, _ = domain.NewDiscount(
				int(discountPercent.Int64),
				discountStartDate.Time,
				discountEndDate.Time,
			)
		}

		effectivePrice := q.pricingCalculator.CalculateEffectivePrice(basePrice, discount, q.clock.Now())

		dto := &contracts.ProductDTO{
			ID:             productID,
			Name:           name,
			Description:    description,
			Category:       category,
			BasePrice:      basePrice.String(),
			EffectivePrice: effectivePrice.String(),
			Status:         status,
			CreatedAt:      createdAt.Time,
			UpdatedAt:      updatedAt.Time,
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

		products = append(products, dto)
	}

	return &contracts.ListResponse{
		Products: products,
		Total:    len(products),
	}, nil
}
