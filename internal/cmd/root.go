package cmd

import (
	"fmt"
	"gcnf/internal/config"
	"gcnf/internal/utils"
	"github.com/spf13/cobra"
	"log"
)

var (
	projectName       string
	projectVersion    string
	projectOwner      string
	projectRepository string
	configs           *config.Configs
)

func Execute(projectPropertiesContent string, clientSecrets []byte) {
	// Load project properties
	props, err := utils.LoadProperties(projectPropertiesContent)
	if err != nil {
		log.Fatalf("Failed to load project properties: %v", err)
	}

	projectName = props["name"]
	projectVersion = props["version"]
	projectOwner = props["owner"]
	projectRepository = props["repository"]

	configs = config.NewConfigs(clientSecrets)

	var rootCmd = &cobra.Command{
		Use:   "gcnf",
		Short: fmt.Sprintf("%s %s - Manage and search Google Sheets data.", projectName, projectVersion),
		Long: fmt.Sprintf(`%s %s
Manage and search Google Sheets data.
Environment variables:
[%s, %s, %s, or use %s]`,
			projectName, projectVersion, config.EnvGoogleCredentialBase64, config.EnvGoogleSheetID, config.EnvStoreConfigFile, config.EnvToken),
	}

	// Initialize root command
	rootCmd.Version = projectVersion
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(unloadCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(injectCmd)
	rootCmd.AddCommand(tokenCmd)
	rootCmd.AddCommand(configCmd)

	// execute commands
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
