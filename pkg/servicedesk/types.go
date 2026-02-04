package servicedesk

import (
	"fmt"
	"time"
)

// ServiceDeskType identifies the external service desk system
type ServiceDeskType string

const (
	// ServiceDeskJira represents Jira Service Desk
	ServiceDeskJira ServiceDeskType = "jira"

	// ServiceDeskWaldur represents Waldur service desk
	ServiceDeskWaldur ServiceDeskType = "waldur"
)

// String returns the string representation
func (t ServiceDeskType) String() string {
	return string(t)
}

// IsValid checks if the service desk type is valid
func (t ServiceDeskType) IsValid() bool {
	return t == ServiceDeskJira || t == ServiceDeskWaldur
}

// SyncDirection indicates the direction of synchronization
type SyncDirection string

const (
	// SyncDirectionOutbound is from VirtEngine to external
	SyncDirectionOutbound SyncDirection = "outbound"

	// SyncDirectionInbound is from external to VirtEngine
	SyncDirectionInbound SyncDirection = "inbound"
)

// SyncStatus represents the synchronization status
type SyncStatus string

const (
	// SyncStatusPending indicates sync is pending
	SyncStatusPending SyncStatus = "pending"

	// SyncStatusSynced indicates successful sync
	SyncStatusSynced SyncStatus = "synced"

	// SyncStatusFailed indicates sync failed
	SyncStatusFailed SyncStatus = "failed"

	// SyncStatusConflict indicates a conflict was detected
	SyncStatusConflict SyncStatus = "conflict"
)

// TicketFieldMapping defines the mapping between on-chain and external fields
type TicketFieldMapping struct {
	// OnChainField is the VirtEngine ticket field name
	OnChainField string `json:"on_chain_field"`

	// ExternalField is the external service desk field name
	ExternalField string `json:"external_field"`

	// Transform is an optional transformation function name
	Transform string `json:"transform,omitempty"`

	// Direction specifies sync direction (both, outbound, inbound)
	Direction string `json:"direction"`
}

// StatusMapping maps on-chain ticket status to external status
type StatusMapping struct {
	// OnChainStatus is the VirtEngine ticket status
	OnChainStatus string `json:"on_chain_status"`

	// JiraStatus is the corresponding Jira status
	JiraStatus string `json:"jira_status"`

	// WaldurStatus is the corresponding Waldur status
	WaldurStatus string `json:"waldur_status"`
}

// PriorityMapping maps on-chain ticket priority to external priority
type PriorityMapping struct {
	// OnChainPriority is the VirtEngine ticket priority
	OnChainPriority string `json:"on_chain_priority"`

	// JiraPriority is the corresponding Jira priority
	JiraPriority string `json:"jira_priority"`

	// WaldurPriority is the corresponding Waldur priority
	WaldurPriority string `json:"waldur_priority"`
}

// CategoryMapping maps on-chain ticket category to external fields
type CategoryMapping struct {
	// OnChainCategory is the VirtEngine ticket category
	OnChainCategory string `json:"on_chain_category"`

	// JiraComponent is the corresponding Jira component
	JiraComponent string `json:"jira_component"`

	// JiraLabels are additional Jira labels
	JiraLabels []string `json:"jira_labels,omitempty"`

	// WaldurType is the corresponding Waldur issue type
	WaldurType string `json:"waldur_type"`
}

// MappingSchema defines the complete mapping configuration
type MappingSchema struct {
	// Version is the schema version
	Version string `json:"version"`

	// FieldMappings are field-level mappings
	FieldMappings []TicketFieldMapping `json:"field_mappings"`

	// StatusMappings are status mappings
	StatusMappings []StatusMapping `json:"status_mappings"`

	// PriorityMappings are priority mappings
	PriorityMappings []PriorityMapping `json:"priority_mappings"`

	// CategoryMappings are category mappings
	CategoryMappings []CategoryMapping `json:"category_mappings"`

	// CustomFieldMappings maps VirtEngine fields to external custom fields
	CustomFieldMappings map[string]string `json:"custom_field_mappings"`
}

