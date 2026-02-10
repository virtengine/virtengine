package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	"github.com/virtengine/virtengine/x/encryption/types"
)

// EphemeralKeyRecord tracks a session-based ephemeral key.
type EphemeralKeyRecord struct {
	SessionID string `json:"session_id"`
	Address   string `json:"address"`
	PublicKey []byte `json:"public_key"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
	UsedAt    int64  `json:"used_at,omitempty"`
}

// CreateEphemeralKey generates a new ephemeral keypair for a session.
func (k Keeper) CreateEphemeralKey(ctx sdk.Context, address sdk.AccAddress, ttlSeconds uint64) (string, []byte, []byte, error) {
	keyPair, err := encryptioncrypto.GenerateKeyPair()
	if err != nil {
		return "", nil, nil, err
	}

	sessionID := buildEphemeralSessionID(address.String(), keyPair.PublicKey[:], ctx.BlockHeight())
	record := EphemeralKeyRecord{
		SessionID: sessionID,
		Address:   address.String(),
		PublicKey: append([]byte(nil), keyPair.PublicKey[:]...),
		CreatedAt: ctx.BlockTime().Unix(),
		ExpiresAt: 0,
	}

	if ttlSeconds > 0 {
		ttl := safeInt64FromUint64(ttlSeconds)
		record.ExpiresAt = ctx.BlockTime().Add(time.Duration(ttl) * time.Second).Unix()
	}

	if err := k.setEphemeralKeyRecord(ctx, record); err != nil {
		return "", nil, nil, err
	}

	return sessionID, record.PublicKey, keyPair.PrivateKey[:], nil
}

// UseEphemeralKey marks an ephemeral key as used.
func (k Keeper) UseEphemeralKey(ctx sdk.Context, sessionID string) error {
	record, found := k.GetEphemeralKey(ctx, sessionID)
	if !found {
		return types.ErrKeyNotFound.Wrap("ephemeral session not found")
	}

	now := ctx.BlockTime().Unix()
	if record.ExpiresAt != 0 && now >= record.ExpiresAt {
		return types.ErrKeyExpired.Wrap("ephemeral key expired")
	}
	if record.UsedAt != 0 {
		return types.ErrKeyRevoked.Wrap("ephemeral key already used")
	}

	record.UsedAt = now
	return k.setEphemeralKeyRecord(ctx, record)
}

// GetEphemeralKey returns an ephemeral key record.
func (k Keeper) GetEphemeralKey(ctx sdk.Context, sessionID string) (EphemeralKeyRecord, bool) {
	bz := ctx.KVStore(k.skey).Get(types.EphemeralKeyRecordKey([]byte(sessionID)))
	if bz == nil {
		return EphemeralKeyRecord{}, false
	}

	var record EphemeralKeyRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return EphemeralKeyRecord{}, false
	}
	return record, true
}

// CleanupEphemeralKeys removes expired or used ephemeral keys.
func (k Keeper) CleanupEphemeralKeys(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixEphemeralKey)
	defer func() {
		_ = iter.Close()
	}()

	now := ctx.BlockTime().Unix()
	var cleaned uint32
	for ; iter.Valid(); iter.Next() {
		var record EphemeralKeyRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		if record.UsedAt != 0 || (record.ExpiresAt != 0 && now >= record.ExpiresAt) {
			store.Delete(iter.Key())
			cleaned++
		}
	}

	return cleaned
}

// DeriveSharedSecret derives a shared secret using an ephemeral private key and peer public key.
func (k Keeper) DeriveSharedSecret(ephemeralPrivateKey, peerPublicKey []byte) ([]byte, error) {
	return encryptioncrypto.DeriveSharedSecret(ephemeralPrivateKey, peerPublicKey)
}

func (k Keeper) setEphemeralKeyRecord(ctx sdk.Context, record EphemeralKeyRecord) error {
	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.EphemeralKeyRecordKey([]byte(record.SessionID)), bz)
	return nil
}

func buildEphemeralSessionID(address string, pubKey []byte, height int64) string {
	h := sha256.New()
	h.Write([]byte(address))
	h.Write(pubKey)
	h.Write([]byte{byte(height >> 56), byte(height >> 48), byte(height >> 40), byte(height >> 32), byte(height >> 24), byte(height >> 16), byte(height >> 8), byte(height)})
	return hex.EncodeToString(h.Sum(nil))
}
