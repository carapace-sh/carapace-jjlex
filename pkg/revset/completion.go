package revset

// ExpectedToken represents a type of token expected at a completion position.
type ExpectedToken int

const (
	// ExpectedExpression means any primary expression is valid
	// (identifier, string, function call, parenthesized expr, @, ::, ..)
	ExpectedExpression ExpectedToken = iota

	// ExpectedOperator means an infix or postfix operator is valid
	ExpectedOperator

	// ExpectedClosingParen means a ')' is expected
	ExpectedClosingParen

	// ExpectedComma means a ',' is expected (in function args)
	ExpectedComma

	// ExpectedEquals means a '=' is expected (for keyword args in functions)
	ExpectedEquals

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
	case ExpectedEquals:
		return "="
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
	// Name is the function name (e.g. "parents")
	Name string `json:"name"`
	// Args are the positional arguments parsed so far
	Args []*Expression `json:"args,omitempty"`
	// KeywordArgs are the keyword arguments parsed so far
	KeywordArgs []KeywordArg `json:"keywordArgs,omitempty"`
	// ArgIndex is the 0-based index of the argument being completed
	ArgIndex int `json:"argIndex"`
	// IsKeywordArg is true when completing a keyword argument name
	IsKeywordArg bool `json:"isKeywordArg"`
	// KeywordArgName is the name of the keyword arg being completed
	KeywordArgName string `json:"keywordArgName,omitempty"`
	// IsZeroArg is true when the function takes no arguments
	// (e.g. root(), all(), none())
	IsZeroArg bool `json:"isZeroArg"`
	// InStringArg is true when the current argument is a partial string literal
	// (e.g. completing author("Ste) — the string value is in PartialString)
	InStringArg bool `json:"inStringArg"`
}

// CompletionContext describes what is expected at the completion position.
type CompletionContext struct {
	// ExpectedTokens lists the types of tokens expected at the position
	ExpectedTokens []ExpectedToken `json:"expectedTokens"`

	// ValidOperators lists the operators valid at this position
	// (only populated when ExpectedOperator is in ExpectedTokens)
	ValidOperators []ValidOperator `json:"validOperators,omitempty"`

	// PartialIdent is the partial identifier being typed (e.g. "par" in "par")
	PartialIdent string `json:"partialIdent,omitempty"`

	// PartialString is the partial string literal content being typed (e.g. "fo" in "fo")
	// without the surrounding quotes
	PartialString string `json:"partialString,omitempty"`

	// StringQuote is the quote character used for the partial string (' or ")
	StringQuote rune `json:"stringQuote,omitempty"`

	// Function is non-nil when the cursor is inside a function call
	Function *FunctionContext `json:"function,omitempty"`

	// InPattern is true when completing inside a pattern (name:...)
	InPattern bool `json:"inPattern"`
	// PatternName is the pattern prefix name (e.g. "exact" in "exact:")
	PatternName string `json:"patternName,omitempty"`

	// InRemoteSymbol is true when completing the remote name part of a
	// name@remote expression (e.g. completing "origin" in "main@origin").
	InRemoteSymbol bool `json:"inRemoteSymbol"`
	// PartialRemote is the partial remote name being typed (e.g. "ori" in "main@ori")
	PartialRemote string `json:"partialRemote,omitempty"`
	// RemoteBookmarkName is the bookmark name before @ in a name@remote expression
	// (e.g. "main" in "main@origin")
	RemoteBookmarkName string `json:"remoteBookmarkName,omitempty"`

	// AttachedRevset is the revset expression that postfix operators are
	// attached to (e.g. "@-" has AttachedRevset "@-"). Used to determine
	// whether ActionAncestors or ActionDescendants should be invoked.
	AttachedRevset string `json:"attachedRevset,omitempty"`
}
