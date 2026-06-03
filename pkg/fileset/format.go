package fileset

import (
	"fmt"
	"strings"
)

// Precedence levels (higher = tighter binding)
const (
	precUnion     = 1
	precInterDiff = 2
	precNegate    = 3
	precPrimary   = 4
)

func (e *Expression) String() string {
	return formatExpression(e, precUnion)
}

func (op UnaryOp) GoString() string {
	switch op {
	case Negate:
		return "fileset.Negate"
	}
	return "fileset.UnknownUnary"
}

func (op BinaryOp) GoString() string {
	switch op {
	case Intersection:
		return "fileset.Intersection"
	case Difference:
		return "fileset.Difference"
	}
	return "fileset.UnknownBinary"
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
	case KindUnionAll:
		return precUnion
	case KindBinary:
		return precInterDiff
	case KindUnary:
		return precNegate
	default:
		return precPrimary
	}
}

func formatExprInner(e *Expression) string {
	switch e.Kind {
	case KindIdentifier:
		return e.payload.(*IdentifierExpr).Name
	case KindString:
		return fmt.Sprintf("%q", e.payload.(*StringExpr).Value)
	case KindBareString:
		return e.payload.(*BareStringExpr).Value
	case KindPattern:
		p := e.payload.(*PatternExpr)
		return fmt.Sprintf("%s:%s", p.Name, formatExpression(p.Value, precPrimary))
	case KindBareStringPattern:
		p := e.payload.(*PatternExpr)
		return fmt.Sprintf("%s:%s", p.Name, p.Value.BareStringValue())
	case KindUnary:
		u := e.payload.(*UnaryExpr)
		return fmt.Sprintf("~%s", formatExpression(u.Arg, precNegate+1))
	case KindBinary:
		b := e.payload.(*BinaryExpr)
		return formatBinaryInner(b)
	case KindUnionAll:
		u := e.payload.(*UnionAllExpr)
		parts := make([]string, len(u.Nodes))
		for i, n := range u.Nodes {
			parts[i] = formatExpression(n, precUnion+1)
		}
		return strings.Join(parts, "|")
	case KindFunctionCall:
		f := e.payload.(*FunctionCallExpr)
		return formatFunctionCall(f)
	}
	return ""
}

func formatBinaryInner(b *BinaryExpr) string {
	var opStr string
	switch b.Op {
	case Intersection:
		opStr = "&"
	case Difference:
		opStr = "~"
	}
	lhs := formatExpression(b.LHS, precInterDiff)
	rhs := formatExpression(b.RHS, precInterDiff+1) // left-assoc
	return fmt.Sprintf("%s%s%s", lhs, opStr, rhs)
}

func formatFunctionCall(f *FunctionCallExpr) string {
	args := make([]string, 0, len(f.Args))
	for _, a := range f.Args {
		args = append(args, formatExpression(a, precPrimary))
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ","))
}