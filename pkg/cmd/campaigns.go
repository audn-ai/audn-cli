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
	campPage, campLimit        int
	campStatus, campAgentID    string
	campID, campName, campDesc string
	campScenarioIDs            []string
)

var campaignsCmd = &cobra.Command{Use: "campaigns", Short: "Manage campaigns"}

var campaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" { env.BearerToken = bearerTokenFlag }
		if m2mClientIDFlag != "" { env.M2MClientID = m2mClientIDFlag }
		if m2mClientSecretFlag != "" { env.M2MClientSecret = m2mClientSecretFlag }
		if apiSecretFlag != "" { env.APISecret = apiSecretFlag }
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		res, err := c.CampaignsList(ctx, campPage, campLimit, campStatus, campAgentID)
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(b))
		} else {
			for _, v := range res.Campaigns {
				fmt.Printf("%s\t%s\t%s\t%d/%d\n", v.ID, v.Name, v.Status, v.Passed, v.Total)
			}
			fmt.Printf("Total: %d\n", res.Pagination.Total)
		}
		return nil
	},
}

var campaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		if campName == "" || campAgentID == "" {
			return fmt.Errorf("--name and --agent-id required")
		}
		if len(campScenarioIDs) == 0 {
			return fmt.Errorf("--scenario-id required: at least one scenario must be specified for the campaign")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" { env.BearerToken = bearerTokenFlag }
		if m2mClientIDFlag != "" { env.M2MClientID = m2mClientIDFlag }
		if m2mClientSecretFlag != "" { env.M2MClientSecret = m2mClientSecretFlag }
		if apiSecretFlag != "" { env.APISecret = apiSecretFlag }
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		out, err := c.CampaignCreate(ctx, api.CampaignCreateReq{Name: campName, Description: campDesc, AgentID: campAgentID, ScenarioIDs: campScenarioIDs})
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("Created: %s\t%s\n", out.ID, out.Name)
		}
		return nil
	},
}

var campaignsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		if campID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" { env.BearerToken = bearerTokenFlag }
		if m2mClientIDFlag != "" { env.M2MClientID = m2mClientIDFlag }
		if m2mClientSecretFlag != "" { env.M2MClientSecret = m2mClientSecretFlag }
		if apiSecretFlag != "" { env.APISecret = apiSecretFlag }
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		out, err := c.GetCampaign(ctx, campID)
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("%+v\n", out)
		}
		return nil
	},
}

var campaignsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		if campID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" { env.BearerToken = bearerTokenFlag }
		if m2mClientIDFlag != "" { env.M2MClientID = m2mClientIDFlag }
		if m2mClientSecretFlag != "" { env.M2MClientSecret = m2mClientSecretFlag }
		if apiSecretFlag != "" { env.APISecret = apiSecretFlag }
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		patch := map[string]any{}
		if campName != "" {
			patch["name"] = campName
		}
		if campDesc != "" {
			patch["description"] = campDesc
		}
		out, err := c.CampaignUpdate(ctx, campID, patch)
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("Updated: %s\t%s\n", out.ID, out.Name)
		}
		return nil
	},
}

var campaignsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		if campID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" { env.BearerToken = bearerTokenFlag }
		if m2mClientIDFlag != "" { env.M2MClientID = m2mClientIDFlag }
		if m2mClientSecretFlag != "" { env.M2MClientSecret = m2mClientSecretFlag }
		if apiSecretFlag != "" { env.APISecret = apiSecretFlag }
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := c.CampaignDelete(ctx, campID); err != nil {
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

var campaignsExecuteCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute a campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		if campID == "" {
			return fmt.Errorf("--id is required")
		}
		env, _ := config.LoadEnv()
		if bearerTokenFlag != "" { env.BearerToken = bearerTokenFlag }
		if m2mClientIDFlag != "" { env.M2MClientID = m2mClientIDFlag }
		if m2mClientSecretFlag != "" { env.M2MClientSecret = m2mClientSecretFlag }
		if apiSecretFlag != "" { env.APISecret = apiSecretFlag }
		hdrs, err := api.ResolveAuthHeaders(env)
		if err != nil {
			return err
		}
		c := api.NewClient(firstNonEmpty(apiURLFlag, env.APIBaseURL, "https://audn.ai"), hdrs)
		ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(env.TimeoutSeconds)*time.Second)
		defer cancel()
		execRes, err := c.ExecuteCampaign(ctx, campID)
		if err != nil {
			return err
		}
		if jsonOutputFlag {
			b, _ := json.MarshalIndent(execRes, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("Campaign %s started. Job: %s.\n", campID, execRes.JobID)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(campaignsCmd)
	campaignsCmd.AddCommand(campaignsListCmd, campaignsCreateCmd, campaignsGetCmd, campaignsUpdateCmd, campaignsDeleteCmd, campaignsExecuteCmd)

	campaignsListCmd.Flags().IntVar(&campPage, "page", 1, "Page number")
	campaignsListCmd.Flags().IntVar(&campLimit, "limit", 10, "Page size")
	campaignsListCmd.Flags().StringVar(&campStatus, "status", "", "Filter by status")
	campaignsListCmd.Flags().StringVar(&campAgentID, "agent-id", "", "Filter by agent id")

	campaignsCreateCmd.Flags().StringVar(&campName, "name", "", "Campaign name")
	campaignsCreateCmd.Flags().StringVar(&campDesc, "description", "", "Description")
	campaignsCreateCmd.Flags().StringVar(&campAgentID, "agent-id", "", "Agent ID")
	campaignsCreateCmd.Flags().StringSliceVar(&campScenarioIDs, "scenario-id", nil, "Scenario ID (repeatable)")

	for _, cmd := range []*cobra.Command{campaignsGetCmd, campaignsUpdateCmd, campaignsDeleteCmd, campaignsExecuteCmd} {
		cmd.Flags().StringVar(&campID, "id", "", "Campaign ID")
	}
	campaignsUpdateCmd.Flags().StringVar(&campName, "name", "", "New name")
	campaignsUpdateCmd.Flags().StringVar(&campDesc, "description", "", "New description")
}
