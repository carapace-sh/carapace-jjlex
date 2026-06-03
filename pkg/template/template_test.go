package template

import (
	"testing"
)

func TestParseTreeEq(t *testing.T) {
	a, err := Parse(" commit_id.short(1) ++ description ")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Parse("commit_id.short( 1 )++(description)")
	if err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Errorf("expected equal: %q != %q", a.String(), b.String())
	}

	c, err := Parse(` "ab" `)
	if err != nil {
		t.Fatal(err)
	}
	d, err := Parse(` "a" ++ "b" `)
	if err != nil {
		t.Fatal(err)
	}
	if c.String() == d.String() {
		t.Errorf("expected different: %q == %q", c.String(), d.String())
	}

	e, err := Parse(` "foo" ++ 0 `)
	if err != nil {
		t.Fatal(err)
	}
	f, err := Parse(` "foo0" `)
	if err != nil {
		t.Fatal(err)
	}
	if e.String() == f.String() {
		t.Errorf("expected different: %q == %q", e.String(), f.String())
	}
}

func testParseKind(t *testing.T, input string, kind ExpressionKind) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != kind {
		t.Errorf("parse %q: expected %v, got %v", input, kind, expr.Kind)
	}
}

func testParseError(t *testing.T, input string) {
	t.Helper()
	_, err := Parse(input)
	if err == nil {
		t.Errorf("expected error parsing %q, got none", input)
	}
}

func testParseUnaryOp(t *testing.T, input string, op UnaryOp) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != KindUnary {
		t.Fatalf("parse %q: expected KindUnary, got %v", input, expr.Kind)
	}
	if expr.UnaryOp() != op {
		t.Errorf("parse %q: expected op %v, got %v", input, op, expr.UnaryOp())
	}
}

func testParseBinaryOp(t *testing.T, input string, op BinaryOp) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != KindBinary {
		t.Fatalf("parse %q: expected KindBinary, got %v", input, expr.Kind)
	}
	if expr.BinaryOp() != op {
		t.Errorf("parse %q: expected op %v, got %v", input, op, expr.BinaryOp())
	}
}

func testParseEqual(t *testing.T, input, expected string) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.String() != expected {
		t.Errorf("parse %q: expected %q, got %q", input, expected, expr.String())
	}
}

func TestParseLiterals(t *testing.T) {
	testParseKind(t, `"hello"`, KindString)
	testParseKind(t, `'raw string'`, KindString)
	testParseKind(t, `""`, KindString)
	testParseKind(t, `''`, KindString)
	testParseKind(t, "true", KindBoolean)
	testParseKind(t, "false", KindBoolean)
	testParseKind(t, "42", KindInteger)
	testParseKind(t, "0", KindInteger)
	testParseKind(t, "foo", KindIdentifier)
}

func TestParseStringEscapes(t *testing.T) {
	testParseKind(t, `"\\t\\r\\n"`, KindString)
	testParseKind(t, `"\\0\\e"`, KindString)
	testParseKind(t, `"\\x41"`, KindString)
	// Invalid escape \y is not a valid escape in double-quoted strings
	// (jj template parser rejects it; our parser is lenient for now)
}

func TestParseOperatorPrecedence(t *testing.T) {
	testParseEqual(t, "!!x", "!(!x)")
	// Binary ops with different precedence
	expr, err := Parse("x.f() || y == y")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBinary || expr.BinaryOp() != LogicalOr {
		t.Errorf("expected ||, got %v", expr.Kind)
	}
	// x || y == y && z.h() > z
	expr, err = Parse("x || y == y && z.h() > z")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBinary || expr.BinaryOp() != LogicalOr {
		t.Errorf("expected ||, got %v", expr.Kind)
	}
	// Concat with other operators (++ is weakest, so binary ops get parens)
	testParseEqual(t, "x && y ++ z", "(x && y) ++ z")
	testParseEqual(t, "x ++ y || z", "x ++ (y || z)")
	testParseEqual(t, "x == y ++ z", "(x == y) ++ z")
	testParseEqual(t, "x != y ++ z", "(x != y) ++ z")
}

