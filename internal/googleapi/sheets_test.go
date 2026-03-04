package googleapi

import (
	"path/filepath"
	"strings"
	"testing"

	"gcnf/internal/config"
)

// ---------------------------------------------------------------------------
// Tests for sheetToMap
// ---------------------------------------------------------------------------

func TestSheetToMap_ValidData(t *testing.T) {
	// Simulate a sheet with header row and data rows.
	// Header: ["Category", "Name", "dev", "prod"]
	// Data rows belong to category "Database" with two keys.
	rows := [][]interface{}{
		{"Category", "Name", "dev", "prod"},
		{"Database", "host", "localhost", "db.prod.example.com"},
		{"Database", "port", "5432", "5432"},
	}

	configs := &config.Configs{
		GoogleSheetID: "test-sheet-id",
	}

	result := sheetToMap(rows, "Sheet1", "dev", configs)

	// Verify metadata keys.
	if result["__ENV__"] != "dev" {
		t.Errorf("expected __ENV__ = 'dev', got %v", result["__ENV__"])
	}
	if result["__SHEET__"] != "Sheet1" {
		t.Errorf("expected __SHEET__ = 'Sheet1', got %v", result["__SHEET__"])
	}
	if result["__SHEET_ID__"] != "test-sheet-id" {
		t.Errorf("expected __SHEET_ID__ = 'test-sheet-id', got %v", result["__SHEET_ID__"])
	}

	// Verify category data.
	catData, ok := result["Database"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'Database' category to be a map")
	}
	if catData["host"] != "localhost" {
		t.Errorf("expected Database.host = 'localhost', got %v", catData["host"])
	}
	if catData["port"] != "5432" {
		t.Errorf("expected Database.port = '5432', got %v", catData["port"])
	}
}

func TestSheetToMap_MultipleCategories(t *testing.T) {
	rows := [][]interface{}{
		{"Category", "Name", "staging"},
		{"Auth", "secret", "staging-secret"},
		{"Auth", "issuer", "staging-issuer"},
		{"Cache", "ttl", "300"},
		{"Cache", "driver", "redis"},
	}

	configs := &config.Configs{
		GoogleSheetID: "sheet-123",
	}

	result := sheetToMap(rows, "Config", "staging", configs)

	authData, ok := result["Auth"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'Auth' category to be a map")
	}
	if authData["secret"] != "staging-secret" {
		t.Errorf("expected Auth.secret = 'staging-secret', got %v", authData["secret"])
	}
	if authData["issuer"] != "staging-issuer" {
		t.Errorf("expected Auth.issuer = 'staging-issuer', got %v", authData["issuer"])
	}

	cacheData, ok := result["Cache"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'Cache' category to be a map")
	}
	if cacheData["ttl"] != "300" {
		t.Errorf("expected Cache.ttl = '300', got %v", cacheData["ttl"])
	}
	if cacheData["driver"] != "redis" {
		t.Errorf("expected Cache.driver = 'redis', got %v", cacheData["driver"])
	}
}

func TestSheetToMap_NonStringCells(t *testing.T) {
	// When cells are not strings, sheetToMap uses fmt.Sprintf("%v", ...) fallback.
	rows := [][]interface{}{
		{"Category", "Name", "dev"},
		{"Numbers", "count", 42},
		{123, 456, "hello"},
	}

	configs := &config.Configs{
		GoogleSheetID: "sheet-456",
	}

	result := sheetToMap(rows, "Sheet1", "dev", configs)

	numData, ok := result["Numbers"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'Numbers' category to be a map")
	}
	// The integer 42 is converted to "42" via fmt.Sprintf.
	if numData["count"] != "42" {
		t.Errorf("expected Numbers.count = '42', got %v", numData["count"])
	}

	// Category 123 becomes "123", subcategory 456 becomes "456".
	intCatData, ok := result["123"].(map[string]interface{})
	if !ok {
		t.Fatal("expected '123' category to be a map")
	}
	if intCatData["456"] != "hello" {
		t.Errorf("expected 123.456 = 'hello', got %v", intCatData["456"])
	}
}

