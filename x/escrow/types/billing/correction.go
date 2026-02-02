// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"encoding/binary"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CorrectionType defines the type of billing correction
type CorrectionType uint8

const (
	// CorrectionTypeInvoiceAdjustment is an adjustment to an invoice
	CorrectionTypeInvoiceAdjustment CorrectionType = 0

	// CorrectionTypeUsageMismatch corrects usage record discrepancies
	CorrectionTypeUsageMismatch CorrectionType = 1

	// CorrectionTypePayoutCorrection corrects provider payout amounts
	CorrectionTypePayoutCorrection CorrectionType = 2

	// CorrectionTypeRefundAdjustment adjusts refund amounts
	CorrectionTypeRefundAdjustment CorrectionType = 3

	// CorrectionTypeTaxCorrection corrects tax calculations
	CorrectionTypeTaxCorrection CorrectionType = 4

	// CorrectionTypeFeeAdjustment adjusts platform or service fees
	CorrectionTypeFeeAdjustment CorrectionType = 5
)

// CorrectionTypeNames maps correction types to human-readable names
var CorrectionTypeNames = map[CorrectionType]string{
	CorrectionTypeInvoiceAdjustment: "invoice_adjustment",
	CorrectionTypeUsageMismatch:     "usage_mismatch",
	CorrectionTypePayoutCorrection:  "payout_correction",
	CorrectionTypeRefundAdjustment:  "refund_adjustment",
	CorrectionTypeTaxCorrection:     "tax_correction",
	CorrectionTypeFeeAdjustment:     "fee_adjustment",
}

// String returns string representation
func (t CorrectionType) String() string {
	if name, ok := CorrectionTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid returns true if the correction type is valid
func (t CorrectionType) IsValid() bool {
	_, ok := CorrectionTypeNames[t]
	return ok
}

// CorrectionStatus defines the status of a correction
type CorrectionStatus uint8

const (
	// CorrectionStatusPending is a pending correction awaiting approval
	CorrectionStatusPending CorrectionStatus = 0

	// CorrectionStatusApproved is an approved correction ready to apply
	CorrectionStatusApproved CorrectionStatus = 1

	// CorrectionStatusApplied is a correction that has been applied
	CorrectionStatusApplied CorrectionStatus = 2

	// CorrectionStatusRejected is a correction that was rejected
	CorrectionStatusRejected CorrectionStatus = 3

	// CorrectionStatusCancelled is a correction that was cancelled
	CorrectionStatusCancelled CorrectionStatus = 4
)

// CorrectionStatusNames maps correction status to human-readable names
var CorrectionStatusNames = map[CorrectionStatus]string{
	CorrectionStatusPending:   "pending",
	CorrectionStatusApproved:  "approved",
	CorrectionStatusApplied:   "applied",
	CorrectionStatusRejected:  "rejected",
	CorrectionStatusCancelled: "cancelled",
}

// String returns string representation
func (s CorrectionStatus) String() string {
	if name, ok := CorrectionStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsTerminal returns true if the status is final
func (s CorrectionStatus) IsTerminal() bool {
	return s == CorrectionStatusApplied || s == CorrectionStatusRejected || s == CorrectionStatusCancelled
}

// IsValid returns true if the correction status is valid
func (s CorrectionStatus) IsValid() bool {
	_, ok := CorrectionStatusNames[s]
	return ok
}

// CorrectionLimit defines limits and constraints for billing corrections
type CorrectionLimit struct {
	// MaxCorrectionAmount is the maximum amount allowed per correction
	MaxCorrectionAmount sdk.Coins `json:"max_correction_amount"`

	// MaxCorrectionsPerPeriod is the maximum number of corrections allowed per period
	MaxCorrectionsPerPeriod uint32 `json:"max_corrections_per_period"`

	// CorrectionWindowSeconds is the time window for counting corrections
	CorrectionWindowSeconds int64 `json:"correction_window_seconds"`

	// RequireApprovalThreshold is the amount above which approval is required
	RequireApprovalThreshold sdk.Coins `json:"require_approval_threshold"`

	// AllowedCorrectionTypes are the correction types allowed
	AllowedCorrectionTypes []CorrectionType `json:"allowed_correction_types"`
}

// DefaultCorrectionLimit returns a default correction limit configuration
func DefaultCorrectionLimit() CorrectionLimit {
	return CorrectionLimit{
		MaxCorrectionAmount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000000000)), // 1000 VIRT
		MaxCorrectionsPerPeriod:  10,
		CorrectionWindowSeconds:  604800,                                             // 7 days
		RequireApprovalThreshold: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 100000000)), // 100 VIRT
		AllowedCorrectionTypes: []CorrectionType{
			CorrectionTypeInvoiceAdjustment,
			CorrectionTypeUsageMismatch,
			CorrectionTypePayoutCorrection,
			CorrectionTypeRefundAdjustment,
			CorrectionTypeTaxCorrection,
			CorrectionTypeFeeAdjustment,
		},
	}
}

