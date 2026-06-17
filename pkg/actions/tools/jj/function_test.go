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
	// Completing inside a partial string should offer matching identifiers
	// with proper quoting (e.g. "feature → "feature-x").
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("feature-x")

		// Top-level partial string: "feature should offer "feature-x" as a completion
		s.Run(`"feature`).Expect(carapace.ActionValuesDescribed(
			"feature-x", "(empty) (no description set)",
		).Prefix(`"`).Suffix(`"`).NoSpace().
			Style(style.Blue).
			Tag("local bookmarks"))
	})
}

func TestActionRevsetsPartialStringInFunction(t *testing.T) {
	// Completing inside a partial string in a function arg should offer
	// matching identifiers with closing quote+paren.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("feature-x")

		// In function: parents("feature should complete to parents("feature-x")
		s.Run(`parents("feature`).Expect(carapace.ActionValuesDescribed(
			"feature-x", "(empty) (no description set)",
		).Prefix(`parents("`).Suffix(`")`).NoSpace().
			Style(style.Blue).
			Tag("local bookmarks"))
	})
}

func TestActionRevsetsPartialStringQuotedBookmark(t *testing.T) {
	// Quoted identifier for a bookmark containing special characters like - which
	// can also look like a postfix operator. Verify that "feature-x- matches correctly.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{LocalBookmarks: true, RemoteBookmarks: false, Commits: 0, HeadCommits: 0, Tags: false, ChangeIds: false})
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("parents-x")

		// Top-level partial string: "paren should match bookmark "parents-x"
		s.Run(`"parents`).Expect(carapace.ActionValuesDescribed(
			"parents-x", "(empty) (no description set)",
		).Prefix(`"`).Suffix(`"`).NoSpace().
			Style(style.Blue).
			Tag("local bookmarks"))
	})
}