func TestSheetToMap_SkipsShortRows(t *testing.T) {
	// Rows that are too short (len(row) < 2 or len(row) <= serverIndex)
	// should be skipped silently.
	rows := [][]interface{}{
		{"Category", "Name", "dev"},
		{"Valid", "key", "value"},
		{"Short"},                   // len < 2, skipped
		{"Also", "short"},           // len(row) == 2, serverIndex == 2, len(row) <= serverIndex, skipped
		{"Another", "valid", "val"}, // valid
	}

	configs := &config.Configs{
		GoogleSheetID: "sheet-789",
	}

	result := sheetToMap(rows, "Sheet1", "dev", configs)

	validData, ok := result["Valid"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'Valid' category to be a map")
	}
	if validData["key"] != "value" {
		t.Errorf("expected Valid.key = 'value', got %v", validData["key"])
	}

	anotherData, ok := result["Another"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'Another' category to be a map")
	}
	if anotherData["valid"] != "val" {
		t.Errorf("expected Another.valid = 'val', got %v", anotherData["valid"])
	}

	// "Short" and "Also" should not appear as categories.
	if _, exists := result["Short"]; exists {
		t.Error("did not expect 'Short' category to exist in result")
	}
	if _, exists := result["Also"]; exists {
		t.Error("did not expect 'Also' category to exist in result")
	}
}

func TestSheetToMap_MergesDuplicateCategories(t *testing.T) {
	// When the same category name appears in non-contiguous blocks,
	// the second block should merge into the first.
	rows := [][]interface{}{
		{"Category", "Name", "dev"},
		{"App", "key1", "val1"},
		{"Other", "x", "y"},
		{"App", "key2", "val2"},
	}

	configs := &config.Configs{
		GoogleSheetID: "sheet-merge",
	}

	result := sheetToMap(rows, "Sheet1", "dev", configs)

	appData, ok := result["App"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'App' category to be a map")
	}
	// Both key1 and key2 should be present due to merging.
	if appData["key1"] != "val1" {
		t.Errorf("expected App.key1 = 'val1', got %v", appData["key1"])
	}
	if appData["key2"] != "val2" {
		t.Errorf("expected App.key2 = 'val2', got %v", appData["key2"])
	}
}

func TestSheetToMap_OnlyHeaderRow(t *testing.T) {
	// A sheet with just a header and no data rows should return only metadata.
	rows := [][]interface{}{
		{"Category", "Name", "dev"},
	}

	configs := &config.Configs{
		GoogleSheetID: "sheet-header-only",
	}

	result := sheetToMap(rows, "Sheet1", "dev", configs)

	if result["__ENV__"] != "dev" {
		t.Errorf("expected __ENV__ = 'dev', got %v", result["__ENV__"])
	}
	// Should only have the 3 metadata keys.
	if len(result) != 3 {
		t.Errorf("expected 3 keys (metadata only), got %d", len(result))
	}
}

// Note: sheetToMap with empty rows calls log.Fatalf which invokes os.Exit(1).
// This cannot be caught in normal tests, so we skip that case.

// Note: sheetToMap with a missing environment in the header also calls log.Fatalf.
// We skip that test case for the same reason.

// ---------------------------------------------------------------------------
// Tests for getValue
// ---------------------------------------------------------------------------

func TestGetValue_NilSheetData(t *testing.T) {
	result := getValue(nil, "Category", "Name")
	if result != "No data found." {
		t.Errorf("expected 'No data found.', got %q", result)
	}
}

func TestGetValue_EmptyCategory(t *testing.T) {
	data := map[string]interface{}{
		"Cat": map[string]interface{}{
			"key": "value",
		},
	}
	result := getValue(data, "", "key")
	if result != "Category and name are required." {
		t.Errorf("expected 'Category and name are required.', got %q", result)
	}
}

