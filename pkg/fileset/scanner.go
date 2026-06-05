package fileset

func (p *parser) scanIdentifier() bool {
	if p.atEnd() {
		return false
	}
	start := p.pos
	for !p.atEnd() && isFilesetIdentifierPart(p.peek()) {
		p.advance()
	}
	return p.pos > start
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
		if ch == '-' {
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

func (p *parser) scanBareString() bool {
	if p.atEnd() {
		return false
	}
	start := p.pos
	for !p.atEnd() && isBareStringPart(p.peek()) {
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
