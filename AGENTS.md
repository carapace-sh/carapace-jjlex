# AGENTS.md

## Project Overview

Go library for parsing [jj (Jujutsu)](https://github.com/jj-vcs/jj) revset expressions into an AST, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework). The module path is `github.com/carapace-sh/carapace-jjlex`.

## Commands

```sh
go test ./...          # run all tests
go test ./pkg/revset/ # run revset package tests only
go build ./...         # build all packages
go run main/main.go "<expr>"                    # parse expression, output AST as JSON
go run main/main.go --complete <cursor> "<expr>" # completion context as JSON
```

No Makefile, no linter config, no CI config present.

## Architecture

Two independent recursive-descent parsers in `pkg/revset/`:

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Strict: rejects partial/invalid input.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input, cursor)` → `*CompletionContext` describing what tokens are valid at the cursor position. Tolerant: recovers from errors at cursor to report expectations.

Both parsers implement the same operator precedence hierarchy (levels 0-6) but independently. The completion parser mirrors the main parser's structure but stops at the cursor and records expectations instead of building a full AST.

### File responsibilities

| File | Purpose |
|---|---|
| `pkg/revset/ast.go` | AST node types (`Expression`, `UnaryOp`, `BinaryOp`, `ExpressionKind`), payload structs (`IdentifierExpr`, `BinaryExpr`, etc.), type-erased accessor methods |
| `pkg/revset/span.go` | `Span` (Start/End byte offsets) and `Pos` types |
| `pkg/revset/parser.go` | Main parser + public API: `Parse()`, `IsIdentifier()`, `ParseSymbol()`, `Format()`. Also contains shared helper functions used by both parsers. |
| `pkg/revset/format.go` | AST → string formatting with precedence-aware parenthesization |
| `pkg/revset/completion.go` | Completion context types: `CompletionContext`, `ExpectedToken`, `ValidOperator`, `FunctionContext` |
| `pkg/revset/completion_parser.go` | Completion parser: `ParseForCompletion()`, `compParser` type, duplicate validation checks (`isFunctionNameCheck`, `isStrictIdentifierCheck`) |
| `main/main.go` | CLI entrypoint (parses args, calls library, outputs JSON). No tests. |

## Key Patterns & Gotchas

### Expression uses a type-erased payload pattern

`Expression.payload` is `any`; accessors (`.Identifier()`, `.UnaryOp()`, `.BinaryLHS()`, etc.) do type checks and return zero values on kind mismatch. Always check `Kind` before calling accessors.

### Two parsers must stay in sync

When modifying operator precedence or parsing rules in `parser.go`, the same changes must be mirrored in `completion_parser.go`. They share helper functions (`isWhitespace`, `isIdentifierPart`, `isStrictIdentifierPart`, `splitIdentifierParts`) but have independent parser types (`parser` vs `compParser`). The completion parser also duplicates `isFunctionName` and `isStrictIdentifier` as `isFunctionNameCheck` and `isStrictIdentifierCheck` — keep these in sync.

### Identifier rules are non-obvious

- Identifiers allow internal `.`, `-`, `+` as connectors (e.g. `foo.bar-v1+7`)
- Multiple consecutive `-` are allowed (`foo--bar`), but `+` and `.` are not repeatable
- `*` and `/` are valid identifier characters (for glob patterns)
- `isIdentifierPart` uses Unicode categories (XID_CONTINUE); `isStrictIdentifierPart` is ASCII-only (used for pattern names, keyword args, and function names)
- Function names require standard identifier syntax (alphanumeric + underscore, must start with letter or underscore)
- Identifiers cannot start or end with `.`; `+` and `-` cannot appear at edges of parts

### Operator ambiguity

- `~` is both prefix (negate) and infix (difference) — context-dependent
- `-` is postfix (parents) but NOT infix — `foo - bar` is an error (suggests `~`)
- `+` is postfix (children) but NOT infix — `foo + bar` is an error (suggests `|`)
- `:` alone is always an error (suggests `::`)
- `^` is always an error (suggests `-`)
- Range operators (`::`, `..`) cannot be nested without parentheses
- Prefix `::` and `..` do not allow whitespace between operator and operand — `":: foo"` is a syntax error

### Pattern syntax (`name:value`)

No whitespace allowed around the `:` in patterns. `exact: foo` is an error. The pattern name must be a strict identifier; the value is parsed as a `neighbors_expression` (postfix ops only, no ranges). Pattern is right-associative: `x:y:z` = `x:(y:z)`.

### `@` suffix parsing

After an identifier or string, `@` triggers workspace/remote symbol parsing:
- `main@` → `KindAtWorkspace` (workspace with no remote)
- `main@origin` → `KindRemoteSymbol` (name + remote)
- Both name and remote parts can be quoted strings
- `"@"` inside quotes is NOT interpreted as workspace syntax — it's `KindString`

### Span tracking

Spans exclude trailing whitespace. The parser tracks `lastContent` (position after last non-whitespace) and clips expression spans accordingly. Parenthesized expressions include the parens in their span.

### Union flattening

The `|` operator flattens into a single `KindUnionAll` node rather than creating nested binary trees. `foo | bar | baz` produces one `UnionAllExpr` with 3 nodes.

### ParseError chaining

`ParseError` has an unexported `origin` field accessible via `.Origin()`. This is for error chaining but the field is not set in current code.

### go.mod has unused dependency

`go.mod` requires `github.com/carapace-sh/revset` but gopls reports it's not used. Don't add dependencies to this — it may be intentional or a leftover.

## Testing

- Tests use standard `testing` package only (no testify or other deps)
- `revset_test.go` — main parser tests using helpers: `testParseKind`, `testParseUnaryOp`, `testParseBinaryOp`, `testParseEqual`, `testParseString`, `testParseError`
- `completion_test.go` — completion tests using `assertHasExpected`, `assertHasOperator`
- Test helpers use `t.Helper()` and `t.Fatalf`/`t.Errorf` patterns
- `main/` has no tests
- No external dependencies (pure stdlib)
- Passing `cursor=-1` to `ParseForCompletion` defaults to `len(input)` (end of input)
