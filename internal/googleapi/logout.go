package googleapi

import (
	"fmt"
	"gcnf/internal/config"
	"gcnf/internal/utils"
	"os"
)

func GoogleLogoutCommand(configs *config.Configs) {
	if utils.FileExists(configs.UserTokenFile) {
		os.Remove(configs.UserTokenFile)
		fmt.Println("Logout successful.")
	} else {
		fmt.Println("No user is currently logged in.")
	}
}
