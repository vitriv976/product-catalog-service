// Package e2e contains end-to-end tests for the product catalog service
package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/domain"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/update_product"
	"product-catalog-service/internal/services"
)

const (
	testProject  = "test-project"
	testInstance = "test-instance"
	testDatabase = "test-database"
)

var (
	spannerClient *spanner.Client
	container     *services.Container
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Set emulator host
	os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")

	// Create database
	databasePath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", testProject, testInstance, testDatabase)
	
	// Note: In real setup, you'd create the instance and database programmatically
	// For simplicity, we assume they exist or create them manually before running tests
	
	// Initialize container
	var err error
	container, err = services.NewContainer(ctx, databasePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize container: %v", err))
	}

	spannerClient = container.SpannerClient

	// Run tests
	code := m.Run()

	// Cleanup
	container.Close()

	os.Exit(code)
}

// TestProductCreationFlow tests the complete product creation flow
func TestProductCreationFlow(t *testing.T) {
	ctx := context.Background()

	// Create product
	productID, err := container.CreateProduct.Execute(ctx, create_product.Request{
		Name:             "Test Product",
		Description:      "A test product for E2E testing",
		Category:         "Electronics",
		PriceNumerator:   1999,
		PriceDenominator: 100,
	})
	require.NoError(t, err)
	require.NotEmpty(t, productID)

	// Verify product in database
	product, err := container.GetProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "Electronics", product.Category)
	assert.Equal(t, "draft", product.Status)

	// Verify outbox event was created
	events := getOutboxEvents(t, ctx, productID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].EventType)
}

// TestProductUpdateFlow tests updating product details
func TestProductUpdateFlow(t *testing.T) {
	ctx := context.Background()

	// Create product
	productID, err := container.CreateProduct.Execute(ctx, create_product.Request{
		Name:             "Original Name",
		Description:      "Original Description",
		Category:         "Category1",
		PriceNumerator:   1000,
		PriceDenominator: 100,
	})
	require.NoError(t, err)

	// Update product
	err = container.UpdateProduct.Execute(ctx, update_product.Request{
		ProductID:   productID,
		Name:        "Updated Name",
		Description: "Updated Description",
		Category:    "Category2",
	})
	require.NoError(t, err)

	// Verify changes
	product, err := container.GetProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", product.Name)
	assert.Equal(t, "Updated Description", product.Description)
	assert.Equal(t, "Category2", product.Category)

	// Verify outbox events
	events := getOutboxEvents(t, ctx, productID)
	require.GreaterOrEqual(t, len(events), 2)
	assert.Equal(t, "product.updated", events[len(events)-1].EventType)
}

// TestDiscountApplicationFlow tests applying and calculating discounts
func TestDiscountApplicationFlow(t *testing.T) {
	ctx := context.Background()

	// Create and activate product
	productID, err := container.CreateProduct.Execute(ctx, create_product.Request{
		Name:             "Discount Test Product",
		Description:      "Testing discounts",
		Category:         "Test",
		PriceNumerator:   10000, // $100.00
		PriceDenominator: 100,
	})
	require.NoError(t, err)

	// Activate product (required for discount)
	err = container.ActivateProduct.Execute(ctx, activate_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	// Apply 20% discount
	now := time.Now()
	err = container.ApplyDiscount.Execute(ctx, apply_discount.Request{
		ProductID:         productID,
		DiscountPercent:   20,
		DiscountStartDate: now.Add(-1 * time.Hour),
		DiscountEndDate:   now.Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Verify effective price is calculated correctly
	product, err := container.GetProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "100.00", product.BasePrice)
	assert.Equal(t, "80.00", product.EffectivePrice) // 20% off
	assert.NotNil(t, product.DiscountPercent)
	assert.Equal(t, 20, *product.DiscountPercent)

	// Verify discount event
	events := getOutboxEvents(t, ctx, productID)
	hasDiscountEvent := false
	for _, event := range events {
		if event.EventType == "discount.applied" {
			hasDiscountEvent = true
			break
		}
	}
	assert.True(t, hasDiscountEvent)
}

// TestActivationDeactivationFlow tests product activation and deactivation
func TestActivationDeactivationFlow(t *testing.T) {
	ctx := context.Background()

	// Create product
	productID, err := container.CreateProduct.Execute(ctx, create_product.Request{
		Name:             "State Test Product",
		Description:      "Testing state transitions",
		Category:         "Test",
		PriceNumerator:   5000,
		PriceDenominator: 100,
	})
	require.NoError(t, err)

	// Activate product
	err = container.ActivateProduct.Execute(ctx, activate_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	product, err := container.GetProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "active", product.Status)

	// Deactivate product
	err = container.DeactivateProduct.Execute(ctx, deactivate_product.Request{
		ProductID: productID,
	})
	require.NoError(t, err)

	product, err = container.GetProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "inactive", product.Status)
}

// TestBusinessRuleValidation tests domain business rules
func TestBusinessRuleValidation(t *testing.T) {
	ctx := context.Background()

	// Create inactive product
	productID, err := container.CreateProduct.Execute(ctx, create_product.Request{
		Name:             "Validation Test Product",
		Description:      "Testing validation rules",
		Category:         "Test",
		PriceNumerator:   1000,
		PriceDenominator: 100,
	})
	require.NoError(t, err)

	// Try to apply discount to inactive product (should fail)
	now := time.Now()
	err = container.ApplyDiscount.Execute(ctx, apply_discount.Request{
		ProductID:         productID,
		DiscountPercent:   10,
		DiscountStartDate: now,
		DiscountEndDate:   now.Add(24 * time.Hour),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrProductNotActive)
}

// TestListProducts tests the list products query
func TestListProducts(t *testing.T) {
	ctx := context.Background()

	// Create multiple products
	category := "ListTest"
	for i := 0; i < 3; i++ {
		productID, err := container.CreateProduct.Execute(ctx, create_product.Request{
			Name:             fmt.Sprintf("List Product %d", i+1),
			Description:      "For testing list",
			Category:         category,
			PriceNumerator:   1000 + int64(i*100),
			PriceDenominator: 100,
		})
		require.NoError(t, err)

		// Activate each product
		err = container.ActivateProduct.Execute(ctx, activate_product.Request{
			ProductID: productID,
		})
		require.NoError(t, err)
	}

	// List products by category
	response, err := container.ListProductsQuery.Execute(ctx, contracts.ListRequest{
		Category: category,
		Status:   "active",
		Limit:    10,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.Products), 3)
}

// Helper function to get outbox events for a product
func getOutboxEvents(t *testing.T, ctx context.Context, productID string) []struct {
	EventType   string
	AggregateID string
} {
	stmt := spanner.Statement{
		SQL: `SELECT event_type, aggregate_id FROM outbox_events WHERE aggregate_id = @productID ORDER BY created_at`,
		Params: map[string]interface{}{
			"productID": productID,
		},
	}

	txn := spannerClient.ReadOnlyTransaction()
	defer txn.Close()

	iter := txn.Query(ctx, stmt)
	defer iter.Stop()

	var events []struct {
		EventType   string
		AggregateID string
	}

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, err)

		var event struct {
			EventType   string
			AggregateID string
		}
		err = row.Columns(&event.EventType, &event.AggregateID)
		require.NoError(t, err)

		events = append(events, event)
	}

	return events
}
