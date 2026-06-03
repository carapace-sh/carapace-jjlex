---
name: jj-bookmarks
description: >
  Reference for jj (Jujutsu VCS) bookmarks — named pointers to revisions that
  replace Git branches. Covers bookmark concepts, remote bookmarks, tracking,
  bookmark conflicts, push safety, and CLI commands. Triggers on: "jj bookmark",
  "jj branch", "jj remote bookmark", "jj tracking", "jj push", "jj git push",
  "bookmark conflict", "bookmark advance".
user-invocable: true
---

# jj Bookmarks Reference

Bookmarks are jj's equivalent of Git branches — named pointers to revisions. However, they behave differently in important ways.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/bookmarks/>. For **CLI commands**, see the `jj-cli` skill. For **revset syntax** used in bookmark commands, see the `jj-revsets` skill.

---

## Key Differences from Git Branches

| Aspect | Git Branches | jj Bookmarks |
|--------|-------------|-------------|
| Active branch | Has a checked-out/active branch | No active/checked-out concept |
| Movement | Moves automatically on new commits | Stays where you put them; doesn't auto-move on new commits |
| On rewrite | Can orphan commits if not moved | Automatically follow their target when commits are rewritten |
| Naming | Branch names | Bookmark names (same Git namespace on push) |
| Detached HEAD | A state to avoid | The normal state — `@` simply points to whatever you're editing |

---

## Local Bookmarks

### Creating

```bash
jj bookmark create feature          # Create bookmark at @
jj bookmark create feature -r <rev> # Create bookmark at specific revision
jj bookmark set feature             # Create or update bookmark at @
jj bookmark set feature -r <rev>    # Create or update bookmark at revision
```

### Moving

```bash
jj bookmark move feature            # Move to @ (default target)
jj bookmark move feature --to <rev> # Move to specific revision
jj bookmark move feature --allow-backwards  # Allow backward/sideways move
```

### Deleting and Forgetting

```bash
jj bookmark delete feature          # Delete bookmark (propagates to remotes on push)
jj bookmark forget feature          # Forget without marking as deletion to push
jj bookmark forget feature --include-remotes  # Also forget remote bookmarks
```

### Listing

```bash
jj bookmark list                    # List all bookmarks
jj bookmark list -t                 # List tracked bookmarks
jj bookmark list -a                 # List all remote bookmarks
jj bookmark list -c                 # List conflicted bookmarks
jj bookmark list -r '@'            # List bookmarks pointing at @
jj bookmark list feature            # List bookmarks matching "feature"
```

### Advancing

```bash
jj bookmark advance feature         # Advance bookmark toward @
jj bookmark advance feature --to <rev>  # Advance to specific revision
```

`advance` moves a bookmark forward (toward descendants) but not backward or sideways, unless `--allow-backwards` is used. The default target is `revsets.bookmark-advance-to` (typically `@`).

### Renaming

```bash
jj bookmark rename old-name new-name       # Rename a bookmark
jj bookmark rename old-name new-name --overwrite-existing  # Allow overwriting
```

---

## Remote Bookmarks

Remote bookmarks represent the state of bookmarks on a remote. Format: `<name>@<remote>` (e.g., `main@origin`).

### Key Properties

- jj stores the **last-seen position** of each remote bookmark
- Remote bookmarks can only be updated by fetching from or pushing to the remote
- There is **no way to manually edit** remote bookmark positions
- Use `jj bookmark list --remote <name>` to see remote bookmarks

### Tracking

**Tracked bookmarks** associate a remote bookmark with a local bookmark of the same name. When you fetch:

1. The remote bookmark position is updated (`main@origin` moves)
2. If tracked, the local bookmark also moves to match
3. If both moved, jj merges the changes
4. If they diverged, the local bookmark becomes **conflicted**

```bash
jj bookmark track main@origin       # Start tracking
jj bookmark untrack main@origin     # Stop tracking
jj bookmark list -t                 # List tracked bookmarks
```

### Auto-Tracking

Set in config to automatically track all newly fetched bookmarks:

```toml
[remotes.origin]
auto-track-bookmarks = "*"
```

---

## Bookmark Conflicts

Bookmarks can become **conflicted** when updated both locally and on a remote, or when a rewrite causes a bookmark to point to multiple commits.

### Identifying Conflicts

| Method | Shows |
|--------|-------|
| `jj status` | Conflicted bookmarks with instructions |
| `jj bookmark list` | Detailed conflict information |
| `jj log` | Bookmark name with `??` suffix (e.g., `main??`) |

### Resolving Conflicts

```bash
# Move the bookmark to the desired target
jj bookmark move main --to <desired-revision>

# Or merge/rebase with the remote version
jj new main          # Create merge commit
jj rebase -r main -o <desired-target>
```

---

## Push Safety

When pushing bookmarks, jj performs three safety checks:

1. **Remote position check** — verifies the remote bookmark matches jj's recorded position (like `git push --force-with-lease`). Prevents accidental overwrites.
2. **Conflict check** — local bookmark must not be conflicted.
3. **Tracking check** — remote bookmark must be tracked if it exists on the remote.

### Push Commands

```bash
jj git push --bookmark feature      # Push a single bookmark
jj git push --all                   # Push all bookmarks (including new)
jj git push --tracked               # Push all tracked bookmarks
jj git push --deleted               # Push deletions of deleted bookmarks
jj git push -c <rev>                # Push by creating a bookmark for a commit
jj git push --dry-run               # Preview what will be pushed
```

---

## Bookmark Revset Functions

See the `jj-revsets` skill for full revset syntax. Key bookmark-related functions:

| Function | Signature | Description |
|----------|-----------|-------------|
| `bookmarks` | `bookmarks([pattern])` | All local bookmark targets |
| `remote_bookmarks` | `remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | All remote bookmark targets |
| `tracked_remote_bookmarks` | `tracked_remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | Targets of tracked remote bookmarks |
| `untracked_remote_bookmarks` | `untracked_remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | Targets of untracked remote bookmarks |
| `tags` | `tags([pattern])` | All tag targets |

---

## CLI Reference

See the `jj-cli` skill for full flag details. Quick reference:

| Command | Alias | Description |
|---------|-------|-------------|
| `jj bookmark list` | `jj b l` | List bookmarks |
| `jj bookmark create <name>` | `jj b c` | Create bookmark |
| `jj bookmark move <name>` | `jj b m` | Move bookmark |
| `jj bookmark delete <name>` | `jj b d` | Delete bookmark |
| `jj bookmark forget <name>` | `jj b f` | Forget bookmark |
| `jj bookmark rename <old> <new>` | `jj b r` | Rename bookmark |
| `jj bookmark set <name>` | `jj b s` | Create or update bookmark |
| `jj bookmark advance <name>` | `jj b a` | Advance bookmark |
| `jj bookmark track <name>` | `jj b t` | Track remote bookmark |
| `jj bookmark untrack <name>` | — | Stop tracking remote bookmark |