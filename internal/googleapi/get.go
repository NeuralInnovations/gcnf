package googleapi

import (
	"fmt"

	"gcnf/internal/config"
)

// GetCommand retrieves a specific value from Google Sheets and prints it to stdout.
func GetCommand(sheet, env, category, name string, configs *config.Configs) {
	value := loadValue(sheet, env, category, name, configs)
	fmt.Println(value)
}
