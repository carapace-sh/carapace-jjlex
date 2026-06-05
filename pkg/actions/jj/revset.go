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
	o.HeadCommits = 10
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

		if ctx.InPattern {
			return actionForPatternValue(ctx)
		}

		if ctx.Function != nil {
			return actionForFunctionArg(ctx, opts)
		}

		if expectsToken(ctx, revset.ExpectedPatternValue) {
			return ActionStringPatterns().Suffix(":").NoSpace()
		}

		if expectsToken(ctx, revset.ExpectedStringClose) && ctx.PartialString != "" {
			return ActionStringPatterns().Suffix(":").Prefix(ctx.PartialString)
		}

		if expectsToken(ctx, revset.ExpectedExpression) {
			return actionExpression(opts, ctx)
		}

		if expectsToken(ctx, revset.ExpectedOperator) {
			return actionOperator(opts, ctx)
		}

		if expectsToken(ctx, revset.ExpectedClosingParen) {
			return carapace.ActionValues(")")
		}

		if expectsToken(ctx, revset.ExpectedComma) {
			return carapace.ActionValues(",")
		}

		if expectsToken(ctx, revset.ExpectedEquals) {
			return carapace.ActionValues("=")
		}

		return actionExpression(opts, ctx)
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
	if strings.HasSuffix(attached, "-") {
		actions = append(actions,
			ActionAncestors(strings.TrimSuffix(attached, "-")).Suppress("doesn't exist"),
		)
	}
	if strings.HasSuffix(attached, "+") {
		actions = append(actions,
			ActionDescendants(strings.TrimSuffix(attached, "+")).Suppress("doesn't exist"),
		)
	}
	return actions
}

func actionExpression(opts RevOpts, ctx *revset.CompletionContext) carapace.Action {
	batch := carapace.Batch(
		ActionRevs(opts),
		ActionRevsetFunctions(),
		ActionRevsetPatterns().Suffix(":"),
		ActionSpecialSymbols(),
		ActionRevsetAliases(true),
	)

	batch = append(batch, postfixActions(ctx)...)

	return batch.ToA().NoSpace()
}

func actionOperator(_ RevOpts, ctx *revset.CompletionContext) carapace.Action {
	batch := carapace.Batch(
		ActionRevsetOperators(true),
	)

	batch = append(batch, postfixActions(ctx)...)

	return batch.ToA().NoSpace()
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
			ActionRevsetFunctions(),
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
		"tags", "remote_tags":
		if fn.ArgIndex >= 1 && !fn.IsKeywordArg {
			return ActionRemotes().NoSpace()
		}
		batch := carapace.Batch(ActionStringPatterns().Suffix(":"))
		switch fn.Name {
		case "bookmarks":
			batch = append(batch, ActionLocalBookmarks())
		case "remote_bookmarks", "tracked_remote_bookmarks", "untracked_remote_bookmarks":
			batch = append(batch, ActionRemoteBookmarks())
		case "tags", "remote_tags":
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
			ActionRevsetFunctions(),
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
