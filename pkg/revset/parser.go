package revset

import (
	"fmt"
	"unicode"
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
	lastContent int // position right after last non-whitespace content
}

// Parse parses a revset expression string into an AST.
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
	// Use lastContent to avoid including trailing whitespace in span
	contentEnd := min(p.lastContent, len(p.input))
	if expr.Span.End > contentEnd || expr.Span.Start < start {
		expr.Span = Span{Start: start, End: contentEnd}
	}
	return expr, nil
}

// IsIdentifier checks if the text is a valid revset identifier.
func IsIdentifier(text string) bool {
	p := &parser{input: text}
	start := p.pos
	if !p.scanIdentifier() {
		return false
	}
	return p.pos == len(text) && p.pos > start
}

// Format returns a normalized string representation of the expression.
// It produces a canonical form that parses to the same AST.
func Format(expr *Expression) string {
	return expr.String()
}

// ParseSymbol parses text as a revset symbol, rejecting empty strings.
// Leading/trailing whitespace is NOT ignored.
func ParseSymbol(text string) (string, error) {
	p := &parser{input: text}
	// Don't skip whitespace - leading whitespace is an error
	name, err := p.parseSymbolName()
	if err != nil {
		return "", err
	}
	// Don't skip whitespace - trailing whitespace is an error
	if p.pos < len(p.input) {
		return "", &ParseError{Message: "unexpected token after symbol", Span: Span{Start: p.pos, End: len(text)}}
	}
	if name == "" {
		return "", &ParseError{Message: "Expected non-empty string", Span: Span{Start: 0, End: len(text)}}
	}
	return name, nil
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

func (p *parser) notPrefixOp(op, similarOp, description string) *ParseError {
	return &ParseError{
		Message: fmt.Sprintf("`%s` is not a prefix operator", op),
		Span:    Span{Start: p.pos, End: p.pos + len(op)},
	}
}

func (p *parser) notPostfixOp(op, similarOp, description string) *ParseError {
	return &ParseError{
		Message: fmt.Sprintf("`%s` is not a postfix operator", op),
		Span:    Span{Start: p.pos, End: p.pos + len(op)},
	}
}

func (p *parser) notInfixOp(op, similarOp, description string) *ParseError {
	return &ParseError{
		Message: fmt.Sprintf("`%s` is not an infix operator", op),
		Span:    Span{Start: p.pos, End: p.pos + len(op)},
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

// Operator precedence (from lowest to highest):
// 0: union (|) - left assoc
// 1: intersection (&), difference (~) - left assoc
// 2: infix dag range (::), range (..) - left assoc but can't nest
// 3: prefix negate (~)
// 4: prefix dag range (::), prefix range (..)
// 5: postfix dag range (::), postfix range (..)
// 6: postfix parents (-), children (+)

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
		} else if ch == '+' {
			return nil, p.notInfixOp("+", "|", "union")
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
		} else if ch == '-' {
			return nil, p.notInfixOp("-", "~", "difference")
		} else {
			break
		}
	}
	return left, nil
}

// Level 2: infix dag range (::), range (..)
// These can't be nested without parentheses.
// Only called when we have a LHS and see a range operator.
// The RHS is parsed as neighbors_expression (not full range_expression).
func (p *parser) parseInfixRangeOp(lhs *Expression) (*Expression, error) {
	p.skipWhitespace()
	if p.matchString("::") {
		p.pos += 2
		p.skipWhitespace()
		right, err := p.parsePostfixOps()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindBinary,
			Span: Span{Start: lhs.Span.Start, End: right.Span.End},
			payload: &BinaryExpr{
				Op:  DagRange,
				LHS: lhs,
				RHS: right,
			},
		}, nil
	}
	if p.matchString(":") {
		return nil, p.notInfixOp(":", "::", "DAG range")
	}
	if p.matchString("..") {
		p.pos += 2
		p.skipWhitespace()
		right, err := p.parsePostfixOps()
		if err != nil {
			return nil, err
		}
		return &Expression{
			Kind: KindBinary,
			Span: Span{Start: lhs.Span.Start, End: right.Span.End},
			payload: &BinaryExpr{
				Op:  Range,
				LHS: lhs,
				RHS: right,
			},
		}, nil
	}
	return lhs, nil
}

// Level 3: prefix negate (~)
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
	return p.parseRangeExpression()
}

