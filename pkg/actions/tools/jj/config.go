package jj

import (
	"strconv"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-bridge/pkg/actions/bridge"
	"github.com/carapace-sh/carapace/pkg/style"
	"github.com/pelletier/go-toml/v2"
)

// ActionConfigs completes jj config keys with their current values as descriptions.
//
//	ui.color (never)
//	ui.editor (hx)
func ActionConfigs(includeDefaults bool) carapace.Action {
	if !includeDefaults {
		return actionConfigs(false)
	}
	return carapace.Batch(
		actionConfigs(true),
		actionConfigs(false).Style(style.Blue),
	).ToA()
}

func actionConfigs(includeDefaults bool) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		args := []string{"config", "list"}
		if includeDefaults {
			args = append(args, "--include-defaults")
		}
		return actionExecJJ(args...)(func(output []byte) carapace.Action {
			var config map[string]any
			if err := toml.Unmarshal(output, &config); err != nil {
				return carapace.ActionMessage(err.Error())
			}

			vals := make([]string, 0)
			for key, value := range flattenConfig(config) {
				switch {
				case strings.Contains(value, "\n"):
					vals = append(vals, key, "")
				default:
					vals = append(vals, key, value)
				}
			}
			return carapace.ActionValuesDescribed(vals...)
		})
	}).UidF(Uid("config"))
}

func flattenConfig(m map[string]any) map[string]string {
	flattened := make(map[string]string)
	for key, value := range m {
		switch v := value.(type) {
		case bool:
			flattened[key] = strconv.FormatBool(v)
		case string:
			flattened[key] = v
		case int:
			flattened[key] = strconv.Itoa(v)
		case map[string]any:
			for k, val := range flattenConfig(v) {
				flattened[key+"."+k] = val
			}
		default:
			flattened[key] = ""
		}
	}
	return flattened
}

