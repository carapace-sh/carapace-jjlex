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

func TestActionRevsetQuotedSymbol(t *testing.T) {
	// A quoted string in a revset is a symbol reference (bookmark, tag,
	// commit ID). Bookmarks with special characters like brackets are
	// displayed with quotes by jj (e.g. "parents(") and must be referenced
	// as a quoted string literal in revsets.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark(`"parents("`)

		s.Run(`"paren`).Expect(carapace.ActionValuesDescribed(
			`parents(`, "",
		).Prefix(`"`).Suffix(`"`).Style(style.Blue).Tag("local bookmarks").NoSpace())

		s.Run(`'paren`).Expect(carapace.ActionValuesDescribed(
			`parents(`, "",
		).Prefix(`'`).Suffix(`'`).Style(style.Blue).Tag("local bookmarks").NoSpace())
	})
}

func TestActionRevsetQuotedSymbolInFunction(t *testing.T) {
	// Inside a function that takes a revset expression (e.g. parents()),
	// a quoted string argument is a symbol reference, not a string pattern.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark(`"parents("`)

		s.Run(`parents("paren`).Expect(carapace.ActionValuesDescribed(
			`parents(`, "",
		).Prefix(`parents("`).Suffix(`"`).Style(style.Blue).Tag("local bookmarks").NoSpace())
	})
}

func TestActionRevsetQuotedSymbolInPattern(t *testing.T) {
	// Inside a pattern value (e.g. exact:"foo), a quoted string is a
	// symbol reference.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark(`"parents("`)

		s.Run(`exact:"paren`).Expect(carapace.ActionValuesDescribed(
			`parents(`, "",
		).Prefix(`"`).Suffix(`"`).Style(style.Blue).Tag("local bookmarks").NoSpace())
	})
}

func TestActionRevsetQuotedRemote(t *testing.T) {
	// A quoted remote name (e.g. main@"ori) is a symbol reference for
	// the remote name.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.AddRemote("origin")

		s.Run(`main@"ori`).Expect(carapace.ActionValues(
			"origin",
		).Prefix(`"`).Suffix(`"`).Tag("remotes").NoSpace())
	})
}

func TestActionRevsetQuotedRegularBookmark(t *testing.T) {
	// A regular bookmark (no special characters) should also be offered
	// inside a quoted string, with the opening/closing quote added.
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("main")
		f.CreateBookmark("patch")

		// "p should match "patch" (raw name starts with "p")
		s.Run(`"p`).Expect(carapace.ActionValuesDescribed(
			"patch", "",
		).Prefix(`"`).Suffix(`"`).Style(style.Blue).Tag("local bookmarks").NoSpace())
	})
}

func TestActionRevsetQuotedEmptyInFunction(t *testing.T) {
	// Empty quoted string inside a function (e.g. parents(") should offer
	// revision symbols with the function prefix and closing quote.
	// Use limited RevOpts to make the test deterministic (only bookmarks).
	sandbox.Action(t, func() carapace.Action {
		opts := RevOpts{}.Default()
		opts.Commits = 0
		opts.HeadCommits = 0
		opts.Tags = false
		opts.ChangeIds = false
		return ActionRevsets(opts)
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("main")

		// parents(" — empty partial string in function
		// prefix is parents(", action returns all bookmarks
		// suffix " is added, prefix parents(" is added
		s.Run(`parents("`).Expect(carapace.ActionValuesDescribed(
			"main", "",
		).Prefix(`parents("`).Suffix(`"`).Style(style.Blue).Tag("local bookmarks").NoSpace())
	})
}

