// Package domain contains pure business logic
package domain

import "errors"

// Sentinel errors for domain validation
var (
	ErrProductNotActive        = errors.New("product is not active")
	ErrProductNotInactive      = errors.New("product is not inactive")
	ErrInvalidDiscountPeriod   = errors.New("invalid discount period")
	ErrInvalidDiscountPercent  = errors.New("discount percent must be between 0 and 100")
	ErrProductAlreadyArchived  = errors.New("product is already archived")
	ErrEmptyProductName        = errors.New("product name cannot be empty")
	ErrEmptyCategory           = errors.New("category cannot be empty")
	ErrNilPrice                = errors.New("price cannot be nil")
	ErrNegativePrice           = errors.New("price cannot be negative")
)
