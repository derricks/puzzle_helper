package cmd

import (
	"reflect"
	"testing"
)

// --- Lexer ---

func TestLexerTokenizesAssignment(t *testing.T) {
	l := newLexer("N = [1,2,3,4]")
	expected := []token{
		{kind: tokIdent, val: "N"},
		{kind: tokEquals, val: "="},
		{kind: tokLBracket, val: "["},
		{kind: tokInt, val: "1"},
		{kind: tokComma, val: ","},
		{kind: tokInt, val: "2"},
		{kind: tokComma, val: ","},
		{kind: tokInt, val: "3"},
		{kind: tokComma, val: ","},
		{kind: tokInt, val: "4"},
		{kind: tokRBracket, val: "]"},
		{kind: tokEOF},
	}
	for _, want := range expected {
		got := l.next()
		if got.kind != want.kind || got.val != want.val {
			t.Errorf("expected token %+v, got %+v", want, got)
		}
	}
}

func TestLexerHandlesWhitespace(t *testing.T) {
	l := newLexer("  MyVar  =  [ 1 , 2 ]  ")
	wantKinds := []tokenKind{tokIdent, tokEquals, tokLBracket, tokInt, tokComma, tokInt, tokRBracket, tokEOF}
	for _, want := range wantKinds {
		got := l.next()
		if got.kind != want {
			t.Errorf("expected kind %d, got %d (%q)", want, got.kind, got.val)
		}
	}
}

func TestLexerHandlesNegativeIntegers(t *testing.T) {
	l := newLexer("X = [-1,-2]")
	expected := []token{
		{kind: tokIdent, val: "X"},
		{kind: tokEquals, val: "="},
		{kind: tokLBracket, val: "["},
		{kind: tokInt, val: "-1"},
		{kind: tokComma, val: ","},
		{kind: tokInt, val: "-2"},
		{kind: tokRBracket, val: "]"},
		{kind: tokEOF},
	}
	for _, want := range expected {
		got := l.next()
		if got.kind != want.kind || got.val != want.val {
			t.Errorf("expected token %+v, got %+v", want, got)
		}
	}
}

func TestLexerTokenizesStringLiteral(t *testing.T) {
	l := newLexer(`Name = "hello"`)
	expected := []token{
		{kind: tokIdent, val: "Name"},
		{kind: tokEquals, val: "="},
		{kind: tokString, val: "hello"},
		{kind: tokEOF},
	}
	for _, want := range expected {
		got := l.next()
		if got.kind != want.kind || got.val != want.val {
			t.Errorf("expected token %+v, got %+v", want, got)
		}
	}
}

func TestLexerTokenizesDotDot(t *testing.T) {
	l := newLexer("N = [1..9]")
	expected := []token{
		{kind: tokIdent, val: "N"},
		{kind: tokEquals, val: "="},
		{kind: tokLBracket, val: "["},
		{kind: tokInt, val: "1"},
		{kind: tokDotDot, val: ".."},
		{kind: tokInt, val: "9"},
		{kind: tokRBracket, val: "]"},
		{kind: tokEOF},
	}
	for _, want := range expected {
		got := l.next()
		if got.kind != want.kind || got.val != want.val {
			t.Errorf("expected token %+v, got %+v", want, got)
		}
	}
}

// --- Parser / AST ---

func TestParseAssignment(t *testing.T) {
	node, err := Parse("N = [1,2,3,4]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.Name != "N" {
		t.Errorf("expected Name=\"N\", got %q", node.Name)
	}
	if len(node.List.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(node.List.Elements))
	}
	for i, want := range []int{1, 2, 3, 4} {
		if got := node.List.Elements[i].Value; got != want {
			t.Errorf("element[%d]: expected %d, got %d", i, want, got)
		}
	}
}

func TestParseMultiCharName(t *testing.T) {
	node, err := Parse("Day = [1,2,3]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.Name != "Day" {
		t.Errorf("expected Name=\"Day\", got %q", node.Name)
	}
}

func TestParseEmptyList(t *testing.T) {
	node, err := Parse("X = []")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(node.List.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(node.List.Elements))
	}
}

