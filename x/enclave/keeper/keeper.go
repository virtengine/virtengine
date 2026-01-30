package keeper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// IKeeper defines the interface for the enclave keeper
type IKeeper interface {
	// Enclave identity management
	RegisterEnclaveIdentity(ctx sdk.Context, identity *types.EnclaveIdentity) error
	GetEnclaveIdentity(ctx sdk.Context, validatorAddr sdk.AccAddress) (*types.EnclaveIdentity, bool)
	UpdateEnclaveIdentity(ctx sdk.Context, identity *types.EnclaveIdentity) error
	RevokeEnclaveIdentity(ctx sdk.Context, validatorAddr sdk.AccAddress, reason string) error

	// Key rotation
	InitiateKeyRotation(ctx sdk.Context, rotation *types.KeyRotationRecord) error
	CompleteKeyRotation(ctx sdk.Context, validatorAddr sdk.AccAddress) error
	GetActiveKeyRotation(ctx sdk.Context, validatorAddr sdk.AccAddress) (*types.KeyRotationRecord, bool)

	// Measurement allowlist
	AddMeasurement(ctx sdk.Context, measurement *types.MeasurementRecord) error
	GetMeasurement(ctx sdk.Context, measurementHash []byte) (*types.MeasurementRecord, bool)
	RevokeMeasurement(ctx sdk.Context, measurementHash []byte, reason string, proposalID uint64) error
	IsMeasurementAllowed(ctx sdk.Context, measurementHash []byte, currentHeight int64) bool

	// Query helpers
	GetActiveValidatorEnclaveKeys(ctx sdk.Context) []types.EnclaveIdentity
	GetCommitteeEnclaveKeys(ctx sdk.Context, epoch uint64) []types.EnclaveIdentity
	GetMeasurementAllowlist(ctx sdk.Context, teeType string, includeRevoked bool) []types.MeasurementRecord
	GetValidKeySet(ctx sdk.Context, forHeight int64) []types.ValidatorKeyInfo

	// Attested results
	SetAttestedResult(ctx sdk.Context, result *types.AttestedScoringResult) error
	GetAttestedResult(ctx sdk.Context, blockHeight int64, scopeID string) (*types.AttestedScoringResult, bool)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Iterators
	WithEnclaveIdentities(ctx sdk.Context, fn func(identity types.EnclaveIdentity) bool)
	WithMeasurements(ctx sdk.Context, fn func(measurement types.MeasurementRecord) bool)

	// Validation
	ValidateAttestation(ctx sdk.Context, identity *types.EnclaveIdentity) error
	VerifyEnclaveSignature(ctx sdk.Context, result *types.AttestedScoringResult) error

	// Codec and store
	Codec() codec.Codec
	StoreKey() storetypes.StoreKey
}

// Keeper of the enclave store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.Codec

	// The address capable of executing governance messages
	authority string
}

