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
		if ctx.PartialString != "" {
			// We're inside a string literal — find the opening quote
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
			} else if ctx.PartialString != "" {
				// String literal remote (e.g. main@"ori)
				atIdx := strings.LastIndex(c.Value, "@")
				if atIdx >= 0 {
					prefix = c.Value[:atIdx+1]
				}
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
			return actionForPatternValue(ctx).Prefix(prefix)
		}

		// Compute postfix actions and suppressed operators early so we can
		// filter operators when building the main action.
		_, suppressOps := postfixActions(ctx)

		if ctx.Function != nil {
			fnAction := actionForFunctionArg(ctx, opts)
			// When inside a function but the cursor is after a postfix operator
			// on an argument (e.g. "parents(bookmark-)"), also include
			// operator and postfix actions so the user can continue the
			// postfix chain or close the function call.
			if expectsToken(ctx, revset.ExpectedOperator) && hasPostfixOps(ctx) {
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
			// When inside a partial string argument of a function that doesn't
			// handle string quoting itself (e.g. parents("paren), add closing
			// quotes and paren to the completions.
			// String-aware functions (author, committer, etc.) already add
			// closing quote+paren in their actionForFunctionArg, so skip them.
			if ctx.Function.InStringArg && expectsToken(ctx, revset.ExpectedStringClose) && !isFunctionStringAware(ctx.Function.Name) {
				quote := string(ctx.StringQuote)
				closeSuffix := quote
				if expectsToken(ctx, revset.ExpectedClosingParen) {
					closeSuffix += ")"
				}
				// Offer identifiers/bookmarks wrapped in quotes with function
				// context (e.g. parents("paren → "parents(")").
				identAction := carapace.ActionCallback(func(c carapace.Context) carapace.Action {
					c.Value = ctx.PartialString
					return fnAction.Invoke(c).Prefix(prefix).Suffix(closeSuffix).ToA().NoSpace()
				})
				// Also offer string patterns (e.g. exact:, regex:) as string content.
				patternAction := carapace.ActionCallback(func(c carapace.Context) carapace.Action {
					c.Value = ctx.PartialString
					patternSuffix := ":"
					if ctx.PartialString != "" {
						patternSuffix = ":" + ctx.PartialString
					}
					return ActionStringPatterns().Suffix(patternSuffix).Invoke(c).Prefix(prefix).Suffix(quote).ToA().NoSpace()
				})
				return carapace.Batch(identAction, patternAction).ToA()
			}
			return fnAction.Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedPatternValue) {
			return ActionStringPatterns().Suffix(":").NoSpace().Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedStringClose) && ctx.StringQuote != 0 {
			quote := string(ctx.StringQuote)
			// Offer identifiers (bookmarks, commits, etc.) that match the partial
			// string content, wrapped in quotes to form valid quoted identifier references.
			identAction := carapace.ActionCallback(func(c carapace.Context) carapace.Action {
				c.Value = ctx.PartialString
				return ActionRevs(opts).Invoke(c).Prefix(prefix).Suffix(quote).ToA().NoSpace()
			})
			// Offer string patterns (e.g. exact:, regex:) as content of the string.
			patternAction := carapace.ActionCallback(func(c carapace.Context) carapace.Action {
				c.Value = ctx.PartialString
				patternSuffix := ":"
				if ctx.PartialString != "" {
					patternSuffix = ":" + ctx.PartialString
				}
				return ActionStringPatterns().Suffix(patternSuffix).Invoke(c).Prefix(prefix).ToA().NoSpace()
			})
			return carapace.Batch(identAction, patternAction).ToA()
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

// isFunctionStringAware returns true for functions that handle partial
// string argument quoting in their actionForFunctionArg case (e.g. author,
// committer). These functions already add closing quotes and parens to their
// completion values, so the caller should not add them again.
func isFunctionStringAware(name string) bool {
	switch name {
	case "author", "author_name", "author_email",
		"committer", "committer_name", "committer_email",
		"description", "subject":
		return true
	}
	return false
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

func actionForFunctionArg(ctx *revset.CompletionContext, opts RevOpts) carapace.Action {
	fn := ctx.Function

	if fn.IsZeroArg {
		return carapace.ActionValues(")")
	}

	if fn.IsKeywordArg && fn.KeywordArgName != "" && !strings.Contains(fn.KeywordArgName, "=") {
		return ActionRevsetKeywordArgs(fn.Name).Suffix("=")
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
