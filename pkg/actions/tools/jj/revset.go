package jj

import (
	"slices"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
)

// RevOpts configures which revision sources to include in completions.
type RevOpts struct {
	LocalBookmarks  bool
	RemoteBookmarks bool
	Commits         int
	HeadCommits     int
	Tags            bool
	ChangeIds       bool
}

func (o RevOpts) Default() RevOpts {
	o.LocalBookmarks = true
	o.RemoteBookmarks = true
	o.Commits = 100
	o.HeadCommits = 0
	o.Tags = true
	o.ChangeIds = true
	return o
}

// ActionRevs completes revision references (bookmarks, tags, commits, change IDs).
//
//	main (last commit message)
//	abc123 (another message)
func ActionRevs(opts RevOpts) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		batch := carapace.Batch()

		if opts.LocalBookmarks {
			batch = append(batch, ActionLocalBookmarks())
		}
		if opts.RemoteBookmarks {
			batch = append(batch, ActionRemoteBookmarks(""))
		}
		if opts.Commits > 0 {
			batch = append(batch, ActionRecentCommits(opts.Commits))
		}
		if opts.HeadCommits > 0 {
			batch = append(batch, ActionHeadCommits(opts.HeadCommits))
		}
		if opts.Tags {
			batch = append(batch, ActionTags())
		}
		if opts.ChangeIds {
			batch = append(batch, ActionChangeIds())
		}

		return batch.ToA()
	})
}