// NewKeeper creates and returns an instance for enclave keeper
func NewKeeper(cdc codec.Codec, skey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		skey:      skey,
		authority: authority,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.Codec {
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
	if err := types.ValidateParams(&params); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&params)
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

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// ============================================================================
// Enclave Identity Management
// ============================================================================

// RegisterEnclaveIdentity registers a new enclave identity for a validator
func (k Keeper) RegisterEnclaveIdentity(ctx sdk.Context, identity *types.EnclaveIdentity) error {
	if err := types.ValidateEnclaveIdentity(identity); err != nil {
		return err
	}

	validatorAddr, err := sdk.AccAddressFromBech32(identity.ValidatorAddress)
	if err != nil {
		return types.ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	// Check if identity already exists
	if _, exists := k.GetEnclaveIdentity(ctx, validatorAddr); exists {
		return types.ErrEnclaveIdentityExists
	}

	// Check rate limits
	if err := k.CheckRegistrationRateLimit(ctx, validatorAddr); err != nil {
		return err
	}

	// Validate TEE type is allowed
	params := k.GetParams(ctx)
	if !types.IsTEETypeAllowed(&params, identity.TeeType) {
		return types.ErrInvalidEnclaveIdentity.Wrapf("TEE type %s is not allowed", identity.TeeType.String())
	}

	// Validate measurement is allowlisted
	if !k.IsMeasurementAllowed(ctx, identity.MeasurementHash, ctx.BlockHeight()) {
		return types.ErrMeasurementNotAllowlisted
	}

	// Validate attestation
	if err := k.ValidateAttestation(ctx, identity); err != nil {
		return err
	}

	// Set defaults if not provided
	if identity.RegisteredAt.IsZero() {
		identity.RegisteredAt = ctx.BlockTime()
	}
	identity.UpdatedAt = ctx.BlockTime()

	if identity.ExpiryHeight == 0 {
		identity.ExpiryHeight = ctx.BlockHeight() + params.DefaultExpiryBlocks
	}

	identity.Status = types.EnclaveIdentityStatusActive

	// Store the identity
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(identity)
	if err != nil {
		return err
	}
	store.Set(types.EnclaveIdentityKey(validatorAddr), bz)

	// Store key fingerprint index
	fingerprint := types.KeyFingerprint(identity.EncryptionPubKey)
	store.Set(types.EnclaveKeyByFingerprintKey([]byte(fingerprint)), validatorAddr)

	// Update rate limit tracking
	k.IncrementBlockRegistrationCount(ctx)
	k.RecordValidatorRegistration(ctx, validatorAddr)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnclaveIdentityRegistered,
			sdk.NewAttribute(types.AttributeKeyValidator, identity.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyTEEType, identity.TeeType.String()),
			sdk.NewAttribute(types.AttributeKeyMeasurementHash, hex.EncodeToString(identity.MeasurementHash)),
			sdk.NewAttribute(types.AttributeKeyEncryptionKeyID, fingerprint),
			sdk.NewAttribute(types.AttributeKeyExpiryHeight, math.NewInt(identity.ExpiryHeight).String()),
		),
	)

	return nil
}

// GetEnclaveIdentity retrieves an enclave identity for a validator
func (k Keeper) GetEnclaveIdentity(ctx sdk.Context, validatorAddr sdk.AccAddress) (*types.EnclaveIdentity, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.EnclaveIdentityKey(validatorAddr))
	if bz == nil {
		return nil, false
	}

	var identity types.EnclaveIdentity
	if err := json.Unmarshal(bz, &identity); err != nil {
		return nil, false
	}
	return &identity, true
}

// UpdateEnclaveIdentity updates an existing enclave identity
func (k Keeper) UpdateEnclaveIdentity(ctx sdk.Context, identity *types.EnclaveIdentity) error {
	if err := types.ValidateEnclaveIdentity(identity); err != nil {
		return err
	}

	validatorAddr, err := sdk.AccAddressFromBech32(identity.ValidatorAddress)
	if err != nil {
		return types.ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	// Check if identity exists
	existing, exists := k.GetEnclaveIdentity(ctx, validatorAddr)
	if !exists {
		return types.ErrEnclaveIdentityNotFound
	}

	// Remove old fingerprint index if key changed
	oldFingerprint := types.KeyFingerprint(existing.EncryptionPubKey)
	newFingerprint := types.KeyFingerprint(identity.EncryptionPubKey)
	if oldFingerprint != newFingerprint {
		store := ctx.KVStore(k.skey)
		store.Delete(types.EnclaveKeyByFingerprintKey([]byte(oldFingerprint)))
		store.Set(types.EnclaveKeyByFingerprintKey([]byte(newFingerprint)), validatorAddr)
	}

	identity.UpdatedAt = ctx.BlockTime()

	// Store updated identity
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(identity)
	if err != nil {
		return err
	}
	store.Set(types.EnclaveIdentityKey(validatorAddr), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnclaveIdentityUpdated,
			sdk.NewAttribute(types.AttributeKeyValidator, identity.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyMeasurementHash, hex.EncodeToString(identity.MeasurementHash)),
		),
	)

	return nil
}

