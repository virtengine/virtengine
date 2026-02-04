package servicedesk

import (
	"testing"
	"time"
)

// Test constants for repeated string literals
const (
	testStatusOpen = "open"
)

func TestDefaultMappingSchema(t *testing.T) {
	schema := DefaultMappingSchema()

	if schema.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", schema.Version)
	}

	// Test status mappings
	tests := []struct {
		onChain string
		jira    string
		waldur  string
	}{
		{testStatusOpen, "Open", "new"},
		{"assigned", "In Progress", "in_progress"},
		{"in_progress", "In Progress", "in_progress"},
		{"waiting_customer", "Waiting for Customer", "waiting"},
		{"resolved", "Resolved", "resolved"},
		{"closed", "Closed", "closed"},
	}

	for _, tt := range tests {
		jira := schema.MapOnChainStatusToJira(tt.onChain)
		if jira != tt.jira {
			t.Errorf("MapOnChainStatusToJira(%s) = %s, want %s", tt.onChain, jira, tt.jira)
		}

		waldur := schema.MapOnChainStatusToWaldur(tt.onChain)
		if waldur != tt.waldur {
			t.Errorf("MapOnChainStatusToWaldur(%s) = %s, want %s", tt.onChain, waldur, tt.waldur)
		}
	}
}

func TestMapStatusRoundTrip(t *testing.T) {
	schema := DefaultMappingSchema()

	// Test Jira round trip
	jiraStatus := schema.MapOnChainStatusToJira(testStatusOpen)
	onChain := schema.MapJiraStatusToOnChain(jiraStatus)
	if onChain != testStatusOpen {
		t.Errorf("Jira round trip failed: open -> %s -> %s", jiraStatus, onChain)
	}

	// Test Waldur round trip
	waldurStatus := schema.MapOnChainStatusToWaldur("resolved")
	onChain = schema.MapWaldurStatusToOnChain(waldurStatus)
	if onChain != "resolved" {
		t.Errorf("Waldur round trip failed: resolved -> %s -> %s", waldurStatus, onChain)
	}
}

func TestPriorityMapping(t *testing.T) {
	schema := DefaultMappingSchema()

	tests := []struct {
		onChain string
		jira    string
		waldur  string
	}{
		{"low", "Low", "low"},
		{"normal", "Medium", "medium"},
		{"high", "High", "high"},
		{"urgent", "Highest", "critical"},
	}

	for _, tt := range tests {
		jira := schema.MapPriorityToJira(tt.onChain)
		if jira != tt.jira {
			t.Errorf("MapPriorityToJira(%s) = %s, want %s", tt.onChain, jira, tt.jira)
		}

		waldur := schema.MapPriorityToWaldur(tt.onChain)
		if waldur != tt.waldur {
			t.Errorf("MapPriorityToWaldur(%s) = %s, want %s", tt.onChain, waldur, tt.waldur)
		}
	}
}

func TestCategoryMapping(t *testing.T) {
	schema := DefaultMappingSchema()

	tests := []struct {
		category  string
		component string
	}{
		{"account", "Account Management"},
		{"identity", "Identity & VEID"},
		{"billing", "Billing & Payments"},
		{"provider", "Provider Services"},
		{"marketplace", "Marketplace"},
		{"technical", "Technical Support"},
		{"security", "Security"},
	}

	for _, tt := range tests {
		component := schema.MapCategoryToJiraComponent(tt.category)
		if component != tt.component {
			t.Errorf("MapCategoryToJiraComponent(%s) = %s, want %s", tt.category, component, tt.component)
		}
	}
}

