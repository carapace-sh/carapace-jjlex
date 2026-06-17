package jj

import (
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/fixture"
	"github.com/carapace-sh/carapace/pkg/sandbox"
	"github.com/carapace-sh/carapace/pkg/style"
)

// TestActionRevsetAliasesBuiltin testes builtin revset aliases (might change between jj versions)
func TestActionRevsetAliasesBuiltin(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetAliases(true)
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)

		s.Run("").Expect(carapace.ActionValuesDescribed(
			"builtin_immutable_heads", "trunk() | tags() | untracked_remote_bookmarks()",
			"hidden", "~visible()",
			"immutable", "::(immutable_heads() | root())",
			"immutable_heads", "builtin_immutable_heads()",
			"mutable", "~immutable()",
			"trunk", "latest(\n  remote_bookmarks(exact:\"main\", exact:\"origin\") |\n  remote_bookmarks(exact:\"master\", exact:\"origin\") |\n  remote_bookmarks(exact:\"trunk\", exact:\"origin\") |\n  remote_bookmarks(exact:\"main\", exact:\"upstream\") |\n  remote_bookmarks(exact:\"master\", exact:\"upstream\") |\n  remote_bookmarks(exact:\"trunk\", exact:\"upstream\") |\n  root()\n)\n",
			"visible", "::visible_heads()",
		).Style(style.Dim).NoSpace().
			Tag("revset aliases"))

		f.ConfigSet("revset-aliases.custom", "parents(@--)")
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"builtin_immutable_heads", "trunk() | tags() | untracked_remote_bookmarks()",
			"hidden", "~visible()",
			"immutable", "::(immutable_heads() | root())",
			"immutable_heads", "builtin_immutable_heads()",
			"mutable", "~immutable()",
			"trunk", "latest(\n  remote_bookmarks(exact:\"main\", exact:\"origin\") |\n  remote_bookmarks(exact:\"master\", exact:\"origin\") |\n  remote_bookmarks(exact:\"trunk\", exact:\"origin\") |\n  remote_bookmarks(exact:\"main\", exact:\"upstream\") |\n  remote_bookmarks(exact:\"master\", exact:\"upstream\") |\n  remote_bookmarks(exact:\"trunk\", exact:\"upstream\") |\n  root()\n)\n",
			"visible", "::visible_heads()",

			// custom
			"custom", "parents(@--)",
		).Style(style.Dim).NoSpace().
			Tag("revset aliases"))
	})
}

func TestActionRevsetAliasesUser(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsetAliases(false)
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)

		f.ConfigSet("revset-aliases.custom", "parents(@--)")
		s.Run("").Expect(carapace.ActionValuesDescribed(
			"custom", "parents(@--)",
		).Style(style.Dim).NoSpace().
			Tag("revset aliases"))
	})
}

func TestActionRevsetAuthor(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")

		s.Run("author(").Expect(carapace.ActionValuesDescribed(
			"fixture", "fixture@nowhere",
			"fixture@nowhere", "fixture",
		).Prefix("author(\"").Suffix("\")").
			Tag("authors"))

		s.Run("author('").Expect(carapace.ActionValuesDescribed(
			"fixture", "fixture@nowhere",
			"fixture@nowhere", "fixture",
		).Prefix("author('").Suffix("')").
			Tag("authors"))

		s.Run("author(\"").Expect(carapace.ActionValuesDescribed(
			"fixture", "fixture@nowhere",
			"fixture@nowhere", "fixture",
		).Prefix("author(\"").Suffix("\")").
			Tag("authors"))
	})
}

func TestActionRevsetsPostfixInFunction(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")

		s.Run("parents(@-").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"-", "third commit",
				"--", "second commit",
				"---", "first commit",
			).Tag("ancestors").Prefix("parents(@"),
			carapace.ActionValues(")", ",").Prefix("parents(@-"),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"+", "children",
			).Prefix("parents(@-"),
		).ToA().NoSpace())
	})
}

func TestActionRevsetsPostfixNoDoublePrefix(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")

		s.Run("@-").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"-", "third commit",
				"--", "second commit",
				"---", "first commit",
			).Tag("ancestors").Prefix("@"),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"+", "children",
			).Prefix("@-"),
		).ToA().NoSpace())
	})
}

func TestActionRevsetsBookmarkPostfixNoDoublePrefix(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")
		f.CreateBookmark("feature-x")

		s.Run("feature-x-").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"-", "third commit",
				"--", "second commit",
				"---", "first commit",
			).Tag("ancestors").Prefix("feature-x"),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"+", "children",
			).Prefix("feature-x-"),
		).ToA().NoSpace())
	})
}

