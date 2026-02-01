package types

import "fmt"

// Event types for the support module
const (
	EventTypeExternalTicketRegistered = "external_ticket_registered"
	EventTypeExternalTicketUpdated    = "external_ticket_updated"
	EventTypeExternalTicketRemoved    = "external_ticket_removed"
)

// Event attribute keys
const (
	AttributeKeyResourceID       = "resource_id"
	AttributeKeyResourceType     = "resource_type"
	AttributeKeyExternalSystem   = "external_system"
	AttributeKeyExternalTicketID = "external_ticket_id"
	AttributeKeyExternalURL      = "external_url"
	AttributeKeyCreatedBy        = "created_by"
	AttributeKeyBlockHeight      = "block_height"
	AttributeKeyTimestamp        = "timestamp"
)

// EventExternalTicketRegistered is emitted when an external ticket reference is registered
type EventExternalTicketRegistered struct {
	ResourceID       string `json:"resource_id"`
	ResourceType     string `json:"resource_type"`
	ExternalSystem   string `json:"external_system"`
	ExternalTicketID string `json:"external_ticket_id"`
	ExternalURL      string `json:"external_url,omitempty"`
	CreatedBy        string `json:"created_by"`
	BlockHeight      int64  `json:"block_height"`
	Timestamp        int64  `json:"timestamp"`
}

// EventExternalTicketUpdated is emitted when an external ticket reference is updated
type EventExternalTicketUpdated struct {
	ResourceID       string `json:"resource_id"`
	ResourceType     string `json:"resource_type"`
	ExternalTicketID string `json:"external_ticket_id"`
	ExternalURL      string `json:"external_url,omitempty"`
	UpdatedBy        string `json:"updated_by"`
	BlockHeight      int64  `json:"block_height"`
	Timestamp        int64  `json:"timestamp"`
}

// EventExternalTicketRemoved is emitted when an external ticket reference is removed
type EventExternalTicketRemoved struct {
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	RemovedBy    string `json:"removed_by"`
	BlockHeight  int64  `json:"block_height"`
	Timestamp    int64  `json:"timestamp"`
}

// Proto message interface stubs for Event types

func (*EventExternalTicketRegistered) ProtoMessage()    {}
func (m *EventExternalTicketRegistered) Reset()         { *m = EventExternalTicketRegistered{} }
func (m *EventExternalTicketRegistered) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventExternalTicketUpdated) ProtoMessage()    {}
func (m *EventExternalTicketUpdated) Reset()         { *m = EventExternalTicketUpdated{} }
func (m *EventExternalTicketUpdated) String() string { return fmt.Sprintf("%+v", *m) }

func (*EventExternalTicketRemoved) ProtoMessage()    {}
func (m *EventExternalTicketRemoved) Reset()         { *m = EventExternalTicketRemoved{} }
func (m *EventExternalTicketRemoved) String() string { return fmt.Sprintf("%+v", *m) }
