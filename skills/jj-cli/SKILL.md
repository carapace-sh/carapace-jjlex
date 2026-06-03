---
name: jj-cli
description: >
  Reference for jj (Jujutsu VCS) command-line interface — all commands,
  subcommands, flags, and argument types. For conceptual model see jj-concepts;
  for bookmarks see jj-bookmarks; for operation log see jj-operations; for
  configuration see jj-config; for template language see jj-templates; for git
  comparison see jj-git-compat; for revset syntax see jj-revsets; for fileset
  syntax see jj-filesets. Triggers on: "jj cli", "jj command", "jj usage",
  "jj subcommand", "jj flag", "jj help".
user-invocable: true
---

# jj CLI Reference

Command-line reference for jj (Jujutsu VCS). Argument types: **REVSET** = revision expression (see `jj-revsets`), **FILESET** = file expression (see `jj-filesets`).

> **Source of truth**: `jj --help`, <https://jj-vcs.github.io/jj/latest/cli-reference/>. For **concepts**, see `jj-concepts`. For **bookmarks**, see `jj-bookmarks`. For **operations**, see `jj-operations`. For **configuration**, see `jj-config`. For **templates**, see `jj-templates`. For **git comparison**, see `jj-git-compat`.

---

## Global Flags

| Flag | Description |
|------|-------------|
| `-R`, `--repository <PATH>` | Path to repository |
| `--ignore-working-copy` | Don't snapshot or update the working copy |
| `--ignore-immutable` | Allow rewriting immutable commits |
| `--at-operation <OP>` / `--at-op` | Load repo at a specific operation (see `jj-operations`) |
| `--no-integrate-operation` | Don't integrate any operations |
| `--debug` | Enable debug logging |
| `--color <WHEN>` | Colorize: `always`, `never`, `debug`, `auto` |
| `--quiet` | Silence non-primary output |
| `--no-pager` | Disable the pager |
| `--config <NAME=VALUE>` | Additional config (repeatable) |
| `--config-file <PATH>` | Additional config file (repeatable) |

---

## Command Summary

### Repository Setup

| Command | Description |
|---------|-------------|
| `jj git init [--no-colocate]` | Create a new Git-backed repo |
| `jj git clone <URL> [<DEST>] [--remote <NAME>]` | Clone a Git repo |
| `jj git init --git-repo=<PATH>` | Create jj repo from existing Git repo |

### Status and Information

| Command | Description |
|---------|-------------|
| `jj status` / `jj st` | Show high-level repo status |
| `jj log [FILESETS]...` | Show revision history |
| `jj show [REVSET]` | Show commit description and changes |
| `jj diff [FILESETS]...` | Compare file contents between revisions |
| `jj root` | Show current workspace root directory |
| `jj version` | Display version information |

### Revision Operations

| Command | Description |
|---------|-------------|
| `jj new [REVSETS]...` | Create a new empty change |
| `jj commit [FILESETS]...` | Describe working copy and create new change on top |
| `jj describe [REVSETS]...` | Update commit message |
| `jj edit <REVSET>` | Set a revision as the working-copy revision |
| `jj duplicate [REVSETS]...` | Create new changes with the same content |
| `jj abandon [REVSETS]...` | Abandon a revision |
| `jj squash [FILESETS]...` | Move changes from one revision into another |
| `jj split [FILESETS]...` | Split a revision in two |
| `jj rebase` | Move revisions to different parent(s) |
| `jj absorb [FILESETS]...` | Move changes into the stack of mutable revisions |
| `jj parallelize [REVSETS]...` | Make revisions siblings |
| `jj simplify-parents` | Simplify parent edges |
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
| `jj fix [FILESETS]...` | Update files with formatting fixes |

### Conflict Resolution

| Command | Description |
|---------|-------------|
| `jj resolve [FILESETS]...` | Resolve conflicted files with an external merge tool |

### Bookmarks (`jj bookmark` / `jj b`)

See `jj-bookmarks` for conceptual details.

| Command | Description |
|---------|-------------|
| `jj bookmark list` / `jj b l` | List bookmarks |
| `jj bookmark create <NAME>` / `jj b c` | Create a bookmark |
| `jj bookmark move <NAME>` / `jj b m` | Move a bookmark |
| `jj bookmark delete <NAME>` / `jj b d` | Delete a bookmark |
| `jj bookmark forget <NAME>` / `jj b f` | Forget a bookmark |
| `jj bookmark rename <OLD> <NEW>` / `jj b r` | Rename a bookmark |
| `jj bookmark set <NAME>` / `jj b s` | Create or update a bookmark |
| `jj bookmark advance` / `jj b a` | Advance a bookmark |
| `jj bookmark track <NAME>` / `jj b t` | Track a remote bookmark |
| `jj bookmark untrack` | Stop tracking a remote bookmark |

