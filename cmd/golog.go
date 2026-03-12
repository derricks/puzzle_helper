package cmd

import (
	"fmt"
	"strconv"
	"sync"
	"unicode"
)

// --- Tokens ---

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokIdent
	tokEquals
	tokLBracket
	tokRBracket
	tokInt
	tokComma
	tokDotDot
	tokString
	tokLParen
	tokRParen
)

type token struct {
	kind tokenKind
	val  string
}

// --- Lexer ---

type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input)}
}

func (l *lexer) skipSpace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *lexer) next() token {
	l.skipSpace()
	if l.pos >= len(l.input) {
		return token{kind: tokEOF}
	}
	ch := l.input[l.pos]
	switch {
	case ch == '=':
		l.pos++
		return token{kind: tokEquals, val: "="}
	case ch == '[':
		l.pos++
		return token{kind: tokLBracket, val: "["}
	case ch == ']':
		l.pos++
		return token{kind: tokRBracket, val: "]"}
	case ch == '(':
		l.pos++
		return token{kind: tokLParen, val: "("}
	case ch == ')':
		l.pos++
		return token{kind: tokRParen, val: ")"}
	case ch == ',':
		l.pos++
		return token{kind: tokComma, val: ","}
	case ch == '.' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '.':
		l.pos += 2
		return token{kind: tokDotDot, val: ".."}
	case unicode.IsLetter(ch) || ch == '_':
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
			l.pos++
		}
		return token{kind: tokIdent, val: string(l.input[start:l.pos])}
	case ch == '"':
		l.pos++ // skip opening quote
		start := l.pos
		for l.pos < len(l.input) && l.input[l.pos] != '"' {
			l.pos++
		}
		val := string(l.input[start:l.pos])
		if l.pos < len(l.input) {
			l.pos++ // skip closing quote
		}
		return token{kind: tokString, val: val}
	case unicode.IsDigit(ch) || (ch == '-' && l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1])):
		start := l.pos
		if ch == '-' {
			l.pos++
		}
		for l.pos < len(l.input) && unicode.IsDigit(l.input[l.pos]) {
			l.pos++
		}
		return token{kind: tokInt, val: string(l.input[start:l.pos])}
	default:
		l.pos++
		return l.next()
	}
}

// --- AST ---

// AssignNode is the root AST node for an assignment statement.
// Exactly one of List or Scalar is non-nil.
type AssignNode struct {
	Name   string
	List   *ListNode
	Scalar *ScalarNode
}

// ScalarNode represents a single fixed value. Exactly one of Int or String is non-nil.
type ScalarNode struct {
	Int    *int
	String *string
}

// ListNode represents a bracketed list. Either Elements or Range is set, not both.
type ListNode struct {
	Elements []*IntLiteralNode
	Range    *RangeNode
}

// RangeNode represents an inclusive integer range: [From..To].
type RangeNode struct {
	From *IntLiteralNode
	To   *IntLiteralNode
}

// IntLiteralNode represents a single integer literal in the list.
type IntLiteralNode struct {
	Value int
}

// --- Parser ---

type parser struct {
	lex     *lexer
	current token
}

func newParser(input string) *parser {
	p := &parser{lex: newLexer(input)}
	p.advance()
	return p
}

func (p *parser) advance() {
	p.current = p.lex.next()
}

// parseInt converts a string to an int, wrapping any error with the offending value.
func parseInt(s string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	return v, nil
}

// parseIntLiteral consumes a tokInt from the stream and returns it as an IntLiteralNode.
func (p *parser) parseIntLiteral() (*IntLiteralNode, error) {
	tok, err := p.expect(tokInt)
	if err != nil {
		return nil, err
	}
	v, err := parseInt(tok.val)
	if err != nil {
		return nil, err
	}
	return &IntLiteralNode{Value: v}, nil
}

func (p *parser) expect(kind tokenKind) (token, error) {
	tok := p.current
	if tok.kind != kind {
		return token{}, fmt.Errorf("expected token kind %d, got %d (%q)", kind, tok.kind, tok.val)
	}
	p.advance()
	return tok, nil
}

