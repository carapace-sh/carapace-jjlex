package jjlex

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

// CompletionType indicates what kind of completion is being requested
type CompletionType int

func (t CompletionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

const (
	CompletionTypeUnknown     CompletionType = iota
	CompletionTypeRevision                   // Completing a revision/symbol/function/alias
	CompletionTypeFunctionArg                // Completing a function argument
	CompletionTypeOperator                   // Completing after an operator
	CompletionTypePattern                    // Completing a pattern (substring:/exact:)
)

func (ct CompletionType) String() string {
	switch ct {
	case CompletionTypeRevision:
		return "Revision"
	case CompletionTypeFunctionArg:
		return "FunctionArgument"
	case CompletionTypeOperator:
		return "Operator"
	case CompletionTypePattern:
		return "Pattern"
	default:
		return "Unknown"
	}
}

// CompletionContext contains information about what to complete
type CompletionContext struct {
	// Type of completion being requested
	Type CompletionType

	// The partial text to complete (e.g., "auth" for "author()")
	Prefix string

	// For function arguments: the function name (e.g., "author")
	FunctionName string

	// For function arguments: the argument index (0-based)
	ArgumentIndex int

	// For function arguments: whether we're at a position expecting a revset or a string pattern
	ExpectingRevset bool

	// The full input up to the cursor
	FullInput string

	// The tokens parsed up to the completion point
	Tokens []Token

	// Whether the context is valid/unambiguous
	IsValid bool

	// Human-readable message explaining the context
	Message string

	// When completing an operator this completes the attached revset
	AttachedRevset string
}

// completionAnalyzer analyzes revset expressions for completion context
type completionAnalyzer struct {
	input string
}

// NewCompletionAnalyzer creates a new analyzer for the given input and cursor position
func newCompletionAnalyzer(input string) *completionAnalyzer {
	return &completionAnalyzer{
		input: input,
	}
}

// Analyze returns the completion context at the cursor position
func (ca *completionAnalyzer) Analyze() CompletionContext {
	// Tokenize up to cursor position
	tokenizer := newTokenizer(ca.input)
	tokens := tokenizer.TokenizeAll()

	ctx := CompletionContext{
		FullInput: ca.input,
		Tokens:    tokens,
		IsValid:   true,
	}

	// If empty or only whitespace
	if strings.TrimSpace(ca.input) == "" {
		ctx.Type = CompletionTypeRevision
		ctx.Prefix = ""
		ctx.Message = "Complete a revision, operator, or function"
		return ctx
	}

	// Remove EOF token
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == TokenEOF {
		tokens = tokens[:len(tokens)-1]
	}

	if len(tokens) == 0 {
		ctx.Type = CompletionTypeRevision
		ctx.Message = "Complete a revision, operator, or function"
		return ctx
	}

	// Check if input ends with whitespace (symbol is complete, expect operator)
	inputTrailingSpace := len(ca.input) > 0 && unicode.IsSpace(rune(ca.input[len(ca.input)-1]))

	lastToken := tokens[len(tokens)-1]

	// Determine prefix for completion
	ctx.Prefix = ca.extractPrefix(lastToken)

	// Analyze based on last token type
	switch lastToken.Type {
	case TokenSymbol, TokenQuotedString, TokenInteger:
		if inputTrailingSpace {
			ctx.Type = CompletionTypeOperator
			ctx.Prefix = ""
			ctx.Message = "Complete an operator or end expression"
			return ctx
		}
		return ca.analyzeSymbolCompletion(ctx, tokens, lastToken)

	case TokenLParen:
		return ca.analyzeFunctionArgument(ctx, tokens)

	case TokenComma:
		return ca.analyzeFunctionArgument(ctx, tokens)

	case TokenAmpersand, TokenPipe, TokenTilde:
		ctx.Type = CompletionTypeRevision
		ctx.Prefix = ""
		ctx.Message = "Complete a revision after operator"
		return ctx

	case TokenMinus, TokenPlus:
		if len(tokens) >= 2 {
			prevToken := tokens[len(tokens)-2]
			prevEnd := prevToken.Pos + len(prevToken.Value)
			if lastToken.Pos == prevEnd {
				ctx.Type = CompletionTypeOperator
				ctx.Prefix = ""
				ctx.AttachedRevset = ca.input
				ctx.Message = "Complete after postfix operator"
				return ctx
			}
		}
		ctx.Type = CompletionTypeRevision
		ctx.Prefix = ""
		ctx.Message = "Complete a revision after operator"
		return ctx

	case TokenDotDot, TokenColonColon:
		if lastToken.Value == "." {
			ctx.Type = CompletionTypeOperator
			ctx.Prefix = "."
			ctx.Message = "Complete after partial range operator"
			return ctx
		}
		if lastToken.Type == TokenColonColon {
			ctx.AttachedRevset = lastToken.Value
		}
		ctx.Type = CompletionTypeRevision
		ctx.Prefix = ""
		ctx.Message = "Complete a revision after operator"
		return ctx

	case TokenColon:
		return ca.analyzePatternPrefix(ctx, tokens)

	case TokenRParen:
		prefix := ""
		if !inputTrailingSpace {
			if funcName := ca.findEnclosingFunctionName(tokens); funcName != nil {
				ctx.AttachedRevset = ca.input[funcName.Pos:]
				if ca.isZeroArgFunctionCall(tokens, funcName) {
					prefix = ca.input[funcName.Pos:]
				}
			} else if funcName := ca.findTopLevelFunctionName(tokens); funcName != nil {
				ctx.AttachedRevset = ca.input[funcName.Pos:]
			}
		}
		ctx.Type = CompletionTypeOperator
		ctx.Prefix = prefix
		ctx.Message = "Complete an operator or end expression"
		return ctx

	case TokenError:
		if strings.HasPrefix(lastToken.Value, "unterminated quoted string") {
			ctx.Type = CompletionTypeRevision
			ctx.Prefix = ca.input[lastToken.Pos:]
			ctx.AttachedRevset = ca.input[lastToken.Pos:]
			ctx.ExpectingRevset = true
			ctx.Message = "Complete revision inside quoted string"
			return ctx
		}
		if strings.HasPrefix(lastToken.Value, "unexpected") {
			ctx.Type = CompletionTypeUnknown
			ctx.IsValid = false
			ctx.Message = fmt.Sprintf("Error: %s", lastToken.Value)
			return ctx
		}
		fallthrough

	default:
		ctx.Type = CompletionTypeRevision
		ctx.Prefix = ""
		ctx.Message = "Complete a revision or function"
		return ctx
	}
}

// analyzeSymbolCompletion handles completion for symbol tokens
func (ca *completionAnalyzer) analyzeSymbolCompletion(ctx CompletionContext, tokens []Token, lastToken Token) CompletionContext {
	ctx.Prefix = ca.extractPrefix(lastToken)

	isRevsetLike := lastToken.Type == TokenSymbol || lastToken.Type == TokenInteger || lastToken.Type == TokenQuotedString

	// Check if the token is directly attached to a preceding token
	if len(tokens) >= 2 {
		prevToken := tokens[len(tokens)-2]
		prevEnd := prevToken.Pos + len(prevToken.Value)
		if lastToken.Pos == prevEnd {
			switch prevToken.Type {
			case TokenLParen:
				ctx.AttachedRevset = lastToken.Value
			case TokenComma:
				ctx.AttachedRevset = lastToken.Value
			case TokenDotDot, TokenColonColon:
				ctx.AttachedRevset = ca.input[prevToken.Pos:]
			default:
				if isBinarySetOperator(prevToken.Type) {
					ctx.AttachedRevset = lastToken.Value
				}
			}
		} else if prevToken.Type == TokenComma {
			ctx.AttachedRevset = lastToken.Value
		} else if isBinarySetOperator(prevToken.Type) && !ca.isInsideBareParens(tokens) {
			ctx.AttachedRevset = lastToken.Value
		}
	}

	// Check if previous token is '(' or ','
	if len(tokens) >= 2 && (tokens[len(tokens)-2].Type == TokenLParen || tokens[len(tokens)-2].Type == TokenComma) {
		funcNameToken := ca.findFunctionName(tokens, len(tokens)-2)
		if funcNameToken != nil {
			argCount := ca.countFunctionArguments(tokens, len(tokens)-2)
			ctx.FunctionName = funcNameToken.Value
			ctx.Type = CompletionTypeFunctionArg
			ctx.ArgumentIndex = argCount
			ctx.IsValid = ca.isValidArgumentCount(funcNameToken.Value, argCount)
			ctx.ExpectingRevset = ca.expectsRevsetArgument(funcNameToken.Value, argCount)
			if !ctx.ExpectingRevset {
				ctx.AttachedRevset = ""
			}
			if !ctx.IsValid {
				ctx.Prefix = ""
				ctx.Type = CompletionTypeUnknown
				ctx.FunctionName = ""
				ctx.ArgumentIndex = 0
				ctx.ExpectingRevset = false
				ctx.AttachedRevset = ""
				ctx.Message = fmt.Sprintf("function '%s' has too many arguments", funcNameToken.Value)
			} else {
				ctx.Message = fmt.Sprintf("Complete argument for function '%s'", funcNameToken.Value)
			}
			return ctx
		}
	}

	ctx.Type = CompletionTypeRevision
	ctx.Message = fmt.Sprintf("Complete revision '%s' (branch, tag, commit ID, or alias)", ctx.Prefix)
	if ctx.AttachedRevset == "" && isRevsetLike && !hasPrecedingOperatorOrParen(tokens) {
		ctx.AttachedRevset = ca.extractPrefix(lastToken)
	}
	if lastToken.Type == TokenQuotedString {
		ctx.ExpectingRevset = true
	}
	return ctx
}

// analyzeFunctionArgument handles completion inside function arguments
func (ca *completionAnalyzer) analyzeFunctionArgument(ctx CompletionContext, tokens []Token) CompletionContext {
	// Find the function name
	funcNameToken := ca.findFunctionName(tokens, len(tokens)-1)
	if funcNameToken == nil {
		ctx.Type = CompletionTypeRevision
		ctx.Message = "Complete a revision or function"
		return ctx
	}

	ctx.FunctionName = funcNameToken.Value
	ctx.Type = CompletionTypeFunctionArg

	// Count arguments
	argCount := ca.countFunctionArguments(tokens, len(tokens)-1)
	ctx.ArgumentIndex = argCount

	ctx.ExpectingRevset = ca.expectsRevsetArgument(funcNameToken.Value, argCount)

	if !ca.isValidArgumentCount(funcNameToken.Value, argCount) {
		ctx.IsValid = false
		ctx.Prefix = ""
		ctx.Type = CompletionTypeUnknown
		ctx.FunctionName = ""
		ctx.ArgumentIndex = 0
		ctx.ExpectingRevset = false
		ctx.AttachedRevset = ""
		ctx.Message = fmt.Sprintf("function '%s' has too many arguments", funcNameToken.Value)
	} else if ctx.ExpectingRevset {
		ctx.Message = fmt.Sprintf("Complete revset argument %d for function '%s'", argCount+1, funcNameToken.Value)
	} else {
		ctx.Message = fmt.Sprintf("Complete string pattern for function '%s'", funcNameToken.Value)
	}

	return ctx
}

// analyzePatternPrefix handles completion after pattern prefix (exact:, substring:)
func (ca *completionAnalyzer) analyzePatternPrefix(ctx CompletionContext, tokens []Token) CompletionContext {
	// Check what's before the colon
	if len(tokens) >= 2 {
		prevToken := tokens[len(tokens)-2]
		if prevToken.Type == TokenSymbol {
			prefix := prevToken.Value
			if prefix == "exact" || prefix == "substring" {
				ctx.Type = CompletionTypePattern
				ctx.Prefix = ""
				ctx.Message = fmt.Sprintf("Complete pattern after '%s:'", prefix)
				return ctx
			}
		}
	}

	ctx.Type = CompletionTypeOperator
	ctx.Prefix = ":"
	ctx.Message = "Complete after operator"
	return ctx
}

// Helper methods

// isInsideBareParens checks if we're inside parentheses that are NOT a function call
func (ca *completionAnalyzer) isInsideBareParens(tokens []Token) bool {
	depth := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		switch tokens[i].Type {
		case TokenRParen:
			depth++
		case TokenLParen:
			if depth > 0 {
				depth--
			} else {
				// Check if this paren is part of a function call
				if i > 0 && tokens[i-1].Type == TokenSymbol {
					return false
				}
				return true
			}
		}
	}
	return false
}

