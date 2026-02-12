// Package m_product provides the product table model
package m_product

// Table and field constants
const (
	TableName = "products"

	// Field names
	ProductID             = "product_id"
	Name                  = "name"
	Description           = "description"
	Category              = "category"
	BasePriceNumerator    = "base_price_numerator"
	BasePriceDenominator  = "base_price_denominator"
	DiscountPercent       = "discount_percent"
	DiscountStartDate     = "discount_start_date"
	DiscountEndDate       = "discount_end_date"
	Status                = "status"
	CreatedAt             = "created_at"
	UpdatedAt             = "updated_at"
	ArchivedAt            = "archived_at"
)
