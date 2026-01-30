// Package keeper implements the staking module keeper.
//
// VE-921: Staking rewards keeper
package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/staking/types"
)

// IKeeper defines the interface for the staking keeper
type IKeeper interface {
	// Performance management
	GetValidatorPerformance(ctx sdk.Context, validatorAddr string, epoch uint64) (types.ValidatorPerformance, bool)
	SetValidatorPerformance(ctx sdk.Context, perf types.ValidatorPerformance) error
	UpdateValidatorPerformance(ctx sdk.Context, validatorAddr string, update PerformanceUpdate) error
	GetCurrentEpochPerformances(ctx sdk.Context) []types.ValidatorPerformance

	// Reward management
	CalculateEpochRewards(ctx sdk.Context, epoch uint64) ([]types.ValidatorReward, error)
	DistributeRewards(ctx sdk.Context, epoch uint64) error
	GetValidatorReward(ctx sdk.Context, validatorAddr string, epoch uint64) (types.ValidatorReward, bool)
	SetValidatorReward(ctx sdk.Context, reward types.ValidatorReward) error
	GetRewardEpoch(ctx sdk.Context, epochNumber uint64) (types.RewardEpoch, bool)
	SetRewardEpoch(ctx sdk.Context, epoch types.RewardEpoch) error
	GetCurrentEpoch(ctx sdk.Context) uint64

	// Identity network rewards
	DistributeIdentityNetworkRewards(ctx sdk.Context, epoch uint64) error
	CalculateVEIDRewards(ctx sdk.Context, epoch uint64) ([]types.ValidatorReward, error)

	// Slashing
	SlashValidator(ctx sdk.Context, validatorAddr string, reason types.SlashReason, infractionHeight int64, evidence string) (*types.SlashRecord, error)
	SlashForDoubleSigning(ctx sdk.Context, validatorAddr string, height int64, evidence types.DoubleSignEvidence) (*types.SlashRecord, error)
	SlashForDowntime(ctx sdk.Context, validatorAddr string, missedBlocks int64) (*types.SlashRecord, error)
	SlashForInvalidAttestation(ctx sdk.Context, validatorAddr string, attestation types.InvalidVEIDAttestation) (*types.SlashRecord, error)
	GetSlashRecord(ctx sdk.Context, slashID string) (types.SlashRecord, bool)
	SetSlashRecord(ctx sdk.Context, record types.SlashRecord) error
	GetSlashingRecordsByValidator(ctx sdk.Context, validatorAddr string) []types.SlashRecord

	// Jailing
	JailValidator(ctx sdk.Context, validatorAddr string, duration time.Duration) error
	UnjailValidator(ctx sdk.Context, validatorAddr string) error
	IsValidatorJailed(ctx sdk.Context, validatorAddr string) bool
	TombstoneValidator(ctx sdk.Context, validatorAddr string) error

	// Signing info
	GetValidatorSigningInfo(ctx sdk.Context, validatorAddr string) (types.ValidatorSigningInfo, bool)
	SetValidatorSigningInfo(ctx sdk.Context, info types.ValidatorSigningInfo) error
	HandleValidatorSignature(ctx sdk.Context, validatorAddr string, signed bool) error

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Block hooks
	BeginBlocker(ctx sdk.Context) error
	EndBlocker(ctx sdk.Context) error

	// Iterators
	WithValidatorPerformances(ctx sdk.Context, fn func(types.ValidatorPerformance) bool)
	WithSlashRecords(ctx sdk.Context, fn func(types.SlashRecord) bool)
	WithRewardEpochs(ctx sdk.Context, fn func(types.RewardEpoch) bool)
	WithValidatorRewards(ctx sdk.Context, fn func(types.ValidatorReward) bool)

	// Genesis
	SetCurrentEpoch(ctx sdk.Context, epoch uint64)
	SetNextSlashSequence(ctx sdk.Context, seq uint64)

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// PerformanceUpdate contains performance update data
type PerformanceUpdate struct {
	BlockProposed            bool
	BlockSigned              bool
	BlockMissed              bool
	VEIDVerificationComplete bool
	VEIDVerificationScore    int64
	UptimeSeconds            int64
	DowntimeSeconds          int64
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// VEIDKeeper defines the expected VEID keeper interface
type VEIDKeeper interface {
	GetValidatorVerificationCount(ctx sdk.Context, validatorAddr string, startHeight, endHeight int64) int64
	GetValidatorAverageVerificationScore(ctx sdk.Context, validatorAddr string, startHeight, endHeight int64) int64
}

// StakingKeeper defines the expected staking keeper interface (cosmos-sdk)
type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) []sdk.AccAddress
	GetValidatorStake(ctx sdk.Context, validatorAddr sdk.AccAddress) int64
	GetTotalStake(ctx sdk.Context) int64
}

// Keeper of the staking store
type Keeper struct {
	skey          storetypes.StoreKey
	cdc           codec.BinaryCodec
	bankKeeper    BankKeeper
	veidKeeper    VEIDKeeper
	stakingKeeper StakingKeeper
	authority     string
}

