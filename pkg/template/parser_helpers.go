package template

import (
	"strconv"
)

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\x0c'
}

func isIdentifierStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || (r >= '0' && r <= '9')
}

func isFunctionName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			if !isIdentifierStart(ch) {
				return false
			}
		} else {
			if !isIdentifierPart(ch) {
				return false
			}
		}
	}
	// Boolean literals cannot be used as function names
	if s == "true" || s == "false" {
		return false
	}
	return true
}

func isPatternIdentifierStart(r rune) bool {
	return isIdentifierStart(r)
}

func isPatternIdentifierPart(r rune) bool {
	return isIdentifierPart(r) || r == '-'
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func hexVal(r rune) rune {
	if r >= '0' && r <= '9' {
		return r - '0'
	}
	if r >= 'a' && r <= 'f' {
		return r - 'a' + 10
	}
	if r >= 'A' && r <= 'F' {
		return r - 'A' + 10
	}
	return 0
}

func (p *parser) scanIdentifier() bool {
	if p.atEnd() {
		return false
	}
	start := p.pos
	if !isIdentifierStart(p.peek()) {
		return false
	}
	for !p.atEnd() && isIdentifierPart(p.peek()) {
		p.advance()
	}
	return p.pos > start
}

// scanPatternIdentifierSuffix extends the current identifier position
// to include dash-separated parts (e.g., "regex-i" for pattern names).
// This should be called after scanIdentifier() when the context might be a pattern.
func (p *parser) scanPatternIdentifierSuffix() {
	for !p.atEnd() && p.peek() == '-' {
		saved := p.pos
		p.advance() // consume -
		if p.atEnd() || !isIdentifierPart(p.peek()) {
			p.pos = saved
			return
		}
		for !p.atEnd() && isIdentifierPart(p.peek()) {
			p.advance()
		}
	}
}

func (p *parser) scanPatternIdentifier() bool {
	if p.atEnd() {
		return false
	}
	start := p.pos
	if !isPatternIdentifierStart(p.peek()) {
		return false
	}
	p.advance()
	for !p.atEnd() && isPatternIdentifierPart(p.peek()) {
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
		p.advance() // consume closing '
	}
	return value
}

func (p *parser) parseIntegerLiteral(start int) (*Expression, error) {
	begin := p.pos
	// Reject leading zeros (except plain "0")
	if p.peek() == '0' {
		p.advance()
		// Check if next char is also a digit (leading zero)
		if !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
			return nil, p.syntaxError("integer literal cannot have leading zeros")
		}
	} else {
		for !p.atEnd() && p.peek() >= '0' && p.peek() <= '9' {
			p.advance()
		}
	}
	text := p.input[begin:p.pos]
	val, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil, p.syntaxError("invalid integer literal")
	}
	return &Expression{
		Kind:    KindInteger,
		Span:    Span{Start: start, End: p.pos},
		payload: &IntegerExpr{Value: val},
	}, nil
}

// Infix operator precedence constants (higher = tighter binding, matching jj's Pest grammar)
// From weakest to strongest:
// ++ (concat, handled at template level)
// || (logical or)
// && (logical and)
// == != (equality)
// >= > <= < (comparison)
// + - (add/sub)
// * / % (mul/div/rem)
// pattern (p:x)
// prefix !, - (handled in parsePrefix)
// method call x.f() (handled in parseTerm)
const (
	precConcat     = 1
	precLogicalOr  = 2
	precLogicalAnd = 3
	precEqual      = 4
	precCompare    = 5
	precAddSub     = 6
	precMulDiv     = 7
	precPattern    = 8
	precPrefix     = 9
	precMethod     = 10
	precPrimary    = 11
)

func peekInfixOp(input string, pos int) (op string, prec int, rightAssoc bool) {
	remaining := input[pos:]
	// Check multi-char operators first
	switch {
	case len(remaining) >= 2 && remaining[:2] == "++":
		// ++ is the concatenation operator, handled at template level, not here
		return "", 0, false
	case len(remaining) >= 2 && remaining[:2] == "||":
		return "||", precLogicalOr, false
	case len(remaining) >= 2 && remaining[:2] == "&&":
		return "&&", precLogicalAnd, false
	case len(remaining) >= 2 && remaining[:2] == "==":
		return "==", precEqual, false
	case len(remaining) >= 2 && remaining[:2] == "!=":
		return "!=", precEqual, false
	case len(remaining) >= 2 && remaining[:2] == ">=":
		return ">=", precCompare, false
	case len(remaining) >= 2 && remaining[:2] == "<=":
		return "<=", precCompare, false
	}
	ch := rune(0)
	if len(remaining) > 0 {
		ch = rune(remaining[0])
	}
	switch ch {
	case '>':
		return ">", precCompare, false
	case '<':
		return "<", precCompare, false
	case '+':
		return "+", precAddSub, false
	case '-':
		return "-", precAddSub, false
	case '*':
		return "*", precMulDiv, false
	case '/':
		return "/", precMulDiv, false
	case '%':
		return "%", precMulDiv, false
	}
	return "", 0, false
}

func infixOpStringToBinaryOp(op string) BinaryOp {
	switch op {
	case "||":
		return LogicalOr
	case "&&":
		return LogicalAnd
	case "==":
		return Equal
	case "!=":
		return NotEqual
	case ">=":
		return GreaterEqual
	case ">":
		return Greater
	case "<=":
		return LessEqual
	case "<":
		return Less
	case "+":
		return Add
	case "-":
		return Sub
	case "*":
		return Mul
	case "/":
		return Div
	case "%":
		return Rem
	}
	return -1
}

func isInfixOpChar(ch rune) bool {
	return ch == '|' || ch == '&' || ch == '=' || ch == '!' ||
		ch == '>' || ch == '<' || ch == '+' || ch == '-' ||
		ch == '*' || ch == '/' || ch == '%'
}