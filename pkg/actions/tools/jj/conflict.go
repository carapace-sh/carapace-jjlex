package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
)

// ActionConflicts completes conflicted files at a given revision.
//
//	fileA
//	fileB
func ActionConflicts(revision string) carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		if revision == "" {
			revision = "@"
		}
		return actionExecJJE("resolve", "--list", "--revision", revision)(func(output []byte, err error) carapace.Action {
			if err != nil {
				return carapace.ActionValues()
			}
			lines := strings.Split(string(output), "\n")

			vals := make([]string, 0)
			for _, line := range lines[:len(lines)-1] {
				parts := strings.SplitN(line, "    ", 2)
				vals = append(vals, parts[0])
				if len(parts) > 1 {
					vals = append(vals, parts[1])
				} else {
					vals = append(vals, "")
				}
			}
			if len(vals) == 0 {
				return carapace.ActionValues()
			}
			return carapace.ActionValuesDescribed(vals...).Tag("conflicts")
		})
	}).UidF(Uid("conflict"))
}
