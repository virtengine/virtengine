package keeper

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	decredsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

type IKeeper interface {
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
	Get(ctx sdk.Context, id sdk.Address) (types.Provider, bool)
	Create(ctx sdk.Context, provider types.Provider) error
	WithProviders(ctx sdk.Context, fn func(types.Provider) bool)
	Update(ctx sdk.Context, provider types.Provider) error
	Delete(ctx sdk.Context, id sdk.Address)
	NewQuerier() Querier
	// Additional methods for cross-module integration
	ProviderExists(ctx sdk.Context, providerAddr sdk.AccAddress) bool
	GetProviderPublicKey(ctx sdk.Context, providerAddr sdk.AccAddress) ([]byte, bool)
	IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool
	// Public key management methods
	SetProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, pubKey []byte, keyType string) error
	GetProviderPublicKeyRecord(ctx sdk.Context, owner sdk.AccAddress) (types.ProviderPublicKeyRecord, bool)
	GetProviderSigningKey(ctx sdk.Context, owner sdk.AccAddress, keyID string, epoch uint64) (types.ProviderPublicKeyRecord, bool)
	GetProviderSigningKeyEpochs(ctx sdk.Context, owner sdk.AccAddress) []types.ProviderPublicKeyRecord
	WithProviderSigningKeys(ctx sdk.Context, fn func(sdk.AccAddress, types.ProviderPublicKeyRecord) bool)
	RotateProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, newKey []byte, keyType string, signature []byte) error
	RevokeProviderSigningKey(ctx sdk.Context, owner sdk.AccAddress, keyID string) error
	DeleteProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress)
	WithProviderPublicKeys(ctx sdk.Context, fn func(sdk.AccAddress, types.ProviderPublicKeyRecord) bool)
	MigrateSigningKeyEpochs(ctx sdk.Context) error
	ImportProviderSigningKeyEpoch(ctx sdk.Context, owner sdk.AccAddress, record types.ProviderPublicKeyRecord, current bool) error
	// Domain verification methods
	GenerateDomainVerificationToken(ctx sdk.Context, providerAddr sdk.AccAddress, domain string) (*DomainVerificationRecord, error)
	VerifyProviderDomain(ctx sdk.Context, providerAddr sdk.AccAddress) error
	GetDomainVerificationRecord(ctx sdk.Context, providerAddr sdk.AccAddress) (*DomainVerificationRecord, bool)
	IsDomainVerified(ctx sdk.Context, providerAddr sdk.AccAddress) bool
	DeleteDomainVerificationRecord(ctx sdk.Context, providerAddr sdk.AccAddress)
	RequestDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress, domain string, method types.VerificationMethod) (*DomainVerificationRecord, string, error)
	ConfirmDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress, proof string) error
	RevokeDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress) error
}

// Keeper of the provider store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec
}

// NewKeeper creates and returns an instance for Provider keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey) IKeeper {
	return Keeper{
		skey: skey,
		cdc:  cdc,
	}
}

