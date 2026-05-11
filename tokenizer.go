package jjlex

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token in a revset expression
type TokenType int

func (t TokenType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

const (
	// Literal tokens
	TokenEOF TokenType = iota
	TokenSymbol
	TokenString
	TokenInteger

	// Operators (binary)
	TokenAmpersand  // &
	TokenPipe       // |
	TokenTilde      // ~
	TokenMinus      // -
	TokenPlus       // +
	TokenDotDot     // ..
	TokenColonColon // ::

	// Operators (postfix)
	TokenMinusSuffix // - (postfix)
	TokenPlusSuffix  // + (postfix)

	// Parentheses
	TokenLParen // (
	TokenRParen // )

	// Function call
	TokenComma // ,

	// Quoted symbol
	TokenQuotedString // "..."

	// Pattern qualifiers
	TokenColon // : (for pattern prefixes like exact:, substring:)

	// Error
	TokenError
)

// Token represents a lexical token from a revset expression
type Token struct {
	Type  TokenType
	Value string
	Pos   int // Position in input
}

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenSymbol:
		return "Symbol"
	case TokenString:
		return "String"
	case TokenInteger:
		return "Integer"
	case TokenAmpersand:
		return "&"
	case TokenPipe:
		return "|"
	case TokenTilde:
		return "~"
	case TokenMinus:
		return "-"
	case TokenPlus:
		return "+"
	case TokenDotDot:
		return ".."
	case TokenColonColon:
		return "::"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenComma:
		return ","
	case TokenQuotedString:
		return "QuotedString"
	case TokenColon:
		return ":"
	case TokenError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Tokenizer lexically analyzes a revset expression
type Tokenizer struct {
	input   string
	pos     int
	current rune
}

// NewTokenizer creates a new Tokenizer for the given input
func NewTokenizer(input string) *Tokenizer {
	t := &Tokenizer{
		input: input,
		pos:   0,
	}
	if len(input) > 0 {
		t.current = rune(input[0])
	}
	return t
}

// NextToken returns the next token from the input
func (t *Tokenizer) NextToken() Token {
	// Skip whitespace
	for t.pos < len(t.input) && unicode.IsSpace(t.current) {
		t.advance()
	}

	if t.pos >= len(t.input) {
		return Token{Type: TokenEOF, Pos: t.pos}
	}

	pos := t.pos

	// Quoted strings (double-quoted)
	if t.current == '"' {
		return t.readQuotedString()
	}

	// Single-quoted strings (literal symbols)
	if t.current == '\'' {
		return t.readSingleQuotedString()
	}

	// Operators
	switch t.current {
	case '&':
		t.advance()
		return Token{Type: TokenAmpersand, Value: "&", Pos: pos}

	case '|':
		t.advance()
		return Token{Type: TokenPipe, Value: "|", Pos: pos}

	case '~':
		t.advance()
		return Token{Type: TokenTilde, Value: "~", Pos: pos}

	case '(':
		t.advance()
		return Token{Type: TokenLParen, Value: "(", Pos: pos}

	case ')':
		t.advance()
		return Token{Type: TokenRParen, Value: ")", Pos: pos}

	case ',':
		t.advance()
		return Token{Type: TokenComma, Value: ",", Pos: pos}

	case '.':
		if t.peek() == '.' {
			t.advance()
			t.advance()
			return Token{Type: TokenDotDot, Value: "..", Pos: pos}
		}
		return Token{Type: TokenError, Value: "unexpected '.'", Pos: pos}

	case ':':
		if t.peek() == ':' {
			t.advance()
			t.advance()
			return Token{Type: TokenColonColon, Value: "::", Pos: pos}
		}
		t.advance()
		return Token{Type: TokenColon, Value: ":", Pos: pos}

	case '+':
		t.advance()
		return Token{Type: TokenPlus, Value: "+", Pos: pos}

	case '-':
		t.advance()
		return Token{Type: TokenMinus, Value: "-", Pos: pos}
	}

	// Numbers
	if unicode.IsDigit(t.current) {
		return t.readNumber()
	}

	// Symbols and identifiers
	if isSymbolStart(t.current) {
		return t.readSymbol()
	}

	// Unknown character
	ch := t.current
	t.advance()
	return Token{
		Type:  TokenError,
		Value: fmt.Sprintf("unexpected character: %q", ch),
		Pos:   pos,
	}
}

// TokenizeAll tokenizes the entire input and returns all tokens
func (t *Tokenizer) TokenizeAll() []Token {
	var tokens []Token
	for {
		tok := t.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

// Private helper methods

func (t *Tokenizer) advance() {
	t.pos++
	if t.pos >= len(t.input) {
		t.current = 0
	} else {
		t.current = rune(t.input[t.pos])
	}
}

func (t *Tokenizer) peek() rune {
	if t.pos+1 >= len(t.input) {
		return 0
	}
	return rune(t.input[t.pos+1])
}

func (t *Tokenizer) peekN(n int) rune {
	if t.pos+n >= len(t.input) {
		return 0
	}
	return rune(t.input[t.pos+n])
}

func (t *Tokenizer) readQuotedString() Token {
	pos := t.pos
	t.advance() // skip opening "

	var sb strings.Builder
	for t.pos < len(t.input) && t.current != '"' {
		if t.current == '\\' {
			t.advance()
			if t.pos >= len(t.input) {
				return Token{
					Type:  TokenError,
					Value: "unterminated quoted string",
					Pos:   pos,
				}
			}
			// Handle escape sequences
			switch t.current {
			case '"':
				sb.WriteRune('"')
			case '\\':
				sb.WriteRune('\\')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case 'n':
				sb.WriteRune('\n')
			case '0':
				sb.WriteRune('\x00')
			case 'e':
				sb.WriteRune('\x1b')
			case 'x':
				// \xHH hex escape
				if t.peekN(1) != 0 && t.peekN(2) != 0 {
					t.advance()
					hex := string([]rune{t.current, t.peek()})
					var b byte
					_, _ = fmt.Sscanf(hex, "%x", &b)
					sb.WriteRune(rune(b))
					t.advance()
				} else {
					return Token{
						Type:  TokenError,
						Value: "invalid \\x escape sequence",
						Pos:   pos,
					}
				}
				t.advance()
				continue
			default:
				return Token{
					Type:  TokenError,
					Value: fmt.Sprintf("invalid escape sequence: \\%q", t.current),
					Pos:   pos,
				}
			}
			t.advance()
		} else {
			sb.WriteRune(t.current)
			t.advance()
		}
	}

	if t.current != '"' {
		return Token{
			Type:  TokenError,
			Value: "unterminated quoted string",
			Pos:   pos,
		}
	}

	t.advance() // skip closing "

	return Token{
		Type:  TokenQuotedString,
		Value: sb.String(),
		Pos:   pos,
	}
}

func (t *Tokenizer) readSingleQuotedString() Token {
	pos := t.pos
	t.advance() // skip opening '

	var sb strings.Builder
	for t.pos < len(t.input) && t.current != '\'' {
		sb.WriteRune(t.current)
		t.advance()
	}

	if t.current != '\'' {
		return Token{
			Type:  TokenError,
			Value: "unterminated single-quoted string",
			Pos:   pos,
		}
	}

	t.advance() // skip closing '

	return Token{
		Type:  TokenSymbol,
		Value: sb.String(),
		Pos:   pos,
	}
}

func (t *Tokenizer) readNumber() Token {
	pos := t.pos
	var sb strings.Builder

	for t.pos < len(t.input) && unicode.IsDigit(t.current) {
		sb.WriteRune(t.current)
		t.advance()
	}

	return Token{
		Type:  TokenInteger,
		Value: sb.String(),
		Pos:   pos,
	}
}

func (t *Tokenizer) readSymbol() Token {
	pos := t.pos
	var sb strings.Builder

	// Read identifier (can include alphanumeric, _, @, etc.)
	for t.pos < len(t.input) && isSymbolChar(t.current) {
		sb.WriteRune(t.current)
		t.advance()
	}

	value := sb.String()
	return Token{
		Type:  TokenSymbol,
		Value: value,
		Pos:   pos,
	}
}

func isSymbolStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '@'
}

func isSymbolChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '@' || ch == '-'
}
