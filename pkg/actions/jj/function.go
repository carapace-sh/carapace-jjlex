package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/pelletier/go-toml/v2"
)

// ActionRevsetFunctions completes revset function names.
//
//	parents (Same as x-)
//	children (Same as x+)
//	all (All visible commits)
func ActionRevsetFunctions() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		noArgs := carapace.ActionValuesDescribed(
			"all", "All visible commits and ancestors of commits explicitly mentioned",
			"builtin_immutable_heads", "Default immutable heads (trunk() | tags() | untracked_remote_bookmarks())",
			"conflicts", "Commits that have files in a conflicted state",
			"divergent", "Commits that are divergent",
			"empty", "Commits modifying no files (includes merges() without user modifications and root())",
			"hidden", "Hidden commits (empty unless hidden revisions are mentioned)",
			"immutable", "Commits that jj treats as immutable",
			"immutable_heads", "Heads of the set of immutable commits",
			"merges", "Merge commits",
			"mine", "Commits where the author's email matches the email of the current user",
			"mutable", "Commits that jj treats as mutable",
			"none", "No commits",
			"root", "The virtual commit that is the oldest ancestor of all other commits",
			"signed", "Commits that are cryptographically signed",
			"trunk", "Head commit for the default bookmark of the default remote",
			"visible", "Visible commits (equal to all() unless hidden revisions are mentioned)",
			"visible_heads", "All visible heads (same as heads(all()))",
			"working_copies", "The working copy commits across all the workspaces",
		).Uid("jj", "revset-function", "args", "false")

		withArgs := carapace.ActionValuesDescribed(
			"ancestors", "Ancestors of x, optionally limited by depth",
			"at_operation", "Evaluate x at the specified operation",
			"author", "Commits with the author's name or email matching pattern",
			"author_date", "Commits with author dates matching date pattern",
			"author_email", "Commits with the author's email matching pattern",
			"author_name", "Commits with the author's name matching pattern",
			"bisect", "Commits where about half the input set are descendants",
			"bookmarks", "All local bookmark targets, optionally filtered by pattern",
			"change_id", "Commits with the given change ID prefix",
			"children", "Same as x+, optionally limited by depth",
			"coalesce", "Commits in the first non-none revset from a list",
			"commit_id", "Commits with the given commit ID prefix",
			"committer", "Commits with the committer's name or email matching pattern",
			"committer_date", "Commits with committer dates matching date pattern",
			"committer_email", "Commits with the committer's email matching pattern",
			"committer_name", "Commits with the committer's name matching pattern",
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
			"fork_point", "The fork point of all commits in x",
			"heads", "Commits in x that are not ancestors of other commits in x",
			"latest", "Latest count commits by committer timestamp",
			"parents", "Same as x-, optionally limited by depth",
			"present", "Same as x, but evaluates to none() if any commit doesn't exist",
			"reachable", "All commits reachable from srcs within domain",
			"remote_bookmarks", "All remote bookmark targets, optionally filtered",
			"remote_tags", "All remote tag targets, optionally filtered",
			"roots", "Commits in x that are not descendants of other commits in x",
			"subject", "Commits with subject (first line of description) matching pattern",
			"tags", "All tag targets, optionally filtered by pattern",
			"tracked_remote_bookmarks", "Targets of tracked remote bookmarks",
			"tracked_remote_tags", "Targets of tracked remote tags",
			"untracked_remote_bookmarks", "Targets of untracked remote bookmarks",
			"untracked_remote_tags", "Targets of untracked remote tags",
		).Uid("jj", "revset-function", "args", "true")

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
				"::", "x::: Descendants of x; x::y: Ancestors of y reachable from x; :: All visible commits",
				"..", "x..: Non-ancestors of x; x..y: Ancestors of y not ancestors of x; .. All visible commits excluding root",
			))
		} else {
			batch = append(batch, carapace.ActionValuesDescribed(
				"::", "::x: Ancestors of x; :: All visible commits",
				"..", "..x: Ancestors of x excluding root; .. All visible commits excluding root",
				"~", "~x: Revisions not in x",
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
		"exact", "Exact match",
		"exact-i", "Exact match (case-insensitive)",
		"substring", "Substring match (default)",
		"substring-i", "Substring match (case-insensitive)",
		"glob", "Glob pattern match",
		"glob-i", "Glob pattern match (case-insensitive)",
		"regex", "Regular expression match",
		"regex-i", "Regular expression match (case-insensitive)",
	).Uid("jj", "revset-pattern").Tag("string patterns")
}

// ActionStringPatterns completes string pattern prefixes for revset functions.
//
//	exact: (exact match)
//	glob: (glob pattern)
func ActionStringPatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"exact", "Exact match",
		"exact-i", "Exact match (case-insensitive)",
		"substring", "Substring match (default)",
		"substring-i", "Substring match (case-insensitive)",
		"glob", "Glob pattern match",
		"glob-i", "Glob pattern match (case-insensitive)",
		"regex", "Regular expression match",
		"regex-i", "Regular expression match (case-insensitive)",
	).Uid("jj", "revset-string-pattern").Tag("string patterns")
}