// Validate validates the correction limit configuration
func (l *CorrectionLimit) Validate() error {
	if !l.MaxCorrectionAmount.IsValid() {
		return fmt.Errorf("max_correction_amount must be valid coins")
	}

	if l.MaxCorrectionsPerPeriod == 0 {
		return fmt.Errorf("max_corrections_per_period must be positive")
	}

	if l.CorrectionWindowSeconds <= 0 {
		return fmt.Errorf("correction_window_seconds must be positive")
	}

	if !l.RequireApprovalThreshold.IsValid() {
		return fmt.Errorf("require_approval_threshold must be valid coins")
	}

	for i, ct := range l.AllowedCorrectionTypes {
		if !ct.IsValid() {
			return fmt.Errorf("allowed_correction_types[%d]: invalid correction type %d", i, ct)
		}
	}

	return nil
}

// IsCorrectionTypeAllowed checks if a correction type is allowed
func (l *CorrectionLimit) IsCorrectionTypeAllowed(ct CorrectionType) bool {
	for _, allowed := range l.AllowedCorrectionTypes {
		if allowed == ct {
			return true
		}
	}
	return false
}

// RequiresApproval checks if an amount requires approval
func (l *CorrectionLimit) RequiresApproval(amount sdk.Coins) bool {
	return amount.IsAllGTE(l.RequireApprovalThreshold)
}

