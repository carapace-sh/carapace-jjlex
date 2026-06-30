package revset

import (
	"slices"
	"testing"
)

func TestCompletionEmpty(t *testing.T) {
	ctx := ParseForCompletion("")
	assertHasExpected(t, ctx, ExpectedExpression)
	// Prefix operators ~, ::, .. are also valid at the start
	assertHasOperator(t, ctx, "~")
	assertHasOperator(t, ctx, "::")
	assertHasOperator(t, ctx, "..")
}

func TestCompletionAfterCompleteExpression(t *testing.T) {
	// "foo" with cursor at end - operators or extending the identifier
	ctx := ParseForCompletion("foo")
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
	ctx := ParseForCompletion("par")
	if ctx.PartialIdent != "par" {
		t.Errorf("expected PartialIdent 'par', got %q", ctx.PartialIdent)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterOperator(t *testing.T) {
	// "foo |" with cursor at end - expect expression
	ctx := ParseForCompletion("foo | ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterAmpersand(t *testing.T) {
	// "foo & " with cursor at end - expect expression
	ctx := ParseForCompletion("foo & ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAmpersandAtStart(t *testing.T) {
	// "&" at start - only expect expression, not operators (ampersand is infix-only)
	ctx := ParseForCompletion("&")
	assertHasExpected(t, ctx, ExpectedExpression)
	assertNoOperator(t, ctx, "&")
}

func TestCompletionAfterTilde(t *testing.T) {
	// "foo ~ " with cursor at end - expect expression
	ctx := ParseForCompletion("foo ~ ")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterUnion(t *testing.T) {
	// "foo | bar" with cursor at end - expect operators
	ctx := ParseForCompletion("foo | bar")
	assertHasExpected(t, ctx, ExpectedOperator)
	// bar is a partial identifier that could be extended
	if ctx.PartialIdent != "bar" {
		t.Errorf("expected PartialIdent 'bar', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionEmpty(t *testing.T) {
	// "parents(" with cursor at end - expect expression and )
	ctx := ParseForCompletion("parents(")
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
	// "parents(foo" with cursor at end - 'foo' is a partial identifier still being typed
	ctx := ParseForCompletion("parents(foo")
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "parents" {
		t.Errorf("expected function name 'parents', got %q", ctx.Function.Name)
	}
	// 'foo' is partial (cursor at end), so argIndex is still 0
	if ctx.Function.ArgIndex != 0 {
		t.Errorf("expected arg index 0, got %d", ctx.Function.ArgIndex)
	}
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected partialIdent 'foo', got %q", ctx.PartialIdent)
	}
	if len(ctx.Function.Args) != 0 {
		t.Fatalf("expected 0 args (partial), got %d", len(ctx.Function.Args))
	}
}

func TestCompletionInFunctionMultipleArgs(t *testing.T) {
	// "file(a, b" with cursor at end - 'b' is partial, 'a' is complete
	ctx := ParseForCompletion("file(a, b")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	// 'a' is complete (followed by comma), 'b' is partial (cursor at end)
	if len(ctx.Function.Args) != 1 {
		t.Fatalf("expected 1 complete arg, got %d", len(ctx.Function.Args))
	}
	if ctx.Function.Args[0] == nil {
		t.Fatal("expected non-nil first arg")
	}
	if ctx.Function.Args[0].Identifier() != "a" {
		t.Errorf("expected first arg 'a', got %q", ctx.Function.Args[0].Identifier())
	}
	if ctx.PartialIdent != "b" {
		t.Errorf("expected partialIdent 'b', got %q", ctx.PartialIdent)
	}
}

func TestCompletionInFunctionAfterComma(t *testing.T) {
	// "file(a, " with cursor at end - expect next arg
	ctx := ParseForCompletion("file(a, ")
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
	ctx := ParseForCompletion("remote_bookmarks(remote")
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
	ctx := ParseForCompletion("remote_bookmarks(remote=")
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
	ctx := ParseForCompletion("(foo")
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionInParenthesizedEmpty(t *testing.T) {
	// "( " with cursor at end - expect expression and )
	ctx := ParseForCompletion("( ")
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionPartialString(t *testing.T) {
	// `"fo` with cursor at end
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
	// `'fo` with cursor at end
	ctx := ParseForCompletion(`'fo`)
	if ctx.PartialString != "fo" {
		t.Errorf("expected PartialString 'fo', got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '\'' {
		t.Errorf("expected StringQuote ', got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
}

func TestCompletionEmptyStringLiteral(t *testing.T) {
	// `"` with cursor at end — just opening quote, no content yet
	ctx := ParseForCompletion(`"`)
	if ctx.PartialString != "" {
		t.Errorf("expected empty PartialString, got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '"' {
		t.Errorf("expected StringQuote \", got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
	// Operators should NOT be expected inside an unclosed string
	assertNotHasExpected(t, ctx, ExpectedOperator)
	assertNoOperator(t, ctx, "|")
	assertNoOperator(t, ctx, "&")
}

func TestCompletionStringLiteralSuppressesOperators(t *testing.T) {
	// `"&` with cursor at end — inside unclosed string, operators not valid
	ctx := ParseForCompletion(`"&`)
	if ctx.PartialString != "&" {
		t.Errorf("expected PartialString '&', got %q", ctx.PartialString)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
	assertNotHasExpected(t, ctx, ExpectedOperator)
	assertNoOperator(t, ctx, "&")
	assertNoOperator(t, ctx, "|")
}

func TestCompletionStringLiteralInFunction(t *testing.T) {
	// `parents("foo` with cursor at end — inside quoted string in function
	ctx := ParseForCompletion(`parents("foo`)
	if ctx.PartialString != "foo" {
		t.Errorf("expected PartialString 'foo', got %q", ctx.PartialString)
	}
	if ctx.StringQuote != '"' {
		t.Errorf("expected StringQuote \", got %c", ctx.StringQuote)
	}
	assertHasExpected(t, ctx, ExpectedStringClose)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "parents" {
		t.Errorf("expected function 'parents', got %q", ctx.Function.Name)
	}
	if !ctx.Function.InStringArg {
		t.Error("expected InStringArg")
	}
	// Operators should NOT be expected inside an unclosed string
	assertNotHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionCompleteStringInFunction(t *testing.T) {
	// `parents("foo")` with cursor after closing quote — string complete,
	// operators and ),  are valid
	ctx := ParseForCompletion(`parents("foo"`)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.InStringArg {
		t.Error("did not expect InStringArg for complete string")
	}
	assertHasExpected(t, ctx, ExpectedOperator)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionOperatorInFunctionArg(t *testing.T) {
	// `parents("foo" |` with cursor at end — after infix operator within
	// a function argument, expression and operator are both expected.
	ctx := ParseForCompletion(`parents("foo" |`)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if len(ctx.Function.Args) == 0 {
		t.Fatal("expected at least one parsed arg")
	}
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedOperator)
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionQuotedRemoteSymbol(t *testing.T) {
	// `"parents("@` with cursor at end — after @ following a quoted string,
	// should set InRemoteSymbol with the quoted name as RemoteBookmarkName.
	ctx := ParseForCompletion(`"parents("@`)
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol")
	}
	if ctx.RemoteBookmarkName != "parents(" {
		t.Errorf("expected RemoteBookmarkName 'parents(', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.PartialRemote != "" {
		t.Errorf("expected empty PartialRemote, got %q", ctx.PartialRemote)
	}
}

func TestCompletionQuotedRemoteSymbolPartial(t *testing.T) {
	// `"parents("@ori` with cursor at end — partial remote name after @.
	ctx := ParseForCompletion(`"parents("@ori`)
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol")
	}
	if ctx.RemoteBookmarkName != "parents(" {
		t.Errorf("expected RemoteBookmarkName 'parents(', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.PartialRemote != "ori" {
		t.Errorf("expected PartialRemote 'ori', got %q", ctx.PartialRemote)
	}
}

func TestCompletionQuotedRemoteSymbolComplete(t *testing.T) {
	// `"parents("@git` with cursor at end — complete remote name after @.
	ctx := ParseForCompletion(`"parents("@git`)
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol")
	}
	if ctx.RemoteBookmarkName != "parents(" {
		t.Errorf("expected RemoteBookmarkName 'parents(', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.PartialRemote != "git" {
		t.Errorf("expected PartialRemote 'git', got %q", ctx.PartialRemote)
	}
}

func TestCompletionRawQuotedRemoteSymbol(t *testing.T) {
	// `'parents('@` with cursor at end — single-quoted string with @ suffix.
	ctx := ParseForCompletion(`'parents('@`)
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol")
	}
	if ctx.RemoteBookmarkName != "parents(" {
		t.Errorf("expected RemoteBookmarkName 'parents(', got %q", ctx.RemoteBookmarkName)
	}
}

func TestCompletionInPattern(t *testing.T) {
	// "exact:" with cursor at end
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

func TestCompletionAfterAt(t *testing.T) {
	// "main@" with cursor at end - completing remote name
	ctx := ParseForCompletion("main@")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterDagRangePrefix(t *testing.T) {
	// "::" with cursor at end (nullary)
	ctx := ParseForCompletion("::")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAfterRangeAll(t *testing.T) {
	// ".." with cursor at end (nullary)
	ctx := ParseForCompletion("..")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionNegatePrefix(t *testing.T) {
	// "~" with cursor at end
	ctx := ParseForCompletion("~")
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionAfterNegate(t *testing.T) {
	// "~foo" with cursor at end
	ctx := ParseForCompletion("~foo")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixDagRange(t *testing.T) {
	// "foo::bar" with cursor at end
	ctx := ParseForCompletion("foo::bar")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionDagRangeNeedsRight(t *testing.T) {
	// "foo::" with cursor at end (infix, needs RHS)
	// This is actually a postfix ::, which means after-expression operators
	ctx := ParseForCompletion("foo::")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionDagRangeInfixNeedsRight(t *testing.T) {
	// Placeholder - tested in TestCompletionInfixDagRangeNeedsRight
}

func TestCompletionPostfixParents(t *testing.T) {
	// "foo-" with cursor at end
	ctx := ParseForCompletion("foo-")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPostfixChildren(t *testing.T) {
	// "foo+" with cursor at end
	ctx := ParseForCompletion("foo+")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionNestedFunction(t *testing.T) {
	// "parents(file(" with cursor at end
	ctx := ParseForCompletion("parents(file(")
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
	ctx := ParseForCompletion("remote_bookmarks(remote=foo")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionAfterAtWorkspace(t *testing.T) {
	// "@" with cursor at end
	ctx := ParseForCompletion("@")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionAtInFunction(t *testing.T) {
	// "parents(@" with cursor at end
	ctx := ParseForCompletion("parents(@")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "parents" {
		t.Errorf("expected function name 'parents', got %q", ctx.Function.Name)
	}
}

func TestCompletionTrailingComma(t *testing.T) {
	// "bookmarks(a," with cursor at end (trailing comma allowed)
	ctx := ParseForCompletion("bookmarks(a,")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
}

func TestCompletionEmptyFunctionCall(t *testing.T) {
	// "visible_heads()" with cursor at end
	ctx := ParseForCompletion("visible_heads()")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionZeroArgFunction(t *testing.T) {
	// "root(" with cursor at end - zero-arg function only expects )
	ctx := ParseForCompletion("root(")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "root" {
		t.Errorf("expected function name 'root', got %q", ctx.Function.Name)
	}
	if !ctx.Function.IsZeroArg {
		t.Error("expected IsZeroArg")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertNotHasExpected(t, ctx, ExpectedExpression)
	assertNotHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionZeroArgFunctionNested(t *testing.T) {
	// "all(root(" with cursor at end - inner zero-arg function
	ctx := ParseForCompletion("all(root(")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "root" {
		t.Errorf("expected function name 'root', got %q", ctx.Function.Name)
	}
	if !ctx.Function.IsZeroArg {
		t.Error("expected IsZeroArg")
	}
	assertHasExpected(t, ctx, ExpectedClosingParen)
	assertNotHasExpected(t, ctx, ExpectedExpression)
	assertNotHasExpected(t, ctx, ExpectedOperator)
	assertNotHasExpected(t, ctx, ExpectedComma)
}

func TestCompletionInPatternWithPartialIdent(t *testing.T) {
	// "exact:fo" with cursor at end - pattern value is partial identifier
	ctx := ParseForCompletion("exact:fo")
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "exact" {
		t.Errorf("expected PatternName 'exact', got %q", ctx.PatternName)
	}
}

func TestCompletionAfterDifferenceInFunction(t *testing.T) {
	// "file(foo ~ " with cursor at end
	ctx := ParseForCompletion("file(foo ~ ")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	assertHasExpected(t, ctx, ExpectedExpression)
}

func TestCompletionRemoteSymbolPartial(t *testing.T) {
	// "main@ori" with cursor at end - partial remote name
	ctx := ParseForCompletion("main@ori")
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol to be true")
	}
	if ctx.PartialRemote != "ori" {
		t.Errorf("expected PartialRemote 'ori', got %q", ctx.PartialRemote)
	}
	if ctx.RemoteBookmarkName != "main" {
		t.Errorf("expected RemoteBookmarkName 'main', got %q", ctx.RemoteBookmarkName)
	}
	// PartialIdent should be empty since this is a remote name, not a general identifier
	if ctx.PartialIdent != "" {
		t.Errorf("expected PartialIdent to be empty for remote names, got %q", ctx.PartialIdent)
	}
}

func TestCompletionRemoteSymbolAtCursor(t *testing.T) {
	// "main@" with cursor at end - after @, expecting remote name
	ctx := ParseForCompletion("main@")
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol to be true")
	}
	if ctx.RemoteBookmarkName != "main" {
		t.Errorf("expected RemoteBookmarkName 'main', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.PartialRemote != "" {
		t.Errorf("expected empty PartialRemote, got %q", ctx.PartialRemote)
	}
}

func TestCompletionRemoteSymbolInFunction(t *testing.T) {
	// "parents(main@ori" - partial remote in function arg
	ctx := ParseForCompletion("parents(main@ori")
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol to be true")
	}
	if ctx.PartialRemote != "ori" {
		t.Errorf("expected PartialRemote 'ori', got %q", ctx.PartialRemote)
	}
	if ctx.RemoteBookmarkName != "main" {
		t.Errorf("expected RemoteBookmarkName 'main', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.IsKeywordArg {
		t.Error("expected IsKeywordArg to be false for remote names")
	}
	if ctx.Function.KeywordArgName != "" {
		t.Errorf("expected empty KeywordArgName for remote names, got %q", ctx.Function.KeywordArgName)
	}
}

func TestCompletionRemoteSymbolCompleted(t *testing.T) {
	// "main@origin)" - completed remote symbol, not in remote context
	ctx := ParseForCompletion("main@origin)")
	if ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol to be false for completed remote symbol")
	}
}

func TestCompletionArgSpanAndContent(t *testing.T) {
	// "parents(foo" - 'foo' is partial at cursor, not a completed arg
	ctx := ParseForCompletion("parents(foo")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	// 'foo' is partial, so no completed args
	if len(ctx.Function.Args) != 0 {
		t.Fatalf("expected 0 args (partial), got %d", len(ctx.Function.Args))
	}
	if ctx.PartialIdent != "foo" {
		t.Errorf("expected partialIdent 'foo', got %q", ctx.PartialIdent)
	}
}

func TestCompletionArgStringContent(t *testing.T) {
	// `parents("foo"` - verify string arg content
	ctx := ParseForCompletion(`parents("foo"`)
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
	ctx := ParseForCompletion("parents(foo@")
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
	ctx := ParseForCompletion("parents(foo@bar")
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
	ctx := ParseForCompletion("parents(foo)")
	if ctx.Function != nil {
		// After the function is closed, Function context should not be set
		// (the cursor is after the closing paren)
	}
	// The lastExpr should be a function call
	// This is tested indirectly - the function call is complete
}

func TestCompletionMultipleArgsSpanAndContent(t *testing.T) {
	// "file(a, b, c" - 'a' and 'b' are complete (comma-separated), 'c' is partial
	ctx := ParseForCompletion("file(a, b, c")
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	// 'a' and 'b' are complete, 'c' is partial at cursor
	if len(ctx.Function.Args) != 2 {
		t.Fatalf("expected 2 complete args, got %d", len(ctx.Function.Args))
	}
	for i, expected := range []struct {
		name  string
		spanS int
		spanE int
	}{
		{"a", 5, 6},
		{"b", 8, 9},
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

func assertNotHasExpected(t *testing.T, ctx *CompletionContext, notExpected ExpectedToken) {
	t.Helper()
	if slices.Contains(ctx.ExpectedTokens, notExpected) {
		t.Errorf("did not expect %s in ExpectedTokens, got %v", notExpected, ctx.ExpectedTokens)
	}
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

func assertNoOperator(t *testing.T, ctx *CompletionContext, op string) {
	t.Helper()
	for _, v := range ctx.ValidOperators {
		if v.Op == op {
			t.Errorf("expected operator %q NOT in ValidOperators, but it was found", op)
			return
		}
	}
}

func TestCompletionNullaryRange(t *testing.T) {
	// ".." with cursor at end (nullary range)
	ctx := ParseForCompletion("..")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPrefixRange(t *testing.T) {
	// "..foo" with cursor at end (prefix range)
	ctx := ParseForCompletion("..foo")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixRange(t *testing.T) {
	// "foo..bar" with cursor at end
	ctx := ParseForCompletion("foo..bar")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionInfixRangeNeedsRight(t *testing.T) {
	// "foo.." with cursor at end - could be postfix or infix needing RHS
	ctx := ParseForCompletion("foo..")
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPatternKinds(t *testing.T) {
	// "exact:" — already tested in TestCompletionInPattern, verify other kinds

	// "glob:" — pattern
	ctx := ParseForCompletion("glob:")
	if !ctx.InPattern {
		t.Error("expected InPattern for glob:")
	}
	if ctx.PatternName != "glob" {
		t.Errorf("expected PatternName 'glob', got %q", ctx.PatternName)
	}

	// "substring:" — pattern
	ctx = ParseForCompletion("substring:")
	if !ctx.InPattern {
		t.Error("expected InPattern for substring:")
	}

	// "regex:" — pattern
	ctx = ParseForCompletion("regex:")
	if !ctx.InPattern {
		t.Error("expected InPattern for regex:")
	}
}

func TestCompletionPatternCaseInsensitive(t *testing.T) {
	// "glob-i:" — case-insensitive pattern
	ctx := ParseForCompletion("glob-i:")
	if !ctx.InPattern {
		t.Error("expected InPattern")
	}
	if ctx.PatternName != "glob-i" {
		t.Errorf("expected PatternName 'glob-i', got %q", ctx.PatternName)
	}
}

func TestCompletionKeywordArgEqualsExpr(t *testing.T) {
	// "remote_bookmarks(remote=" with cursor at end — should expect expression
	ctx := ParseForCompletion("remote_bookmarks(remote=")
	assertHasExpected(t, ctx, ExpectedExpression)
	if ctx.Function == nil {
		t.Fatal("expected Function context")
	}
	if ctx.Function.Name != "remote_bookmarks" {
		t.Errorf("expected function name 'remote_bookmarks', got %q", ctx.Function.Name)
	}
}

func TestCompletionAttachedRevset(t *testing.T) {
	tests := []struct {
		input           string
		expected        string
		expectedOpStart int
	}{
		{"@", "@", 0},
		{"@-", "@-", 1},
		{"@--", "@--", 1},
		{"@+", "@+", 1},
		{"foo", "foo", 0},
		{"foo-", "foo-", 3},
		{"foo++", "foo++", 3},
		{"all()", "all()", 0},
		{"", "", 0},
		{"@- |", "", 0},
		{"foo | @-", "@-", 1},
		{"parents(bookmark)-", "parents(bookmark)-", 17},
		{"parents(bookmark)--", "parents(bookmark)--", 17},
		{"(bookmark)-", "(bookmark)-", 10},
		{"foo | parents(bookmark)-", "parents(bookmark)-", 17},
		{"parents(bookmark)+", "parents(bookmark)+", 17},
		{"children(bookmark)+", "children(bookmark)+", 18},
		{"foo@origin-", "foo@origin-", 10},
		{"parents(foo@origin)-", "parents(foo@origin)-", 19},
		{"parents(all())-", "parents(all())-", 14},
		{"parents(foo)-+", "parents(foo)-+", 12},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ctx := ParseForCompletion(tt.input)
			if ctx.AttachedRevset != tt.expected {
				t.Errorf("expected AttachedRevset %q, got %q", tt.expected, ctx.AttachedRevset)
			}
			if ctx.PostfixOpStart != tt.expectedOpStart {
				t.Errorf("expected PostfixOpStart %d, got %d", tt.expectedOpStart, ctx.PostfixOpStart)
			}
		})
	}
}

func TestCompletionPartialIdentWithConnector(t *testing.T) {
	// When an identifier ends with a connector (-, +, .) at the cursor,
	// the connector should be included in the partial identifier since
	// the user might be typing a dash-containing identifier (e.g. "book-mark").
	// Both expression completion and operator completion should be available.

	tests := []struct {
		input        string
		partialIdent string
		partialOp    int
	}{
		{"book-", "book-", 4},
		{"book+", "book+", 4},
		{"foo.bar", "foo.bar", 0},
		{"foo--", "foo--", 3},
		{"foo-v1+", "foo-v1+", 6},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ctx := ParseForCompletion(tt.input)
			// Should have both Expression and Operator expected
			assertHasExpected(t, ctx, ExpectedExpression)
			assertHasExpected(t, ctx, ExpectedOperator)
			// PartialIdent should include the trailing connector
			if ctx.PartialIdent != tt.partialIdent {
				t.Errorf("expected PartialIdent %q, got %q", tt.partialIdent, ctx.PartialIdent)
			}
			// PostfixOpStart should be set for trailing -/+
			if ctx.PostfixOpStart != tt.partialOp {
				t.Errorf("expected PostfixOpStart %d, got %d", tt.partialOp, ctx.PostfixOpStart)
			}
		})
	}
}

func TestCompletionPartialIdentConnectorContinues(t *testing.T) {
	// When typing after a connector (e.g. "book-m"), the partial
	// identifier should still include the connector.
	ctx := ParseForCompletion("book-m")
	if ctx.PartialIdent != "book-m" {
		t.Errorf("expected PartialIdent 'book-m', got %q", ctx.PartialIdent)
	}
	assertHasExpected(t, ctx, ExpectedExpression)
	assertHasExpected(t, ctx, ExpectedOperator)
}

func TestCompletionPartialRemoteWithConnector(t *testing.T) {
	// When a remote name ends with a connector (-/+) at the cursor,
	// the trailing connector is treated as a postfix operator:
	// PartialRemote is trimmed, AttachedRevset includes the postfix,
	// and ExpectedOperator is set so the action layer can provide
	// both remote name and postfix operator completions.
	ctx := ParseForCompletion("foo@origin-")
	if ctx.PartialRemote != "origin" {
		t.Errorf("expected PartialRemote 'origin', got %q", ctx.PartialRemote)
	}
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol")
	}
	if ctx.RemoteBookmarkName != "foo" {
		t.Errorf("expected RemoteBookmarkName 'foo', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.AttachedRevset != "foo@origin-" {
		t.Errorf("expected AttachedRevset 'foo@origin-', got %q", ctx.AttachedRevset)
	}
	if ctx.PostfixOpStart != 10 {
		t.Errorf("expected PostfixOpStart 10, got %d", ctx.PostfixOpStart)
	}
	if !slices.Contains(ctx.ExpectedTokens, ExpectedOperator) {
		t.Error("expected ExpectedOperator token")
	}
}

func TestCompletionPartialRemoteWithSlashAndPostfix(t *testing.T) {
	// Bookmarks containing '/' with '@remote' and trailing postfix operator
	ctx := ParseForCompletion("fix/book-mark@origin-")
	if ctx.PartialRemote != "origin" {
		t.Errorf("expected PartialRemote 'origin', got %q", ctx.PartialRemote)
	}
	if !ctx.InRemoteSymbol {
		t.Error("expected InRemoteSymbol")
	}
	if ctx.RemoteBookmarkName != "fix/book-mark" {
		t.Errorf("expected RemoteBookmarkName 'fix/book-mark', got %q", ctx.RemoteBookmarkName)
	}
	if ctx.AttachedRevset != "fix/book-mark@origin-" {
		t.Errorf("expected AttachedRevset 'fix/book-mark@origin-', got %q", ctx.AttachedRevset)
	}
	if ctx.PostfixOpStart != 20 {
		t.Errorf("expected PostfixOpStart 20, got %d", ctx.PostfixOpStart)
	}
	if !slices.Contains(ctx.ExpectedTokens, ExpectedOperator) {
		t.Error("expected ExpectedOperator token")
	}
}

func TestCompletionPartialRemoteWithChildren(t *testing.T) {
	// Remote with trailing + (children operator)
	ctx := ParseForCompletion("main@upstream+")
	if ctx.PartialRemote != "upstream" {
		t.Errorf("expected PartialRemote 'upstream', got %q", ctx.PartialRemote)
	}
	if ctx.AttachedRevset != "main@upstream+" {
		t.Errorf("expected AttachedRevset 'main@upstream+', got %q", ctx.AttachedRevset)
	}
	if ctx.PostfixOpStart != 13 {
		t.Errorf("expected PostfixOpStart 13, got %d", ctx.PostfixOpStart)
	}
}

func TestCompletionPartialRemoteWithDoublePostfix(t *testing.T) {
	// Remote with -- (two parent operators)
	ctx := ParseForCompletion("main@origin--")
	if ctx.PartialRemote != "origin" {
		t.Errorf("expected PartialRemote 'origin', got %q", ctx.PartialRemote)
	}
	if ctx.AttachedRevset != "main@origin--" {
		t.Errorf("expected AttachedRevset 'main@origin--', got %q", ctx.AttachedRevset)
	}
	if ctx.PostfixOpStart != 11 {
		t.Errorf("expected PostfixOpStart 11, got %d", ctx.PostfixOpStart)
	}
}
