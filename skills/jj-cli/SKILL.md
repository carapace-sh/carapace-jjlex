---
name: jj-cli
description: >
  Comprehensive reference for the jj (Jujutsu VCS) command-line interface. Covers
  all commands, subcommands, flags, the working-copy model, bookmarks, the
  operation log, configuration, the template language, and git interoperability.
  For revset expression syntax see the jj-revsets skill; for fileset expression
  syntax see the jj-filesets skill. Triggers on: "jj cli", "jj command",
  "jj usage", "jj subcommand", "jj flag", "jj workflow", "jujutsu cli",
  "jj help", "jj how to".
user-invocable: true
---

# jj CLI Reference

Jujutsu (jj) is a version-control system that uses Git as its storage backend but provides a fundamentally different user model. The working copy is automatically committed, there is no staging area, bookmarks replace branches, and every operation is recorded in an undoable operation log.

> **Source of truth**: `jj --help`, `docs/`, <https://jj-vcs.github.io/jj/latest/cli-reference/>. For **revset expression syntax**, see the `jj-revsets` skill. For **fileset expression syntax**, see the `jj-filesets` skill.

---

## Conceptual Model

### Working Copy Is a Commit

Unlike Git, the jj working copy **is** a commit (the `@` revision). Every jj command automatically snapshots the working copy before operating. There is no staging area / index. Files are implicitly tracked by default.

**Three-step command lifecycle:**
1. Snapshot the working copy (recorded as an operation)
2. Perform the operation, creating new commits in memory
3. Update the working copy to match the new state

If step 3 doesn't complete (crash, ^C), the working copy becomes "stale". Fix with `jj workspace update-stale`.

### Change ID vs Commit ID

Every commit has two identifiers:

| ID | Example | Stability |
|----|---------|-----------|
| **Change ID** | `kntqzsqt` | Stable — persists across rewrites (amends, rebases). Uses letters `k`–`z` |
| **Commit ID** | `d7439b06` | Unstable — changes when the commit content changes. Git-compatible hex hash |

Use change IDs to refer to commits in jj commands; they survive history rewrites.

### No Active Branch / No Detached HEAD

jj has no concept of a "current branch". Bookmarks are just named pointers to commits — they don't move automatically on new commits. There is no detached HEAD state; `@` simply points to whichever commit you're editing.

### Conflicts Are Committable

Conflicts are recorded directly in commits with structured conflict markers. Commands never fail due to unresolved conflicts. Resolve them later with `jj resolve`, by editing files, or with `jj squash`.

### Automatic Descendant Rebasing

When a commit is rewritten (rebase, amend, describe, etc.), all its descendants are automatically rebased onto the new version. Bookmarks and the working copy also update automatically.

### Immutable Revisions

jj protects certain revisions from accidental modification. The default immutable set is `::trunk() | tags() | untracked_remote_bookmarks()`. Override with `immutable_heads()` revset alias. Use `--ignore-immutable` to bypass.

### Virtual Root Commit

Every jj repo has a single root commit (ID `00000000...`) that all commits descend from. There is no Git-style "unborn branch" state.

---

## Global Flags

These flags apply to all jj commands:

| Flag | Description |
|------|-------------|
| `-R`, `--repository <PATH>` | Path to repository to operate on |
| `--ignore-working-copy` | Don't snapshot or update the working copy |
| `--ignore-immutable` | Allow rewriting immutable commits |
| `--at-operation <OP>` / `--at-op` | Load repo at a specific operation (read-only use recommended) |
| `--no-integrate-operation` | Run command but don't integrate any operations |
| `--debug` | Enable debug logging |
| `--color <WHEN>` | When to colorize: `always`, `never`, `debug`, `auto` |
| `--quiet` | Silence non-primary output |
| `--no-pager` | Disable the pager |
| `--config <NAME=VALUE>` | Additional config option (repeatable) |
| `--config-file <PATH>` | Additional config file (repeatable) |

---

## Command Reference

Commands are grouped by function. Argument types: **REVSET** = revision selection expression (see `jj-revsets` skill), **FILESET** = file selection expression (see `jj-filesets` skill).

### Repository Setup

| Command | Description |
|---------|-------------|
| `jj git init [--no-colocate]` | Create a new Git-backed repo. Default: colocated with `.git` |
| `jj git clone <URL> [<DEST>] [--remote <NAME>]` | Clone a Git repo |
| `jj git init --git-repo=<PATH>` | Create jj repo from existing Git repo |

### Status and Information

| Command | Description |
|---------|-------------|
| `jj status` / `jj st` | Show high-level repo status |
| `jj log [FILESETS]...` | Show revision history. Default revset: `revsets.log` config |
| `jj show [REVSET]` | Show commit description and changes. Default: `@` |
| `jj diff [FILESETS]...` | Compare file contents between revisions. Default: `@` vs parent |
| `jj root` | Show current workspace root directory |
| `jj version` | Display version information |

### Revision Operations

