package cmd

import (
	"fmt"
	"gcnf/internal/utils"
	"github.com/spf13/cobra"
)

// Delete Command
var unloadCmd = &cobra.Command{
	Use:     "unload",
	Aliases: []string{"d"},
	Short:   "Unload the locally saved data",
	Run: func(cmd *cobra.Command, args []string) {
		if utils.DeleteFile(configs.ConfigFile) {
			fmt.Printf("%s deleted.\n", configs.ConfigFile)
		} else {
			fmt.Printf("%s does not exist.\n", configs.ConfigFile)
		}
	},
}
