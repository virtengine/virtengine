// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the bank module.
type BankKeeper interface {
	// SpendableCoins returns the spendable coins for an account.
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	// SendCoins transfers coins from one account to another.
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error

	// SendCoinsFromModuleToAccount transfers coins from a module to an account.
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	// SendCoinsFromAccountToModule transfers coins from an account to a module.
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	// SendCoinsFromModuleToModule transfers coins between modules.
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error

	// MintCoins mints new coins in a module account.
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	// BurnCoins burns coins from a module account.
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	// GetBalance returns the balance of a specific denomination for an address.
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// OracleKeeper defines the expected interface for the oracle module.
type OracleKeeper interface {
	// GetAggregatedPrice returns the aggregated price for a denomination.
	GetAggregatedPrice(ctx sdk.Context, denom, baseDenom string) (price sdkmath.LegacyDec, healthy bool, err error)
}

// AccountKeeper defines the expected interface for the auth module.
type AccountKeeper interface {
	// GetModuleAddress returns the module account address for a given module name.
	GetModuleAddress(moduleName string) sdk.AccAddress

	// GetModuleAccount returns the module account for a given module name.
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}
