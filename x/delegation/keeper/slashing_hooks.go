// Package keeper implements the delegation module keeper.
//
// VE-922: Slashing hooks for validator misbehavior
package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ stakingtypes.StakingHooks = SlashingHooks{}

// SlashingHooks provides staking hooks for delegator slashing.
type SlashingHooks struct {
	k Keeper
}

// Hooks returns the delegation module's staking hooks.
func (k Keeper) Hooks() SlashingHooks {
	return SlashingHooks{k: k}
}

// AfterValidatorCreated is a no-op hook.
func (h SlashingHooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error { return nil }

// BeforeValidatorModified is a no-op hook.
func (h SlashingHooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error { return nil }

// AfterValidatorRemoved is a no-op hook.
func (h SlashingHooks) AfterValidatorRemoved(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorBonded is a no-op hook.
func (h SlashingHooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterValidatorBeginUnbonding is a no-op hook.
func (h SlashingHooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeDelegationCreated is a no-op hook.
func (h SlashingHooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeDelegationSharesModified is a no-op hook.
func (h SlashingHooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeDelegationRemoved is a no-op hook.
func (h SlashingHooks) BeforeDelegationRemoved(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// AfterDelegationModified is a no-op hook.
func (h SlashingHooks) AfterDelegationModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// BeforeValidatorSlashed slashes delegations proportionally.
func (h SlashingHooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction sdkmath.LegacyDec) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return h.k.SlashDelegations(sdkCtx, valAddr.String(), fraction, sdkCtx.BlockHeight())
}

// AfterUnbondingInitiated is a no-op hook.
func (h SlashingHooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error { return nil }
