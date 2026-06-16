package jj

import (
	"strconv"
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

func TestActionRevsetsAncestorsDepth(t *testing.T) {
	sandbox.Action(t, func() carapace.Action {
		return ActionRevsets(RevOpts{}.Default())
	})(func(s *sandbox.Sandbox) {
		f := fixture.InitT(t, s)
		f.CommitAdd("a.txt", "a", "first commit")
		f.CreateBookmark("feature-x")

		depthVals := make([]string, 20)
		for i := 1; i <= 20; i++ {
			depthVals[i-1] = strconv.Itoa(i)
		}

		s.Run("ancestors(feature-x,").Expect(carapace.Batch(
			carapace.ActionValues(")").Prefix("ancestors(feature-x,"),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"-", "parents",
				"+", "children",
			).Prefix("ancestors(feature-x,"),
			carapace.ActionValues(depthVals...).Style(style.Blue).Tag("depth").Prefix("ancestors(feature-x,"),
		).ToA().NoSpace())

		s.Run("ancestors(feature-x, ").Expect(carapace.Batch(
			carapace.ActionValues(")").Prefix("ancestors(feature-x, "),
			carapace.ActionValuesDescribed(
				"&", "intersection",
				"|", "union",
				"~", "difference",
				"::", "DAG range",
				"..", "range",
				"-", "parents",
				"+", "children",
			).Prefix("ancestors(feature-x, "),
			carapace.ActionValues(depthVals...).Style(style.Blue).Tag("depth").Prefix("ancestors(feature-x, "),
		).ToA().NoSpace())
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

		depthVals := make([]string, 20)
		for i := 1; i <= 20; i++ {
			depthVals[i-1] = strconv.Itoa(i)
		}

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
			carapace.ActionValues(depthVals...).Tag("depth").Style(style.Blue).Prefix("parents(@-"),
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
