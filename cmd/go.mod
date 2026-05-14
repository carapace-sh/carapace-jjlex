module github.com/carapace-sh/carapace-jjlex/cmd

go 1.26.1

require (
	github.com/carapace-sh/carapace v1.11.4
	github.com/carapace-sh/carapace-bin v1.6.5
	github.com/carapace-sh/carapace-jjlex v1.1.1
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/carapace-sh/carapace-shlex v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kevinburke/ssh_config v1.4.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/mod v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/carapace-sh/carapace-jjlex => ../

replace github.com/kevinburke/ssh_config => github.com/carapace-sh/ssh_config v1.4.1-0.20260319075335-4f04016b8b4b

replace github.com/carapace-sh/carapace-bin => ../../carapace-bin/
