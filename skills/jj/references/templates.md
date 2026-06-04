# jj Template Language Reference

The jj template language is a functional language for customizing command output. Most display commands accept `-T`/`--template` to customize output format.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/templates/>. For **CLI commands**, see [cli.md](cli.md).


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

