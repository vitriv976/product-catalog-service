// Package domain contains pure business logic
package domain

import (
	"time"
)

// Discount represents a percentage-based discount with validity period
type Discount struct {
	percentage int     // Percentage as integer (0-100)
	startDate  time.Time
	endDate    time.Time
}

// NewDiscount creates a new discount with validation
func NewDiscount(percentage int, startDate, endDate time.Time) (*Discount, error) {
	if percentage < 0 || percentage > 100 {
		return nil, ErrInvalidDiscountPercent
	}

	if !endDate.After(startDate) {
		return nil, ErrInvalidDiscountPeriod
	}

	return &Discount{
		percentage: percentage,
		startDate:  startDate,
		endDate:    endDate,
	}, nil
}

// Percentage returns the discount percentage (0-100)
func (d *Discount) Percentage() int {
	if d == nil {
		return 0
	}
	return d.percentage
}

// StartDate returns the discount start date
func (d *Discount) StartDate() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.startDate
}

// EndDate returns the discount end date
func (d *Discount) EndDate() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.endDate
}

// IsValidAt checks if the discount is valid at the given time
func (d *Discount) IsValidAt(t time.Time) bool {
	if d == nil {
		return false
	}
	return !t.Before(d.startDate) && !t.After(d.endDate)
}

// IsActive checks if the discount is currently active
func (d *Discount) IsActive(now time.Time) bool {
	return d.IsValidAt(now)
}
