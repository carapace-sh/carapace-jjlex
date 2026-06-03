---
name: jj-concepts
description: >
  Reference for jj (Jujutsu VCS) core concepts — the working-copy model, change
  IDs vs commit IDs, automatic descendant rebasing, committable conflicts,
  immutable revisions, the virtual root commit, and how jj differs from Git's
  mental model. Triggers on: "jj concept", "jj model", "jj working copy",
  "jj change id", "jj commit id", "jj immutable", "jj conflict", "jj rebase",
  "jj descendant", "jj root commit", "jj stash", "jj index", "jj staging".
user-invocable: true
---

# jj Core Concepts

Jujutsu (jj) is a version-control system that uses Git as its storage backend but provides a fundamentally different user model. Understanding these concepts is essential for using jj effectively.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/>. For **CLI commands and flags**, see the `jj-cli` skill. For **bookmark details**, see the `jj-bookmarks` skill. For **operation log details**, see the `jj-operations` skill.

---

## Working Copy Is a Commit

Unlike Git, the jj working copy **is** a commit — the `@` revision. Every jj command automatically snapshots the working copy before operating. There is no staging area / index.

**Three-step command lifecycle:**
1. Snapshot the working copy (recorded as an operation)
2. Perform the operation, creating new commits in memory
3. Update the working copy to match the new state

If step 3 doesn't complete (crash, ^C), the working copy becomes "stale". Fix with `jj workspace update-stale`.

### No Staging Area

jj has no staging area / index. Instead, use:
- `jj split` — split a commit into two (replaces `git add -p; git commit`)
- `jj squash -i` — interactively move changes into parent (replaces `git add -p; git commit --amend`)
- `jj squash <file>` — move a specific file's changes into parent

### Auto-Tracking

Files are implicitly tracked by default. Adding a file automatically includes it; removing one automatically untracks it. Configure with `snapshot.auto-track`. Use `jj file track` / `jj file untrack` for manual control.

### Stale Working Copy

The working copy tracks which operation it was last updated to (stored in `.jj/working_copy/`). If this becomes out of sync (e.g., concurrent commands, crashed process, operation lost via `jj op abandon`), run `jj workspace update-stale` to recover.

---

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

---

## No Active Branch / No Detached HEAD

jj has no concept of a "current branch" or "detached HEAD". Bookmarks are just named pointers — they don't move automatically on new commits. The `@` symbol simply points to whichever commit you're currently editing.

This means:
- There is no need to "switch branches" — just `jj new main` or `jj edit <rev>`
- Creating a commit doesn't update any bookmark
- You're always in a "detached HEAD" state, and that's normal

---

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

---

## Automatic Descendant Rebasing

When a commit is rewritten (via rebase, amend, describe, squash, etc.), all its descendants are automatically rebased onto the new version. Bookmarks and the working copy also update automatically.

This means:
- `jj describe @-` automatically rebases `@` onto the new version of `@-`
- `jj rebase -s foo -o bar` moves `foo` and all its descendants
- No manual descendant management is ever needed

### Evil Merges

Unlike Git, jj considers changes in merge commits to be first-class. There's no concept of "evil merge" — content changes in merge commits are valid and tracked properly.

---

## Immutable Revisions

jj protects certain revisions from accidental modification (describe, edit, rebase, etc.). The default immutable set is:

```
::trunk() | tags() | untracked_remote_bookmarks()
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

---

## Virtual Root Commit

Every jj repo has a single virtual root commit (ID `00000000...`). All commits descend from this root. There is no Git-style "unborn branch" state.

Key properties:
- `root()` revset function returns the root commit
- The root commit is always immutable
- The root commit has no parents and an empty tree

---

## Visible and Hidden Revisions

| Concept | Meaning |
|---------|---------|
| **Visible** | Commits reachable from visible heads (working copies, bookmarks, tags, explicit references) |
| **Hidden** | Commits not visible — abandoned or rewritten commits |
| **Abandoned** | Explicitly abandoned with `jj abandon` — becomes hidden |

Hidden commits are only accessible by commit ID or when explicitly referenced. They are included in `all()`, `x::`, `~x`, etc., but not in `..visible_heads()`.

---

## Multiple Workspaces

jj supports multiple workspaces backed by a single repo, similar to Git worktrees but more powerful:

- Each workspace has its own `@` (working-copy commit)
- `jj workspace add <path>` — create a new workspace
- `jj workspace list` — list all workspaces
- `jj workspace forget <name>` — remove a workspace from tracking
- Workspaces share bookmarks, tags, and commit graph

---

## Git Interoperability

jj uses Git as its storage backend. Key points:

- **Colocated repos** (default): `.jj` and `.git` coexist in the same directory
- `jj git export` — push jj state to the underlying Git repo
- `jj git import` — pull Git state into jj
- Colocated repos auto-import/export on every jj command
- Git commands can be used alongside jj (read-only recommended)
- `jj git colocation enable/disable/status` — manage colocation