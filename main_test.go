package main

import (
	"testing"

	"github.com/carapace-sh/carapace-jjlex/pkg/fileset"
	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
	"github.com/carapace-sh/carapace-jjlex/pkg/template"
)

// Realistic revset examples sourced from jj's source code
// (docs/revsets.md, lib/tests/test_revset.rs, cli/src/config/revsets.toml)

var parseSuccessCases = []string{
	// Primitive symbols
	"root()",
	"none()",
	"all()",
	"visible_heads()",
	"@",

	// Postfix operators
	"@-",
	"root()-",
	"none()-",
	"@+",
	"root()+",
	"none()+",

	// Prefix dag range / range
	"::@",
	"::root()",
	"::none()",
	"@::",
	"root()::",
	"..@",
	"..root()",
	"@..",
	"root()..",

	// Binary dag range / range
	"@::@",
	"root()::root()",
	"@..@",
	"root()..root()",

	// Nullary dag range / range
	"::",
	"..",

	// Union / intersection / difference
	"tags() | bookmarks()",
	"root() & none()",
	"all() ~ root()",

	// Negate
	"~root()",
	"~all()",
	"~~foo",

	// Function calls - nullary
	"visible_heads()",
	"git_refs()",
	"git_head()",
	"merges()",
	"conflicts()",
	"divergent()",
	"empty()",
	"mine()",
	"signed()",
	"bookmarks()",
	"tags()",
	"remote_bookmarks()",
	"tracked_remote_bookmarks()",
	"untracked_remote_bookmarks()",
	"remote_tags()",
	"tracked_remote_tags()",
	"untracked_remote_tags()",

	// Function calls - unary
	"heads(all())",
	"heads(none())",
	"roots(all())",
	"roots(none())",
	"latest(all())",
	"latest(all(), 1)",
	"latest(all(), 0)",
	"fork_point(none())",
	"fork_point(all())",
	"connected(none())",
	"connected(root())",
	"present(@)",
	"exactly(none(), 0)",

	// Function calls - with string pattern args
	`description(substring:commit)`,
	`description("")`,
	`description("foo\n")`,
	`description(exact:"commit 1\n")`,
	`subject("commit 2")`,
	`author(substring:name)`,
	`author(*name2*)`,
	`author(*email3*)`,
	`author(substring-i:Name)`,
	`committer(substring:name)`,
	`committer(*name2*)`,

	// Function calls - bookmark/tag with patterns
	`bookmarks(bookmark1)`,
	`bookmarks(substring:bookmark)`,
	`bookmarks(exact:bookmark1)`,
	`bookmarks(glob:"Bookmark?")`,
	`bookmarks(glob-i:"Bookmark?")`,
	`bookmarks(regex:'ookmark')`,
	`bookmarks(regex-i:'BOOKmark')`,
	`tags(tag1)`,
	`tags(substring:tag)`,
	`tags(exact:tag1)`,
	`tags(glob:"Tag?")`,
	`tags(glob-i:"Tag?")`,
	`tags(regex:'ag')`,
	`tags(regex:'^[Tt]ag1$')`,

	// Function calls - remote_bookmarks with keyword args
	`remote_bookmarks()`,
	`remote_bookmarks(bookmark1)`,
	`remote_bookmarks(*, origin)`,
	`remote_bookmarks(remote=git)`,
	`remote_bookmarks(remote=*)`,
	`remote_bookmarks(bookmark1, origin)`,
	`remote_bookmarks(remote=foo)`,

	// Function calls - date
	`author_date(after:'2023-03-25 12:00')`,
	`author_date(before:'2023-03-25 12:00')`,
	`committer_date(after:'2023-03-25 12:00')`,
	`committer_date(before:'2023-03-25 12:00')`,

	// Function calls - files
	`files("repo/added_clean_clean")`,
	`files("added_clean_clean"|"added_modified_clean")`,

	// Function calls - diff_lines
	`diff_lines(*2*)`,
	`diff_lines_added(regex:'[1234]', 'file1')`,
	`diff_lines_removed(regex:'[1234]', 'file2')`,

	// Function calls - commit_id / change_id
	`commit_id(019f179b4479a4f3d1373b772866037929e4f63c)`,
	`commit_id('')`,
	`change_id(zvlyxpuvtsoopsqzlkorrpqrszrqvlnx)`,
	`change_id('')`,

	// Function calls - at_operation
	`at_operation(@, all())`,
	`at_operation(@-, all())`,
	`at_operation(@--, all())`,
	`at_operation(000000000000-, all())`,

	// Function calls - coalesce
	`coalesce()`,
	`coalesce(none())`,
	`coalesce(all(), @)`,
	`coalesce(none(), none(), @)`,

	// Function calls - reachable
	`reachable(@, root()..)`,

	// Function calls - children with count
	`children(root(), 2)`,

	// Function calls - first_parent
	`first_parent(root())`,

	// Function calls - bisect
	`bisect(none())`,
	`bisect(root())`,

	// Complex compound expressions from docs/config
	"tags() | bookmarks()",
	"remote_bookmarks()..",
	"remote_bookmarks(remote=origin)..",
	"(remote_bookmarks()..@)::",
	`author(*martinvonz*) & description(*reset*)`,
	"reachable(@, mutable())",
	"present(@) | ancestors(immutable_heads().., 2) | trunk()",
	"present(@)",
	"mutable() | immutable_heads()",
	"heads(::to & bookmarks())",
	"trunk() | tags() | untracked_remote_bookmarks()",
	"::(immutable_heads() | root())",
	"~immutable()",
	"::visible_heads()",
	"~visible()",

	// Remote symbols
	"main@origin",
	`"foo bar"@origin`,
	`main@"foo bar"`,

	// At workspace
	"main@",
	`"foo bar"@`,

	// Parenthesized expressions with operators
	"(D|A)-",
	"(C|B)+",
	"(C|B)::",
	"(C|B)..",
	"(B|root())+",
	"(C|B)::(C|B)",
	"(C|B)..(C|B)",

	// Chained operators
	"root()++",
	"foo---",
	"foo+++",
	"((foo-)-)-",

	// Precedence combinations
	"~x|y",
	"x&~y",
	"x|y&z",
	"x|y~z",
	"x~~y",

	// Pattern expressions
	`substring:"foo"`,
	`exact:foo`,
	"exact:@",
	`glob:"ci/*"`,
	`regex:"pattern"`,
	`glob-i:"fix*jpeg*"`,
	"exact:( 'foo' )",
	"x:f(y)",
	"x:@-+",
	"x:y::z",
	"x:y&z",
	"x:y:z",

	// Whitespace handling
	" \t\r\n\x0call()",
	"  description(  arg1 ) ~    file(  arg1 ,   arg2 )  ~ visible_heads(  )  ",
	"remote_bookmarks( remote  =   foo  )",
}

