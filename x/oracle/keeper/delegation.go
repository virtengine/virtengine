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

// Delegation store key prefixes
var (
	OracleStakePrefix     = []byte{0x30}
	DelegationPrefix      = []byte{0x31}
	DelegatorIndexPrefix  = []byte{0x32}
	TotalStakedKey        = []byte{0x33}
)

// OracleStakeKey returns the key for an oracle's total stake.
func OracleStakeKey(oracleAddr sdk.AccAddress) []byte {
	return append(OracleStakePrefix, oracleAddr.Bytes()...)
}

// DelegationKey returns the key for a delegation record.
func DelegationKey(delegator, oracleAddr sdk.AccAddress) []byte {
	key := make([]byte, 0, len(DelegationPrefix)+len(delegator.Bytes())+len(oracleAddr.Bytes()))
	key = append(key, DelegationPrefix...)
	key = append(key, delegator.Bytes()...)
	key = append(key, oracleAddr.Bytes()...)
	return key
}

// DelegatorIndexKey returns the key for indexing delegations by delegator.
func DelegatorIndexKey(delegator sdk.AccAddress) []byte {
	return append(DelegatorIndexPrefix, delegator.Bytes()...)
}

// DelegationRecord stores information about a delegation.
type DelegationRecord struct {
	Delegator   string    `json:"delegator"`
	Oracle      string    `json:"oracle"`
	Amount      sdk.Coins `json:"amount"`
	StartHeight int64     `json:"start_height"`
}

// DelegateStake allows a delegator to stake tokens on an oracle.
// The staked tokens are transferred from the delegator to the oracle module.
func (k *Keeper) DelegateStake(ctx sdk.Context, delegator, oracleAddr sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: delegation amount must be positive", ErrInvalidAmount)
	}

	// Check if delegator has sufficient funds
	spendable := k.bankKeeper.SpendableCoins(ctx, delegator)
	if !spendable.IsAllGTE(amount) {
		return fmt.Errorf("%w: delegator has %s, needs %s", ErrInsufficientFunds, spendable, amount)
	}

	// Transfer tokens from delegator to oracle module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegator, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to transfer stake: %w", err)
	}

	// Update delegation record
	delegation := k.getDelegation(ctx, delegator, oracleAddr)
	delegation.Delegator = delegator.String()
	delegation.Oracle = oracleAddr.String()
	delegation.Amount = delegation.Amount.Add(amount...)
	if delegation.StartHeight == 0 {
		delegation.StartHeight = ctx.BlockHeight()
	}
	k.setDelegation(ctx, delegator, oracleAddr, delegation)

	// Update oracle's total stake
	oracleStake := k.GetOracleStake(ctx, oracleAddr)
	oracleStake = oracleStake.Add(amount...)
	k.setOracleStake(ctx, oracleAddr, oracleStake)

	// Update total staked
	totalStaked := k.getTotalStaked(ctx)
	totalStaked = totalStaked.Add(amount...)
	k.setTotalStaked(ctx, totalStaked)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "delegate_stake"),
			sdk.NewAttribute("delegator", delegator.String()),
			sdk.NewAttribute("oracle", oracleAddr.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// UndelegateStake allows a delegator to remove staked tokens from an oracle.
// The tokens are transferred from the oracle module back to the delegator.
func (k *Keeper) UndelegateStake(ctx sdk.Context, delegator, oracleAddr sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: undelegation amount must be positive", ErrInvalidAmount)
	}

	// Get current delegation
	delegation := k.getDelegation(ctx, delegator, oracleAddr)
	if delegation.Amount.IsZero() {
		return fmt.Errorf("no delegation found from %s to %s", delegator.String(), oracleAddr.String())
	}

	// Check if undelegation amount is valid
	if !delegation.Amount.IsAllGTE(amount) {
		return fmt.Errorf("%w: delegated %s, requested %s", ErrInsufficientFunds, delegation.Amount, amount)
	}

	// Transfer tokens from oracle module to delegator
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegator, amount); err != nil {
		return fmt.Errorf("failed to return stake: %w", err)
	}

	// Update delegation record
	delegation.Amount = delegation.Amount.Sub(amount...)
	if delegation.Amount.IsZero() {
		k.deleteDelegation(ctx, delegator, oracleAddr)
	} else {
		k.setDelegation(ctx, delegator, oracleAddr, delegation)
	}

	// Update oracle's total stake
	oracleStake := k.GetOracleStake(ctx, oracleAddr)
	oracleStake = oracleStake.Sub(amount...)
	k.setOracleStake(ctx, oracleAddr, oracleStake)

	// Update total staked
	totalStaked := k.getTotalStaked(ctx)
	totalStaked = totalStaked.Sub(amount...)
	k.setTotalStaked(ctx, totalStaked)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "undelegate_stake"),
			sdk.NewAttribute("delegator", delegator.String()),
			sdk.NewAttribute("oracle", oracleAddr.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// GetOracleStake returns the total stake for an oracle.
func (k *Keeper) GetOracleStake(ctx sdk.Context, oracleAddr sdk.AccAddress) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(OracleStakeKey(oracleAddr))
	if bz == nil {
		return sdk.Coins{}
	}

	var stake sdk.Coins
	if err := json.Unmarshal(bz, &stake); err != nil {
		return sdk.Coins{}
	}
	return stake
}

