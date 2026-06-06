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
			batch = append(batch, ActionRemoteBookmarks())
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

		// Compute the prefix: everything before the partial identifier being typed.
		// Sub-actions filter against c.Value, so we need to strip this prefix
		// before invoking them and re-attach it to the completion values.
		prefix := c.Value[:len(c.Value)-len(ctx.PartialIdent)]

		if ctx.InRemoteSymbol {
			// Completing a remote name after @ (e.g. "main@ori")
			// Strip the remote part from the prefix so ActionRemotes filters correctly
			if ctx.PartialRemote != "" {
				prefix = c.Value[:len(c.Value)-len(ctx.PartialRemote)]
			} else if ctx.PartialString != "" {
				// String literal remote (e.g. main@"ori)
				atIdx := strings.LastIndex(c.Value, "@")
				if atIdx >= 0 {
					prefix = c.Value[:atIdx+1]
				}
			} else {
				// Bare @ with no remote text yet (e.g. "main@")
				atIdx := strings.LastIndex(c.Value, "@")
				if atIdx >= 0 {
					prefix = c.Value[:atIdx+1]
				}
			}
			return ActionRemotes().Prefix(prefix).NoSpace()
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
	batch := carapace.Batch(
		ActionRevs(opts),
		ActionRevsetFunctions().Suffix("("),
		ActionRevsetPatterns().Suffix(":"),
		ActionSpecialSymbols(),
		ActionRevsetAliases(true),
	)

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

	if fn.IsKeywordArg && fn.KeywordArgName != "" && !strings.Contains(fn.KeywordArgName, "=") {
		return ActionRevsetKeywordArgs(fn.Name).Suffix("=")
	}

	switch fn.Name {
	case "parents", "children", "ancestors", "descendants",
		"first_parent", "first_ancestors",
		"heads", "roots", "latest", "fork_point", "bisect",
		"present", "connected", "exactly", "reachable", "coalesce":
		return carapace.Batch(
			ActionRevs(opts),
			ActionRevsetFunctions().Suffix("("),
			ActionSpecialSymbols(),
			ActionRevsetAliases(true),
		).ToA().NoSpace()

	case "author", "author_name", "author_email",
		"committer", "committer_name", "committer_email",
		"description", "subject":
		return ActionStringPatterns().Suffix(":").NoSpace()

	case "author_date", "committer_date":
		return ActionDatePatterns().Suffix(":").NoSpace()

	case "diff_lines", "diff_lines_added", "diff_lines_removed":
		if fn.IsKeywordArg && fn.KeywordArgName == "files" {
			return ActionFilesetPatterns().Suffix(":").NoSpace()
		}
		return ActionStringPatterns().Suffix(":").NoSpace()

	case "files":
		return ActionFilesetPatterns().Suffix(":").NoSpace()

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
			batch = append(batch, ActionRemoteBookmarks())
		case "tags", "remote_tags", "tracked_remote_tags", "untracked_remote_tags":
			batch = append(batch, ActionTags())
		}
		return batch.ToA().NoSpace()

	case "at_operation":
		return ActionOperations().NoSpace()

	case "change_id", "commit_id":
		return ActionRevs(opts).NoSpace()

	default:
		return carapace.Batch(
			ActionRevs(opts),
			ActionRevsetFunctions().Suffix("("),
			ActionSpecialSymbols(),
			ActionRevsetAliases(true),
		).ToA().NoSpace()
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
