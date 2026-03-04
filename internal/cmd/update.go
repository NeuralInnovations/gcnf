package cmd

import (
	"gcnf/internal/updater"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update gcnf to the latest version",
	Long:  `Download and install the latest version of gcnf from GitHub releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		updater.UpdateGCNFCommand(projectOwner, projectRepository)
	},
}
