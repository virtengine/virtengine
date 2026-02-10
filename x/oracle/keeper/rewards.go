// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	otypes "github.com/virtengine/virtengine/x/oracle/types"
)

// Reward store key prefixes
var (
	RewardPoolPrefix     = []byte{0x10}
	OracleRewardPrefix   = []byte{0x11}
	EpochRewardPrefix    = []byte{0x12}
	LastRewardEpochKey   = []byte{0x13}
	OraclePerformanceKey = []byte{0x14}
)

// RewardPoolKey returns the key for the reward pool.
func RewardPoolKey() []byte {
	return RewardPoolPrefix
}

// OracleRewardKey returns the key for an oracle's accumulated rewards.
func OracleRewardKey(oracleAddr sdk.AccAddress) []byte {
	return append(OracleRewardPrefix, oracleAddr.Bytes()...)
}

// EpochRewardKey returns the key for epoch reward data.
func EpochRewardKey(epoch uint64) []byte {
	key := make([]byte, 0, len(EpochRewardPrefix)+8)
	key = append(key, EpochRewardPrefix...)
	key = append(key, byte(epoch>>56), byte(epoch>>48), byte(epoch>>40), byte(epoch>>32))
	key = append(key, byte(epoch>>24), byte(epoch>>16), byte(epoch>>8), byte(epoch))
	return key
}

// OraclePerformance tracks an oracle's performance metrics.
type OraclePerformance struct {
	Address           string `json:"address"`
	TotalSubmissions  uint64 `json:"total_submissions"`
	ValidSubmissions  uint64 `json:"valid_submissions"`
	LastActiveHeight  int64  `json:"last_active_height"`
	ConsecutiveMisses uint32 `json:"consecutive_misses"`
}

// OracleRewardRecord tracks accumulated rewards for an oracle.
type OracleRewardRecord struct {
	Address            string    `json:"address"`
	AccumulatedRewards sdk.Coins `json:"accumulated_rewards"`
	LastClaimHeight    int64     `json:"last_claim_height"`
}

// EpochRewardInfo stores information about rewards for an epoch.
type EpochRewardInfo struct {
	Epoch         uint64    `json:"epoch"`
	TotalRewards  sdk.Coins `json:"total_rewards"`
	DistributedAt int64     `json:"distributed_at"`
	NumOracles    uint32    `json:"num_oracles"`
}

// DistributeRewards distributes rewards to oracle operators for a given epoch.
// Rewards are distributed proportionally based on oracle performance.
func (k *Keeper) DistributeRewards(ctx sdk.Context, epoch uint64) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	params := k.GetParams(ctx)

	// Get all active oracles (sources) from params
	if len(params.Sources) == 0 {
		return nil // No oracles to reward
	}

	// Calculate reward per oracle (equal distribution for now)
	// In a more sophisticated system, this would be based on performance metrics
	rewardPool := k.getRewardPool(ctx)
	if rewardPool.IsZero() {
		return nil // No rewards to distribute
	}

	numOracles := len(params.Sources)
	if numOracles == 0 {
		return nil
	}

	// Calculate per-oracle reward
	perOracleReward := sdk.Coins{}
	for _, coin := range rewardPool {
		amountPerOracle := coin.Amount.QuoRaw(int64(numOracles))
		if amountPerOracle.IsPositive() {
			perOracleReward = perOracleReward.Add(sdk.NewCoin(coin.Denom, amountPerOracle))
		}
	}

	if perOracleReward.IsZero() {
		return nil
	}

	totalDistributed := sdk.Coins{}

	// Distribute to each oracle
	for _, sourceAddr := range params.Sources {
		oracleAddr, err := sdk.AccAddressFromBech32(sourceAddr)
		if err != nil {
			continue // Skip invalid addresses
		}

		// Check oracle performance (only reward active oracles)
		perf := k.getOraclePerformance(ctx, oracleAddr)
		if perf.ConsecutiveMisses > 10 {
			continue // Skip inactive oracles
		}

		// Transfer rewards from module to oracle
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, oracleAddr, perOracleReward); err != nil {
			return fmt.Errorf("failed to distribute rewards to oracle %s: %w", sourceAddr, err)
		}

		totalDistributed = totalDistributed.Add(perOracleReward...)

		// Update oracle reward record
		rewardRecord := k.getOracleRewardRecord(ctx, oracleAddr)
		rewardRecord.Address = sourceAddr
		rewardRecord.AccumulatedRewards = rewardRecord.AccumulatedRewards.Add(perOracleReward...)
		rewardRecord.LastClaimHeight = ctx.BlockHeight()
		k.setOracleRewardRecord(ctx, oracleAddr, rewardRecord)
	}

	// Update reward pool (subtract distributed)
	remainingPool := rewardPool.Sub(totalDistributed...)
	k.setRewardPool(ctx, remainingPool)

	// Store epoch reward info
	epochInfo := EpochRewardInfo{
		Epoch:         epoch,
		TotalRewards:  totalDistributed,
		DistributedAt: ctx.BlockHeight(),
		NumOracles:    uint32(min(numOracles, int(^uint32(0)))), //nolint:gosec // numOracles is bounded by len(sources)
	}
	k.setEpochRewardInfo(ctx, epoch, epochInfo)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "distribute_rewards"),
			sdk.NewAttribute("epoch", fmt.Sprintf("%d", epoch)),
			sdk.NewAttribute("total_distributed", totalDistributed.String()),
			sdk.NewAttribute("num_oracles", fmt.Sprintf("%d", numOracles)),
		),
	)

	return nil
}

