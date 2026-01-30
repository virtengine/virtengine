package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// IKeeper defines the interface for the encryption keeper
type IKeeper interface {
	// Key management
	RegisterRecipientKey(ctx sdk.Context, address sdk.AccAddress, publicKey []byte, algorithmID, label string) (string, error)
	RevokeRecipientKey(ctx sdk.Context, address sdk.AccAddress, keyFingerprint string) error
	UpdateKeyLabel(ctx sdk.Context, address sdk.AccAddress, keyFingerprint, label string) error
	GetRecipientKeys(ctx sdk.Context, address sdk.AccAddress) []types.RecipientKeyRecord
	GetRecipientKeyByFingerprint(ctx sdk.Context, fingerprint string) (types.RecipientKeyRecord, bool)
	GetActiveRecipientKey(ctx sdk.Context, address sdk.AccAddress) (types.RecipientKeyRecord, bool)

	// Envelope validation
	ValidateEnvelope(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) error
	ValidateEnvelopeRecipients(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) ([]string, error)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper of the encryption store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
}

// NewKeeper creates and returns an instance for encryption keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		skey:      skey,
		authority: authority,
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

// paramsStore is the stored format of params
type paramsStore struct {
	MaxRecipientsPerEnvelope uint32   `json:"max_recipients_per_envelope"`
	MaxKeysPerAccount        uint32   `json:"max_keys_per_account"`
	AllowedAlgorithms        []string `json:"allowed_algorithms"`
	RequireSignature         bool     `json:"require_signature"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		MaxRecipientsPerEnvelope: params.MaxRecipientsPerEnvelope,
		MaxKeysPerAccount:        params.MaxKeysPerAccount,
		AllowedAlgorithms:        params.AllowedAlgorithms,
		RequireSignature:         params.RequireSignature,
	})
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey(), bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey())
	if bz == nil {
		return types.DefaultParams()
	}

	var ps paramsStore
	if err := json.Unmarshal(bz, &ps); err != nil {
		return types.DefaultParams()
	}
	return types.Params{
		MaxRecipientsPerEnvelope: ps.MaxRecipientsPerEnvelope,
		MaxKeysPerAccount:        ps.MaxKeysPerAccount,
		AllowedAlgorithms:        ps.AllowedAlgorithms,
		RequireSignature:         ps.RequireSignature,
	}
}

// recipientKeyStore is the stored format of a recipient key record
type recipientKeyStore struct {
	Address        string `json:"address"`
	PublicKey      []byte `json:"public_key"`
	KeyFingerprint string `json:"key_fingerprint"`
	AlgorithmID    string `json:"algorithm_id"`
	RegisteredAt   int64  `json:"registered_at"`
	RevokedAt      int64  `json:"revoked_at"`
	Label          string `json:"label"`
}

// RegisterRecipientKey registers a new recipient public key
func (k Keeper) RegisterRecipientKey(ctx sdk.Context, address sdk.AccAddress, publicKey []byte, algorithmID, label string) (string, error) {
	params := k.GetParams(ctx)

	// Validate algorithm is allowed
	if !types.IsAlgorithmAllowed(&params, algorithmID) {
		return "", types.ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not allowed", algorithmID)
	}

	// Validate public key size
	algInfo, err := types.GetAlgorithmInfo(algorithmID)
	if err != nil {
		return "", err
	}
	if len(publicKey) != algInfo.KeySize {
		return "", types.ErrInvalidPublicKey.Wrapf("expected %d bytes, got %d", algInfo.KeySize, len(publicKey))
	}

	// Compute key fingerprint
	fingerprint := types.ComputeKeyFingerprint(publicKey)

	// Check if key already exists (by fingerprint)
	if _, found := k.GetRecipientKeyByFingerprint(ctx, fingerprint); found {
		return "", types.ErrKeyAlreadyExists.Wrapf("key with fingerprint %s already registered", fingerprint)
	}

	// Check max keys per account
	existingKeys := k.GetRecipientKeys(ctx, address)
	if uint32(len(existingKeys)) >= params.MaxKeysPerAccount {
		return "", types.ErrInvalidPublicKey.Wrapf("account has reached max keys limit: %d", params.MaxKeysPerAccount)
	}

	store := ctx.KVStore(k.skey)

	record := recipientKeyStore{
		Address:        address.String(),
		PublicKey:      publicKey,
		KeyFingerprint: fingerprint,
		AlgorithmID:    algorithmID,
		RegisteredAt:   ctx.BlockTime().Unix(),
		RevokedAt:      0,
		Label:          label,
	}

	bz, err := json.Marshal(&record)
	if err != nil {
		return "", err
	}

	// Store by address
	store.Set(types.RecipientKeyKey(address.Bytes()), bz)

	// Store fingerprint -> address mapping
	store.Set(types.KeyByFingerprintKey([]byte(fingerprint)), address.Bytes())

	return fingerprint, nil
}

// RevokeRecipientKey revokes a recipient's public key
func (k Keeper) RevokeRecipientKey(ctx sdk.Context, address sdk.AccAddress, keyFingerprint string) error {
	store := ctx.KVStore(k.skey)

	// Get existing key record
	key := types.RecipientKeyKey(address.Bytes())
	bz := store.Get(key)
	if bz == nil {
		return types.ErrKeyNotFound.Wrapf("no key found for address %s", address.String())
	}

	var record recipientKeyStore
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.ErrKeyNotFound.Wrapf("failed to unmarshal key record: %v", err)
	}

	// Verify fingerprint matches
	if record.KeyFingerprint != keyFingerprint {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found for address %s", keyFingerprint, address.String())
	}

	// Check if already revoked
	if record.RevokedAt != 0 {
		return types.ErrKeyRevoked.Wrapf("key %s is already revoked", keyFingerprint)
	}

	// Mark as revoked
	record.RevokedAt = ctx.BlockTime().Unix()

	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// UpdateKeyLabel updates a recipient key's label
func (k Keeper) UpdateKeyLabel(ctx sdk.Context, address sdk.AccAddress, keyFingerprint, label string) error {
	store := ctx.KVStore(k.skey)

	// Get existing key record
	key := types.RecipientKeyKey(address.Bytes())
	bz := store.Get(key)
	if bz == nil {
		return types.ErrKeyNotFound.Wrapf("no key found for address %s", address.String())
	}

	var record recipientKeyStore
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.ErrKeyNotFound.Wrapf("failed to unmarshal key record: %v", err)
	}

	// Verify fingerprint matches
	if record.KeyFingerprint != keyFingerprint {
		return types.ErrKeyNotFound.Wrapf("key fingerprint %s not found for address %s", keyFingerprint, address.String())
	}

	// Update label
	record.Label = label

	bz, err := json.Marshal(&record)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// GetRecipientKeys returns all keys registered for an address
func (k Keeper) GetRecipientKeys(ctx sdk.Context, address sdk.AccAddress) []types.RecipientKeyRecord {
	store := ctx.KVStore(k.skey)

	key := types.RecipientKeyKey(address.Bytes())
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var record recipientKeyStore
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil
	}

	return []types.RecipientKeyRecord{{
		Address:        record.Address,
		PublicKey:      record.PublicKey,
		KeyFingerprint: record.KeyFingerprint,
		AlgorithmID:    record.AlgorithmID,
		RegisteredAt:   record.RegisteredAt,
		RevokedAt:      record.RevokedAt,
		Label:          record.Label,
	}}
}

// GetRecipientKeyByFingerprint returns a key record by its fingerprint
func (k Keeper) GetRecipientKeyByFingerprint(ctx sdk.Context, fingerprint string) (types.RecipientKeyRecord, bool) {
	store := ctx.KVStore(k.skey)

	// Lookup address by fingerprint
	addrBytes := store.Get(types.KeyByFingerprintKey([]byte(fingerprint)))
	if addrBytes == nil {
		return types.RecipientKeyRecord{}, false
	}

	// Get key record by address
	bz := store.Get(types.RecipientKeyKey(addrBytes))
	if bz == nil {
		return types.RecipientKeyRecord{}, false
	}

	var record recipientKeyStore
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.RecipientKeyRecord{}, false
	}

	// Verify fingerprint matches
	if record.KeyFingerprint != fingerprint {
		return types.RecipientKeyRecord{}, false
	}

	return types.RecipientKeyRecord{
		Address:        record.Address,
		PublicKey:      record.PublicKey,
		KeyFingerprint: record.KeyFingerprint,
		AlgorithmID:    record.AlgorithmID,
		RegisteredAt:   record.RegisteredAt,
		RevokedAt:      record.RevokedAt,
		Label:          record.Label,
	}, true
}

// GetActiveRecipientKey returns the active (non-revoked) key for an address
func (k Keeper) GetActiveRecipientKey(ctx sdk.Context, address sdk.AccAddress) (types.RecipientKeyRecord, bool) {
	keys := k.GetRecipientKeys(ctx, address)
	for _, key := range keys {
		if key.IsActive() {
			return key, true
		}
	}
	return types.RecipientKeyRecord{}, false
}

// ValidateEnvelope validates an encrypted payload envelope
func (k Keeper) ValidateEnvelope(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) error {
	if envelope == nil {
		return types.ErrInvalidEnvelope.Wrap("envelope cannot be nil")
	}

	// Basic structural validation
	if err := envelope.Validate(); err != nil {
		return err
	}

	params := k.GetParams(ctx)

	// Check algorithm is allowed
	if !types.IsAlgorithmAllowed(&params, envelope.AlgorithmID) {
		return types.ErrUnsupportedAlgorithm.Wrapf("algorithm %s is not allowed", envelope.AlgorithmID)
	}

	// Check max recipients
	if uint32(len(envelope.RecipientKeyIDs)) > params.MaxRecipientsPerEnvelope {
		return types.ErrMaxRecipientsExceeded.Wrapf("envelope has %d recipients, max is %d",
			len(envelope.RecipientKeyIDs), params.MaxRecipientsPerEnvelope)
	}

	// Validate signature if required
	if params.RequireSignature && len(envelope.SenderSignature) == 0 {
		return types.ErrInvalidSignature.Wrap("signature required but not provided")
	}

	return nil
}

// ValidateEnvelopeRecipients validates that all recipients have registered keys
// Returns list of missing key fingerprints
func (k Keeper) ValidateEnvelopeRecipients(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) ([]string, error) {
	var missingKeys []string

	for _, keyID := range envelope.RecipientKeyIDs {
		record, found := k.GetRecipientKeyByFingerprint(ctx, keyID)
		if !found {
			missingKeys = append(missingKeys, keyID)
			continue
		}

		if !record.IsActive() {
			return nil, types.ErrKeyRevoked.Wrapf("recipient key %s has been revoked", keyID)
		}
	}

	return missingKeys, nil
}

// WithRecipientKeys iterates all recipient keys
func (k Keeper) WithRecipientKeys(ctx sdk.Context, fn func(record types.RecipientKeyRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixRecipientKey)

	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var record recipientKeyStore
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}

		keyRecord := types.RecipientKeyRecord{
			Address:        record.Address,
			PublicKey:      record.PublicKey,
			KeyFingerprint: record.KeyFingerprint,
			AlgorithmID:    record.AlgorithmID,
			RegisteredAt:   record.RegisteredAt,
			RevokedAt:      record.RevokedAt,
			Label:          record.Label,
		}

		if stop := fn(keyRecord); stop {
			break
		}
	}
}