var parseErrorCases = []string{
	// Curly braces are not valid revset syntax
	"{}",
	"{path}",

	// Invalid operator usage
	":foo",
	"foo^",
	"foo + bar",
	"foo - bar",

	// Space in prefix range operators
	" :: foo ",
	" .. foo ",

	// Incomplete expression
	"foo | -",
	"parents(foo",

	// Invalid identifier
	".foo",
	"foo.",
	"foo.+bar",
	"foo++bar",
	"foo+-bar",

	// Invalid string escapes
	`"\y"`,
	`"\x"`,
	`"\xf"`,
	`"\xgg"`,

	// Trailing comma issues
	"bookmarks(,)",
	"bookmarks(,a)",
	"bookmarks(a,,)",
	"file(a,,b)",

	// Repeated range operators
	":::foo",
	"::::foo",
	"foo:::",
	"foo::::",
	"foo:::bar",
	"::::",
	"....foo",
	"foo....",
	"....",
}

func TestParseSuccess(t *testing.T) {
	for _, input := range parseSuccessCases {
		t.Run(input, func(t *testing.T) {
			_, err := revset.Parse(input)
			if err != nil {
				t.Fatalf("expected success for %q, got error: %v", input, err)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	for _, input := range parseErrorCases {
		t.Run(input, func(t *testing.T) {
			_, err := revset.Parse(input)
			if err == nil {
				t.Fatalf("expected error for %q, got success", input)
			}
		})
	}
}

// Fileset integration tests sourced from jj's documentation and source code

var filesetSuccessCases = []string{
	// Primitives
	"all()",
	"none()",

	// File patterns
	`glob:"*.rs"`,
	`glob:"**/*.rs"`,
	`exact:foo`,
	`cwd:"src/lib.rs"`,
	`root:"src/lib.rs"`,
	`root-file:"src/lib.rs"`,
	`root-glob:"*.rs"`,
	`cwd-file:"src/lib.rs"`,
	`cwd-glob:"*.rs"`,
	`prefix-glob:"*.d"`,
	`cwd-prefix-glob:"*.d"`,
	`root-prefix-glob:"src"`,

	// Case-insensitive patterns
	`glob-i:"*.TXT"`,
	`cwd-glob-i:"*.txt"`,
	`prefix-glob-i:"*.d"`,
	`root-glob-i:"*.rs"`,
	`root-prefix-glob-i:"src"`,

	// Operators
	"foo | bar",
	"foo & bar",
	"foo ~ bar",
	"~foo",
	"~~foo",

	// Precedence
	"foo | bar & baz",
	"foo & bar ~ baz",

	// Parenthesized
	"(foo | bar) & baz",
	"~(foo | bar)",

	// String literals
	`"src/lib.rs"`,
	`'src/lib.rs'`,

	// Identifiers with special chars (path separators, glob chars)
	"src/lib.rs",
	"*.rs",
	"**/*.txt",
	"foo-bar",
	"foo+bar",

	// Function calls in expressions
	"all() | none()",
	"~all()",
	"all() & none()",
}

var filesetErrorCases = []string{
	// Incomplete expression
	"foo | ~",
	"all(",
	")",
	"foo &",
	"foo |",
}

func TestFilesetParseSuccess(t *testing.T) {
	for _, input := range filesetSuccessCases {
		t.Run(input, func(t *testing.T) {
			_, err := fileset.Parse(input)
			if err != nil {
				t.Fatalf("expected success for %q, got error: %v", input, err)
			}
		})
	}
}

func TestFilesetParseError(t *testing.T) {
	for _, input := range filesetErrorCases {
		t.Run(input, func(t *testing.T) {
			_, err := fileset.Parse(input)
			if err == nil {
				t.Fatalf("expected error for %q, got success", input)
			}
		})
	}
}

// Template integration tests sourced from jj's documentation and template language

var templateSuccessCases = []string{
	// Literals
	"true",
	"false",
	"42",
	`"hello"`,
	`'raw string'`,
	`"escapes: \t\n\r\0\e\x41"`,
	`"backslash: \\"`,

	// Identifiers
	"commit_id",
	"change_id",
	"author",
	"description",

	// Operators
	"x && y",
	"x || y",
	"!x",
	"-42",
	"x == y",
	"x != y",
	"x > y",
	"x < y",
	"x >= y",
	"x <= y",
	"x + y",
	"x - y",
	"x * y",
	"x / y",
	"x % y",

	// Concatenation
	"x ++ y",
	"x ++ y ++ z",

	// Precedence
	"!x.f()",
	"x == y && z",
	"x || y && z",
	"(x || y) && z",
	"x + y * z",

	// Method calls
	"x.f()",
	"x.f().g()",
	"x.f(a)",
	"x.f(a, b)",

	// Function calls
	"if(true, x, y)",
	"label(separate(' ', x, y))",
	"coalesce(x, y, z)",

	// Keyword arguments
	"label(color='red', x)",

	// Patterns
	`exact:foo`,
	`glob:"*.rs"`,
	`regex:"pattern"`,
	`substring:"bar"`,
	`glob-i:"*.TXT"`,

	// Nested patterns
	`x:y:z`,

	// Lambda expressions
	"|| x",
	"|a| x",
	"|a, b| x + y",

	// Parenthesized expressions
	"(x || y)",
	"(x + y) * z",

	// Boolean literals in context
	"if(true, x, y)",
	"if(false, x, y)",

	// Pattern value types: boolean, integer
	"regex:false",
	"regex:0",

	// Pattern with dash name and infix operators
	"x+y-z:a",
	"x-y+z:a",

	// Lambda vs logical operator
	"x||||y",
	"||||x",

	// Lambda binding
	"|| x ++ y",
	"f(|| x, || y)",

	// Complex expressions
	"change_id.short()",
	"commit_id.short()",
	"description ++ '\\n'",
	"if(divergent, label('divergent', 'D'), '')",
	"parents.map(|c| c.commit_id().short()).join(\",\")",
}

var templateErrorCases = []string{
	// Incomplete expressions
	"foo ++",
	"foo &&",
	"foo ||",
	"(",
	")",
	"if(",

	// Invalid syntax
	"true(",
	"false(",
	"true:foo",

	// Boolean as keyword arg name
	"f(true=0)",
	"f(false=0)",

	// Invalid string escapes in double quotes
	`"\y"`,

	// Integer with leading zero
	"042",
}

func TestTemplateParseSuccess(t *testing.T) {
	for _, input := range templateSuccessCases {
		t.Run(input, func(t *testing.T) {
			_, err := template.Parse(input)
			if err != nil {
				t.Fatalf("expected success for %q, got error: %v", input, err)
			}
		})
	}
}

func TestTemplateParseError(t *testing.T) {
	for _, input := range templateErrorCases {
		t.Run(input, func(t *testing.T) {
			_, err := template.Parse(input)
			if err == nil {
				t.Fatalf("expected error for %q, got success", input)
			}
		})
	}
}

func TestTemplateRoundTrip(t *testing.T) {
	cases := []string{
		"true",
		"false",
		"42",
		`"hello"`,
		"commit_id",
		"x ++ y",
		"x && y",
		"x || y",
		"!x",
		"x == y",
		"x != y",
		"x > y",
		"x + y",
		"x - y",
		"x * y",
		"x.f()",
		"if(true, x, y)",
		`exact:foo`,
		"|| x",
		"|a| x",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			expr, err := template.Parse(input)
			if err != nil {
				t.Fatalf("parse error for %q: %v", input, err)
			}
			formatted := template.Format(expr)
			expr2, err := template.Parse(formatted)
			if err != nil {
				t.Fatalf("re-parse error for %q (formatted as %q): %v", input, formatted, err)
			}
			formatted2 := template.Format(expr2)
			if formatted != formatted2 {
				t.Errorf("round-trip mismatch for %q: got %q then %q", input, formatted, formatted2)
			}
		})
	}
}
