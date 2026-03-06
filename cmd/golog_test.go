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
