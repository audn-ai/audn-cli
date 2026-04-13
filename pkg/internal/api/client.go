package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"encoding/base64"

	"github.com/audn-ai/audn-cli/pkg/internal/auth"
	"github.com/audn-ai/audn-cli/pkg/internal/config"
)

type Client struct {
	baseURL string
	headers map[string]string
	http    *http.Client
}

func NewClient(baseURL string, headers map[string]string) *Client {
	return &Client{baseURL: strings.TrimSuffix(baseURL, "/"), headers: headers, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) do(ctx context.Context, method, p string, body any, out any) error {
	var buf *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		buf = bytes.NewReader(b)
	} else {
		buf = bytes.NewReader(nil)
	}
	url := c.baseURL + path.Clean("/"+p)
	req, _ := http.NewRequestWithContext(ctx, method, url, buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Error == "" {
			e.Error = resp.Status
		}
		return fmt.Errorf("api error: %s", e.Error)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

type startRunReq struct {
	AgentID string `json:"agentId"`
	Source  string `json:"source"`
}

type StartRunResponse struct {
	RunID        string `json:"runId"`
	Status       string `json:"status"`
	DashboardURL string `json:"dashboardUrl"`
}

func (c *Client) StartRun(ctx context.Context, agentID string) (*StartRunResponse, error) {
	var out StartRunResponse
	err := c.do(ctx, http.MethodPost, "/api/v1/runs", startRunReq{AgentID: agentID, Source: "cli"}, &out)
	if err != nil {
		return nil, err
	}
	if out.RunID == "" {
		return nil, errors.New("invalid response: missing runId")
	}
	return &out, nil
}

// Audn.ai platform-compatible endpoints
type ExecuteCampaignResponse struct {
	Success bool   `json:"success"`
	JobID   string `json:"jobId"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (c *Client) ExecuteCampaign(ctx context.Context, campaignID string) (*ExecuteCampaignResponse, error) {
	var out ExecuteCampaignResponse
	path := fmt.Sprintf("/api/campaigns/%s/execute", campaignID)
	err := c.do(ctx, http.MethodPost, path, map[string]any{}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type Campaign struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Total  int    `json:"total_tests"`
	Passed int    `json:"passed_tests"`
	Failed int    `json:"failed_tests"`
}

func (c *Client) GetCampaign(ctx context.Context, campaignID string) (*Campaign, error) {
	var out Campaign
	path := fmt.Sprintf("/api/campaigns/%s", campaignID)
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Results API response (subset)
type ResultsResponse struct {
	Results []struct {
		Status string  `json:"status"`
		Risk   float64 `json:"risk_score"`
		Attack struct {
			Severity string `json:"severity"`
		} `json:"attack_scenarios"`
	} `json:"results"`
}

func (c *Client) ListResults(ctx context.Context, campaignID string) (*ResultsResponse, error) {
	var out ResultsResponse
	p := fmt.Sprintf("/api/results?campaignId=%s&limit=1000", campaignID)
	if err := c.do(ctx, http.MethodGet, p, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Agents API
type AgentsListResponse struct {
	Agents []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Platform string `json:"platform"`
		Status   string `json:"status"`
		Phone    string `json:"phone_number"`
	} `json:"agents"`
	Pagination struct {
		Total int `json:"total"`
	} `json:"pagination"`
}

func (c *Client) AgentsList(ctx context.Context, page, limit int, status string) (*AgentsListResponse, error) {
	qs := fmt.Sprintf("?page=%d&limit=%d", page, limit)
	if strings.TrimSpace(status) != "" {
		qs += "&status=" + urlQueryEscape(status)
	}
	var out AgentsListResponse
	if err := c.do(ctx, http.MethodGet, "/api/agents"+qs, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Status      string `json:"status"`
	Phone       string `json:"phone_number"`
	Description string `json:"description"`
}

type AgentCreateReq struct {
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Description string `json:"description,omitempty"`
}

type AgentUpdateReq struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Platform    string `json:"platform,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Description string `json:"description,omitempty"`
}

func (c *Client) AgentCreate(ctx context.Context, reqBody AgentCreateReq) (*Agent, error) {
	var out Agent
	if err := c.do(ctx, http.MethodPost, "/api/agents", reqBody, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AgentGet(ctx context.Context, id string) (*Agent, error) {
	// agents/[id] GET
	var out Agent
	if err := c.do(ctx, http.MethodGet, "/api/agents/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AgentUpdate(ctx context.Context, reqBody AgentUpdateReq) (*Agent, error) {
	var out Agent
	if err := c.do(ctx, http.MethodPut, "/api/agents", reqBody, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AgentDelete(ctx context.Context, id string) error {
	p := "/api/agents?id=" + urlQueryEscape(id)
	return c.do(ctx, http.MethodDelete, p, nil, nil)
}

// Campaigns API
type CampaignsListResponse struct {
	Campaigns  []Campaign `json:"campaigns"`
	Pagination struct {
		Total int `json:"total"`
	} `json:"pagination"`
}

func (c *Client) CampaignsList(ctx context.Context, page, limit int, status, agentID string) (*CampaignsListResponse, error) {
	qs := fmt.Sprintf("?page=%d&limit=%d", page, limit)
	if strings.TrimSpace(status) != "" {
		qs += "&status=" + urlQueryEscape(status)
	}
	if strings.TrimSpace(agentID) != "" {
		qs += "&agentId=" + urlQueryEscape(agentID)
	}
	var out CampaignsListResponse
	if err := c.do(ctx, http.MethodGet, "/api/campaigns"+qs, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type CampaignCreateReq struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	AgentID     string   `json:"agent_id"`
	ScenarioIDs []string `json:"scenario_ids,omitempty"`
}

func (c *Client) CampaignCreate(ctx context.Context, body CampaignCreateReq) (*Campaign, error) {
	var out Campaign
	if err := c.do(ctx, http.MethodPost, "/api/campaigns", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CampaignUpdate(ctx context.Context, id string, patch map[string]any) (*Campaign, error) {
	patch["id"] = id
	var out Campaign
	if err := c.do(ctx, http.MethodPut, "/api/campaigns/"+id, patch, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CampaignDelete(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/campaigns/"+id, nil, nil)
}

// small helpers
func urlQueryEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", "+"), "\n", "")
}

type RunSummary struct {
	VastGrade              string `json:"vastGrade"`
	HighestSeverityFinding string `json:"highestSeverityFinding"`
	TotalFindings          int    `json:"totalFindings"`
}

type GetRunResponse struct {
	RunID     string     `json:"runId"`
	Status    string     `json:"status"`
	Summary   RunSummary `json:"summary"`
	ReportURL string     `json:"reportUrl"`
	Error     string     `json:"error"`
}

func (c *Client) GetRun(ctx context.Context, runID string) (*GetRunResponse, error) {
	var out GetRunResponse
	err := c.do(ctx, http.MethodGet, "/api/v1/runs/"+runID, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveAuthHeaders chooses between CI auth (sk_ API key), M2M, bearer token, and user auth
func ResolveAuthHeaders(env config.Env) (map[string]string, error) {
	// CI path - sk_ API key as Bearer token (highest priority)
	if strings.TrimSpace(env.APISecret) != "" && strings.HasPrefix(env.APISecret, "sk_") {
		return map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", env.APISecret),
		}, nil
	}

	// M2M path - Auth0 Machine-to-Machine (for CI/CD)
	if strings.TrimSpace(env.M2MClientID) != "" && strings.TrimSpace(env.M2MClientSecret) != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		tokenResp, err := auth.GetM2MToken(ctx, env)
		if err != nil {
			return nil, fmt.Errorf("failed to get M2M token: %w", err)
		}
		return map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", tokenResp.AccessToken),
		}, nil
	}

	// Bearer token path (pre-obtained token)
	if strings.TrimSpace(env.BearerToken) != "" {
		return map[string]string{"Authorization": "Bearer " + env.BearerToken}, nil
	}
	// Developer path via stored credentials
	cred, err := config.ReadCredentials()
	if err != nil {
		return nil, err
	}
	if cred == nil {
		return nil, errors.New("you are not logged in and no API secret provided. Run 'audn-cli login' or pass --api-secret")
	}
	// Refresh if needed
	if _, err := auth.EnsureValidCredentials(context.Background(), env, cred); err != nil {
		// proceed anyway with old token; server may still accept if valid by ID token exp
		_ = err
	}
	// Construct appSession cookie expected by Next.js getSession()
	sess := map[string]any{
		"user": map[string]any{
			// minimal claims; decode from ID token if needed in the future
		},
		"accessToken":  cred.AccessToken,
		"idToken":      cred.IDToken,
		"refreshToken": cred.RefreshToken,
		"tokenType":    "Bearer",
		"expiresAt":    cred.ExpiresAt,
	}
	// Try to decode email/sub from ID token for better user mapping
	if u, e := extractUserFromIDToken(cred.IDToken); e == nil {
		sess["user"] = u
	} else if strings.TrimSpace(cred.UserEmail) != "" {
		sess["user"] = map[string]any{"sub": "auth0|cli-user", "email": cred.UserEmail}
	}
	b, _ := json.Marshal(sess)
	cookie := base64Std(b)
	return map[string]string{
		"Cookie": "appSession=" + cookie + "; auth0.is_authenticated=true",
	}, nil
}

// Helpers for cookie base64 encoding
func base64Std(b []byte) string {
	return toBase64(b)
}

// small indirection for testability
var toBase64 = func(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

// extract minimal user fields from ID token payload
func extractUserFromIDToken(idt string) (map[string]any, error) {
	parts := strings.Split(idt, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid id token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var c map[string]any
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, err
	}
	user := map[string]any{
		"sub":            c["sub"],
		"email":          c["email"],
		"name":           c["name"],
		"picture":        c["picture"],
		"updated_at":     c["updated_at"],
		"email_verified": c["email_verified"],
	}
	return user, nil
}
