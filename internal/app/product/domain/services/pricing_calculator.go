// Package services contains domain services
package services

import (
	"math/big"
	"time"

	"product-catalog-service/internal/app/product/domain"
)

// PricingCalculator is a domain service for calculating effective prices
type PricingCalculator struct{}

// NewPricingCalculator creates a new pricing calculator
func NewPricingCalculator() *PricingCalculator {
	return &PricingCalculator{}
}

// CalculateEffectivePrice calculates the effective price considering active discounts
func (pc *PricingCalculator) CalculateEffectivePrice(
	basePrice *domain.Money,
	discount *domain.Discount,
	now time.Time,
) *domain.Money {
	if basePrice == nil {
		return nil
	}

	// No discount or discount not active
	if discount == nil || !discount.IsActive(now) {
		return basePrice
	}

	// Calculate discount amount
	discountPercentage := big.NewRat(int64(discount.Percentage()), 100)
	discountAmount := basePrice.Multiply(discountPercentage)

	// Calculate effective price
	effectivePrice := basePrice.Subtract(discountAmount)

	return effectivePrice
}