// RevokeEnclaveIdentity revokes an enclave identity
func (k Keeper) RevokeEnclaveIdentity(ctx sdk.Context, validatorAddr sdk.AccAddress, reason string) error {
	identity, exists := k.GetEnclaveIdentity(ctx, validatorAddr)
	if !exists {
		return types.ErrEnclaveIdentityNotFound
	}

	identity.Status = types.EnclaveIdentityStatusRevoked
	identity.UpdatedAt = ctx.BlockTime()

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(identity)
	if err != nil {
		return err
	}
	store.Set(types.EnclaveIdentityKey(validatorAddr), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnclaveIdentityRevoked,
			sdk.NewAttribute(types.AttributeKeyValidator, identity.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	return nil
}

// ============================================================================
// Key Rotation
// ============================================================================

// InitiateKeyRotation starts a key rotation for a validator
func (k Keeper) InitiateKeyRotation(ctx sdk.Context, rotation *types.KeyRotationRecord) error {
	if err := types.ValidateKeyRotationRecord(rotation); err != nil {
		return err
	}

	validatorAddr, err := sdk.AccAddressFromBech32(rotation.ValidatorAddress)
	if err != nil {
		return types.ErrInvalidEnclaveIdentity.Wrapf("invalid validator address: %v", err)
	}

	// Check if there's already an active rotation
	if _, exists := k.GetActiveKeyRotation(ctx, validatorAddr); exists {
		return types.ErrKeyRotationInProgress
	}

	rotation.Status = types.KeyRotationStatusActive
	rotation.InitiatedAt = ctx.BlockTime()

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(rotation)
	if err != nil {
		return err
	}
	store.Set(types.KeyRotationKey(validatorAddr, rotation.Epoch), bz)

	// Update identity status
	identity, exists := k.GetEnclaveIdentity(ctx, validatorAddr)
	if exists {
		identity.Status = types.EnclaveIdentityStatusRotating
		k.UpdateEnclaveIdentity(ctx, identity)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnclaveKeyRotated,
			sdk.NewAttribute(types.AttributeKeyValidator, rotation.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyOldKeyFingerprint, rotation.OldKeyFingerprint),
			sdk.NewAttribute(types.AttributeKeyNewKeyFingerprint, rotation.NewKeyFingerprint),
			sdk.NewAttribute(types.AttributeKeyOverlapStartHeight, math.NewInt(rotation.OverlapStartHeight).String()),
			sdk.NewAttribute(types.AttributeKeyOverlapEndHeight, math.NewInt(rotation.OverlapEndHeight).String()),
		),
	)

	return nil
}

// CompleteKeyRotation completes a key rotation
func (k Keeper) CompleteKeyRotation(ctx sdk.Context, validatorAddr sdk.AccAddress) error {
	rotation, exists := k.GetActiveKeyRotation(ctx, validatorAddr)
	if !exists {
		return types.ErrNoActiveRotation
	}

	now := ctx.BlockTime()
	rotation.Status = types.KeyRotationStatusCompleted
	rotation.CompletedAt = &now

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(rotation)
	if err != nil {
		return err
	}
	store.Set(types.KeyRotationKey(validatorAddr, rotation.Epoch), bz)

	// Update identity status back to active
	identity, exists := k.GetEnclaveIdentity(ctx, validatorAddr)
	if exists {
		identity.Status = types.EnclaveIdentityStatusActive
		k.UpdateEnclaveIdentity(ctx, identity)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKeyRotationCompleted,
			sdk.NewAttribute(types.AttributeKeyValidator, rotation.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyNewKeyFingerprint, rotation.NewKeyFingerprint),
		),
	)

	// Record metrics
	k.RecordKeyRotationCompleted()

	return nil
}