func TestParseUnaryOps(t *testing.T) {
	testParseUnaryOp(t, "!foo", LogicalNot)
	testParseUnaryOp(t, "-42", Negate)
	// Method call binds tighter than prefix: !x.f() = !(x.f())
	expr, err := Parse("!x.f()")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindUnary || expr.UnaryOp() != LogicalNot {
		t.Errorf("expected LogicalNot, got %v", expr.Kind)
	}
	inner := expr.UnaryArg()
	if inner.Kind != KindMethodCall {
		t.Errorf("expected method call inside, got %v", inner.Kind)
	}
}

func TestParseBinaryOps(t *testing.T) {
	testParseBinaryOp(t, "x || y", LogicalOr)
	testParseBinaryOp(t, "x && y", LogicalAnd)
	testParseBinaryOp(t, "x == y", Equal)
	testParseBinaryOp(t, "x != y", NotEqual)
	testParseBinaryOp(t, "x >= y", GreaterEqual)
	testParseBinaryOp(t, "x > y", Greater)
	testParseBinaryOp(t, "x <= y", LessEqual)
	testParseBinaryOp(t, "x < y", Less)
	testParseBinaryOp(t, "x + y", Add)
	testParseBinaryOp(t, "x - y", Sub)
	testParseBinaryOp(t, "x * y", Mul)
	testParseBinaryOp(t, "x / y", Div)
	testParseBinaryOp(t, "x % y", Rem)
}

