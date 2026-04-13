package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "audn-cli",
		Short: "Audn.ai VAST CLI for CI/CD gating and developer workflows",
		Long:  "audn-cli integrates Audn.ai VAST (Voice Agent Security Testing) into CI/CD pipelines and local developer flows.",
	}

	// Global flags
	bearerTokenFlag     string
	jsonOutputFlag      bool
	apiURLFlag          string
	m2mClientIDFlag     string
	m2mClientSecretFlag string
	apiSecretFlag       string
)

// Execute runs the root command.
func Execute() error { return rootCmd.Execute() }

func init() {
	// Ensure color output is not disabled by CI unless explicitly set
	if os.Getenv("NO_COLOR") != "" {
		fmt.Print("") // noop placeholder; color control is handled inline
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&bearerTokenFlag, "bearer-token", "", "Use Authorization: Bearer token (alias to AUDN_BEARER_TOKEN)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutputFlag, "json", false, "Output results as JSON where applicable")
	rootCmd.PersistentFlags().StringVar(&apiURLFlag, "api-url", "", "API base URL (default: https://audn.ai)")
	rootCmd.PersistentFlags().StringVar(&m2mClientIDFlag, "m2m-client-id", "", "Auth0 M2M Client ID (alias to AUTH0_M2M_CLIENT_ID)")
	rootCmd.PersistentFlags().StringVar(&m2mClientSecretFlag, "m2m-client-secret", "", "Auth0 M2M Client Secret (alias to AUTH0_M2M_CLIENT_SECRET)")
	rootCmd.PersistentFlags().StringVar(&apiSecretFlag, "api-secret", "", "API secret key (alias to AUDN_API_SECRET)")
}
