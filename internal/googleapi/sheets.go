// Package googleapi provides Google Sheets API integration for reading configuration data.
package googleapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gcnf/internal/config"
	"gcnf/internal/utils"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// getCachePath returns the cache file path for a specific sheet+env combination.
// sanitizeCacheComponent replaces filesystem-unsafe characters in a cache path component.
func sanitizeCacheComponent(s string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", "..", "_", ":", "_", "\x00", "_")
	return r.Replace(s)
}

func getCachePath(sheet, env string, configs *config.Configs) string {
	dir := filepath.Dir(configs.ConfigFile)
	safeSheet := sanitizeCacheComponent(sheet)
	safeEnv := sanitizeCacheComponent(env)
	return filepath.Join(dir, fmt.Sprintf("gcnf_cache_%s_%s.json", safeSheet, safeEnv))
}

// GetCachePath is the exported version of getCachePath.
func GetCachePath(sheet, env string, configs *config.Configs) string {
	return getCachePath(sheet, env, configs)
}

func loadSheet(sheet string, env string, configs *config.Configs) {
	rows := loadGoogleSheet(sheet, configs)
	data := sheetToMap(rows, sheet, env, configs)
	cachePath := getCachePath(sheet, env, configs)
	saveToFileAt(data, cachePath)
}

func loadValue(sheet, env, category, name string, configs *config.Configs) string {
	cachePath := getCachePath(sheet, env, configs)
	var needLoad = true
	var tryCount = 2
	var data map[string]interface{}

	for needLoad && tryCount > 0 {
		tryCount--
		needLoad = false
		if !utils.FileExists(cachePath) || utils.IsCacheExpired(cachePath, configs.CacheTTL) {
			loadSheet(sheet, env, configs)
		}
		data = utils.LoadFileContentAsJson(cachePath)
		if data == nil {
			log.Fatal("No data found.")
		}

		envValid := data["__ENV__"] == env
		sheetValid := data["__SHEET__"] == sheet
		sheetIdValid := data["__SHEET_ID__"] == configs.GoogleSheetID
		if !envValid || !sheetValid || !sheetIdValid {
			needLoad = true
			_, _ = utils.DeleteFile(cachePath)
		}
	}

	if data == nil {
		log.Fatal("No data found.")
	}

	// Final validation: ensure data matches the requested env/sheet after retries
	if data["__ENV__"] != env || data["__SHEET__"] != sheet || data["__SHEET_ID__"] != configs.GoogleSheetID {
		log.Fatalf("Config mismatch after retries: expected env=%s sheet=%s sheetID=%s", env, sheet, configs.GoogleSheetID)
	}

	return getValue(data, category, name)
}

func getValue(sheetData map[string]interface{}, category, name string) string {
	if sheetData == nil {
		return "No data found."
	}
	if category == "" || name == "" {
		return "Category and name are required."
	}
	catData, ok := sheetData[category].(map[string]interface{})
	if !ok {
		log.Fatalf("Category '%s' not found.", category)
	}
	value, ok := catData[name].(string)
	if !ok {
		log.Fatalf("Name '%s' not found in category '%s'.", name, category)
	}
	return value
}

// getSheetService creates a Google Sheets API service using the configured credentials.
func getSheetService(configs *config.Configs) *sheets.Service {
	ctx := context.Background()
	var srv *sheets.Service
	var err error

	if configs.GoogleCredentialBase64 != "" {
		var credentialBytes []byte
		credentialBytes, err = base64.StdEncoding.DecodeString(configs.GoogleCredentialBase64)
		if err != nil {
			log.Fatalf("Failed to decode Google credentials: %v", err)
		}
		conf, jwtErr := google.JWTConfigFromJSON(credentialBytes, configs.Scopes...)
		if jwtErr != nil {
			log.Fatalf("Failed to create JWT config: %v", jwtErr)
		}
		client := conf.Client(ctx)
		srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	} else {
		clientSecret := configs.GetClientSecret()
		client := utils.GetUserTokenClient(clientSecret, configs.UserTokenFile)
		if client == nil {
			log.Fatal("Failed to get Google credentials. Use 'gcnf login' to authenticate with Google account.")
		}
		srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	}

	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv
}

func loadGoogleSheet(sheetName string, configs *config.Configs) [][]interface{} {
	srv := getSheetService(configs)
	resp, err := srv.Spreadsheets.Values.Get(configs.GoogleSheetID, sheetName).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	return resp.Values
}

func sheetToMap(rows [][]interface{}, sheet, env string, configs *config.Configs) map[string]interface{} {
	if len(rows) == 0 {
		log.Fatalf("Sheet '%s' is empty.", sheet)
	}

	headerRow := rows[0]
	rows = rows[1:]
	serverIndex := -1
	for i, v := range headerRow {
		s, ok := v.(string)
		if ok && s == env {
			serverIndex = i
			break
		}
	}
	if serverIndex == -1 {
		log.Fatalf("Environment '%s' not found in header.", env)
	}

	data := map[string]interface{}{
		"__ENV__":      env,
		"__SHEET__":    sheet,
		"__SHEET_ID__": configs.GoogleSheetID,
	}
	currentCategory := ""
	categoryData := map[string]interface{}{}

	for _, row := range rows {
		if len(row) <= serverIndex || len(row) < 2 {
			continue
		}
		category, ok := row[0].(string)
		if !ok {
			category = fmt.Sprintf("%v", row[0])
		}
		subcategory, ok := row[1].(string)
		if !ok {
			subcategory = fmt.Sprintf("%v", row[1])
		}
		value, ok := row[serverIndex].(string)
		if !ok {
			value = fmt.Sprintf("%v", row[serverIndex])
		}

		if category != currentCategory {
			if currentCategory != "" {
				if existingData, ok := data[currentCategory]; ok {
					if existingMap, ok := existingData.(map[string]interface{}); ok {
						mergedData := utils.MergeMaps(existingMap, categoryData)
						data[currentCategory] = mergedData
					} else {
						data[currentCategory] = categoryData
					}
				} else {
					data[currentCategory] = categoryData
				}
			}
			currentCategory = category
			categoryData = map[string]interface{}{}
		}
		categoryData[subcategory] = value
	}
	if currentCategory != "" {
		if existingData, ok := data[currentCategory]; ok {
			if existingMap, ok := existingData.(map[string]interface{}); ok {
				mergedData := utils.MergeMaps(existingMap, categoryData)
				data[currentCategory] = mergedData
			} else {
				data[currentCategory] = categoryData
			}
		} else {
			data[currentCategory] = categoryData
		}
	}
	return data
}

func saveToFileAt(data map[string]interface{}, filePath string) {
	if err := utils.EnsureDirectoryExists(filepath.Dir(filePath)); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to write to config file: %v", err)
	}
}
