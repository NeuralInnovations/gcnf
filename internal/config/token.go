package config

import (
	"encoding/base64"
	"gcnf/internal/utils"
	"strings"
)

// ProjectToken represents a composite token that bundles authentication credentials,
// sheet ID, and config file path into a single encoded string.
type ProjectToken struct {
	GoogleCredBase64 string
	GoogleSheetID    string
	StoreConfigFile  string
}

// EncodeToken encodes a ProjectToken into a base64-encoded dot-separated string.
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

// DecodeToken decodes a base64-encoded token string into a ProjectToken.
// Returns nil if the token is invalid or cannot be decoded.
func DecodeToken(encodedToken string) *ProjectToken {
	decoded, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return nil
	}
	parts := strings.Split(string(decoded), ".")
	if len(parts) == 3 {
		cred, err1 := utils.FromBase64(parts[0])
		sheetID, err2 := utils.FromBase64(parts[1])
		configFile, err3 := utils.FromBase64(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			return nil
		}
		return &ProjectToken{
			GoogleCredBase64: cred,
			GoogleSheetID:    sheetID,
			StoreConfigFile:  configFile,
		}
	}
	// Backward compatibility: 4-part tokens from an older format had a deprecated
	// field at index 2 (formerly GoogleSheetName). It is silently skipped.
	// parts[0]=cred, parts[1]=sheetID, parts[2]=deprecated, parts[3]=configFile
	if len(parts) == 4 {
		cred, err1 := utils.FromBase64(parts[0])
		sheetID, err2 := utils.FromBase64(parts[1])
		configFile, err3 := utils.FromBase64(parts[3])
		if err1 != nil || err2 != nil || err3 != nil {
			return nil
		}
		return &ProjectToken{
			GoogleCredBase64: cred,
			GoogleSheetID:    sheetID,
			StoreConfigFile:  configFile,
		}
	}
	return nil
}

// GenerateToken creates an encoded token string from the current configuration.
func GenerateToken(configs *Configs) string {
	t := ProjectToken{
		GoogleCredBase64: configs.GoogleCredentialBase64,
		GoogleSheetID:    configs.GoogleSheetID,
		StoreConfigFile:  configs.ConfigFile,
	}
	return EncodeToken(t)
}
