---
name: jj-templates
description: >
  Reference for the jj (Jujutsu VCS) template language — the functional language
  for customizing command output with -T/--template flags. Covers operators,
  types, methods, global functions, and examples. Triggers on: "jj template",
  "jj -T", "jj --template", "template language", "jj format", "jj output format",
  "jj custom log", "template alias".
user-invocable: false
---

# jj Template Language Reference

The jj template language is a functional language for customizing command output. Most display commands accept `-T`/`--template` to customize output format.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/templates/>. For **CLI commands**, see the `jj-cli` skill.

---

## Syntax

Templates are expressions that can reference keywords (context objects), call methods, use operators, and call global functions.

### Literals

- Boolean: `true`, `false`
- Integer: `42`
- String: `"hello"` (with escapes: `\t`, `\r`, `\n`, `\0`, `\e`, `\xHH`, `\"`, `\\`), `'raw string'`

### Keywords

All 0-argument methods of the context object are available as keywords:
- In `jj log` and similar commands: all `Commit` methods (e.g., `commit_id` = `self.commit_id()`)
- In `jj op log`: all `Operation` methods

### String Escape Sequences (double-quoted only)

| Escape | Meaning |
|--------|---------|
| `\t` | Tab |
| `\r` | Carriage return |
| `\n` | Newline |
| `\0` | Null |
| `\e` | Escape |
| `\xHH` | Hex byte |
| `\"` | Double quote |
| `\\` | Backslash |

---

## Operators

Listed from **strongest** to **weakest** binding:

| Priority | Operator | Meaning |
|----------|----------|---------|
| 1 | `x.f()` | Method call |
| 2 | `f(x)` | Function call |
| 3 | `-x` | Negate integer |
| 4 | `!x` | Logical not |
| 5 | `p:x` | String pattern |
| 6 | `x * y`, `x / y`, `x % y` | Multiplication/division/remainder |
| 7 | `x + y`, `x - y` | Addition/subtraction |
| 8 | `x >= y`, `x > y`, `x <= y`, `x < y` | Comparison |
| 9 | `x == y`, `x != y` | Equality |
| 10 | `x && y` | Logical and |
| 11 | `x \|\| y` | Logical or |
| 12 | `x ++ y` | Concatenation (weakest) |

Use parentheses to override precedence.

---

## Types and Methods

### Commit

Available as the context object in `jj log`, `jj show`, `jj obslog`.

| Method | Returns | Description |
|--------|---------|-------------|
| `.commit_id()` | CommitId | Full commit ID |
| `.change_id()` | ChangeId | Change ID |
| `.description()` | String | Commit description |
| `.parents()` | List\<Commit\> | Parent commits |
| `.author()` | Signature | Author info |
| `.committer()` | Signature | Committer info |
| `.trailers()` | List\<Trailer\> | Git trailers from description |
| `.bookmarks()` | List\<CommitRef\> | All bookmarks (local + remote) |
| `.local_bookmarks()` | List\<CommitRef\> | Local bookmarks only |
| `.remote_bookmarks()` | List\<CommitRef\> | Remote bookmarks only |
| `.tags()` | List\<CommitRef\> | All tags |
| `.mine()` | Boolean | True if author's email matches current user |
| `.hidden()` | Boolean | True if commit is abandoned |
| `.divergent()` | Boolean | True if multiple commits share change ID |
| `.conflict()` | Boolean | True if contains merge conflicts |
| `.empty()` | Boolean | True if no files modified |
| `.root()` | Boolean | True if root commit |
| `.diff([files])` | TreeDiff | Changes from parents |
| `.files([files])` | List\<TreeEntry\> | Files in commit |
| `.conflicted_files()` | List\<TreeEntry\> | Conflicted files |
| `.signature()` | Option\<CryptographicSignature\> | Cryptographic signature |
| `.working_copies()` | List\<WorkspaceRef\> | Working copies of this commit |
| `.current_working_copy()` | Boolean | True if this is the current working-copy commit |
| `.local_tags()` | List\<CommitRef\> | Local tags pointing to this commit |
| `.remote_tags()` | List\<CommitRef\> | Remote tags pointing to this commit |
| `.change_offset()` | Option\<Integer\> | Offset for hidden/divergent change IDs |
| `.immutable()` | Boolean | True if commit is immutable |
| `.contained_in(revset)` | Boolean | True if commit is in the given revset |

### ChangeId

