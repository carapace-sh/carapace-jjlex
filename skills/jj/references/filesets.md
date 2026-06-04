# jj Fileset Reference

Jujutsu filesets are expressions that select sets of files. The language consists of **file patterns**, **operators**, and **functions**.

> **Source of truth**: `lib/src/fileset.pest` (grammar), `lib/src/fileset.rs` (`BUILTIN_FUNCTION_MAP`), `docs/filesets.md`. Discrepancies between docs and source are noted.

## Operators

Listed from **strongest** to **weakest** binding power:

### 1. Function Call — `f(x)`

Highest binding power.

### 2. Pattern — `p:x`

Pattern prefix applied to a primary expression value. No whitespace around `:`. Right-associative: `x:y:z` = `x:(y:z)`.

See [File Patterns](#file-patterns) for available pattern kinds.

### 3. Prefix Negate — `~x`

Matches everything **not** in `x`.

### 4. Intersection / Difference

| Op | Meaning |
|----|---------|
| `x & y` | Intersection — matches files in both `x` and `y` |
| `x ~ y` | Difference — matches files in `x` but not in `y` |

Left-associative. Infix `~` is difference (not negate).

### 5. Union — `x | y`

Lowest binding power. Left-associative. Flattens: `x | y | z` produces one `UnionAll` node.

Use parentheses to override precedence: `~(x | y)`, `(x | y) & z`.

### Operator Examples

```
~foo           # everything except foo
foo & bar      # files matching both
foo ~ bar      # files matching foo but not bar
foo | bar      # files matching foo or bar
~(foo | bar)  # everything except foo or bar
```

## File Patterns

Pattern syntax: `kind:value` where `kind` is a strict identifier and `value` is a primary expression (typically a string or identifier). No whitespace around `:`.

| Kind | Description |
|------|-------------|
| `cwd` | Matches cwd-relative path prefix (file or files under directory recursively) |
| `file` or `cwd-file` | Matches cwd-relative file (or exact) path |
| `glob` or `cwd-glob` | Matches file paths with cwd-relative Unix-style shell wildcard pattern |
| `prefix-glob` or `cwd-prefix-glob` | Like `glob:`, but also matches path prefix (equivalent to `glob:"*.d" \| glob:"*.d/**"`) |
| `root` | Matches workspace-relative path prefix |
| `root-file` | Matches workspace-relative file (or exact) path |
| `root-glob` | Matches file paths with workspace-relative wildcard pattern |
| `root-prefix-glob` | Like `root-glob:`, but also matches path prefix |

### Case-Insensitive Variants

Append `-i` to any glob pattern kind:

| Kind | Description |
|------|-------------|
| `glob-i` or `cwd-glob-i` | Case-insensitive cwd-relative glob |
| `prefix-glob-i` or `cwd-prefix-glob-i` | Case-insensitive cwd-relative prefix-glob |
| `root-glob-i` | Case-insensitive workspace-relative glob |
| `root-prefix-glob-i` | Case-insensitive workspace-relative prefix-glob |

### Glob Syntax

Uses Unix-style shell wildcards (via globset crate): `*` matches any filename, `**` matches any path, `?` matches single character, `[...]` matches character class.

Examples:
- `glob:"*.c"` — all `.c` files in cwd non-recursively
- `glob:"**/*.rs"` — all `.rs` files recursively
- `root-glob:"*.c"` — all `.c` files in workspace root

## Built-in Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `all()` | `all()` | Matches everything |
| `none()` | `none()` | Matches nothing |

Only two built-in functions currently. Additional functions can be defined via aliases.

## Identifiers

Fileset identifiers allow a broader character set than revset identifiers:

```
identifier = (XID_CONTINUE | "+" | "-" | "." | "@" | "_" | "*" | "?" | "[" | "]" | "/" | "\\")+
```

This means path separators (`/`, `\`), glob characters (`*`, `?`, `[`, `]`), and `+`, `-`, `.`, `@` are all valid inside identifiers **without quoting**.

### Quoting Rules

- File names containing **whitespace or meta characters** must be quoted
- As a special case, quotes can be omitted if the expression has **no operators or function calls** (bare string fallback)
- Glob characters (`*`, `?`, `[`, `]`) are **NOT** meta characters in fileset — they're valid identifier chars

### Strict Identifiers

Used for pattern names and function names:

```
strict_identifier = strict_identifier_part ~ ("-" ~ strict_identifier_part)*
strict_identifier_part = (ASCII_ALPHANUMERIC | "_")+
```

Function names additionally require starting with a letter or underscore: `(ASCII_ALPHA | "_") ~ (ASCII_ALPHANUMERIC | "_")*`.

## Bare Strings and Bare String Patterns

When an expression **cannot be parsed** (contains spaces without operators, or would be a syntax error), jj falls back to:

1. **Bare string pattern**: `kind:value` where `value` is a bare string (allows spaces)
2. **Bare string**: Just the raw text (treated as a path)

```
bare_string = ( ASCII_ALPHANUMERIC | " " | "+" | "-" | "." | "@" | "_" | "*" | "?" | "[" | "]" | "/" | "\\" | '\u{80}'..'\u{10ffff}' )+
```

Examples:
- `jj diff 'Foo Bar'` — shell quotes required, inner quotes optional
- `jj diff ~"Foo Bar"` — both shell and inner quotes required
- `jj diff '~glob:**/*.rs'` — glob chars are not meta, still need shell quotes

## String Literals

- Double-quoted: `"..."` with escapes (`\t`, `\r`, `\n`, `\0`, `\e`, `\xHH`, `\"`, `\\`)
- Single-quoted raw: `'...'` — no escape processing

## Function Calls

```
function = function_name ~ "(" ~ whitespace* ~ function_arguments ~ whitespace* ~ ")"
function_arguments = expression ~ (whitespace* ~ "," ~ whitespace* ~ expression)* ~ (whitespace* ~ ",")?
                  | ""
```

- Trailing commas are allowed: `all(foo,)`
- No keyword arguments (unlike revsets)

## Aliases

Define custom fileset symbols, functions, and patterns in config:

```toml
[fileset-aliases]
'LOCK' = '**/Cargo.lock | **/package-lock.json | **/uv.lock'
'not:x' = '~x'
```

Alias functions can be overloaded by parameter count. Built-in functions are shadowed by name and cannot co-exist with aliases.

## Examples

```bash
# Show diff excluding Cargo.lock
jj diff '~Cargo.lock'

# List files in src excluding Rust sources
jj file list 'src ~ glob:"**/*.rs"'

# Split a revision, putting foo into the second commit
jj split '~foo'

# Match all .txt files in current directory
jj file list 'glob:"*.txt"'

# Match all .rs files recursively
jj file list 'glob:"**/*.rs"'

# Combine patterns
jj file list 'glob:"*.rs" & ~glob:"test*"'

# Case-insensitive glob
jj file list 'glob-i:"*.TXT"'
```

## Grammar Summary

Derived from the Pest grammar (`lib/src/fileset.pest`):

```
expression       = (negate_op ~ whitespace*)* ~ primary
                   ~ (whitespace* ~ infix_op ~ whitespace* ~ (negate_op ~ whitespace*)* ~ primary)*
primary          = "(" ~ whitespace* ~ expression ~ whitespace* ~ ")"
                 | function
                 | pattern
                 | identifier
                 | string_literal
                 | raw_string_literal
pattern          = strict_identifier ~ pattern_kind_op ~ primary
function         = function_name ~ "(" ~ whitespace* ~ function_arguments ~ whitespace* ~ ")"
function_arguments = expression ~ (whitespace* ~ "," ~ whitespace* ~ expression)* ~ (whitespace* ~ ",")?
                 | ""
function_name    = (ASCII_ALPHA | "_") ~ (ASCII_ALPHANUMERIC | "_")*
identifier       = (XID_CONTINUE | "+" | "-" | "." | "@" | "_" | "*" | "?" | "[" | "]" | "/" | "\\")+
strict_identifier = strict_identifier_part ~ ("-" ~ strict_identifier_part)*
strict_identifier_part = (ASCII_ALPHANUMERIC | "_")+

# Bare string fallback (for program_or_bare_string rule)
bare_string      = ( ASCII_ALPHANUMERIC | " " | "+" | "-" | "." | "@" | "_" | "*" | "?"
                   | "[" | "]" | "/" | "\\" | '\u{80}'..'\u{10ffff}' )+
bare_string_pattern = strict_identifier ~ pattern_kind_op ~ bare_string

pattern_kind_op  = ":"
negate_op        = "~"
union_op         = "|"
intersection_op  = "&"
difference_op    = "~"
infix_ops        = union_op | intersection_op | difference_op
```

**Key grammar notes:**
- `~` is both prefix (negate) and infix (difference) — context determines which
- No `::`, `..`, `-`, `+` postfix operators (unlike revsets)
- No `@` workspace syntax (unlike revsets)
- No keyword arguments in function calls (unlike revsets)
- Pattern colon (`name:value`) requires no whitespace around `:`
- String literals: `"..."` with escapes, `'...'` raw strings
- Trailing comma allowed in function calls: `all(foo,)`
- Whitespace: space, tab, CR, LF, FF (`\x0c`)