// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EvidenceType defines the type of dispute evidence
type EvidenceType uint8

const (
	// DocumentEvidence is documentary evidence (contracts, agreements)
	DocumentEvidence EvidenceType = 0

	// UsageLogEvidence is usage log evidence
	UsageLogEvidence EvidenceType = 1

	// CommunicationEvidence is communication records (emails, messages)
	CommunicationEvidence EvidenceType = 2

	// ScreenshotEvidence is screenshot evidence
	ScreenshotEvidence EvidenceType = 3

	// ContractEvidence is contract-specific evidence
	ContractEvidence EvidenceType = 4
)

// EvidenceTypeNames maps evidence types to human-readable names
var EvidenceTypeNames = map[EvidenceType]string{
	DocumentEvidence:      "document",
	UsageLogEvidence:      "usage_log",
	CommunicationEvidence: "communication",
	ScreenshotEvidence:    "screenshot",
	ContractEvidence:      "contract",
}

// String returns string representation
func (t EvidenceType) String() string {
	if name, ok := EvidenceTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid returns true if the evidence type is valid
func (t EvidenceType) IsValid() bool {
	_, ok := EvidenceTypeNames[t]
	return ok
}

// DisputeEvidence represents evidence submitted for a dispute
type DisputeEvidence struct {
	// EvidenceID is the unique identifier for this evidence
	EvidenceID string `json:"evidence_id"`

	// DisputeID is the dispute this evidence belongs to
	DisputeID string `json:"dispute_id"`

	// Type is the type of evidence
	Type EvidenceType `json:"type"`

	// Description describes the evidence
	Description string `json:"description"`

	// ArtifactCID is the content-addressable identifier for the evidence file
	ArtifactCID string `json:"artifact_cid"`

	// UploadedBy is the address that uploaded the evidence
	UploadedBy string `json:"uploaded_by"`

	// UploadedAt is when the evidence was uploaded
	UploadedAt time.Time `json:"uploaded_at"`

	// FileSize is the size of the evidence file in bytes
	FileSize int64 `json:"file_size"`

	// ContentType is the MIME type of the evidence file
	ContentType string `json:"content_type"`

	// IsVerified indicates if the evidence has been verified
	IsVerified bool `json:"is_verified"`

	// VerifiedBy is the address that verified the evidence (optional)
	VerifiedBy string `json:"verified_by,omitempty"`

	// VerifiedAt is when the evidence was verified (optional)
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// NewDisputeEvidence creates a new dispute evidence
func NewDisputeEvidence(
	evidenceID string,
	disputeID string,
	evidenceType EvidenceType,
	description string,
	artifactCID string,
	uploadedBy string,
	fileSize int64,
	contentType string,
	now time.Time,
) *DisputeEvidence {
	return &DisputeEvidence{
		EvidenceID:  evidenceID,
		DisputeID:   disputeID,
		Type:        evidenceType,
		Description: description,
		ArtifactCID: artifactCID,
		UploadedBy:  uploadedBy,
		UploadedAt:  now,
		FileSize:    fileSize,
		ContentType: contentType,
		IsVerified:  false,
	}
}

// Validate validates the dispute evidence
func (e *DisputeEvidence) Validate() error {
	if e.EvidenceID == "" {
		return fmt.Errorf("evidence_id is required")
	}

	if len(e.EvidenceID) > 64 {
		return fmt.Errorf("evidence_id exceeds maximum length of 64")
	}

	if e.DisputeID == "" {
		return fmt.Errorf("dispute_id is required")
	}

	if !e.Type.IsValid() {
		return fmt.Errorf("invalid evidence type: %d", e.Type)
	}

	if e.Description == "" {
		return fmt.Errorf("description is required")
	}

	if len(e.Description) > 1024 {
		return fmt.Errorf("description exceeds maximum length of 1024")
	}

	if e.ArtifactCID == "" {
		return fmt.Errorf("artifact_cid is required")
	}

	if len(e.ArtifactCID) > 128 {
		return fmt.Errorf("artifact_cid exceeds maximum length of 128")
	}

	if _, err := sdk.AccAddressFromBech32(e.UploadedBy); err != nil {
		return fmt.Errorf("invalid uploaded_by address: %w", err)
	}

	if e.UploadedAt.IsZero() {
		return fmt.Errorf("uploaded_at is required")
	}

	if e.FileSize <= 0 {
		return fmt.Errorf("file_size must be positive")
	}

	if e.ContentType == "" {
		return fmt.Errorf("content_type is required")
	}

	if e.VerifiedBy != "" {
		if _, err := sdk.AccAddressFromBech32(e.VerifiedBy); err != nil {
			return fmt.Errorf("invalid verified_by address: %w", err)
		}
	}

	return nil
}

// Verify marks the evidence as verified
func (e *DisputeEvidence) Verify(verifiedBy string, now time.Time) error {
	if e.IsVerified {
		return fmt.Errorf("evidence already verified")
	}

	if _, err := sdk.AccAddressFromBech32(verifiedBy); err != nil {
		return fmt.Errorf("invalid verified_by address: %w", err)
	}

	e.IsVerified = true
	e.VerifiedBy = verifiedBy
	e.VerifiedAt = &now
	return nil
}

// DisputeCategory defines the category of a dispute
type DisputeCategory uint8

const (
	// DisputeCategoryBillingError is for billing calculation errors
	DisputeCategoryBillingError DisputeCategory = 0

	// DisputeCategoryUsageMismatch is for usage record discrepancies
	DisputeCategoryUsageMismatch DisputeCategory = 1

	// DisputeCategoryServiceNotDelivered is for services not delivered
	DisputeCategoryServiceNotDelivered DisputeCategory = 2

	// DisputeCategoryQualityIssue is for service quality issues
	DisputeCategoryQualityIssue DisputeCategory = 3

	// DisputeCategoryContractViolation is for contract violations
	DisputeCategoryContractViolation DisputeCategory = 4

	// DisputeCategoryDuplicateCharge is for duplicate charges
	DisputeCategoryDuplicateCharge DisputeCategory = 5

	// DisputeCategoryUnauthorizedCharge is for unauthorized charges
	DisputeCategoryUnauthorizedCharge DisputeCategory = 6

	// DisputeCategoryRefundRequest is for refund requests
	DisputeCategoryRefundRequest DisputeCategory = 7
)

// DisputeCategoryNames maps categories to human-readable names
var DisputeCategoryNames = map[DisputeCategory]string{
	DisputeCategoryBillingError:        "billing_error",
	DisputeCategoryUsageMismatch:       "usage_mismatch",
	DisputeCategoryServiceNotDelivered: "service_not_delivered",
	DisputeCategoryQualityIssue:        "quality_issue",
	DisputeCategoryContractViolation:   "contract_violation",
	DisputeCategoryDuplicateCharge:     "duplicate_charge",
	DisputeCategoryUnauthorizedCharge:  "unauthorized_charge",
	DisputeCategoryRefundRequest:       "refund_request",
}

// String returns string representation
func (c DisputeCategory) String() string {
	if name, ok := DisputeCategoryNames[c]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", c)
}

// IsValid returns true if the dispute category is valid
func (c DisputeCategory) IsValid() bool {
	_, ok := DisputeCategoryNames[c]
	return ok
}

// DisputeAuditEntry represents an audit trail entry for a dispute
type DisputeAuditEntry struct {
	// Action is the action performed
	Action string `json:"action"`

	// PerformedBy is the address that performed the action
	PerformedBy string `json:"performed_by"`

	// Details contains additional action details
	Details string `json:"details"`

	// Timestamp is when the action was performed
	Timestamp time.Time `json:"timestamp"`
}

// newDisputeWorkflowAuditEntry creates a new dispute workflow audit entry
func newDisputeWorkflowAuditEntry(action, performedBy, details string, timestamp time.Time) DisputeAuditEntry {
	return DisputeAuditEntry{
		Action:      action,
		PerformedBy: performedBy,
		Details:     details,
		Timestamp:   timestamp,
	}
}

// Validate validates the dispute audit entry
func (e *DisputeAuditEntry) Validate() error {
	if e.Action == "" {
		return fmt.Errorf("action is required")
	}

	if len(e.Action) > 128 {
		return fmt.Errorf("action exceeds maximum length of 128")
	}

	if e.PerformedBy == "" {
		return fmt.Errorf("performed_by is required")
	}

	if e.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	return nil
}

// DisputeWorkflow manages the complete lifecycle of a dispute
type DisputeWorkflow struct {
	// DisputeID is the unique identifier for this dispute
	DisputeID string `json:"dispute_id"`

	// InvoiceID is the invoice being disputed
	InvoiceID string `json:"invoice_id"`

	// InitiatedBy is the address that initiated the dispute
	InitiatedBy string `json:"initiated_by"`

	// Category is the category of the dispute
	Category DisputeCategory `json:"category"`

	// Subject is a brief subject line for the dispute
	Subject string `json:"subject"`

	// Description is a detailed description of the dispute
	Description string `json:"description"`

	// DisputedAmount is the amount being disputed
	DisputedAmount sdk.Coins `json:"disputed_amount"`

	// Evidence is the list of evidence submitted for this dispute
	Evidence []DisputeEvidence `json:"evidence"`

	// Window is the dispute window configuration
	Window *DisputeWindow `json:"window,omitempty"`

	// CorrectionLimit defines correction limits for this dispute
	CorrectionLimit *CorrectionLimit `json:"correction_limit,omitempty"`

	// Status is the current status of the dispute
	Status DisputeStatus `json:"status"`

	// Resolution is the resolution type once resolved
	Resolution DisputeResolutionType `json:"resolution"`

	// ResolutionDetails contains details about the resolution
	ResolutionDetails string `json:"resolution_details,omitempty"`

	// ResolvedBy is the address that resolved the dispute
	ResolvedBy string `json:"resolved_by,omitempty"`

	// RefundAmount is any refund amount agreed upon
	RefundAmount sdk.Coins `json:"refund_amount,omitempty"`

	// AuditTrail is the audit trail of actions on this dispute
	AuditTrail []DisputeAuditEntry `json:"audit_trail"`

	// CreatedAt is when the dispute was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the dispute was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// ResolvedAt is when the dispute was resolved (optional)
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// BlockHeight is the block height when the dispute was created
	BlockHeight int64 `json:"block_height"`
}

// NewDisputeWorkflow creates a new dispute workflow
func NewDisputeWorkflow(
	disputeID string,
	invoiceID string,
	initiatedBy string,
	category DisputeCategory,
	subject string,
	description string,
	disputedAmount sdk.Coins,
	window *DisputeWindow,
	correctionLimit *CorrectionLimit,
	blockHeight int64,
	now time.Time,
) *DisputeWorkflow {
	return &DisputeWorkflow{
		DisputeID:       disputeID,
		InvoiceID:       invoiceID,
		InitiatedBy:     initiatedBy,
		Category:        category,
		Subject:         subject,
		Description:     description,
		DisputedAmount:  disputedAmount,
		Evidence:        make([]DisputeEvidence, 0),
		Window:          window,
		CorrectionLimit: correctionLimit,
		Status:          DisputeStatusOpen,
		Resolution:      DisputeResolutionNone,
		AuditTrail: []DisputeAuditEntry{
			newDisputeWorkflowAuditEntry("dispute_initiated", initiatedBy, "Dispute created", now),
		},
		CreatedAt:   now,
		UpdatedAt:   now,
		BlockHeight: blockHeight,
	}
}

// Validate validates the dispute workflow
func (w *DisputeWorkflow) Validate() error {
	if w.DisputeID == "" {
		return fmt.Errorf("dispute_id is required")
	}

	if len(w.DisputeID) > 64 {
		return fmt.Errorf("dispute_id exceeds maximum length of 64")
	}

	if w.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(w.InitiatedBy); err != nil {
		return fmt.Errorf("invalid initiated_by address: %w", err)
	}

	if !w.Category.IsValid() {
		return fmt.Errorf("invalid dispute category: %d", w.Category)
	}

	if w.Subject == "" {
		return fmt.Errorf("subject is required")
	}

	if len(w.Subject) > 256 {
		return fmt.Errorf("subject exceeds maximum length of 256")
	}

	if w.Description == "" {
		return fmt.Errorf("description is required")
	}

	if len(w.Description) > 4096 {
		return fmt.Errorf("description exceeds maximum length of 4096")
	}

	if !w.DisputedAmount.IsValid() {
		return fmt.Errorf("disputed_amount must be valid coins")
	}

	if w.DisputedAmount.IsZero() {
		return fmt.Errorf("disputed_amount cannot be zero")
	}

	// Validate evidence
	for i, ev := range w.Evidence {
		if err := ev.Validate(); err != nil {
			return fmt.Errorf("evidence[%d]: %w", i, err)
		}
	}

	// Validate window if present
	if w.Window != nil {
		if err := w.Window.Validate(); err != nil {
			return fmt.Errorf("window: %w", err)
		}
	}

	// Validate correction limit if present
	if w.CorrectionLimit != nil {
		if err := w.CorrectionLimit.Validate(); err != nil {
			return fmt.Errorf("correction_limit: %w", err)
		}
	}

	// Validate audit trail
	for i, entry := range w.AuditTrail {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("audit_trail[%d]: %w", i, err)
		}
	}

	if w.ResolvedBy != "" {
		if _, err := sdk.AccAddressFromBech32(w.ResolvedBy); err != nil {
			return fmt.Errorf("invalid resolved_by address: %w", err)
		}
	}

	if w.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	if w.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}

	if w.BlockHeight < 0 {
		return fmt.Errorf("block_height must be non-negative")
	}

	return nil
}

