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
	FiatConversionStateRequested      FiatConversionState = "requested"
	FiatConversionStateSwapping       FiatConversionState = "swapping"
	FiatConversionStateOffRampPending FiatConversionState = "off_ramp_pending"
	FiatConversionStateCompleted      FiatConversionState = "completed"
	FiatConversionStateFailed         FiatConversionState = "failed"
	FiatConversionStateCancelled      FiatConversionState = "cancelled"
)

// IsValid returns true when the state is recognized.
func (s FiatConversionState) IsValid() bool {
	switch s {
	case FiatConversionStateRequested, FiatConversionStateSwapping, FiatConversionStateOffRampPending,
		FiatConversionStateCompleted, FiatConversionStateFailed, FiatConversionStateCancelled:
		return true
	default:
		return false
	}
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

// FiatConversionRecord stores conversion details.
type FiatConversionRecord struct {
	ConversionID        string                      `json:"conversion_id"`
	InvoiceID           string                      `json:"invoice_id,omitempty"`
	SettlementID        string                      `json:"settlement_id,omitempty"`
	PayoutID            string                      `json:"payout_id,omitempty"`
	EscrowID            string                      `json:"escrow_id,omitempty"`
	OrderID             string                      `json:"order_id,omitempty"`
	LeaseID             string                      `json:"lease_id,omitempty"`
	Provider            string                      `json:"provider"`
	Customer            string                      `json:"customer"`
	RequestedBy         string                      `json:"requested_by"`
	RequestedAt         time.Time                   `json:"requested_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	State               FiatConversionState         `json:"state"`
	CryptoToken         TokenSpec                   `json:"crypto_token"`
	StableToken         TokenSpec                   `json:"stable_token"`
	CryptoAmount        sdk.Coin                    `json:"crypto_amount"`
	StableAmount        sdk.Coin                    `json:"stable_amount"`
	FiatCurrency        string                      `json:"fiat_currency"`
	FiatAmount          string                      `json:"fiat_amount"`
	PaymentMethod       string                      `json:"payment_method,omitempty"`
	DestinationRef      string                      `json:"destination_ref,omitempty"`
	DestinationHash     string                      `json:"destination_hash"`
	DestinationRegion   string                      `json:"destination_region,omitempty"`
	SlippageTolerance   float64                     `json:"slippage_tolerance"`
	DexAdapter          string                      `json:"dex_adapter,omitempty"`
	SwapQuoteID         string                      `json:"swap_quote_id,omitempty"`
	SwapTxHash          string                      `json:"swap_tx_hash,omitempty"`
	SwapStatus          string                      `json:"swap_status,omitempty"`
	OffRampProvider     string                      `json:"off_ramp_provider,omitempty"`
	OffRampQuoteID      string                      `json:"off_ramp_quote_id,omitempty"`
	OffRampID           string                      `json:"off_ramp_id,omitempty"`
	OffRampStatus       string                      `json:"off_ramp_status,omitempty"`
	OffRampReference    string                      `json:"off_ramp_reference,omitempty"`
	ComplianceStatus    string                      `json:"compliance_status,omitempty"`
	ComplianceRiskScore int32                       `json:"compliance_risk_score,omitempty"`
	ComplianceCheckedAt int64                       `json:"compliance_checked_at,omitempty"`
	FailureReason       string                      `json:"failure_reason,omitempty"`
	AuditTrail          []FiatConversionAuditEntry  `json:"audit_trail,omitempty"`
	EncryptedPayload    *EncryptedSettlementPayload `json:"encrypted_payload,omitempty"`
}

// Validate validates the conversion record.
func (r *FiatConversionRecord) Validate() error {
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
		State:             FiatConversionStateRequested,
		CryptoToken:       request.CryptoToken,
		StableToken:       request.StableToken,
		CryptoAmount:      payout,
		StableAmount:      sdk.NewCoin(request.StableToken.Denom, sdkmath.ZeroInt()),
		FiatCurrency:      request.FiatCurrency,
		PaymentMethod:     request.PaymentMethod,
		DestinationRef:    encryptedRef,
		DestinationHash:   destinationHash,
		DestinationRegion: request.DestinationRegion,
		SlippageTolerance: request.SlippageTolerance,
		DexAdapter:        request.PreferredDEX,
		OffRampProvider:   request.PreferredOffRamp,
		AuditTrail:        []FiatConversionAuditEntry{},
		EncryptedPayload:  request.EncryptedPayload,
	}
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

// MarkSwapping transitions to swapping state.
func (r *FiatConversionRecord) MarkSwapping(ts time.Time) error {
	if !r.State.IsValid() {
		return ErrInvalidSettlement.Wrap("invalid conversion state")
	}
	r.State = FiatConversionStateSwapping
	r.UpdatedAt = ts
	return nil
}

// MarkOffRampPending transitions to off-ramp pending.
func (r *FiatConversionRecord) MarkOffRampPending(ts time.Time) error {
	r.State = FiatConversionStateOffRampPending
	r.UpdatedAt = ts
	return nil
}

// MarkCompleted transitions to completed.
func (r *FiatConversionRecord) MarkCompleted(ts time.Time) error {
	r.State = FiatConversionStateCompleted
	r.UpdatedAt = ts
	return nil
}

// MarkFailed transitions to failed.
func (r *FiatConversionRecord) MarkFailed(reason string, ts time.Time) error {
	r.State = FiatConversionStateFailed
	r.FailureReason = reason
	r.UpdatedAt = ts
	return nil
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