func TestActionRevsetsPostfixDescendants(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")

		s.Run("@+").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"-", "parents",
			).Prefix("@+"),
		).ToA().NoSpace())
	})
}

func TestActionRevsetsPostfixMultiLevel(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")

		s.Run("@---").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"-", "first commit",
			).Tag("ancestors").Prefix("@--"),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"+", "children",
			).Prefix("@---"),
		).ToA().NoSpace())
	})
}

func TestActionRevsetsRemoteBookmarkPostfix(t *testing.T) {
	// Verify that ActionAncestors works with remote bookmark syntax (@remote).
	// Uses the implicit @git remote which always exists in colocated repos.
	sandbox.Action(t, func() carapace.Action {
		return ActionAncestors("book-mark@git")
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")
		f.CreateBookmark("book-mark")
		f.AddRemote("origin")
		f.Run("git", "push", "--remote", "origin")

		s.Run("").Expect(carapace.ActionValuesDescribed(
			"-", "third commit",
			"--", "second commit",
			"---", "first commit",
		).Tag("ancestors").Prefix("book-mark@git"))
	})
}

func TestActionRevsetsBookmarkWithSlashPostfix(t *testing.T) {
	// Verify postfix completion works for bookmarks containing '/'.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CommitAdd("b.txt", "b", "second commit")
		f.CommitAdd("c.txt", "c", "third commit")
		f.CreateBookmark("fix/book-mark")

		s.Run("fix/book-mark-").Expect(carapace.Batch(
			carapace.ActionValuesDescribed(
				"-", "third commit",
				"--", "second commit",
				"---", "first commit",
			).Tag("ancestors").Prefix("fix/book-mark"),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"+", "children",
			).Prefix("fix/book-mark-"),
		).ToA().NoSpace())
	})
}

func TestActionRevsetsPartialString(t *testing.T) {
	// Typing a prefix that matches a bookmark needing quoting should offer
	// the bookmark. Non-simple bookmarks appear from ActionRevs alongside
	// the local bookmarks tag.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.RunGit("branch", "parents(")

		// Top-level: paren should offer parents( as a local bookmark entry
		s.Run(`paren`).Expect(carapace.ActionValuesDescribed(
			`parents(`, "first commit",
		).Tag("local bookmarks").Style(style.Blue).NoSpace())
	})
}

func TestActionRevsetsPartialStringSimpleFiltered(t *testing.T) {
	// Simple bookmarks like "main" should NOT be offered by ActionQuotedRevs
	// since jj rejects quoted simple identifiers.
	sandbox.Action(t, func() carapace.Action {
		return ActionQuotedRevs(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("main")

		// Simple bookmarks should not appear in quoted form
		s.Run(`m`).Expect(carapace.ActionValues())
	})
}

func TestActionRevsetsPartialStringInFunction(t *testing.T) {
	// Completing inside a partial string in a function arg should offer
	// bookmarks needing quoting with closing quote+paren.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.RunGit("branch", "parents(")

		// In function: parents("paren should complete to parents("parents(")
		s.Run(`parents("paren`).Expect(carapace.ActionValuesDescribed(
			`parents(`, "first commit",
		).Prefix(`parents("`).Suffix(`")`).NoSpace().
			Tag("quoted bookmarks").Style(style.Blue))
	})
}

func TestActionRevsetsPartialStringQuotedBookmark(t *testing.T) {
	// ExpectedStringClose path: when parser sees a closing quote is expected,
	// offer bookmarks needing quoting and string patterns.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("main")
		f.RunGit("branch", "parents(")

		// Top-level " should only offer bookmarks needing quoting and patterns.
		// Note: Invoke().Prefix().Suffix() pipeline strips Uids from the
		// string pattern values, so we build the expected action without them.
		patternsAction := carapace.ActionValuesDescribed(
			"exact", "Exact match",
			"exact-i", "Exact match (case-insensitive)",
			"glob", "Glob pattern match",
			"glob-i", "Glob pattern match (case-insensitive)",
			"regex", "Regular expression match",
			"regex-i", "Regular expression match (case-insensitive)",
			"substring", "Substring match (default)",
			"substring-i", "Substring match (case-insensitive)",
		).Suffix(":").Prefix(`"`).NoSpace().Tag("string patterns")

		quotedBookmarkAction := carapace.ActionValuesDescribed(
			`parents(`, "first commit",
		).Prefix(`"`).Suffix(`"`).NoSpace().Tag("quoted bookmarks").Style(style.Blue)

		s.Run(`"`).Expect(carapace.Batch(
			quotedBookmarkAction,
			patternsAction,
		).ToA())
	})
}
