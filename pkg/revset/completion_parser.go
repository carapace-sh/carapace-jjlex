package revset

import (
	"unicode/utf8"
)

// ParseForCompletion parses a partial revset expression and returns a
// CompletionContext describing what is expected at the end of the input.
// Partial expressions are allowed - the parser recovers from errors to
// report what tokens would be valid at the cursor position.
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
	// before reaching the cursor. This distinguishes "expecting first expression"
	// from "after an expression, expecting operator".
	consumed bool

	// afterOperator is true when we have consumed an operator but haven't
	// started parsing the RHS. Used to avoid setting AttachedRevset in this
	// case.
	afterOperator bool

	// Stack of function parse states for nested calls
	funcStack []*funcParseState

	// innermostFunc is the deepest function context we've set
	innermostFunc *FunctionContext

	// lastExpr is the most recently parsed expression (set by parsePrimary)
	lastExpr *Expression

	// exprStart is the input position where the current expression started
	exprStart int

	// postfixStart is the input position where the current postfix chain
	// started (set at the beginning of parsePostfixOps). Used to compute
	// AttachedRevset.
	postfixStart int
}

type funcParseState struct {
	name        string
	args        []*Expression
	keywordArgs []KeywordArg
	argIndex    int
}

// atCursorOrEnd returns true if we're at or past the effective end (cursor or input end).
func (p *compParser) atCursorOrEnd() bool {
	return p.pos >= len(p.input) || p.pos >= p.cursor
}

// peek returns the next rune without advancing.
func (p *compParser) peek() rune {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	return r
}

// advance advances one rune.
func (p *compParser) advance() rune {
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return 0
	}
	r, w := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos += w
	p.consumed = true
	return r
}

// skipWS skips whitespace up to the cursor.
func (p *compParser) skipWS() {
	for p.pos < len(p.input) && p.pos < p.cursor {
		r, w := utf8.DecodeRuneInString(p.input[p.pos:])
		if !isWhitespace(r) {
			break
		}
		p.pos += w
	}
}

// matchString checks if the input at pos matches s (within cursor bounds).
func (p *compParser) matchString(s string) bool {
	end := p.pos + len(s)
	if end > len(p.input) || end > p.cursor {
		return false
	}
	return p.input[p.pos:end] == s
}

// addExpected adds an expected token type.
func (p *compParser) addExpected(t ExpectedToken) {
	p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, t)
}

// addOperator adds a valid operator.
func (p *compParser) addOperator(op, desc string) {
	p.ctx.ValidOperators = append(p.ctx.ValidOperators, ValidOperator{Op: op, Description: desc})
}

// afterExpression adds the tokens valid after a complete expression.
func (p *compParser) afterExpression() {
	p.addExpected(ExpectedOperator)
	p.addOperator("|", "union")
	p.addOperator("&", "intersection")
	p.addOperator("~", "difference")
	p.addOperator("::", "DAG range")
	p.addOperator("..", "range")
	p.addOperator("-", "parents")
	p.addOperator("+", "children")

	// Set AttachedRevset to the input from the start of the current postfix
	// chain to the current position. This tells the action layer what revset
	// the postfix operators are attached to (e.g. "@-" or "main--").
	// Don't set if we're right after an operator (before RHS started).
	if p.postfixStart < p.pos && p.ctx.AttachedRevset == "" && !p.afterOperator {
		p.ctx.AttachedRevset = p.input[p.postfixStart:p.pos]
	}

	// If inside a function, also expect ) and ,
	if len(p.funcStack) > 0 {
		p.addExpected(ExpectedClosingParen)
		p.addExpected(ExpectedComma)
	}
}

// beforeExpression adds the tokens valid when expecting an expression.
func (p *compParser) beforeExpression() {
	p.addExpected(ExpectedExpression)
}

// --- Parsing methods ---

