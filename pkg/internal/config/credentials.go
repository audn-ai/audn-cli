package config

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "runtime"
)

const credentialsFile = "credentials.json"

func SaveCredentials(c Credentials) error {
    dir, err := configDir()
    if err != nil { return err }
    if err := os.MkdirAll(dir, 0o700); err != nil { return err }
    b, err := json.MarshalIndent(c, "", "  ")
    if err != nil { return err }
    path := filepath.Join(dir, credentialsFile)
    // Use 0600 on POSIX; Windows will ignore but it's fine
    if err := os.WriteFile(path, b, 0o600); err != nil { return err }
    return nil
}

func ReadCredentials() (*Credentials, error) {
    dir, err := configDir()
    if err != nil { return nil, err }
    path := filepath.Join(dir, credentialsFile)
    b, err := os.ReadFile(path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) { return nil, nil }
        return nil, err
    }
    var c Credentials
    if err := json.Unmarshal(b, &c); err != nil { return nil, err }
    return &c, nil
}

func DeleteCredentials() error {
    dir, err := configDir()
    if err != nil { return err }
    path := filepath.Join(dir, credentialsFile)
    if err := os.Remove(path); err != nil {
        if errors.Is(err, os.ErrNotExist) { return nil }
        return err
    }
    return nil
}

// Ensure owner-only perms if possible (POSIX)
func CheckPermissions() error {
    if runtime.GOOS == "windows" { return nil }
    dir, err := configDir()
    if err != nil { return err }
    path := filepath.Join(dir, credentialsFile)
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    mode := info.Mode().Perm()
    if mode&0o077 != 0 {
        return fmt.Errorf("credentials file %s permissions too open: %v; expected 0600", path, fs.FileMode(mode))
    }
    return nil
}

