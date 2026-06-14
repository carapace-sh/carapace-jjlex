package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace/pkg/style"
)

// ActionRevFiles completes files at a given revision.
//
//	go.mod
//	go.sum
func ActionRevFiles(revision string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		if revision == "" {
			revision = "@"
		}
		return actionExecJJ("file", "list", "--revision", revision)(func(output []byte) carapace.Action {
			lines := strings.Split(string(output), "\n")
			return carapace.ActionValues(lines[:len(lines)-1]...).MultiParts("/").StyleF(style.ForPathExt).Tag("files")
		})
	}).UidF(Uid("file"))
}
