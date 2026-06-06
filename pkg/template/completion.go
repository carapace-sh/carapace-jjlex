package template

// ExpectedToken represents a type of token expected at a completion position.
type ExpectedToken int

const (
	ExpectedExpression ExpectedToken = iota
	ExpectedOperator
	ExpectedClosingParen
	ExpectedComma
	ExpectedEquals
	ExpectedStringClose
	ExpectedPatternValue
	ExpectedLambdaClose
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
	case ExpectedLambdaClose:
		return "|"
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
	// Name is the function name (e.g. "if", "label")
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
	// IsMethod is true if this is a method call rather than a function call
	IsMethod bool `json:"isMethod"`
	// MethodObject is the expression the method is called on (if IsMethod)
	MethodObject *Expression `json:"methodObject,omitempty"`
	// IsZeroArg is true when the function takes no arguments
	IsZeroArg bool `json:"isZeroArg"`
}

// CompletionContext describes what is expected at the completion position.
type CompletionContext struct {
	// ExpectedTokens lists the types of tokens expected at the position
	ExpectedTokens []ExpectedToken `json:"expectedTokens"`

	// ValidOperators lists the operators valid at this position
	ValidOperators []ValidOperator `json:"validOperators,omitempty"`

	// PartialIdent is the partial identifier being typed (e.g. "commi" in "commi")
	PartialIdent string `json:"partialIdent,omitempty"`

	// PartialString is the partial string literal content being typed
	PartialString string `json:"partialString,omitempty"`

	// StringQuote is the quote character used for the partial string (' or ")
	StringQuote rune `json:"stringQuote,omitempty"`

	// Function is non-nil when the cursor is inside a function/method call
	Function *FunctionContext `json:"function,omitempty"`

	// InPattern is true when completing inside a pattern (name:...)
	InPattern bool `json:"inPattern"`
	// PatternName is the pattern prefix name (e.g. "exact" in "exact:")
	PatternName string `json:"patternName,omitempty"`

	// InLambda is true when the cursor is inside a lambda body
	InLambda bool `json:"inLambda"`
	// LambdaParams are the parameter names of the enclosing lambda
	LambdaParams []string `json:"lambdaParams,omitempty"`
}
