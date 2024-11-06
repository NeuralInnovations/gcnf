package googleapi

import (
	"context"
	"encoding/json"
	"fmt"
	"gcnf/internal/config"
	"gcnf/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func GoogleLoginCommand(configs *config.Configs) {
	_ = context.Background()

	clientSecret := configs.GetClientSecret()

	conf, err := google.ConfigFromJSON(clientSecret, configs.Scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	tok, err := getTokenFromWeb(conf)
	if err != nil {
		log.Fatalf("Unable to get token from web: %v", err)
	}

	utils.EnsureDirectoryExists(filepath.Dir(configs.UserTokenFile))

	f, err := os.OpenFile(configs.UserTokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tok)
	fmt.Println("Login successful.")
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default: // "linux" and others
		cmd = "xdg-open"
	}

	if cmd == "rundll32" {
		args = append(args, url)
		return exec.Command(cmd, args...).Start()
	} else {
		args = append(args, url)
		return exec.Command(cmd, args...).Start()
	}
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	// Set up a local server to receive the authorization code
	listener, err := net.Listen("tcp", "localhost:0") // Use a random available port
	if err != nil {
		return nil, fmt.Errorf("could not start local server: %v", err)
	}
	defer listener.Close()

	// Update the redirect URL to the local server's address
	redirectURL := fmt.Sprintf("http://%s/", listener.Addr().String())
	config.RedirectURL = redirectURL

	// Generate the authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Open the authorization URL in the user's default browser
	fmt.Println("Opening browser for authorization...")
	err = openBrowser(authURL)
	if err != nil {
		fmt.Printf("Please open the following URL in your browser:\n%v\n", authURL)
	}

	// Channel to receive the authorization code
	codeCh := make(chan string)

	// Start a server to handle the callback
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Get the authorization code from the query parameters
			code := r.URL.Query().Get("code")
			if code == "" {
				http.Error(w, "Authorization code not found", http.StatusBadRequest)
				return
			}
			fmt.Fprintf(w, "Authorization successful. You can close this window.")
			codeCh <- code
		})
		http.Serve(listener, nil)
	}()

	// Wait for the authorization code
	code := <-codeCh

	// Exchange the authorization code for an access token
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange token: %v", err)
	}
	return tok, nil
}
