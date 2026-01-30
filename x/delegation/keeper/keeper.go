// Package keeper implements the delegation module keeper.
//
// VE-922: Delegated staking keeper
package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/delegation/types"
)

// BasisPointsMax is 100% in basis points
const BasisPointsMax int64 = 10000

// DefaultValidatorCommissionRate is the default validator commission rate (10% = 1000 basis points)
const DefaultValidatorCommissionRate int64 = 1000

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// StakingRewardsKeeper defines the expected staking rewards keeper interface
type StakingRewardsKeeper interface {
	GetValidatorReward(ctx sdk.Context, validatorAddr string, epoch uint64) (interface{}, bool)
	GetCurrentEpoch(ctx sdk.Context) uint64
}

// Keeper of the delegation store
type Keeper struct {
	skey                 storetypes.StoreKey
	cdc                  codec.BinaryCodec
	bankKeeper           BankKeeper
	stakingRewardsKeeper StakingRewardsKeeper
	authority            string
}

// NewKeeper creates and returns an instance for delegation keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	bankKeeper BankKeeper,
	stakingRewardsKeeper StakingRewardsKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:                  cdc,
		skey:                 skey,
		bankKeeper:           bankKeeper,
		stakingRewardsKeeper: stakingRewardsKeeper,
		authority:            authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ============================================================================
// Parameters
// ============================================================================

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// ============================================================================
// Sequence Management
// ============================================================================

// Note: The following sequence functions are reserved for future use when
// delegation IDs require unique identifiers. They are currently unused but
// kept for API completeness.

// getNextDelegationSequence returns and increments the delegation sequence
func (k Keeper) getNextDelegationSequence(ctx sdk.Context) uint64 { //nolint:unused // Reserved for future use
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyDelegation)
	seq := uint64(1)
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq+1)
	store.Set(types.SequenceKeyDelegation, bz)
	return seq
}

// getNextUnbondingSequence returns and increments the unbonding sequence
func (k Keeper) getNextUnbondingSequence(ctx sdk.Context) uint64 { //nolint:unused // Reserved for future use
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyUnbonding)
	seq := uint64(1)
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq+1)
	store.Set(types.SequenceKeyUnbonding, bz)
	return seq
}

// getNextRedelegationSequence returns and increments the redelegation sequence
func (k Keeper) getNextRedelegationSequence(ctx sdk.Context) uint64 { //nolint:unused // Reserved for future use
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyRedelegation)
	seq := uint64(1)
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq+1)
	store.Set(types.SequenceKeyRedelegation, bz)
	return seq
}

// SetDelegationSequence sets the delegation sequence
func (k Keeper) SetDelegationSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.SequenceKeyDelegation, bz)
}

// SetUnbondingSequence sets the unbonding sequence
func (k Keeper) SetUnbondingSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.SequenceKeyUnbonding, bz)
}

// SetRedelegationSequence sets the redelegation sequence
func (k Keeper) SetRedelegationSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.SequenceKeyRedelegation, bz)
}

// generateUnbondingID generates a deterministic unbonding ID
func (k Keeper) generateUnbondingID(ctx sdk.Context, delegatorAddr, validatorAddr string) string {
	seq := k.getNextUnbondingSequence(ctx)
	data := fmt.Sprintf("%s:%s:%d:%d", delegatorAddr, validatorAddr, ctx.BlockHeight(), seq)
	hash := sha256.Sum256([]byte(data))
	return "ubd-" + hex.EncodeToString(hash[:8])
}

// generateRedelegationID generates a deterministic redelegation ID
func (k Keeper) generateRedelegationID(ctx sdk.Context, delegatorAddr, srcValidator, dstValidator string) string {
	seq := k.getNextRedelegationSequence(ctx)
	data := fmt.Sprintf("%s:%s:%s:%d:%d", delegatorAddr, srcValidator, dstValidator, ctx.BlockHeight(), seq)
	hash := sha256.Sum256([]byte(data))
	return "red-" + hex.EncodeToString(hash[:8])
}

