package fileset

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
			return string(content)
		}
		if ch == '\\' {
			p.advance()
			if p.atCursorOrEnd() {
				p.ctx.PartialString = string(content)
				p.ctx.StringQuote = '"'
				p.addExpected(ExpectedStringClose)
				return string(content)
			}
			p.advance()
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

	if p.pos >= p.cursor && identStart < p.cursor {
		p.ctx.PartialIdent = p.input[identStart:p.cursor]
		p.lastExpr = &Expression{Kind: KindIdentifier, Span: Span{Start: identStart, End: p.cursor}, payload: &IdentifierExpr{Name: p.ctx.PartialIdent}}
		p.beforeExpression()
		return
	}

	ident := p.input[identStart:identEnd]

	p.skipWS()
	if p.atCursorOrEnd() {
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
			p.parsePrimary()
			p.lastExpr = &Expression{Kind: KindPattern, Span: Span{Start: identStart, End: p.pos}, payload: &PatternExpr{Name: ident, Value: p.lastExpr}}
			return
		}
	}

	p.lastExpr = &Expression{Kind: KindIdentifier, Span: Span{Start: identStart, End: identEnd}, payload: &IdentifierExpr{Name: ident}}
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
			p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args}}
			return
		}

		// Parse argument
		p.parseExpr()
		fs.args = append(fs.args, p.lastExpr)
		argIndex++

		p.skipWS()
		if p.atCursorOrEnd() {
			p.setFunctionContext(fs, argIndex)
			p.addExpected(ExpectedClosingParen)
			p.addExpected(ExpectedComma)
			p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args}}
			return
		}
		if p.peek() == ',' {
			p.advance()
			p.skipWS()
			if p.atCursorOrEnd() {
				p.setFunctionContext(fs, argIndex)
				p.addExpected(ExpectedClosingParen)
				p.beforeExpression()
				p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args}}
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
		p.beforeExpression()
		p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args}}
		return
	}
	if p.peek() == ')' {
		p.advance()
	}
	p.lastExpr = &Expression{Kind: KindFunctionCall, Span: Span{Start: funcStart, End: p.pos}, payload: &FunctionCallExpr{Name: name, Args: fs.args}}
}

func (p *compParser) setFunctionContext(fs *funcParseState, argIndex int) {
	if p.innermostFunc != nil {
		return
	}
	ctx := &FunctionContext{
		Name:     fs.name,
		ArgIndex: argIndex,
	}
	ctx.Args = make([]*Expression, len(fs.args))
	copy(ctx.Args, fs.args)
	p.innermostFunc = ctx
	p.ctx.Function = ctx
}

func (p *compParser) scanIdentifierCompletion() {
	if p.atCursorOrEnd() {
		return
	}
	if !p.scanIdentPartCompletion() {
		return
	}
	for !p.atCursorOrEnd() {
		ch := p.peek()
		if isFilesetIdentifierPart(ch) {
			p.advance()
		} else {
			break
		}
	}
}

func (p *compParser) scanIdentPartCompletion() bool {
	start := p.pos
	for !p.atCursorOrEnd() && isFilesetIdentifierPart(p.peek()) {
		p.advance()
	}
	return p.pos > start
}

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
	parts := splitIdentifierParts(ident, "-")
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, ch := range part {
			if !isStrictIdentifierPart(ch) {
				return false
			}
		}
	}
	return true
}

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
