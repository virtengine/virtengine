package types

import (
	"fmt"
	"time"
)

// SupportEventType identifies support event types.
type SupportEventType string

const (
	SupportEventTypeRequestCreated       SupportEventType = "support_request_created"
	SupportEventTypeRequestUpdated       SupportEventType = "support_request_updated"
	SupportEventTypeStatusChanged        SupportEventType = "support_request_status_changed"
	SupportEventTypeResponseAdded        SupportEventType = "support_response_added"
	SupportEventTypeRequestArchived      SupportEventType = "support_request_archived"
	SupportEventTypeRequestPurged        SupportEventType = "support_request_purged"
	SupportEventTypeExternalTicketLinked SupportEventType = "support_external_ticket_linked"
)

// SupportEvent represents a support event for off-chain bridges.
type SupportEvent interface {
	GetEventType() SupportEventType
	GetEventID() string
	GetBlockHeight() int64
	GetSequence() uint64
}

// SupportEventWrapper is the typed event emitted for subscription.
type SupportEventWrapper struct {
	EventType   string `json:"event_type"`
	EventID     string `json:"event_id"`
	BlockHeight int64  `json:"block_height"`
	Sequence    uint64 `json:"sequence"`
}

// ProtoMessage interface stubs.
func (m *SupportEventWrapper) ProtoMessage()  {}
func (m *SupportEventWrapper) Reset()         { *m = SupportEventWrapper{} }
func (m *SupportEventWrapper) String() string { return fmt.Sprintf("%+v", *m) }

