package fileset

import (
	"testing"
)

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("")
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasOperator(t, ctx, "~")
}

func TestCompletionAfterCompleteExpression(t *testing.T) {
	ctx := ParseForCompletion("foo")
	assertHasExpected(t, ctx, ExpectedOperator)
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected PartialIdent 'foo', got %q", ctx.PartialIdent)
	}
	assertHasOperator(t, ctx, "|")
	assertHasOperator(t, ctx, "&")
	assertHasOperator(t, ctx, "~")
}

func TestCompletionPartialIdentifier(t *testing.T) {
	ctx := ParseForCompletion("al")
	if ctx.PartialIdent != "al" {
		t.Errorf("expected PartialIdent 'al', got %q", ctx.PartialIdent)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterOperator(t *testing.T) {
	ctx := ParseForCompletion("foo | ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterAmpersand(t *testing.T) {
	ctx := ParseForCompletion("foo & ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterTilde(t *testing.T) {
	ctx := ParseForCompletion("foo ~ ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegate(t *testing.T) {
	ctx := ParseForCompletion("~foo")
	assertHasExpected(t, ctx, ExpectedOperator)
	assertHasOperator(t, ctx, "|")
	assertHasOperator(t, ctx, "&")
	assertHasOperator(t, ctx, "~")
}

func TestCompletionNegatePrefix(t *testing.T) {
	ctx := ParseForCompletion("~")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterUnion(t *testing.T) {
	ctx := ParseForCompletion("foo | bar")
	assertHasExpected(t, ctx, ExpectedOperator)
	if ctx.PartialIdent != "bar" {
		t.Errorf("expected PartialIdent 'bar', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionEmpty(t *testing.T) {
	ctx := ParseForCompletion("all(")
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "all" {
		t.Errorf("expected function name 'all', got %q", ctx.Function.Name)
	}
	if ctx.Function.ArgIndex != 0 {
		t.Errorf("expected arg index 0, got %d", ctx.Function.ArgIndex)
	}
	if !ctx.Function.IsZeroArg {
		t.Error("expected IsZeroArg")
	}
}

func TestCompletionInFunctionAfterArg(t *testing.T) {
	ctx := ParseForCompletion("all(foo")
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "all" {
		t.Errorf("expected function name 'all', got %q", ctx.Function.Name)
	}
}

func TestCompletionInFunctionMultipleArgs(t *testing.T) {
	ctx := ParseForCompletion("all(a, b")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	// 'a' is complete (followed by comma), 'b' is partial (cursor at end)
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 complete arg, got %d", len(ctx.Function.Args))
	}
	if ctx.PartialIdent != "b" {
		t.Errorf("expected partialIdent 'b', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionAfterComma(t *testing.T) {
	ctx := ParseForCompletion("all(a, ")
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if ctx.Function != nil && ctx.Function.ArgIndex != 1 {
		t.Errorf("expected arg index 1, got %d", ctx.Function.ArgIndex)
	}
}

func TestCompletionInParenthesized(t *testing.T) {
	ctx := ParseForCompletion("(foo")
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionInParenthesizedEmpty(t *testing.T) {
	ctx := ParseForCompletion("( ")
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionPartialString(t *testing.T) {
	ctx := ParseForCompletion(`"fo`)
	if ctx.PartialString != "fo" {
		t.Errorf("expected PartialString 'fo', got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '"' {
		t.Errorf("expected StringQuote \", got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
}

func TestCompletionInPattern(t *testing.T) {
	ctx := ParseForCompletion("glob:")
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "glob" {
		t.Errorf("expected PatternName 'glob', got %q", ctx.PatternName)
	}
	assertHasExpected(t, ctx, ExpectedPatternValue)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionInPatternWithPartial(t *testing.T) {
	ctx := ParseForCompletion("exact:fo")
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "exact" {
		t.Errorf("expected PatternName 'exact', got %q", ctx.PatternName)
	}
}

func TestCompletionAfterDifference(t *testing.T) {
	ctx := ParseForCompletion("foo ~ ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionIntersection(t *testing.T) {
	ctx := ParseForCompletion("foo & bar")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionNestedFunction(t *testing.T) {
	ctx := ParseForCompletion("all(all(")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "all" {
		t.Errorf("expected function name 'all', got %q", ctx.Function.Name)
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if !ctx.Function.IsZeroArg {
		t.Error("expected IsZeroArg for inner all()")
	}
}

func TestCompletionEmptyFunctionCall(t *testing.T) {
	ctx := ParseForCompletion("all()")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionTrailingComma(t *testing.T) {
	ctx := ParseForCompletion("all(a,")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func assertHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	for _, tok := range ctx.ExpectedTokens {
		if tok == expected {
			return
		}
	}
	t.Errorf("expected token %v not found in %v", expected, ctx.ExpectedTokens)
}

func assertHasOperator(t *testing.T, ctx *CompletionContext, op string) {
	t.Helper()
	for _, v := range ctx.ValidOperators {
		if v.Op == op {
			return
		}
	}
	t.Errorf("expected operator %q not found in %v", op, ctx.ValidOperators)
}
