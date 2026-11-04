# jj Template Language Reference

The jj template language is a functional language for customizing command output. Most display commands accept `-T`/`--template` to customize output format.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/templates/>. For **CLI commands**, see [cli.md](cli.md).


## Operators

Listed from **strongest** to **weakest** binding:

| Priority | Operator | Meaning |
|----------|----------|---------|
| 1 | `x.f()` | Method call |
| 2 | `f(x)` | Function call |
| 3 | `{k => v, ..}` | Map literal |
| 4 | `-x` | Negate integer |
| 5 | `!x` | Logical not |
| 6 | `p:x` | String pattern |
| 7 | `x * y`, `x / y`, `x % y` | Multiplication/division/remainder |
| 8 | `x + y`, `x - y` | Addition/subtraction |
| 9 | `x >= y`, `x > y`, `x <= y`, `x < y` | Comparison |
| 10 | `x == y`, `x != y` | Equality |
| 11 | `x && y` | Logical and |
| 12 | `x \|\| y` | Logical or |
| 13 | `x ++ y` | Concatenation (weakest) |

Use parentheses to override precedence.

Map literals (`{ template => template, ... }`) are comma-separated with an optional trailing comma; `{}` is an empty map. They are accepted anywhere a primary expression is allowed, and are mainly used as JSON objects in `json()`:

```javascript
json({"commit_id" => commit_id, "tags" => tags})
```


## Global Functions

### Formatting Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `fill(width, content, [break_words])` | Template | Fill lines at width; break words longer than width if `break_words` is true |
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
| `json(value)` | String | Serialize to JSON; accepts map literals: `json({"a" => value})` |
| `revset(revset)` | List\<Commit\> | Evaluate a revset during templating; constants are evaluated at parse time, dynamic queries e.g. `revset(self.change_id())` per call |

### Conditional/Utility Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `if(condition, then, [else])` | Any | Conditional evaluation |
| `try(expr, [fallback...])` | Any | First successful result, suppresses runtime errors |
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


### Optional Named Arguments

Functions accept optional named parameters by keyword-style syntax (these are the only keyword arguments jj declares; other optional parameters such as `hyperlink`'s fallback, `if`'s else branch, or `config`'s name are positional only):

```javascript
fill(80, description, break_words=true)
pad_start(10, change_id, fill_char=".")
truncate_start(10, description, ellipsis="...")
git_web_url(remote="upstream")
```


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

