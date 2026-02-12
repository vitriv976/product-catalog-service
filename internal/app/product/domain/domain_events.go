// Package domain contains pure business logic
package domain

import "time"

// DomainEvent is a marker interface for domain events
type DomainEvent interface {
	EventType() string
}

// ProductCreatedEvent is emitted when a product is created
type ProductCreatedEvent struct {
	ProductID   string
	Name        string
	Category    string
	BasePrice   *Money
	OccurredAt  time.Time
}

func (e *ProductCreatedEvent) EventType() string {
	return "product.created"
}

// ProductUpdatedEvent is emitted when product details change
type ProductUpdatedEvent struct {
	ProductID   string
	Name        string
	Category    string
	OccurredAt  time.Time
}

func (e *ProductUpdatedEvent) EventType() string {
	return "product.updated"
}

// ProductActivatedEvent is emitted when a product is activated
type ProductActivatedEvent struct {
	ProductID  string
	OccurredAt time.Time
}

func (e *ProductActivatedEvent) EventType() string {
	return "product.activated"
}

// ProductDeactivatedEvent is emitted when a product is deactivated
type ProductDeactivatedEvent struct {
	ProductID  string
	OccurredAt time.Time
}

func (e *ProductDeactivatedEvent) EventType() string {
	return "product.deactivated"
}

// DiscountAppliedEvent is emitted when a discount is applied
type DiscountAppliedEvent struct {
	ProductID       string
	DiscountPercent int
	StartDate       time.Time
	EndDate         time.Time
	OccurredAt      time.Time
}

func (e *DiscountAppliedEvent) EventType() string {
	return "discount.applied"
}

// DiscountRemovedEvent is emitted when a discount is removed
type DiscountRemovedEvent struct {
	ProductID  string
	OccurredAt time.Time
}

func (e *DiscountRemovedEvent) EventType() string {
	return "discount.removed"
}
