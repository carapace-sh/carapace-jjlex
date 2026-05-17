package revset

type UnaryOp int

const (
	Negate UnaryOp = iota
	DagRangePre
	DagRangePost
	RangePre
	RangePost
	Parents
	Children
)

func (op UnaryOp) String() string {
	switch op {
	case Negate:
		return "~"
	case DagRangePre:
		return "::"
	case DagRangePost:
		return "::"
	case RangePre:
		return ".."
	case RangePost:
		return ".."
	case Parents:
		return "-"
	case Children:
		return "+"
	}
	return ""
}

type BinaryOp int

const (
	Intersection BinaryOp = iota
	Difference
	DagRange
	Range
)

func (op BinaryOp) String() string {
	switch op {
	case Intersection:
		return "&"
	case Difference:
		return "~"
	case DagRange:
		return "::"
	case Range:
		return ".."
	}
	return ""
}

type ExpressionKind int

const (
	KindIdentifier ExpressionKind = iota
	KindString
	KindPattern
	KindRemoteSymbol
	KindAtWorkspace
	KindAtCurrentWorkspace
	KindDagRangeAll
	KindRangeAll
	KindUnary
	KindBinary
	KindUnionAll
	KindFunctionCall
)

func (k ExpressionKind) String() string {
	switch k {
	case KindIdentifier:
		return "Identifier"
	case KindString:
		return "String"
	case KindPattern:
		return "Pattern"
	case KindRemoteSymbol:
		return "RemoteSymbol"
	case KindAtWorkspace:
		return "AtWorkspace"
	case KindAtCurrentWorkspace:
		return "AtCurrentWorkspace"
	case KindDagRangeAll:
		return "DagRangeAll"
	case KindRangeAll:
		return "RangeAll"
	case KindUnary:
		return "Unary"
	case KindBinary:
		return "Binary"
	case KindUnionAll:
		return "UnionAll"
	case KindFunctionCall:
		return "FunctionCall"
	}
	return "Unknown"
}

type Expression struct {
	Kind    ExpressionKind
	Span    Span
	payload any
}

func (e *Expression) Identifier() string {
	if e.Kind != KindIdentifier {
		return ""
	}
	return e.payload.(*IdentifierExpr).Name
}

func (e *Expression) StringValue() string {
	if e.Kind != KindString {
		return ""
	}
	return e.payload.(*StringExpr).Value
}

func (e *Expression) PatternName() string {
	if e.Kind != KindPattern {
		return ""
	}
	return e.payload.(*PatternExpr).Name
}

func (e *Expression) PatternValue() *Expression {
	if e.Kind != KindPattern {
		return nil
	}
	return e.payload.(*PatternExpr).Value
}

func (e *Expression) RemoteSymbolName() string {
	if e.Kind != KindRemoteSymbol {
		return ""
	}
	return e.payload.(*RemoteSymbolExpr).Name
}

func (e *Expression) RemoteSymbolRemote() string {
	if e.Kind != KindRemoteSymbol {
		return ""
	}
	return e.payload.(*RemoteSymbolExpr).Remote
}

func (e *Expression) AtWorkspaceName() string {
	if e.Kind != KindAtWorkspace {
		return ""
	}
	return e.payload.(*AtWorkspaceExpr).Name
}

func (e *Expression) UnaryOp() UnaryOp {
	if e.Kind != KindUnary {
		return -1
	}
	return e.payload.(*UnaryExpr).Op
}

func (e *Expression) UnaryArg() *Expression {
	if e.Kind != KindUnary {
		return nil
	}
	return e.payload.(*UnaryExpr).Arg
}

func (e *Expression) BinaryOp() BinaryOp {
	if e.Kind != KindBinary {
		return -1
	}
	return e.payload.(*BinaryExpr).Op
}

func (e *Expression) BinaryLHS() *Expression {
	if e.Kind != KindBinary {
		return nil
	}
	return e.payload.(*BinaryExpr).LHS
}

func (e *Expression) BinaryRHS() *Expression {
	if e.Kind != KindBinary {
		return nil
	}
	return e.payload.(*BinaryExpr).RHS
}

func (e *Expression) UnionNodes() []*Expression {
	if e.Kind != KindUnionAll {
		return nil
	}
	return e.payload.(*UnionAllExpr).Nodes
}

func (e *Expression) FunctionName() string {
	if e.Kind != KindFunctionCall {
		return ""
	}
	return e.payload.(*FunctionCallExpr).Name
}

func (e *Expression) FunctionArgs() []*Expression {
	if e.Kind != KindFunctionCall {
		return nil
	}
	return e.payload.(*FunctionCallExpr).Args
}

func (e *Expression) FunctionKeywordArgs() []KeywordArg {
	if e.Kind != KindFunctionCall {
		return nil
	}
	return e.payload.(*FunctionCallExpr).KeywordArgs
}

type IdentifierExpr struct {
	Name string
}

type StringExpr struct {
	Value string
}

type PatternExpr struct {
	Name  string
	Value *Expression
}

type RemoteSymbolExpr struct {
	Name   string
	Remote string
}

type AtWorkspaceExpr struct {
	Name string
}

type UnaryExpr struct {
	Op  UnaryOp
	Arg *Expression
}

type BinaryExpr struct {
	Op  BinaryOp
	LHS *Expression
	RHS *Expression
}

type UnionAllExpr struct {
	Nodes []*Expression
}

type FunctionCallExpr struct {
	Name        string
	Args        []*Expression
	KeywordArgs []KeywordArg
}

type KeywordArg struct {
	Name  string
	Value *Expression
}