// ActionRevsets completes revset expressions with full context-awareness
// using the revset completion parser to determine what is expected at the cursor.
//
//	all()
//	trunk() | @-
func ActionRevsets(opts RevOpts) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		ctx := revset.ParseForCompletion(c.Value)

		// Compute the prefix: everything before the partial identifier or string being typed.
		// Sub-actions filter against c.Value, so we need to strip this prefix
		// before invoking them and re-attach it to the completion values.
		var prefix string
		if ctx.StringQuote != 0 {
			// We're inside a string literal — prefix is everything before
			// the string content (includes the opening quote)
			prefix = c.Value[:len(c.Value)-len(ctx.PartialString)]
		} else {
			prefix = c.Value[:len(c.Value)-len(ctx.PartialIdent)]
		}

		if ctx.InRemoteSymbol {
			// Completing after @ in a name@... expression.
			// This could be a remote name (e.g. "main@origin") or
			// a workspace name (e.g. "other@").
			// Strip the remote/workspace part from the prefix so actions filter correctly.
			if ctx.PartialRemote != "" {
				prefix = c.Value[:len(c.Value)-len(ctx.PartialRemote)]
			} else if ctx.StringQuote != 0 {
				// String literal remote (e.g. main@"ori) — a quoted remote
				// name is a symbol reference, so offer remotes/workspaces
				// with the opening quote as prefix and closing quote as suffix.
				atIdx := strings.LastIndex(c.Value, "@")
				if atIdx >= 0 {
					prefix = c.Value[:atIdx+1]
				}
				quote := string(ctx.StringQuote)
				remoteAction := carapace.Batch(
					ActionRemotes(),
					ActionWorkspaces(),
				).ToA().Prefix(quote).Suffix(quote).NoSpace()

				if ctx.PostfixOpStart > 0 {
					return mergeWithPostfix(remoteAction, ctx)
				}
				return remoteAction
			} else {
				// Bare @ with no text yet (e.g. "main@")
				atIdx := strings.LastIndex(c.Value, "@")
				if atIdx >= 0 {
					prefix = c.Value[:atIdx+1]
				}
			}
			remoteAction := carapace.Batch(
				ActionRemotes(),
				ActionWorkspaces(),
			).ToA().Prefix(prefix).NoSpace()

			// When the partial remote name ends with trailing -/+ that are
			// postfix operators (e.g. "fix/book-mark@origin-"), also offer
			// ancestor/descendant completions.
			if ctx.PostfixOpStart > 0 {
				return mergeWithPostfix(remoteAction, ctx)
			}
			return remoteAction
		}

		if ctx.InPattern {
			// When completing inside a quoted pattern value (e.g. exact:"foo),
			// the quoted string is a symbol reference. Offer revision symbols
			// with the opening quote as prefix and closing quote as suffix.
			if expectsToken(ctx, revset.ExpectedStringClose) && ctx.StringQuote != 0 {
				quote := string(ctx.StringQuote)
				return actionQuotedRevsetArg(opts).Prefix(quote).Suffix(quote).NoSpace()
			}
			return actionForPatternValue(ctx).Prefix(prefix)
		}

		// Compute postfix actions and suppressed operators early so we can
		// filter operators when building the main action.
		_, suppressOps := postfixActions(ctx)

		if ctx.Function != nil {
			fnAction := actionForFunctionArg(ctx, opts)
			// When inside a function and operators are expected after a
			// complete expression (not also expecting a new expression),
			// include operators, ), and , along with the function arg action.
			if expectsToken(ctx, revset.ExpectedOperator) && !expectsToken(ctx, revset.ExpectedExpression) {
				batch := carapace.Batch(fnAction)
				if expectsToken(ctx, revset.ExpectedClosingParen) {
					batch = append(batch, carapace.ActionValues(")"))
				}
				if expectsToken(ctx, revset.ExpectedComma) {
					batch = append(batch, carapace.ActionValues(","))
				}
				for _, op := range ctx.ValidOperators {
					if !suppressOps[op.Op] {
						batch = append(batch, carapace.ActionValuesDescribed(op.Op, op.Description))
					}
				}
				return mergeWithPostfix(batch.ToA().NoSpace().Prefix(prefix), ctx)
			}
			// When inside a function and both expression and operator are
			// expected after an operator within an already-started argument
			// (e.g. parents("foo" |), offer general revset expressions plus
			// operators, ), and , — not the function-arg-specific action
			// which may be empty (e.g. parents arg 1 is an integer, not a
			// revset). Only do this when at least one arg has been parsed,
			// so that first-arg completion (e.g. author() still uses the
			// function-specific action.
			if expectsToken(ctx, revset.ExpectedExpression) && expectsToken(ctx, revset.ExpectedOperator) &&
				len(ctx.Function.Args) > 0 {
				batch := carapace.Batch(
					actionExpression(opts, ctx),
					actionOperator(opts, ctx, suppressOps),
				)
				if expectsToken(ctx, revset.ExpectedClosingParen) {
					batch = append(batch, carapace.ActionValues(")"))
				}
				if expectsToken(ctx, revset.ExpectedComma) {
					batch = append(batch, carapace.ActionValues(","))
				}
				return mergeWithPostfix(batch.ToA().Prefix(prefix), ctx)
			}
			return fnAction.Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedPatternValue) {
			return ActionStringPatterns().Suffix(":").NoSpace().Prefix(prefix)
		}

		// Handle partial string literals at the top level (not inside a
		// function, pattern, or remote symbol). A quoted string in a revset
		// is a symbol reference (bookmark, tag, commit ID, change ID) —
		// e.g. "parents(" is a bookmark whose name contains brackets.
		// Invoke with the partial string content (without the opening quote)
		// so that both regular bookmarks (e.g. "main") and already-quoted
		// bookmarks (e.g. ""parents(") match correctly. Add the opening
		// quote as prefix and closing quote as suffix on the result.
		if expectsToken(ctx, revset.ExpectedStringClose) && ctx.StringQuote != 0 {
			quote := string(ctx.StringQuote)
			return actionQuotedRevsetArg(opts).Prefix(quote).Suffix(quote).NoSpace()
		}

		if expectsToken(ctx, revset.ExpectedExpression) && expectsToken(ctx, revset.ExpectedOperator) {
			// Both expression and operator are valid - combine both actions.
			// When PartialIdent includes a trailing connector (e.g. "feature-x-"),
			// the prefix is empty (the PartialIdent was stripped from c.Value).
			// Expression completions (bookmarks, etc.) need the empty prefix since
			// they produce full identifiers. But operator completions need the
			// full user input as prefix since operators are appended after it.
			operatorPrefix := prefix
			if operatorPrefix == "" && ctx.PostfixOpStart > 0 {
				operatorPrefix = c.Value
			}
			batch := carapace.Batch(
				actionExpression(opts, ctx).Prefix(prefix),
				actionOperator(opts, ctx, suppressOps).Prefix(operatorPrefix),
			)
			return mergeWithPostfix(batch.ToA(), ctx)
		}

		if expectsToken(ctx, revset.ExpectedExpression) {
			batch := carapace.Batch(actionExpression(opts, ctx))
			return mergeWithPostfix(batch.ToA().Prefix(prefix), ctx)
		}

		if expectsToken(ctx, revset.ExpectedOperator) {
			batch := carapace.Batch(actionOperator(opts, ctx, suppressOps))
			return mergeWithPostfix(batch.ToA().Prefix(prefix), ctx)
		}

		if expectsToken(ctx, revset.ExpectedClosingParen) {
			return carapace.ActionValues(")").Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedComma) {
			return carapace.ActionValues(",").Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedEquals) {
			return carapace.ActionValues("=").Prefix(prefix)
		}

		return actionExpression(opts, ctx).Prefix(prefix)
	})
}

