package cmd

import (
    "fmt"

    "github.com/audn-ai/audn-cli/pkg/internal/config"
    "github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
    Use:   "logout",
    Short: "Remove stored credentials",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := config.DeleteCredentials(); err != nil {
            return err
        }
        fmt.Println("✅ Logged out. Credentials removed.")
        return nil
    },
}

func init() { rootCmd.AddCommand(logoutCmd) }

