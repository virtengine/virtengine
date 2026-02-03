// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// Escrow store key prefixes
var (
	EscrowPrefix = []byte{0x10}
)

// EscrowKey returns the store key for an escrow record.
func EscrowKey(orderID string) []byte {
	return append(EscrowPrefix, []byte(orderID)...)
}

// EscrowRecord represents a held escrow for an order.
type EscrowRecord struct {
	OrderID   string    `json:"order_id"`
	Depositor string    `json:"depositor"`
	Amount    sdk.Coins `json:"amount"`
	Height    int64     `json:"height"`
}

// HoldEscrow locks tokens from a depositor for an order.
// The tokens are transferred from the depositor's account to the BME module account.
func (k *Keeper) HoldEscrow(ctx sdk.Context, orderID string, depositor sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	// Check if depositor has sufficient funds
	spendable := k.bankKeeper.SpendableCoins(ctx, depositor)
	if !spendable.IsAllGTE(amount) {
		return fmt.Errorf("%w: depositor has %s, needs %s", ErrInsufficientFunds, spendable, amount)
	}

	// Transfer tokens from depositor to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to transfer tokens to escrow: %w", err)
	}

	// Store escrow record
	record := EscrowRecord{
		OrderID:   orderID,
		Depositor: depositor.String(),
		Amount:    amount,
		Height:    ctx.BlockHeight(),
	}

	if err := k.setEscrowRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to store escrow record: %w", err)
	}

	// Update vault state
	state := k.GetState(ctx)
	state.Balances = state.Balances.Add(amount...)
	if err := k.SetState(ctx, state); err != nil {
		return fmt.Errorf("failed to update vault state: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "hold_escrow"),
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("depositor", depositor.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// ReleaseEscrow releases escrowed tokens to a recipient.
// This is typically called when an order is fulfilled or cancelled.
func (k *Keeper) ReleaseEscrow(ctx sdk.Context, orderID string, recipient sdk.AccAddress) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	// Get escrow record
	record, found := k.getEscrowRecord(ctx, orderID)
	if !found {
		return fmt.Errorf("%w: order %s", ErrEscrowNotFound, orderID)
	}

	if record.Amount.IsZero() {
		return fmt.Errorf("%w: escrow has no funds", ErrInvalidAmount)
	}

	// Transfer tokens from module account to recipient
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, record.Amount); err != nil {
		return fmt.Errorf("failed to release escrow: %w", err)
	}

	// Update vault state
	state := k.GetState(ctx)
	state.Balances = state.Balances.Sub(record.Amount...)
	if err := k.SetState(ctx, state); err != nil {
		return fmt.Errorf("failed to update vault state: %w", err)
	}

	// Delete escrow record
	k.deleteEscrowRecord(ctx, orderID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "release_escrow"),
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("recipient", recipient.String()),
			sdk.NewAttribute("amount", record.Amount.String()),
		),
	)

	return nil
}

// GetEscrowBalance returns the escrowed amount for an order.
func (k *Keeper) GetEscrowBalance(ctx sdk.Context, orderID string) sdk.Coins {
	record, found := k.getEscrowRecord(ctx, orderID)
	if !found {
		return sdk.Coins{}
	}
	return record.Amount
}

// setEscrowRecord stores an escrow record.
func (k *Keeper) setEscrowRecord(ctx sdk.Context, record EscrowRecord) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(EscrowKey(record.OrderID), bz)
	return nil
}

// getEscrowRecord retrieves an escrow record.
func (k *Keeper) getEscrowRecord(ctx sdk.Context, orderID string) (EscrowRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(EscrowKey(orderID))
	if bz == nil {
		return EscrowRecord{}, false
	}

	var record EscrowRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return EscrowRecord{}, false
	}
	return record, true
}

// deleteEscrowRecord removes an escrow record.
func (k *Keeper) deleteEscrowRecord(ctx sdk.Context, orderID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(EscrowKey(orderID))
}

// RefundEscrow refunds escrowed tokens back to the original depositor.
// This is called when an order is cancelled before fulfillment.
func (k *Keeper) RefundEscrow(ctx sdk.Context, orderID string) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	// Get escrow record
	record, found := k.getEscrowRecord(ctx, orderID)
	if !found {
		return fmt.Errorf("%w: order %s", ErrEscrowNotFound, orderID)
	}

	depositor, err := sdk.AccAddressFromBech32(record.Depositor)
	if err != nil {
		return fmt.Errorf("invalid depositor address: %w", err)
	}

	// Release back to original depositor
	return k.ReleaseEscrow(ctx, orderID, depositor)
}
