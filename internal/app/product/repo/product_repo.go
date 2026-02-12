// Package repo implements repository patterns
package repo

import (
	"context"
	"time"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/models/m_product"
)

// ProductRepo implements the product repository
type ProductRepo struct {
	model *m_product.Model
}

// NewProductRepo creates a new product repository
func NewProductRepo() *ProductRepo {
	return &ProductRepo{
		model: m_product.NewModel(),
	}
}

// InsertMut returns a mutation to insert a new product
func (r *ProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	data := &m_product.Data{
		ProductID:            p.ID(),
		Name:                 p.Name(),
		Description:          p.Description(),
		Category:             p.Category(),
		BasePriceNumerator:   p.BasePrice().Numerator(),
		BasePriceDenominator: p.BasePrice().Denominator(),
		Status:               string(p.Status()),
		CreatedAt:            p.CreatedAt(),
		UpdatedAt:            p.UpdatedAt(),
	}

	// Handle discount if present
	if discount := p.Discount(); discount != nil {
		data.DiscountPercent = spanner.NullInt64{Int64: int64(discount.Percentage()), Valid: true}
		data.DiscountStartDate = spanner.NullTime{Time: discount.StartDate(), Valid: true}
		data.DiscountEndDate = spanner.NullTime{Time: discount.EndDate(), Valid: true}
	}

	// Handle archived_at if present
	if archivedAt := p.ArchivedAt(); archivedAt != nil {
		data.ArchivedAt = spanner.NullTime{Time: *archivedAt, Valid: true}
	}

	return r.model.InsertMut(data)
}

// UpdateMut returns a mutation to update a product (only changed fields)
func (r *ProductRepo) UpdateMut(p *domain.Product) *spanner.Mutation {
	updates := make(map[string]interface{})

	// Only include changed fields
	if p.Changes().Dirty(domain.FieldName) {
		updates[m_product.Name] = p.Name()
	}

	if p.Changes().Dirty(domain.FieldDescription) {
		updates[m_product.Description] = p.Description()
	}

	if p.Changes().Dirty(domain.FieldCategory) {
		updates[m_product.Category] = p.Category()
	}

	if p.Changes().Dirty(domain.FieldBasePrice) {
		updates[m_product.BasePriceNumerator] = p.BasePrice().Numerator()
		updates[m_product.BasePriceDenominator] = p.BasePrice().Denominator()
	}

	if p.Changes().Dirty(domain.FieldDiscount) {
		if discount := p.Discount(); discount != nil {
			updates[m_product.DiscountPercent] = spanner.NullInt64{Int64: int64(discount.Percentage()), Valid: true}
			updates[m_product.DiscountStartDate] = spanner.NullTime{Time: discount.StartDate(), Valid: true}
			updates[m_product.DiscountEndDate] = spanner.NullTime{Time: discount.EndDate(), Valid: true}
		} else {
			// Clear discount
			updates[m_product.DiscountPercent] = spanner.NullInt64{Valid: false}
			updates[m_product.DiscountStartDate] = spanner.NullTime{Valid: false}
			updates[m_product.DiscountEndDate] = spanner.NullTime{Valid: false}
		}
	}

	if p.Changes().Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(p.Status())
		if archivedAt := p.ArchivedAt(); archivedAt != nil {
			updates[m_product.ArchivedAt] = spanner.NullTime{Time: *archivedAt, Valid: true}
		}
	}

	// Always update updated_at if there are changes
	if len(updates) > 0 {
		updates[m_product.UpdatedAt] = p.UpdatedAt()
	}

	if len(updates) == 0 {
		return nil // No changes
	}

	return r.model.UpdateMut(p.ID(), updates)
}

// FindByID retrieves a product by ID
func (r *ProductRepo) FindByID(ctx context.Context, txn *spanner.ReadOnlyTransaction, id string) (*domain.Product, error) {
	data, err := r.model.ReadByID(ctx, txn, id)
	if err != nil {
		return nil, err
	}

	// Map database model to domain
	basePrice := domain.NewMoney(data.BasePriceNumerator, data.BasePriceDenominator)

	var discount *domain.Discount
	if data.DiscountPercent.Valid {
		discount, _ = domain.NewDiscount(
			int(data.DiscountPercent.Int64),
			data.DiscountStartDate.Time,
			data.DiscountEndDate.Time,
		)
	}

	var archivedAt *time.Time
	if data.ArchivedAt.Valid {
		archivedAt = &data.ArchivedAt.Time
	}

	product := domain.LoadProduct(
		data.ProductID,
		data.Name,
		data.Description,
		data.Category,
		basePrice,
		discount,
		domain.ProductStatus(data.Status),
		data.CreatedAt,
		data.UpdatedAt,
		archivedAt,
	)

	return product, nil
}
