package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
)

// ActionRevsetFunctions completes revset function names.
//
//	parents (Same as x-)
//	children (Same as x+)
func ActionRevsetFunctions(brackets bool) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		noArgs := carapace.ActionValuesDescribed(
			"all", "All visible commits and ancestors of explicitly mentioned commits",
			"conflicts", "Commits with conflicted files",
			"divergent", "Divergent commits",
			"empty", "Commits modifying no files (includes merges() without user modifications and root())",
			"merges", "Merge commits (2+ parents)",
			"mine", "Commits where author email matches current user",
			"none", "No commits",
			"root", "The virtual root commit",
			"signed", "Cryptographically signed commits",
			"visible_heads", "All visible heads (same as heads(all()))",
			"working_copies", "Working copy commits across all workspaces",
		).Uid("jj", "revset-function-noargs")

		withArgs := carapace.ActionValuesDescribed(
			"ancestors", "Ancestors of x, optionally limited by depth",
			"at_operation", "Evaluate x at the specified operation",
			"author", "Commits with author name or email matching pattern",
			"author_date", "Commits with author date matching date pattern",
			"author_email", "Commits with author email matching pattern",
			"author_name", "Commits with author name matching pattern",
			"bisect", "Commits where about half the input set are descendants",
			"bookmarks", "All local bookmark targets, optionally filtered by pattern",
			"change_id", "Commits with given change ID prefix",
			"children", "Same as x+, optionally limited by depth",
			"coalesce", "First non-none revset from a list",
			"commit_id", "Commits with given commit ID prefix",
			"committer", "Commits with committer name or email matching pattern",
			"committer_date", "Commits with committer date matching date pattern",
			"committer_email", "Commits with committer email matching pattern",
			"committer_name", "Commits with committer name matching pattern",
			"connected", "Same as x::x",
			"descendants", "Same as x::, optionally limited by depth",
			"description", "Commits with description matching pattern",
			"diff_lines", "Commits with diffs matching text pattern",
			"diff_lines_added", "Like diff_lines() but matches only added lines",
			"diff_lines_removed", "Like diff_lines() but matches only removed lines",
			"exactly", "Returns x if exactly count commits, otherwise errors",
			"files", "Commits modifying paths matching fileset expression",
			"first_ancestors", "Like ancestors() but only traverses first parent",
			"first_parent", "Like parents() but for merges returns only first parent",
			"fork_point", "Common ancestor(s) with no descendant that is also a common ancestor",
			"heads", "Commits in x that are not ancestors of other commits in x",
			"latest", "Latest count commits by committer timestamp",
			"parents", "Same as x-, optionally limited by depth",
			"present", "Same as x, but evaluates to none() if any commit doesn't exist",
			"reachable", "All commits reachable from srcs within domain",
			"remote_bookmarks", "All remote bookmark targets, optionally filtered",
			"roots", "Commits in x that are not descendants of other commits in x",
			"subject", "Commits with subject (first line of description) matching pattern",
			"tags", "All tag targets, optionally filtered by pattern",
			"tracked_remote_bookmarks", "Targets of tracked remote bookmarks",
			"tracked_remote_tags", "Targets of tracked remote tags",
			"untracked_remote_bookmarks", "Targets of untracked remote bookmarks",
			"untracked_remote_tags", "Targets of untracked remote tags",
		).Uid("jj", "revset-function-withargs")

		if brackets {
			return carapace.Batch(noArgs.Suffix("()"), withArgs.Suffix("(")).ToA()
		}
		return carapace.Batch(noArgs, withArgs).ToA()
	}).Tag("revset functions")
}

// ActionRevsetOperators completes revset operators.
//
//	| (union)
//	& (intersection)
func ActionRevsetOperators(attached bool) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		batch := carapace.Batch()

		if attached {
			batch = append(batch, carapace.ActionValuesDescribed(
				"-", "x-: Parents of x (repeatable)",
				"+", "x+: Children of x (repeatable)",
				"::", "x::: Descendants of x (inclusive); x::y: Ancestors of y reachable from x",
				"..", "x..: Non-ancestors of x; x..y: Ancestors of y not ancestors of x",
			))
		} else {
			batch = append(batch, carapace.ActionValuesDescribed(
				"::", "::x: Ancestors of x (inclusive); prefix DAG range",
				"..", "..x: Ancestors of x excluding root; prefix range",
				"~", "~x: Revisions not in x; prefix negate",
			))
		}

		batch = append(batch, carapace.ActionValuesDescribed(
			"&", "x & y: Intersection (both x and y)",
			"|", "x | y: Union (either x or y)",
			"~", "x ~ y: Difference (in x but not in y)",
		))

		return batch.ToA().Uid("jj", "revset-operator")
	}).Tag("revset operators")
}

