package jj

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-bridge/pkg/actions/bridge"
)

// ActionSigningKeys completes signing keys based on the user's configuration.
//
//	ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... (/home/user/.ssh/id_ed25519)
//	ABCDEF1234567890 (some GPG key)
func ActionSigningKeys() carapace.Action {
	return actionExecJJ("config", "get", "signing.backend")(func(output []byte) carapace.Action {
		backend := strings.TrimSpace(string(output))

		switch backend {
		case "ssh":
			return bridge.ActionMacro("carapace", "net.ssh.SigningKeys")
		case "gpg":
			return bridge.ActionMacro("carapace", "os.GpgKeyIds")
		default:
			return carapace.ActionValues()
		}
	})
}
