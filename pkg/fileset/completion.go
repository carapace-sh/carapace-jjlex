package fileset

// ExpectedToken represents a type of token expected at a completion position.
type ExpectedToken int

const (
	// ExpectedExpression means any primary expression is valid
	// (identifier, string, function call, parenthesized expr, pattern)
	ExpectedExpression ExpectedToken = iota

	// ExpectedOperator means an infix or prefix operator is valid
	ExpectedOperator

	// ExpectedClosingParen means a ')' is expected
	ExpectedClosingParen

	// ExpectedComma means a ',' is expected (in function args)
	ExpectedComma

	// ExpectedStringClose means a closing quote is expected
	ExpectedStringClose

	// ExpectedPatternValue means a pattern value after ':' is expected
	ExpectedPatternValue
)

func (t ExpectedToken) String() string {
	switch t {
	case ExpectedExpression:
		return "Expression"
	case ExpectedOperator:
		return "Operator"
	case ExpectedClosingParen:
		return ")"
	case ExpectedComma:
		return ","
	case ExpectedStringClose:
		return "quote"
	case ExpectedPatternValue:
		return "PatternValue"
	}
	return "Unknown"
}

func (t ExpectedToken) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// ValidOperator represents an operator that could be valid at a completion position.
type ValidOperator struct {
	Op          string `json:"op"`
	Description string `json:"description"`
}

// FunctionContext provides details about an ongoing function call at the completion position.
type FunctionContext struct {
	// Name is the function name (e.g. "all")
	Name string `json:"name"`
	// Args are the positional arguments parsed so far
	Args []*Expression `json:"args,omitempty"`
	// ArgIndex is the 0-based index of the argument being completed
	ArgIndex int `json:"argIndex"`
}

// CompletionContext describes what is expected at the completion position.
type CompletionContext struct {
	// ExpectedTokens lists the types of tokens expected at the position
	ExpectedTokens []ExpectedToken `json:"expectedTokens"`

	// ValidOperators lists the operators valid at this position
	// (only populated when ExpectedOperator is in ExpectedTokens)
	ValidOperators []ValidOperator `json:"validOperators,omitempty"`

	// PartialIdent is the partial identifier being typed (e.g. "al" in "al")
	PartialIdent string `json:"partialIdent,omitempty"`

	// PartialString is the partial string literal content being typed
	// without the surrounding quotes
	PartialString string `json:"partialString,omitempty"`

	// StringQuote is the quote character used for the partial string (' or ")
	StringQuote rune `json:"stringQuote,omitempty"`

	// Function is non-nil when the cursor is inside a function call
	Function *FunctionContext `json:"function,omitempty"`

	// InPattern is true when completing inside a pattern (name:...)
	InPattern bool `json:"inPattern"`
	// PatternName is the pattern prefix name (e.g. "glob" in "glob:")
	PatternName string `json:"patternName,omitempty"`
}
