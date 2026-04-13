package auth

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"

    "github.com/audn-ai/audn-cli/pkg/internal/config"
)

func LoginWithPKCE(ctx context.Context, env config.Env, redirectURI string, openBrowser bool) (*TokenResponse, string, error) {
    issuer := strings.TrimSuffix(env.Auth0Domain, "/")
    authz := issuer + "/authorize"
    token := issuer + "/oauth/token"

    verifier, challenge, err := genPKCE()
    if err != nil { return nil, "", err }

    state := randString(24)
    q := url.Values{}
    q.Set("client_id", env.Auth0ClientID)
    q.Set("response_type", "code")
    q.Set("redirect_uri", redirectURI)
    q.Set("scope", "openid profile email offline_access")
    if env.Auth0Audience != "" { q.Set("audience", env.Auth0Audience) }
    q.Set("code_challenge", challenge)
    q.Set("code_challenge_method", "S256")
    q.Set("state", state)
    authURL := authz + "?" + q.Encode()

    // Start local callback server
    codeCh := make(chan string, 1)
    errCh := make(chan error, 1)
    srv := &http.Server{Addr: getAddrFromRedirect(redirectURI)}
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Query().Get("state") != state {
            http.Error(w, "invalid state", http.StatusBadRequest)
            errCh <- fmt.Errorf("invalid state")
            return
        }
        code := r.URL.Query().Get("code")
        if code == "" {
            http.Error(w, "missing code", http.StatusBadRequest)
            errCh <- fmt.Errorf("missing code")
            return
        }
        io.WriteString(w, "Login successful. You can close this window.")
        codeCh <- code
        go func(){ _ = srv.Shutdown(context.Background()) }()
    })

    go func(){ _ = srv.ListenAndServe() }()

    fmt.Printf("Open this URL to authenticate:\n%s\n", authURL)
    if openBrowser { OpenBrowser(authURL) }

    // Wait for code or context cancellation
    var authCode string
    select {
    case <-ctx.Done():
        _ = srv.Shutdown(context.Background())
        return nil, "", ctx.Err()
    case err := <-errCh:
        _ = srv.Shutdown(context.Background())
        return nil, "", err
    case authCode = <-codeCh:
    }

    // Exchange code for tokens
    form := url.Values{}
    form.Set("grant_type", "authorization_code")
    form.Set("code", authCode)
    form.Set("redirect_uri", redirectURI)
    form.Set("client_id", env.Auth0ClientID)
    form.Set("code_verifier", verifier)
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, token, strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return nil, "", err }
    defer resp.Body.Close()
    b, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        return nil, "", fmt.Errorf("token exchange failed: %s", string(b))
    }
    var tr TokenResponse
    if err := json.Unmarshal(b, &tr); err != nil { return nil, "", err }
    email := extractEmailFromIDToken(tr.IDToken)
    return &tr, email, nil
}

func genPKCE() (verifier, challenge string, err error) {
    // Verifier: 43-128 chars, unpadded base64url
    b := make([]byte, 32)
    if _, err = rand.Read(b); err != nil { return }
    verifier = base64.RawURLEncoding.EncodeToString(b)
    sum := sha256.Sum256([]byte(verifier))
    challenge = base64.RawURLEncoding.EncodeToString(sum[:])
    return
}

func randString(n int) string {
    b := make([]byte, n)
    _, _ = rand.Read(b)
    return base64.RawURLEncoding.EncodeToString(b)
}

func getAddrFromRedirect(redirectURI string) string {
    u, _ := url.Parse(redirectURI)
    host := u.Host
    if host == "" {
        host = "127.0.0.1:8765"
    }
    return host
}
