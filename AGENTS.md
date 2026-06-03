# AGENTS.md

## Project Overview

Go library for parsing [jj (Jujutsu)](https://github.com/jj-vcs/jj) revset, fileset, and template expressions into ASTs, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-jjlex`.

## Commands

```sh
# From repo root:
go test ./...                              # run all tests
go test ./pkg/revset/                      # run revset package tests only
go test ./pkg/fileset/                      # run fileset package tests only
go test -run TestParseRevset ./pkg/revset/  # run specific test
go build ./...                              # build all packages
go run main/main.go revset "<expr>"                    # parse revset expression, output AST as JSON
go run main/main.go revset-complete <cursor> "<expr>"  # revset completion context as JSON
go run main/main.go fileset "<expr>"                   # parse fileset expression, output AST as JSON
go run main/main.go fileset-complete <cursor> "<expr>" # fileset completion context as JSON
go run main/main.go fileset-bare "<expr>"              # parse fileset with bare string fallback
go run main/main.go template "<expr>"                    # parse template expression, output AST as JSON
go run main/main.go template-complete <cursor> "<expr>"  # template completion context as JSON
```

No Makefile, no linter config, no CI config present.

## Architecture

Three pairs of independent recursive-descent parsers:

### Revset (`pkg/revset/`)

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Strict: rejects partial/invalid input.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input, cursor)` → `*CompletionContext` describing what tokens are valid at the cursor position. Tolerant: recovers from errors at cursor to report expectations.

Both parsers implement the same operator precedence hierarchy (levels 0-6) but independently. The completion parser mirrors the main parser's structure but stops at the cursor and records expectations instead of building a full AST.

### Fileset (`pkg/fileset/`)

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Also `ParseProgramOrBareString()` for bare string/pattern fallback (matching jj's `program_or_bare_string` rule).
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input, cursor)` → `*CompletionContext`.

Fileset grammar is simpler than revset: no `::`, `..`, `-`, `+` operators; no `@` workspace syntax; no remote symbols. Operators are `|` (union), `&` (intersection), `~` (negate/difference). Precedence: `|` < `&`/`~` (infix) < `~` (prefix) < `p:x` (pattern) < primary.

### Template (`pkg/template/`)

- **`parser.go`** — Full parser using Pratt parser for expression precedence. `Parse()` → `*Expression` AST with spans. `++` (concat) is handled at the template level, outside the Pratt parser.
- **`completion_parser.go`** — Completion parser entry point. `ParseForCompletion(input, cursor)` → `*CompletionContext`.
- **`completion_parser_impl.go`** — Completion parser implementation with cursor-aware scanning.
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
| `main/main.go` | CLI entrypoint with subcommands for revset, fileset, and template |
| `main/main_test.go` | Integration tests with realistic examples from jj source |
| `pkg/actions/jj/function.go` | Completion action definitions: revset/fileset/string/date patterns, operators, functions, keyword args, special symbols |
| `pkg/actions/jj/revset.go` | Revset completion wiring: maps CompletionContext to carapace actions using function arg dispatch and pattern value dispatch |
| `pkg/actions/jj/revision.go` | Dynamic completion actions (bookmarks, tags, remotes, commits, operations, workspaces) that shell out to `jj` |
| `pkg/actions/jj/helpers.go` | Parsing helpers for `jj` CLI output (bookmarks, lines, toml aliases) |
| `pkg/actions/jj/exec.go` | `actionExecJJ` helper to run `jj` commands and process output |
| `pkg/actions/jj/uid.go` | UID generation helper for action deduplication |

## Key Patterns & Gotchas

### Completion actions return pure value lists

Action functions in `pkg/actions/jj/function.go` return raw value lists with Uid and Tag set, but without formatting modifiers (`.Suffix`, `.NoSpace`, `.Prefix`). These are applied at call sites in `pkg/actions/jj/revset.go` where context determines what suffix/no-space behavior is needed. This keeps actions reusable. Uid is set before any formatting modifiers so that suffix/prefix characters don't leak into the UID.

### Expression uses a type-erased payload pattern

`Expression.payload` is `any`; accessors (`.Identifier()`, `.UnaryOp()`, `.BinaryLHS()`, etc.) do type checks and return zero values on kind mismatch. Always check `Kind` before calling accessors.

### Two parsers must stay in sync (both packages)

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

## Testing

- Tests use standard `testing` package only (no testify or other deps)
- `revset_test.go` / `fileset_test.go` — parser tests using helpers
- `completion_test.go` (both packages) — completion tests using `assertHasExpected`, `assertHasOperator`
- `main/main_test.go` — integration tests with realistic examples from jj source code
- No external dependencies (pure stdlib)
- Passing `cursor=-1` to `ParseForCompletion` defaults to `len(input)` (end of input)

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
1. Skill (`skills/jj-revsets/SKILL.md`)
2. Parser (`pkg/revset/parser.go`)
3. Completion parser (`pkg/revset/completion_parser.go`)
4. Completion actions (`pkg/actions/jj/function.go`, `pkg/actions/jj/revset.go`)

Check: `lib/src/revset.pest`, `lib/src/revset.rs` (BUILTIN_FUNCTION_MAP), `docs/revsets.md`

### Fileset

When jj fileset syntax changes, update:
1. Skill (`skills/jj-filesets/SKILL.md`)
2. Parser (`pkg/fileset/parser.go`, `parser_helpers.go`, `scanner.go`, `helpers.go`)
3. Completion parser (`pkg/fileset/completion_parser.go`, `completion_helpers.go`)
4. AST (`pkg/fileset/ast.go`)
5. Completion actions (`pkg/actions/jj/function.go`, `pkg/actions/jj/revset.go`)

Check: `lib/src/fileset.pest`, `lib/src/fileset.rs` (BUILTIN_FUNCTION_MAP), `docs/filesets.md`

### Template

When jj template syntax changes, update:
1. Skill (`skills/jj-templates/SKILL.md`)
2. Parser (`pkg/template/parser.go`, `parser_helpers.go`)
3. Completion parser (`pkg/template/completion_parser.go`, `completion_parser_impl.go`, `completion_helpers.go`)
4. AST (`pkg/template/ast.go`)
5. Format (`pkg/template/format.go`)

Check: `lib/src/template.pest`, `lib/src/template_parser.rs`, `docs/templates.md`

### CLI, Concepts, Bookmarks, Operations, Config, Templates, Git Compat

When jj CLI or concepts change, update the corresponding skill in `skills/`. Source of truth is the jj documentation and CLI help output.

Check: `docs/`, `jj --help`, <https://jj-vcs.github.io/jj/latest/>

### Conflicts

When jj conflict handling or marker syntax changes, update:
1. Skill (`skills/jj-conflicts/SKILL.md`)

Check: `docs/conflicts/`, <https://jj-vcs.github.io/jj/latest/conflicts/>

### Divergence

When jj divergence handling or change offset syntax changes, update:
1. Skill (`skills/jj-divergence/SKILL.md`)

Check: `docs/guides/divergence/`, <https://jj-vcs.github.io/jj/latest/guides/divergence/>

### Forge Workflows

When jj forge integration (GitHub, GitLab, Gerrit) changes, update:
1. Skill (`skills/jj-forge-workflows/SKILL.md`)

Check: `docs/github/`, `docs/gerrit/`, <https://jj-vcs.github.io/jj/latest/github/>, <https://jj-vcs.github.io/jj/latest/gerrit/>
