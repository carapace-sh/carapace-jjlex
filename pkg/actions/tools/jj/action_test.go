package jj

import (
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/fixture"
	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
	"github.com/carapace-sh/carapace/pkg/sandbox"
)

func TestActionSpecialSymbols(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionSpecialSymbols()
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"@", "Current working copy commit",
		).Tag("special symbols"))
	})
}

func TestActionStringPatterns(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionStringPatterns()
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"exact", "Exact match",
			"exact-i", "Exact match (case-insensitive)",
			"substring", "Substring match (default)",
			"substring-i", "Substring match (case-insensitive)",
			"glob", "Glob pattern match",
			"glob-i", "Glob pattern match (case-insensitive)",
			"regex", "Regular expression match",
			"regex-i", "Regular expression match (case-insensitive)",
		).Tag("string patterns"))
	})
}

func TestActionDatePatterns(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionDatePatterns()
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"after", "Matches dates at or after the given date",
			"before", "Matches dates before (not including) the given date",
		).Tag("date patterns"))
	})
}

func TestActionRevsetPatterns(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetPatterns()
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"exact", "Exact match",
			"exact-i", "Exact match (case-insensitive)",
			"substring", "Substring match (default)",
			"substring-i", "Substring match (case-insensitive)",
			"glob", "Glob pattern match",
			"glob-i", "Glob pattern match (case-insensitive)",
			"regex", "Regular expression match",
			"regex-i", "Regular expression match (case-insensitive)",
		).Tag("string patterns"))
	})
}

func TestActionFilesetPatterns(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionFilesetPatterns()
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
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
		).Tag("fileset patterns"))
	})
}

func TestActionRevsetKeywordArgs(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetKeywordArgs("remote_bookmarks")
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"remote", "Filter by remote name",
		).Tag("keyword arguments"))
	})
}

func TestActionRevsetKeywordArgsDiffLines(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetKeywordArgs("diff_lines")
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"files", "Narrow search to fileset expression",
		).Tag("keyword arguments"))
	})
}

func TestActionRevsetKeywordArgsNone(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetKeywordArgs("parents")
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.ActionValues())
	})
}

func TestActionRevsetOperatorsAttached(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetOperators(true)
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"-", "x-: Parents of x (repeatable)",
				"+", "x+: Children of x (repeatable)",
				"::", "x::: Descendants of x; x::y: Ancestors of y reachable from x; :: All visible commits",
				"..", "x..: Non-ancestors of x; x..y: Ancestors of y not ancestors of x; .. All visible commits excluding root",
			),
			carapace.ActionValuesDescribed(
				"&", "x & y: Intersection (both x and y)",
				"|", "x | y: Union (either x or y)",
				"~", "x ~ y: Difference (in x but not in y)",
			),
		).ToA().Tag("revset operators"))
	})
}

func TestActionRevsetOperatorsDetached(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetOperators(false)
	})(func(s *sandbox.Sandbox) {
		s.Run("").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"::", "::x: Ancestors of x; :: All visible commits",
				"..", "..x: Ancestors of x excluding root; .. All visible commits excluding root",
				"~", "~x: Revisions not in x",
			),
			carapace.ActionValuesDescribed(
				"&", "x & y: Intersection (both x and y)",
				"|", "x | y: Union (either x or y)",
				"~", "x ~ y: Difference (in x but not in y)",
			),
		).ToA().Tag("revset operators"))
	})
}

