package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

func (k Keeper) getRecipientKeyStore(ctx sdk.Context, address sdk.AccAddress, fingerprint string) (recipientKeyStore, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.RecipientKeyKey(address.Bytes(), []byte(fingerprint)))
	if bz == nil {
		return recipientKeyStore{}, false
	}

	var record recipientKeyStore
	if err := json.Unmarshal(bz, &record); err != nil {
		return recipientKeyStore{}, false
	}

	if record.KeyFingerprint != fingerprint {
		return recipientKeyStore{}, false
	}

	return record, true
}

func (k Keeper) setRecipientKeyStore(ctx sdk.Context, record recipientKeyStore) error {
	addr, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}

	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.RecipientKeyKey(addr.Bytes(), []byte(record.KeyFingerprint)), bz)
	store.Set(types.KeyByFingerprintKey([]byte(record.KeyFingerprint)), addr.Bytes())
	store.Set(types.RecipientKeyVersionKey(addr.Bytes(), record.KeyVersion), []byte(record.KeyFingerprint))
	return nil
}

func (k Keeper) deleteRecipientKeyStore(ctx sdk.Context, address sdk.AccAddress, record recipientKeyStore) {
	store := ctx.KVStore(k.skey)
	store.Delete(types.RecipientKeyKey(address.Bytes(), []byte(record.KeyFingerprint)))
	store.Delete(types.KeyByFingerprintKey([]byte(record.KeyFingerprint)))
	store.Delete(types.RecipientKeyVersionKey(address.Bytes(), record.KeyVersion))
}

// ImportRecipientKeyRecord stores a recipient key record directly.
func (k Keeper) ImportRecipientKeyRecord(ctx sdk.Context, record types.RecipientKeyRecord) error {
	storeRecord := recipientKeyStore{
		Address:        record.Address,
		PublicKey:      record.PublicKey,
		KeyFingerprint: record.KeyFingerprint,
		KeyVersion:     record.KeyVersion,
		AlgorithmID:    record.AlgorithmID,
		RegisteredAt:   record.RegisteredAt,
		RevokedAt:      record.RevokedAt,
		DeprecatedAt:   record.DeprecatedAt,
		ExpiresAt:      record.ExpiresAt,
		PurgeAt:        record.PurgeAt,
		Label:          record.Label,
	}

	return k.setRecipientKeyStore(ctx, storeRecord)
}
