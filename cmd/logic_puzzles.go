package cmd

import (
	"cmp"
	"fmt"
)

// Incrementer is a generic interface for types that track an incrementing value
// with a defined maximum, supporting reset and current-value inspection.
type Incrementer[T any] interface {
	Increment()
	IsMaxed() bool
	Reset()
	CurrentValue() T
	Name() string
}

// SliceIncrementer is an Incrementer[T] backed by a predefined slice of values.
// It steps through the slice in order. IsMaxed returns true when the index is
// at the last element of the slice.
type SliceIncrementer[T any] struct {
	name   string
	values []T
	index  int
}

func (s *SliceIncrementer[T]) Increment()      { s.index++ }
func (s *SliceIncrementer[T]) IsMaxed() bool   { return s.index >= len(s.values)-1 }
func (s *SliceIncrementer[T]) Reset()          { s.index = 0 }
func (s *SliceIncrementer[T]) CurrentValue() T { return s.values[s.index] }
func (s *SliceIncrementer[T]) Name() string    { return s.name }

// Odometer is a composite Incrementer whose value is a slice of individual
// Incrementer values. It implements Incrementer[[]T]: its CurrentValue returns
// the collected values of all digits. Incrementing works like a physical
// odometer — carries propagate from the rightmost digit leftward.
type Odometer[T any] struct {
	name   string
	names  []string
	digits []Incrementer[T]
	vars   map[string]T
}

func (o *Odometer[T]) Name() string { return o.name }

// Increment advances the odometer by one. Starting from the rightmost digit,
// if a digit is already maxed it is reset and the carry moves left. The first
// non-maxed digit found is incremented and the carry stops. If all digits are
// maxed, Increment is a no-op.
func (o *Odometer[T]) Increment() {
	if o.IsMaxed() {
		return
	}
	for i := len(o.digits) - 1; i >= 0; i-- {
		if !o.digits[i].IsMaxed() {
			o.digits[i].Increment()
			break
		}
		o.digits[i].Reset()
	}
}

// IsMaxed returns true when every digit has reached its maximum value.
func (o *Odometer[T]) IsMaxed() bool {
	for _, d := range o.digits {
		if !d.IsMaxed() {
			return false
		}
	}
	return len(o.digits) > 0
}

// Reset resets all digits to their initial values.
func (o *Odometer[T]) Reset() {
	for _, d := range o.digits {
		d.Reset()
	}
}

// CurrentValue returns a slice of each digit's current value, ordered from
// leftmost to rightmost.
func (o *Odometer[T]) CurrentValue() []T {
	values := make([]T, len(o.digits))
	for i, d := range o.digits {
		values[i] = d.CurrentValue()
	}
	return values
}

// OdometerVars returns a name→value snapshot of an Odometer's current state.
// The map is allocated once on the first call and reused on every subsequent
// call, avoiding per-iteration heap allocations in solver loops.
func OdometerVars[T any](o *Odometer[T]) map[string]T {
	if o.vars == nil {
		o.vars = make(map[string]T, len(o.digits))
	}
	for i, d := range o.digits {
		o.vars[o.names[i]] = d.CurrentValue()
	}
	return o.vars
}

// ComparisonOp is the operator used in a binary Constraint.
type ComparisonOp int

const (
	OpLT  ComparisonOp = iota // <
	OpGT                      // >
	OpLTE                     // <=
	OpGTE                     // >=
	OpEQ                      // ==
	OpNEQ                     // !=
)

// Constraint is a binary ordering rule between two named Incrementer values.
// T must be an ordered type (int, float64, string, etc.) so that all six
// comparison operators are well-defined.
// Left and Right are Incrementer names; Op is the comparison operator.
type Constraint[T cmp.Ordered] struct {
	Left  string
	Op    ComparisonOp
	Right string
}

// Check evaluates the constraint against a name→value snapshot. It returns an
// error if either variable name is not present in vars.
func (c Constraint[T]) Check(vars map[string]T) (bool, error) {
	l, lok := vars[c.Left]
	r, rok := vars[c.Right]
	if !lok {
		return false, fmt.Errorf("constraint references unknown variable %q", c.Left)
	}
	if !rok {
		return false, fmt.Errorf("constraint references unknown variable %q", c.Right)
	}
	switch c.Op {
	case OpLT:
		return l < r, nil
	case OpGT:
		return l > r, nil
	case OpLTE:
		return l <= r, nil
	case OpGTE:
		return l >= r, nil
	case OpEQ:
		return l == r, nil
	case OpNEQ:
		return l != r, nil
	default:
		return false, nil
	}
}

// CheckAll returns true only if every constraint in the slice is satisfied by
// vars. It returns an error if any constraint references an unknown variable.
func CheckAll[T cmp.Ordered](constraints []Constraint[T], vars map[string]T) (bool, error) {
	for _, c := range constraints {
		ok, err := c.Check(vars)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
