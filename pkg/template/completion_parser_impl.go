package template

import "unicode/utf8"

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

func (p *compParser) matchString(s string) bool {
	end := p.pos + len(s)
	if end > len(p.input) || end > p.cursor {
		return false
	}
	return p.input[p.pos:end] == s
}

func (p *compParser) addExpected(t ExpectedToken) {
	p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, t)
}

func (p *compParser) addOperator(op, desc string) {
	p.ctx.ValidOperators = append(p.ctx.ValidOperators, ValidOperator{Op: op, Description: desc})
}

func (p *compParser) afterExpression() {
	p.addExpected(ExpectedOperator)
	p.addOperator("||", "logical or")
	p.addOperator("&&", "logical and")
	p.addOperator("==", "equal")
	p.addOperator("!=", "not equal")
	p.addOperator(">=", "greater or equal")
	p.addOperator(">", "greater")
	p.addOperator("<=", "less or equal")
	p.addOperator("<", "less")
	p.addOperator("+", "add")
	p.addOperator("-", "subtract")
	p.addOperator("*", "multiply")
	p.addOperator("/", "divide")
	p.addOperator("%", "remainder")
	p.addOperator("++", "concatenate")
	if len(p.funcStack) > 0 {
		p.addExpected(ExpectedClosingParen)
		p.addExpected(ExpectedComma)
	}
}

func (p *compParser) beforeExpression() {
	p.addExpected(ExpectedExpression)
}

func (p *compParser) setFunctionContext(name string, isMethod bool, methodObj *Expression) {
	fctx := &FunctionContext{
		Name:         name,
		IsMethod:     isMethod,
		MethodObject: methodObj,
	}
	p.ctx.Function = fctx
	p.innermostFunc = fctx
}

func (p *compParser) updateFunctionArgIndex() {
	if len(p.funcStack) > 0 {
		fs := p.funcStack[len(p.funcStack)-1]
		if p.ctx.Function != nil {
			p.ctx.Function.ArgIndex = fs.argIndex
			p.ctx.Function.Args = fs.args
			p.ctx.Function.KeywordArgs = fs.keywordArgs
			p.ctx.Function.IsMethod = fs.isMethod
			p.ctx.Function.MethodObject = fs.methodObj
		}
	}
}

func (p *compParser) parseTemplateComp() {
	p.parseExpressionComp()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.consumed {
				p.addOperator("++", "concatenate")
			}
			return
		}
		if p.matchString("++") {
			p.pos += 2
			p.skipWS()
			if p.atCursorOrEnd() {
				p.afterExpression()
				p.beforeExpression()
				return
			}
			p.parseExpressionComp()
		} else {
			return
		}
	}
}

func (p *compParser) parseExpressionComp() {
	p.parsePrattComp(precLogicalOr)
}

func (p *compParser) parsePrattComp(minPrec int) {
	p.parsePrefixComp()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.consumed {
				p.afterExpression()
			}
			return
		}
		op, prec, _ := peekInfixOp(p.input, p.pos)
		if op == "" || prec < minPrec {
			return
		}
		p.pos += len(op)
		p.skipWS()
		if p.atCursorOrEnd() {
			p.afterExpression()
			p.beforeExpression()
			return
		}
		p.parsePrattComp(prec + 1)
	}
}

func (p *compParser) parsePrefixComp() {
	p.skipWS()
	if p.atCursorOrEnd() {
		if !p.consumed {
			p.beforeExpression()
			p.addOperator("!", "logical not")
			p.addOperator("-", "negate")
		} else {
			p.afterExpression()
		}
		return
	}
	if p.peek() == '!' {
		p.advance()
		p.skipWS()
		if p.atCursorOrEnd() {
			p.beforeExpression()
			return
		}
		p.parsePrefixComp()
		return
	}
	if p.peek() == '-' {
		// Check if it's a prefix negate (not an infix minus)
		// If we haven't consumed anything, it's prefix
		if !p.consumed {
			p.advance()
			p.skipWS()
			if p.atCursorOrEnd() {
				p.beforeExpression()
				return
			}
			p.parsePrefixComp()
			return
		}
	}
	p.parseTermComp()
}

func (p *compParser) parseTermComp() {
	p.parsePrimaryComp()
	for {
		p.skipWS()
		if p.atCursorOrEnd() {
			if p.consumed {
				p.afterExpression()
			}
			return
		}
		if p.peek() == '.' {
			saved := p.pos
			p.advance()
			p.skipWS()
			if p.atCursorOrEnd() || !isIdentifierStart(p.peek()) {
				p.pos = saved
				return
			}
			identStart := p.pos
			p.scanIdentifierCompletion()
			ident := p.input[identStart:p.pos]
			p.skipWS()
			if p.atCursorOrEnd() {
				// Completing method name
				p.ctx.PartialIdent = ident
				p.beforeExpression()
				return
			}
			if p.peek() == '(' && isFunctionName(ident) {
				p.parseFunctionCallComp(ident, true, nil)
			} else {
				p.pos = saved
				return
			}
		} else {
			return
		}
	}
}

