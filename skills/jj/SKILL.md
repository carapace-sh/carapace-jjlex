---
name: jj
description: >
  Use when working with jj (Jujutsu VCS) — commands, concepts, bookmarks, revsets,
  filesets, templates, operations, conflicts, divergence, config, forge workflows,
  or Git compatibility. Triggers on: "jj", "jujutsu", "jj vcs", "jj version control",
  "jj command", "jj revset", "jj fileset", "jj template", "jj bookmark", "jj operation",
  "jj config", "jj conflict", "jj divergent", "jj git", "jj push", "jj forge",
  "jj gerrit", "jj github", "jj gitlab".
user-invocable: true
---

# jj — Jujutsu VCS Reference

Comprehensive reference for jj (Jujutsu VCS), a version-control system that uses Git as its storage backend with a fundamentally different user model.

## Sub-Resources

Load the reference that matches your task. When in doubt, load multiple references.

| Keywords | Reference |
|----------|----------|
| cli, command, subcommand, flag, help, usage, argument, jj log, jj rebase, jj squash, jj commit, jj new, jj describe, jj edit, jj abandon, jj split, jj diff, jj show, jj status, jj bookmark, jj tag, jj file, jj git, jj op, jj workspace, jj config, jj resolve, jj next, jj prev, jj duplicate, jj restore, jj revert, jj parallelize, jj fix, jj diffedit, jj interdiff, jj metaedit, jj sign, jj unsign, jj arrange, jj evolog, jj bisect, jj sparse, jj gerrit, jj util, REVSET, FILESET, template flag | [references/cli.md](references/cli.md) |
| concept, model, working copy, change id, commit id, immutable, conflict, rebase, descendant, root commit, stash, index, staging, auto-track, stale, visible, hidden, evil merge, workspace, symbol resolution, no active branch, detached HEAD | [references/concepts.md](references/concepts.md) |
| bookmark, branch, remote bookmark, tracking, bookmark conflict, push safety, jj bookmark create, jj bookmark move, jj bookmark delete, jj bookmark forget, jj bookmark track, jj bookmark untrack, jj bookmark advance, jj bookmark rename, jj bookmark list, jj git push | [references/bookmarks.md](references/bookmarks.md) |
| revset, revset expression, revset function, revset operator, select revisions, ancestors, descendants, parents, children, union, intersection, difference, range, dag range, pattern, string pattern, date pattern, revset alias, grammar, BUILTIN_FUNCTION_MAP | [references/revsets.md](references/revsets.md) |
| fileset, fileset expression, fileset function, fileset operator, select files, file pattern, glob, cwd, root, prefix-glob, bare string, bare string pattern, strict identifier, grammar | [references/filesets.md](references/filesets.md) |
| template, -T, --template, template language, format, output format, custom log, template alias, method, global function, concat, label, fill, indent, if, coalesce, Commit, ChangeId, CommitId, String, List, Timestamp, Signature, TreeDiff, Operation, color label | [references/templates.md](templates.md) |
| operation, op, undo, redo, at-op, operation log, op log, op show, op diff, op restore, op revert, op abandon, op integrate, lock-free concurrency, operation id | [references/operations.md](references/operations.md) |
| conflict, resolve, conflict marker, conflict style, diff style, snapshot style, git style, long conflict marker, missing terminating newline, merge conflict, conflict resolution, jj resolve | [references/conflicts.md](references/conflicts.md) |
| divergent, divergence, divergent change, change offset, change id suffix, metaedit change id, resolve divergence, concurrent modification | [references/divergence.md](references/divergence.md) |
| config, configuration, settings, toml, config edit, config set, config list, config path, config get, config unset, user config, repo config, workspace config, conditional config, revset-aliases, fileset-aliases, template-aliases, colors, signing, merge-tools, fix tools, snapshot, fsmonitor | [references/config.md](references/config.md) |
| git, vs git, git comparison, git equivalent, git to jj, colocate, colocation, git compatibility, migration, git init, git clone, git fetch, git push, git export, git import, command equivalents | [references/git-compat.md](references/git-compat.md) |
| github, gitlab, gerrit, forge, push, pr, pull request, merge request, code review, gerrit upload, remote, upstream, push options, multi-remote, Change-Id, Link trailer | [references/forge-workflows.md](references/forge-workflows.md) |

## Quick Guide

- **How do I use jj commands?** → [references/cli.md](references/cli.md)
- **How does jj differ from Git conceptually?** → [references/concepts.md](references/concepts.md) and [references/git-compat.md](references/git-compat.md)
- **How do bookmarks work?** → [references/bookmarks.md](references/bookmarks.md)
- **How do I write revset expressions?** → [references/revsets.md](references/revsets.md)
- **How do I write fileset expressions?** → [references/filesets.md](references/filesets.md)
- **How do I customize output with templates?** → [references/templates.md](references/templates.md)
- **How do I undo/redo operations?** → [references/operations.md](references/operations.md)
- **How do conflicts work in jj?** → [references/conflicts.md](references/conflicts.md)
- **How do I resolve divergent changes?** → [references/divergence.md](references/divergence.md)
- **How do I configure jj?** → [references/config.md](references/config.md)
- **How do I migrate from Git?** → [references/git-compat.md](references/git-compat.md)
- **How do I work with GitHub/GitLab/Gerrit?** → [references/forge-workflows.md](references/forge-workflows.md)
- **What are the Git command equivalents?** → [references/git-compat.md](references/git-compat.md)
- **How do I push bookmarks safely?** → [references/bookmarks.md](references/bookmarks.md)
- **What revset functions filter by author?** → [references/revsets.md](references/revsets.md)
- **What template methods are available on Commit?** → [references/templates.md](references/templates.md)

## Cross-Project References

For carapace shell completion integration with jj (revset/fileset/template completion actions), see the **carapace** and **carapace-dev** skills (in the carapace and carapace-bin repos).