// ============================================================================
// Delegation Storage
// ============================================================================

// GetDelegation returns a delegation
func (k Keeper) GetDelegation(ctx sdk.Context, delegatorAddr, validatorAddr string) (types.Delegation, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetDelegationKey(delegatorAddr, validatorAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.Delegation{}, false
	}

	var del types.Delegation
	if err := json.Unmarshal(bz, &del); err != nil {
		return types.Delegation{}, false
	}
	return del, true
}

// SetDelegation stores a delegation
func (k Keeper) SetDelegation(ctx sdk.Context, del types.Delegation) error {
	if err := del.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetDelegationKey(del.DelegatorAddress, del.ValidatorAddress)
	bz, err := json.Marshal(del)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// DeleteDelegation deletes a delegation
func (k Keeper) DeleteDelegation(ctx sdk.Context, delegatorAddr, validatorAddr string) {
	store := ctx.KVStore(k.skey)
	key := types.GetDelegationKey(delegatorAddr, validatorAddr)
	store.Delete(key)
}

// GetDelegatorDelegations returns all delegations for a delegator
func (k Keeper) GetDelegatorDelegations(ctx sdk.Context, delegatorAddr string) []types.Delegation {
	var delegations []types.Delegation

	k.WithDelegations(ctx, func(del types.Delegation) bool {
		if del.DelegatorAddress == delegatorAddr {
			delegations = append(delegations, del)
		}
		return false
	})

	return delegations
}

// GetValidatorDelegations returns all delegations for a validator
func (k Keeper) GetValidatorDelegations(ctx sdk.Context, validatorAddr string) []types.Delegation {
	var delegations []types.Delegation

	k.WithDelegations(ctx, func(del types.Delegation) bool {
		if del.ValidatorAddress == validatorAddr {
			delegations = append(delegations, del)
		}
		return false
	})

	return delegations
}

// WithDelegations iterates over all delegations
func (k Keeper) WithDelegations(ctx sdk.Context, fn func(types.Delegation) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.DelegationPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var del types.Delegation
		if err := json.Unmarshal(iter.Value(), &del); err != nil {
			continue
		}
		if fn(del) {
			break
		}
	}
}

// ============================================================================
// Validator Shares
// ============================================================================

// GetValidatorShares returns validator shares
func (k Keeper) GetValidatorShares(ctx sdk.Context, validatorAddr string) (types.ValidatorShares, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetValidatorSharesKey(validatorAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorShares{}, false
	}

	var shares types.ValidatorShares
	if err := json.Unmarshal(bz, &shares); err != nil {
		return types.ValidatorShares{}, false
	}
	return shares, true
}

// SetValidatorShares stores validator shares
func (k Keeper) SetValidatorShares(ctx sdk.Context, shares types.ValidatorShares) error {
	if err := shares.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetValidatorSharesKey(shares.ValidatorAddress)
	bz, err := json.Marshal(shares)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetOrCreateValidatorShares gets or creates validator shares record
func (k Keeper) GetOrCreateValidatorShares(ctx sdk.Context, validatorAddr string) types.ValidatorShares {
	shares, found := k.GetValidatorShares(ctx, validatorAddr)
	if !found {
		shares = *types.NewValidatorShares(validatorAddr, ctx.BlockTime())
	}
	return shares
}

// WithValidatorShares iterates over all validator shares
func (k Keeper) WithValidatorShares(ctx sdk.Context, fn func(types.ValidatorShares) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorSharesPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var shares types.ValidatorShares
		if err := json.Unmarshal(iter.Value(), &shares); err != nil {
			continue
		}
		if fn(shares) {
			break
		}
	}
}

// ============================================================================
// Unbonding Delegation Storage
// ============================================================================

