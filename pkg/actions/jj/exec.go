package jj

import (
	"github.com/carapace-sh/carapace"
)

func actionExecJJ(arg ...string) func(func(output []byte) carapace.Action) carapace.Action {
	return func(f func(output []byte) carapace.Action) carapace.Action {
		return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			args := []string{"--color", "never"}
			if repository, ok := c.LookupEnv("JJ_REPOSITORY"); ok {
				args = append(args, "--repository", repository)
			}
			if operation, ok := c.LookupEnv("JJ_OPERATION"); ok {
				args = append(args, "--at-operation", operation)
			}
			args = append(args, arg...)
			return carapace.ActionExecCommand("jj", args...)(func(output []byte) carapace.Action {
				return f(output)
			})
		})
	}
}
