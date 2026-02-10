package keeper

import (
	"encoding/json"
	"math"
	"time"

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
	RotateRecipientKey(ctx sdk.Context, address sdk.AccAddress, oldFingerprint string, newPublicKey []byte, newAlgorithmID, newLabel, reason string, newKeyTTLSeconds uint64) (string, error)
	GetRecipientKeys(ctx sdk.Context, address sdk.AccAddress) []types.RecipientKeyRecord
	GetRecipientKeyByFingerprint(ctx sdk.Context, fingerprint string) (types.RecipientKeyRecord, bool)
	GetActiveRecipientKey(ctx sdk.Context, address sdk.AccAddress) (types.RecipientKeyRecord, bool)

	// Envelope validation
	ValidateEnvelope(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) error
	ValidateEnvelopeRecipients(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) ([]string, error)

	// Access control
	CheckEnvelopeAccess(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, requester sdk.AccAddress) error
	CheckEnvelopeAccessByFingerprint(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, keyFingerprint string) error
	ValidateAndCheckAccess(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, requester sdk.AccAddress) error
	GetEnvelopeRecipients(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope) ([]sdk.AccAddress, error)
	EnforceEncryptedPayloadRequired(envelope *types.EncryptedPayloadEnvelope, fieldName string) error

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

	hooks types.EncryptionHooks
}

// NewKeeper creates and returns an instance for encryption keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		skey:      skey,
		authority: authority,
	}
}

// SetHooks sets the encryption hooks.
func (k *Keeper) SetHooks(hooks types.EncryptionHooks) *Keeper {
	k.hooks = hooks
	return k
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
	RevocationGracePeriod    uint64   `json:"revocation_grace_period_seconds"`
	KeyExpiryWarningSeconds  []uint64 `json:"key_expiry_warning_seconds"`
	RotationBatchSize        uint32   `json:"rotation_batch_size"`
	DefaultKeyTTLSeconds     uint64   `json:"default_key_ttl_seconds"`
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
		RevocationGracePeriod:    params.RevocationGracePeriodSeconds,
		KeyExpiryWarningSeconds:  params.KeyExpiryWarningSeconds,
		RotationBatchSize:        params.RotationBatchSize,
		DefaultKeyTTLSeconds:     params.DefaultKeyTtlSeconds,
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
		MaxRecipientsPerEnvelope:     ps.MaxRecipientsPerEnvelope,
		MaxKeysPerAccount:            ps.MaxKeysPerAccount,
		AllowedAlgorithms:            ps.AllowedAlgorithms,
		RequireSignature:             ps.RequireSignature,
		RevocationGracePeriodSeconds: ps.RevocationGracePeriod,
		KeyExpiryWarningSeconds:      ps.KeyExpiryWarningSeconds,
		RotationBatchSize:            ps.RotationBatchSize,
		DefaultKeyTtlSeconds:         ps.DefaultKeyTTLSeconds,
	}
}

