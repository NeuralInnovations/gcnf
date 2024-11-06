package cmd

import (
	"encoding/json"
	"fmt"
	"gcnf/internal/utils"
	"github.com/spf13/cobra"
)

// Status Command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Status of the gcnf tool",
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		statusCommand(format)
	},
}

func init() {
	statusCmd.Flags().StringP("format", "f", "yaml", "Output format (yaml or json)")
}

func statusCommand(format string) {
	status := map[string]string{
		"name":                  projectName,
		"version":               projectVersion,
		"owner":                 projectOwner,
		"repository":            projectRepository,
		"google_sheet_id":       configs.GoogleSheetID,
		"google_sheet_name":     configs.GoogleSheetName,
		"storage_config_file":   configs.ConfigFile,
		"credentials_status":    configs.GetCredentialsStatus(),
		"user_token_file":       configs.GetUserTokenStatus(),
		"google_credential_b64": configs.GetBase64CredentialStatus(),
	}

	if format == "json" {
		data, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(data))
	} else {
		utils.PrintYAML(status)
	}
}
