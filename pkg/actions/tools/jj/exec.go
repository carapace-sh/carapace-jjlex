package jj

import (
	"github.com/carapace-sh/carapace"
)

// Environment variable names for contextual completion.
// These are not official jj environment variables — they are conventions
// used by carapace-sh completers to scope completions to a specific
// repository and operation when invoking `jj` commands.
const (
	// EnvRepository overrides the repository path passed to `jj --repository`.
	EnvRepository = "JJ_REPOSITORY"
	// EnvOperation overrides the operation passed to `jj --at-operation`.
	EnvOperation = "JJ_OPERATION"
)

func actionExecJJ(arg ...string) func(func(output []byte) carapace.Action) carapace.Action {
	return func(f func(output []byte) carapace.Action) carapace.Action {
		return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			args := []string{"--color", "never"}
			if repository, ok := c.LookupEnv(EnvRepository); ok {
				args = append(args, "--repository", repository)
			}
			if operation, ok := c.LookupEnv(EnvOperation); ok {
				args = append(args, "--at-operation", operation)
			}
			args = append(args, arg...)
			return carapace.ActionExecCommand("jj", args...)(func(output []byte) carapace.Action {
				return f(output)
			})
		})
	}
}

func actionExecJJE(arg ...string) func(func(output []byte, err error) carapace.Action) carapace.Action {
	return func(f func(output []byte, err error) carapace.Action) carapace.Action {
		return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			args := []string{"--color", "never"}
			if repository, ok := c.LookupEnv(EnvRepository); ok {
				args = append(args, "--repository", repository)
			}
			args = append(args, arg...)
			return carapace.ActionExecCommandE("jj", args...)(func(output []byte, err error) carapace.Action {
				return f(output, err)
			})
		})
	}
}