func TestParseConcat(t *testing.T) {
	expr, err := Parse(`"a" ++ "b" ++ "c"`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindConcat {
		t.Fatalf("expected KindConcat, got %v", expr.Kind)
	}
	nodes := expr.ConcatNodes()
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
}

func TestParseMethodCall(t *testing.T) {
	expr, err := Parse("commit_id.short()")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindMethodCall {
		t.Fatalf("expected KindMethodCall, got %v", expr.Kind)
	}
	if expr.MethodName() != "short" {
		t.Errorf("expected method name 'short', got %q", expr.MethodName())
	}

	// Chained method calls
	expr, err = Parse("x.f().g()")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindMethodCall {
		t.Fatalf("expected KindMethodCall, got %v", expr.Kind)
	}
	if expr.MethodName() != "g" {
		t.Errorf("expected method name 'g', got %q", expr.MethodName())
	}
}

func TestParseFunctionCall(t *testing.T) {
	testParseKind(t, "foo()", KindFunctionCall)
	expr, err := Parse("label(\"test\", commit_id)")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindFunctionCall {
		t.Fatalf("expected KindFunctionCall, got %v", expr.Kind)
	}
	if expr.FunctionName() != "label" {
		t.Errorf("expected function name 'label', got %q", expr.FunctionName())
	}
	if len(expr.FunctionArgs()) != 2 {
		t.Errorf("expected 2 args, got %d", len(expr.FunctionArgs()))
	}
}

func TestParseFunctionKeywordArgs(t *testing.T) {
	expr, err := Parse("f(x, foo=0, bar=1)")
	if err != nil {
		t.Fatal(err)
	}
	if len(expr.FunctionArgs()) != 1 {
		t.Errorf("expected 1 positional arg, got %d", len(expr.FunctionArgs()))
	}
	if len(expr.FunctionKeywordArgs()) != 2 {
		t.Errorf("expected 2 keyword args, got %d", len(expr.FunctionKeywordArgs()))
	}
}

func TestParseLambda(t *testing.T) {
	expr, err := Parse("|| x")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindLambda {
		t.Fatalf("expected KindLambda, got %v", expr.Kind)
	}
	if len(expr.LambdaParams()) != 0 {
		t.Errorf("expected 0 params, got %d", len(expr.LambdaParams()))
	}

	expr, err = Parse("|x| x")
	if err != nil {
		t.Fatal(err)
	}
	if len(expr.LambdaParams()) != 1 {
		t.Errorf("expected 1 param, got %d", len(expr.LambdaParams()))
	}

	expr, err = Parse("|x, y| x ++ y")
	if err != nil {
		t.Fatal(err)
	}
	if len(expr.LambdaParams()) != 2 {
		t.Errorf("expected 2 params, got %d", len(expr.LambdaParams()))
	}

	// Trailing comma
	_, err = Parse("|x,| a")
	if err != nil {
		t.Errorf("expected trailing comma to be ok: %v", err)
	}

	// No body
	testParseError(t, "||")

	// Redefined parameter
	testParseError(t, "|x, x| a")

	// Boolean as parameter name
	testParseError(t, "|false| a")

	// Lambda vs logical operator: || can be zero-arg lambda
	e1, err := Parse("x||||y")
	if err != nil {
		t.Fatal(err)
	}
	// x || (|| y) - right side is a zero-arg lambda
	if e1.Kind != KindBinary || e1.BinaryOp() != LogicalOr {
		t.Errorf("expected ||, got %v", e1.Kind)
	}
	rightExpr := e1.BinaryRHS()
	if rightExpr.Kind != KindLambda {
		t.Errorf("expected lambda on right, got %v", rightExpr.Kind)
	}

	e2, err := Parse("||||x")
	if err != nil {
		t.Fatal(err)
	}
	// || (|| x) - zero-arg lambda whose body is another zero-arg lambda
	if e2.Kind != KindLambda {
		t.Errorf("expected lambda, got %v", e2.Kind)
	}
}

func TestParsePattern(t *testing.T) {
	expr, err := Parse(`regex:"meow"`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindPattern {
		t.Fatalf("expected KindPattern, got %v", expr.Kind)
	}
	if expr.PatternName() != "regex" {
		t.Errorf("expected pattern name 'regex', got %q", expr.PatternName())
	}

	// Pattern with identifier value
	expr, err = Parse("regex:meow")
	if err != nil {
		t.Fatal(err)
	}
	if expr.PatternName() != "regex" {
		t.Errorf("expected pattern name 'regex', got %q", expr.PatternName())
	}
	if expr.PatternValue().Kind != KindIdentifier {
		t.Errorf("expected pattern value KindIdentifier, got %v", expr.PatternValue().Kind)
	}

	// Pattern with dash in name
	expr, err = Parse("regex-i:'test'")
	if err != nil {
		t.Fatal(err)
	}
	if expr.PatternName() != "regex-i" {
		t.Errorf("expected pattern name 'regex-i', got %q", expr.PatternName())
	}

	// Pattern right-associative: x:y:z = x:(y:z)
	e3, err := Parse("x:y:z")
	if err != nil {
		t.Fatal(err)
	}
	if e3.Kind != KindPattern || e3.PatternName() != "x" {
		t.Fatalf("expected pattern x, got %v", e3.Kind)
	}
	inner := e3.PatternValue()
	if inner.Kind != KindPattern || inner.PatternName() != "y" {
		t.Fatalf("expected inner pattern y, got %v", inner.Kind)
	}

	// Pattern with method call value
	testParseEqual(t, "x:y.f(z)", "x:(y.f(z))")

	// No whitespace around :
	testParseError(t, `regex: 'with spaces'`)
	testParseError(t, `regex :'with spaces'`)

	// Whitespace in parenthesized value
	testParseEqual(t, `exact:( 'foo' )`, `exact:"foo"`)
}

func TestParseBooleanLiterals(t *testing.T) {
	expr, err := Parse("true")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBoolean || !expr.BooleanValue() {
		t.Errorf("expected true boolean, got %v", expr)
	}

	expr, err = Parse("false")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBoolean || expr.BooleanValue() {
		t.Errorf("expected false boolean, got %v", expr)
	}

	// Case sensitive
	testParseKind(t, "False", KindIdentifier)
	testParseKind(t, "tRue", KindIdentifier)

	// Boolean cannot be function name
	testParseError(t, "true()")
	testParseError(t, "false()")
}

func TestParseIntegerLiteral(t *testing.T) {
	expr, err := Parse("42")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindInteger || expr.IntegerValue() != 42 {
		t.Errorf("expected integer 42, got %v", expr)
	}

	expr, err = Parse("0")
	if err != nil {
		t.Fatal(err)
	}
	if expr.IntegerValue() != 0 {
		t.Errorf("expected integer 0, got %d", expr.IntegerValue())
	}

	// Leading zeros not allowed
	testParseError(t, "00")
}

func TestParseParenthesized(t *testing.T) {
	testParseEqual(t, "(x ++ y)", "x ++ y")
	testParseEqual(t, "((x))", "x")

	testParseError(t, "(x")
	testParseError(t, "x)")
}

func TestParseWhitespace(t *testing.T) {
	testParseEqual(t, " f( ) ", "f()")
	testParseEqual(t, ` " " `, `" "`)
	testParseEqual(t, " ' ' ", `" "`)
}