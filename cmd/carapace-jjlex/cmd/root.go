package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-bin/pkg/actions/tools/jj"
	"github.com/carapace-sh/carapace-jjlex"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "carapace-jjlex",
	Long: "simple jujutsu revset lexer",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := jjlex.Split(args[0])
		m, err := json.MarshalIndent(ctx, "", "  ")
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
	carapace.Gen(rootCmd).PositionalCompletion(
		carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			batch := carapace.Batch()
			ctx := jjlex.Split(c.Value)
			switch ctx.Type {
			case jjlex.CompletionTypeOperator:
				// TODO
			case jjlex.CompletionTypeFunctionArg:
				// TODO complete corresponding type (e.g. lexer should return revision)
				batch = append(batch,
					jj.ActionRevs(jj.RevOption{}.Default()),
					jj.ActionRevsetFunctions().Suffix("("),
				)
			case jjlex.CompletionTypeRevision:
				batch = append(batch,
					jj.ActionRevs(jj.RevOption{}.Default()),
					jj.ActionRevsetFunctions().Suffix("("),
				)
			}
			return batch.ToA().Prefix(strings.TrimSuffix(ctx.FullInput, ctx.Prefix)).NoSpace()
		}),
	)
}