### Tags (`jj tag`)

| Command | Description |
|---------|-------------|
| `jj tag list` / `jj tag l` | List tags |
| `jj tag set <NAME>` / `jj tag s` | Create or update a tag |
| `jj tag delete <NAME>` / `jj tag d` | Delete tags |

### File Operations (`jj file`)

| Command | Description |
|---------|-------------|
| `jj file list [FILESETS]...` | List files in a revision |
| `jj file show <PATH>` | Print file contents |
| `jj file annotate <PATH>` | Show source change per line |
| `jj file chmod <MODE> <PATH>` | Set or remove executable bit |
| `jj file search [FILESETS]...` | Search for content in files |
| `jj file track <PATH>` | Start tracking a path |
| `jj file untrack <PATH>` | Stop tracking a path |

### Git Integration (`jj git`)

See `jj-git-compat` for full Git comparison and migration guide.

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

See `jj-operations` for conceptual details.

| Command | Description |
|---------|-------------|
| `jj op log` | Show the operation log |
| `jj op show <OP>` | Show changes in an operation |
| `jj op diff` | Compare changes between two operations |
| `jj op restore <OP>` | Restore repo to earlier state |
| `jj op revert <OP>` | Revert a specific operation |
| `jj op abandon` | Abandon operation history |
| `jj op integrate` | Make an operation part of the log |
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

See `jj-config` for all config sections and settings.

| Command | Description |
|---------|-------------|
| `jj config edit --user/--repo/--workspace` | Edit config in editor |
| `jj config get <NAME>` | Get config value |
| `jj config list [NAME]` | List config variables |
| `jj config path --user/--repo/--workspace` | Print config file path |
| `jj config set --user/--repo/--workspace <NAME> <VALUE>` | Set config value |
| `jj config unset --user/--repo/--workspace <NAME>` | Unset config value |

### Other Commands

| Command | Description |
|---------|-------------|
| `jj sparse list/set/edit/reset` | Manage sparse working copy patterns |
| `jj bisect run --range <REVSETS> <CMD>` | Find bad revision by bisection |
| `jj gerrit upload` | Upload changes to Gerrit |
| `jj util completion <SHELL>` | Print shell completion script |
| `jj util config-schema` | Print JSON schema for jj TOML config |
| `jj util exec <CMD>` | Execute an external command |
| `jj util gc` | Run backend garbage collection |
| `jj util install-man-pages <PATH>` | Install man pages |
| `jj util markdown-help` | Print CLI help in Markdown |
| `jj util snapshot` | Snapshot working copy if needed |
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
| `-n`, `--limit <N>` | Limit number of revisions |
| `--reversed` | Show older revisions first |
| `-G`, `--no-graph` | Flat list (no graph) |
| `-T`, `--template <TMPL>` | Custom output template (see `jj-templates`) |
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
| `-r`, `--revisions <REVSETS>` | Show changes in these revisions |
| `-f`, `--from <REVSET>` | Show changes from this revision |
| `-t`, `--to <REVSET>` | Show changes to this revision |
| `-T`, `--template <TMPL>` | Custom output template |
| `-s`, `--summary` / `--stat` / `--types` / `--name-only` | Diff format |
| `--git` / `--color-words` | Diff format |
| `--tool <TOOL>` | Use external diff tool |
| `--context <N>` | Lines of context |
| `-w`, `--ignore-all-space` / `-b`, `--ignore-space-change` | Whitespace options |

### `jj show`

```
jj show [OPTIONS] [REVSET]
```

| Flag | Description |
|------|-------------|
| `<REVSET>` | Revision to show. Default: `@` |
| `-T`, `--template <TMPL>` | Custom output template (see `jj-templates`) |
| `--no-patch` | Don't show the patch |
| Diff format flags | `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words`, `--tool`, `--context` |
| Whitespace flags | `-w`, `-b` |

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
| `-A`, `--insert-after <REVSETS>` | Insert after given commit(s) |
| `-B`, `--insert-before <REVSETS>` | Insert before given commit(s) |
| `--skip-emptied` | Abandon commits that become empty after rebase |
| `--keep-divergent` | Keep divergent commits |
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
| `--restore-descendants` | Don't modify content of children |

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
| `-T`, `--template <TMPL>` | Custom output template (see `jj-templates`) |
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
| `--named <NAME=REVISION>` | New bookmark name and revision |
| `--dry-run` | Preview what will be pushed |
| `--allow-empty-description` | Allow pushing commits with empty descriptions |
| `--allow-private` | Allow pushing private commits |
| `-o`, `--option <OPTION>` | Git push options |

