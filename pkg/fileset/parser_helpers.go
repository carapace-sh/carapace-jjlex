package fileset

func (p *parser) parseSymbolOrFunctionOrPattern(start int) (*Expression, error) {
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

	// Check for pattern: strict_identifier : primary (no whitespace around :)
	if identEnd < len(p.input) && p.input[identEnd] == ':' && p.isStrictIdentifier(ident) {
		p.pos = identEnd + 1 // consume :
		if p.atEnd() || isWhitespace(p.peek()) {
			p.pos = identEnd
		} else {
			value, err := p.parsePrimary()
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

	// Plain identifier
	return &Expression{
		Kind:    KindIdentifier,
		Span:    Span{Start: start, End: p.pos},
		payload: &IdentifierExpr{Name: ident},
	}, nil
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

func (p *parser) parseFunctionCall(name string, start int) (*Expression, error) {
	p.advance() // consume (
	p.skipWhitespace()

	var args []*Expression

	for !p.atEnd() && p.peek() != ')' {
		p.skipWhitespace()
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		p.skipWhitespace()
		if p.peek() == ',' {
			p.advance()
			p.skipWhitespace()
			if p.peek() == ')' {
				break
			}
		} else {
			break
		}
	}

	p.skipWhitespace()
	if p.peek() != ')' {
		return nil, p.syntaxError("expected ')' in function call")
	}
	p.advance()

	return &Expression{
		Kind: KindFunctionCall,
		Span: Span{Start: start, End: p.pos},
		payload: &FunctionCallExpr{
			Name: name,
			Args: args,
		},
	}, nil
}

func (p *parser) tryBareStringPattern() *Expression {
	p.skipWhitespace()
	start := p.pos
	if !p.scanStrictIdentifier() {
		return nil
	}
	identEnd := p.pos
	ident := p.input[start:identEnd]

	if identEnd >= len(p.input) || p.input[identEnd] != ':' {
		return nil
	}
	p.pos = identEnd + 1

	bareStart := p.pos
	if !p.scanBareString() {
		return nil
	}
	bareValue := p.input[bareStart:p.pos]

	return &Expression{
		Kind:    KindBareStringPattern,
		Span:    Span{Start: start, End: p.pos},
		payload: &PatternExpr{
			Name: ident,
			Value: &Expression{
				Kind:    KindBareString,
				Span:    Span{Start: bareStart, End: p.pos},
				payload: &BareStringExpr{Value: bareValue},
			},
		},
	}
}

func (p *parser) tryBareString() *Expression {
	p.skipWhitespace()
	start := p.pos
	if !p.scanBareString() {
		return nil
	}
	if p.pos == start {
		return nil
	}
	value := p.input[start:p.pos]
	return &Expression{
		Kind:    KindBareString,
		Span:    Span{Start: start, End: p.pos},
		payload: &BareStringExpr{Value: value},
	}
}