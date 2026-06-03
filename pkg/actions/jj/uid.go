package jj

import (
	"fmt"
	"net/url"

	"github.com/carapace-sh/carapace/pkg/traverse"
	"github.com/carapace-sh/carapace/pkg/uid"
)

func Uid(host string, opts ...string) func(s string, uc uid.Context) (*url.URL, error) {
	return func(s string, uc uid.Context) (*url.URL, error) {
		if length := len(opts); length%2 != 0 {
			return nil, fmt.Errorf("invalid amount of arguments [jj.Uid]: %v", length)
		}

		repository, ok := uc.LookupEnv("JJ_REPOSITORY")
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
		if operation, ok := uc.LookupEnv("JJ_OPERATION"); ok {
			values.Add("JJ_OPERATION", operation)
		}
		values.Add("JJ_REPOSITORY", repository)
		for i := 0; i < len(opts); i += 2 {
			if opts[i+1] != "" {
				values.Add(opts[i], opts[i+1])
			}
		}
		uid.RawQuery = values.Encode()

		return uid, nil
	}
}
