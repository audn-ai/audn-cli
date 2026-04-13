package auth

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/audn-ai/audn-cli/pkg/internal/config"
)

func TestRefreshTokens(t *testing.T) {
    // Mock Auth0 token endpoint
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/oauth/token" || r.Method != http.MethodPost {
            http.Error(w, "not found", 404)
            return
        }
        if err := r.ParseForm(); err != nil { t.Fatalf("parse form: %v", err) }
        if r.Form.Get("grant_type") != "refresh_token" { t.Fatalf("grant type") }
        if r.Form.Get("client_id") != "cid" { t.Fatalf("client id") }
        if r.Form.Get("refresh_token") != "rt" { t.Fatalf("refresh token") }
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"access_token":"new_at","id_token":"new_id","refresh_token":"new_rt","expires_in":3600,"token_type":"Bearer"}`))
    }))
    defer srv.Close()

    env := config.Env{ Auth0Domain: srv.URL, Auth0ClientID: "cid" }
    tr, err := RefreshTokens(context.Background(), env, "rt")
    if err != nil { t.Fatalf("refresh failed: %v", err) }
    if tr.AccessToken != "new_at" || tr.IDToken != "new_id" || tr.RefreshToken != "new_rt" || tr.ExpiresIn != 3600 { t.Fatalf("unexpected: %#v", tr) }
}

func TestEnsureValidCredentials_NoRefreshNeeded(t *testing.T) {
    cred := &config.Credentials{ AccessToken: "at", RefreshToken: "rt", ExpiresAt: 32503680000000 } // year 3000
    env := config.Env{}
    out, err := EnsureValidCredentials(context.Background(), env, cred)
    if err != nil || out != cred { t.Fatalf("expected unchanged credentials") }
}