### `jj git fetch`

| Flag | Description |
|------|-------------|
| `--remote <REMOTE>` | Remote to fetch from |
| `-b`, `--branch <BRANCH>` | Branch to fetch (repeatable) |
| `--tracked` | Fetch only tracked bookmarks |
| `--all-remotes` | Fetch from all remotes |

### `jj operation` Subcommands

See `jj-operations` for conceptual details.

#### `jj op log`

| Flag | Description |
|------|-------------|
| `-n`, `--limit <N>` | Limit number of operations |
| `--reversed` | Older operations first |
| `-G`, `--no-graph` | Flat list |
| `-T`, `--template <TMPL>` | Custom template (see `jj-templates`) |
| `-d`, `--op-diff` | Show repo changes at each operation |
| `-p`, `--patch` | Show patch (implies --op-diff) |
| Diff format flags | `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words` |
| `--show-changes-in <REVSETS>` | Filter changed revisions |

#### `jj op show [<OP>]`

| Flag | Description |
|------|-------------|
| `<OP>` | Operation to show. Default: `@` |
| `-G`, `--no-graph` / `-T`, `--template <TMPL>` | Display options |
| `--no-op-diff` | Don't show operation diff |
| Diff format flags | `-p`, `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words` |

### `jj evolog`

| Flag | Description |
|------|-------------|
| `-r`, `--revisions <REVSETS>` | Which revisions to follow. Default: `@` |
| `-n`, `--limit <N>` | Limit number of revisions |
| `--reversed` / `-G`, `--no-graph` / `-T`, `--template <TMPL>` | Display options |
| Diff format flags | `-p`, `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words`, `--tool`, `--context` |

### `jj next` / `jj prev`

| Flag | Description |
|------|-------------|
| `<OFFSET>` | How many revisions to move (default: 1) |
| `-e`, `--edit` | Edit the target commit directly |
| `-n`, `--no-edit` | Don't edit the target commit |
| `--conflict` | Jump to next/previous conflicted descendant/ancestor |

### `jj metaedit`

| Flag | Description |
|------|-------------|
| `<REVSETS>` | Revision(s) to modify. Default: `@` |
| `--update-change-id` | Generate a new change-id |
| `-m`, `--message <MSG>` | Update the change description |
| `--update-author-timestamp` | Update the author timestamp |
| `--update-author` | Update the author to the configured user |
| `--author <AUTHOR>` | Set author to the provided string |
| `--author-timestamp <TS>` | Set the author date |
| `--force-rewrite` | Rewrite even if no metadata changed |

### `jj fix`

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Fix only these paths |
| `-s`, `--source <REVSETS>` | Fix files in these revision(s) and descendants |
| `--include-unchanged-files` | Fix unchanged files too |
| `-a`, `--all-lines` | Format all lines instead of only modified lines |

### `jj resolve`

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Only resolve conflicts in these paths |
| `-r`, `--revision <REVSET>` | Revision with conflicts. Default: `@` |
| `-l`, `--list` | List all conflicts instead of resolving |
| `--tool <NAME>` | Specify 3-way merge tool |

### `jj interdiff`

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Restrict the diff to these paths |
| `-f`, `--from <REVSET>` | First revision to compare |
| `-t`, `--to <REVSET>` | Second revision to compare |
| Diff format flags | `-s`, `--stat`, `--types`, `--name-only`, `--git`, `--color-words`, `--tool`, `--context` |
| Whitespace flags | `-w`, `-b` |

### `jj diffedit`

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Edit only these paths |
| `-r`, `--revision <REVSET>` | Revision to touch up. Default: `@` |
| `-f`, `--from <REVSET>` / `-t`, `--to <REVSET>` | Diff source/target |
| `--tool <NAME>` | Specify diff editor |
| `--restore-descendants` | Preserve content when rebasing descendants |

### `jj revert`

| Flag | Description |
|------|-------------|
| `-r`, `--revision <REVSETS>` | The revision(s) to apply the reverse of |
| `-o`, `--onto <REVSETS>` | Revision(s) to apply reverse changes on top of |
| `-A`, `--insert-after <REVSETS>` | Insert reverse changes after |
| `-B`, `--insert-before <REVSETS>` | Insert reverse changes before |

### `jj sign` / `jj unsign`

| Flag | Description |
|------|-------------|
| `-r`, `--revision <REVSETS>` | What revision(s) to sign/unsign |
| `--key <KEY>` | The key used for signing (sign only) |

