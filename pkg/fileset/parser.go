package fileset

import (
	"fmt"
	"unicode/utf8"
)

type ParseError struct {
	Message string
	Span    Span
	origin  *ParseError
}

func (e *ParseError) Error() string {
	return e.Message
}

func (e *ParseError) Origin() *ParseError {
	return e.origin
}

type parser struct {
	input       string
	pos         int
	lastContent int
}

// Parse parses a fileset expression string into an AST.
func Parse(input string) (*Expression, error) {
	p := &parser{input: input}
	p.skipWhitespace()
	start := p.pos
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return nil, p.syntaxError("unexpected token")
	}
	contentEnd := min(p.lastContent, len(p.input))
	if expr.Span.End > contentEnd || expr.Span.Start < start {
		expr.Span = Span{Start: start, End: contentEnd}
	}
	return expr, nil
}

// ParseProgramOrBareString parses input as either a fileset expression
// or a bare string / bare string pattern (fallback when no operators are present).
func ParseProgramOrBareString(input string) (*Expression, error) {
	p := &parser{input: input}
	p.skipWhitespace()
	start := p.pos

	// Try parsing as expression first
	saved := p.pos
	expr, err := p.parseExpression()
	if err == nil {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			contentEnd := min(p.lastContent, len(p.input))
			if expr.Span.End > contentEnd || expr.Span.Start < start {
				expr.Span = Span{Start: start, End: contentEnd}
			}
			return expr, nil
		}
	}

	// Expression didn't consume all input or failed; try bare_string_pattern
	p.pos = saved
	if barePattern := p.tryBareStringPattern(); barePattern != nil {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return barePattern, nil
		}
	}

	// Try bare_string
	p.pos = saved
	if bareStr := p.tryBareString(); bareStr != nil {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return bareStr, nil
		}
	}

	if err != nil {
		return nil, err
	}
	return nil, p.syntaxError("unexpected token")
}

// IsIdentifier checks if the text is a valid fileset identifier.
func IsIdentifier(text string) bool {
	p := &parser{input: text}
	start := p.pos
	if !p.scanIdentifier() {
		return false
	}
	return p.pos == len(text) && p.pos > start
}

func (p *parser) syntaxError(msg string) *ParseError {
	return &ParseError{
		Message: msg,
		Span:    Span{Start: p.pos, End: min(p.pos+1, len(p.input))},
	}
}

func (p *parser) syntaxErrorf(format string, args ...any) *ParseError {
	return &ParseError{
		Message: fmt.Sprintf(format, args...),
		Span:    Span{Start: p.pos, End: min(p.pos+1, len(p.input))},
	}
}

func (p *parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *parser) advance() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += w
	p.lastContent = p.pos
	return r
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) {
		r, w := utf8.DecodeRuneInString(p.input[p.pos:])
		if !isWhitespace(r) {
			break
		}
		p.pos += w
	}
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.input)
}



// Operator precedence (from lowest to highest):
// 0: union (|) - left assoc
// 1: intersection (&), difference (~) - left assoc
// 2: prefix negate (~)
// 3: primary (identifier, string, pattern, function, parens)

func (p *parser) parseExpression() (*Expression, error) {
	return p.parseInfixLevel0()
}

// Level 0: union (|)
func (p *parser) parseInfixLevel0() (*Expression, error) {
	left, err := p.parseInfixLevel1()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.atEnd() {
			break
		}
		ch := p.peek()
		if ch == '|' {
			p.advance()
			p.skipWhitespace()
			right, err := p.parseInfixLevel1()
			if err != nil {
				return nil, err
			}
			left = p.unionNodes(left, right)
		} else {
			break
		}
	}
	return left, nil
}

// Level 1: intersection (&), difference (~)
func (p *parser) parseInfixLevel1() (*Expression, error) {
	left, err := p.parseNegatePrefix()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.atEnd() {
			break
		}
		ch := p.peek()
		if ch == '&' {
			p.advance()
			p.skipWhitespace()
			right, err := p.parseNegatePrefix()
			if err != nil {
				return nil, err
			}
			leftStart := left.Span.Start
			left = &Expression{
				Kind: KindBinary,
				Span: Span{Start: leftStart, End: right.Span.End},
				payload: &BinaryExpr{
					Op:  Intersection,
					LHS: left,
					RHS: right,
				},
			}
		} else if ch == '~' {
			p.advance()
			p.skipWhitespace()
			right, err := p.parseNegatePrefix()
			if err != nil {
				return nil, err
			}
			leftStart := left.Span.Start
			left = &Expression{
				Kind: KindBinary,
				Span: Span{Start: leftStart, End: right.Span.End},
				payload: &BinaryExpr{
					Op:  Difference,
					LHS: left,
					RHS: right,
				},
			}
		} else {
			break
		}
	}
	return left, nil
}

// Level 2: prefix negate (~)
func (p *parser) parseNegatePrefix() (*Expression, error) {
	p.skipWhitespace()
	if p.peek() == '~' {
		start := p.pos
		p.advance()
		p.skipWhitespace()
		arg, err := p.parseNegatePrefix()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: start, End: p.pos},
			payload: &UnaryExpr{
				Op:  Negate,
				Arg: arg,
			},
		}, nil
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (*Expression, error) {
	p.skipWhitespace()
	start := p.pos

	ch := p.peek()
	switch {
	case ch == '(':
		p.advance()
		p.skipWhitespace()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.peek() != ')' {
			return nil, p.syntaxError("expected ')'")
		}
		p.advance()
		expr.Span = Span{Start: start, End: p.pos}
		return expr, nil

	case ch == '"':
		s, err := p.parseStringLiteralValue()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind:    KindString,
			Span:    Span{Start: start, End: p.pos},
			payload: &StringExpr{Value: s},
		}, nil

	case ch == '\'':
		s := p.parseRawStringLiteralValue()
		return &Expression{
			Kind:    KindString,
			Span:    Span{Start: start, End: p.pos},
			payload: &StringExpr{Value: s},
		}, nil

	case isFilesetIdentifierStart(ch):
		return p.parseSymbolOrFunctionOrPattern(start)

	default:
		if p.atEnd() {
			return nil, p.syntaxError("unexpected end of input")
		}
		return nil, p.syntaxErrorf("unexpected character %q", ch)
	}
}