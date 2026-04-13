package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/audn-ai/audn-cli/pkg/internal/api"
	"github.com/audn-ai/audn-cli/pkg/internal/config"
	"github.com/audn-ai/audn-cli/pkg/internal/policy"
	"github.com/audn-ai/audn-cli/pkg/internal/validation"
	"github.com/spf13/cobra"
)

var (
	agentIDFlag        string
	campaignIDFlag     string
	failOnSeverityFlag string
	minGradeFlag       string
	waitFlag           bool
	timeoutFlag        int
)

var gateCmd = &cobra.Command{Use: "gate", Short: "Security gate commands"}

var gateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Run a VAST gate check for an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadEnv()
		if err != nil {
			return err
		}

		// Merge precedence: flag > env > default
		cfg.AgentID = firstNonEmpty(agentIDFlag, os.Getenv("AUDN_AGENT_ID"), cfg.AgentID)
		cfg.APISecret = firstNonEmpty(apiSecretFlag, os.Getenv("AUDN_API_SECRET"), cfg.APISecret)
		cfg.M2MClientID = firstNonEmpty(m2mClientIDFlag, os.Getenv("AUTH0_M2M_CLIENT_ID"), cfg.M2MClientID)
		cfg.M2MClientSecret = firstNonEmpty(m2mClientSecretFlag, os.Getenv("AUTH0_M2M_CLIENT_SECRET"), cfg.M2MClientSecret)
		cfg.FailOnSeverity = strings.ToLower(firstNonEmpty(failOnSeverityFlag, os.Getenv("AUDN_FAIL_ON_SEVERITY"), "critical"))
		cfg.MinGrade = strings.ToUpper(firstNonEmpty(minGradeFlag, os.Getenv("AUDN_MIN_GRADE"), cfg.MinGrade))
		cfg.Wait = parseBoolDefault(os.Getenv("AUDN_WAIT"), true)
		if cmd.Flags().Changed("wait") {
			cfg.Wait = waitFlag
		}
		cfg.TimeoutSeconds = parseIntDefault(os.Getenv("AUDN_TIMEOUT"), 600)
		if cmd.Flags().Changed("timeout") {
			cfg.TimeoutSeconds = timeoutFlag
		}
		cfg.APIBaseURL = firstNonEmpty(apiURLFlag, os.Getenv("AUDN_API_URL"), defaultIfEmpty(cfg.APIBaseURL, "https://audn.ai"))

		// Resolve auth: API key/secret for CI, else credentials from login
		authHdrs, err := api.ResolveAuthHeaders(cfg)
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIBaseURL, authHdrs)

		ctx, cancel := context.WithTimeout(cmd.Context(), time.Duration(cfg.TimeoutSeconds)*time.Second)
		defer cancel()

		if strings.TrimSpace(campaignIDFlag) != "" {
			// Validate campaign ID format
			if err := validation.ValidateCampaignID(campaignIDFlag); err != nil {
				return fmt.Errorf("validation error: %w", err)
			}

			// Platform campaign flow
			execRes, err := client.ExecuteCampaign(ctx, campaignIDFlag)
			if err != nil {
				return fmt.Errorf("failed to execute campaign: %w", err)
			}
			fmt.Printf("Campaign %s started. Job: %s. Waiting for completion...\n", campaignIDFlag, execRes.JobID)

			if !cfg.Wait {
				fmt.Println("✅ Started successfully (not waiting).")
				return nil
			}

			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("timeout waiting for campaign %s", campaignIDFlag)
				case <-ticker.C:
					camp, err := client.GetCampaign(ctx, campaignIDFlag)
					if err != nil {
						return fmt.Errorf("poll error: %w", err)
					}
					fmt.Printf("Status: %s (passed=%d failed=%d total=%d)\n", camp.Status, camp.Passed, camp.Failed, camp.Total)
					if camp.Status == "failed" {
						return fmt.Errorf("❌ Campaign failed")
					}
					if camp.Status == "completed" {
						// Fetch results to evaluate severity
						rr, err := client.ListResults(ctx, campaignIDFlag)
						if err != nil {
							return fmt.Errorf("failed to fetch results: %w", err)
						}
						// compute highest severity
						hs := highestSeverity(rr)
						res := policy.Evaluate(policy.Input{
							MinGrade:        cfg.MinGrade,
							FailOnSeverity:  cfg.FailOnSeverity,
							VastGrade:       "", // not available
							HighestSeverity: hs,
						})

						// JSON output for machine parsing
						if cmd.Flags().Changed("json") || os.Getenv("AUDN_OUTPUT_JSON") == "true" {
							result := map[string]any{
								"status":           "pass",
								"campaign_id":      campaignIDFlag,
								"highest_severity": hs,
								"total_results":    len(rr.Results),
								"passed_tests":     camp.Passed,
								"failed_tests":     camp.Failed,
								"total_tests":      camp.Total,
								"timestamp":        time.Now().UTC().Format(time.RFC3339),
							}
							if !res.Pass {
								result["status"] = "fail"
								result["reason"] = res.Reason
							}
							b, _ := json.Marshal(result)
							fmt.Println(string(b))
							if !res.Pass {
								return fmt.Errorf("gate failed: %s", res.Reason)
							}
							return nil
						}

						// Human-readable output
						if res.Pass {
							fmt.Printf("\x1b[32m✅ VAST Gate PASSED. Highest Severity: %s.\x1b[0m\n", hs)
							return nil
						}
						fmt.Printf("\x1b[31m❌ VAST Gate FAILED. Reason: %s (highest=%s)\x1b[0m\n", res.Reason, hs)
						return fmt.Errorf("gate failed: %s", res.Reason)
					}
				}
			}
		}

		// Legacy run API (if available)
		if cfg.AgentID == "" {
			return errors.New("missing required --agent-id or use --campaign-id")
		}

		// Validate agent ID format
		if err := validation.ValidateAgentID(cfg.AgentID); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		run, err := client.StartRun(ctx, cfg.AgentID)
		if err != nil {
			return fmt.Errorf("failed to start VAST run: %w", err)
		}
		fmt.Printf("VAST run initiated. Run ID: %s. Waiting for results...\n", run.RunID)
		if run.DashboardURL != "" {
			fmt.Printf("Dashboard: %s\n", run.DashboardURL)
		}
		if !cfg.Wait {
			fmt.Println("✅ Started successfully (not waiting).")
			return nil
		}
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timeout waiting for run %s", run.RunID)
			case <-ticker.C:
				status, err := client.GetRun(ctx, run.RunID)
				if err != nil {
					return fmt.Errorf("poll error: %w", err)
				}
				fmt.Printf("Status: %s\n", status.Status)
				if status.Status == "FAILED" {
					return fmt.Errorf("❌ VAST run failed: %s", status.Error)
				}
				if status.Status == "COMPLETED" {
					res := policy.Evaluate(policy.Input{MinGrade: cfg.MinGrade, FailOnSeverity: cfg.FailOnSeverity, VastGrade: status.Summary.VastGrade, HighestSeverity: status.Summary.HighestSeverityFinding})
					if status.ReportURL != "" {
						fmt.Printf("Full Report: %s\n", status.ReportURL)
					}
					if res.Pass {
						fmt.Printf("\x1b[32m✅ VAST Gate PASSED. Grade: %s. Highest Severity: %s.\x1b[0m\n", status.Summary.VastGrade, status.Summary.HighestSeverityFinding)
						return nil
					}
					fmt.Printf("\x1b[31m❌ VAST Gate FAILED. Reason: %s\x1b[0m\n", res.Reason)
					return fmt.Errorf("gate failed: %s", res.Reason)
				}
			}
		}
	},
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseBoolDefault(s string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "y":
		return true
	case "false", "0", "no", "n":
		return false
	default:
		return def
	}
}

