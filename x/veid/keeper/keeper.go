package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// IKeeper defines the interface for the veid keeper
type IKeeper interface {
	// Identity record management
	GetIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (types.IdentityRecord, bool)
	SetIdentityRecord(ctx sdk.Context, record types.IdentityRecord) error
	CreateIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (*types.IdentityRecord, error)

	// Scope management
	UploadScope(ctx sdk.Context, address sdk.AccAddress, scope *types.IdentityScope) error
	GetScope(ctx sdk.Context, address sdk.AccAddress, scopeID string) (types.IdentityScope, bool)
	GetScopesByType(ctx sdk.Context, address sdk.AccAddress, scopeType types.ScopeType) []types.IdentityScope
	RevokeScope(ctx sdk.Context, address sdk.AccAddress, scopeID string, reason string) error

	// Identity Wallet management (VE-209)
	CreateIdentityWallet(ctx sdk.Context, accountAddr sdk.AccAddress, bindingSignature, bindingPubKey []byte) (*types.IdentityWallet, error)
	GetWallet(ctx sdk.Context, address sdk.AccAddress) (*types.IdentityWallet, bool)
	GetWalletByID(ctx sdk.Context, walletID string) (*types.IdentityWallet, bool)
	SetWallet(ctx sdk.Context, wallet *types.IdentityWallet) error
	AddScopeToWallet(ctx sdk.Context, accountAddr sdk.AccAddress, scopeRef types.ScopeReference, userSignature []byte) error
	RevokeScopeFromWallet(ctx sdk.Context, accountAddr sdk.AccAddress, scopeID string, reason string, userSignature []byte) error
	UpdateConsent(ctx sdk.Context, accountAddr sdk.AccAddress, update types.ConsentUpdateRequest, userSignature []byte) error
	UpdateDerivedFeatures(ctx sdk.Context, accountAddr sdk.AccAddress, update *types.DerivedFeaturesUpdate) error
	UpdateWalletScore(ctx sdk.Context, accountAddr sdk.AccAddress, newScore uint32, newStatus types.AccountStatus, modelVersion string, validatorAddress string, scopesEvaluated []string, reason string) error
	GetWalletPublicMetadata(ctx sdk.Context, accountAddr sdk.AccAddress) (types.PublicWalletInfo, bool)

	// Verification management
	UpdateVerificationStatus(ctx sdk.Context, address sdk.AccAddress, scopeID string, status types.VerificationStatus, reason string, validatorAddr string) error
	UpdateScore(ctx sdk.Context, address sdk.AccAddress, score uint32, scoreVersion string) error

	// Score retrieval
	GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool)

	// Signature validation
	ValidateClientSignature(ctx sdk.Context, clientID string, signature []byte, payload []byte) error
	ValidateUserSignature(ctx sdk.Context, address sdk.AccAddress, signature []byte, payload []byte) error
	ValidateSaltBinding(ctx sdk.Context, salt []byte, metadata *types.UploadMetadata, payloadHash []byte) error

	// Approved clients
	GetApprovedClient(ctx sdk.Context, clientID string) (types.ApprovedClient, bool)
	IsClientApproved(ctx sdk.Context, clientID string) bool

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Iterators
	WithIdentityRecords(ctx sdk.Context, fn func(record types.IdentityRecord) bool)
	WithScopes(ctx sdk.Context, address sdk.AccAddress, fn func(scope types.IdentityScope) bool)
	WithWallets(ctx sdk.Context, fn func(wallet *types.IdentityWallet) bool)

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// StakingKeeper defines the interface for the cosmos staking keeper that veid needs
// for validator authorization checks on verification updates
type StakingKeeper interface {
	// GetValidator returns the validator with the given operator address
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
}

// Keeper of the veid store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string

	// mfaKeeper is the MFA keeper reference for borderline fallback operations
	// This is set via SetMFAKeeper after module initialization to avoid circular imports
	mfaKeeper MFAKeeper

	// stakingKeeper is the staking keeper reference for validator authorization
	// This is set via SetStakingKeeper after module initialization
	stakingKeeper StakingKeeper

	// zkSystem is the ZK proof system for privacy-preserving proofs
	// Initialized during keeper setup with compiled circuits
	zkSystem *ZKProofSystem

	// randSource provides deterministic randomness derived from tx context.
	// It must never be nil during state transitions to preserve consensus safety.
	randSource RandomSource
}

