# AGENTS.md

## Project Overview

Go library for parsing [jj (Jujutsu)](https://github.com/jj-vcs/jj) revset, fileset, and template expressions into ASTs, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-jjlex`.

## Commands

```sh
go test ./...                              # run all tests
go test ./pkg/revset/                      # run revset package tests only
go test ./pkg/fileset/                      # run fileset package tests only
go test ./pkg/template/                     # run template package tests only
go test -run TestParseRevset ./pkg/revset/  # run specific test
go build ./...                              # build all packages
go run . revset "<expr>"                   # parse revset expression, output AST as JSON
go run . revset-complete "<expr>"          # revset completion context as JSON
go run . fileset "<expr>"                  # parse fileset expression, output AST as JSON
go run . fileset-complete "<expr>"        # fileset completion context as JSON
go run . fileset-bare "<expr>"             # parse fileset with bare string fallback
go run . fileset-bare-complete "<expr>"    # fileset-bare completion context as JSON
go run . template "<expr>"                 # parse template expression, output AST as JSON
go run . template-complete "<expr>"        # template completion context as JSON
```

No Makefile, no linter config, no CI config present.

## Architecture

Cobra-based CLI (`cmd/`) wrapping three pairs of independent recursive-descent parsers (`pkg/revset/`, `pkg/fileset/`, `pkg/template/`) with completion actions that wire the parsers to carapace (`pkg/actions/tools/jj/`).

### CLI (`cmd/`)

- **`root.go`** — Root cobra command with `carapace.Gen(rootCmd).Standalone()`
- **`revset.go`** — `revset` and `revset-complete` subcommands, wired to `jj.ActionRevsets()` for shell completion
- **`fileset.go`** — `fileset`, `fileset-complete`, `fileset-bare`, and `fileset-bare-complete` subcommands
- **`template.go`** — `template` and `template-complete` subcommands

Entry point is `main.go` at repo root, which calls `cmd.Execute()`.

Three pairs of independent recursive-descent parsers:

### Revset (`pkg/revset/`)

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Strict: rejects partial/invalid input.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input)` → `*CompletionContext` describing what tokens are valid at end of input. Tolerant: recovers from errors to report expectations.

Both parsers implement the same operator precedence hierarchy (levels 0-6) but independently. The completion parser mirrors the main parser's structure but stops at end of input and records expectations instead of building a full AST.

### Fileset (`pkg/fileset/`)

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Also `ParseProgramOrBareString()` for bare string/pattern fallback (matching jj's `program_or_bare_string` rule).
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input)` → `*CompletionContext`.

Fileset grammar is simpler than revset: no `::`, `..`, `-`, `+` operators; no `@` workspace syntax; no remote symbols. Operators are `|` (union), `&` (intersection), `~` (negate/difference). Precedence: `|` < `&`/`~` (infix) < `~` (prefix) < `p:x` (pattern) < primary.

### Template (`pkg/template/`)

- **`parser.go`** — Full parser using Pratt parser for expression precedence. `Parse()` → `*Expression` AST with spans. `++` (concat) is handled at the template level, outside the Pratt parser.
- **`completion_parser.go`** — Completion parser entry point. `ParseForCompletion(input)` → `*CompletionContext`.
- **`completion_parser_impl.go`** — Completion parser implementation.
- **`completion_helpers.go`** — Dedup helpers for tokens and operators.
- **`parser_helpers.go`** — Shared helper functions: character classification (`isWhitespace`, `isIdentifierStart`, `isIdentifierPart`, `isFunctionName`, `isPatternIdentifierStart`, `isPatternIdentifierPart`, `isStrictIdentifierPart`), string/integer literal scanning, infix operator peeking.
- **`ast.go`** — Template AST node types (`UnaryOp`, `BinaryOp`, `ExpressionKind`, payload types including `ConcatExpr`, `FunctionCallExpr`, `MethodCallExpr`, `LambdaExpr`, `KeywordArg`).
- **`span.go`** — `Span` (Start/End byte offsets) and `Pos` types.
- **`format.go`** — Template AST → string formatting with precedence-aware parenthesization.
- **`completion.go`** — Template completion context types (`ExpectedToken`, `ValidOperator`, `FunctionContext`, `CompletionContext`).
- **`template_test.go`** — Template parser tests.
- **`completion_test.go`** — Template completion tests.

