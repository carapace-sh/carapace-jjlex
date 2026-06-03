package template

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

// Parse parses a jj template expression string into an AST.
func Parse(input string) (*Expression, error) {
	p := &parser{input: input}
	p.skipWhitespace()
	start := p.pos
	expr, err := p.parseTemplate()
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

// IsIdentifier checks if the text is a valid template identifier.
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

func (p *parser) matchString(s string) bool {
	if p.pos+len(s) > len(p.input) {
		return false
	}
	return p.input[p.pos:p.pos+len(s)] == s
}

func (p *parser) save() int {
	return p.pos
}

func (p *parser) restore(pos int) {
	p.pos = pos
}

// parseTemplate parses the concatenation level: expression ++ expression ++
func (p *parser) parseTemplate() (*Expression, error) {
	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	var nodes []*Expression
	if left.Kind == KindConcat {
		nodes = append(nodes, left.ConcatNodes()...)
	} else {
		nodes = append(nodes, left)
	}
	for {
		p.skipWhitespace()
		if !p.matchString("++") {
			break
		}
		p.pos += 2
		p.skipWhitespace()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, right)
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return &Expression{
		Kind: KindConcat,
		Span: Span{Start: nodes[0].Span.Start, End: nodes[len(nodes)-1].Span.End},
		payload: &ConcatExpr{Nodes: nodes},
	}, nil
}

// parseExpression uses a Pratt parser for all infix/prefix operators
func (p *parser) parseExpression() (*Expression, error) {
	return p.parsePratt(precLogicalOr)
}

func (p *parser) parsePratt(minPrec int) (*Expression, error) {
	left, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		op, prec, rightAssoc := peekInfixOp(p.input, p.pos)
		if op == "" || prec < minPrec {
			break
		}
		nextMinPrec := prec + 1
		if rightAssoc {
			nextMinPrec = prec
		}
		p.pos += len(op)
		p.skipWhitespace()
		right, err := p.parsePratt(nextMinPrec)
		if err != nil {
			return nil, err
		}
		binOp := infixOpStringToBinaryOp(op)
		left = &Expression{
			Kind: KindBinary,
			Span: Span{Start: left.Span.Start, End: right.Span.End},
			payload: &BinaryExpr{
				Op:  binOp,
				LHS: left,
				RHS: right,
			},
		}
	}
	return left, nil
}

func (p *parser) parsePrefix() (*Expression, error) {
	p.skipWhitespace()
	start := p.pos
	if p.peek() == '!' {
		p.advance()
		p.skipWhitespace()
		arg, err := p.parsePrefix()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: start, End: arg.Span.End},
			payload: &UnaryExpr{
				Op:  LogicalNot,
				Arg: arg,
			},
		}, nil
	}
	if p.peek() == '-' {
		p.advance()
		p.skipWhitespace()
		arg, err := p.parsePrefix()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: start, End: arg.Span.End},
			payload: &UnaryExpr{
				Op:  Negate,
				Arg: arg,
			},
		}, nil
	}
	return p.parseTerm()
}

// parseTerm handles: primary followed by zero or more method calls (.function())
func (p *parser) parseTerm() (*Expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		p.skipWhitespace()
		if p.peek() != '.' {
			break
		}
		saved := p.save()
		p.advance() // consume .
		p.skipWhitespace()
		if p.atEnd() || !isIdentifierStart(p.peek()) {
			p.restore(saved)
			break
		}
		identStart := p.pos
		if !p.scanIdentifier() {
			p.restore(saved)
			break
		}
		ident := p.input[identStart:p.pos]
		p.skipWhitespace()
		if p.peek() != '(' {
			p.restore(saved)
			break
		}
		if !isFunctionName(ident) {
			p.restore(saved)
			break
		}
		args, keywordArgs, err := p.parseFunctionArgs()
		if err != nil {
			return nil, err
		}
		funcExpr := &FunctionCallExpr{
			Name:        ident,
			Args:        args,
			KeywordArgs: keywordArgs,
		}
		expr = &Expression{
			Kind: KindMethodCall,
			Span: Span{Start: expr.Span.Start, End: p.pos},
			payload: &MethodCallExpr{
				Object:   expr,
				Function: funcExpr,
			},
		}
	}
	return expr, nil
}

