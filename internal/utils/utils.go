package utils

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
)

func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = filepath.Join(dir, path[2:])
	}
	return filepath.Clean(path)
}

func EnsureDirectoryExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", dir)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return
		}
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DeleteFile(path string) bool {
	if FileExists(path) {
		os.Remove(path)
		return true
	}
	return false
}

func GetFileContent(path string) map[string]interface{} {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return nil
	}
	return data
}

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

func LoadProperties(content string) (map[string]string, error) {
	props := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	re := regexp.MustCompile(`^([^=]+)=(.*)$`)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			props[key] = value
		}
	}
	return props, nil
}

func PrintYAML(data interface{}) {
	out, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalf("Failed to marshal YAML: %v", err)
	}
	fmt.Println(string(out))
}

func ToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func FromBase64(s string) string {
	decoded, _ := base64.StdEncoding.DecodeString(s)
	return string(decoded)
}

func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func IsValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func SplitLines(s string) []string {
	return strings.Split(s, "\n")
}

func StripWhitespace(s string) string {
	return strings.TrimSpace(s)
}

func IsCommentLine(s string) bool {
	return strings.HasPrefix(s, "#") || strings.HasPrefix(s, "//") || strings.HasPrefix(s, ";")
}

func ContainsEqualSign(s string) bool {
	return strings.Contains(s, "=")
}

func SplitKeyValue(s string) (string, string) {
	parts := strings.SplitN(s, "=", 2)
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func ResolveValue(envVars map[string]string, value string, resolveGCNFURL func(string) (string, error)) string {
	_ = value
	isQuoted := strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")

	// Remove quotes for processing
	if isQuoted {
		value = value[1 : len(value)-1]
	}

	// Patterns
	basicPattern := regexp.MustCompile(`\$(\w+)`)
	bracedPattern := regexp.MustCompile(`\${(\w+)(?::-([^}]+))?(?::\?([^}]+))?}`)

	// Replace $VAR and ${VAR}
	value = basicPattern.ReplaceAllStringFunc(value, func(s string) string {
		varName := s[1:]
		val, ok := envVars[varName]
		if ok {
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = val[1 : len(val)-1]
			}
			if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
				val = val[1 : len(val)-1]
			}
		}
		val = os.Getenv(varName)
		return val
	})

	value = bracedPattern.ReplaceAllStringFunc(value, func(s string) string {
		matches := bracedPattern.FindStringSubmatch(s)
		varName := matches[1]
		defaultValue := matches[2]
		errorMessage := matches[3]
		if val, ok := envVars[varName]; ok {
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = val[1 : len(val)-1]
			}
			if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
				val = val[1 : len(val)-1]
			}
			return val
		} else if val := os.Getenv(varName); val != "" {
			return val
		} else if defaultValue != "" {
			return defaultValue
		} else if errorMessage != "" {
			log.Fatalf(errorMessage)
		}
		return ""
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

func LoadEmbeddedFile(path string) ([]byte, error) {
	// Try to load the file from the executable directory
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exeDir := filepath.Dir(exePath)
	fullPath := filepath.Join(exeDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		// Try to load from the working directory
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func GetUserTokenClient(clientSecret []byte, tokenFile string) *http.Client {
	// If modifying these scopes, delete your previously saved token.json.
	cnf, err := google.ConfigFromJSON(clientSecret, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Get a token, save the token, then create a client
	client := GetHttpClient(cnf, tokenFile)

	return client
}

func GetHttpClient(config *oauth2.Config, tokenFile string) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
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
