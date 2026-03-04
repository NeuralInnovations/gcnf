// Package cmd defines the CLI commands using the cobra framework.
package cmd

import (
	"fmt"
	"io"
	"log"

	"gcnf/internal/config"
	"gcnf/internal/utils"

	"github.com/spf13/cobra"
)

var (
	projectName       string
	projectVersion    string
	projectOwner      string
	projectRepository string
	configs           *config.Configs
	quietMode         bool
	verboseMode       bool
)

// verboseLog prints a log message only when verbose mode is enabled.
func verboseLog(format string, args ...interface{}) {
	if verboseMode {
		log.Printf("[verbose] "+format, args...)
	}
}

// Execute is the main entry point for the CLI application.
func Execute(projectPropertiesContent string, clientSecrets []byte) {
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
		Long: fmt.Sprintf(`%s %s - Use Google Sheets as a configuration source.

Environment variables:
  %s    Base64-encoded Google service account JSON
  %s                 Google Sheets document ID
  %s            Path to local config cache file
  %s                          Composite token (alternative to individual vars)`,
			projectName, projectVersion,
			config.EnvGoogleCredentialBase64, config.EnvGoogleSheetID,
			config.EnvStoreConfigFile, config.EnvToken),
	}

	rootCmd.Version = projectVersion
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Suppress non-data output")
	rootCmd.PersistentFlags().BoolVar(&verboseMode, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if quietMode {
			log.SetOutput(io.Discard)
		}
	}

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
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
