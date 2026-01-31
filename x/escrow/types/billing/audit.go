// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"
)

// AuditActionType defines types of auditable billing actions
type AuditActionType uint8

const (
	// AuditActionReconciliationReportGenerated is when a reconciliation report is generated
	AuditActionReconciliationReportGenerated AuditActionType = 0

	// AuditActionReconciliationReportCompleted is when a reconciliation report completes
	AuditActionReconciliationReportCompleted AuditActionType = 1

	// AuditActionDisputeInitiated is when a dispute is initiated
	AuditActionDisputeInitiated AuditActionType = 2

	// AuditActionDisputeEvidenceUploaded is when evidence is uploaded for a dispute
	AuditActionDisputeEvidenceUploaded AuditActionType = 3

	// AuditActionDisputeResolved is when a dispute is resolved
	AuditActionDisputeResolved AuditActionType = 4

	// AuditActionDisputeEscalated is when a dispute is escalated
	AuditActionDisputeEscalated AuditActionType = 5

	// AuditActionDisputeExpired is when a dispute expires
	AuditActionDisputeExpired AuditActionType = 6

	// AuditActionCorrectionRequested is when a correction is requested
	AuditActionCorrectionRequested AuditActionType = 7

	// AuditActionCorrectionApproved is when a correction is approved
	AuditActionCorrectionApproved AuditActionType = 8

	// AuditActionCorrectionApplied is when a correction is applied
	AuditActionCorrectionApplied AuditActionType = 9

	// AuditActionCorrectionRejected is when a correction is rejected
	AuditActionCorrectionRejected AuditActionType = 10

	// AuditActionAlertTriggered is when an alert is triggered
	AuditActionAlertTriggered AuditActionType = 11

	// AuditActionAlertAcknowledged is when an alert is acknowledged
	AuditActionAlertAcknowledged AuditActionType = 12

	// AuditActionAlertResolved is when an alert is resolved
	AuditActionAlertResolved AuditActionType = 13

	// AuditActionExportGenerated is when an export is generated
	AuditActionExportGenerated AuditActionType = 14
)

// AuditActionTypeNames maps action types to human-readable names
var AuditActionTypeNames = map[AuditActionType]string{
	AuditActionReconciliationReportGenerated: "reconciliation_report_generated",
	AuditActionReconciliationReportCompleted: "reconciliation_report_completed",
	AuditActionDisputeInitiated:              "dispute_initiated",
	AuditActionDisputeEvidenceUploaded:       "dispute_evidence_uploaded",
	AuditActionDisputeResolved:               "dispute_resolved",
	AuditActionDisputeEscalated:              "dispute_escalated",
	AuditActionDisputeExpired:                "dispute_expired",
	AuditActionCorrectionRequested:           "correction_requested",
	AuditActionCorrectionApproved:            "correction_approved",
	AuditActionCorrectionApplied:             "correction_applied",
	AuditActionCorrectionRejected:            "correction_rejected",
	AuditActionAlertTriggered:                "alert_triggered",
	AuditActionAlertAcknowledged:             "alert_acknowledged",
	AuditActionAlertResolved:                 "alert_resolved",
	AuditActionExportGenerated:               "export_generated",
}

