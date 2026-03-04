package googleapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gcnf/internal/config"
	"gcnf/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleLoginCommand initiates the OAuth2 login flow and saves the token to disk.
func GoogleLoginCommand(configs *config.Configs) {
	clientSecret := configs.GetClientSecret()

	conf, err := google.ConfigFromJSON(clientSecret, configs.Scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	tok, err := getTokenFromWeb(conf)
	if err != nil {
		log.Fatalf("Unable to get token from web: %v", err)
	}

	if err := utils.EnsureDirectoryExists(filepath.Dir(configs.UserTokenFile)); err != nil {
		log.Fatalf("Unable to create token directory: %v", err)
	}

	f, err := os.OpenFile(configs.UserTokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(tok); err != nil {
		log.Fatalf("Failed to save oauth token: %v", err)
	}
	fmt.Println("Login successful.")
}

// openBrowser opens the given URL in the user's default browser.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

func getTokenFromWeb(oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("could not start local server: %v", err)
	}

	redirectURL := fmt.Sprintf("http://%s/", listener.Addr().String())
	oauthConfig.RedirectURL = redirectURL

	// Generate random state for CSRF protection
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		listener.Close()
		return nil, fmt.Errorf("could not generate state token: %v", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("Opening browser for authorization...")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Please open the following URL in your browser:\n%v\n", authURL)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Use a local ServeMux to avoid polluting the global DefaultServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			errCh <- fmt.Errorf("invalid OAuth state parameter")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			errCh <- fmt.Errorf("authorization code not found in callback")
			return
		}
		fmt.Fprintf(w, "Authorization successful. You can close this window.")
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		server.Close()
		return nil, err
	}

	// Gracefully shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	tok, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange token: %v", err)
	}
	return tok, nil
}
