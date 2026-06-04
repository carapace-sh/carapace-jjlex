# jj Conflicts Reference

Unlike most VCSs, jj records conflicted states in commits. Rebase operations never fail due to conflicts — the conflict is recorded and the rebase succeeds. Conflicts can be rebased, merged, or backed out. The stored representation is logical (not textual markers), so rebasing doesn't create nested conflict markers.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/conflicts/>. For **core concepts** about conflicts, see [concepts.md](concepts.md). For **CLI commands** like `jj resolve`, see [cli.md](cli.md). For **conflict marker style configuration**, see [config.md](config.md).


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

