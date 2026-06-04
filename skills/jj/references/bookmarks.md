# jj Bookmarks Reference

Bookmarks are jj's equivalent of Git branches — named pointers to revisions. However, they behave differently in important ways.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/bookmarks/>. For **CLI commands**, see [cli.md](cli.md). For **revset syntax** used in bookmark commands, see [revsets.md](revsets.md).


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


## Bookmark Revset Functions

See [revsets.md](revsets.md) for full revset syntax. Key bookmark-related functions:

| Function | Signature | Description |
|----------|-----------|-------------|
| `bookmarks` | `bookmarks([pattern])` | All local bookmark targets |
| `remote_bookmarks` | `remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | All remote bookmark targets |
| `tracked_remote_bookmarks` | `tracked_remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | Targets of tracked remote bookmarks |
| `untracked_remote_bookmarks` | `untracked_remote_bookmarks([name_pattern], [[remote=]remote_pattern])` | Targets of untracked remote bookmarks |
| `tags` | `tags([pattern])` | All tag targets |

