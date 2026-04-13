package cmd

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/audn-ai/audn-cli/pkg/internal/auth"
    "github.com/audn-ai/audn-cli/pkg/internal/config"
    "github.com/spf13/cobra"
)

var (
    noBrowserFlag    bool
    loginTimeoutFlag int
    usePKCEFlag      bool
    redirectURIFlag  string
    portFlag         int
)

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Authenticate via Auth0 Device Code flow",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := config.LoadEnv()
        if err != nil {
            return err
        }

        if cfg.Auth0Domain == "" || cfg.Auth0ClientID == "" || cfg.Auth0Audience == "" {
            return errors.New("missing Auth0 configuration: set AUTH0_ISSUER_BASE_URL, AUTH0_CLIENT_ID, AUTH0_AUDIENCE")
        }

        timeout := 15 * time.Minute
        if loginTimeoutFlag > 0 {
            timeout = time.Duration(loginTimeoutFlag) * time.Second
        }
        ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
        defer cancel()

        var tokens *auth.TokenResponse
        var email string

        if usePKCEFlag {
            // Authorization Code + PKCE fallback (no Device Flow needed)
            if redirectURIFlag == "" {
                if portFlag <= 0 { portFlag = 8765 }
                redirectURIFlag = fmt.Sprintf("http://127.0.0.1:%d/callback", portFlag)
            }
            fmt.Printf("Starting PKCE login using redirect URI: %s\n", redirectURIFlag)
            t, e, err := auth.LoginWithPKCE(ctx, cfg, redirectURIFlag, !noBrowserFlag)
            if err != nil { return err }
            tokens, email = t, e
        } else {
            // Device Code flow (preferred for CLIs)
            dc, err := auth.StartDeviceCode(ctx, cfg)
            if err != nil {
                return err
            }

            fmt.Printf("Your one-time device code is: %s\n", dc.UserCode)
            if noBrowserFlag {
                fmt.Printf("Go to %s and enter the code.\n", dc.VerificationURI)
            } else {
                fmt.Printf("Opening your browser for login...\n")
                fmt.Printf("Or, go to %s and enter the code.\n", dc.VerificationURI)
                auth.OpenBrowser(dc.VerificationURIComplete)
            }

            t, e, err := auth.PollForTokens(ctx, cfg, dc)
            if err != nil { return err }
            tokens, email = t, e
        }

        cred := config.Credentials{
            UserEmail:    email,
            AccessToken:  tokens.AccessToken,
            IDToken:      tokens.IDToken,
            RefreshToken: tokens.RefreshToken,
            // Store as milliseconds since epoch to align with server checks
            ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second).UnixMilli(),
        }
        if err := config.SaveCredentials(cred); err != nil {
            return err
        }

        fmt.Printf("✅ Login successful. Authenticated as %s.\n", email)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(loginCmd)
    loginCmd.Flags().BoolVar(&noBrowserFlag, "no-browser", false, "Do not open a browser automatically")
    loginCmd.Flags().IntVar(&loginTimeoutFlag, "login-timeout", 900, "Login timeout in seconds")
    loginCmd.Flags().BoolVar(&usePKCEFlag, "pkce", false, "Use Authorization Code + PKCE instead of Device Flow")
    loginCmd.Flags().StringVar(&redirectURIFlag, "redirect-uri", "", "Redirect URI for PKCE flow (must be allowed in Auth0 app)")
    loginCmd.Flags().IntVar(&portFlag, "port", 8765, "Local callback port for PKCE flow when redirect-uri not provided")
}
