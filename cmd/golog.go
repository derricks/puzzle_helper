package cmd 

import (
	"github.com/spf13/cobra"
)

// Incrementer is an interface that defines methods for incrementing and managing a value
type Incrementer[T any] interface {
	// Name returns the name of the incrementer
	Name() string
	// IsMaxed returns true if the incrementer has reached its maximum value
	IsMaxed() bool
	
	// Increment increases the current value and returns the new value
	Increment() T
	
	// Reset sets the incrementer back to its initial state
	Reset()
	
	// GetCurrentValue returns the current value without incrementing
	GetCurrentValue() T
}

// SliceIncrementer implements Incrementer for a slice of any type
type SliceIncrementer[T any] struct {
	name      string
	values    []T
	currIndex int
}

// NewSliceIncrementer creates a new SliceIncrementer with the provided name and values
func NewSliceIncrementer[T any](name string, values []T) *SliceIncrementer[T] {
	return &SliceIncrementer[T]{
		name:      name,
		values:    values,
		currIndex: -1, // Start at -1 so first Increment() returns index 0
	}
}

// Name returns the name of the incrementer
func (si *SliceIncrementer[T]) Name() string {
	return si.name
}

// IsMaxed returns true if we've reached the end of the slice
func (si *SliceIncrementer[T]) IsMaxed() bool {
	return si.currIndex >= len(si.values)-1
}

// Increment moves to the next value in the slice and returns it
// If already at the end, returns the last value
func (si *SliceIncrementer[T]) Increment() T {
	if !si.IsMaxed() {
		si.currIndex++
	}
	return si.values[si.currIndex]
}

// Reset sets the index back to the start
func (si *SliceIncrementer[T]) Reset() {
	si.currIndex = -1
}

// GetCurrentValue returns the current value without incrementing
func (si *SliceIncrementer[T]) GetCurrentValue() T {
	if si.currIndex == -1 {
		return si.values[0]
	}
	return si.values[si.currIndex]
}

// IncrementerIncrementer implements Incrementer and is backed by a slice of Incrementers
// It acts like an odometer: incrementing the last, and rolling over to the next when maxed
// T is the return type of the innermost Incrementer

type IncrementerIncrementer[T any] struct {
	name        string
	incrementers []Incrementer[T]
}

// NewIncrementerIncrementer creates a new IncrementerIncrementer with the provided name and slice of Incrementers
func NewIncrementerIncrementer[T any](name string, incrementers []Incrementer[T]) *IncrementerIncrementer[T] {
	return &IncrementerIncrementer[T]{
		name:         name,
		incrementers: incrementers,
	}
}

// Name returns the name of the IncrementerIncrementer
func (ii *IncrementerIncrementer[T]) Name() string {
	return ii.name
}

// IsMaxed returns true if all incrementers are maxed
func (ii *IncrementerIncrementer[T]) IsMaxed() bool {
	for _, inc := range ii.incrementers {
		if !inc.IsMaxed() {
			return false
		}
	}
	return true
}

// Increment increments the last incrementer, rolling over as needed
func (ii *IncrementerIncrementer[T]) Increment() T {
	n := len(ii.incrementers)
	if n == 0 {
		var zero T
		return zero
	}
	for i := n - 1; i >= 0; i-- {
		if !ii.incrementers[i].IsMaxed() {
			val := ii.incrementers[i].Increment()
			// Reset all incrementers after i
			for j := i + 1; j < n; j++ {
				ii.incrementers[j].Reset()
			}
			return val
		}
	}
	// If all are maxed, just return the last one's current value
	return ii.incrementers[n-1].GetCurrentValue()
}

// Reset resets all incrementers
func (ii *IncrementerIncrementer[T]) Reset() {
	for _, inc := range ii.incrementers {
		inc.Reset()
	}
}

// GetCurrentValue returns the current value of the last incrementer
func (ii *IncrementerIncrementer[T]) GetCurrentValue() T {
	n := len(ii.incrementers)
	if n == 0 {
		var zero T
		return zero
	}
	return ii.incrementers[n-1].GetCurrentValue()
}

// GetAllIncrementerValues returns a map of incrementer names to their current values
func (ii *IncrementerIncrementer[T]) GetAllIncrementerValues() map[string]T {
	result := make(map[string]T)
	for _, inc := range ii.incrementers {
		result[inc.Name()] = inc.GetCurrentValue()
	}
	return result
}

// gologCmd represents the golog command
var gologCmd = &cobra.Command{
	Use:   "golog",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Add your command logic here
	},
}

func init() {
	rootCmd.AddCommand(gologCmd)
}