func TestRevsetKeywordArgsLogic(t *testing.T) {
	args := revsetKeywordArgs("remote_bookmarks")
	if len(args) != 1 || args[0].name != "remote" {
		t.Errorf("expected remote keyword for remote_bookmarks, got %v", args)
	}

	args = revsetKeywordArgs("parents")
	if len(args) != 0 {
		t.Errorf("expected no keywords for parents, got %v", args)
	}

	args = revsetKeywordArgs("tracked_remote_tags")
	if len(args) != 1 || args[0].name != "remote" {
		t.Errorf("expected remote keyword for tracked_remote_tags, got %v", args)
	}

	args = revsetKeywordArgs("untracked_remote_bookmarks")
	if len(args) != 1 || args[0].name != "remote" {
		t.Errorf("expected remote keyword for untracked_remote_bookmarks, got %v", args)
	}

	args = revsetKeywordArgs("diff_lines")
	if len(args) != 1 || args[0].name != "files" {
		t.Errorf("expected files keyword for diff_lines, got %v", args)
	}

	args = revsetKeywordArgs("diff_lines_added")
	if len(args) != 1 || args[0].name != "files" {
		t.Errorf("expected files keyword for diff_lines_added, got %v", args)
	}

	args = revsetKeywordArgs("diff_lines_removed")
	if len(args) != 1 || args[0].name != "files" {
		t.Errorf("expected files keyword for diff_lines_removed, got %v", args)
	}
}

func TestParseBookmarkOutput(t *testing.T) {
	output := "main: qvlkomxp 9a2e553b some commit message\nfeature: numonpmw ad5c8efd another commit\nmain@origin: qvlkomxp 9a2e553b main commit\n"
	vals := parseBookmarkValues([]byte(output), false)
	if len(vals) != 4 {
		t.Fatalf("expected 4 values (2 name/desc pairs), got %d", len(vals))
	}
	if vals[0] != "main" || vals[1] != "some commit message" {
		t.Errorf("expected 'main'/'some commit message', got %q/%q", vals[0], vals[1])
	}
	if vals[2] != "feature" || vals[3] != "another commit" {
		t.Errorf("expected 'feature'/'another commit', got %q/%q", vals[2], vals[3])
	}
}

func TestParseBookmarkOutputRemoteOnly(t *testing.T) {
	output := "main: qvlkomxp 9a2e553b local commit\nmain@origin: qvlkomxp 9a2e553b main commit\ndevelop@upstream: numonpmw ad5c8efd dev commit\n"
	vals := parseBookmarkValues([]byte(output), true)
	if len(vals) != 4 {
		t.Fatalf("expected 4 values (2 name/desc pairs), got %d", len(vals))
	}
	if vals[0] != "main@origin" || vals[1] != "main commit" {
		t.Errorf("expected 'main@origin'/'main commit', got %q/%q", vals[0], vals[1])
	}
	if vals[2] != "develop@upstream" || vals[3] != "dev commit" {
		t.Errorf("expected 'develop@upstream'/'dev commit', got %q/%q", vals[2], vals[3])
	}
}

func TestParseBookmarkOutputEmpty(t *testing.T) {
	vals := parseBookmarkValues([]byte(""), false)
	if len(vals) != 0 {
		t.Errorf("expected 0 values, got %d", len(vals))
	}
}

func TestParseLines(t *testing.T) {
	output := "origin https://github.com/foo\nupstream https://github.com/bar\n"
	lines := parseLines([]byte(output))
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "origin https://github.com/foo" {
		t.Errorf("expected 'origin https://github.com/foo', got %q", lines[0])
	}
	if lines[1] != "upstream https://github.com/bar" {
		t.Errorf("expected 'upstream https://github.com/bar', got %q", lines[1])
	}
}

func TestParseLinesWithDescriptions(t *testing.T) {
	output := "abc123 fix bug\ndef456 add feature\n"
	result := parseDescribedLines([]byte(output))
	if len(result) != 4 {
		t.Fatalf("expected 4 values (2 pairs), got %d", len(result))
	}
	if result[0] != "abc123" || result[1] != "fix bug" {
		t.Errorf("expected 'abc123'/'fix bug', got %q/%q", result[0], result[1])
	}
	if result[2] != "def456" || result[3] != "add feature" {
		t.Errorf("expected 'def456'/'add feature', got %q/%q", result[2], result[3])
	}
}

