package cmd

import (
	"log"

	"gcnf/internal/googleapi"

	"github.com/spf13/cobra"
)

var unloadCmd = &cobra.Command{
	Use:     "unload",
	Aliases: []string{"d"},
	Short:   "Unload the locally saved data",
	Long: `Delete locally cached configuration files downloaded from Google Sheets.

Without flags, deletes all cache files. With --sheet and --env, deletes only
the cache for that specific sheet+env combination.`,
	Example: `  gcnf unload
  gcnf unload --sheet Env --env staging
  gcnf d`,
	Run: func(cmd *cobra.Command, args []string) {
		sheet, _ := cmd.Flags().GetString("sheet")
		env, _ := cmd.Flags().GetString("env")
		if sheet != "" && env != "" {
			googleapi.UnloadCache(sheet, env, configs)
		} else if sheet != "" || env != "" {
			log.Fatal("Both --sheet and --env are required for targeted deletion.")
		} else {
			googleapi.UnloadAllCaches(configs)
		}
	},
}

func init() {
	unloadCmd.Flags().StringP("sheet", "s", "", "Sheet name (delete specific cache)")
	unloadCmd.Flags().StringP("env", "e", "", "Environment (delete specific cache)")
}