| Command | Description |
|---------|-------------|
| `jj new [REVSETS]...` | Create a new empty change. Default parent: `@` |
| `jj commit [FILESETS]...` | Describe working copy and create new change on top |
| `jj describe [REVSETS]...` | Update commit message. Default: `@` |
| `jj edit <REVSET>` | Set a revision as the working-copy revision |
| `jj duplicate [REVSETS]...` | Create new changes with the same content as existing ones |
| `jj abandon [REVSETS]...` | Abandon a revision (descendants rebased onto parents) |
| `jj squash [FILESETS]...` | Move changes from one revision into another |
| `jj split [FILESETS]...` | Split a revision in two |
| `jj rebase` | Move revisions to different parent(s) |
| `jj absorb [FILESETS]...` | Move changes into the stack of mutable revisions |
| `jj parallelize [REVSETS]...` | Make revisions siblings instead of ancestors/descendants |
| `jj simplify-parents` | Simplify parent edges for revisions |
| `jj revert -r <REVSETS>` | Apply the reverse of given revision(s) |
| `jj restore [FILESETS]...` | Restore paths from another revision |
| `jj metaedit [REVSETS]...` | Modify revision metadata without changing content |
| `jj sign` | Cryptographically sign a revision |
| `jj unsign` | Drop a cryptographic signature |

### Navigation

| Command | Description |
|---------|-------------|
| `jj next` | Move working-copy commit to the child revision |
| `jj prev` | Move working-copy commit to the parent revision |

### Diff Editing

| Command | Description |
|---------|-------------|
| `jj diffedit [FILESETS]...` | Touch up content changes with a diff editor |
| `jj interdiff [FILESETS]...` | Show differences between the diffs of two revisions |
| `jj fix [FILESETS]...` | Update files with formatting fixes or other automated changes |

### Conflict Resolution

| Command | Description |
|---------|-------------|
| `jj resolve [FILESETS]...` | Resolve conflicted files with an external merge tool |

### Bookmarks (`jj bookmark` / `jj b`)

| Command | Description |
|---------|-------------|
| `jj bookmark list` / `jj b l` | List bookmarks and their targets |
| `jj bookmark create <NAME>` / `jj b c` | Create a new bookmark |
| `jj bookmark move <NAME>` / `jj b m` | Move bookmark to target revision |
| `jj bookmark delete <NAME>` / `jj b d` | Delete bookmark (propagates to remotes on push) |
| `jj bookmark forget <NAME>` / `jj b f` | Forget bookmark without marking as deletion |
| `jj bookmark rename <OLD> <NEW>` / `jj b r` | Rename a bookmark |
| `jj bookmark set <NAME>` / `jj b s` | Create or update a bookmark by name |
| `jj bookmark advance` / `jj b a` | Advance closest bookmarks to a target revision |
| `jj bookmark track <NAME>` / `jj b t` | Start tracking a remote bookmark |
| `jj bookmark untrack` | Stop tracking a remote bookmark |

### Tags (`jj tag`)

| Command | Description |
|---------|-------------|
| `jj tag list` / `jj tag l` | List tags and their targets |
| `jj tag set <NAME>` / `jj tag s` | Create or update a tag |
| `jj tag delete <NAME>` / `jj tag d` | Delete tags |

### File Operations (`jj file`)

| Command | Description |
|---------|-------------|
| `jj file list [FILESETS]...` | List files in a revision |
| `jj file show <PATH>` | Print file contents in a revision |
| `jj file annotate <PATH>` | Show source change for each line (like `git blame`) |
| `jj file chmod <MODE> <PATH>` | Set or remove executable bit |
| `jj file search [FILESETS]...` | Search for content in files |
| `jj file track <PATH>` | Start tracking a path in the working copy |
| `jj file untrack <PATH>` | Stop tracking a path in the working copy |

### Git Integration (`jj git`)

| Command | Description |
|---------|-------------|
| `jj git fetch [--remote <NAME>]` | Fetch from a Git remote |
| `jj git push [--all] [--bookmark <NAME>]` | Push to a Git remote |
| `jj git export` | Update underlying Git repo with jj changes |
| `jj git import` | Update jj with changes from underlying Git repo |
| `jj git remote add/remove/list/rename/set-url` | Manage Git remotes |
| `jj git root` | Show the underlying Git directory |
| `jj git colocation enable/disable/status` | Manage colocation |

### Operation Log (`jj operation` / `jj op`)

| Command | Description |
|---------|-------------|
| `jj op log` | Show the operation log |
| `jj op show <OP>` | Show changes in an operation |
| `jj op diff` | Compare changes between two operations |
| `jj op restore <OP>` | Restore repo to earlier state |
| `jj op revert <OP>` | Revert a specific earlier operation |
| `jj op abandon` | Abandon operation history |
| `jj op integrate` | Make an operation part of the operation log |
| `jj undo` | Undo the last operation |
| `jj redo` | Redo the most recently undone operation |

### Workspaces (`jj workspace`)

| Command | Description |
|---------|-------------|
| `jj workspace add <NAME>` | Add a workspace |
| `jj workspace list` | List workspaces |
| `jj workspace forget <NAME>` | Stop tracking a workspace |
| `jj workspace rename <NAME>` | Rename the current workspace |
| `jj workspace root` | Show workspace root directory |
| `jj workspace update-stale` | Update a stale workspace |

### Configuration (`jj config`)

| Command | Description |
|---------|-------------|
| `jj config edit` / `jj config e` | Edit config file in editor |
| `jj config get <NAME>` | Get config value |
| `jj config list` / `jj config l` | List config variables and values |
| `jj config path` / `jj config p` | Print paths to config files |
| `jj config set <NAME> <VALUE>` | Set a config option |
| `jj config unset <NAME>` | Unset a config option |

