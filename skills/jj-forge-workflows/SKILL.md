---
name: jj-forge-workflows
description: >
  Reference for jj (Jujutsu VCS) integration with forges — GitHub, GitLab,
  and Gerrit workflows including pushing, pull requests, merge requests,
  code review, multi-remote setup, push options, and Gerrit upload. Triggers on:
  "jj github", "jj gitlab", "jj gerrit", "jj push", "jj forge", "jj pr",
  "jj pull request", "jj merge request", "jj code review", "jj gerrit upload",
  "jj remote", "jj upstream", "push options".
user-invocable: false
---

# jj Forge Workflows Reference

jj integrates with code hosting platforms (forges) through Git remotes. This reference covers workflows for GitHub, GitLab, and Gerrit Code Review.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/github/>, <https://jj-vcs.github.io/jj/latest/gerrit/>. For **bookmark details and push safety**, see the `jj-bookmarks` skill. For **CLI commands**, see the `jj-cli` skill. For **config settings**, see the `jj-config` skill.

---

## GitHub / GitLab Workflow

### Basic Workflow: Generated Bookmark Name

```bash
# Start a new commit off of the default bookmark
jj new main

# Make changes, describe, and create a new commit
jj commit -m 'refactor(foo): restructure foo()'
jj commit -m 'feat(bar): add support for bar'

# Push the parent commit with an auto-generated bookmark name
jj git push --change @-   # or -c
```

`--change` / `-c` creates a bookmark named `push-<change-id-prefix>` and pushes it. Subsequent pushes with the same change ID update the same bookmark.

### Basic Workflow: Named Bookmark

```bash
# Start a new commit off of the default bookmark
jj new main

# Make changes and describe
jj commit -m 'refactor(foo): restructure foo()'
jj commit -m 'feat(bar): add support for bar'

# Create a bookmark on the working-copy commit's parent
jj bookmark create my-feature -r @-

# Track the bookmark on the remote
jj bookmark track my-feature@origin

# Push
jj git push
```

**Note:** Unlike Git, jj does not automatically move bookmarks when creating new commits. You must manually move bookmarks with `jj bookmark move`.

### Updating the Repository

jj has no direct equivalent of `git pull`. Use a two-step process:

```bash
# 1. Fetch everything from the remote
jj git fetch

# 2. Rebase your work on top of the updated main
jj rebase -o main
```

If you have multiple outstanding branches, use `-b` for each:

```bash
jj rebase -b feature-a -o main
jj rebase -b feature-b -o main
```

### Addressing Review Comments

#### Adding New Commits (GitHub-style)

```bash
# Create a new commit on top of the bookmark
jj new your-feature

# Make changes, describe
jj diff
jj commit -m 'address pr comments'

# Move the bookmark forward
jj bookmark move your-feature --to @-

# Push
jj git push
```

#### Rewriting Commits (clean history preference)

```bash
# Create a commit on top of the commit that needs fixing
jj new your-feature-

# Make changes
jj diff

# Squash changes into the parent commit
jj squash

# Force push (jj handles this automatically)
jj git push --bookmark your-feature
```

The hyphen in `your-feature-` is revset syntax for "the parent of `your-feature`".

### Working with Other People's Bookmarks

By default, `jj git clone` only imports the default remote bookmark (usually `main`). To work with others' bookmarks:

```bash
# Option 1: Create a commit on top of a remote bookmark directly
jj new <bookmark>@<remote>

# Option 2: Track all remote bookmarks automatically
# In config: remotes.<name>.auto-track-bookmarks = "*"
jj new <bookmark>
```

### Using GitHub CLI

GitHub CLI needs a Git repo to work. For non-colocated jj repos, set the `GIT_DIR` environment variable:

```bash
GIT_DIR=$(jj git root) gh issue list
```

Automate this with direnv:

```bash
# .envrc
export GIT_DIR=$(jj git root)
```

Then `direnv allow` to activate.

### Useful Revsets for Forge Workflows

```bash
# All revisions across local bookmarks not on main or any remote
jj log -r 'bookmarks() & ~(main | remote_bookmarks())'

# All revisions you authored, across bookmarks not on any remote
jj log -r 'mine() & bookmarks() & ~remote_bookmarks()'

# All remote bookmarks you authored or committed to
jj log -r 'remote_bookmarks() & (mine() | committer(your@email.com))'

# All ancestors of working copy not on any remote
jj log -r 'remote_bookmarks()..@'
```

---

## Multiple Remotes

Common pattern: `upstream` for the source repo, `origin` for your fork.

```bash
# Clone with explicit upstream remote
jj git clone --remote upstream https://github.com/upstream-org/repo

cd repo

# Add your fork as origin
jj git remote add origin git@github.com:your-org/your-repo-fork
```

### Configuring Default Remotes

```toml
# Basic: fetch from upstream, push to origin
[git]
fetch = "upstream"
push = "origin"

# Fetch from both (to keep your own bookmarks synchronized)
[git]
fetch = ["upstream", "origin"]
push = "origin"
```

Default for both `git.fetch` and `git.push` is `"origin"`.

---

## Git Push Options

`jj git push` supports passing Git push options to the server via `-o`/`--option`:

