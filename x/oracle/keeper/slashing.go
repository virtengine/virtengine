// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
)

// Slashing store key prefixes
var (
	SlashRecordPrefix = []byte{0x20}
	TotalSlashedKey   = []byte{0x21}
)

// SlashRecordKey returns the key for a slash record.
func SlashRecordKey(oracleAddr sdk.AccAddress, height int64) []byte {
	key := make([]byte, 0, len(SlashRecordPrefix)+len(oracleAddr.Bytes())+8)
	key = append(key, SlashRecordPrefix...)
	key = append(key, oracleAddr.Bytes()...)
	key = append(key, byte(height>>56), byte(height>>48), byte(height>>40), byte(height>>32))
	key = append(key, byte(height>>24), byte(height>>16), byte(height>>8), byte(height))
	return key
}

// SlashRecord stores information about a slashing event.
type SlashRecord struct {
	OracleAddress string    `json:"oracle_address"`
	Amount        sdk.Coins `json:"amount"`
	Reason        string    `json:"reason"`
	Height        int64     `json:"height"`
}

// SlashDeposit slashes an oracle's staked deposit for misbehavior.
// The slashed tokens are burned from the oracle module.
func (k *Keeper) SlashDeposit(ctx sdk.Context, oracleAddr sdk.AccAddress, amount sdk.Coins, reason string) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: slash amount must be positive", ErrInvalidAmount)
	}

	// Get oracle's current stake
	stake := k.GetOracleStake(ctx, oracleAddr)
	if stake.IsZero() {
		return fmt.Errorf("%w: oracle %s has no stake", ErrOracleNotFound, oracleAddr.String())
	}

	// Calculate actual slash amount (can't slash more than staked)
	actualSlash := sdk.Coins{}
	for _, slashCoin := range amount {
		stakeAmount := stake.AmountOf(slashCoin.Denom)
		if stakeAmount.IsZero() {
			continue
		}
		if slashCoin.Amount.LTE(stakeAmount) {
			actualSlash = actualSlash.Add(slashCoin)
		} else {
			actualSlash = actualSlash.Add(sdk.NewCoin(slashCoin.Denom, stakeAmount))
		}
	}

	if actualSlash.IsZero() {
		return nil // Nothing to slash
	}

	// Burn the slashed tokens from the oracle module
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, actualSlash); err != nil {
		return fmt.Errorf("failed to burn slashed tokens: %w", err)
	}

	// Update oracle's stake record
	remainingStake := stake.Sub(actualSlash...)
	k.setOracleStake(ctx, oracleAddr, remainingStake)

	// Update total slashed
	totalSlashed := k.getTotalSlashed(ctx)
	totalSlashed = totalSlashed.Add(actualSlash...)
	k.setTotalSlashed(ctx, totalSlashed)

	// Store slash record
	record := SlashRecord{
		OracleAddress: oracleAddr.String(),
		Amount:        actualSlash,
		Reason:        reason,
		Height:        ctx.BlockHeight(),
	}
	k.setSlashRecord(ctx, oracleAddr, ctx.BlockHeight(), record)

	// Update oracle performance
	perf := k.getOraclePerformance(ctx, oracleAddr)
	perf.ConsecutiveMisses++ // Slashing counts as a miss
	k.setOraclePerformance(ctx, oracleAddr, perf)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "slash_deposit"),
			sdk.NewAttribute("oracle", oracleAddr.String()),
			sdk.NewAttribute("amount", actualSlash.String()),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// SlashForInactivity slashes an oracle for being inactive.
