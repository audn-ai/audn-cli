package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Env config used across commands
type Env struct {
	// Auth0 device flow
	Auth0Domain   string // AUTH0_ISSUER_BASE_URL
	Auth0ClientID string // AUTH0_CLIENT_ID
	Auth0Audience string // AUTH0_AUDIENCE

	// Auth0 M2M (for CI/CD)
	M2MClientID     string // AUTH0_M2M_CLIENT_ID
	M2MClientSecret string // AUTH0_M2M_CLIENT_SECRET

	// Gate settings
	AgentID        string
	APISecret      string
	BearerToken    string
	FailOnSeverity string
	MinGrade       string
	Wait           bool
	TimeoutSeconds int
	APIBaseURL     string
}

func LoadEnv() (Env, error) {
	domain := os.Getenv("AUTH0_ISSUER_BASE_URL")
	if domain == "" {
		if d := os.Getenv("AUTH0_DOMAIN"); d != "" {
			// Construct issuer base URL from domain
			if !strings.HasPrefix(d, "http://") && !strings.HasPrefix(d, "https://") {
				domain = "https://" + d
			} else {
				domain = d
			}
		} else {
			// Default to Audn.ai Auth0 tenant
			domain = "https://auth.audn.ai"
		}
	}
	clientID := os.Getenv("AUTH0_CLI_INTERACTIVE_CLIENT_ID")
	if strings.TrimSpace(clientID) == "" {
		clientID = os.Getenv("AUTH0_CLIENT_ID")
		if strings.TrimSpace(clientID) == "" {
			// Default to Audn.ai CLI client ID
			clientID = "gsf12oQDyphSpuRIt4KCfENZFYTeqNgf"
		}
	}
	audience := os.Getenv("AUTH0_AUDIENCE")
	if strings.TrimSpace(audience) == "" {
		// Default to Management API audience (most compatible)
		audience = domain + "/api/v2/"
	}

	return Env{
		Auth0Domain:     domain,
		Auth0ClientID:   clientID,
		Auth0Audience:   audience,
		M2MClientID:     os.Getenv("AUTH0_M2M_CLIENT_ID"),
		M2MClientSecret: os.Getenv("AUTH0_M2M_CLIENT_SECRET"),
		BearerToken:     os.Getenv("AUDN_BEARER_TOKEN"),
		APIBaseURL:      os.Getenv("AUDN_API_URL"),
	}, nil
}

// Credentials persisted after login
type Credentials struct {
	UserEmail    string `json:"user_email"`
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func configDir() (string, error) {
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			return "", errors.New("APPDATA not set")
		}
		return filepath.Join(base, "audn"), nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "audn"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "audn"), nil
}