// Correction represents a billing correction for reconciliation
type Correction struct {
	// CorrectionID is the unique identifier
	CorrectionID string `json:"correction_id"`

	// InvoiceID is the invoice being corrected
	InvoiceID string `json:"invoice_id"`

	// SettlementID is the settlement being corrected (optional)
	SettlementID string `json:"settlement_id,omitempty"`

	// Type is the type of correction
	Type CorrectionType `json:"type"`

	// Status is the current correction status
	Status CorrectionStatus `json:"status"`

	// OriginalAmount is the original amount before correction
	OriginalAmount sdk.Coins `json:"original_amount"`

	// CorrectedAmount is the new corrected amount
	CorrectedAmount sdk.Coins `json:"corrected_amount"`

	// Difference is the difference between original and corrected (can be negative)
	Difference sdk.Coins `json:"difference"`

	// Reason is the reason code for the correction
	Reason string `json:"reason"`

	// Description is a detailed description of the correction
	Description string `json:"description"`

	// RequestedBy is the address that requested the correction
	RequestedBy string `json:"requested_by"`

	// ApprovedBy is the address that approved the correction (optional)
	ApprovedBy string `json:"approved_by,omitempty"`

	// RequestedAt is when the correction was requested
	RequestedAt time.Time `json:"requested_at"`

	// AppliedAt is when the correction was applied (optional)
	AppliedAt *time.Time `json:"applied_at,omitempty"`

	// EvidenceArtifactCIDs are content-addressable identifiers for evidence documents
	EvidenceArtifactCIDs []string `json:"evidence_artifact_cids,omitempty"`

	// BlockHeight is when the correction was created on-chain
	BlockHeight int64 `json:"block_height"`

	// Metadata contains additional correction details
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewCorrection creates a new correction
func NewCorrection(
	correctionID string,
	invoiceID string,
	settlementID string,
	correctionType CorrectionType,
	originalAmount sdk.Coins,
	correctedAmount sdk.Coins,
	reason string,
	description string,
	requestedBy string,
	blockHeight int64,
	now time.Time,
) *Correction {
	// Calculate difference (corrected - original)
	difference := correctedAmount.Sub(originalAmount...)

	return &Correction{
		CorrectionID:         correctionID,
		InvoiceID:            invoiceID,
		SettlementID:         settlementID,
		Type:                 correctionType,
		Status:               CorrectionStatusPending,
		OriginalAmount:       originalAmount,
		CorrectedAmount:      correctedAmount,
		Difference:           difference,
		Reason:               reason,
		Description:          description,
		RequestedBy:          requestedBy,
		RequestedAt:          now,
		EvidenceArtifactCIDs: make([]string, 0),
		BlockHeight:          blockHeight,
		Metadata:             make(map[string]string),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// Validate validates the correction
func (c *Correction) Validate() error {
	if c.CorrectionID == "" {
		return fmt.Errorf("correction_id is required")
	}

	if len(c.CorrectionID) > 64 {
		return fmt.Errorf("correction_id exceeds maximum length of 64")
	}

	if c.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if !c.Type.IsValid() {
		return fmt.Errorf("invalid correction type: %d", c.Type)
	}

	if !c.Status.IsValid() {
		return fmt.Errorf("invalid correction status: %d", c.Status)
	}

	if !c.OriginalAmount.IsValid() {
		return fmt.Errorf("original_amount must be valid coins")
	}

	if !c.CorrectedAmount.IsValid() {
		return fmt.Errorf("corrected_amount must be valid coins")
	}

	if c.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	if len(c.Reason) > 256 {
		return fmt.Errorf("reason exceeds maximum length of 256")
	}

	if _, err := sdk.AccAddressFromBech32(c.RequestedBy); err != nil {
		return fmt.Errorf("invalid requested_by address: %w", err)
	}

	if c.ApprovedBy != "" {
		if _, err := sdk.AccAddressFromBech32(c.ApprovedBy); err != nil {
			return fmt.Errorf("invalid approved_by address: %w", err)
		}
	}

	// Validate evidence CIDs format
	for i, cid := range c.EvidenceArtifactCIDs {
		if cid == "" {
			return fmt.Errorf("evidence_artifact_cids[%d]: empty CID", i)
		}
		if len(cid) > 128 {
			return fmt.Errorf("evidence_artifact_cids[%d]: exceeds maximum length of 128", i)
		}
	}

	return nil
}

// Approve approves the correction
func (c *Correction) Approve(approvedBy string, now time.Time) error {
	if c.Status != CorrectionStatusPending {
		return fmt.Errorf("can only approve pending corrections, current status: %s", c.Status)
	}

	c.Status = CorrectionStatusApproved
	c.ApprovedBy = approvedBy
	c.UpdatedAt = now
	return nil
}

// Apply applies the correction
func (c *Correction) Apply(now time.Time) error {
	if c.Status != CorrectionStatusApproved && c.Status != CorrectionStatusPending {
		return fmt.Errorf("can only apply approved or pending corrections, current status: %s", c.Status)
	}

	c.Status = CorrectionStatusApplied
	c.AppliedAt = &now
	c.UpdatedAt = now
	return nil
}

// Reject rejects the correction
func (c *Correction) Reject(rejectedBy string, now time.Time) error {
	if c.Status.IsTerminal() {
		return fmt.Errorf("cannot reject terminal correction, current status: %s", c.Status)
	}

	c.Status = CorrectionStatusRejected
	c.ApprovedBy = rejectedBy // Store who rejected in ApprovedBy field
	c.UpdatedAt = now
	return nil
}

// Cancel cancels the correction
func (c *Correction) Cancel(now time.Time) error {
	if c.Status.IsTerminal() {
		return fmt.Errorf("cannot cancel terminal correction, current status: %s", c.Status)
	}

	c.Status = CorrectionStatusCancelled
	c.UpdatedAt = now
	return nil
}

// AddEvidence adds evidence artifact CID to the correction
func (c *Correction) AddEvidence(cid string) error {
	if cid == "" {
		return fmt.Errorf("evidence CID cannot be empty")
	}

	if len(cid) > 128 {
		return fmt.Errorf("evidence CID exceeds maximum length of 128")
	}

	c.EvidenceArtifactCIDs = append(c.EvidenceArtifactCIDs, cid)
	return nil
}

// GetAbsoluteDifference returns the absolute value of the difference
func (c *Correction) GetAbsoluteDifference() sdk.Coins {
	// Since sdk.Coins can't be negative, we return whichever is larger
	if c.CorrectedAmount.IsAllGTE(c.OriginalAmount) {
		return c.CorrectedAmount.Sub(c.OriginalAmount...)
	}
	return c.OriginalAmount.Sub(c.CorrectedAmount...)
}

// CorrectionLedgerEntry represents a state change in the correction ledger
// This is used for deterministic tracking of all correction state changes
type CorrectionLedgerEntry struct {
	// EntryID is the unique identifier for this entry
	EntryID string `json:"entry_id"`

	// CorrectionID is the correction this entry relates to
	CorrectionID string `json:"correction_id"`

	// InvoiceID is the related invoice
	InvoiceID string `json:"invoice_id"`

	// EntryType is the type of ledger entry
	EntryType CorrectionLedgerEntryType `json:"entry_type"`

	// PreviousStatus is the status before this change
	PreviousStatus CorrectionStatus `json:"previous_status"`

	// NewStatus is the status after this change
	NewStatus CorrectionStatus `json:"new_status"`

	// Amount is the correction amount
	Amount sdk.Coins `json:"amount"`

	// Description describes the entry
	Description string `json:"description"`

	// Initiator is who initiated this change
	Initiator string `json:"initiator"`

	// TransactionHash is the on-chain transaction hash (if applicable)
	TransactionHash string `json:"transaction_hash,omitempty"`

	// BlockHeight is when this entry was created
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when this entry was created
	Timestamp time.Time `json:"timestamp"`
}

// CorrectionLedgerEntryType defines types of correction ledger entries
type CorrectionLedgerEntryType uint8

const (
	// CorrectionLedgerEntryTypeCreated is when correction is created
	CorrectionLedgerEntryTypeCreated CorrectionLedgerEntryType = 0

	// CorrectionLedgerEntryTypeApproved is when correction is approved
	CorrectionLedgerEntryTypeApproved CorrectionLedgerEntryType = 1

	// CorrectionLedgerEntryTypeApplied is when correction is applied
	CorrectionLedgerEntryTypeApplied CorrectionLedgerEntryType = 2

	// CorrectionLedgerEntryTypeRejected is when correction is rejected
	CorrectionLedgerEntryTypeRejected CorrectionLedgerEntryType = 3

	// CorrectionLedgerEntryTypeCancelled is when correction is cancelled
	CorrectionLedgerEntryTypeCancelled CorrectionLedgerEntryType = 4

	// CorrectionLedgerEntryTypeEvidenceAdded is when evidence is added
	CorrectionLedgerEntryTypeEvidenceAdded CorrectionLedgerEntryType = 5
)

// CorrectionLedgerEntryTypeNames maps types to names
var CorrectionLedgerEntryTypeNames = map[CorrectionLedgerEntryType]string{
	CorrectionLedgerEntryTypeCreated:       "created",
	CorrectionLedgerEntryTypeApproved:      "approved",
	CorrectionLedgerEntryTypeApplied:       "applied",
	CorrectionLedgerEntryTypeRejected:      "rejected",
	CorrectionLedgerEntryTypeCancelled:     "cancelled",
	CorrectionLedgerEntryTypeEvidenceAdded: "evidence_added",
}

// String returns string representation
func (t CorrectionLedgerEntryType) String() string {
	if name, ok := CorrectionLedgerEntryTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// NewCorrectionLedgerEntry creates a new correction ledger entry
func NewCorrectionLedgerEntry(
	entryID string,
	correctionID string,
	invoiceID string,
	entryType CorrectionLedgerEntryType,
	previousStatus CorrectionStatus,
	newStatus CorrectionStatus,
	amount sdk.Coins,
	description string,
	initiator string,
	txHash string,
	blockHeight int64,
	timestamp time.Time,
) *CorrectionLedgerEntry {
	return &CorrectionLedgerEntry{
		EntryID:         entryID,
		CorrectionID:    correctionID,
		InvoiceID:       invoiceID,
		EntryType:       entryType,
		PreviousStatus:  previousStatus,
		NewStatus:       newStatus,
		Amount:          amount,
		Description:     description,
		Initiator:       initiator,
		TransactionHash: txHash,
		BlockHeight:     blockHeight,
		Timestamp:       timestamp,
	}
}

// Validate validates the correction ledger entry
func (e *CorrectionLedgerEntry) Validate() error {
	if e.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	if e.CorrectionID == "" {
		return fmt.Errorf("correction_id is required")
	}

	if e.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if e.Initiator == "" {
		return fmt.Errorf("initiator is required")
	}

	return nil
}

// CorrectionSummary provides a summary view of a correction
type CorrectionSummary struct {
	CorrectionID    string           `json:"correction_id"`
	InvoiceID       string           `json:"invoice_id"`
	Type            CorrectionType   `json:"type"`
	Status          CorrectionStatus `json:"status"`
	OriginalAmount  sdk.Coins        `json:"original_amount"`
	CorrectedAmount sdk.Coins        `json:"corrected_amount"`
	Difference      sdk.Coins        `json:"difference"`
	Reason          string           `json:"reason"`
	RequestedBy     string           `json:"requested_by"`
	RequestedAt     time.Time        `json:"requested_at"`
}

// ToSummary creates a summary from the correction
func (c *Correction) ToSummary() CorrectionSummary {
	return CorrectionSummary{
		CorrectionID:    c.CorrectionID,
		InvoiceID:       c.InvoiceID,
		Type:            c.Type,
		Status:          c.Status,
		OriginalAmount:  c.OriginalAmount,
		CorrectedAmount: c.CorrectedAmount,
		Difference:      c.Difference,
		Reason:          c.Reason,
		RequestedBy:     c.RequestedBy,
		RequestedAt:     c.RequestedAt,
	}
}

// Store key prefixes for correction types
var (
	// CorrectionPrefix is the prefix for correction storage
	CorrectionPrefix = []byte{0x80}

	// CorrectionByInvoicePrefix indexes corrections by invoice
	CorrectionByInvoicePrefix = []byte{0x81}

	// CorrectionByStatusPrefix indexes corrections by status
	CorrectionByStatusPrefix = []byte{0x82}

	// CorrectionLedgerEntryPrefix is the prefix for correction ledger entries
	CorrectionLedgerEntryPrefix = []byte{0x83}

	// CorrectionLedgerEntryByCorrectionPrefix indexes entries by correction
	CorrectionLedgerEntryByCorrectionPrefix = []byte{0x84}

	// CorrectionBySettlementPrefix indexes corrections by settlement
	CorrectionBySettlementPrefix = []byte{0x85}

	// CorrectionByRequesterPrefix indexes corrections by requester
	CorrectionByRequesterPrefix = []byte{0x86}

	// CorrectionLimitPrefix is the prefix for correction limits
	CorrectionLimitPrefix = []byte{0x87}
)

// BuildCorrectionKey builds the key for a correction
func BuildCorrectionKey(correctionID string) []byte {
	return append(CorrectionPrefix, []byte(correctionID)...)
}

// ParseCorrectionKey parses a correction key
func ParseCorrectionKey(key []byte) (string, error) {
	if len(key) <= len(CorrectionPrefix) {
		return "", fmt.Errorf("invalid correction key length")
	}
	return string(key[len(CorrectionPrefix):]), nil
}

// BuildCorrectionByInvoiceKey builds the index key for corrections by invoice
func BuildCorrectionByInvoiceKey(invoiceID string, correctionID string) []byte {
	key := make([]byte, 0, len(CorrectionByInvoicePrefix)+len(invoiceID)+len(correctionID)+1)
	key = append(key, CorrectionByInvoicePrefix...)
	key = append(key, []byte(invoiceID)...)
	key = append(key, byte('/'))
	return append(key, []byte(correctionID)...)
}

// BuildCorrectionByInvoicePrefix builds the prefix for invoice's corrections
func BuildCorrectionByInvoicePrefix(invoiceID string) []byte {
	key := make([]byte, 0, len(CorrectionByInvoicePrefix)+len(invoiceID)+1)
	key = append(key, CorrectionByInvoicePrefix...)
	key = append(key, []byte(invoiceID)...)
	return append(key, byte('/'))
}

// BuildCorrectionByStatusKey builds the index key for corrections by status
func BuildCorrectionByStatusKey(status CorrectionStatus, correctionID string) []byte {
	key := make([]byte, 0, len(CorrectionByStatusPrefix)+len(correctionID)+2)
	key = append(key, CorrectionByStatusPrefix...)
	key = append(key, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(correctionID)...)
}

// BuildCorrectionByStatusPrefix builds the prefix for corrections by status
func BuildCorrectionByStatusPrefix(status CorrectionStatus) []byte {
	key := make([]byte, 0, len(CorrectionByStatusPrefix)+2)
	key = append(key, CorrectionByStatusPrefix...)
	key = append(key, byte(status))
	return append(key, byte('/'))
}

// BuildCorrectionBySettlementKey builds the index key for corrections by settlement
func BuildCorrectionBySettlementKey(settlementID string, correctionID string) []byte {
	key := make([]byte, 0, len(CorrectionBySettlementPrefix)+len(settlementID)+len(correctionID)+1)
	key = append(key, CorrectionBySettlementPrefix...)
	key = append(key, []byte(settlementID)...)
	key = append(key, byte('/'))
	return append(key, []byte(correctionID)...)
}

// BuildCorrectionBySettlementPrefix builds the prefix for settlement's corrections
func BuildCorrectionBySettlementPrefix(settlementID string) []byte {
	key := make([]byte, 0, len(CorrectionBySettlementPrefix)+len(settlementID)+1)
	key = append(key, CorrectionBySettlementPrefix...)
	key = append(key, []byte(settlementID)...)
	return append(key, byte('/'))
}

// BuildCorrectionByRequesterKey builds the index key for corrections by requester
func BuildCorrectionByRequesterKey(requester string, correctionID string) []byte {
	key := make([]byte, 0, len(CorrectionByRequesterPrefix)+len(requester)+len(correctionID)+1)
	key = append(key, CorrectionByRequesterPrefix...)
	key = append(key, []byte(requester)...)
	key = append(key, byte('/'))
	return append(key, []byte(correctionID)...)
}

// BuildCorrectionByRequesterPrefix builds the prefix for requester's corrections
func BuildCorrectionByRequesterPrefix(requester string) []byte {
	key := make([]byte, 0, len(CorrectionByRequesterPrefix)+len(requester)+1)
	key = append(key, CorrectionByRequesterPrefix...)
	key = append(key, []byte(requester)...)
	return append(key, byte('/'))
}

// BuildCorrectionLedgerEntryKey builds the key for a correction ledger entry
func BuildCorrectionLedgerEntryKey(entryID string) []byte {
	key := make([]byte, 0, len(CorrectionLedgerEntryPrefix)+len(entryID))
	key = append(key, CorrectionLedgerEntryPrefix...)
	return append(key, []byte(entryID)...)
}

// BuildCorrectionLedgerEntryByCorrectionKey builds the index key for entries by correction
func BuildCorrectionLedgerEntryByCorrectionKey(correctionID string, timestamp int64) []byte {
	key := make([]byte, 0, len(CorrectionLedgerEntryByCorrectionPrefix)+len(correctionID)+1+8)
	key = append(key, CorrectionLedgerEntryByCorrectionPrefix...)
	key = append(key, []byte(correctionID)...)
	key = append(key, byte('/'))

	// Append timestamp as big-endian uint64 for deterministic ordering
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
	return append(key, tsBytes...)
}

// BuildCorrectionLedgerEntryByCorrectionPrefix builds the prefix for correction's entries
func BuildCorrectionLedgerEntryByCorrectionPrefix(correctionID string) []byte {
	key := make([]byte, 0, len(CorrectionLedgerEntryByCorrectionPrefix)+len(correctionID)+1)
	key = append(key, CorrectionLedgerEntryByCorrectionPrefix...)
	key = append(key, []byte(correctionID)...)
	return append(key, byte('/'))
}

// BuildCorrectionLimitKey builds the key for correction limits
func BuildCorrectionLimitKey(provider string) []byte {
	key := make([]byte, 0, len(CorrectionLimitPrefix)+len(provider))
	key = append(key, CorrectionLimitPrefix...)
	return append(key, []byte(provider)...)
}

// CorrectionSequenceKey is the key for correction ID sequence
var CorrectionSequenceKey = []byte("correction_sequence")

// NextCorrectionID generates the next correction ID
func NextCorrectionID(currentSequence uint64, prefix string) string {
	return fmt.Sprintf("%s-CORR-%08d", prefix, currentSequence+1)
}
