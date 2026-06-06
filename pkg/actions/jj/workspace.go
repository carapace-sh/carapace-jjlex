package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
)

// ActionWorkspaces completes workspace names with their current commit descriptions.
//
//	default (qzzmpvmx 2ceef6bf (no description set))
//	another (oxtpukyp 00a745c4 (empty) (no description set))
func ActionWorkspaces() carapace.Action {
	return actionExecJJ("workspace", "list")(func(output []byte) carapace.Action {
		lines := strings.Split(string(output), "\n")

		vals := make([]string, 0)
		for _, line := range lines[:len(lines)-1] {
			vals = append(vals, strings.SplitN(line, ": ", 2)...)
		}
		if len(vals) == 0 {
			return carapace.ActionValues()
		}
		return carapace.ActionValuesDescribed(vals...).Tag("workspaces")
	}).UidF(Uid("workspace"))
}