// GetActiveKeyRotation retrieves the active key rotation for a validator
func (k Keeper) GetActiveKeyRotation(ctx sdk.Context, validatorAddr sdk.AccAddress) (*types.KeyRotationRecord, bool) {
	store := ctx.KVStore(k.skey)
	prefix := append(types.PrefixKeyRotation, validatorAddr...)
	iterator := store.ReverseIterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var rotation types.KeyRotationRecord
		if err := json.Unmarshal(iterator.Value(), &rotation); err != nil {
			continue
		}

		if rotation.Status == types.KeyRotationStatusActive {
			return &rotation, true
		}
	}

	return nil, false
}

// ============================================================================
// Measurement Allowlist
// ============================================================================

// AddMeasurement adds a measurement to the allowlist
func (k Keeper) AddMeasurement(ctx sdk.Context, measurement *types.MeasurementRecord) error {
	if err := types.ValidateMeasurementRecord(measurement); err != nil {
		return err
	}

	measurement.AddedAt = ctx.BlockTime()

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(measurement)
	if err != nil {
		return err
	}
	store.Set(types.MeasurementAllowlistKey(measurement.MeasurementHash), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMeasurementAdded,
			sdk.NewAttribute(types.AttributeKeyMeasurementHash, hex.EncodeToString(measurement.MeasurementHash)),
			sdk.NewAttribute(types.AttributeKeyTEEType, measurement.TeeType.String()),
			sdk.NewAttribute(types.AttributeKeyDescription, measurement.Description),
		),
	)

	return nil
}

// GetMeasurement retrieves a measurement from the allowlist
func (k Keeper) GetMeasurement(ctx sdk.Context, measurementHash []byte) (*types.MeasurementRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.MeasurementAllowlistKey(measurementHash))
	if bz == nil {
		return nil, false
	}

	var measurement types.MeasurementRecord
	if err := json.Unmarshal(bz, &measurement); err != nil {
		return nil, false
	}
	return &measurement, true
}

// RevokeMeasurement revokes a measurement from the allowlist
func (k Keeper) RevokeMeasurement(ctx sdk.Context, measurementHash []byte, reason string, proposalID uint64) error {
	measurement, exists := k.GetMeasurement(ctx, measurementHash)
	if !exists {
		return types.ErrMeasurementNotAllowlisted
	}

	now := ctx.BlockTime()
	measurement.Revoked = true
	measurement.RevokedAt = &now
	measurement.RevokedByProposal = proposalID

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(measurement)
	if err != nil {
		return err
	}
	store.Set(types.MeasurementAllowlistKey(measurementHash), bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMeasurementRevoked,
			sdk.NewAttribute(types.AttributeKeyMeasurementHash, hex.EncodeToString(measurementHash)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	return nil
}

// IsMeasurementAllowed checks if a measurement is in the allowlist and valid
func (k Keeper) IsMeasurementAllowed(ctx sdk.Context, measurementHash []byte, currentHeight int64) bool {
	measurement, exists := k.GetMeasurement(ctx, measurementHash)
	if !exists {
		return false
	}
	return types.IsMeasurementValid(measurement, currentHeight)
}

// ============================================================================
// Query Helpers
// ============================================================================

// GetActiveValidatorEnclaveKeys returns all active validator enclave keys
func (k Keeper) GetActiveValidatorEnclaveKeys(ctx sdk.Context) []types.EnclaveIdentity {
	var identities []types.EnclaveIdentity
	currentHeight := ctx.BlockHeight()

	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		if identity.Status == types.EnclaveIdentityStatusActive &&
			!types.IsIdentityExpired(&identity, currentHeight) {
			identities = append(identities, identity)
		}
		return false
	})

	return identities
}

// GetCommitteeEnclaveKeys is now implemented in committee.go with deterministic selection

