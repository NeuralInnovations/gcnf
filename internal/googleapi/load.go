package googleapi

import (
	"fmt"
	"gcnf/internal/config"
)

func LoadCommand(sheet, env string, configs *config.Configs) {
	loadSheet(sheet, env, configs)
	fmt.Printf("Data loaded and saved to %s.\n", configs.ConfigFile)
}