// Operator precedence (same as main parser):
// 0: union (|)
// 1: intersection (&), difference (~)
// 2: infix dag range (::), range (..)
// 3: prefix negate (~)
// 4: prefix dag range (::), prefix range (..)
// 5: postfix dag range (::), postfix range (..)
// 6: postfix parents (-), children (+)

func (p *compParser) parseExpr() {
	p.parseInfixLevel0()
}

func (p *compParser) parseInfixLevel0() {
	p.parseInfixLevel1()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			// Don't offer operators right after an operator (expecting RHS)
			if p.lastExpr != nil && !p.afterOperator {
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
				// After | with no RHS - only expect expression
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
	p.skipWS()
	// At cursor with no left operand - call parseNegatePrefix to get prefix operators
	if p.atCursorOrEnd() && p.lastExpr == nil {
		p.parseNegatePrefix()
		return
	}
	p.parseNegatePrefix()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.lastExpr != nil {
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
				// After & with no RHS - only expect expression
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
				// After ~ with no RHS - could be difference or negate
				// Since ~ can be both prefix and infix, we handle it carefully
				if p.lastExpr != nil {
					p.afterExpression()
				}
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
		if p.lastExpr != nil {
			// After an expression, ~ is difference operator (handled by afterExpression)
			// But also ~ is a valid prefix operator here
			p.addExpected(ExpectedOperator)
			p.addOperator("~", "negate/difference")
		} else {
			p.beforeExpression()
			p.addExpected(ExpectedOperator)
			p.addOperator("~", "negate")
			p.addOperator("::", "DAG range prefix")
			p.addOperator("..", "range prefix")
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
	p.parseRangeExpr()
}

func (p *compParser) parseRangeExpr() {
	p.skipWS()
	if p.atCursorOrEnd() {
		if p.lastExpr != nil {
			p.afterExpression()
		} else {
			p.beforeExpression()
			p.addExpected(ExpectedOperator)
			p.addOperator("::", "DAG range")
			p.addOperator("..", "range")
		}
		return
	}

	// Check for nullary/prefix :: or ..
	if p.matchString("::") {
		saved := p.pos
		p.pos += 2
		p.skipWS()
		if p.atCursorOrEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' {
			// Nullary ::
			p.consumed = true
			p.lastExpr = &Expression{Kind: KindDagRangeAll, Span: Span{Start: saved, End: saved + 2}}
			return
		}
		// Prefix :: (no whitespace allowed after)
		if saved+2 < len(p.input) && isWhitespace(rune(p.input[saved+2])) {
			p.pos = saved
			return
		}
		p.parsePostfixOps()
		return
	}

	if p.matchString("..") {
		saved := p.pos
		p.pos += 2
		p.skipWS()
		if p.atCursorOrEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' {
			p.consumed = true
			p.lastExpr = &Expression{Kind: KindRangeAll, Span: Span{Start: saved, End: saved + 2}}
			return
		}
		if saved+2 < len(p.input) && isWhitespace(rune(p.input[saved+2])) {
			p.pos = saved
			return
		}
		p.parsePostfixOps()
		return
	}

	// Parse neighbors_expression
	p.parsePostfixOps()

	// Check for postfix or infix range operators
	p.skipWS()
	if p.atCursorOrEnd() {
		if p.lastExpr != nil {
			p.afterExpression()
			p.addOperator("::", "DAG range")
			p.addOperator("..", "range")
		}
		return
	}

	if p.matchString("::") {
		saved := p.pos
		p.pos += 2
		p.afterOperator = true
		p.skipWS()
		if p.atCursorOrEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' {
			// After :: at cursor - only expect expression (not operators)
			if p.lastExpr != nil {
				p.afterExpression()
			}
			p.beforeExpression()
			return
		}
		// Infix :: - parse RHS
		p.pos = saved
		p.parseInfixRangeOp()
		return
	}

	if p.matchString("..") {
		saved := p.pos
		p.pos += 2
		p.afterOperator = true
		p.skipWS()
		if p.atCursorOrEnd() || p.peek() == ')' || p.peek() == ',' || p.peek() == '|' || p.peek() == '&' || p.peek() == '~' {
			// After .. at cursor - only expect expression (not operators)
			if p.lastExpr != nil {
				p.afterExpression()
			}
			p.beforeExpression()
			return
		}
		p.pos = saved
		p.parseInfixRangeOp()
		return
	}
}

