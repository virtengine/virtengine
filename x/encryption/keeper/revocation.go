package keeper

import (
	"bytes"
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// RevokeRecipientKey revokes a recipient's public key.
func (k Keeper) RevokeRecipientKey(ctx sdk.Context, address sdk.AccAddress, keyFingerprint string) error {
	store := ctx.KVStore(k.skey)

	if address.String() != k.authority {
		addrBytes := store.Get(types.KeyByFingerprintKey([]byte(keyFingerprint)))
		if addrBytes != nil && !bytes.Equal(addrBytes, address.Bytes()) {
			return types.ErrUnauthorized.Wrap("sender is not key owner or authority")
		}
	}

	targetAddr := address
	if address.String() == k.authority {
		addrBytes := store.Get(types.KeyByFingerprintKey([]byte(keyFingerprint)))
		if addrBytes == nil {
			return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found", keyFingerprint)
		}
		targetAddr = sdk.AccAddress(addrBytes)
	}

	key := types.RecipientKeyKey(targetAddr.Bytes(), []byte(keyFingerprint))
	bz := store.Get(key)
	if bz == nil {
		return types.ErrKeyNotFound.Wrapf("no key found for address %s", targetAddr.String())
	}

	var record recipientKeyStore
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.ErrKeyNotFound.Wrapf("failed to unmarshal key record: %v", err)
	}

	if record.KeyFingerprint != keyFingerprint {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found for address %s", keyFingerprint, targetAddr.String())
	}

	if record.RevokedAt != 0 {
		return types.ErrKeyRevoked.Wrapf("key %s is already revoked", keyFingerprint)
	}

	record.RevokedAt = ctx.BlockTime().Unix()
	graceSeconds := k.GetParams(ctx).RevocationGracePeriodSeconds
	if graceSeconds > 0 {
		record.PurgeAt = record.RevokedAt + safeInt64FromUint64(graceSeconds)
	}

	if err := k.setRecipientKeyStore(ctx, record); err != nil {
		return err
	}

	activeKey := store.Get(types.ActiveKeyKey(targetAddr.Bytes()))
	if string(activeKey) == keyFingerprint {
		k.setLatestActiveKey(ctx, targetAddr)
	}

	if k.hooks != nil {
		_ = k.hooks.AfterKeyRevoked(ctx, targetAddr, keyFingerprint)
	}

	return nil
}

// PurgeRevokedKeys deletes key material after grace period.
func (k Keeper) PurgeRevokedKeys(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixRecipientKey)
	defer func() {
		_ = iter.Close()
	}()

	var purged uint32
	now := ctx.BlockTime().Unix()
	for ; iter.Valid(); iter.Next() {
		var record recipientKeyStore
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		if record.RevokedAt == 0 || record.PurgeAt == 0 || now < record.PurgeAt {
			continue
		}

		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			continue
		}

		k.deleteRecipientKeyStore(ctx, addr, record)
		purged++
	}

	return purged
}
