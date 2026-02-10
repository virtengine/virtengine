package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type EscrowKeeper interface {
	GetAccount(ctx sdk.Context, id escrowid.Account) (etypes.Account, error)
	GetPayment(ctx sdk.Context, id escrowid.Payment) (etypes.Payment, error)
	AccountClose(ctx sdk.Context, id escrowid.Account) error
	PaymentClose(ctx sdk.Context, id escrowid.Payment) error
}

// VEIDKeeper defines the interface for the VEID module keeper.
// This interface is used by the market module to check identity requirements
// before allowing order creation.
//
// Task Reference: MARKET-VEID-001 - VE-301 Marketplace gating
type VEIDKeeper interface {
	// GetVEIDScore returns the identity score for an account (0-100).
	// Returns false if no identity record exists.
	GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool)

	// GetIdentityRecord returns the full identity record for an account.
	// Returns false if no identity record exists.
	GetIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (veidtypes.IdentityRecord, bool)

	// GetScopesByType returns all scopes of a specific type for an address.
	GetScopesByType(ctx sdk.Context, address sdk.AccAddress, scopeType veidtypes.ScopeType) []veidtypes.IdentityScope
}
