package fileset

import (
	"testing"
)

func TestParseTreeEq(t *testing.T) {
	a, err := Parse(" all( ) | ~foo ")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Parse("(all())|(~(foo))")
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

func TestParseFileset(t *testing.T) {
	// Parse a quoted string
	testParseKind(t, `"foo"`, KindString)
	testParseKind(t, `'foo'`, KindString)

	// Parse the "negate" operator
	testParseUnaryOp(t, "~ foo", Negate)
	testParseEqual(t, "~~ foo", "~(~foo)")

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

	// Parenthesized expressions
	testParseEqual(t, "(foo)", "foo")

	// Space around expressions
	testParseEqual(t, " ~ foo ", "~foo")

	// Incomplete parse
	testParseError(t, "foo | ~")
}

func TestParseWhitespace(t *testing.T) {
	testParseEqual(t, " \t\r\n\x0call()", "all()")
}

func TestParseIdentifier(t *testing.T) {
	// Standard identifier
	testParseKind(t, "foo_bar", KindIdentifier)

	// Path with /
	testParseKind(t, "src/lib.rs", KindIdentifier)

	// Glob characters *, ?, [], /
	testParseKind(t, "*.rs", KindIdentifier)
	testParseKind(t, "foo?.txt", KindIdentifier)
	testParseKind(t, "src[0-9]/test", KindIdentifier)

	// Internal . and - and @
	testParseKind(t, "foo.bar", KindIdentifier)
	testParseKind(t, "foo-bar", KindIdentifier)
	testParseKind(t, "foo@bar", KindIdentifier)

	// + is part of identifier in fileset (unlike revset where it's a postfix operator)
	testParseKind(t, "foo+bar", KindIdentifier)

	// \ as path separator
	testParseKind(t, "src\\test", KindIdentifier)

	// Parenthesized identifier
	testParseEqual(t, "(foo)", "foo")
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
}

func TestParsePattern(t *testing.T) {
	// Pattern with string value
	expr, err := Parse(`glob:"*.rs"`)
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindPattern {
		t.Fatalf("expected KindPattern, got %v", expr.Kind)
	}
	if expr.PatternName() != "glob" {
		t.Errorf("expected name 'glob', got %q", expr.PatternName())
	}
	if expr.PatternValue().Kind != KindString || expr.PatternValue().StringValue() != "*.rs" {
		t.Errorf("expected string value '*.rs', got %v", expr.PatternValue())
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
	testParseKind(t, `"glob:foo"`, KindString)

	// Whitespace isn't allowed in pattern
	testParseError(t, "glob: foo")
	testParseError(t, "glob :foo")

	// Parenthesized pattern value
	testParseEqual(t, "glob:( 'foo' )", `glob:"foo"`)

	// Pattern is right-associative and binds tighter than operators
	testParseEqual(t, "glob:foo&bar", "glob:foo&bar")
	testParseEqual(t, "glob:foo|bar", "glob:foo|bar")
}

func TestParseFunction(t *testing.T) {
	// No-arg function
	expr, err := Parse("all()")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindFunctionCall {
		t.Fatalf("expected KindFunctionCall, got %v", expr.Kind)
	}
	if expr.FunctionName() != "all" {
		t.Errorf("expected name 'all', got %q", expr.FunctionName())
	}
	if len(expr.FunctionArgs()) != 0 {
		t.Errorf("expected 0 args, got %d", len(expr.FunctionArgs()))
	}

	// Function with args
	expr, err = Parse("all()")
	if err != nil {
		t.Fatal(err)
	}
	if expr.FunctionName() != "all" {
		t.Errorf("expected 'all', got %q", expr.FunctionName())
	}

	// None function
	expr, err = Parse("none()")
	if err != nil {
		t.Fatal(err)
	}
	if expr.FunctionName() != "none" {
		t.Errorf("expected 'none', got %q", expr.FunctionName())
	}

	// Function in expression
	testParseEqual(t, "all() & none()", "all()&none()")
	testParseEqual(t, "~all()", "~all()")
}

func TestParseProgramOrBareString(t *testing.T) {
	// Expression parses normally
	expr, err := ParseProgramOrBareString("foo | bar")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindUnionAll {
		t.Errorf("expected KindUnionAll, got %v", expr.Kind)
	}

	// Path-like identifiers are valid expressions (no need for bare string fallback)
	expr, err = ParseProgramOrBareString("src/lib.rs")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindIdentifier {
		t.Errorf("expected KindIdentifier, got %v", expr.Kind)
	}
	if expr.Identifier() != "src/lib.rs" {
		t.Errorf("expected 'src/lib.rs', got %q", expr.Identifier())
	}

	// Bare string with spaces (can't be parsed as expression, falls back)
	expr, err = ParseProgramOrBareString("Foo Bar")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBareString {
		t.Errorf("expected KindBareString, got %v", expr.Kind)
	}
	if expr.BareStringValue() != "Foo Bar" {
		t.Errorf("expected 'Foo Bar', got %q", expr.BareStringValue())
	}

	// Pattern with glob-like identifier value (parses as expression pattern)
	expr, err = ParseProgramOrBareString("glob:*.rs")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindPattern {
		t.Errorf("expected KindPattern, got %v", expr.Kind)
	}
	if expr.PatternName() != "glob" {
		t.Errorf("expected 'glob', got %q", expr.PatternName())
	}
	if expr.PatternValue().Kind != KindIdentifier || expr.PatternValue().Identifier() != "*.rs" {
		t.Errorf("expected identifier value '*.rs', got %v", expr.PatternValue())
	}

	// Bare string pattern with spaces in value (falls back)
	expr, err = ParseProgramOrBareString("glob:*.rs foo")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindBareStringPattern {
		t.Errorf("expected KindBareStringPattern, got %v", expr.Kind)
	}
	if expr.PatternName() != "glob" {
		t.Errorf("expected 'glob', got %q", expr.PatternName())
	}

	// Identifier with + and - (valid as expression)
	expr, err = ParseProgramOrBareString("foo+bar-baz")
	if err != nil {
		t.Fatal(err)
	}
	if expr.Kind != KindIdentifier {
		t.Errorf("expected KindIdentifier, got %v", expr.Kind)
	}
}

func TestParsePrecedence(t *testing.T) {
	// & binds tighter than | (unambiguous without parens since & is higher precedence)
	testParseEqual(t, "x | y & z", "x|y&z")

	// ~ binds tighter than & (unambiguous without parens)
	testParseEqual(t, "x & ~y", "x&~y")

	// Pattern binds tighter than operators
	testParseEqual(t, "glob:*.rs | exact:foo", "glob:*.rs|exact:foo")

	// ~ binds tighter than ~ (infix)
	testParseEqual(t, "x ~ y", "x~y")
	testParseEqual(t, "x ~ ~y", "x~~y")

	// Parens force evaluation order
	testParseEqual(t, "(x | y) & z", "(x|y)&z")
	testParseEqual(t, "x | (y & z)", "x|y&z") // redundant parens
}

func TestParseExpressionSpans(t *testing.T) {
	tests := []struct {
		input   string
		spanStr string
	}{
		{" ~ x ", "~ x"},
		{" x |y ", "x |y"},
		{" (x) ", "(x)"},
		{"~( x|y) ", "~( x|y)"},
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

// Helper functions

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

func testParseError(t *testing.T, input string) {
	t.Helper()
	_, err := Parse(input)
	if err == nil {
		t.Errorf("parse %q: expected error, got nil", input)
	}
}

func testParseString(t *testing.T, input, expected string) {
	t.Helper()
	expr, err := Parse(input)
	if err != nil {
		t.Fatalf("parse %q: %v", input, err)
	}
	if expr.Kind != KindString {
		t.Fatalf("parse %q: expected KindString, got %v", input, expr.Kind)
	}
	if expr.StringValue() != expected {
		t.Errorf("parse %q: expected %q, got %q", input, expected, expr.StringValue())
	}
}