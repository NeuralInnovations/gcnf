package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetBinaryName(t *testing.T) {
	expected := fmt.Sprintf("gcnf-%s-%s", runtime.GOOS, runtime.GOARCH)
	got := getBinaryName()
	if got != expected {
		t.Errorf("getBinaryName() = %q, want %q", got, expected)
	}
}

// roundTripFunc is an adapter to allow using ordinary functions as http.RoundTripper.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// redirectingClient returns an *http.Client that redirects all requests to the given test server URL.
func redirectingClient(serverURL string) *http.Client {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = parsed.Scheme
			req.URL.Host = parsed.Host
			return http.DefaultTransport.RoundTrip(req)
		}),
	}
}

// withMockHTTPClient temporarily overrides the package-level httpClient for a test,
// restoring it when the test completes.
func withMockHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()
	original := httpClient
	httpClient = client
	t.Cleanup(func() { httpClient = original })
}

func TestGetLatestReleaseURL_ValidResponseWithMatchingAsset(t *testing.T) {
	binaryName := getBinaryName()
	expectedURL := "https://github.com/example/repo/releases/download/v1.0.0/" + binaryName

	payload := map[string]interface{}{
		"assets": []interface{}{
			map[string]interface{}{
				"name":                 "gcnf-other-os",
				"browser_download_url": "https://example.com/other",
			},
			map[string]interface{}{
				"name":                 binaryName,
				"browser_download_url": expectedURL,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	withMockHTTPClient(t, redirectingClient(server.URL))

	result := getLatestReleaseURL("owner", "repo")
	if result != expectedURL {
		t.Errorf("getLatestReleaseURL() = %q, want %q", result, expectedURL)
	}
}

func TestGetLatestReleaseURL_NoMatchingAsset(t *testing.T) {
	payload := map[string]interface{}{
		"assets": []interface{}{
			map[string]interface{}{
				"name":                 "gcnf-nonexistent-os-nonexistent-arch",
				"browser_download_url": "https://example.com/wrong-binary",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	withMockHTTPClient(t, redirectingClient(server.URL))

	result := getLatestReleaseURL("owner", "repo")
	if result != "" {
		t.Errorf("getLatestReleaseURL() = %q, want empty string", result)
	}
}

func TestGetLatestReleaseURL_Non200Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	withMockHTTPClient(t, redirectingClient(server.URL))

	result := getLatestReleaseURL("owner", "repo")
	if result != "" {
		t.Errorf("getLatestReleaseURL() = %q, want empty string for non-200", result)
	}
}

func TestGetLatestReleaseURL_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	withMockHTTPClient(t, redirectingClient(server.URL))

	result := getLatestReleaseURL("owner", "repo")
	if result != "" {
		t.Errorf("getLatestReleaseURL() = %q, want empty string for malformed JSON", result)
	}
}

func TestGetLatestReleaseURL_MissingAssetsField(t *testing.T) {
	payload := map[string]interface{}{
		"tag_name": "v1.0.0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	withMockHTTPClient(t, redirectingClient(server.URL))

	result := getLatestReleaseURL("owner", "repo")
	if result != "" {
		t.Errorf("getLatestReleaseURL() = %q, want empty string for missing assets", result)
	}
}

func TestGetLatestReleaseURL_AssetWithoutNameField(t *testing.T) {
	payload := map[string]interface{}{
		"assets": []interface{}{
			map[string]interface{}{
				"browser_download_url": "https://example.com/some-binary",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	withMockHTTPClient(t, redirectingClient(server.URL))

	result := getLatestReleaseURL("owner", "repo")
	if result != "" {
		t.Errorf("getLatestReleaseURL() = %q, want empty string for asset without name", result)
	}
}

func TestDownloadBinary_Success(t *testing.T) {
	expectedContent := "fake-binary-content-here"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	withMockHTTPClient(t, server.Client())

	downloadPath, err := downloadBinary(server.URL + "/download")
	if err != nil {
		t.Fatalf("downloadBinary() returned unexpected error: %v", err)
	}
	defer os.Remove(downloadPath)

	if downloadPath == "" {
		t.Fatal("downloadBinary() returned empty path")
	}

	// Verify the temp file exists
	info, err := os.Stat(downloadPath)
	if err != nil {
		t.Fatalf("downloaded file does not exist: %v", err)
	}
	if info.Size() == 0 {
		t.Error("downloaded file is empty")
	}

	// Verify file content
	content, err := os.ReadFile(downloadPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != expectedContent {
		t.Errorf("file content = %q, want %q", string(content), expectedContent)
	}
}

func TestDownloadBinary_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	withMockHTTPClient(t, server.Client())

	_, err := downloadBinary(server.URL + "/download")
	if err == nil {
		t.Error("downloadBinary() should return error for non-200 response")
	}
}

func TestDownloadBinary_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	withMockHTTPClient(t, server.Client())

	_, err := downloadBinary(server.URL + "/download")
	if err == nil {
		t.Error("downloadBinary() should return error for server error response")
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary source file
	srcContent := []byte("binary content for copy test")
	srcFile, err := os.CreateTemp("", "copy-src-*.tmp")
	if err != nil {
		t.Fatalf("failed to create temp source file: %v", err)
	}
	defer os.Remove(srcFile.Name())

	if _, err := srcFile.Write(srcContent); err != nil {
		t.Fatalf("failed to write to source file: %v", err)
	}
	srcFile.Close()

	// Create a temporary destination path
	dstPath := filepath.Join(os.TempDir(), "copy-dst-test-binary")
	defer os.Remove(dstPath)

	err = copyFile(srcFile.Name(), dstPath)
	if err != nil {
		t.Fatalf("copyFile() returned unexpected error: %v", err)
	}

	// Verify contents match
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(dstContent) != string(srcContent) {
		t.Errorf("destination content = %q, want %q", string(dstContent), string(srcContent))
	}

	// Verify permissions (0755)
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat destination file: %v", err)
	}
	// On Unix, check the permission bits. On Windows, permissions work differently.
	if runtime.GOOS != "windows" {
		perm := info.Mode().Perm()
		if perm != 0755 {
			t.Errorf("destination file permissions = %o, want 0755", perm)
		}
	}
}

func TestInstallBinary(t *testing.T) {
	path, err := installBinary()
	if err != nil {
		t.Fatalf("installBinary() returned unexpected error: %v", err)
	}

	if runtime.GOOS == "windows" {
		expectedDir := filepath.Join(os.Getenv("USERPROFILE"), "bin")
		expectedPath := filepath.Join(expectedDir, "gcnf.exe")
		if path != expectedPath {
			t.Errorf("installBinary() = %q, want %q", path, expectedPath)
		}
	} else {
		expectedPath := "/usr/local/bin/gcnf"
		if path != expectedPath {
			t.Errorf("installBinary() = %q, want %q", path, expectedPath)
		}
	}
}
