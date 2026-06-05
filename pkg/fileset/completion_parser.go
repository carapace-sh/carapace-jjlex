package fileset

import (
	"unicode/utf8"
)

// ParseForCompletion parses a partial fileset expression and returns a
// CompletionContext describing what is expected at the end of the input.
func ParseForCompletion(input string) *CompletionContext {
	cursor := len(input)
	p := &compParser{
		input:  input,
		pos:    0,
		cursor: cursor,
		ctx:    &CompletionContext{},
	}
	p.skipWS()
	p.parseExpr()
	if len(p.ctx.ExpectedTokens) == 0 {
		p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, ExpectedExpression)
	}
	p.ctx.ExpectedTokens = dedupTokens(p.ctx.ExpectedTokens)
	p.ctx.ValidOperators = dedupOperators(p.ctx.ValidOperators)
	return p.ctx
}

type compParser struct {
	input  string
	pos    int
	cursor int
	ctx    *CompletionContext

	// consumed is true when we have consumed at least one token of input
	// before reaching the cursor.
	consumed bool

	// afterOperator is true when we have consumed an operator but haven't
	// started parsing the RHS.
	afterOperator bool

	// Stack of function parse states for nested calls
	funcStack []*funcParseState

	// innermostFunc is the deepest function context we've set
	innermostFunc *FunctionContext

	// lastExpr is the most recently parsed expression (set by parsePrimary)
	lastExpr *Expression

	// exprStart is the input position where the current expression started
	exprStart int
}

type funcParseState struct {
	name     string
	args     []*Expression
	argIndex int
}

func (p *compParser) atCursorOrEnd() bool {
	return p.pos >= len(p.input) || p.pos >= p.cursor
}

func (p *compParser) peek() rune {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

func (p *compParser) advance() rune {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += w
	p.consumed = true
	return r
}

func (p *compParser) skipWS() {
	for p.pos < len(p.input) && p.pos < p.cursor {
		r, w := utf8.DecodeRuneInString(p.input[p.pos:])
		if !isWhitespace(r) {
			break
		}
		p.pos += w
	}
}

func (p *compParser) matchStringAtFullInput(pos int, s string) bool {
	end := pos + len(s)
	if end > len(p.input) {
		return false
	}
	return p.input[pos:end] == s
}

func (p *compParser) addExpected(t ExpectedToken) {
	p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, t)
}

func (p *compParser) addOperator(op, desc string) {
	p.ctx.ValidOperators = append(p.ctx.ValidOperators, ValidOperator{Op: op, Description: desc})
}

func (p *compParser) afterExpression() {
	p.addExpected(ExpectedOperator)
	p.addOperator("|", "union")
	p.addOperator("&", "intersection")
	p.addOperator("~", "difference")

	if len(p.funcStack) > 0 {
		p.addExpected(ExpectedClosingParen)
		p.addExpected(ExpectedComma)
	}
}

func (p *compParser) beforeExpression() {
	p.addExpected(ExpectedExpression)
}

func (p *compParser) parseExpr() {
	p.parseInfixLevel0()
}

// Operator precedence (same as main parser):
// 0: union (|)
// 1: intersection (&), difference (~)
// 2: prefix negate (~)
// 3: primary

func (p *compParser) parseInfixLevel0() {
	p.parseInfixLevel1()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.consumed {
				p.afterExpression()
			} else {
				p.beforeExpression()
			}
			return
		}
		ch := p.peek()
		if ch == '|' {
			p.advance()
			p.afterOperator = true
			p.skipWS()
			if p.atCursorOrEnd() {
				p.afterExpression()
				p.beforeExpression()
				return
			}
			p.afterOperator = false
			p.parseInfixLevel1()
		} else {
			break
		}
	}
}

func (p *compParser) parseInfixLevel1() {
	p.parseNegatePrefix()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.consumed {
				p.afterExpression()
			}
			return
		}
		ch := p.peek()
		if ch == '&' {
			p.advance()
			p.afterOperator = true
			p.skipWS()
			if p.atCursorOrEnd() {
				p.afterExpression()
				p.beforeExpression()
				return
			}
			p.afterOperator = false
			p.parseNegatePrefix()
		} else if ch == '~' {
			p.advance()
			p.afterOperator = true
			p.skipWS()
			if p.atCursorOrEnd() {
				p.afterExpression()
				p.beforeExpression()
				return
			}
			p.afterOperator = false
			p.parseNegatePrefix()
		} else {
			break
		}
	}
}

func (p *compParser) parseNegatePrefix() {
	p.skipWS()
	if p.atCursorOrEnd() {
		if p.consumed {
			p.addExpected(ExpectedOperator)
			p.addOperator("~", "negate/difference")
		} else {
			p.beforeExpression()
			p.addExpected(ExpectedOperator)
			p.addOperator("~", "negate")
		}
		return
	}
	if p.peek() == '~' {
		p.advance()
		p.skipWS()
		if p.atCursorOrEnd() {
			p.beforeExpression()
			return
		}
		p.parseNegatePrefix()
		return
	}
	p.parsePrimary()
}

func (p *compParser) parsePrimary() {
	p.skipWS()
	start := p.pos
	p.exprStart = start
	if p.atCursorOrEnd() {
		p.beforeExpression()
		return
	}

	ch := p.peek()
	switch {
	case ch == '(':
		p.parseParenthesized()
		if p.lastExpr != nil {
			p.lastExpr.Span = Span{Start: start, End: p.pos}
		}
	case ch == '"':
		value := p.parseStringLiteralCompletion()
		p.lastExpr = &Expression{Kind: KindString, Span: Span{Start: start, End: p.pos}, payload: &StringExpr{Value: value}}
	case ch == '\'':
		value := p.parseRawStringLiteralCompletion()
		p.lastExpr = &Expression{Kind: KindString, Span: Span{Start: start, End: p.pos}, payload: &StringExpr{Value: value}}
	case isFilesetIdentifierStart(ch):
		p.parseSymbolOrFunctionCompletion()
	default:
		p.beforeExpression()
	}
}