// parsePrimary handles: (template) | function | lambda | pattern | identifier | string | integer | boolean
func (p *parser) parsePrimary() (*Expression, error) {
	p.skipWhitespace()
	start := p.pos
	ch := p.peek()

	switch {
	case ch == '(':
		p.advance()
		p.skipWhitespace()
		expr, err := p.parseTemplate()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.atEnd() || p.peek() != ')' {
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

	case ch == '|':
		return p.parseLambdaOrLogicalOr(start)

	case isIdentifierStart(ch):
		return p.parseIdentFuncOrPattern(start)

	case ch >= '0' && ch <= '9':
		return p.parseIntegerLiteral(start)

	default:
		if p.atEnd() {
			return nil, p.syntaxError("unexpected end of input")
		}
		return nil, p.syntaxErrorf("unexpected character %q", ch)
	}
}

// parseLambdaOrLogicalOr handles the ambiguity between || (logical OR) and || (lambda).
// In primary position, || is always a zero-arg lambda. The || infix operator is
// handled by the Pratt parser at the expression level.
func (p *parser) parseLambdaOrLogicalOr(start int) (*Expression, error) {
	// If we see | followed by another |, it's a zero-arg lambda || expr
	if p.pos+1 < len(p.input) && p.input[p.pos+1] == '|' {
		// Zero-arg lambda: || expr
		p.advance() // consume first |
		p.advance() // consume second |
		p.skipWhitespace()
		if p.atEnd() {
			return nil, p.syntaxError("unexpected end of input after lambda parameters")
		}
		body, err := p.parseTemplate()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindLambda,
			Span: Span{Start: start, End: body.Span.End},
			payload: &LambdaExpr{
				Params: nil,
				Body:   body,
			},
		}, nil
	}
	// Regular lambda: |params| expr
	return p.parseLambda(start)
}

func (p *parser) parseLambda(start int) (*Expression, error) {
	p.advance() // consume first |
	p.skipWhitespace()
	var params []string

	if p.peek() != '|' {
		for {
			p.skipWhitespace()
			paramStart := p.pos
			if !p.scanIdentifier() {
				return nil, p.syntaxError("expected parameter name in lambda")
			}
			param := p.input[paramStart:p.pos]
			if param == "true" || param == "false" {
				return nil, p.syntaxErrorf("keyword %q cannot be used as parameter name", param)
			}
			for _, existing := range params {
				if existing == param {
					return nil, p.syntaxErrorf("redefined function parameter %q", param)
				}
			}
			params = append(params, param)
			p.skipWhitespace()
			if p.peek() == ',' {
				p.advance()
				p.skipWhitespace()
				if p.peek() == '|' {
					// trailing comma before | is ok
					break
				}
				continue
			}
			break
		}
	}

	p.skipWhitespace()
	if p.atEnd() || p.peek() != '|' {
		return nil, p.syntaxError("expected '|' to close lambda parameters")
	}
	p.advance() // consume closing |

	p.skipWhitespace()
	if p.atEnd() {
		return nil, p.syntaxError("unexpected end of input after lambda parameters")
	}

	body, err := p.parseTemplate()
	if err != nil {
		return nil, err
	}

	return &Expression{
		Kind: KindLambda,
		Span: Span{Start: start, End: body.Span.End},
		payload: &LambdaExpr{
			Params: params,
			Body:   body,
		},
	}, nil
}

func (p *parser) parseIdentFuncOrPattern(start int) (*Expression, error) {
	identStart := p.pos
	if !p.scanIdentifier() {
		return nil, p.syntaxError("expected identifier")
	}
	baseIdentEnd := p.pos
	// Extend identifier to include pattern_identifier (with dashes) if followed by dash
	// Pattern names like "regex-i" include dashes before the colon
	// Save position so we can backtrack if pattern doesn't match
	p.scanPatternIdentifierSuffix()
	ident := p.input[identStart:p.pos]
	identEnd := p.pos

	// Check for boolean literals
	if ident == "true" || ident == "false" {
		p.skipWhitespace()
		if p.peek() == '(' {
			return nil, p.syntaxErrorf("keyword %q cannot be used as function name", ident)
		}
		if identEnd < len(p.input) && p.input[identEnd] == ':' && !p.matchStringAt(identEnd, "::") {
			return nil, p.syntaxErrorf("keyword %q cannot be used as pattern name", ident)
		}
		val := ident == "true"
		return &Expression{
			Kind:    KindBoolean,
			Span:    Span{Start: start, End: p.pos},
			payload: &BooleanExpr{Value: val},
		}, nil
	}

	// Check for function call: identifier(
	// Function names cannot contain dashes, so use baseIdentEnd
	p.skipWhitespace()
	if p.peek() == '(' && isFunctionName(p.input[identStart:baseIdentEnd]) {
		// Backtrack to base identifier for function call
		p.pos = baseIdentEnd
		args, keywordArgs, err := p.parseFunctionArgs()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindFunctionCall,
			Span: Span{Start: start, End: p.pos},
			payload: &FunctionCallExpr{
				Name:        p.input[identStart:baseIdentEnd],
				Args:        args,
				KeywordArgs: keywordArgs,
			},
		}, nil
	}

	// Check for pattern: identifier:value (no whitespace allowed around :)
	if identEnd < len(p.input) && p.input[identEnd] == ':' && !p.matchStringAt(identEnd, "::") && isPatternIdentifier(ident) {
		p.pos = identEnd + 1 // consume :
		// Pattern value must immediately follow : with no whitespace (unless parenthesized)
		if p.atEnd() || isWhitespace(p.peek()) {
			// Space after : in pattern is not allowed without parentheses
			p.pos = identEnd
		} else {
			value, err := p.parsePrefix()
			if err != nil {
				return nil, err
			}
			return &Expression{
				Kind:    KindPattern,
				Span:    Span{Start: start, End: p.pos},
				payload: &PatternExpr{Name: ident, Value: value},
			}, nil
		}
	}

	// Not a pattern — backtrack to base identifier (without dash suffix)
	// so that "x-y" becomes identifier "x" with "-y" parsed as infix subtraction
	p.pos = baseIdentEnd
	ident = p.input[identStart:baseIdentEnd]

	// Plain identifier
	return &Expression{
		Kind:    KindIdentifier,
		Span:    Span{Start: start, End: p.pos},
		payload: &IdentifierExpr{Name: ident},
	}, nil
}