func TestExternalTicketRefValidation(t *testing.T) {
	tests := []struct {
		name    string
		ref     ExternalTicketRef
		wantErr bool
	}{
		{
			name: "valid jira ref",
			ref: ExternalTicketRef{
				Type:       ServiceDeskJira,
				ExternalID: "PROJ-123",
			},
			wantErr: false,
		},
		{
			name: "valid waldur ref",
			ref: ExternalTicketRef{
				Type:       ServiceDeskWaldur,
				ExternalID: "abc-123",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			ref: ExternalTicketRef{
				Type:       "invalid",
				ExternalID: "123",
			},
			wantErr: true,
		},
		{
			name: "missing external ID",
			ref: ExternalTicketRef{
				Type:       ServiceDeskJira,
				ExternalID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCallbackPayloadValidation(t *testing.T) {
	tests := []struct {
		name    string
		payload CallbackPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			payload: CallbackPayload{
				EventType:       "status_changed",
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      "PROJ-123",
				OnChainTicketID: "TKT-00000001",
				Signature:       "abc123",
				Nonce:           "nonce123",
				Timestamp:       time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing event type",
			payload: CallbackPayload{
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      "PROJ-123",
				OnChainTicketID: "TKT-00000001",
				Signature:       "abc123",
				Nonce:           "nonce123",
			},
			wantErr: true,
		},
		{
			name: "missing signature",
			payload: CallbackPayload{
				EventType:       "status_changed",
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      "PROJ-123",
				OnChainTicketID: "TKT-00000001",
				Nonce:           "nonce123",
			},
			wantErr: true,
		},
		{
			name: "missing nonce",
			payload: CallbackPayload{
				EventType:       "status_changed",
				ServiceDeskType: ServiceDeskJira,
				ExternalID:      "PROJ-123",
				OnChainTicketID: "TKT-00000001",
				Signature:       "abc123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyncEventCanRetry(t *testing.T) {
	event := &SyncEvent{
		ID:         "test-1",
		MaxRetries: 3,
		RetryCount: 0,
	}

	if !event.CanRetry() {
		t.Error("expected CanRetry() = true for RetryCount 0")
	}

	event.RetryCount = 2
	if !event.CanRetry() {
		t.Error("expected CanRetry() = true for RetryCount 2")
	}

	event.RetryCount = 3
	if event.CanRetry() {
		t.Error("expected CanRetry() = false for RetryCount 3")
	}

	event.RetryCount = 5
	if event.CanRetry() {
		t.Error("expected CanRetry() = false for RetryCount 5")
	}
}

func TestTicketSyncRecordGetExternalRef(t *testing.T) {
	record := &TicketSyncRecord{
		TicketID: "TKT-00000001",
		ExternalRefs: []ExternalTicketRef{
			{Type: ServiceDeskJira, ExternalID: "PROJ-123"},
			{Type: ServiceDeskWaldur, ExternalID: "waldur-456"},
		},
	}

	jiraRef := record.GetExternalRef(ServiceDeskJira)
	if jiraRef == nil {
		t.Fatal("expected to find Jira ref")
	}
	if jiraRef.ExternalID != "PROJ-123" {
		t.Errorf("expected ExternalID PROJ-123, got %s", jiraRef.ExternalID)
	}

	waldurRef := record.GetExternalRef(ServiceDeskWaldur)
	if waldurRef == nil {
		t.Fatal("expected to find Waldur ref")
	}
	if waldurRef.ExternalID != "waldur-456" {
		t.Errorf("expected ExternalID waldur-456, got %s", waldurRef.ExternalID)
	}

	// Test non-existent
	unknownRef := record.GetExternalRef("unknown")
	if unknownRef != nil {
		t.Error("expected nil for unknown service desk type")
	}
}

func TestServiceDeskTypeIsValid(t *testing.T) {
	tests := []struct {
		t       ServiceDeskType
		isValid bool
	}{
		{ServiceDeskJira, true},
		{ServiceDeskWaldur, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		if tt.t.IsValid() != tt.isValid {
			t.Errorf("ServiceDeskType(%s).IsValid() = %v, want %v", tt.t, tt.t.IsValid(), tt.isValid)
		}
	}
}
