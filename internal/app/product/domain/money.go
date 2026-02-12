// Package domain contains pure business logic
package domain

import (
	"math/big"
)

// Money represents a monetary amount using precise decimal arithmetic
type Money struct {
	amount *big.Rat
}

// NewMoney creates a new Money value from numerator and denominator
func NewMoney(numerator, denominator int64) *Money {
	return &Money{
		amount: big.NewRat(numerator, denominator),
	}
}

// NewMoneyFromRat creates a new Money value from a big.Rat
func NewMoneyFromRat(rat *big.Rat) *Money {
	if rat == nil {
		return nil
	}
	// Create a copy to avoid external mutations
	newRat := new(big.Rat).Set(rat)
	return &Money{amount: newRat}
}

// Amount returns a copy of the underlying big.Rat
func (m *Money) Amount() *big.Rat {
	if m == nil || m.amount == nil {
		return nil
	}
	return new(big.Rat).Set(m.amount)
}

// Numerator returns the numerator for database storage
func (m *Money) Numerator() int64 {
	if m == nil || m.amount == nil {
		return 0
	}
	return m.amount.Num().Int64()
}

// Denominator returns the denominator for database storage
func (m *Money) Denominator() int64 {
	if m == nil || m.amount == nil {
		return 1
	}
	return m.amount.Denom().Int64()
}

// IsNegative returns true if the amount is negative
func (m *Money) IsNegative() bool {
	if m == nil || m.amount == nil {
		return false
	}
	return m.amount.Sign() < 0
}

// IsZero returns true if the amount is zero
func (m *Money) IsZero() bool {
	if m == nil || m.amount == nil {
		return true
	}
	return m.amount.Sign() == 0
}

// Multiply returns a new Money that is this Money multiplied by the given percentage (0-1)
func (m *Money) Multiply(percentage *big.Rat) *Money {
	if m == nil || m.amount == nil || percentage == nil {
		return nil
	}
	result := new(big.Rat).Mul(m.amount, percentage)
	return &Money{amount: result}
}

// Subtract returns a new Money that is this Money minus the given Money
func (m *Money) Subtract(other *Money) *Money {
	if m == nil || m.amount == nil {
		return nil
	}
	if other == nil || other.amount == nil {
		return &Money{amount: new(big.Rat).Set(m.amount)}
	}
	result := new(big.Rat).Sub(m.amount, other.amount)
	return &Money{amount: result}
}

// String returns a string representation of the Money
func (m *Money) String() string {
	if m == nil || m.amount == nil {
		return "0"
	}
	return m.amount.FloatString(2)
}