### Other Commands

| Command | Description |
|---------|-------------|
| `jj sparse list/set/edit/reset` | Manage sparse working copy patterns |
| `jj bisect run --range <REVSETS> <CMD>` | Find bad revision by bisection |
| `jj gerrit upload` | Upload changes to Gerrit for code review |
| `jj util completion <SHELL>` | Print shell completion script |
| `jj util config-schema` | Print JSON schema for jj TOML config |
| `jj util exec <CMD>` | Execute an external command via jj |
| `jj util gc` | Run backend garbage collection |
| `jj util install-man-pages <PATH>` | Install man pages |
| `jj util markdown-help` | Print CLI help in Markdown |
| `jj util snapshot` | Snapshot the working copy if needed |
| `jj arrange [REVSETS]` | Interactively arrange the commit graph |
| `jj evolog` | Show how a change has evolved over time |

---

## Detailed Command Flags

### `jj log`

```
jj log [OPTIONS] [FILESETS]...
```

| Flag | Description |
|------|-------------|
| `-r`, `--revision <REVSETS>` | Which revisions to show. Default: `revsets.log` config |
| `-n`, `--limit <N>` | Limit number of revisions shown |
| `--reversed` | Show older revisions first |
| `-G`, `--no-graph` | Flat list (no graph) |
| `-T`, `--template <TMPL>` | Custom output template |
| `-p`, `--patch` | Show patch for each revision |
| `--count` | Print number of commits |
| `-s`, `--summary` | Show modified/added/deleted per path |
| `--stat` | Show histogram of changes |
| `--types` | Show type before and after per path |
| `--name-only` | Show only path names |
| `--git` | Git-format diff |
| `--color-words` | Word-level diff |
| `--tool <TOOL>` | Use external diff tool |
| `--context <N>` | Lines of context |

### `jj diff`

```
jj diff [OPTIONS] [FILESETS]...
```

| Flag | Description |
|------|-------------|
| `-r`, `--revisions <REVSETS>` | Show changes in these revisions. Default: `@` |
| `-f`, `--from <REVSET>` | Show changes from this revision |
| `-t`, `--to <REVSET>` | Show changes to this revision |
| `-T`, `--template <TMPL>` | Custom output template |
| `-s`, `--summary` | Show modified/added/deleted per path |
| `--stat` | Show histogram of changes |
| `--types` | Show type before and after |
| `--name-only` | Show only path names |
| `--git` | Git-format diff |
| `--color-words` | Word-level diff |
| `--tool <TOOL>` | Use external diff tool |
| `--context <N>` | Lines of context |
| `-w`, `--ignore-all-space` | Ignore whitespace |
| `-b`, `--ignore-space-change` | Ignore whitespace changes |

### `jj show`

```
jj show [OPTIONS] [REVSET]
```

| Flag | Description |
|------|-------------|
| `<REVSET>` | Revision to show. Default: `@` |
| `-T`, `--template <TMPL>` | Custom output template |
| `-s`, `--summary` / `--stat` / `--types` / `--name-only` | Diff format options |
| `--git` / `--color-words` | Diff format options |
| `--tool <TOOL>` | Use external diff tool |
| `--context <N>` | Lines of context |
| `--no-patch` | Don't show the patch |
| `-w` / `-b` | Ignore whitespace options |

### `jj rebase`

```
jj rebase [OPTIONS] <--onto|--insert-after|--insert-before>
```

| Flag | Description |
|------|-------------|
| `-b`, `--branch <REVSETS>` | Rebase whole branch relative to destination's ancestors |
| `-s`, `--source <REVSETS>` | Rebase specified revision(s) with descendants |
| `-r`, `--revision <REVSETS>` | Rebase given revisions, rebasing descendants onto parent(s) |
| `-o`, `--onto <REVSETS>` | Revision(s) to rebase onto (repeatable for merge) |
| `-A`, `--insert-after <REVSETS>` | Revision(s) to insert after |
| `-B`, `--insert-before <REVSETS>` | Revision(s) to insert before |
| `--skip-emptied` | Abandon commits that become empty after rebase |
| `--keep-divergent` | Keep divergent commits while rebasing |
| `--simplify-parents` | Simplify parents of rebased commits |

### `jj squash`

```
jj squash [OPTIONS] [FILESETS]...
```

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Move only changes to these paths |
| `-r`, `--revision <REVSET>` | Revision to squash into its parent. Default: `@` |
| `-f`, `--from <REVSETS>` | Revision(s) to squash from. Default: `@` |
| `-t`, `--into <REVSET>` | Revision to squash into. Default: `@` |
| `-o`, `--onto <REVSETS>` | (Experimental) Parent for the new commit |
| `-A`, `--insert-after <REVSETS>` | (Experimental) Insert after |
| `-B`, `--insert-before <REVSETS>` | (Experimental) Insert before |
| `-m`, `--message <MSG>` | Description for squashed revision |
| `-u`, `--use-destination-message` | Use destination's description |
| `--editor` | Open editor to edit description |
| `-i`, `--interactive` | Interactively choose which parts |
| `--tool <NAME>` | Specify diff editor (implies --interactive) |
| `-k`, `--keep-emptied` | Don't abandon source revision |

### `jj new`

