# AGENTS.md

## Project Overview

Go library for parsing [jj (Jujutsu)](https://github.com/jj-vcs/jj) revset and fileset expressions into ASTs, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-jjlex`.

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
```

No Makefile, no linter config, no CI config present.

## Architecture

Two pairs of independent recursive-descent parsers:

### Revset (`pkg/revset/`)

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Strict: rejects partial/invalid input.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input, cursor)` → `*CompletionContext` describing what tokens are valid at the cursor position. Tolerant: recovers from errors at cursor to report expectations.

Both parsers implement the same operator precedence hierarchy (levels 0-6) but independently. The completion parser mirrors the main parser's structure but stops at the cursor and records expectations instead of building a full AST.

### Fileset (`pkg/fileset/`)

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Also `ParseProgramOrBareString()` for bare string/pattern fallback (matching jj's `program_or_bare_string` rule).
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input, cursor)` → `*CompletionContext`.

Fileset grammar is simpler than revset: no `::`, `..`, `-`, `+` operators; no `@` workspace syntax; no remote symbols. Operators are `|` (union), `&` (intersection), `~` (negate/difference). Precedence: `|` < `&`/`~` (infix) < `~` (prefix) < `p:x` (pattern) < primary.

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
| `main/main.go` | CLI entrypoint with subcommands for revset and fileset |
| `main/main_test.go` | Integration tests with realistic examples from jj source |

## Key Patterns & Gotchas

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

### Completion parser `innermostFunc` guard

`setFunctionContext()` only sets `p.ctx.Function` if `p.innermostFunc` is nil. The innermost (deepest) function call wins.

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
| `jj-revsets` | Revset expression syntax: symbols, operators, functions, string/date patterns, aliases | false |
| `jj-filesets` | Fileset expression syntax: operators, pattern kinds, built-in functions, bare strings | false |
| `jj-bookmarks` | Bookmarks (jj's branches): tracking, conflicts, push safety, CLI commands | false |
| `jj-operations` | Operation log: undo/redo, --at-op, lock-free concurrency | false |
| `jj-config` | Configuration: levels, sections, conditional config, CLI commands | false |
| `jj-templates` | Template language: operators, types, methods, global functions, aliases | false |
| `jj-git-compat` | Git comparison: conceptual differences, command equivalents, colocation, migration | false |

Skills cross-reference each other. For example, `jj-cli` references `jj-revsets` for revset syntax and `jj-filesets` for fileset syntax.

## Syncing with jj Upstream

### Revset

When jj revset syntax changes, update:
1. Skill (`skills/jj-revsets/SKILL.md`)
2. Parser (`pkg/revset/parser.go`)
3. Completion parser (`pkg/revset/completion_parser.go`)

Check: `lib/src/revset.pest`, `lib/src/revset.rs` (BUILTIN_FUNCTION_MAP), `docs/revsets.md`

### Fileset

When jj fileset syntax changes, update:
1. Skill (`skills/jj-filesets/SKILL.md`)
2. Parser (`pkg/fileset/parser.go`, `parser_helpers.go`, `scanner.go`, `helpers.go`)
3. Completion parser (`pkg/fileset/completion_parser.go`, `completion_helpers.go`)
4. AST (`pkg/fileset/ast.go`)

Check: `lib/src/fileset.pest`, `lib/src/fileset.rs` (BUILTIN_FUNCTION_MAP), `docs/filesets.md`

### CLI, Concepts, Bookmarks, Operations, Config, Templates, Git Compat

When jj CLI or concepts change, update the corresponding skill in `skills/`. Source of truth is the jj documentation and CLI help output.

Check: `docs/`, `jj --help`, <https://jj-vcs.github.io/jj/latest/>
