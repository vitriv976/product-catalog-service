// Package m_product provides the product table model
package m_product

import (
	"context"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

// Data represents a product row in the database
// Spanner struct tags map Go fields to table column names (snake_case)
type Data struct {
	ProductID            string           `spanner:"product_id"`
	Name                 string           `spanner:"name"`
	Description          string           `spanner:"description"`
	Category             string           `spanner:"category"`
	BasePriceNumerator   int64            `spanner:"base_price_numerator"`
	BasePriceDenominator int64            `spanner:"base_price_denominator"`
	DiscountPercent      spanner.NullInt64 `spanner:"discount_percent"`
	DiscountStartDate    spanner.NullTime  `spanner:"discount_start_date"`
	DiscountEndDate      spanner.NullTime  `spanner:"discount_end_date"`
	Status               string           `spanner:"status"`
	CreatedAt            time.Time        `spanner:"created_at"`
	UpdatedAt            time.Time        `spanner:"updated_at"`
	ArchivedAt           spanner.NullTime  `spanner:"archived_at"`
}

// Model provides typed Spanner mutation operations
type Model struct{}

// NewModel creates a new product model
func NewModel() *Model {
	return &Model{}
}

// InsertMut creates an insert mutation
func (m *Model) InsertMut(data *Data) *spanner.Mutation {
	return spanner.InsertOrUpdate(TableName, []string{
		ProductID,
		Name,
		Description,
		Category,
		BasePriceNumerator,
		BasePriceDenominator,
		DiscountPercent,
		DiscountStartDate,
		DiscountEndDate,
		Status,
		CreatedAt,
		UpdatedAt,
		ArchivedAt,
	}, []interface{}{
		data.ProductID,
		data.Name,
		data.Description,
		data.Category,
		data.BasePriceNumerator,
		data.BasePriceDenominator,
		data.DiscountPercent,
		data.DiscountStartDate,
		data.DiscountEndDate,
		data.Status,
		data.CreatedAt,
		data.UpdatedAt,
		data.ArchivedAt,
	})
}

// UpdateMut creates an update mutation for specific fields
func (m *Model) UpdateMut(productID string, updates map[string]interface{}) *spanner.Mutation {
	if len(updates) == 0 {
		return nil
	}

	// Always include the primary key
	updates[ProductID] = productID

	// Convert map to columns and values
	columns := make([]string, 0, len(updates))
	values := make([]interface{}, 0, len(updates))

	// Ensure product_id is first
	columns = append(columns, ProductID)
	values = append(values, productID)

	for col, val := range updates {
		if col != ProductID {
			columns = append(columns, col)
			values = append(values, val)
		}
	}

	return spanner.Update(TableName, columns, values)
}

// DeleteMut creates a delete mutation
func (m *Model) DeleteMut(productID string) *spanner.Mutation {
	return spanner.Delete(TableName, spanner.Key{productID})
}

// ReadByID reads a product by ID
func (m *Model) ReadByID(ctx context.Context, txn *spanner.ReadOnlyTransaction, productID string) (*Data, error) {
	row, err := txn.ReadRow(ctx, TableName, spanner.Key{productID}, []string{
		ProductID,
		Name,
		Description,
		Category,
		BasePriceNumerator,
		BasePriceDenominator,
		DiscountPercent,
		DiscountStartDate,
		DiscountEndDate,
		Status,
		CreatedAt,
		UpdatedAt,
		ArchivedAt,
	})
	if err != nil {
		return nil, err
	}

	var data Data
	err = row.ToStruct(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// ListByCategory lists products by category
func (m *Model) ListByCategory(ctx context.Context, txn *spanner.ReadOnlyTransaction, category string, limit int) ([]*Data, error) {
	stmt := spanner.Statement{
		SQL: `SELECT product_id, name, description, category, base_price_numerator, base_price_denominator,
			discount_percent, discount_start_date, discount_end_date, status, created_at, updated_at, archived_at
			FROM products
			WHERE category = @category AND status = @status
			ORDER BY created_at DESC
			LIMIT @limit`,
		Params: map[string]interface{}{
			"category": category,
			"status":   "active",
			"limit":    limit,
		},
	}

	iter := txn.Query(ctx, stmt)
	defer iter.Stop()

	var results []*Data
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var data Data
		if err := row.ToStruct(&data); err != nil {
			return nil, err
		}
		results = append(results, &data)
	}

	return results, nil
}
