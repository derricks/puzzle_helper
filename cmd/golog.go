package cmd

import (
	"fmt"
	"strconv"
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
		return nil, fmt.Errorf("expected '[', integer, or string after '=', got %q", p.current.val)
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
