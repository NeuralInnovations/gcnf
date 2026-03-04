package googleapi

import (
	"fmt"
	"log"

	"gcnf/internal/config"
)

// extractEnvNames returns environment column names from a header row (skipping first 2 columns).
func extractEnvNames(headerRow []interface{}) []string {
	var envs []string
	for i := 2; i < len(headerRow); i++ {
		envs = append(envs, fmt.Sprintf("%v", headerRow[i]))
	}
	return envs
}

// extractCategories returns unique category names from sheet rows (skipping header).
func extractCategories(rows [][]interface{}) []string {
	seen := make(map[string]bool)
	var categories []string
	for _, row := range rows {
		if len(row) > 0 {
			cat := fmt.Sprintf("%v", row[0])
			if cat != "" && !seen[cat] {
				seen[cat] = true
				categories = append(categories, cat)
			}
		}
	}
	return categories
}

// ListSheetsCommand lists all sheet tab names in the spreadsheet.
func ListSheetsCommand(configs *config.Configs) {
	srv := getSheetService(configs)
	spreadsheet, err := srv.Spreadsheets.Get(configs.GoogleSheetID).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve spreadsheet metadata: %v", err)
	}
	for _, sheet := range spreadsheet.Sheets {
		fmt.Println(sheet.Properties.Title)
	}
}

// ListEnvsCommand lists environment column names from a sheet's header row.
func ListEnvsCommand(sheetName string, configs *config.Configs) {
	rows := loadGoogleSheet(sheetName, configs)
	if len(rows) == 0 {
		log.Fatalf("Sheet '%s' is empty.", sheetName)
	}
	for _, env := range extractEnvNames(rows[0]) {
		fmt.Println(env)
	}
}

// ListCategoriesCommand lists unique category names from a sheet.
func ListCategoriesCommand(sheetName string, configs *config.Configs) {
	rows := loadGoogleSheet(sheetName, configs)
	if len(rows) < 2 {
		log.Fatalf("Sheet '%s' has no data rows.", sheetName)
	}
	for _, cat := range extractCategories(rows[1:]) {
		fmt.Println(cat)
	}
}
