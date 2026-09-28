package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/itokun99/blogger-go"
	"github.com/itokun99/blogger-mcp/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Run(ctx context.Context, cfg config.Config, port int) error {
	credsData, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return fmt.Errorf("auth: read credentials: %w", err)
	}

	oauthCfg, err := google.ConfigFromJSON(credsData, blogger.Scopes...)
	if err != nil {
		return fmt.Errorf("auth: parse credentials: %w", err)
	}

	state := randomState()
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	fmt.Fprintln(os.Stderr, "Please visit the following URL to authorize:")
	fmt.Fprintln(os.Stderr, authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\n", err)
	}

	codeCh := make(chan string, 1)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("auth: listen on port %d: %w", port, err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			errMsg := r.URL.Query().Get("error")
			if errMsg != "" {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, fmt.Sprintf("Authorization error: %s", errMsg))
				return
			}
			if code == "" {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "No code received")
				return
			}
			codeCh <- code
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "Authorization complete. You can close this tab.")
		}),
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "auth: server error: %v\n", err)
		}
	}()

	code, ok := <-codeCh
	if !ok {
		server.Close()
		return fmt.Errorf("auth: no authorization code received")
	}

	token, err := oauthCfg.Exchange(ctx, code)
	server.Close()
	if err != nil {
		return fmt.Errorf("auth: exchange code: %w", err)
	}

	tokenDir := filepath.Dir(cfg.TokenFile)
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		return fmt.Errorf("auth: create token directory: %w", err)
	}

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("auth: marshal token: %w", err)
	}

	if err := os.WriteFile(cfg.TokenFile, data, 0o600); err != nil {
		return fmt.Errorf("auth: write token: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Token saved to %s\n", cfg.TokenFile)
	return nil
}

func randomState() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(i)
	}
	return fmt.Sprintf("%x", b)
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