```
jj new [OPTIONS] [REVSETS]...
```

| Flag | Description |
|------|-------------|
| `<REVSETS>` | Parent(s) of the new change. Default: `@` |
| `-m`, `--message <MSG>` | Change description |
| `--no-edit` | Don't edit the newly created change |
| `-A`, `--insert-after <REVSETS>` | Insert after given commit(s) |
| `-B`, `--insert-before <REVSETS>` | Insert before given commit(s) |

### `jj commit`

```
jj commit [OPTIONS] [FILESETS]...
```

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Put these paths in the current commit |
| `-i`, `--interactive` | Interactively choose which changes |
| `--tool <NAME>` | Specify diff editor (implies --interactive) |
| `-m`, `--message <MSG>` | Change description (don't open editor) |
| `--editor` | Open editor to edit description |

### `jj describe`

```
jj describe [OPTIONS] [REVSETS]...
```

| Flag | Description |
|------|-------------|
| `<REVSETS>` | Revision(s) to describe. Default: `@` |
| `-m`, `--message <MSG>` | Change description (don't open editor) |
| `--stdin` | Read description from stdin |
| `--editor` | Open editor to edit description |

### `jj split`

```
jj split [OPTIONS] [FILESETS]...
```

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Files for the selected changes |
| `-r`, `--revision <REVSET>` | The revision to split. Default: `@` |
| `-o`, `--onto <REVSETS>` | Revision(s) to rebase selected changes onto |
| `-A`, `--insert-after <REVSETS>` | Insert after |
| `-B`, `--insert-before <REVSETS>` | Insert before |
| `-m`, `--message <MSG>` | Description for selected changes |
| `--editor` | Open editor to edit descriptions |
| `-p`, `--parallel` | Split into parallel revisions instead of parent/child |
| `-i`, `--interactive` | Interactively choose which parts |
| `--tool <NAME>` | Specify diff editor |

### `jj abandon`

```
jj abandon [OPTIONS] [REVSETS]...
```

| Flag | Description |
|------|-------------|
| `<REVSETS>` | Revision(s) to abandon. Default: `@` |
| `--retain-bookmarks` | Don't delete bookmarks pointing to abandoned revisions |
| `--restore-descendants` | Don't modify content of children of abandoned commits |

### `jj duplicate`

```
jj duplicate [OPTIONS] [REVSETS]...
```

| Flag | Description |
|------|-------------|
| `<REVSETS>` | Revision(s) to duplicate. Default: `@` |
| `-o`, `--onto <REVSETS>` | Revision(s) to duplicate onto |
| `-A`, `--insert-after <REVSETS>` | Insert after |
| `-B`, `--insert-before <REVSETS>` | Insert before |

### `jj restore`

```
jj restore [OPTIONS] [FILESETS]...
```

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Restore only these paths |
| `-f`, `--from <REVSET>` | Revision to restore from (source) |
| `-t`, `--into <REVSET>` | Revision to restore into (destination) |
| `-c`, `--changes-in <REVSET>` | Undo changes in a revision vs merge of parents |
| `-i`, `--interactive` | Interactively choose parts |
| `--tool <NAME>` | Specify diff editor |
| `--restore-descendants` | Preserve content when rebasing descendants |

### `jj bookmark` Subcommands

#### `jj bookmark list` / `jj b l`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Filter by name patterns |
| `-a`, `--all-remotes` | Show all tracked and untracked remote bookmarks |
| `--remote <REMOTE>` | Filter by remote |
| `-t`, `--tracked` | Show tracked remote bookmarks only |
| `-c`, `--conflicted` | Show conflicted bookmarks only |
| `-r`, `--revision <REVSETS>` | Show bookmarks pointing to these revisions |
| `-T`, `--template <TMPL>` | Custom output template |
| `--sort <KEY>` | Sort by: `name`, `author-date`, `committer-date`, etc. (suffix `-` for reverse) |

#### `jj bookmark create` / `jj b c`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Bookmark names to create |
| `-r`, `--revision <REVSET>` | Target revision. Default: `@` |

#### `jj bookmark move` / `jj b m`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Bookmark name patterns to move |
| `-f`, `--from <REVSETS>` | Move bookmarks from these revisions |
| `-t`, `--to <REVSET>` | Move bookmarks to this revision. Default: `@` |
| `-B`, `--allow-backwards` | Allow moving backwards or sideways |

#### `jj bookmark set` / `jj b s`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Bookmark names to create or update |
| `-r`, `--revision <REVSET>` | Target revision. Default: `@` |
| `-B`, `--allow-backwards` | Allow moving backwards or sideways |

#### `jj bookmark delete` / `jj b d`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Bookmark names to delete |

#### `jj bookmark forget` / `jj b f`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Bookmark names to forget |
| `--include-remotes` | Also forget corresponding remote bookmarks |

#### `jj bookmark rename` / `jj b r`

| Flag | Description |
|------|-------------|
| `<OLD>` | Current bookmark name |
| `<NEW>` | New bookmark name |
| `--overwrite-existing` | Allow overwriting an existing bookmark |

#### `jj bookmark advance` / `jj b a`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Bookmark name patterns |
| `-t`, `--to <REVSET>` | Target revision. Default: `revsets.bookmark-advance-to` |

#### `jj bookmark track` / `jj b t`

| Flag | Description |
|------|-------------|
| `<BOOKMARK>` | Bookmark names to track |
| `--remote <REMOTE>` | Remote names to track |

### `jj config` Subcommands

All `jj config edit/set/unset` commands require one of `--user`, `--repo`, or `--workspace` to specify the config level.

| Subcommand | Description |
|------------|-------------|
| `jj config edit --user/--repo/--workspace` | Open config file in editor |
| `jj config get <NAME>` | Print config value |
| `jj config list [NAME]` | List config variables |
| `jj config path --user/--repo/--workspace` | Print config file path |
| `jj config set --user/--repo/--workspace <NAME> <VALUE>` | Set config value |
| `jj config unset --user/--repo/--workspace <NAME>` | Unset config value |

### `jj git push`

| Flag | Description |
|------|-------------|
| `--remote <REMOTE>` | Remote to push to |
| `-b`, `--bookmark <NAME>` | Push this bookmark (repeatable) |
| `--all` | Push all bookmarks (including new) |
| `--tracked` | Push all tracked bookmarks |
| `--deleted` | Push all deleted bookmarks |
| `-r`, `--revision <REVSETS>` | Push bookmarks pointing to these commits |
| `-c`, `--change <REVSETS>` | Push by creating a bookmark for each commit |
| `--named <NAME=REVISION>` | Specify a new bookmark name and revision |
| `--dry-run` | Only display what will change on remote |
| `--allow-empty-description` | Allow pushing commits with empty descriptions |
| `--allow-private` | Allow pushing commits that are private |
| `-o`, `--option <OPTION>` | Git push options |

### `jj git fetch`

| Flag | Description |
|------|-------------|
| `--remote <REMOTE>` | Remote to fetch from |
| `-b`, `--branch <BRANCH>` | Branch to fetch (repeatable) |
| `--tracked` | Fetch only tracked bookmarks |
| `--all-remotes` | Fetch from all remotes |

### `jj operation` Subcommands

#### `jj op log`

| Flag | Description |
|------|-------------|
| `-n`, `--limit <N>` | Limit number of operations |
| `--reversed` | Older operations first |
| `-G`, `--no-graph` | Flat list |
| `-T`, `--template <TMPL>` | Custom template |
| `-d`, `--op-diff` | Show repo changes at each operation |
| `-p`, `--patch` | Show patch (implies --op-diff) |
| Diff format flags | `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words`, `--tool`, `--context` |

#### `jj op show [<OP>]`

| Flag | Description |
|------|-------------|
| `<OP>` | Operation to show. Default: `@` |
| `-G`, `--no-graph` | Flat list |
| `-T`, `--template <TMPL>` | Custom template |
| `--no-op-diff` | Don't show operation diff |
| Diff format flags | `-p`, `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words`, `--tool`, `--context` |

### `jj evolog`

| Flag | Description |
|------|-------------|
| `-r`, `--revisions <REVSETS>` | Which revisions to follow. Default: `@` |
| `-n`, `--limit <N>` | Limit number of revisions |
| `--reversed` | Older first |
| `-G`, `--no-graph` | Flat list |
| `-T`, `--template <TMPL>` | Custom template |
| Diff format flags | `-p`, `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words`, `--tool`, `--context` |
---

## Bookmarks

Bookmarks are named pointers to revisions, similar to Git branches but with key differences:

- **No active/current bookmark** — there is no concept of "checking out" a bookmark
- **Don't auto-move on new commits** — bookmarks stay where you put them
- **Auto-follow on rewrite** — when a commit is rewritten (rebase, amend), bookmarks pointing to it automatically update
- **Remote bookmarks** — format `<name>@<remote>` (e.g., `main@origin`), store the last-seen position on the remote

### Tracking

Tracked remote bookmarks sync with local bookmarks of the same name. On `jj git fetch`, if a tracked remote bookmark moved, the local bookmark also moves. If both moved, the bookmark becomes **conflicted** (shown as `main??`).

```bash
jj bookmark track main@origin   # Start tracking
jj bookmark untrack main@origin # Stop tracking
```

Set `remotes.<name>.auto-track-bookmarks = "*"` in config to auto-track all fetched bookmarks.

### Bookmark Conflicts

Conflicted bookmarks show `??` suffix in `jj log`. Resolve with `jj bookmark move <name> --to <target>`.

### Safety on Push

Push checks: (1) remote position matches jj's record, (2) local bookmark is not conflicted, (3) remote bookmark must be tracked if it exists on remote.

---

## Operation Log

Every jj command that modifies the repo creates an **operation** in the operation log. Each operation stores: a view snapshot (bookmark positions, heads, working-copy commits), pointers to parent operations, and metadata (timestamp, username, description).

```bash
jj op log          # View operation history
jj undo            # Undo last operation
jj redo            # Redo last undone operation
jj op restore <op> # Restore repo to earlier state
jj op revert <op>  # Revert a specific earlier operation
jj op show <op>    # Show changes in an operation
jj op diff         # Compare changes between two operations
```

The `--at-op` global flag loads the repo at a specific operation. This is read-only — the `@` symbol resolves to the working-copy commit recorded in that operation's view. Automatic working-copy snapshotting is disabled with `--at-op`.

---

## Configuration

jj uses TOML config files with layered precedence (later overrides earlier):

1. **Built-in settings** — cannot be edited
2. **User settings** — `jj config edit --user` (e.g., `~/.jjconfig.toml` or `$XDG_CONFIG_HOME/jj/config.toml`)
3. **Repo settings** — `jj config edit --repo`
4. **Workspace settings** — `jj config edit --workspace`
5. **Command-line** — `--config <NAME=VALUE>` or `--config-file <PATH>`

### Key Config Sections

| Section | Purpose | Example Keys |
|---------|---------|--------------|
| `[user]` | Identity | `name`, `email` |
| `[ui]` | UI settings | `default-command`, `editor`, `diff-formatter`, `pager`, `color`, `graph.style` |
| `[revsets]` | Default revsets | `log`, `short-prefixes`, `bookmark-advance-from`, `bookmark-advance-to` |
| `[revset-aliases]` | Custom revset aliases | `immutable_heads()`, `trunk()`, user-defined |
| `[fileset-aliases]` | Custom fileset aliases | User-defined |
| `[templates]` | Template customization | `log`, `config_list`, `draft_commit_description` |
| `[template-aliases]` | Template aliases | User-defined |
| `[aliases]` | Command aliases | `l = ["log", "-r", "main@origin..@"]` |
| `[colors]` | Color/style settings | `commit_id = "green"`, `"working_copy commit_id" = { underline = true }` |
| `[git]` | Git integration | `push = "origin"`, `fetch = "origin"`, `colocate`, `private-commits`, `sign-on-push` |
| `[merge-tools.<name>]` | Diff/merge tool | `program`, `edit-args`, `merge-args` |
| `[fix.tools.<name>]` | Code formatters | `command`, `patterns`, `enabled` |
| `[signing]` | Commit signing | `behavior`, `backend`, `key` |
| `[snapshot]` | Snapshot behavior | `auto-track`, `max-new-file-size` |
| `[remotes.<name>]` | Remote config | `fetch-bookmarks`, `fetch-tags`, `auto-track-bookmarks` |

### Conditional Config

```toml
[[--scope]]
--when.repositories = ["~/oss"]
[--scope.user]
email = "oss@example.org"

[[--scope]]
--when.commands = ["diff", "show"]
[--scope.ui]
pager = "delta"
```

Available conditions: `--when.repositories`, `--when.workspaces`, `--when.hostnames`, `--when.commands`, `--when.platforms`, `--when.environments`.

---

## Template Language

The jj template language is a functional language for customizing command output (used with `-T`/`--template` flags). Most display commands accept templates.

### Operators (strongest to weakest)

| Priority | Operator | Meaning |
|----------|----------|---------|
| 1 | `x.f()` | Method call |
| 2 | `f(x)` | Function call |
| 3 | `-x` | Negate integer |
| 4 | `!x` | Logical not |
| 5 | `p:x` | String pattern |
| 6 | `x * y`, `x / y`, `x % y` | Multiplication/division/remainder |
| 7 | `x + y`, `x - y` | Addition/subtraction |
| 8 | `x >= y`, `x > y`, `x <= y`, `x < y` | Comparison |
| 9 | `x == y`, `x != y` | Equality |
| 10 | `x && y` | Logical and |
| 11 | `x \|\| y` | Logical or |
| 12 | `x ++ y` | Concatenation |

### Key Types and Methods

**Commit**: `.commit_id()`, `.change_id()`, `.description()`, `.parents()`, `.author()`, `.committer()`, `.trailers()`, `.bookmarks()`, `.local_bookmarks()`, `.remote_bookmarks()`, `.tags()`, `.mine()`, `.hidden()`, `.divergent()`, `.conflict()`, `.empty()`, `.root()`, `.diff([files])`, `.files([files])`, `.conflicted_files()`, `.signature()`

**ChangeId / CommitId**: `.short([len])`, `.shortest([min_len])`, `.normal_hex()` (ChangeId only)

**String**: `.len()`, `.contains(needle)`, `.starts_with()`, `.ends_with()`, `.remove_prefix()`, `.remove_suffix()`, `.trim()`, `.upper()`, `.lower()`, `.substr(start, [end])`, `.first_line()`, `.lines()`, `.split(pattern, [limit])`, `.replace(pattern, replacement, [limit])`, `.escape_json()`

**List**: `.len()`, `.join(sep)`, `.filter(\|x\| expr)`, `.map(\|x\| expr)`, `.any(\|x\| expr)`, `.all(\|x\| expr)`, `.first()`, `.last()`, `.get(index)`, `.reverse()`, `.skip(n)`, `.take(n)`

**Timestamp**: `.ago()`, `.format(fmt)`, `.utc()`, `.local()`, `.after(date)`, `.before(date)`

**Signature**: `.name()`, `.email()`, `.timestamp()`

**TreeDiff**: `.files()`, `.color_words([context])`, `.git([context])`, `.stat([width])`, `.summary()`

### Global Functions

`fill(width, content)`, `indent(prefix, content)`, `pad_start(width, content, [fill])`, `pad_end(width, content, [fill])`, `pad_centered(width, content, [fill])`, `truncate_start(width, content, [ellipsis])`, `truncate_end(width, content, [ellipsis])`, `label(label, content)`, `if(cond, then, [else])`, `coalesce(content...)`, `concat(content...)`, `join(sep, content...)`, `separate(sep, content...)`, `surround(prefix, suffix, content)`, `stringify(content)`, `json(value)`, `config(name)`, `hyperlink(url, text, [fallback])`, `replace(pattern, content, replacement)`

### Template Examples

```bash
# Short commit IDs of working-copy parents
jj log -G -r @ -T 'parents.map(|c| c.commit_id().short()).join(",")'

# Machine-readable full IDs
jj log -G -T 'commit_id ++ " " ++ change_id ++ "\n"'

# Custom log format with diff stats
jj log -T 'change_id.shortest() ++ " " ++ label("diff", if(empty, "", diff.stat().total_added())) ++ "\n"'

# Discover color labels
jj log --color=debug
```

---

## Git Comparison

### Key Conceptual Differences

| Aspect | Git | jj |
|--------|-----|-----|
| Working copy | Must manually commit | Automatically committed as a regular commit |
| Staging area | Has index for partial commits | No index — use `jj split`, `jj squash -i` instead |
| Branches | Active branch, moves on commit | Bookmarks: no active concept, don't auto-move |
| Detached HEAD | Special state to avoid | Normal — `@` points to whatever you're editing |
| Conflicts | Commands fail on conflicts | Conflicts are committable, resolved later |
| Descendant rebasing | Must manually rebase | Automatic when parent is rewritten |
| Operation log | Per-ref reflogs | Global operation log tracking all changes |
| Root commit | "Unborn branch" state | Virtual root commit (`00000000...`) |

### Command Equivalents

| Use Case | Git | jj |
|----------|-----|-----|
| Create repo | `git init` | `jj git init [--no-colocate]` |
| Clone | `git clone <url>` | `jj git clone <url>` |
| Status | `git status` | `jj st` |
| Diff | `git diff HEAD` | `jj diff` |
| Diff specific revision | `git diff <rev>^ <rev>` | `jj diff -r <rev>` |
| Diff from A to B | `git diff A B` | `jj diff --from A --to B` |
| Commit | `git commit -a` | `jj commit` |
| Amend | `git commit --amend -a` | `jj squash` |
| Interactive amend | `git add -p; git commit --amend` | `jj squash -i` |
| Split | `git commit -p` | `jj split` |
| Edit description | `git commit --amend -m "msg"` | `jj describe` |
| Stash | `git stash` | `jj new @-` |
| Switch branch | `git switch -c topic main` | `jj new main` |
| Log ancestors | `git log --oneline --graph` | `jj log -r ::@` |
| Log all | `git log --oneline --graph --all` | `jj log -r 'all()'` |
| Rebase branch | `git rebase B A` | `jj rebase -b A -o B` |
| Undo | `git reset` (limited) | `jj undo` |
| Branch list | `git branch` | `jj bookmark list` / `jj b l` |
| Create branch | `git branch <name> <rev>` | `jj bookmark create <name> -r <rev>` |
| Move branch | `git branch -f <name> <rev>` | `jj bookmark move <name> --to <rev>` |
| Delete branch | `git branch -d <name>` | `jj bookmark delete <name>` |
| Push all | `git push --all` | `jj git push --all` |
| Push single | `git push origin <name>` | `jj git push --bookmark <name>` |
| Fetch | `git fetch` | `jj git fetch` |
| Blame | `git blame <file>` | `jj file annotate <file>` |
| Restore file | `git restore <paths>` | `jj restore <paths>` |
| Revert commit | `git revert <rev>` | `jj revert -r <rev>` |
| Show commit | `git show <rev>` | `jj show <rev>` |
| Search in diffs | `git log -G pattern` | `jj log -r 'diff_lines(regex:pattern)'` |
| Search in files | `git grep pattern` | `jj file search --pattern pattern` |

---

## Common Workflows

### Starting a New Project

```bash
jj git init                    # Create new repo (colocated with .git by default)
jj git clone <url>             # Or clone existing repo
```

### Day-to-Day Development

```bash
# Edit files directly — no staging needed
# Files are auto-tracked on next jj command

jj status                      # Check current state
jj diff                        # See current changes
jj describe                    # Set/edit commit message
jj new                         # Create new change on top, make @ empty
jj commit                      # Describe working copy and create new change on top
jj squash                      # Move working-copy changes into parent
jj squash -i                   # Interactively select changes to move into parent
jj squash --from <rev> --into <rev>  # Move changes between arbitrary revisions
```

### Navigating History

```bash
jj log                         # Show default log (ancestors of visible heads)
jj log -r ::@                  # Ancestors of working copy
jj log -r 'all()'              # All visible commits
jj log -r 'main..@'            # Commits between main and working copy
jj log -r 'author(mine())'     # My commits
jj log -r 'diff_lines(regex:TODO)'  # Commits with TODO in diff
jj show <rev>                  # Show a specific commit
jj next                        # Move to next child
jj prev                        # Move to parent
jj edit <rev>                  # Edit an existing commit directly
```

### Bookmarks and Remotes

```bash
jj bookmark create feature     # Create bookmark at @
jj bookmark move feature --to <rev>  # Move bookmark
jj bookmark track main@origin  # Track a remote bookmark
jj git fetch                   # Fetch from remote
jj git push --bookmark feature # Push a bookmark
jj git push --all              # Push all bookmarks
jj git push -c <rev>           # Push by creating a bookmark for a commit
```

### Undoing Mistakes

```bash
jj undo                        # Undo last operation
jj op log                      # View operation history
jj op show <op>                # Show changes in an operation
jj op restore <op>             # Restore entire repo to earlier state
jj op revert <op>              # Revert a specific operation
```

### Conflict Resolution

```bash
# Conflicts are recorded in commits — commands don't fail
# Create a new commit on top of a conflicted one
jj new <conflicted-commit>
# Edit conflict markers in files, then:
jj squash                      # Move resolution into the conflicted commit
# Or use an external merge tool:
jj resolve --tool meld
```

### History Restructuring

```bash
jj rebase -s <rev> -o <target>      # Rebase source and descendants onto target
jj rebase -b <rev> -o <target>      # Rebase whole branch onto target
jj rebase -r <rev> --before <target>  # Insert revision before target
jj split -r <rev>                     # Split revision into two
jj absorb                              # Move changes into appropriate ancestors
jj parallelize <revs>                  # Make revisions siblings
jj simplify-parents -r <rev>           # Remove unnecessary merge edges
jj abandon <rev>                       # Abandon revision, rebase descendants
jj duplicate <rev>                     # Duplicate a commit
jj revert -r <rev>                     # Create reverse commit
```

### Working with Git

```bash
jj git init --git-repo=<path>   # Create jj repo from existing Git repo
jj git export                   # Push jj changes to underlying Git repo
jj git import                   # Pull Git changes into jj
jj git colocation status       # Check if colocated
jj git colocation enable       # Enable colocation
```

### Sparse Working Copy

```bash
jj sparse list                  # List current sparse patterns
jj sparse set --add 'glob:*.md'  # Add patterns
jj sparse set --remove 'src/'    # Remove patterns
jj sparse reset                  # Reset to include all files
jj sparse edit                   # Edit patterns in editor
```

---

## Arguments Accepting Revsets

Many jj commands accept **REVSET** arguments (see the `jj-revsets` skill for full syntax). Commands that accept revsets:

| Command | Revset Argument(s) |
|---------|-------------------|
| `jj abandon` | `<REVSETS>` |
| `jj arrange` | `<REVSETS>` |
| `jj bisect run` | `--range <REVSETS>` |
| `jj bookmark list` | `--revision <REVSETS>` |
| `jj bookmark move` | `--from <REVSETS>`, `--to <REVSET>` |
| `jj bookmark set` | `--revision <REVSET>` |
| `jj describe` | `<REVSETS>` |
| `jj diff` | `-r <REVSETS>`, `--from <REVSET>`, `--to <REVSET>` |
| `jj diffedit` | `-r <REVSET>`, `--from <REVSET>`, `--to <REVSET>` |
| `jj duplicate` | `<REVSETS>`, `--onto`, `--insert-after`, `--insert-before` |
| `jj edit` | `<REVSET>` |
| `jj evolog` | `--revisions <REVSETS>` |
| `jj file annotate/list/search/show` | `--revision <REVSET>` |
| `jj fix` | `--source <REVSETS>` |
| `jj interdiff` | `--from <REVSET>`, `--to <REVSET>` |
| `jj log` | `-r <REVSETS>` |
| `jj metaedit` | `<REVSETS>` |
| `jj new` | `<REVSETS>`, `--insert-after`, `--insert-before` |
| `jj parallelize` | `<REVSETS>` |
| `jj rebase` | `--branch`, `--source`, `--revision`, `--onto`, `--insert-after`, `--insert-before` |
| `jj restore` | `--from <REVSET>`, `--into <REVSET>`, `--changes-in <REVSET>` |
| `jj revert` | `--revision <REVSETS>`, `--onto`, `--insert-after`, `--insert-before` |
| `jj show` | `<REVSET>` |
| `jj sign/unsign` | `--revision <REVSETS>` |
| `jj simplify-parents` | `--source <REVSETS>`, `--revision <REVSETS>` |
| `jj split` | `--revision <REVSET>`, `--onto`, `--insert-after`, `--insert-before` |
| `jj squash` | `--revision <REVSET>`, `--from <REVSETS>`, `--into <REVSET>`, `--onto`, `--insert-after`, `--insert-before` |
| `jj tag list/set` | `--revision <REVSETS>` |
| `jj workspace add` | `--revision <REVSETS>` |

---

## Arguments Accepting Filesets

Commands that accept **FILESET** arguments (see the `jj-filesets` skill for full syntax):

| Command | Fileset Argument(s) |
|---------|---------------------|
| `jj absorb` | `<FILESETS>` |
| `jj commit` | `<FILESETS>` |
| `jj diff` | `<FILESETS>` |
| `jj diffedit` | `<FILESETS>` |
| `jj file chmod` | `<FILESETS>` |
| `jj file list` | `<FILESETS>` |
| `jj file search` | `<FILESETS>` |
| `jj file show` | `<FILESETS>` |
| `jj file track` | `<FILESETS>` |
| `jj file untrack` | `<FILESETS>` |
| `jj fix` | `<FILESETS>` |
| `jj interdiff` | `<FILESETS>` |
| `jj log` | `<FILESETS>` |
| `jj resolve` | `<FILESETS>` |
| `jj restore` | `<FILESETS>` |
| `jj split` | `<FILESETS>` |
| `jj squash` | `<FILESETS>` |
| `jj status` | `<FILESETS>` |
