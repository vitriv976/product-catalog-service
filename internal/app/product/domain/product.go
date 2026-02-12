// Package domain contains pure business logic
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProductStatus represents the state of a product
type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// Product is the aggregate root for product management
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time
	changes     *ChangeTracker
	events      []DomainEvent
}

// NewProduct creates a new product (factory method)
func NewProduct(name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	// Validation
	if name == "" {
		return nil, ErrEmptyProductName
	}
	if category == "" {
		return nil, ErrEmptyCategory
	}
	if basePrice == nil {
		return nil, ErrNilPrice
	}
	if basePrice.IsNegative() {
		return nil, ErrNegativePrice
	}

	p := &Product{
		id:          uuid.New().String(),
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    nil,
		status:      ProductStatusDraft,
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
	}

	// Capture domain event
	p.events = append(p.events, &ProductCreatedEvent{
		ProductID:  p.id,
		Name:       p.name,
		Category:   p.category,
		BasePrice:  p.basePrice,
		OccurredAt: now,
	})

	return p, nil
}

// LoadProduct loads an existing product from the database (for updates)
func LoadProduct(
	id, name, description, category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	createdAt, updatedAt time.Time,
	archivedAt *time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(),
		events:      make([]DomainEvent, 0),
	}
}

// ID returns the product ID
func (p *Product) ID() string {
	return p.id
}

// Name returns the product name
func (p *Product) Name() string {
	return p.name
}

// Description returns the product description
func (p *Product) Description() string {
	return p.description
}

// Category returns the product category
func (p *Product) Category() string {
	return p.category
}

// BasePrice returns a copy of the base price
func (p *Product) BasePrice() *Money {
	return p.basePrice
}

// Discount returns the current discount (may be nil)
func (p *Product) Discount() *Discount {
	return p.discount
}

// Status returns the product status
func (p *Product) Status() ProductStatus {
	return p.status
}

// CreatedAt returns the creation timestamp
func (p *Product) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update timestamp
func (p *Product) UpdatedAt() time.Time {
	return p.updatedAt
}

// ArchivedAt returns the archival timestamp (may be nil)
func (p *Product) ArchivedAt() *time.Time {
	return p.archivedAt
}

// Changes returns the change tracker
func (p *Product) Changes() *ChangeTracker {
	return p.changes
}

// DomainEvents returns captured domain events
func (p *Product) DomainEvents() []DomainEvent {
	return p.events
}

// UpdateDetails updates product name, description, and/or category
func (p *Product) UpdateDetails(name, description, category string, now time.Time) error {
	if name == "" {
		return ErrEmptyProductName
	}
	if category == "" {
		return ErrEmptyCategory
	}

	if p.name != name {
		p.name = name
		p.changes.MarkDirty(FieldName)
	}

	if p.description != description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
	}

	if p.category != category {
		p.category = category
		p.changes.MarkDirty(FieldCategory)
	}

	if p.changes.HasChanges() {
		p.updatedAt = now
		p.events = append(p.events, &ProductUpdatedEvent{
			ProductID:  p.id,
			Name:       p.name,
			Category:   p.category,
			OccurredAt: now,
		})
	}

	return nil
}

// ApplyDiscount applies a discount to the product
func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}

	if !discount.IsValidAt(now) {
		return ErrInvalidDiscountPeriod
	}

	p.discount = discount
	p.changes.MarkDirty(FieldDiscount)
	p.updatedAt = now

	p.events = append(p.events, &DiscountAppliedEvent{
		ProductID:       p.id,
		DiscountPercent: discount.Percentage(),
		StartDate:       discount.StartDate(),
		EndDate:         discount.EndDate(),
		OccurredAt:      now,
	})

	return nil
}

// RemoveDiscount removes the current discount
func (p *Product) RemoveDiscount(now time.Time) {
	if p.discount == nil {
		return
	}

	p.discount = nil
	p.changes.MarkDirty(FieldDiscount)
	p.updatedAt = now

	p.events = append(p.events, &DiscountRemovedEvent{
		ProductID:  p.id,
		OccurredAt: now,
	})
}

// Activate activates the product
func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductAlreadyArchived
	}

	if p.status == ProductStatusActive {
		return nil // Already active
	}

	p.status = ProductStatusActive
	p.changes.MarkDirty(FieldStatus)
	p.updatedAt = now

	p.events = append(p.events, &ProductActivatedEvent{
		ProductID:  p.id,
		OccurredAt: now,
	})

	return nil
}

// Deactivate deactivates the product
func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductAlreadyArchived
	}

	if p.status == ProductStatusInactive {
		return nil // Already inactive
	}

	// Remove discount when deactivating
	if p.discount != nil {
		p.discount = nil
		p.changes.MarkDirty(FieldDiscount)
	}

	p.status = ProductStatusInactive
	p.changes.MarkDirty(FieldStatus)
	p.updatedAt = now

	p.events = append(p.events, &ProductDeactivatedEvent{
		ProductID:  p.id,
		OccurredAt: now,
	})

	return nil
}

// Archive archives the product (soft delete)
func (p *Product) Archive(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductAlreadyArchived
	}

	p.status = ProductStatusArchived
	p.archivedAt = &now
	p.changes.MarkDirty(FieldStatus)
	p.updatedAt = now

	return nil
}
