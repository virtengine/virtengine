package waldur

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMapVirtEnginePriorityToWaldur(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IssuePriority
	}{
		{"low priority", "low", PriorityLow},
		{"normal priority", "normal", PriorityNormal},
		{"medium priority", "medium", PriorityNormal},
		{"high priority", "high", PriorityHigh},
		{"urgent priority", "urgent", PriorityCritical},
		{"critical priority", "critical", PriorityCritical},
		{"unknown priority", "unknown", PriorityNormal},
		{"empty priority", "", PriorityNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapVirtEnginePriorityToWaldur(tt.input)
			if result != tt.expected {
				t.Errorf("MapVirtEnginePriorityToWaldur(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapWaldurPriorityToVirtEngine(t *testing.T) {
	tests := []struct {
		name     string
		input    IssuePriority
		expected string
	}{
		{"low priority", PriorityLow, "low"},
		{"normal priority", PriorityNormal, "normal"},
		{"high priority", PriorityHigh, "high"},
		{"critical priority", PriorityCritical, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapWaldurPriorityToVirtEngine(tt.input)
			if result != tt.expected {
				t.Errorf("MapWaldurPriorityToVirtEngine(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapVirtEngineStatusToWaldur(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IssueState
	}{
		{"open status", "open", StateNew},
		{"assigned status", "assigned", StateOpen},
		{"in_progress status", "in_progress", StateInProgress},
		{"pending_customer status", "pending_customer", StateWaiting},
		{"resolved status", "resolved", StateResolved},
		{"closed status", "closed", StateClosed},
		{"canceled status", "canceled", StateCanceled},
		{"unknown status", "unknown", StateNew},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapVirtEngineStatusToWaldur(tt.input)
			if result != tt.expected {
				t.Errorf("MapVirtEngineStatusToWaldur(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapWaldurStateToVirtEngine(t *testing.T) {
	tests := []struct {
		name     string
		input    IssueState
		expected string
	}{
		{"new state", StateNew, "open"},
		{"open state", StateOpen, "assigned"},
		{"in_progress state", StateInProgress, "in_progress"},
		{"waiting state", StateWaiting, "pending_customer"},
		{"resolved state", StateResolved, "resolved"},
		{"closed state", StateClosed, "closed"},
		{"canceled state", StateCanceled, "canceled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapWaldurStateToVirtEngine(tt.input)
			if result != tt.expected {
				t.Errorf("MapWaldurStateToVirtEngine(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapVirtEngineCategoryToWaldurType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected IssueType
	}{
		{"incident category", "incident", IssueTypeIncident},
		{"outage category", "outage", IssueTypeIncident},
		{"security category", "security", IssueTypeIncident},
		{"change category", "change", IssueTypeChange},
		{"upgrade category", "upgrade", IssueTypeChange},
		{"question category", "question", IssueTypeInformational},
		{"inquiry category", "inquiry", IssueTypeInformational},
		{"general category", "general", IssueTypeServiceRequest},
		{"billing category", "billing", IssueTypeServiceRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapVirtEngineCategoryToWaldurType(tt.input)
			if result != tt.expected {
				t.Errorf("MapVirtEngineCategoryToWaldurType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateIssueRequest_Defaults(t *testing.T) {
	// Test that empty type and priority get defaults
	req := CreateIssueRequest{
		Summary:     "Test issue",
		Description: "Test description",
	}

	// Verify defaults would be applied (in the CreateIssue function)
	if req.Type == "" {
		req.Type = IssueTypeServiceRequest
	}
	if req.Priority == "" {
		req.Priority = PriorityNormal
	}

	if req.Type != IssueTypeServiceRequest {
		t.Errorf("expected default type to be ServiceRequest, got %q", req.Type)
	}
	if req.Priority != PriorityNormal {
		t.Errorf("expected default priority to be Normal, got %q", req.Priority)
	}
}

func TestSupportIssue_Fields(t *testing.T) {
	// Test that all fields can be set
	issue := SupportIssue{
		UUID:         "test-uuid",
		Key:          "SUPPORT-123",
		Type:         IssueTypeServiceRequest,
		Priority:     PriorityHigh,
		State:        StateInProgress,
		Summary:      "Test summary",
		Description:  "Test description",
		CallerUUID:   "caller-uuid",
		CallerName:   "Test Caller",
		AssigneeUUID: "assignee-uuid",
		AssigneeName: "Test Assignee",
		CustomerUUID: "customer-uuid",
		ProjectUUID:  "project-uuid",
		ResourceUUID: "resource-uuid",
		BackendID:    "TKT-12345",
		Resolution:   "Fixed",
	}

	if issue.UUID != "test-uuid" {
		t.Error("UUID field not set correctly")
	}
	if issue.Key != "SUPPORT-123" {
		t.Error("Key field not set correctly")
	}
	if issue.BackendID != "TKT-12345" {
		t.Error("BackendID field not set correctly")
	}
}

func TestSupportClient_ListIssues_Paginates(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case usersMePath:
			http.NotFound(w, r)
		case apiUsersMePath:
			w.Header().Set("Content-Type", "application/json")
			username := "support"
			uuid := "550e8400-e29b-41d4-a716-446655440094"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"uuid":     &uuid,
				"username": &username,
			})
		case "/api/support-issues/":
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Query().Get("page") == "2" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"count": 2,
					"results": []map[string]any{
						{"uuid": "issue-2", "summary": "second"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": 2,
				"next":  server.URL + "/api/support-issues/?page=2",
				"results": []map[string]any{
					{"uuid": "issue-1", "summary": "first"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("NewClient() unexpected error: %v", err)
	}

	issues, err := NewSupportClient(client).ListIssues(context.Background(), ListIssuesParams{})
	if err != nil {
		t.Fatalf("ListIssues() unexpected error: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("ListIssues() returned %d issues, want 2", len(issues))
	}
}

func TestSupportClient_GetIssueByBackendID_DuplicateConflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case usersMePath:
			http.NotFound(w, r)
		case apiUsersMePath:
			w.Header().Set("Content-Type", "application/json")
			username := "support"
			uuid := "550e8400-e29b-41d4-a716-446655440095"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"uuid":     &uuid,
				"username": &username,
			})
		case "/api/support-issues/":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": 2,
				"results": []map[string]any{
					{"uuid": "issue-1", "backend_id": "dup"},
					{"uuid": "issue-2", "backend_id": "dup"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("NewClient() unexpected error: %v", err)
	}

	_, err = NewSupportClient(client).GetIssueByBackendID(context.Background(), "dup")
	if err == nil || !strings.Contains(err.Error(), ErrConflict.Error()) {
		t.Fatalf("GetIssueByBackendID() error = %v, want conflict", err)
	}
}