func hasPrecedingOperatorOrParen(tokens []Token) bool {
	for i := len(tokens) - 2; i >= 0; i-- {
		tt := tokens[i].Type
		if tt == TokenLParen || isOperatorToken(tt) || tt == TokenRParen {
			return true
		}
		if tt == TokenComma {
			continue
		}
		break
	}
	return false
}

func isBinarySetOperator(tt TokenType) bool {
	switch tt {
	case TokenAmpersand, TokenPipe, TokenTilde, TokenMinus, TokenPlus:
		return true
	}
	return false
}

func isOperatorToken(tt TokenType) bool {
	switch tt {
	case TokenAmpersand, TokenPipe, TokenTilde, TokenMinus, TokenPlus, TokenDotDot, TokenColonColon:
		return true
	}
	return false
}

// isZeroArgFunctionCall checks if the given function name token starts a zero-arg call
// by looking for Symbol LParen RParen pattern
func (ca *completionAnalyzer) isZeroArgFunctionCall(tokens []Token, funcName *Token) bool {
	for i, tok := range tokens {
		if tok.Pos == funcName.Pos && tok.Type == TokenSymbol {
			if i+2 < len(tokens) && tokens[i+1].Type == TokenLParen && tokens[i+2].Type == TokenRParen {
				return true
			}
			return false
		}
	}
	return false
}

