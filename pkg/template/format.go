package template

import (
	"fmt"
	"strings"
)

// Format returns a normalized string representation of the expression.
func Format(expr *Expression) string {
	return expr.String()
}

func (e *Expression) String() string {
	return formatExprInner(e)
}

func (op UnaryOp) GoString() string {
	switch op {
	case LogicalNot:
		return "template.LogicalNot"
	case Negate:
		return "template.Negate"
	}
	return "template.UnknownUnary"
}

func (op BinaryOp) GoString() string {
	switch op {
	case LogicalOr:
		return "template.LogicalOr"
	case LogicalAnd:
		return "template.LogicalAnd"
	case Equal:
		return "template.Equal"
	case NotEqual:
		return "template.NotEqual"
	case GreaterEqual:
		return "template.GreaterEqual"
	case Greater:
		return "template.Greater"
	case LessEqual:
		return "template.LessEqual"
	case Less:
		return "template.Less"
	case Add:
		return "template.Add"
	case Sub:
		return "template.Sub"
	case Mul:
		return "template.Mul"
	case Div:
		return "template.Div"
	case Rem:
		return "template.Rem"
	}
	return "template.UnknownBinary"
}

func formatExpression(e *Expression, ctxPrec int) string {
	result := formatExprInner(e)
	if needsParens(e, ctxPrec) {
		return "(" + result + ")"
	}
	return result
}

func needsParens(e *Expression, ctxPrec int) bool {
	return exprPrec(e) < ctxPrec
}

func exprPrec(e *Expression) int {
	switch e.Kind {
	case KindConcat:
		return precConcat
	case KindBinary:
		b := e.payload.(*BinaryExpr)
		switch b.Op {
		case LogicalOr:
			return precLogicalOr
		case LogicalAnd:
			return precLogicalAnd
		case Equal, NotEqual:
			return precEqual
		case GreaterEqual, Greater, LessEqual, Less:
			return precCompare
		case Add, Sub:
			return precAddSub
		case Mul, Div, Rem:
			return precMulDiv
		}
	case KindUnary:
		u := e.payload.(*UnaryExpr)
		switch u.Op {
		case LogicalNot, Negate:
			return precPrefix
		}
	case KindPattern:
		return precPattern
	case KindMethodCall:
		return precMethod
	default:
		return precPrimary
	}
	return precPrimary
}

func formatExprInner(e *Expression) string {
	switch e.Kind {
	case KindIdentifier:
		return e.payload.(*IdentifierExpr).Name
	case KindBoolean:
		if e.payload.(*BooleanExpr).Value {
			return "true"
		}
		return "false"
	case KindInteger:
		return fmt.Sprintf("%d", e.payload.(*IntegerExpr).Value)
	case KindString:
		return fmt.Sprintf("%q", e.payload.(*StringExpr).Value)
	case KindPattern:
		p := e.payload.(*PatternExpr)
		return fmt.Sprintf("%s:%s", p.Name, formatExpression(p.Value, precPrimary))
	case KindUnary:
		u := e.payload.(*UnaryExpr)
		return formatUnaryInner(u)
	case KindBinary:
		b := e.payload.(*BinaryExpr)
		return formatBinaryInner(b)
	case KindConcat:
		u := e.payload.(*ConcatExpr)
		parts := make([]string, len(u.Nodes))
		for i, n := range u.Nodes {
			parts[i] = formatExpression(n, precPrimary)
		}
		return strings.Join(parts, " ++ ")
	case KindFunctionCall:
		f := e.payload.(*FunctionCallExpr)
		return formatFunctionCall(f)
	case KindMethodCall:
		m := e.payload.(*MethodCallExpr)
		obj := formatExpression(m.Object, precMethod+1)
		method := formatFunctionCall(m.Function)
		return obj + "." + method
	case KindLambda:
		l := e.payload.(*LambdaExpr)
		params := strings.Join(l.Params, ", ")
		body := formatExpression(l.Body, precPrimary)
		if len(params) == 0 {
			return fmt.Sprintf("|| %s", body)
		}
		return fmt.Sprintf("|%s| %s", params, body)
	}
	return ""
}

func formatUnaryInner(u *UnaryExpr) string {
	switch u.Op {
	case LogicalNot:
		return fmt.Sprintf("!%s", formatExpression(u.Arg, precPrefix+1))
	case Negate:
		return fmt.Sprintf("-%s", formatExpression(u.Arg, precPrefix+1))
	}
	return ""
}

func formatBinaryInner(b *BinaryExpr) string {
	var opStr string
	var prec int
	switch b.Op {
	case LogicalOr:
		opStr = " || "
		prec = precLogicalOr
	case LogicalAnd:
		opStr = " && "
		prec = precLogicalAnd
	case Equal:
		opStr = " == "
		prec = precEqual
	case NotEqual:
		opStr = " != "
		prec = precEqual
	case GreaterEqual:
		opStr = " >= "
		prec = precCompare
	case Greater:
		opStr = " > "
		prec = precCompare
	case LessEqual:
		opStr = " <= "
		prec = precCompare
	case Less:
		opStr = " < "
		prec = precCompare
	case Add:
		opStr = " + "
		prec = precAddSub
	case Sub:
		opStr = " - "
		prec = precAddSub
	case Mul:
		opStr = " * "
		prec = precMulDiv
	case Div:
		opStr = " / "
		prec = precMulDiv
	case Rem:
		opStr = " % "
		prec = precMulDiv
	}
	lhs := formatExpression(b.LHS, prec)
	rhs := formatExpression(b.RHS, prec+1)
	return lhs + opStr + rhs
}

func formatFunctionCall(f *FunctionCallExpr) string {
	args := make([]string, 0, len(f.Args)+len(f.KeywordArgs))
	for _, a := range f.Args {
		args = append(args, formatExpression(a, precPrimary))
	}
	for _, kw := range f.KeywordArgs {
		args = append(args, fmt.Sprintf("%s=%s", kw.Name, formatExpression(kw.Value, precPrimary)))
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", "))
}