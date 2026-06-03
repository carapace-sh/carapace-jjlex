package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/carapace-sh/carapace-jjlex/pkg/fileset"
	"github.com/spf13/cobra"
)

var filesetCmd = &cobra.Command{
	Use:   "fileset <expression>",
	Short: "Parse a fileset expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expression, err := fileset.Parse(args[0])
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

var filesetCompleteCmd = &cobra.Command{
	Use:   "fileset-complete <expression>",
	Short: "Get completion context for a fileset expression",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := fileset.ParseForCompletion(args[0])
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

var filesetBareCmd = &cobra.Command{
	Use:   "fileset-bare <expression>",
	Short: "Parse a fileset expression with bare string fallback",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		expression, err := fileset.ParseProgramOrBareString(args[0])
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

var filesetBareCompleteCmd = &cobra.Command{
	Use:   "fileset-bare-complete <expression>",
	Short: "Get completion context for a fileset expression with bare string fallback",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := fileset.ParseForCompletion(args[0])
		m, err := json.Marshal(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(m))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(filesetCmd)
	rootCmd.AddCommand(filesetCompleteCmd)
	rootCmd.AddCommand(filesetBareCmd)
	rootCmd.AddCommand(filesetBareCompleteCmd)
}