func TestGetValue_EmptyName(t *testing.T) {
	data := map[string]interface{}{
		"Cat": map[string]interface{}{
			"key": "value",
		},
	}
	result := getValue(data, "Cat", "")
	if result != "Category and name are required." {
		t.Errorf("expected 'Category and name are required.', got %q", result)
	}
}

func TestGetValue_BothEmpty(t *testing.T) {
	data := map[string]interface{}{
		"Cat": map[string]interface{}{
			"key": "value",
		},
	}
	result := getValue(data, "", "")
	if result != "Category and name are required." {
		t.Errorf("expected 'Category and name are required.', got %q", result)
	}
}

func TestGetValue_ExistingCategoryAndName(t *testing.T) {
	data := map[string]interface{}{
		"Database": map[string]interface{}{
			"host": "localhost",
			"port": "5432",
		},
		"Cache": map[string]interface{}{
			"driver": "redis",
		},
	}

	result := getValue(data, "Database", "host")
	if result != "localhost" {
		t.Errorf("expected 'localhost', got %q", result)
	}

	result = getValue(data, "Database", "port")
	if result != "5432" {
		t.Errorf("expected '5432', got %q", result)
	}

	result = getValue(data, "Cache", "driver")
	if result != "redis" {
		t.Errorf("expected 'redis', got %q", result)
	}
}

// Note: getValue with a missing category or missing name calls log.Fatalf,
// so those cases cannot be tested without os.Exit interception. We skip them.

// ---------------------------------------------------------------------------
// Tests for getCachePath
// ---------------------------------------------------------------------------

func TestGetCachePath_Basic(t *testing.T) {
	configs := &config.Configs{ConfigFile: "/tmp/gcnf/gcnf_config.json"}
	path := getCachePath("Env", "staging", configs)
	expected := filepath.Join("/tmp/gcnf", "gcnf_cache_Env_staging.json")
	if path != expected {
		t.Errorf("getCachePath() = %q, want %q", path, expected)
	}
}

func TestGetCachePath_SanitizesSlashes(t *testing.T) {
	configs := &config.Configs{ConfigFile: "/tmp/gcnf/config.json"}
	path := getCachePath("Sheet/Tab", "env/test", configs)
	base := filepath.Base(path)
	if strings.Contains(base, "/") {
		t.Errorf("filename should not contain slashes, got: %s", base)
	}
	if base != "gcnf_cache_Sheet_Tab_env_test.json" {
		t.Errorf("unexpected filename: %s", base)
	}
}

func TestGetCachePath_SanitizesUnsafeChars(t *testing.T) {
	configs := &config.Configs{ConfigFile: "/tmp/gcnf/config.json"}
	path := getCachePath("../evil", "env\\bad", configs)
	base := filepath.Base(path)
	if strings.Contains(base, "..") || strings.Contains(base, "\\") {
		t.Errorf("filename should not contain unsafe chars, got: %s", base)
	}
}

func TestGetCachePath_DifferentEnvsDifferentPaths(t *testing.T) {
	configs := &config.Configs{ConfigFile: "/tmp/gcnf/config.json"}
	path1 := getCachePath("Env", "staging", configs)
	path2 := getCachePath("Env", "production", configs)
	if path1 == path2 {
		t.Errorf("different envs should produce different cache paths, both got: %s", path1)
	}
}

func TestGetCachePath_DifferentSheetsDifferentPaths(t *testing.T) {
	configs := &config.Configs{ConfigFile: "/tmp/gcnf/config.json"}
	path1 := getCachePath("Env", "staging", configs)
	path2 := getCachePath("Secrets", "staging", configs)
	if path1 == path2 {
		t.Errorf("different sheets should produce different cache paths, both got: %s", path1)
	}
}
