package googleapi

import (
	"fmt"
	"gcnf/internal/config"
)

func GetCommand(sheet, env, category, name string, configs *config.Configs) {
	value := loadValue(sheet, env, category, name, configs)
	fmt.Println(value)
}
