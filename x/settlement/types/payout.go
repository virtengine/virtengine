package types

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PayoutState represents the state of a payout
type PayoutState string

const (
	PayoutStatePending    PayoutState = "pending"
	PayoutStateProcessing PayoutState = "processing"
	PayoutStateCompleted  PayoutState = "completed"
	PayoutStateFailed     PayoutState = "failed"
	PayoutStateHeld       PayoutState = "held"     // Held due to dispute
	PayoutStateRefunded   PayoutState = "refunded" // Refunded to customer
	PayoutStateCancelled  PayoutState = "cancelled"
)

// IsValidPayoutState checks if the state is valid
func IsValidPayoutState(state PayoutState) bool {
	switch state {
	case PayoutStatePending, PayoutStateProcessing, PayoutStateCompleted,
		PayoutStateFailed, PayoutStateHeld, PayoutStateRefunded, PayoutStateCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the payout is in a terminal state
func (s PayoutState) IsTerminal() bool {
	switch s {
	case PayoutStateCompleted, PayoutStateFailed, PayoutStateRefunded, PayoutStateCancelled:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if transition to the new state is valid
func (s PayoutState) CanTransitionTo(newState PayoutState) bool {
	switch s {
	case PayoutStatePending:
		return newState == PayoutStateProcessing || newState == PayoutStateHeld ||
			newState == PayoutStateCancelled
	case PayoutStateProcessing:
		return newState == PayoutStateCompleted || newState == PayoutStateFailed ||
			newState == PayoutStateHeld
	case PayoutStateHeld:
		return newState == PayoutStateProcessing || newState == PayoutStateRefunded ||
			newState == PayoutStateCancelled
	case PayoutStateFailed:
		return newState == PayoutStateProcessing // Allow retry
	default:
		return false
	}
}

// PayoutRecord represents a payout execution record
type PayoutRecord struct {
	// PayoutID is the unique identifier for this payout
	PayoutID string `json:"payout_id"`

	// InvoiceID is the linked invoice
	InvoiceID string `json:"invoice_id"`

	// SettlementID is the linked settlement record
	SettlementID string `json:"settlement_id"`

	// EscrowID is the linked escrow account
	EscrowID string `json:"escrow_id"`

	// OrderID is the linked marketplace order
	OrderID string `json:"order_id"`

	// LeaseID is the linked marketplace lease
	LeaseID string `json:"lease_id"`

	// Provider is the payout recipient
	Provider string `json:"provider"`

	// Customer is the original depositor
	Customer string `json:"customer"`

	// GrossAmount is the total amount before fees and holdbacks
	GrossAmount sdk.Coins `json:"gross_amount"`

	// PlatformFee is the platform fee deducted
	PlatformFee sdk.Coins `json:"platform_fee"`

	// ValidatorFee is the validator fee deducted
	ValidatorFee sdk.Coins `json:"validator_fee"`

	// HoldbackAmount is the amount held back (for disputes or guarantees)
	HoldbackAmount sdk.Coins `json:"holdback_amount"`

	// NetAmount is the net amount to the provider (gross - fees - holdback)
	NetAmount sdk.Coins `json:"net_amount"`

	// State is the current payout state
	State PayoutState `json:"state"`

	// DisputeID links to active dispute (if held)
	DisputeID string `json:"dispute_id,omitempty"`

	// HoldReason explains why payout is held
	HoldReason string `json:"hold_reason,omitempty"`

	// IdempotencyKey prevents duplicate execution
	IdempotencyKey string `json:"idempotency_key"`

	// ExecutionAttempts is the number of execution attempts
	ExecutionAttempts uint32 `json:"execution_attempts"`

	// LastAttemptAt is when the last execution was attempted
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastError is the last error message if failed
	LastError string `json:"last_error,omitempty"`

	// TxHash is the on-chain transaction hash when completed
	TxHash string `json:"tx_hash,omitempty"`

	// CreatedAt is when the payout was created
	CreatedAt time.Time `json:"created_at"`

	// ProcessedAt is when the payout was processed
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// CompletedAt is when the payout was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// BlockHeight is when the payout was created
	BlockHeight int64 `json:"block_height"`
}

// NewPayoutRecord creates a new payout record
func NewPayoutRecord(
	payoutID string,
	invoiceID string,
	settlementID string,
	escrowID string,
	orderID string,
	leaseID string,
	provider string,
	customer string,
	grossAmount sdk.Coins,
	platformFee sdk.Coins,
	validatorFee sdk.Coins,
	holdbackAmount sdk.Coins,
	blockTime time.Time,
	blockHeight int64,
) *PayoutRecord {
	// Calculate net amount
	netAmount := grossAmount.Sub(platformFee...).Sub(validatorFee...).Sub(holdbackAmount...)

	return &PayoutRecord{
		PayoutID:          payoutID,
		InvoiceID:         invoiceID,
		SettlementID:      settlementID,
		EscrowID:          escrowID,
		OrderID:           orderID,
		LeaseID:           leaseID,
		Provider:          provider,
		Customer:          customer,
		GrossAmount:       grossAmount,
		PlatformFee:       platformFee,
		ValidatorFee:      validatorFee,
		HoldbackAmount:    holdbackAmount,
		NetAmount:         netAmount,
		State:             PayoutStatePending,
		IdempotencyKey:    fmt.Sprintf("payout-%s-%s", invoiceID, settlementID),
		ExecutionAttempts: 0,
		CreatedAt:         blockTime,
		BlockHeight:       blockHeight,
	}
}

// Validate validates a payout record
func (p *PayoutRecord) Validate() error {
	if p.PayoutID == "" {
		return ErrInvalidPayout.Wrap("payout_id cannot be empty")
	}

	if len(p.PayoutID) > 64 {
		return ErrInvalidPayout.Wrap("payout_id exceeds maximum length")
	}

	if p.InvoiceID == "" && p.SettlementID == "" {
		return ErrInvalidPayout.Wrap("must have invoice_id or settlement_id")
	}

	if _, err := sdk.AccAddressFromBech32(p.Provider); err != nil {
		return ErrInvalidPayout.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(p.Customer); err != nil {
		return ErrInvalidPayout.Wrap("invalid customer address")
	}

	if !p.GrossAmount.IsValid() || p.GrossAmount.IsZero() {
		return ErrInvalidPayout.Wrap("gross_amount must be valid and non-zero")
	}

	if !p.NetAmount.IsValid() {
		return ErrInvalidPayout.Wrap("net_amount must be valid")
	}

	if !IsValidPayoutState(p.State) {
		return ErrInvalidPayout.Wrapf("invalid state: %s", p.State)
	}

	return nil
}

// MarkProcessing transitions to processing state
func (p *PayoutRecord) MarkProcessing(blockTime time.Time) error {
	if !p.State.CanTransitionTo(PayoutStateProcessing) {
		return ErrInvalidStateTransition.Wrapf("cannot transition from %s to processing", p.State)
	}

	p.State = PayoutStateProcessing
	p.ProcessedAt = &blockTime
	p.ExecutionAttempts++
	p.LastAttemptAt = &blockTime
	return nil
}

// MarkCompleted transitions to completed state
func (p *PayoutRecord) MarkCompleted(txHash string, blockTime time.Time) error {
	if !p.State.CanTransitionTo(PayoutStateCompleted) {
		return ErrInvalidStateTransition.Wrapf("cannot transition from %s to completed", p.State)
	}

	p.State = PayoutStateCompleted
	p.TxHash = txHash
	p.CompletedAt = &blockTime
	p.LastError = ""
	return nil
}

// MarkFailed transitions to failed state
func (p *PayoutRecord) MarkFailed(errorMsg string, blockTime time.Time) error {
	if !p.State.CanTransitionTo(PayoutStateFailed) {
		return ErrInvalidStateTransition.Wrapf("cannot transition from %s to failed", p.State)
	}

	p.State = PayoutStateFailed
	p.LastError = errorMsg
	p.LastAttemptAt = &blockTime
	return nil
}

// Hold places a hold on the payout
func (p *PayoutRecord) Hold(disputeID string, reason string, blockTime time.Time) error {
	if !p.State.CanTransitionTo(PayoutStateHeld) {
		return ErrInvalidStateTransition.Wrapf("cannot transition from %s to held", p.State)
	}

	p.State = PayoutStateHeld
	p.DisputeID = disputeID
	p.HoldReason = reason
	return nil
}

// ReleaseHold releases the hold and resumes processing
func (p *PayoutRecord) ReleaseHold() error {
	if p.State != PayoutStateHeld {
		return ErrInvalidStateTransition.Wrap("can only release held payouts")
	}

	p.State = PayoutStatePending
	p.DisputeID = ""
	p.HoldReason = ""
	return nil
}

// Refund transitions to refunded state
func (p *PayoutRecord) Refund(reason string, blockTime time.Time) error {
	if !p.State.CanTransitionTo(PayoutStateRefunded) {
		return ErrInvalidStateTransition.Wrapf("cannot transition from %s to refunded", p.State)
	}

	p.State = PayoutStateRefunded
	p.HoldReason = reason
	return nil
}

// MarshalJSON implements json.Marshaler
func (p PayoutRecord) MarshalJSON() ([]byte, error) {
	type Alias PayoutRecord
	return json.Marshal(&struct {
		Alias
		GrossAmount    []sdk.Coin `json:"gross_amount"`
		PlatformFee    []sdk.Coin `json:"platform_fee"`
		ValidatorFee   []sdk.Coin `json:"validator_fee"`
		HoldbackAmount []sdk.Coin `json:"holdback_amount"`
		NetAmount      []sdk.Coin `json:"net_amount"`
	}{
		Alias:          (Alias)(p),
		GrossAmount:    p.GrossAmount,
		PlatformFee:    p.PlatformFee,
		ValidatorFee:   p.ValidatorFee,
		HoldbackAmount: p.HoldbackAmount,
		NetAmount:      p.NetAmount,
	})
}

// PayoutLedgerEntry represents a state change in the payout ledger
type PayoutLedgerEntry struct {
	// EntryID is the unique identifier for this entry
	EntryID string `json:"entry_id"`

	// PayoutID is the payout this entry relates to
	PayoutID string `json:"payout_id"`

	// EntryType is the type of ledger entry
	EntryType PayoutLedgerEntryType `json:"entry_type"`

	// PreviousState is the state before this change
	PreviousState PayoutState `json:"previous_state"`

	// NewState is the state after this change
	NewState PayoutState `json:"new_state"`

	// Amount is the amount involved in this entry
	Amount sdk.Coins `json:"amount,omitempty"`

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

// PayoutLedgerEntryType defines types of payout ledger entries
type PayoutLedgerEntryType uint8

const (
	PayoutLedgerEntryCreated    PayoutLedgerEntryType = 0
	PayoutLedgerEntryProcessing PayoutLedgerEntryType = 1
	PayoutLedgerEntryCompleted  PayoutLedgerEntryType = 2
	PayoutLedgerEntryFailed     PayoutLedgerEntryType = 3
	PayoutLedgerEntryHeld       PayoutLedgerEntryType = 4
	PayoutLedgerEntryReleased   PayoutLedgerEntryType = 5
	PayoutLedgerEntryRefunded   PayoutLedgerEntryType = 6
	PayoutLedgerEntryCancelled  PayoutLedgerEntryType = 7
)

// PayoutLedgerEntryTypeNames maps types to names
var PayoutLedgerEntryTypeNames = map[PayoutLedgerEntryType]string{
	PayoutLedgerEntryCreated:    "created",
	PayoutLedgerEntryProcessing: "processing",
	PayoutLedgerEntryCompleted:  "completed",
	PayoutLedgerEntryFailed:     "failed",
	PayoutLedgerEntryHeld:       "held",
	PayoutLedgerEntryReleased:   "released",
	PayoutLedgerEntryRefunded:   "refunded",
	PayoutLedgerEntryCancelled:  "cancelled",
}

// String returns string representation
func (t PayoutLedgerEntryType) String() string {
	if name, ok := PayoutLedgerEntryTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// NewPayoutLedgerEntry creates a new payout ledger entry
func NewPayoutLedgerEntry(
	entryID string,
	payoutID string,
	entryType PayoutLedgerEntryType,
	previousState PayoutState,
	newState PayoutState,
	amount sdk.Coins,
	description string,
	initiator string,
	txHash string,
	blockHeight int64,
	timestamp time.Time,
) *PayoutLedgerEntry {
	return &PayoutLedgerEntry{
		EntryID:         entryID,
		PayoutID:        payoutID,
		EntryType:       entryType,
		PreviousState:   previousState,
		NewState:        newState,
		Amount:          amount,
		Description:     description,
		Initiator:       initiator,
		TransactionHash: txHash,
		BlockHeight:     blockHeight,
		Timestamp:       timestamp,
	}
}

// TreasuryRecord tracks platform treasury accounting
type TreasuryRecord struct {
	// RecordID is the unique identifier
	RecordID string `json:"record_id"`

	// RecordType is the type of treasury record
	RecordType TreasuryRecordType `json:"record_type"`

	// PayoutID links to the payout (if applicable)
	PayoutID string `json:"payout_id,omitempty"`

	// SettlementID links to the settlement
	SettlementID string `json:"settlement_id,omitempty"`

	// Amount is the amount credited/debited
	Amount sdk.Coins `json:"amount"`

	// BalanceAfter is the balance after this transaction
	BalanceAfter sdk.Coins `json:"balance_after"`

	// Description describes the transaction
	Description string `json:"description"`

	// BlockHeight is when this record was created
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when this record was created
	Timestamp time.Time `json:"timestamp"`
}

// TreasuryRecordType defines types of treasury records
type TreasuryRecordType uint8

const (
	TreasuryRecordPlatformFee  TreasuryRecordType = 0
	TreasuryRecordValidatorFee TreasuryRecordType = 1
	TreasuryRecordHoldback     TreasuryRecordType = 2
	TreasuryRecordRefund       TreasuryRecordType = 3
	TreasuryRecordWithdrawal   TreasuryRecordType = 4
)

// TreasuryRecordTypeNames maps types to names
var TreasuryRecordTypeNames = map[TreasuryRecordType]string{
	TreasuryRecordPlatformFee:  "platform_fee",
	TreasuryRecordValidatorFee: "validator_fee",
	TreasuryRecordHoldback:     "holdback",
	TreasuryRecordRefund:       "refund",
	TreasuryRecordWithdrawal:   "withdrawal",
}

// String returns string representation
func (t TreasuryRecordType) String() string {
	if name, ok := TreasuryRecordTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// Store key prefixes for payout types
var (
	// PrefixPayout is the prefix for payout record storage
	PrefixPayout = []byte{0x20}

	// PrefixPayoutByInvoice indexes payouts by invoice
	PrefixPayoutByInvoice = []byte{0x21}

	// PrefixPayoutBySettlement indexes payouts by settlement
	PrefixPayoutBySettlement = []byte{0x22}

	// PrefixPayoutByProvider indexes payouts by provider
	PrefixPayoutByProvider = []byte{0x23}

	// PrefixPayoutByState indexes payouts by state
	PrefixPayoutByState = []byte{0x24}

	// PrefixPayoutLedgerEntry stores payout ledger entries
	PrefixPayoutLedgerEntry = []byte{0x25}

	// PrefixPayoutLedgerByPayout indexes ledger entries by payout
	PrefixPayoutLedgerByPayout = []byte{0x26}

	// PrefixTreasuryRecord stores treasury records
	PrefixTreasuryRecord = []byte{0x27}

	// PrefixTreasuryBalance stores treasury balance
	PrefixTreasuryBalance = []byte{0x28}

	// PrefixPayoutIdempotency stores idempotency keys
	PrefixPayoutIdempotency = []byte{0x29}

	// PrefixPayoutSequence stores payout sequence counter
	PrefixPayoutSequence = []byte{0x2A}
)

// PayoutKey returns the store key for a payout record
func PayoutKey(payoutID string) []byte {
	key := make([]byte, 0, len(PrefixPayout)+len(payoutID))
	key = append(key, PrefixPayout...)
	key = append(key, []byte(payoutID)...)
	return key
}

// PayoutByInvoiceKey returns the store key for payout lookup by invoice
func PayoutByInvoiceKey(invoiceID string) []byte {
	key := make([]byte, 0, len(PrefixPayoutByInvoice)+len(invoiceID))
	key = append(key, PrefixPayoutByInvoice...)
	key = append(key, []byte(invoiceID)...)
	return key
}

// PayoutBySettlementKey returns the store key for payout lookup by settlement
func PayoutBySettlementKey(settlementID string) []byte {
	key := make([]byte, 0, len(PrefixPayoutBySettlement)+len(settlementID))
	key = append(key, PrefixPayoutBySettlement...)
	key = append(key, []byte(settlementID)...)
	return key
}

// PayoutByProviderKey returns the store key for payout lookup by provider
func PayoutByProviderKey(provider string, payoutID string) []byte {
	key := make([]byte, 0, len(PrefixPayoutByProvider)+len(provider)+1+len(payoutID))
	key = append(key, PrefixPayoutByProvider...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	key = append(key, []byte(payoutID)...)
	return key
}

// PayoutByProviderPrefixKey returns the prefix for provider's payouts
func PayoutByProviderPrefixKey(provider string) []byte {
	key := make([]byte, 0, len(PrefixPayoutByProvider)+len(provider)+1)
	key = append(key, PrefixPayoutByProvider...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return key
}

// PayoutByStateKey returns the store key for payout lookup by state
func PayoutByStateKey(state PayoutState, payoutID string) []byte {
	stateBytes := []byte(state)
	key := make([]byte, 0, len(PrefixPayoutByState)+len(stateBytes)+1+len(payoutID))
	key = append(key, PrefixPayoutByState...)
	key = append(key, stateBytes...)
	key = append(key, byte('/'))
	key = append(key, []byte(payoutID)...)
	return key
}

// PayoutByStatePrefixKey returns the prefix for payouts by state
func PayoutByStatePrefixKey(state PayoutState) []byte {
	stateBytes := []byte(state)
	key := make([]byte, 0, len(PrefixPayoutByState)+len(stateBytes)+1)
	key = append(key, PrefixPayoutByState...)
	key = append(key, stateBytes...)
	key = append(key, byte('/'))
	return key
}

// PayoutLedgerEntryKey returns the store key for a payout ledger entry
func PayoutLedgerEntryKey(entryID string) []byte {
	key := make([]byte, 0, len(PrefixPayoutLedgerEntry)+len(entryID))
	key = append(key, PrefixPayoutLedgerEntry...)
	key = append(key, []byte(entryID)...)
	return key
}

// PayoutLedgerByPayoutKey returns the index key for ledger entries by payout
func PayoutLedgerByPayoutKey(payoutID string, entryID string) []byte {
	key := make([]byte, 0, len(PrefixPayoutLedgerByPayout)+len(payoutID)+1+len(entryID))
	key = append(key, PrefixPayoutLedgerByPayout...)
	key = append(key, []byte(payoutID)...)
	key = append(key, byte('/'))
	key = append(key, []byte(entryID)...)
	return key
}

// TreasuryRecordKey returns the store key for a treasury record
func TreasuryRecordKey(recordID string) []byte {
	key := make([]byte, 0, len(PrefixTreasuryRecord)+len(recordID))
	key = append(key, PrefixTreasuryRecord...)
	key = append(key, []byte(recordID)...)
	return key
}

// PayoutIdempotencyKey returns the store key for idempotency check
func PayoutIdempotencyKey(idempotencyKey string) []byte {
	key := make([]byte, 0, len(PrefixPayoutIdempotency)+len(idempotencyKey))
	key = append(key, PrefixPayoutIdempotency...)
	key = append(key, []byte(idempotencyKey)...)
	return key
}

// PayoutSequenceKey returns the store key for payout sequence
func PayoutSequenceKey() []byte {
	return PrefixPayoutSequence
}
