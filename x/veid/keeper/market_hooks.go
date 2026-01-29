package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// VEID Market Hooks Interface (VE-3028)
// ============================================================================

// VEIDMarketHooks defines the interface that market module can use to validate
// VEID requirements for marketplace participation. These hooks are called by
// the market module during order, bid, and lease lifecycle events.
type VEIDMarketHooks interface {
	// BeforeOrderCreate is called before order creation to validate tenant VEID.
	// Returns error if the tenant does not meet VEID requirements for the market type.
	BeforeOrderCreate(ctx sdk.Context, tenantAddress sdk.AccAddress, marketType types.MarketType) error

	// BeforeBidCreate is called before bid submission to validate provider VEID.
	// Returns error if the provider does not meet VEID requirements.
	BeforeBidCreate(ctx sdk.Context, providerAddress sdk.AccAddress, marketType types.MarketType) error

	// BeforeLeaseCreate is called before lease signing to verify both parties.
	// Returns error if either party does not meet requirements.
	BeforeLeaseCreate(ctx sdk.Context, tenantAddress, providerAddress sdk.AccAddress, marketType types.MarketType) error

	// GetRequiredVEIDLevel returns the required VEID level for marketplace participation.
	GetRequiredVEIDLevel(ctx sdk.Context, marketType types.MarketType) (types.VEIDLevel, error)

	// GetParticipantStatus returns the VEID status for a marketplace participant.
	GetParticipantStatus(ctx sdk.Context, address sdk.AccAddress, marketType types.MarketType) (*types.ParticipantVEIDStatus, error)

	// CheckOrderEligibility performs a detailed eligibility check for order creation.
	CheckOrderEligibility(ctx sdk.Context, tenantAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error)

	// CheckBidEligibility performs a detailed eligibility check for bid submission.
	CheckBidEligibility(ctx sdk.Context, providerAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error)

	// CheckLeaseEligibility performs a detailed eligibility check for lease signing.
	CheckLeaseEligibility(ctx sdk.Context, tenantAddress, providerAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error)
}

// Ensure Keeper implements VEIDMarketHooks
var _ VEIDMarketHooks = Keeper{}

// ============================================================================
// Hook Implementations
// ============================================================================

// BeforeOrderCreate validates that the tenant meets VEID requirements for order creation.
// This hook is called by the market module before processing MsgCreateOrder.
func (k Keeper) BeforeOrderCreate(ctx sdk.Context, tenantAddress sdk.AccAddress, marketType types.MarketType) error {
	k.Logger(ctx).Debug("BeforeOrderCreate hook called",
		"tenant", tenantAddress.String(),
		"market_type", marketType,
	)

	result, err := k.CheckOrderEligibility(ctx, tenantAddress, marketType)
	if err != nil {
		return err
	}

	if !result.Eligible {
		return types.ErrMarketVEIDNotMet.Wrapf("tenant %s: %s", tenantAddress.String(), result.Reason)
	}

	return nil
}

// BeforeBidCreate validates that the provider meets VEID requirements for bid submission.
// This hook is called by the market module before processing MsgCreateBid.
func (k Keeper) BeforeBidCreate(ctx sdk.Context, providerAddress sdk.AccAddress, marketType types.MarketType) error {
	k.Logger(ctx).Debug("BeforeBidCreate hook called",
		"provider", providerAddress.String(),
		"market_type", marketType,
	)

	result, err := k.CheckBidEligibility(ctx, providerAddress, marketType)
	if err != nil {
		return err
	}

	if !result.Eligible {
		return types.ErrProviderVEIDNotMet.Wrapf("provider %s: %s", providerAddress.String(), result.Reason)
	}

	return nil
}

// BeforeLeaseCreate validates that both tenant and provider meet VEID requirements.
// This hook is called by the market module before creating a lease.
func (k Keeper) BeforeLeaseCreate(ctx sdk.Context, tenantAddress, providerAddress sdk.AccAddress, marketType types.MarketType) error {
	k.Logger(ctx).Debug("BeforeLeaseCreate hook called",
		"tenant", tenantAddress.String(),
		"provider", providerAddress.String(),
		"market_type", marketType,
	)

	result, err := k.CheckLeaseEligibility(ctx, tenantAddress, providerAddress, marketType)
	if err != nil {
		return err
	}

	if !result.Eligible {
		return types.ErrMarketVEIDNotMet.Wrap(result.Reason)
	}

	return nil
}

