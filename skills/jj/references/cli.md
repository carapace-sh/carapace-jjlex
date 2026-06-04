# jj CLI Reference

Command-line reference for jj (Jujutsu VCS). Argument types: **REVSET** = revision expression (see [revsets.md](revsets.md)), **FILESET** = file expression (see [filesets.md](filesets.md)).

> **Source of truth**: `jj --help`, <https://jj-vcs.github.io/jj/latest/cli-reference/>. For **concepts**, see [concepts.md](concepts.md). For **bookmarks**, see [bookmarks.md](bookmarks.md). For **operations**, see [operations.md](operations.md). For **configuration**, see [config.md](config.md). For **templates**, see [templates.md](templates.md). For **git comparison**, see [git-compat.md](git-compat.md).


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
| `jj show [REVSETS]` | Show commit description and changes |
| `jj diff [FILESETS]...` | Compare file contents between revisions |
| `jj root` | Show current workspace root directory |
| `jj version` | Display version information |

### Revision Operations

| Command | Description |
|---------|-------------|
| `jj new [REVSETS]...` | Create a new empty change |
| `jj commit [FILESETS]...` / `jj ci` | Describe working copy and create new change on top |
| `jj describe [REVSETS]...` / `jj desc` | Update commit message |
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

See [bookmarks.md](bookmarks.md) for conceptual details.

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

See [git-compat.md](git-compat.md) for full Git comparison and migration guide.

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

See [operations.md](operations.md) for conceptual details.

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

See [config.md](config.md) for all config sections and settings.

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
| `jj util backend name` | Print the commit backend being used |
| `jj util completion <SHELL>` | Print shell completion script |
| `jj util config-schema` | Print JSON schema for jj TOML config |
| `jj util exec <CMD>` | Execute an external command |
| `jj util gc` | Run backend garbage collection |
| `jj util install-man-pages <PATH>` | Install man pages |
| `jj util markdown-help` | Print CLI help in Markdown |
| `jj util snapshot` | Snapshot working copy if needed |
| `jj arrange [REVSETS]` | Interactively arrange the commit graph |
| `jj evolog` / `jj evolution-log` | Show how a change has evolved over time |

## Arguments Accepting Revsets

Many jj commands accept **REVSET** arguments (see [revsets.md](revsets.md) for full syntax):

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
| `jj show` | `<REVSETS>` |
| `jj sign/unsign` | `--revision <REVSETS>` |
| `jj simplify-parents` | `--source <REVSETS>`, `--revision <REVSETS>` |
| `jj split` | `--revision <REVSET>`, `--onto`, `--insert-after`, `--insert-before` |
| `jj squash` | `--revision <REVSET>`, `--from <REVSETS>`, `--into <REVSET>`, `--onto`, `--insert-after`, `--insert-before` |
| `jj tag list/set` | `--revision <REVSETS>` |
| `jj workspace add` | `--revision <REVSETS>` |


## Arguments Accepting Templates

Commands that accept `-T`/`--template` flags (see [templates.md](templates.md) for the template language):

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