// DefaultMappingSchema returns the default mapping schema
func DefaultMappingSchema() *MappingSchema {
	return &MappingSchema{
		Version: "1.0",
		FieldMappings: []TicketFieldMapping{
			{OnChainField: "ticket_id", ExternalField: "virtengine_ticket_id", Direction: "outbound"},
			{OnChainField: "customer_address", ExternalField: "virtengine_customer", Direction: "outbound"},
			{OnChainField: "category", ExternalField: "component", Direction: "outbound"},
			{OnChainField: "priority", ExternalField: "priority", Direction: "both"},
			{OnChainField: "status", ExternalField: "status", Direction: "both"},
			{OnChainField: "assigned_to", ExternalField: "assignee", Direction: "both"},
		},
		StatusMappings: []StatusMapping{
			{OnChainStatus: "open", JiraStatus: "Open", WaldurStatus: "new"},
			{OnChainStatus: "assigned", JiraStatus: "In Progress", WaldurStatus: "in_progress"},
			{OnChainStatus: "in_progress", JiraStatus: "In Progress", WaldurStatus: "in_progress"},
			{OnChainStatus: "waiting_customer", JiraStatus: "Waiting for Customer", WaldurStatus: "waiting"},
			{OnChainStatus: "waiting_support", JiraStatus: "Waiting for Support", WaldurStatus: "waiting"},
			{OnChainStatus: "resolved", JiraStatus: "Resolved", WaldurStatus: "resolved"},
			{OnChainStatus: "closed", JiraStatus: "Closed", WaldurStatus: "closed"},
			{OnChainStatus: "archived", JiraStatus: "Closed", WaldurStatus: "closed"},
		},
		PriorityMappings: []PriorityMapping{
			{OnChainPriority: "low", JiraPriority: "Low", WaldurPriority: "low"},
			{OnChainPriority: "normal", JiraPriority: "Medium", WaldurPriority: "medium"},
			{OnChainPriority: "high", JiraPriority: "High", WaldurPriority: "high"},
			{OnChainPriority: "urgent", JiraPriority: "Highest", WaldurPriority: "critical"},
		},
		CategoryMappings: []CategoryMapping{
			{OnChainCategory: "account", JiraComponent: "Account Management", WaldurType: "account"},
			{OnChainCategory: "identity", JiraComponent: "Identity & VEID", WaldurType: "identity"},
			{OnChainCategory: "billing", JiraComponent: "Billing & Payments", WaldurType: "billing"},
			{OnChainCategory: "provider", JiraComponent: "Provider Services", WaldurType: "provider"},
			{OnChainCategory: "marketplace", JiraComponent: "Marketplace", WaldurType: "marketplace"},
			{OnChainCategory: "technical", JiraComponent: "Technical Support", WaldurType: "technical"},
			{OnChainCategory: "security", JiraComponent: "Security", WaldurType: "security"},
		},
		CustomFieldMappings: map[string]string{
			"ticket_id":        "customfield_10100",
			"ticket_number":    "customfield_10101",
			"customer_address": "customfield_10102",
			"related_entity":   "customfield_10103",
			"provider_address": "customfield_10104",
			"blockchain_tx":    "customfield_10105",
		},
	}
}

// MapOnChainStatusToJira maps on-chain status to Jira status
func (s *MappingSchema) MapOnChainStatusToJira(onChainStatus string) string {
	for _, m := range s.StatusMappings {
		if m.OnChainStatus == onChainStatus {
			return m.JiraStatus
		}
	}
	return "Open" // default
}

// MapOnChainStatusToWaldur maps on-chain status to Waldur status
func (s *MappingSchema) MapOnChainStatusToWaldur(onChainStatus string) string {
	for _, m := range s.StatusMappings {
		if m.OnChainStatus == onChainStatus {
			return m.WaldurStatus
		}
	}
	return "new" // default
}

// MapJiraStatusToOnChain maps Jira status to on-chain status
func (s *MappingSchema) MapJiraStatusToOnChain(jiraStatus string) string {
	for _, m := range s.StatusMappings {
		if m.JiraStatus == jiraStatus {
			return m.OnChainStatus
		}
	}
	return "open" // default
}

// MapWaldurStatusToOnChain maps Waldur status to on-chain status
func (s *MappingSchema) MapWaldurStatusToOnChain(waldurStatus string) string {
	for _, m := range s.StatusMappings {
		if m.WaldurStatus == waldurStatus {
			return m.OnChainStatus
		}
	}
	return "open" // default
}

// MapPriorityToJira maps on-chain priority to Jira priority
func (s *MappingSchema) MapPriorityToJira(onChainPriority string) string {
	for _, m := range s.PriorityMappings {
		if m.OnChainPriority == onChainPriority {
			return m.JiraPriority
		}
	}
	return "Medium" // default
}

// MapPriorityToWaldur maps on-chain priority to Waldur priority
func (s *MappingSchema) MapPriorityToWaldur(onChainPriority string) string {
	for _, m := range s.PriorityMappings {
		if m.OnChainPriority == onChainPriority {
			return m.WaldurPriority
		}
	}
	return "medium" // default
}

// MapCategoryToJiraComponent maps on-chain category to Jira component
func (s *MappingSchema) MapCategoryToJiraComponent(category string) string {
	for _, m := range s.CategoryMappings {
		if m.OnChainCategory == category {
			return m.JiraComponent
		}
	}
	return "General" // default
}

