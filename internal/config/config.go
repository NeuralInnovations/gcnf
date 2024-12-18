package config

import (
	"context"
	"gcnf/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
	"path/filepath"
)

const (
	EnvGoogleCredentialBase64 = "GCNF_GOOGLE_CREDENTIAL_BASE64"
	EnvGoogleSheetID          = "GCNF_GOOGLE_SHEET_ID"
	EnvStoreConfigFile        = "GCNF_STORE_CONFIG_FILE"
	EnvToken                  = "GCNF_TOKEN"
)

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
}

func (c *Configs) GetClientSecret() []byte {
	clientSecret, err := utils.LoadEmbeddedFile("client_secret.json")
	if err != nil {
		clientSecret = c.GoogleClientTokenContent
	}
	return clientSecret
}

func NewConfigs(googleClientTokenContent []byte) *Configs {
	c := &Configs{
		Scopes:                   []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
		UserTokenFile:            utils.NormalizePath("~/.gcnf/.token"),
		UserGoogleSheetIDFile:    utils.NormalizePath("~/.gcnf/.google_sheet_id"),
		UserStoreConfigFile:      utils.NormalizePath("~/.gcnf/.gcnf_config.json"),
		GoogleClientTokenContent: googleClientTokenContent,
	}

	// Ensure the token directory exists
	utils.EnsureDirectoryExists(filepath.Dir(c.UserTokenFile))

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

	return c
}

func (c *Configs) GetCredentialsStatus() string {
	if c.GoogleCredentialBase64 != "" && utils.IsValidBase64(c.GoogleCredentialBase64) {
		return "service_account"
	}
	if utils.FileExists(c.UserTokenFile) {
		return "user_token"
	}
	return "none"
}

func (c *Configs) GetBase64CredentialStatus() string {
	if c.GoogleCredentialBase64 == "" {
		return "empty"
	}
	if utils.IsValidBase64(c.GoogleCredentialBase64) {
		return "available"
	}
	return "invalid"
}

func (c *Configs) GetUserTokenStatus() string {
	if utils.FileExists(c.UserTokenFile) {
		clientSecret := c.GetClientSecret()

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(clientSecret, sheets.SpreadsheetsReadonlyScope)
		if err != nil {
			return "invalid"
		}

		// Get a token, save the token, then create a client
		client := getClient(config, c.UserTokenFile)
		if client == nil {
			return "invalid"
		}

		// Create a new Sheets service
		_, err = sheets.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			return "invalid"
		}
		return "valid"
	}
	return "not_found"
}

func getClient(config *oauth2.Config, tokFile string) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := utils.TokenFromFile(tokFile)
	if err != nil {
		return nil
	}
	return config.Client(context.Background(), tok)
}
