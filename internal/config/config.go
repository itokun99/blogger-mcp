package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/itokun99/blogger-go"
)

type Config struct {
	CredentialsFile string
	TokenFile       string
}

func Load(credentialsFlag, tokenFlag string) Config {
	cfg := Config{}

	if credentialsFlag != "" {
		cfg.CredentialsFile = credentialsFlag
	} else if v := os.Getenv("BLOGGER_MCP_CREDENTIALS"); v != "" {
		cfg.CredentialsFile = v
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir = "."
		}
		cfg.CredentialsFile = filepath.Join(configDir, "blogger-mcp", "credentials.json")
	}

	if tokenFlag != "" {
		cfg.TokenFile = tokenFlag
	} else if v := os.Getenv("BLOGGER_MCP_TOKEN"); v != "" {
		cfg.TokenFile = v
	} else {
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir = "."
		}
		cfg.TokenFile = filepath.Join(configDir, "blogger-mcp", "token.json")
	}

	return cfg
}

func (c Config) Client(ctx context.Context) (*blogger.Client, error) {
	if _, err := os.Stat(c.CredentialsFile); err != nil {
		return nil, fmt.Errorf(
			"blogger_credentials: credentials file not found at %s (set via -credentials flag or BLOGGER_MCP_CREDENTIALS env var, or run blogger-mcp auth)",
			c.CredentialsFile,
		)
	}

	ts, err := blogger.TokenSourceFromJSON(c.CredentialsFile, c.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("blogger_credentials: %w (run blogger-mcp auth to obtain a token)", err)
	}

	httpClient := blogger.NewHTTPClient(ctx, ts)
	return blogger.NewRawClient(httpClient), nil
}
