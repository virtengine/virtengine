package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SupportCategory identifies support request categories.
type SupportCategory string

const (
	SupportCategoryAccount     SupportCategory = "account"
	SupportCategoryIdentity    SupportCategory = "identity"
	SupportCategoryBilling     SupportCategory = "billing"
	SupportCategoryProvider    SupportCategory = "provider"
	SupportCategoryMarketplace SupportCategory = "marketplace"
	SupportCategoryTechnical   SupportCategory = "technical"
	SupportCategorySecurity    SupportCategory = "security"
	SupportCategoryOther       SupportCategory = "other"
)

// IsValid checks if the category is valid.
func (c SupportCategory) IsValid() bool {
	switch c {
	case SupportCategoryAccount,
		SupportCategoryIdentity,
		SupportCategoryBilling,
		SupportCategoryProvider,
		SupportCategoryMarketplace,
		SupportCategoryTechnical,
		SupportCategorySecurity,
		SupportCategoryOther:
		return true
	default:
		return false
	}
}

// SupportPriority identifies support request priority.
type SupportPriority string

const (
	SupportPriorityLow    SupportPriority = "low"
	SupportPriorityNormal SupportPriority = "normal"
	SupportPriorityHigh   SupportPriority = "high"
	SupportPriorityUrgent SupportPriority = "urgent"
)

// IsValid checks if the priority is valid.
func (p SupportPriority) IsValid() bool {
	switch p {
	case SupportPriorityLow,
		SupportPriorityNormal,
		SupportPriorityHigh,
		SupportPriorityUrgent:
		return true
	default:
		return false
	}
}

// SupportStatus represents the lifecycle status of a support request.
type SupportStatus uint8

const (
	SupportStatusUnspecified     SupportStatus = 0
	SupportStatusOpen            SupportStatus = 1
	SupportStatusAssigned        SupportStatus = 2
	SupportStatusInProgress      SupportStatus = 3
	SupportStatusWaitingCustomer SupportStatus = 4
	SupportStatusWaitingSupport  SupportStatus = 5
	SupportStatusResolved        SupportStatus = 6
	SupportStatusClosed          SupportStatus = 7
	SupportStatusArchived        SupportStatus = 8
)

// SupportStatusNames maps status to string.
var SupportStatusNames = map[SupportStatus]string{
	SupportStatusUnspecified:     "unspecified",
	SupportStatusOpen:            "open",
	SupportStatusAssigned:        "assigned",
	SupportStatusInProgress:      "in_progress",
	SupportStatusWaitingCustomer: "waiting_customer",
	SupportStatusWaitingSupport:  "waiting_support",
	SupportStatusResolved:        "resolved",
	SupportStatusClosed:          "closed",
	SupportStatusArchived:        "archived",
}

