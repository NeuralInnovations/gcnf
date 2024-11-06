package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

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
	os.Remove(downloadPath)
	fmt.Printf("Update complete. You can now use the latest version of 'gcnf' at %s.\n", installPath)
}

func getLatestReleaseURL(owner, repo string) string {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return ""
	}
	assets := result["assets"].([]interface{})
	for _, asset := range assets {
		assetMap := asset.(map[string]interface{})
		name := assetMap["name"].(string)
		if name == getBinaryName() {
			return assetMap["browser_download_url"].(string)
		}
	}
	return ""
}

func getBinaryName() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "x86_64" {
		arch = "amd64"
	}
	return fmt.Sprintf("gcnf-%s-%s", osName, arch)
}

func downloadBinary(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download binary")
	}
	defer resp.Body.Close()
	tempFile, err := os.CreateTemp("", "gcnf-update.tmp")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func installBinary() (string, error) {
	installDir := "/usr/local/bin"
	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("USERPROFILE"), "bin")
	}

	installPath := filepath.Join(installDir, "gcnf")
	if runtime.GOOS == "windows" {
		installPath += ".exe"
	}

	// Make the binary executable
	if runtime.GOOS != "windows" {
		err := os.Chmod(installPath, 0755)
		if err != nil {
			return "", err
		}
	}

	return installPath, nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func copyFile(src string, dst string) error {
	// Read all content of src to data, may cause OOM for a large file.
	data, err := os.ReadFile(src)
	checkErr(err)
	// Write data to dst
	err = os.WriteFile(dst, data, os.FileMode(0644))
	checkErr(err)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, f.Mode())
			if err != nil {
				return err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), f.Mode()); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
