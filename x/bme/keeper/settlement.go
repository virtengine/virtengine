// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
)

// Settlement store key prefixes
var (
	SettlementPrefix = []byte{0x11}
	LeasePrefix      = []byte{0x12}
)

// SettlementKey returns the store key for a settlement record.
func SettlementKey(leaseID string) []byte {
	return append(SettlementPrefix, []byte(leaseID)...)
}

// LeaseKey returns the store key for a lease's escrow association.
func LeaseKey(leaseID string) []byte {
	return append(LeasePrefix, []byte(leaseID)...)
}

// LeaseEscrowInfo stores the association between a lease and its escrow.
type LeaseEscrowInfo struct {
	LeaseID   string `json:"lease_id"`
	OrderID   string `json:"order_id"`
	Provider  string `json:"provider"`
	Depositor string `json:"depositor"`
}

// SettlementRecord stores settlement history for a lease.
type SettlementRecord struct {
	LeaseID     string    `json:"lease_id"`
	Provider    string    `json:"provider"`
	TotalPaid   sdk.Coins `json:"total_paid"`
	LastSettled int64     `json:"last_settled"`
}

// SettleBilling calculates and transfers payment from escrow to a provider based on usage.
// This is called periodically to settle billing for active leases.
func (k *Keeper) SettleBilling(ctx sdk.Context, leaseID string, provider sdk.AccAddress, usageAmount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !usageAmount.IsValid() {
		return fmt.Errorf("%w: invalid usage amount", ErrInvalidAmount)
	}

	// If usage amount is zero, nothing to settle
	if usageAmount.IsZero() {
		return nil
	}

	// Get lease escrow info
	leaseInfo, found := k.getLeaseEscrowInfo(ctx, leaseID)
	if !found {
		return fmt.Errorf("lease escrow info not found for lease %s", leaseID)
	}

	// Get current escrow balance
	escrowBalance := k.GetEscrowBalance(ctx, leaseInfo.OrderID)
	if escrowBalance.IsZero() {
		return fmt.Errorf("%w: no funds in escrow for order %s", ErrInsufficientFunds, leaseInfo.OrderID)
	}

	// Calculate amount to pay (min of usage and available escrow)
	paymentAmount := sdk.Coins{}
	for _, usageCoin := range usageAmount {
		escrowCoin := escrowBalance.AmountOf(usageCoin.Denom)
		if escrowCoin.IsZero() {
			continue
		}
		if usageCoin.Amount.LTE(escrowCoin) {
			paymentAmount = paymentAmount.Add(usageCoin)
		} else {
			paymentAmount = paymentAmount.Add(sdk.NewCoin(usageCoin.Denom, escrowCoin))
		}
	}

	if paymentAmount.IsZero() {
		return nil
	}

	// Get params for fee calculation
	params := k.GetParams(ctx)

	// Calculate settlement fee (in basis points)
	feeAmount := sdk.Coins{}
	providerAmount := sdk.Coins{}
	for _, coin := range paymentAmount {
		// Fee = amount * settleSpreadBps / 10000
		fee := coin.Amount.MulRaw(int64(params.SettleSpreadBps)).QuoRaw(10000)
		if fee.IsPositive() {
			feeAmount = feeAmount.Add(sdk.NewCoin(coin.Denom, fee))
		}
		providerAmt := coin.Amount.Sub(fee)
		if providerAmt.IsPositive() {
			providerAmount = providerAmount.Add(sdk.NewCoin(coin.Denom, providerAmt))
		}
	}

	// Transfer provider payment from module to provider
	if !providerAmount.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, provider, providerAmount); err != nil {
			return fmt.Errorf("failed to pay provider: %w", err)
		}
	}

	// Burn the fee (or could send to fee collector)
	if !feeAmount.IsZero() {
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, feeAmount); err != nil {
			return fmt.Errorf("failed to burn settlement fees: %w", err)
		}
	}

	// Update escrow record with remaining balance
	escrowRecord, _ := k.getEscrowRecord(ctx, leaseInfo.OrderID)
	escrowRecord.Amount = escrowRecord.Amount.Sub(paymentAmount...)
	if err := k.setEscrowRecord(ctx, escrowRecord); err != nil {
		return fmt.Errorf("failed to update escrow record: %w", err)
	}

	// Update vault state
	state := k.GetState(ctx)
	state.Balances = state.Balances.Sub(paymentAmount...)
	if err := k.SetState(ctx, state); err != nil {
		return fmt.Errorf("failed to update vault state: %w", err)
	}

	// Update settlement record
	settlementRecord, _ := k.getSettlementRecord(ctx, leaseID)
	settlementRecord.LeaseID = leaseID
	settlementRecord.Provider = provider.String()
	settlementRecord.TotalPaid = settlementRecord.TotalPaid.Add(providerAmount...)
	settlementRecord.LastSettled = ctx.BlockHeight()
	if err := k.setSettlementRecord(ctx, settlementRecord); err != nil {
		return fmt.Errorf("failed to update settlement record: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "settle_billing"),
			sdk.NewAttribute("lease_id", leaseID),
			sdk.NewAttribute("provider", provider.String()),
			sdk.NewAttribute("payment", providerAmount.String()),
			sdk.NewAttribute("fee", feeAmount.String()),
		),
	)

	return nil
}

// RegisterLeaseEscrow associates a lease with an escrow order.
func (k *Keeper) RegisterLeaseEscrow(ctx sdk.Context, leaseID, orderID string, provider, depositor sdk.AccAddress) error {
	info := LeaseEscrowInfo{
		LeaseID:   leaseID,
		OrderID:   orderID,
		Provider:  provider.String(),
		Depositor: depositor.String(),
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&info)
	if err != nil {
		return err
	}
	store.Set(LeaseKey(leaseID), bz)
	return nil
}

// getLeaseEscrowInfo retrieves the escrow info for a lease.
func (k *Keeper) getLeaseEscrowInfo(ctx sdk.Context, leaseID string) (LeaseEscrowInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(LeaseKey(leaseID))
	if bz == nil {
		return LeaseEscrowInfo{}, false
	}

	var info LeaseEscrowInfo
	if err := json.Unmarshal(bz, &info); err != nil {
		return LeaseEscrowInfo{}, false
	}
	return info, true
}

// setSettlementRecord stores a settlement record.
func (k *Keeper) setSettlementRecord(ctx sdk.Context, record SettlementRecord) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(SettlementKey(record.LeaseID), bz)
	return nil
}

// getSettlementRecord retrieves a settlement record.
func (k *Keeper) getSettlementRecord(ctx sdk.Context, leaseID string) (SettlementRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(SettlementKey(leaseID))
	if bz == nil {
		return SettlementRecord{
			TotalPaid: sdk.Coins{},
		}, false
	}

	var record SettlementRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return SettlementRecord{
			TotalPaid: sdk.Coins{},
		}, false
	}
	return record, true
}

// GetSettlementRecord returns the settlement record for a lease (public accessor).
func (k *Keeper) GetSettlementRecord(ctx sdk.Context, leaseID string) (SettlementRecord, bool) {
	return k.getSettlementRecord(ctx, leaseID)
}
