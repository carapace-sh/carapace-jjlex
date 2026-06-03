package template

type UnaryOp int

const (
	LogicalNot UnaryOp = iota
	Negate
)

func (op UnaryOp) String() string {
	switch op {
	case LogicalNot:
		return "!"
	case Negate:
		return "-"
	}
	return ""
}

type BinaryOp int

const (
	LogicalOr BinaryOp = iota
	LogicalAnd
	Equal
	NotEqual
	GreaterEqual
	Greater
	LessEqual
	Less
	Add
	Sub
	Mul
	Div
	Rem
)

func (op BinaryOp) String() string {
	switch op {
	case LogicalOr:
		return "||"
	case LogicalAnd:
		return "&&"
	case Equal:
		return "=="
	case NotEqual:
		return "!="
	case GreaterEqual:
		return ">="
	case Greater:
		return ">"
	case LessEqual:
		return "<="
	case Less:
		return "<"
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"
	case Rem:
		return "%"
	}
	return ""
}

type ExpressionKind int

const (
	KindIdentifier ExpressionKind = iota
	KindBoolean
	KindInteger
	KindString
	KindPattern
	KindUnary
	KindBinary
	KindConcat
	KindFunctionCall
	KindMethodCall
	KindLambda
)

func (k ExpressionKind) String() string {
	switch k {
	case KindIdentifier:
		return "Identifier"
	case KindBoolean:
		return "Boolean"
	case KindInteger:
		return "Integer"
	case KindString:
		return "String"
	case KindPattern:
		return "Pattern"
	case KindUnary:
		return "Unary"
	case KindBinary:
		return "Binary"
	case KindConcat:
		return "Concat"
	case KindFunctionCall:
		return "FunctionCall"
	case KindMethodCall:
		return "MethodCall"
	case KindLambda:
		return "Lambda"
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

func (e *Expression) BooleanValue() bool {
	if e.Kind != KindBoolean {
		return false
	}
	return e.payload.(*BooleanExpr).Value
}

func (e *Expression) IntegerValue() int64 {
	if e.Kind != KindInteger {
		return 0
	}
	return e.payload.(*IntegerExpr).Value
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

func (e *Expression) ConcatNodes() []*Expression {
	if e.Kind != KindConcat {
		return nil
	}
	return e.payload.(*ConcatExpr).Nodes
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

func (e *Expression) MethodObject() *Expression {
	if e.Kind != KindMethodCall {
		return nil
	}
	return e.payload.(*MethodCallExpr).Object
}

func (e *Expression) MethodName() string {
	if e.Kind != KindMethodCall {
		return ""
	}
	return e.payload.(*MethodCallExpr).Function.Name
}

func (e *Expression) MethodArgs() []*Expression {
	if e.Kind != KindMethodCall {
		return nil
	}
	return e.payload.(*MethodCallExpr).Function.Args
}

func (e *Expression) MethodKeywordArgs() []KeywordArg {
	if e.Kind != KindMethodCall {
		return nil
	}
	return e.payload.(*MethodCallExpr).Function.KeywordArgs
}

func (e *Expression) LambdaParams() []string {
	if e.Kind != KindLambda {
		return nil
	}
	return e.payload.(*LambdaExpr).Params
}

func (e *Expression) LambdaBody() *Expression {
	if e.Kind != KindLambda {
		return nil
	}
	return e.payload.(*LambdaExpr).Body
}

type IdentifierExpr struct {
	Name string
}

type BooleanExpr struct {
	Value bool
}

type IntegerExpr struct {
	Value int64
}

type StringExpr struct {
	Value string
}

type PatternExpr struct {
	Name  string
	Value *Expression
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

type ConcatExpr struct {
	Nodes []*Expression
}

type FunctionCallExpr struct {
	Name        string
	Args        []*Expression
	KeywordArgs []KeywordArg
}

type MethodCallExpr struct {
	Object   *Expression
	Function *FunctionCallExpr
}

type LambdaExpr struct {
	Params []string
	Body   *Expression
}

type KeywordArg struct {
	Name  string
	Value *Expression
}
