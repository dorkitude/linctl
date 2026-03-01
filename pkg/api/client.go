package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://api.linear.app/graphql"
)

type Client struct {
	httpClient *http.Client
	authHeader string
	baseURL    string
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
	// Extensions may contain rate-limit/cost details and other server metadata.
	Extensions json.RawMessage `json:"extensions,omitempty"`
}

type GraphQLError struct {
	Message   string                 `json:"message"`
	Locations []GraphQLErrorLocation `json:"locations,omitempty"`
	Path      []interface{}          `json:"path,omitempty"`
}

type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// NewClient creates a new Linear API client
func NewClient(authHeader string) *Client {
	return NewClientWithURL(BaseURL, authHeader)
}

// NewClientWithURL creates a new Linear API client with custom URL
func NewClientWithURL(baseURL, authHeader string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authHeader: authHeader,
		baseURL:    baseURL,
	}
}

// Execute performs a GraphQL request
func (c *Client) Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	gqlResp, err := c.ExecuteRaw(ctx, query, variables)
	if err != nil {
		return err
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	if result != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	return nil
}

// ExecuteRaw performs a GraphQL request and returns raw GraphQL response fields.
// Unlike Execute, this does not treat GraphQL errors as transport errors.
func (c *Client) ExecuteRaw(ctx context.Context, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("User-Agent", "linctl/0.1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &gqlResp, nil
}

// Rate limiting helper
func (c *Client) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	// This would query Linear's rate limiting info
	// For now, we'll return a placeholder
	return &RateLimit{
		Limit:     5000,
		Remaining: 4999,
		Reset:     time.Now().Add(time.Hour),
	}, nil
}

type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}
