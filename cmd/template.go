package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/actions/jj"
	"github.com/carapace-sh/carapace-jjlex/pkg/template"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template <expression>",
	Short: "Parse a template expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expression, err := template.Parse(args[0])
		if err != nil {
			return err
		}
		m, err := json.Marshal(expression)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

var templateCompleteCmd = &cobra.Command{
	Use:   "template-complete <expression>",
	Short: "Get completion context for a template expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := template.ParseForCompletion(args[0])
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(templateCompleteCmd)

	carapace.Gen(templateCmd).PositionalCompletion(
		jj.ActionTemplates(),
	)

	carapace.Gen(templateCompleteCmd).PositionalCompletion(
		jj.ActionTemplates(),
	)
}
