package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
)

// Load Command
var loadCmd = &cobra.Command{
	Use:     "load",
	Aliases: []string{"l"},
	Short:   "Load data from Google Sheets and save it locally",
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.Flags().GetString("env")
		sheet, _ := cmd.Flags().GetString("sheet")
		checkRequirements()
		googleapi.LoadCommand(sheet, env, configs)
	},
}

func init() {
	loadCmd.Flags().StringP("env", "e", "", "Environment to download")
	loadCmd.Flags().StringP("sheet", "s", "", "Sheet name")
	loadCmd.MarkFlagRequired("env")
	loadCmd.MarkFlagRequired("sheet")
}
