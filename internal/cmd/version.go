package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Version Command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(projectVersion)
	},
}
