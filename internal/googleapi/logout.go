package googleapi

import (
	"fmt"
	"log"
	"os"

	"gcnf/internal/config"
	"gcnf/internal/utils"
)

// GoogleLogoutCommand removes the user's OAuth2 token file.
func GoogleLogoutCommand(configs *config.Configs) {
	if utils.FileExists(configs.UserTokenFile) {
		if err := os.Remove(configs.UserTokenFile); err != nil {
			log.Fatalf("Failed to remove token file: %v", err)
		}
		fmt.Println("Logout successful.")
	} else {
		fmt.Println("No user is currently logged in.")
	}
}