// NewKeeper creates and returns an instance for staking keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	bankKeeper BankKeeper,
	veidKeeper VEIDKeeper,
	stakingKeeper StakingKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:           cdc,
		skey:          skey,
		bankKeeper:    bankKeeper,
		veidKeeper:    veidKeeper,
		stakingKeeper: stakingKeeper,
		authority:     authority,
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
	if err := params.Validate(); err != nil {
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
// Epoch Management
// ============================================================================

// GetCurrentEpoch returns the current epoch number
func (k Keeper) GetCurrentEpoch(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get([]byte("current_epoch"))
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// SetCurrentEpoch sets the current epoch number
func (k Keeper) SetCurrentEpoch(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, epoch)
	store.Set([]byte("current_epoch"), bz)
}

// getNextSlashSequence returns and increments the slash sequence
func (k Keeper) getNextSlashSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeySlash)
	seq := uint64(1)
	if bz != nil {
		seq = binary.BigEndian.Uint64(bz)
	}
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq+1)
	store.Set(types.SequenceKeySlash, bz)
	return seq
}

// SetNextSlashSequence sets the next slash sequence
func (k Keeper) SetNextSlashSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, seq)
	store.Set(types.SequenceKeySlash, bz)
}

// generateSlashID generates a deterministic slash ID
func (k Keeper) generateSlashID(ctx sdk.Context, validatorAddr string, reason types.SlashReason) string {
	seq := k.getNextSlashSequence(ctx)
	data := fmt.Sprintf("%s:%s:%d:%d", validatorAddr, reason, ctx.BlockHeight(), seq)
	hash := sha256.Sum256([]byte(data))
	return "slash-" + hex.EncodeToString(hash[:8])
}

// ============================================================================
// Validator Performance
// ============================================================================

// GetValidatorPerformance returns a validator's performance for an epoch
func (k Keeper) GetValidatorPerformance(ctx sdk.Context, validatorAddr string, epoch uint64) (types.ValidatorPerformance, bool) {
	store := ctx.KVStore(k.skey)
	key := k.validatorPerformanceKey(validatorAddr, epoch)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorPerformance{}, false
	}

	var perf types.ValidatorPerformance
	if err := json.Unmarshal(bz, &perf); err != nil {
		return types.ValidatorPerformance{}, false
	}
	return perf, true
}

// SetValidatorPerformance sets a validator's performance
func (k Keeper) SetValidatorPerformance(ctx sdk.Context, perf types.ValidatorPerformance) error {
	if err := types.ValidateValidatorPerformance(&perf); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := k.validatorPerformanceKey(perf.ValidatorAddress, perf.EpochNumber)
	bz, err := json.Marshal(perf)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Also store under validator prefix for iteration
	perfStore := prefix.NewStore(store, types.ValidatorPerformancePrefix)
	perfStore.Set([]byte(perf.ValidatorAddress+":"+strconv.FormatUint(perf.EpochNumber, 10)), bz)

	return nil
}

// validatorPerformanceKey returns the key for validator performance
func (k Keeper) validatorPerformanceKey(validatorAddr string, epoch uint64) []byte {
	return append(types.ValidatorPerformancePrefix, []byte(fmt.Sprintf("%s:%d", validatorAddr, epoch))...)
}

// UpdateValidatorPerformance updates validator performance metrics
func (k Keeper) UpdateValidatorPerformance(ctx sdk.Context, validatorAddr string, update PerformanceUpdate) error {
	epoch := k.GetCurrentEpoch(ctx)
	perf, found := k.GetValidatorPerformance(ctx, validatorAddr, epoch)
	if !found {
		perf = *types.NewValidatorPerformance(validatorAddr, epoch)
	}

	if update.BlockProposed {
		perf.BlocksProposed++
		perf.LastProposedHeight = ctx.BlockHeight()
	}

	if update.BlockSigned {
		perf.TotalSignatures++
		perf.LastSignedHeight = ctx.BlockHeight()
		perf.ConsecutiveMissedBlocks = 0
	}

	if update.BlockMissed {
		perf.BlocksMissed++
		perf.ConsecutiveMissedBlocks++
	}

	if update.VEIDVerificationComplete {
		perf.VEIDVerificationsCompleted++
	}

	if update.VEIDVerificationScore > 0 {
		// Running average
		if perf.VEIDVerificationsCompleted > 0 {
			perf.VEIDVerificationScore = (perf.VEIDVerificationScore*(perf.VEIDVerificationsCompleted-1) + update.VEIDVerificationScore) / perf.VEIDVerificationsCompleted
		} else {
			perf.VEIDVerificationScore = update.VEIDVerificationScore
		}
	}

	perf.UptimeSeconds += update.UptimeSeconds
	perf.DowntimeSeconds += update.DowntimeSeconds
	blockTime := ctx.BlockTime()
	perf.UpdatedAt = &blockTime

	// Recompute overall score
	types.ComputeOverallScore(&perf)

	return k.SetValidatorPerformance(ctx, perf)
}

// GetCurrentEpochPerformances returns all validator performances for the current epoch
func (k Keeper) GetCurrentEpochPerformances(ctx sdk.Context) []types.ValidatorPerformance {
	epoch := k.GetCurrentEpoch(ctx)
	var performances []types.ValidatorPerformance

	k.WithValidatorPerformances(ctx, func(perf types.ValidatorPerformance) bool {
		if perf.EpochNumber == epoch {
			performances = append(performances, perf)
		}
		return false
	})

	return performances
}

// WithValidatorPerformances iterates over all validator performances
func (k Keeper) WithValidatorPerformances(ctx sdk.Context, fn func(types.ValidatorPerformance) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorPerformancePrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var perf types.ValidatorPerformance
		if err := json.Unmarshal(iter.Value(), &perf); err != nil {
			continue
		}
		if fn(perf) {
			break
		}
	}
}