func (k Keeper) NewQuerier() Querier {
	return Querier{k}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// Get returns a provider with given provider id
func (k Keeper) Get(ctx sdk.Context, id sdk.Address) (types.Provider, bool) {
	store := ctx.KVStore(k.skey)
	key := ProviderKey(id)

	if !store.Has(key) {
		return types.Provider{}, false
	}

	buf := store.Get(key)
	var val types.Provider
	k.cdc.MustUnmarshal(buf, &val)
	return val, true
}

// Create creates a new provider or returns an error if the provider exists already
func (k Keeper) Create(ctx sdk.Context, provider types.Provider) error {
	store := ctx.KVStore(k.skey)
	owner, err := sdk.AccAddressFromBech32(provider.Owner)
	if err != nil {
		return err
	}

	key := ProviderKey(owner)

	if store.Has(key) {
		return types.ErrProviderExists
	}

	store.Set(key, k.cdc.MustMarshal(&provider))

	err = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderCreated{
			Owner: owner.String(),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// WithProviders iterates all providers
func (k Keeper) WithProviders(ctx sdk.Context, fn func(types.Provider) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.ProviderPrefix())

	iter := store.Iterator(nil, nil)
	defer func() {
		_ = iter.Close()
	}()
	for ; iter.Valid(); iter.Next() {
		var val types.Provider
		k.cdc.MustUnmarshal(iter.Value(), &val)
		if stop := fn(val); stop {
			break
		}
	}
}

// Update updates a provider details
func (k Keeper) Update(ctx sdk.Context, provider types.Provider) error {
	store := ctx.KVStore(k.skey)
	owner, err := sdk.AccAddressFromBech32(provider.Owner)
	if err != nil {
		return err
	}

	key := ProviderKey(owner)

	if !store.Has(key) {
		return types.ErrProviderNotFound
	}
	store.Set(key, k.cdc.MustMarshal(&provider))

	err = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderUpdated{
			Owner: owner.String(),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a provider from the store and emits a deletion event.
// If the provider does not exist, this is a no-op.
func (k Keeper) Delete(ctx sdk.Context, id sdk.Address) {
	store := ctx.KVStore(k.skey)
	key := ProviderKey(id)

	if !store.Has(key) {
		return
	}

	// Retrieve provider before deletion to get the owner address for the event
	provider, found := k.Get(ctx, id)
	store.Delete(key)

	// Use provider.Owner if available, otherwise derive from id bytes
	owner := sdk.AccAddress(id.Bytes()).String()
	if found && provider.Owner != "" {
		owner = provider.Owner
	}

	_ = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderDeleted{
			Owner: owner,
		},
	)
}

// ProviderExists checks if a provider exists
func (k Keeper) ProviderExists(ctx sdk.Context, providerAddr sdk.AccAddress) bool {
	_, exists := k.Get(ctx, providerAddr)
	return exists
}

// GetProviderPublicKey returns the public key for a provider.
// This is used for benchmark signature verification and encrypted communication.
func (k Keeper) GetProviderPublicKey(ctx sdk.Context, providerAddr sdk.AccAddress) ([]byte, bool) {
	record, found := k.GetProviderPublicKeyRecord(ctx, providerAddr)
	if !found {
		return nil, false
	}
	return record.PublicKey, true
}

// GetProviderPublicKeyRecord returns the full public key record for a provider
func (k Keeper) GetProviderPublicKeyRecord(ctx sdk.Context, owner sdk.AccAddress) (types.ProviderPublicKeyRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := ProviderPublicKeyKey(owner)

	bz := store.Get(key)
	if bz == nil {
		return types.ProviderPublicKeyRecord{}, false
	}

	var record types.ProviderPublicKeyRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		// Log error but return not found to avoid breaking callers
		return types.ProviderPublicKeyRecord{}, false
	}
	return record.NormalizeLegacy(), true
}

// SetProviderPublicKey stores a public key for a provider.
// The provider must already exist in the store.
func (k Keeper) SetProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, pubKey []byte, keyType string) error {
	// Verify provider exists
	if !k.ProviderExists(ctx, owner) {
		return types.ErrProviderNotFound.Wrapf("cannot set public key for non-existent provider: %s", owner.String())
	}

	// Create and validate the record
	activationHeight := ctx.BlockHeight()
	if activationHeight <= 0 {
		activationHeight = 1
	}
	record := types.NewProviderPublicKeyRecord(pubKey, keyType, activationHeight)
	record.ActivatedAtUnix = ctx.BlockTime().Unix()
	if err := record.Validate(); err != nil {
		return err
	}

	// Existing keys may only change through RotateProviderPublicKey, which
	// requires a detached proof from the retiring key.
	existingRecord, found := k.GetProviderPublicKeyRecord(ctx, owner)
	if found {
		if existingRecord.KeyType == keyType && bytes.Equal(existingRecord.PublicKey, pubKey) {
			return nil
		}
		return types.ErrInvalidRotationSignature.Wrap("existing provider key must be changed through proof-based rotation")
	}

	// Store the record
	store := ctx.KVStore(k.skey)
	key := ProviderPublicKeyKey(owner)

	bz, err := json.Marshal(&record)
	if err != nil {
		return types.ErrInternal.Wrapf("failed to marshal public key record: %v", err)
	}
	store.Set(key, bz)
	store.Set(ProviderSigningKeyEpochKey(owner, record.Epoch), bz)

	// Emit event
	_ = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderUpdated{
			Owner: owner.String(),
		},
	)

	return nil
}

