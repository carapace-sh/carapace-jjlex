---
name: jj
description: >
  Main entrypoint for jj (Jujutsu VCS) skills. Loads and refers to focused
  sub-skills covering specific jj topics. Triggers on: "jj", "jujutsu",
  "jj vcs", "jj version control".
user-invocable: true
---

# jj — Jujutsu VCS Skill Index

This is the main entrypoint for jj (Jujutsu VCS) knowledge. When a user asks
about jj, determine which sub-skill(s) are relevant and load them by their name.

## Sub-skills

| Skill | Name | When to load |
|-------|------|-------------|
| **jj-cli** | `jj-cli` | Commands, subcommands, flags, argument types, help output |
| **jj-concepts** | `jj-concepts` | Working-copy model, change IDs, commit IDs, rebasing, immutable revisions, conflicts, root commit |
| **jj-bookmarks** | `jj-bookmarks` | Bookmarks (branches), remote bookmarks, tracking, bookmark conflicts, push safety |
| **jj-config** | `jj-config` | Config file locations, layered precedence, sections, conditional config, config CLI commands |
| **jj-revsets** | `jj-revsets` | Revset expressions: symbols, operators, built-in functions, string/date patterns, aliases |
| **jj-filesets** | `jj-filesets` | Fileset expressions: operators, pattern kinds, built-in functions, bare strings, grammar |
| **jj-templates** | `jj-templates` | Template language: operators, types, methods, global functions, output formatting |
| **jj-operations** | `jj-operations` | Operation log: undo/redo, --at-op, lock-free concurrency |
| **jj-git-compat** | `jj-git-compat` | Git interoperability: differences, command equivalents, colocated repos, migration |

## How to use

1. When the user asks a jj-related question, identify the relevant sub-skill(s).
2. Load the sub-skill by calling `view` on its skill name (e.g. `jj-cli`, `jj-concepts`).
3. Follow the instructions in the loaded sub-skill.
4. If the question spans multiple topics, load multiple sub-skills as needed.
5. If unsure which sub-skill applies, start with **jj-concepts** for conceptual
   questions or **jj-cli** for command-related questions.