func (p *compParser) parseInfixRangeOp() {
	p.skipWS()
	if p.matchString("::") {
		p.pos += 2
		p.afterOperator = true
		p.skipWS()
		if p.atCursorOrEnd() {
			// After :: at cursor - only expect expression (not operators)
			if p.lastExpr != nil {
				p.afterExpression()
			}
			p.beforeExpression()
			return
		}
		p.afterOperator = false
		p.parsePostfixOps()
		return
	}
	if p.matchString("..") {
		p.pos += 2
		p.afterOperator = true
		p.skipWS()
		if p.atCursorOrEnd() {
			// After .. at cursor - only expect expression (not operators)
			if p.lastExpr != nil {
				p.afterExpression()
			}
			p.beforeExpression()
			return
		}
		p.afterOperator = false
		p.parsePostfixOps()
		return
	}
}

func (p *compParser) parsePostfixOps() {
	p.postfixStart = p.pos
	p.parsePrimary()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.lastExpr != nil {
				p.afterExpression()
			}
			return
		}
		ch := p.peek()
		if ch == '-' {
			p.advance()
		} else if ch == '+' {
			p.advance()
		} else {
			return
		}
	}
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
	case ch == '@':
		p.advance()
		p.lastExpr = &Expression{Kind: KindAtCurrentWorkspace, Span: Span{Start: start, End: p.pos}}
	case isIdentifierStart(ch):
		p.parseSymbolOrFunctionCompletion()
	default:
		p.beforeExpression()
	}
}

func (p *compParser) parseParenthesized() {
	p.advance() // consume (
	p.skipWS()
	if p.atCursorOrEnd() {
		p.beforeExpression()
		p.addExpected(ExpectedClosingParen)
		return
	}
	p.parseExpr()
	p.skipWS()
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedClosingParen)
		return
	}
	if p.peek() == ')' {
		p.advance()
	}
}

func (p *compParser) parseStringLiteralCompletion() string {
	p.advance() // consume opening "
	var content []rune
	for {
		if p.atCursorOrEnd() {
			p.ctx.PartialString = string(content)
			p.ctx.StringQuote = '"'
			p.addExpected(ExpectedStringClose)
			return string(content)
		}
		ch := p.peek()
		if ch == '"' {
			p.advance()
			return string(content) // complete string
		}
		if ch == '\\' {
			p.advance()
			if p.atCursorOrEnd() {
				p.ctx.PartialString = string(content)
				p.ctx.StringQuote = '"'
				p.addExpected(ExpectedStringClose)
				return string(content)
			}
			p.advance() // consume escaped char
		} else {
			content = append(content, ch)
			p.advance()
		}
	}
}

func (p *compParser) parseRawStringLiteralCompletion() string {
	p.advance() // consume opening '
	begin := p.pos
	for {
		if p.atCursorOrEnd() {
			p.ctx.PartialString = p.input[begin:p.pos]
			p.ctx.StringQuote = '\''
			p.addExpected(ExpectedStringClose)
			return p.input[begin:p.pos]
		}
		if p.peek() == '\'' {
			value := p.input[begin:p.pos]
			p.advance()
			return value
		}
		p.advance()
	}
}