// String returns the string representation.
func (s SupportStatus) String() string {
	if name, ok := SupportStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// SupportStatusFromString parses a status string.
func SupportStatusFromString(value string) SupportStatus {
	for status, name := range SupportStatusNames {
		if name == value {
			return status
		}
	}
	return SupportStatusUnspecified
}

// IsValid checks if the status is valid.
func (s SupportStatus) IsValid() bool {
	return s >= SupportStatusOpen && s <= SupportStatusArchived
}

// IsTerminal checks if the status is terminal.
func (s SupportStatus) IsTerminal() bool {
	return s == SupportStatusClosed || s == SupportStatusArchived
}

// CanTransitionTo checks if a status transition is allowed.
func (s SupportStatus) CanTransitionTo(next SupportStatus) bool {
	transitions := map[SupportStatus][]SupportStatus{
		SupportStatusOpen:            {SupportStatusAssigned, SupportStatusInProgress, SupportStatusWaitingSupport, SupportStatusWaitingCustomer, SupportStatusClosed},
		SupportStatusAssigned:        {SupportStatusInProgress, SupportStatusWaitingCustomer, SupportStatusWaitingSupport, SupportStatusResolved, SupportStatusClosed},
		SupportStatusInProgress:      {SupportStatusWaitingCustomer, SupportStatusWaitingSupport, SupportStatusResolved, SupportStatusClosed},
		SupportStatusWaitingCustomer: {SupportStatusInProgress, SupportStatusResolved, SupportStatusClosed},
		SupportStatusWaitingSupport:  {SupportStatusInProgress, SupportStatusResolved, SupportStatusClosed},
		SupportStatusResolved:        {SupportStatusInProgress, SupportStatusClosed},
		SupportStatusClosed:          {SupportStatusArchived},
	}
	allowed, ok := transitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

// SupportRequestID uniquely identifies a support request.
type SupportRequestID struct {
	SubmitterAddress string `json:"submitter_address"`
	Sequence         uint64 `json:"sequence"`
}

// String returns string representation.
func (id SupportRequestID) String() string {
	return fmt.Sprintf("%s/support/%d", id.SubmitterAddress, id.Sequence)
}

// Validate validates the ID.
func (id SupportRequestID) Validate() error {
	if id.SubmitterAddress == "" {
		return fmt.Errorf("submitter address is required")
	}
	if id.Sequence == 0 {
		return fmt.Errorf("sequence must be positive")
	}
	return nil
}

// ParseSupportRequestID parses a support request ID string.
func ParseSupportRequestID(value string) (SupportRequestID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 3 {
		return SupportRequestID{}, fmt.Errorf("invalid support request id: %s", value)
	}
	if parts[1] != "support" {
		return SupportRequestID{}, fmt.Errorf("invalid support request id segment: %s", value)
	}
	seq, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return SupportRequestID{}, fmt.Errorf("invalid support request sequence: %w", err)
	}
	return SupportRequestID{
		SubmitterAddress: parts[0],
		Sequence:         seq,
	}, nil
}

// RelatedEntity links a support request to an on-chain resource.
type RelatedEntity struct {
	Type ResourceType `json:"type"`
	ID   string       `json:"id"`
}

// Validate validates the related entity.
func (r *RelatedEntity) Validate() error {
	if r == nil {
		return nil
	}
	if r.ID == "" {
		return ErrInvalidResourceRef.Wrap("related entity id is required")
	}
	if !r.Type.IsValid() {
		return ErrInvalidResourceRef.Wrapf("invalid related entity type: %s", r.Type)
	}
	return nil
}

// SupportRequest represents an on-chain support request.
type SupportRequest struct {
	ID               SupportRequestID        `json:"id"`
	TicketNumber     string                  `json:"ticket_number"`
	SubmitterAddress string                  `json:"submitter_address"`
	Category         SupportCategory         `json:"category"`
	Priority         SupportPriority         `json:"priority"`
	Status           SupportStatus           `json:"status"`
	Payload          EncryptedSupportPayload `json:"payload"`
	PublicMetadata   map[string]string       `json:"public_metadata,omitempty"`
	RelatedEntity    *RelatedEntity          `json:"related_entity,omitempty"`

	Recipients    []string   `json:"recipients,omitempty"`
	AssignedAgent string     `json:"assigned_agent,omitempty"`
	AssignedAt    *time.Time `json:"assigned_at,omitempty"`

	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastResponseAt *time.Time `json:"last_response_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`

	RetentionPolicy *RetentionPolicy `json:"retention_policy,omitempty"`
	Archived        bool             `json:"archived"`
	ArchivedAt      *time.Time       `json:"archived_at,omitempty"`
	ArchiveReason   string           `json:"archive_reason,omitempty"`
	Purged          bool             `json:"purged"`
	PurgedAt        *time.Time       `json:"purged_at,omitempty"`
	PurgeReason     string           `json:"purge_reason,omitempty"`
}

// NewSupportRequest creates a new support request.
func NewSupportRequest(
	id SupportRequestID,
	ticketNumber string,
	submitterAddress string,
	category SupportCategory,
	priority SupportPriority,
	payload EncryptedSupportPayload,
	now time.Time,
) *SupportRequest {
	created := now.UTC()
	return &SupportRequest{
		ID:               id,
		TicketNumber:     ticketNumber,
		SubmitterAddress: submitterAddress,
		Category:         category,
		Priority:         priority,
		Status:           SupportStatusOpen,
		Payload:          payload,
		PublicMetadata:   map[string]string{},
		CreatedAt:        created,
		UpdatedAt:        created,
	}
}

// Validate validates the support request.
func (r *SupportRequest) Validate() error {
	if r == nil {
		return ErrInvalidSupportRequest.Wrap("request is nil")
	}
	if err := r.ID.Validate(); err != nil {
		return ErrInvalidSupportRequest.Wrapf("invalid id: %v", err)
	}
	if r.SubmitterAddress == "" {
		return ErrInvalidAddress.Wrap("submitter address is required")
	}
	if r.SubmitterAddress != r.ID.SubmitterAddress {
		return ErrInvalidSupportRequest.Wrap("submitter address mismatch")
	}
	if !r.Category.IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid category: %s", r.Category)
	}
	if !r.Priority.IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid priority: %s", r.Priority)
	}
	if !r.Status.IsValid() {
		return ErrInvalidSupportRequest.Wrapf("invalid status: %s", r.Status)
	}
	if err := r.Payload.Validate(); err != nil {
		if !r.Purged || r.Payload.Envelope != nil {
			return err
		}
	}
	if err := r.RelatedEntity.Validate(); err != nil {
		return err
	}
	if err := r.RetentionPolicy.Validate(); err != nil {
		return err
	}
	if r.Archived && r.Status != SupportStatusArchived {
		return ErrInvalidSupportRequest.Wrap("archived request must have archived status")
	}
	return nil
}

// SetStatus transitions the request to a new status.
func (r *SupportRequest) SetStatus(next SupportStatus, now time.Time) error {
	if !r.Status.CanTransitionTo(next) {
		return ErrInvalidStatusTransition.Wrapf("invalid transition: %s -> %s", r.Status, next)
	}
	r.Status = next
	t := now.UTC()
	r.UpdatedAt = t
	switch next {
	case SupportStatusResolved:
		r.ResolvedAt = &t
	case SupportStatusClosed:
		r.ClosedAt = &t
	case SupportStatusArchived:
		r.Archived = true
		r.ArchivedAt = &t
	}
	return nil
}

// MarkArchived marks the request as archived.
func (r *SupportRequest) MarkArchived(reason string, now time.Time) {
	r.Archived = true
	r.ArchiveReason = reason
	t := now.UTC()
	r.ArchivedAt = &t
	r.Status = SupportStatusArchived
	r.UpdatedAt = t
}

// MarkPurged marks the request as purged.
func (r *SupportRequest) MarkPurged(reason string, now time.Time) {
	r.Purged = true
	r.PurgeReason = reason
	t := now.UTC()
	r.PurgedAt = &t
	r.UpdatedAt = t
}
