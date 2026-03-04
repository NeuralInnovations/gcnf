// Package config manages application configuration loading from environment variables, files, and tokens.
package config

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gcnf/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	// EnvGoogleCredentialBase64 is the env var for base64-encoded Google service account credentials.
	EnvGoogleCredentialBase64 = "GCNF_GOOGLE_CREDENTIAL_BASE64"
	// EnvGoogleSheetID is the env var for the Google Sheets document ID.
	EnvGoogleSheetID = "GCNF_GOOGLE_SHEET_ID"
	// EnvStoreConfigFile is the env var for the local config file path.
	EnvStoreConfigFile = "GCNF_STORE_CONFIG_FILE"
	// EnvToken is the env var for the composite GCNF token.
	EnvToken = "GCNF_TOKEN"
	// EnvCacheTTL is the env var for cache time-to-live duration (e.g. "30m", "1h").
	EnvCacheTTL = "GCNF_CACHE_TTL"
)

// Configs holds all configuration values for the application.
type Configs struct {
	GoogleCredentialBase64   string
	GoogleSheetID            string
	ConfigFile               string
	Scopes                   []string
	UserTokenFile            string
	UserGoogleSheetIDFile    string
	UserStoreConfigFile      string
	TokenValue               string
	GoogleClientTokenContent []byte
	CacheTTL                 time.Duration
}

// GetClientSecret returns the OAuth2 client secret, preferring the embedded content.
func (c *Configs) GetClientSecret() []byte {
	clientSecret, err := utils.LoadEmbeddedFile("client_secret.json")
	if err != nil {
		clientSecret = c.GoogleClientTokenContent
	}
	return clientSecret
}

// NewConfigs creates a new Configs instance, loading values from environment
// variables, GCNF_TOKEN, and local config files with appropriate precedence.
func NewConfigs(googleClientTokenContent []byte) *Configs {
	c := &Configs{
		Scopes:                   []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
		UserTokenFile:            utils.NormalizePath("~/.gcnf/.token"),
		UserGoogleSheetIDFile:    utils.NormalizePath("~/.gcnf/.google_sheet_id"),
		UserStoreConfigFile:      utils.NormalizePath("~/.gcnf/.gcnf_config.json"),
		GoogleClientTokenContent: googleClientTokenContent,
	}

	// Ensure the token directory exists
	if err := utils.EnsureDirectoryExists(filepath.Dir(c.UserTokenFile)); err != nil {
		log.Printf("Warning: could not create token directory: %v", err)
	}

	// Load from environment variables
	tokenStr := os.Getenv(EnvToken)
	var tkn *ProjectToken
	if tokenStr != "" {
		tkn = DecodeToken(tokenStr)
	}
	if tkn == nil {
		tkn = &ProjectToken{}
		tkn.GoogleSheetID, _ = utils.LoadFileContentAsString(c.UserGoogleSheetIDFile, true)
		tkn.StoreConfigFile, _ = utils.LoadFileContentAsString(c.UserStoreConfigFile, true)
	}
	c.GoogleCredentialBase64 = utils.Coalesce(os.Getenv(EnvGoogleCredentialBase64), tkn.GoogleCredBase64)
	c.GoogleSheetID = utils.Coalesce(os.Getenv(EnvGoogleSheetID), tkn.GoogleSheetID)
	c.ConfigFile = utils.Coalesce(os.Getenv(EnvStoreConfigFile), tkn.StoreConfigFile, "./gcnf_config.json")
	c.ConfigFile = utils.NormalizePath(c.ConfigFile)

	if ttlStr := os.Getenv(EnvCacheTTL); ttlStr != "" {
		ttl, err := time.ParseDuration(ttlStr)
		if err != nil {
			log.Printf("Warning: invalid %s value %q, ignoring: %v", EnvCacheTTL, ttlStr, err)
		} else {
			c.CacheTTL = ttl
		}
	}

	return c
}

// GetCredentialsStatus returns the type of credentials configured: "service_account", "user_token", or "none".
func (c *Configs) GetCredentialsStatus() string {
	if c.GoogleCredentialBase64 != "" && utils.IsValidBase64(c.GoogleCredentialBase64) {
		return "service_account"
	}
	if utils.FileExists(c.UserTokenFile) {
		return "user_token"
	}
	return "none"
}

// GetBase64CredentialStatus returns the status of base64 credentials: "available", "empty", or "invalid".
func (c *Configs) GetBase64CredentialStatus() string {
	if c.GoogleCredentialBase64 == "" {
		return "empty"
	}
	if utils.IsValidBase64(c.GoogleCredentialBase64) {
		return "available"
	}
	return "invalid"
}

// GetUserTokenStatus returns the validity of the user's OAuth2 token: "valid", "invalid", or "not_found".
func (c *Configs) GetUserTokenStatus() string {
	if utils.FileExists(c.UserTokenFile) {
		clientSecret := c.GetClientSecret()

		// If modifying these scopes, delete your previously saved token at ~/.gcnf/.token.
		config, err := google.ConfigFromJSON(clientSecret, sheets.SpreadsheetsReadonlyScope)
		if err != nil {
			return "invalid"
		}

		client := getClient(config, c.UserTokenFile)
		if client == nil {
			return "invalid"
		}

		_, err = sheets.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			return "invalid"
		}
		return "valid"
	}
	return "not_found"
}

func getClient(config *oauth2.Config, tokFile string) *http.Client {
	// The token file at ~/.gcnf/.token stores the user's access and refresh tokens,
	// and is created automatically when the authorization flow completes for the first time.
	tok, err := utils.TokenFromFile(tokFile)
	if err != nil {
		return nil
	}
	return config.Client(context.Background(), tok)
}
