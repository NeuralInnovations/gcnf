package googleapi

import (
	"fmt"
	"log"
	"path/filepath"

	"gcnf/internal/config"
	"gcnf/internal/utils"
)

// LoadCommand downloads data from a Google Sheets spreadsheet and saves it locally.
func LoadCommand(sheet, env string, configs *config.Configs) {
	loadSheet(sheet, env, configs)
	cachePath := getCachePath(sheet, env, configs)
	fmt.Printf("Data loaded and saved to %s.\n", cachePath)
}

// UnloadAllCaches deletes all cache files in the config directory.
func UnloadAllCaches(configs *config.Configs) {
	dir := filepath.Dir(configs.ConfigFile)
	matches, err := filepath.Glob(filepath.Join(dir, "gcnf_cache_*.json"))
	if err != nil {
		log.Printf("Warning: could not list cache files: %v", err)
		return
	}
	// Also remove the old single-file cache for backward compat,
	// but only if it's the default gcnf_config.json name (not a user-specified file)
	if filepath.Base(configs.ConfigFile) == "gcnf_config.json" && utils.FileExists(configs.ConfigFile) {
		matches = append(matches, configs.ConfigFile)
	}
	if len(matches) == 0 {
		fmt.Println("No cache files found.")
		return
	}
	for _, f := range matches {
		deleted, delErr := utils.DeleteFile(f)
		if deleted {
			fmt.Printf("%s deleted.\n", f)
		}
		if delErr != nil {
			log.Printf("Warning: failed to delete %s: %v", f, delErr)
		}
	}
}

// UnloadCache deletes a specific cache file for a given sheet+env combination.
func UnloadCache(sheet, env string, configs *config.Configs) {
	cachePath := getCachePath(sheet, env, configs)
	deleted, err := utils.DeleteFile(cachePath)
	if err != nil {
		log.Fatalf("Failed to delete cache file: %v", err)
	}
	if deleted {
		fmt.Printf("%s deleted.\n", cachePath)
	} else {
		fmt.Printf("%s does not exist.\n", cachePath)
	}
}
