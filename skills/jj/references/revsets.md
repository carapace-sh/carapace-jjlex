# jj Revset Reference

Jujutsu revsets are expressions that select sets of revisions (commits). The language consists of **symbols**, **operators**, and **functions**.

> **Source of truth**: `lib/src/revset.pest` (grammar), `lib/src/revset.rs` (`BUILTIN_FUNCTION_MAP`), `docs/revsets.md`. Discrepancies between docs and source are noted.

## Hidden Revisions

Most revsets search only **visible** commits. Hidden commits are included only if explicitly mentioned (by commit ID, `<name>@<remote>`, or `at_operation()`). If hidden commits are specified, their ancestors also become available. They are included in `all()`, `x..`, `~x`, etc., but **not** in `..visible_heads()`.

## Symbols

| Symbol | Meaning |
|--------|---------|
| `@` | Working copy commit in the current workspace |
| `<name>@` | Working copy commit in workspace `<name>` |
| `<name>@<remote>` | Remote-tracking bookmark |
| `<hex prefix>` | Commit ID or change ID prefix (must be unique) |
| `"<symbol>"` or `'<symbol>'` | Quoted symbol — prevents interpretation as expression. Git branches with special characters like brackets (e.g. `parents(`) are displayed quoted by jj (`"parents("`) and must be referenced with quotes in revsets |
| `<change_id>/<offset>` | Change ID with offset, for hidden or divergent changes |

### Symbol Resolution Priority

1. Tag name
2. Bookmark name
3. Git ref
4. Commit ID or change ID

Override with `commit_id()` or `change_id()`.

## Operators

Listed from **strongest** to **weakest** binding:

### 1. Function Call — `f(x)`

Highest binding power.

### 2. Postfix Operators

| Op | Meaning | Example |
|----|---------|---------|
| `x-` | Parents of `x` (can be empty) | `@-` |
| `x+` | Children of `x` (can be empty) | `root()+` |

Repeatable: `x---` = great-grandparents, `x+++` = great-grandchildren.

### 3. Pattern — `p:x`