// ActionConfigValues completes config values for the given config key.
//
//	auto
//	true
func ActionConfigValues(config string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		_bool := carapace.ActionValues("true", "false").StyleF(style.ForKeyword)
		if a, ok := (carapace.ActionMap{
			"diff.color-words.conflict":                    carapace.ActionValues("materialize", "pair").StyleF(style.ForKeyword),
			"diff.color-words.context":                     carapace.ActionValues().Usage("number of context lines"),
			"diff.color-words.max-inline-alternation":      carapace.ActionValues().Usage("max alternation before falling back to full diff, -1 for unlimited"),
			"diff.git.context":                             carapace.ActionValues().Usage("number of context lines"),
			"diff.git.show-path-prefix":                    _bool,
			"diff.stat.context":                            carapace.ActionValues().Usage("number of context lines"),
			"fsmonitor.backend":                            carapace.ActionValues("none", "watchman").StyleF(style.ForKeyword),
			"fsmonitor.watchman.register-snapshot-trigger": _bool,
			"gerrit.default-remote":                        ActionRemotes(),
			"git.abandon-unreachable-commits":              _bool,
			"git.colocate":                                 _bool,
			"git.fetch":                                    ActionRemotes(),
			"git.object-hash":                              carapace.ActionValues("sha1", "sha256"),
			"git.private-commits":                          ActionRevsets(RevOpts{}.Default()),
			"git.push":                                     ActionRemotes(),
			"git.record-synthetic-predecessors":            _bool,
			"git.sign-on-push":                             _bool,
			"git.track-default-bookmark-on-clone":          _bool,
			"git.write-change-id-header":                   _bool,
			"hints.<name>":                                 _bool,
			"merge.hunk-level":                             carapace.ActionValues("line", "word"),
			"merge.same-change":                            carapace.ActionValues("keep", "accept"),
			"signing.backend":                              carapace.ActionValues("none", "gpg", "gpgsm", "ssh").StyleF(style.ForKeyword),
			"signing.backends.gpg.allow-expired-keys":      _bool,
			"signing.backends.gpg.program":                 bridge.ActionCarapaceBin().Split(),
			"signing.backends.gpgsm.allow-expired-keys":    _bool,
			"signing.backends.gpgsm.program":               bridge.ActionCarapaceBin().Split(),
			"signing.backends.ssh.allowed-signers":         carapace.ActionFiles(),
			"signing.backends.ssh.program":                 bridge.ActionCarapaceBin().Split(),
			"signing.backends.ssh.revocation-list":         carapace.ActionFiles(),
			"signing.behavior":                             carapace.ActionValues("drop", "keep", "own", "force").StyleF(style.ForKeyword),
			"signing.key":                                  ActionSigningKeys(),
			"snapshot.auto-track":                          ActionFilesets(),
			"snapshot.auto-update-stale":                   _bool,
			"snapshot.max-new-file-size":                   carapace.ActionValues().Usage("size in bytes or human-readable (e.g. 1MiB)"),
			"split.legacy-bookmark-behavior":               _bool,
			"ui.bookmark-list-sort-keys":                   actionConfigSortKeys(),
			"ui.color":                                     carapace.ActionValues("auto", "always", "never", "debug").StyleF(style.ForKeyword),
			"ui.conflict-marker-style":                     actionConfigConflictMarkerStyles(),
			"ui.default-command":                           bridge.ActionCarapaceBin("jj").Split(),
			"ui.diff-editor": carapace.Batch(
				carapace.ActionValues(":builtin", ":ours", ":theirs", "diffedit3", "diffedit3-ssh", "meld", "meld-3", "vimdiff").Prefix(":").NoSpace(),
				carapace.ActionValues("diffedit3", "diffedit3-ssh", "meld", "meld-3", "vimdiff"),
				bridge.ActionCarapaceBin().Split(),
			).ToA(),
			"ui.diff-formatter": carapace.Batch(
				carapace.ActionValues(":color-words", ":git", ":summary", ":stat", ":types", ":name-only").Prefix(":").NoSpace(),
				bridge.ActionCarapaceBin().Split(),
			).ToA(),
			"ui.diff-instructions":          _bool,
			"ui.graph.style":                carapace.ActionValues("curved", "square", "ascii", "ascii-large").StyleF(style.ForKeyword),
			"ui.log-synthetic-elided-nodes": _bool,
			"ui.log-word-wrap":              _bool,
			"ui.merge-editor": carapace.Batch(
				carapace.ActionValues(":builtin", ":ours", ":theirs", "kdiff3", "meld", "mergiraf", "smerge", "vimdiff", "vscode", "vscodium").Prefix(":").NoSpace(),
				carapace.ActionValues("kdiff3", "meld", "mergiraf", "smerge", "vimdiff", "vscode", "vscodium"),
				bridge.ActionCarapaceBin().Split(),
			).ToA(),
			"ui.movement.edit": _bool,
			"ui.paginate":      carapace.ActionValues("auto", "never").StyleF(style.ForKeyword),
			"ui.pager": carapace.Batch(
				carapace.ActionValues(":builtin"),
				bridge.ActionCarapaceBin().Split(),
			).ToA(),
			"ui.progress-indicator":            _bool,
			"ui.quiet":                         _bool,
			"ui.show-cryptographic-signatures": _bool,
			"ui.streampager.interface":         carapace.ActionValues("quit-if-one-page", "full-screen-clear-output", "quit-quickly-or-clear-output"),
			"ui.streampager.show-ruler":        _bool,
			"ui.streampager.wrapping":          carapace.ActionValues("anywhere", "word", "none"),
			"ui.tag-list-sort-keys":            actionConfigSortKeys(),
			"working-copy.eol-conversion":      carapace.ActionValues("none", "input", "input-output"),
			"working-copy.exec-bit-change":     carapace.ActionValues("respect", "ignore", "auto").StyleF(style.ForKeyword),
		}[config]); ok {
			return a
		}

		splitted := strings.Split(config, ".")
		last := splitted[len(splitted)-1]
		switch splitted[0] {
		case "aliases", "ui.default-command":
			return bridge.ActionCarapaceBin("jj").Split()
		case "fix":
			if splitted[1] == "tools" {
				switch last {
				case "command":
					return bridge.ActionCarapaceBin().Split()
				case "diff-do-chdir", "enabled", "run-tool-if-zero-line-ranges":
					return _bool
				case "patterns":
					return ActionFilesets()
				default:
					return carapace.ActionValues().Usage("template with $first and $last variables")
				}
			}
		case "experimental-advance-branches":
			return ActionStringPatterns()
		case "fileset-aliases":
			return ActionFilesets()
		case "hints":
			return _bool
		case "merge-tools":
			switch last {
			case "conflict-marker-style":
				return actionConfigConflictMarkerStyles()
			case "diff-invocation-mode", "edit-invocation-mode":
				return carapace.ActionValues("dir", "file-by-file").StyleF(style.ForKeyword)
			case "merge-tool-edits-conflict-markers":
				return _bool
			case "program":
				return carapace.ActionFiles()
			default:
				return carapace.ActionValues().Usage("command arguments")
			}
		case "remotes":
			switch last {
			case "fetch-bookmarks", "fetch-tags", "auto-track-bookmarks", "auto-track-created-bookmarks":
				return ActionStringPatterns()
			}
		case "revset-aliases", "revsets":
			return ActionRevsets(RevOpts{}.Default())
		case "template-aliases", "templates":
			return ActionTemplates()
		}
		return carapace.ActionValues()
	})
}

func actionConfigConflictMarkerStyles() carapace.Action {
	return carapace.ActionValues("diff", "diff-experimental", "snapshot", "git").StyleF(style.ForKeyword)
}

func actionConfigSortKeys() carapace.Action {
	return carapace.ActionValuesDescribed(
		"name", "bookmark or tag name",
		"name-", "bookmark or tag name, descending",
		"author-name", "author name",
		"author-name-", "author name, descending",
		"author-email", "author email",
		"author-email-", "author email, descending",
		"author-date", "author date",
		"author-date-", "author date, descending",
		"committer-name", "committer name",
		"committer-name-", "committer name, descending",
		"committer-email", "committer email",
		"committer-email-", "committer email, descending",
		"committer-date", "committer date",
		"committer-date-", "committer date, descending",
	)
}