Template grammar uses a Pratt parser for infix operators (precedence from weakest to strongest: `||`, `&&`, `==`/`!=`, comparisons, `+`/`-`, `*`/`/`/`%`), with `++` (concat) handled at the top template level. Prefix operators (`!`, `-`), method calls (`x.f()`), function calls (`f(x)`), patterns (`name:value`), lambdas (`|| expr`, `|x| expr`), and keyword arguments (`f(x, key=val)`) are also supported.

### File responsibilities

| File | Purpose |
|---|---|
| `pkg/revset/ast.go` | Revset AST node types, payload structs, accessor methods |
| `pkg/revset/span.go` | `Span` (Start/End byte offsets) and `Pos` types |
| `pkg/revset/parser.go` | Revset main parser + public API: `Parse()`, `IsIdentifier()`, `ParseSymbol()`, `Format()`. Shared helper functions. |
| `pkg/revset/format.go` | Revset AST → string formatting with precedence-aware parenthesization |
| `pkg/revset/completion.go` | Revset completion context types: `CompletionContext`, `ExpectedToken`, `ValidOperator`, `FunctionContext` |
| `pkg/revset/completion_parser.go` | Revset completion parser |
| `pkg/revset/revset_test.go` | Revset parser tests |
| `pkg/revset/completion_test.go` | Revset completion tests |
| `pkg/fileset/ast.go` | Fileset AST node types (includes `KindBareString`, `KindBareStringPattern` not in revset) |
| `pkg/fileset/span.go` | `Span` and `Pos` types (same structure as revset) |
| `pkg/fileset/parser.go` | Fileset main parser + `Parse()`, `ParseProgramOrBareString()`, `IsIdentifier()` |
| `pkg/fileset/parser_helpers.go` | Fileset parser helper methods (`parseSymbolOrFunctionOrPattern`, `parseFunctionCall`, `tryBareStringPattern`, `tryBareString`, `isFunctionName`, `isStrictIdentifier`) |
| `pkg/fileset/scanner.go` | Fileset scanner methods (`scanIdentifier`, `scanStrictIdentifier`, `scanBareString`, string literal parsing, `unionNodes`) |
| `pkg/fileset/helpers.go` | Shared helper functions: `isWhitespace`, `isFilesetIdentifierPart`, `isFilesetIdentifierStart`, `isStrictIdentifierPart`, `isBareStringPart`, `isHexDigit`, `hexVal`, `splitIdentifierParts`, `containsRune` |
| `pkg/fileset/format.go` | Fileset AST → string formatting with precedence-aware parenthesization |
| `pkg/fileset/completion.go` | Fileset completion context types (same structure as revset, no keyword args) |
| `pkg/fileset/completion_parser.go` | Fileset completion parser |
| `pkg/fileset/completion_helpers.go` | Fileset completion parser helper methods and dedup functions |
| `pkg/fileset/fileset_test.go` | Fileset parser tests |
| `pkg/fileset/completion_test.go` | Fileset completion tests |
| `pkg/template/ast.go` | Template AST node types (`UnaryOp`, `BinaryOp`, `ExpressionKind`, payload structs) |
| `pkg/template/span.go` | `Span` and `Pos` types |
| `pkg/template/parser.go` | Template main parser + `Parse()`, `IsIdentifier()`, `Format()`, `parseFunctionArgs()`, `tryParseKeywordArg()` |
| `pkg/template/parser_helpers.go` | Template parser helpers: character classification, string/integer scanning, infix op peek/mapping, precedence constants |
| `pkg/template/format.go` | Template AST → string formatting with precedence-aware parenthesization |
| `pkg/template/completion.go` | Template completion context types (`ExpectedToken`, `ValidOperator`, `FunctionContext`, `CompletionContext`) |
| `pkg/template/completion_parser.go` | Template completion parser entry point |
| `pkg/template/completion_parser_impl.go` | Template completion parser implementation |
| `pkg/template/completion_helpers.go` | Template completion parser dedup helpers |
| `pkg/template/template_test.go` | Template parser tests |
| `pkg/template/completion_test.go` | Template completion tests |
| `main.go` | CLI entrypoint, calls `cmd.Execute()` |
| `main_test.go` | Integration tests with realistic examples from jj source |
| `cmd/root.go` | Root cobra command |
| `cmd/revset.go` | Revset subcommands |
| `cmd/fileset.go` | Fileset subcommands (including bare string variants) |
| `cmd/template.go` | Template subcommands |
| `pkg/actions/tools/jj/function.go` | Static action definitions: patterns, operators, functions, keyword args, special symbols |
| `pkg/actions/tools/jj/revset.go` | Completion wiring: maps CompletionContext to carapace actions. Includes `actionQuotedRevsetArg`, `isStringPatternFunction`, `actionForFunctionArg` |
| `pkg/actions/tools/jj/revision.go` | Dynamic completion actions (bookmarks, tags, remotes, commits, operations) that shell out to `jj`. Includes `actionLocalBookmarksRaw`, `actionTagsRaw`/`actionTags`, `stripDisplayQuotes` |
| `pkg/actions/tools/jj/helpers.go` | Parsing helpers for `jj` CLI output (bookmarks, lines) |
| `pkg/actions/tools/jj/exec.go` | `actionExecJJ`/`actionExecJJE` helpers to run `jj` commands |
| `pkg/actions/tools/jj/uid.go` | UID generation helper for action deduplication |
| `pkg/actions/tools/jj/action_test.go` | Sandbox tests for actions and unit tests for parsing helpers |
| `pkg/fixture/fixture.go` | Fixture for creating jj test repos: commits, bookmarks, tags, state inspection |
| `pkg/fixture/fixture_t.go` | Testing.T wrapper with fatal helpers for fixture operations |