// GetMeasurementAllowlist returns all measurements in the allowlist
func (k Keeper) GetMeasurementAllowlist(ctx sdk.Context, teeType string, includeRevoked bool) []types.MeasurementRecord {
	var measurements []types.MeasurementRecord
	currentHeight := ctx.BlockHeight()

	k.WithMeasurements(ctx, func(measurement types.MeasurementRecord) bool {
		// Filter by TEE type if specified
		if teeType != "" && measurement.TeeType.String() != teeType {
			return false
		}

		// Filter revoked unless requested
		if !includeRevoked && measurement.Revoked {
			return false
		}

		// Filter expired
		if measurement.ExpiryHeight > 0 && currentHeight >= measurement.ExpiryHeight {
			return false
		}

		measurements = append(measurements, measurement)
		return false
	})

	return measurements
}

// GetValidKeySet returns the valid key set for a given block height
func (k Keeper) GetValidKeySet(ctx sdk.Context, forHeight int64) []types.ValidatorKeyInfo {
	var keys []types.ValidatorKeyInfo

	k.WithEnclaveIdentities(ctx, func(identity types.EnclaveIdentity) bool {
		if identity.Status != types.EnclaveIdentityStatusActive &&
			identity.Status != types.EnclaveIdentityStatusRotating {
			return false
		}

		if types.IsIdentityExpired(&identity, forHeight) {
			return false
		}

		// Check if in rotation
		validatorAddr, _ := sdk.AccAddressFromBech32(identity.ValidatorAddress)
		rotation, hasRotation := k.GetActiveKeyRotation(ctx, validatorAddr)
		isInRotation := hasRotation && types.IsInOverlapPeriod(rotation, forHeight)

		keys = append(keys, types.ValidatorKeyInfo{
			ValidatorAddress: identity.ValidatorAddress,
			EncryptionKeyId:  types.KeyFingerprint(identity.EncryptionPubKey),
			EncryptionPubKey: identity.EncryptionPubKey,
			MeasurementHash:  identity.MeasurementHash,
			ExpiryHeight:     identity.ExpiryHeight,
			IsInRotation:     isInRotation,
		})

		return false
	})

	return keys
}

// ============================================================================
// Attested Results
// ============================================================================

// SetAttestedResult stores an attested scoring result
func (k Keeper) SetAttestedResult(ctx sdk.Context, result *types.AttestedScoringResult) error {
	if err := types.ValidateAttestedScoringResult(result); err != nil {
		return err
	}

	// Verify the enclave measurement is allowlisted
	if !k.IsMeasurementAllowed(ctx, result.EnclaveMeasurementHash, ctx.BlockHeight()) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeVEIDScoreRejectedAttestation,
				sdk.NewAttribute(types.AttributeKeyScopeID, result.ScopeId),
				sdk.NewAttribute(types.AttributeKeyMeasurementHash, types.MeasurementHashHex(result.EnclaveMeasurementHash)),
				sdk.NewAttribute(types.AttributeKeyReason, "measurement not allowlisted"),
			),
		)
		return types.ErrMeasurementNotAllowlisted
	}

	// Verify the enclave signature
	if err := k.VerifyEnclaveSignature(ctx, result); err != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeVEIDScoreRejectedAttestation,
				sdk.NewAttribute(types.AttributeKeyScopeID, result.ScopeId),
				sdk.NewAttribute(types.AttributeKeyReason, "invalid enclave signature"),
			),
		)
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(result)
	if err != nil {
		return err
	}
	store.Set(types.AttestedResultKey(result.BlockHeight, result.ScopeId), bz)

	// Emit success event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVEIDScoreComputedAttested,
			sdk.NewAttribute(types.AttributeKeyScopeID, result.ScopeId),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, result.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyScore, math.NewInt(int64(result.Score)).String()),
			sdk.NewAttribute(types.AttributeKeyStatus, result.Status),
			sdk.NewAttribute(types.AttributeKeyMeasurementHash, types.MeasurementHashHex(result.EnclaveMeasurementHash)),
			sdk.NewAttribute(types.AttributeKeyValidator, result.ValidatorAddress),
		),
	)

	return nil
}

