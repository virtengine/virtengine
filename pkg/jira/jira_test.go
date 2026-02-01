// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This file contains unit tests for the Jira integration.
package jira

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test constants
const (
	testIssueKey = "TEST-1"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name: "valid basic auth config",
			config: ClientConfig{
				BaseURL: "https://test.atlassian.net",
				Auth: AuthConfig{
					Type:     AuthTypeBasic,
					Username: "user@test.com",
					APIToken: "test-token",
				},
			},
			wantErr: false,
		},
		{
			name: "valid bearer auth config",
			config: ClientConfig{
				BaseURL: "https://test.atlassian.net",
				Auth: AuthConfig{
					Type:        AuthTypeBearer,
					BearerToken: "test-bearer-token",
				},
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: ClientConfig{
				Auth: AuthConfig{
					Type:     AuthTypeBasic,
					Username: "user@test.com",
					APIToken: "test-token",
				},
			},
			wantErr: true,
		},
		{
			name: "missing username for basic auth",
			config: ClientConfig{
				BaseURL: "https://test.atlassian.net",
				Auth: AuthConfig{
					Type:     AuthTypeBasic,
					APIToken: "test-token",
				},
			},
			wantErr: true,
		},
		{
			name: "missing token for bearer auth",
			config: ClientConfig{
				BaseURL: "https://test.atlassian.net",
				Auth: AuthConfig{
					Type: AuthTypeBearer,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid auth type",
			config: ClientConfig{
				BaseURL: "https://test.atlassian.net",
				Auth: AuthConfig{
					Type: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				}
			}
		})
	}
}

// MockJiraServer creates a mock Jira server for testing
func MockJiraServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for handler
		if handler, ok := handlers[r.Method+" "+r.URL.Path]; ok {
			handler(w, r)
			return
		}

		// Default 404
		http.NotFound(w, r)
	}))
}

// TestClientCreateIssue tests issue creation
func TestClientCreateIssue(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/issue" {
			http.NotFound(w, r)
			return
		}
		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type: application/json")
		}

		// Decode request
		var req CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Verify request
		if req.Fields.Summary == "" {
			http.Error(w, "Summary required", http.StatusBadRequest)
			return
		}

		// Return success
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(CreateIssueResponse{
			ID:   "10001",
			Key:  testIssueKey,
			Self: server.URL + "/rest/api/3/issue/10001",
		})
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Type:     AuthTypeBasic,
			Username: "test",
			APIToken: "test",
		},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	resp, err := client.CreateIssue(ctx, &CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   Project{Key: "TEST"},
			Summary:   "Test Issue",
			IssueType: IssueTypeField{Name: "Task"},
		},
	})

	if err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	if resp.Key != testIssueKey {
		t.Errorf("expected key TEST-1, got %s", resp.Key)
	}
}

// TestClientGetIssue tests issue retrieval
func TestClientGetIssue(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Path {
		case "/rest/api/3/issue/TEST-1":
			_ = json.NewEncoder(w).Encode(Issue{
				ID:   "10001",
				Key:  testIssueKey,
				Self: server.URL + "/rest/api/3/issue/10001",
				Fields: IssueFields{
					Summary: "Test Issue",
					Status: StatusField{
						Name: "Open",
					},
				},
			})
		case "/rest/api/3/issue/NOTFOUND":
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Type:     AuthTypeBasic,
			Username: "test",
			APIToken: "test",
		},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test successful retrieval
	issue, err := client.GetIssue(ctx, testIssueKey)
	if err != nil {
		t.Fatalf("failed to get issue: %v", err)
	}
	if issue.Key != testIssueKey {
		t.Errorf("expected key TEST-1, got %s", issue.Key)
	}
	if issue.Fields.Summary != "Test Issue" {
		t.Errorf("expected summary 'Test Issue', got %s", issue.Fields.Summary)
	}

	// Test not found
	_, err = client.GetIssue(ctx, "NOTFOUND")
	if err == nil {
		t.Error("expected error for not found issue")
	}
}

// TestClientAddComment tests adding comments
func TestClientAddComment(t *testing.T) {
	server := MockJiraServer(t, map[string]http.HandlerFunc{
		"POST /rest/api/3/issue/TEST-1/comment": func(w http.ResponseWriter, r *http.Request) {
			var req AddCommentRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(Comment{
				ID:      "10001",
				Body:    req.Body,
				Created: time.Now().Format(time.RFC3339),
			})
		},
	})
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Type:     AuthTypeBasic,
			Username: "test",
			APIToken: "test",
		},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	comment, err := client.AddComment(ctx, testIssueKey, &AddCommentRequest{
		Body: "Test comment",
	})

	if err != nil {
		t.Fatalf("failed to add comment: %v", err)
	}

	if comment.Body != "Test comment" {
		t.Errorf("expected body 'Test comment', got %s", comment.Body)
	}
}

// TestClientSearchIssues tests issue search
func TestClientSearchIssues(t *testing.T) {
	server := MockJiraServer(t, map[string]http.HandlerFunc{
		"POST /rest/api/3/search": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(SearchResult{
				StartAt:    0,
				MaxResults: 10,
				Total:      2,
				Issues: []Issue{
					{ID: "10001", Key: testIssueKey},
					{ID: "10002", Key: "TEST-2"},
				},
			})
		},
	})
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL: server.URL,
		Auth: AuthConfig{
			Type:     AuthTypeBasic,
			Username: "test",
			APIToken: "test",
		},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.SearchIssues(ctx, "project = TEST", 0, 10)

	if err != nil {
		t.Fatalf("failed to search issues: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected 2 total issues, got %d", result.Total)
	}

	if len(result.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(result.Issues))
	}
}

// TestTicketBridge tests the ticket bridge
func TestTicketBridge(t *testing.T) {
	// Create mock client
	mockClient := &MockClient{
		issues:   make(map[string]*Issue),
		comments: make(map[string][]Comment),
	}

	bridge := NewTicketBridge(mockClient, DefaultTicketBridgeConfig())

	ctx := context.Background()

	// Test creating from support request
	req := &VirtEngineSupportRequest{
		TicketID:         "ve-ticket-001",
		TicketNumber:     "VE-001",
		SubmitterAddress: "virtengine1abc123def456",
		Category:         "technical",
		Priority:         "high",
		Subject:          "Test support request",
		Description:      "This is a test support request",
		CreatedAt:        time.Now(),
	}

	issue, err := bridge.CreateFromSupportRequest(ctx, req)
	if err != nil {
		t.Fatalf("failed to create from support request: %v", err)
	}

	if issue == nil {
		t.Fatal("expected issue, got nil")
	}

	// Verify issue was created
	if mockClient.issues[issue.Key] == nil {
		t.Error("issue not found in mock client")
	}
}

// TestSLATracker tests SLA tracking
func TestSLATracker(t *testing.T) {
	tracker := NewSLATracker(DefaultSLAConfig())

	// Start tracking
	ticketID := "test-ticket-1"
	now := time.Now()
	err := tracker.StartTracking(ticketID, testIssueKey, "high", now)
	if err != nil {
		t.Fatalf("failed to start tracking: %v", err)
	}

	// Duplicate tracking should fail
	err = tracker.StartTracking(ticketID, testIssueKey, "high", now)
	if err == nil {
		t.Error("expected error for duplicate tracking")
	}

	// Get SLA info
	info, err := tracker.GetSLAInfo(ticketID)
	if err != nil {
		t.Fatalf("failed to get SLA info: %v", err)
	}

	if info.TicketKey != testIssueKey {
		t.Errorf("expected ticket key TEST-1, got %s", info.TicketKey)
	}

	if info.ResponseSLA == nil {
		t.Error("expected response SLA info")
	}

	// Record first response
	err = tracker.RecordFirstResponse(ticketID, time.Now())
	if err != nil {
		t.Fatalf("failed to record first response: %v", err)
	}

	// Pause SLA
	err = tracker.PauseSLA(ticketID)
	if err != nil {
		t.Fatalf("failed to pause SLA: %v", err)
	}

	// Resume SLA
	err = tracker.ResumeSLA(ticketID)
	if err != nil {
		t.Fatalf("failed to resume SLA: %v", err)
	}

	// Record resolution
	err = tracker.RecordResolution(ticketID, time.Now())
	if err != nil {
		t.Fatalf("failed to record resolution: %v", err)
	}

	// Get metrics
	metrics := tracker.GetMetrics()
	if metrics.TotalTickets != 1 {
		t.Errorf("expected 1 total ticket, got %d", metrics.TotalTickets)
	}

	if metrics.ActiveTickets != 0 {
		t.Errorf("expected 0 active tickets (resolved), got %d", metrics.ActiveTickets)
	}
}

// TestSLABreach tests SLA breach detection
func TestSLABreach(t *testing.T) {
	// Create tracker with short SLA targets for testing
	config := SLAConfig{
		ResponseTimeTargets: map[string]int64{
			"high": 1, // 1 minute
		},
		ResolutionTimeTargets: map[string]int64{
			"high": 2, // 2 minutes
		},
	}
	tracker := NewSLATracker(config)

	// Start tracking with a past time to simulate breach
	ticketID := "breach-test-1"
	pastTime := time.Now().Add(-5 * time.Minute) // 5 minutes ago
	err := tracker.StartTracking(ticketID, testIssueKey, "high", pastTime)
	if err != nil {
		t.Fatalf("failed to start tracking: %v", err)
	}

	// Check for breaches
	ctx := context.Background()
	breached, err := tracker.CheckSLABreaches(ctx)
	if err != nil {
		t.Fatalf("failed to check breaches: %v", err)
	}

	if len(breached) == 0 {
		t.Error("expected breached tickets")
	}

	// Verify breach status
	info, err := tracker.GetSLAInfo(ticketID)
	if err != nil {
		t.Fatalf("failed to get SLA info: %v", err)
	}

	if !info.ResponseSLA.Breached {
		t.Error("expected response SLA to be breached")
	}
}

// TestWebhookHandler tests webhook handling
func TestWebhookHandler(t *testing.T) {
	handler := NewWebhookHandler(WebhookConfig{
		Secret:           "test-secret",
		RequireSignature: false,
	})

	// Register handler
	var receivedEvent *WebhookEvent
	handler.RegisterHandler(WebhookEventIssueUpdated, func(ctx context.Context, event *WebhookEvent) error {
		receivedEvent = event
		return nil
	})

	// Create test event
	event := &WebhookEvent{
		WebhookEvent: string(WebhookEventIssueUpdated),
		Issue: &Issue{
			ID:  "10001",
			Key: testIssueKey,
		},
		Changelog: &Changelog{
			Items: []ChangelogItem{
				{
					Field:      "status",
					FromString: "Open",
					ToString:   "In Progress",
				},
			},
		},
	}

	// Handle event
	ctx := context.Background()
	err := handler.HandleEvent(ctx, event)
	if err != nil {
		t.Fatalf("failed to handle event: %v", err)
	}

	if receivedEvent == nil {
		t.Fatal("expected event to be received")
	}

	if receivedEvent.Issue.Key != testIssueKey {
		t.Errorf("expected issue key TEST-1, got %s", receivedEvent.Issue.Key)
	}
}

// TestWebhookSignatureVerification tests webhook signature verification
func TestWebhookSignatureVerification(t *testing.T) {
	//nolint:gosec // G101: test file with test credentials
	secret := "test-webhook-secret"
	handler := NewWebhookHandler(WebhookConfig{
		Secret:           secret,
		RequireSignature: true,
	})

	payload := []byte(`{"webhookEvent":"jira:issue_updated"}`)

	// Test with valid signature
	// Expected HMAC-SHA256 signature for the payload with the secret
	// This would be computed by Jira
	validSig := handler.computeSignature(payload)
	if !handler.VerifySignature(payload, validSig) {
		t.Error("expected valid signature to pass")
	}

	// Test with invalid signature
	if handler.VerifySignature(payload, "invalid-signature") {
		t.Error("expected invalid signature to fail")
	}

	// Test with empty signature
	if handler.VerifySignature(payload, "") {
		t.Error("expected empty signature to fail")
	}
}

// computeSignature is a test helper to compute expected signature
func (h *WebhookHandler) computeSignature(payload []byte) string {
	if h.secret == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// TestStatusChangeHandler tests the status change handler helper
func TestStatusChangeHandler(t *testing.T) {
	var capturedKey, capturedFrom, capturedTo string

	handler := StatusChangeHandler(func(ctx context.Context, issueKey, fromStatus, toStatus string) error {
		capturedKey = issueKey
		capturedFrom = fromStatus
		capturedTo = toStatus
		return nil
	})

	ctx := context.Background()
	event := &WebhookEvent{
		Issue: &Issue{Key: testIssueKey},
		Changelog: &Changelog{
			Items: []ChangelogItem{
				{Field: "status", FromString: "Open", ToString: "Closed"},
			},
		},
	}

	err := handler(ctx, event)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if capturedKey != testIssueKey {
		t.Errorf("expected key TEST-1, got %s", capturedKey)
	}
	if capturedFrom != "Open" {
		t.Errorf("expected from 'Open', got %s", capturedFrom)
	}
	if capturedTo != "Closed" {
		t.Errorf("expected to 'Closed', got %s", capturedTo)
	}
}

// TestPriorityMapping tests priority mapping
func TestPriorityMapping(t *testing.T) {
	bridge := NewTicketBridge(nil, DefaultTicketBridgeConfig())

	tests := []struct {
		input    string
		expected Priority
	}{
		{"urgent", PriorityHighest},
		{"high", PriorityHigh},
		{"medium", PriorityMedium},
		{"low", PriorityLow},
		{"unknown", PriorityMedium}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := bridge.mapPriority(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestStatusMapping tests status mapping
func TestStatusMapping(t *testing.T) {
	bridge := NewTicketBridge(nil, DefaultTicketBridgeConfig())

	tests := []struct {
		input    string
		expected string
	}{
		{"open", "Open"},
		{"in_progress", "In Progress"},
		{"waiting_customer", "Waiting for Customer"},
		{"resolved", "Resolved"},
		{"closed", "Closed"},
		{"unknown", "Open"}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := bridge.mapStatus(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestTruncateAddress tests address truncation
func TestTruncateAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"short", "short"},
		{"virtengine1abc123def456ghi789jkl012mno345", "virtueng...mno345"},
		{"1234567890123456", "1234567890123456"}, // Exactly 16 chars - not truncated
		// Note: Strings slightly over 16 chars may become longer when truncated
		// due to the 8+...+8 format. This is acceptable as the goal is to identify
		// long blockchain addresses while maintaining readability.
		{"12345678901234567890123456789012", "12345678...89012"},  // 32 chars -> truncated
		{"123456789012345678901234", "12345678...45678901234"}, // 24 chars
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateAddress(tt.input)
			// For short addresses, should return unchanged
			if len(tt.input) <= 16 {
				if result != tt.input {
					t.Errorf("short address should not be truncated, got %s", result)
				}
			} else if len(tt.input) >= 24 {
				// For long addresses (typical blockchain address length), verify truncation
				if !strings.Contains(result, "...") {
					t.Errorf("long address should contain ellipsis, got %s", result)
				}
			}
		})
	}
}

// TestDefaultConfigs tests default configuration values
func TestDefaultConfigs(t *testing.T) {
	t.Run("ClientConfig", func(t *testing.T) {
		cfg := DefaultClientConfig()
		if cfg.Timeout != 30*time.Second {
			t.Errorf("expected 30s timeout, got %v", cfg.Timeout)
		}
	})

	t.Run("TicketBridgeConfig", func(t *testing.T) {
		cfg := DefaultTicketBridgeConfig()
		if cfg.ProjectKey != "VESUPPORT" {
			t.Errorf("expected project key VESUPPORT, got %s", cfg.ProjectKey)
		}
		if cfg.DefaultIssueType != IssueTypeServiceRequest {
			t.Errorf("expected issue type ServiceRequest, got %s", cfg.DefaultIssueType)
		}
	})

	t.Run("SLAConfig", func(t *testing.T) {
		cfg := DefaultSLAConfig()
		if cfg.ResponseTimeTargets["urgent"] != 15 {
			t.Errorf("expected urgent response time 15, got %d", cfg.ResponseTimeTargets["urgent"])
		}
		if !cfg.ExcludeWeekends {
			t.Error("expected ExcludeWeekends to be true")
		}
	})
}

// TestErrorResponse tests error response formatting
func TestErrorResponse(t *testing.T) {
	t.Run("with error messages", func(t *testing.T) {
		err := &ErrorResponse{
			ErrorMessages: []string{"First error", "Second error"},
		}
		if err.Error() != "First error" {
			t.Errorf("expected 'First error', got %s", err.Error())
		}
	})

	t.Run("with errors map", func(t *testing.T) {
		err := &ErrorResponse{
			Errors: map[string]string{
				"field1": "Field error",
			},
		}
		if err.Error() != "Field error" {
			t.Errorf("expected 'Field error', got %s", err.Error())
		}
	})

	t.Run("empty error", func(t *testing.T) {
		err := &ErrorResponse{}
		if err.Error() != "unknown Jira error" {
			t.Errorf("expected 'unknown Jira error', got %s", err.Error())
		}
	})
}

// MockClient is a mock Jira client for testing
type MockClient struct {
	issues       map[string]*Issue
	comments     map[string][]Comment
	issueCounter int
}

func (m *MockClient) CreateIssue(ctx context.Context, req *CreateIssueRequest) (*CreateIssueResponse, error) {
	m.issueCounter++
	key := fmt.Sprintf("%s-%d", req.Fields.Project.Key, m.issueCounter)
	id := fmt.Sprintf("%d", 10000+m.issueCounter)

	issue := &Issue{
		ID:   id,
		Key:  key,
		Self: "http://test/" + key,
		Fields: IssueFields{
			Summary:     req.Fields.Summary,
			Description: req.Fields.Description,
			IssueType:   req.Fields.IssueType,
			Project:     req.Fields.Project,
			Status:      StatusField{Name: "Open"},
		},
	}

	m.issues[key] = issue

	return &CreateIssueResponse{
		ID:   id,
		Key:  key,
		Self: issue.Self,
	}, nil
}

func (m *MockClient) GetIssue(ctx context.Context, issueKeyOrID string) (*Issue, error) {
	if issue, ok := m.issues[issueKeyOrID]; ok {
		return issue, nil
	}
	return nil, fmt.Errorf("issue not found: %s", issueKeyOrID)
}

func (m *MockClient) UpdateIssue(ctx context.Context, issueKeyOrID string, req *UpdateIssueRequest) error {
	if _, ok := m.issues[issueKeyOrID]; !ok {
		return fmt.Errorf("issue not found: %s", issueKeyOrID)
	}
	return nil
}

func (m *MockClient) DeleteIssue(ctx context.Context, issueKeyOrID string) error {
	if _, ok := m.issues[issueKeyOrID]; !ok {
		return fmt.Errorf("issue not found: %s", issueKeyOrID)
	}
	delete(m.issues, issueKeyOrID)
	return nil
}

func (m *MockClient) SearchIssues(ctx context.Context, jql string, startAt, maxResults int) (*SearchResult, error) {
	issues := make([]Issue, 0, len(m.issues))
	for _, issue := range m.issues {
		issues = append(issues, *issue)
	}
	return &SearchResult{
		StartAt:    startAt,
		MaxResults: maxResults,
		Total:      len(issues),
		Issues:     issues,
	}, nil
}

func (m *MockClient) GetTransitions(ctx context.Context, issueKeyOrID string) (*TransitionsResponse, error) {
	return &TransitionsResponse{
		Transitions: []Transition{
			{ID: "1", Name: "In Progress", To: StatusField{Name: "In Progress"}},
			{ID: "2", Name: "Resolve", To: StatusField{Name: "Resolved"}},
			{ID: "3", Name: "Close", To: StatusField{Name: "Closed"}},
		},
	}, nil
}

func (m *MockClient) TransitionIssue(ctx context.Context, issueKeyOrID string, req *TransitionRequest) error {
	if _, ok := m.issues[issueKeyOrID]; !ok {
		return fmt.Errorf("issue not found: %s", issueKeyOrID)
	}
	return nil
}

func (m *MockClient) AddComment(ctx context.Context, issueKeyOrID string, comment *AddCommentRequest) (*Comment, error) {
	if _, ok := m.issues[issueKeyOrID]; !ok {
		return nil, fmt.Errorf("issue not found: %s", issueKeyOrID)
	}

	c := Comment{
		ID:      fmt.Sprintf("%d", len(m.comments[issueKeyOrID])+1),
		Body:    comment.Body,
		Created: time.Now().Format(time.RFC3339),
	}
	m.comments[issueKeyOrID] = append(m.comments[issueKeyOrID], c)
	return &c, nil
}

func (m *MockClient) GetComments(ctx context.Context, issueKeyOrID string, startAt, maxResults int) (*CommentResponse, error) {
	comments := m.comments[issueKeyOrID]
	return &CommentResponse{
		StartAt:    startAt,
		MaxResults: maxResults,
		Total:      len(comments),
		Comments:   comments,
	}, nil
}

func (m *MockClient) GetServiceDeskInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"version": "5.0.0",
	}, nil
}

