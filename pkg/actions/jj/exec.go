package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
)

func actionExecJJ(args ...string) func(func(output []byte) carapace.Action) carapace.Action {
	return func(f func(output []byte) carapace.Action) carapace.Action {
		return carapace.ActionExecCommand("jj", args...)(func(output []byte) carapace.Action {
			return f(output)
		})
	}
}

func actionExecJJE(args ...string) func(func(output []byte, err error) carapace.Action) carapace.Action {
	return func(f func(output []byte, err error) carapace.Action) carapace.Action {
		return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			cmd := c.Command("jj", args...)
			cmd.Stdin = strings.NewReader("")
			output, err := cmd.CombinedOutput()
			return f(output, err)
		})
	}
}
