// Package product implements gRPC handlers for product service
package product

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"product-catalog-service/internal/app/product/contracts"
	pb "product-catalog-service/proto/product/v1"
)

// mapDTOToProto maps a ProductDTO to a protobuf Product message
func mapDTOToProto(dto *contracts.ProductDTO) *pb.Product {
	product := &pb.Product{
		Id:             dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		BasePrice:      dto.BasePrice,
		EffectivePrice: dto.EffectivePrice,
		Status:         dto.Status,
		CreatedAt:      timestamppb.New(dto.CreatedAt),
		UpdatedAt:      timestamppb.New(dto.UpdatedAt),
	}

	if dto.DiscountPercent != nil {
		product.DiscountPercent = proto.Int32(int32(*dto.DiscountPercent))
	}

	if dto.DiscountStartDate != nil {
		product.DiscountStartDate = timestamppb.New(*dto.DiscountStartDate)
	}

	if dto.DiscountEndDate != nil {
		product.DiscountEndDate = timestamppb.New(*dto.DiscountEndDate)
	}

	return product
}
