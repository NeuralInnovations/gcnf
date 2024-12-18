package config

import (
	"encoding/base64"
	"gcnf/internal/utils"
	"strings"
)

// Token handling functions moved here

type ProjectToken struct {
	GoogleCredBase64 string
	GoogleSheetID    string
	StoreConfigFile  string
}

func EncodeToken(t ProjectToken) string {
	items := []string{
		utils.ToBase64(t.GoogleCredBase64),
		utils.ToBase64(t.GoogleSheetID),
		utils.ToBase64(t.StoreConfigFile),
	}
	strToEncode := strings.Join(items, ".")
	encoded := base64.StdEncoding.EncodeToString([]byte(strToEncode))
	return encoded
}

func DecodeToken(encodedToken string) *ProjectToken {
	decoded, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return nil
	}
	parts := strings.Split(string(decoded), ".")
	if len(parts) == 3 {
		return &ProjectToken{
			GoogleCredBase64: utils.FromBase64(parts[0]),
			GoogleSheetID:    utils.FromBase64(parts[1]),
			StoreConfigFile:  utils.FromBase64(parts[2]),
		}
	}
	if len(parts) == 4 {
		return &ProjectToken{
			GoogleCredBase64: utils.FromBase64(parts[0]),
			GoogleSheetID:    utils.FromBase64(parts[1]),

			StoreConfigFile: utils.FromBase64(parts[3]),
		}
	}
	return nil
}

func GenerateToken(configs *Configs) string {
	t := ProjectToken{
		GoogleCredBase64: configs.GoogleCredentialBase64,
		GoogleSheetID:    configs.GoogleSheetID,
		StoreConfigFile:  configs.ConfigFile,
	}
	return EncodeToken(t)
}
