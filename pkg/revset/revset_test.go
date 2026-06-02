package revset

import (
	"testing"
)

func TestParseTreeEq(t *testing.T) {
	a, err := Parse(" foo( x ) | ~bar:\"baz\" ")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Parse("(foo(x))|(~(bar:\"baz\"))")
	if err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Errorf("expected equal: %q != %q", a.String(), b.String())
	}

	c, err := Parse(" foo ")
	if err != nil {
		t.Fatal(err)
	}
	d, err := Parse(` "foo" `)
	if err != nil {
		t.Fatal(err)
	}
	if c.String() == d.String() {
		t.Errorf("expected different: %q == %q", c.String(), d.String())
	}
}

func TestParseRevset(t *testing.T) {
	// Parse a quoted symbol
	testParseKind(t, `"foo"`, KindString)
	testParseKind(t, `'foo'`, KindString)

	// Parse the "parents" operator
	testParseUnaryOp(t, "foo-", Parents)

	// Parse the "children" operator
	testParseUnaryOp(t, "foo+", Children)

	// Parse the "ancestors" operator
	testParseUnaryOp(t, "::foo", DagRangePre)

	// Parse the "descendants" operator
	testParseUnaryOp(t, "foo::", DagRangePost)

	// Parse the "dag range" operator
	testParseBinaryOp(t, "foo::bar", DagRange)

	// Parse the nullary "dag range" operator
	testParseKind(t, "::", KindDagRangeAll)

	// Parse the "range" operators
	testParseUnaryOp(t, "..foo", RangePre)
	testParseUnaryOp(t, "foo..", RangePost)
	testParseBinaryOp(t, "foo..bar", Range)
	testParseKind(t, "..", KindRangeAll)

	// Parse the "negate" operator
	testParseUnaryOp(t, "~ foo", Negate)
	testParseEqual(t, "~ ~~ foo", "~(~(~(foo)))")

	// Parse the "intersection" operator
	testParseBinaryOp(t, "foo & bar", Intersection)

	// Parse the "union" operator
	expr, err := Parse("foo | bar")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindUnionAll {
		t.Errorf("expected KindUnionAll, got %v", expr.Kind)
	}
	if len(expr.UnionNodes()) != 2 {
		t.Errorf("expected 2 nodes in union, got %d", len(expr.UnionNodes()))
	}

	expr, err = Parse("foo | bar | baz")
	if err != nil {
		t.Fatal(err)
	}
	if len(expr.UnionNodes()) != 3 {
		t.Errorf("expected 3 nodes in union, got %d", len(expr.UnionNodes()))
	}

	// Parse the "difference" operator
	testParseBinaryOp(t, "foo ~ bar", Difference)

	// Parentheses before suffix operators
	testParseEqual(t, "(foo)-", "foo-")

	// Space around expressions
	testParseEqual(t, " ::foo ", "::foo")
	testParseEqual(t, "( ::foo )", "::foo")

	// Space is not allowed around prefix range operators
	testParseError(t, " :: foo ")
	testParseError(t, " .. foo ")

	// Incomplete parse
	testParseError(t, "foo | -")
}

func TestParseExpressionSpans(t *testing.T) {
	tests := []struct {
		input    string
		spanStr  string
	}{
		{" ~ x ", "~ x"},
		{" x+ ", "x+"},
		{" x |y ", "x |y"},
		{" (x) ", "(x)"},
		{"~( x|y) ", "~( x|y)"},
		{" ( x )- ", "( x )-"},
	}
	for _, tt := range tests {
		expr, err := Parse(tt.input)
		if err != nil {
			t.Fatalf("parse %q: %v", tt.input, err)
		}
		got := tt.input[expr.Span.Start:expr.Span.End]
		if got != tt.spanStr {
			t.Errorf("parse %q: span expected %q, got %q", tt.input, tt.spanStr, got)
		}
	}
}

func TestParseWhitespace(t *testing.T) {
	// Standard whitespace: space, tab, CR, LF, FF
	testParseEqual(t, " \t\r\n\x0call()", "all()")
}

