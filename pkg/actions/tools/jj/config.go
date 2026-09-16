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
			"ui.bookmark-list-sort-keys":                   actionConfigTOMLArray(actionConfigSortKeys()),
			"ui.color":                                     carapace.ActionValues("auto", "always", "never", "debug").StyleF(style.ForKeyword),
			"ui.conflict-marker-style":                     actionConfigConflictMarkerStyles(),
			"ui.default-command":                           actionConfigTOMLArray(bridge.ActionCarapaceBin("jj")),
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
			"ui.tag-list-sort-keys":            actionConfigTOMLArray(actionConfigSortKeys()),
			"working-copy.eol-conversion":      carapace.ActionValues("none", "input", "input-output"),
			"working-copy.exec-bit-change":     carapace.ActionValues("respect", "ignore", "auto").StyleF(style.ForKeyword),
		}[config]); ok {
			return a
		}

		splitted := strings.Split(config, ".")
		last := splitted[len(splitted)-1]
		switch splitted[0] {
		case "aliases":
			if last == "doc" {
				return carapace.ActionValues().Usage("documentation string")
			}
			return actionConfigTOMLArray(bridge.ActionCarapaceBin("jj"))
		case "ui.default-command":
			return actionConfigTOMLArray(bridge.ActionCarapaceBin("jj"))
		case "fix":
			if splitted[1] == "tools" {
				switch last {
				case "command":
					return actionConfigTOMLArray(bridge.ActionCarapaceBin())
				case "diff-do-chdir", "enabled", "run-tool-if-zero-line-ranges":
					return _bool
				case "patterns":
					return ActionFilesets()
				default:
					return carapace.ActionValues().Usage("template with $first and $last variables")
				}
			}
		case "experimental-advance-branches":
			return actionConfigTOMLArray(ActionStringPatterns())
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
			case "diff-args":
				return actionConfigTOMLArray(carapace.Batch(
					carapace.ActionValues("$left", "$right", "$path", "$marker_length", "$width"),
					bridge.ActionCarapaceBin(),
				).ToA())
			case "edit-args":
				return actionConfigTOMLArray(carapace.Batch(
					carapace.ActionValues("$left", "$right"),
					bridge.ActionCarapaceBin(),
				).ToA())
			case "merge-args":
				return actionConfigTOMLArray(carapace.Batch(
					carapace.ActionValues("$base", "$left", "$right", "$output"),
					bridge.ActionCarapaceBin(),
				).ToA())
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

// actionConfigTOMLArray completes a one-line TOML array of strings.
//
//	["log", "-r", "main"]
func actionConfigTOMLArray(elements carapace.Action) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		value := c.Value
		switch {
		case value == "" || value == "[":
			return carapace.ActionValues(`["`).NoSpace()
		case !strings.HasPrefix(value, `["`):
			return carapace.ActionValues()
		}

		// parse the partial value by closing it with various suffixes
		var completed []string
		open, found := false, false
		for _, suffix := range []string{"", `]`, `"]`, `""]`} {
			var parsed map[string]any
			if err := toml.Unmarshal([]byte(`v = `+value+suffix), &parsed); err != nil {
				continue
			}
			array, ok := parsed["v"].([]any)
			if !ok || len(array) == 0 {
				continue
			}
			elems := make([]string, 0, len(array))
			for _, elem := range array {
				s, ok := elem.(string)
				if !ok {
					elems = nil
					break
				}
				elems = append(elems, s)
			}
			if elems == nil {
				continue
			}
			completed = elems[:len(elems)-1]
			open, found = false, true
			switch suffix {
			case "":
			case `]`:
				if strings.HasSuffix(strings.TrimRight(value, " \t"), `,`) { // trailing comma: next element expected
					completed = elems
					open = true
				}
			default:
				open = true
			}
			break
		}
		if !found || !open {
			return carapace.ActionValues()
		}

		quote := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
		prefix := `[`
		for i, elem := range completed {
			if i > 0 {
				prefix += `, `
			}
			prefix += `"` + quote.Replace(elem) + `"`
		}
		if open {
			if len(completed) > 0 {
				prefix += `, `
			}
			prefix += `"`
		}

		c.Args = completed // previous elements as context for the completion

		// Strip the prefix from c.Value so the elements action sees just the
		// partial element being completed. We invoke elements directly (rather
		// than using Action.Prefix/Suffix) so that c.Args is correctly updated
		// for bridge actions that rely on it.
		switch {
		case strings.HasPrefix(c.Value, prefix):
			c.Value = c.Value[len(prefix):]
		case strings.HasPrefix(prefix, c.Value):
			c.Value = ""
		default:
			return carapace.ActionValues()
		}
		return elements.Invoke(c).Prefix(prefix).Suffix(`"`).ToA().NoSpace()
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
