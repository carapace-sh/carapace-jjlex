package jj

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/style"
)

// ActionHeadCommits completes head commits (@, @-, @--, ...).
//
//	@  (HEAD)
//	@- (commit message)
func ActionHeadCommits(limit int) carapace.Action {
	return actionExecJJ("log", "--no-graph", "--template", `description.first_line() ++ "\n"`, "--revisions", "::@", "--limit", strconv.Itoa(limit))(func(output []byte) carapace.Action {
		lines := parseLines(output)
		vals := make([]string, 0)
		for i, line := range lines {
			vals = append(vals, "@"+strings.Repeat("-", i), line)
		}
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("head commits").Style(style.Blue)
	}).UidF(Uid("commit"))
}

// ActionLocalBookmarks completes local bookmarks.
//
//	main (last commit message)
//	feature (another message)
func ActionLocalBookmarks() carapace.Action {
	return actionExecJJ("bookmark", "list")(func(output []byte) carapace.Action {
		vals := parseBookmarkValues(output, false)
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("local bookmarks").Style(style.Blue)
	}).UidF(Uid("bookmark"))
}

// ActionRemoteBookmarks completes remote bookmarks.
//
//	main@origin (last commit message)
//	develop@upstream (another message)
func ActionRemoteBookmarks() carapace.Action {
	return actionExecJJ("bookmark", "list", "--all-remotes")(func(output []byte) carapace.Action {
		vals := parseBookmarkValues(output, true)
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("remote bookmarks").Style(style.Cyan)
	}).UidF(Uid("remote-bookmark"))
}

// ActionTags completes tags.
//
//	v1.0 (release message)
//	v2.0 (another release)
func ActionTags() carapace.Action {
	return actionExecJJ("log", "--no-graph", "--revisions", "tags()", "--template", `tags ++ "\t" ++ description.first_line() ++ "\n"`)(func(output []byte) carapace.Action {
		lines := strings.Split(string(output), "\n")
		vals := make([]string, 0)
		for _, line := range lines[:len(lines)-1] {
			splitted := strings.SplitN(line, "\t", 2)
			for tag := range strings.SplitSeq(splitted[0], " ") {
				vals = append(vals, tag, splitted[1])
			}
		}
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("tags").Style(style.Yellow)
	}).UidF(Uid("change"))
}

// ActionRecentCommits completes recent commits by commit ID.
//
//	abc123 (commit message)
//	def456 (another message)
func ActionRecentCommits(limit int) carapace.Action {
	return actionExecJJ("log", "--no-graph", "--template", `commit_id.shortest() ++ " " ++ description.first_line() ++ "\n"`, "--limit", fmt.Sprintf("%d", limit))(func(output []byte) carapace.Action {
		vals := parseDescribedLines(output)
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("commits").Style(style.Dim)
	}).UidF(Uid("commit"))
}

// ActionChangeIds completes change IDs.
//
//	t (message)
//	x (another message)
func ActionChangeIds() carapace.Action {
	return actionExecJJ("log", "--no-graph", "--template", `change_id.shortest() ++ " " ++ description.first_line() ++ "\n"`)(func(output []byte) carapace.Action {
		vals := parseDescribedLines(output)
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("change ids").Style(style.Magenta)
	}).UidF(Uid("change-id"))
}

// ActionRemotes completes remote names.
//
//	origin
//	upstream
func ActionRemotes() carapace.Action {
	return actionExecJJ("git", "remote", "list")(func(output []byte) carapace.Action {
		lines := parseLines(output)
		names := make([]string, 0, len(lines))
		for _, line := range lines {
			parts := strings.SplitN(line, " ", 2)
			names = append(names, parts[0])
		}
		if len(names) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValues(names...).Tag("remotes")
	}).UidF(Uid("remote"))
}

// ActionOperations completes operation IDs.
//
//	abc123 (operation description)
func ActionOperations() carapace.Action {
	return actionExecJJ("op", "log", "--limit", "20", "--template", `id.short() ++ " " ++ description.first_line() ++ "\n"`)(func(output []byte) carapace.Action {
		vals := parseDescribedLines(output)
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("operations").Style(style.Dim)
	}).UidF(Uid("operation"))
}

// ActionAncestors completes ancestor postfix operators for a given revset.
//
//	- (message)
//	-- (message)
func ActionAncestors(revset string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		if revset == "" {
			revset = "@"
		}
		return actionExecJJ("log", "--no-graph", "--template", `description.first_line() ++ "\n"`, "--revisions", fmt.Sprintf("first_ancestors(%v)", revset), "--limit", "20")(func(output []byte) carapace.Action {
			lines := parseLines(output)
			vals := make([]string, 0)
			for i, line := range lines {
				if i == 0 {
					continue
				}
				vals = append(vals, strings.Repeat("-", i), line)
			}
			return carapace.ActionValuesDescribed(vals...).Prefix(revset).Tag("ancestors")
		})
	}).UidF(Uid("revset"))
}

// ActionDescendants completes descendant postfix operators for a given revset.
//
//	+ (message)
//	++ (message)
func ActionDescendants(revset string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		if revset == "" {
			revset = "@"
		}
		revsetArgs := make([]string, 0, 20)
		for d := 1; d <= 20; d++ {
			revsetArgs = append(revsetArgs, fmt.Sprintf("children(%v, %d)", revset, d))
		}
		revsetExpr := strings.Join(revsetArgs, " | ")
		depthChecks := ""
		for d := 20; d >= 1; d-- {
			inner := fmt.Sprintf(`self.contained_in("children(%v, %d)")`, revset, d)
			if depthChecks == "" {
				depthChecks = fmt.Sprintf(`if(%s, "%d", "unknown")`, inner, d)
			} else {
				depthChecks = fmt.Sprintf(`if(%s, "%d", %s)`, inner, d, depthChecks)
			}
		}
		tmpl := fmt.Sprintf(`%s ++ "\t" ++ description.first_line() ++ "\n"`, depthChecks)
		return actionExecJJ("log", "--no-graph", "--template", tmpl, "--revisions", revsetExpr, "--limit", "100")(func(output []byte) carapace.Action {
			lines := parseLines(output)
			byDepth := make(map[int][]string)
			maxDepth := 0
			for _, line := range lines {
				parts := strings.SplitN(line, "\t", 2)
				if len(parts) < 2 {
					continue
				}
				var d int
				if _, err := fmt.Sscanf(parts[0], "%d", &d); err != nil || d < 1 || d > 20 {
					continue
				}
				byDepth[d] = append(byDepth[d], parts[1])
				if d > maxDepth {
					maxDepth = d
				}
			}
			vals := make([]string, 0)
			for d := 1; d <= maxDepth; d++ {
				descs := byDepth[d]
				if len(descs) == 0 {
					continue
				}
				if len(descs) == 1 {
					vals = append(vals, strings.Repeat("+", d), descs[0])
				} else {
					vals = append(vals, strings.Repeat("+", d), fmt.Sprintf("%d children", len(descs)))
				}
			}
			if len(vals) == 0 {
				return carapace.ActionValues()
			}
			return carapace.ActionValuesDescribed(vals...).Prefix(revset).Tag("descendants")
		})
	}).UidF(Uid("revset"))
}