func TestParseNegativeIntegers(t *testing.T) {
	node, err := Parse("Temp = [-10,-5,0,5,10]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	wantVals := []int{-10, -5, 0, 5, 10}
	for i, want := range wantVals {
		if got := node.List.Elements[i].Value; got != want {
			t.Errorf("element[%d]: expected %d, got %d", i, want, got)
		}
	}
}

func TestParseIntScalar(t *testing.T) {
	node, err := Parse("N = 5")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.List != nil {
		t.Errorf("expected List to be nil for a scalar, got %+v", node.List)
	}
	if node.Scalar == nil || node.Scalar.Int == nil {
		t.Fatal("expected Scalar.Int to be set")
	}
	if got := *node.Scalar.Int; got != 5 {
		t.Errorf("expected Scalar.Int=5, got %d", got)
	}
}

func TestParseStringScalar(t *testing.T) {
	node, err := Parse(`Color = "red"`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.List != nil {
		t.Errorf("expected List to be nil for a scalar, got %+v", node.List)
	}
	if node.Scalar == nil || node.Scalar.String == nil {
		t.Fatal("expected Scalar.String to be set")
	}
	if got := *node.Scalar.String; got != "red" {
		t.Errorf("expected Scalar.String=\"red\", got %q", got)
	}
}

func TestParseRange(t *testing.T) {
	node, err := Parse("N = [1..9]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.Name != "N" {
		t.Errorf("expected Name=\"N\", got %q", node.Name)
	}
	if node.List.Range == nil {
		t.Fatal("expected Range to be set, got nil")
	}
	if node.List.Elements != nil {
		t.Errorf("expected Elements to be nil for a range, got %v", node.List.Elements)
	}
	if got := node.List.Range.From.Value; got != 1 {
		t.Errorf("expected Range.From=1, got %d", got)
	}
	if got := node.List.Range.To.Value; got != 9 {
		t.Errorf("expected Range.To=9, got %d", got)
	}
}

func TestParseRangeErrorMissingTo(t *testing.T) {
	if _, err := Parse("N = [1..]"); err == nil {
		t.Error("expected error for missing 'to' value in range, got nil")
	}
}

func TestParseErrorMissingEquals(t *testing.T) {
	if _, err := Parse("N [1,2,3]"); err == nil {
		t.Error("expected error for missing '=', got nil")
	}
}

func TestParseErrorInvalidRHS(t *testing.T) {
	// A comma immediately after '=' is not a valid list, scalar int, or scalar string.
	if _, err := Parse("N = ,1,2,3"); err == nil {
		t.Error("expected error for invalid RHS starting with ',', got nil")
	}
}

func TestParseErrorMissingCloseBracket(t *testing.T) {
	if _, err := Parse("N = [1,2,3"); err == nil {
		t.Error("expected error for missing ']', got nil")
	}
}

// --- ToIncrementer ---

func TestToIncrementerName(t *testing.T) {
	node, err := Parse("N = [1,2,3,4]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if inc.Name() != "N" {
		t.Errorf("expected Name=\"N\", got %q", inc.Name())
	}
}

func TestToIncrementerInitialValue(t *testing.T) {
	node, err := Parse("N = [1,2,3,4]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if got := inc.CurrentValue(); got != 1 {
		t.Errorf("expected initial value 1, got %d", got)
	}
}

func TestToIncrementerFullCycle(t *testing.T) {
	node, err := Parse("N = [1,2,3,4]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	for _, want := range []int{1, 2, 3, 4} {
		if got := inc.CurrentValue(); got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
		inc.Increment()
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true after full cycle")
	}
}

func TestToIncrementerFromIntScalar(t *testing.T) {
	node, err := Parse("N = 42")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if inc.Name() != "N" {
		t.Errorf("expected Name=\"N\", got %q", inc.Name())
	}
	if got := inc.CurrentValue(); got != 42 {
		t.Errorf("expected CurrentValue=42, got %d", got)
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true for a scalar (single-value) incrementer")
	}
}

func TestToStringIncrementerFromStringScalar(t *testing.T) {
	node, err := Parse(`Color = "blue"`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToStringIncrementer(node)
	if inc.Name() != "Color" {
		t.Errorf("expected Name=\"Color\", got %q", inc.Name())
	}
	if got := inc.CurrentValue(); got != "blue" {
		t.Errorf("expected CurrentValue=\"blue\", got %q", got)
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true for a scalar (single-value) incrementer")
	}
}

func TestToIncrementerFromRange(t *testing.T) {
	node, err := Parse("N = [1..9]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if inc.Name() != "N" {
		t.Errorf("expected Name=\"N\", got %q", inc.Name())
	}
	for _, want := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} {
		if got := inc.CurrentValue(); got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
		inc.Increment()
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true after full range cycle")
	}
}

func TestToIncrementerSingleValueRange(t *testing.T) {
	node, err := Parse("X = [5..5]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if got := inc.CurrentValue(); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true for single-value range")
	}
}

// --- prime() DSL ---

func TestLexerTokenizesPrimeCall(t *testing.T) {
	l := newLexer("N = prime(10,999)")
	expected := []token{
		{kind: tokIdent, val: "N"},
		{kind: tokEquals, val: "="},
		{kind: tokIdent, val: "prime"},
		{kind: tokLParen, val: "("},
		{kind: tokInt, val: "10"},
		{kind: tokComma, val: ","},
		{kind: tokInt, val: "999"},
		{kind: tokRParen, val: ")"},
		{kind: tokEOF},
	}
	for _, want := range expected {
		got := l.next()
		if got.kind != want.kind || got.val != want.val {
			t.Errorf("expected token %+v, got %+v", want, got)
		}
	}
}

func TestParsePrimeCallFirstElement(t *testing.T) {
	node, err := Parse("N = prime(10, 999)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.List == nil || len(node.List.Elements) == 0 {
		t.Fatal("expected non-empty element list from prime()")
	}
	// First prime >= 10 is 11.
	if got := node.List.Elements[0].Value; got != 11 {
		t.Errorf("expected first element 11, got %d", got)
	}
}

func TestParsePrimeCallLastElement(t *testing.T) {
	node, err := Parse("N = prime(10, 999)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	last := node.List.Elements[len(node.List.Elements)-1]
	// Largest prime <= 999 is 997.
	if got := last.Value; got != 997 {
		t.Errorf("expected last element 997, got %d", got)
	}
}

func TestParsePrimeCallElementCount(t *testing.T) {
	node, err := Parse("N = prime(10, 999)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	// 168 primes up to 999; subtract {2,3,5,7} below 10 → 164.
	if got := len(node.List.Elements); got != 164 {
		t.Errorf("expected 164 primes in [10,999], got %d", got)
	}
}

func TestParsePrimeCallRangeHasNoPrimes(t *testing.T) {
	// 14..16 contains no primes (14=2×7, 15=3×5, 16=2^4).
	node, err := Parse("N = prime(14, 16)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if got := len(node.List.Elements); got != 0 {
		t.Errorf("expected 0 primes in [14,16], got %d", got)
	}
}

func TestParsePrimeErrorUnknownFunction(t *testing.T) {
	if _, err := Parse("N = fibonacci(1, 10)"); err == nil {
		t.Error("expected error for unknown function, got nil")
	}
}

func TestParsePrimeErrorMissingOpenParen(t *testing.T) {
	if _, err := Parse("N = prime 10, 999)"); err == nil {
		t.Error("expected error for missing '(', got nil")
	}
}

func TestParsePrimeErrorMissingComma(t *testing.T) {
	if _, err := Parse("N = prime(10 999)"); err == nil {
		t.Error("expected error for missing ',' between prime() arguments, got nil")
	}
}

func TestToIncrementerFromPrime(t *testing.T) {
	node, err := Parse("N = prime(10, 20)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if inc.Name() != "N" {
		t.Errorf("expected Name=\"N\", got %q", inc.Name())
	}
	// Primes in [10,20]: 11, 13, 17, 19.
	for _, want := range []int{11, 13, 17, 19} {
		if got := inc.CurrentValue(); got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
		inc.Increment()
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true after cycling all primes in [10,20]")
	}
}

func TestPrimeSieveCacheIsReused(t *testing.T) {
	// Reset cache to a clean slate so this test is independent of run order.
	sieveMu.Lock()
	sieveCache = nil
	sieveCacheMax = 0
	sieveMu.Unlock()

	primesInRange(10, 100)

	sieveMu.Lock()
	firstPtr := reflect.ValueOf(sieveCache).Pointer()
	sieveMu.Unlock()

	// Smaller upper bound: existing sieve covers it, no new allocation.
	primesInRange(10, 50)

	sieveMu.Lock()
	secondPtr := reflect.ValueOf(sieveCache).Pointer()
	sieveMu.Unlock()

	if firstPtr != secondPtr {
		t.Error("expected sieve cache to be reused for a smaller range, but a new allocation occurred")
	}

	// Larger upper bound: sieve must be rebuilt.
	primesInRange(10, 200)

	sieveMu.Lock()
	thirdPtr := reflect.ValueOf(sieveCache).Pointer()
	cachedMax := sieveCacheMax
	sieveMu.Unlock()

	if thirdPtr == firstPtr {
		t.Error("expected sieve cache to be replaced for a larger range, but old allocation was kept")
	}
	if cachedMax != 200 {
		t.Errorf("expected sieveCacheMax = 200 after expanding cache, got %d", cachedMax)
	}
}

func TestSquareCacheIsReused(t *testing.T) {
	// Reset cache to a clean slate so this test is independent of run order.
	squareMu.Lock()
	squareCache = nil
	squareCacheMax = -1
	squareCacheRoot = -1
	squareMu.Unlock()

	squaresInRange(0, 25)

	squareMu.Lock()
	lenAfterFirst := len(squareCache)
	maxAfterFirst := squareCacheMax
	squareMu.Unlock()

	// Smaller upper bound: cache should not change.
	squaresInRange(0, 16)

	squareMu.Lock()
	lenAfterSmaller := len(squareCache)
	maxAfterSmaller := squareCacheMax
	squareMu.Unlock()

	if lenAfterSmaller != lenAfterFirst {
		t.Errorf("expected cache length to stay %d for smaller range, got %d", lenAfterFirst, lenAfterSmaller)
	}
	if maxAfterSmaller != maxAfterFirst {
		t.Errorf("expected squareCacheMax to stay %d for smaller range, got %d", maxAfterFirst, maxAfterSmaller)
	}

	// Larger upper bound: cache should be extended incrementally.
	squaresInRange(0, 100)

	squareMu.Lock()
	lenAfterLarger := len(squareCache)
	maxAfterLarger := squareCacheMax
	squareMu.Unlock()

	if lenAfterLarger <= lenAfterFirst {
		t.Errorf("expected cache to grow beyond %d entries for larger range, got %d", lenAfterFirst, lenAfterLarger)
	}
	if maxAfterLarger != 100 {
		t.Errorf("expected squareCacheMax = 100 after extending cache, got %d", maxAfterLarger)
	}
}

// --- square() DSL ---

func TestParseSqaureCallFirstElement(t *testing.T) {
	node, err := Parse("N = square(10, 1000)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if node.List == nil || len(node.List.Elements) == 0 {
		t.Fatal("expected non-empty element list from square()")
	}
	// First perfect square >= 10 is 16 (4²).
	if got := node.List.Elements[0].Value; got != 16 {
		t.Errorf("expected first element 16, got %d", got)
	}
}

func TestParseSquareCallLastElement(t *testing.T) {
	node, err := Parse("N = square(10, 1000)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	last := node.List.Elements[len(node.List.Elements)-1]
	// Largest perfect square <= 1000 is 961 (31²).
	if got := last.Value; got != 961 {
		t.Errorf("expected last element 961, got %d", got)
	}
}

func TestParseSquareCallElementCount(t *testing.T) {
	node, err := Parse("N = square(10, 1000)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	// Squares in [10,1000]: 4²=16 through 31²=961 → 28 values.
	if got := len(node.List.Elements); got != 28 {
		t.Errorf("expected 28 squares in [10,1000], got %d", got)
	}
}

func TestParseSquareCallBoundaryIsInclusive(t *testing.T) {
	// 25 (5²) and 36 (6²) are both in [25, 36].
	node, err := Parse("N = square(25, 36)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	wantVals := []int{25, 36}
	if got := len(node.List.Elements); got != len(wantVals) {
		t.Fatalf("expected %d elements, got %d", len(wantVals), got)
	}
	for i, want := range wantVals {
		if got := node.List.Elements[i].Value; got != want {
			t.Errorf("element[%d]: expected %d, got %d", i, want, got)
		}
	}
}

func TestParseSquareCallRangeHasNoSquares(t *testing.T) {
	// 10..15 contains no perfect squares (9 and 16 are both outside).
	node, err := Parse("N = square(10, 15)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if got := len(node.List.Elements); got != 0 {
		t.Errorf("expected 0 squares in [10,15], got %d", got)
	}
}

func TestToIncrementerFromSquare(t *testing.T) {
	node, err := Parse("N = square(1, 25)")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	inc := ToIncrementer(node)
	if inc.Name() != "N" {
		t.Errorf("expected Name=\"N\", got %q", inc.Name())
	}
	// Perfect squares in [1,25]: 1, 4, 9, 16, 25.
	for _, want := range []int{1, 4, 9, 16, 25} {
		if got := inc.CurrentValue(); got != want {
			t.Errorf("expected %d, got %d", want, got)
		}
		inc.Increment()
	}
	if !inc.IsMaxed() {
		t.Error("expected IsMaxed() = true after cycling all squares in [1,25]")
	}
}

// --- Integration: Parse → Odometer ---

func TestParseIntoOdometer(t *testing.T) {
	tensNode, err := Parse("tens = [0,1,2]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	onesNode, err := Parse("ones = [0,1,2]")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	tens := ToIncrementer(tensNode)
	ones := ToIncrementer(onesNode)
	o := &Odometer[int]{
		name:   "counter",
		names:  []string{tens.Name(), ones.Name()},
		digits: []Incrementer[int]{tens, ones},
	}

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
		t.Errorf("expected odometer to be maxed, got %v", o.CurrentValue())
	}
}
