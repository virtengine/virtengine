// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Fraud report type definitions
// This file defines the FraudReport type with encrypted evidence,
// moderator queue routing, and comprehensive audit trail.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Description constraints
const (
	// MinDescriptionLength is the minimum description length
	MinDescriptionLength = 20

	// MaxDescriptionLength is the maximum description length
	MaxDescriptionLength = 5000

	// MaxResolutionNotesLength is the maximum resolution notes length
	MaxResolutionNotesLength = 2000
)

// FraudReportStatus represents the state of a fraud report
type FraudReportStatus uint8

const (
	// FraudReportStatusUnspecified represents an unspecified status
	FraudReportStatusUnspecified FraudReportStatus = 0

	// FraudReportStatusSubmitted indicates the report has been submitted and is awaiting review
	FraudReportStatusSubmitted FraudReportStatus = 1

	// FraudReportStatusReviewing indicates a moderator is actively reviewing the report
	FraudReportStatusReviewing FraudReportStatus = 2

	// FraudReportStatusResolved indicates the report has been resolved (action taken)
	FraudReportStatusResolved FraudReportStatus = 3

	// FraudReportStatusRejected indicates the report was rejected (no fraud found)
	FraudReportStatusRejected FraudReportStatus = 4

	// FraudReportStatusEscalated indicates the report has been escalated to admin
	FraudReportStatusEscalated FraudReportStatus = 5
)

// FraudReportStatusNames maps fraud report statuses to human-readable names
var FraudReportStatusNames = map[FraudReportStatus]string{
	FraudReportStatusUnspecified: "unspecified",
	FraudReportStatusSubmitted:   "submitted",
	FraudReportStatusReviewing:   "reviewing",
	FraudReportStatusResolved:    "resolved",
	FraudReportStatusRejected:    "rejected",
	FraudReportStatusEscalated:   "escalated",
}

