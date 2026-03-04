package cmd

import (
	"gcnf/internal/googleapi"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two environments in a sheet",
	Long:  `Load data for two environments from the same sheet and display the differences.`,
	Example: `  gcnf diff --sheet Env --env1 staging --env2 production
  gcnf diff -s Env --env1 dev --env2 staging`,
	Run: func(cmd *cobra.Command, args []string) {
		sheet, _ := cmd.Flags().GetString("sheet")
		env1, _ := cmd.Flags().GetString("env1")
		env2, _ := cmd.Flags().GetString("env2")
		checkRequirements()
		googleapi.DiffCommand(sheet, env1, env2, configs)
	},
}

func init() {
	diffCmd.Flags().StringP("sheet", "s", "", "Sheet name")
	diffCmd.Flags().String("env1", "", "First environment")
	diffCmd.Flags().String("env2", "", "Second environment")
	diffCmd.MarkFlagRequired("sheet")
	diffCmd.MarkFlagRequired("env1")
	diffCmd.MarkFlagRequired("env2")
}