// recipientKeyStore is the stored format of a recipient key record
type recipientKeyStore struct {
	Address        string `json:"address"`
	PublicKey      []byte `json:"public_key"`
	KeyFingerprint string `json:"key_fingerprint"`
	KeyVersion     uint32 `json:"key_version"`
	AlgorithmID    string `json:"algorithm_id"`
	RegisteredAt   int64  `json:"registered_at"`
	RevokedAt      int64  `json:"revoked_at"`
	DeprecatedAt   int64  `json:"deprecated_at"`
	ExpiresAt      int64  `json:"expires_at"`
	PurgeAt        int64  `json:"purge_at"`
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
	if len(existingKeys) > int(^uint32(0)) {
		return "", types.ErrInvalidPublicKey.Wrap("existing keys overflow")
	}
	if safeUint32FromInt(len(existingKeys)) >= params.MaxKeysPerAccount {
		return "", types.ErrInvalidPublicKey.Wrapf("account has reached max keys limit: %d", params.MaxKeysPerAccount)
	}

	store := ctx.KVStore(k.skey)

	nextVersion := uint32(1)
	for _, key := range existingKeys {
		if key.KeyVersion >= nextVersion {
			nextVersion = key.KeyVersion + 1
		}
	}

	record := recipientKeyStore{
		Address:        address.String(),
		PublicKey:      publicKey,
		KeyFingerprint: fingerprint,
		KeyVersion:     nextVersion,
		AlgorithmID:    algorithmID,
		RegisteredAt:   ctx.BlockTime().Unix(),
		RevokedAt:      0,
		DeprecatedAt:   0,
		ExpiresAt:      0,
		PurgeAt:        0,
		Label:          label,
	}

	if params.DefaultKeyTtlSeconds > 0 {
		ttl := safeInt64FromUint64(params.DefaultKeyTtlSeconds)
		record.ExpiresAt = ctx.BlockTime().Add(time.Duration(ttl) * time.Second).Unix()
	}

	bz, err := json.Marshal(&record)
	if err != nil {
		return "", err
	}

	// Store by address + fingerprint
	store.Set(types.RecipientKeyKey(address.Bytes(), []byte(fingerprint)), bz)

	// Store fingerprint -> address mapping
	store.Set(types.KeyByFingerprintKey([]byte(fingerprint)), address.Bytes())

	// Store version -> fingerprint mapping
	store.Set(types.RecipientKeyVersionKey(address.Bytes(), nextVersion), []byte(fingerprint))

	// Update active key pointer
	store.Set(types.ActiveKeyKey(address.Bytes()), []byte(fingerprint))

	return fingerprint, nil
}

// UpdateKeyLabel updates a recipient key's label
func (k Keeper) UpdateKeyLabel(ctx sdk.Context, address sdk.AccAddress, keyFingerprint, label string) error {
	store := ctx.KVStore(k.skey)

	// Get existing key record
	key := types.RecipientKeyKey(address.Bytes(), []byte(keyFingerprint))
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

	iter := storetypes.KVStorePrefixIterator(store, types.RecipientKeyPrefix(address.Bytes()))
	defer func() {
		_ = iter.Close()
	}()

	keys := make([]types.RecipientKeyRecord, 0)
	for ; iter.Valid(); iter.Next() {
		var record recipientKeyStore
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		keys = append(keys, types.RecipientKeyRecord{
			Address:        record.Address,
			PublicKey:      record.PublicKey,
			KeyFingerprint: record.KeyFingerprint,
			AlgorithmID:    record.AlgorithmID,
			RegisteredAt:   record.RegisteredAt,
			RevokedAt:      record.RevokedAt,
			DeprecatedAt:   record.DeprecatedAt,
			ExpiresAt:      record.ExpiresAt,
			PurgeAt:        record.PurgeAt,
			Label:          record.Label,
			KeyVersion:     record.KeyVersion,
		})
	}

	return keys
}

// GetRecipientKeyByFingerprint returns a key record by its fingerprint
func (k Keeper) GetRecipientKeyByFingerprint(ctx sdk.Context, fingerprint string) (types.RecipientKeyRecord, bool) {
	store := ctx.KVStore(k.skey)

	// Lookup address by fingerprint
	addrBytes := store.Get(types.KeyByFingerprintKey([]byte(fingerprint)))
	if addrBytes == nil {
		return types.RecipientKeyRecord{}, false
	}

	// Get key record by address + fingerprint
	bz := store.Get(types.RecipientKeyKey(addrBytes, []byte(fingerprint)))
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
		DeprecatedAt:   record.DeprecatedAt,
		ExpiresAt:      record.ExpiresAt,
		PurgeAt:        record.PurgeAt,
		Label:          record.Label,
		KeyVersion:     record.KeyVersion,
	}, true
}

// GetRecipientKeyByVersion returns a key record by address and version.
func (k Keeper) GetRecipientKeyByVersion(ctx sdk.Context, address sdk.AccAddress, version uint32) (types.RecipientKeyRecord, bool) {
	store := ctx.KVStore(k.skey)
	fp := store.Get(types.RecipientKeyVersionKey(address.Bytes(), version))
	if len(fp) == 0 {
		return types.RecipientKeyRecord{}, false
	}
	return k.GetRecipientKeyByFingerprint(ctx, string(fp))
}

