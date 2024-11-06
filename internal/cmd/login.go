package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
)

// Login Command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Google account",
	Run: func(cmd *cobra.Command, args []string) {
		googleapi.GoogleLoginCommand(configs)
	},
}