// SupportEventCheckpoint tracks event consumption for a subscriber.
type SupportEventCheckpoint struct {
	SubscriberID string    `json:"subscriber_id"`
	Sequence     uint64    `json:"sequence"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Validate validates the checkpoint.
func (c *SupportEventCheckpoint) Validate() error {
	if c == nil {
		return ErrInvalidParams.Wrap("checkpoint is nil")
	}
	if c.SubscriberID == "" {
		return ErrInvalidParams.Wrap("subscriber_id is required")
	}
	return nil
}

// SupportRequestCreatedEvent is emitted when a request is created.
type SupportRequestCreatedEvent struct {
	EventType     string                   `json:"event_type"`
	EventID       string                   `json:"event_id"`
	BlockHeight   int64                    `json:"block_height"`
	Sequence      uint64                   `json:"sequence"`
	TicketID      string                   `json:"ticket_id"`
	TicketNumber  string                   `json:"ticket_number"`
	Submitter     string                   `json:"submitter"`
	Category      string                   `json:"category"`
	Priority      string                   `json:"priority"`
	Status        string                   `json:"status"`
	PayloadHash   string                   `json:"payload_hash,omitempty"`
	EnvelopeRef   string                   `json:"envelope_ref,omitempty"`
	Payload       *EncryptedSupportPayload `json:"payload,omitempty"`
	Recipients    []string                 `json:"recipients,omitempty"`
	RelatedEntity *RelatedEntity           `json:"related_entity,omitempty"`
	Timestamp     int64                    `json:"timestamp"`
}

func (e SupportRequestCreatedEvent) GetEventType() SupportEventType {
	return SupportEventTypeRequestCreated
}
func (e SupportRequestCreatedEvent) GetEventID() string    { return e.EventID }
func (e SupportRequestCreatedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportRequestCreatedEvent) GetSequence() uint64   { return e.Sequence }

// SupportRequestUpdatedEvent is emitted when a request is updated.
type SupportRequestUpdatedEvent struct {
	EventType     string                   `json:"event_type"`
	EventID       string                   `json:"event_id"`
	BlockHeight   int64                    `json:"block_height"`
	Sequence      uint64                   `json:"sequence"`
	TicketID      string                   `json:"ticket_id"`
	UpdatedBy     string                   `json:"updated_by"`
	Priority      string                   `json:"priority,omitempty"`
	Category      string                   `json:"category,omitempty"`
	AssignedAgent string                   `json:"assigned_agent,omitempty"`
	Status        string                   `json:"status,omitempty"`
	PayloadHash   string                   `json:"payload_hash,omitempty"`
	EnvelopeRef   string                   `json:"envelope_ref,omitempty"`
	Payload       *EncryptedSupportPayload `json:"payload,omitempty"`
	Timestamp     int64                    `json:"timestamp"`
}

func (e SupportRequestUpdatedEvent) GetEventType() SupportEventType {
	return SupportEventTypeRequestUpdated
}
func (e SupportRequestUpdatedEvent) GetEventID() string    { return e.EventID }
func (e SupportRequestUpdatedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportRequestUpdatedEvent) GetSequence() uint64   { return e.Sequence }

// SupportStatusChangedEvent is emitted when status changes.
type SupportStatusChangedEvent struct {
	EventType   string `json:"event_type"`
	EventID     string `json:"event_id"`
	BlockHeight int64  `json:"block_height"`
	Sequence    uint64 `json:"sequence"`
	TicketID    string `json:"ticket_id"`
	OldStatus   string `json:"old_status"`
	NewStatus   string `json:"new_status"`
	UpdatedBy   string `json:"updated_by"`
	Timestamp   int64  `json:"timestamp"`
}

func (e SupportStatusChangedEvent) GetEventType() SupportEventType {
	return SupportEventTypeStatusChanged
}
func (e SupportStatusChangedEvent) GetEventID() string    { return e.EventID }
func (e SupportStatusChangedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportStatusChangedEvent) GetSequence() uint64   { return e.Sequence }

// SupportResponseAddedEvent is emitted when a response is added.
type SupportResponseAddedEvent struct {
	EventType   string                   `json:"event_type"`
	EventID     string                   `json:"event_id"`
	BlockHeight int64                    `json:"block_height"`
	Sequence    uint64                   `json:"sequence"`
	TicketID    string                   `json:"ticket_id"`
	ResponseID  string                   `json:"response_id"`
	Author      string                   `json:"author"`
	IsAgent     bool                     `json:"is_agent"`
	PayloadHash string                   `json:"payload_hash,omitempty"`
	EnvelopeRef string                   `json:"envelope_ref,omitempty"`
	Payload     *EncryptedSupportPayload `json:"payload,omitempty"`
	Timestamp   int64                    `json:"timestamp"`
}

func (e SupportResponseAddedEvent) GetEventType() SupportEventType {
	return SupportEventTypeResponseAdded
}
func (e SupportResponseAddedEvent) GetEventID() string    { return e.EventID }
func (e SupportResponseAddedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportResponseAddedEvent) GetSequence() uint64   { return e.Sequence }

// SupportRequestArchivedEvent is emitted when a request is archived.
type SupportRequestArchivedEvent struct {
	EventType   string `json:"event_type"`
	EventID     string `json:"event_id"`
	BlockHeight int64  `json:"block_height"`
	Sequence    uint64 `json:"sequence"`
	TicketID    string `json:"ticket_id"`
	ArchivedBy  string `json:"archived_by"`
	Reason      string `json:"reason,omitempty"`
	Timestamp   int64  `json:"timestamp"`
}

func (e SupportRequestArchivedEvent) GetEventType() SupportEventType {
	return SupportEventTypeRequestArchived
}
func (e SupportRequestArchivedEvent) GetEventID() string    { return e.EventID }
func (e SupportRequestArchivedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportRequestArchivedEvent) GetSequence() uint64   { return e.Sequence }

// SupportRequestPurgedEvent is emitted when a payload is purged.
type SupportRequestPurgedEvent struct {
	EventType   string `json:"event_type"`
	EventID     string `json:"event_id"`
	BlockHeight int64  `json:"block_height"`
	Sequence    uint64 `json:"sequence"`
	TicketID    string `json:"ticket_id"`
	PurgedBy    string `json:"purged_by"`
	Reason      string `json:"reason,omitempty"`
	Timestamp   int64  `json:"timestamp"`
}

func (e SupportRequestPurgedEvent) GetEventType() SupportEventType {
	return SupportEventTypeRequestPurged
}
func (e SupportRequestPurgedEvent) GetEventID() string    { return e.EventID }
func (e SupportRequestPurgedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportRequestPurgedEvent) GetSequence() uint64   { return e.Sequence }

// SupportExternalTicketLinkedEvent is emitted when an external ticket is linked.
type SupportExternalTicketLinkedEvent struct {
	EventType        string `json:"event_type"`
	EventID          string `json:"event_id"`
	BlockHeight      int64  `json:"block_height"`
	Sequence         uint64 `json:"sequence"`
	TicketID         string `json:"ticket_id"`
	ExternalSystem   string `json:"external_system"`
	ExternalTicketID string `json:"external_ticket_id"`
	ExternalURL      string `json:"external_url,omitempty"`
	LinkedBy         string `json:"linked_by"`
	Timestamp        int64  `json:"timestamp"`
}

func (e SupportExternalTicketLinkedEvent) GetEventType() SupportEventType {
	return SupportEventTypeExternalTicketLinked
}
func (e SupportExternalTicketLinkedEvent) GetEventID() string    { return e.EventID }
func (e SupportExternalTicketLinkedEvent) GetBlockHeight() int64 { return e.BlockHeight }
func (e SupportExternalTicketLinkedEvent) GetSequence() uint64   { return e.Sequence }
