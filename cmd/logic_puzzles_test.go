package cmd

import (
	"reflect"
	"testing"
)

// --- Constraint.Check ---

func TestConstraintLT(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpLT, Right: "B"}
	if got, err := c.Check(map[string]int{"A": 1, "B": 2}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 1 < 2 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 2, "B": 1}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 2 < 1 to be false")
	}
	if got, err := c.Check(map[string]int{"A": 2, "B": 2}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 2 < 2 to be false")
	}
}

func TestConstraintGT(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpGT, Right: "B"}
	if got, err := c.Check(map[string]int{"A": 3, "B": 1}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 3 > 1 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 1, "B": 3}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 1 > 3 to be false")
	}
}

func TestConstraintLTE(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpLTE, Right: "B"}
	if got, err := c.Check(map[string]int{"A": 2, "B": 2}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 2 <= 2 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 1, "B": 2}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 1 <= 2 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 3, "B": 2}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 3 <= 2 to be false")
	}
}

func TestConstraintGTE(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpGTE, Right: "B"}
	if got, err := c.Check(map[string]int{"A": 2, "B": 2}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 2 >= 2 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 3, "B": 2}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 3 >= 2 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 1, "B": 2}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 1 >= 2 to be false")
	}
}

func TestConstraintEQ(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpEQ, Right: "B"}
	if got, err := c.Check(map[string]int{"A": 5, "B": 5}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 5 == 5 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 5, "B": 6}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 5 == 6 to be false")
	}
}

func TestConstraintNEQ(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpNEQ, Right: "B"}
	if got, err := c.Check(map[string]int{"A": 5, "B": 6}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected 5 != 6 to be true")
	}
	if got, err := c.Check(map[string]int{"A": 5, "B": 5}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected 5 != 5 to be false")
	}
}

func TestConstraintMissingNameReturnsError(t *testing.T) {
	c := Constraint[int]{Left: "A", Op: OpLT, Right: "B"}
	_, err := c.Check(map[string]int{"B": 1})
	if err == nil {
		t.Fatal("expected an error for missing variable \"A\", got nil")
	}
}

func TestConstraintString(t *testing.T) {
	// Constraints work on string incrementers via lexicographic ordering.
	c := Constraint[string]{Left: "First", Op: OpLT, Right: "Second"}
	if got, err := c.Check(map[string]string{"First": "apple", "Second": "banana"}); err != nil {
		t.Fatal(err)
	} else if !got {
		t.Error("expected \"apple\" < \"banana\" to be true")
	}
	if got, err := c.Check(map[string]string{"First": "zebra", "Second": "apple"}); err != nil {
		t.Fatal(err)
	} else if got {
		t.Error("expected \"zebra\" < \"apple\" to be false")
	}
}

// --- CheckAll ---

func TestCheckAllEmptySlice(t *testing.T) {
	got, err := CheckAll[int](nil, map[string]int{"A": 1})
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected CheckAll with no constraints to return true")
	}
}

func TestCheckAllAllPass(t *testing.T) {
	constraints := []Constraint[int]{
		{Left: "A", Op: OpLT, Right: "B"},
		{Left: "A", Op: OpGT, Right: "C"},
	}
	// C < A < B: C=1, A=3, B=5
	got, err := CheckAll(constraints, map[string]int{"A": 3, "B": 5, "C": 1})
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected all constraints to pass")
	}
}

func TestCheckAllFirstFails(t *testing.T) {
	constraints := []Constraint[int]{
		{Left: "A", Op: OpLT, Right: "B"}, // fails: 5 < 3 is false
		{Left: "A", Op: OpGT, Right: "C"}, // would pass
	}
	got, err := CheckAll(constraints, map[string]int{"A": 5, "B": 3, "C": 1})
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected CheckAll to return false when first constraint fails")
	}
}

func TestCheckAllSecondFails(t *testing.T) {
	constraints := []Constraint[int]{
		{Left: "A", Op: OpLT, Right: "B"}, // passes: 3 < 5
		{Left: "A", Op: OpGT, Right: "C"}, // fails: 3 > 9 is false
	}
	got, err := CheckAll(constraints, map[string]int{"A": 3, "B": 5, "C": 9})
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected CheckAll to return false when second constraint fails")
	}
}

// --- OdometerVars ---