// AddEvidence adds evidence to the dispute
func (w *DisputeWorkflow) AddEvidence(evidence DisputeEvidence, now time.Time) error {
	if w.Status == DisputeStatusResolved || w.Status == DisputeStatusClosed {
		return fmt.Errorf("cannot add evidence to resolved or closed dispute")
	}

	if err := evidence.Validate(); err != nil {
		return fmt.Errorf("invalid evidence: %w", err)
	}

	w.Evidence = append(w.Evidence, evidence)
	w.UpdatedAt = now
	w.AuditTrail = append(w.AuditTrail, newDisputeWorkflowAuditEntry(
		"evidence_uploaded",
		evidence.UploadedBy,
		fmt.Sprintf("Evidence %s uploaded: %s", evidence.EvidenceID, evidence.Type.String()),
		now,
	))
	return nil
}

// SubmitForReview moves the dispute to under review status
func (w *DisputeWorkflow) SubmitForReview(submittedBy string, now time.Time) error {
	if w.Status != DisputeStatusOpen {
		return fmt.Errorf("can only submit open disputes for review, current status: %s", w.Status)
	}

	w.Status = DisputeStatusUnderReview
	w.UpdatedAt = now
	w.AuditTrail = append(w.AuditTrail, newDisputeWorkflowAuditEntry(
		"submitted_for_review",
		submittedBy,
		"Dispute submitted for review",
		now,
	))
	return nil
}

