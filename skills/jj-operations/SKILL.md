---
name: jj-operations
description: >
  Reference for jj (Jujutsu VCS) operation log — the global undo history that
  records every repo-modifying operation. Covers operation concepts, viewing,
  undoing, redoing, restoring, reverting, abandoning, the --at-op flag, and
  lock-free concurrency. Triggers on: "jj operation", "jj op", "jj undo",
  "jj redo", "jj at-op", "operation log", "jj op log", "jj op restore",
  "jj op revert".
user-invocable: false
---

# jj Operation Log Reference

Every jj command that modifies the repo creates an **operation** in the operation log. This provides a complete undo history, far more powerful than Git's per-ref reflog.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/operation-log/>. For **CLI commands**, see the `jj-cli` skill.

---

## How It Works

Each operation contains:
- A **view** — snapshot of the entire repo state: where each bookmark, tag, and Git ref pointed; the set of heads; working-copy commits in each workspace
- **Parent pointers** — links to the operation(s) immediately before it
- **Metadata** — timestamps, username, hostname, description of what happened

Operations form an append-only log. The latest operation is `@` (not to be confused with `@` for the working-copy commit — context determines which).

---

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

---

## Undoing and Redoing

```bash
jj undo        # Undo the last operation
jj redo        # Redo the most recently undone operation
```

`jj undo` creates a new operation that reverses the effect of the last operation. It does not modify the operation log — the undone operation remains in history. `jj redo` reverses the undo.

You can undo multiple operations by running `jj undo` repeatedly.

---

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

---

## Abandoning Operations

```bash
jj op abandon <op-id>    # Remove an operation from history
```

This permanently removes an operation from the log. Use with caution — it can cause working copies to become stale if the abandoned operation affected their state.

---

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

---

## Lock-Free Concurrency

The operation log enables **lock-free concurrency**:

- Multiple `jj` commands can run simultaneously without corrupting the repo
- Works even across different machines accessing the repo via distributed file systems
- Each command loads the repo at the latest operation
- Conflicts are reported via `jj status` and `jj log`
- If concurrent commands conflict, the later one will create a new operation that merges both changes

---

## Integration with Other Operations

### `jj op integrate`

```bash
jj op integrate <op-id>   # Make an external operation part of the operation log
```

This is used to integrate operations that were created outside the normal jj command flow (e.g., by `jj git import`).

---

## Common Workflows

```bash
# View what happened recently
jj op log

# Undo the last thing you did
jj undo

# Undo multiple operations
jj undo   # undo once
jj undo   # undo twice

# Redo what you just undid
jj redo

# View the repo state at a specific point in time
jj log --at-op <op-id>

# Restore the entire repo to an earlier state
jj op restore <op-id>

# Revert a specific earlier operation (not just the last one)
jj op revert <op-id>

# See what changed in an operation
jj op show <op-id>
```