# jj Forge Workflows Reference

jj integrates with code hosting platforms (forges) through Git remotes. This reference covers workflows for GitHub, GitLab, and Gerrit Code Review.

> **Source of truth**: <https://jj-vcs.github.io/jj/latest/github/>, <https://jj-vcs.github.io/jj/latest/gerrit/>. For **bookmark details and push safety**, see [bookmarks.md](bookmarks.md). For **CLI commands**, see [cli.md](cli.md). For **config settings**, see [config.md](config.md).


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

