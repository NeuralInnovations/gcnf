package cmd

import (
	"gcnf/internal/googleapi"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List sheets, environments, or categories",
	Long:  `List available sheet tabs, environment columns, or category names from a Google Sheets spreadsheet.`,
}

var listSheetsCmd = &cobra.Command{
	Use:     "sheets",
	Short:   "List all sheet tabs in the spreadsheet",
	Example: `  gcnf list sheets`,
	Run: func(cmd *cobra.Command, args []string) {
		checkRequirements()
		googleapi.ListSheetsCommand(configs)
	},
}

var listEnvsCmd = &cobra.Command{
	Use:     "envs",
	Short:   "List environment columns in a sheet",
	Example: `  gcnf list envs --sheet Env`,
	Run: func(cmd *cobra.Command, args []string) {
		sheet, _ := cmd.Flags().GetString("sheet")
		checkRequirements()
		googleapi.ListEnvsCommand(sheet, configs)
	},
}

var listCategoriesCmd = &cobra.Command{
	Use:     "categories",
	Short:   "List categories in a sheet",
	Example: `  gcnf list categories --sheet Env`,
	Run: func(cmd *cobra.Command, args []string) {
		sheet, _ := cmd.Flags().GetString("sheet")
		checkRequirements()
		googleapi.ListCategoriesCommand(sheet, configs)
	},
}

func init() {
	listEnvsCmd.Flags().StringP("sheet", "s", "", "Sheet name")
	listEnvsCmd.MarkFlagRequired("sheet")

	listCategoriesCmd.Flags().StringP("sheet", "s", "", "Sheet name")
	listCategoriesCmd.MarkFlagRequired("sheet")

	listCmd.AddCommand(listSheetsCmd)
	listCmd.AddCommand(listEnvsCmd)
	listCmd.AddCommand(listCategoriesCmd)
}
