package config

import (
	"encoding/base64"
	"strings"
	"testing"

	"gcnf/internal/utils"
)

func TestEncodeDecodeToken_RoundTrip(t *testing.T) {
	original := ProjectToken{
		GoogleCredBase64: "my-google-cred-base64-data",
		GoogleSheetID:    "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
		StoreConfigFile:  "/path/to/config.json",
	}

	encoded := EncodeToken(original)
	decoded := DecodeToken(encoded)

	if decoded == nil {
		t.Fatal("DecodeToken returned nil for a valid encoded token")
	}
	if decoded.GoogleCredBase64 != original.GoogleCredBase64 {
		t.Errorf("GoogleCredBase64 mismatch: got %q, want %q", decoded.GoogleCredBase64, original.GoogleCredBase64)
	}
	if decoded.GoogleSheetID != original.GoogleSheetID {
		t.Errorf("GoogleSheetID mismatch: got %q, want %q", decoded.GoogleSheetID, original.GoogleSheetID)
	}
	if decoded.StoreConfigFile != original.StoreConfigFile {
		t.Errorf("StoreConfigFile mismatch: got %q, want %q", decoded.StoreConfigFile, original.StoreConfigFile)
	}
}

func TestDecodeToken_FourPartBackwardCompatibility(t *testing.T) {
	// Build a 4-part token manually: cred.sheetID.deprecated.configFile
	cred := "test-credential"
	sheetID := "sheet-id-123"
	deprecated := "old-sheet-name"
	configFile := "./config.json"

	parts := []string{
		utils.ToBase64(cred),
		utils.ToBase64(sheetID),
		utils.ToBase64(deprecated),
		utils.ToBase64(configFile),
	}
	inner := strings.Join(parts, ".")
	encoded := base64.StdEncoding.EncodeToString([]byte(inner))

	decoded := DecodeToken(encoded)
	if decoded == nil {
		t.Fatal("DecodeToken returned nil for a valid 4-part token")
	}
	if decoded.GoogleCredBase64 != cred {
		t.Errorf("GoogleCredBase64 mismatch: got %q, want %q", decoded.GoogleCredBase64, cred)
	}
	if decoded.GoogleSheetID != sheetID {
		t.Errorf("GoogleSheetID mismatch: got %q, want %q", decoded.GoogleSheetID, sheetID)
	}
	// parts[2] (deprecated) should be skipped
	if decoded.StoreConfigFile != configFile {
		t.Errorf("StoreConfigFile mismatch: got %q, want %q", decoded.StoreConfigFile, configFile)
	}
}

func TestDecodeToken_InvalidOuterBase64(t *testing.T) {
	result := DecodeToken("!!!not-valid-base64!!!")
	if result != nil {
		t.Error("DecodeToken should return nil for invalid outer base64")
	}
}

func TestDecodeToken_InvalidInnerBase64(t *testing.T) {
	// Create a token where the inner parts are not valid base64
	inner := "not-base64-part1.not-base64-part2.not-base64-part3"
	encoded := base64.StdEncoding.EncodeToString([]byte(inner))

	result := DecodeToken(encoded)
	if result != nil {
		t.Error("DecodeToken should return nil when inner base64 parts are invalid")
	}
}

func TestDecodeToken_WrongNumberOfParts(t *testing.T) {
	testCases := []struct {
		name      string
		numParts  int
	}{
		{"1 part", 1},
		{"2 parts", 2},
		{"5 parts", 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parts := make([]string, tc.numParts)
			for i := range parts {
				parts[i] = utils.ToBase64("value")
			}
			inner := strings.Join(parts, ".")
			encoded := base64.StdEncoding.EncodeToString([]byte(inner))

			result := DecodeToken(encoded)
			if result != nil {
				t.Errorf("DecodeToken should return nil for %d parts, but got non-nil", tc.numParts)
			}
		})
	}
}

func TestEncodeDecodeToken_EmptyFields(t *testing.T) {
	original := ProjectToken{
		GoogleCredBase64: "",
		GoogleSheetID:    "",
		StoreConfigFile:  "",
	}

	encoded := EncodeToken(original)
	decoded := DecodeToken(encoded)

	if decoded == nil {
		t.Fatal("DecodeToken returned nil for a token with empty fields")
	}
	if decoded.GoogleCredBase64 != "" {
		t.Errorf("GoogleCredBase64 should be empty, got %q", decoded.GoogleCredBase64)
	}
	if decoded.GoogleSheetID != "" {
		t.Errorf("GoogleSheetID should be empty, got %q", decoded.GoogleSheetID)
	}
	if decoded.StoreConfigFile != "" {
		t.Errorf("StoreConfigFile should be empty, got %q", decoded.StoreConfigFile)
	}
}

func TestGenerateToken(t *testing.T) {
	configs := &Configs{
		GoogleCredentialBase64: "test-cred-base64",
		GoogleSheetID:         "test-sheet-id",
		ConfigFile:            "/etc/gcnf/config.json",
	}

	token := GenerateToken(configs)
	if token == "" {
		t.Fatal("GenerateToken returned empty string")
	}

	decoded := DecodeToken(token)
	if decoded == nil {
		t.Fatal("DecodeToken returned nil for a GenerateToken output")
	}
	if decoded.GoogleCredBase64 != configs.GoogleCredentialBase64 {
		t.Errorf("GoogleCredBase64 mismatch: got %q, want %q", decoded.GoogleCredBase64, configs.GoogleCredentialBase64)
	}
	if decoded.GoogleSheetID != configs.GoogleSheetID {
		t.Errorf("GoogleSheetID mismatch: got %q, want %q", decoded.GoogleSheetID, configs.GoogleSheetID)
	}
	if decoded.StoreConfigFile != configs.ConfigFile {
		t.Errorf("StoreConfigFile mismatch: got %q, want %q", decoded.StoreConfigFile, configs.ConfigFile)
	}
}
