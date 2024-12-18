package cmd

import (
	"gcnf/internal/config"
	"gcnf/internal/utils"
	"log"
)

func checkRequirements() {
	if configs.GoogleCredentialBase64 == "" && !utils.FileExists(configs.UserTokenFile) {
		log.Fatalf("No authentication method provided. Use 'gcnf login' or set %s.", config.EnvGoogleCredentialBase64)
	}
	if configs.GoogleSheetID == "" {
		log.Fatalf("%s environment variable or --google_sheet_id parameter is not set.", config.EnvGoogleSheetID)
	}
	if configs.ConfigFile == "" || configs.ConfigFile == "." {
		log.Fatalf("%s environment variable or --config_file parameter is not set.", config.EnvStoreConfigFile)
	}
}
