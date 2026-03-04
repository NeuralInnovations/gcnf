package googleapi

import (
	"fmt"
	"strings"

	"gcnf/internal/config"
)

// ReadGCNFURL parses a gcnf:// URL and retrieves the corresponding value from Google Sheets.
// URL format: gcnf://SHEET/ENV/CATEGORY/NAME
func ReadGCNFURL(gcnfURL string, configs *config.Configs) (string, error) {
	if !strings.HasPrefix(gcnfURL, "gcnf://") {
		return "", fmt.Errorf("invalid gcnf URL: must start with 'gcnf://'")
	}
	parts := strings.SplitN(gcnfURL[7:], "/", 4)
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid gcnf URL format, expected: gcnf://SHEET/ENV/CATEGORY/NAME")
	}
	sheet, env, category, name := parts[0], parts[1], parts[2], parts[3]
	value := loadValue(sheet, env, category, name, configs)
	return value, nil
}
