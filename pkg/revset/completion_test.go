package revset

import (
	"slices"
	"testing"
)

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("", 0)
	assertHasExpected(t, ctx, ExpectedExpression)
	// Prefix operators ~, ::, .. are also valid at the start
	assertHasOperator(t, ctx, "~")
	assertHasOperator(t, ctx, "::")
	assertHasOperator(t, ctx, "..")
}

func TestCompletionAfterCompleteExpression(t *testing.T) {
	// "foo" with cursor at end - operators or extending the identifier
	ctx := ParseForCompletion("foo", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
	// PartialIdent is "foo" since cursor is at end of identifier
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected PartialIdent 'foo', got %q", ctx.PartialIdent)
	}
	assertHasOperator(t, ctx, "|")
	assertHasOperator(t, ctx, "&")
	assertHasOperator(t, ctx, "~")
	assertHasOperator(t, ctx, "::")
	assertHasOperator(t, ctx, "..")
	assertHasOperator(t, ctx, "-")
	assertHasOperator(t, ctx, "+")
}

func TestCompletionPartialIdentifier(t *testing.T) {
	// "par" with cursor at end - partial identifier
	ctx := ParseForCompletion("par", -1)
	if ctx.PartialIdent != "par" {
		t.Errorf("expected PartialIdent 'par', got %q", ctx.PartialIdent)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterOperator(t *testing.T) {
	// "foo |" with cursor at end - expect expression
	ctx := ParseForCompletion("foo | ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterAmpersand(t *testing.T) {
	// "foo & " with cursor at end - expect expression
	ctx := ParseForCompletion("foo & ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterTilde(t *testing.T) {
	// "foo ~ " with cursor at end - expect expression
	ctx := ParseForCompletion("foo ~ ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterUnion(t *testing.T) {
	// "foo | bar" with cursor at end - expect operators
	ctx := ParseForCompletion("foo | bar", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
	// bar is a partial identifier that could be extended
	if ctx.PartialIdent != "bar" {
		t.Errorf("expected PartialIdent 'bar', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionEmpty(t *testing.T) {
	// "parents(" with cursor at end - expect expression and )
	ctx := ParseForCompletion("parents(", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "parents" {
		t.Errorf("expected function name 'parents', got %q", ctx.Function.Name)
	}
	if ctx.Function.ArgIndex != 0 {
		t.Errorf("expected arg index 0, got %d", ctx.Function.ArgIndex)
	}
}

func TestCompletionInFunctionAfterArg(t *testing.T) {
	// "parents(foo" with cursor at end
	ctx := ParseForCompletion("parents(foo", -1)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "parents" {
		t.Errorf("expected function name 'parents', got %q", ctx.Function.Name)
	}
	if ctx.Function.ArgIndex != 1 {
		t.Errorf("expected arg index 1, got %d", ctx.Function.ArgIndex)
	}
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ctx.Function.Args))
	}
	if ctx.Function.Args[0] == nil {
		t.Fatal("expected non-nil arg")
	}
	if ctx.Function.Args[0].Kind != KindIdentifier {
		t.Errorf("expected identifier arg, got %v", ctx.Function.Args[0].Kind)
	}
	if ctx.Function.Args[0].Identifier() != "foo" {
		t.Errorf("expected arg identifier 'foo', got %q", ctx.Function.Args[0].Identifier())
	}
}

func TestCompletionInFunctionMultipleArgs(t *testing.T) {
	// "file(a, b" with cursor at end
	ctx := ParseForCompletion("file(a, b", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(ctx.Function.Args))
	}
	if ctx.Function.Args[0] == nil {
		t.Fatal("expected non-nil first arg")
	}
	if ctx.Function.Args[0].Identifier() != "a" {
		t.Errorf("expected first arg 'a', got %q", ctx.Function.Args[0].Identifier())
	}
	if ctx.Function.Args[1] == nil {
		t.Fatal("expected non-nil second arg")
	}
	if ctx.Function.Args[1].Identifier() != "b" {
		t.Errorf("expected second arg 'b', got %q", ctx.Function.Args[1].Identifier())
	}
}

func TestCompletionInFunctionAfterComma(t *testing.T) {
	// "file(a, " with cursor at end - expect next arg
	ctx := ParseForCompletion("file(a, ", -1)
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
	// "remote_bookmarks(remote" with cursor at end
	ctx := ParseForCompletion("remote_bookmarks(remote", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	// At this point 'remote' could be a keyword arg name or a positional arg
	// Since we haven't seen '=' yet, it's ambiguous
	if ctx.Function.KeywordArgName != "remote" {
		t.Errorf("expected keyword arg name 'remote', got %q", ctx.Function.KeywordArgName)
	}
}

func TestCompletionInFunctionAfterKeywordEquals(t *testing.T) {
	// "remote_bookmarks(remote=" with cursor at end
	ctx := ParseForCompletion("remote_bookmarks(remote=", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "remote_bookmarks" {
		t.Errorf("expected function name 'remote_bookmarks', got %q", ctx.Function.Name)
	}
}

func TestCompletionInParenthesized(t *testing.T) {
	// "(foo" with cursor at end - expect )
	ctx := ParseForCompletion("(foo", -1)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionInParenthesizedEmpty(t *testing.T) {
	// "( " with cursor at end - expect expression and )
	ctx := ParseForCompletion("( ", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionPartialString(t *testing.T) {
	// `"fo` with cursor at end
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
	// `'fo` with cursor at end
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
	// "exact:" with cursor at end
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

func TestCompletionAfterAt(t *testing.T) {
	// "main@" with cursor at end - completing remote name
	ctx := ParseForCompletion("main@", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAtCursorPosition(t *testing.T) {
	// "foo | bar |" with cursor at position of second |
	input := "foo | bar |"
	cursor := len(input)
	ctx := ParseForCompletion(input, cursor)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionCursorInMiddle(t *testing.T) {
	// "foo | bar" with cursor after "foo " (before the |)
	input := "foo | bar"
	cursor := 4 // position after 'foo '
	ctx := ParseForCompletion(input, cursor)
	// At position 4 we're in whitespace before |
	// After completing 'foo', operators are valid
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAfterDagRangePrefix(t *testing.T) {
	// "::" with cursor at end (nullary)
	ctx := ParseForCompletion("::", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAfterRangeAll(t *testing.T) {
	// ".." with cursor at end (nullary)
	ctx := ParseForCompletion("..", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionNegatePrefix(t *testing.T) {
	// "~" with cursor at end
	ctx := ParseForCompletion("~", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegate(t *testing.T) {
	// "~foo" with cursor at end
	ctx := ParseForCompletion("~foo", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixDagRange(t *testing.T) {
	// "foo::bar" with cursor at end
	ctx := ParseForCompletion("foo::bar", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionDagRangeNeedsRight(t *testing.T) {
	// "foo::" with cursor at end (infix, needs RHS)
	// This is actually a postfix ::, which means after-expression operators
	ctx := ParseForCompletion("foo::", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionDagRangeInfixNeedsRight(t *testing.T) {
	// Placeholder - tested in TestCompletionInfixDagRangeNeedsRight
}

func TestCompletionPostfixParents(t *testing.T) {
	// "foo-" with cursor at end
	ctx := ParseForCompletion("foo-", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPostfixChildren(t *testing.T) {
	// "foo+" with cursor at end
	ctx := ParseForCompletion("foo+", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionNestedFunction(t *testing.T) {
	// "parents(file(" with cursor at end
	ctx := ParseForCompletion("parents(file(", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "file" {
		t.Errorf("expected function name 'file', got %q", ctx.Function.Name)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionAfterKeywordArgValue(t *testing.T) {
	// "remote_bookmarks(remote=foo" with cursor at end
	ctx := ParseForCompletion("remote_bookmarks(remote=foo", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionInfixDagRangeNeedsRight(t *testing.T) {
	// "foo::" with cursor right after :: - needs RHS expression
	input := "foo::bar"
	cursor := 5 // position right after "foo::"
	ctx := ParseForCompletion(input, cursor)
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterAtWorkspace(t *testing.T) {
	// "@" with cursor at end
	ctx := ParseForCompletion("@", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAtInFunction(t *testing.T) {
	// "parents(@" with cursor at end
	ctx := ParseForCompletion("parents(@", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "parents" {
		t.Errorf("expected function name 'parents', got %q", ctx.Function.Name)
	}
}

func TestCompletionTrailingComma(t *testing.T) {
	// "bookmarks(a," with cursor at end (trailing comma allowed)
	ctx := ParseForCompletion("bookmarks(a,", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionEmptyFunctionCall(t *testing.T) {
	// "visible_heads()" with cursor at end
	ctx := ParseForCompletion("visible_heads()", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInPatternWithPartialIdent(t *testing.T) {
	// "exact:fo" with cursor at end - pattern value is partial identifier
	ctx := ParseForCompletion("exact:fo", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "exact" {
		t.Errorf("expected PatternName 'exact', got %q", ctx.PatternName)
	}
}

func TestCompletionAfterDifferenceInFunction(t *testing.T) {
	// "file(foo ~ " with cursor at end
	ctx := ParseForCompletion("file(foo ~ ", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionRemoteSymbolPartial(t *testing.T) {
	// "main@ori" with cursor at end - partial remote name
	ctx := ParseForCompletion("main@ori", -1)
	if ctx.PartialIdent != "ori" {
		t.Errorf("expected PartialIdent 'ori', got %q", ctx.PartialIdent)
	}
}

func TestCompletionArgSpanAndContent(t *testing.T) {
	// "parents(foo" - verify arg span and content
	ctx := ParseForCompletion("parents(foo", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ctx.Function.Args))
	}
	arg := ctx.Function.Args[0]
	if arg.Kind != KindIdentifier {
		t.Errorf("expected KindIdentifier, got %v", arg.Kind)
	}
	if arg.Identifier() != "foo" {
		t.Errorf("expected identifier 'foo', got %q", arg.Identifier())
	}
	if arg.Span.Start != 8 || arg.Span.End != 11 {
		t.Errorf("expected span [8,11), got [%d,%d)", arg.Span.Start, arg.Span.End)
	}
}

func TestCompletionArgStringContent(t *testing.T) {
	// `parents("foo"` - verify string arg content
	ctx := ParseForCompletion(`parents("foo"`, -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ctx.Function.Args))
	}
	arg := ctx.Function.Args[0]
	if arg.Kind != KindString {
		t.Errorf("expected KindString, got %v", arg.Kind)
	}
	if arg.StringValue() != "foo" {
		t.Errorf("expected string value 'foo', got %q", arg.StringValue())
	}
	if arg.Span.Start != 8 || arg.Span.End != 13 {
		t.Errorf("expected span [8,13), got [%d,%d)", arg.Span.Start, arg.Span.End)
	}
}

func TestCompletionArgAtWorkspaceContent(t *testing.T) {
	// "parents(foo@" - verify at workspace arg
	ctx := ParseForCompletion("parents(foo@", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ctx.Function.Args))
	}
	arg := ctx.Function.Args[0]
	if arg.Kind != KindAtWorkspace {
		t.Errorf("expected KindAtWorkspace, got %v", arg.Kind)
	}
	if arg.AtWorkspaceName() != "foo" {
		t.Errorf("expected name 'foo', got %q", arg.AtWorkspaceName())
	}
}

func TestCompletionArgRemoteSymbolContent(t *testing.T) {
	// "parents(foo@bar" - verify remote symbol arg
	ctx := ParseForCompletion("parents(foo@bar", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(ctx.Function.Args))
	}
	arg := ctx.Function.Args[0]
	if arg.Kind != KindRemoteSymbol {
		t.Errorf("expected KindRemoteSymbol, got %v", arg.Kind)
	}
	if arg.RemoteSymbolName() != "foo" {
		t.Errorf("expected name 'foo', got %q", arg.RemoteSymbolName())
	}
	if arg.RemoteSymbolRemote() != "bar" {
		t.Errorf("expected remote 'bar', got %q", arg.RemoteSymbolRemote())
	}
}

func TestCompletionFunctionCallExpr(t *testing.T) {
	// "parents(foo)" - verify function call expression
	ctx := ParseForCompletion("parents(foo)", -1)
	if ctx.Function != nil {
		// After the function is closed, Function context should not be set
		// (the cursor is after the closing paren)
	}
	// The lastExpr should be a function call
	// This is tested indirectly - the function call is complete
}

func TestCompletionMultipleArgsSpanAndContent(t *testing.T) {
	// "file(a, b, c" - verify all args have correct span and content
	ctx := ParseForCompletion("file(a, b, c", -1)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(ctx.Function.Args))
	}
	for i, expected := range []struct {
		name  string
		spanS int
		spanE int
	}{
		{"a", 5, 6},
		{"b", 8, 9},
		{"c", 11, 12},
	} {
		arg := ctx.Function.Args[i]
		if arg == nil {
			t.Fatalf("arg %d is nil", i)
		}
		if arg.Identifier() != expected.name {
			t.Errorf("arg %d: expected identifier %q, got %q", i, expected.name, arg.Identifier())
		}
		if arg.Span.Start != expected.spanS || arg.Span.End != expected.spanE {
			t.Errorf("arg %d: expected span [%d,%d), got [%d,%d)", i, expected.spanS, expected.spanE, arg.Span.Start, arg.Span.End)
		}
	}
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

func TestCompletionNullaryRange(t *testing.T) {
	// ".." with cursor at end (nullary range)
	ctx := ParseForCompletion("..", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPrefixRange(t *testing.T) {
	// "..foo" with cursor at end (prefix range)
	ctx := ParseForCompletion("..foo", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixRange(t *testing.T) {
	// "foo..bar" with cursor at end
	ctx := ParseForCompletion("foo..bar", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixRangeNeedsRight(t *testing.T) {
	// "foo.." with cursor at end - could be postfix or infix needing RHS
	ctx := ParseForCompletion("foo..", -1)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPatternKinds(t *testing.T) {
	// "exact:" — already tested in TestCompletionInPattern, verify other kinds

	// "glob:" — pattern
	ctx := ParseForCompletion("glob:", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern for glob:")
	}
	if ctx.PatternName != "glob" {
		t.Errorf("expected PatternName 'glob', got %q", ctx.PatternName)
	}

	// "substring:" — pattern
	ctx = ParseForCompletion("substring:", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern for substring:")
	}

	// "regex:" — pattern
	ctx = ParseForCompletion("regex:", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern for regex:")
	}
}

func TestCompletionPatternCaseInsensitive(t *testing.T) {
	// "glob-i:" — case-insensitive pattern
	ctx := ParseForCompletion("glob-i:", -1)
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "glob-i" {
		t.Errorf("expected PatternName 'glob-i', got %q", ctx.PatternName)
	}
}

func TestCompletionKeywordArgEqualsExpr(t *testing.T) {
	// "remote_bookmarks(remote=" with cursor at end — should expect expression
	ctx := ParseForCompletion("remote_bookmarks(remote=", -1)
	assertHasExpected(t, ctx, ExpectedExpression)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "remote_bookmarks" {
		t.Errorf("expected function name 'remote_bookmarks', got %q", ctx.Function.Name)
	}
}
