---
name: jj-conflicts
description: >
  Reference for jj (Jujutsu VCS) first-class conflicts — conflict marker styles
  (diff, snapshot, git), long conflict markers, missing terminating newlines,
  resolution workflows, and conflict marker syntax. Triggers on: "jj conflict",
  "jj resolve", "conflict marker", "conflict style", "jj conflict resolution",
  "conflict marker style", "merge conflict", "conflict edge case".
user-invocable: false
---

# jj Conflicts Reference

Unlike most VCSs, jj records conflicted states in commits. Rebase operations never fail due to conflicts — the conflict is recorded and the rebase succeeds. Conflicts can be rebased, merged, or backed out. The stored representation is logical (not textual markers), so rebasing doesn't create nested conflict markers.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/conflicts/>. For **core concepts** about conflicts, see the `jj-concepts` skill. For **CLI commands** like `jj resolve`, see the `jj-cli` skill. For **conflict marker style configuration**, see the `jj-config` skill.

---

## Advantages of First-Class Conflicts

1. **Single workflow** — No need for `git rebase/merge/cherry-pick --continue`; rebase always succeeds
2. **Auto-rebase** — Descendants of rewritten commits automatically get rewritten; conflicts are propagated
3. **Proper merge handling** — Changes in merge commits are compared to auto-merged parents, making "evil merges" safe
4. **Postponable resolution** — Conflicts can be resolved when ready; work-in-progress commits can stay rebased onto upstream
5. **Trivial criss-cross and octopus merges** — Cases Git can't handle or would nest markers are auto-resolved
6. **Collaborative resolution** — Conflicts can be shared (though not recommended with Git backend)

---

## Conflict Marker Styles

Configured via `ui.conflict-marker-style`. Default: `"diff"`.

### Diff Style (default)

Shows a diff from the merge base to one side, plus the other side in full.

```
<<<<<<< conflict 1 of 1
%%%%%%% diff from: vpxusssl 38d49363 "merge base"
\\\\\\\        to: rtsqusxu 2768b0b9 "commit A"
 apple
-grape
+grapefruit
 orange
+++++++ ysrnknol 7a20f389 "commit B"
APPLE
GRAPE
ORANGE
>>>>>>> conflict 1 of 1 ends
```

**Markers:**

| Marker | Purpose |
|--------|---------|
| `<<<<<<<` | Conflict start |
| `>>>>>>>` | Conflict end |
| `+++++++` | Snapshot section start (full content of one side) |
| `%%%%%%%` | Diff section start (diff from base to one side) |
| `\\\\\\\` | Diff direction indicator (which side the diff is to) |

**Resolution:** Apply the diff (the `%%%%%%%`/`\\\\\\\` section) to the snapshot side (the `+++++++` section).

### Snapshot Style

Enabled via `ui.conflict-marker-style = "snapshot"`. Shows full contents of each side and the base, no diffs.

```
<<<<<<< conflict 1 of 1
+++++++ rtsqusxu 2768b0b9 "commit A"
apple
grapefruit
orange
------- vpxusssl 38d49363 "merge base"
apple
grape
orange
+++++++ ysrnknol 7a20f389 "commit B"
APPLE
GRAPE
ORANGE
>>>>>>> conflict 1 of 1 ends
```

Additional marker:
- `-------` — Base/merge-base content separator

### Git Style (diff3)

Enabled via `ui.conflict-marker-style = "git"`. Git-compatible format showing base, left, and right sides.

```
<<<<<<< rtsqusxu 2768b0b9 "commit A"
apple
grapefruit
orange
||||||| vpxusssl 38d49363 "merge base"
apple
grape
orange
=======
APPLE
GRAPE
ORANGE
>>>>>>> ysrnknol 7a20f389 "commit B"
```

Additional marker:
- `|||||||` — Base content separator

**Limitation:** Only supports 2-sided conflicts. For conflicts with more than 2 sides, falls back to "snapshot" style.

---

## Long Conflict Markers

When files may contain lines that could be mistaken for standard conflict markers (e.g., lines starting with `=======`), jj uses longer markers for unambiguous parsing.

```
<<<<<<<<<<<<<<< conflict 1 of 1
%%%%%%%%%%%%%%% diff from: wqvuxsty cb9217d5 "merge base"
\\\\\\\\\\\\\\\        to: kwntsput 0e15b770 "commit A"
-Heading
+HEADING
 =======
+++++++++++++++ mpnwrytz 52020ed6 "commit B"
New Heading
===========
>>>>>>>>>>>>>>> conflict 1 of 1 ends
```

The number of characters in the marker is consistent within a conflict block and long enough to disambiguate from file content.

---

## Conflicts with Missing Terminating Newline

Conflict markers require their own lines. When a conflict occurs in content without a trailing newline, jj adds an extra newline to each term but omits the terminating newline from the `>>>>>>>` marker.

```
<<<<<<< conflict 1 of 1
+++++++ tlwwkqxk d121763d "commit A" (no terminating newline)
grapefruit
%%%%%%% diff from: qwpqssno fe561d93 "merge base" (no terminating newline)
\\\\\\\        to: poxkmrxy c735fe02 "commit B"
 grape
+
>>>>>>> conflict 1 of 1 ends
```

The `(no terminating newline)` annotation appears in the section header. The resolution (`grapefruit\n`) includes the newline that was added during marker serialization.

---

## Conflict Resolution Workflows

### Method 1: New Commit + Squash (recommended)

```bash
# Create a working-copy commit on top of the conflicted commit
jj new <conflicted-commit>

# Resolve the conflict by editing files (replace markers with resolved text)
# Then review the resolution
jj diff

# Move the resolution into the conflicted commit
jj squash
```

**Advantages:** Easy to review the resolution with `jj diff` before squashing.

### Method 2: Edit the Conflicted Commit Directly

```bash
jj edit <conflicted-commit>

# Resolve the conflict by editing files
# The commit is amended automatically
```

**Disadvantage:** Harder to inspect the resolution separately from the rest of the commit's changes.

### Method 3: External Merge Tool

```bash
jj resolve [path]
```

Uses the configured merge editor (`ui.merge-editor`). Requires 2 sides and a base (3-way merge). Can resolve one file at a time or all conflicts at once.

### Method 4: Restore One Side

```bash
# Restore the file to the state from one parent
jj restore --from <parent> <path>
```

Useful when you want to simply accept one side of the conflict.

### Partial Resolution

You don't need to resolve all conflicts at once. You can resolve part of a conflict by updating different parts of the conflict markers, and leave the rest unresolved.

---

## Conflict Marker Configuration

| Setting | Values | Default |
|---------|--------|---------|
| `ui.conflict-marker-style` | `"diff"`, `"snapshot"`, `"git"` | `"diff"` |

Merge tool specific conflict marker style:

```toml
[merge-tools.<name>]
conflict-marker-style = "diff" | "snapshot" | "git"
merge-tool-edits-conflict-markers = true | false
```

---

## Limitations

- No good way to resolve conflicts between directories, files, and symlinks (directory/file/symlink type conflicts)
- `jj restore` can choose one side for type conflicts, but there's no way to see where the involved parts came from
- Git-style markers only support 2-sided conflicts (falls back to snapshot for 3+ sides)
- Git tools will have trouble with revisions containing conflicted files — jj renders these with conflict markers, but Git sees the non-human-readable internal representation