// String returns string representation
func (t AuditActionType) String() string {
	if name, ok := AuditActionTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid checks if the action type is valid
func (t AuditActionType) IsValid() bool {
	_, ok := AuditActionTypeNames[t]
	return ok
}

// AuditEntityType defines the type of entity being audited
type AuditEntityType string

const (
	// AuditEntityTypeReconciliation is a reconciliation report
	AuditEntityTypeReconciliation AuditEntityType = "reconciliation"

	// AuditEntityTypeDispute is a dispute
	AuditEntityTypeDispute AuditEntityType = "dispute"

	// AuditEntityTypeCorrection is a correction
	AuditEntityTypeCorrection AuditEntityType = "correction"

	// AuditEntityTypeInvoice is an invoice
	AuditEntityTypeInvoice AuditEntityType = "invoice"

	// AuditEntityTypeSettlement is a settlement
	AuditEntityTypeSettlement AuditEntityType = "settlement"
)

// AuditEntityTypes is the list of valid entity types
var AuditEntityTypes = []AuditEntityType{
	AuditEntityTypeReconciliation,
	AuditEntityTypeDispute,
	AuditEntityTypeCorrection,
	AuditEntityTypeInvoice,
	AuditEntityTypeSettlement,
}

// IsValid checks if the entity type is valid
func (t AuditEntityType) IsValid() bool {
	for _, et := range AuditEntityTypes {
		if t == et {
			return true
		}
	}
	return false
}

// AuditOutcome defines the outcome of an audited action
type AuditOutcome string

const (
	// AuditOutcomeSuccess indicates the action succeeded
	AuditOutcomeSuccess AuditOutcome = "success"

	// AuditOutcomeFailure indicates the action failed
	AuditOutcomeFailure AuditOutcome = "failure"
)

// IsValid checks if the outcome is valid
func (o AuditOutcome) IsValid() bool {
	return o == AuditOutcomeSuccess || o == AuditOutcomeFailure
}

// AuditEntry represents a single audit log entry for billing actions
type AuditEntry struct {
	// EntryID is the unique identifier for this entry
	EntryID string `json:"entry_id"`

	// Action is the type of action performed
	Action AuditActionType `json:"action"`

	// EntityType is the type of entity being audited
	EntityType AuditEntityType `json:"entity_type"`

	// EntityID is the ID of the entity being audited
	EntityID string `json:"entity_id"`

	// Actor is the address of who performed the action
	Actor string `json:"actor"`

	// ActorRole is the role of the actor (e.g., "provider", "customer", "arbitrator", "system")
	ActorRole string `json:"actor_role"`

	// Details contains action-specific details
	Details map[string]string `json:"details,omitempty"`

	// Outcome is the result of the action
	Outcome AuditOutcome `json:"outcome"`

	// ErrorMessage contains the error message if outcome is failure
	ErrorMessage string `json:"error_message,omitempty"`

	// IPAddress is the IP address of the actor (optional, for off-chain actions)
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the user agent string (optional, for off-chain actions)
	UserAgent string `json:"user_agent,omitempty"`

	// BlockHeight is the block height when this action was recorded
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the action occurred
	Timestamp time.Time `json:"timestamp"`
}

// NewAuditEntry creates a new audit entry with required fields
func NewAuditEntry(
	entryID string,
	action AuditActionType,
	entityType AuditEntityType,
	entityID string,
	actor string,
	actorRole string,
	outcome AuditOutcome,
	blockHeight int64,
	timestamp time.Time,
) *AuditEntry {
	return &AuditEntry{
		EntryID:     entryID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Actor:       actor,
		ActorRole:   actorRole,
		Outcome:     outcome,
		BlockHeight: blockHeight,
		Timestamp:   timestamp,
		Details:     make(map[string]string),
	}
}

// WithDetails adds details to the audit entry
func (e *AuditEntry) WithDetails(details map[string]string) *AuditEntry {
	e.Details = details
	return e
}

// AddDetail adds a single detail to the audit entry
func (e *AuditEntry) AddDetail(key, value string) *AuditEntry {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// WithError sets the error message for a failed action
func (e *AuditEntry) WithError(errMsg string) *AuditEntry {
	e.ErrorMessage = errMsg
	e.Outcome = AuditOutcomeFailure
	return e
}

// WithClientInfo sets the IP address and user agent
func (e *AuditEntry) WithClientInfo(ipAddress, userAgent string) *AuditEntry {
	e.IPAddress = ipAddress
	e.UserAgent = userAgent
	return e
}

// Validate validates the audit entry
func (e *AuditEntry) Validate() error {
	if e.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	if !e.Action.IsValid() {
		return fmt.Errorf("invalid action type: %d", e.Action)
	}

	if !e.EntityType.IsValid() {
		return fmt.Errorf("invalid entity type: %s", e.EntityType)
	}

	if e.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}

	if e.Actor == "" {
		return fmt.Errorf("actor is required")
	}

	if e.ActorRole == "" {
		return fmt.Errorf("actor_role is required")
	}

	if !e.Outcome.IsValid() {
		return fmt.Errorf("invalid outcome: %s", e.Outcome)
	}

	if e.BlockHeight < 0 {
		return fmt.Errorf("block_height must be non-negative")
	}

	if e.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}

// Pagination defines pagination parameters for audit log queries
type Pagination struct {
	// Offset is the number of entries to skip
	Offset uint64 `json:"offset"`

	// Limit is the maximum number of entries to return
	Limit uint64 `json:"limit"`
}

// DefaultPagination returns default pagination settings
func DefaultPagination() Pagination {
	return Pagination{
		Offset: 0,
		Limit:  100,
	}
}

// AuditLog defines the interface for audit logging
type AuditLog interface {
	// Log records an audit entry
	Log(ctx context.Context, entry *AuditEntry) error

	// GetByEntity retrieves audit entries for a specific entity
	GetByEntity(ctx context.Context, entityType AuditEntityType, entityID string) ([]*AuditEntry, error)

	// GetByActor retrieves audit entries for a specific actor
	GetByActor(ctx context.Context, actor string, pagination Pagination) ([]*AuditEntry, error)

	// GetByAction retrieves audit entries for a specific action type
	GetByAction(ctx context.Context, action AuditActionType, pagination Pagination) ([]*AuditEntry, error)

	// GetByTimeRange retrieves audit entries within a time range
	GetByTimeRange(ctx context.Context, start, end time.Time, pagination Pagination) ([]*AuditEntry, error)
}

// Store key prefixes for audit types
var (
	// AuditEntryPrefix is the prefix for audit entries
	AuditEntryPrefix = []byte{0x90}

	// AuditByEntityPrefix indexes audit entries by entity
	AuditByEntityPrefix = []byte{0x91}

	// AuditByActorPrefix indexes audit entries by actor
	AuditByActorPrefix = []byte{0x92}

	// AuditByActionPrefix indexes audit entries by action type
	AuditByActionPrefix = []byte{0x93}
)

// BuildAuditEntryKey builds the key for an audit entry
func BuildAuditEntryKey(entryID string) []byte {
	return append(AuditEntryPrefix, []byte(entryID)...)
}

// ParseAuditEntryKey parses an audit entry key
func ParseAuditEntryKey(key []byte) (string, error) {
	if len(key) <= len(AuditEntryPrefix) {
		return "", fmt.Errorf("invalid audit entry key length")
	}
	return string(key[len(AuditEntryPrefix):]), nil
}

// BuildAuditByEntityKey builds the index key for audit entries by entity
func BuildAuditByEntityKey(entityType AuditEntityType, entityID string, entryID string) []byte {
	key := append(AuditByEntityPrefix, []byte(entityType)...)
	key = append(key, byte('/'))
	key = append(key, []byte(entityID)...)
	key = append(key, byte('/'))
	return append(key, []byte(entryID)...)
}

// BuildAuditByEntityPrefix builds the prefix for an entity's audit entries
func BuildAuditByEntityPrefix(entityType AuditEntityType, entityID string) []byte {
	key := append(AuditByEntityPrefix, []byte(entityType)...)
	key = append(key, byte('/'))
	key = append(key, []byte(entityID)...)
	return append(key, byte('/'))
}

// BuildAuditByActorKey builds the index key for audit entries by actor
func BuildAuditByActorKey(actor string, timestamp int64, entryID string) []byte {
	key := append(AuditByActorPrefix, []byte(actor)...)
	key = append(key, byte('/'))
	// Append timestamp as big-endian uint64 for proper ordering
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
	key = append(key, tsBytes...)
	key = append(key, byte('/'))
	return append(key, []byte(entryID)...)
}

// BuildAuditByActorPrefix builds the prefix for an actor's audit entries
func BuildAuditByActorPrefix(actor string) []byte {
	key := append(AuditByActorPrefix, []byte(actor)...)
	return append(key, byte('/'))
}

// BuildAuditByActionKey builds the index key for audit entries by action type
func BuildAuditByActionKey(action AuditActionType, timestamp int64, entryID string) []byte {
	key := append(AuditByActionPrefix, byte(action))
	key = append(key, byte('/'))
	// Append timestamp as big-endian uint64 for proper ordering
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
	key = append(key, tsBytes...)
	key = append(key, byte('/'))
	return append(key, []byte(entryID)...)
}

// BuildAuditByActionPrefix builds the prefix for action type's audit entries
func BuildAuditByActionPrefix(action AuditActionType) []byte {
	key := append(AuditByActionPrefix, byte(action))
	return append(key, byte('/'))
}

// AuditActorRoles defines common actor roles
const (
	AuditActorRoleProvider   = "provider"
	AuditActorRoleCustomer   = "customer"
	AuditActorRoleArbitrator = "arbitrator"
	AuditActorRoleSystem     = "system"
	AuditActorRoleAdmin      = "admin"
)

// ValidAuditActorRoles lists all valid actor roles
var ValidAuditActorRoles = []string{
	AuditActorRoleProvider,
	AuditActorRoleCustomer,
	AuditActorRoleArbitrator,
	AuditActorRoleSystem,
	AuditActorRoleAdmin,
}

// IsValidActorRole checks if the given role is valid
func IsValidActorRole(role string) bool {
	for _, r := range ValidAuditActorRoles {
		if r == role {
			return true
		}
	}
	return false
}

// NewReconciliationAuditEntry creates an audit entry for reconciliation actions
func NewReconciliationAuditEntry(
	entryID string,
	action AuditActionType,
	reportID string,
	actor string,
	actorRole string,
	outcome AuditOutcome,
	blockHeight int64,
	timestamp time.Time,
) *AuditEntry {
	return NewAuditEntry(
		entryID,
		action,
		AuditEntityTypeReconciliation,
		reportID,
		actor,
		actorRole,
		outcome,
		blockHeight,
		timestamp,
	)
}

// NewDisputeAuditEntry creates an audit entry for dispute actions
func NewDisputeAuditEntry(
	entryID string,
	action AuditActionType,
	disputeID string,
	actor string,
	actorRole string,
	outcome AuditOutcome,
	blockHeight int64,
	timestamp time.Time,
) *AuditEntry {
	return NewAuditEntry(
		entryID,
		action,
		AuditEntityTypeDispute,
		disputeID,
		actor,
		actorRole,
		outcome,
		blockHeight,
		timestamp,
	)
}

// NewCorrectionAuditEntry creates an audit entry for correction actions
func NewCorrectionAuditEntry(
	entryID string,
	action AuditActionType,
	correctionID string,
	actor string,
	actorRole string,
	outcome AuditOutcome,
	blockHeight int64,
	timestamp time.Time,
) *AuditEntry {
	return NewAuditEntry(
		entryID,
		action,
		AuditEntityTypeCorrection,
		correctionID,
		actor,
		actorRole,
		outcome,
		blockHeight,
		timestamp,
	)
}

// NewInvoiceAuditEntry creates an audit entry for invoice actions
func NewInvoiceAuditEntry(
	entryID string,
	action AuditActionType,
	invoiceID string,
	actor string,
	actorRole string,
	outcome AuditOutcome,
	blockHeight int64,
	timestamp time.Time,
) *AuditEntry {
	return NewAuditEntry(
		entryID,
		action,
		AuditEntityTypeInvoice,
		invoiceID,
		actor,
		actorRole,
		outcome,
		blockHeight,
		timestamp,
	)
}

// NewSettlementAuditEntry creates an audit entry for settlement actions
func NewSettlementAuditEntry(
	entryID string,
	action AuditActionType,
	settlementID string,
	actor string,
	actorRole string,
	outcome AuditOutcome,
	blockHeight int64,
	timestamp time.Time,
) *AuditEntry {
	return NewAuditEntry(
		entryID,
		action,
		AuditEntityTypeSettlement,
		settlementID,
		actor,
		actorRole,
		outcome,
		blockHeight,
		timestamp,
	)
}
