package googleapi

import (
	"fmt"
	"gcnf/internal/config"
	"strings"
)

func ReadGCNFURL(gcnfURL string, configs *config.Configs) (string, error) {
	if !strings.HasPrefix(gcnfURL, "gcnf://") {
		return "", fmt.Errorf("Invalid gcnf URL. Must start with 'gcnf://'.")
	}
	parts := strings.SplitN(gcnfURL[7:], "/", 4)
	if len(parts) != 4 {
		return "", fmt.Errorf("Invalid gcnf URL format. Expect: 'gcnf://SHEET/ENV/CATEGORY/NAME'")
	}
	sheet, env, category, name := parts[0], parts[1], parts[2], parts[3]
	value := loadValue(sheet, env, category, name, configs)
	return value, nil
}