func parseIntDefault(s string, def int) int {
	var v int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &v)
	if err != nil {
		return def
	}
	return v
}

func defaultIfEmpty(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func highestSeverity(rr *api.ResultsResponse) string {
	order := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	top := ""
	best := -1
	for _, r := range rr.Results {
		sev := strings.ToLower(strings.TrimSpace(r.Attack.Severity))
		rank := order[sev]
		if rank > best {
			best = rank
			top = sev
		}
	}
	if top == "" {
		top = "low"
	}
	return top
}

func init() {
	rootCmd.AddCommand(gateCmd)
	gateCmd.AddCommand(gateCheckCmd)

	gateCheckCmd.Flags().StringVar(&agentIDFlag, "agent-id", "", "Agent ID to test (or AUDN_AGENT_ID)")
	gateCheckCmd.Flags().StringVar(&campaignIDFlag, "campaign-id", "", "Campaign ID to execute (preferred)")
	gateCheckCmd.Flags().StringVar(&failOnSeverityFlag, "fail-on-severity", "", "Fail if finding at or above this severity")
	gateCheckCmd.Flags().StringVar(&minGradeFlag, "min-grade", "", "Fail if VAST Grade is below this (A/B/C/D/F)")
	gateCheckCmd.Flags().BoolVar(&waitFlag, "wait", true, "Wait for completion (or AUDN_WAIT)")
	gateCheckCmd.Flags().IntVar(&timeoutFlag, "timeout", 600, "Timeout seconds (or AUDN_TIMEOUT)")
	// Note: --api-secret, --api-url are now global persistent flags
}