func (p *compParser) parseSymbolOrFunctionCompletion() {
	identStart := p.pos
	p.scanIdentifierCompletion()
	identEnd := p.pos

	// Check if cursor is within or right after the identifier
	if p.pos >= p.cursor && identStart < p.cursor {
		p.ctx.PartialIdent = p.input[identStart:p.cursor]
		p.lastExpr = &Expression{Kind: KindIdentifier, Span: Span{Start: identStart, End: p.cursor}, payload: &IdentifierExpr{Name: p.ctx.PartialIdent}}
		p.beforeExpression()
		return
	}

	ident := p.input[identStart:identEnd]

	p.skipWS()
	if p.atCursorOrEnd() {
		// After a complete identifier, could continue with operators or become function call
		p.lastExpr = &Expression{Kind: KindIdentifier, Span: Span{Start: identStart, End: identEnd}, payload: &IdentifierExpr{Name: ident}}
		p.afterExpression()
		return
	}

	// Check for function call
	if p.peek() == '(' && isFunctionNameCheck(ident) {
		p.parseFunctionCallCompletion(ident)
		return
	}

	// Check for pattern: strict_identifier : value
	if identEnd < len(p.input) && p.input[identEnd] == ':' && !p.matchStringAtFullInput(identEnd, "::") && isStrictIdentifierCheck(ident) {
		p.pos = identEnd + 1 // consume :
		if p.atCursorOrEnd() || (p.pos < len(p.input) && isWhitespace(rune(p.input[p.pos]))) {
			if p.atCursorOrEnd() {
				p.ctx.InPattern = true
				p.ctx.PatternName = ident
				p.addExpected(ExpectedPatternValue)
				p.beforeExpression()
				p.lastExpr = &Expression{Kind: KindPattern, Span: Span{Start: identStart, End: p.pos}, payload: &PatternExpr{Name: ident}}
				return
			}
			p.pos = identEnd
		} else {
			p.ctx.InPattern = true
			p.ctx.PatternName = ident
			p.parsePatternValueCompletion()
			p.lastExpr = &Expression{Kind: KindPattern, Span: Span{Start: identStart, End: p.pos}, payload: &PatternExpr{Name: ident, Value: p.lastExpr}}
			return
		}
	}

	// Check for @ suffix
	p.skipWS()
	if !p.atCursorOrEnd() && p.peek() == '@' {
		p.advance() // consume @
		p.skipWS()
		if p.atCursorOrEnd() {
			// After @ at cursor: could be completing a remote name
			// or the expression is complete as AtWorkspace
			p.ctx.InRemoteSymbol = true
			p.ctx.RemoteBookmarkName = ident
			p.beforeExpression()
			p.lastExpr = &Expression{Kind: KindAtWorkspace, Span: Span{Start: identStart, End: p.pos}, payload: &AtWorkspaceExpr{Name: ident}}
			return
		}
		// Parse remote part
		ch := p.peek()
		var remote string
		if ch == '"' {
			remote = p.parseStringLiteralCompletion()
			if p.atCursorOrEnd() {
				p.ctx.InRemoteSymbol = true
				p.ctx.RemoteBookmarkName = ident
			}
		} else if ch == '\'' {
			remote = p.parseRawStringLiteralCompletion()
			if p.atCursorOrEnd() {
				p.ctx.InRemoteSymbol = true
				p.ctx.RemoteBookmarkName = ident
			}
		} else if isIdentifierStart(ch) {
			remoteStart := p.pos
			p.scanIdentifierCompletion()
			if p.pos >= p.cursor && remoteStart < p.cursor {
				p.ctx.PartialRemote = p.input[remoteStart:p.cursor]
				remote = p.ctx.PartialRemote
				// Clear PartialIdent since this is a remote name, not a general identifier
				p.ctx.PartialIdent = ""
				p.ctx.InRemoteSymbol = true
				p.ctx.RemoteBookmarkName = ident
			} else {
				remote = p.input[remoteStart:p.pos]
			}
		}
		p.lastExpr = &Expression{Kind: KindRemoteSymbol, Span: Span{Start: identStart, End: p.pos}, payload: &RemoteSymbolExpr{Name: ident, Remote: remote}}
		return
	}

	p.lastExpr = &Expression{Kind: KindIdentifier, Span: Span{Start: identStart, End: identEnd}, payload: &IdentifierExpr{Name: ident}}
}