// ExternalTicketRef represents a reference to an external ticket
type ExternalTicketRef struct {
	// Type is the service desk type
	Type ServiceDeskType `json:"type"`

	// ExternalID is the external ticket ID (e.g., JIRA-123)
	ExternalID string `json:"external_id"`

	// ExternalURL is the URL to the external ticket
	ExternalURL string `json:"external_url"`

	// ProjectKey is the external project key
	ProjectKey string `json:"project_key"`

	// SyncStatus is the current sync status
	SyncStatus SyncStatus `json:"sync_status"`

	// LastSyncAt is the last successful sync timestamp
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`

	// LastSyncError is the last sync error message
	LastSyncError string `json:"last_sync_error,omitempty"`

	// SyncVersion is used for optimistic locking
	SyncVersion int64 `json:"sync_version"`

	// CreatedAt is when the external ticket was created
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the external ticket reference
func (r *ExternalTicketRef) Validate() error {
	if !r.Type.IsValid() {
		return fmt.Errorf("invalid service desk type: %s", r.Type)
	}
	if r.ExternalID == "" {
		return fmt.Errorf("external ID is required")
	}
	return nil
}

// TicketSyncRecord tracks sync state for a ticket
type TicketSyncRecord struct {
	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticket_id"`

	// ExternalRefs are references to external tickets
	ExternalRefs []ExternalTicketRef `json:"external_refs"`

	// LastOnChainUpdate is the last on-chain update timestamp
	LastOnChainUpdate time.Time `json:"last_on_chain_update"`

	// LastOnChainVersion is the last on-chain version (block height)
	LastOnChainVersion int64 `json:"last_on_chain_version"`

	// PendingSync indicates if there's a pending sync operation
	PendingSync bool `json:"pending_sync"`

	// ConflictResolution is the conflict resolution strategy
	ConflictResolution ConflictResolution `json:"conflict_resolution"`
}

// GetExternalRef returns the external reference for the given type
func (r *TicketSyncRecord) GetExternalRef(t ServiceDeskType) *ExternalTicketRef {
	for i := range r.ExternalRefs {
		if r.ExternalRefs[i].Type == t {
			return &r.ExternalRefs[i]
		}
	}
	return nil
}

// ConflictResolution defines how to resolve conflicts
type ConflictResolution string

const (
	// ConflictResolutionOnChainWins uses on-chain data as source of truth
	ConflictResolutionOnChainWins ConflictResolution = "on_chain_wins"

	// ConflictResolutionExternalWins uses external data as source of truth
	ConflictResolutionExternalWins ConflictResolution = "external_wins"

	// ConflictResolutionNewestWins uses the newest update
	ConflictResolutionNewestWins ConflictResolution = "newest_wins"

	// ConflictResolutionManual requires manual resolution
	ConflictResolutionManual ConflictResolution = "manual"
)

// SyncEvent represents an event that needs to be synced
type SyncEvent struct {
	// ID is the unique event ID
	ID string `json:"id"`

	// Type is the event type
	Type string `json:"type"`

	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticket_id"`

	// Direction is the sync direction
	Direction SyncDirection `json:"direction"`

	// Payload is the event payload
	Payload map[string]interface{} `json:"payload"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height (for on-chain events)
	BlockHeight int64 `json:"block_height,omitempty"`

	// RetryCount is the number of retry attempts
	RetryCount int `json:"retry_count"`

	// MaxRetries is the maximum retry attempts
	MaxRetries int `json:"max_retries"`

	// NextRetryAt is when to retry next
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// Status is the processing status
	Status SyncStatus `json:"status"`

	// Error is the last error message
	Error string `json:"error,omitempty"`
}

// CanRetry checks if the event can be retried
func (e *SyncEvent) CanRetry() bool {
	return e.RetryCount < e.MaxRetries
}

// AttachmentSync represents an attachment sync request
type AttachmentSync struct {
	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticket_id"`

	// ArtifactAddress is the artifact store content address
	ArtifactAddress string `json:"artifact_address"`

	// FileName is the original file name
	FileName string `json:"file_name"`

	// ContentType is the MIME content type
	ContentType string `json:"content_type"`

	// Size is the file size in bytes
	Size int64 `json:"size"`

	// ExternalAttachmentID is the external attachment ID after sync
	ExternalAttachmentID string `json:"external_attachment_id,omitempty"`

	// AccessToken is a temporary access token for the attachment
	AccessToken string `json:"access_token,omitempty"`

	// ExpiresAt is when the access token expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CallbackPayload is the payload for signed callbacks from external systems
type CallbackPayload struct {
	// EventType is the external event type
	EventType string `json:"event_type"`

	// ServiceDeskType is the service desk type
	ServiceDeskType ServiceDeskType `json:"service_desk_type"`

	// ExternalID is the external ticket ID
	ExternalID string `json:"external_id"`

	// OnChainTicketID is the VirtEngine ticket ID
	OnChainTicketID string `json:"on_chain_ticket_id"`

	// Changes are the field changes
	Changes map[string]interface{} `json:"changes"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Signature is the HMAC signature
	Signature string `json:"signature"`

	// Nonce is a unique nonce to prevent replay attacks
	Nonce string `json:"nonce"`
}

// Validate validates the callback payload
func (p *CallbackPayload) Validate() error {
	if p.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if !p.ServiceDeskType.IsValid() {
		return fmt.Errorf("invalid service_desk_type: %s", p.ServiceDeskType)
	}
	if p.ExternalID == "" {
		return fmt.Errorf("external_id is required")
	}
	if p.OnChainTicketID == "" {
		return fmt.Errorf("on_chain_ticket_id is required")
	}
	if p.Signature == "" {
		return fmt.Errorf("signature is required")
	}
	if p.Nonce == "" {
		return fmt.Errorf("nonce is required")
	}
	return nil
}
