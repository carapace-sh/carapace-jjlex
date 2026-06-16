package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/actions/tools/jj"
	spec "github.com/carapace-sh/carapace-spec"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "carapace-jjlex",
	Short: "Parse jj revset, fileset, and template expressions",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	carapace.Gen(rootCmd)

	// TODO add other actions as macros
	spec.AddMacroI(jj.ActionAncestors)
	spec.AddMacroI(jj.ActionDescendants)
	spec.Register(rootCmd)
}