// Resolve resolves the dispute
func (w *DisputeWorkflow) Resolve(
	resolution DisputeResolutionType,
	details string,
	resolvedBy string,
	refundAmount sdk.Coins,
	now time.Time,
) error {
	if w.Status == DisputeStatusResolved || w.Status == DisputeStatusClosed {
		return fmt.Errorf("dispute already resolved or closed")
	}

	w.Status = DisputeStatusResolved
	w.Resolution = resolution
	w.ResolutionDetails = details
	w.ResolvedBy = resolvedBy
	w.RefundAmount = refundAmount
	w.ResolvedAt = &now
	w.UpdatedAt = now
	w.AuditTrail = append(w.AuditTrail, newDisputeWorkflowAuditEntry(
		"dispute_resolved",
		resolvedBy,
		fmt.Sprintf("Resolved with %s: %s", resolution.String(), details),
		now,
	))
	return nil
}

// Escalate escalates the dispute
func (w *DisputeWorkflow) Escalate(escalateTo, reason, escalatedBy string, now time.Time) error {
	if w.Status == DisputeStatusResolved || w.Status == DisputeStatusClosed {
		return fmt.Errorf("cannot escalate resolved or closed dispute")
	}

	w.Status = DisputeStatusEscalated
	w.UpdatedAt = now
	w.AuditTrail = append(w.AuditTrail, newDisputeWorkflowAuditEntry(
		"dispute_escalated",
		escalatedBy,
		fmt.Sprintf("Escalated to %s: %s", escalateTo, reason),
		now,
	))
	return nil
}