| Method | Returns | Description |
|--------|---------|-------------|
| `.short([len])` | String | Short hex representation (default 12) |
| `.shortest([min_len])` | ShortestIdPrefix | Shortest unique prefix |
| `.normal_hex()` | String | Normal hex (0-9a-f) instead of reversed (z-k) |

### ShortestIdPrefix

| Method | Returns | Description |
|--------|---------|-------------|
| `.prefix()` | String | The shortest unique prefix string |
| `.rest()` | String | The remaining characters after the prefix |
| `.upper()` | ShortestIdPrefix | Uppercase variant |
| `.lower()` | ShortestIdPrefix | Lowercase variant |

### CommitId

| Method | Returns | Description |
|--------|---------|-------------|
| `.short([len])` | String | Short hex representation |
| `.shortest([min_len])` | ShortestIdPrefix | Shortest unique prefix |

### String

| Method | Returns | Description |
|--------|---------|-------------|
| `.len()` | Integer | UTF-8 byte length |
| `.contains(needle)` | Boolean | Check substring |
| `.match(pattern)` | String | Extract first match |
| `.starts_with(needle)` | Boolean | Check prefix |
| `.ends_with(needle)` | Boolean | Check suffix |
| `.remove_prefix(needle)` | String | Remove prefix if present |
| `.remove_suffix(needle)` | String | Remove suffix if present |
| `.trim()` | String | Remove leading/trailing whitespace |
| `.trim_start()` | String | Remove leading whitespace |
| `.trim_end()` | String | Remove trailing whitespace |
| `.upper()` | String | Uppercase |
| `.lower()` | String | Lowercase |
| `.substr(start, [end])` | String | Extract substring (byte indices) |
| `.first_line()` | String | First line |
| `.lines()` | List\<String\> | Split into lines |
| `.split(pattern, [limit])` | List\<String\> | Split by pattern |
| `.replace(pattern, replacement, [limit])` | String | Replace with capture groups ($0, $1...) |
| `.escape_json()` | String | JSON-serialized |

### List

| Method | Returns | Description |
|--------|---------|-------------|
| `.len()` | Integer | Element count |
| `.join(separator)` | Template | Concatenate with separator |
| `.filter(\|x\| expr)` | List | Filter by predicate |
| `.map(\|x\| expr)` | AnyList | Transform each element |
| `.any(\|x\| expr)` | Boolean | Any element satisfies |
| `.all(\|x\| expr)` | Boolean | All elements satisfy |
| `.first()` | T | First element |
| `.last()` | T | Last element |
| `.get(index)` | T | Element at index |
| `.reverse()` | List | Reverse order |
| `.skip(count)` | List | Skip first N elements |
| `.take(count)` | List | Take first N elements |

**Type-specific list methods:**

- `List<Trailer>.contains_key(key)` — True if any trailer has the given key

### Timestamp

| Method | Returns | Description |
|--------|---------|-------------|
| `.ago()` | String | Relative timestamp (e.g., "2 hours ago") |
| `.format(format)` | String | strftime format |
| `.utc()` | Timestamp | Convert to UTC |
| `.local()` | Timestamp | Convert to local timezone |
| `.after(date)` | Boolean | Date comparison |
| `.before(date)` | Boolean | Date comparison |
| `.since(start)` | TimestampRange | Duration since start timestamp |

### TimestampRange

| Method | Returns | Description |
|--------|---------|-------------|
| `.start()` | Timestamp | Start of the range |
| `.end()` | Timestamp | End of the range |
| `.duration()` | String | Duration as human-readable string |

### Signature

| Method | Returns | Description |
|--------|---------|-------------|
| `.name()` | String | Author/committer name |
| `.email()` | Email | Email info |
| `.timestamp()` | Timestamp | Timestamp |

### Email

| Method | Returns | Description |
|--------|---------|-------------|
| `.local()` | String | Local part (before @) |
| `.domain()` | String | Domain part (after @) |

### TreeDiff

| Method | Returns | Description |
|--------|---------|-------------|
| `.files()` | List\<TreeDiffEntry\> | Changed files |
| `.color_words([context])` | Template | Word-level colored diff |
| `.git([context])` | Template | Git diff format |
| `.stat([width])` | DiffStats | Diff statistics |
| `.summary()` | Template | Status/path list |

### DiffStats

| Method | Returns | Description |
|--------|---------|-------------|
| `.files()` | List\<DiffStatEntry\> | Per-file stats |
| `.total_added()` | Integer | Total lines added |
| `.total_removed()` | Integer | Total lines removed |

### DiffStatEntry

