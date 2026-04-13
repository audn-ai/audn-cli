package config

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"
)

func TestCredentialsRW(t *testing.T) {
    // isolate config dir via XDG_CONFIG_HOME
    tmp := t.TempDir()
    if runtime.GOOS == "windows" {
        os.Setenv("APPDATA", tmp)
    } else {
        os.Setenv("XDG_CONFIG_HOME", tmp)
    }
    defer func(){
        os.Unsetenv("XDG_CONFIG_HOME")
        os.Unsetenv("APPDATA")
    }()

    c := Credentials{UserEmail: "u@example.com", AccessToken: "at", RefreshToken: "rt", ExpiresAt: 123}
    if err := SaveCredentials(c); err != nil {
        t.Fatalf("save: %v", err)
    }
    got, err := ReadCredentials()
    if err != nil { t.Fatalf("read: %v", err) }
    if got == nil || got.UserEmail != c.UserEmail || got.AccessToken != c.AccessToken {
        t.Fatalf("mismatch: %#v", got)
    }
    dir, _ := configDir()
    if runtime.GOOS != "windows" {
        info, err := os.Stat(filepath.Join(dir, "credentials.json"))
        if err != nil { t.Fatalf("stat: %v", err) }
        if info.Mode().Perm()&0o077 != 0 {
            t.Fatalf("expected 0600 perms, got %v", info.Mode().Perm())
        }
    }
    if err := DeleteCredentials(); err != nil { t.Fatalf("delete: %v", err) }
}

