package cmd

import (
	"fmt"
	"gcnf/internal/googleapi"
	"github.com/spf13/cobra"
	"log"
)

var readCmd = &cobra.Command{
	Use:     "read",
	Aliases: []string{"r"},
	Short:   "Read value from Google Sheets using gcnf URL",
	Long:    `Read a configuration value from Google Sheets using the gcnf:// URL format.`,
	Example: `  gcnf read "gcnf://Env/Staging/Database/ConnectionString"
  gcnf r "gcnf://Env/Production/App/ApiKey"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("gcnf url is required")
		}
		gcnfURL := args[0]
		checkRequirements()
		value, err := googleapi.ReadGCNFURL(gcnfURL, configs)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(value)
	},
}