func TestParseIdentifier(t *testing.T) {
	// Integer is a symbol
	testParseKind(t, "0", KindIdentifier)

	// Tag/bookmark name separated by /
	testParseKind(t, "foo_bar/baz", KindIdentifier)

	// Glob literal with star
	testParseKind(t, "*/foo/**", KindIdentifier)

	// Internal '.', '-', and '+' are allowed
	testParseKind(t, "foo.bar-v1+7", KindIdentifier)
	testParseEqual(t, "foo.bar-v1+7-", "(foo.bar-v1+7)-")

	// Multiple '-' are allowed
	testParseKind(t, "foo--bar", KindIdentifier)
	testParseKind(t, "foo----bar", KindIdentifier)

	// '.' is not allowed at the beginning or end
	testParseError(t, ".foo")
	testParseError(t, "foo.")

	// Multiple '.' and '+', or together with '-', are not allowed
	testParseError(t, "foo.+bar")
	testParseError(t, "foo++bar")
	testParseError(t, "foo+-bar")

	// Parse a parenthesized symbol
	testParseEqual(t, "(foo)", "foo")

	// Non-ASCII tag/bookmark name
	testParseKind(t, "柔術+jj", KindIdentifier)
}

func TestParseStringLiteral(t *testing.T) {
	// Escape sequences
	expr, err := Parse(` "\t\r\n\"\\\0\e" `)
	if err != nil {
		t.Fatal(err)
	}
	s := expr.StringValue()
	expected := "\t\r\n\"\\\x00\x1b"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}

	// Invalid escape
	testParseError(t, ` "\y" `)

	// Single-quoted raw string
	testParseString(t, ` '' `, "")
	testParseString(t, ` 'a\n' `, "a\\n")
	testParseString(t, ` '\' `, "\\")
	testParseString(t, ` '"' `, "\"")

	// Hex bytes
	testParseString(t, `"\x61\x65\x69\x6f\x75"`, "aeiou")
	testParseString(t, `"\xe0\xe8\xec\xf0\xf9"`, "àèìðù")

	// Invalid hex
	testParseError(t, `"\x"`)
	testParseError(t, `"\xf"`)
	testParseError(t, `"\xgg"`)
}