### `jj workspace add`

| Flag | Description |
|------|-------------|
| `<DESTINATION>` | Where to create the new workspace |
| `--name <NAME>` | A name for the workspace |
| `-r`, `--revision <REVSETS>` | Parent revisions for the working-copy commit |
| `-m`, `--message <MSG>` | Change description |
| `--sparse-patterns <MODE>` | `copy`, `full`, or `empty` (default: `copy`) |

### `jj sparse set`

| Flag | Description |
|------|-------------|
| `--add <PATTERN>` | Patterns to add |
| `--remove <PATTERN>` | Patterns to remove |
| `--clear` | Include no files (combine with --add) |

### `jj tag list` / `jj tag l`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Filter by name patterns |
| `-a`, `--all-remotes` | Show all remote tags |
| `-c`, `--conflicted` | Show conflicted tags only |
| `-r`, `--revision <REVSETS>` | Show tags pointing to these revisions |
| `-T`, `--template <TMPL>` | Custom output template |
| `--sort <KEY>` | Sort key (same options as bookmarks) |

### `jj tag set` / `jj tag s`

| Flag | Description |
|------|-------------|
| `<NAMES>` | Tag names to create or update |
| `-r`, `--revision <REVSET>` | Target revision. Default: `@` |
| `--allow-move` | Allow moving existing tags |

### `jj git clone`

| Flag | Description |
|------|-------------|
| `<SOURCE>` | URL or path of the Git repo to clone |
| `<DESTINATION>` | Target directory (optional) |
| `--remote <NAME>` | Name of the remote (default: `origin`) |
| `--colocate` / `--no-colocate` | Control colocation |
| `--depth <DEPTH>` | Create a shallow clone |
| `--fetch-tags <MODE>` | When to fetch tags: `all`, `included`, `none` |
| `-b`, `--branch <BRANCH>` | Branch to fetch and use as parent (repeatable) |

### `jj git init`

| Flag | Description |
|------|-------------|
| `<DESTINATION>` | Target directory (default: `.`) |
| `--colocate` / `--no-colocate` | Control colocation |
| `--git-repo <PATH>` | Path to existing Git repo as backend |

### `jj file` Subcommands

#### `jj file annotate`

| Flag | Description |
|------|-------------|
| `<PATH>` | File to annotate |
| `-r`, `--revision <REVSET>` | Starting revision |
| `-T`, `--template <TMPL>` | Custom output template per line |

#### `jj file list`

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Filter by paths |
| `-r`, `--revision <REVSET>` | Revision to list files in. Default: `@` |
| `-T`, `--template <TMPL>` | Custom output template |

#### `jj file search`

| Flag | Description |
|------|-------------|
| `<FILESETS>` | Filter by paths |
| `-r`, `--revision <REVSET>` | Revision to search. Default: `@` |
| `-p`, `--pattern <PATTERN>` | Pattern to search for |

### `jj bisect run`

| Flag | Description |
|------|-------------|
| `--range <REVSETS>` / `-r` | Range of revisions to bisect |
| `--find-good` | Find first good revision instead |
| `<COMMAND>` | Command to run |

### `jj gerrit upload`

| Flag | Description |
|------|-------------|
| `-r`, `--revision <REVSETS>` | Revisions to send |
| `-b`, `--remote-branch <BRANCH>` | Target branch |
| `--remote <REMOTE>` | Gerrit remote |
| `-n`, `--dry-run` | Preview only |
| `--reviewer <EMAIL>` / `--cc <EMAIL>` | Reviewers/CCs |
| `-l`, `--label <LABEL>` / `--hashtag <TAG>` / `--topic <TOPIC>` | Metadata |
| `-m`, `--message <MSG>` | Patch set description |
| `--wip` / `--ready` / `--private` / `--remove-private` | Change visibility |
| `--submit` / `--skip-validation` | Submit options |

---

## Arguments Accepting Revsets

Many jj commands accept **REVSET** arguments (see the `jj-revsets` skill for full syntax):

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

---

## Arguments Accepting Templates

Commands that accept `-T`/`--template` flags (see the `jj-templates` skill for the template language):

| Command | Template Context |
|---------|-----------------|
| `jj log` | Commit |
| `jj show` | Commit |
| `jj evolog` | Commit |
| `jj op log` | Operation |
| `jj op show` | Operation |
| `jj bookmark list` | Bookmark info |
| `jj tag list` | Tag info |
| `jj workspace list` | Workspace info |
| `jj diff` | File diff entry |
| `jj file list` | File entry |
| `jj file annotate` | Per-line info |