### Man Pages (`man/`)

YAML documentation files for completion value UIDs. Organized as `man/jj/<host>/<host>.yaml` where the UID `jj://host/path` maps to key `path` in the YAML. Two kinds of content:

- **Known concepts** — fixed meaning, enumerable keys (e.g. `revset-function/parents`, `fileset-operator/&`)
- **Live entities** — context-dependent, runtime keys (e.g. `bookmark/main`, `commit/abc123`)

## Key Patterns & Gotchas

### Completion actions return pure value lists

Action functions in `pkg/actions/tools/jj/function.go` return raw value lists with Uid and Tag set, but without formatting modifiers (`.Suffix`, `.NoSpace`, `.Prefix`). These are applied at call sites in `pkg/actions/tools/jj/revset.go` where context determines what suffix/no-space behavior is needed. This keeps actions reusable. Uid is set before any formatting modifiers so that suffix/prefix characters don't leak into the UID.

### Expression uses a type-erased payload pattern

`Expression.payload` is `any`; accessors (`.Identifier()`, `.UnaryOp()`, `.BinaryLHS()`, etc.) do type checks and return zero values on kind mismatch. Always check `Kind` before calling accessors.

### Two parsers must stay in sync (all three packages)

When modifying operator precedence or parsing rules in `parser.go`, the same changes must be mirrored in `completion_parser.go`. They share helper functions but have independent parser types. The completion parsers also duplicate `isFunctionName`/`isStrictIdentifier` as `isFunctionNameCheck`/`isStrictIdentifierCheck` — keep these in sync.

### Revset identifier rules