func (k *Keeper) SlashForInactivity(ctx sdk.Context, oracleAddr sdk.AccAddress, missedBlocks uint32) error {
	// Calculate slash amount based on missed blocks
	// Using a simple formula: slash 0.1% per 10 missed blocks
	slashBps := int64(missedBlocks / 10)
	if slashBps == 0 {
		return nil // Not enough missed blocks to slash
	}
	if slashBps > 1000 {
		slashBps = 1000 // Cap at 10%
	}

	stake := k.GetOracleStake(ctx, oracleAddr)
	if stake.IsZero() {
		return nil
	}

	slashAmount := sdk.Coins{}
	for _, coin := range stake {
		slashAmt := coin.Amount.MulRaw(slashBps).QuoRaw(10000)
		if slashAmt.IsPositive() {
			slashAmount = slashAmount.Add(sdk.NewCoin(coin.Denom, slashAmt))
		}
	}

	if slashAmount.IsZero() {
		return nil
	}

	reason := fmt.Sprintf("inactivity: missed %d blocks", missedBlocks)
	return k.SlashDeposit(ctx, oracleAddr, slashAmount, reason)
}

// SlashForBadPrice slashes an oracle for submitting a bad price.
func (k *Keeper) SlashForBadPrice(ctx sdk.Context, oracleAddr sdk.AccAddress, deviation uint64) error {
	params := k.GetParams(ctx)

	// Only slash if deviation exceeds max allowed
	if deviation <= params.MaxPriceDeviationBps {
		return nil
	}

	// Calculate slash amount based on deviation
	// Slash 1% for every 100 bps of excess deviation
	excessDeviation := deviation - params.MaxPriceDeviationBps
	slashBps := int64(min(excessDeviation/100, 500)) //nolint:gosec // Value is bounded by min(n, 500)
	if slashBps == 0 {
		return nil
	}

	stake := k.GetOracleStake(ctx, oracleAddr)
	if stake.IsZero() {
		return nil
	}

	slashAmount := sdk.Coins{}
	for _, coin := range stake {
		slashAmt := coin.Amount.MulRaw(slashBps).QuoRaw(10000)
		if slashAmt.IsPositive() {
			slashAmount = slashAmount.Add(sdk.NewCoin(coin.Denom, slashAmt))
		}
	}

	if slashAmount.IsZero() {
		return nil
	}

	reason := fmt.Sprintf("bad price: deviation %d bps exceeds max %d", deviation, params.MaxPriceDeviationBps)
	return k.SlashDeposit(ctx, oracleAddr, slashAmount, reason)
}

// getTotalSlashed retrieves the total slashed amount.
func (k *Keeper) getTotalSlashed(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(TotalSlashedKey)
	if bz == nil {
		return sdk.Coins{}
	}

	var total sdk.Coins
	if err := json.Unmarshal(bz, &total); err != nil {
		return sdk.Coins{}
	}
	return total
}

// setTotalSlashed stores the total slashed amount.
func (k *Keeper) setTotalSlashed(ctx sdk.Context, total sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&total)
	if err != nil {
		return
	}
	store.Set(TotalSlashedKey, bz)
}

// setSlashRecord stores a slash record.
func (k *Keeper) setSlashRecord(ctx sdk.Context, oracleAddr sdk.AccAddress, height int64, record SlashRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&record)
	if err != nil {
		return
	}
	store.Set(SlashRecordKey(oracleAddr, height), bz)
}

// GetSlashRecords returns all slash records for an oracle.
func (k *Keeper) GetSlashRecords(ctx sdk.Context, oracleAddr sdk.AccAddress) []SlashRecord {
	store := ctx.KVStore(k.storeKey)
	prefix := make([]byte, 0, len(SlashRecordPrefix)+len(oracleAddr.Bytes()))
	prefix = append(prefix, SlashRecordPrefix...)
	prefix = append(prefix, oracleAddr.Bytes()...)

	var records []SlashRecord
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record SlashRecord
		if err := json.Unmarshal(iter.Value(), &record); err == nil {
			records = append(records, record)
		}
	}

	return records
}

// GetTotalSlashed returns the total amount slashed across all oracles.
func (k *Keeper) GetTotalSlashed(ctx sdk.Context) sdk.Coins {
	return k.getTotalSlashed(ctx)
}
