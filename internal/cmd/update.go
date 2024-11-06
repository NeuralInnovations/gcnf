package cmd

import (
	"fmt"
	"gcnf/internal/updater"
	"github.com/spf13/cobra"
)

// Update Command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: fmt.Sprintf("Update %s to the latest version", projectName),
	Run: func(cmd *cobra.Command, args []string) {
		updater.UpdateGCNFCommand(projectOwner, projectRepository)
	},
}