// mergeWithPostfix combines a main action (with prefix already applied) with
// postfix actions. Postfix actions (ancestors/descendants) have their prefix
// set to the base revset (e.g. "@"), but when the AttachedRevset appears
// inside a larger expression (e.g. "parents(@-"), the values need an
// additional outer prefix prepended. mergeWithPostfix computes this
// outer prefix from the input and the AttachedRevset position.
func mergeWithPostfix(main carapace.Action, ctx *revset.CompletionContext) carapace.Action {
	postfix, _ := postfixActions(ctx)
	if len(postfix) == 0 {
		return main
	}
	postfixBatch := carapace.Batch()
	for _, a := range postfix {
		postfixBatch = append(postfixBatch, a)
	}
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		// Compute outer prefix: the part of the input before the AttachedRevset.
		// For "parents(@-", AttachedRevset="@-", outerPrefix="parents(".
		outerPrefix := ""
		attachedLen := len(ctx.AttachedRevset)
		if idx := len(c.Value) - attachedLen; idx > 0 {
			outerPrefix = c.Value[:idx]
		}
		// Invoke postfix actions with a trimmed c.Value so that the
		// .Prefix(baseRevset) on ActionAncestors matches correctly.
		// For "parents(@-", set c.Value to "@-" (the AttachedRevset).
		if attachedLen > 0 && len(c.Value) >= attachedLen {
			c.Value = c.Value[len(c.Value)-attachedLen:]
		}
		invoked := postfixBatch.ToA().NoSpace().Invoke(c)
		if outerPrefix != "" {
			invoked = invoked.Prefix(outerPrefix)
		}
		return carapace.Batch(main, invoked.ToA()).ToA()
	})
}

// hasPostfixOps returns true when the completion parser determined that
// postfix operators are present in the AttachedRevset (indicated by
// PostfixOpStart > 0). This correctly distinguishes between a dash that's
// part of an identifier (e.g. "book-" in "book-mark") and a postfix operator
// (e.g. "-" in "bookmark-").
func hasPostfixOps(ctx *revset.CompletionContext) bool {
	return ctx.PostfixOpStart > 0
}

func expectsToken(ctx *revset.CompletionContext, token revset.ExpectedToken) bool {
	return slices.Contains(ctx.ExpectedTokens, token)
}

// postfixActions returns additional actions for postfix operator completion
// based on the AttachedRevset from the completion parser. Only returns
// ancestor actions when the attached revset ends with "-", and only returns
// descendant actions when it ends with "+". This avoids the expensive
// ActionDescendants (which runs 20 jj show commands) when not needed.
//
// The base revset is computed by trimming one trailing "-"/"+" at a time.
// For example, for "@-", base="@"; for "@--", base="@-"; for "bookmark-+", base="bookmark-".
// This ensures ActionAncestors/ActionDescendants start from the correct
// parent/child level rather than from the root, so the completion suffixes
// continue from what the user already typed.
//
// The returned actions produce suffixes that continue from what was typed.
// For example, for "@-" (user typed 1 dash), the completions are "-"
// (current level), "--" (one more parent), etc.
//
// The returned actions already have the correct prefix baked in via
// .Prefix(base). The caller should NOT apply an additional
// .Prefix() to these actions.
//
// The returned suppressOps set contains operator strings that should be
// suppressed from the operator list because they are already covered by
// the postfix suffix actions (which provide commit descriptions).
func postfixActions(ctx *revset.CompletionContext) ([]carapace.Action, map[string]bool) {
	attached := ctx.AttachedRevset
	if attached == "" || ctx.PostfixOpStart == 0 {
		return nil, nil
	}
	var actions []carapace.Action
	suppressOps := make(map[string]bool)
	if hasPostfixOp(attached, '-') {
		attached = strings.TrimSuffix(attached, "-")
		actions = append(actions,
			ActionAncestors(attached).Suppress("doesn't exist"),
		)
		suppressOps["-"] = true
	}
	if hasPostfixOp(attached, '+') {
		attached = strings.TrimSuffix(attached, "+")
		actions = append(actions,
			ActionDescendants(attached).Suppress("doesn't exist"),
		)
		suppressOps["+"] = true
	}
	return actions, suppressOps
}