func TestParseTomlAliases(t *testing.T) {
	output := []byte("revset-aliases.'HEAD' = '@-'\nrevset-aliases.trunk = 'main@origin'\nrevset-aliases.'grep:x' = 'description(regex:x)'\nother.key = 'value'\n")
	action := parseTomlAliases(output, "revset-aliases")
	sandbox.Action(t, func() carapace.Action {
		return action
	})(func(s *sandbox.Sandbox) {
		fixture.InitT(t, s)
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"HEAD", "@-",
			"trunk", "main@origin",
			"grep", "description(regex:x)",
		))
	})
}

func TestParseTomlAliasesEmpty(t *testing.T) {
	output := []byte("")
	action := parseTomlAliases(output, "revset-aliases")
	ctx := carapace.NewContext()
	result := action.Invoke(ctx).ToA()
	_ = result
}

func TestParseLinesEmpty(t *testing.T) {
	lines := parseLines([]byte(""))
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestParseLinesTrailingNewline(t *testing.T) {
	lines := parseLines([]byte("a\nb\n"))
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestFlattenConfig(t *testing.T) {
	input := map[string]any{
		"ui": map[string]any{
			"color":  "auto",
			"editor": "hx",
		},
		"signing": map[string]any{
			"enabled": true,
		},
	}
	result := flattenConfig(input)
	if result["ui.color"] != "auto" {
		t.Errorf("expected 'auto', got %q", result["ui.color"])
	}
	if result["ui.editor"] != "hx" {
		t.Errorf("expected 'hx', got %q", result["ui.editor"])
	}
	if result["signing.enabled"] != "true" {
		t.Errorf("expected 'true', got %q", result["signing.enabled"])
	}
}

func TestFlattenConfigEmpty(t *testing.T) {
	result := flattenConfig(map[string]any{})
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestHasPostfixOps(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"@", false},
		{"bookmark", false},
		{"all()", false},
		{"bookmark-", true},
		{"bookmark+", true},
		{"bookmark--", true},
		{"bookmark++", true},
		{"parents(bookmark)-", true},
		{"parents(bookmark)+", true},
		{"parents(bookmark)", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ctx := revset.ParseForCompletion(tt.input)
			if result := hasPostfixOps(ctx); result != tt.expected {
				t.Errorf("hasPostfixOps(%q) = %v, want %v (AttachedRevset=%q)", tt.input, result, tt.expected, ctx.AttachedRevset)
			}
		})
	}
}

func TestStripDisplayQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"parents("`, `parents(`},
		{`"main"`, `main`},
		{`main`, `main`},
		{``, ``},
		{`"`, `"`},
		{`""`, ``},
		{`"unclosed`, `"unclosed`},
		{`a"b"c`, `a"b"c`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := stripDisplayQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("stripDisplayQuotes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsStringPatternFunction(t *testing.T) {
	stringPatternFuncs := []string{
		"author", "author_name", "author_email",
		"committer", "committer_name", "committer_email",
		"description", "subject",
		"diff_lines", "diff_lines_added", "diff_lines_removed",
	}
	for _, name := range stringPatternFuncs {
		t.Run(name, func(t *testing.T) {
			if !isStringPatternFunction(name) {
				t.Errorf("isStringPatternFunction(%q) = false, want true", name)
			}
		})
	}

	nonStringPatternFuncs := []string{
		"parents", "children", "ancestors", "descendants",
		"heads", "roots", "all", "none", "root",
		"bookmarks", "tags", "files",
		"change_id", "commit_id",
		"author_date", "committer_date",
		"at_operation", "coalesce", "present", "connected",
	}
	for _, name := range nonStringPatternFuncs {
		t.Run(name, func(t *testing.T) {
			if isStringPatternFunction(name) {
				t.Errorf("isStringPatternFunction(%q) = true, want false", name)
			}
		})
	}
}
