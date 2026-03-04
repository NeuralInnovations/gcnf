package cmd

import (
	"log"

	"gcnf/internal/config"
	"gcnf/internal/utils"
)

func checkRequirements() {
	if configs.GoogleCredentialBase64 == "" && !utils.FileExists(configs.UserTokenFile) {
		log.Fatalf("No authentication method provided. Use 'gcnf login' or set %s.", config.EnvGoogleCredentialBase64)
	}
	if configs.GoogleSheetID == "" {
		log.Fatalf("%s environment variable is not set. Use 'gcnf config --sheetId <ID>' to set it.", config.EnvGoogleSheetID)
	}
	if configs.ConfigFile == "" || configs.ConfigFile == "." {
		log.Fatalf("%s environment variable is not set.", config.EnvStoreConfigFile)
	}
}
