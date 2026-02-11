package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	markettypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
)

// MarketKeeper defines the subset of the market keeper used for GDPR exports.
type MarketKeeper interface {
	WithOrders(ctx sdk.Context, fn func(markettypes.Order) bool)
	WithBids(ctx sdk.Context, fn func(markettypes.Bid) bool)
	WithLeases(ctx sdk.Context, fn func(mv1.Lease) bool)
}

// EscrowKeeper defines the subset of the escrow keeper used for GDPR exports.
type EscrowKeeper interface {
	WithAccounts(ctx sdk.Context, fn func(etypes.Account) bool)
	WithPayments(ctx sdk.Context, fn func(etypes.Payment) bool)
}

// DelegationQueryKeeper defines the subset of the delegation keeper used for GDPR exports.
type DelegationQueryKeeper interface {
	GetDelegatorDelegations(ctx sdk.Context, delegatorAddr string) []delegationtypes.Delegation
	GetDelegatorUnbondingDelegations(ctx sdk.Context, delegatorAddr string) []delegationtypes.UnbondingDelegation
	GetDelegatorRedelegations(ctx sdk.Context, delegatorAddr string) []delegationtypes.Redelegation
	GetDelegatorUnclaimedRewards(ctx sdk.Context, delegatorAddr string) []delegationtypes.DelegatorReward
	GetDelegatorSlashingEvents(ctx sdk.Context, delegatorAddr string) []delegationtypes.DelegatorSlashingEvent
}
