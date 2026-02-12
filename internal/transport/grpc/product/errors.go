// Package product implements gRPC handlers for product service
package product

import (
	"errors"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"product-catalog-service/internal/app/product/domain"
)

// mapDomainErrorToGRPC maps domain errors to gRPC status codes
func mapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	// Domain errors
	switch {
	case errors.Is(err, domain.ErrProductNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrProductNotInactive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrInvalidDiscountPeriod):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidDiscountPercent):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrProductAlreadyArchived):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrEmptyProductName):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrEmptyCategory):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNilPrice):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNegativePrice):
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Spanner errors - NotFound when row doesn't exist
	if spanner.ErrCode(err) == codes.NotFound {
		return status.Error(codes.NotFound, "product not found")
	}

	// Default - include original error message for debugging
	return status.Error(codes.Internal, "internal server error: "+err.Error())
}
