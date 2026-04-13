package auth

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "encoding/base64"
    "net/http"
    "net/url"
    "os/exec"
    "runtime"
    "strings"
    "time"

    "github.com/audn-ai/audn-cli/pkg/internal/config"
)

type DeviceCodeResponse struct {
    DeviceCode              string `json:"device_code"`
    UserCode                string `json:"user_code"`
    VerificationURI         string `json:"verification_uri"`
    VerificationURIComplete string `json:"verification_uri_complete"`
    ExpiresIn               int    `json:"expires_in"`
    Interval                int    `json:"interval"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    IDToken      string `json:"id_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    TokenType    string `json:"token_type"`
}

func StartDeviceCode(ctx context.Context, env config.Env) (*DeviceCodeResponse, error) {
    endpoint := strings.TrimSuffix(env.Auth0Domain, "/") + "/oauth/device/code"
    data := url.Values{}
    data.Set("client_id", env.Auth0ClientID)
    data.Set("scope", "openid profile email offline_access")
    data.Set("audience", env.Auth0Audience)

    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        b, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("device code error: %s", string(b))
    }
    var out DeviceCodeResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return nil, err }
    return &out, nil
}

func PollForTokens(ctx context.Context, env config.Env, dc *DeviceCodeResponse) (*TokenResponse, string, error) {
    endpoint := strings.TrimSuffix(env.Auth0Domain, "/") + "/oauth/token"
    interval := dc.Interval
    if interval <= 0 { interval = 5 }
    ticker := time.NewTicker(time.Duration(interval) * time.Second)
    defer ticker.Stop()

    deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
    for {
        select {
        case <-ctx.Done():
            return nil, "", ctx.Err()
        case <-ticker.C:
            if time.Now().After(deadline) {
                return nil, "", errors.New("device code expired; please run login again")
            }
            data := url.Values{}
            data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
            data.Set("device_code", dc.DeviceCode)
            data.Set("client_id", env.Auth0ClientID)

            req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
            req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
            resp, err := http.DefaultClient.Do(req)
            if err != nil { return nil, "", err }
            body, _ := io.ReadAll(resp.Body)
            resp.Body.Close()

            if resp.StatusCode == 200 {
                var tr TokenResponse
                if err := json.Unmarshal(body, &tr); err != nil { return nil, "", err }
                email := extractEmailFromIDToken(tr.IDToken)
                return &tr, email, nil
            }
            // Handle polling errors
            if resp.StatusCode == 403 || resp.StatusCode == 400 {
                if strings.Contains(string(body), "authorization_pending") {
                    // keep polling
                    continue
                }
                if strings.Contains(string(body), "slow_down") {
                    interval += 5
                    ticker.Reset(time.Duration(interval) * time.Second)
                    continue
                }
                if strings.Contains(string(body), "expired_token") || strings.Contains(string(body), "access_denied") {
                    return nil, "", fmt.Errorf("login failed: %s", string(body))
                }
            }
            return nil, "", fmt.Errorf("unexpected token response: %d %s", resp.StatusCode, string(body))
        }
    }
}

// Very light-weight open browser helper with no external deps
func OpenBrowser(url string) {
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "darwin":
        cmd = exec.Command("open", url)
    case "windows":
        cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
    default:
        cmd = exec.Command("xdg-open", url)
    }
    _ = cmd.Start()
}

// extractEmailFromIDToken decodes the JWT payload (without verification) to read email
func extractEmailFromIDToken(idToken string) string {
    parts := strings.Split(idToken, ".")
    if len(parts) != 3 { return "" }
    // base64url decode
    payload, err := base64URLDecode(parts[1])
    if err != nil { return "" }
    var claims map[string]any
    if err := json.Unmarshal(payload, &claims); err != nil { return "" }
    if v, ok := claims["email"].(string); ok { return v }
    return ""
}

func base64URLDecode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }
