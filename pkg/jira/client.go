// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This file implements the Jira REST API client.
//
// CRITICAL: Never log API tokens or sensitive ticket content.
package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// API path constants
const (
	apiPathIssueCreate     = "/rest/api/3/issue"
	apiPathIssue           = "/rest/api/3/issue/%s"
	apiPathIssueTransition = "/rest/api/3/issue/%s/transitions"
	apiPathIssueComment    = "/rest/api/3/issue/%s/comment"
	apiPathSearch          = "/rest/api/3/search"
	apiPathServiceDeskInfo = "/rest/servicedeskapi/info"
)

// Client implements the Jira REST API client
type Client struct {
	// baseURL is the Jira instance base URL
	baseURL string

	// httpClient is the HTTP client
	httpClient *http.Client

	// auth holds authentication details
	auth AuthConfig

	// userAgent is the User-Agent header
	userAgent string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Type is the authentication type
	Type AuthType

	// Username is the Jira username (for basic auth)
	Username string

	// APIToken is the API token (for basic auth)
	// CRITICAL: Never log this value
	APIToken string

	// BearerToken is the bearer token (for OAuth)
	// CRITICAL: Never log this value
	BearerToken string
}

// AuthType represents authentication types
type AuthType string

const (
	// AuthTypeBasic uses username + API token
	AuthTypeBasic AuthType = "basic"

	// AuthTypeBearer uses OAuth bearer token
	AuthTypeBearer AuthType = "bearer"
)

// ClientConfig holds client configuration
type ClientConfig struct {
	// BaseURL is the Jira instance URL (e.g., "https://company.atlassian.net")
	BaseURL string

	// Auth is the authentication configuration
	Auth AuthConfig

	// Timeout is the HTTP timeout
	Timeout time.Duration

	// UserAgent is the User-Agent header
	UserAgent string
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Timeout:   30 * time.Second,
		UserAgent: "VirtEngine-Jira-Client/1.0",
	}
}

// NewClient creates a new Jira API client
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("jira: base URL is required")
	}

	// Normalize base URL
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")

	// Validate auth
	switch cfg.Auth.Type {
	case AuthTypeBasic:
		if cfg.Auth.Username == "" || cfg.Auth.APIToken == "" {
			return nil, fmt.Errorf("jira: username and API token required for basic auth")
		}
	case AuthTypeBearer:
		if cfg.Auth.BearerToken == "" {
			return nil, fmt.Errorf("jira: bearer token required for OAuth auth")
		}
	default:
		return nil, fmt.Errorf("jira: invalid auth type: %s", cfg.Auth.Type)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = "VirtEngine-Jira-Client/1.0"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		auth:      cfg.Auth,
		userAgent: userAgent,
	}, nil
}

// IClient defines the interface for the Jira client
type IClient interface {
	// Issue operations
	CreateIssue(ctx context.Context, req *CreateIssueRequest) (*CreateIssueResponse, error)
	GetIssue(ctx context.Context, issueKeyOrID string) (*Issue, error)
	UpdateIssue(ctx context.Context, issueKeyOrID string, req *UpdateIssueRequest) error
	DeleteIssue(ctx context.Context, issueKeyOrID string) error
	SearchIssues(ctx context.Context, jql string, startAt, maxResults int) (*SearchResult, error)

	// Transitions
	GetTransitions(ctx context.Context, issueKeyOrID string) (*TransitionsResponse, error)
	TransitionIssue(ctx context.Context, issueKeyOrID string, req *TransitionRequest) error

	// Comments
	AddComment(ctx context.Context, issueKeyOrID string, comment *AddCommentRequest) (*Comment, error)
	GetComments(ctx context.Context, issueKeyOrID string, startAt, maxResults int) (*CommentResponse, error)

	// Service Desk specific
	GetServiceDeskInfo(ctx context.Context) (map[string]interface{}, error)
}

// Ensure Client implements IClient
var _ IClient = (*Client)(nil)

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("jira: failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("jira: failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Set auth header
	// CRITICAL: Auth header is set but never logged
	switch c.auth.Type {
	case AuthTypeBasic:
		req.SetBasicAuth(c.auth.Username, c.auth.APIToken)
	case AuthTypeBearer:
		req.Header.Set("Authorization", "Bearer "+c.auth.BearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("jira: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("jira: failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// CreateIssue creates a new Jira issue
func (c *Client) CreateIssue(ctx context.Context, req *CreateIssueRequest) (*CreateIssueResponse, error) {
	respBody, statusCode, err := c.doRequest(ctx, http.MethodPost, apiPathIssueCreate, req)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusCreated {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
			return nil, &errResp
		}
		return nil, fmt.Errorf("jira: create issue failed with status %d", statusCode)
	}

	var result CreateIssueResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("jira: failed to parse create issue response: %w", err)
	}

	return &result, nil
}

// GetIssue retrieves a Jira issue by key or ID
func (c *Client) GetIssue(ctx context.Context, issueKeyOrID string) (*Issue, error) {
	path := fmt.Sprintf(apiPathIssue, url.PathEscape(issueKeyOrID))
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, fmt.Errorf("jira: issue not found: %s", issueKeyOrID)
	}

	if statusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
			return nil, &errResp
		}
		return nil, fmt.Errorf("jira: get issue failed with status %d", statusCode)
	}

	var issue Issue
	if err := json.Unmarshal(respBody, &issue); err != nil {
		return nil, fmt.Errorf("jira: failed to parse issue response: %w", err)
	}

	return &issue, nil
}