func (p *compParser) parsePrimaryComp() {
	p.skipWS()
	if p.atCursorOrEnd() {
		if !p.consumed {
			p.beforeExpression()
		}
		return
	}

	ch := p.peek()
	switch {
	case ch == '(':
		p.parseParenthesizedComp()
	case ch == '"':
		p.parseStringLiteralComp('"')
	case ch == '\'':
		p.parseStringLiteralComp('\'')
	case ch == '|':
		p.parseLambdaComp()
	case ch >= '0' && ch <= '9':
		for !p.atCursorOrEnd() && p.peek() >= '0' && p.peek() <= '9' {
			p.advance()
		}
		p.consumed = true
	case isIdentifierStart(ch):
		p.parseIdentFuncOrPatternComp()
	default:
		p.beforeExpression()
	}
}

func (p *compParser) parseParenthesizedComp() {
	p.advance() // consume (
	p.skipWS()
	if p.atCursorOrEnd() {
		p.beforeExpression()
		p.addExpected(ExpectedClosingParen)
		return
	}
	p.parseTemplateComp()
	p.skipWS()
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedClosingParen)
		return
	}
	if p.peek() == ')' {
		p.advance()
	}
}

func (p *compParser) parseStringLiteralComp(quote rune) {
	p.advance() // consume opening quote
	contentStart := p.pos
	for {
		if p.atCursorOrEnd() {
			p.ctx.PartialString = p.input[contentStart:p.pos]
			p.ctx.StringQuote = quote
			p.addExpected(ExpectedStringClose)
			return
		}
		ch := p.peek()
		if ch == quote {
			p.advance()
			p.consumed = true
			return
		}
		if quote == '"' && ch == '\\' {
			p.advance()
			if p.atCursorOrEnd() {
				p.ctx.PartialString = p.input[contentStart:p.pos]
				p.ctx.StringQuote = quote
				p.addExpected(ExpectedStringClose)
				return
			}
			p.advance()
		} else {
			p.advance()
		}
	}
}

func (p *compParser) parseLambdaComp() {
	p.advance() // consume first |
	p.skipWS()

	var params []string
	if p.atCursorOrEnd() || p.peek() != '|' {
		// Parse parameters
		for {
			p.skipWS()
			if p.atCursorOrEnd() {
				p.addExpected(ExpectedLambdaClose)
				p.beforeExpression()
				p.ctx.InLambda = true
				p.ctx.LambdaParams = params
				return
			}
			if p.peek() == '|' {
				break
			}
			identStart := p.pos
			p.scanIdentifierCompletion()
			if p.pos > identStart {
				param := p.input[identStart:p.pos]
				params = append(params, param)
			}
			p.skipWS()
			if p.atCursorOrEnd() {
				p.addExpected(ExpectedComma)
				p.addExpected(ExpectedLambdaClose)
				p.ctx.InLambda = true
				p.ctx.LambdaParams = params
				return
			}
			if p.peek() == ',' {
				p.advance()
				p.skipWS()
				if p.peek() == '|' {
					break
				}
				continue
			}
			break
		}
	}

	p.skipWS()
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedLambdaClose)
		p.ctx.InLambda = true
		p.ctx.LambdaParams = params
		return
	}
	if p.peek() != '|' {
		p.beforeExpression()
		p.ctx.InLambda = true
		p.ctx.LambdaParams = params
		return
	}
	p.advance() // consume closing |

	p.skipWS()
	if p.atCursorOrEnd() {
		p.beforeExpression()
		p.ctx.InLambda = true
		p.ctx.LambdaParams = params
		return
	}

	p.ctx.InLambda = true
	p.ctx.LambdaParams = params
	p.parseTemplateComp()
	p.ctx.InLambda = false
	p.ctx.LambdaParams = nil
}