func (p *parser) parseAssignment() (*AssignNode, error) {
	nameTok, err := p.expect(tokIdent)
	if err != nil {
		return nil, fmt.Errorf("expected variable name: %w", err)
	}
	if _, err := p.expect(tokEquals); err != nil {
		return nil, fmt.Errorf("expected '=': %w", err)
	}
	node := &AssignNode{Name: nameTok.val}
	switch p.current.kind {
	case tokLBracket:
		list, err := p.parseList()
		if err != nil {
			return nil, err
		}
		node.List = list
	case tokIdent:
		funcName := p.current.val
		p.advance()
		list, err := p.parseFuncCall(funcName)
		if err != nil {
			return nil, err
		}
		node.List = list
	case tokInt:
		lit, err := p.parseIntLiteral()
		if err != nil {
			return nil, err
		}
		node.Scalar = &ScalarNode{Int: &lit.Value}
	case tokString:
		s := p.current.val
		p.advance()
		node.Scalar = &ScalarNode{String: &s}
	default:
		return nil, fmt.Errorf("expected '[', function call, integer, or string after '=', got %q", p.current.val)
	}
	return node, nil
}

func (p *parser) parseList() (*ListNode, error) {
	if _, err := p.expect(tokLBracket); err != nil {
		return nil, fmt.Errorf("expected '[': %w", err)
	}
	node := &ListNode{}

	// Empty list.
	if p.current.kind == tokRBracket {
		p.advance()
		return node, nil
	}

	// Parse the first integer, which is present in both list and range forms.
	firstLit, err := p.parseIntLiteral()
	if err != nil {
		return nil, fmt.Errorf("expected integer in list: %w", err)
	}

	// Range form: [from..to]
	if p.current.kind == tokDotDot {
		p.advance()
		toLit, err := p.parseIntLiteral()
		if err != nil {
			return nil, fmt.Errorf("expected integer after '..': %w", err)
		}
		if _, err := p.expect(tokRBracket); err != nil {
			return nil, fmt.Errorf("expected ']' after range: %w", err)
		}
		node.Range = &RangeNode{From: firstLit, To: toLit}
		return node, nil
	}

	// List form: first element already parsed, continue with the rest.
	node.Elements = append(node.Elements, firstLit)
	if p.current.kind == tokComma {
		p.advance()
	}
	for p.current.kind != tokRBracket && p.current.kind != tokEOF {
		lit, err := p.parseIntLiteral()
		if err != nil {
			return nil, fmt.Errorf("expected integer in list: %w", err)
		}
		node.Elements = append(node.Elements, lit)
		if p.current.kind == tokComma {
			p.advance()
		}
	}
	if _, err := p.expect(tokRBracket); err != nil {
		return nil, fmt.Errorf("expected ']': %w", err)
	}
	return node, nil
}

// funcRegistry maps DSL function names to their range-value producers.
// To add a new two-argument function to the DSL, register it here.
var funcRegistry = map[string]func(from, to int) []int{
	"prime":  primesInRange,
	"square": squaresInRange,
}

// parseFuncCall looks up name in funcRegistry and delegates to parseTwoArgRangeCall.
func (p *parser) parseFuncCall(name string) (*ListNode, error) {
	rangeFn, ok := funcRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown function %q", name)
	}
	return p.parseTwoArgRangeCall(name, rangeFn)
}

