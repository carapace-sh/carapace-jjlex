package revset

import (
	"fmt"
	"strings"
)

// Precedence levels (higher = tighter binding)
const (
	precUnion      = 1
	precInterDiff  = 2
	precNegate     = 3
	precRange      = 4
	precPrefixRng  = 5
	precPostfixRng = 6
	precPostfix    = 7
	precPrimary    = 8
)

func (e *Expression) String() string {
	return formatExprInner(e)
}

func (op UnaryOp) GoString() string {
	switch op {
	case Negate:
		return "revset.Negate"
	case DagRangePre:
		return "revset.DagRangePre"
	case DagRangePost:
		return "revset.DagRangePost"
	case RangePre:
		return "revset.RangePre"
	case RangePost:
		return "revset.RangePost"
	case Parents:
		return "revset.Parents"
	case Children:
		return "revset.Children"
	}
	return "revset.UnknownUnary"
}

func (op BinaryOp) GoString() string {
	switch op {
	case Intersection:
		return "revset.Intersection"
	case Difference:
		return "revset.Difference"
	case DagRange:
		return "revset.DagRange"
	case Range:
		return "revset.Range"
	}
	return "revset.UnknownBinary"
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
		b := e.payload.(*BinaryExpr)
		switch b.Op {
		case Intersection:
			return precInterDiff
		case Difference:
			return precInterDiff
		case DagRange:
			return precRange
		case Range:
			return precRange
		}
	case KindUnary:
		u := e.payload.(*UnaryExpr)
		switch u.Op {
		case Negate:
			return precNegate
		case DagRangePre, RangePre:
			return precPrefixRng
		case DagRangePost, RangePost:
			return precPostfixRng
		case Parents, Children:
			return precPostfix
		}
	default:
		return precPrimary
	}
	return precPrimary
}

func formatExprInner(e *Expression) string {
	switch e.Kind {
	case KindIdentifier:
		return e.payload.(*IdentifierExpr).Name
	case KindString:
		return fmt.Sprintf("%q", e.payload.(*StringExpr).Value)
	case KindPattern:
		p := e.payload.(*PatternExpr)
		return fmt.Sprintf("%s:%s", p.Name, formatExpression(p.Value, precPostfix))
	case KindRemoteSymbol:
		r := e.payload.(*RemoteSymbolExpr)
		return fmt.Sprintf("%s@%s", QuoteIfNeeded(r.Name), QuoteIfNeeded(r.Remote))
	case KindAtWorkspace:
		a := e.payload.(*AtWorkspaceExpr)
		return fmt.Sprintf("%s@", QuoteIfNeeded(a.Name))
	case KindAtCurrentWorkspace:
		return "@"
	case KindDagRangeAll:
		return "::"
	case KindRangeAll:
		return ".."
	case KindUnary:
		u := e.payload.(*UnaryExpr)
		return formatUnaryInner(u)
	case KindBinary:
		b := e.payload.(*BinaryExpr)
		return formatBinaryInner(b)
	case KindUnionAll:
		u := e.payload.(*UnionAllExpr)
		parts := make([]string, len(u.Nodes))
		for i, n := range u.Nodes {
			parts[i] = formatExpression(n, precUnion+1) // left-assoc: RHS needs higher prec
		}
		return strings.Join(parts, "|")
	case KindFunctionCall:
		f := e.payload.(*FunctionCallExpr)
		return formatFunctionCall(f)
	}
	return ""
}

func formatUnaryInner(u *UnaryExpr) string {
	switch u.Op {
	case Negate:
		return fmt.Sprintf("~%s", formatExpression(u.Arg, precNegate+1))
	case DagRangePre:
		return fmt.Sprintf("::%s", formatExpression(u.Arg, precPostfix))
	case DagRangePost:
		return fmt.Sprintf("%s::", formatExpression(u.Arg, precPostfixRng+1))
	case RangePre:
		return fmt.Sprintf("..%s", formatExpression(u.Arg, precPostfix))
	case RangePost:
		return fmt.Sprintf("%s..", formatExpression(u.Arg, precPostfixRng+1))
	case Parents:
		return fmt.Sprintf("%s-", formatExpression(u.Arg, precPostfix+1))
	case Children:
		return fmt.Sprintf("%s+", formatExpression(u.Arg, precPostfix+1))
	}
	return ""
}

func formatBinaryInner(b *BinaryExpr) string {
	var opStr string
	var prec int
	switch b.Op {
	case Intersection:
		opStr = "&"
		prec = precInterDiff
	case Difference:
		opStr = "~"
		prec = precInterDiff
	case DagRange:
		opStr = "::"
		prec = precRange
	case Range:
		opStr = ".."
		prec = precRange
	}
	lhs := formatExpression(b.LHS, prec)
	rhs := formatExpression(b.RHS, prec+1) // left-assoc
	return fmt.Sprintf("%s%s%s", lhs, opStr, rhs)
}

func formatFunctionCall(f *FunctionCallExpr) string {
	args := make([]string, 0, len(f.Args)+len(f.KeywordArgs))
	for _, a := range f.Args {
		args = append(args, formatExpression(a, precPrimary))
	}
	for _, kw := range f.KeywordArgs {
		args = append(args, fmt.Sprintf("%s=%s", kw.Name, formatExpression(kw.Value, precPrimary)))
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ","))
}

func QuoteIfNeeded(s string) string {
	if IsSimpleIdentifier(s) {
		return s
	}
	return fmt.Sprintf("%q", s)
}

func IsSimpleIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Match the full revset identifier grammar:
	//   identifier = identifier_part ~ (("." | "-"+ | "+") ~ identifier_part)*
	// Each identifier_part is one or more characters from (XID_CONTINUE | "_" | "*" | "/").
	pos := 0
	runes := []rune(s)

	// First part: must be a valid identifier_part
	n := scanIdentifierPart(runes)
	if n == 0 {
		return false
	}
	pos += n

	for pos < len(runes) {
		// Connector: ".", one or more "-", or "+"
		connectorLen := scanConnector(runes[pos:])
		if connectorLen == 0 {
			return false
		}
		pos += connectorLen

		// After connector: must be another identifier_part
		n = scanIdentifierPart(runes[pos:])
		if n == 0 {
			return false
		}
		pos += n
	}
	return pos == len(runes)
}

// scanIdentifierPart scans one or more characters matching isIdentifierPart.
// Returns the number of runes consumed.
func scanIdentifierPart(runes []rune) int {
	n := 0
	for n < len(runes) && isIdentifierPart(runes[n]) {
		n++
	}
	return n
}

// scanConnector scans a valid identifier connector: ".", one or more "-", or "+".
// Returns the number of runes consumed.
func scanConnector(runes []rune) int {
	if len(runes) == 0 {
		return 0
	}
	switch runes[0] {
	case '.':
		return 1
	case '-':
		// One or more dashes
		n := 0
		for n < len(runes) && runes[n] == '-' {
			n++
		}
		return n
	case '+':
		return 1
	}
	return 0
}
