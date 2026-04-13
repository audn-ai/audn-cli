package cmd

import (
    "errors"
    "fmt"

    "github.com/audn-ai/audn-cli/pkg/internal/config"
    "github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
    Use:   "whoami",
    Short: "Show the authenticated user's email",
    RunE: func(cmd *cobra.Command, args []string) error {
        cred, err := config.ReadCredentials()
        if err != nil {
            return err
        }
        if cred == nil || cred.UserEmail == "" {
            return errors.New("you are not logged in. Run 'audn-cli login'.")
        }
        fmt.Println(cred.UserEmail)
        return nil
    },
}

func init() { rootCmd.AddCommand(whoamiCmd) }