// String returns the string representation of a FraudReportStatus
func (s FraudReportStatus) String() string {
	if name, ok := FraudReportStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the status is valid
func (s FraudReportStatus) IsValid() bool {
	return s >= FraudReportStatusSubmitted && s <= FraudReportStatusEscalated
}

// IsTerminal returns true if this is a terminal status (no further transitions)
func (s FraudReportStatus) IsTerminal() bool {
	return s == FraudReportStatusResolved || s == FraudReportStatusRejected
}

// IsPending returns true if the report is still pending moderator action
func (s FraudReportStatus) IsPending() bool {
	return s == FraudReportStatusSubmitted || s == FraudReportStatusReviewing || s == FraudReportStatusEscalated
}

// CanTransitionTo returns true if the status can transition to the target status
func (s FraudReportStatus) CanTransitionTo(target FraudReportStatus) bool {
	switch s {
	case FraudReportStatusSubmitted:
		return target == FraudReportStatusReviewing || target == FraudReportStatusRejected
	case FraudReportStatusReviewing:
		return target == FraudReportStatusResolved || target == FraudReportStatusRejected || target == FraudReportStatusEscalated
	case FraudReportStatusEscalated:
		return target == FraudReportStatusResolved || target == FraudReportStatusRejected
	case FraudReportStatusResolved, FraudReportStatusRejected:
		return false // Terminal states
	default:
		return false
	}
}

// FraudCategory represents the category of fraud being reported
type FraudCategory uint8

const (
	// FraudCategoryUnspecified represents an unspecified category
	FraudCategoryUnspecified FraudCategory = 0

	// FraudCategoryFakeIdentity indicates fake or stolen identity
	FraudCategoryFakeIdentity FraudCategory = 1

	// FraudCategoryPaymentFraud indicates payment-related fraud
	FraudCategoryPaymentFraud FraudCategory = 2

	// FraudCategoryServiceMisrepresentation indicates misrepresented services
	FraudCategoryServiceMisrepresentation FraudCategory = 3

	// FraudCategoryResourceAbuse indicates abuse of allocated resources
	FraudCategoryResourceAbuse FraudCategory = 4

	// FraudCategorySybilAttack indicates suspected sybil attack
	FraudCategorySybilAttack FraudCategory = 5

	// FraudCategoryMaliciousContent indicates malicious content or software
	FraudCategoryMaliciousContent FraudCategory = 6

	// FraudCategoryTermsViolation indicates terms of service violation
	FraudCategoryTermsViolation FraudCategory = 7

	// FraudCategoryOther indicates other fraud types
	FraudCategoryOther FraudCategory = 8
)

// FraudCategoryNames maps fraud categories to human-readable names
var FraudCategoryNames = map[FraudCategory]string{
	FraudCategoryUnspecified:              "unspecified",
	FraudCategoryFakeIdentity:             "fake_identity",
	FraudCategoryPaymentFraud:             "payment_fraud",
	FraudCategoryServiceMisrepresentation: "service_misrepresentation",
	FraudCategoryResourceAbuse:            "resource_abuse",
	FraudCategorySybilAttack:              "sybil_attack",
	FraudCategoryMaliciousContent:         "malicious_content",
	FraudCategoryTermsViolation:           "terms_violation",
	FraudCategoryOther:                    "other",
}

// String returns the string representation of a FraudCategory
func (c FraudCategory) String() string {
	if name, ok := FraudCategoryNames[c]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", c)
}

// IsValid returns true if the category is valid
func (c FraudCategory) IsValid() bool {
	return c >= FraudCategoryFakeIdentity && c <= FraudCategoryOther
}

// FraudCategoryFromString converts a string to a FraudCategory
func FraudCategoryFromString(s string) (FraudCategory, error) {
	for cat, name := range FraudCategoryNames {
		if name == s {
			return cat, nil
		}
	}
	return FraudCategoryUnspecified, fmt.Errorf("unknown fraud category: %s", s)
}

// ResolutionType represents the type of resolution for a fraud report
type ResolutionType uint8

const (
	// ResolutionTypeUnspecified represents no resolution yet
	ResolutionTypeUnspecified ResolutionType = 0

	// ResolutionTypeWarning indicates a warning was issued
	ResolutionTypeWarning ResolutionType = 1

	// ResolutionTypeSuspension indicates account suspension
	ResolutionTypeSuspension ResolutionType = 2

	// ResolutionTypeTermination indicates account termination
	ResolutionTypeTermination ResolutionType = 3

	// ResolutionTypeRefund indicates a refund was processed
	ResolutionTypeRefund ResolutionType = 4

	// ResolutionTypeNoAction indicates no action taken (rejected)
	ResolutionTypeNoAction ResolutionType = 5
)

// ResolutionTypeNames maps resolution types to human-readable names
var ResolutionTypeNames = map[ResolutionType]string{
	ResolutionTypeUnspecified: "unspecified",
	ResolutionTypeWarning:     "warning",
	ResolutionTypeSuspension:  "suspension",
	ResolutionTypeTermination: "termination",
	ResolutionTypeRefund:      "refund",
	ResolutionTypeNoAction:    "no_action",
}

// String returns the string representation of a ResolutionType
func (r ResolutionType) String() string {
	if name, ok := ResolutionTypeNames[r]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", r)
}

// IsValid returns true if the resolution type is valid
func (r ResolutionType) IsValid() bool {
	return r >= ResolutionTypeWarning && r <= ResolutionTypeNoAction
}

// EncryptedEvidence holds encrypted evidence for a fraud report
type EncryptedEvidence struct {
	// AlgorithmID identifies the encryption algorithm used
	AlgorithmID string `json:"algorithm_id"`

	// RecipientKeyIDs are the fingerprints of moderator public keys
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// EncryptedKeys contains the data encryption key encrypted for each recipient
	EncryptedKeys [][]byte `json:"encrypted_keys,omitempty"`

	// Nonce is the initialization vector for encryption
	Nonce []byte `json:"nonce"`

	// Ciphertext is the encrypted evidence data
	Ciphertext []byte `json:"ciphertext"`

	// SenderSignature is the signature for authenticity verification
	SenderSignature []byte `json:"sender_signature"`

	// SenderPubKey is the sender's public key
	SenderPubKey []byte `json:"sender_pub_key"`

	// ContentType indicates the type of evidence (e.g., "image/png", "application/json")
	ContentType string `json:"content_type"`

	// EvidenceHash is SHA256 hash of the original evidence for integrity verification
	EvidenceHash string `json:"evidence_hash"`
}

// Validate validates the encrypted evidence structure
func (e *EncryptedEvidence) Validate() error {
	if e == nil {
		return ErrMissingEvidence
	}
	if e.AlgorithmID == "" {
		return ErrInvalidEvidence.Wrap("algorithm ID is required")
	}
	if len(e.RecipientKeyIDs) == 0 {
		return ErrInvalidEvidence.Wrap("at least one moderator recipient required")
	}
	if len(e.Nonce) == 0 {
		return ErrInvalidEvidence.Wrap("nonce is required")
	}
	if len(e.Ciphertext) == 0 {
		return ErrInvalidEvidence.Wrap("ciphertext is required")
	}
	if len(e.SenderPubKey) == 0 {
		return ErrInvalidEvidence.Wrap("sender public key is required")
	}
	if e.EvidenceHash == "" {
		return ErrInvalidEvidence.Wrap("evidence hash is required")
	}
	return nil
}

// ComputeCiphertextHash computes a hash of the ciphertext for audit logging
func (e *EncryptedEvidence) ComputeCiphertextHash() string {
	hash := sha256.Sum256(e.Ciphertext)
	return hex.EncodeToString(hash[:])
}

// FraudReportID is the unique identifier for a fraud report
type FraudReportID struct {
	// Sequence is the globally unique sequential report number
	Sequence uint64 `json:"sequence"`
}

// String returns the string representation of the fraud report ID
func (id FraudReportID) String() string {
	return fmt.Sprintf("fraud-report-%d", id.Sequence)
}

// Validate validates the fraud report ID
func (id FraudReportID) Validate() error {
	if id.Sequence == 0 {
		return ErrInvalidReportID.Wrap("sequence must be positive")
	}
	return nil
}

// FraudReport represents a fraud report submitted by a provider
type FraudReport struct {
	// ID is the unique identifier for this report
	ID string `json:"id"`

	// Reporter is the provider address who submitted the report
	Reporter string `json:"reporter"`

	// ReportedParty is the address of the party being reported
	ReportedParty string `json:"reported_party"`

	// Category is the type of fraud being reported
	Category FraudCategory `json:"category"`

	// Description is the detailed description of the fraud
	Description string `json:"description"`

	// Evidence contains the encrypted evidence attachments
	Evidence []EncryptedEvidence `json:"evidence"`

	// Status is the current status of the report
	Status FraudReportStatus `json:"status"`

	// AssignedModerator is the moderator assigned to review this report
	AssignedModerator string `json:"assigned_moderator,omitempty"`

	// Resolution is the resolution type if resolved
	Resolution ResolutionType `json:"resolution,omitempty"`

	// ResolutionNotes are notes provided by the moderator upon resolution
	ResolutionNotes string `json:"resolution_notes,omitempty"`

	// SubmittedAt is when the report was submitted
	SubmittedAt time.Time `json:"submitted_at"`

	// UpdatedAt is when the report was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// ResolvedAt is when the report was resolved (if applicable)
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// BlockHeight is the block height when submitted
	BlockHeight int64 `json:"block_height"`

	// ContentHash is SHA256 hash of the report content for integrity
	ContentHash string `json:"content_hash"`

	// RelatedOrderIDs are order IDs related to this fraud report
	RelatedOrderIDs []string `json:"related_order_ids,omitempty"`
}

// NewFraudReport creates a new fraud report
func NewFraudReport(
	id string,
	reporter string,
	reportedParty string,
	category FraudCategory,
	description string,
	evidence []EncryptedEvidence,
	blockHeight int64,
	submittedAt time.Time,
) *FraudReport {
	report := &FraudReport{
		ID:            id,
		Reporter:      reporter,
		ReportedParty: reportedParty,
		Category:      category,
		Description:   description,
		Evidence:      evidence,
		Status:        FraudReportStatusSubmitted,
		SubmittedAt:   submittedAt,
		UpdatedAt:     submittedAt,
		BlockHeight:   blockHeight,
	}
	report.ContentHash = report.ComputeContentHash()
	return report
}

// Validate validates the fraud report
func (r *FraudReport) Validate() error {
	if r.ID == "" {
		return ErrInvalidReportID.Wrap("report ID is required")
	}
	if r.Reporter == "" {
		return ErrInvalidReporter.Wrap("reporter address is required")
	}
	if r.ReportedParty == "" {
		return ErrInvalidReportedParty.Wrap("reported party address is required")
	}
	if r.Reporter == r.ReportedParty {
		return ErrSelfReport
	}
	if !r.Category.IsValid() {
		return ErrInvalidCategory
	}
	if len(r.Description) < MinDescriptionLength {
		return ErrDescriptionTooShort.Wrapf("minimum %d characters required", MinDescriptionLength)
	}
	if len(r.Description) > MaxDescriptionLength {
		return ErrDescriptionTooLong.Wrapf("maximum %d characters allowed", MaxDescriptionLength)
	}
	if len(r.Evidence) == 0 {
		return ErrMissingEvidence
	}
	for i, ev := range r.Evidence {
		if err := ev.Validate(); err != nil {
			return ErrInvalidEvidence.Wrapf("evidence %d: %v", i, err)
		}
	}
	if r.SubmittedAt.IsZero() {
		return fmt.Errorf("submitted_at is required")
	}
	return nil
}

// ComputeContentHash computes a hash of the report content
func (r *FraudReport) ComputeContentHash() string {
	// Hash the immutable content fields
	content := fmt.Sprintf("%s|%s|%s|%d|%s|%s",
		r.ID,
		r.Reporter,
		r.ReportedParty,
		r.Category,
		r.Description,
		r.SubmittedAt.Format(time.RFC3339Nano),
	)
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// UpdateStatus updates the report status with validation
func (r *FraudReport) UpdateStatus(newStatus FraudReportStatus, updatedAt time.Time) error {
	if !r.Status.CanTransitionTo(newStatus) {
		return ErrInvalidStatus.Wrapf("cannot transition from %s to %s",
			r.Status.String(), newStatus.String())
	}
	r.Status = newStatus
	r.UpdatedAt = updatedAt
	return nil
}

// Resolve marks the report as resolved with the given resolution
func (r *FraudReport) Resolve(
	resolution ResolutionType,
	notes string,
	resolvedAt time.Time,
) error {
	if r.Status.IsTerminal() {
		return ErrReportAlreadyResolved
	}
	if !resolution.IsValid() {
		return ErrInvalidResolution
	}
	if len(notes) > MaxResolutionNotesLength {
		return ErrInvalidResolutionNotes.Wrapf("maximum %d characters allowed", MaxResolutionNotesLength)
	}

	r.Status = FraudReportStatusResolved
	r.Resolution = resolution
	r.ResolutionNotes = notes
	r.ResolvedAt = &resolvedAt
	r.UpdatedAt = resolvedAt
	return nil
}

// Reject marks the report as rejected
func (r *FraudReport) Reject(notes string, rejectedAt time.Time) error {
	if r.Status.IsTerminal() {
		return ErrReportAlreadyResolved
	}
	if len(notes) > MaxResolutionNotesLength {
		return ErrInvalidResolutionNotes.Wrapf("maximum %d characters allowed", MaxResolutionNotesLength)
	}

	r.Status = FraudReportStatusRejected
	r.Resolution = ResolutionTypeNoAction
	r.ResolutionNotes = notes
	r.ResolvedAt = &rejectedAt
	r.UpdatedAt = rejectedAt
	return nil
}

// AuditAction represents the type of action recorded in the audit log
type AuditAction uint8

const (
	// AuditActionUnspecified represents an unspecified action
	AuditActionUnspecified AuditAction = 0

	// AuditActionSubmitted indicates report was submitted
	AuditActionSubmitted AuditAction = 1

	// AuditActionAssigned indicates report was assigned to moderator
	AuditActionAssigned AuditAction = 2

	// AuditActionStatusChanged indicates status was changed
	AuditActionStatusChanged AuditAction = 3

	// AuditActionEvidenceViewed indicates evidence was viewed by moderator
	AuditActionEvidenceViewed AuditAction = 4

	// AuditActionResolved indicates report was resolved
	AuditActionResolved AuditAction = 5

	// AuditActionRejected indicates report was rejected
	AuditActionRejected AuditAction = 6

	// AuditActionEscalated indicates report was escalated
	AuditActionEscalated AuditAction = 7

	// AuditActionCommentAdded indicates a comment was added
	AuditActionCommentAdded AuditAction = 8
)

// AuditActionNames maps audit actions to human-readable names
var AuditActionNames = map[AuditAction]string{
	AuditActionUnspecified:    "unspecified",
	AuditActionSubmitted:      "submitted",
	AuditActionAssigned:       "assigned",
	AuditActionStatusChanged:  "status_changed",
	AuditActionEvidenceViewed: "evidence_viewed",
	AuditActionResolved:       "resolved",
	AuditActionRejected:       "rejected",
	AuditActionEscalated:      "escalated",
	AuditActionCommentAdded:   "comment_added",
}

// String returns the string representation of an AuditAction
func (a AuditAction) String() string {
	if name, ok := AuditActionNames[a]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", a)
}

// IsValid returns true if the action is valid
func (a AuditAction) IsValid() bool {
	return a >= AuditActionSubmitted && a <= AuditActionCommentAdded
}

// FraudAuditLog represents an audit log entry for a fraud report
type FraudAuditLog struct {
	// ID is the unique identifier for this log entry
	ID string `json:"id"`

	// ReportID is the associated fraud report ID
	ReportID string `json:"report_id"`

	// Action is the type of action performed
	Action AuditAction `json:"action"`

	// Actor is the address that performed the action
	Actor string `json:"actor"`

	// PreviousStatus is the status before the action (if applicable)
	PreviousStatus FraudReportStatus `json:"previous_status,omitempty"`

	// NewStatus is the status after the action (if applicable)
	NewStatus FraudReportStatus `json:"new_status,omitempty"`

	// Details contains additional action-specific details
	Details string `json:"details,omitempty"`

	// Timestamp is when the action was performed
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height when the action occurred
	BlockHeight int64 `json:"block_height"`

	// TxHash is the transaction hash (if applicable)
	TxHash string `json:"tx_hash,omitempty"`
}

// NewFraudAuditLog creates a new audit log entry
func NewFraudAuditLog(
	id string,
	reportID string,
	action AuditAction,
	actor string,
	previousStatus FraudReportStatus,
	newStatus FraudReportStatus,
	details string,
	timestamp time.Time,
	blockHeight int64,
) *FraudAuditLog {
	return &FraudAuditLog{
		ID:             id,
		ReportID:       reportID,
		Action:         action,
		Actor:          actor,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Details:        details,
		Timestamp:      timestamp,
		BlockHeight:    blockHeight,
	}
}

// Validate validates the audit log entry
func (l *FraudAuditLog) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("audit log ID is required")
	}
	if l.ReportID == "" {
		return fmt.Errorf("report ID is required")
	}
	if !l.Action.IsValid() {
		return fmt.Errorf("invalid audit action")
	}
	if l.Actor == "" {
		return fmt.Errorf("actor address is required")
	}
	if l.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}

// ModeratorQueueEntry represents an entry in the moderator queue
type ModeratorQueueEntry struct {
	// ReportID is the fraud report ID
	ReportID string `json:"report_id"`

	// Priority is the queue priority (higher = more urgent)
	Priority uint8 `json:"priority"`

	// QueuedAt is when the report was added to the queue
	QueuedAt time.Time `json:"queued_at"`

	// Category is the fraud category for routing
	Category FraudCategory `json:"category"`

	// AssignedTo is the moderator assigned (empty if unassigned)
	AssignedTo string `json:"assigned_to,omitempty"`
}

// NewModeratorQueueEntry creates a new queue entry
func NewModeratorQueueEntry(reportID string, category FraudCategory, priority uint8, queuedAt time.Time) *ModeratorQueueEntry {
	return &ModeratorQueueEntry{
		ReportID: reportID,
		Priority: priority,
		QueuedAt: queuedAt,
		Category: category,
	}
}