- Identifiers allow internal `.`, `-`, `+` as connectors (e.g. `foo.bar-v1+7`)
- Multiple consecutive `-` are allowed (`foo--bar`), but `+` and `.` are not repeatable
- `*` and `/` are valid identifier characters (for glob patterns)
- `isIdentifierPart` uses Unicode categories (XID_CONTINUE); `isStrictIdentifierPart` is ASCII-only

### Fileset identifier rules

- Fileset identifiers include `+`, `-`, `.`, `@`, `_`, `*`, `?`, `[`, `]`, `/`, `\` plus XID_CONTINUE
- This is broader than revset identifiers (e.g., `+` is a regular char, not a postfix operator)
- `isFilesetIdentifierPart` checks for these characters
- Bare strings (`bare_string`) allow spaces and even more characters — used in `ParseProgramOrBareString` fallback
- Strict identifiers (for pattern names and function names) are ASCII alphanumeric + `_` + `-` separators

### Fileset bare string and bare string pattern

- `ParseProgramOrBareString()` tries expression parsing first, then falls back to `bare_string_pattern` (strict_identifier:bare_string) and `bare_string`
- Bare strings allow spaces, `+`, `-`, `.`, `@`, `_`, `*`, `?`, `[`, `]`, `/`, `\`, and non-ASCII
- This matches jj's `program_or_bare_string` grammar rule for user-friendly input

### Revset operator ambiguity

- `~` is both prefix (negate) and infix (difference) — context-dependent
- `-` is postfix (parents) but NOT infix — `foo - bar` is an error (suggests `~`)
- `+` is postfix (children) but NOT infix — `foo + bar` is an error (suggests `|`)
- `:` alone is always an error (suggests `::`)
- Range operators (`::`, `..`) cannot be nested without parentheses
- Prefix `::` and `..` do not allow whitespace between operator and operand

### Fileset operator simplicity

- `~` is both prefix (negate) and infix (difference) — same as revset
- No `::`, `..`, `-`, `+` postfix operators
- No `@` workspace/remote syntax
- No keyword arguments in function calls
- Only built-in functions: `all()` and `none()`

### Pattern syntax (`name:value`)

No whitespace allowed around the `:` in patterns. `exact: foo` is an error. The pattern name must be a strict identifier; the value is parsed as a primary expression. Pattern is right-associative: `x:y:z` = `x:(y:z)`.

### Span tracking

Spans exclude trailing whitespace. The parser tracks `lastContent` and clips expression spans accordingly.

### Union flattening

The `|` operator flattens into a single `KindUnionAll` node rather than creating nested binary trees.

### Completion parser `consumed` flag

Tracks whether any input was consumed before reaching the cursor. Distinguishes "expecting first expression" from "after an expression, expecting operator".

### Template `++` is weakest binding operator

The `++` (concatenation) operator is handled at the template level, outside the Pratt parser. This means all other binary operators (`||`, `&&`, `==`, `!=`, comparisons, `+`, `-`, `*`, `/`, `%`) bind tighter than `++`. So `x && y ++ z` parses as `(x && y) ++ z`. The format function uses `precPrimary` context for concat operands to force parentheses around any non-primary expression inside a concat.

### Template pattern identifiers allow dashes

Pattern names like `regex-i` include dashes before the colon. The parser uses `scanPatternIdentifierSuffix()` to extend identifiers after the initial scan. The completion parser mirrors this with `scanPatternIdentifierSuffixComp()`.

### Template keyword arguments

Template function calls support keyword arguments (`f(x, key=val)`). The parser's `tryParseKeywordArg()` checks for `identifier = expression` sequences. The completion parser handles keyword arg name completion.

### Completion parser `innermostFunc` guard

`setFunctionContext()` always updates to the most recently entered function, so nested function calls correctly report the innermost function context.

### Quoted string completions and display quoting

jj display-quotes bookmark/tag names with special characters (e.g. a git branch `parents(` is displayed as `"parents("` by `jj bookmark list`). This creates two challenges for completion:

1. **Raw vs display-quoted names**: `ActionLocalBookmarks` returns display-quoted values from `jj bookmark list`. When completing inside a quoted string (e.g. `"paren`), these values won't match because they already include the opening quote. `actionLocalBookmarksRaw` and `actionTagsRaw` return raw (unquoted) names using `name.substr()` (bookmarks) or `stripDisplayQuotes` (tags) so that `Prefix(quote).Suffix(quote)` can add consistent quoting at the completion site.

2. **String pattern vs symbol reference**: A quoted string in a revset is a symbol reference (bookmark/tag/commit), not a string pattern. `isStringPatternFunction` distinguishes functions where a quoted argument is a pattern value (`author`, `description`, `diff_lines`) from functions where it's a symbol reference (`parents`, `children`). `actionQuotedRevsetArg` offers only symbols valid inside quotes (no remote bookmarks with `@remote`, no functions, no patterns).

The `StringQuote` field (not `PartialString`) is used to detect "inside a quoted string" — it's set even when no content has been typed yet (just the opening quote).

### Operators suppressed inside unclosed string literals

The completion parser's `afterExpression()` and `parseRangeExpr()` check `p.ctx.StringQuote != 0` and skip adding `ExpectedOperator`/`ValidOperators` when inside an unclosed string. This prevents operators from being offered while the user is still typing a quoted symbol (e.g. `"paren` should offer bookmarks, not `|`, `&`, etc.). Operators are correctly offered after the string is closed (e.g. `"main"`).

### Operators after complete expression in function args

When a complete expression (e.g. `"parents("`) is followed by the cursor inside a function, `ExpectedOperator` is set but `ExpectedExpression` is not. The action layer offers operators, `)`, and `,`. When an infix operator (e.g. `|`) has been consumed within an argument and at least one arg was parsed (`len(ctx.Function.Args) > 0`), both `ExpectedExpression` and `ExpectedOperator` are set — general revset expressions are offered for the RHS instead of the function-arg-specific action (which may be empty, e.g. `parents` arg 1 is an integer depth).

## Testing

- Tests use standard `testing` package only (no testify or other deps)
- `revset_test.go` / `fileset_test.go` — parser tests using helpers
- `completion_test.go` (all three packages) — completion tests using `assertHasExpected`, `assertHasOperator`
- `main_test.go` — integration tests with realistic examples from jj source code
- `pkg/actions/tools/jj/action_test.go` — sandbox tests for carapace actions and unit tests for parsing helpers
- `pkg/fixture/` — jj repo fixture for creating test repos with commits, bookmarks, tags, etc.
- Parser/completion packages have no external dependencies (pure stdlib); `pkg/actions/tools/jj` depends on carapace and cobra

## Skills

The `skills/` directory contains reference skills for jj concepts, CLI, and expression languages. Each skill is a self-contained SKILL.md with YAML frontmatter (`name`, `description`, `user-invocable`).

The `jj` skill is the user-invocable entrypoint that refers to focused sub-skills. The sub-skills have `user-invocable: false` — they trigger automatically via their `description` patterns but are not directly invoked by name.

| Skill | Description | User-invocable |
|-------|-------------|----------------|
| `jj` | Main entrypoint for jj skills — refers to sub-skills by topic | true |
| `jj-cli` | CLI commands, subcommands, flags, and argument types | false |
| `jj-concepts` | Core concepts: working copy model, change IDs, conflicts, immutable revisions, descendant rebasing | false |
| `jj-conflicts` | First-class conflicts, marker styles (diff/snapshot/git), long markers, resolution workflows | false |
| `jj-bookmarks` | Bookmarks (jj's branches): tracking, conflicts, push safety, CLI commands | false |
| `jj-operations` | Operation log: undo/redo, --at-op, lock-free concurrency | false |
| `jj-config` | Configuration: levels, sections, conditional config, CLI commands | false |
| `jj-revsets` | Revset expression syntax: symbols, operators, functions, string/date patterns, aliases | false |
| `jj-filesets` | Fileset expression syntax: operators, pattern kinds, built-in functions, bare strings | false |
| `jj-templates` | Template language: operators, types, methods, global functions, aliases | false |
| `jj-divergence` | Divergent changes: what divergence is, how it happens, resolution strategies, change offsets | false |
| `jj-forge-workflows` | GitHub/GitLab/Gerrit integration, push options, pull requests, code review, multi-remote setup | false |
| `jj-git-compat` | Git comparison: conceptual differences, command equivalents, colocation, migration | false |

Skills cross-reference each other. For example, `jj-cli` references `jj-revsets` for revset syntax and `jj-filesets` for fileset syntax.

## Syncing with jj Upstream

### Revset

When jj revset syntax changes, update:
1. Skill (`skills/jj/references/revsets.md`)
2. Parser (`pkg/revset/parser.go`)
3. Completion parser (`pkg/revset/completion_parser.go`)
4. Completion actions (`pkg/actions/tools/jj/function.go`, `pkg/actions/tools/jj/revset.go`)
5. Man pages (`man/jj/revset-function/`, `man/jj/revset-operator/`, `man/jj/revset-pattern/`, etc.)

Check: `lib/src/revset.pest`, `lib/src/revset.rs` (BUILTIN_FUNCTION_MAP), `docs/revsets.md`

### Fileset

When jj fileset syntax changes, update:
1. Skill (`skills/jj/references/filesets.md`)
2. Parser (`pkg/fileset/parser.go`, `parser_helpers.go`, `scanner.go`, `helpers.go`)
3. Completion parser (`pkg/fileset/completion_parser.go`, `completion_helpers.go`)
4. AST (`pkg/fileset/ast.go`)
5. Completion actions (`pkg/actions/tools/jj/function.go`, `pkg/actions/tools/jj/revset.go`)
6. Man pages (`man/jj/fileset-function/`, `man/jj/fileset-operator/`, `man/jj/fileset-pattern/`)

Check: `lib/src/fileset.pest`, `lib/src/fileset.rs` (BUILTIN_FUNCTION_MAP), `docs/filesets.md`

### Template

When jj template syntax changes, update:
1. Skill (`skills/jj/references/templates.md`)
2. Parser (`pkg/template/parser.go`, `parser_helpers.go`)
3. Completion parser (`pkg/template/completion_parser.go`, `completion_parser_impl.go`, `completion_helpers.go`)
4. AST (`pkg/template/ast.go`)
5. Format (`pkg/template/format.go`)
6. Man pages (`man/jj/template-function/`, `man/jj/template-operator/`)

Check: `lib/src/template.pest`, `lib/src/template_parser.rs`, `docs/templates.md`

### CLI, Concepts, Bookmarks, Operations, Config, Templates, Git Compat

When jj CLI or concepts change, update the corresponding reference in `skills/jj/references/`. Source of truth is the jj documentation and CLI help output.

Check: `docs/`, `jj --help`, <https://jj-vcs.github.io/jj/latest/>

### Conflicts

When jj conflict handling or marker syntax changes, update:
1. Skill (`skills/jj/references/conflicts.md`)

Check: `docs/conflicts/`, <https://jj-vcs.github.io/jj/latest/conflicts/>

### Divergence

When jj divergence handling or change offset syntax changes, update:
1. Skill (`skills/jj/references/divergence.md`)

Check: `docs/guides/divergence/`, <https://jj-vcs.github.io/jj/latest/guides/divergence/>

### Forge Workflows

When jj forge integration (GitHub, GitLab, Gerrit) changes, update:
1. Skill (`skills/jj/references/forge-workflows.md`)

Check: `docs/github/`, `docs/gerrit/`, <https://jj-vcs.github.io/jj/latest/github/>, <https://jj-vcs.github.io/jj/latest/gerrit/>