// parseRangeExpression handles the range_expression rule from Pest grammar:
//
//	range_expression = neighbors_expression ~ range_ops ~ neighbors_expression
//	                  | neighbors_expression ~ range_post_ops
//	                  | range_pre_ops ~ neighbors_expression
//	                  | neighbors_expression
//	                  | range_all_ops
//
// Key property: range expressions CANNOT be nested without parentheses.
func (p *parser) parseRangeExpression() (*Expression, error) {
	p.skipWhitespace()

	// Check for nullary :: or .. first
	if p.matchString("::") {
		saved := p.pos
		p.pos += 2
		p.skipWhitespace()
		// Nullary :: if followed by end, ), comma, or a non-expression-start operator
		if p.atEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' {
			return &Expression{
				Kind: KindDagRangeAll,
				Span: Span{Start: saved, End: saved + 2},
			}, nil
		}

		// Check for whitespace after :: (not allowed for prefix)
		if saved+2 < len(p.input) && isWhitespace(rune(p.input[saved+2])) {
			p.restore(saved)
			return nil, p.syntaxError("space not allowed after `::` prefix operator")
		}

		// Prefix :: (ancestors)
		arg, err := p.parsePostfixOps()
		if err != nil {
			return nil, err
		}
		// After prefix ::, the result is a range_expression which CANNOT
		// participate in further range operations. In the Rust Pest grammar,
		// the infix range operator's operands must be neighbors_expression,
		// not range_expression.
		// However, we do need to check for a single infix :: right after,
		// because "::foo::bar" could mean (::foo)::bar (prefix then infix).
		// But the Rust grammar rejects this - ranges can't be nested.
		// We do NOT call parseInfixRangeOp here.
		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: saved, End: p.pos},
			payload: &UnaryExpr{
				Op:  DagRangePre,
				Arg: arg,
			},
		}, nil
	}

	if p.matchString("..") {
		saved := p.pos
		p.pos += 2
		p.skipWhitespace()
		if p.atEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' {
			return &Expression{
				Kind: KindRangeAll,
				Span: Span{Start: saved, End: saved + 2},
			}, nil
		}

		if saved+2 < len(p.input) && isWhitespace(rune(p.input[saved+2])) {
			p.restore(saved)
			return nil, p.syntaxError("space not allowed after `..` prefix operator")
		}

		arg, err := p.parsePostfixOps()
		if err != nil {
			return nil, err
		}
		// Same as :: prefix above: don't allow further range operations
		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: saved, End: p.pos},
			payload: &UnaryExpr{
				Op:  RangePre,
				Arg: arg,
			},
		}, nil
	}

	// Check for compat prefix : (should error)
	if p.peek() == ':' && !p.matchStringAt(p.pos, "::") {
		return nil, p.notPrefixOp(":", "::", "ancestors")
	}

	// Parse neighbors_expression
	node, err := p.parsePostfixOps()
	if err != nil {
		return nil, err
	}

	// Check for postfix or infix range operators
	p.skipWhitespace()

	if p.matchString("::") {
		saved := p.pos
		p.pos += 2
		p.skipWhitespace()

		// If followed by something that could be an expression, it's infix
		if !p.atEnd() && p.peek() != ')' && p.peek() != ',' && p.peek() != '|' && p.peek() != '&' && p.peek() != '~' {
			// Infix ::
			p.restore(saved)
			return p.parseInfixRangeOp(node)
		}

		// Postfix ::
		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: node.Span.Start, End: p.pos},
			payload: &UnaryExpr{
				Op:  DagRangePost,
				Arg: node,
			},
		}, nil
	}

	if p.matchString("..") {
		saved := p.pos
		p.pos += 2
		p.skipWhitespace()

		if !p.atEnd() && p.peek() != ')' && p.peek() != ',' && p.peek() != '|' && p.peek() != '&' && p.peek() != '~' {
			p.restore(saved)
			return p.parseInfixRangeOp(node)
		}

		return &Expression{
			Kind: KindUnary,
			Span: Span{Start: node.Span.Start, End: p.pos},
			payload: &UnaryExpr{
				Op:  RangePost,
				Arg: node,
			},
		}, nil
	}

	// Check for single : (compat, should error as infix)
	if p.peek() == ':' {
		return nil, p.notInfixOp(":", "::", "DAG range")
	}

	return node, nil
}

