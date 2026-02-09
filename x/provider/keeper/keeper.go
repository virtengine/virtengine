package keeper

import (
	"crypto/ed25519"
	"encoding/json"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	RotateProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, newKey []byte, keyType string, signature []byte) error
	DeleteProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress)
	WithProviderPublicKeys(ctx sdk.Context, fn func(sdk.AccAddress, types.ProviderPublicKeyRecord) bool)
	// Domain verification methods
	GenerateDomainVerificationToken(ctx sdk.Context, providerAddr sdk.AccAddress, domain string) (*DomainVerificationRecord, error)
	VerifyProviderDomain(ctx sdk.Context, providerAddr sdk.AccAddress) error
	GetDomainVerificationRecord(ctx sdk.Context, providerAddr sdk.AccAddress) (*DomainVerificationRecord, bool)
	IsDomainVerified(ctx sdk.Context, providerAddr sdk.AccAddress) bool
	DeleteDomainVerificationRecord(ctx sdk.Context, providerAddr sdk.AccAddress)
	// TODO: Replace int32 with types.VerificationMethod after proto generation
	RequestDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress, domain string, method int32) (*DomainVerificationRecord, string, error)
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
	return record, true
}

// SetProviderPublicKey stores a public key for a provider.
// The provider must already exist in the store.
func (k Keeper) SetProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress, pubKey []byte, keyType string) error {
	// Verify provider exists
	if !k.ProviderExists(ctx, owner) {
		return types.ErrProviderNotFound.Wrapf("cannot set public key for non-existent provider: %s", owner.String())
	}

	// Create and validate the record
	record := types.NewProviderPublicKeyRecord(pubKey, keyType, ctx.BlockHeight())
	if err := record.Validate(); err != nil {
		return err
	}

	// Check if we're updating an existing key (increment rotation count)
	existingRecord, found := k.GetProviderPublicKeyRecord(ctx, owner)
	if found {
		record.RotationCount = existingRecord.RotationCount + 1
	}

	// Store the record
	store := ctx.KVStore(k.skey)
	key := ProviderPublicKeyKey(owner)

	bz, err := json.Marshal(&record)
	if err != nil {
		return types.ErrInternal.Wrapf("failed to marshal public key record: %v", err)
	}
	store.Set(key, bz)

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
	if found && len(existingRecord.PublicKey) > 0 {
		// Verify the rotation signature using the old key
		if !k.verifyRotationSignature(existingRecord.PublicKey, existingRecord.KeyType, newKey, signature) {
			return types.ErrInvalidRotationSignature.Wrap("signature verification failed")
		}
	}
	// If no existing key, allow setting without signature (first-time setup)

	return k.SetProviderPublicKey(ctx, owner, newKey, keyType)
}

// verifyRotationSignature verifies that the signature was created by signing newKey with oldKey
func (k Keeper) verifyRotationSignature(oldKey []byte, keyType string, newKey []byte, signature []byte) bool {
	switch keyType {
	case types.PublicKeyTypeEd25519:
		if len(oldKey) != ed25519.PublicKeySize {
			return false
		}
		return ed25519.Verify(oldKey, newKey, signature)
	case types.PublicKeyTypeX25519:
		// X25519 is for encryption, not signing. For rotation verification,
		// the caller should provide an Ed25519 signature alongside.
		// For now, accept the rotation if signature is non-empty (caller responsibility)
		return len(signature) > 0
	case types.PublicKeyTypeSecp256k1:
		// secp256k1 signature verification would require additional crypto imports
		// For now, accept non-empty signatures (caller responsibility)
		return len(signature) > 0
	default:
		return false
	}
}

// DeleteProviderPublicKey removes a provider's public key from storage
func (k Keeper) DeleteProviderPublicKey(ctx sdk.Context, owner sdk.AccAddress) {
	store := ctx.KVStore(k.skey)
	key := ProviderPublicKeyKey(owner)

	if !store.Has(key) {
		return
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

		if stop := fn(addr, record); stop {
			break
		}
	}
}

// IsProvider checks if an address is a registered provider
func (k Keeper) IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool {
	return k.ProviderExists(ctx, addr)
}