| Method | Returns | Description |
|--------|---------|-------------|
| `.bytes_delta()` | Integer | Size difference |
| `.lines_added()` | Integer | Lines added |
| `.lines_removed()` | Integer | Lines removed |
| `.path()` | RepoPath | File path |
| `.status()` | String | `modified`, `added`, `removed`, `copied`, `renamed` |
| `.status_char()` | String | `M`, `A`, `D`, `C`, `R` |

### Trailer

| Method | Returns | Description |
|--------|---------|-------------|
| `.key()` | String | Trailer key |
| `.value()` | String | Trailer value |

### CryptographicSignature

| Method | Returns | Description |
|--------|---------|-------------|
| `.status()` | String | Signature status: `good`, `bad`, `unknown`, or `invalid` |
| `.key()` | String | Key identifier |
| `.display()` | String | Display string |

### CommitRef

| Method | Returns | Description |
|--------|---------|-------------|
| `.name()` | RefSymbol | Bookmark or tag name |
| `.remote()` | Option\<RefSymbol\> | Remote name (if remote bookmark/tag) |
| `.present()` | Boolean | True if the ref is present (not a conflict) |
| `.conflict()` | Boolean | True if the ref is conflicted |
| `.normal_target()` | Option\<Commit\> | The target commit (if not conflicted) |
| `.removed_targets()` | List\<Commit\> | Removed targets (in a conflict) |
| `.added_targets()` | List\<Commit\> | Added targets (in a conflict) |
| `.tracked()` | Boolean | True if the remote ref is tracked |
| `.tracking_present()` | Boolean | True if the local tracking ref exists |
| `.tracking_ahead_count()` | SizeHint | How many local commits ahead of remote |
| `.tracking_behind_count()` | SizeHint | How many remote commits ahead of local |
| `.synced()` | Boolean | True if local and remote are in sync |

### TreeEntry

| Method | Returns | Description |
|--------|---------|-------------|
| `.path()` | RepoPath | File path |
| `.conflict()` | Boolean | True if the file has conflicts |
| `.conflict_side_count()` | Integer | Number of sides in the merge conflict |
| `.file_type()` | String | `file`, `symlink`, `tree`, `git-submodule`, or `conflict` |
| `.executable()` | Boolean | True if the file is executable |

### TreeDiffEntry

| Method | Returns | Description |
|--------|---------|-------------|
| `.path()` | RepoPath | File path |
| `.display_diff_path()` | String | Display path (accounts for copy/rename) |
| `.status()` | String | `modified`, `added`, `removed`, `copied`, or `renamed` |
| `.status_char()` | String | `M`, `A`, `D`, `C`, or `R` |
| `.source()` | TreeEntry | Source file entry |
| `.target()` | TreeEntry | Target file entry |

### RepoPath

| Method | Returns | Description |
|--------|---------|-------------|
| `.absolute()` | String | Absolute path |
| `.display()` | String | Path relative to current working directory |
| `.parent()` | Option\<RepoPath\> | Parent directory path |

### SizeHint

| Method | Returns | Description |
|--------|---------|-------------|
| `.lower()` | Integer | Minimum count |
| `.upper()` | Option\<Integer\> | Maximum count (None if unknown) |
| `.exact()` | Option\<Integer\> | Exact count (None if approximate) |
| `.zero()` | Boolean | True if count is definitely zero |

### ConfigValue

| Method | Returns | Description |
|--------|---------|-------------|
| `.as_boolean()` | Boolean | Convert to boolean |
| `.as_integer()` | Integer | Convert to integer |
| `.as_string()` | String | Convert to string |
| `.as_string_list()` | List\<String\> | Convert to string list |

### AnnotationLine

Available as context in `jj file annotate -T`.

| Method | Returns | Description |
|--------|---------|-------------|
| `.commit()` | Commit | Commit that authored this line |
| `.content()` | ByteString | Line content |
| `.line_number()` | Integer | Line number |
| `.original_line_number()` | Integer | Original line number |
| `.first_line_in_hunk()` | Boolean | True if this is the first line in a hunk |

### Operation

Available as context in `jj op log -T`.

| Method | Returns | Description |
|--------|---------|-------------|
| `.current_operation()` | Boolean | True if this is the current operation |
| `.description()` | String | Operation description |
| `.id()` | OperationId | Operation ID |
| `.attributes()` | String | Operation attributes |
| `.time()` | TimestampRange | Time range |
| `.user()` | String | Username |
| `.snapshot()` | Boolean | True if this is a snapshot operation |
| `.workspace_name()` | String | Workspace name |
| `.root()` | Boolean | True if this is the root operation |
| `.parents()` | List\<Operation\> | Parent operations |