// Close closes the dispute
func (w *DisputeWorkflow) Close(closedBy, reason string, now time.Time) error {
	if w.Status == DisputeStatusClosed {
		return fmt.Errorf("dispute already closed")
	}

	w.Status = DisputeStatusClosed
	w.UpdatedAt = now
	w.AuditTrail = append(w.AuditTrail, newDisputeWorkflowAuditEntry(
		"dispute_closed",
		closedBy,
		reason,
		now,
	))
	return nil
}

// Expire expires the dispute
func (w *DisputeWorkflow) Expire(now time.Time) error {
	if w.Status == DisputeStatusResolved || w.Status == DisputeStatusClosed {
		return fmt.Errorf("cannot expire resolved or closed dispute")
	}

	w.Status = DisputeStatusExpired
	w.UpdatedAt = now
	w.AuditTrail = append(w.AuditTrail, newDisputeWorkflowAuditEntry(
		"dispute_expired",
		"system",
		"Dispute expired without resolution",
		now,
	))
	return nil
}

// GetEvidenceByID returns evidence by ID
func (w *DisputeWorkflow) GetEvidenceByID(evidenceID string) (*DisputeEvidence, bool) {
	for i := range w.Evidence {
		if w.Evidence[i].EvidenceID == evidenceID {
			return &w.Evidence[i], true
		}
	}
	return nil, false
}

