package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FiatConversionState represents the state of a fiat conversion.
type FiatConversionState string

const (
	FiatConversionStateCreated         FiatConversionState = "created"
	FiatConversionStateSwapPending     FiatConversionState = "swap_pending"
	FiatConversionStateSwapSubmitted   FiatConversionState = "swap_submitted"
	FiatConversionStateSwapSettled     FiatConversionState = "swap_settled"
	FiatConversionStateOffRampPending  FiatConversionState = "offramp_pending"
	FiatConversionStatePayoutPending   FiatConversionState = "payout_pending"
	FiatConversionStatePayoutSubmitted FiatConversionState = "payout_submitted"
	FiatConversionStatePayoutCompleted FiatConversionState = "payout_completed"
	FiatConversionStateFailed          FiatConversionState = "failed"
	FiatConversionStateCancelled       FiatConversionState = "cancelled"

	// Legacy aliases retained for compatibility with existing data and tests.
	FiatConversionStateRequested FiatConversionState = FiatConversionStateCreated
	FiatConversionStateSwapping  FiatConversionState = FiatConversionStateSwapPending
	FiatConversionStateCompleted FiatConversionState = FiatConversionStatePayoutCompleted
)

var fiatConversionTransitions = map[FiatConversionState]map[FiatConversionState]struct{}{
	FiatConversionStateCreated: {
		FiatConversionStateSwapPending: {},
		FiatConversionStateFailed:      {},
		FiatConversionStateCancelled:   {},
	},
	FiatConversionStateSwapPending: {
		FiatConversionStateSwapSubmitted: {},
		FiatConversionStateFailed:        {},
		FiatConversionStateCancelled:     {},
	},
	FiatConversionStateSwapSubmitted: {
		FiatConversionStateSwapSettled: {},
		FiatConversionStateFailed:      {},
		FiatConversionStateCancelled:   {},
	},
	FiatConversionStateSwapSettled: {
		FiatConversionStateOffRampPending: {},
		FiatConversionStateFailed:         {},
		FiatConversionStateCancelled:      {},
	},
	FiatConversionStateOffRampPending: {
		FiatConversionStatePayoutPending: {},
		FiatConversionStateFailed:        {},
		FiatConversionStateCancelled:     {},
	},
	FiatConversionStatePayoutPending: {
		FiatConversionStatePayoutSubmitted: {},
		FiatConversionStateFailed:          {},
		FiatConversionStateCancelled:       {},
	},
	FiatConversionStatePayoutSubmitted: {
		FiatConversionStatePayoutCompleted: {},
		FiatConversionStateFailed:          {},
		FiatConversionStateCancelled:       {},
	},
	FiatConversionStateFailed: {
		FiatConversionStateSwapPending:   {},
		FiatConversionStatePayoutPending: {},
		FiatConversionStateCancelled:     {},
	},
}

func normalizeFiatConversionState(state FiatConversionState) FiatConversionState {
	switch state {
	case "requested":
		return FiatConversionStateCreated
	case "swapping":
		return FiatConversionStateSwapPending
	case "off_ramp_pending":
		return FiatConversionStateOffRampPending
	case "completed":
		return FiatConversionStatePayoutCompleted
	default:
		return state
	}
}

// IsValid returns true when the state is recognized.
func (s FiatConversionState) IsValid() bool {
	switch normalizeFiatConversionState(s) {
	case FiatConversionStateCreated, FiatConversionStateSwapPending, FiatConversionStateSwapSubmitted,
		FiatConversionStateSwapSettled, FiatConversionStateOffRampPending, FiatConversionStatePayoutPending,
		FiatConversionStatePayoutSubmitted, FiatConversionStatePayoutCompleted, FiatConversionStateFailed,
		FiatConversionStateCancelled:
		return true
	}
	return false
}

// IsTerminal returns true when no further transitions are allowed.
func (s FiatConversionState) IsTerminal() bool {
	switch normalizeFiatConversionState(s) {
	case FiatConversionStatePayoutCompleted, FiatConversionStateFailed, FiatConversionStateCancelled:
		return true
	}
	return false
}

