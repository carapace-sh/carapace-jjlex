---
name: jj-config
description: >
  Reference for jj (Jujutsu VCS) configuration — config file locations, layered
  precedence, all config sections and key settings, conditional config, and the
  config CLI commands. Triggers on: "jj config", "jj configuration", "jj settings",
  "jj toml", "jj config edit", "jj config set", "jj config list", "jj config path".
user-invocable: false
---

# jj Configuration Reference

jj uses TOML configuration files with layered precedence. This reference covers all config sections, key settings, and the config CLI.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/config/>. For **CLI commands**, see the `jj-cli` skill. For **revset aliases**, see the `jj-revsets` skill.

---

## Config Levels

jj loads config settings in order of precedence (later overrides earlier):

| Level | Location | Edit Command |
|-------|----------|-------------|
| 1. Built-in | Hardcoded defaults (not editable) | — |
| 2. User | `~/.jjconfig.toml` or `$XDG_CONFIG_HOME/jj/config.toml` | `jj config edit --user` |
| 3. Repo | Per-repo config (not inside repo dir for security) | `jj config edit --repo` |
| 4. Workspace | Per-workspace config | `jj config edit --workspace` |
| 5. Command-line | `--config <NAME=VALUE>` or `--config-file <PATH>` | — |

Find config file paths with `jj config path --user/--repo/--workspace`.

User config also loads from `<platform-config-dir>/jj/conf.d/*.toml` files in lexicographic order. Override the config directory with the `JJ_CONFIG` environment variable.

---

## Config CLI Commands

| Command | Description |
|---------|-------------|
| `jj config edit --user/--repo/--workspace` | Open config file in editor |
| `jj config get <NAME>` | Print a config value |
| `jj config list [NAME]` | List config variables and values |
| `jj config path --user/--repo/--workspace` | Print config file path |
| `jj config set --user/--repo/--workspace <NAME> <VALUE>` | Set a config value |
| `jj config unset --user/--repo/--workspace <NAME>` | Unset a config value |

`jj config list` flags: `--include-defaults`, `--include-overridden`, `-T/--template`.

---

## Config Sections

### `[user]` — Identity

```toml
[user]
name = "Your Name"
email = "you@example.com"
```

### `[ui]` — UI Settings

```toml
[ui]
default-command = ["log", "--reversed"]  # Default: "log"
editor = "nvim"                           # Default: $VISUAL or $EDITOR or "vi"
diff-formatter = ":git"                   # ":color-words", ":summary", or external tool
pager = "less -FRX"                       # Pager command
paginate = "auto"                         # "auto", "never"
color = "auto"                             # "always", "never", "debug", "auto"
graph.style = "curved"                    # "curved", "square", "ascii", "ascii-large"
show-cryptographic-signatures = false     # Show signature info in log
conflict-marker-style = "diff"            # "diff", "snapshot", "git"
```

### `[revsets]` — Default Revsets

```toml
[revsets]
log = "present(@) | ancestors(immutable_heads().., 2) | trunk()"  # Default revset for jj log
short-prefixes = "<revsets.log>"  # Defaults to same as revsets.log
bookmark-advance-from = "heads(::to & bookmarks())"
bookmark-advance-to = "@"
op-diff-changes-in = "mutable() | trunk()"
simplify-parents = "reachable(@, mutable())"
```

### `[revset-aliases]` — Custom Revset Aliases

```toml
[revset-aliases]
'HEAD' = '@-'
'user()' = 'author("me@example.org")'
'user(x)' = 'author(x) | committer(x)'
'immutable_heads()' = 'builtin_immutable_heads() | release@origin'
```

Alias functions can be overloaded by parameter count. Built-in functions are shadowed by name and cannot co-exist with aliases.

### `[fileset-aliases]` — Custom Fileset Aliases

```toml
[fileset-aliases]
'LOCK' = '**/Cargo.lock | **/package-lock.json | **/uv.lock'
'not:x' = '~x'
```

### `[templates]` — Template Customization

```toml
[templates]
log = "builtin_log_detailed"
config_list = "builtin_config_list_detailed"
draft_commit_description = '''...'''
new_description = '''...'''
commit_trailers = '''...'''
git_push_bookmark = '"push-" ++ change_id.short()'
```

### `[template-aliases]` — Template Aliases

```toml
[template-aliases]
'format_short_id(id)' = 'id.shortest(12)'
```

### `[aliases]` — Command Aliases

```toml
[aliases]
l = ["log", "-r", "(main..@):: | (main..@)-"]
```

