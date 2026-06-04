# jj Core Concepts

Jujutsu (jj) is a version-control system that uses Git as its storage backend but provides a fundamentally different user model. Understanding these concepts is essential for using jj effectively.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/>. For **CLI commands and flags**, see [cli.md](cli.md). For **bookmark details**, see [bookmarks.md](bookmarks.md). For **operation log details**, see [operations.md](operations.md).


## Change ID vs Commit ID

Every commit has two identifiers:

| ID | Example | Stability |
|----|---------|-----------|
| **Change ID** | `kntqzsqt` | Stable — persists across rewrites (amends, rebases). Uses letters `k`–`z` instead of `0`–`9`/`a`–`f` |
| **Commit ID** | `d7439b06` | Unstable — changes when the commit content changes. Git-compatible hex hash |

Use **change IDs** to refer to commits in jj commands — they survive history rewrites. The commit ID is the underlying Git hash and changes whenever the commit is modified.

### Symbol Resolution Priority

When a symbol could match multiple things, jj resolves in this order:
1. Tag name
2. Bookmark name
3. Git ref
4. Commit ID or change ID prefix (must be unique)

Override with `commit_id()` or `change_id()` revset functions.


## Conflicts Are Committable

Conflicts are recorded directly in commits with structured conflict markers. Commands never fail due to unresolved conflicts.

### Conflict Markers

When a commit has conflicts, jj writes structured markers into the files:

```
<<<<<<<
%%%%%%
content from the other side
+++++++
content from this side
>>>>>>>
```

- `<<<<<<<` / `>>>>>>>` — conflict boundaries
- `%%%%%%` — separator between sides
- `+++++++` — separator for added content (in 3-way conflicts)

### Resolution Strategies

1. **Create a new commit on top** — `jj new <conflicted-commit>`, edit files, then `jj squash` the resolution into the conflicted commit
2. **Edit the commit directly** — `jj edit <conflicted-commit>`, resolve, then `jj new` to move on
3. **Use external merge tool** — `jj resolve --tool <tool>`
4. **Restore one side** — `jj restore --from <side> <paths>`


## Immutable Revisions

jj protects certain revisions from accidental modification (describe, edit, rebase, etc.). The default immutable set is:

```
present(trunk()) | tags() | untracked_remote_bookmarks()
```

Override with the `immutable_heads()` revset alias in config:

```toml
[revset-aliases]
'immutable_heads()' = 'builtin_immutable_heads() | release@origin'
```

Use `--ignore-immutable` global flag to bypass immutability checks.

### Built-in Aliases

| Alias | Definition | Notes |
|-------|------------|-------|
| `trunk()` | Effectively `present(trunk()) \| tags() \| untracked_remote_bookmarks()` | Resolves to the default bookmark head, falls back to `root()` |
| `builtin_immutable_heads()` | `present(trunk()) \| tags() \| untracked_remote_bookmarks()` | Don't redefine this; redefine `immutable_heads()` instead |
| `immutable()` | `::(immutable_heads() \| root())` | Don't redefine |
| `mutable()` | `~immutable()` | Don't redefine |


## Visible and Hidden Revisions

| Concept | Meaning |
|---------|---------|
| **Visible** | Commits reachable from visible heads (working copies, bookmarks, tags, explicit references) |
| **Hidden** | Commits not visible — abandoned or rewritten commits |
| **Abandoned** | Explicitly abandoned with `jj abandon` — becomes hidden |

Hidden commits are only accessible by commit ID or when explicitly referenced. They are included in `all()`, `x::`, `~x`, etc., but not in `..visible_heads()`.


## Git Interoperability

jj uses Git as its storage backend. Key points:

- **Colocated repos** (default): `.jj` and `.git` coexist in the same directory
- `jj git export` — push jj state to the underlying Git repo
- `jj git import` — pull Git state into jj
- Colocated repos auto-import/export on every jj command
- Git commands can be used alongside jj (read-only recommended)
- `jj git colocation enable/disable/status` — manage colocation