// ActionRevsetPatterns completes revset string pattern prefixes.
//
//	exact: (exact match)
//	glob: (glob pattern)
func ActionRevsetPatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"exact:", "Exact match",
		"exact-i:", "Exact match (case-insensitive)",
		"substring:", "Substring match (default)",
		"substring-i:", "Substring match (case-insensitive)",
		"glob:", "Glob pattern match",
		"glob-i:", "Glob pattern match (case-insensitive)",
		"regex:", "Regular expression match",
		"regex-i:", "Regular expression match (case-insensitive)",
	).Suffix("").Tag("string patterns").Uid("jj", "revset-pattern")
}

// ActionStringPatterns completes string pattern prefixes for revset functions.
//
//	exact: (exact match)
//	glob: (glob pattern)
func ActionStringPatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"exact:", "Exact match",
		"exact-i:", "Exact match (case-insensitive)",
		"substring:", "Substring match (default)",
		"substring-i:", "Substring match (case-insensitive)",
		"glob:", "Glob pattern match",
		"glob-i:", "Glob pattern match (case-insensitive)",
		"regex:", "Regular expression match",
		"regex-i:", "Regular expression match (case-insensitive)",
	).Suffix("").Tag("string patterns").Uid("jj", "revset-string-pattern")
}

// ActionDatePatterns completes date pattern prefixes for date-matching revset functions.
//
//	after: (after date)
//	before: (before date)
func ActionDatePatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"after:", "Matches dates at or after the given date",
		"before:", "Matches dates before (not including) the given date",
	).Suffix("").Tag("date patterns").Uid("jj", "revset-date-pattern")
}

// ActionFilesetPatterns completes fileset pattern prefixes.
//
//	exact: (exact match)
//	glob: (glob pattern)
func ActionFilesetPatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"exact:", "Exact match",
		"exact-i:", "Exact match (case-insensitive)",
		"substring:", "Substring match",
		"substring-i:", "Substring match (case-insensitive)",
		"glob:", "Glob pattern match",
		"glob-i:", "Glob pattern match (case-insensitive)",
		"regex:", "Regular expression match",
		"regex-i:", "Regular expression match (case-insensitive)",
	).Suffix("").Tag("fileset patterns").Uid("jj", "revset-fileset-pattern")
}

// ActionSpecialSymbols completes revset special symbols.
//
//	@ (current working copy)
func ActionSpecialSymbols() carapace.Action {
	return carapace.ActionValuesDescribed(
		"@", "Current working copy commit",
	).Tag("special symbols").Uid("jj", "revset-symbol")
}

// ActionRevsetAliases completes revset aliases from jj config.
//
//	trunk() (main@origin)
//	immutable() (::(immutable_heads() | root()))
func ActionRevsetAliases(includeDefaults bool) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		args := []string{"config", "list", "revset-aliases"}
		if includeDefaults {
			args = append(args, "--include-defaults")
		}
		return actionExecJJ(args...)(func(output []byte) carapace.Action {
			return parseTomlAliases(output, "revset-aliases")
		})
	}).Tag("revset aliases").Uid("jj", "revset-alias")
}

// ActionRevsetKeywordArgs completes keyword argument names for revset functions.
//
//	remote= (filter by remote name)
func ActionRevsetKeywordArgs(funcName string) carapace.Action {
	keywords := revsetKeywordArgs(funcName)
	if len(keywords) == 0 {
		return carapace.ActionValues()
	}
	vals := make([]string, 0, len(keywords)*2)
	for _, kw := range keywords {
		vals = append(vals, kw.name, kw.description)
	}
	return carapace.ActionValuesDescribed(vals...).Suffix("=").
		Tag("keyword arguments").Uid("jj", "revset-keyword-arg", "fn", funcName)
}

type keywordArg struct {
	name        string
	description string
}

func revsetKeywordArgs(funcName string) []keywordArg {
	switch funcName {
	case "remote_bookmarks", "tracked_remote_bookmarks", "untracked_remote_bookmarks",
		"remote_tags", "tracked_remote_tags", "untracked_remote_tags":
		return []keywordArg{{name: "remote", description: "Filter by remote name"}}
	default:
		return nil
	}
}

func parseTomlAliases(output []byte, topLevelKey string) carapace.Action {
	lines := strings.Split(string(output), "\n")
	vals := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if strings.HasPrefix(key, topLevelKey+".") || strings.HasPrefix(key, topLevelKey+".'") {
			aliasName := strings.TrimPrefix(key, topLevelKey+".")
			vals = append(vals, aliasName, val)
		}
	}
	if len(vals) == 0 {
		return carapace.ActionValues()
	}
	return carapace.ActionValuesDescribed(vals...)
}
