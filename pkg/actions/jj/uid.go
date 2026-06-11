package jj

import (
	"fmt"
	"net/url"

	"github.com/carapace-sh/carapace/pkg/traverse"
	"github.com/carapace-sh/carapace/pkg/uid"
)

// uidWithPrefix is like Uid but prepends prefix to the value before using it
// as the URL path. This is needed for postfix operator completions where the
// value is just the operator suffix (e.g. "--") but the UID path should
// contain the full revset expression (e.g. "main--").
func uidWithPrefix(host string, prefix string, opts ...string) func(s string, uc uid.Context) (*url.URL, error) {
	return func(s string, uc uid.Context) (*url.URL, error) {
		return Uid(host, opts...)(prefix+s, uc)
	}
}

func Uid(host string, opts ...string) func(s string, uc uid.Context) (*url.URL, error) {
	return func(s string, uc uid.Context) (*url.URL, error) {
		if length := len(opts); length%2 != 0 {
			return nil, fmt.Errorf("invalid amount of arguments [jj.Uid]: %v", length)
		}

		repository, ok := uc.LookupEnv(EnvRepository)
		if !ok {
			var err error
			repository, err = traverse.Parent(".jj")(uc)
			if err != nil {
				return nil, err
			}
		}

		uid := &url.URL{
			Scheme: "jj",
			Host:   host,
			Path:   s,
		}
		values := uid.Query()
		if operation, ok := uc.LookupEnv(EnvOperation); ok {
			values.Add(EnvOperation, operation)
		}
		values.Add(EnvRepository, repository)
		for i := 0; i < len(opts); i += 2 {
			if opts[i+1] != "" {
				values.Add(opts[i], opts[i+1])
			}
		}
		uid.RawQuery = values.Encode()

		return uid, nil
	}
}