// GetVerifiedEvidenceCount returns the count of verified evidence
func (w *DisputeWorkflow) GetVerifiedEvidenceCount() int {
	count := 0
	for _, ev := range w.Evidence {
		if ev.IsVerified {
			count++
		}
	}
	return count
}

// DisputeRules defines rules and constraints for disputes
type DisputeRules struct {
	// MinDisputeAmount is the minimum amount that can be disputed
	MinDisputeAmount sdk.Coins `json:"min_dispute_amount"`

	// MaxEvidenceCount is the maximum number of evidence items per dispute
	MaxEvidenceCount uint32 `json:"max_evidence_count"`

	// MaxEvidenceSize is the maximum size of a single evidence file in bytes
	MaxEvidenceSize int64 `json:"max_evidence_size"`

	// AllowedEvidenceTypes is the list of allowed evidence types
	AllowedEvidenceTypes []EvidenceType `json:"allowed_evidence_types"`

	// RequireEvidenceForAmount is the amount threshold above which evidence is required
	RequireEvidenceForAmount sdk.Coins `json:"require_evidence_for_amount"`

	// AutoEscalateAfterDays is the number of days after which unresolved disputes auto-escalate
	AutoEscalateAfterDays uint32 `json:"auto_escalate_after_days"`
}

// DefaultDisputeRules returns default dispute rules
func DefaultDisputeRules() DisputeRules {
	return DisputeRules{
		MinDisputeAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000000)), // 1 VIRT
		MaxEvidenceCount: 10,
		MaxEvidenceSize:  10485760, // 10 MB
		AllowedEvidenceTypes: []EvidenceType{
			DocumentEvidence,
			UsageLogEvidence,
			CommunicationEvidence,
			ScreenshotEvidence,
			ContractEvidence,
		},
		RequireEvidenceForAmount: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000000)), // 100 VIRT
		AutoEscalateAfterDays:    14,
	}
}