func (p *parser) parseFunctionArgs() ([]*Expression, []KeywordArg, error) {
	if p.peek() != '(' {
		return nil, nil, p.syntaxError("expected '('")
	}
	p.advance() // consume (
	p.skipWhitespace()

	var args []*Expression
	var keywordArgs []KeywordArg

	if p.peek() == ')' {
		p.advance()
		return args, keywordArgs, nil
	}

	for {
		p.skipWhitespace()

		// Check for end of arguments
		if p.peek() == ')' {
			break
		}

		// Check for keyword argument: identifier = expression
		if kwName, isKw, err := p.tryParseKeywordArg(); err != nil {
			return nil, nil, err
		} else if isKw {
			p.skipWhitespace()
			p.advance() // consume =
			p.skipWhitespace()
			value, err := p.parseTemplate()
			if err != nil {
				return nil, nil, err
			}
			keywordArgs = append(keywordArgs, KeywordArg{Name: kwName, Value: value})
			p.skipWhitespace()
			if p.peek() == ',' {
				p.advance()
				p.skipWhitespace()
			}
			continue
		}

		// Positional argument
		arg, err := p.parseTemplate()
		if err != nil {
			return nil, nil, err
		}
		args = append(args, arg)

		p.skipWhitespace()
		if p.peek() == ',' {
			p.advance()
			p.skipWhitespace()
			// Check for trailing comma before )
			if p.peek() == ')' {
				break
			}
		} else {
			break
		}
	}

	p.skipWhitespace()
	if p.atEnd() || p.peek() != ')' {
		return nil, nil, p.syntaxError("expected ')' in function call")
	}
	p.advance() // consume )

	return args, keywordArgs, nil
}

func (p *parser) tryParseKeywordArg() (string, bool, error) {
	saved := p.save()
	start := p.pos
	if !p.scanIdentifier() {
		return "", false, nil
	}
	name := p.input[start:p.pos]
	nameEnd := p.pos
	if name == "true" || name == "false" {
		// Check if this boolean is followed by = (keyword arg syntax)
		// If so, return an error — boolean literals cannot be keyword arg names
		pos := nameEnd
		for pos < len(p.input) && isWhitespace(rune(p.input[pos])) {
			pos++
		}
		if pos < len(p.input) && p.input[pos] == '=' {
			return "", false, p.syntaxErrorf("keyword %q cannot be used as function parameter name", name)
		}
		p.restore(saved)
		return "", false, nil
	}

	// Skip whitespace and check for =
	pos := nameEnd
	for pos < len(p.input) && isWhitespace(rune(p.input[pos])) {
		pos++
	}
	if pos < len(p.input) && p.input[pos] == '=' {
		p.pos = nameEnd
		return name, true, nil
	}

	p.restore(saved)
	return "", false, nil
}

func (p *parser) matchStringAt(pos int, s string) bool {
	if pos+len(s) > len(p.input) {
		return false
	}
	return p.input[pos:pos+len(s)] == s
}

func isPatternIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			if !isPatternIdentifierStart(ch) {
				return false
			}
		} else {
			if !isPatternIdentifierPart(ch) {
				return false
			}
		}
	}
	return true
}