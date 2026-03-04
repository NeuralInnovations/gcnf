package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:     "load",
	Aliases: []string{"l"},
	Short:   "Load data from Google Sheets and save it locally",
	Long:    `Download data from a Google Sheets spreadsheet for a given environment and sheet name, saving it to the local config file.`,
	Example: `  gcnf load --sheet Env --env staging
  gcnf l -s Env -e production`,
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
