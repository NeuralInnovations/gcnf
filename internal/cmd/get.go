package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
)

// Get Command
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get a specific value in the loaded data",
	Run: func(cmd *cobra.Command, args []string) {
		sheet, _ := cmd.Flags().GetString("sheet")
		env, _ := cmd.Flags().GetString("env")
		category, _ := cmd.Flags().GetString("category")
		name, _ := cmd.Flags().GetString("name")
		checkRequirements()
		googleapi.GetCommand(sheet, env, category, name, configs)
	},
}

func init() {
	getCmd.Flags().StringP("sheet", "s", "", "Sheet name")
	getCmd.Flags().StringP("env", "e", "", "Environment column")
	getCmd.Flags().StringP("category", "c", "", "Category")
	getCmd.Flags().StringP("name", "n", "", "Name")
	getCmd.MarkFlagRequired("sheet")
	getCmd.MarkFlagRequired("env")
	getCmd.MarkFlagRequired("category")
	getCmd.MarkFlagRequired("name")
}
