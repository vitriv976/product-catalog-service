// Package domain contains pure business logic
package domain

// Field constants for change tracking
const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldBasePrice   = "base_price"
	FieldDiscount    = "discount"
	FieldStatus      = "status"
)

// ChangeTracker tracks which fields have been modified
type ChangeTracker struct {
	dirtyFields map[string]bool
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirtyFields: make(map[string]bool),
	}
}

// MarkDirty marks a field as dirty (changed)
func (ct *ChangeTracker) MarkDirty(field string) {
	if ct == nil {
		return
	}
	ct.dirtyFields[field] = true
}

// Dirty checks if a field is dirty
func (ct *ChangeTracker) Dirty(field string) bool {
	if ct == nil {
		return false
	}
	return ct.dirtyFields[field]
}

// HasChanges returns true if any fields are dirty
func (ct *ChangeTracker) HasChanges() bool {
	if ct == nil {
		return false
	}
	return len(ct.dirtyFields) > 0
}

// Clear clears all dirty flags
func (ct *ChangeTracker) Clear() {
	if ct == nil {
		return
	}
	ct.dirtyFields = make(map[string]bool)
}
