// Package clock provides time abstraction for testing
package clock

import "time"

// Clock provides current time
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using actual system time
type RealClock struct{}

// NewRealClock creates a new real clock
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current time
func (rc *RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock implements Clock with a fixed time (for testing)
type FixedClock struct {
	fixed time.Time
}

// NewFixedClock creates a new fixed clock for testing
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{fixed: t}
}

// Now returns the fixed time
func (fc *FixedClock) Now() time.Time {
	return fc.fixed
}