// NewKeeper creates and returns an instance for veid keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, authority string) Keeper {
	// Initialize ZK proof system
	// Note: Circuit compilation and trusted setup happens here
	// In production, this would load pre-compiled circuits and verification keys
	zkSystem, err := NewZKProofSystem()
	if err != nil {
		// Log warning but don't fail - fall back to hash-based proofs
		// In production, this should be a hard requirement
		log.NewNopLogger().Error("Failed to initialize ZK proof system", "error", err)
	}

	return Keeper{
		cdc:        cdc,
		skey:       skey,
		authority:  authority,
		zkSystem:   zkSystem,
		randSource: DeterministicRandomSource{},
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

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetStakingKeeper sets the staking keeper reference for validator authorization
func (k *Keeper) SetStakingKeeper(stakingKeeper StakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// SetRandomSource overrides the default deterministic random source.
// Passing nil resets to the default DeterministicRandomSource implementation.
func (k *Keeper) SetRandomSource(src RandomSource) {
	if src == nil {
		k.randSource = DeterministicRandomSource{}
		return
	}
	k.randSource = src
}

// IsValidator checks if the given account address is a bonded validator
// This is used to authorize validator-only operations like UpdateVerificationStatus and UpdateScore
func (k Keeper) IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.stakingKeeper == nil {
		// If staking keeper is not set, deny authorization for safety
		k.Logger(ctx).Error("staking keeper not set, denying validator authorization")
		return false
	}

	// Convert account address to validator address
	valAddr := sdk.ValAddress(addr)

	// Get the validator
	validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil {
		// Validator not found
		return false
	}

	// Check if validator is bonded (active)
	return validator.IsBonded()
}

// ============================================================================
// Parameters
// ============================================================================

// paramsStore is the stored format of params
type paramsStore struct {
	MaxScopesPerAccount    uint32            `json:"max_scopes_per_account"`
	MaxScopesPerType       uint32            `json:"max_scopes_per_type"`
	SaltMinBytes           uint32            `json:"salt_min_bytes"`
	SaltMaxBytes           uint32            `json:"salt_max_bytes"`
	RequireClientSignature bool              `json:"require_client_signature"`
	RequireUserSignature   bool              `json:"require_user_signature"`
	VerificationExpiryDays uint32            `json:"verification_expiry_days"`
	MinScoreForTier        map[string]uint32 `json:"min_score_for_tier"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		MaxScopesPerAccount:    params.MaxScopesPerAccount,
		MaxScopesPerType:       params.MaxScopesPerType,
		SaltMinBytes:           params.SaltMinBytes,
		SaltMaxBytes:           params.SaltMaxBytes,
		RequireClientSignature: params.RequireClientSignature,
		RequireUserSignature:   params.RequireUserSignature,
		VerificationExpiryDays: params.VerificationExpiryDays,
		MinScoreForTier:        params.MinScoreForTier,
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
		MaxScopesPerAccount:    ps.MaxScopesPerAccount,
		MaxScopesPerType:       ps.MaxScopesPerType,
		SaltMinBytes:           ps.SaltMinBytes,
		SaltMaxBytes:           ps.SaltMaxBytes,
		RequireClientSignature: ps.RequireClientSignature,
		RequireUserSignature:   ps.RequireUserSignature,
		VerificationExpiryDays: ps.VerificationExpiryDays,
		MinScoreForTier:        ps.MinScoreForTier,
	}
}

// ============================================================================
// Identity Records
// ============================================================================

// identityRecordStore is the stored format of an identity record
type identityRecordStore struct {
	AccountAddress string             `json:"account_address"`
	ScopeRefs      []types.ScopeRef   `json:"scope_refs"`
	CurrentScore   uint32             `json:"current_score"`
	ScoreVersion   string             `json:"score_version"`
	LastVerifiedAt *int64             `json:"last_verified_at,omitempty"`
	CreatedAt      int64              `json:"created_at"`
	UpdatedAt      int64              `json:"updated_at"`
	Tier           types.IdentityTier `json:"tier"`
	Flags          []string           `json:"flags,omitempty"`
	Locked         bool               `json:"locked"`
	LockedReason   string             `json:"locked_reason,omitempty"`
}

// GetIdentityRecord returns an identity record by address
func (k Keeper) GetIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (types.IdentityRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := types.IdentityRecordKey(address.Bytes())
	bz := store.Get(key)
	if bz == nil {
		return types.IdentityRecord{}, false
	}

	var rs identityRecordStore
	if err := json.Unmarshal(bz, &rs); err != nil {
		return types.IdentityRecord{}, false
	}

	record := types.IdentityRecord{
		AccountAddress: rs.AccountAddress,
		ScopeRefs:      rs.ScopeRefs,
		CurrentScore:   rs.CurrentScore,
		ScoreVersion:   rs.ScoreVersion,
		CreatedAt:      time.Unix(rs.CreatedAt, 0),
		UpdatedAt:      time.Unix(rs.UpdatedAt, 0),
		Tier:           rs.Tier,
		Flags:          rs.Flags,
		Locked:         rs.Locked,
		LockedReason:   rs.LockedReason,
	}

	if rs.LastVerifiedAt != nil {
		t := time.Unix(*rs.LastVerifiedAt, 0)
		record.LastVerifiedAt = &t
	}

	return record, true
}

// SetIdentityRecord stores an identity record
func (k Keeper) SetIdentityRecord(ctx sdk.Context, record types.IdentityRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	address, err := sdk.AccAddressFromBech32(record.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	rs := identityRecordStore{
		AccountAddress: record.AccountAddress,
		ScopeRefs:      record.ScopeRefs,
		CurrentScore:   record.CurrentScore,
		ScoreVersion:   record.ScoreVersion,
		CreatedAt:      record.CreatedAt.Unix(),
		UpdatedAt:      record.UpdatedAt.Unix(),
		Tier:           record.Tier,
		Flags:          record.Flags,
		Locked:         record.Locked,
		LockedReason:   record.LockedReason,
	}

	if record.LastVerifiedAt != nil {
		ts := record.LastVerifiedAt.Unix()
		rs.LastVerifiedAt = &ts
	}

	bz, err := json.Marshal(&rs)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.IdentityRecordKey(address.Bytes()), bz)
	return nil
}

// CreateIdentityRecord creates a new identity record for an address
func (k Keeper) CreateIdentityRecord(ctx sdk.Context, address sdk.AccAddress) (*types.IdentityRecord, error) {
	// Check if record already exists
	if _, found := k.GetIdentityRecord(ctx, address); found {
		return nil, types.ErrInvalidIdentityRecord.Wrap("identity record already exists")
	}

	record := types.NewIdentityRecord(address.String(), ctx.BlockTime())
	if err := k.SetIdentityRecord(ctx, *record); err != nil {
		return nil, err
	}

	return record, nil
}

// WithIdentityRecords iterates over all identity records
func (k Keeper) WithIdentityRecords(ctx sdk.Context, fn func(record types.IdentityRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixIdentityRecord)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rs identityRecordStore
		if err := json.Unmarshal(iter.Value(), &rs); err != nil {
			continue
		}

		record := types.IdentityRecord{
			AccountAddress: rs.AccountAddress,
			ScopeRefs:      rs.ScopeRefs,
			CurrentScore:   rs.CurrentScore,
			ScoreVersion:   rs.ScoreVersion,
			CreatedAt:      time.Unix(rs.CreatedAt, 0),
			UpdatedAt:      time.Unix(rs.UpdatedAt, 0),
			Tier:           rs.Tier,
			Flags:          rs.Flags,
			Locked:         rs.Locked,
			LockedReason:   rs.LockedReason,
		}

		if rs.LastVerifiedAt != nil {
			t := time.Unix(*rs.LastVerifiedAt, 0)
			record.LastVerifiedAt = &t
		}

		if fn(record) {
			break
		}
	}
}

// ============================================================================
// Scopes
// ============================================================================

// scopeStore is the stored format of an identity scope
type scopeStore struct {
	ScopeID          string                   `json:"scope_id"`
	ScopeType        types.ScopeType          `json:"scope_type"`
	Version          uint32                   `json:"version"`
	EncryptedPayload json.RawMessage          `json:"encrypted_payload"`
	UploadMetadata   types.UploadMetadata     `json:"upload_metadata"`
	Status           types.VerificationStatus `json:"status"`
	UploadedAt       int64                    `json:"uploaded_at"`
	VerifiedAt       *int64                   `json:"verified_at,omitempty"`
	ExpiresAt        *int64                   `json:"expires_at,omitempty"`
	Revoked          bool                     `json:"revoked"`
	RevokedAt        *int64                   `json:"revoked_at,omitempty"`
	RevokedReason    string                   `json:"revoked_reason,omitempty"`
}

// UploadScope uploads a new identity scope
func (k Keeper) UploadScope(ctx sdk.Context, address sdk.AccAddress, scope *types.IdentityScope) error {
	if err := scope.Validate(); err != nil {
		return err
	}

	params := k.GetParams(ctx)

	// Validate salt binding, client signature, and user signature
	if err := k.ValidateSaltBinding(ctx, scope.UploadMetadata.Salt, &scope.UploadMetadata, scope.UploadMetadata.PayloadHash); err != nil {
		return err
	}
	if err := k.ValidateClientSignature(ctx, scope.UploadMetadata.ClientID, scope.UploadMetadata.ClientSignature, scope.UploadMetadata.SigningPayload()); err != nil {
		return err
	}
	if err := k.ValidateUserSignature(ctx, address, scope.UploadMetadata.UserSignature, scope.UploadMetadata.UserSigningPayload()); err != nil {
		return err
	}

	// Get or create identity record
	record, found := k.GetIdentityRecord(ctx, address)
	if !found {
		newRecord, err := k.CreateIdentityRecord(ctx, address)
		if err != nil {
			return err
		}
		record = *newRecord
	}

	// Check if identity is locked
	if record.Locked {
		return types.ErrIdentityLocked.Wrap(record.LockedReason)
	}

	// Check scope limits
	if safeUint32FromIntBiometric(len(record.ScopeRefs)) >= params.MaxScopesPerAccount {
		return types.ErrMaxScopesExceeded.Wrapf("maximum %d scopes per account", params.MaxScopesPerAccount)
	}

	typeCount := 0
	for _, ref := range record.ScopeRefs {
		if ref.ScopeType == scope.ScopeType {
			typeCount++
		}
	}
	if safeUint32FromIntBiometric(typeCount) >= params.MaxScopesPerType {
		return types.ErrMaxScopesExceeded.Wrapf("maximum %d scopes of type %s", params.MaxScopesPerType, scope.ScopeType)
	}

	// Check if scope already exists
	if _, found := k.GetScope(ctx, address, scope.ScopeID); found {
		return types.ErrScopeAlreadyExists.Wrapf("scope %s already exists", scope.ScopeID)
	}

	// Store the scope
	payloadBz, err := json.Marshal(scope.EncryptedPayload)
	if err != nil {
		return err
	}

	ss := scopeStore{
		ScopeID:          scope.ScopeID,
		ScopeType:        scope.ScopeType,
		Version:          scope.Version,
		EncryptedPayload: payloadBz,
		UploadMetadata:   scope.UploadMetadata,
		Status:           scope.Status,
		UploadedAt:       scope.UploadedAt.Unix(),
		Revoked:          scope.Revoked,
		RevokedReason:    scope.RevokedReason,
	}

	if scope.VerifiedAt != nil {
		ts := scope.VerifiedAt.Unix()
		ss.VerifiedAt = &ts
	}
	if scope.ExpiresAt != nil {
		ts := scope.ExpiresAt.Unix()
		ss.ExpiresAt = &ts
	}
	if scope.RevokedAt != nil {
		ts := scope.RevokedAt.Unix()
		ss.RevokedAt = &ts
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ScopeKey(address.Bytes(), scope.ScopeID), bz)

	// Mark salt as used
	k.markSaltUsed(ctx, scope.UploadMetadata.Salt)

	// Update identity record with scope reference
	record.AddScopeRef(types.NewScopeRef(scope))
	record.UpdatedAt = ctx.BlockTime()
	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return err
	}

	return nil
}

// GetScope returns a scope by address and scope ID
func (k Keeper) GetScope(ctx sdk.Context, address sdk.AccAddress, scopeID string) (types.IdentityScope, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ScopeKey(address.Bytes(), scopeID)
	bz := store.Get(key)
	if bz == nil {
		return types.IdentityScope{}, false
	}

	var ss scopeStore
	if err := json.Unmarshal(bz, &ss); err != nil {
		return types.IdentityScope{}, false
	}

	return k.scopeStoreToScope(ss), true
}

// GetScopesByType returns all scopes of a specific type for an address
func (k Keeper) GetScopesByType(ctx sdk.Context, address sdk.AccAddress, scopeType types.ScopeType) []types.IdentityScope {
	var scopes []types.IdentityScope

	k.WithScopes(ctx, address, func(scope types.IdentityScope) bool {
		if scope.ScopeType == scopeType {
			scopes = append(scopes, scope)
		}
		return false
	})

	return scopes
}

// WithScopes iterates over all scopes for an address
func (k Keeper) WithScopes(ctx sdk.Context, address sdk.AccAddress, fn func(scope types.IdentityScope) bool) {
	store := ctx.KVStore(k.skey)
	prefix := types.ScopePrefixKey(address.Bytes())
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var ss scopeStore
		if err := json.Unmarshal(iter.Value(), &ss); err != nil {
			continue
		}

		if fn(k.scopeStoreToScope(ss)) {
			break
		}
	}
}

// RevokeScope revokes an identity scope
func (k Keeper) RevokeScope(ctx sdk.Context, address sdk.AccAddress, scopeID string, reason string) error {
	scope, found := k.GetScope(ctx, address, scopeID)
	if !found {
		return types.ErrScopeNotFound.Wrapf("scope %s not found", scopeID)
	}

	if scope.Revoked {
		return types.ErrScopeRevoked.Wrapf("scope %s already revoked", scopeID)
	}

	// Update scope
	now := ctx.BlockTime()
	scope.Revoked = true
	scope.RevokedAt = &now
	scope.RevokedReason = reason

	// Store updated scope
	if err := k.setScope(ctx, address, &scope); err != nil {
		return err
	}

	// Update identity record
	record, found := k.GetIdentityRecord(ctx, address)
	if found {
		for i, ref := range record.ScopeRefs {
			if ref.ScopeID == scopeID {
				record.ScopeRefs[i].Status = types.VerificationStatusExpired
				break
			}
		}
		record.UpdatedAt = now
		if err := k.SetIdentityRecord(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

// setScope stores a scope (internal helper)
func (k Keeper) setScope(ctx sdk.Context, address sdk.AccAddress, scope *types.IdentityScope) error {
	payloadBz, err := json.Marshal(scope.EncryptedPayload)
	if err != nil {
		return err
	}

	ss := scopeStore{
		ScopeID:          scope.ScopeID,
		ScopeType:        scope.ScopeType,
		Version:          scope.Version,
		EncryptedPayload: payloadBz,
		UploadMetadata:   scope.UploadMetadata,
		Status:           scope.Status,
		UploadedAt:       scope.UploadedAt.Unix(),
		Revoked:          scope.Revoked,
		RevokedReason:    scope.RevokedReason,
	}

	if scope.VerifiedAt != nil {
		ts := scope.VerifiedAt.Unix()
		ss.VerifiedAt = &ts
	}
	if scope.ExpiresAt != nil {
		ts := scope.ExpiresAt.Unix()
		ss.ExpiresAt = &ts
	}
	if scope.RevokedAt != nil {
		ts := scope.RevokedAt.Unix()
		ss.RevokedAt = &ts
	}

	bz, err := json.Marshal(&ss)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ScopeKey(address.Bytes(), scope.ScopeID), bz)
	return nil
}

// scopeStoreToScope converts a scopeStore to IdentityScope
func (k Keeper) scopeStoreToScope(ss scopeStore) types.IdentityScope {
	scope := types.IdentityScope{
		ScopeID:        ss.ScopeID,
		ScopeType:      ss.ScopeType,
		Version:        ss.Version,
		UploadMetadata: ss.UploadMetadata,
		Status:         ss.Status,
		UploadedAt:     time.Unix(ss.UploadedAt, 0),
		Revoked:        ss.Revoked,
		RevokedReason:  ss.RevokedReason,
	}

	// Unmarshal encrypted payload
	if len(ss.EncryptedPayload) > 0 {
		_ = json.Unmarshal(ss.EncryptedPayload, &scope.EncryptedPayload)
	}

	if ss.VerifiedAt != nil {
		t := time.Unix(*ss.VerifiedAt, 0)
		scope.VerifiedAt = &t
	}
	if ss.ExpiresAt != nil {
		t := time.Unix(*ss.ExpiresAt, 0)
		scope.ExpiresAt = &t
	}
	if ss.RevokedAt != nil {
		t := time.Unix(*ss.RevokedAt, 0)
		scope.RevokedAt = &t
	}

	return scope
}

// ============================================================================
// Salt Management
// ============================================================================

// ValidateSaltBindingWithSignature validates salt binding with cryptographic signature verification.
// This ensures the salt is cryptographically bound to the user, scope, and timestamp.
//
// Task Reference: VE-3022 - Cryptographic Signature Verification
func (k Keeper) ValidateSaltBindingWithSignature(
	ctx sdk.Context,
	salt []byte,
	address sdk.AccAddress,
	scopeID string,
	timestamp int64,
	signature []byte,
	pubKey []byte,
	algorithm string,
) error {
	// First, validate basic salt properties
	if err := k.validateSaltBasics(ctx, salt); err != nil {
		return err
	}
	if err := k.checkSaltUnused(ctx, types.ComputeSaltHash(salt)); err != nil {
		return err
	}

	// Validate cryptographic binding using the signature verification
	bindingTime := time.Unix(timestamp, 0)
	currentTime := ctx.BlockTime()

	if err := VerifySaltBinding(
		salt,
		address,
		scopeID,
		bindingTime,
		signature,
		pubKey,
		algorithm,
		currentTime,
	); err != nil {
		return types.ErrSaltBindingInvalid.Wrapf("salt binding verification failed: %v", err)
	}

	return nil
}

// markSaltUsed marks a salt as used
func (k Keeper) markSaltUsed(ctx sdk.Context, salt []byte) {
	saltHash := sha256.Sum256(salt)
	store := ctx.KVStore(k.skey)
	store.Set(types.SaltRegistryKey(saltHash[:]), []byte{1})
}

// ============================================================================
// Approved Clients
// ============================================================================

// approvedClientStore is the stored format of an approved client
type approvedClientStore struct {
	ClientID      string            `json:"client_id"`
	Name          string            `json:"name"`
	PublicKey     []byte            `json:"public_key"`
	Algorithm     string            `json:"algorithm"`
	Active        bool              `json:"active"`
	RegisteredAt  int64             `json:"registered_at"`
	DeactivatedAt int64             `json:"deactivated_at,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// GetApprovedClient returns an approved client by ID
func (k Keeper) GetApprovedClient(ctx sdk.Context, clientID string) (types.ApprovedClient, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ApprovedClientKey(clientID)
	bz := store.Get(key)
	if bz == nil {
		return types.ApprovedClient{}, false
	}

	var cs approvedClientStore
	if err := json.Unmarshal(bz, &cs); err != nil {
		return types.ApprovedClient{}, false
	}

	return types.ApprovedClient{
		ClientID:      cs.ClientID,
		Name:          cs.Name,
		PublicKey:     cs.PublicKey,
		Algorithm:     cs.Algorithm,
		Active:        cs.Active,
		RegisteredAt:  cs.RegisteredAt,
		DeactivatedAt: cs.DeactivatedAt,
		Metadata:      cs.Metadata,
	}, true
}

// SetApprovedClient stores an approved client
func (k Keeper) SetApprovedClient(ctx sdk.Context, client types.ApprovedClient) error {
	if err := client.Validate(); err != nil {
		return err
	}

	cs := approvedClientStore{
		ClientID:      client.ClientID,
		Name:          client.Name,
		PublicKey:     client.PublicKey,
		Algorithm:     client.Algorithm,
		Active:        client.Active,
		RegisteredAt:  client.RegisteredAt,
		DeactivatedAt: client.DeactivatedAt,
		Metadata:      client.Metadata,
	}

	bz, err := json.Marshal(&cs)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	store.Set(types.ApprovedClientKey(client.ClientID), bz)
	return nil
}

// IsClientApproved checks if a client is approved and active
func (k Keeper) IsClientApproved(ctx sdk.Context, clientID string) bool {
	client, found := k.GetApprovedClient(ctx, clientID)
	if !found {
		return false
	}
	return client.Active
}

// WithApprovedClients iterates over all approved clients
func (k Keeper) WithApprovedClients(ctx sdk.Context, fn func(client types.ApprovedClient) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.PrefixApprovedClient)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var cs approvedClientStore
		if err := json.Unmarshal(iter.Value(), &cs); err != nil {
			continue
		}

		client := types.ApprovedClient{
			ClientID:      cs.ClientID,
			Name:          cs.Name,
			PublicKey:     cs.PublicKey,
			Algorithm:     cs.Algorithm,
			Active:        cs.Active,
			RegisteredAt:  cs.RegisteredAt,
			DeactivatedAt: cs.DeactivatedAt,
			Metadata:      cs.Metadata,
		}

		if fn(client) {
			break
		}
	}
}