// GetUnbondingDelegation returns an unbonding delegation
func (k Keeper) GetUnbondingDelegation(ctx sdk.Context, unbondingID string) (types.UnbondingDelegation, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetUnbondingDelegationKey(unbondingID)
	bz := store.Get(key)
	if bz == nil {
		return types.UnbondingDelegation{}, false
	}

	var ubd types.UnbondingDelegation
	if err := json.Unmarshal(bz, &ubd); err != nil {
		return types.UnbondingDelegation{}, false
	}
	return ubd, true
}

// SetUnbondingDelegation stores an unbonding delegation
func (k Keeper) SetUnbondingDelegation(ctx sdk.Context, ubd types.UnbondingDelegation) error {
	if err := ubd.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetUnbondingDelegationKey(ubd.ID)
	bz, err := json.Marshal(ubd)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// DeleteUnbondingDelegation deletes an unbonding delegation
func (k Keeper) DeleteUnbondingDelegation(ctx sdk.Context, unbondingID string) {
	store := ctx.KVStore(k.skey)
	key := types.GetUnbondingDelegationKey(unbondingID)
	store.Delete(key)
}

// GetDelegatorUnbondingDelegations returns all unbonding delegations for a delegator
func (k Keeper) GetDelegatorUnbondingDelegations(ctx sdk.Context, delegatorAddr string) []types.UnbondingDelegation {
	var unbondings []types.UnbondingDelegation

	k.WithUnbondingDelegations(ctx, func(ubd types.UnbondingDelegation) bool {
		if ubd.DelegatorAddress == delegatorAddr {
			unbondings = append(unbondings, ubd)
		}
		return false
	})

	return unbondings
}

// WithUnbondingDelegations iterates over all unbonding delegations
func (k Keeper) WithUnbondingDelegations(ctx sdk.Context, fn func(types.UnbondingDelegation) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.UnbondingDelegationPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ubd types.UnbondingDelegation
		if err := json.Unmarshal(iter.Value(), &ubd); err != nil {
			continue
		}
		if fn(ubd) {
			break
		}
	}
}

// ============================================================================
// Redelegation Storage
// ============================================================================

// GetRedelegation returns a redelegation
func (k Keeper) GetRedelegation(ctx sdk.Context, redelegationID string) (types.Redelegation, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetRedelegationKey(redelegationID)
	bz := store.Get(key)
	if bz == nil {
		return types.Redelegation{}, false
	}

	var red types.Redelegation
	if err := json.Unmarshal(bz, &red); err != nil {
		return types.Redelegation{}, false
	}
	return red, true
}

// SetRedelegation stores a redelegation
func (k Keeper) SetRedelegation(ctx sdk.Context, red types.Redelegation) error {
	if err := red.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetRedelegationKey(red.ID)
	bz, err := json.Marshal(red)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// DeleteRedelegation deletes a redelegation
func (k Keeper) DeleteRedelegation(ctx sdk.Context, redelegationID string) {
	store := ctx.KVStore(k.skey)
	key := types.GetRedelegationKey(redelegationID)
	store.Delete(key)
}

// GetDelegatorRedelegations returns all redelegations for a delegator
func (k Keeper) GetDelegatorRedelegations(ctx sdk.Context, delegatorAddr string) []types.Redelegation {
	var redelegations []types.Redelegation

	k.WithRedelegations(ctx, func(red types.Redelegation) bool {
		if red.DelegatorAddress == delegatorAddr {
			redelegations = append(redelegations, red)
		}
		return false
	})

	return redelegations
}

// WithRedelegations iterates over all redelegations
func (k Keeper) WithRedelegations(ctx sdk.Context, fn func(types.Redelegation) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.RedelegationPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var red types.Redelegation
		if err := json.Unmarshal(iter.Value(), &red); err != nil {
			continue
		}
		if fn(red) {
			break
		}
	}
}

// HasRedelegation checks if there is an active redelegation from source validator
func (k Keeper) HasRedelegation(ctx sdk.Context, delegatorAddr, srcValidator string) bool {
	hasRed := false
	k.WithRedelegations(ctx, func(red types.Redelegation) bool {
		if red.DelegatorAddress == delegatorAddr && red.ValidatorSrcAddress == srcValidator {
			hasRed = true
			return true
		}
		return false
	})
	return hasRed
}