func TestOdometerVars(t *testing.T) {
	a := &SliceIncrementer[int]{name: "A", values: []int{1, 2, 3}}
	b := &SliceIncrementer[int]{name: "B", values: []int{4, 5, 6}}
	o := &Odometer[int]{
		name:   "o",
		names:  []string{"A", "B"},
		digits: []Incrementer[int]{a, b},
	}

	got := OdometerVars(o)
	want := map[string]int{"A": 1, "B": 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestOdometerVarsReusesMap(t *testing.T) {
	a := &SliceIncrementer[int]{name: "A", values: []int{1, 2}}
	b := &SliceIncrementer[int]{name: "B", values: []int{10, 20}}
	o := &Odometer[int]{
		name:   "o",
		names:  []string{"A", "B"},
		digits: []Incrementer[int]{a, b},
	}

	first := OdometerVars(o)
	o.Increment()
	second := OdometerVars(o)

	if reflect.ValueOf(first).Pointer() != reflect.ValueOf(second).Pointer() {
		t.Error("OdometerVars allocated a new map on the second call; expected the same map to be reused")
	}
}

func TestOdometerVarsAfterIncrement(t *testing.T) {
	a := &SliceIncrementer[int]{name: "A", values: []int{1, 2}}
	b := &SliceIncrementer[int]{name: "B", values: []int{10, 20}}
	o := &Odometer[int]{
		name:   "o",
		names:  []string{"A", "B"},
		digits: []Incrementer[int]{a, b},
	}
	o.Increment()

	got := OdometerVars(o)
	want := map[string]int{"A": 1, "B": 20}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

// --- Integration: Odometer + constraints ---

func TestOdometerWithConstraintFiltersResults(t *testing.T) {
	// A in [1,2,3], B in [1,2,3]; collect pairs where A < B.
	a := &SliceIncrementer[int]{name: "A", values: []int{1, 2, 3}}
	b := &SliceIncrementer[int]{name: "B", values: []int{1, 2, 3}}
	o := &Odometer[int]{
		name:   "o",
		names:  []string{"A", "B"},
		digits: []Incrementer[int]{a, b},
	}
	constraints := []Constraint[int]{{Left: "A", Op: OpLT, Right: "B"}}

	var solutions [][]int
	for !o.IsMaxed() {
		vars := OdometerVars(o)
		ok, err := CheckAll(constraints, vars)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			cv := o.CurrentValue()
			solutions = append(solutions, []int{cv[0], cv[1]})
		}
		o.Increment()
	}
	// Check the final state too.
	vars := OdometerVars(o)
	ok, err := CheckAll(constraints, vars)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		cv := o.CurrentValue()
		solutions = append(solutions, []int{cv[0], cv[1]})
	}

	want := [][]int{{1, 2}, {1, 3}, {2, 3}}
	if !reflect.DeepEqual(solutions, want) {
		t.Errorf("expected %v, got %v", want, solutions)
	}
}

// counter is a simple test-only Incrementer[int] with a configurable max.
type counter struct {
	value   int
	maxVal  int
}

func newCounter(max int) *counter       { return &counter{maxVal: max} }
func (c *counter) Increment()           { c.value++ }
func (c *counter) IsMaxed() bool        { return c.value >= c.maxVal }
func (c *counter) Reset()               { c.value = 0 }
func (c *counter) CurrentValue() int    { return c.value }
func (c *counter) Name() string         { return "" }

// --- SliceIncrementer ---

func TestSliceIncrementerCurrentValueInitial(t *testing.T) {
	s := &SliceIncrementer[int]{values: []int{10, 20, 30}}
	if got := s.CurrentValue(); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
}

func TestSliceIncrementerIncrement(t *testing.T) {
	s := &SliceIncrementer[int]{values: []int{10, 20, 30}}
	s.Increment()
	if got := s.CurrentValue(); got != 20 {
		t.Errorf("expected 20 after one Increment, got %d", got)
	}
}

func TestSliceIncrementerIsMaxedFalse(t *testing.T) {
	s := &SliceIncrementer[int]{values: []int{10, 20, 30}}
	if s.IsMaxed() {
		t.Error("expected IsMaxed() = false at first element")
	}
}

func TestSliceIncrementerIsMaxedTrueAtLastElement(t *testing.T) {
	s := &SliceIncrementer[int]{values: []int{10, 20, 30}}
	s.Increment()
	s.Increment() // now at index 2, the last element
	if !s.IsMaxed() {
		t.Errorf("expected IsMaxed() = true at last element, CurrentValue=%v", s.CurrentValue())
	}
}

func TestSliceIncrementerReset(t *testing.T) {
	s := &SliceIncrementer[int]{values: []int{10, 20, 30}}
	s.Increment()
	s.Increment()
	s.Reset()
	if got := s.CurrentValue(); got != 10 {
		t.Errorf("expected 10 after Reset, got %d", got)
	}
	if s.IsMaxed() {
		t.Error("expected IsMaxed() = false after Reset")
	}
}

func TestSliceIncrementerFullCycle(t *testing.T) {
	s := &SliceIncrementer[int]{values: []int{7, 14, 21}}
	expected := []int{7, 14, 21}
	for _, want := range expected {
		if got := s.CurrentValue(); got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
		s.Increment()
	}
	if !s.IsMaxed() {
		t.Error("expected IsMaxed() = true after stepping past last element")
	}
}

func TestSliceIncrementerInOdometer(t *testing.T) {
	// Odometer with two SliceIncrementers: first cycles [A,B], second cycles [1,2,3].
	// Full sequence: [A,1] [A,2] [A,3] [B,1] [B,2] [B,3] then maxed.
	o := &Odometer[int]{digits: []Incrementer[int]{
		&SliceIncrementer[int]{values: []int{0, 1}},
		&SliceIncrementer[int]{values: []int{1, 2, 3}},
	}}
	expected := [][]int{
		{0, 1}, {0, 2}, {0, 3},
		{1, 1}, {1, 2}, {1, 3},
	}
	for _, want := range expected {
		if got := o.CurrentValue(); !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
		o.Increment()
	}
	if !o.IsMaxed() {
		t.Errorf("expected odometer to be maxed, got %v", o.CurrentValue())
	}
}

func TestSliceIncrementerName(t *testing.T) {
	s := &SliceIncrementer[int]{name: "digit", values: []int{1, 2, 3}}
	if got := s.Name(); got != "digit" {
		t.Errorf("expected name \"digit\", got %q", got)
	}
}

func TestOdometerName(t *testing.T) {
	o := &Odometer[int]{name: "myOdometer", names: []string{"tens", "ones"}, digits: []Incrementer[int]{
		newCounter(9), newCounter(9),
	}}
	if got := o.Name(); got != "myOdometer" {
		t.Errorf("expected name \"myOdometer\", got %q", got)
	}
}

func TestOdometerTracksBothNamesAndDigits(t *testing.T) {
	tens := &SliceIncrementer[int]{name: "tens", values: []int{0, 1, 2}}
	ones := &SliceIncrementer[int]{name: "ones", values: []int{0, 1, 2}}
	o := &Odometer[int]{
		name:   "counter",
		names:  []string{tens.Name(), ones.Name()},
		digits: []Incrementer[int]{tens, ones},
	}

	if len(o.names) != len(o.digits) {
		t.Fatalf("expected names and digits slices to have equal length")
	}
	if o.names[0] != "tens" {
		t.Errorf("expected names[0] = \"tens\", got %q", o.names[0])
	}
	if o.names[1] != "ones" {
		t.Errorf("expected names[1] = \"ones\", got %q", o.names[1])
	}

	o.Increment()
	o.Increment()
	// After two increments: ones digit is at index 2 (value=2)
	if got := o.digits[1].CurrentValue(); got != 2 {
		t.Errorf("expected ones digit = 2, got %d", got)
	}
}

func TestSliceIncrementerWithStrings(t *testing.T) {
	s := &SliceIncrementer[string]{values: []string{"low", "mid", "high"}}
	if got := s.CurrentValue(); got != "low" {
		t.Errorf("expected \"low\", got %q", got)
	}
	s.Increment()
	if got := s.CurrentValue(); got != "mid" {
		t.Errorf("expected \"mid\", got %q", got)
	}
	if s.IsMaxed() {
		t.Error("expected IsMaxed() = false at middle element")
	}
	s.Increment()
	if got := s.CurrentValue(); got != "high" {
		t.Errorf("expected \"high\", got %q", got)
	}
	if !s.IsMaxed() {
		t.Error("expected IsMaxed() = true at last element")
	}
}

// --- Odometer.CurrentValue ---

func TestOdometerCurrentValueInitial(t *testing.T) {
	o := &Odometer[int]{digits: []Incrementer[int]{
		newCounter(9), newCounter(9), newCounter(9),
	}}
	if got := o.CurrentValue(); !reflect.DeepEqual(got, []int{0, 0, 0}) {
		t.Errorf("expected [0 0 0], got %v", got)
	}
}

// --- Odometer.Increment basic ---

func TestOdometerIncrementRightmostDigit(t *testing.T) {
	o := &Odometer[int]{digits: []Incrementer[int]{
		newCounter(9), newCounter(9), newCounter(9),
	}}
	o.Increment()
	if got := o.CurrentValue(); !reflect.DeepEqual(got, []int{0, 0, 1}) {
		t.Errorf("expected [0 0 1], got %v", got)
	}
}

func TestOdometerIncrementCarryOnce(t *testing.T) {
	// Rightmost must already be maxed for a carry to occur.
	rightmost := newCounter(1)
	rightmost.Increment() // value=1, IsMaxed()=true
	o := &Odometer[int]{digits: []Incrementer[int]{
		newCounter(9), newCounter(9), rightmost,
	}}
	o.Increment() // rightmost maxed: reset, carry → middle increments
	if got := o.CurrentValue(); !reflect.DeepEqual(got, []int{0, 1, 0}) {
		t.Errorf("expected [0 1 0], got %v", got)
	}
}

func TestOdometerIncrementCarryChain(t *testing.T) {
	// Two rightmost digits pre-maxed; carry cascades to leftmost.
	right := newCounter(1)
	right.Increment() // maxed
	mid := newCounter(1)
	mid.Increment() // maxed
	o := &Odometer[int]{digits: []Incrementer[int]{
		newCounter(9), mid, right,
	}}
	o.Increment()
	if got := o.CurrentValue(); !reflect.DeepEqual(got, []int{1, 0, 0}) {
		t.Errorf("expected [1 0 0], got %v", got)
	}
}

// --- Odometer.IsMaxed ---

func TestOdometerIsMaxedFalseInitially(t *testing.T) {
	o := &Odometer[int]{digits: []Incrementer[int]{newCounter(3)}}
	if o.IsMaxed() {
		t.Error("expected IsMaxed() = false on fresh odometer")
	}
}

func TestOdometerIsMaxedTrueWhenAllDigitsMaxed(t *testing.T) {
	d := newCounter(1)
	d.Increment() // now at max
	o := &Odometer[int]{digits: []Incrementer[int]{d}}
	if !o.IsMaxed() {
		t.Error("expected IsMaxed() = true when sole digit is maxed")
	}
}

func TestOdometerIsMaxedRequiresAllDigitsMaxed(t *testing.T) {
	maxed := newCounter(1)
	maxed.Increment()
	notMaxed := newCounter(9)

	o := &Odometer[int]{digits: []Incrementer[int]{maxed, notMaxed}}
	if o.IsMaxed() {
		t.Error("expected IsMaxed() = false when only one of two digits is maxed")
	}
}

// --- Odometer.Increment is a no-op when fully maxed ---

func TestOdometerIncrementNoOpWhenMaxed(t *testing.T) {
	d := newCounter(1)
	d.Increment() // maxed
	o := &Odometer[int]{digits: []Incrementer[int]{d}}

	o.Increment() // must not reset d back to 0
	if got := o.CurrentValue(); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("expected [1] after no-op increment, got %v", got)
	}
	if !o.IsMaxed() {
		t.Error("expected odometer to remain maxed")
	}
}

