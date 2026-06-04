# jj Configuration Reference

jj uses TOML configuration files with layered precedence. This reference covers all config sections, key settings, and the config CLI.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/config/>. For **CLI commands**, see [cli.md](cli.md). For **revset aliases**, see [revsets.md](revsets.md).


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