// CountDelegatorRedelegations counts active redelegations for a delegator
func (k Keeper) CountDelegatorRedelegations(ctx sdk.Context, delegatorAddr string) int {
	count := 0
	k.WithRedelegations(ctx, func(red types.Redelegation) bool {
		if red.DelegatorAddress == delegatorAddr {
			count++
		}
		return false
	})
	return count
}

// ============================================================================
// Delegator Rewards
// ============================================================================

// GetDelegatorReward returns a delegator reward
func (k Keeper) GetDelegatorReward(ctx sdk.Context, delegatorAddr, validatorAddr string, epoch uint64) (types.DelegatorReward, bool) {
	store := ctx.KVStore(k.skey)
	key := types.GetDelegatorRewardsKey(delegatorAddr, validatorAddr, epoch)
	bz := store.Get(key)
	if bz == nil {
		return types.DelegatorReward{}, false
	}

	var reward types.DelegatorReward
	if err := json.Unmarshal(bz, &reward); err != nil {
		return types.DelegatorReward{}, false
	}
	return reward, true
}

// SetDelegatorReward stores a delegator reward
func (k Keeper) SetDelegatorReward(ctx sdk.Context, reward types.DelegatorReward) error {
	if err := reward.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.GetDelegatorRewardsKey(reward.DelegatorAddress, reward.ValidatorAddress, reward.EpochNumber)
	bz, err := json.Marshal(reward)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetDelegatorUnclaimedRewards returns all unclaimed rewards for a delegator
func (k Keeper) GetDelegatorUnclaimedRewards(ctx sdk.Context, delegatorAddr string) []types.DelegatorReward {
	var rewards []types.DelegatorReward

	k.WithDelegatorRewards(ctx, func(reward types.DelegatorReward) bool {
		if reward.DelegatorAddress == delegatorAddr && !reward.Claimed {
			rewards = append(rewards, reward)
		}
		return false
	})

	return rewards
}

// GetDelegatorValidatorUnclaimedRewards returns unclaimed rewards for a delegator from a specific validator
func (k Keeper) GetDelegatorValidatorUnclaimedRewards(ctx sdk.Context, delegatorAddr, validatorAddr string) []types.DelegatorReward {
	var rewards []types.DelegatorReward

	k.WithDelegatorRewards(ctx, func(reward types.DelegatorReward) bool {
		if reward.DelegatorAddress == delegatorAddr && reward.ValidatorAddress == validatorAddr && !reward.Claimed {
			rewards = append(rewards, reward)
		}
		return false
	})

	return rewards
}

// WithDelegatorRewards iterates over all delegator rewards
func (k Keeper) WithDelegatorRewards(ctx sdk.Context, fn func(types.DelegatorReward) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.DelegatorRewardsPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var reward types.DelegatorReward
		if err := json.Unmarshal(iter.Value(), &reward); err != nil {
			continue
		}
		if fn(reward) {
			break
		}
	}
}

// ============================================================================
// Unbonding Queue
// ============================================================================

// addToUnbondingQueue adds an unbonding delegation to the completion queue
func (k Keeper) addToUnbondingQueue(ctx sdk.Context, completionTime time.Time, unbondingID string) {
	store := ctx.KVStore(k.skey)
	key := types.GetUnbondingQueueKey(completionTime.Unix())

	// Get existing entries at this time
	var entries []string
	if bz := store.Get(key); bz != nil {
		_ = json.Unmarshal(bz, &entries)
	}

	entries = append(entries, unbondingID)
	bz, _ := json.Marshal(entries)
	store.Set(key, bz)
}