// GetAttestedResult retrieves an attested scoring result
func (k Keeper) GetAttestedResult(ctx sdk.Context, blockHeight int64, scopeID string) (*types.AttestedScoringResult, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AttestedResultKey(blockHeight, scopeID))
	if bz == nil {
		return nil, false
	}

	var result types.AttestedScoringResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return nil, false
	}
	return &result, true
}

// ============================================================================
// Validation
// ============================================================================

// ValidateAttestation validates an enclave's attestation
func (k Keeper) ValidateAttestation(ctx sdk.Context, identity *types.EnclaveIdentity) error {
	params := k.GetParams(ctx)

	// Validate quote version
	if identity.QuoteVersion < params.MinQuoteVersion {
		return types.ErrInvalidQuoteVersion.Wrapf(
			"quote version %d is below minimum %d",
			identity.QuoteVersion, params.MinQuoteVersion,
		)
	}

	// Validate debug mode
	if identity.DebugMode {
		return types.ErrDebugModeEnabled
	}

	// Validate attestation chain if required
	if params.RequireAttestationChain && len(identity.AttestationChain) == 0 {
		return types.ErrAttestationInvalid.Wrap("attestation chain required but not provided")
	}

	// Validate measurement is allowlisted with proper ISVSVN
	measurement, exists := k.GetMeasurement(ctx, identity.MeasurementHash)
	if !exists {
		return types.ErrMeasurementNotAllowlisted
	}

	if identity.IsvSvn < measurement.MinIsvSvn {
		return types.ErrISVSVNTooLow.Wrapf(
			"ISVSVN %d is below minimum %d for measurement",
			identity.IsvSvn, measurement.MinIsvSvn,
		)
	}

	// In production, this would verify the attestation quote cryptographically
	// against the platform root of trust
	return nil
}

// VerifyEnclaveSignature verifies the enclave signature on an attested result
func (k Keeper) VerifyEnclaveSignature(ctx sdk.Context, result *types.AttestedScoringResult) error {
	// Get the validator's enclave identity
	validatorAddr, err := sdk.AccAddressFromBech32(result.ValidatorAddress)
	if err != nil {
		return types.ErrEnclaveSignatureInvalid.Wrapf("invalid validator address: %v", err)
	}

	identity, exists := k.GetEnclaveIdentity(ctx, validatorAddr)
	if !exists {
		return types.ErrEnclaveIdentityNotFound
	}

	// Verify the measurement matches
	if !bytes.Equal(identity.MeasurementHash, result.EnclaveMeasurementHash) {
		return types.ErrEnclaveSignatureInvalid.Wrap("measurement hash mismatch")
	}

	// In production, this would verify the signature cryptographically
	// using the enclave's signing public key
	if len(result.EnclaveSignature) == 0 {
		return types.ErrEnclaveSignatureInvalid.Wrap("empty signature")
	}

	return nil
}

// ============================================================================
// Iterators
// ============================================================================

// WithEnclaveIdentities iterates over all enclave identities
func (k Keeper) WithEnclaveIdentities(ctx sdk.Context, fn func(identity types.EnclaveIdentity) bool) {
	store := ctx.KVStore(k.skey)
	iterator := store.Iterator(types.PrefixEnclaveIdentity, storetypes.PrefixEndBytes(types.PrefixEnclaveIdentity))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var identity types.EnclaveIdentity
		if err := json.Unmarshal(iterator.Value(), &identity); err != nil {
			continue
		}

		if fn(identity) {
			break
		}
	}
}

// WithMeasurements iterates over all measurements
func (k Keeper) WithMeasurements(ctx sdk.Context, fn func(measurement types.MeasurementRecord) bool) {
	store := ctx.KVStore(k.skey)
	iterator := store.Iterator(types.PrefixMeasurementAllowlist, storetypes.PrefixEndBytes(types.PrefixMeasurementAllowlist))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var measurement types.MeasurementRecord
		if err := json.Unmarshal(iterator.Value(), &measurement); err != nil {
			continue
		}

		if fn(measurement) {
			break
		}
	}
}
