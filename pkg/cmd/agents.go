package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/audn-ai/audn-cli/pkg/internal/api"
	"github.com/audn-ai/audn-cli/pkg/internal/config"
	"github.com/spf13/cobra"
)

var (
	agentsPage   int
	agentsLimit  int
	agentsStatus string

	agentID       string
	agentName     string
	agentPlatform string
	agentPhone    string
	agentDesc     string
)

var agentsCmd = &cobra.Command{Use: "agents", Short: "Manage voice agents"}

var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" {
			env.BearerToken = bearerTokenFlag
		}
		if m2mClientIDFlag != "" {
			env.M2MClientID = m2mClientIDFlag
		}
		if m2mClientSecretFlag != "" {
			env.M2MClientSecret = m2mClientSecretFlag
		}
		if apiSecretFlag != "" {
			env.APISecret = apiSecretFlag
		}
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		res, err := c.AgentsList(ctx, agentsPage, agentsLimit, agentsStatus)
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		for _, a := range res.Agents {
			fmt.Printf("%s\t%s\t%s\t%s\n", a.ID, a.Name, a.Platform, a.Status)
		}
		fmt.Printf("Total: %d\n", res.Pagination.Total)
		return nil
	},
}

var agentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		if agentName == "" || agentPlatform == "" {
			return fmt.Errorf("--name and --platform are required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" {
			env.BearerToken = bearerTokenFlag
		}
		if m2mClientIDFlag != "" {
			env.M2MClientID = m2mClientIDFlag
		}
		if m2mClientSecretFlag != "" {
			env.M2MClientSecret = m2mClientSecretFlag
		}
		if apiSecretFlag != "" {
			env.APISecret = apiSecretFlag
		}
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ag, err := c.AgentCreate(ctx, api.AgentCreateReq{Name: agentName, Platform: agentPlatform, PhoneNumber: agentPhone, Description: agentDesc})
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(ag, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("Created: %s\t%s\n", ag.ID, ag.Name)
		}
		return nil
	},
}

var agentsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		if agentID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" {
			env.BearerToken = bearerTokenFlag
		}
		if m2mClientIDFlag != "" {
			env.M2MClientID = m2mClientIDFlag
		}
		if m2mClientSecretFlag != "" {
			env.M2MClientSecret = m2mClientSecretFlag
		}
		if apiSecretFlag != "" {
			env.APISecret = apiSecretFlag
		}
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ag, err := c.AgentGet(ctx, agentID)
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(ag, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("%+v\n", ag)
		}
		return nil
	},
}

var agentsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		if agentID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" {
			env.BearerToken = bearerTokenFlag
		}
		if m2mClientIDFlag != "" {
			env.M2MClientID = m2mClientIDFlag
		}
		if m2mClientSecretFlag != "" {
			env.M2MClientSecret = m2mClientSecretFlag
		}
		if apiSecretFlag != "" {
			env.APISecret = apiSecretFlag
		}
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ag, err := c.AgentUpdate(ctx, api.AgentUpdateReq{ID: agentID, Name: agentName, Platform: agentPlatform, PhoneNumber: agentPhone, Description: agentDesc})
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(ag, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("Updated: %s\t%s\n", ag.ID, ag.Name)
		}
		return nil
	},
}

var agentsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		if agentID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" {
			env.BearerToken = bearerTokenFlag
		}
		if m2mClientIDFlag != "" {
			env.M2MClientID = m2mClientIDFlag
		}
		if m2mClientSecretFlag != "" {
			env.M2MClientSecret = m2mClientSecretFlag
		}
		if apiSecretFlag != "" {
			env.APISecret = apiSecretFlag
		}
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := c.AgentDelete(ctx, agentID); err != nil {
			return err
		}
		if jsonOutputFlag {
			fmt.Println(`{"deleted":true}`)
		} else {
			fmt.Println("Deleted")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentsListCmd, agentsCreateCmd, agentsGetCmd, agentsUpdateCmd, agentsDeleteCmd)

	agentsListCmd.Flags().IntVar(&agentsPage, "page", 1, "Page number")
	agentsListCmd.Flags().IntVar(&agentsLimit, "limit", 10, "Page size")
	agentsListCmd.Flags().StringVar(&agentsStatus, "status", "", "Filter by status")

	for _, cmd := range []*cobra.Command{agentsCreateCmd, agentsUpdateCmd} {
		cmd.Flags().StringVar(&agentName, "name", "", "Agent name")
		cmd.Flags().StringVar(&agentPlatform, "platform", "", "Agent platform (twilio/genesys/amazon_connect/custom)")
		cmd.Flags().StringVar(&agentPhone, "phone", "", "Agent phone number (E.164)")
		cmd.Flags().StringVar(&agentDesc, "description", "", "Agent description")
	}
	for _, cmd := range []*cobra.Command{agentsGetCmd, agentsUpdateCmd, agentsDeleteCmd} {
		cmd.Flags().StringVar(&agentID, "id", "", "Agent ID")
	}
}