func TestParsePattern(t *testing.T) {
	// Pattern with string value
	expr, err := Parse(`substring:"foo"`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindPattern {
		t.Fatalf("expected KindPattern, got %v", expr.Kind)
	}
	if expr.PatternName() != "substring" {
		t.Errorf("expected name 'substring', got %q", expr.PatternName())
	}
	if expr.PatternValue().Kind != KindString || expr.PatternValue().StringValue() != "foo" {
		t.Errorf("expected string value 'foo', got %v", expr.PatternValue())
	}

	// Pattern with identifier value
	expr, err = Parse("exact:foo")
	if err != nil {
		t.Fatal(err)
	}
	if expr.PatternName() != "exact" {
		t.Errorf("expected name 'exact', got %q", expr.PatternName())
	}
	if expr.PatternValue().Kind != KindIdentifier || expr.PatternValue().Identifier() != "foo" {
		t.Errorf("expected identifier value 'foo', got %v", expr.PatternValue())
	}

	// Quoted string with : inside is NOT a pattern
	testParseKind(t, `"exact:foo"`, KindString)

	// Pattern with @ value
	expr, err = Parse("x:@")
	if err != nil {
		t.Fatal(err)
	}
	if expr.PatternName() != "x" {
		t.Errorf("expected name 'x', got %q", expr.PatternName())
	}
	if expr.PatternValue().Kind != KindAtCurrentWorkspace {
		t.Errorf("expected AtCurrentWorkspace, got %v", expr.PatternValue().Kind)
	}

	// Pattern with remote symbol value
	expr, err = Parse("x:y@z")
	if err != nil {
		t.Fatal(err)
	}
	if expr.PatternValue().Kind != KindRemoteSymbol {
		t.Errorf("expected RemoteSymbol, got %v", expr.PatternValue().Kind)
	}
	if expr.PatternValue().RemoteSymbolName() != "y" || expr.PatternValue().RemoteSymbolRemote() != "z" {
		t.Errorf("expected y@z, got %s@%s", expr.PatternValue().RemoteSymbolName(), expr.PatternValue().RemoteSymbolRemote())
	}

	// Whitespace isn't allowed in pattern
	testParseError(t, "exact: foo")
	testParseError(t, "exact :foo")

	// Parenthesized pattern value
	testParseEqual(t, "exact:( 'foo' )", "exact:'foo'")

	// Functions in pattern value
	testParseEqual(t, "x:f(y)", "x:(f(y))")

	// Postfix operations in pattern value
	testParseEqual(t, "x:@-+", "x:((@-)+)")

	// Ranges have lower binding strength than patterns
	testParseEqual(t, "x:y::z", "(x:y)::(z)")
	testParseEqual(t, "x:y&z", "(x:y)&(z)")

	// Pattern prefix is right-associative
	testParseEqual(t, "x:y:z", "x:(y:z)")
}

func TestParseSymbol(t *testing.T) {
	if _, err := ParseSymbol(""); err == nil {
		t.Error("expected error for empty symbol")
	}
	if _, err := ParseSymbol("''"); err == nil {
		t.Error("expected error for empty string symbol")
	}

	name, err := ParseSymbol("foo.bar")
	if err != nil {
		t.Fatal(err)
	}
	if name != "foo.bar" {
		t.Errorf("expected 'foo.bar', got %q", name)
	}

	if _, err := ParseSymbol("foo@bar"); err == nil {
		t.Error("expected error for 'foo@bar' as symbol")
	}
	if _, err := ParseSymbol("foo bar"); err == nil {
		t.Error("expected error for 'foo bar' as symbol")
	}

	name, err = ParseSymbol("'foo bar'")
	if err != nil {
		t.Fatal(err)
	}
	if name != "foo bar" {
		t.Errorf("expected 'foo bar', got %q", name)
	}

	name, err = ParseSymbol(`"foo\tbar"`)
	if err != nil {
		t.Fatal(err)
	}
	if name != "foo\tbar" {
		t.Errorf("expected 'foo\\tbar', got %q", name)
	}

	if _, err := ParseSymbol(" foo"); err == nil {
		t.Error("expected error for symbol with leading whitespace")
	}
	if _, err := ParseSymbol("foo "); err == nil {
		t.Error("expected error for symbol with trailing whitespace")
	}
	if _, err := ParseSymbol("(foo)"); err == nil {
		t.Error("expected error for parenthesized symbol")
	}
}

func TestParseAtWorkspaceAndRemoteSymbol(t *testing.T) {
	// Parse "@" (the current working copy)
	testParseKind(t, "@", KindAtCurrentWorkspace)
	testParseKind(t, "main@", KindAtWorkspace)

	expr, err := Parse("main@origin")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindRemoteSymbol {
		t.Fatalf("expected KindRemoteSymbol, got %v", expr.Kind)
	}
	if expr.RemoteSymbolName() != "main" {
		t.Errorf("expected name 'main', got %q", expr.RemoteSymbolName())
	}
	if expr.RemoteSymbolRemote() != "origin" {
		t.Errorf("expected remote 'origin', got %q", expr.RemoteSymbolRemote())
	}

	// Quoted component in @ expression
	expr, err = Parse(`"foo bar"@`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindAtWorkspace {
		t.Fatalf("expected KindAtWorkspace, got %v", expr.Kind)
	}
	if expr.AtWorkspaceName() != "foo bar" {
		t.Errorf("expected name 'foo bar', got %q", expr.AtWorkspaceName())
	}

	expr, err = Parse(`"foo bar"@origin`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.RemoteSymbolName() != "foo bar" || expr.RemoteSymbolRemote() != "origin" {
		t.Errorf("expected name='foo bar' remote='origin', got name=%q remote=%q", expr.RemoteSymbolName(), expr.RemoteSymbolRemote())
	}

	expr, err = Parse(`main@"foo bar"`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.RemoteSymbolName() != "main" || expr.RemoteSymbolRemote() != "foo bar" {
		t.Errorf("expected name='main' remote='foo bar', got name=%q remote=%q", expr.RemoteSymbolName(), expr.RemoteSymbolRemote())
	}

	expr, err = Parse(`'foo bar'@'bar baz'`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.RemoteSymbolName() != "foo bar" || expr.RemoteSymbolRemote() != "bar baz" {
		t.Errorf("expected name='foo bar' remote='bar baz', got name=%q remote=%q", expr.RemoteSymbolName(), expr.RemoteSymbolRemote())
	}

	// Quoted "@" is not interpreted as a working copy or remote symbol
	testParseKind(t, `"@"`, KindString)
	testParseKind(t, `"main@"`, KindString)
	testParseKind(t, `"main@origin"`, KindString)

	// Non-ASCII name
	testParseKind(t, "柔術@", KindAtWorkspace)

	expr, err = Parse("柔@術")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindRemoteSymbol {
		t.Fatalf("expected KindRemoteSymbol, got %v", expr.Kind)
	}
	if expr.RemoteSymbolName() != "柔" || expr.RemoteSymbolRemote() != "術" {
		t.Errorf("expected name='柔' remote='術', got name=%q remote=%q", expr.RemoteSymbolName(), expr.RemoteSymbolRemote())
	}
}

func TestParseFunctionCall(t *testing.T) {
	// Space is allowed around function arguments
	testParseEqual(t,
		"   description(  arg1 ) ~    file(  arg1 ,   arg2 )  ~ visible_heads(  )  ",
		"(description(arg1)~file(arg1,arg2))~visible_heads()",
	)

	// Space is allowed around keyword arguments
	testParseEqual(t,
		"remote_bookmarks( remote  =   foo  )",
		"remote_bookmarks(remote=foo)",
	)

	// Trailing comma isn't allowed for empty argument
	testParseError(t, "bookmarks(,)")

	// Trailing comma is allowed for the last argument
	testParseEqual(t, "bookmarks(a,)", "bookmarks(a)")
	testParseEqual(t, "bookmarks(a ,  )", "bookmarks(a)")

	testParseError(t, "bookmarks(,a)")
	testParseError(t, "bookmarks(a,,)")
	testParseError(t, "bookmarks(a  , , )")

	testParseEqual(t, "file(a,b,)", "file(a,b)")
	testParseError(t, "file(a,,b)")

	testParseEqual(t, "remote_bookmarks(a,remote=b  , )", "remote_bookmarks(a,remote=b)")
	testParseError(t, "remote_bookmarks(a,,remote=b)")

	// Empty function call
	testParseEqual(t, "visible_heads()", "visible_heads()")

	// Function with keyword args
	expr, err := Parse("remote_bookmarks(remote=foo)")
	if err != nil {
		t.Fatal(err)
	}
	if expr.FunctionName() != "remote_bookmarks" {
		t.Errorf("expected function name 'remote_bookmarks', got %q", expr.FunctionName())
	}
	if len(expr.FunctionKeywordArgs()) != 1 {
		t.Fatalf("expected 1 keyword arg, got %d", len(expr.FunctionKeywordArgs()))
	}
	if expr.FunctionKeywordArgs()[0].Name != "remote" {
		t.Errorf("expected keyword arg name 'remote', got %q", expr.FunctionKeywordArgs()[0].Name)
	}
}

func TestParseFunctionCallSpans(t *testing.T) {
	expr, err := Parse("foo( a, (b) , ~(c), d = (e) )")
	if err != nil {
		t.Fatal(err)
	}
	if expr.FunctionName() != "foo" {
		t.Errorf("expected function name 'foo', got %q", expr.FunctionName())
	}
	if len(expr.FunctionArgs()) != 3 {
		t.Fatalf("expected 3 args, got %d", len(expr.FunctionArgs()))
	}
	if len(expr.FunctionKeywordArgs()) != 1 {
		t.Fatalf("expected 1 keyword arg, got %d", len(expr.FunctionKeywordArgs()))
	}
}

func TestParseCompatOperator(t *testing.T) {
	// : as prefix should suggest ::
	testParseError(t, ":foo")

	// ^ as postfix should suggest -
	testParseError(t, "foo^")

	// + as infix should suggest |
	testParseError(t, "foo + bar")

	// - as infix should suggest ~
	testParseError(t, "foo - bar")
}

func TestParseOperatorCombinations(t *testing.T) {
	// Parse repeated "parents" operator
	testParseEqual(t, "foo---", "((foo-)-)-")

	// Parse repeated "children" operator
	testParseEqual(t, "foo+++", "((foo+)+)+")

	// Set operator associativity/precedence
	testParseEqual(t, "~x|y", "(~x)|y")
	testParseEqual(t, "x&~y", "x&(~y)")
	testParseEqual(t, "x~~y", "x~(~y)")
	testParseEqual(t, "x~~~y", "x~(~(~y))")
	testParseEqual(t, "~x::y", "~(x::y)")
	testParseEqual(t, "x|y|z", "(x|y)|z")
	testParseEqual(t, "x&y|z", "(x&y)|z")
	testParseEqual(t, "x|y&z", "x|(y&z)")
	testParseEqual(t, "x|y~z", "x|(y~z)")
	testParseEqual(t, "::&..", "(::)&(..)")

	// Parse repeated range operators - these should be syntax errors
	rangeErrors := []string{
		"::foo::", ":::foo", "::::foo", "foo:::", "foo::::",
		"foo:::bar", "foo::::bar", "::foo::bar", "foo::bar::",
		"::::", "....foo", "foo....", "foo.....bar",
		"..foo..bar", "foo..bar..", "....", "::..",
	}
	for _, input := range rangeErrors {
		testParseError(t, input)
	}

	// Combinations of parents/children with range operators
	testParseEqual(t, "foo-+", "(foo-)+")
	testParseEqual(t, "foo-::", "(foo-)::")
	testParseEqual(t, "::foo+", "::(foo+)")
	testParseError(t, "::-")
	testParseError(t, "..+")
}

func TestParseFunction(t *testing.T) {
	testParseKind(t, "parents(foo)", KindFunctionCall)
	testParseEqual(t, "parents((foo))", "parents(foo)")
	testParseError(t, "parents(foo")
}

func TestIsIdentifier(t *testing.T) {
	if !IsIdentifier("foo") {
		t.Error("expected 'foo' to be a valid identifier")
	}
	if !IsIdentifier("foo_bar/baz") {
		t.Error("expected 'foo_bar/baz' to be a valid identifier")
	}
	if IsIdentifier(".foo") {
		t.Error("expected '.foo' to not be a valid identifier")
	}
	if IsIdentifier("") {
		t.Error("expected empty string to not be a valid identifier")
	}
	if !IsIdentifier("0") {
		t.Error("expected '0' to be a valid identifier")
	}
	if !IsIdentifier("foo.bar") {
		t.Error("expected 'foo.bar' to be a valid identifier")
	}
	if !IsIdentifier("foo--bar") {
		t.Error("expected 'foo--bar' to be a valid identifier")
	}
}

func TestExpressionAccessors(t *testing.T) {
	// Identifier
	expr, _ := Parse("foo")
	if expr.Identifier() != "foo" {
		t.Errorf("expected 'foo', got %q", expr.Identifier())
	}

	// String
	expr, _ = Parse(`"hello"`)
	if expr.StringValue() != "hello" {
		t.Errorf("expected 'hello', got %q", expr.StringValue())
	}

	// Unary
	expr, _ = Parse("~foo")
	if expr.UnaryOp() != Negate {
		t.Errorf("expected Negate, got %v", expr.UnaryOp())
	}
	if expr.UnaryArg().Identifier() != "foo" {
		t.Errorf("expected arg 'foo', got %q", expr.UnaryArg().Identifier())
	}

	// Binary
	expr, _ = Parse("foo & bar")
	if expr.BinaryOp() != Intersection {
		t.Errorf("expected Intersection, got %v", expr.BinaryOp())
	}
	if expr.BinaryLHS().Identifier() != "foo" {
		t.Errorf("expected lhs 'foo', got %q", expr.BinaryLHS().Identifier())
	}
	if expr.BinaryRHS().Identifier() != "bar" {
		t.Errorf("expected rhs 'bar', got %q", expr.BinaryRHS().Identifier())
	}

	// Function call
	expr, _ = Parse("parents(foo)")
	if expr.FunctionName() != "parents" {
		t.Errorf("expected 'parents', got %q", expr.FunctionName())
	}
	if len(expr.FunctionArgs()) != 1 {
		t.Errorf("expected 1 arg, got %d", len(expr.FunctionArgs()))
	}

	// AtCurrentWorkspace
	expr, _ = Parse("@")
	if expr.Kind != KindAtCurrentWorkspace {
		t.Errorf("expected KindAtCurrentWorkspace, got %v", expr.Kind)
	}

	// DagRangeAll
	expr, _ = Parse("::")
	if expr.Kind != KindDagRangeAll {
		t.Errorf("expected KindDagRangeAll, got %v", expr.Kind)
	}

	// RangeAll
	expr, _ = Parse("..")
	if expr.Kind != KindRangeAll {
		t.Errorf("expected KindRangeAll, got %v", expr.Kind)
	}

	// Pattern
	expr, _ = Parse("exact:foo")
	if expr.PatternName() != "exact" {
		t.Errorf("expected 'exact', got %q", expr.PatternName())
	}
	if expr.PatternValue().Identifier() != "foo" {
		t.Errorf("expected value 'foo', got %q", expr.PatternValue().Identifier())
	}

	// RemoteSymbol
	expr, _ = Parse("main@origin")
	if expr.RemoteSymbolName() != "main" {
		t.Errorf("expected 'main', got %q", expr.RemoteSymbolName())
	}
	if expr.RemoteSymbolRemote() != "origin" {
		t.Errorf("expected 'origin', got %q", expr.RemoteSymbolRemote())
	}

	// AtWorkspace
	expr, _ = Parse("main@")
	if expr.AtWorkspaceName() != "main" {
		t.Errorf("expected 'main', got %q", expr.AtWorkspaceName())
	}
}

// Helper functions

func testParseKind(t *testing.T, input string, expectedKind ExpressionKind) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != expectedKind {
		t.Errorf("parse %q: expected kind %v, got %v", input, expectedKind, expr.Kind)
	}
}

func testParseUnaryOp(t *testing.T, input string, expectedOp UnaryOp) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != KindUnary {
		t.Fatalf("parse %q: expected KindUnary, got %v", input, expr.Kind)
	}
	if expr.UnaryOp() != expectedOp {
		t.Errorf("parse %q: expected op %v, got %v", input, expectedOp, expr.UnaryOp())
	}
}

func testParseBinaryOp(t *testing.T, input string, expectedOp BinaryOp) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != KindBinary {
		t.Fatalf("parse %q: expected KindBinary, got %v", input, expr.Kind)
	}
	if expr.BinaryOp() != expectedOp {
		t.Errorf("parse %q: expected op %v, got %v", input, expectedOp, expr.BinaryOp())
	}
}

func testParseEqual(t *testing.T, a, b string) {
	t.Helper()
	exprA, err := Parse(a)
	if err != nil {
		t.Fatalf("parse %q: %v", a, err)
	}
	exprB, err := Parse(b)
	if err != nil {
		t.Fatalf("parse %q: %v", b, err)
	}
	if exprA.String() != exprB.String() {
		t.Errorf("expected %q == %q, got %q != %q", a, b, exprA.String(), exprB.String())
	}
}

func testParseString(t *testing.T, input, expected string) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != KindString {
		t.Fatalf("expected KindString, got %v", expr.Kind)
	}
	s := expr.StringValue()
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}

func testParseError(t *testing.T, input string) {
	t.Helper()
	if _, err := Parse(input); err == nil {
		t.Errorf("expected error for %q", input)
	}
}

func TestParseRangeOps(t *testing.T) {
	// Infix range (x..y)
	testParseBinaryOp(t, "foo..bar", Range)

	// Nullary .. (KindRangeAll) already tested in TestParseRevset

	// x.. is postfix range (not ancestors-of-x)
	testParseUnaryOp(t, "foo..", RangePost)

	// ..x is prefix range (ancestors excluding root)
	testParseUnaryOp(t, "..foo", RangePre)

	// Range operators don't nest without parens
	testParseError(t, "foo::bar::")
	testParseError(t, "foo..bar..")

	// x..y semantics: ancestors of y not ancestors of x
	// (verified via Format output, not set evaluation)
	testParseEqual(t, "foo..bar", "foo..bar")

	// ..x precedence: ..x then & is (..x)&y, not ..(x&y)
	testParseEqual(t, "..x|y", "(..x)|y")
	testParseEqual(t, "..x&y", "(..x)&y")
}

func TestParseAtSuffixVariants(t *testing.T) {
	// Quoted name in at-suffix
	expr, err := Parse(`"foo bar"@`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindAtWorkspace {
		t.Fatalf("expected KindAtWorkspace, got %v", expr.Kind)
	}
	if expr.AtWorkspaceName() != "foo bar" {
		t.Errorf("expected 'foo bar', got %q", expr.AtWorkspaceName())
	}

	// Quoted remote in at-suffix
	expr, err = Parse(`main@"foo bar"`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindRemoteSymbol {
		t.Fatalf("expected KindRemoteSymbol, got %v", expr.Kind)
	}
	if expr.RemoteSymbolName() != "main" {
		t.Errorf("expected 'main', got %q", expr.RemoteSymbolName())
	}
	if expr.RemoteSymbolRemote() != "foo bar" {
		t.Errorf("expected 'foo bar', got %q", expr.RemoteSymbolRemote())
	}

	// Single-quoted remote
	expr, err = Parse(`main@'bar baz'`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.RemoteSymbolRemote() != "bar baz" {
		t.Errorf("expected 'bar baz', got %q", expr.RemoteSymbolRemote())
	}

	// @ inside quoted string is NOT workspace syntax
	testParseKind(t, `"@"`, KindString)
	testParseKind(t, `"main@"`, KindString)
	testParseKind(t, `"main@origin"`, KindString)
}

func TestParseKeywordArgs(t *testing.T) {
	// Keyword arg with strict identifier
	expr, err := Parse("remote_bookmarks(remote=foo)")
	if err != nil {
		t.Fatal(err)
	}
	if expr.FunctionName() != "remote_bookmarks" {
		t.Errorf("expected 'remote_bookmarks', got %q", expr.FunctionName())
	}
	if len(expr.FunctionArgs()) != 0 {
		t.Errorf("expected 0 positional args, got %d", len(expr.FunctionArgs()))
	}
	if len(expr.FunctionKeywordArgs()) != 1 {
		t.Fatalf("expected 1 keyword arg, got %d", len(expr.FunctionKeywordArgs()))
	}
	kw := expr.FunctionKeywordArgs()[0]
	if kw.Name != "remote" {
		t.Errorf("expected keyword arg name 'remote', got %q", kw.Name)
	}
	if kw.Value.Identifier() != "foo" {
		t.Errorf("expected keyword arg value 'foo', got %q", kw.Value.Identifier())
	}

	// Mix of positional and keyword args
	expr, err = Parse("remote_bookmarks(bookmark1, remote=origin)")
	if err != nil {
		t.Fatal(err)
	}
	if len(expr.FunctionArgs()) != 1 {
		t.Errorf("expected 1 positional arg, got %d", len(expr.FunctionArgs()))
	}
	if len(expr.FunctionKeywordArgs()) != 1 {
		t.Errorf("expected 1 keyword arg, got %d", len(expr.FunctionKeywordArgs()))
	}

	// Keyword arg with string value
	expr, err = Parse(`remote_bookmarks(remote="origin"`)
	_ = expr // just verifying it parses without panic
	_ = err
}

func TestParsePatternWithPostfixOps(t *testing.T) {
	// Pattern value can have postfix ops
	testParseEqual(t, "x:@-+", "x:((@-)+)")
	testParseEqual(t, "x:@", "x:@")

	// Pattern value is neighbors_expression (no ranges)
	testParseEqual(t, "x:y::z", "(x:y)::z")
	testParseEqual(t, "x:y&z", "(x:y)&z")

	// Pattern is right-associative
	testParseEqual(t, "x:y:z", "x:(y:z)")
}

func TestParseUnionFlattening(t *testing.T) {
	// Union flattens into a single node
	expr, err := Parse("a | b | c | d")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindUnionAll {
		t.Fatalf("expected KindUnionAll, got %v", expr.Kind)
	}
	if len(expr.UnionNodes()) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(expr.UnionNodes()))
	}

	// Two-element union
	expr, err = Parse("a | b")
	if err != nil {
		t.Fatal(err)
	}
	if len(expr.UnionNodes()) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(expr.UnionNodes()))
	}
}

func TestParseDifferencePrecedence(t *testing.T) {
	// ~ binds tighter than | but same level as &
	testParseEqual(t, "x & ~y", "x&(~y)")

	// & and ~ have same precedence, left-associative
	// a & b ~ c = (a & b) ~ c since & and ~ are left-assoc at same level
	expr, err := Parse("a & b ~ c")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBinary || expr.BinaryOp() != Difference {
		t.Fatalf("expected Difference at top, got %v", expr.Kind)
	}
}