// UpdateIssue updates a Jira issue
func (c *Client) UpdateIssue(ctx context.Context, issueKeyOrID string, req *UpdateIssueRequest) error {
	path := fmt.Sprintf(apiPathIssue, url.PathEscape(issueKeyOrID))
	respBody, statusCode, err := c.doRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNoContent || statusCode == http.StatusOK {
		return nil
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
		return &errResp
	}
	return fmt.Errorf("jira: update issue failed with status %d", statusCode)
}

// DeleteIssue deletes a Jira issue
func (c *Client) DeleteIssue(ctx context.Context, issueKeyOrID string) error {
	path := fmt.Sprintf(apiPathIssue, url.PathEscape(issueKeyOrID))
	respBody, statusCode, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNoContent || statusCode == http.StatusOK {
		return nil
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
		return &errResp
	}
	return fmt.Errorf("jira: delete issue failed with status %d", statusCode)
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, startAt, maxResults int) (*SearchResult, error) {
	searchReq := map[string]interface{}{
		"jql":        jql,
		"startAt":    startAt,
		"maxResults": maxResults,
	}

	respBody, statusCode, err := c.doRequest(ctx, http.MethodPost, apiPathSearch, searchReq)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
			return nil, &errResp
		}
		return nil, fmt.Errorf("jira: search failed with status %d", statusCode)
	}

	var result SearchResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("jira: failed to parse search response: %w", err)
	}

	return &result, nil
}

// GetTransitions retrieves available transitions for an issue
func (c *Client) GetTransitions(ctx context.Context, issueKeyOrID string) (*TransitionsResponse, error) {
	path := fmt.Sprintf(apiPathIssueTransition, url.PathEscape(issueKeyOrID))
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
			return nil, &errResp
		}
		return nil, fmt.Errorf("jira: get transitions failed with status %d", statusCode)
	}

	var result TransitionsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("jira: failed to parse transitions response: %w", err)
	}

	return &result, nil
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(ctx context.Context, issueKeyOrID string, req *TransitionRequest) error {
	path := fmt.Sprintf(apiPathIssueTransition, url.PathEscape(issueKeyOrID))
	respBody, statusCode, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return err
	}

	if statusCode == http.StatusNoContent || statusCode == http.StatusOK {
		return nil
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
		return &errResp
	}
	return fmt.Errorf("jira: transition failed with status %d", statusCode)
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(ctx context.Context, issueKeyOrID string, comment *AddCommentRequest) (*Comment, error) {
	path := fmt.Sprintf(apiPathIssueComment, url.PathEscape(issueKeyOrID))
	respBody, statusCode, err := c.doRequest(ctx, http.MethodPost, path, comment)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusCreated && statusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
			return nil, &errResp
		}
		return nil, fmt.Errorf("jira: add comment failed with status %d", statusCode)
	}

	var result Comment
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("jira: failed to parse comment response: %w", err)
	}

	return &result, nil
}

// GetComments retrieves comments for an issue
func (c *Client) GetComments(ctx context.Context, issueKeyOrID string, startAt, maxResults int) (*CommentResponse, error) {
	path := fmt.Sprintf(apiPathIssueComment+"?startAt=%d&maxResults=%d",
		url.PathEscape(issueKeyOrID), startAt, maxResults)
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && (len(errResp.ErrorMessages) > 0 || len(errResp.Errors) > 0) {
			return nil, &errResp
		}
		return nil, fmt.Errorf("jira: get comments failed with status %d", statusCode)
	}

	var result CommentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("jira: failed to parse comments response: %w", err)
	}

	return &result, nil
}

// GetServiceDeskInfo retrieves service desk information
func (c *Client) GetServiceDeskInfo(ctx context.Context) (map[string]interface{}, error) {
	respBody, statusCode, err := c.doRequest(ctx, http.MethodGet, apiPathServiceDeskInfo, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		// Service Desk API might not be available
		return nil, fmt.Errorf("jira: service desk info unavailable (status %d)", statusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("jira: failed to parse service desk info: %w", err)
	}

	return result, nil
}

// FindIssueByVirtEngineTicketID finds a Jira issue by VirtEngine ticket ID
func (c *Client) FindIssueByVirtEngineTicketID(ctx context.Context, ticketID, customFieldID string) (*Issue, error) {
	jql := fmt.Sprintf(`"%s" ~ "%s"`, customFieldID, ticketID)
	result, err := c.SearchIssues(ctx, jql, 0, 1)
	if err != nil {
		return nil, err
	}

	if len(result.Issues) == 0 {
		return nil, nil
	}

	return &result.Issues[0], nil
}

// FindIssuesBySubmitterAddress finds Jira issues by VirtEngine submitter address
func (c *Client) FindIssuesBySubmitterAddress(ctx context.Context, address, customFieldID string, startAt, maxResults int) (*SearchResult, error) {
	jql := fmt.Sprintf(`"%s" ~ "%s"`, customFieldID, address)
	return c.SearchIssues(ctx, jql, startAt, maxResults)
}

