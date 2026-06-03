package template

import (
	"slices"
	"testing"
)

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("", 0)
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasOperator(t, ctx, "!")
	assertHasOperator(t, ctx, "-")
}

func TestCompletionAfterCompleteExpression(t *testing.T) {
	ctx := ParseForCompletion("foo", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected PartialIdent 'foo', got %q", ctx.PartialIdent)
	}
	assertHasOperator(t, ctx, "||")
	assertHasOperator(t, ctx, "&&")
	assertHasOperator(t, ctx, "==")
	assertHasOperator(t, ctx, "++")
}

func TestCompletionPartialIdentifier(t *testing.T) {
	ctx := ParseForCompletion("commi", -1)
	if ctx.PartialIdent != "commi" {
		t.Errorf("expected PartialIdent 'commi', got %q", ctx.PartialIdent)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterOperator(t *testing.T) {
	ctx := ParseForCompletion("foo && ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterLogicalOr(t *testing.T) {
	ctx := ParseForCompletion("foo || ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterConcat(t *testing.T) {
	ctx := ParseForCompletion("foo ++ ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionInFunctionEmpty(t *testing.T) {
	ctx := ParseForCompletion("if(", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "if" {
		t.Errorf("expected function name 'if', got %q", ctx.Function.Name)
	}
	if ctx.Function.ArgIndex != 0 {
		t.Errorf("expected arg index 0, got %d", ctx.Function.ArgIndex)
	}
}

func TestCompletionInFunctionAfterArg(t *testing.T) {
	ctx := ParseForCompletion("if(foo", -1)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.ArgIndex != 1 {
		t.Errorf("expected arg index 1, got %d", ctx.Function.ArgIndex)
	}
}

func TestCompletionInFunctionMultipleArgs(t *testing.T) {
	ctx := ParseForCompletion("if(a, b", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(ctx.Function.Args))
	}
}

func TestCompletionInFunctionAfterComma(t *testing.T) {
	ctx := ParseForCompletion("if(a, ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.ArgIndex != 1 {
		t.Errorf("expected arg index 1, got %d", ctx.Function.ArgIndex)
	}
}

func TestCompletionInFunctionKeywordArg(t *testing.T) {
	ctx := ParseForCompletion("label(foo", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected PartialIdent 'foo', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionAfterKeywordEquals(t *testing.T) {
	ctx := ParseForCompletion("label(color=", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.KeywordArgName != "color" {
		t.Errorf("expected keyword arg name 'color', got %q", ctx.Function.KeywordArgName)
	}
}

func TestCompletionInParenthesized(t *testing.T) {
	ctx := ParseForCompletion("(foo", -1)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionInParenthesizedEmpty(t *testing.T) {
	ctx := ParseForCompletion("( ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionPartialString(t *testing.T) {
	ctx := ParseForCompletion(`"fo`, -1)
	if ctx.PartialString != "fo" {
		t.Errorf("expected PartialString 'fo', got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '"' {
		t.Errorf("expected StringQuote \", got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
}

func TestCompletionPartialRawString(t *testing.T) {
	ctx := ParseForCompletion(`'fo`, -1)
	if ctx.PartialString != "fo" {
		t.Errorf("expected PartialString 'fo', got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '\'' {
		t.Errorf("expected StringQuote ', got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
}

func TestCompletionInPattern(t *testing.T) {
	ctx := ParseForCompletion("exact:", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "exact" {
		t.Errorf("expected PatternName 'exact', got %q", ctx.PatternName)
	}
	assertHasExpected(t, ctx, ExpectedPatternValue)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionInPatternWithPartialIdent(t *testing.T) {
	ctx := ParseForCompletion("exact:fo", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "exact" {
		t.Errorf("expected PatternName 'exact', got %q", ctx.PatternName)
	}
}

func TestCompletionAfterNot(t *testing.T) {
	ctx := ParseForCompletion("!", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegate(t *testing.T) {
	ctx := ParseForCompletion("-", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegateExpression(t *testing.T) {
	ctx := ParseForCompletion("-foo", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionMethodCall(t *testing.T) {
	ctx := ParseForCompletion("foo.", -1)
	if ctx.PartialIdent != "" {
		t.Errorf("expected empty PartialIdent, got %q", ctx.PartialIdent)
	}
}

func TestCompletionNestedFunction(t *testing.T) {
	ctx := ParseForCompletion("if(label(", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "label" {
		t.Errorf("expected function name 'label', got %q", ctx.Function.Name)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionEmptyFunctionCall(t *testing.T) {
	ctx := ParseForCompletion("if(true)", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionBooleanLiteral(t *testing.T) {
	ctx := ParseForCompletion("true", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionLambdaZeroArgs(t *testing.T) {
	ctx := ParseForCompletion("|| ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	if !ctx.InLambda {
		t.Error("expected InLambda")
	}
}

func TestCompletionLambdaWithArgs(t *testing.T) {
	ctx := ParseForCompletion("|x| ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	if !ctx.InLambda {
		t.Error("expected InLambda")
	}
	if len(ctx.LambdaParams) != 1 || ctx.LambdaParams[0] != "x" {
		t.Errorf("expected lambda params [x], got %v", ctx.LambdaParams)
	}
}

func TestCompletionInLambdaParams(t *testing.T) {
	ctx := ParseForCompletion("|x", -1)
	assertHasExpected(t, ctx, ExpectedLambdaClose)
	if !ctx.InLambda {
		t.Error("expected InLambda")
	}
}

func TestCompletionConcatAfterExpression(t *testing.T) {
	ctx := ParseForCompletion("foo ++ bar", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInConcatMiddle(t *testing.T) {
	// "foo ++ " with cursor at end - expect expression
	ctx := ParseForCompletion("foo ++ ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionPrefixNot(t *testing.T) {
	ctx := ParseForCompletion("!foo", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixOperator(t *testing.T) {
	ctx := ParseForCompletion("foo && bar", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAfterKeywordArgValue(t *testing.T) {
	ctx := ParseForCompletion("label(color=red", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionMethodCallOnExpression(t *testing.T) {
	ctx := ParseForCompletion("foo.bar(", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "bar" {
		t.Errorf("expected function name 'bar', got %q", ctx.Function.Name)
	}
	if !ctx.Function.IsMethod {
		t.Error("expected IsMethod")
	}
}

func TestCompletionPatternCaseInsensitive(t *testing.T) {
	ctx := ParseForCompletion("glob-i:", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "glob-i" {
		t.Errorf("expected PatternName 'glob-i', got %q", ctx.PatternName)
	}
}

func TestCompletionTrailingComma(t *testing.T) {
	ctx := ParseForCompletion("if(a,", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionCursorPosition(t *testing.T) {
	input := "foo ++ bar"
	cursor := 4 // position after "foo "
	ctx := ParseForCompletion(input, cursor)
	assertHasExpected(t, ctx, ExpectedOperator)
}

// --- Helpers ---

func assertHasExpected(t *testing.T, ctx *CompletionContext, expected ExpectedToken) {
	t.Helper()
	if slices.Contains(ctx.ExpectedTokens, expected) {
		return
	}
	t.Errorf("expected %s in ExpectedTokens, got %v", expected, ctx.ExpectedTokens)
}

func assertHasOperator(t *testing.T, ctx *CompletionContext, op string) {
	t.Helper()
	for _, v := range ctx.ValidOperators {
		if v.Op == op {
			return
		}
	}
	t.Errorf("expected operator %q in ValidOperators, got %v", op, ctx.ValidOperators)
}