func (p *compParser) parsePatternValueCompletion() {
	p.skipWS()
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedPatternValue)
		p.beforeExpression()
		return
	}
	p.parsePostfixOps()
}

func (p *compParser) parseFunctionCallCompletion(name string) {
	fs := &funcParseState{name: name}
	funcStart := p.exprStart
	p.funcStack = append(p.funcStack, fs)
	defer func() {
		p.funcStack = p.funcStack[:len(p.funcStack)-1]
	}()

	p.advance() // consume (
	p.skipWS()
	if p.atCursorOrEnd() {
		p.setFunctionContext(fs, 0)
		p.beforeExpression()
		p.addExpected(ExpectedClosingParen)
		p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name}}
		return
	}

	argIndex := 0
	for !p.atCursorOrEnd() && p.peek() != ')' {
		p.skipWS()
		if p.atCursorOrEnd() {
			p.setFunctionContext(fs, argIndex)
			p.beforeExpression()
			p.addExpected(ExpectedClosingParen)
			p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
			return
		}

		// Try keyword arg detection
		kwName, kwSaved := p.tryKeywordArgLookahead()
		if kwName != "" {
			// We have identifier = pattern - scan the keyword name
			kwIdentStart := p.pos
			p.scanIdentifierCompletion() // consume the keyword name
			p.skipWS()
			if p.atCursorOrEnd() {
				p.setFunctionContext(fs, argIndex)
				p.ctx.Function.IsKeywordArg = true
				p.ctx.Function.KeywordArgName = p.input[kwIdentStart:min(p.pos, p.cursor)]
				p.addExpected(ExpectedEquals)
				p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
				return
			}
			if p.peek() == '=' {
				p.advance()
				p.skipWS()
				if p.atCursorOrEnd() {
					p.setFunctionContext(fs, argIndex)
					p.beforeExpression()
					p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
					return
				}
				p.parseExpr()
				fs.keywordArgs = append(fs.keywordArgs, KeywordArg{Name: kwName, Value: p.lastExpr})
				p.skipWS()
				if p.atCursorOrEnd() {
					p.setFunctionContext(fs, argIndex+1)
					p.addExpected(ExpectedClosingParen)
					p.addExpected(ExpectedComma)
					p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
					return
				}
				if p.peek() == ',' {
					p.advance()
					argIndex++
				}
				continue
			}
			// Not actually a keyword arg, restore
			p.pos = kwSaved
		}

		// Regular positional argument
		p.parseExpr()
		fs.args = append(fs.args, p.lastExpr)
		argIndex++

		// If the parsed expression ended with a partial identifier, it might
		// be a keyword arg name (e.g., "remote" in "remote_bookmarks(remote")
		if p.atCursorOrEnd() && p.ctx.PartialIdent != "" {
			// The partial identifier could be a keyword arg name
			p.setFunctionContext(fs, argIndex)
			if p.ctx.Function != nil {
				p.ctx.Function.KeywordArgName = p.ctx.PartialIdent
			}
			p.addExpected(ExpectedClosingParen)
			p.addExpected(ExpectedComma)
			p.addExpected(ExpectedEquals)
			p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
			return
		}

		p.skipWS()
		if p.atCursorOrEnd() {
			p.setFunctionContext(fs, argIndex)
			p.addExpected(ExpectedClosingParen)
			p.addExpected(ExpectedComma)
			p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
			return
		}
		if p.peek() == ',' {
			p.advance()
			p.skipWS()
			if p.atCursorOrEnd() {
				p.setFunctionContext(fs, argIndex)
				p.addExpected(ExpectedClosingParen)
				p.beforeExpression()
				p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
				return
			}
		} else {
			break
		}
	}

	p.skipWS()
	if p.atCursorOrEnd() {
		p.setFunctionContext(fs, argIndex)
		p.addExpected(ExpectedClosingParen)
		p.beforeExpression() // could add more args
		p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
		return
	}
	if p.peek() == ')' {
		p.advance()
	}
	p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args, KeywordArgs: fs.keywordArgs}}
}

