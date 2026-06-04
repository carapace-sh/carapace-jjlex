# jj Operation Log Reference

Every jj command that modifies the repo creates an **operation** in the operation log. This provides a complete undo history, far more powerful than Git's per-ref reflog.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/operation-log/>. For **CLI commands**, see [cli.md](cli.md).


## Viewing Operations

```bash
jj op log              # View operation history
jj op show              # Show changes in the latest operation
jj op show <op-id>      # Show changes in a specific operation
jj op diff              # Compare changes between two operations
jj op diff --from <op1> --to <op2>  # Compare specific operations
```

### `jj op log` Flags

| Flag | Description |
|------|-------------|
| `-n`, `--limit <N>` | Limit number of operations |
| `--reversed` | Older operations first |
| `-G`, `--no-graph` | Flat list (no graph) |
| `-T`, `--template <TMPL>` | Custom output template |
| `-d`, `--op-diff` | Show repo changes at each operation |
| `-p`, `--patch` | Show patch (implies --op-diff) |
| Diff format flags | `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words` |
| `--show-changes-in <REVSETS>` | Show only changed revisions matching revset |

### `jj op show` Flags

| Flag | Description |
|------|-------------|
| `<OP>` | Operation to show. Default: `@` |
| `-G`, `--no-graph` | Flat list |
| `-T`, `--template <TMPL>` | Custom template |
| `--no-op-diff` | Don't show operation diff |
| Diff format flags | `-p`, `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words` |

### `jj op diff` Flags

| Flag | Description |
|------|-------------|
| `--operation <OP>` | Show changes in this operation |
| `-f`, `--from <OP>` | Show changes from this operation |
| `-t`, `--to <OP>` | Show changes to this operation |
| `--show-changes-in <REVSETS>` | Filter changed revisions |
| Diff format flags | `-p`, `-s`, `--stat`, etc. |


## Restoring and Reverting

### Restore to a Previous State

```bash
jj op restore <op-id>   # Restore the entire repo to the state at this operation
```

This creates a new operation that sets the repo state to match the target operation's view.

### Revert a Specific Operation

```bash
jj op revert <op-id>    # Revert the effect of a specific operation
```

Unlike `jj undo`, this can revert any operation — not just the most recent one. It creates a new operation that reverses the target operation's changes.

Both commands accept a `--what` option:
- `--what repo` — restore/revert only repo state (bookmarks, heads, etc.)
- `--what remote-tracking` — restore/revert only remote-tracking bookmark positions


## The `--at-op` Flag

The `--at-operation` (alias `--at-op`) global flag loads the repo at a specific operation:

```bash
jj log --at-op <op-id>          # View log at a historical operation
jj status --at-op <op-id>       # View status at a historical operation
jj diff --at-op <op-id>         # View diff at a historical operation
```

### Important Behaviors

- **Read-only recommended** — while mutation commands work with `--at-op`, they simulate running the command back when that operation was most recent
- **Working copy not snapshotted** — automatic working-copy snapshotting is disabled
- **`@` resolves to the working-copy commit** recorded in that operation's view, not the current working copy

### Operation ID Formats

- Use the short operation ID shown by `jj op log`
- `@` refers to the current operation
- `@-` refers to the parent of `@`
- Operation IDs support the same postfix operators as revsets: `@-` (parent), `@+` (child)


## Integration with Other Operations

### `jj op integrate`

```bash
jj op integrate <op-id>   # Make an external operation part of the operation log
```

This is used to integrate operations that were created outside the normal jj command flow (e.g., by `jj git import`).