// hasPostfixOp returns true when the AttachedRevset ends with the given
// postfix operator character.
func hasPostfixOp(attached string, op byte) bool {
	if len(attached) == 0 {
		return false
	}
	return attached[len(attached)-1] == op
}

func actionExpression(opts RevOpts, _ *revset.CompletionContext) carapace.Action {
	batch := carapace.Batch(actionRevsetArg(opts))
	return batch.ToA().NoSpace()
}

func actionOperator(_ RevOpts, ctx *revset.CompletionContext, suppressOps map[string]bool) carapace.Action {
	// If ValidOperators is populated, filter to only those operators
	if len(ctx.ValidOperators) > 0 {
		batch := carapace.Batch()
		for _, op := range ctx.ValidOperators {
			if !suppressOps[op.Op] {
				batch = append(batch, carapace.ActionValuesDescribed(op.Op, op.Description))
			}
		}

		return batch.ToA().NoSpace()
	}

	// No valid operators from completion context
	return carapace.ActionValues()
}

// actionQuotedRevsetArg returns completions for revision symbols that are
// valid inside a quoted string literal. Unlike actionRevsetArg, this excludes
// remote bookmarks (which have @remote suffixes that don't work inside quotes)
// and revset functions/patterns (which are not symbol references).
// It uses raw (unquoted) bookmark/tag names so that the caller can add
// consistent quoting via Prefix/Suffix without double-quoting values that
// jj would otherwise display with quotes (e.g. "parents(").
func actionQuotedRevsetArg(opts RevOpts) carapace.Action {
	batch := carapace.Batch()
	if opts.LocalBookmarks {
		batch = append(batch, actionLocalBookmarksRaw())
	}
	if opts.Commits > 0 {
		batch = append(batch, ActionRecentCommits(opts.Commits))
	}
	if opts.HeadCommits > 0 {
		batch = append(batch, ActionHeadCommits(opts.HeadCommits))
	}
	if opts.Tags {
		batch = append(batch, actionTagsRaw())
	}
	if opts.ChangeIds {
		batch = append(batch, ActionChangeIds())
	}
	return batch.ToA().NoSpace()
}

// isStringPatternFunction returns true for functions whose argument is a
// string pattern (author, description, etc.) rather than a revset expression.
// For these functions, a quoted string argument is the pattern value itself,
// not a symbol reference, so InStringArg handling is done in the function's
// case branch instead of the generic InStringArg path.
func isStringPatternFunction(name string) bool {
	switch name {
	case "author", "author_name", "author_email",
		"committer", "committer_name", "committer_email",
		"description", "subject",
		"diff_lines", "diff_lines_added", "diff_lines_removed":
		return true
	}
	return false
}