// Validate validates the dispute rules
func (r *DisputeRules) Validate() error {
	if !r.MinDisputeAmount.IsValid() {
		return fmt.Errorf("min_dispute_amount must be valid coins")
	}

	if r.MaxEvidenceCount == 0 {
		return fmt.Errorf("max_evidence_count must be positive")
	}

	if r.MaxEvidenceSize <= 0 {
		return fmt.Errorf("max_evidence_size must be positive")
	}

	if len(r.AllowedEvidenceTypes) == 0 {
		return fmt.Errorf("at least one evidence type must be allowed")
	}

	for i, et := range r.AllowedEvidenceTypes {
		if !et.IsValid() {
			return fmt.Errorf("allowed_evidence_types[%d]: invalid evidence type %d", i, et)
		}
	}

	if !r.RequireEvidenceForAmount.IsValid() {
		return fmt.Errorf("require_evidence_for_amount must be valid coins")
	}

	if r.AutoEscalateAfterDays == 0 {
		return fmt.Errorf("auto_escalate_after_days must be positive")
	}

	return nil
}

// IsEvidenceTypeAllowed checks if an evidence type is allowed
func (r *DisputeRules) IsEvidenceTypeAllowed(et EvidenceType) bool {
	for _, allowed := range r.AllowedEvidenceTypes {
		if allowed == et {
			return true
		}
	}
	return false
}

// RequiresEvidence checks if evidence is required for the given amount
func (r *DisputeRules) RequiresEvidence(amount sdk.Coins) bool {
	return amount.IsAllGTE(r.RequireEvidenceForAmount)
}

// DisputeService defines the interface for dispute management
type DisputeService interface {
	// InitiateDispute initiates a new dispute
	InitiateDispute(
		ctx context.Context,
		invoiceID string,
		category DisputeCategory,
		subject string,
		description string,
		disputedAmount sdk.Coins,
		initiator string,
	) (*DisputeWorkflow, error)

	// UploadEvidence uploads evidence for a dispute
	UploadEvidence(ctx context.Context, disputeID string, evidence DisputeEvidence) error

	// VerifyEvidence verifies evidence for a dispute
	VerifyEvidence(ctx context.Context, disputeID string, evidenceID string, verifier string) error

	// SubmitForReview submits a dispute for review
	SubmitForReview(ctx context.Context, disputeID string) error

	// ResolveDispute resolves a dispute
	ResolveDispute(
		ctx context.Context,
		disputeID string,
		resolution DisputeResolutionType,
		details string,
		resolver string,
		refundAmount sdk.Coins,
	) error

	// EscalateDispute escalates a dispute
	EscalateDispute(ctx context.Context, disputeID string, escalateTo string, reason string) error

	// GetDispute retrieves a dispute by ID
	GetDispute(ctx context.Context, disputeID string) (*DisputeWorkflow, error)

	// GetDisputesByInvoice retrieves all disputes for an invoice
	GetDisputesByInvoice(ctx context.Context, invoiceID string) ([]*DisputeWorkflow, error)

	// GetDisputesByStatus retrieves disputes by status with pagination
	GetDisputesByStatus(ctx context.Context, status DisputeStatus, pagination Pagination) ([]*DisputeWorkflow, error)

	// CheckDisputeWindowOpen checks if the dispute window is open for an invoice
	CheckDisputeWindowOpen(ctx context.Context, invoiceID string, now time.Time) (bool, time.Duration, error)

	// ApplyCorrection applies a correction to a dispute
	ApplyCorrection(ctx context.Context, disputeID string, correction *Correction) error
}

// Store key prefixes for dispute workflow types
var (
	// DisputeWorkflowPrefix is the prefix for dispute workflow storage
	DisputeWorkflowPrefix = []byte{0x43}

	// DisputeWorkflowByInvoicePrefix indexes disputes by invoice
	DisputeWorkflowByInvoicePrefix = []byte{0x44}

	// DisputeEvidencePrefix is the prefix for dispute evidence storage
	DisputeEvidencePrefix = []byte{0x45}

	// DisputeRulesPrefix is the prefix for dispute rules storage
	DisputeRulesPrefix = []byte{0x46}
)

// BuildDisputeWorkflowKey builds the key for a dispute workflow
func BuildDisputeWorkflowKey(disputeID string) []byte {
	return append(DisputeWorkflowPrefix, []byte(disputeID)...)
}

