package cmd

import (
	"testing"
)

func TestNewSliceIncrementer(t *testing.T) {
	values := []int{1, 2, 3}
	inc := NewSliceIncrementer("test", values)

	if inc == nil {
		t.Error("NewSliceIncrementer returned nil")
	}

	if inc.currIndex != -1 {
		t.Errorf("Initial currIndex should be -1, got %d", inc.currIndex)
	}

	if len(inc.values) != len(values) {
		t.Errorf("Expected values length %d, got %d", len(values), len(inc.values))
	}
}

func TestSliceIncrementer_IsMaxed(t *testing.T) {
	tests := []struct {
		name     string
		values   []int
		index    int
		expected bool
	}{
		{
			name:     "empty slice",
			values:   []int{},
			index:    -1,
			expected: true,
		},
		{
			name:     "at start",
			values:   []int{1, 2, 3},
			index:    -1,
			expected: false,
		},
		{
			name:     "in middle",
			values:   []int{1, 2, 3},
			index:    1,
			expected: false,
		},
		{
			name:     "at end",
			values:   []int{1, 2, 3},
			index:    2,
			expected: true,
		},
		{
			name:     "single element",
			values:   []int{1},
			index:    0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inc := NewSliceIncrementer("test", tt.values)
			inc.currIndex = tt.index
			if got := inc.IsMaxed(); got != tt.expected {
				t.Errorf("IsMaxed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSliceIncrementer_Increment(t *testing.T) {
	tests := []struct {
		name           string
		values         []int
		incrementCount int
		expectedValue  int
		expectedIndex  int
	}{
		{
			name:           "first increment",
			values:         []int{1, 2, 3},
			incrementCount: 1,
			expectedValue:  1,
			expectedIndex:  0,
		},
		{
			name:           "middle increment",
			values:         []int{1, 2, 3},
			incrementCount: 2,
			expectedValue:  2,
			expectedIndex:  1,
		},
		{
			name:           "increment to end",
			values:         []int{1, 2, 3},
			incrementCount: 3,
			expectedValue:  3,
			expectedIndex:  2,
		},
		{
			name:           "increment beyond end",
			values:         []int{1, 2, 3},
			incrementCount: 4,
			expectedValue:  3,
			expectedIndex:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inc := NewSliceIncrementer("test", tt.values)
			var got int
			for i := 0; i < tt.incrementCount; i++ {
				got = inc.Increment()
			}
			if got != tt.expectedValue {
				t.Errorf("Increment() = %v, want %v", got, tt.expectedValue)
			}
			if inc.currIndex != tt.expectedIndex {
				t.Errorf("currIndex = %v, want %v", inc.currIndex, tt.expectedIndex)
			}
		})
	}
}

func TestSliceIncrementer_Reset(t *testing.T) {
	tests := []struct {
		name          string
		values        []int
		incrementsNum int
	}{
		{
			name:          "reset from start",
			values:        []int{1, 2, 3},
			incrementsNum: 0,
		},
		{
			name:          "reset from middle",
			values:        []int{1, 2, 3},
			incrementsNum: 2,
		},
		{
			name:          "reset from end",
			values:        []int{1, 2, 3},
			incrementsNum: 3,
		},
		{
			name:          "reset from beyond end",
			values:        []int{1, 2, 3},
			incrementsNum: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inc := NewSliceIncrementer("test", tt.values)
			for i := 0; i < tt.incrementsNum; i++ {
				inc.Increment()
			}
			inc.Reset()
			if inc.currIndex != -1 {
				t.Errorf("After Reset(), currIndex = %v, want -1", inc.currIndex)
			}
		})
	}
}

func TestSliceIncrementer_GetCurrentValue(t *testing.T) {
	tests := []struct {
		name           string
		values         []int
		incrementCount int
		expected       int
	}{
		{
			name:           "get initial value",
			values:         []int{1, 2, 3},
			incrementCount: 0,
			expected:      1,
		},
		{
			name:           "get after one increment",
			values:         []int{1, 2, 3},
			incrementCount: 1,
			expected:      1,
		},
		{
			name:           "get from middle",
			values:         []int{1, 2, 3},
			incrementCount: 2,
			expected:      2,
		},
		{
			name:           "get from end",
			values:         []int{1, 2, 3},
			incrementCount: 3,
			expected:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inc := NewSliceIncrementer("test", tt.values)
			for i := 0; i < tt.incrementCount; i++ {
				inc.Increment()
			}
			if got := inc.GetCurrentValue(); got != tt.expected {
				t.Errorf("GetCurrentValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSliceIncrementer_WithDifferentTypes(t *testing.T) {
	t.Run("string slice", func(t *testing.T) {
		inc := NewSliceIncrementer("test", []string{"a", "b", "c"})
		expected := []string{"a", "b", "c"}
		for i, want := range expected {
			got := inc.Increment()
			if got != want {
				t.Errorf("Increment() at index %d = %v, want %v", i, got, want)
			}
		}
	})

	t.Run("float slice", func(t *testing.T) {
		inc := NewSliceIncrementer("test", []float64{1.1, 2.2, 3.3})
		expected := []float64{1.1, 2.2, 3.3}
		for i, want := range expected {
			got := inc.Increment()
			if got != want {
				t.Errorf("Increment() at index %d = %v, want %v", i, got, want)
			}
		}
	})

	type custom struct {
		value int
	}
	t.Run("custom type slice", func(t *testing.T) {
		inc := NewSliceIncrementer("test", []custom{{1}, {2}, {3}})
		expected := []custom{{1}, {2}, {3}}
		for i, want := range expected {
			got := inc.Increment()
			if got.value != want.value {
				t.Errorf("Increment() at index %d = %v, want %v", i, got.value, want.value)
			}
		}
	})
}

func TestIncrementerIncrementer_Behavior(t *testing.T) {
	// Create three SliceIncrementers for digits 0-1, 0-2, 0-2
	inc1 := NewSliceIncrementer("ones", []int{0, 1})
	inc2 := NewSliceIncrementer("twos", []int{0, 1, 2})
	inc3 := NewSliceIncrementer("threes", []int{0, 1, 2})
	combo := NewIncrementerIncrementer("combo", []Incrementer[int]{inc1, inc2, inc3})

	// Helper to get the current state of all incrementers
	getState := func() (a, b, c int) {
		return inc1.GetCurrentValue(), inc2.GetCurrentValue(), inc3.GetCurrentValue()
	}

	// Initial state: all at 0
	combo.Reset()
	if a, b, c := getState(); a != 0 || b != 0 || c != 0 {
		t.Errorf("Initial state: got (%d,%d,%d), want (0,0,0)", a, b, c)
	}

	// Increment through all combinations
	seen := make(map[[3]int]bool)
	maxed := false
	for i := 0; i < 18; i++ { // 2*3*3 = 18 possible states
		val := combo.Increment()
		state := [3]int{inc1.GetCurrentValue(), inc2.GetCurrentValue(), inc3.GetCurrentValue()}
		seen[state] = true
		if combo.IsMaxed() {
			maxed = true
			break
		}
		_ = val // just to use val
	}
	if !maxed {
		t.Error("Expected combo to be maxed after all combinations")
	}
	if len(seen) != 18 {
		t.Errorf("Expected 18 unique states, got %d", len(seen))
	}

	// After maxed, further increments should not change state
	finalState := [3]int{inc1.GetCurrentValue(), inc2.GetCurrentValue(), inc3.GetCurrentValue()}
	for i := 0; i < 3; i++ {
		combo.Increment()
		state := [3]int{inc1.GetCurrentValue(), inc2.GetCurrentValue(), inc3.GetCurrentValue()}
		if state != finalState {
			t.Errorf("State changed after maxed: got %v, want %v", state, finalState)
		}
	}

	// Test Reset
	combo.Reset()
	if a, b, c := getState(); a != 0 || b != 0 || c != 0 {
		t.Errorf("After Reset: got (%d,%d,%d), want (0,0,0)", a, b, c)
	}

	// Test GetCurrentValue returns the last incrementer's value
	combo.Reset()
	inc3.Increment() // ones:0, twos:0, threes:1
	if got := combo.GetCurrentValue(); got != 1 {
		t.Errorf("GetCurrentValue() = %v, want 1", got)
	}
}

func TestIncrementerIncrementer_GetAllIncrementerValues(t *testing.T) {
	inc1 := NewSliceIncrementer("ones", []int{0, 1})
	inc2 := NewSliceIncrementer("twos", []int{0, 1, 2})
	inc3 := NewSliceIncrementer("threes", []int{0, 1, 2})
	combo := NewIncrementerIncrementer("combo", []Incrementer[int]{inc1, inc2, inc3})

	// Initial state: all at 0
	combo.Reset()
	values := combo.GetAllIncrementerValues()
	expected := map[string]int{"ones": 0, "twos": 0, "threes": 0}
	for k, v := range expected {
		if values[k] != v {
			t.Errorf("GetAllIncrementerValues()[%q] = %v, want %v", k, values[k], v)
		}
	}

	// Increment some incrementers
	inc1.Increment() // ones: 1
	inc2.Increment() // twos: 1
	values = combo.GetAllIncrementerValues()
	expected = map[string]int{"ones": 1, "twos": 1, "threes": 0}
	for k, v := range expected {
		if values[k] != v {
			t.Errorf("After increment, GetAllIncrementerValues()[%q] = %v, want %v", k, values[k], v)
		}
	}
} 