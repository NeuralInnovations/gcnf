package cmd

import (
	"fmt"
	"gcnf/internal/config"
	"github.com/spf13/cobra"
)

// Token Command
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Token commands",
}

func init() {
	tokenCmd.AddCommand(generateTokenCmd)
}

// Generate Token Command
var generateTokenCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate service token",
	Run: func(cmd *cobra.Command, args []string) {
		checkRequirements()
		tkn := config.GenerateToken(configs)
		fmt.Println(tkn)
	},
}
