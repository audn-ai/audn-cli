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

// EnsureValidCredentials refreshes tokens if expired (or near expiry) and persists them.
func EnsureValidCredentials(ctx context.Context, env config.Env, cred *config.Credentials) (*config.Credentials, error) {
    if cred == nil { return nil, fmt.Errorf("no credentials") }
    // treat expiry as ms; refresh if exp < now + 30s skew
    now := time.Now().UnixMilli()
    if cred.ExpiresAt != 0 && cred.ExpiresAt > now+30000 {
        return cred, nil
    }
    if strings.TrimSpace(cred.RefreshToken) == "" {
        return cred, nil // cannot refresh; let server attempt and fail if needed
    }
    tr, err := RefreshTokens(ctx, env, cred.RefreshToken)
    if err != nil { return nil, err }
    // Update credentials
    if tr.AccessToken != "" { cred.AccessToken = tr.AccessToken }
    if tr.IDToken != "" { cred.IDToken = tr.IDToken }
    if tr.RefreshToken != "" { cred.RefreshToken = tr.RefreshToken }
    if tr.ExpiresIn > 0 { cred.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second).UnixMilli() }
    if err := config.SaveCredentials(*cred); err != nil { return nil, err }
    return cred, nil
}

// RefreshTokens performs Auth0 refresh_token grant.
func RefreshTokens(ctx context.Context, env config.Env, refreshToken string) (*TokenResponse, error) {
    endpoint := strings.TrimSuffix(env.Auth0Domain, "/") + "/oauth/token"
    form := url.Values{}
    form.Set("grant_type", "refresh_token")
    form.Set("client_id", env.Auth0ClientID)
    form.Set("refresh_token", refreshToken)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("refresh failed: %s", string(body))
    }
    var tr TokenResponse
    if err := json.Unmarshal(body, &tr); err != nil { return nil, err }
    return &tr, nil
}