// setFunctionContext populates the CompletionContext's Function field.
// Only sets the first time (innermost/deepest function wins).
func (p *compParser) setFunctionContext(fs *funcParseState, argIndex int) {
	if p.innermostFunc != nil {
		return // already set by a deeper function call
	}
	ctx := &FunctionContext{
		Name:        fs.name,
		KeywordArgs: fs.keywordArgs,
		ArgIndex:    argIndex,
	}
	// Copy args
	ctx.Args = make([]*Expression, len(fs.args))
	copy(ctx.Args, fs.args)
	p.innermostFunc = ctx
	p.ctx.Function = ctx
}

// tryKeywordArgLookahead checks if the current position looks like a keyword arg.
// Returns the keyword name and the position to restore to.
func (p *compParser) tryKeywordArgLookahead() (string, int) {
	saved := p.pos
	start := p.pos
	if p.pos >= len(p.input) || p.pos >= p.cursor {
		return "", saved
	}
	if !isStrictIdentifierPart(p.peek()) {
		return "", saved
	}
	for p.pos < len(p.input) && p.pos < p.cursor && isStrictIdentifierPart(p.peek()) {
		p.advance()
	}
	name := p.input[start:p.pos]
	// Skip whitespace (looking beyond cursor for lookahead)
	pos := p.pos
	for pos < len(p.input) && isWhitespace(rune(p.input[pos])) {
		pos++
	}
	// Check for =
	if pos < len(p.input) && p.input[pos] == '=' {
		p.pos = saved
		return name, saved
	}
	p.pos = saved
	return "", saved
}

// scanIdentifierCompletion scans a revset identifier, stopping at cursor.
func (p *compParser) scanIdentifierCompletion() {
	if p.atCursorOrEnd() {
		return
	}
	if !p.scanIdentPartCompletion() {
		return
	}
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if ch == '.' {
			saved := p.pos
			p.advance()
			if p.atCursorOrEnd() || !p.scanIdentPartCompletion() {
				p.pos = saved
				break
			}
		} else if ch == '-' {
			saved := p.pos
			for !p.atCursorOrEnd() && p.peek() == '-' {
				p.advance()
			}
			if p.pos > saved && !p.atCursorOrEnd() && p.scanIdentPartCompletion() {
				continue
			}
			p.pos = saved
			break
		} else if ch == '+' {
			saved := p.pos
			p.advance()
			if p.atCursorOrEnd() || !p.scanIdentPartCompletion() {
				p.pos = saved
				break
			}
		} else {
			break
		}
	}
}

func (p *compParser) scanIdentPartCompletion() bool {
	start := p.pos
	for !p.atCursorOrEnd() && isIdentifierPart(p.peek()) {
		p.advance()
	}
	return p.pos > start
}

// matchStringAtFullInput checks match at a position in the full input.
func (p *compParser) matchStringAtFullInput(pos int, s string) bool {
	end := pos + len(s)
	if end > len(p.input) {
		return false
	}
	return p.input[pos:end] == s
}

// Helper checks that don't need a parser.
func isFunctionNameCheck(ident string) bool {
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

func isStrictIdentifierCheck(ident string) bool {
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

// dedupTokens removes duplicate expected tokens.
func dedupTokens(tokens []ExpectedToken) []ExpectedToken {
	seen := make(map[ExpectedToken]bool)
	result := make([]ExpectedToken, 0, len(tokens))
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	return result
}

// dedupOperators removes duplicate operators.
func dedupOperators(ops []ValidOperator) []ValidOperator {
	seen := make(map[string]bool)
	result := make([]ValidOperator, 0, len(ops))
	for _, op := range ops {
		if !seen[op.Op] {
			seen[op.Op] = true
			result = append(result, op)
		}
	}
	return result
}