// findEnclosingFunctionName finds the function name for the innermost function call
// that the last token (a closing paren) belongs to
func (ca *completionAnalyzer) findEnclosingFunctionName(tokens []Token) *Token {
	depth := 0
	for i := len(tokens) - 1; i >= 0; i-- {
		switch tokens[i].Type {
		case TokenRParen:
			depth++
		case TokenLParen:
			depth--
			if depth == 0 && i > 0 && tokens[i-1].Type == TokenSymbol {
				return &tokens[i-1]
			}
		}
	}
	return nil
}

func (ca *completionAnalyzer) findTopLevelFunctionName(tokens []Token) *Token {
	if len(tokens) < 3 {
		return nil
	}
	if tokens[0].Type != TokenSymbol || tokens[1].Type != TokenLParen {
		return nil
	}
	if tokens[len(tokens)-1].Type != TokenRParen {
		return nil
	}
	// Verify parens are balanced and the last RParen matches the first LParen
	depth := 0
	for i := 1; i < len(tokens)-1; i++ {
		switch tokens[i].Type {
		case TokenLParen:
			depth++
		case TokenRParen:
			depth--
			if depth < 0 {
				return nil
			}
		}
	}
	if depth != 0 {
		return nil
	}
	return &tokens[0]
}

func (ca *completionAnalyzer) extractPrefix(lastToken Token) string {
	if lastToken.Type == TokenSymbol || lastToken.Type == TokenInteger {
		return lastToken.Value
	}
	if lastToken.Type == TokenQuotedString {
		return `"` + lastToken.Value + `"`
	}
	return ""
}

