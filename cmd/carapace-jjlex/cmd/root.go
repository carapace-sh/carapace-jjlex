package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-bin/pkg/actions/tools/jj"
	jjlex "github.com/carapace-sh/carapace-jjlex"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "carapace-jjlex revset",
	Long: "simple jujutsu revset lexer",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := jjlex.Split(args[0])

		var output any = ctx
		if cmd.Flag("allowed-operators").Changed {
			output = ctx.AllowedOperators()
		}

		m, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(m))
		return nil
	},
}

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Bool("allowed-operators", false, "list of allowed operators at this point")
	carapace.Gen(rootCmd).PositionalCompletion(
		jj.ActionRevsets(jj.RevOption{}.Default()),
	)
}
