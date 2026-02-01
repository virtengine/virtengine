// Package billing provides billing and invoice types for the escrow module.
//
// This file defines treasury accounting types for fee allocation,
// platform revenue, and community pool contributions.
package billing

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FeeType defines the type of fee
type FeeType uint8

const (
	// FeeTypePlatform is the platform service fee
	FeeTypePlatform FeeType = 0

	// FeeTypeNetwork is the network usage fee
	FeeTypeNetwork FeeType = 1

	// FeeTypeCommunity is the community pool contribution
	FeeTypeCommunity FeeType = 2

	// FeeTypeTake is the take rate fee (for providers)
	FeeTypeTake FeeType = 3

	// FeeTypeValidator is the validator fee share
	FeeTypeValidator FeeType = 4
)

// FeeTypeNames maps fee types to names
var FeeTypeNames = map[FeeType]string{
	FeeTypePlatform:  "platform",
	FeeTypeNetwork:   "network",
	FeeTypeCommunity: "community",
	FeeTypeTake:      "take",
	FeeTypeValidator: "validator",
}

// String returns string representation
func (t FeeType) String() string {
	if name, ok := FeeTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// TreasuryAllocation represents a fee allocation to a treasury account
type TreasuryAllocation struct {
	// AllocationID is the unique identifier
	AllocationID string `json:"allocation_id"`

	// FeeType is the type of fee
	FeeType FeeType `json:"fee_type"`

	// InvoiceID links to the source invoice
	InvoiceID string `json:"invoice_id"`

	// SettlementID links to the settlement record
	SettlementID string `json:"settlement_id"`

	// Amount is the allocated amount
	Amount sdk.Coins `json:"amount"`

	// Destination is the destination account/module
	Destination string `json:"destination"`

	// Description describes the allocation
	Description string `json:"description"`

	// BlockHeight is when the allocation was made
	BlockHeight int64 `json:"block_height"`

	// Timestamp is when the allocation was made
	Timestamp time.Time `json:"timestamp"`

	// TransactionHash is the transaction hash
	TransactionHash string `json:"transaction_hash,omitempty"`

	// Status is the allocation status
	Status TreasuryAllocationStatus `json:"status"`
}

// TreasuryAllocationStatus defines the status of a treasury allocation
type TreasuryAllocationStatus uint8

const (
	// TreasuryAllocationStatusPending is pending allocation
	TreasuryAllocationStatusPending TreasuryAllocationStatus = 0

	// TreasuryAllocationStatusCompleted is completed allocation
	TreasuryAllocationStatusCompleted TreasuryAllocationStatus = 1

	// TreasuryAllocationStatusFailed is failed allocation
	TreasuryAllocationStatusFailed TreasuryAllocationStatus = 2

	// TreasuryAllocationStatusRolledBack is rolled back allocation
	TreasuryAllocationStatusRolledBack TreasuryAllocationStatus = 3
)

// TreasuryAllocationStatusNames maps statuses to names
var TreasuryAllocationStatusNames = map[TreasuryAllocationStatus]string{
	TreasuryAllocationStatusPending:    "pending",
	TreasuryAllocationStatusCompleted:  "completed",
	TreasuryAllocationStatusFailed:     "failed",
	TreasuryAllocationStatusRolledBack: "rolled_back",
}

// String returns string representation
func (s TreasuryAllocationStatus) String() string {
	if name, ok := TreasuryAllocationStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// Validate validates the treasury allocation
func (ta *TreasuryAllocation) Validate() error {
	if ta.AllocationID == "" {
		return fmt.Errorf("allocation_id is required")
	}

	if ta.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if ta.SettlementID == "" {
		return fmt.Errorf("settlement_id is required")
	}

	if !ta.Amount.IsValid() || ta.Amount.IsZero() {
		return fmt.Errorf("amount must be valid and non-zero")
	}

	if ta.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	return nil
}

// FeeConfig defines fee configuration for settlements
type FeeConfig struct {
	// PlatformFeeRate is the platform fee rate (basis points, e.g., 250 = 2.5%)
	PlatformFeeRate sdkmath.LegacyDec `json:"platform_fee_rate"`

	// NetworkFeeRate is the network fee rate (basis points)
	NetworkFeeRate sdkmath.LegacyDec `json:"network_fee_rate"`

	// CommunityPoolRate is the community pool contribution rate (basis points)
	CommunityPoolRate sdkmath.LegacyDec `json:"community_pool_rate"`

	// TakeRate is the provider take rate (from x/take module)
	TakeRate sdkmath.LegacyDec `json:"take_rate"`

	// MinFeeAmount is the minimum fee amount
	MinFeeAmount sdk.Coins `json:"min_fee_amount"`

	// MaxFeeAmount is the maximum fee amount (0 = no max)
	MaxFeeAmount sdk.Coins `json:"max_fee_amount"`

	// FeeExemptAddresses are addresses exempt from fees
	FeeExemptAddresses []string `json:"fee_exempt_addresses,omitempty"`
}

// DefaultFeeConfig returns the default fee configuration
func DefaultFeeConfig() FeeConfig {
	return FeeConfig{
		PlatformFeeRate:    sdkmath.LegacyNewDecWithPrec(250, 4),  // 2.5%
		NetworkFeeRate:     sdkmath.LegacyNewDecWithPrec(50, 4),   // 0.5%
		CommunityPoolRate:  sdkmath.LegacyNewDecWithPrec(100, 4),  // 1.0%
		TakeRate:           sdkmath.LegacyNewDecWithPrec(400, 4),  // 4.0% (from take module)
		MinFeeAmount:       sdk.NewCoins(),
		MaxFeeAmount:       sdk.NewCoins(),
		FeeExemptAddresses: make([]string, 0),
	}
}

// Validate validates the fee configuration
func (fc *FeeConfig) Validate() error {
	if fc.PlatformFeeRate.IsNegative() {
		return fmt.Errorf("platform_fee_rate cannot be negative")
	}

	if fc.NetworkFeeRate.IsNegative() {
		return fmt.Errorf("network_fee_rate cannot be negative")
	}

	if fc.CommunityPoolRate.IsNegative() {
		return fmt.Errorf("community_pool_rate cannot be negative")
	}

	if fc.TakeRate.IsNegative() {
		return fmt.Errorf("take_rate cannot be negative")
	}

	// Total rate should not exceed 100%
	totalRate := fc.PlatformFeeRate.Add(fc.NetworkFeeRate).Add(fc.CommunityPoolRate).Add(fc.TakeRate)
	if totalRate.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("total fee rate exceeds 100%%: %s", totalRate.String())
	}

	return nil
}

// CalculateFees calculates all fees for a given amount
func (fc *FeeConfig) CalculateFees(amount sdk.Coins) FeeBreakdown {
	breakdown := FeeBreakdown{
		GrossAmount:   amount,
		PlatformFee:   sdk.NewCoins(),
		NetworkFee:    sdk.NewCoins(),
		CommunityFee:  sdk.NewCoins(),
		TakeFee:       sdk.NewCoins(),
		TotalFees:     sdk.NewCoins(),
		NetAmount:     sdk.NewCoins(),
	}

	for _, coin := range amount {
		// Calculate each fee type
		platformFee := coin.Amount.ToLegacyDec().Mul(fc.PlatformFeeRate).TruncateInt()
		networkFee := coin.Amount.ToLegacyDec().Mul(fc.NetworkFeeRate).TruncateInt()
		communityFee := coin.Amount.ToLegacyDec().Mul(fc.CommunityPoolRate).TruncateInt()
		takeFee := coin.Amount.ToLegacyDec().Mul(fc.TakeRate).TruncateInt()

		totalFee := platformFee.Add(networkFee).Add(communityFee).Add(takeFee)
		netAmount := coin.Amount.Sub(totalFee)

		if platformFee.IsPositive() {
			breakdown.PlatformFee = breakdown.PlatformFee.Add(sdk.NewCoin(coin.Denom, platformFee))
		}
		if networkFee.IsPositive() {
			breakdown.NetworkFee = breakdown.NetworkFee.Add(sdk.NewCoin(coin.Denom, networkFee))
		}
		if communityFee.IsPositive() {
			breakdown.CommunityFee = breakdown.CommunityFee.Add(sdk.NewCoin(coin.Denom, communityFee))
		}
		if takeFee.IsPositive() {
			breakdown.TakeFee = breakdown.TakeFee.Add(sdk.NewCoin(coin.Denom, takeFee))
		}
		if totalFee.IsPositive() {
			breakdown.TotalFees = breakdown.TotalFees.Add(sdk.NewCoin(coin.Denom, totalFee))
		}
		if netAmount.IsPositive() {
			breakdown.NetAmount = breakdown.NetAmount.Add(sdk.NewCoin(coin.Denom, netAmount))
		}
	}

	return breakdown
}

// IsExempt checks if an address is exempt from fees
func (fc *FeeConfig) IsExempt(address string) bool {
	for _, exempt := range fc.FeeExemptAddresses {
		if exempt == address {
			return true
		}
	}
	return false
}

// FeeBreakdown represents the breakdown of fees for a settlement
type FeeBreakdown struct {
	// GrossAmount is the total amount before fees
	GrossAmount sdk.Coins `json:"gross_amount"`

	// PlatformFee is the platform service fee
	PlatformFee sdk.Coins `json:"platform_fee"`

	// NetworkFee is the network usage fee
	NetworkFee sdk.Coins `json:"network_fee"`

	// CommunityFee is the community pool contribution
	CommunityFee sdk.Coins `json:"community_fee"`

	// TakeFee is the take rate fee
	TakeFee sdk.Coins `json:"take_fee"`

	// TotalFees is the sum of all fees
	TotalFees sdk.Coins `json:"total_fees"`

	// NetAmount is the amount after fees (provider payout)
	NetAmount sdk.Coins `json:"net_amount"`
}

// Validate validates the fee breakdown
func (fb *FeeBreakdown) Validate() error {
	// Net + Total Fees should equal Gross
	expectedGross := fb.NetAmount.Add(fb.TotalFees...)
	if !expectedGross.Equal(fb.GrossAmount) {
		return fmt.Errorf("fee breakdown does not reconcile: gross=%s, net=%s, fees=%s",
			fb.GrossAmount.String(), fb.NetAmount.String(), fb.TotalFees.String())
	}
	return nil
}

// SettlementRecord represents a settlement transaction record
type SettlementRecord struct {
	// SettlementID is the unique identifier
	SettlementID string `json:"settlement_id"`

	// InvoiceID links to the invoice
	InvoiceID string `json:"invoice_id"`

	// EscrowID links to the escrow account
	EscrowID string `json:"escrow_id"`

	// Provider is the provider address
	Provider string `json:"provider"`

	// Customer is the customer address
	Customer string `json:"customer"`

	// GrossAmount is the total amount before fees
	GrossAmount sdk.Coins `json:"gross_amount"`

	// FeeBreakdown contains the fee breakdown
	FeeBreakdown FeeBreakdown `json:"fee_breakdown"`

	// NetPayout is the net amount paid to provider
	NetPayout sdk.Coins `json:"net_payout"`

	// HoldbackAmount is the amount held back (e.g., for disputes)
	HoldbackAmount sdk.Coins `json:"holdback_amount"`

	// DisputeHoldbackReason is the reason for holdback
	DisputeHoldbackReason string `json:"dispute_holdback_reason,omitempty"`

	// Status is the settlement status
	Status SettlementStatus `json:"status"`

	// SettledAt is when the settlement was completed
	SettledAt *time.Time `json:"settled_at,omitempty"`

	// BlockHeight is when the settlement was created
	BlockHeight int64 `json:"block_height"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Allocations tracks all treasury allocations for this settlement
	Allocations []TreasuryAllocation `json:"allocations,omitempty"`

	// Metadata contains additional settlement details
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SettlementStatus defines the status of a settlement
type SettlementStatus uint8

const (
	// SettlementStatusPending is pending settlement
	SettlementStatusPending SettlementStatus = 0

	// SettlementStatusProcessing is currently processing
	SettlementStatusProcessing SettlementStatus = 1

	// SettlementStatusCompleted is completed settlement
	SettlementStatusCompleted SettlementStatus = 2

	// SettlementStatusFailed is failed settlement
	SettlementStatusFailed SettlementStatus = 3

	// SettlementStatusHeldBack has funds held back
	SettlementStatusHeldBack SettlementStatus = 4

	// SettlementStatusDisputed is under dispute
	SettlementStatusDisputed SettlementStatus = 5

	// SettlementStatusCancelled is cancelled settlement
	SettlementStatusCancelled SettlementStatus = 6
)

// SettlementStatusNames maps statuses to names
var SettlementStatusNames = map[SettlementStatus]string{
	SettlementStatusPending:    "pending",
	SettlementStatusProcessing: "processing",
	SettlementStatusCompleted:  "completed",
	SettlementStatusFailed:     "failed",
	SettlementStatusHeldBack:   "held_back",
	SettlementStatusDisputed:   "disputed",
	SettlementStatusCancelled:  "cancelled",
}

// String returns string representation
func (s SettlementStatus) String() string {
	if name, ok := SettlementStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsTerminal returns true if the status is final
func (s SettlementStatus) IsTerminal() bool {
	return s == SettlementStatusCompleted || s == SettlementStatusFailed || s == SettlementStatusCancelled
}

// NewSettlementRecord creates a new settlement record
func NewSettlementRecord(
	settlementID string,
	invoiceID string,
	escrowID string,
	provider string,
	customer string,
	grossAmount sdk.Coins,
	feeBreakdown FeeBreakdown,
	blockHeight int64,
	now time.Time,
) *SettlementRecord {
	return &SettlementRecord{
		SettlementID:   settlementID,
		InvoiceID:      invoiceID,
		EscrowID:       escrowID,
		Provider:       provider,
		Customer:       customer,
		GrossAmount:    grossAmount,
		FeeBreakdown:   feeBreakdown,
		NetPayout:      feeBreakdown.NetAmount,
		HoldbackAmount: sdk.NewCoins(),
		Status:         SettlementStatusPending,
		BlockHeight:    blockHeight,
		CreatedAt:      now,
		UpdatedAt:      now,
		Allocations:    make([]TreasuryAllocation, 0),
		Metadata:       make(map[string]string),
	}
}

// Validate validates the settlement record
func (sr *SettlementRecord) Validate() error {
	if sr.SettlementID == "" {
		return fmt.Errorf("settlement_id is required")
	}

	if sr.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if sr.EscrowID == "" {
		return fmt.Errorf("escrow_id is required")
	}

	if _, err := sdk.AccAddressFromBech32(sr.Provider); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(sr.Customer); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if !sr.GrossAmount.IsValid() {
		return fmt.Errorf("gross_amount must be valid")
	}

	if err := sr.FeeBreakdown.Validate(); err != nil {
		return fmt.Errorf("invalid fee breakdown: %w", err)
	}

	return nil
}

// SetHoldback sets a holdback amount
func (sr *SettlementRecord) SetHoldback(amount sdk.Coins, reason string) error {
	if sr.Status.IsTerminal() {
		return fmt.Errorf("cannot set holdback on terminal settlement: %s", sr.Status)
	}

	// Ensure holdback doesn't exceed net payout
	if amount.IsAllGT(sr.NetPayout) {
		return fmt.Errorf("holdback amount %s exceeds net payout %s", amount.String(), sr.NetPayout.String())
	}

	sr.HoldbackAmount = amount
	sr.DisputeHoldbackReason = reason
	sr.Status = SettlementStatusHeldBack
	return nil
}

// ReleaseHoldback releases the holdback
func (sr *SettlementRecord) ReleaseHoldback() error {
	if sr.Status != SettlementStatusHeldBack {
		return fmt.Errorf("settlement is not held back: %s", sr.Status)
	}

	sr.HoldbackAmount = sdk.NewCoins()
	sr.DisputeHoldbackReason = ""
	sr.Status = SettlementStatusPending
	return nil
}

// Complete marks the settlement as completed
func (sr *SettlementRecord) Complete(now time.Time) error {
	if sr.Status.IsTerminal() {
		return fmt.Errorf("settlement already in terminal state: %s", sr.Status)
	}

	sr.Status = SettlementStatusCompleted
	sr.SettledAt = &now
	sr.UpdatedAt = now
	return nil
}

// Fail marks the settlement as failed
func (sr *SettlementRecord) Fail(now time.Time) error {
	if sr.Status.IsTerminal() {
		return fmt.Errorf("settlement already in terminal state: %s", sr.Status)
	}

	sr.Status = SettlementStatusFailed
	sr.UpdatedAt = now
	return nil
}

// AddAllocation adds a treasury allocation to the settlement
func (sr *SettlementRecord) AddAllocation(allocation TreasuryAllocation) {
	sr.Allocations = append(sr.Allocations, allocation)
}

// GetAllocationsTotal returns the total of all allocations
func (sr *SettlementRecord) GetAllocationsTotal() sdk.Coins {
	total := sdk.NewCoins()
	for _, alloc := range sr.Allocations {
		total = total.Add(alloc.Amount...)
	}
	return total
}

// TreasurySummary provides a summary of treasury activity
type TreasurySummary struct {
	// PeriodStart is the start of the summary period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the summary period
	PeriodEnd time.Time `json:"period_end"`

	// TotalSettlements is the number of settlements
	TotalSettlements uint32 `json:"total_settlements"`

	// TotalGrossVolume is the total gross volume
	TotalGrossVolume sdk.Coins `json:"total_gross_volume"`

	// TotalPlatformFees is total platform fees collected
	TotalPlatformFees sdk.Coins `json:"total_platform_fees"`

	// TotalNetworkFees is total network fees collected
	TotalNetworkFees sdk.Coins `json:"total_network_fees"`

	// TotalCommunityFees is total community pool contributions
	TotalCommunityFees sdk.Coins `json:"total_community_fees"`

	// TotalTakeFees is total take fees collected
	TotalTakeFees sdk.Coins `json:"total_take_fees"`

	// TotalFees is total of all fees
	TotalFees sdk.Coins `json:"total_fees"`

	// TotalPayouts is total provider payouts
	TotalPayouts sdk.Coins `json:"total_payouts"`

	// TotalHoldbacks is total held back amounts
	TotalHoldbacks sdk.Coins `json:"total_holdbacks"`

	// GeneratedAt is when the summary was generated
	GeneratedAt time.Time `json:"generated_at"`

	// BlockHeight is the block height at generation
	BlockHeight int64 `json:"block_height"`
}

// NewTreasurySummary creates a new treasury summary
func NewTreasurySummary(periodStart, periodEnd time.Time, blockHeight int64, now time.Time) *TreasurySummary {
	return &TreasurySummary{
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		TotalSettlements:   0,
		TotalGrossVolume:   sdk.NewCoins(),
		TotalPlatformFees:  sdk.NewCoins(),
		TotalNetworkFees:   sdk.NewCoins(),
		TotalCommunityFees: sdk.NewCoins(),
		TotalTakeFees:      sdk.NewCoins(),
		TotalFees:          sdk.NewCoins(),
		TotalPayouts:       sdk.NewCoins(),
		TotalHoldbacks:     sdk.NewCoins(),
		GeneratedAt:        now,
		BlockHeight:        blockHeight,
	}
}

// AddSettlement adds a settlement to the summary
func (ts *TreasurySummary) AddSettlement(settlement *SettlementRecord) {
	ts.TotalSettlements++
	ts.TotalGrossVolume = ts.TotalGrossVolume.Add(settlement.GrossAmount...)
	ts.TotalPlatformFees = ts.TotalPlatformFees.Add(settlement.FeeBreakdown.PlatformFee...)
	ts.TotalNetworkFees = ts.TotalNetworkFees.Add(settlement.FeeBreakdown.NetworkFee...)
	ts.TotalCommunityFees = ts.TotalCommunityFees.Add(settlement.FeeBreakdown.CommunityFee...)
	ts.TotalTakeFees = ts.TotalTakeFees.Add(settlement.FeeBreakdown.TakeFee...)
	ts.TotalFees = ts.TotalFees.Add(settlement.FeeBreakdown.TotalFees...)
	ts.TotalPayouts = ts.TotalPayouts.Add(settlement.NetPayout...)
	ts.TotalHoldbacks = ts.TotalHoldbacks.Add(settlement.HoldbackAmount...)
}