```bash
# Single option
jj git push -o <push_option>
jj git push --option <push_option>

# Multiple options
jj git push -o foo -o bar=val

# Value with spaces (use double quotes)
jj git push -o 'key=value with spaces'
```

Support is server-dependent (GitLab supports push options for CI and merge requests; other platforms may not).

### GitLab Push Option Examples

```bash
# Skip CI for a push
jj git push -o ci.skip

# Pass CI variables
jj git push -o 'ci.variable=MAX_RETRIES=10' -o 'ci.variable=MAX_TIME=600'

# Create a merge request on push
jj git push --allow-new \
    -o merge_request.create \
    -o merge_request.target=main \
    -o 'merge_request.title=Add feature X' \
    -o merge_request.draft

# Auto-merge when pipeline succeeds
jj git push \
    -o merge_request.merge_when_pipeline_succeeds \
    -o merge_request.remove_source_branch

# Add/remove labels
jj git push \
    -o 'merge_request.label=label1' \
    -o 'merge_request.label=label2' \
    -o 'merge_request.unlabel=label3'

# Assign/unassign users
jj git push \
    -o 'merge_request.assign=user1' \
    -o 'merge_request.unassign=user2'
```

---

## Gerrit Integration

jj and Gerrit share the same mental model: both track a "change identity" across rewrites. jj's change IDs and Gerrit's `Change-Id` trailers are philosophically aligned. `jj gerrit upload` bridges the gap.

### Setup

```bash
# Option 1: Start jj in an existing Git repo with Gerrit remotes
jj git init --colocate

# Option 2: Add a Gerrit remote to a jj repo
jj git remote add gerrit https://review.gerrithub.io/yourname/yourproject

# Option 3: Clone with Gerrit remote
jj git clone https://review.gerrithub.io/your/project
```

### Configuration

```toml
[gerrit]
default-remote = "gerrit"       # Git remote to push to
default-remote-branch = "main"   # Target branch in Gerrit
```

### Upload Changes

```bash
# Upload @ if it has a description, otherwise upload @-
jj gerrit upload

# Upload a specific revision
jj gerrit upload -r @-

# Upload multiple revisions (expands to stack of commits)
jj gerrit upload -r <revset>

# Preview without pushing
jj gerrit upload -r '@-' --remote-branch main --dry-run
```

Each jj change maps to a single Gerrit change based on the jj change ID.

### Target Branch and Remote Override

```bash
# One-time branch override
jj gerrit upload --remote-branch <branch>

# One-time remote override
jj gerrit upload --remote <remote>

# Persistent config
jj config set --repo gerrit.default-remote-branch <branch>
jj config set --repo gerrit.default-remote <remote>
```

### Update Changes After Review

Address review feedback, then re-upload with the same revsets. Gerrit creates new patch sets on existing changes:

```bash
# Edit an earlier commit
jj edit xcv
# ... apply changes ...
jj gerrit upload -r xcv
```

### Change-Id Management

If you don't provide a `Change-Id` footer, `jj gerrit upload` generates a transient one based on the jj change ID. As long as the jj change ID stays the same (and no explicit Change-Id is added), uploads create new patch sets on the existing change.

**Tips for splitting and squashing:**
- When **splitting**, the portion you want associated with the original Gerrit change should stay in the original jj change
- When **squashing**, squash into the change that was previously uploaded to Gerrit

**Manual Change-Id:** Copy a Gerrit `Change-Id` footer into the jj commit description to directly assign a jj change to an existing Gerrit change.

### Automatic Change-Id Footer Configuration

```toml
[templates]
commit_trailers = '''
if(
  !trailers.contains_key("Change-Id"),
  format_gerrit_change_id_trailer(self)
)
'''
```

**Important caveats when using automatic Change-Id footers:**
- The Gerrit change mapping is defined entirely by the `Change-Id` footers
- Keep `Change-Id` footers associated with the desired changes when splitting/squashing
- Never duplicate the same `Change-Id` across different changes
- Gerrit rejects pushes with duplicate `Change-Id`s
- Separate uploads may unintentionally overwrite existing changes

### Alternative: Link Trailer

Since Gerrit 3.3.1, a `Link` trailer can replace `Change-Id`:

```bash
jj config set --repo gerrit.review-url <reviewUrl>
```

This causes `jj gerrit upload` to use `Link` trailers in the format `<reviewUrl>/id/I<changeid>`.

---

## Working in Colocated vs Pure jj Repos

### Colocated Workspace

After `jj git init` or `jj git clone`, the `.jj` and `.git` directories are siblings. Every `jj` command auto-syncs with Git. Git will be in detached HEAD state.

```bash
# Colocated workflow
nvim docs/tutorial.md
jj commit -m "Update tutorial"
jj bookmark create doc-update -r @-
jj bookmark track doc-update@origin
jj git push
```

### Pure jj Repository

No `.git` directory visible. Bookmarks must be explicitly imported/exported:

```bash
jj git import    # pull Git refs into jj
jj git export    # push jj refs to Git
```

Or generate a bookmark for push:

```bash
jj git push -c mw    # creates bookmark "push-mwmpwkwknuz"
```