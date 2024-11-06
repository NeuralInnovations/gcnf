package googleapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"gcnf/internal/config"
	"gcnf/internal/utils"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"os"
	"path/filepath"
)

func loadSheet(sheet string, env string, configs *config.Configs) {
	rows := loadGoogleSheet(sheet, configs)
	data := sheetToMap(rows, sheet, env, configs)
	saveToFile(data, configs)
}

func loadValue(sheet, env, category, name string, configs *config.Configs) string {
	var needLoad = true
	var tryCount = 2
	var data map[string]interface{} = nil

	for needLoad && tryCount > 0 {
		tryCount -= 1
		needLoad = false
		if !utils.FileExists(configs.ConfigFile) {
			loadSheet(sheet, env, configs)
		}
		data = utils.GetFileContent(configs.ConfigFile)
		if data == nil {
			log.Fatal("No data found.")
		}

		envValid := data["__ENV__"] == env
		sheetValid := data["__SHEET__"] == sheet
		sheetIdValid := data["__SHEET_ID__"] == configs.GoogleSheetID
		if !envValid || !sheetValid || !sheetIdValid {
			needLoad = true
			utils.DeleteFile(configs.ConfigFile)
		}
	}

	if data == nil {
		log.Fatal("No data found.")
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

func loadGoogleSheet(sheetName string, configs *config.Configs) [][]interface{} {
	ctx := context.Background()
	var srv *sheets.Service
	var err error

	if configs.GoogleCredentialBase64 != "" {
		credentialBytes, err := base64.StdEncoding.DecodeString(configs.GoogleCredentialBase64)
		if err != nil {
			log.Fatalf("Failed to decode Google credentials: %v", err)
		}
		config, err := google.JWTConfigFromJSON(credentialBytes, configs.Scopes...)
		if err != nil {
			log.Fatalf("Failed to create JWT config: %v", err)
		}
		client := config.Client(ctx)
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

	resp, err := srv.Spreadsheets.Values.Get(configs.GoogleSheetID, sheetName).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	return resp.Values
}

func sheetToMap(rows [][]interface{}, sheet, env string, configs *config.Configs) map[string]interface{} {
	headerRow := rows[0]
	rows = rows[1:]
	serverIndex := -1
	for i, v := range headerRow {
		if v.(string) == env {
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
		if len(row) > serverIndex {
			category := row[0].(string)
			subcategory := row[1].(string)
			value := row[serverIndex].(string)

			if category != currentCategory {
				if currentCategory != "" {
					if existingData, ok := data[currentCategory]; ok {
						mergedData := utils.MergeMaps(existingData.(map[string]interface{}), categoryData)
						data[currentCategory] = mergedData
					} else {
						data[currentCategory] = categoryData
					}
				}
				currentCategory = category
				categoryData = map[string]interface{}{}
			}
			categoryData[subcategory] = value
		}
	}
	if currentCategory != "" {
		if existingData, ok := data[currentCategory]; ok {
			mergedData := utils.MergeMaps(existingData.(map[string]interface{}), categoryData)
			data[currentCategory] = mergedData
		} else {
			data[currentCategory] = categoryData
		}
	}
	return data
}

func saveToFile(data map[string]interface{}, configs *config.Configs) {
	utils.EnsureDirectoryExists(filepath.Dir(configs.ConfigFile))
	file, err := os.Create(configs.ConfigFile)
	if err != nil {
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		log.Fatalf("Failed to write to config file: %v", err)
	}
}
