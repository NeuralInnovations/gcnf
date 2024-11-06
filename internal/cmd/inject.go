package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
	"log"
)

// Inject Command
var injectCmd = &cobra.Command{
	Use:     "inject",
	Aliases: []string{"i"},
	Short:   "Inject data from Google Sheets into a template file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("template_path is required")
		}
		templatePath := args[0]
		skipComments, _ := cmd.Flags().GetBool("skip-comments")
		checkRequirements()
		err := googleapi.InjectCommand(templatePath, skipComments, configs)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	injectCmd.Flags().BoolP("skip-comments", "c", false, "Skip comments")
}
