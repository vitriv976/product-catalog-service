// Package product implements gRPC handlers for product service
package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "product-catalog-service/proto/product/v1"

	"product-catalog-service/internal/app/product/contracts"
	"product-catalog-service/internal/app/product/queries/get_product"
	"product-catalog-service/internal/app/product/queries/list_products"
	"product-catalog-service/internal/app/product/usecases/activate_product"
	"product-catalog-service/internal/app/product/usecases/apply_discount"
	"product-catalog-service/internal/app/product/usecases/create_product"
	"product-catalog-service/internal/app/product/usecases/deactivate_product"
	"product-catalog-service/internal/app/product/usecases/remove_discount"
	"product-catalog-service/internal/app/product/usecases/update_product"
)

// Handler implements the ProductService gRPC server
type Handler struct {
	pb.UnimplementedProductServiceServer
	
	// Commands (use cases)
	createProduct     *create_product.Interactor
	updateProduct     *update_product.Interactor
	activateProduct   *activate_product.Interactor
	deactivateProduct *deactivate_product.Interactor
	applyDiscount     *apply_discount.Interactor
	removeDiscount    *remove_discount.Interactor
	
	// Queries
	getProductQuery  *get_product.Query
	listProductsQuery *list_products.Query
}

// NewHandler creates a new product gRPC handler
func NewHandler(
	createProduct *create_product.Interactor,
	updateProduct *update_product.Interactor,
	activateProduct *activate_product.Interactor,
	deactivateProduct *deactivate_product.Interactor,
	applyDiscount *apply_discount.Interactor,
	removeDiscount *remove_discount.Interactor,
	getProductQuery *get_product.Query,
	listProductsQuery *list_products.Query,
) *Handler {
	return &Handler{
		createProduct:     createProduct,
		updateProduct:     updateProduct,
		activateProduct:   activateProduct,
		deactivateProduct: deactivateProduct,
		applyDiscount:     applyDiscount,
		removeDiscount:    removeDiscount,
		getProductQuery:   getProductQuery,
		listProductsQuery: listProductsQuery,
	}
}

// CreateProduct creates a new product
func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductReply, error) {
	// 1. Validate proto request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Category == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	if req.PriceDenominator == 0 {
		return nil, status.Error(codes.InvalidArgument, "price_denominator cannot be zero")
	}

	// 2. Map proto to application request
	appReq := create_product.Request{
		Name:             req.Name,
		Description:      req.Description,
		Category:         req.Category,
		PriceNumerator:   req.PriceNumerator,
		PriceDenominator: req.PriceDenominator,
	}

	// 3. Call usecase
	productID, err := h.createProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	// 4. Return response
	return &pb.CreateProductReply{
		ProductId: productID,
	}, nil
}

// UpdateProduct updates product details
func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductReply, error) {
	// Validate
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Category == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}

	// Map and execute
	appReq := update_product.Request{
		ProductID:   req.ProductId,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	}

	err := h.updateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.UpdateProductReply{Success: true}, nil
}

// ActivateProduct activates a product
func (h *Handler) ActivateProduct(ctx context.Context, req *pb.ActivateProductRequest) (*pb.ActivateProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := activate_product.Request{
		ProductID: req.ProductId,
	}

	err := h.activateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ActivateProductReply{Success: true}, nil
}

// DeactivateProduct deactivates a product
func (h *Handler) DeactivateProduct(ctx context.Context, req *pb.DeactivateProductRequest) (*pb.DeactivateProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := deactivate_product.Request{
		ProductID: req.ProductId,
	}

	err := h.deactivateProduct.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.DeactivateProductReply{Success: true}, nil
}

// ApplyDiscount applies a discount to a product
func (h *Handler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.DiscountPercent < 0 || req.DiscountPercent > 100 {
		return nil, status.Error(codes.InvalidArgument, "discount_percent must be between 0 and 100")
	}
	if req.StartDate == nil || req.EndDate == nil {
		return nil, status.Error(codes.InvalidArgument, "start_date and end_date are required")
	}

	appReq := apply_discount.Request{
		ProductID:         req.ProductId,
		DiscountPercent:   int(req.DiscountPercent),
		DiscountStartDate: req.StartDate.AsTime(),
		DiscountEndDate:   req.EndDate.AsTime(),
	}

	err := h.applyDiscount.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.ApplyDiscountReply{Success: true}, nil
}

// RemoveDiscount removes a discount from a product
func (h *Handler) RemoveDiscount(ctx context.Context, req *pb.RemoveDiscountRequest) (*pb.RemoveDiscountReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	appReq := remove_discount.Request{
		ProductID: req.ProductId,
	}

	err := h.removeDiscount.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &pb.RemoveDiscountReply{Success: true}, nil
}

// GetProduct retrieves a product by ID
func (h *Handler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductReply, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	dto, err := h.getProductQuery.Execute(ctx, req.ProductId)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	product := mapDTOToProto(dto)

	return &pb.GetProductReply{
		Product: product,
	}, nil
}

// ListProducts retrieves a list of products
func (h *Handler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	appReq := contracts.ListRequest{
		Category: req.Category,
		Status:   req.Status,
		Limit:    int(req.Limit),
		Offset:   int(req.Offset),
	}

	response, err := h.listProductsQuery.Execute(ctx, appReq)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	products := make([]*pb.Product, 0, len(response.Products))
	for _, dto := range response.Products {
		products = append(products, mapDTOToProto(dto))
	}

	return &pb.ListProductsReply{
		Products: products,
		Total:    int32(response.Total),
	}, nil
}
