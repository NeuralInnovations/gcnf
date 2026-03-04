package cmd

import (
	"fmt"
	"os"

	"gcnf/internal/googleapi"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [template_path]",
	Args:  cobra.ExactArgs(1),
	Short: "Validate a template file for issues",
	Long: `Parse a template file and report missing variables, malformed gcnf:// URLs,
and empty values. Exits with code 0 if valid, 1 if issues found.`,
	Example: `  gcnf validate .env.template`,
	Run: func(cmd *cobra.Command, args []string) {
		templatePath := args[0]
		issues := googleapi.ValidateCommand(templatePath, configs)
		if len(issues) > 0 {
			for _, issue := range issues {
				fmt.Fprintln(os.Stderr, issue)
			}
			os.Exit(1)
		}
		fmt.Println("Template is valid.")
	},
}