// setOracleStake stores an oracle's total stake.
func (k *Keeper) setOracleStake(ctx sdk.Context, oracleAddr sdk.AccAddress, stake sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	if stake.IsZero() {
		store.Delete(OracleStakeKey(oracleAddr))
		return
	}
	bz, err := json.Marshal(&stake)
	if err != nil {
		return // Silently fail if marshal fails (should not happen for valid coins)
	}
	store.Set(OracleStakeKey(oracleAddr), bz)
}

// getDelegation retrieves a delegation record.
func (k *Keeper) getDelegation(ctx sdk.Context, delegator, oracleAddr sdk.AccAddress) DelegationRecord {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(DelegationKey(delegator, oracleAddr))
	if bz == nil {
		return DelegationRecord{Amount: sdk.Coins{}}
	}

	var record DelegationRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return DelegationRecord{Amount: sdk.Coins{}}
	}
	return record
}

// setDelegation stores a delegation record.
func (k *Keeper) setDelegation(ctx sdk.Context, delegator, oracleAddr sdk.AccAddress, record DelegationRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&record)
	if err != nil {
		return // Silently fail if marshal fails
	}
	store.Set(DelegationKey(delegator, oracleAddr), bz)
}

// deleteDelegation removes a delegation record.
func (k *Keeper) deleteDelegation(ctx sdk.Context, delegator, oracleAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(DelegationKey(delegator, oracleAddr))
}

// getTotalStaked retrieves the total staked amount.
func (k *Keeper) getTotalStaked(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(TotalStakedKey)
	if bz == nil {
		return sdk.Coins{}
	}

	var total sdk.Coins
	if err := json.Unmarshal(bz, &total); err != nil {
		return sdk.Coins{}
	}
	return total
}

// setTotalStaked stores the total staked amount.
func (k *Keeper) setTotalStaked(ctx sdk.Context, total sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&total)
	if err != nil {
		return // Silently fail if marshal fails
	}
	store.Set(TotalStakedKey, bz)
}

// GetTotalStaked returns the total amount staked across all oracles.
func (k *Keeper) GetTotalStaked(ctx sdk.Context) sdk.Coins {
	return k.getTotalStaked(ctx)
}

// GetDelegation returns a delegation record (public accessor).
func (k *Keeper) GetDelegation(ctx sdk.Context, delegator, oracleAddr sdk.AccAddress) DelegationRecord {
	return k.getDelegation(ctx, delegator, oracleAddr)
}

// GetDelegatorDelegations returns all delegations for a delegator.
func (k *Keeper) GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []DelegationRecord {
	store := ctx.KVStore(k.storeKey)
	prefix := make([]byte, 0, len(DelegationPrefix)+len(delegator.Bytes()))
	prefix = append(prefix, DelegationPrefix...)
	prefix = append(prefix, delegator.Bytes()...)

	var delegations []DelegationRecord
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record DelegationRecord
		if err := json.Unmarshal(iter.Value(), &record); err == nil {
			delegations = append(delegations, record)
		}
	}

	return delegations
}
