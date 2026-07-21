package types

import (
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SettlementRecord represents a completed settlement for an order
type SettlementRecord struct {
	// SettlementID is the unique identifier for this settlement
	SettlementID string `json:"settlement_id"`

	// EscrowID is the linked escrow account
	EscrowID string `json:"escrow_id"`

	// OrderID is the linked marketplace order
	OrderID string `json:"order_id"`

	// LeaseID is the linked marketplace lease
	LeaseID string `json:"lease_id"`

	// Provider is the service provider address
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// TotalAmount is the total amount settled
	TotalAmount sdk.Coins `json:"total_amount"`

	// ProviderShare is the amount going to the provider
	ProviderShare sdk.Coins `json:"provider_share"`

	// PlatformFee is the platform fee taken
	PlatformFee sdk.Coins `json:"platform_fee"`

	// ValidatorFee is the fee distributed to validators
	ValidatorFee sdk.Coins `json:"validator_fee,omitempty"`

	// SettledAt is when the settlement occurred
	SettledAt time.Time `json:"settled_at"`

	// UsageRecordIDs are the linked usage records
	UsageRecordIDs []string `json:"usage_record_ids,omitempty"`

	// TotalUsageUnits is the total usage units in this settlement
	TotalUsageUnits uint64 `json:"total_usage_units"`

	// PeriodStart is the start of the billing period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the billing period
	PeriodEnd time.Time `json:"period_end"`

	// BlockHeight is when the settlement was recorded
	BlockHeight int64 `json:"block_height"`

	// SettlementType indicates the type of settlement
	SettlementType SettlementType `json:"settlement_type"`

	// IsFinal indicates if this is the final settlement for the order
	IsFinal bool `json:"is_final"`
}

// SettlementType defines the type of settlement
type SettlementType string

const (
	SettlementTypePeriodic   SettlementType = "periodic"
	SettlementTypeUsageBased SettlementType = "usage_based"
	SettlementTypeFinal      SettlementType = "final"
	SettlementTypeRefund     SettlementType = "refund"
)

// IsValidSettlementType checks if the type is valid
func IsValidSettlementType(t SettlementType) bool {
	switch t {
	case SettlementTypePeriodic, SettlementTypeUsageBased, SettlementTypeFinal, SettlementTypeRefund:
		return true
	default:
		return false
	}
}

// NewSettlementRecord creates a new settlement record
func NewSettlementRecord(
	settlementID string,
	escrowID string,
	orderID string,
	leaseID string,
	provider string,
	customer string,
	totalAmount sdk.Coins,
	providerShare sdk.Coins,
	platformFee sdk.Coins,
	validatorFee sdk.Coins,
	usageRecordIDs []string,
	totalUsageUnits uint64,
	periodStart time.Time,
	periodEnd time.Time,
	settlementType SettlementType,
	isFinal bool,
	blockTime time.Time,
	blockHeight int64,
) *SettlementRecord {
	return &SettlementRecord{
		SettlementID:    settlementID,
		EscrowID:        escrowID,
		OrderID:         orderID,
		LeaseID:         leaseID,
		Provider:        provider,
		Customer:        customer,
		TotalAmount:     totalAmount,
		ProviderShare:   providerShare,
		PlatformFee:     platformFee,
		ValidatorFee:    validatorFee,
		SettledAt:       blockTime,
		UsageRecordIDs:  usageRecordIDs,
		TotalUsageUnits: totalUsageUnits,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		BlockHeight:     blockHeight,
		SettlementType:  settlementType,
		IsFinal:         isFinal,
	}
}

// Validate validates a settlement record
func (s *SettlementRecord) Validate() error {
	if s.SettlementID == "" {
		return ErrInvalidSettlement.Wrap("settlement_id cannot be empty")
	}

	if len(s.SettlementID) > 64 {
		return ErrInvalidSettlement.Wrap("settlement_id exceeds maximum length")
	}

	if s.EscrowID == "" {
		return ErrInvalidSettlement.Wrap("escrow_id cannot be empty")
	}

	if s.OrderID == "" {
		return ErrInvalidSettlement.Wrap("order_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(s.Provider); err != nil {
		return ErrInvalidSettlement.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(s.Customer); err != nil {
		return ErrInvalidSettlement.Wrap("invalid customer address")
	}

	if !s.TotalAmount.IsValid() || s.TotalAmount.IsZero() {
		return ErrInvalidSettlement.Wrap("total_amount must be valid and non-zero")
	}

	if !s.ProviderShare.IsValid() {
		return ErrInvalidSettlement.Wrap("provider_share must be valid")
	}

	if !s.PlatformFee.IsValid() {
		return ErrInvalidSettlement.Wrap("platform_fee must be valid")
	}

	if !IsValidSettlementType(s.SettlementType) {
		return ErrInvalidSettlement.Wrapf("invalid settlement type: %s", s.SettlementType)
	}

	// Validate that shares sum to total
	total := s.ProviderShare.Add(s.PlatformFee...)
	if s.ValidatorFee != nil && !s.ValidatorFee.IsZero() {
		total = total.Add(s.ValidatorFee...)
	}
	if !total.Equal(s.TotalAmount) {
		return ErrInvalidSettlement.Wrap("provider_share + platform_fee + validator_fee must equal total_amount")
	}

	return nil
}

// UsageRecord represents a signed usage record from a provider
type UsageRecord struct {
	// ExactDuplicate is a transient response marker and is never persisted.
	ExactDuplicate bool `json:"-"`

	// UsageID is the unique identifier for this usage record
	UsageID string `json:"usage_id"`

	// OrderID is the linked marketplace order
	OrderID string `json:"order_id"`

	// LeaseID is the linked marketplace lease
	LeaseID string `json:"lease_id"`

	// AllocationID is the canonical allocation lineage when available.
	AllocationID string `json:"allocation_id,omitempty"`

	// ChainID is the chain domain authenticated by the provider.
	ChainID string `json:"chain_id,omitempty"`

	// Provider is the provider that submitted this record
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// UsageUnits is the number of usage units consumed
	UsageUnits uint64 `json:"usage_units"`

	// UsageType describes the type of usage (compute, storage, bandwidth, etc.)
	UsageType string `json:"usage_type"`

	// PeriodStart is the start of the usage period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the usage period
	PeriodEnd time.Time `json:"period_end"`

	// UnitPrice is the price per usage unit
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// Metrics are exact raw integer deltas for the signed period.
	Metrics RawUsageMetrics `json:"metrics"`

	PricingVersion uint32 `json:"pricing_version,omitempty"`
	FormulaVersion uint32 `json:"formula_version,omitempty"`
	ModelVersion   uint32 `json:"model_version,omitempty"`
	Sequence       uint64 `json:"stream_sequence,omitempty"`
	Nonce          []byte `json:"nonce,omitempty"`
	IdempotencyKey []byte `json:"idempotency_key,omitempty"`

	ProviderKeyEpoch uint64 `json:"provider_key_epoch,omitempty"`
	ProviderKeyID    string `json:"provider_key_id,omitempty"`
	IssuedAtHeight   int64  `json:"issued_at_height,omitempty"`
	ExpiresAtHeight  int64  `json:"expires_at_height,omitempty"`
	IssuedAtUnix     int64  `json:"issued_at_unix,omitempty"`
	ExpiresAtUnix    int64  `json:"expires_at_unix,omitempty"`
	SignatureVersion uint32 `json:"signature_version,omitempty"`

	// TotalCost is the total cost for this usage
	TotalCost sdk.Coins `json:"total_cost"`

	// Signature is the provider's signature on this record
	ProviderSignature []byte `json:"provider_signature"`

	// CustomerAcknowledged indicates if customer acknowledged this usage
	CustomerAcknowledged bool `json:"customer_acknowledged"`

	// CustomerSignature is the customer's acknowledgment signature (optional)
	CustomerSignature []byte `json:"customer_signature,omitempty"`

	// UsageDigest is SHA-256 over the exact canonical provider sign bytes.
	UsageDigest []byte `json:"usage_digest,omitempty"`

	// SignatureVerified and AuthenticationStatus make verification explicit.
	SignatureVerified    bool   `json:"signature_verified"`
	AuthenticationStatus string `json:"authentication_status,omitempty"`
	LegacyUnverified     bool   `json:"legacy_unverified"`

	CustomerAckReplayKey        []byte `json:"customer_ack_replay_key,omitempty"`
	CustomerAckIssuedAtHeight   int64  `json:"customer_ack_issued_at_height,omitempty"`
	CustomerAckExpiresAtHeight  int64  `json:"customer_ack_expires_at_height,omitempty"`
	CustomerAckIssuedAtUnix     int64  `json:"customer_ack_issued_at_unix,omitempty"`
	CustomerAckExpiresAtUnix    int64  `json:"customer_ack_expires_at_unix,omitempty"`
	CustomerAckSignatureVersion uint32 `json:"customer_ack_signature_version,omitempty"`

	// Settled indicates if this usage has been settled
	Settled bool `json:"settled"`

	// SettlementID is the settlement that included this usage (if settled)
	SettlementID string `json:"settlement_id,omitempty"`

	// SubmittedAt is when this record was submitted
	SubmittedAt time.Time `json:"submitted_at"`

	// BlockHeight is when the record was submitted
	BlockHeight int64 `json:"block_height"`

	// Metadata contains additional usage details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewUsageRecord creates a new usage record
func NewUsageRecord(
	usageID string,
	orderID string,
	leaseID string,
	provider string,
	customer string,
	usageUnits uint64,
	usageType string,
	periodStart time.Time,
	periodEnd time.Time,
	unitPrice sdk.DecCoin,
	providerSignature []byte,
	blockTime time.Time,
	blockHeight int64,
) *UsageRecord {
	// Calculate total cost
	totalCost := sdk.NewCoin(
		unitPrice.Denom,
		unitPrice.Amount.MulInt64(safeInt64FromUint64(usageUnits)).TruncateInt(),
	)

	return &UsageRecord{
		UsageID:           usageID,
		OrderID:           orderID,
		LeaseID:           leaseID,
		Provider:          provider,
		Customer:          customer,
		UsageUnits:        usageUnits,
		UsageType:         usageType,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		UnitPrice:         unitPrice,
		TotalCost:         sdk.NewCoins(totalCost),
		ProviderSignature: providerSignature,
		SubmittedAt:       blockTime,
		BlockHeight:       blockHeight,
		Metadata:          make(map[string]string),
	}
}

func safeInt64FromUint64(value uint64) int64 {
	if value > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1)
	}
	return int64(value)
}

// Validate validates a usage record
func (u *UsageRecord) Validate() error {
	if u.UsageID == "" {
		return ErrInvalidUsageRecord.Wrap("usage_id cannot be empty")
	}

	if u.OrderID == "" {
		return ErrInvalidUsageRecord.Wrap("order_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(u.Provider); err != nil {
		return ErrInvalidUsageRecord.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(u.Customer); err != nil {
		return ErrInvalidUsageRecord.Wrap("invalid customer address")
	}

	if u.UsageUnits == 0 {
		return ErrInvalidUsageRecord.Wrap("usage_units must be greater than zero")
	}

	if u.UsageType == "" {
		return ErrInvalidUsageRecord.Wrap("usage_type cannot be empty")
	}

	if u.PeriodEnd.Before(u.PeriodStart) {
		return ErrInvalidUsageRecord.Wrap("period_end must be after period_start")
	}

	if len(u.ProviderSignature) == 0 {
		return ErrInvalidUsageRecord.Wrap("provider_signature cannot be empty")
	}

	if u.SignatureVersion != 0 {
		if _, err := CanonicalUsageSignBytes(u.CanonicalUsagePayload(u.ChainID)); err != nil {
			return ErrInvalidUsageRecord.Wrap(err.Error())
		}
		if u.SignatureVerified && len(u.UsageDigest) != DigestSize {
			return ErrInvalidUsageRecord.Wrap("verified usage_digest must be 32 bytes")
		}
	}

	return nil
}

const (
	UsageAuthenticationStatusVerified = "verified_v1"
	UsageAuthenticationStatusLegacy   = "legacy_unverified"
)

// CanonicalUsagePayload projects persisted state into the canonical envelope.
func (u UsageRecord) CanonicalUsagePayload(chainID string) CanonicalUsagePayload {
	if chainID == "" {
		chainID = u.ChainID
	}
	return CanonicalUsagePayload{
		SignatureVersion: u.SignatureVersion,
		ChainID:          chainID,
		Domain:           UsageProviderDomainV1,
		SignerRole:       SignerRoleProvider,
		Provider:         u.Provider,
		Customer:         u.Customer,
		OrderID:          u.OrderID,
		LeaseID:          u.LeaseID,
		AllocationID:     u.AllocationID,
		PeriodStart:      u.PeriodStart.Unix(),
		PeriodEnd:        u.PeriodEnd.Unix(),
		Metrics:          u.Metrics,
		PricingVersion:   u.PricingVersion,
		UsageUnits:       u.UsageUnits,
		UsageType:        u.UsageType,
		UnitPriceDenom:   u.UnitPrice.Denom,
		UnitPriceAmount:  u.UnitPrice.Amount.String(),
		FormulaVersion:   u.FormulaVersion,
		ModelVersion:     u.ModelVersion,
		Sequence:         u.Sequence,
		Nonce:            u.Nonce,
		IdempotencyKey:   u.IdempotencyKey,
		ProviderKeyEpoch: u.ProviderKeyEpoch,
		ProviderKeyID:    u.ProviderKeyID,
		IssuedAtHeight:   u.IssuedAtHeight,
		ExpiresAtHeight:  u.ExpiresAtHeight,
		IssuedAtUnix:     u.IssuedAtUnix,
		ExpiresAtUnix:    u.ExpiresAtUnix,
	}
}

// IsAuthenticated reports whether the record can trigger value-bearing effects.
func (u UsageRecord) IsAuthenticated() bool {
	return u.SignatureVersion == SignatureVersionV1 &&
		u.SignatureVerified &&
		!u.LegacyUnverified &&
		u.AuthenticationStatus == UsageAuthenticationStatusVerified &&
		len(u.UsageDigest) == DigestSize
}

// UsageAcknowledgmentProof contains all customer-supplied detached proof fields.
type UsageAcknowledgmentProof struct {
	Signature        []byte
	UsageDigest      []byte
	ReplayKey        []byte
	IssuedAtHeight   int64
	ExpiresAtHeight  int64
	IssuedAtUnix     int64
	ExpiresAtUnix    int64
	SignatureVersion uint32
}

// CanonicalPayload projects an acknowledgment proof into its sign bytes.
func (p UsageAcknowledgmentProof) CanonicalPayload(chainID, customer, usageID string) CanonicalAcknowledgmentPayload {
	return CanonicalAcknowledgmentPayload{
		SignatureVersion: p.SignatureVersion,
		ChainID:          chainID,
		Domain:           UsageCustomerDomainV1,
		SignerRole:       SignerRoleCustomer,
		Customer:         customer,
		UsageID:          usageID,
		UsageDigest:      p.UsageDigest,
		ReplayKey:        p.ReplayKey,
		IssuedAtHeight:   p.IssuedAtHeight,
		ExpiresAtHeight:  p.ExpiresAtHeight,
		IssuedAtUnix:     p.IssuedAtUnix,
		ExpiresAtUnix:    p.ExpiresAtUnix,
	}
}

// MarkSettled marks the usage record as settled
func (u *UsageRecord) MarkSettled(settlementID string) {
	u.Settled = true
	u.SettlementID = settlementID
}

// MarshalJSON implements json.Marshaler
func (s SettlementRecord) MarshalJSON() ([]byte, error) {
	type Alias SettlementRecord
	return json.Marshal(&struct {
		Alias
		TotalAmount   []sdk.Coin `json:"total_amount"`
		ProviderShare []sdk.Coin `json:"provider_share"`
		PlatformFee   []sdk.Coin `json:"platform_fee"`
		ValidatorFee  []sdk.Coin `json:"validator_fee,omitempty"`
	}{
		Alias:         (Alias)(s),
		TotalAmount:   s.TotalAmount,
		ProviderShare: s.ProviderShare,
		PlatformFee:   s.PlatformFee,
		ValidatorFee:  s.ValidatorFee,
	})
}
