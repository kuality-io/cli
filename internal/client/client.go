package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kuality-io/cli/internal/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

var version = "dev"

func SetVersion(v string) {
	version = v
}

func New(cfg *config.Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key configured. Run 'kuality auth login' or set KUALITY_API_KEY")
	}

	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL %q: %w", cfg.BaseURL, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, fmt.Errorf("base URL must use https:// or http:// scheme")
	}

	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "kuality-cli/" + version,
	}, nil
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	endpoint := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("cannot encode request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) decodeResponse(resp *http.Response, target any) error {
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 10*1024*1024)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("authentication failed (401). Check your API key with 'ku auth status'")
	}
	if resp.StatusCode == 403 {
		return fmt.Errorf("forbidden (403). Your plan may not support this operation")
	}
	if resp.StatusCode == 429 {
		return fmt.Errorf("rate limited (429). Wait a moment and try again")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(data))
	}

	if target != nil {
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("cannot decode response: %w", err)
		}
	}

	return nil
}

// Scan types

type CreateScanRequest struct {
	Target   string `json:"target"`
	ScanType string `json:"scan_type"`
}

type ScanResponse struct {
	ScanID   string `json:"scan_id"`
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
	Target   string `json:"target"`
	PollURL  string `json:"poll_url"`
}

type ScanStatus struct {
	ScanID   string `json:"scan_id"`
	ReportID string `json:"report_id"`
	Status   string `json:"status"`
	State    string `json:"state"`
	Target   string `json:"target"`
}

type Report struct {
	ID          string          `json:"id"`
	ReportID    string          `json:"report_id"`
	Target      string          `json:"target"`
	TypeOfScan  string          `json:"type_of_scan"`
	State       string          `json:"state"`
	Score       json.Number     `json:"score"`
	High        int             `json:"high"`
	Medium      int             `json:"medium"`
	Low         int             `json:"low"`
	Info        int             `json:"info"`
	Total       int             `json:"total"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	Error       string          `json:"error"`
	Findings    json.RawMessage `json:"findings"`
	RawFindings json.RawMessage `json:"report_ex"`
}

type ReportListItem struct {
	ID         string      `json:"id"`
	Target     string      `json:"target"`
	TypeOfScan string      `json:"type_of_scan"`
	State      string      `json:"state"`
	Score      json.Number `json:"score"`
	High       int         `json:"high"`
	Medium     int         `json:"medium"`
	Low        int         `json:"low"`
	Info       int         `json:"info"`
	CreatedAt  string      `json:"created_at"`
}

type TargetItem struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

type ScoreItem struct {
	Target string      `json:"target"`
	Score  json.Number `json:"score"`
	Type   string      `json:"type"`
}

type GateResult struct {
	Pass    bool   `json:"pass"`
	Score   json.Number `json:"score"`
	Details string `json:"details"`
}

func (c *Client) CreateScan(target, scanType string) (*ScanResponse, error) {
	resp, err := c.do("POST", "/api/v1/scans", &CreateScanRequest{
		Target:   target,
		ScanType: scanType,
	})
	if err != nil {
		return nil, err
	}

	var result ScanResponse
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetScanStatus(scanID string) (*ScanStatus, error) {
	resp, err := c.do("GET", "/api/v1/scans/"+url.PathEscape(scanID), nil)
	if err != nil {
		return nil, err
	}

	var result ScanStatus
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetReport(reportID string) (*Report, error) {
	resp, err := c.do("GET", "/api/v1/reports/"+url.PathEscape(reportID), nil)
	if err != nil {
		return nil, err
	}

	var result Report
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetReportJUnit(reportID string) ([]byte, error) {
	resp, err := c.do("GET", "/api/v1/reports/"+url.PathEscape(reportID)+"/junit", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
}

func (c *Client) ListReports(scanType, target string) ([]ReportListItem, error) {
	path := "/api/v1/reports"
	params := url.Values{}
	if scanType != "" {
		params.Set("scan_type", scanType)
	}
	if target != "" {
		params.Set("target", target)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result []ReportListItem
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListTargets() ([]TargetItem, error) {
	resp, err := c.do("GET", "/api/v1/targets", nil)
	if err != nil {
		return nil, err
	}

	var result []TargetItem
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) ListScores() ([]ScoreItem, error) {
	resp, err := c.do("GET", "/api/v1/scores", nil)
	if err != nil {
		return nil, err
	}

	var result []ScoreItem
	if err := c.decodeResponse(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}