// parseTwoArgRangeCall parses (from, to) for any two-argument range function and
// returns a ListNode pre-populated with the values produced by rangeFn.
func (p *parser) parseTwoArgRangeCall(funcName string, rangeFn func(from, to int) []int) (*ListNode, error) {
	if _, err := p.expect(tokLParen); err != nil {
		return nil, fmt.Errorf("expected '(' after %q: %w", funcName, err)
	}
	fromLit, err := p.parseIntLiteral()
	if err != nil {
		return nil, fmt.Errorf("expected 'from' argument in %s(): %w", funcName, err)
	}
	if _, err := p.expect(tokComma); err != nil {
		return nil, fmt.Errorf("expected ',' between %s() arguments: %w", funcName, err)
	}
	toLit, err := p.parseIntLiteral()
	if err != nil {
		return nil, fmt.Errorf("expected 'to' argument in %s(): %w", funcName, err)
	}
	if _, err := p.expect(tokRParen); err != nil {
		return nil, fmt.Errorf("expected ')' after %s() arguments: %w", funcName, err)
	}
	vals := rangeFn(fromLit.Value, toLit.Value)
	elements := make([]*IntLiteralNode, len(vals))
	for i, v := range vals {
		elements[i] = &IntLiteralNode{Value: v}
	}
	return &ListNode{Elements: elements}, nil
}

// sieveMu guards sieveCache and sieveCacheMax.
var (
	sieveMu       sync.Mutex
	sieveCache    []bool
	sieveCacheMax int
)

// primesInRange returns all prime numbers in [from, to] inclusive using the
// Sieve of Eratosthenes. The sieve is cached at the package level and only
// recomputed when to exceeds the previously cached upper bound.
func primesInRange(from, to int) []int {
	if to < 2 {
		return nil
	}
	sieveMu.Lock()
	if to > sieveCacheMax {
		sieve := make([]bool, to+1)
		for i := range sieve {
			sieve[i] = true
		}
		sieve[0] = false
		sieve[1] = false
		for i := 2; i*i <= to; i++ {
			if sieve[i] {
				for j := i * i; j <= to; j += i {
					sieve[j] = false
				}
			}
		}
		sieveCache = sieve
		sieveCacheMax = to
	}
	sieve := sieveCache
	sieveMu.Unlock()

	lo := from
	if lo < 2 {
		lo = 2
	}
	var primes []int
	for i := lo; i <= to; i++ {
		if sieve[i] {
			primes = append(primes, i)
		}
	}
	return primes
}

// squareMu guards squareCache, squareCacheMax, and squareCacheRoot.
var (
	squareMu        sync.Mutex
	squareCache     []int
	squareCacheMax  = -1
	squareCacheRoot = -1
)

// squaresInRange returns all perfect squares in [from, to] inclusive.
// The list of squares is cached and extended incrementally — each call only
// computes squares beyond the previously cached upper bound.
func squaresInRange(from, to int) []int {
	squareMu.Lock()
	if to > squareCacheMax {
		for i := squareCacheRoot + 1; i*i <= to; i++ {
			squareCache = append(squareCache, i*i)
			squareCacheRoot = i
		}
		squareCacheMax = to
	}
	cache := squareCache
	squareMu.Unlock()

	var squares []int
	for _, sq := range cache {
		if sq > to {
			break
		}
		if sq >= from {
			squares = append(squares, sq)
		}
	}
	return squares
}

// Parse parses a DSL assignment statement and returns its AST.
func Parse(input string) (*AssignNode, error) {
	return newParser(input).parseAssignment()
}

// ToIncrementer converts a parsed AssignNode into a named SliceIncrementer[int].
// Handles int lists, int ranges, and int scalars.
func ToIncrementer(node *AssignNode) *SliceIncrementer[int] {
	var values []int
	if node.Scalar != nil {
		values = []int{*node.Scalar.Int}
	} else if node.List.Range != nil {
		from := node.List.Range.From.Value
		to := node.List.Range.To.Value
		for i := from; i <= to; i++ {
			values = append(values, i)
		}
	} else {
		values = make([]int, len(node.List.Elements))
		for i, el := range node.List.Elements {
			values[i] = el.Value
		}
	}
	return &SliceIncrementer[int]{name: node.Name, values: values}
}

// ToStringIncrementer converts a parsed AssignNode into a named SliceIncrementer[string].
// Handles string scalars.
func ToStringIncrementer(node *AssignNode) *SliceIncrementer[string] {
	return &SliceIncrementer[string]{name: node.Name, values: []string{*node.Scalar.String}}
}