// GetMatureUnbondings returns unbonding delegations that are ready to complete
func (k Keeper) GetMatureUnbondings(ctx sdk.Context) []types.UnbondingDelegation {
	var matureUnbondings []types.UnbondingDelegation
	now := ctx.BlockTime()

	k.WithUnbondingDelegations(ctx, func(ubd types.UnbondingDelegation) bool {
		for _, entry := range ubd.Entries {
			if !entry.CompletionTime.After(now) {
				matureUnbondings = append(matureUnbondings, ubd)
				break
			}
		}
		return false
	})

	return matureUnbondings
}

// ============================================================================
// Redelegation Queue
// ============================================================================

// addToRedelegationQueue adds a redelegation to the completion queue
func (k Keeper) addToRedelegationQueue(ctx sdk.Context, completionTime time.Time, redelegationID string) {
	store := ctx.KVStore(k.skey)
	key := types.GetRedelegationQueueKey(completionTime.Unix())

	// Get existing entries at this time
	var entries []string
	if bz := store.Get(key); bz != nil {
		_ = json.Unmarshal(bz, &entries)
	}

	entries = append(entries, redelegationID)
	bz, _ := json.Marshal(entries)
	store.Set(key, bz)
}

// GetMatureRedelegations returns redelegations that are ready to complete
func (k Keeper) GetMatureRedelegations(ctx sdk.Context) []types.Redelegation {
	var matureRedelegations []types.Redelegation
	now := ctx.BlockTime()

	k.WithRedelegations(ctx, func(red types.Redelegation) bool {
		for _, entry := range red.Entries {
			if !entry.CompletionTime.After(now) {
				matureRedelegations = append(matureRedelegations, red)
				break
			}
		}
		return false
	})

	return matureRedelegations
}

// ============================================================================
// Share Calculations (Deterministic)
// ============================================================================

// CalculateDelegatorProportion calculates the delegator's proportion of validator shares
// Returns proportion as basis points (0-10000)
func (k Keeper) CalculateDelegatorProportion(ctx sdk.Context, delegatorAddr, validatorAddr string) (int64, error) {
	del, found := k.GetDelegation(ctx, delegatorAddr, validatorAddr)
	if !found {
		return 0, types.ErrDelegationNotFound
	}

	valShares, found := k.GetValidatorShares(ctx, validatorAddr)
	if !found {
		return 0, types.ErrValidatorNotFound
	}

	delegatorShares := del.GetSharesBigInt()
	totalShares := valShares.GetTotalSharesBigInt()

	if totalShares.Sign() == 0 {
		return 0, nil
	}

	// proportion = (delegatorShares * BasisPointsMax) / totalShares
	proportion := new(big.Int).Mul(delegatorShares, big.NewInt(BasisPointsMax))
	proportion.Div(proportion, totalShares)

	return proportion.Int64(), nil
}

// CalculateDelegatorRewardAmount calculates the reward amount for a delegator
// based on their share proportion and validator commission
func (k Keeper) CalculateDelegatorRewardAmount(ctx sdk.Context, delegatorAddr, validatorAddr string, validatorReward string) (string, error) {
	proportion, err := k.CalculateDelegatorProportion(ctx, delegatorAddr, validatorAddr)
	if err != nil {
		return "0", err
	}

	rewardBig, ok := new(big.Int).SetString(validatorReward, 10)
	if !ok || rewardBig.Sign() <= 0 {
		return "0", nil
	}

	// Calculate commission (goes to validator) using default commission rate
	// commission = validatorReward * commissionRate / BasisPointsMax
	commission := new(big.Int).Mul(rewardBig, big.NewInt(DefaultValidatorCommissionRate))
	commission.Div(commission, big.NewInt(BasisPointsMax))

	// Distributable = validatorReward - commission
	distributable := new(big.Int).Sub(rewardBig, commission)

	// Delegator reward = distributable * proportion / BasisPointsMax
	delegatorReward := new(big.Int).Mul(distributable, big.NewInt(proportion))
	delegatorReward.Div(delegatorReward, big.NewInt(BasisPointsMax))

	return delegatorReward.String(), nil
}