### `[colors]` — Color and Style

```toml
[colors]
commit_id = "green"
change_id = "#ff1525"
"working_copy commit_id" = { underline = true }
```

Available colors: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `default` (with `bright_` prefix variants). Styles: `bold`, `italic`, `underline`, `blink`, `reverse`, `hidden`.

Discover all labels with `jj log --color=debug`.

### `[git]` — Git Integration

```toml
[git]
colocate = true                           # Default: true for init/clone
fetch = "origin"                          # Or list: '["origin", "upstream"]'
push = "origin"                           # Default remote for push
private-commits = "description('wip:*')"  # Commits that shouldn't be pushed
executable-path = "/usr/bin/git"           # Git binary path
sign-on-push = false                      # Sign commits during push
```

### `[remotes.<name>]` — Remote Configuration

```toml
[remotes.origin]
fetch-bookmarks = "~gh-pages"             # Pattern to filter fetched bookmarks
fetch-tags = "v*"                          # Pattern to filter fetched tags
auto-track-bookmarks = "*"                # Auto-track all fetched bookmarks
```

### `[merge-tools.<name>]` — Diff/Merge Tools

```toml
[merge-tools.meld]
program = "/usr/bin/meld"
edit-args = ["--newtab", "$left", "$right"]
merge-args = ["$left", "$base", "$right", "-o", "$output", "--auto-merge"]
diff-args = ["--color=always", "$left", "$right"]
```

### `[fix.tools.<name>]` — Code Formatters

```toml
[fix.tools.clang-format]
command = ["/usr/bin/clang-format", "--sort-includes", "--assume-filename=$path"]
patterns = ["glob:'**/*.c'", "glob:'**/*.h'"]
enabled = true
line-range-arg = "--lines=$first:$last"
```

### `[signing]` — Commit Signing

```toml
[signing]
behavior = "own"        # "drop", "keep", "own", "force"
backend = "gpg"         # "gpg", "gpgsm", "ssh"
key = "..."             # Signing key identifier

[signing.backends.gpg]
program = "gpg2"

[signing.backends.ssh]
allowed-signers = "/path/to/allowed-signers"
```

### `[snapshot]` — Snapshot Behavior

```toml
[snapshot]
auto-track = ["glob:'**/*.txt'"]          # Patterns for auto-tracking
max-new-file-size = "10MiB"               # Max size for new files
auto-update-stale = false                  # Auto-update stale working copies
```

### `[working-copy]` — Working Copy Settings

```toml
[working-copy]
eol-conversion = "none"      # "none", "input", "input-output"
exec-bit-change = "respect"  # "respect", "ignore", "auto"
```

### `[merge]` — Merge Behavior

```toml
[merge]
hunk-level = "line"    # "line", "word"
same-change = "keep"   # "keep", "accept"
```

### `[fsmonitor]` — Filesystem Monitoring

```toml
[fsmonitor]
backend = "watchman"    # Enable watchman-based fsmonitor

[fsmonitor.watchman]
register-snapshot-trigger = true
```

### `[diff.color-words]` and `[diff.git]` — Diff Settings

```toml
[diff.color-words]
max-inline-alternation = 3
context = 3

[diff.git]
context = 3
show-path-prefix = true
```

---

## Conditional Config

Apply config settings conditionally using `[[--scope]]` sections:

```toml
[[--scope]]
--when.repositories = ["~/oss"]
[--scope.user]
email = "oss@example.org"

[[--scope]]
--when.hostnames = ["work-laptop", "work-desktop"]
[--scope.ui]
pager = "delta"

[[--scope]]
--when.commands = ["diff", "show"]
[--scope.ui]
pager = "delta"

[[--scope]]
--when.platforms = ["windows"]
[--scope.ui]
editor = "code -w"

[[--scope]]
--when.environments = ["CI=true"]
[--scope.ui]
paginate = "never"
```

### Available Conditions

| Condition | Matches |
|-----------|---------|
| `--when.repositories` | Repository path prefix |
| `--when.workspaces` | Workspace path prefix |
| `--when.hostnames` | `operation.hostname` value |
| `--when.commands` | Subcommands by prefix (e.g., `["diff", "log"]`) |
| `--when.platforms` | Platform: `windows`, `linux`, `macos`, `unix`, etc. |
| `--when.environments` | Environment variable values |

---

## JSON Schema

Add this line to config files for editor validation:

```toml
#:schema https://docs.jj-vcs.dev/latest/config-schema.json
```