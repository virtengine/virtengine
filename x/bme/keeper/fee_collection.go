// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// Fee collection constants
const (
	// FeeCollectorName is the module name for the fee collector (distribution module)
	FeeCollectorName = distrtypes.ModuleName
)

// CollectFees collects fees from a payer and sends them to the fee pool.
// This is used for transaction fees, protocol fees, etc.
func (k *Keeper) CollectFees(ctx sdk.Context, payer sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: fee amount must be positive", ErrInvalidAmount)
	}

	// Check if payer has sufficient funds
	spendable := k.bankKeeper.SpendableCoins(ctx, payer)
	if !spendable.IsAllGTE(amount) {
		return fmt.Errorf("%w: payer has %s, needs %s for fees", ErrInsufficientFunds, spendable, amount)
	}

	// Transfer fees to the BME module first
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to collect fees: %w", err)
	}

	// Then transfer from BME module to the distribution module (fee collector)
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, FeeCollectorName, amount); err != nil {
		return fmt.Errorf("failed to transfer fees to fee collector: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "collect_fees"),
			sdk.NewAttribute("payer", payer.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// MintTokens mints new tokens and sends them to a recipient.
// This is used for reward distribution, collateral release, etc.
func (k *Keeper) MintTokens(ctx sdk.Context, recipient sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: mint amount must be positive", ErrInvalidAmount)
	}

	// Mint tokens to the BME module account
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to mint tokens: %w", err)
	}

	// Transfer minted tokens to recipient
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, amount); err != nil {
		return fmt.Errorf("failed to send minted tokens to recipient: %w", err)
	}

	// Update vault state
	state := k.GetState(ctx)
	state.TotalMinted = state.TotalMinted.Add(amount...)
	if err := k.SetState(ctx, state); err != nil {
		return fmt.Errorf("failed to update vault state: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "mint_tokens"),
			sdk.NewAttribute("recipient", recipient.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// BurnTokens burns tokens from an account.
// This is used for token destruction, collateral burning, etc.
func (k *Keeper) BurnTokens(ctx sdk.Context, from sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: burn amount must be positive", ErrInvalidAmount)
	}

	// Check if account has sufficient funds
	spendable := k.bankKeeper.SpendableCoins(ctx, from)
	if !spendable.IsAllGTE(amount) {
		return fmt.Errorf("%w: account has %s, needs %s to burn", ErrInsufficientFunds, spendable, amount)
	}

	// Transfer tokens from account to BME module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to transfer tokens for burning: %w", err)
	}

	// Burn tokens from BME module
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to burn tokens: %w", err)
	}

	// Update vault state
	state := k.GetState(ctx)
	state.TotalBurned = state.TotalBurned.Add(amount...)
	if err := k.SetState(ctx, state); err != nil {
		return fmt.Errorf("failed to update vault state: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "burn_tokens"),
			sdk.NewAttribute("from", from.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// TransferTokens transfers tokens between two accounts through the BME module.
// This provides an auditable transfer mechanism for billing operations.
func (k *Keeper) TransferTokens(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: transfer amount must be positive", ErrInvalidAmount)
	}

	// Direct transfer using bank keeper
	if err := k.bankKeeper.SendCoins(ctx, from, to, amount); err != nil {
		return fmt.Errorf("failed to transfer tokens: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "transfer_tokens"),
			sdk.NewAttribute("from", from.String()),
			sdk.NewAttribute("to", to.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// GetModuleBalance returns the current balance held by the BME module.
func (k *Keeper) GetModuleBalance(ctx sdk.Context, denom string) sdk.Coin {
	if k.bankKeeper == nil {
		return sdk.NewCoin(denom, math.ZeroInt())
	}

	// Get BME module address - we need to compute it from the module name
	// The module address is derived from the module name
	moduleAddr := sdk.AccAddress([]byte(types.ModuleName))
	return k.bankKeeper.GetBalance(ctx, moduleAddr, denom)
}