// GetRequiredVEIDLevel returns the required VEID level for a marketplace type.
// If no specific requirements are set, returns default levels based on market type.
func (k Keeper) GetRequiredVEIDLevel(ctx sdk.Context, marketType types.MarketType) (types.VEIDLevel, error) {
	// Try to get configured requirements
	requirements, found := k.GetMarketRequirements(ctx, marketType)
	if !found {
		// Return defaults based on market type
		return k.getDefaultVEIDLevel(marketType), nil
	}

	// Convert trust score to VEID level
	return k.trustScoreToVEIDLevel(requirements.MinTrustScore), nil
}

// getDefaultVEIDLevel returns the default VEID level for a market type
func (k Keeper) getDefaultVEIDLevel(marketType types.MarketType) types.VEIDLevel {
	switch marketType {
	case types.MarketTypeTEE:
		// TEE marketplace requires premium verification
		return types.VEIDLevelPremium
	case types.MarketTypeHPC, types.MarketTypeGPU:
		// High-value resources require standard verification
		return types.VEIDLevelStandard
	case types.MarketTypeCompute, types.MarketTypeStorage:
		// Basic marketplace participation
		return types.VEIDLevelBasic
	default:
		return types.VEIDLevelBasic
	}
}

// trustScoreToVEIDLevel converts a trust score to a VEID level
func (k Keeper) trustScoreToVEIDLevel(score sdkmath.LegacyDec) types.VEIDLevel {
	scoreInt := score.TruncateInt64()

	switch {
	case scoreInt >= int64(types.ThresholdPremium):
		return types.VEIDLevelPremium
	case scoreInt >= int64(types.ThresholdStandard):
		return types.VEIDLevelStandard
	case scoreInt >= int64(types.ThresholdBasic):
		return types.VEIDLevelBasic
	default:
		return types.VEIDLevelNone
	}
}

// ============================================================================
// Hook Registration
// ============================================================================

// MarketHooksWrapper wraps the keeper to implement hooks for the market module.
// This allows the market module to import and use the hooks without circular dependencies.
type MarketHooksWrapper struct {
	k Keeper
}

// NewMarketHooksWrapper creates a new MarketHooksWrapper
func NewMarketHooksWrapper(k Keeper) *MarketHooksWrapper {
	return &MarketHooksWrapper{k: k}
}

// BeforeOrderCreate implements VEIDMarketHooks
func (w *MarketHooksWrapper) BeforeOrderCreate(ctx sdk.Context, tenantAddress sdk.AccAddress, marketType types.MarketType) error {
	return w.k.BeforeOrderCreate(ctx, tenantAddress, marketType)
}

// BeforeBidCreate implements VEIDMarketHooks
func (w *MarketHooksWrapper) BeforeBidCreate(ctx sdk.Context, providerAddress sdk.AccAddress, marketType types.MarketType) error {
	return w.k.BeforeBidCreate(ctx, providerAddress, marketType)
}

// BeforeLeaseCreate implements VEIDMarketHooks
func (w *MarketHooksWrapper) BeforeLeaseCreate(ctx sdk.Context, tenantAddress, providerAddress sdk.AccAddress, marketType types.MarketType) error {
	return w.k.BeforeLeaseCreate(ctx, tenantAddress, providerAddress, marketType)
}

// GetRequiredVEIDLevel implements VEIDMarketHooks
func (w *MarketHooksWrapper) GetRequiredVEIDLevel(ctx sdk.Context, marketType types.MarketType) (types.VEIDLevel, error) {
	return w.k.GetRequiredVEIDLevel(ctx, marketType)
}

// GetParticipantStatus implements VEIDMarketHooks
func (w *MarketHooksWrapper) GetParticipantStatus(ctx sdk.Context, address sdk.AccAddress, marketType types.MarketType) (*types.ParticipantVEIDStatus, error) {
	return w.k.GetParticipantStatus(ctx, address, marketType)
}

// CheckOrderEligibility implements VEIDMarketHooks
func (w *MarketHooksWrapper) CheckOrderEligibility(ctx sdk.Context, tenantAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error) {
	return w.k.CheckOrderEligibility(ctx, tenantAddress, marketType)
}

// CheckBidEligibility implements VEIDMarketHooks
func (w *MarketHooksWrapper) CheckBidEligibility(ctx sdk.Context, providerAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error) {
	return w.k.CheckBidEligibility(ctx, providerAddress, marketType)
}

// CheckLeaseEligibility implements VEIDMarketHooks
func (w *MarketHooksWrapper) CheckLeaseEligibility(ctx sdk.Context, tenantAddress, providerAddress sdk.AccAddress, marketType types.MarketType) (*types.MarketEligibilityResult, error) {
	return w.k.CheckLeaseEligibility(ctx, tenantAddress, providerAddress, marketType)
}