func (p *compParser) parseIdentFuncOrPatternComp() {
	identStart := p.pos
	p.scanIdentifierCompletion()
	baseIdentEnd := p.pos
	// Extend for pattern identifier suffix (dashes like "regex-i")
	// Save position so we can backtrack if pattern doesn't match
	p.scanPatternIdentifierSuffixComp()
	identEnd := p.pos

	// Check if cursor is within or right after the identifier
	if p.pos >= p.cursor && identStart < p.cursor {
		p.ctx.PartialIdent = p.input[identStart:p.cursor]
		p.consumed = true
		p.beforeExpression()
		return
	}

	ident := p.input[identStart:identEnd]
	p.consumed = true

	// Boolean literals
	if ident == "true" || ident == "false" {
		p.skipWS()
		if p.atCursorOrEnd() {
			p.afterExpression()
			return
		}
		return
	}

	p.skipWS()
	if p.atCursorOrEnd() {
		p.ctx.PartialIdent = ident
		p.afterExpression()
		return
	}

	// Function call (function names cannot contain dashes)
	baseIdent := p.input[identStart:baseIdentEnd]
	if p.peek() == '(' && isFunctionName(baseIdent) {
		p.pos = baseIdentEnd
		p.parseFunctionCallComp(baseIdent, false, nil)
		return
	}

	// Pattern: identifier:value
	if identEnd < len(p.input) && p.input[identEnd] == ':' {
		isDoubleColon := identEnd+2 <= len(p.input) && p.input[identEnd:identEnd+2] == "::"
		if !isDoubleColon && isPatternIdentifier(ident) {
			p.pos = identEnd + 1
			p.ctx.InPattern = true
			p.ctx.PatternName = ident
			if p.atCursorOrEnd() || (p.pos < len(p.input) && isWhitespace(rune(p.input[p.pos]))) {
				if p.atCursorOrEnd() {
					p.addExpected(ExpectedPatternValue)
					p.beforeExpression()
					return
				}
			} else {
				p.parsePrefixComp()
				return
			}
		}
	}

	// Not a pattern — backtrack to base identifier (without dash suffix)
	// so that "x-y" becomes identifier "x" with "-y" parsed as infix subtraction
	p.pos = baseIdentEnd
}

func (p *compParser) scanIdentifierCompletion() {
	for !p.atCursorOrEnd() && isIdentifierPart(p.peek()) {
		p.advance()
	}
}

func (p *compParser) scanPatternIdentifierSuffixComp() {
	for !p.atCursorOrEnd() && p.peek() == '-' {
		saved := p.pos
		p.advance() // consume -
		if p.atCursorOrEnd() || !isIdentifierPart(p.peek()) {
			p.pos = saved
			return
		}
		for !p.atCursorOrEnd() && isIdentifierPart(p.peek()) {
			p.advance()
		}
	}
}

func (p *compParser) parseFunctionCallComp(name string, isMethod bool, methodObj *Expression) {
	p.setFunctionContext(name, isMethod, methodObj)
	fs := &funcParseState{name: name, isMethod: isMethod, methodObj: methodObj}
	p.funcStack = append(p.funcStack, fs)

	p.advance() // consume (
	p.skipWS()
	if p.atCursorOrEnd() {
		p.beforeExpression()
		p.addExpected(ExpectedClosingParen)
		p.updateFunctionArgIndex()
		return
	}

	for !p.atCursorOrEnd() && p.peek() != ')' {
		p.skipWS()
		if p.atCursorOrEnd() {
			break
		}

		// Check for keyword argument
		saved := p.pos
		if isIdentifierStart(p.peek()) {
			identStart := p.pos
			p.scanIdentifierCompletion()
			kwName := p.input[identStart:p.pos]
			p.skipWS()
			if !p.atCursorOrEnd() && p.peek() == '=' {
				p.advance() // consume =
				p.skipWS()
				fs.keywordArgs = append(fs.keywordArgs, KeywordArg{Name: kwName})
				if p.atCursorOrEnd() {
					p.beforeExpression()
					if p.ctx.Function != nil {
						p.ctx.Function.KeywordArgName = kwName
					}
					p.addExpected(ExpectedClosingParen)
					p.updateFunctionArgIndex()
					return
				}
				p.parseTemplateComp()
				p.skipWS()
				if p.peek() == ',' {
					p.advance()
					p.skipWS()
				}
				continue
			}
			p.pos = saved
		}

		fs.argIndex = len(fs.args)
		p.updateFunctionArgIndex()

		p.parseTemplateComp()
		fs.args = append(fs.args, p.lastExpr)
		fs.argIndex = len(fs.args)
		p.updateFunctionArgIndex()

		p.skipWS()
		if p.atCursorOrEnd() {
			p.addExpected(ExpectedClosingParen)
			p.addExpected(ExpectedComma)
			return
		}
		if p.peek() == ',' {
			p.advance()
			p.skipWS()
			if p.atCursorOrEnd() {
				p.beforeExpression()
				p.addExpected(ExpectedClosingParen)
				p.updateFunctionArgIndex()
				return
			}
		} else {
			break
		}
	}

	p.skipWS()
	if p.atCursorOrEnd() {
		p.addExpected(ExpectedClosingParen)
		p.updateFunctionArgIndex()
		return
	}
	if p.peek() == ')' {
		p.advance()
	}

	p.funcStack = p.funcStack[:len(p.funcStack)-1]
}