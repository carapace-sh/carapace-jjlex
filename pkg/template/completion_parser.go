package template

// ParseForCompletion parses a partial template expression and returns a
// CompletionContext describing what is expected at the end of the input.
func ParseForCompletion(input string) *CompletionContext {
	cursor := len(input)
	p := &compParser{
		input:  input,
		pos:    0,
		cursor: cursor,
		ctx:    &CompletionContext{},
	}
	p.skipWS()
	p.parseTemplateComp()
	if len(p.ctx.ExpectedTokens) == 0 {
		p.ctx.ExpectedTokens = append(p.ctx.ExpectedTokens, ExpectedExpression)
	}
	p.ctx.ExpectedTokens = dedupTokens(p.ctx.ExpectedTokens)
	p.ctx.ValidOperators = dedupOperators(p.ctx.ValidOperators)
	return p.ctx
}

type compParser struct {
	input         string
	pos           int
	cursor        int
	ctx           *CompletionContext
	consumed      bool
	funcStack     []*funcParseState
	innermostFunc *FunctionContext
	lastExpr      *Expression
}

type funcParseState struct {
	name        string
	args        []*Expression
	keywordArgs []KeywordArg
	argIndex    int
	isMethod    bool
	methodObj   *Expression
}