// RotateProviderPublicKey rotates a provider's public key with signature verification.
// The signature must be created by signing the new key with the old key.
func (k Keeper) RotateProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, newKey []byte, keyType string, signature []byte) error {
	// Verify provider exists
	if !k.ProviderExists(ctx, owner) {
		return types.ErrProviderNotFound.Wrapf("cannot rotate key for non-existent provider: %s", owner.String())
	}

	// Get existing key for signature verification
	existingRecord, found := k.GetProviderPublicKeyRecord(ctx, owner)
	if !found || len(existingRecord.PublicKey) == 0 {
		return k.SetProviderPublicKey(ctx, owner, newKey, keyType)
	}
	if !existingRecord.IsSigningAlgorithm() || keyType == types.PublicKeyTypeX25519 {
		return types.ErrInvalidPublicKeyType.Wrap("x25519 cannot authenticate provider signing-key rotation")
	}
	activationUnix := ctx.BlockTime().Unix()
	rotationBytes, err := types.ProviderKeyRotationSignBytes(types.ProviderKeyRotationPayload{
		ChainID:          ctx.ChainID(),
		Provider:         owner.String(),
		OldKeyID:         existingRecord.KeyID,
		OldEpoch:         existingRecord.Epoch,
		NewKeyType:       keyType,
		NewPublicKey:     newKey,
		NewEpoch:         existingRecord.Epoch + 1,
		ActivationHeight: ctx.BlockHeight(),
		ActivationUnix:   activationUnix,
		OverlapEndHeight: ctx.BlockHeight() + types.ProviderSigningKeyOverlapBlocks,
		OverlapEndUnix:   activationUnix + types.ProviderSigningKeyOverlapSeconds,
		SignatureVersion: types.ProviderKeyRotationSignatureVersionV1,
	})
	if err != nil {
		return err
	}
	if !verifyProviderDetachedSignature(existingRecord, rotationBytes, signature) {
		return types.ErrInvalidRotationSignature.Wrap("signature verification failed")
	}
	// If no existing key, allow setting without signature (first-time setup)

	newRecord := types.NewProviderPublicKeyRecord(newKey, keyType, ctx.BlockHeight())
	newRecord.ActivatedAtUnix = activationUnix
	newRecord.Epoch = existingRecord.Epoch + 1
	newRecord.RotationCount = existingRecord.RotationCount + 1
	newRecord.PreviousKeyID = existingRecord.KeyID
	if err := newRecord.Validate(); err != nil {
		return err
	}

	existingRecord.RetiredAtHeight = ctx.BlockHeight() + types.ProviderSigningKeyOverlapBlocks
	existingRecord.RetiredAtUnix = activationUnix + types.ProviderSigningKeyOverlapSeconds
	existingRecord.UpdatedAt = ctx.BlockHeight()

	store := ctx.KVStore(k.skey)
	oldBytes, err := json.Marshal(&existingRecord)
	if err != nil {
		return types.ErrInternal.Wrapf("failed to marshal retiring key: %v", err)
	}
	newBytes, err := json.Marshal(&newRecord)
	if err != nil {
		return types.ErrInternal.Wrapf("failed to marshal active key: %v", err)
	}
	store.Set(ProviderSigningKeyEpochKey(owner, existingRecord.Epoch), oldBytes)
	store.Set(ProviderSigningKeyEpochKey(owner, newRecord.Epoch), newBytes)
	store.Set(ProviderPublicKeyKey(owner), newBytes)
	return nil
}

func verifyProviderDetachedSignature(record types.ProviderPublicKeyRecord, message, signature []byte) bool {
	switch record.KeyType {
	case types.PublicKeyTypeEd25519:
		if len(record.PublicKey) != ed25519.PublicKeySize || len(signature) != ed25519.SignatureSize {
			return false
		}
		return ed25519.Verify(record.PublicKey, message, signature)
	case types.PublicKeyTypeSecp256k1:
		if len(signature) != 64 {
			return false
		}
		var s decredsecp256k1.ModNScalar
		if s.SetByteSlice(signature[32:]) || s.IsZero() || s.IsOverHalfOrder() {
			return false
		}
		return (&cosmossecp256k1.PubKey{Key: record.PublicKey}).VerifySignature(message, signature)
	default:
		return false
	}
}

// GetProviderSigningKey returns an exact immutable epoch record.
func (k Keeper) GetProviderSigningKey(ctx sdk.Context, owner sdk.AccAddress, keyID string, epoch uint64) (types.ProviderPublicKeyRecord, bool) {
	if epoch == 0 || keyID == "" {
		return types.ProviderPublicKeyRecord{}, false
	}
	store := ctx.KVStore(k.skey)
	bz := store.Get(ProviderSigningKeyEpochKey(owner, epoch))
	if bz == nil {
		current, found := k.GetProviderPublicKeyRecord(ctx, owner)
		if !found || current.Epoch != epoch || current.KeyID != keyID {
			return types.ProviderPublicKeyRecord{}, false
		}
		return current, true
	}
	var record types.ProviderPublicKeyRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.ProviderPublicKeyRecord{}, false
	}
	record = record.NormalizeLegacy()
	if record.KeyID != keyID || record.Epoch != epoch {
		return types.ProviderPublicKeyRecord{}, false
	}
	return record, true
}

// GetProviderSigningKeyEpochs lists a provider's epoch history in store order.
func (k Keeper) GetProviderSigningKeyEpochs(ctx sdk.Context, owner sdk.AccAddress) []types.ProviderPublicKeyRecord {
	store := ctx.KVStore(k.skey)
	prefixKey := ProviderSigningKeyEpochOwnerPrefix(owner)
	iter := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iter.Close()
	records := make([]types.ProviderPublicKeyRecord, 0)
	for ; iter.Valid(); iter.Next() {
		var record types.ProviderPublicKeyRecord
		if err := json.Unmarshal(iter.Value(), &record); err == nil {
			records = append(records, record.NormalizeLegacy())
		}
	}
	return records
}