Pattern prefix applied to value. See [String Patterns](#string-patterns).

### 4. Range / DAG Range Operators

| Op | Meaning | Equivalent |
|----|---------|-----------|
| `::x` | Ancestors of `x` (inclusive) | `root()::x` |
| `..x` | Ancestors of `x`, excluding root | `::x ~ root()` |
| `x::` | Descendants of `x` (inclusive) | `x::visible_heads()` |
| `x..` | Non-ancestors of `x` | `~::x` |
| `x::y` | Ancestors of `y` reachable from `x` | `x:: & ::y` |
| `x..y` | Ancestors of `y` not ancestors of `x` | `::y ~ ::x` |
| `::` | All visible commits | `all()` |
| `..` | All visible commits excluding root | `~root()` |

**Rules:**
- Range operators **cannot nest** without parentheses: `::foo::` is a syntax error
- Prefix `::`/`..` do **not** allow whitespace before operand: `":: foo"` is a syntax error
- `..` does **not** distribute over `|` on its left side
- `x..y` = Git's `x..y`; `x::y` = Git's `--ancestry-path x..y`

### 5. Prefix Negate — `~x`

Revisions **not** in `x`. `~` is also the infix difference operator — context determines which.

### 6. Intersection / Difference

| Op | Meaning |
|----|---------|
| `x & y` | Intersection — in both `x` and `y` |
| `x ~ y` | Difference — in `x` but not in `y` |

Left-associative.

### 7. Union — `x | y`

Lowest binding power. Left-associative. Revisions in `x` or `y` (or both).

Use parentheses to override precedence: `(x & y) | z`, `~(x | y)`.

### Operator Examples

Given this DAG:
```
o D
|\
| o C
| |
o | B
|/
o A
|
o root()
```

| Expression | Result |
|-----------|--------|
| `D-` | `{C, B}` |
| `root()-` | `{}` (empty) |
| `A+` | `{B, C}` |
| `A::` | `{D, C, B, A}` |
| `A..` | `{D, C, B}` |
| `::A` | `{A, root()}` |
| `..A` | `{A}` |
| `B::D` | `{D, B}` |
| `B..D` | `{D, C}` |
| `::` | `{D, C, B, A, root()}` |
| `..` | `{D, C, B, A}` |

## Functions

Function argument notation: `[arg]` = optional. Named arguments can be specified by label: `remote=origin`.

### Traversal

| Function | Signature | Description |
|----------|-----------|-------------|
| `parents` | `parents(x, [depth])` | `parents(x)` = `x-`. With depth: `parents(x, 3)` = `x---` |
| `children` | `children(x, [depth])` | `children(x)` = `x+`. With depth: `children(x, 3)` = `x+++` |
| `ancestors` | `ancestors(x, [depth])` | `ancestors(x)` = `::x`. With depth: limited to given depth |
| `descendants` | `descendants(x, [depth])` | `descendants(x)` = `x::`. With depth: limited to given depth |
| `first_parent` | `first_parent(x, [depth])` | Like `parents(x)`, but for merges returns only first parent. With depth: `first_parent(x, 2)` = `first_parent(first_parent(x))` |
| `first_ancestors` | `first_ancestors(x, [depth])` | Like `ancestors(x)`, but only traverses first parent. Useful for Git-style first-parent history |
| `reachable` | `reachable(srcs, domain)` | All commits reachable from `srcs` within `domain`, traversing all parent and child edges. `srcs` outside `domain` are not considered |
| `connected` | `connected(x)` | Same as `x::x`. Useful when `x` includes several commits |

### Set Operations

| Function | Signature | Description |
|----------|-----------|-------------|
| `all` | `all()` | All visible commits and ancestors of explicitly mentioned commits |
| `none` | `none()` | No commits |
| `heads` | `heads(x)` | Commits in `x` that are not ancestors of other commits in `x`. Eqv: `x ~ ::x-`. **Note:** differs from Mercurial's `heads(x)` |
| `roots` | `roots(x)` | Commits in `x` that are not descendants of other commits in `x`. Eqv: `x ~ x+::`. **Note:** differs from Mercurial's `roots(x)` |
| `latest` | `latest(x, [count])` | Latest `count` commits by committer timestamp. Default count = 1 |
| `fork_point` | `fork_point(x)` | Common ancestor(s) with no descendant that is also a common ancestor. If `x` is a single commit, returns that commit |
| `bisect` | `bisect(x)` | Commits where about half the input set are descendants. Handles non-linear history imperfectly |
| `exactly` | `exactly(x, count)` | Returns `x` if exactly `count` commits, otherwise errors. Useful with `count=1` |
| `present` | `present(x)` | Same as `x`, but evaluates to `none()` if any commit in `x` doesn't exist |
| `coalesce` | `coalesce([revsets...])` | Commits in the first revset that doesn't evaluate to `none()`. Zero args returns `none()`. At least one argument is recommended |

### Identity

| Function | Signature | Description |
|----------|-----------|-------------|
| `change_id` | `change_id(prefix)` | Commits with given change ID prefix. Divergent changes resolve to multiple commits. Non-unique prefix is an error. Unmatched prefix is not an error |
| `commit_id` | `commit_id(prefix)` | Commits with given commit ID prefix. Non-unique prefix is an error. Unmatched prefix is not an error |

### Bookmarks and Tags

| Function | Signature | Description |
|----------|-----------|-------------|
| `bookmarks` | `bookmarks([pattern])` | All local bookmark targets. With pattern: filter by name |
| `remote_bookmarks` | `remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | All remote bookmark targets. Examples: `remote_bookmarks()`, `remote_bookmarks("main")`, `remote_bookmarks("main", "origin")`, `remote_bookmarks(remote="origin")`. Git-tracking bookmarks excluded by default; use `remote=exact:"git"` or `remote=glob:"*"` to include |
| `tracked_remote_bookmarks` | `tracked_remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | Targets of tracked remote bookmarks. Same optional args as `remote_bookmarks()` |
| `untracked_remote_bookmarks` | `untracked_remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | Targets of untracked remote bookmarks. Same optional args as `remote_bookmarks()` |
| `tags` | `tags([pattern])` | All tag targets. With pattern: filter by name |
| `remote_tags` | `remote_tags([name_pattern], [[remote=]remote_pattern])` | All remote tag targets. Same optional args pattern as `remote_bookmarks()` |
| `tracked_remote_tags` | `tracked_remote_tags([name_pattern], [[remote=]remote_pattern])` | Targets of tracked remote tags |
| `untracked_remote_tags` | `untracked_remote_tags([name_pattern], [[remote=]remote_pattern])` | Targets of untracked remote tags |

### Filtering

| Function | Signature | Description |
|----------|-----------|-------------|
| `merges` | `merges()` | Merge commits (2+ parents) |
| `description` | `description(pattern)` | Commits with description matching pattern. `description(exact:"")` matches empty description; `description(exact:"foo\n")` matches `"foo\n"` |
| `subject` | `subject(pattern)` | Commits with subject (first line of description) matching pattern |
| `author` | `author(pattern)` | `author_name(pattern) \| author_email(pattern)` |
| `author_name` | `author_name(pattern)` | Commits with author name matching pattern |
| `author_email` | `author_email(pattern)` | Commits with author email matching pattern |
| `author_date` | `author_date(pattern)` | Commits with author date matching date pattern |
| `mine` | `mine()` | `author_email(exact-i:<user-email>)` |
| `committer` | `committer(pattern)` | `committer_name(pattern) \| committer_email(pattern)` |
| `committer_name` | `committer_name(pattern)` | Commits with committer name matching pattern |
| `committer_email` | `committer_email(pattern)` | Commits with committer email matching pattern |
| `committer_date` | `committer_date(pattern)` | Commits with committer date matching date pattern |
| `signed` | `signed()` | Cryptographically signed commits |
| `empty` | `empty()` | Commits modifying no files. Includes `merges()` without user modifications and `root()` |
| `conflicts` | `conflicts()` | Commits with conflicted files |
| `divergent` | `divergent()` | Divergent commits |

### File and Diff

| Function | Signature | Description |
|----------|-----------|-------------|
| `files` | `files(expression)` | Commits modifying paths matching fileset expression. Paths relative to cwd. Directory matches all files in it: `files(foo)` matches `foo`, `foo/bar`, but not `foobar`. Some patterns need quoting: `files(".")` |
| `diff_lines` | `diff_lines(text, [files])` | Commits with diffs matching text pattern (both added and removed lines). Optional `files` narrows search. **Note:** the docs previously called this `diff_contains`; `diff_contains` is now a deprecated alias |
| `diff_lines_added` | `diff_lines_added(text, [files])` | Like `diff_lines()` but matches only added lines |
| `diff_lines_removed` | `diff_lines_removed(text, [files])` | Like `diff_lines()` but matches only removed lines |

### Workspace and Operations

| Function | Signature | Description |
|----------|-----------|-------------|
| `working_copies` | `working_copies()` | Working copy commits across all workspaces |
| `at_operation` | `at_operation(op, x)` | Evaluate `x` at the specified operation. E.g. `at_operation(@-, visible_heads())` = heads visible at previous operation. Brings hidden commits from that operation into scope |
| `visible_heads` | `visible_heads()` | All visible heads. Same as `heads(all())` if no hidden revisions mentioned |
| `root` | `root()` | The virtual root commit (oldest ancestor of all others) |

### Deprecated Functions

| Function | Replacement | Notes |
|----------|-------------|-------|
| `git_refs` | `remote_bookmarks()` / `tags()` | Removed in jj 0.43+ |
| `git_head` | `first_parent(@)` | Removed in jj 0.43+ |
| `diff_contains` | `diff_lines()` | Deprecated alias, to be removed in jj 0.44+ |

## String Patterns

Functions that perform string matching accept a pattern argument. Quotes around the value are optional for bare identifiers.

| Kind | Syntax | Description |
|------|--------|-------------|
| `substring` | `substring:"str"` | Matches strings containing `str` (default kind when unqualified) |
| `exact` | `exact:"str"` | Matches strings exactly equal to `str` |
| `glob` | `glob:"pattern"` | Matches with Unix shell wildcards (`*`, `?`) |
| `regex` | `regex:"pattern"` | Matches substrings with regular expression |

Append `-i` for case-insensitive matching: `glob-i:"fix*jpeg*"`, `exact-i:"FOO"`, `substring-i:"bar"`, `regex-i:"pattern"`.

**Default pattern kind is changing:** Currently unquoted values default to `substring:`. A future release will change the default to `glob:`. Enable the new behavior with `ui.revsets-use-glob-by-default=true`.

### String Pattern Operators

Pattern arguments can be combined with logical operators:
- `~x` — not x
- `x & y` — both x and y
- `x | y` — either x or y
- `x ~ y` — x but not y

Example: `bookmarks(~glob:"ci/*")` selects bookmarks not matching `ci/*`.

### Pattern Argument Syntax for Bookmark/Tag Functions

Functions like `bookmarks()`, `remote_bookmarks()`, `tags()`, `remote_tags()` and their tracked/untracked variants accept:
- No argument — match all
- A positional pattern — filter by name: `bookmarks(push)` (matches `push-123`, `repushed`)
- A named `remote=` pattern — filter by remote: `remote_bookmarks(remote=origin)`
- Both: `remote_bookmarks("main", "origin")` or `remote_bookmarks("main", remote="origin")`

The name and remote patterns themselves accept string pattern kinds: `remote_bookmarks(glob:"feat*", remote=exact:"origin")`.

## Date Patterns

Date-matching functions (`author_date`, `committer_date`) require a pattern kind:

| Kind | Syntax | Description |
|------|--------|-------------|
| `after` | `after:"2024-02-01"` | Matches dates at or after the given date |
| `before` | `before:"2024-02-01"` | Matches dates before (not including) the given date |

Date string forms:
- `2024-02-01`
- `2024-02-01T12:00:00`
- `2024-02-01T12:00:00-08:00`
- `2024-02-01 12:00:00`
- `2 days ago`, `5 minutes ago`
- `yesterday`, `yesterday 5pm`, `yesterday 10:30`, `yesterday 15:30`

## Aliases

Define custom symbols, functions, and pattern aliases in config:

```toml
[revset-aliases]
'HEAD' = '@-'
'user()' = 'user("me@example.org")'
'user(x)' = 'author(x) | committer(x)'
```

Alias functions can be overloaded by parameter count. Built-in functions are shadowed by name and cannot co-exist with aliases.

### Built-in Aliases

| Alias | Definition | Notes |
|-------|------------|-------|
| `trunk()` | `present(trunk()) \| tags() \| untracked_remote_bookmarks()` (effectively) | Resolves to the default bookmark head. Falls back to `root()` if unresolved. Override with `[revset-aliases] 'trunk()' = 'your-bookmark@your-remote'` |
| `builtin_immutable_heads()` | `present(trunk()) \| tags() \| untracked_remote_bookmarks()` | Default for `immutable_heads()`. Don't redefine this; redefine `immutable_heads()` instead |
| `immutable_heads()` | `builtin_immutable_heads()` | Override as needed |
| `immutable()` | `::(immutable_heads() \| root())` | Don't redefine |
| `mutable()` | `~immutable()` | Don't redefine |
| `visible()` | `::visible_heads()` | Equal to `all()` unless hidden revisions are mentioned |
| `hidden()` | `~visible()` | Empty unless hidden revisions are mentioned. Not the set of all previously visible commits |

## Grammar Summary

Derived from the Pest grammar (`lib/src/revset.pest`):

```
expression       = (negate_op ~ whitespace*)* ~ range_expression
                   ~ (whitespace* ~ infix_op ~ whitespace* ~ (negate_op ~ whitespace*)* ~ range_expression)*
range_expression = neighbors_expression ~ range_ops ~ neighbors_expression
                 | neighbors_expression ~ range_post_ops
                 | range_pre_ops ~ neighbors_expression
                 | neighbors_expression
                 | range_all_ops
neighbors_expression = primary ~ (parents_op | children_op)*
primary          = "(" ~ whitespace* ~ expression ~ whitespace* ~ ")"
                 | function
                 | pattern
                 | symbol ~ at_op ~ symbol    # name@remote
                 | symbol ~ at_op             # name@  (workspace)
                 | symbol
                 | at_op                      # @ (current workspace)
pattern          = strict_identifier ~ ":" ~ pattern_value_expression
pattern_value_expression = neighbors_expression    # no ranges allowed in pattern value
function         = function_name ~ "(" ~ whitespace* ~ function_arguments ~ whitespace* ~ ")"
function_arguments = argument ~ (whitespace* ~ "," ~ whitespace* ~ argument)* ~ (whitespace* ~ ",")?
                 | ""    # empty argument list
argument         = keyword_argument | expression
keyword_argument = strict_identifier ~ whitespace* ~ "=" ~ whitespace* ~ expression
symbol           = identifier | string_literal | raw_string_literal

identifier       = identifier_part ~ (("." | "-"+ | "+") ~ identifier_part)*
identifier_part  = (XID_CONTINUE | "_" | "*" | "/")+
strict_identifier = strict_identifier_part ~ (("." | "-" | "+") ~ strict_identifier_part)*
strict_identifier_part = (ASCII_ALPHANUMERIC | "_" | "/")+
function_name    = (ASCII_ALPHA | "_") ~ (ASCII_ALPHANUMERIC | "_")*

infix_op         = union_op | intersection_op | difference_op
                 # compat_add_op("+") and compat_sub_op("-") are in the grammar
                 # but produce errors (suggest | and ~ respectively)
negate_op        = "~"
parents_op       = "-"
children_op      = "+"
                 # compat_parents_op("^") is in the grammar but produces an error (suggests -)
range_ops        = "::" | ".."    # also compat_dag_range_op(":") — error, suggests ::
range_pre_ops    = "::" | ".."    # also compat_dag_range_pre_op(":") — error
range_post_ops  = "::" | ".."    # also compat_dag_range_post_op(":") — error
range_all_ops    = "::" | ".."
```

**Key grammar notes:**
- `-` is postfix (parents), NOT infix — `foo - bar` is a syntax error (use `~`)
- `+` is postfix (children), NOT infix — `foo + bar` is a syntax error (use `|`)
- `^` is in the grammar as a compat postfix operator but always produces an error (suggests `-`)
- `:` alone is in the grammar as a compat range operator but always produces an error (suggests `::`)
- Pattern colon (`name:value`) requires no whitespace around `:`
- Pattern value is `neighbors_expression` — postfix ops allowed but NO ranges: `x:y::z` parses as `(x:y)::z`, not `x:(y::z)`
- String literals: `"..."` with escapes (`\t`, `\r`, `\n`, `\0`, `\e`, `\xHH`, `\"`, `\\`), `'...'` raw strings (no escape processing)
- Trailing comma allowed in function calls: `bookmarks(a,)`
- Empty function calls: `visible_heads()` (no args)
- `function_name` uses strict identifier rules (ASCII alphanumeric + underscore, must start with letter or underscore)
- Whitespace: space, tab, CR, LF, FF (`\x0c`)