func (ca *completionAnalyzer) findFunctionName(tokens []Token, upToIndex int) *Token {
	// Search backwards for a symbol followed by '('
	for i := upToIndex; i >= 0; i-- {
		if tokens[i].Type == TokenLParen && i > 0 && tokens[i-1].Type == TokenSymbol {
			return &tokens[i-1]
		}
	}
	return nil
}

func (ca *completionAnalyzer) countFunctionArguments(tokens []Token, upToIndex int) int {
	// Find the opening paren
	parenIndex := -1
	for i := upToIndex; i >= 0; i-- {
		if tokens[i].Type == TokenLParen {
			parenIndex = i
			break
		}
	}

	if parenIndex == -1 {
		return 0
	}

	// Count commas between opening paren and current position
	commaCount := 0
	depth := 0
	for i := parenIndex + 1; i <= upToIndex && i < len(tokens); i++ {
		switch tokens[i].Type {
		case TokenLParen:
			depth++
		case TokenRParen:
			depth--
		case TokenComma:
			if depth == 0 {
				commaCount++
			}
		}
	}

	return commaCount
}

func (ca *completionAnalyzer) expectsRevsetArgument(funcName string, argIndex int) bool {
	// Functions that take revset arguments in their first position
	revsetFirstFuncs := map[string]bool{
		"parents":     true,
		"children":    true,
		"ancestors":   true,
		"descendants": true,
		"connected":   true,
		"heads":       true,
		"roots":       true,
		"latest":      true,
		"present":     true,
	}

	// Functions that take only revset arguments
	revsetOnlyFuncs := map[string]bool{
		"all":           true,
		"none":          true,
		"tags":          true,
		"git_refs":      true,
		"git_head":      true,
		"visible_heads": true,
		"root":          true,
		"merges":        true,
		"empty":         true,
		"conflict":      true,
	}

	// Functions that take string patterns
	patternFuncs := map[string]bool{
		"branches":        true,
		"remote_branches": true,
		"description":     true,
		"author":          true,
		"committer":       true,
		"file":            true,
		"mine":            true,
	}

	// No-argument functions
	noArgFuncs := map[string]bool{
		"all":           true,
		"none":          true,
		"tags":          true,
		"git_refs":      true,
		"git_head":      true,
		"visible_heads": true,
		"root":          true,
		"merges":        true,
		"empty":         true,
		"conflict":      true,
		"mine":          true,
	}

	if noArgFuncs[funcName] {
		return false
	}

	if revsetOnlyFuncs[funcName] {
		return true
	}

	if revsetFirstFuncs[funcName] {
		return argIndex == 0
	}

	if patternFuncs[funcName] {
		return false
	}

	// latest(x, count) - first is revset, second is integer
	if funcName == "latest" {
		return argIndex == 0
	}

	// remote_branches(pattern, [remote=pattern]) - both are patterns
	if funcName == "remote_branches" {
		return false
	}

	// Default: assume revset
	return true
}

func maxFunctionArguments(funcName string) (int, bool) {
	maxArgs := map[string]int{
		"all":             0,
		"none":            0,
		"tags":            0,
		"git_refs":        0,
		"git_head":        0,
		"visible_heads":   0,
		"root":            0,
		"merges":          0,
		"empty":           0,
		"conflict":        0,
		"mine":            0,
		"parents":         2,
		"children":        2,
		"ancestors":       2,
		"descendants":     2,
		"connected":       1,
		"heads":           1,
		"roots":           1,
		"latest":          2,
		"present":         1,
		"branches":        1,
		"remote_branches": 2,
		"description":     1,
		"author":          1,
		"committer":       1,
		"file":            1,
		"bookmark":        0,
	}
	if max, ok := maxArgs[funcName]; ok {
		return max, true
	}
	return 0, false
}

func (ca *completionAnalyzer) isValidArgumentCount(funcName string, argCount int) bool {
	if max, known := maxFunctionArguments(funcName); known {
		return argCount < max
	}
	return true
}