// ResolveRecipientKeyID resolves a recipient key ID (fingerprint or versioned) to a key record.
func (k Keeper) ResolveRecipientKeyID(ctx sdk.Context, address sdk.AccAddress, keyID string) (types.RecipientKeyRecord, bool) {
	fingerprint, version, ok := types.ParseRecipientKeyID(keyID)
	if ok {
		return k.GetRecipientKeyByVersion(ctx, address, version)
	}
	return k.GetRecipientKeyByFingerprint(ctx, fingerprint)
}

// GetActiveRecipientKey returns the active (non-revoked) key for an address
func (k Keeper) GetActiveRecipientKey(ctx sdk.Context, address sdk.AccAddress) (types.RecipientKeyRecord, bool) {
	store := ctx.KVStore(k.skey)
	activeFingerprint := store.Get(types.ActiveKeyKey(address.Bytes()))
	if len(activeFingerprint) > 0 {
		record, found := k.GetRecipientKeyByFingerprint(ctx, string(activeFingerprint))
		if found && k.isKeyUsableAt(ctx.BlockTime().Unix(), record) {
			return record, true
		}
	}

	k.setLatestActiveKey(ctx, address)
	activeFingerprint = store.Get(types.ActiveKeyKey(address.Bytes()))
	if len(activeFingerprint) > 0 {
		record, found := k.GetRecipientKeyByFingerprint(ctx, string(activeFingerprint))
		if found && k.isKeyUsableAt(ctx.BlockTime().Unix(), record) {
			return record, true
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
	if len(envelope.RecipientKeyIDs) > int(^uint32(0)) {
		return types.ErrInvalidEnvelope.Wrap("recipient count overflow")
	}
	if safeUint32FromInt(len(envelope.RecipientKeyIDs)) > params.MaxRecipientsPerEnvelope {
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
		fingerprint := types.NormalizeRecipientKeyID(keyID)
		record, found := k.GetRecipientKeyByFingerprint(ctx, fingerprint)
		if !found {
			missingKeys = append(missingKeys, keyID)
			continue
		}

		if !k.isKeyUsableAt(ctx.BlockTime().Unix(), record) {
			if record.RevokedAt != 0 {
				return nil, types.ErrKeyRevoked.Wrapf("recipient key %s has been revoked", keyID)
			}
			if record.DeprecatedAt != 0 {
				return nil, types.ErrKeyDeprecated.Wrapf("recipient key %s is deprecated", keyID)
			}
			if record.ExpiresAt != 0 {
				return nil, types.ErrKeyExpired.Wrapf("recipient key %s has expired", keyID)
			}
		}
	}

	return missingKeys, nil
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
}

func (k Keeper) isKeyUsableAt(blockTime int64, record types.RecipientKeyRecord) bool {
	if record.RevokedAt != 0 {
		return false
	}
	if record.DeprecatedAt != 0 {
		return false
	}
	if record.ExpiresAt != 0 && blockTime >= record.ExpiresAt {
		return false
	}
	return true
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
			DeprecatedAt:   record.DeprecatedAt,
			ExpiresAt:      record.ExpiresAt,
			PurgeAt:        record.PurgeAt,
			Label:          record.Label,
			KeyVersion:     record.KeyVersion,
		}

		if stop := fn(keyRecord); stop {
			break
		}
	}
}

func (k Keeper) setLatestActiveKey(ctx sdk.Context, address sdk.AccAddress) {
	keys := k.GetRecipientKeys(ctx, address)
	var latest types.RecipientKeyRecord
	found := false
	for _, key := range keys {
		if !k.isKeyUsableAt(ctx.BlockTime().Unix(), key) {
			continue
		}
		if !found || key.KeyVersion > latest.KeyVersion {
			latest = key
			found = true
		}
	}
	if !found {
		return
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ActiveKeyKey(address.Bytes()), []byte(latest.KeyFingerprint))
}

// RefreshActiveRecipientKey recalculates and stores the latest active key for an address.
func (k Keeper) RefreshActiveRecipientKey(ctx sdk.Context, address sdk.AccAddress) {
	k.setLatestActiveKey(ctx, address)
}
