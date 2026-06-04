# jj Divergent Changes Reference

A **divergent change** occurs when multiple visible commits share the same change ID. Divergent changes appear in `jj log` with a "divergent" label and a change offset suffix (e.g., `mzvwutvl/0`).

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/guides/divergence/>. For **core concepts** about change IDs, see [concepts.md](concepts.md). For **CLI commands**, see [cli.md](cli.md).


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


## Preventing Divergence

- **Avoid amending the same change from multiple workspaces** — Use `jj edit` instead of parallel `jj describe` or `jj squash`
- **Be cautious with background fetches** — IDE integrations that auto-fetch can cause divergence in colocated repos
- **Push promptly** — Pushing your changes reduces the window for concurrent modifications
- **Use `--at-op` carefully** — Running commands at a past operation can create forks in the operation log