// --- Odometer.Reset ---

func TestOdometerReset(t *testing.T) {
	o := &Odometer[int]{digits: []Incrementer[int]{
		newCounter(9), newCounter(9), newCounter(9),
	}}
	o.Increment()
	o.Increment()
	o.Reset()
	if got := o.CurrentValue(); !reflect.DeepEqual(got, []int{0, 0, 0}) {
		t.Errorf("expected [0 0 0] after reset, got %v", got)
	}
}

// --- Full odometer cycle ---

func TestOdometerFullCycle(t *testing.T) {
	// 2-digit base-3 odometer: counts 00→01→02→10→11→12→20→21→22 then maxed.
	// counter(max=2): IsMaxed when value>=2, so valid values are 0,1,2.
	o := &Odometer[int]{digits: []Incrementer[int]{
		newCounter(2), newCounter(2),
	}}

	expected := [][]int{
		{0, 0}, {0, 1}, {0, 2},
		{1, 0}, {1, 1}, {1, 2},
		{2, 0}, {2, 1}, {2, 2},
	}

	for _, want := range expected {
		if got := o.CurrentValue(); !reflect.DeepEqual(got, want) {
			t.Errorf("expected %v, got %v", want, got)
		}
		o.Increment()
	}

	if !o.IsMaxed() {
		t.Errorf("expected odometer to be maxed after full cycle, got %v", o.CurrentValue())
	}
}