func actionForFunctionArg(ctx *revset.CompletionContext, opts RevOpts) carapace.Action {
	fn := ctx.Function

	if fn.IsZeroArg {
		return carapace.ActionValues(")")
	}

	if fn.IsKeywordArg && fn.KeywordArgName != "" && !strings.Contains(fn.KeywordArgName, "=") {
		return ActionRevsetKeywordArgs(fn.Name).Suffix("=")
	}

	// When inside a quoted string argument for a function that takes a
	// revset expression (not a string pattern function like author/description),
	// the quoted string is a symbol reference (bookmark, tag, commit ID).
	// The outer Prefix(prefix) in ActionRevsets strips the function name and
	// opening quote from c.Value, so we only need to add the closing quote
	// as suffix here. The opening quote is already part of the prefix.
	if fn.InStringArg && !isStringPatternFunction(fn.Name) {
		quote := string(ctx.StringQuote)
		return actionQuotedRevsetArg(opts).Suffix(quote).NoSpace()
	}

	switch fn.Name {
	// Traversal: (x, [depth]) — arg 0 = revset, arg 1 = integer depth
	case "parents", "children", "ancestors", "descendants",
		"first_parent", "first_ancestors":
		if fn.ArgIndex == 0 {
			return actionRevsetArg(opts).NoSpace()
		}
		return carapace.ActionValues()

	// Set operations with revset arg(s)
	case "heads", "roots", "fork_point", "bisect", "present", "connected":
		return actionRevsetArg(opts).NoSpace()

	case "latest":
		if fn.ArgIndex == 0 {
			return actionRevsetArg(opts).NoSpace()
		}
		return carapace.ActionValues()

	case "exactly":
		if fn.ArgIndex == 0 {
			return actionRevsetArg(opts).NoSpace()
		}
		return carapace.ActionValues()

	case "reachable":
		return actionRevsetArg(opts).NoSpace()

	case "coalesce":
		return actionRevsetArg(opts).NoSpace()

	// Identity: string prefix
	case "change_id", "commit_id":
		return ActionRevs(opts).NoSpace()

	// String pattern functions
	case "author", "author_name", "author_email",
		"committer", "committer_name", "committer_email",
		"description", "subject":
		// Complete authors with proper quoting
		if fn.InStringArg || fn.ArgIndex == 0 {
			quote := `"`
			if ctx.StringQuote != 0 {
				quote = string(ctx.StringQuote)
			}
			// Only add prefix quote if we haven't already consumed one
			if ctx.PartialString == "" && ctx.StringQuote != 0 {
				// At opening quote like author(' — quote already typed, don't add it again
				return ActionAuthors().Suffix(quote + ")")
			}
			// No quote yet: add quote prefix and quote+parens suffix
			return ActionAuthors().Prefix(quote).Suffix(quote + ")")
		}
		return ActionStringPatterns().Suffix(":").NoSpace()

	// Date pattern functions
	case "author_date", "committer_date":
		return ActionDatePatterns().Suffix(":").NoSpace()

	// Diff functions: (text_pattern, [files=])
	case "diff_lines", "diff_lines_added", "diff_lines_removed":
		if fn.IsKeywordArg && fn.KeywordArgName == "files" {
			return ActionFilesetPatterns().Suffix(":").NoSpace()
		}
		if fn.ArgIndex == 0 {
			return ActionStringPatterns().Suffix(":").NoSpace()
		}
		return carapace.ActionValues()

	// Fileset expression
	case "files":
		return carapace.Batch(
			ActionFilesetPatterns().Suffix(":"),
			ActionRevFiles("@"),
		).ToA().NoSpace()

	// Bookmark/tag functions: ([name_pattern], [remote=remote_pattern])
	case "bookmarks", "remote_bookmarks", "tracked_remote_bookmarks", "untracked_remote_bookmarks",
		"tags", "remote_tags", "tracked_remote_tags", "untracked_remote_tags":
		if fn.ArgIndex >= 1 && !fn.IsKeywordArg {
			return ActionRemotes().NoSpace()
		}
		batch := carapace.Batch(ActionStringPatterns().Suffix(":"))
		switch fn.Name {
		case "bookmarks":
			batch = append(batch, ActionLocalBookmarks())
		case "remote_bookmarks", "tracked_remote_bookmarks", "untracked_remote_bookmarks":
			remote := keywordArgValue(fn, "remote")
			batch = append(batch, ActionRemoteBookmarks(remote))
		case "tags", "remote_tags", "tracked_remote_tags", "untracked_remote_tags":
			batch = append(batch, ActionTags())
		}
		return batch.ToA().NoSpace()

	// Operation: (op, x) — arg 0 = operation, arg 1 = revset
	case "at_operation":
		if fn.ArgIndex == 0 {
			return ActionOperations(100).NoSpace()
		}
		return actionRevsetArg(opts).NoSpace()

	default:
		return actionRevsetArg(opts).NoSpace()
	}
}

func actionForPatternValue(ctx *revset.CompletionContext) carapace.Action {
	switch ctx.PatternName {
	case "exact", "exact-i", "substring", "substring-i", "glob", "glob-i", "regex", "regex-i":
		return ActionStringPatterns().Suffix(":").NoSpace()
	case "after", "before":
		return ActionDatePatterns().Suffix(":").NoSpace()
	case "cwd", "file", "cwd-file", "prefix-glob", "cwd-prefix-glob",
		"root", "root-file", "root-glob", "root-prefix-glob",
		"cwd-glob-i", "prefix-glob-i", "cwd-prefix-glob-i",
		"root-glob-i", "root-prefix-glob-i":
		return ActionFilesetPatterns().Suffix(":").NoSpace()
	default:
		return carapace.ActionValues()
	}
}

// keywordArgValue returns the formatted value of the first keyword argument
// with the given name, or "" if not present.
func keywordArgValue(fn *revset.FunctionContext, name string) string {
	for _, ka := range fn.KeywordArgs {
		if ka.Name == name && ka.Value != nil {
			return revset.Format(ka.Value)
		}
	}
	return ""
}

// actionRevsetArg returns completions for a revset expression argument position.
func actionRevsetArg(opts RevOpts) carapace.Action {
	return carapace.Batch(
		ActionRevs(opts),
		ActionRevsetFunctions(),
		ActionRevsetPatterns().Suffix(":"),
		ActionSpecialSymbols(),
		ActionRevsetAliases(true),
		ActionWorkspaces(),
	).ToA()
}