// AddToRewardPool adds tokens to the oracle reward pool.
func (k *Keeper) AddToRewardPool(ctx sdk.Context, from sdk.AccAddress, amount sdk.Coins) error {
	if k.bankKeeper == nil {
		return ErrBankKeeperNotSet
	}

	if !amount.IsValid() || amount.IsZero() {
		return fmt.Errorf("%w: amount must be positive", ErrInvalidAmount)
	}

	// Transfer from sender to oracle module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to add to reward pool: %w", err)
	}

	// Update reward pool
	pool := k.getRewardPool(ctx)
	pool = pool.Add(amount...)
	k.setRewardPool(ctx, pool)

	return nil
}

// GetRewardPool returns the current reward pool balance.
func (k *Keeper) GetRewardPool(ctx sdk.Context) sdk.Coins {
	return k.getRewardPool(ctx)
}

// getRewardPool retrieves the reward pool from store.
func (k *Keeper) getRewardPool(ctx sdk.Context) sdk.Coins {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(RewardPoolKey())
	if bz == nil {
		return sdk.Coins{}
	}

	var pool sdk.Coins
	if err := json.Unmarshal(bz, &pool); err != nil {
		return sdk.Coins{}
	}
	return pool
}

// setRewardPool stores the reward pool.
func (k *Keeper) setRewardPool(ctx sdk.Context, pool sdk.Coins) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&pool)
	if err != nil {
		return
	}
	store.Set(RewardPoolKey(), bz)
}

// getOracleRewardRecord retrieves an oracle's reward record.
func (k *Keeper) getOracleRewardRecord(ctx sdk.Context, oracleAddr sdk.AccAddress) OracleRewardRecord {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(OracleRewardKey(oracleAddr))
	if bz == nil {
		return OracleRewardRecord{AccumulatedRewards: sdk.Coins{}}
	}

	var record OracleRewardRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return OracleRewardRecord{AccumulatedRewards: sdk.Coins{}}
	}
	return record
}

// setOracleRewardRecord stores an oracle's reward record.
func (k *Keeper) setOracleRewardRecord(ctx sdk.Context, oracleAddr sdk.AccAddress, record OracleRewardRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&record)
	if err != nil {
		return
	}
	store.Set(OracleRewardKey(oracleAddr), bz)
}

// getOraclePerformance retrieves an oracle's performance metrics.
func (k *Keeper) getOraclePerformance(ctx sdk.Context, oracleAddr sdk.AccAddress) OraclePerformance {
	store := ctx.KVStore(k.storeKey)
	key := make([]byte, 0, len(OraclePerformanceKey)+len(oracleAddr.Bytes()))
	key = append(key, OraclePerformanceKey...)
	key = append(key, oracleAddr.Bytes()...)
	bz := store.Get(key)
	if bz == nil {
		return OraclePerformance{}
	}

	var perf OraclePerformance
	if err := json.Unmarshal(bz, &perf); err != nil {
		return OraclePerformance{}
	}
	return perf
}

// setOraclePerformance stores an oracle's performance metrics.
func (k *Keeper) setOraclePerformance(ctx sdk.Context, oracleAddr sdk.AccAddress, perf OraclePerformance) {
	store := ctx.KVStore(k.storeKey)
	key := make([]byte, 0, len(OraclePerformanceKey)+len(oracleAddr.Bytes()))
	key = append(key, OraclePerformanceKey...)
	key = append(key, oracleAddr.Bytes()...)
	bz, err := json.Marshal(&perf)
	if err != nil {
		return
	}
	store.Set(key, bz)
}

// setEpochRewardInfo stores epoch reward information.
func (k *Keeper) setEpochRewardInfo(ctx sdk.Context, epoch uint64, info EpochRewardInfo) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(&info)
	if err != nil {
		return
	}
	store.Set(EpochRewardKey(epoch), bz)
}

// UpdateOraclePerformance updates an oracle's performance after a price submission.
func (k *Keeper) UpdateOraclePerformance(ctx sdk.Context, oracleAddr sdk.AccAddress, valid bool) {
	perf := k.getOraclePerformance(ctx, oracleAddr)
	perf.Address = oracleAddr.String()
	perf.TotalSubmissions++
	perf.LastActiveHeight = ctx.BlockHeight()

	if valid {
		perf.ValidSubmissions++
		perf.ConsecutiveMisses = 0
	} else {
		perf.ConsecutiveMisses++
	}

	k.setOraclePerformance(ctx, oracleAddr, perf)
}

// Ensure unused import is used
var _ = otypes.ParamsPrefix
