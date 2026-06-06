package template

import (
	"slices"
	"testing"
)

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("")
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasOperator(t, ctx, "!")
	assertHasOperator(t, ctx, "-")
}

func TestCompletionAfterCompleteExpression(t *testing.T) {
	ctx := ParseForCompletion("foo")
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
	ctx := ParseForCompletion("commi")
	if ctx.PartialIdent != "commi" {
		t.Errorf("expected PartialIdent 'commi', got %q", ctx.PartialIdent)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterOperator(t *testing.T) {
	ctx := ParseForCompletion("foo && ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterLogicalOr(t *testing.T) {
	ctx := ParseForCompletion("foo || ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterConcat(t *testing.T) {
	ctx := ParseForCompletion("foo ++ ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionInFunctionEmpty(t *testing.T) {
	ctx := ParseForCompletion("if(")
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
	ctx := ParseForCompletion("if(foo")
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	// 'foo' is partial (cursor at end), so argIndex is still 0
	if ctx.Function.ArgIndex != 0 {
		t.Errorf("expected arg index 0, got %d", ctx.Function.ArgIndex)
	}
}

func TestCompletionInFunctionMultipleArgs(t *testing.T) {
	ctx := ParseForCompletion("if(a, b")
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
	ctx := ParseForCompletion("if(a, ")
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
	ctx := ParseForCompletion("label(foo")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected PartialIdent 'foo', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionAfterKeywordEquals(t *testing.T) {
	ctx := ParseForCompletion("label(color=")
	assertHasExpected(t, ctx, ExpectedExpression)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.KeywordArgName != "color" {
		t.Errorf("expected keyword arg name 'color', got %q", ctx.Function.KeywordArgName)
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

func TestCompletionPartialRawString(t *testing.T) {
	ctx := ParseForCompletion(`'fo`)
	if ctx.PartialString != "fo" {
		t.Errorf("expected PartialString 'fo', got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '\'' {
		t.Errorf("expected StringQuote ', got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
}

func TestCompletionInPattern(t *testing.T) {
	ctx := ParseForCompletion("exact:")
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
	ctx := ParseForCompletion("exact:fo")
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "exact" {
		t.Errorf("expected PatternName 'exact', got %q", ctx.PatternName)
	}
}

func TestCompletionAfterNot(t *testing.T) {
	ctx := ParseForCompletion("!")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegate(t *testing.T) {
	ctx := ParseForCompletion("-")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegateExpression(t *testing.T) {
	ctx := ParseForCompletion("-foo")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionMethodCall(t *testing.T) {
	ctx := ParseForCompletion("foo.")
	if ctx.PartialIdent != "" {
		t.Errorf("expected empty PartialIdent, got %q", ctx.PartialIdent)
	}
}

func TestCompletionNestedFunction(t *testing.T) {
	ctx := ParseForCompletion("if(label(")
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
	ctx := ParseForCompletion("if(true)")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionBooleanLiteral(t *testing.T) {
	ctx := ParseForCompletion("true")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionLambdaZeroArgs(t *testing.T) {
	ctx := ParseForCompletion("|| ")
	assertHasExpected(t, ctx, ExpectedExpression)
	if !ctx.InLambda {
		t.Error("expected InLambda")
	}
}

func TestCompletionLambdaWithArgs(t *testing.T) {
	ctx := ParseForCompletion("|x| ")
	assertHasExpected(t, ctx, ExpectedExpression)
	if !ctx.InLambda {
		t.Error("expected InLambda")
	}
	if len(ctx.LambdaParams) != 1 || ctx.LambdaParams[0] != "x" {
		t.Errorf("expected lambda params [x], got %v", ctx.LambdaParams)
	}
}

func TestCompletionInLambdaParams(t *testing.T) {
	ctx := ParseForCompletion("|x")
	assertHasExpected(t, ctx, ExpectedLambdaClose)
	if !ctx.InLambda {
		t.Error("expected InLambda")
	}
}

func TestCompletionConcatAfterExpression(t *testing.T) {
	ctx := ParseForCompletion("foo ++ bar")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInConcatMiddle(t *testing.T) {
	// "foo ++ " with cursor at end - expect expression
	ctx := ParseForCompletion("foo ++ ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionPrefixNot(t *testing.T) {
	ctx := ParseForCompletion("!foo")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixOperator(t *testing.T) {
	ctx := ParseForCompletion("foo && bar")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAfterKeywordArgValue(t *testing.T) {
	ctx := ParseForCompletion("label(color=red")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionMethodCallOnExpression(t *testing.T) {
	ctx := ParseForCompletion("foo.bar(")
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

func TestCompletionMethodTypeSelf(t *testing.T) {
	ctx := ParseForCompletion("self.")
	if ctx.MethodType != "Commit" {
		t.Errorf("expected MethodType 'Commit', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeSelfPartial(t *testing.T) {
	ctx := ParseForCompletion("self.desc")
	if ctx.MethodType != "Commit" {
		t.Errorf("expected MethodType 'Commit', got %q", ctx.MethodType)
	}
	if ctx.PartialIdent != "desc" {
		t.Errorf("expected PartialIdent 'desc', got %q", ctx.PartialIdent)
	}
}

func TestCompletionMethodTypeChain(t *testing.T) {
	ctx := ParseForCompletion("self.description().")
	if ctx.MethodType != "String" {
		t.Errorf("expected MethodType 'String', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeDeepChain(t *testing.T) {
	ctx := ParseForCompletion("self.author().email().")
	if ctx.MethodType != "Email" {
		t.Errorf("expected MethodType 'Email', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeChangeId(t *testing.T) {
	ctx := ParseForCompletion("self.change_id().")
	if ctx.MethodType != "ChangeId" {
		t.Errorf("expected MethodType 'ChangeId', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeCommitKeyword(t *testing.T) {
	ctx := ParseForCompletion("description.sh")
	if ctx.MethodType != "String" {
		t.Errorf("expected MethodType 'String', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeGlobalFunction(t *testing.T) {
	ctx := ParseForCompletion("config().")
	if ctx.MethodType != "Option<ConfigValue>" {
		t.Errorf("expected MethodType 'Option<ConfigValue>', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeStringLiteral(t *testing.T) {
	ctx := ParseForCompletion(`"hello".`)
	if ctx.MethodType != "String" {
		t.Errorf("expected MethodType 'String', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeInteger(t *testing.T) {
	ctx := ParseForCompletion("42.")
	if ctx.MethodType != "Integer" {
		t.Errorf("expected MethodType 'Integer', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeBoolean(t *testing.T) {
	ctx := ParseForCompletion("true.")
	if ctx.MethodType != "Boolean" {
		t.Errorf("expected MethodType 'Boolean', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeShortestIdPrefix(t *testing.T) {
	ctx := ParseForCompletion("self.commit_id().shortest().")
	if ctx.MethodType != "ShortestIdPrefix" {
		t.Errorf("expected MethodType 'ShortestIdPrefix', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeSignature(t *testing.T) {
	ctx := ParseForCompletion("self.author().")
	if ctx.MethodType != "Signature" {
		t.Errorf("expected MethodType 'Signature', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeTimestamp(t *testing.T) {
	ctx := ParseForCompletion("self.author().timestamp().")
	if ctx.MethodType != "Timestamp" {
		t.Errorf("expected MethodType 'Timestamp', got %q", ctx.MethodType)
	}
}

func TestCompletionMethodTypeOptionConfigValue(t *testing.T) {
	ctx := ParseForCompletion("config().as_string().")
	if ctx.MethodType != "String" {
		t.Errorf("expected MethodType 'String', got %q", ctx.MethodType)
	}
}

func TestCompletionPatternCaseInsensitive(t *testing.T) {
	ctx := ParseForCompletion("glob-i:")
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "glob-i" {
		t.Errorf("expected PatternName 'glob-i', got %q", ctx.PatternName)
	}
}

func TestCompletionTrailingComma(t *testing.T) {
	ctx := ParseForCompletion("if(a,")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
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
