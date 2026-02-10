package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// HandleKeyExpiry checks for expiring keys, emits warnings, and revokes or rotates expired keys.
func (k Keeper) HandleKeyExpiry(ctx sdk.Context) (uint32, uint32) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixRecipientKey)
	defer func() {
		_ = iter.Close()
	}()

	params := k.GetParams(ctx)
	now := ctx.BlockTime().Unix()
	var warnings uint32
	var expired uint32

	for ; iter.Valid(); iter.Next() {
		var record recipientKeyStore
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		if record.RevokedAt != 0 {
			continue
		}

		if record.ExpiresAt == 0 {
			continue
		}

		if record.ExpiresAt <= now {
			expired++
			addr, err := sdk.AccAddressFromBech32(record.Address)
			if err != nil {
				continue
			}

			record.RevokedAt = now
			if record.DeprecatedAt == 0 {
				record.DeprecatedAt = now
			}
			if record.PurgeAt == 0 && params.RevocationGracePeriodSeconds > 0 {
				record.PurgeAt = now + safeInt64FromUint64(params.RevocationGracePeriodSeconds)
			}

			_ = k.setRecipientKeyStore(ctx, record)

			if err := ctx.EventManager().EmitTypedEvent(&types.EventKeyExpiredPB{
				Address:     record.Address,
				Fingerprint: record.KeyFingerprint,
				ExpiredAt:   now,
			}); err != nil {
				continue
			}

			if k.hooks != nil {
				_ = k.hooks.AfterKeyExpired(ctx, addr, record.KeyFingerprint)
			}

			newKey, found := k.findLatestUsableKey(ctx, addr, record.KeyFingerprint)
			if found {
				_, _, _, _ = k.queueReencryptionJobs(ctx, record.KeyFingerprint, newKey.KeyFingerprint, params.RotationBatchSize, nil)
				_ = ctx.EventManager().EmitTypedEvent(&types.EventKeyRotatedPB{
					Address:        record.Address,
					OldFingerprint: record.KeyFingerprint,
					NewFingerprint: newKey.KeyFingerprint,
					RotatedAt:      now,
				})
			}

			continue
		}

		for _, window := range params.KeyExpiryWarningSeconds {
			if window == 0 {
				continue
			}
			if record.ExpiresAt-now > safeInt64FromUint64(window) {
				continue
			}

			warnKey := types.ExpiryWarningKey([]byte(record.KeyFingerprint), window)
			if store.Get(warnKey) != nil {
				continue
			}

			if err := ctx.EventManager().EmitTypedEvent(&types.EventKeyExpiryWarningPB{
				Address:              record.Address,
				Fingerprint:          record.KeyFingerprint,
				ExpiresAt:            record.ExpiresAt,
				WarningWindowSeconds: window,
			}); err != nil {
				continue
			}

			store.Set(warnKey, []byte{0x01})
			warnings++
		}
	}

	return warnings, expired
}

func (k Keeper) findLatestUsableKey(ctx sdk.Context, address sdk.AccAddress, excludeFingerprint string) (types.RecipientKeyRecord, bool) {
	keys := k.GetRecipientKeys(ctx, address)
	var latest types.RecipientKeyRecord
	found := false
	for _, key := range keys {
		if key.KeyFingerprint == excludeFingerprint {
			continue
		}
		if !k.isKeyUsableAt(ctx.BlockTime().Unix(), key) {
			continue
		}
		if !found || key.KeyVersion > latest.KeyVersion {
			latest = key
			found = true
		}
	}
	return latest, found
}
