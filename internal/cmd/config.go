package cmd

import (
	"log"

	"gcnf/internal/utils"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set configuration",
	Long:  `View or set persistent configuration values such as the Google Sheet ID.`,
	Example: `  gcnf config --sheetId YOUR_SHEET_ID
  gcnf config`,
	Run: func(cmd *cobra.Command, args []string) {
		sheetId, _ := cmd.Flags().GetString("sheetId")
		if sheetId == "" {
			log.Printf("GoogleSheetID %s\n", configs.GoogleSheetID)
		} else {
			configs.GoogleSheetID = sheetId
			if err := utils.WriteStringToFile(configs.UserGoogleSheetIDFile, configs.GoogleSheetID); err != nil {
				log.Fatalf("Failed to save sheet ID: %v", err)
			}
		}
	},
}

func init() {
	configCmd.Flags().StringP("sheetId", "i", "", "Sheet ID")
}