// ParseDisputeWorkflowKey parses a dispute workflow key
func ParseDisputeWorkflowKey(key []byte) (string, error) {
	if len(key) <= len(DisputeWorkflowPrefix) {
		return "", fmt.Errorf("invalid dispute workflow key length")
	}
	return string(key[len(DisputeWorkflowPrefix):]), nil
}

// BuildDisputeWorkflowByInvoiceKey builds the index key for disputes by invoice
func BuildDisputeWorkflowByInvoiceKey(invoiceID string, disputeID string) []byte {
	key := append(DisputeWorkflowByInvoicePrefix, []byte(invoiceID)...)
	key = append(key, byte('/'))
	return append(key, []byte(disputeID)...)
}

// BuildDisputeWorkflowByInvoicePrefix builds the prefix for invoice's disputes
func BuildDisputeWorkflowByInvoicePrefix(invoiceID string) []byte {
	key := append(DisputeWorkflowByInvoicePrefix, []byte(invoiceID)...)
	return append(key, byte('/'))
}

// BuildDisputeWorkflowByStatusKey builds the index key for disputes by status
func BuildDisputeWorkflowByStatusKey(status DisputeStatus, disputeID string) []byte {
	key := append(DisputeByStatusPrefix, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(disputeID)...)
}

// BuildDisputeWorkflowByStatusPrefix builds the prefix for disputes by status
func BuildDisputeWorkflowByStatusPrefix(status DisputeStatus) []byte {
	key := append(DisputeByStatusPrefix, byte(status))
	return append(key, byte('/'))
}

// BuildDisputeEvidenceKey builds the key for dispute evidence
func BuildDisputeEvidenceKey(evidenceID string) []byte {
	return append(DisputeEvidencePrefix, []byte(evidenceID)...)
}

// ParseDisputeEvidenceKey parses a dispute evidence key
func ParseDisputeEvidenceKey(key []byte) (string, error) {
	if len(key) <= len(DisputeEvidencePrefix) {
		return "", fmt.Errorf("invalid dispute evidence key length")
	}
	return string(key[len(DisputeEvidencePrefix):]), nil
}

// BuildDisputeEvidenceByDisputeKey builds the index key for evidence by dispute
func BuildDisputeEvidenceByDisputeKey(disputeID string, evidenceID string) []byte {
	key := append(DisputeEvidencePrefix, []byte(disputeID)...)
	key = append(key, byte('/'))
	return append(key, []byte(evidenceID)...)
}

// BuildDisputeEvidenceByDisputePrefix builds the prefix for dispute's evidence
func BuildDisputeEvidenceByDisputePrefix(disputeID string) []byte {
	key := append(DisputeEvidencePrefix, []byte(disputeID)...)
	return append(key, byte('/'))
}

// BuildDisputeRulesKey builds the key for dispute rules
func BuildDisputeRulesKey(provider string) []byte {
	return append(DisputeRulesPrefix, []byte(provider)...)
}

// BuildDisputeEvidenceByTimestampKey builds the key for evidence sorted by timestamp
func BuildDisputeEvidenceByTimestampKey(disputeID string, timestamp int64, evidenceID string) []byte {
	key := append(DisputeEvidencePrefix, []byte(disputeID)...)
	key = append(key, byte('/'))

	// Append timestamp as big-endian uint64 for deterministic ordering
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
	key = append(key, tsBytes...)
	key = append(key, byte('/'))
	return append(key, []byte(evidenceID)...)
}

// DisputeWorkflowSequenceKey is the key for dispute workflow ID sequence
var DisputeWorkflowSequenceKey = []byte("dispute_workflow_sequence")

// NextDisputeWorkflowID generates the next dispute workflow ID
func NextDisputeWorkflowID(currentSequence uint64, prefix string) string {
	return fmt.Sprintf("%s-DISP-%08d", prefix, currentSequence+1)
}

// DisputeEvidenceSequenceKey is the key for dispute evidence ID sequence
var DisputeEvidenceSequenceKey = []byte("dispute_evidence_sequence")

// NextDisputeEvidenceID generates the next dispute evidence ID
func NextDisputeEvidenceID(currentSequence uint64, disputeID string) string {
	return fmt.Sprintf("%s-EV-%04d", disputeID, currentSequence+1)
}