// WithProviderSigningKeys iterates all immutable provider key epochs.
func (k Keeper) WithProviderSigningKeys(ctx sdk.Context, fn func(sdk.AccAddress, types.ProviderPublicKeyRecord) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.ProviderSigningKeyEpochPrefix())
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) < 1+8 {
			continue
		}
		addressLength := int(key[0])
		if addressLength <= 0 || len(key) != 1+addressLength+8 {
			continue
		}
		var record types.ProviderPublicKeyRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if fn(sdk.AccAddress(key[1:1+addressLength]), record.NormalizeLegacy()) {
			return
		}
	}
}

// RevokeProviderSigningKey permanently revokes an exact key epoch.
func (k Keeper) RevokeProviderSigningKey(ctx sdk.Context, owner sdk.AccAddress, keyID string) error {
	current, found := k.GetProviderPublicKeyRecord(ctx, owner)
	if !found || current.KeyID != keyID {
		return types.ErrInvalidPublicKey.Wrap("active signing key not found")
	}
	current.RevokedAtHeight = ctx.BlockHeight()
	current.RevokedAtUnix = ctx.BlockTime().Unix()
	current.UpdatedAt = ctx.BlockHeight()
	bz, err := json.Marshal(&current)
	if err != nil {
		return types.ErrInternal.Wrapf("failed to marshal revoked key: %v", err)
	}
	store := ctx.KVStore(k.skey)
	store.Set(ProviderSigningKeyEpochKey(owner, current.Epoch), bz)
	store.Set(ProviderPublicKeyKey(owner), bz)
	return nil
}

// DeleteProviderPublicKey removes a provider's public key from storage
func (k Keeper) DeleteProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress) {
	store := ctx.KVStore(k.skey)
	key := ProviderPublicKeyKey(owner)

	if !store.Has(key) {
		return
	}

	if current, found := k.GetProviderPublicKeyRecord(ctx, owner); found {
		current.RevokedAtHeight = ctx.BlockHeight()
		current.RevokedAtUnix = ctx.BlockTime().Unix()
		if bz, err := json.Marshal(&current); err == nil {
			store.Set(ProviderSigningKeyEpochKey(owner, current.Epoch), bz)
		}
	}
	store.Delete(key)

	_ = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderUpdated{
			Owner: owner.String(),
		},
	)
}

// WithProviderPublicKeys iterates over all provider public keys
func (k Keeper) WithProviderPublicKeys(ctx sdk.Context, fn func(sdk.AccAddress, types.ProviderPublicKeyRecord) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.ProviderPublicKeyPrefix())

	iter := store.Iterator(nil, nil)
	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var record types.ProviderPublicKeyRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		// Extract address from key (skip length prefix byte)
		keyBytes := iter.Key()
		if len(keyBytes) < 2 {
			continue
		}
		addrLen := int(keyBytes[0])
		if len(keyBytes) < 1+addrLen {
			continue
		}
		addr := sdk.AccAddress(keyBytes[1 : 1+addrLen])

		if stop := fn(addr, record.NormalizeLegacy()); stop {
			break
		}
	}
}

// IsProvider checks if an address is a registered provider
func (k Keeper) IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool {
	return k.ProviderExists(ctx, addr)
}

// MigrateSigningKeyEpochs backfills immutable history for pre-84B current keys.
func (k Keeper) MigrateSigningKeyEpochs(ctx sdk.Context) error {
	store := ctx.KVStore(k.skey)
	var migrateErr error
	k.WithProviderPublicKeys(ctx, func(owner sdk.AccAddress, record types.ProviderPublicKeyRecord) bool {
		record = record.NormalizeLegacy()
		if record.ActivatedAtUnix == 0 {
			record.ActivatedAtUnix = ctx.BlockTime().Unix()
		}
		bz, err := json.Marshal(&record)
		if err != nil {
			migrateErr = err
			return true
		}
		store.Set(ProviderPublicKeyKey(owner), bz)
		store.Set(ProviderSigningKeyEpochKey(owner, record.Epoch), bz)
		return false
	})
	return migrateErr
}

// ImportProviderSigningKeyEpoch imports validated genesis history.
func (k Keeper) ImportProviderSigningKeyEpoch(ctx sdk.Context, owner sdk.AccAddress, record types.ProviderPublicKeyRecord, current bool) error {
	record = record.NormalizeLegacy()
	if err := record.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.skey)
	store.Set(ProviderSigningKeyEpochKey(owner, record.Epoch), bz)
	if current {
		store.Set(ProviderPublicKeyKey(owner), bz)
	}
	return nil
}
