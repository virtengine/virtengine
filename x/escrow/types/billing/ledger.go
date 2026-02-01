// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InvoiceLedgerRecord is the immutable on-chain invoice record.
// This contains a hash of the full invoice for verification,
// while the full invoice document is stored in the artifact store.
type InvoiceLedgerRecord struct {
	// InvoiceID is the unique identifier
	InvoiceID string `json:"invoice_id"`

	// InvoiceNumber is the human-readable invoice number
	InvoiceNumber string `json:"invoice_number"`

	// EscrowID links to the escrow account
	EscrowID string `json:"escrow_id"`

	// OrderID links to the marketplace order
	OrderID string `json:"order_id"`

	// LeaseID links to the marketplace lease
	LeaseID string `json:"lease_id"`

	// Provider is the service provider address
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// Status is the current invoice status
	Status InvoiceStatus `json:"status"`

	// Currency is the primary currency/denomination
	Currency string `json:"currency"`

	// Subtotal is the sum of line items before adjustments
	Subtotal sdk.Coins `json:"subtotal"`

	// DiscountTotal is the total discount amount
	DiscountTotal sdk.Coins `json:"discount_total"`

	// TaxTotal is the total tax amount
	TaxTotal sdk.Coins `json:"tax_total"`

	// Total is the final amount due
	Total sdk.Coins `json:"total"`

	// AmountPaid is the amount already paid
	AmountPaid sdk.Coins `json:"amount_paid"`

	// AmountDue is the remaining amount due
	AmountDue sdk.Coins `json:"amount_due"`

	// LineItemCount is the number of line items
	LineItemCount uint32 `json:"line_item_count"`

	// BillingPeriodStart is the billing period start
	BillingPeriodStart time.Time `json:"billing_period_start"`

	// BillingPeriodEnd is the billing period end
	BillingPeriodEnd time.Time `json:"billing_period_end"`

	// DueDate is when payment is due
	DueDate time.Time `json:"due_date"`

	// IssuedAt is when the invoice was issued
	IssuedAt time.Time `json:"issued_at"`

	// PaidAt is when fully paid (if paid)
	PaidAt *time.Time `json:"paid_at,omitempty"`

	// ContentHash is the SHA-256 hash of the full invoice document
	ContentHash string `json:"content_hash"`

	// ArtifactCID is the content-addressable identifier for the full invoice
	ArtifactCID string `json:"artifact_cid"`

	// SettlementID links to the settlement record (if settled)
	SettlementID string `json:"settlement_id,omitempty"`

	// BlockHeight is when the invoice was created on-chain
	BlockHeight int64 `json:"block_height"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewInvoiceLedgerRecord creates a ledger record from a full invoice
func NewInvoiceLedgerRecord(inv *Invoice, artifactCID string, blockHeight int64, now time.Time) (*InvoiceLedgerRecord, error) {
	// Compute content hash from invoice JSON
	invJSON, err := json.Marshal(inv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal invoice for hashing: %w", err)
	}

	hash := sha256.Sum256(invJSON)
	contentHash := hex.EncodeToString(hash[:])

	return &InvoiceLedgerRecord{
		InvoiceID:          inv.InvoiceID,
		InvoiceNumber:      inv.InvoiceNumber,
		EscrowID:           inv.EscrowID,
		OrderID:            inv.OrderID,
		LeaseID:            inv.LeaseID,
		Provider:           inv.Provider,
		Customer:           inv.Customer,
		Status:             inv.Status,
		Currency:           inv.Currency,
		Subtotal:           inv.Subtotal,
		DiscountTotal:      inv.DiscountTotal,
		TaxTotal:           inv.TaxTotal,
		Total:              inv.Total,
		AmountPaid:         inv.AmountPaid,
		AmountDue:          inv.AmountDue,
		LineItemCount:      uint32(len(inv.LineItems)),
		BillingPeriodStart: inv.BillingPeriod.StartTime,
		BillingPeriodEnd:   inv.BillingPeriod.EndTime,
		DueDate:            inv.DueDate,
		IssuedAt:           inv.IssuedAt,
		PaidAt:             inv.PaidAt,
		ContentHash:        contentHash,
		ArtifactCID:        artifactCID,
		SettlementID:       inv.SettlementID,
		BlockHeight:        blockHeight,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// Validate validates the ledger record
func (r *InvoiceLedgerRecord) Validate() error {
	if r.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if len(r.InvoiceID) > 64 {
		return fmt.Errorf("invoice_id exceeds maximum length of 64")
	}

	if r.EscrowID == "" {
		return fmt.Errorf("escrow_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(r.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if r.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if r.ContentHash == "" {
		return fmt.Errorf("content_hash is required")
	}

	if len(r.ContentHash) != 64 {
		return fmt.Errorf("content_hash must be 64 hex characters (SHA-256)")
	}

	if r.BillingPeriodEnd.Before(r.BillingPeriodStart) {
		return fmt.Errorf("billing_period_end must be after billing_period_start")
	}

	if !r.Total.IsValid() {
		return fmt.Errorf("total must be valid coins")
	}

	return nil
}

// VerifyContentHash verifies the content hash against a full invoice
func (r *InvoiceLedgerRecord) VerifyContentHash(inv *Invoice) (bool, error) {
	invJSON, err := json.Marshal(inv)
	if err != nil {
		return false, fmt.Errorf("failed to marshal invoice for verification: %w", err)
	}

	hash := sha256.Sum256(invJSON)
	expectedHash := hex.EncodeToString(hash[:])

	return r.ContentHash == expectedHash, nil
}

// ToSummary creates a summary view from the ledger record
func (r *InvoiceLedgerRecord) ToSummary() InvoiceSummary {
	return InvoiceSummary{
		InvoiceID:     r.InvoiceID,
		InvoiceNumber: r.InvoiceNumber,
		Provider:      r.Provider,
		Customer:      r.Customer,
		Status:        r.Status,
		Total:         r.Total,
		AmountPaid:    r.AmountPaid,
		AmountDue:     r.AmountDue,
		DueDate:       r.DueDate,
		IssuedAt:      r.IssuedAt,
	}
}

// InvoiceLedgerEntry represents a state change in the invoice ledger.
// Entries form an immutable hash chain for audit integrity.
type InvoiceLedgerEntry struct {
	// EntryID is the unique identifier for this entry
	EntryID string `json:"entry_id"`

	// InvoiceID is the invoice this entry relates to
	InvoiceID string `json:"invoice_id"`

	// EntryType is the type of ledger entry
	EntryType InvoiceLedgerEntryType `json:"entry_type"`

	// PreviousStatus is the status before this change
	PreviousStatus InvoiceStatus `json:"previous_status"`

	// NewStatus is the status after this change
	NewStatus InvoiceStatus `json:"new_status"`

	// Amount is the amount involved in this entry (for payments/refunds)
	Amount sdk.Coins `json:"amount,omitempty"`

	// Description describes the entry
	Description string `json:"description"`

	// Initiator is who initiated this change
	Initiator string `json:"initiator"`

	// TransactionHash is the on-chain transaction hash (if applicable)
	TransactionHash string `json:"transaction_hash,omitempty"`

	// PreviousEntryHash is the SHA-256 hash of the previous entry in the chain.
	// For the first entry (genesis), this is the zero hash (64 zeros).
	// This creates an immutable hash chain for audit integrity.
	PreviousEntryHash string `json:"previous_entry_hash"`

	// EntryHash is the SHA-256 hash of this entry (computed from all fields except EntryHash)
	EntryHash string `json:"entry_hash"`

	// SequenceNumber is the position in the invoice's ledger chain (1-indexed)
	SequenceNumber uint64 `json:"sequence_number"`

	// BlockHeight is when this entry was created
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when this entry was created
	Timestamp time.Time `json:"timestamp"`
}

// InvoiceLedgerEntryType defines types of ledger entries
type InvoiceLedgerEntryType uint8

const (
	// LedgerEntryTypeCreated is when invoice is created
	LedgerEntryTypeCreated InvoiceLedgerEntryType = 0

	// LedgerEntryTypeIssued is when invoice is issued
	LedgerEntryTypeIssued InvoiceLedgerEntryType = 1

	// LedgerEntryTypePayment is when payment is recorded
	LedgerEntryTypePayment InvoiceLedgerEntryType = 2

	// LedgerEntryTypeDisputed is when invoice is disputed
	LedgerEntryTypeDisputed InvoiceLedgerEntryType = 3

	// LedgerEntryTypeResolved is when dispute is resolved
	LedgerEntryTypeResolved InvoiceLedgerEntryType = 4

	// LedgerEntryTypeCancelled is when invoice is cancelled
	LedgerEntryTypeCancelled InvoiceLedgerEntryType = 5

	// LedgerEntryTypeRefunded is when invoice is refunded
	LedgerEntryTypeRefunded InvoiceLedgerEntryType = 6

	// LedgerEntryTypeOverdue is when invoice becomes overdue
	LedgerEntryTypeOverdue InvoiceLedgerEntryType = 7

	// LedgerEntryTypeArtifactStored is when artifact is stored
	LedgerEntryTypeArtifactStored InvoiceLedgerEntryType = 8
)

// LedgerEntryTypeNames maps types to names
var LedgerEntryTypeNames = map[InvoiceLedgerEntryType]string{
	LedgerEntryTypeCreated:        "created",
	LedgerEntryTypeIssued:         "issued",
	LedgerEntryTypePayment:        "payment",
	LedgerEntryTypeDisputed:       "disputed",
	LedgerEntryTypeResolved:       "resolved",
	LedgerEntryTypeCancelled:      "cancelled",
	LedgerEntryTypeRefunded:       "refunded",
	LedgerEntryTypeOverdue:        "overdue",
	LedgerEntryTypeArtifactStored: "artifact_stored",
}

// String returns string representation
func (t InvoiceLedgerEntryType) String() string {
	if name, ok := LedgerEntryTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// ZeroHash is the hash value used for genesis entries (first in chain)
const ZeroHash = "0000000000000000000000000000000000000000000000000000000000000000"

// NewInvoiceLedgerEntry creates a new ledger entry
func NewInvoiceLedgerEntry(
	entryID string,
	invoiceID string,
	entryType InvoiceLedgerEntryType,
	previousStatus InvoiceStatus,
	newStatus InvoiceStatus,
	amount sdk.Coins,
	description string,
	initiator string,
	txHash string,
	previousEntryHash string,
	sequenceNumber uint64,
	blockHeight int64,
	timestamp time.Time,
) *InvoiceLedgerEntry {
	entry := &InvoiceLedgerEntry{
		EntryID:           entryID,
		InvoiceID:         invoiceID,
		EntryType:         entryType,
		PreviousStatus:    previousStatus,
		NewStatus:         newStatus,
		Amount:            amount,
		Description:       description,
		Initiator:         initiator,
		TransactionHash:   txHash,
		PreviousEntryHash: previousEntryHash,
		SequenceNumber:    sequenceNumber,
		BlockHeight:       blockHeight,
		Timestamp:         timestamp,
	}

	// Compute entry hash
	entry.EntryHash = entry.ComputeHash()

	return entry
}

// NewGenesisLedgerEntry creates the first ledger entry for an invoice
func NewGenesisLedgerEntry(
	entryID string,
	invoiceID string,
	description string,
	initiator string,
	txHash string,
	blockHeight int64,
	timestamp time.Time,
) *InvoiceLedgerEntry {
	return NewInvoiceLedgerEntry(
		entryID,
		invoiceID,
		LedgerEntryTypeCreated,
		InvoiceStatusDraft,
		InvoiceStatusDraft,
		sdk.Coins{},
		description,
		initiator,
		txHash,
		ZeroHash, // Genesis entry uses zero hash
		1,        // First in sequence
		blockHeight,
		timestamp,
	)
}

// ComputeHash computes the SHA-256 hash of this entry
func (e *InvoiceLedgerEntry) ComputeHash() string {
	// Create a deterministic representation excluding EntryHash itself
	data := fmt.Sprintf(
		"%s|%s|%d|%d|%d|%s|%s|%s|%s|%s|%d|%d|%s",
		e.EntryID,
		e.InvoiceID,
		e.EntryType,
		e.PreviousStatus,
		e.NewStatus,
		e.Amount.String(),
		e.Description,
		e.Initiator,
		e.TransactionHash,
		e.PreviousEntryHash,
		e.SequenceNumber,
		e.BlockHeight,
		e.Timestamp.UTC().Format(time.RFC3339Nano),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Validate validates the ledger entry
func (e *InvoiceLedgerEntry) Validate() error {
	if e.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	if e.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if e.Initiator == "" {
		return fmt.Errorf("initiator is required")
	}

	if e.PreviousEntryHash == "" {
		return fmt.Errorf("previous_entry_hash is required")
	}

	if len(e.PreviousEntryHash) != 64 {
		return fmt.Errorf("previous_entry_hash must be 64 hex characters (SHA-256)")
	}

	if e.EntryHash == "" {
		return fmt.Errorf("entry_hash is required")
	}

	if len(e.EntryHash) != 64 {
		return fmt.Errorf("entry_hash must be 64 hex characters (SHA-256)")
	}

	if e.SequenceNumber == 0 {
		return fmt.Errorf("sequence_number must be positive")
	}

	// Verify hash integrity
	if !e.VerifyHash() {
		return fmt.Errorf("entry_hash does not match computed hash")
	}

	// Verify genesis entry has zero hash
	if e.SequenceNumber == 1 && e.PreviousEntryHash != ZeroHash {
		return fmt.Errorf("genesis entry (sequence 1) must have zero previous_entry_hash")
	}

	// Verify non-genesis entries don't have zero hash
	if e.SequenceNumber > 1 && e.PreviousEntryHash == ZeroHash {
		return fmt.Errorf("non-genesis entry must have non-zero previous_entry_hash")
	}

	return nil
}

// VerifyHash verifies that EntryHash matches the computed hash
func (e *InvoiceLedgerEntry) VerifyHash() bool {
	return e.EntryHash == e.ComputeHash()
}

// VerifyChainLink verifies that this entry correctly links to the previous entry
func (e *InvoiceLedgerEntry) VerifyChainLink(previousEntry *InvoiceLedgerEntry) bool {
	if previousEntry == nil {
		// This should be the genesis entry
		return e.SequenceNumber == 1 && e.PreviousEntryHash == ZeroHash
	}

	// Verify sequence is consecutive
	if e.SequenceNumber != previousEntry.SequenceNumber+1 {
		return false
	}

	// Verify hash chain link
	return e.PreviousEntryHash == previousEntry.EntryHash
}

// InvoiceLedgerChain represents an ordered chain of ledger entries for an invoice
type InvoiceLedgerChain struct {
	InvoiceID string                `json:"invoice_id"`
	Entries   []*InvoiceLedgerEntry `json:"entries"`
}

// NewInvoiceLedgerChain creates a new ledger chain
func NewInvoiceLedgerChain(invoiceID string) *InvoiceLedgerChain {
	return &InvoiceLedgerChain{
		InvoiceID: invoiceID,
		Entries:   make([]*InvoiceLedgerEntry, 0),
	}
}

// AddEntry adds an entry to the chain
func (c *InvoiceLedgerChain) AddEntry(entry *InvoiceLedgerEntry) error {
	if entry.InvoiceID != c.InvoiceID {
		return fmt.Errorf("entry invoice_id %s does not match chain invoice_id %s",
			entry.InvoiceID, c.InvoiceID)
	}

	expectedSeq := uint64(len(c.Entries) + 1)
	if entry.SequenceNumber != expectedSeq {
		return fmt.Errorf("entry sequence_number %d does not match expected %d",
			entry.SequenceNumber, expectedSeq)
	}

	// Verify chain link
	var previousEntry *InvoiceLedgerEntry
	if len(c.Entries) > 0 {
		previousEntry = c.Entries[len(c.Entries)-1]
	}

	if !entry.VerifyChainLink(previousEntry) {
		return fmt.Errorf("entry does not correctly link to previous entry")
	}

	c.Entries = append(c.Entries, entry)
	return nil
}

// Validate validates the entire chain
func (c *InvoiceLedgerChain) Validate() error {
	if c.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	for i, entry := range c.Entries {
		if err := entry.Validate(); err != nil {
			return fmt.Errorf("entry %d: %w", i, err)
		}

		// Verify chain continuity
		var previousEntry *InvoiceLedgerEntry
		if i > 0 {
			previousEntry = c.Entries[i-1]
		}

		if !entry.VerifyChainLink(previousEntry) {
			return fmt.Errorf("entry %d: chain link verification failed", i)
		}
	}

	return nil
}

// LastEntry returns the last entry in the chain
func (c *InvoiceLedgerChain) LastEntry() *InvoiceLedgerEntry {
	if len(c.Entries) == 0 {
		return nil
	}
	return c.Entries[len(c.Entries)-1]
}

// GetPreviousHash returns the hash to use for the next entry
func (c *InvoiceLedgerChain) GetPreviousHash() string {
	if len(c.Entries) == 0 {
		return ZeroHash
	}
	return c.Entries[len(c.Entries)-1].EntryHash
}

// NextSequenceNumber returns the next sequence number
func (c *InvoiceLedgerChain) NextSequenceNumber() uint64 {
	return uint64(len(c.Entries) + 1)
}

// Store key prefixes for ledger types
var (
	// InvoiceLedgerRecordPrefix is the prefix for ledger records
	InvoiceLedgerRecordPrefix = []byte{0x60}

	// InvoiceLedgerEntryPrefix is the prefix for ledger entries
	InvoiceLedgerEntryPrefix = []byte{0x61}

	// InvoiceLedgerEntryByInvoicePrefix indexes entries by invoice
	InvoiceLedgerEntryByInvoicePrefix = []byte{0x62}

	// InvoiceArtifactPrefix is the prefix for artifact references
	InvoiceArtifactPrefix = []byte{0x63}

	// InvoiceLedgerEntrySeqPrefix indexes entries by invoice and sequence number
	InvoiceLedgerEntrySeqPrefix = []byte{0x64}
)

// BuildInvoiceLedgerRecordKey builds the key for a ledger record
func BuildInvoiceLedgerRecordKey(invoiceID string) []byte {
	return append(InvoiceLedgerRecordPrefix, []byte(invoiceID)...)
}

// BuildInvoiceLedgerEntryKey builds the key for a ledger entry
func BuildInvoiceLedgerEntryKey(entryID string) []byte {
	return append(InvoiceLedgerEntryPrefix, []byte(entryID)...)
}

// BuildInvoiceLedgerEntryByInvoiceKey builds the index key for entries by invoice
func BuildInvoiceLedgerEntryByInvoiceKey(invoiceID string, entryID string) []byte {
	key := append(InvoiceLedgerEntryByInvoicePrefix, []byte(invoiceID)...)
	key = append(key, byte('/'))
	return append(key, []byte(entryID)...)
}

// BuildInvoiceLedgerEntryByInvoicePrefix builds the prefix for invoice's entries
func BuildInvoiceLedgerEntryByInvoicePrefix(invoiceID string) []byte {
	key := append(InvoiceLedgerEntryByInvoicePrefix, []byte(invoiceID)...)
	return append(key, byte('/'))
}

// BuildInvoiceArtifactKey builds the key for an artifact reference
func BuildInvoiceArtifactKey(cid string) []byte {
	return append(InvoiceArtifactPrefix, []byte(cid)...)
}

// BuildInvoiceLedgerEntrySeqKey builds the key for entries indexed by sequence
func BuildInvoiceLedgerEntrySeqKey(invoiceID string, seqNum uint64) []byte {
	key := append(InvoiceLedgerEntrySeqPrefix, []byte(invoiceID)...)
	key = append(key, byte('/'))
	// Use fixed-width 8-byte encoding for proper ordering
	seqBytes := make([]byte, 8)
	seqBytes[0] = byte(seqNum >> 56)
	seqBytes[1] = byte(seqNum >> 48)
	seqBytes[2] = byte(seqNum >> 40)
	seqBytes[3] = byte(seqNum >> 32)
	seqBytes[4] = byte(seqNum >> 24)
	seqBytes[5] = byte(seqNum >> 16)
	seqBytes[6] = byte(seqNum >> 8)
	seqBytes[7] = byte(seqNum)
	return append(key, seqBytes...)
}

// BuildInvoiceLedgerEntrySeqPrefix builds the prefix for sequence-indexed entries
func BuildInvoiceLedgerEntrySeqPrefix(invoiceID string) []byte {
	key := append(InvoiceLedgerEntrySeqPrefix, []byte(invoiceID)...)
	return append(key, byte('/'))
}
