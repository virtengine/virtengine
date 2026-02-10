// Package delegation implements the delegation module for VirtEngine.
//
// VE-922: Genesis initialization
package delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/keeper"
	"github.com/virtengine/virtengine/x/delegation/types"
)

// InitGenesis initializes the delegation module's state from a genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Set sequences
	k.SetDelegationSequence(ctx, data.DelegationSequence)
	k.SetUnbondingSequence(ctx, data.UnbondingSequence)
	k.SetRedelegationSequence(ctx, data.RedelegationSequence)

	// Initialize validator shares
	for _, shares := range data.ValidatorShares {
		if err := k.SetValidatorShares(ctx, shares); err != nil {
			panic(err)
		}
	}

	// Initialize delegations
	for _, del := range data.Delegations {
		if err := k.SetDelegation(ctx, del); err != nil {
			panic(err)
		}
	}

	// Initialize unbonding delegations
	for _, ubd := range data.UnbondingDelegations {
		if err := k.SetUnbondingDelegation(ctx, ubd); err != nil {
			panic(err)
		}
	}

	// Initialize redelegations
	for _, red := range data.Redelegations {
		if err := k.SetRedelegation(ctx, red); err != nil {
			panic(err)
		}
	}

	// Initialize delegator rewards
	for _, reward := range data.DelegatorRewards {
		if err := k.SetDelegatorReward(ctx, reward); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the delegation module's state to a genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	var delegations []types.Delegation
	k.WithDelegations(ctx, func(del types.Delegation) bool {
		delegations = append(delegations, del)
		return false
	})

	var unbondingDelegations []types.UnbondingDelegation
	k.WithUnbondingDelegations(ctx, func(ubd types.UnbondingDelegation) bool {
		unbondingDelegations = append(unbondingDelegations, ubd)
		return false
	})

	var redelegations []types.Redelegation
	k.WithRedelegations(ctx, func(red types.Redelegation) bool {
		redelegations = append(redelegations, red)
		return false
	})

	var validatorShares []types.ValidatorShares
	k.WithValidatorShares(ctx, func(shares types.ValidatorShares) bool {
		validatorShares = append(validatorShares, shares)
		return false
	})

	var delegatorRewards []types.DelegatorReward
	k.WithDelegatorRewards(ctx, func(reward types.DelegatorReward) bool {
		delegatorRewards = append(delegatorRewards, reward)
		return false
	})

	return &types.GenesisState{
		Params:               k.GetParams(ctx),
		Delegations:          delegations,
		UnbondingDelegations: unbondingDelegations,
		Redelegations:        redelegations,
		ValidatorShares:      validatorShares,
		DelegatorRewards:     delegatorRewards,
		DelegationSequence:   k.GetDelegationSequence(ctx),
		UnbondingSequence:    k.GetUnbondingSequence(ctx),
		RedelegationSequence: k.GetRedelegationSequence(ctx),
	}
}
