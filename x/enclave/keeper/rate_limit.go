package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// Registration rate limiting store keys
var (
	// PrefixRegistrationCount stores per-block registration count
	PrefixRegistrationCount = []byte{0x10}

	// PrefixLastRegistration stores last registration height per validator
	PrefixLastRegistration = []byte{0x11}
)

// registrationCountKey creates store key for block registration count
func registrationCountKey(height int64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(height))
	return append(PrefixRegistrationCount, bz...)
}

// lastRegistrationKey creates store key for validator's last registration
func lastRegistrationKey(validatorAddr sdk.AccAddress) []byte {
	return append(PrefixLastRegistration, validatorAddr...)
}

// CheckRegistrationRateLimit checks if registration is allowed
// Returns error if rate limit would be exceeded
func (k Keeper) CheckRegistrationRateLimit(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	params := k.GetParams(ctx)

	// Check per-block registration limit
	if params.MaxRegistrationsPerBlock > 0 {
		if err := k.checkBlockRegistrationLimit(ctx, params.MaxRegistrationsPerBlock); err != nil {
			return err
		}
	}

	// Check validator cooldown period
	if params.RegistrationCooldownBlocks > 0 {
		if err := k.checkValidatorCooldown(ctx, validatorAddr, params.RegistrationCooldownBlocks); err != nil {
			return err
		}
	}

	return nil
}

// checkBlockRegistrationLimit verifies block registration limit
func (k Keeper) checkBlockRegistrationLimit(ctx sdk.Context, maxPerBlock uint32) error {
	currentHeight := ctx.BlockHeight()
	store := ctx.KVStore(k.skey)

	key := registrationCountKey(currentHeight)
	bz := store.Get(key)

	var count uint32
	if bz != nil {
		count = binary.BigEndian.Uint32(bz)
	}

	if count >= maxPerBlock {
		return types.ErrTooManyRegistrations.Wrapf(
			"block %d has %d registrations, max is %d",
			currentHeight, count, maxPerBlock,
		)
	}

	return nil
}

// checkValidatorCooldown verifies validator cooldown period
func (k Keeper) checkValidatorCooldown(ctx sdk.Context, validatorAddr sdk.AccAddress, cooldownBlocks int64) error {
	store := ctx.KVStore(k.skey)

	key := lastRegistrationKey(validatorAddr)
	bz := store.Get(key)

	if bz != nil {
		lastHeight := int64(binary.BigEndian.Uint64(bz))
		currentHeight := ctx.BlockHeight()
		blocksSinceLastReg := currentHeight - lastHeight

		if blocksSinceLastReg < cooldownBlocks {
			blocksRemaining := cooldownBlocks - blocksSinceLastReg
			return types.ErrRegistrationCooldown.Wrapf(
				"validator must wait %d more blocks (last registration at height %d)",
				blocksRemaining, lastHeight,
			)
		}
	}

	return nil
}

// IncrementBlockRegistrationCount increments the registration count for current block
func (k Keeper) IncrementBlockRegistrationCount(ctx sdk.Context) {
	params := k.GetParams(ctx)
	if params.MaxRegistrationsPerBlock == 0 {
		// Rate limiting disabled
		return
	}

	currentHeight := ctx.BlockHeight()
	store := ctx.KVStore(k.skey)

	key := registrationCountKey(currentHeight)
	bz := store.Get(key)

	var count uint32
	if bz != nil {
		count = binary.BigEndian.Uint32(bz)
	}

	count++

	newBz := make([]byte, 4)
	binary.BigEndian.PutUint32(newBz, count)
	store.Set(key, newBz)
}

// RecordValidatorRegistration records validator's last registration height
func (k Keeper) RecordValidatorRegistration(ctx sdk.Context, validatorAddr sdk.AccAddress) {
	params := k.GetParams(ctx)
	if params.RegistrationCooldownBlocks == 0 {
		// Cooldown disabled
		return
	}

	currentHeight := ctx.BlockHeight()
	store := ctx.KVStore(k.skey)

	key := lastRegistrationKey(validatorAddr)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(currentHeight))
	store.Set(key, bz)
}

// CleanupOldRegistrationCounts removes old per-block registration counts
// Should be called periodically to avoid unbounded store growth
func (k Keeper) CleanupOldRegistrationCounts(ctx sdk.Context) {
	currentHeight := ctx.BlockHeight()
	store := ctx.KVStore(k.skey)

	// Keep last 100,000 blocks of data (roughly 1 week at 5s blocks)
	cutoffHeight := currentHeight - 100000
	if cutoffHeight <= 0 {
		return
	}

	// Iterate and delete old entries
	iterator := store.Iterator(PrefixRegistrationCount, registrationCountKey(cutoffHeight))
	defer iterator.Close()

	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		keysToDelete = append(keysToDelete, iterator.Key())
	}

	for _, key := range keysToDelete {
		store.Delete(key)
	}

	if len(keysToDelete) > 0 {
		k.Logger(ctx).Debug(
			"cleaned up old registration counts",
			"count", len(keysToDelete),
			"cutoff_height", cutoffHeight,
		)
	}
}

// GetBlockRegistrationCount returns current registration count for a block
func (k Keeper) GetBlockRegistrationCount(ctx sdk.Context, height int64) uint32 {
	store := ctx.KVStore(k.skey)
	key := registrationCountKey(height)
	bz := store.Get(key)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint32(bz)
}

// GetValidatorLastRegistrationHeight returns last registration height for a validator
func (k Keeper) GetValidatorLastRegistrationHeight(ctx sdk.Context, validatorAddr sdk.AccAddress) (int64, bool) {
	store := ctx.KVStore(k.skey)
	key := lastRegistrationKey(validatorAddr)
	bz := store.Get(key)

	if bz == nil {
		return 0, false
	}

	return int64(binary.BigEndian.Uint64(bz)), true
}
