package cmd

import (
	"gcnf/internal/utils"
	"github.com/spf13/cobra"
	"log"
)

// Config Command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set configuration",
	Run: func(cmd *cobra.Command, args []string) {
		sheetId, _ := cmd.Flags().GetString("sheetId")
		if sheetId == "" {
			log.Printf("GoogleSheetID %s\n", configs.GoogleSheetID)
		} else {
			configs.GoogleSheetID = sheetId
			utils.WriteStringToFile(configs.UserGoogleSheetIDFile, configs.GoogleSheetID)
		}
	},
}

func init() {
	configCmd.Flags().StringP("sheetId", "i", "", "Sheet ID")
}
