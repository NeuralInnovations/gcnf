package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
)

// Logout Command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Google account",
	Run: func(cmd *cobra.Command, args []string) {
		googleapi.GoogleLogoutCommand(configs)
	},
}