// parsePostfixOps handles postfix -, +, ^ (parents/children/compat)
func (p *parser) parsePostfixOps() (*Expression, error) {
	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		p.skipWhitespace()
		ch := p.peek()
		switch ch {
		case '-':
			nodeStart := node.Span.Start
			p.advance()
			node = &Expression{
				Kind: KindUnary,
				Span: Span{Start: nodeStart, End: p.pos},
				payload: &UnaryExpr{
					Op:  Parents,
					Arg: node,
				},
			}
		case '+':
			nodeStart := node.Span.Start
			p.advance()
			node = &Expression{
				Kind: KindUnary,
				Span: Span{Start: nodeStart, End: p.pos},
				payload: &UnaryExpr{
					Op:  Children,
					Arg: node,
				},
			}
		case '^':
			p.advance()
			return nil, p.notPostfixOp("^", "-", "parents")
		default:
			return node, nil
		}
	}
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
		// Preserve span including parentheses
		expr.Span = Span{Start: start, End: p.pos}
		return expr, nil

	case ch == '"':
		s, err := p.parseStringLiteralValue()
		if err != nil {
			return nil, err
		}
		stringEnd := p.pos
		p.skipWhitespace()
		expr := &Expression{
			Kind:    KindString,
			Span:    Span{Start: start, End: stringEnd},
			payload: &StringExpr{Value: s},
		}
		return p.parseAtSuffix(expr, start), nil

	case ch == '\'':
		s := p.parseRawStringLiteralValue()
		stringEnd := p.pos
		p.skipWhitespace()
		expr := &Expression{
			Kind:    KindString,
			Span:    Span{Start: start, End: stringEnd},
			payload: &StringExpr{Value: s},
		}
		return p.parseAtSuffix(expr, start), nil

	case ch == '@':
		p.advance()
		return &Expression{
			Kind: KindAtCurrentWorkspace,
			Span: Span{Start: start, End: p.pos},
		}, nil

	case isIdentifierStart(ch):
		return p.parseSymbolOrFunction(start)

	default:
		if p.atEnd() {
			return nil, p.syntaxError("unexpected end of input")
		}
		return nil, p.syntaxErrorf("unexpected character %q", ch)
	}
}

// parseAtSuffix checks for @ suffix after a string/identifier.
func (p *parser) parseAtSuffix(expr *Expression, start int) *Expression {
	if p.peek() != '@' {
		return expr
	}
	p.advance() // consume @
	p.skipWhitespace()

	if p.atEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' || p.peek() == '-' || p.peek() == '+' {
		name := p.extractString(expr)
		return &Expression{
			Kind:    KindAtWorkspace,
			Span:    Span{Start: start, End: p.pos},
			payload: &AtWorkspaceExpr{Name: name},
		}
	}

	name := p.extractString(expr)
	remote := p.parseAtRemotePart()
	return &Expression{
		Kind:    KindRemoteSymbol,
		Span:    Span{Start: start, End: p.pos},
		payload: &RemoteSymbolExpr{Name: name, Remote: remote},
	}
}

func (p *parser) extractString(expr *Expression) string {
	switch expr.Kind {
	case KindIdentifier:
		return expr.payload.(*IdentifierExpr).Name
	case KindString:
		return expr.payload.(*StringExpr).Value
	}
	return ""
}

func (p *parser) parseAtRemotePart() string {
	ch := p.peek()
	switch {
	case ch == '"':
		s, _ := p.parseStringLiteralValue()
		return s
	case ch == '\'':
		return p.parseRawStringLiteralValue()
	case isIdentifierStart(ch):
		start := p.pos
		p.scanIdentifier()
		return p.input[start:p.pos]
	default:
		return ""
	}
}

func (p *parser) parseSymbolOrFunction(start int) (*Expression, error) {
	identStart := p.pos
	if !p.scanIdentifier() {
		return nil, p.syntaxError("expected identifier")
	}
	ident := p.input[identStart:p.pos]
	identEnd := p.pos

	// Check for function call: identifier(
	p.skipWhitespace()
	if p.peek() == '(' && p.isFunctionName(ident) {
		return p.parseFunctionCall(ident, start)
	}

	// Check for pattern: strict_identifier : value (no whitespace allowed around :)
	// Pattern name must be a strict_identifier, and : must immediately follow
	// the identifier with no whitespace, and the value must immediately follow :
	if identEnd < len(p.input) && p.input[identEnd] == ':' && !p.matchStringAt(identEnd, "::") && p.isStrictIdentifier(ident) {
		p.pos = identEnd + 1 // consume : (no whitespace skipped)
		// Pattern value must immediately follow : with no whitespace
		// (matching Rust Pest grammar behavior where "exact: foo" is a syntax error)
		if p.atEnd() || isWhitespace(p.peek()) {
			// Space after : in pattern is not allowed without parentheses
			// Revert and fall through to treat this as an expression
			p.pos = identEnd
		} else {
			value, err := p.parsePatternValue()
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

	// Plain identifier (possibly with @ suffix)
	expr := &Expression{
		Kind:    KindIdentifier,
		Span:    Span{Start: start, End: p.pos},
		payload: &IdentifierExpr{Name: ident},
	}
	return p.parseAtSuffix(expr, start), nil
}

func (p *parser) isFunctionName(ident string) bool {
	if len(ident) == 0 {
		return false
	}
	for i, ch := range ident {
		if i == 0 {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
				return false
			}
		} else {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return false
			}
		}
	}
	return true
}

