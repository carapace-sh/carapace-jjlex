package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/carapace-sh/carapace"
	"github.com/carapace-sh/carapace-jjlex/pkg/actions/jj"
	"github.com/carapace-sh/carapace-jjlex/pkg/revset"
	"github.com/spf13/cobra"
)

var revsetCmd = &cobra.Command{
	Use:   "revset <expression>",
	Short: "Parse a revset expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expression, err := revset.Parse(args[0])
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

var revsetCompleteCmd = &cobra.Command{
	Use:   "revset-complete <cursor> <expression>",
	Short: "Get completion context for a revset expression",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cursor, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid cursor: %v", err)
		}
		ctx := revset.ParseForCompletion(args[1], cursor)
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(revsetCmd)
	rootCmd.AddCommand(revsetCompleteCmd)

	carapace.Gen(revsetCmd).PositionalCompletion(
		jj.ActionRevsets(jj.RevOpts{}.Default()),
	)

	carapace.Gen(revsetCompleteCmd).PositionalCompletion(
		jj.ActionRevsets(jj.RevOpts{}.Default()),
	)
}