// ActionDatePatterns completes date pattern prefixes for date-matching revset functions.
//
//	after: (after date)
//	before: (before date)
func ActionDatePatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"after", "Matches dates at or after the given date",
		"before", "Matches dates before (not including) the given date",
	).Uid("jj", "revset-date-pattern").Tag("date patterns")
}

// ActionFilesetPatterns completes fileset pattern prefixes.
//
//	cwd: (cwd-relative path prefix)
//	glob: (cwd-relative glob pattern)
func ActionFilesetPatterns() carapace.Action {
	return carapace.ActionValuesDescribed(
		"cwd", "Cwd-relative path prefix (file or directory)",
		"file", "Cwd-relative exact file path (alias for cwd-file)",
		"cwd-file", "Cwd-relative exact file path",
		"glob", "Cwd-relative glob pattern (alias for cwd-glob)",
		"cwd-glob", "Cwd-relative glob pattern",
		"prefix-glob", "Cwd-relative prefix-glob pattern (alias for cwd-prefix-glob)",
		"cwd-prefix-glob", "Cwd-relative prefix-glob pattern",
		"root", "Workspace-relative path prefix (file or directory)",
		"root-file", "Workspace-relative exact file path",
		"root-glob", "Workspace-relative glob pattern",
		"root-prefix-glob", "Workspace-relative prefix-glob pattern",
		"glob-i", "Cwd-relative glob pattern, case-insensitive (alias for cwd-glob-i)",
		"cwd-glob-i", "Cwd-relative glob pattern (case-insensitive)",
		"prefix-glob-i", "Cwd-relative prefix-glob pattern, case-insensitive (alias for cwd-prefix-glob-i)",
		"cwd-prefix-glob-i", "Cwd-relative prefix-glob pattern (case-insensitive)",
		"root-glob-i", "Workspace-relative glob pattern (case-insensitive)",
		"root-prefix-glob-i", "Workspace-relative prefix-glob pattern (case-insensitive)",
	).Uid("jj", "fileset-pattern").Tag("fileset patterns")
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
	}).Tag("revset aliases").NoSpace().UidF(Uid("revset"))
}

// ActionRevsetKeywordArgs completes keyword argument names for revset functions.
//
//	remote= (filter by remote name)
func ActionRevsetKeywordArgs(funcName string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		keywords := revsetKeywordArgs(funcName)
		if len(keywords) == 0 {
			return carapace.ActionValues()
		}
		vals := make([]string, 0, len(keywords)*2)
		for _, kw := range keywords {
			vals = append(vals, kw.name, kw.description)
		}
		return carapace.ActionValuesDescribed(vals...)
	}).Tag("keyword arguments").
		Uid("jj", "revset-keyword-arg", "fn", funcName)
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
	case "diff_lines", "diff_lines_added", "diff_lines_removed":
		return []keywordArg{{name: "files", description: "Narrow search to fileset expression"}}
	default:
		return nil
	}
}

func parseTomlAliases(output []byte, topLevelKey string) carapace.Action {
	var config map[string]map[string]string
	if err := toml.Unmarshal(output, &config); err != nil {
		return carapace.ActionMessage(err.Error())
	}
	aliases, ok := config[topLevelKey]
	if !ok || len(aliases) == 0 {
		return carapace.ActionValues()
	}
	vals := make([]string, 0, len(aliases)*2)
	for name, val := range aliases {
		displayName := cleanAliasName(name)
		vals = append(vals, displayName, val)
	}
	return carapace.ActionValuesDescribed(vals...)
}

func cleanAliasName(name string) string {
	if idx := strings.Index(name, "("); idx >= 0 {
		name = name[:idx]
	}
	if idx := strings.Index(name, ":"); idx >= 0 {
		name = name[:idx]
	}
	return name
}