func (p *parser) isStrictIdentifier(ident string) bool {
	if len(ident) == 0 {
		return false
	}
	parts := splitIdentifierParts(ident, ".-+")
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, ch := range part {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '/') {
				return false
			}
		}
	}
	return true
}

func (p *parser) parsePatternValue() (*Expression, error) {
	return p.parsePostfixOps()
}

func (p *parser) parseFunctionCall(name string, start int) (*Expression, error) {
	p.advance() // consume (
	p.skipWhitespace()

	var args []*Expression
	var keywordArgs []KeywordArg

	for !p.atEnd() && p.peek() != ')' {
		p.skipWhitespace()

		// Check for keyword argument: strict_identifier = expression
		if kwName, isKw := p.tryParseKeywordArg(); isKw {
			p.skipWhitespace()
			if p.peek() == '=' {
				p.advance() // consume =
				p.skipWhitespace()
				value, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				keywordArgs = append(keywordArgs, KeywordArg{Name: kwName, Value: value})
				p.skipWhitespace()
				if p.peek() == ',' {
					p.advance()
					p.skipWhitespace()
				}
				continue
			}
		}

		// Regular positional argument
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		p.skipWhitespace()
		if p.peek() == ',' {
			p.advance()
			p.skipWhitespace()
		} else {
			break
		}
	}

	p.skipWhitespace()
	if p.peek() != ')' {
		return nil, p.syntaxError("expected ')' in function call")
	}
	p.advance() // consume )

	return &Expression{
		Kind: KindFunctionCall,
		Span: Span{Start: start, End: p.pos},
		payload: &FunctionCallExpr{
			Name:        name,
			Args:        args,
			KeywordArgs: keywordArgs,
		},
	}, nil
}

func (p *parser) tryParseKeywordArg() (string, bool) {
	saved := p.save()
	start := p.pos
	if !p.scanStrictIdentifier() {
		return "", false
	}
	name := p.input[start:p.pos]
	nameEnd := p.pos

	// Skip whitespace and check for =
	pos := nameEnd
	for pos < len(p.input) && isWhitespace(rune(p.input[pos])) {
		pos++
	}
	if pos < len(p.input) && p.input[pos] == '=' {
		p.pos = nameEnd
		return name, true
	}

	p.restore(saved)
	return "", false
}

func (p *parser) scanStrictIdentifier() bool {
	if p.atEnd() {
		return false
	}
	start := p.pos
	if !isStrictIdentifierPart(p.peek()) {
		return false
	}
	for !p.atEnd() && isStrictIdentifierPart(p.peek()) {
		p.advance()
	}
	for !p.atEnd() {
		ch := p.peek()
		if ch == '.' || ch == '-' || ch == '+' {
			saved := p.pos
			p.advance()
			if p.atEnd() || !isStrictIdentifierPart(p.peek()) {
				p.pos = saved
				break
			}
			for !p.atEnd() && isStrictIdentifierPart(p.peek()) {
				p.advance()
			}
		} else {
			break
		}
	}
	return p.pos > start
}

// scanIdentifier scans a jj revset identifier.
// identifier = identifier_part ~ (("." | "-"+ | "+") ~ identifier_part)*
func (p *parser) scanIdentifier() bool {
	if p.atEnd() {
		return false
	}
	start := p.pos
	if !p.scanIdentifierPart() {
		return false
	}

	for !p.atEnd() {
		ch := p.peek()
		if ch == '.' {
			saved := p.pos
			p.advance()
			if p.atEnd() || !p.scanIdentifierPart() {
				p.pos = saved
				break
			}
		} else if ch == '-' {
			saved := p.pos
			dashCount := 0
			for !p.atEnd() && p.peek() == '-' {
				p.advance()
				dashCount++
			}
			if dashCount > 0 && !p.atEnd() && p.scanIdentifierPart() {
				continue
			}
			p.pos = saved
			break
		} else if ch == '+' {
			saved := p.pos
			p.advance()
			if p.atEnd() || !p.scanIdentifierPart() {
				p.pos = saved
				break
			}
		} else {
			break
		}
	}

	return p.pos > start
}

