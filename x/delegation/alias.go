// Package delegation implements the delegation module for VirtEngine.
//
// VE-922: Module aliases
package delegation

import (
	"pkg.akt.dev/node/x/delegation/keeper"
	"pkg.akt.dev/node/x/delegation/types"
)

const (
	// ModuleName is the name of the delegation module
	ModuleName = types.ModuleName
	// StoreKey is the store key for the delegation module
	StoreKey = types.StoreKey
	// RouterKey is the router key for the delegation module
	RouterKey = types.RouterKey
)

var (
	// NewKeeper creates a new delegation keeper
	NewKeeper = keeper.NewKeeper
)

type (
	// Keeper is an alias for the delegation keeper
	Keeper = keeper.Keeper

	// GenesisState is an alias for the genesis state
	GenesisState = types.GenesisState

	// Params is an alias for the module parameters
	Params = types.Params

	// Delegation is an alias for delegation type
	Delegation = types.Delegation

	// UnbondingDelegation is an alias for unbonding delegation type
	UnbondingDelegation = types.UnbondingDelegation

	// Redelegation is an alias for redelegation type
	Redelegation = types.Redelegation

	// ValidatorShares is an alias for validator shares type
	ValidatorShares = types.ValidatorShares

	// DelegatorReward is an alias for delegator reward type
	DelegatorReward = types.DelegatorReward

	// MsgDelegate is an alias for delegate message
	MsgDelegate = types.MsgDelegate

	// MsgUndelegate is an alias for undelegate message
	MsgUndelegate = types.MsgUndelegate

	// MsgRedelegate is an alias for redelegate message
	MsgRedelegate = types.MsgRedelegate

	// MsgClaimRewards is an alias for claim rewards message
	MsgClaimRewards = types.MsgClaimRewards

	// MsgClaimAllRewards is an alias for claim all rewards message
	MsgClaimAllRewards = types.MsgClaimAllRewards
)
