// Package services provides dependency injection container
package services

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/repo"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
	spannercp "product-catalog-service/internal/commitplan/spanner"
	"product-catalog-service/internal/pkg/clock"
	"product-catalog-service/internal/pkg/committer"
	"product-catalog-service/internal/transport/grpc/product"
)

// Container holds all service dependencies
type Container struct {
	// Infrastructure
	SpannerClient *spanner.Client
	Committer     *committer.Committer
	Clock         clock.Clock

	// Repositories
	ProductRepo contracts.ProductRepository
	OutboxRepo  contracts.OutboxRepository

	// Use cases
	CreateProduct     *create_product.Interactor
	UpdateProduct     *update_product.Interactor
	ActivateProduct   *activate_product.Interactor
	DeactivateProduct *deactivate_product.Interactor
	ApplyDiscount     *apply_discount.Interactor
	RemoveDiscount    *remove_discount.Interactor

	// Queries
	GetProductQuery   *get_product.Query
	ListProductsQuery *list_products.Query

	// Handlers
	ProductHandler *product.Handler
}

// NewContainer creates and wires up all dependencies
func NewContainer(ctx context.Context, spannerDatabase string) (*Container, error) {
	// Create Spanner client
	client, err := spanner.NewClient(ctx, spannerDatabase)
	if err != nil {
		return nil, fmt.Errorf("failed to create spanner client: %w", err)
	}

	// Create infrastructure components
	spannerCommitter := spannercp.NewCommitter(client)
	comm := committer.New(spannerCommitter)
	clk := clock.NewRealClock()

	// Create repositories
	productRepo := repo.NewProductRepo()
	outboxRepo := repo.NewOutboxRepo()

	// Create use cases
	createProductUsecase := create_product.NewInteractor(productRepo, outboxRepo, comm, clk)
	updateProductUsecase := update_product.NewInteractor(client, productRepo, outboxRepo, comm, clk)
	activateProductUsecase := activate_product.NewInteractor(client, productRepo, outboxRepo, comm, clk)
	deactivateProductUsecase := deactivate_product.NewInteractor(client, productRepo, outboxRepo, comm, clk)
	applyDiscountUsecase := apply_discount.NewInteractor(client, productRepo, outboxRepo, comm, clk)
	removeDiscountUsecase := remove_discount.NewInteractor(client, productRepo, outboxRepo, comm, clk)

	// Create queries
	getProductQuery := get_product.NewQuery(client, clk)
	listProductsQuery := list_products.NewQuery(client, clk)

	// Create handlers
	productHandler := product.NewHandler(
		createProductUsecase,
		updateProductUsecase,
		activateProductUsecase,
		deactivateProductUsecase,
		applyDiscountUsecase,
		removeDiscountUsecase,
		getProductQuery,
		listProductsQuery,
	)

	return &Container{
		SpannerClient:     client,
		Committer:         comm,
		Clock:             clk,
		ProductRepo:       productRepo,
		OutboxRepo:        outboxRepo,
		CreateProduct:     createProductUsecase,
		UpdateProduct:     updateProductUsecase,
		ActivateProduct:   activateProductUsecase,
		DeactivateProduct: deactivateProductUsecase,
		ApplyDiscount:     applyDiscountUsecase,
		RemoveDiscount:    removeDiscountUsecase,
		GetProductQuery:   getProductQuery,
		ListProductsQuery: listProductsQuery,
		ProductHandler:    productHandler,
	}, nil
}

// Close closes all resources
func (c *Container) Close() {
	if c.SpannerClient != nil {
		c.SpannerClient.Close()
	}
}
