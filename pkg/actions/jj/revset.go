package jj

import (
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
			return carapace.Batch(
				ActionRemotes(),
				ActionWorkspaces(),
			).ToA().Prefix(prefix).NoSpace()
		}

		if ctx.InPattern {
			return actionForPatternValue(ctx).Prefix(prefix)
		}

		if ctx.Function != nil {
			return actionForFunctionArg(ctx, opts).Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedPatternValue) {
			return ActionStringPatterns().Suffix(":").NoSpace().Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedStringClose) && ctx.PartialString != "" {
			return ActionStringPatterns().Suffix(":").Prefix(ctx.PartialString)
		}

		if expectsToken(ctx, revset.ExpectedExpression) && expectsToken(ctx, revset.ExpectedOperator) {
			// Both expression and operator are valid - combine both actions.
			// When there's a partialIdent, the user is typing an expression
			// so don't show postfix operators (attachedRevset is just the
			// partial identifier, not a completed revset).
			hasPartial := ctx.PartialIdent != ""
			batch := carapace.Batch(
				actionExpression(opts, ctx),
				actionOperator(opts, ctx, !hasPartial),
			)
			return batch.ToA().Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedExpression) {
			return actionExpression(opts, ctx).Prefix(prefix)
		}

		if expectsToken(ctx, revset.ExpectedOperator) {
			return actionOperator(opts, ctx, true).Prefix(prefix)
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

func expectsToken(ctx *revset.CompletionContext, token revset.ExpectedToken) bool {
	for _, t := range ctx.ExpectedTokens {
		if t == token {
			return true
		}
	}
	return false
}

// postfixActions returns additional actions for postfix operator completion
// based on the AttachedRevset from the completion parser. Only returns
// ActionAncestors when the attached revset ends with "-", and only returns
// ActionDescendants when it ends with "+". This avoids the expensive
// ActionDescendants (which runs 20 jj show commands) when not needed.
func postfixActions(ctx *revset.CompletionContext) []carapace.Action {
	attached := ctx.AttachedRevset
	if attached == "" {
		return nil
	}
	var actions []carapace.Action
	if before, ok := strings.CutSuffix(attached, "-"); ok {
		actions = append(actions,
			ActionAncestors(before).Suppress("doesn't exist"),
		)
	}
	if before, ok := strings.CutSuffix(attached, "+"); ok {
		actions = append(actions,
			ActionDescendants(before).Suppress("doesn't exist"),
		)
	}
	return actions
}

func actionExpression(opts RevOpts, ctx *revset.CompletionContext) carapace.Action {
	batch := carapace.Batch(actionRevsetArg(opts))

	batch = append(batch, postfixActions(ctx)...)

	return batch.ToA().NoSpace()
}

func actionOperator(_ RevOpts, ctx *revset.CompletionContext, allowPostfix bool) carapace.Action {
	// If ValidOperators is populated, filter to only those operators
	if len(ctx.ValidOperators) > 0 {
		batch := carapace.Batch()
		for _, op := range ctx.ValidOperators {
			batch = append(batch, carapace.ActionValuesDescribed(op.Op, op.Description))
		}

		attached := allowPostfix && ctx.AttachedRevset != ""
		if attached {
			batch = append(batch, postfixActions(ctx)...)
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
		if fn.InStringArg {
			return ActionAuthors().NoSpace()
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
