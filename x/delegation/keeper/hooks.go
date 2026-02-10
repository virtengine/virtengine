// Package keeper implements the delegation module keeper.
//
// VE-922: Block hooks for unbonding queue processing
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker is called at the beginning of every block
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Process mature unbonding delegations
	matureUnbondings := k.GetMatureUnbondings(ctx)
	for _, ubd := range matureUnbondings {
		if err := k.CompleteUnbonding(ctx, ubd.ID); err != nil {
			k.Logger(ctx).Error("failed to complete unbonding",
				"unbonding_id", ubd.ID,
				"delegator", ubd.DelegatorAddress,
				"error", err,
			)
			// Continue processing other unbondings
		}
	}

	return nil
}

// EndBlocker is called at the end of every block
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// Process mature redelegations
	matureRedelegations := k.GetMatureRedelegations(ctx)
	for _, red := range matureRedelegations {
		if err := k.CompleteRedelegation(ctx, red.ID); err != nil {
			k.Logger(ctx).Error("failed to complete redelegation",
				"redelegation_id", red.ID,
				"delegator", red.DelegatorAddress,
				"error", err,
			)
			// Continue processing other redelegations
		}
	}

	return nil
}