### OperationId

| Method | Returns | Description |
|--------|---------|-------------|
| `.short([len])` | String | Short hex representation |

### CommitEvolutionEntry

Available as context in `jj evolog -T`.

| Method | Returns | Description |
|--------|---------|-------------|
| `.commit()` | Commit | The commit at this evolution step |
| `.operation()` | Operation | The operation that created this step |
| `.predecessors()` | List\<Commit\> | Predecessor commits |
| `.inter_diff([files])` | TreeDiff | Diff from predecessor to this commit |

### WorkspaceRef

| Method | Returns | Description |
|--------|---------|-------------|
| `.name()` | RefSymbol | Workspace name |
| `.target()` | Commit | Working-copy commit for this workspace |
| `.root()` | Template | Workspace root path |

---

## Global Functions

### Formatting Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `fill(width, content)` | Template | Fill lines at width |
| `indent(prefix, content)` | Template | Indent non-empty lines with prefix |
| `pad_start(width, content, [fill_char])` | Template | Left-justify with fill chars |
| `pad_end(width, content, [fill_char])` | Template | Right-justify with fill chars |
| `pad_centered(width, content, [fill_char])` | Template | Center with fill chars |
| `truncate_start(width, content, [ellipsis])` | Template | Truncate from start |
| `truncate_end(width, content, [ellipsis])` | Template | Truncate from end |

### Text/Content Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `hash(content)` | String | Hash and return hex digest |
| `label(label, content)` | Template | Apply color label |
| `hyperlink(url, text, [fallback])` | Template | Render OSC 8 hyperlink |
| `raw_escape_sequence(content)` | Template | Preserve escape sequences |
| `stringify(content)` | String | Remove color labels |
| `json(value)` | String | Serialize to JSON |

### Conditional/Utility Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `if(condition, then, [else])` | Any | Conditional evaluation |
| `coalesce(content...)` | Template | First non-empty content |
| `concat(content...)` | Template | Concatenate all |
| `join(separator, content...)` | Template | Insert separator between items |
| `separate(separator, content...)` | Template | Insert separator between non-empty items |
| `surround(prefix, suffix, content)` | Template | Wrap non-empty content |
| `config(name)` | Option\<ConfigValue\> | Look up config value |
| `git_web_url([remote])` | String | Convert git URL to HTTPS browse URL |

### Replace Function

```javascript
replace(pattern, content, replacement) -> Template
```

Replace matches using a lambda with `RegexCaptures`:
- `.get(index)` — access capture groups
- `.name(name)` — access named groups
- `.len()` — number of captures

---

## Template Aliases

Define reusable template expressions in config:

```toml
[template-aliases]
'format_short_id(id)' = 'id.shortest(12)'
'commit_change_ids' = '''
concat(
  format_field("Commit ID", commit_id),
  format_field("Change ID", change_id),
)
'''
'format_field(key, value)' = 'key ++ ": " ++ value ++ "\n"'
```

---

## Color Labels

Use `label(label_name, content)` to apply custom colors:

```bash
jj log -T '"ID: " ++ self.id().short().label("id short")'
```

Discover all labels with `--color=debug`:

```bash
jj log --color=debug
```

Define custom colors in config:

```toml
[colors]
"id short" = "red"
commit_id = { fg = "green", bold = true }
```

---

## Examples

```bash
# Short commit IDs of working-copy parents
jj log -G -r @ -T 'parents.map(|c| c.commit_id().short()).join(",")'

# Machine-readable full IDs
jj log -G -T 'commit_id ++ " " ++ change_id ++ "\n"'

# Description with fallback
jj log -G -r @ -T 'coalesce(description, "(no description set)\n")'

# Custom log format with diff stats
jj log -T 'change_id.shortest() ++ " " ++ label("diff", if(empty, "", " " ++ diff.stat().total_added() ++ if(diff.stat().total_added(), " ", "") ++ diff.stat().total_removed())) ++ "\n"'

# List bookmarks pointing to each commit
jj log -T 'change_id.short() ++ " " ++ bookmarks.join(", ") ++ "\n"'

# Author and date
jj log -T 'change_id.short() ++ " " ++ author.name() ++ " " ++ author.timestamp().ago() ++ "\n"'

# Custom op log format
jj op log -T '"ID: " ++ self.id().short().label("id short")'

# Bookmarks and tags inline
jj log -T 'separate(" ", change_id.shortest(), bookmarks, tags) ++ "\n"'
```