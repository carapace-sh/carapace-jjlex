# AGENTS.md

## Project Overview

Go library for parsing [jj (Jujutsu)](https://github.com/jj-vcs/jj) revset expressions into an AST, with completion support. Part of the [carapace-sh](https://github.com/carapace-sh) ecosystem (shell completion framework).

## Commands

```sh
go test ./...          # run all tests
go build ./...         # build all packages
go run main/main.go "<expr>"                    # parse expression, output AST as JSON
go run main/main.go --complete <cursor> "<expr>" # completion context as JSON
```

## Architecture

Two independent recursive-descent parsers in the same package:

- **`parser.go`** — Full parser. `Parse()` → `*Expression` AST with spans. Strict: rejects partial/invalid input.
- **`completion_parser.go`** — Completion parser. `ParseForCompletion(input, cursor)` → `*CompletionContext` describing what tokens are valid at the cursor position. Tolerant: recovers from errors at cursor to report expectations.

Both parsers implement the same operator precedence hierarchy (levels 0-6) but independently. The completion parser mirrors the main parser's structure but stops at the cursor and records expectations instead of building a full AST.

### File responsibilities

| File | Purpose |
|---|---|
| `ast.go` | AST node types (`Expression`, `UnaryOp`, `BinaryOp`, `ExpressionKind`), payload structs, accessor methods |
| `span.go` | `Span` (Start/End byte offsets) and `Pos` types |
| `parser.go` | Main parser + public API: `Parse()`, `IsIdentifier()`, `ParseSymbol()`, `Format()` |
| `format.go` | AST → string formatting with precedence-aware parenthesization |
| `completion.go` | Completion context types: `CompletionContext`, `ExpectedToken`, `ValidOperator`, `FunctionContext` |
| `completion_parser.go` | Completion parser: `ParseForCompletion()` and the `compParser` type |
| `main/main.go` | CLI entrypoint (parses args, calls library, outputs JSON) |

## Key Patterns & Gotchas

### Expression uses a type-erased payload pattern

`Expression.payload` is `any`; accessors (`.Identifier()`, `.UnaryOp()`, `.BinaryLHS()`, etc.) do type checks and return zero values on kind mismatch. Always check `Kind` before calling accessors.

### Two parsers must stay in sync

When modifying operator precedence or parsing rules in `parser.go`, the same changes must be mirrored in `completion_parser.go`. They share helper functions (`isWhitespace`, `isIdentifierPart`, `isStrictIdentifierPart`, `splitIdentifierParts`) but have independent parser types (`parser` vs `compParser`).

### Identifier rules are non-obvious

- Identifiers allow internal `.`, `-`, `+` as connectors (e.g. `foo.bar-v1+7`)
- Multiple consecutive `-` are allowed (`foo--bar`), but `+` and `.` are not repeatable
- `*` and `/` are valid identifier characters (for glob patterns)
- `isIdentifierPart` uses Unicode categories; `isStrictIdentifierPart` is ASCII-only (used for pattern names, keyword args)
- Function names require standard identifier syntax (alphanumeric + underscore)

### Operator ambiguity

- `~` is both prefix (negate) and infix (difference) — context-dependent
- `-` is postfix (parents) but NOT infix — `foo - bar` is an error (suggests `~`)
- `+` is postfix (children) but NOT infix — `foo + bar` is an error (suggests `|`)
- `:` alone is always an error (suggests `::`)
- `^` is always an error (suggests `-`)
- Range operators (`::`, `..`) cannot be nested without parentheses

### Pattern syntax (`name:value`)

No whitespace allowed around the `:` in patterns. `exact: foo` is an error. The pattern name must be a strict identifier; the value is parsed as a `neighbors_expression` (postfix ops only, no ranges).

### Span tracking

Spans exclude trailing whitespace. The parser tracks `lastContent` (position after last non-whitespace) and clips expression spans accordingly. Parenthesized expressions include the parens in their span.

### Union flattening

The `|` operator flattens into a single `KindUnionAll` node rather than creating nested binary trees. `foo | bar | baz` produces one `UnionAllExpr` with 3 nodes.

### ParseError chaining

`ParseError` has an unexported `origin` field accessible via `.Origin()`. This is for error chaining but the field is not set in current code.

## Testing

- Tests use standard `testing` package (no testify or other deps)
- `revset_test.go` — main parser tests using helpers: `testParseKind`, `testParseUnaryOp`, `testParseBinaryOp`, `testParseEqual`, `testParseString`, `testParseError`
- `completion_test.go` — completion tests using `assertHasExpected`, `assertHasOperator`
- `main/` has no tests
- No external dependencies (pure stdlib)
