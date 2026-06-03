---
name: jj-divergence
description: >
  Reference for jj (Jujutsu VCS) divergent changes — what divergence is, how it
  happens, how to identify it, and resolution strategies (abandon, update change ID,
  squash, ignore). Triggers on: "jj divergent", "jj divergence", "divergent change",
  "change offset", "change id suffix", "jj metaedit change id", "resolve divergence".
user-invocable: false
---

# jj Divergent Changes Reference

A **divergent change** occurs when multiple visible commits share the same change ID. Divergent changes appear in `jj log` with a "divergent" label and a change offset suffix (e.g., `mzvwutvl/0`).

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/guides/divergence/>. For **core concepts** about change IDs, see the `jj-concepts` skill. For **CLI commands**, see the `jj-cli` skill.

---

## What Is Divergence?

Normally, when a commit is rewritten, the original becomes hidden and the new commit (successor) is visible. Only one commit with a given change ID is visible at a time.

Divergence happens when **two or more visible commits share the same change ID**. In `jj log`, divergent changes show a change offset:

```
@  mzvwutvl/0 test.user@example.com 2001-02-03 08:05:12 29d07a2d (divergent)
│  a divergent change
```

The change offset `/n` disambiguates which commit you mean:
- `/0` — most recent commit with that change ID
- `/1` — next most recent
- `/2` — and so on

---

## How Divergence Happens

### Hidden Commits Becoming Visible

A hidden commit can become visible again when:

1. **A visible descendant is added locally** — `jj new REV` makes `REV` visible even if it was hidden
2. **A visible descendant is fetched from a remote** — Others may base new commits off your pushed change
3. **Made the working copy** — `jj edit REV` makes `REV` and all its ancestors visible
4. **A bookmark is added** — `jj bookmark create REV` makes `REV` visible with the assumption you are working with that commit again

### Concurrent Modifications

Two different users/processes amending the same change:

- Another author modifies commits in a branch you also modified locally
- Operations on the same change from different workspaces of the same repository
- Two programs modifying the repository simultaneously (e.g., you run `jj describe` while an IDE integration fetches and rebases)

---

## Identifying Divergent Changes

Divergent changes appear in `jj log` with:

- The **"divergent"** label
- A **change offset** suffix (`/0`, `/1`, etc.) after the change ID
- In `jj status`, divergence warnings are shown

### Referring to Divergent Changes

Since the change ID alone is ambiguous for divergent changes, use one of:

- **Commit ID** — `jj log 29d07a2d`
- **Change ID with offset** — `mzvwutvl/0`, `mzvwutvl/1`
- **Full commit ID hash**

Plain change ID references like `mzvwutvl` will fail or resolve to multiple revisions when the change is divergent.

---

## Resolution Strategies

### Strategy 1: Abandon One of the Commits

The simplest approach when one version is clearly wrong or obsolete:

```bash
# Abandon the unwanted commit
jj abandon <unwanted-commit-id>

# Abandon multiple at once
jj abandon abc def 123
jj abandon abc::
```

### Strategy 2: Generate a New Change ID

Keep both versions as separate changes with different change IDs:

```bash
jj metaedit --update-change-id <commit-id>
```

This preserves both versions of the content while resolving the divergence by giving one commit a new change ID.

### Strategy 3: Squash the Commits Together

Combine the content from both divergent commits into one:

```bash
jj squash --from <source-commit-id> --into <target-commit-id>
```

The source commit is abandoned after squashing. Only one commit with the original change ID remains.

### Strategy 4: Ignore the Divergence

Divergence isn't an error. If it doesn't cause immediate problems, you can leave it as-is. However:

- You cannot refer to divergent changes unambiguously using their change ID alone
- If both commits are part of immutable history, ignoring may be your only option

---

## Divergence vs. Conflicts

| Aspect | Divergence | Conflict |
|--------|-----------|----------|
| What | Two visible commits with same change ID | Conflicting file contents in a single commit |
| Cause | Same change ID on multiple commits | Incompatible edits to the same file |
| Resolution | Abandon, re-ID, squash, or ignore | Edit files to resolve markers |
| Indicator | "divergent" label + change offset | "conflict" label |

---

## Preventing Divergence

- **Avoid amending the same change from multiple workspaces** — Use `jj edit` instead of parallel `jj describe` or `jj squash`
- **Be cautious with background fetches** — IDE integrations that auto-fetch can cause divergence in colocated repos
- **Push promptly** — Pushing your changes reduces the window for concurrent modifications
- **Use `--at-op` carefully** — Running commands at a past operation can create forks in the operation log