// Package utils provides file I/O, encoding, and environment variable resolution utilities.
package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/yaml.v3"
)

// Package-level compiled regexes.
var (
	// BasicPattern matches simple variable references like $VAR.
	BasicPattern = regexp.MustCompile(`\$(\w+)`)
	// BracedPattern matches braced variable references like ${VAR}, ${VAR:-default}, ${VAR:?error}.
	BracedPattern   = regexp.MustCompile(`\${(\w+)(?::-([^}]*))?(?::\?([^}]*))?}`)
	propertyPattern = regexp.MustCompile(`^([^=]+)=(.*)$`)
)

// NormalizePath expands ~ to the user's home directory and cleans the path.
func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		var suffix string
		if len(path) > 1 {
			suffix = path[2:]
		}
		usr, err := user.Current()
		if err != nil {
			// Fallback: try HOME or USERPROFILE env var
			home := os.Getenv("HOME")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			if home == "" {
				return filepath.Clean(path)
			}
			path = filepath.Join(home, suffix)
		} else {
			path = filepath.Join(usr.HomeDir, suffix)
		}
	}
	return filepath.Clean(path)
}

// EnsureDirectoryExists creates the directory if it does not exist, with 0700 permissions.
func EnsureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(dir, 0700); mkErr != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, mkErr)
			}
		} else {
			return fmt.Errorf("failed to stat directory %s: %w", dir, err)
		}
	}
	return nil
}

// FileExists returns true if the file exists and is accessible.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsCacheExpired returns true if the file at path is older than ttl.
// Returns false if ttl is 0 (disabled) or if the file cannot be stat'd.
func IsCacheExpired(path string, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) > ttl
}

// DeleteFile removes the file at path. Returns (true, nil) if deleted,
// (false, nil) if file doesn't exist, or (false, error) on failure.
func DeleteFile(path string) (bool, error) {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to delete %s: %w", path, err)
	}
	return true, nil
}

// LoadFileContentAsString reads the entire file as a string.
// If trim is true, leading/trailing whitespace is removed.
func LoadFileContentAsString(path string, trim bool) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	if trim {
		content = strings.TrimSpace(content)
	}
	return content, nil
}

// LoadFileContentAsJson reads a JSON file and returns it as a map.
// Returns nil if the file cannot be read or parsed.
func LoadFileContentAsJson(path string) map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil
	}
	return data
}

// MergeMaps returns a new map containing all entries from a and b.
// Values from b override values from a for duplicate keys.
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

// LoadProperties parses a properties file content into a key-value map.
// Lines without '=' (e.g., section headers) are silently skipped.
func LoadProperties(content string) (map[string]string, error) {
	props := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		if matches := propertyPattern.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			props[key] = value
		}
	}
	return props, nil
}

// PrintYAML marshals data to YAML and prints it to stdout.
func PrintYAML(data interface{}) {
	out, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to marshal YAML: %v", err)
	}
	fmt.Println(string(out))
}

// ToBase64 encodes a string to base64.
func ToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// FromBase64 decodes a base64 string. Returns an error if the input is not valid base64.
func FromBase64(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("base64 decode failed: %w", err)
	}
	return string(decoded), nil
}

// Coalesce returns the first non-empty string from the given values.
func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// IsValidBase64 returns true if s is a valid base64-encoded string.
func IsValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// SplitLines splits a string by newline characters.
func SplitLines(s string) []string {
	return strings.Split(s, "\n")
}

// StripWhitespace trims leading and trailing whitespace.
func StripWhitespace(s string) string {
	return strings.TrimSpace(s)
}

// IsCommentLine returns true if the line is a comment (starts with #, //, or ;).
func IsCommentLine(s string) bool {
	return strings.HasPrefix(s, "#") || strings.HasPrefix(s, "//") || strings.HasPrefix(s, ";")
}

// ContainsEqualSign returns true if s contains an '=' character.
func ContainsEqualSign(s string) bool {
	return strings.Contains(s, "=")
}

// SplitKeyValue splits a "key=value" string into its parts.
// If no '=' is found, returns the trimmed string as key and empty string as value.
func SplitKeyValue(s string) (string, string) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// stripQuotes removes surrounding double or single quotes from a string.
func stripQuotes(val string) string {
	if len(val) >= 2 {
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			return val[1 : len(val)-1]
		}
		if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
			return val[1 : len(val)-1]
		}
	}
	return val
}

// ResolveValue resolves environment variable references and gcnf:// URLs in a value string.
// Supports $VAR, ${VAR}, ${VAR:-default}, and ${VAR:?error} patterns.
// Variables are first looked up in envVars, then in os environment.
func ResolveValue(envVars map[string]string, value string, resolveGCNFURL func(string) (string, error)) string {
	isQuoted := len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")

	// Remove quotes for processing
	if isQuoted {
		value = value[1 : len(value)-1]
	}

	// Process ${VAR}, ${VAR:-default}, ${VAR:?error} FIRST (before $VAR)
	// to avoid BasicPattern matching inside braced expressions.
	value = BracedPattern.ReplaceAllStringFunc(value, func(s string) string {
		matches := BracedPattern.FindStringSubmatch(s)
		varName := matches[1]
		defaultValue := matches[2]
		errorMessage := matches[3]
		if val, ok := envVars[varName]; ok {
			return stripQuotes(val)
		} else if val := os.Getenv(varName); val != "" {
			return val
		} else if defaultValue != "" {
			return defaultValue
		} else if errorMessage != "" {
			log.Fatalf("%s", errorMessage)
		}
		return ""
	})

	// Replace $VAR (simple variable references)
	value = BasicPattern.ReplaceAllStringFunc(value, func(s string) string {
		varName := s[1:]
		if val, ok := envVars[varName]; ok {
			return stripQuotes(val)
		}
		return os.Getenv(varName)
	})

	// Handle gcnf:// values
	if strings.HasPrefix(value, "gcnf://") {
		resolvedValue, err := resolveGCNFURL(value)
		if err != nil {
			log.Fatalf("Error resolving gcnf URL: %v", err)
		}
		value = resolvedValue
	}

	// Restore quotes if necessary
	if isQuoted {
		value = fmt.Sprintf("\"%s\"", value)
	}

	return value
}

// LoadEmbeddedFile loads a file from the executable directory or working directory.
func LoadEmbeddedFile(path string) ([]byte, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)
	fullPath := filepath.Join(exeDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// GetUserTokenClient creates an authenticated HTTP client using OAuth2 user credentials.
func GetUserTokenClient(clientSecret []byte, tokenFile string) *http.Client {
	// If modifying these scopes, delete your previously saved token at ~/.gcnf/.token.
	cnf, err := google.ConfigFromJSON(clientSecret, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return GetHttpClient(cnf, tokenFile)
}

// GetHttpClient creates an HTTP client from a saved OAuth2 token file.
// Returns nil if the token file cannot be read.
func GetHttpClient(config *oauth2.Config, tokenFile string) *http.Client {
	// The token file stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first time.
	tok, err := TokenFromFile(tokenFile)
	if err != nil {
		return nil
	}
	return config.Client(context.Background(), tok)
}

// TokenFromFile retrieves a Token from a given file path.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// WriteStringToFile writes a string to a file with 0600 permissions.
// Creates parent directories if they don't exist.
func WriteStringToFile(file string, value string) error {
	if err := EnsureDirectoryExists(filepath.Dir(file)); err != nil {
		return err
	}
	return os.WriteFile(file, []byte(value), 0600)
}
