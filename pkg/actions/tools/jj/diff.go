package jj

import (
	"net/url"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/style"
	"github.com/carapace-sh/carapace/pkg/uid"
)

// ActionRevDiffs completes changed files between revisions.
// Accepts up to two revision arguments:
//
//   - 0: compare working copy to parent
//
//   - 1: compare given revision to its parent
//
//   - 2: compare first revision to second revision
//
//     go.mod (M)
//     go.sum (A)
func ActionRevDiffs(revisions ...string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		var from, to string
		switch len(revisions) {
		case 0:
			from = "@"
			to = "@-"
		case 1:
			from = revisions[0]
			to = revisions[0] + "-"
		case 2:
			from = revisions[0]
			to = revisions[1]
		default:
			return carapace.ActionMessage("ActionRevDiffs: at most 2 revision arguments")
		}

		return actionExecJJ("diff", "--summary", "--from", from, "--to", to)(func(output []byte) carapace.Action {
			lines := strings.Split(string(output), "\n")

			vals := make([]string, 0)
			for _, line := range lines[:len(lines)-1] {
				if splitted := strings.SplitN(line, " ", 2); splitted != nil {
					vals = append(vals, splitted[1], splitted[0])
				}
			}
			a := carapace.ActionValuesDescribed(vals...)
			if len(revisions) > 1 {
				a = a.MultiParts("/")
			}
			return a.StyleF(style.ForPathExt)
		}).Tag("rev diffs").UidF(func(s string, uc uid.Context) (*url.URL, error) {
			return Uid("rev-diff", "from", from, "to", to)(s, uc)
		})
	})
}

// ActionRevChanges completes files changed in given revisions with add/remove status.
//
//	go.mod (M)
//	new_file.rs (A)
func ActionRevChanges(revisions ...string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		args := []string{"log", "--summary", "--no-graph", "--template", ""}
		for _, revision := range revisions {
			args = append(args, "--revisions", revision)
		}
		return actionExecJJ(args...)(func(output []byte) carapace.Action {
			lines := strings.Split(string(output), "\n")

			vals := make([]string, 0)
			for _, line := range lines[:len(lines)-1] {
				if splitted := strings.SplitN(line, " ", 2); splitted != nil {
					vals = append(vals, splitted[1], splitted[0])
				}
			}
			return carapace.ActionValuesDescribed(vals...).MultiParts("/").StyleF(style.ForPathExt)
		})
	}).Tag("rev changes").UidF(func(s string, uc uid.Context) (*url.URL, error) {
		opts := make([]string, 0, len(revisions)*2)
		for _, rev := range revisions {
			opts = append(opts, "revision", rev)
		}
		return Uid("changes", opts...)(s, uc)
	})
}