func (p *parser) scanIdentifierPart() bool {
	start := p.pos
	for !p.atEnd() && isIdentifierPart(p.peek()) {
		p.advance()
	}
	return p.pos > start
}

func (p *parser) parseStringLiteralValue() (string, error) {
	if p.peek() != '"' {
		return "", p.syntaxError("expected string literal")
	}
	p.advance()

	var result []rune
	for {
		if p.atEnd() {
			return "", p.syntaxError("unterminated string literal")
		}
		ch := p.peek()
		if ch == '"' {
			p.advance()
			return string(result), nil
		}
		if ch == '\\' {
			p.advance()
			if p.atEnd() {
				return "", p.syntaxError("unterminated escape sequence")
			}
			escaped := p.advance()
			switch escaped {
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case 'n':
				result = append(result, '\n')
			case '0':
				result = append(result, '\000')
			case 'e':
				result = append(result, '\x1b')
			case 'x':
				if p.atEnd() {
					return "", p.syntaxError("incomplete hex escape")
				}
				h1 := p.advance()
				if p.atEnd() {
					return "", p.syntaxError("incomplete hex escape")
				}
				h2 := p.advance()
				if !isHexDigit(h1) || !isHexDigit(h2) {
					return "", p.syntaxError("invalid hex escape sequence")
				}
				result = append(result, rune(hexVal(h1)<<4|hexVal(h2)))
			default:
				return "", p.syntaxErrorf("invalid escape sequence \\%c", escaped)
			}
		} else {
			result = append(result, ch)
			p.advance()
		}
	}
}

func (p *parser) parseRawStringLiteralValue() string {
	if p.peek() != '\'' {
		return ""
	}
	p.advance()
	start := p.pos
	for !p.atEnd() && p.peek() != '\'' {
		p.advance()
	}
	value := p.input[start:p.pos]
	if !p.atEnd() {
		p.advance()
	}
	return value
}

func (p *parser) parseSymbolName() (string, error) {
	ch := p.peek()
	switch {
	case ch == '"':
		return p.parseStringLiteralValue()
	case ch == '\'':
		return p.parseRawStringLiteralValue(), nil
	case isIdentifierStart(ch):
		start := p.pos
		p.scanIdentifier()
		return p.input[start:p.pos], nil
	default:
		return "", p.syntaxError("expected symbol")
	}
}

func (p *parser) matchStringAt(pos int, s string) bool {
	if pos+len(s) > len(p.input) {
		return false
	}
	return p.input[pos:pos+len(s)] == s
}

func (p *parser) unionNodes(lhs, rhs *Expression) *Expression {
	var nodes []*Expression
	if lhs.Kind == KindUnionAll {
		nodes = append(nodes, lhs.payload.(*UnionAllExpr).Nodes...)
	} else {
		nodes = append(nodes, lhs)
	}
	nodes = append(nodes, rhs)
	return &Expression{
		Kind:    KindUnionAll,
		Span:    Span{Start: lhs.Span.Start, End: rhs.Span.End},
		payload: &UnionAllExpr{Nodes: nodes},
	}
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\x0c'
}

func isIdentifierStart(r rune) bool {
	return isIdentifierPart(r)
}

func isIdentifierPart(r rune) bool {
	// identifier_part = (XID_CONTINUE | "_" | "*" | "/")+
	// XID_CONTINUE includes unicode letters, digits, combining marks, etc.
	// NOT included: ".", "-", "+" (those are connectors at identifier level)
	if r == '_' || r == '*' || r == '/' {
		return true
	}
	return unicode.Is(unicode.Letter, r) || unicode.IsDigit(r) || unicode.IsMark(r)
}

func isStrictIdentifierPart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '/'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func hexVal(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'a' && r <= 'f':
		return int(r - 'a' + 10)
	case r >= 'A' && r <= 'F':
		return int(r - 'A' + 10)
	}
	return 0
}

func splitIdentifierParts(ident string, seps string) []string {
	var parts []string
	current := ""
	for _, ch := range ident {
		if containsRune(seps, ch) {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	parts = append(parts, current)
	return parts
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}
