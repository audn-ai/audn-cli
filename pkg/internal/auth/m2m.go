package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/audn-ai/audn-cli/pkg/internal/config"
)

// M2MTokenResponse represents the response from Auth0 M2M token endpoint
type M2MTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// GetM2MToken fetches an access token using Auth0 Machine-to-Machine flow
func GetM2MToken(ctx context.Context, env config.Env) (*M2MTokenResponse, error) {
	if env.Auth0Domain == "" {
		return nil, fmt.Errorf("AUTH0_DOMAIN or AUTH0_ISSUER_BASE_URL is required for M2M authentication")
	}
	if env.M2MClientID == "" {
		return nil, fmt.Errorf("AUTH0_M2M_CLIENT_ID is required for M2M authentication")
	}
	if env.M2MClientSecret == "" {
		return nil, fmt.Errorf("AUTH0_M2M_CLIENT_SECRET is required for M2M authentication")
	}
	if env.Auth0Audience == "" {
		return nil, fmt.Errorf("AUTH0_AUDIENCE is required for M2M authentication")
	}

	// Prepare token request
	tokenURL := fmt.Sprintf("%s/oauth/token", strings.TrimSuffix(env.Auth0Domain, "/"))

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", env.M2MClientID)
	data.Set("client_secret", env.M2MClientSecret)
	data.Set("audience", env.Auth0Audience)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request M2M token: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp M2MTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	return &tokenResp, nil
}
