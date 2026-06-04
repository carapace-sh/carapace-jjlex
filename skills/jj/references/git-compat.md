# jj Git Compatibility Reference

jj uses Git as its storage backend and provides seamless interoperability. This reference covers conceptual differences, command equivalents, colocation, and migration.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/git-comparison/>, <https://jj-vcs.github.io/jj/latest/git-compatibility/>. For **CLI commands**, see [cli.md](cli.md). For **bookmark details**, see [bookmarks.md](bookmarks.md). For **core concepts**, see [concepts.md](concepts.md).


## Command Equivalents

### Repository Setup

| Use Case | Git | jj |
|----------|-----|-----|
| Create repo | `git init` | `jj git init [--no-colocate]` |
| Clone | `git clone <url>` | `jj git clone <url>` |
| Clone with branch | `git clone -b <branch> <url>` | `jj git clone -b <branch> <url>` |

### Remote Operations

| Use Case | Git | jj |
|----------|-----|-----|
| Fetch | `git fetch [<remote>]` | `jj git fetch [--remote <remote>]` |
| Push all | `git push --all` | `jj git push --all [--remote <remote>]` |
| Push single | `git push <remote> <branch>` | `jj git push --bookmark <name>` |
| Push by commit | No direct equivalent | `jj git push -c <rev>` |
| Add remote | `git remote add <name> <url>` | `jj git remote add <name> <url>` |
| List remotes | `git remote -v` | `jj git remote list` |
| Remove remote | `git remote remove <name>` | `jj git remote remove <name>` |

### Status and Diff

| Use Case | Git | jj |
|----------|-----|-----|
| Status | `git status` | `jj st` |
| Diff current | `git diff HEAD` | `jj diff` |
| Diff specific revision | `git diff <rev>^ <rev>` | `jj diff -r <rev>` |
| Diff from A to current | `git diff <rev>` | `jj diff --from <rev>` |
| Diff from A to B | `git diff A B` | `jj diff --from A --to B` |
| Diff all in range | `git diff A...B` | `jj diff -r A..B` |
| Show commit | `git show <rev>` | `jj show <rev>` |

### File Operations

| Use Case | Git | jj |
|----------|-----|-----|
| Add file | `touch filename; git add filename` | `touch filename` |
| Remove file | `git rm filename` | `rm filename` |
| Untrack (keep file) | `git rm --cached filename` | `jj file untrack filename` |
| List files | `git ls-files` | `jj file list` |
| Blame | `git blame <file>` | `jj file annotate <file>` |
| Search in files | `git grep pattern` | `jj file search --pattern pattern` |

### Commit Operations

| Use Case | Git | jj |
|----------|-----|-----|
| Commit all changes | `git commit -a` | `jj commit` |
| Create new empty commit | No direct equivalent | `jj new` |
| Abandon current changes | `git reset --hard` | `jj abandon` |
| Make current empty | `git reset --hard` | `jj restore` |
| Squash into parent | `git commit --amend -a` | `jj squash` |
| Interactive squash | `git add -p; git commit --amend` | `jj squash -i` |
| Move changes into ancestor | `git commit --fixup=X; git rebase --autosquash X^` | `jj squash --into X` |
| Move specific files into parent | No direct equivalent | `jj squash <file>` |
| Split changes | `git commit -p` | `jj split` |
| Split arbitrary commit | No direct equivalent | `jj split -r <rev>` |
| Edit diff directly | No direct equivalent | `jj diffedit -r <rev>` |
| Edit description | `git commit --amend -m "msg"` | `jj describe` |
| Edit any description | `git commit --fixup=reword:X; git rebase --autosquash` | `jj describe X` |
| Discard parent, keep diff | `git reset --soft HEAD~` | `jj squash --from @-` |

### History and Navigation

| Use Case | Git | jj |
|----------|-----|-----|
| Log ancestors | `git log --oneline --graph` | `jj log -r ::@` |
| Log all | `git log --oneline --graph --all` | `jj log -r 'all()'` or `jj log -r ::` |
| Log not on main | `git log --oneline --branches --not upstream/main` | `jj log` |
| Search diffs | `git log -G pattern` | `jj log -r 'diff_lines(regex:pattern)'` |
| Move to child | No direct equivalent | `jj next` |
| Move to parent | No direct equivalent | `jj prev` |
| Edit any commit | `git checkout <rev>` (detached) | `jj edit <rev>` |

### Branch / Bookmark Operations

| Use Case | Git | jj |
|----------|-----|-----|
| List | `git branch` | `jj bookmark list` / `jj b l` |
| Create | `git branch <name> <rev>` | `jj bookmark create <name> -r <rev>` |
| Move forward | `git branch -f <name> <rev>` | `jj bookmark move <name> --to <rev>` |
| Move backward | `git branch -f <name> <rev>` | `jj bookmark move <name> --to <rev> --allow-backwards` |
| Delete | `git branch -d <name>` | `jj bookmark delete <name>` |
| Start new branch | `git switch -c topic main` | `jj new main` |
| Merge branch A into current | `git merge A` | `jj new @ A` |

### Rebase and Restructure

| Use Case | Git | jj |
|----------|-----|-----|
| Rebase branch A onto B | `git rebase B A` | `jj rebase -b A -o B` |
| Rebase with descendants | `git rebase --onto B A^ <descendant>` | `jj rebase -s A -o B` |
| Reorder commits | `git rebase -i A` | `jj rebase -r C --before B` or `jj arrange` |
| Parallelize commits | No direct equivalent | `jj parallelize A B` |
| Simplify parents | No direct equivalent | `jj simplify-parents` |
| Cherry-pick | `git cherry-pick <rev>` | `jj duplicate <rev> -o <dest>` |

### Stash

| Use Case | Git | jj |
|----------|-----|-----|
| Stash changes | `git stash` | `jj new @-` |
| Pop stash | `git stash pop` | `jj squash --from @` |

### Undo and Operations

| Use Case | Git | jj |
|----------|-----|-----|
| Undo last operation | `git reset` (limited) | `jj undo` |
| Redo | No direct equivalent | `jj redo` |
| View operation history | No equivalent | `jj op log` |
| Restore to earlier state | `git reset --hard <ref>` | `jj op restore <op>` |
| Revert specific operation | No equivalent | `jj op revert <op>` |

### Tags

| Use Case | Git | jj |
|----------|-----|-----|
| List | `git tag -l` | `jj tag list` |
| Create | `git tag <name> <rev>` | `jj tag set <name> -r <rev>` |
| Delete | `git tag -d <name>` | `jj tag delete <name>` |
| List tags containing rev | `git tag --contains <rev>` | `jj tag list -r '<rev>::'` |
| List tags merged into rev | `git tag --merged <rev>` | `jj tag list -r '::<rev>'` |


## Migration from Git

### Existing Git Repo

```bash
# In your existing Git repo directory:
jj git init --git-repo=.    # Use existing .git as backend
# Or for colocated setup (default):
jj git init                   # Creates .jj alongside .git
```

### Key Workflow Changes

1. **No more staging** — Edit files directly, they're auto-tracked. Use `jj squash -i` or `jj split` for partial commits.

2. **No more branches** — Use bookmarks. They don't auto-move, so explicitly set them with `jj bookmark set` or `jj bookmark move`.

3. **Undo is powerful** — `jj undo` reverses any operation. `jj op log` shows complete history. No more reflog hacks.

4. **Conflicts don't block** — Merge conflicts are recorded in commits. Resolve later with `jj resolve` or by editing files.

5. **Everything auto-rebases** — Amending a commit automatically rebases all descendants. No manual rebase needed.