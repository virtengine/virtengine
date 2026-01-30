package types

import (
	"fmt"
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// TicketStatus represents the lifecycle status of a support ticket
type TicketStatus uint8

const (
	// TicketStatusUnspecified is the default/invalid status
	TicketStatusUnspecified TicketStatus = iota

	// TicketStatusOpen is a newly created ticket awaiting assignment
	TicketStatusOpen

	// TicketStatusAssigned is assigned to an agent but not yet in progress
	TicketStatusAssigned

	// TicketStatusInProgress is actively being worked on
	TicketStatusInProgress

	// TicketStatusPendingCustomer is waiting for customer response
	TicketStatusPendingCustomer

	// TicketStatusResolved is resolved by the agent
	TicketStatusResolved

	// TicketStatusClosed is closed (by customer or admin)
	TicketStatusClosed
)

// String returns the string representation of a ticket status
func (s TicketStatus) String() string {
	switch s {
	case TicketStatusUnspecified:
		return "unspecified"
	case TicketStatusOpen:
		return "open"
	case TicketStatusAssigned:
		return "assigned"
	case TicketStatusInProgress:
		return "in_progress"
	case TicketStatusPendingCustomer:
		return "pending_customer"
	case TicketStatusResolved:
		return "resolved"
	case TicketStatusClosed:
		return "closed"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// TicketStatusFromString converts a string to a TicketStatus
func TicketStatusFromString(s string) (TicketStatus, error) {
	switch s {
	case "open":
		return TicketStatusOpen, nil
	case "assigned":
		return TicketStatusAssigned, nil
	case "in_progress":
		return TicketStatusInProgress, nil
	case "pending_customer":
		return TicketStatusPendingCustomer, nil
	case "resolved":
		return TicketStatusResolved, nil
	case "closed":
		return TicketStatusClosed, nil
	default:
		return TicketStatusUnspecified, fmt.Errorf("unknown ticket status: %s", s)
	}
}

// IsValid checks if the status is a valid ticket status
func (s TicketStatus) IsValid() bool {
	return s >= TicketStatusOpen && s <= TicketStatusClosed
}

// IsActive checks if the ticket is still active (not closed)
func (s TicketStatus) IsActive() bool {
	return s != TicketStatusClosed && s != TicketStatusUnspecified
}

// CanTransitionTo checks if transition to target status is allowed
func (s TicketStatus) CanTransitionTo(target TicketStatus) bool {
	switch s {
	case TicketStatusOpen:
		return target == TicketStatusAssigned || target == TicketStatusClosed
	case TicketStatusAssigned:
		return target == TicketStatusInProgress || target == TicketStatusOpen || target == TicketStatusClosed
	case TicketStatusInProgress:
		return target == TicketStatusPendingCustomer || target == TicketStatusResolved || target == TicketStatusClosed
	case TicketStatusPendingCustomer:
		return target == TicketStatusInProgress || target == TicketStatusResolved || target == TicketStatusClosed
	case TicketStatusResolved:
		return target == TicketStatusClosed || target == TicketStatusOpen // reopen
	case TicketStatusClosed:
		return target == TicketStatusOpen // reopen
	default:
		return false
	}
}

// TicketPriority represents the priority level of a support ticket
type TicketPriority uint8

const (
	// TicketPriorityUnspecified is the default/invalid priority
	TicketPriorityUnspecified TicketPriority = iota

	// TicketPriorityLow is for non-urgent issues
	TicketPriorityLow

	// TicketPriorityNormal is the default priority
	TicketPriorityNormal

	// TicketPriorityHigh is for urgent issues
	TicketPriorityHigh

	// TicketPriorityUrgent is for critical issues requiring immediate attention
	TicketPriorityUrgent
)

// String returns the string representation of a ticket priority
func (p TicketPriority) String() string {
	switch p {
	case TicketPriorityUnspecified:
		return "unspecified"
	case TicketPriorityLow:
		return "low"
	case TicketPriorityNormal:
		return "normal"
	case TicketPriorityHigh:
		return "high"
	case TicketPriorityUrgent:
		return "urgent"
	default:
		return fmt.Sprintf("unknown(%d)", p)
	}
}

// TicketPriorityFromString converts a string to a TicketPriority
func TicketPriorityFromString(s string) (TicketPriority, error) {
	switch s {
	case "low":
		return TicketPriorityLow, nil
	case "normal":
		return TicketPriorityNormal, nil
	case "high":
		return TicketPriorityHigh, nil
	case "urgent":
		return TicketPriorityUrgent, nil
	default:
		return TicketPriorityUnspecified, fmt.Errorf("unknown ticket priority: %s", s)
	}
}

// IsValid checks if the priority is a valid ticket priority
func (p TicketPriority) IsValid() bool {
	return p >= TicketPriorityLow && p <= TicketPriorityUrgent
}

// ResourceReference identifies a related on-chain resource (order, lease, deployment)
type ResourceReference struct {
	// Type is the resource type (e.g., "order", "lease", "deployment")
	Type string `json:"type"`

	// ID is the resource identifier
	ID string `json:"id"`

	// Owner is the resource owner address
	Owner string `json:"owner,omitempty"`
}

// Validate validates the resource reference
func (r ResourceReference) Validate() error {
	if r.Type == "" && r.ID == "" {
		// Empty reference is valid (ticket may not be related to a specific resource)
		return nil
	}

	validTypes := map[string]bool{
		"order":      true,
		"lease":      true,
		"deployment": true,
		"provider":   true,
		"bid":        true,
	}

	if !validTypes[r.Type] {
		return ErrInvalidResourceRef.Wrapf("invalid resource type: %s", r.Type)
	}

	if r.ID == "" {
		return ErrInvalidResourceRef.Wrap("resource ID required when type is specified")
	}

	return nil
}

// IsEmpty checks if the resource reference is empty
func (r ResourceReference) IsEmpty() bool {
	return r.Type == "" && r.ID == ""
}

// SupportTicket represents an encrypted support request stored on-chain
type SupportTicket struct {
	// TicketID is the unique ticket identifier
	TicketID string `json:"ticket_id"`

	// CustomerAddress is the ticket creator's address
	CustomerAddress string `json:"customer_address"`

	// ProviderAddress is the related provider (optional)
	ProviderAddress string `json:"provider_address,omitempty"`

	// ResourceRef is the related resource reference (optional)
	ResourceRef ResourceReference `json:"resource_ref,omitempty"`

	// Status is the current ticket status
	Status TicketStatus `json:"status"`

	// Priority is the ticket priority
	Priority TicketPriority `json:"priority"`

	// Category is the issue category
	Category string `json:"category"`

	// EncryptedPayload contains the encrypted ticket content
	// Uses MultiRecipientEnvelope for multi-party decryption
	EncryptedPayload encryptiontypes.MultiRecipientEnvelope `json:"encrypted_payload"`

	// AssignedTo is the assigned support agent address
	AssignedTo string `json:"assigned_to,omitempty"`

	// ResponseCount is the number of responses on this ticket
	ResponseCount uint32 `json:"response_count"`

	// CreatedAt is the ticket creation timestamp
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time `json:"updated_at"`

	// ResolvedAt is when the ticket was resolved (if applicable)
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// ClosedAt is when the ticket was closed (if applicable)
	ClosedAt *time.Time `json:"closed_at,omitempty"`

	// Resolution summary (encrypted reference or hash)
	ResolutionRef string `json:"resolution_ref,omitempty"`
}

// NewSupportTicket creates a new support ticket
func NewSupportTicket(
	ticketID string,
	customer string,
	category string,
	priority TicketPriority,
	encPayload encryptiontypes.MultiRecipientEnvelope,
	createdAt time.Time,
) *SupportTicket {
	return &SupportTicket{
		TicketID:         ticketID,
		CustomerAddress:  customer,
		Status:           TicketStatusOpen,
		Priority:         priority,
		Category:         category,
		EncryptedPayload: encPayload,
		ResponseCount:    0,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}
}

// Validate validates the support ticket
func (t *SupportTicket) Validate() error {
	if t.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	if t.CustomerAddress == "" {
		return ErrInvalidAddress.Wrap("customer address cannot be empty")
	}

	if !t.Status.IsValid() {
		return ErrInvalidTicketStatus.Wrapf("invalid status: %d", t.Status)
	}

	if !t.Priority.IsValid() {
		return ErrInvalidTicketPriority.Wrapf("invalid priority: %d", t.Priority)
	}

	if t.Category == "" {
		return ErrInvalidCategory.Wrap("category cannot be empty")
	}

	if err := t.ResourceRef.Validate(); err != nil {
		return err
	}

	if err := t.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidEncryptedPayload.Wrap(err.Error())
	}

	return nil
}

// TicketResponse represents an encrypted response to a support ticket
type TicketResponse struct {
	// TicketID is the parent ticket identifier
	TicketID string `json:"ticket_id"`

	// ResponseIndex is the sequential response number
	ResponseIndex uint32 `json:"response_index"`

	// ResponderAddress is the response author's address
	ResponderAddress string `json:"responder_address"`

	// IsAgent indicates if the responder is a support agent
	IsAgent bool `json:"is_agent"`

	// EncryptedPayload contains the encrypted response content
	EncryptedPayload encryptiontypes.MultiRecipientEnvelope `json:"encrypted_payload"`

	// CreatedAt is the response creation timestamp
	CreatedAt time.Time `json:"created_at"`
}

// Validate validates the ticket response
func (r *TicketResponse) Validate() error {
	if r.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket ID cannot be empty")
	}

	if r.ResponderAddress == "" {
		return ErrInvalidAddress.Wrap("responder address cannot be empty")
	}

	if err := r.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidEncryptedPayload.Wrap(err.Error())
	}

	return nil
}

// Proto message interface stubs

func (*SupportTicket) ProtoMessage()      {}
func (t *SupportTicket) Reset()           { *t = SupportTicket{} }
func (t *SupportTicket) String() string   { return fmt.Sprintf("%+v", *t) }

func (*TicketResponse) ProtoMessage()     {}
func (r *TicketResponse) Reset()          { *r = TicketResponse{} }
func (r *TicketResponse) String() string  { return fmt.Sprintf("%+v", *r) }

func (*ResourceReference) ProtoMessage()    {}
func (r *ResourceReference) Reset()         { *r = ResourceReference{} }
func (r *ResourceReference) String() string { return fmt.Sprintf("%+v", *r) }
