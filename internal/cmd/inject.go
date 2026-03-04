package cmd

import (
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
	"log"
)

var injectCmd = &cobra.Command{
	Use:     "inject",
	Aliases: []string{"i"},
	Short:   "Inject data from Google Sheets into a template file",
	Long: `Process a template file by resolving environment variable references ($VAR, ${VAR:-default})
and gcnf:// URLs, printing the result to stdout or to a file.`,
	Example: `  gcnf inject .env.template > .env
  gcnf inject -o .env .env.template
  gcnf i -c .env.template > .env`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("template_path is required")
		}
		templatePath := args[0]
		skipComments, _ := cmd.Flags().GetBool("skip-comments")
		outputPath, _ := cmd.Flags().GetString("output")
		checkRequirements()
		err := googleapi.InjectCommand(templatePath, skipComments, outputPath, configs)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	injectCmd.Flags().BoolP("skip-comments", "c", false, "Skip comments")
	injectCmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")
}