// CanTransitionTo returns true when a state transition is legal.
func (s FiatConversionState) CanTransitionTo(target FiatConversionState) bool {
	current := normalizeFiatConversionState(s)
	next := normalizeFiatConversionState(target)
	if current == next {
		return true
	}
	allowed, ok := fiatConversionTransitions[current]
	if !ok {
		return false
	}
	_, exists := allowed[next]
	return exists
}

// TokenSpec captures token metadata for swaps.
type TokenSpec struct {
	Symbol   string `json:"symbol"`
	Denom    string `json:"denom"`
	Decimals uint8  `json:"decimals"`
	ChainID  string `json:"chain_id,omitempty"`
}

// Validate validates the token spec.
func (t TokenSpec) Validate() error {
	if t.Symbol == "" || t.Denom == "" {
		return ErrInvalidParams.Wrap("token spec requires symbol and denom")
	}
	return nil
}

// FiatPayoutPreference configures provider fiat conversion preferences.
type FiatPayoutPreference struct {
	Provider          string                      `json:"provider"`
	Enabled           bool                        `json:"enabled"`
	FiatCurrency      string                      `json:"fiat_currency"`
	PaymentMethod     string                      `json:"payment_method,omitempty"`
	DestinationRef    string                      `json:"destination_ref,omitempty"`
	DestinationHash   string                      `json:"destination_hash"`
	DestinationRegion string                      `json:"destination_region,omitempty"`
	PreferredDEX      string                      `json:"preferred_dex,omitempty"`
	PreferredOffRamp  string                      `json:"preferred_off_ramp,omitempty"`
	SlippageTolerance float64                     `json:"slippage_tolerance"`
	CryptoToken       TokenSpec                   `json:"crypto_token"`
	StableToken       TokenSpec                   `json:"stable_token"`
	EncryptedPayload  *EncryptedSettlementPayload `json:"encrypted_payload,omitempty"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         time.Time                   `json:"updated_at"`
}

// FiatConversionRequest captures a conversion request.
type FiatConversionRequest struct {
	InvoiceID         string                      `json:"invoice_id,omitempty"`
	SettlementID      string                      `json:"settlement_id,omitempty"`
	PayoutID          string                      `json:"payout_id,omitempty"`
	Provider          string                      `json:"provider"`
	Customer          string                      `json:"customer"`
	RequestedBy       string                      `json:"requested_by"`
	CryptoAmount      sdk.Coin                    `json:"crypto_amount"`
	FiatCurrency      string                      `json:"fiat_currency"`
	PaymentMethod     string                      `json:"payment_method,omitempty"`
	Destination       string                      `json:"destination,omitempty"`
	DestinationHash   string                      `json:"destination_hash,omitempty"`
	DestinationRegion string                      `json:"destination_region,omitempty"`
	PreferredDEX      string                      `json:"preferred_dex,omitempty"`
	PreferredOffRamp  string                      `json:"preferred_off_ramp,omitempty"`
	SlippageTolerance float64                     `json:"slippage_tolerance"`
	CryptoToken       TokenSpec                   `json:"crypto_token"`
	StableToken       TokenSpec                   `json:"stable_token"`
	EncryptedPayload  *EncryptedSettlementPayload `json:"encrypted_payload,omitempty"`
}

// FiatConversionAuditEntry is an audit log entry for conversions.
type FiatConversionAuditEntry struct {
	Action    string            `json:"action"`
	Actor     string            `json:"actor"`
	Reason    string            `json:"reason,omitempty"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// FiatConversionStateTransition records a state transition for durable observability.
type FiatConversionStateTransition struct {
	From      FiatConversionState `json:"from"`
	To        FiatConversionState `json:"to"`
	Event     string              `json:"event"`
	Reason    string              `json:"reason,omitempty"`
	Timestamp int64               `json:"timestamp"`
	Metadata  map[string]string   `json:"metadata,omitempty"`
}

// FiatConversionRecord stores conversion details.
type FiatConversionRecord struct {
	ConversionID        string                          `json:"conversion_id"`
	InvoiceID           string                          `json:"invoice_id,omitempty"`
	SettlementID        string                          `json:"settlement_id,omitempty"`
	PayoutID            string                          `json:"payout_id,omitempty"`
	EscrowID            string                          `json:"escrow_id,omitempty"`
	OrderID             string                          `json:"order_id,omitempty"`
	LeaseID             string                          `json:"lease_id,omitempty"`
	Provider            string                          `json:"provider"`
	Customer            string                          `json:"customer"`
	RequestedBy         string                          `json:"requested_by"`
	RequestedAt         time.Time                       `json:"requested_at"`
	UpdatedAt           time.Time                       `json:"updated_at"`
	State               FiatConversionState             `json:"state"`
	CryptoToken         TokenSpec                       `json:"crypto_token"`
	StableToken         TokenSpec                       `json:"stable_token"`
	CryptoAmount        sdk.Coin                        `json:"crypto_amount"`
	StableAmount        sdk.Coin                        `json:"stable_amount"`
	FiatCurrency        string                          `json:"fiat_currency"`
	FiatAmount          string                          `json:"fiat_amount"`
	IdempotencyKey      string                          `json:"idempotency_key"`
	PaymentMethod       string                          `json:"payment_method,omitempty"`
	DestinationRef      string                          `json:"destination_ref,omitempty"`
	DestinationHash     string                          `json:"destination_hash"`
	DestinationRegion   string                          `json:"destination_region,omitempty"`
	SlippageTolerance   float64                         `json:"slippage_tolerance"`
	SwapAttempts        uint32                          `json:"swap_attempts"`
	OffRampAttempts     uint32                          `json:"off_ramp_attempts"`
	PayoutAttempts      uint32                          `json:"payout_attempts"`
	DexAdapter          string                          `json:"dex_adapter,omitempty"`
	SwapQuoteID         string                          `json:"swap_quote_id,omitempty"`
	SwapTxHash          string                          `json:"swap_tx_hash,omitempty"`
	SwapStatus          string                          `json:"swap_status,omitempty"`
	OffRampProvider     string                          `json:"off_ramp_provider,omitempty"`
	OffRampQuoteID      string                          `json:"off_ramp_quote_id,omitempty"`
	OffRampID           string                          `json:"off_ramp_id,omitempty"`
	OffRampStatus       string                          `json:"off_ramp_status,omitempty"`
	OffRampReference    string                          `json:"off_ramp_reference,omitempty"`
	ComplianceStatus    string                          `json:"compliance_status,omitempty"`
	ComplianceRiskScore int32                           `json:"compliance_risk_score,omitempty"`
	ComplianceCheckedAt int64                           `json:"compliance_checked_at,omitempty"`
	FailureReason       string                          `json:"failure_reason,omitempty"`
	LastErrorAt         int64                           `json:"last_error_at,omitempty"`
	LastError           string                          `json:"last_error,omitempty"`
	AuditTrail          []FiatConversionAuditEntry      `json:"audit_trail,omitempty"`
	TransitionHistory   []FiatConversionStateTransition `json:"transition_history,omitempty"`
	EncryptedPayload    *EncryptedSettlementPayload     `json:"encrypted_payload,omitempty"`
}

// Validate validates the conversion record.
func (r *FiatConversionRecord) Validate() error {
	r.State = normalizeFiatConversionState(r.State)
	if r.ConversionID == "" {
		return ErrInvalidSettlement.Wrap("conversion_id cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return ErrInvalidAddress.Wrap("invalid provider address")
	}
	if _, err := sdk.AccAddressFromBech32(r.Customer); err != nil {
		return ErrInvalidAddress.Wrap("invalid customer address")
	}
	if !r.State.IsValid() {
		return ErrInvalidSettlement.Wrapf("invalid conversion state: %s", r.State)
	}
	if r.IdempotencyKey == "" {
		r.IdempotencyKey = r.DefaultIdempotencyKey()
	}
	if !r.CryptoAmount.IsValid() || !r.CryptoAmount.IsPositive() {
		return ErrInvalidAmount.Wrap("crypto_amount must be positive")
	}
	if r.CryptoAmount.Denom != "" && r.CryptoToken.Denom != "" && r.CryptoAmount.Denom != r.CryptoToken.Denom {
		return ErrInvalidAmount.Wrap("crypto_amount denom must match crypto_token")
	}
	if r.FiatCurrency == "" {
		return ErrInvalidParams.Wrap("fiat_currency required")
	}
	if r.SlippageTolerance < 0 || r.SlippageTolerance > 1 {
		return ErrInvalidParams.Wrap("slippage_tolerance must be between 0 and 1")
	}
	if err := r.CryptoToken.Validate(); err != nil {
		return err
	}
	if err := r.StableToken.Validate(); err != nil {
		return err
	}
	if r.EncryptedPayload != nil {
		if err := r.EncryptedPayload.Validate(); err != nil {
			return ErrInvalidParams.Wrapf("invalid encrypted payload: %v", err)
		}
		if r.DestinationHash == "" {
			return ErrInvalidParams.Wrap("destination_hash required")
		}
	}
	if r.DestinationRef != "" {
		if r.EncryptedPayload == nil || r.DestinationRef != r.EncryptedPayload.EnvelopeRef {
			return ErrInvalidParams.Wrap("plaintext conversion fields are not allowed")
		}
	}
	return nil
}

// HashDestination hashes a destination string to avoid storing raw PII.
func HashDestination(destination string) string {
	if destination == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(destination))
	return hex.EncodeToString(sum[:])
}

// NewFiatConversionRecord builds a conversion record from a request.
func NewFiatConversionRecord(id string, request FiatConversionRequest, payout sdk.Coin, now time.Time) *FiatConversionRecord {
	destinationHash := request.DestinationHash
	if destinationHash == "" && request.Destination != "" {
		destinationHash = HashDestination(request.Destination)
	}
	encryptedRef := ""
	if request.EncryptedPayload != nil {
		encryptedRef = request.EncryptedPayload.EnvelopeRef
	}

	return &FiatConversionRecord{
		ConversionID:      id,
		InvoiceID:         request.InvoiceID,
		SettlementID:      request.SettlementID,
		PayoutID:          request.PayoutID,
		Provider:          request.Provider,
		Customer:          request.Customer,
		RequestedBy:       request.RequestedBy,
		RequestedAt:       now,
		UpdatedAt:         now,
		State:             FiatConversionStateCreated,
		CryptoToken:       request.CryptoToken,
		StableToken:       request.StableToken,
		CryptoAmount:      payout,
		StableAmount:      sdk.NewCoin(request.StableToken.Denom, sdkmath.ZeroInt()),
		FiatCurrency:      request.FiatCurrency,
		IdempotencyKey:    defaultFiatConversionIdempotencyKey(request.InvoiceID, request.SettlementID, request.PayoutID, request.Provider),
		PaymentMethod:     request.PaymentMethod,
		DestinationRef:    encryptedRef,
		DestinationHash:   destinationHash,
		DestinationRegion: request.DestinationRegion,
		SlippageTolerance: request.SlippageTolerance,
		DexAdapter:        request.PreferredDEX,
		OffRampProvider:   request.PreferredOffRamp,
		AuditTrail:        []FiatConversionAuditEntry{},
		TransitionHistory: []FiatConversionStateTransition{},
		EncryptedPayload:  request.EncryptedPayload,
	}
}

func defaultFiatConversionIdempotencyKey(invoiceID, settlementID, payoutID, provider string) string {
	return fmt.Sprintf("fiatconv:%s:%s:%s:%s", invoiceID, settlementID, payoutID, provider)
}

// DefaultIdempotencyKey computes the canonical idempotency key.
func (r *FiatConversionRecord) DefaultIdempotencyKey() string {
	return defaultFiatConversionIdempotencyKey(r.InvoiceID, r.SettlementID, r.PayoutID, r.Provider)
}

// AddAuditEntry appends an audit entry.
func (r *FiatConversionRecord) AddAuditEntry(action, actor, reason string, metadata map[string]string, ts time.Time) {
	r.AuditTrail = append(r.AuditTrail, FiatConversionAuditEntry{
		Action:    action,
		Actor:     actor,
		Reason:    reason,
		Timestamp: ts.Unix(),
		Metadata:  metadata,
	})
	r.UpdatedAt = ts
}

// TransitionTo enforces legal state transitions and records transition history.
func (r *FiatConversionRecord) TransitionTo(next FiatConversionState, event string, reason string, metadata map[string]string, ts time.Time) error {
	current := normalizeFiatConversionState(r.State)
	target := normalizeFiatConversionState(next)
	if !current.CanTransitionTo(target) {
		return ErrInvalidStateTransition.Wrapf("invalid fiat conversion transition: %s -> %s", current, target)
	}
	if current == target {
		r.UpdatedAt = ts
		return nil
	}

	r.State = target
	r.UpdatedAt = ts
	r.TransitionHistory = append(r.TransitionHistory, FiatConversionStateTransition{
		From:      current,
		To:        target,
		Event:     event,
		Reason:    reason,
		Timestamp: ts.Unix(),
		Metadata:  metadata,
	})
	auditMeta := map[string]string{
		"from":  string(current),
		"to":    string(target),
		"event": event,
	}
	for key, value := range metadata {
		auditMeta[key] = value
	}
	r.AuditTrail = append(r.AuditTrail, FiatConversionAuditEntry{
		Action:    "state_transition",
		Actor:     "system",
		Reason:    reason,
		Timestamp: ts.Unix(),
		Metadata:  auditMeta,
	})

	switch target {
	case FiatConversionStateFailed:
		r.FailureReason = reason
		r.LastError = reason
		r.LastErrorAt = ts.Unix()
	case FiatConversionStatePayoutCompleted:
		r.FailureReason = ""
		r.LastError = ""
		r.LastErrorAt = 0
	}

	return nil
}

// MarkSwapping transitions to swapping state.
func (r *FiatConversionRecord) MarkSwapping(ts time.Time) error {
	return r.MarkSwapPending(ts)
}

// MarkSwapPending transitions to swap_pending.
func (r *FiatConversionRecord) MarkSwapPending(ts time.Time) error {
	r.SwapAttempts++
	return r.TransitionTo(FiatConversionStateSwapPending, "swap_pending", "", nil, ts)
}

// MarkSwapSubmitted transitions to swap_submitted.
func (r *FiatConversionRecord) MarkSwapSubmitted(quoteID string, ts time.Time) error {
	if quoteID != "" {
		r.SwapQuoteID = quoteID
	}
	r.SwapStatus = "submitted"
	return r.TransitionTo(FiatConversionStateSwapSubmitted, "swap_submitted", "", map[string]string{
		"swap_quote_id": quoteID,
	}, ts)
}

// MarkSwapSettled transitions to swap_settled.
func (r *FiatConversionRecord) MarkSwapSettled(txHash string, stableAmount sdk.Coin, ts time.Time) error {
	if txHash != "" {
		r.SwapTxHash = txHash
	}
	if stableAmount.IsValid() && !stableAmount.Amount.IsNil() {
		r.StableAmount = stableAmount
	}
	r.SwapStatus = "settled"
	return r.TransitionTo(FiatConversionStateSwapSettled, "swap_settled", "", map[string]string{
		"swap_tx_hash": txHash,
	}, ts)
}

// MarkOffRampPending transitions to off-ramp pending.
func (r *FiatConversionRecord) MarkOffRampPending(ts time.Time) error {
	return r.TransitionTo(FiatConversionStateOffRampPending, "offramp_pending", "", nil, ts)
}

// MarkPayoutPending transitions to payout_pending.
func (r *FiatConversionRecord) MarkPayoutPending(ts time.Time) error {
	r.OffRampAttempts++
	return r.TransitionTo(FiatConversionStatePayoutPending, "payout_pending", "", nil, ts)
}

// MarkPayoutSubmitted transitions to payout_submitted.
func (r *FiatConversionRecord) MarkPayoutSubmitted(offRampID string, offRampStatus string, reference string, ts time.Time) error {
	r.PayoutAttempts++
	if offRampID != "" {
		r.OffRampID = offRampID
	}
	if offRampStatus != "" {
		r.OffRampStatus = offRampStatus
	}
	if reference != "" {
		r.OffRampReference = reference
	}
	return r.TransitionTo(FiatConversionStatePayoutSubmitted, "payout_submitted", "", map[string]string{
		"offramp_id": offRampID,
		"status":     offRampStatus,
	}, ts)
}

// MarkCompleted transitions to completed.
func (r *FiatConversionRecord) MarkCompleted(ts time.Time) error {
	return r.MarkPayoutCompleted(ts)
}

// MarkPayoutCompleted transitions to payout_completed.
func (r *FiatConversionRecord) MarkPayoutCompleted(ts time.Time) error {
	return r.TransitionTo(FiatConversionStatePayoutCompleted, "payout_completed", "", nil, ts)
}

// MarkFailed transitions to failed.
func (r *FiatConversionRecord) MarkFailed(reason string, ts time.Time) error {
	return r.TransitionTo(FiatConversionStateFailed, "failed", reason, nil, ts)
}

// MarkCancelled transitions to cancelled.
func (r *FiatConversionRecord) MarkCancelled(reason string, ts time.Time) error {
	return r.TransitionTo(FiatConversionStateCancelled, "cancelled", reason, nil, ts)
}

// ValidatePreference validates payout preference.
func (p *FiatPayoutPreference) Validate() error {
	if _, err := sdk.AccAddressFromBech32(p.Provider); err != nil {
		return ErrInvalidAddress.Wrap("invalid provider address")
	}
	if p.Enabled {
		if p.FiatCurrency == "" || p.PaymentMethod == "" {
			return ErrInvalidParams.Wrap("fiat_currency and payment_method required")
		}
		if p.EncryptedPayload == nil {
			return ErrInvalidParams.Wrap("encrypted_payload required")
		}
		if err := p.EncryptedPayload.Validate(); err != nil {
			return ErrInvalidParams.Wrapf("invalid encrypted payload: %v", err)
		}
		if p.DestinationHash == "" {
			return ErrInvalidParams.Wrap("destination_hash required")
		}
		if err := p.CryptoToken.Validate(); err != nil {
			return err
		}
		if err := p.StableToken.Validate(); err != nil {
			return err
		}
		if p.SlippageTolerance < 0 || p.SlippageTolerance > 1 {
			return ErrInvalidParams.Wrap("slippage_tolerance must be between 0 and 1")
		}
	}
	if p.DestinationRef != "" {
		if p.EncryptedPayload == nil || p.DestinationRef != p.EncryptedPayload.EnvelopeRef {
			return ErrInvalidParams.Wrap("plaintext payout fields are not allowed")
		}
	}
	return nil
}

// ValidateRequest validates a conversion request.
func (r *FiatConversionRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.Provider); err != nil {
		return ErrInvalidAddress.Wrap("invalid provider address")
	}
	if _, err := sdk.AccAddressFromBech32(r.Customer); err != nil {
		return ErrInvalidAddress.Wrap("invalid customer address")
	}
	if r.FiatCurrency == "" || r.PaymentMethod == "" {
		return ErrInvalidParams.Wrap("fiat_currency and payment_method required")
	}
	if r.EncryptedPayload == nil {
		return ErrInvalidParams.Wrap("encrypted_payload required")
	}
	if err := r.EncryptedPayload.Validate(); err != nil {
		return ErrInvalidParams.Wrapf("invalid encrypted payload: %v", err)
	}
	if r.DestinationHash == "" && r.Destination == "" {
		return ErrInvalidParams.Wrap("destination_hash required")
	}
	if !r.CryptoAmount.IsValid() || !r.CryptoAmount.IsPositive() {
		return ErrInvalidAmount.Wrap("crypto_amount must be positive")
	}
	if r.CryptoAmount.Denom != "" && r.CryptoToken.Denom != "" && r.CryptoAmount.Denom != r.CryptoToken.Denom {
		return ErrInvalidAmount.Wrap("crypto_amount denom must match crypto_token")
	}
	if err := r.CryptoToken.Validate(); err != nil {
		return err
	}
	if err := r.StableToken.Validate(); err != nil {
		return err
	}
	if r.SlippageTolerance < 0 || r.SlippageTolerance > 1 {
		return ErrInvalidParams.Wrap("slippage_tolerance must be between 0 and 1")
	}
	if r.Destination != "" {
		return ErrInvalidParams.Wrap("plaintext conversion fields are not allowed")
	}
	return nil
}

// FormatComplianceSnapshot formats compliance summary.
func FormatComplianceSnapshot(status string, riskScore int32) string {
	return fmt.Sprintf("%s/%d", status, riskScore)
}
