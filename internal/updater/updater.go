// Package updater implements self-update functionality by downloading releases from GitHub.
package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// httpClient is used for all HTTP requests with a reasonable timeout.
var httpClient = &http.Client{Timeout: 30 * time.Second}

// UpdateGCNFCommand downloads and installs the latest release of gcnf from GitHub.
func UpdateGCNFCommand(owner, repo string) {
	updateURL := getLatestReleaseURL(owner, repo)
	if updateURL == "" {
		fmt.Println("No suitable release found for your platform.")
		return
	}
	downloadPath, err := downloadBinary(updateURL)
	if err != nil {
		fmt.Printf("Failed to download gcnf: %v\n", err)
		return
	}
	defer os.Remove(downloadPath)
	installPath, err := installBinary()
	if err != nil {
		fmt.Printf("Failed to install gcnf: %v\n", err)
		return
	}
	err = copyFile(downloadPath, installPath)
	if err != nil {
		fmt.Printf("Failed to copy gcnf: %v\n", err)
		return
	}
	fmt.Printf("Update complete. You can now use the latest version of 'gcnf' at %s.\n", installPath)
}

// getLatestReleaseURL fetches the download URL for the current platform's binary from the latest GitHub release.
func getLatestReleaseURL(owner, repo string) string {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	assetsRaw, ok := result["assets"]
	if !ok {
		return ""
	}
	assets, ok := assetsRaw.([]interface{})
	if !ok {
		return ""
	}

	binaryName := getBinaryName()
	for _, asset := range assets {
		assetMap, ok := asset.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := assetMap["name"].(string)
		if !ok {
			continue
		}
		if name == binaryName {
			url, ok := assetMap["browser_download_url"].(string)
			if !ok {
				continue
			}
			return url
		}
	}
	return ""
}

// getBinaryName returns the expected binary name for the current OS and architecture.
func getBinaryName() string {
	return fmt.Sprintf("gcnf-%s-%s", runtime.GOOS, runtime.GOARCH)
}

// downloadBinary downloads a binary from the given URL to a temporary file.
func downloadBinary(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download binary: HTTP %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "gcnf-update.tmp")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

// installBinary returns the target installation path for the gcnf binary.
func installBinary() (string, error) {
	installDir := "/usr/local/bin"
	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("USERPROFILE"), "bin")
	}

	installPath := filepath.Join(installDir, "gcnf")
	if runtime.GOOS == "windows" {
		installPath += ".exe"
	}

	return installPath, nil
}

// copyFile copies src to dst with executable permissions (0755).
func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	if err := os.WriteFile(dst, data, 0755); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}
	return nil
}
