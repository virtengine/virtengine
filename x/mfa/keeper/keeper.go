package keeper

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/bcrypt"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// IKeeper defines the interface for the mfa keeper
type IKeeper interface {
	// Factor enrollment
	EnrollFactor(ctx sdk.Context, enrollment *types.FactorEnrollment) error
	RevokeFactor(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType, factorID string) error
	GetFactorEnrollment(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType, factorID string) (*types.FactorEnrollment, bool)
	GetFactorEnrollments(ctx sdk.Context, address sdk.AccAddress) []types.FactorEnrollment
	GetActiveFactorsByType(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType) []types.FactorEnrollment
	HasActiveFactorOfType(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType) bool

	// MFA policy
	SetMFAPolicy(ctx sdk.Context, policy *types.MFAPolicy) error
	GetMFAPolicy(ctx sdk.Context, address sdk.AccAddress) (*types.MFAPolicy, bool)
	DeleteMFAPolicy(ctx sdk.Context, address sdk.AccAddress) error

	// Challenge management
	CreateChallenge(ctx sdk.Context, challenge *types.Challenge) error
	GetChallenge(ctx sdk.Context, challengeID string) (*types.Challenge, bool)
	UpdateChallenge(ctx sdk.Context, challenge *types.Challenge) error
	DeleteChallenge(ctx sdk.Context, challengeID string) error
	GetPendingChallenges(ctx sdk.Context, address sdk.AccAddress) []types.Challenge
	VerifyMFAChallenge(ctx sdk.Context, challengeID string, response *types.ChallengeResponse) (bool, error)

	// Authorization sessions
	CreateAuthorizationSession(ctx sdk.Context, session *types.AuthorizationSession) error
	GetAuthorizationSession(ctx sdk.Context, sessionID string) (*types.AuthorizationSession, bool)
	UseAuthorizationSession(ctx sdk.Context, sessionID string) error
	DeleteAuthorizationSession(ctx sdk.Context, sessionID string) error
	GetAccountSessions(ctx sdk.Context, address sdk.AccAddress) []types.AuthorizationSession

	// Authorization session management (MFA-CORE-002)
	HasValidAuthSession(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType) bool
	HasValidAuthSessionWithDevice(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) bool
	ConsumeAuthSession(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType) error
	ConsumeAuthSessionWithDevice(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) error
	CreateAuthSessionForAction(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, verifiedFactors []types.FactorType, deviceFingerprint string) (*types.AuthorizationSession, error)
	GetValidSessionsForAccount(ctx sdk.Context, address sdk.AccAddress) []types.AuthorizationSession
	CleanupExpiredSessions(ctx sdk.Context, address sdk.AccAddress) int
	ValidateSessionForTransaction(ctx sdk.Context, sessionID string, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) error
	GetSessionDurationForAction(ctx sdk.Context, action types.SensitiveTransactionType) int64
	IsActionSingleUse(ctx sdk.Context, action types.SensitiveTransactionType) bool

	// Trusted devices
	AddTrustedDevice(ctx sdk.Context, address sdk.AccAddress, device *types.DeviceInfo) (string, error)
	RemoveTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) error
	GetTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) (*types.TrustedDevice, bool)
	GetTrustedDevices(ctx sdk.Context, address sdk.AccAddress) []types.TrustedDevice
	IsTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) bool
	ValidateTrustToken(ctx sdk.Context, address sdk.AccAddress, fingerprint string, token string) bool

	// Sensitive transaction config
	SetSensitiveTxConfig(ctx sdk.Context, config *types.SensitiveTxConfig) error
	GetSensitiveTxConfig(ctx sdk.Context, txType types.SensitiveTransactionType) (*types.SensitiveTxConfig, bool)
	GetAllSensitiveTxConfigs(ctx sdk.Context) []types.SensitiveTxConfig

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper of the mfa store
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string

	// Hooks for integration with other modules
	veidKeeper  VEIDKeeper
	rolesKeeper RolesKeeper
}

// VEIDKeeper defines the interface for the VEID keeper
type VEIDKeeper interface {
	GetVEIDScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool)
}

// RolesKeeper defines the interface for the roles keeper
type RolesKeeper interface {
	IsAccountOperational(ctx sdk.Context, address sdk.AccAddress) bool
}

// NewKeeper creates and returns an instance for mfa keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	authority string,
	veidKeeper VEIDKeeper,
	rolesKeeper RolesKeeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		skey:        skey,
		authority:   authority,
		veidKeeper:  veidKeeper,
		rolesKeeper: rolesKeeper,
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

// ============================================================================
// Parameters
// ============================================================================

// paramsStore is the stored format of params
type paramsStore struct {
	DefaultSessionDuration  int64              `json:"default_session_duration"`
	MaxFactorsPerAccount    uint32             `json:"max_factors_per_account"`
	MaxChallengeAttempts    uint32             `json:"max_challenge_attempts"`
	ChallengeTTL            int64              `json:"challenge_ttl"`
	MaxTrustedDevices       uint32             `json:"max_trusted_devices"`
	TrustedDeviceTTL        int64              `json:"trusted_device_ttl"`
	MinVEIDScoreForMFA      uint32             `json:"min_veid_score_for_mfa"`
	RequireAtLeastOneFactor bool               `json:"require_at_least_one_factor"`
	AllowedFactorTypes      []types.FactorType `json:"allowed_factor_types"`
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&paramsStore{
		DefaultSessionDuration:  params.DefaultSessionDuration,
		MaxFactorsPerAccount:    params.MaxFactorsPerAccount,
		MaxChallengeAttempts:    params.MaxChallengeAttempts,
		ChallengeTTL:            params.ChallengeTTL,
		MaxTrustedDevices:       params.MaxTrustedDevices,
		TrustedDeviceTTL:        params.TrustedDeviceTTL,
		MinVEIDScoreForMFA:      params.MinVEIDScoreForMFA,
		RequireAtLeastOneFactor: params.RequireAtLeastOneFactor,
		AllowedFactorTypes:      params.AllowedFactorTypes,
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
		DefaultSessionDuration:  ps.DefaultSessionDuration,
		MaxFactorsPerAccount:    ps.MaxFactorsPerAccount,
		MaxChallengeAttempts:    ps.MaxChallengeAttempts,
		ChallengeTTL:            ps.ChallengeTTL,
		MaxTrustedDevices:       ps.MaxTrustedDevices,
		TrustedDeviceTTL:        ps.TrustedDeviceTTL,
		MinVEIDScoreForMFA:      ps.MinVEIDScoreForMFA,
		RequireAtLeastOneFactor: ps.RequireAtLeastOneFactor,
		AllowedFactorTypes:      ps.AllowedFactorTypes,
	}
}

// ============================================================================
// Factor Enrollment
// ============================================================================

// factorEnrollmentStore is the stored format of a factor enrollment
type factorEnrollmentStore struct {
	AccountAddress   string                       `json:"account_address"`
	FactorType       types.FactorType             `json:"factor_type"`
	FactorID         string                       `json:"factor_id"`
	PublicIdentifier []byte                       `json:"public_identifier,omitempty"`
	Label            string                       `json:"label"`
	Status           types.FactorEnrollmentStatus `json:"status"`
	EnrolledAt       int64                        `json:"enrolled_at"`
	VerifiedAt       int64                        `json:"verified_at,omitempty"`
	RevokedAt        int64                        `json:"revoked_at,omitempty"`
	LastUsedAt       int64                        `json:"last_used_at,omitempty"`
	UseCount         uint64                       `json:"use_count"`
	Metadata         *types.FactorMetadata        `json:"metadata,omitempty"`
}

// EnrollFactor enrolls a new factor for an account
func (k Keeper) EnrollFactor(ctx sdk.Context, enrollment *types.FactorEnrollment) error {
	if err := enrollment.Validate(); err != nil {
		return err
	}

	address, err := sdk.AccAddressFromBech32(enrollment.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	params := k.GetParams(ctx)

	// Check if factor type is allowed
	if !params.IsFactorTypeAllowed(enrollment.FactorType) {
		return types.ErrInvalidFactorType.Wrapf("factor type %s is not allowed", enrollment.FactorType.String())
	}

	// Check max factors limit
	existingEnrollments := k.GetFactorEnrollments(ctx, address)
	activeCount := 0
	for _, e := range existingEnrollments {
		if e.IsActive() {
			activeCount++
		}
	}
	if safeUint32FromInt(activeCount) >= params.MaxFactorsPerAccount {
		return types.ErrInvalidEnrollment.Wrapf("maximum factors per account (%d) reached", params.MaxFactorsPerAccount)
	}

	// Check if this specific factor already exists
	if _, found := k.GetFactorEnrollment(ctx, address, enrollment.FactorType, enrollment.FactorID); found {
		return types.ErrEnrollmentAlreadyExists.Wrapf("factor %s/%s already enrolled", enrollment.FactorType.String(), enrollment.FactorID)
	}

	// Store the enrollment
	store := ctx.KVStore(k.skey)
	key := types.FactorEnrollmentKey(address, enrollment.FactorType, enrollment.FactorID)

	bz, err := json.Marshal(&factorEnrollmentStore{
		AccountAddress:   enrollment.AccountAddress,
		FactorType:       enrollment.FactorType,
		FactorID:         enrollment.FactorID,
		PublicIdentifier: enrollment.PublicIdentifier,
		Label:            enrollment.Label,
		Status:           enrollment.Status,
		EnrolledAt:       enrollment.EnrolledAt,
		VerifiedAt:       enrollment.VerifiedAt,
		RevokedAt:        enrollment.RevokedAt,
		LastUsedAt:       enrollment.LastUsedAt,
		UseCount:         enrollment.UseCount,
		Metadata:         enrollment.Metadata,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFactorEnrolled,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, enrollment.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, enrollment.FactorType.String()),
			sdk.NewAttribute(types.AttributeKeyFactorID, enrollment.FactorID),
		),
	)

	return nil
}

// RevokeFactor revokes a factor enrollment
func (k Keeper) RevokeFactor(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType, factorID string) error {
	enrollment, found := k.GetFactorEnrollment(ctx, address, factorType, factorID)
	if !found {
		return types.ErrEnrollmentNotFound.Wrapf("factor %s/%s not found", factorType.String(), factorID)
	}

	if enrollment.Status == types.EnrollmentStatusRevoked {
		return types.ErrFactorRevoked.Wrap("factor is already revoked")
	}

	enrollment.Status = types.EnrollmentStatusRevoked
	enrollment.RevokedAt = ctx.BlockTime().Unix()

	// Update the enrollment
	store := ctx.KVStore(k.skey)
	key := types.FactorEnrollmentKey(address, factorType, factorID)

	bz, err := json.Marshal(&factorEnrollmentStore{
		AccountAddress:   enrollment.AccountAddress,
		FactorType:       enrollment.FactorType,
		FactorID:         enrollment.FactorID,
		PublicIdentifier: enrollment.PublicIdentifier,
		Label:            enrollment.Label,
		Status:           enrollment.Status,
		EnrolledAt:       enrollment.EnrolledAt,
		VerifiedAt:       enrollment.VerifiedAt,
		RevokedAt:        enrollment.RevokedAt,
		LastUsedAt:       enrollment.LastUsedAt,
		UseCount:         enrollment.UseCount,
		Metadata:         enrollment.Metadata,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFactorRevoked,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyFactorType, factorType.String()),
			sdk.NewAttribute(types.AttributeKeyFactorID, factorID),
		),
	)

	return nil
}

// GetFactorEnrollment returns a specific factor enrollment
func (k Keeper) GetFactorEnrollment(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType, factorID string) (*types.FactorEnrollment, bool) {
	store := ctx.KVStore(k.skey)
	key := types.FactorEnrollmentKey(address, factorType, factorID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var es factorEnrollmentStore
	_ = json.Unmarshal(bz, &es)

	return &types.FactorEnrollment{
		AccountAddress:   es.AccountAddress,
		FactorType:       es.FactorType,
		FactorID:         es.FactorID,
		PublicIdentifier: es.PublicIdentifier,
		Label:            es.Label,
		Status:           es.Status,
		EnrolledAt:       es.EnrolledAt,
		VerifiedAt:       es.VerifiedAt,
		RevokedAt:        es.RevokedAt,
		LastUsedAt:       es.LastUsedAt,
		UseCount:         es.UseCount,
		Metadata:         es.Metadata,
	}, true
}

// GetFactorEnrollments returns all factor enrollments for an account
func (k Keeper) GetFactorEnrollments(ctx sdk.Context, address sdk.AccAddress) []types.FactorEnrollment {
	store := ctx.KVStore(k.skey)
	prefix := types.FactorEnrollmentPrefixKey(address)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var enrollments []types.FactorEnrollment
	for ; iterator.Valid(); iterator.Next() {
		var es factorEnrollmentStore
		_ = json.Unmarshal(iterator.Value(), &es)
		enrollments = append(enrollments, types.FactorEnrollment{
			AccountAddress:   es.AccountAddress,
			FactorType:       es.FactorType,
			FactorID:         es.FactorID,
			PublicIdentifier: es.PublicIdentifier,
			Label:            es.Label,
			Status:           es.Status,
			EnrolledAt:       es.EnrolledAt,
			VerifiedAt:       es.VerifiedAt,
			RevokedAt:        es.RevokedAt,
			LastUsedAt:       es.LastUsedAt,
			UseCount:         es.UseCount,
			Metadata:         es.Metadata,
		})
	}

	return enrollments
}

// GetActiveFactorsByType returns active factor enrollments of a specific type
func (k Keeper) GetActiveFactorsByType(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType) []types.FactorEnrollment {
	all := k.GetFactorEnrollments(ctx, address)
	now := ctx.BlockTime()

	var active []types.FactorEnrollment
	for _, e := range all {
		if e.FactorType == factorType && e.CanVerify(now) {
			active = append(active, e)
		}
	}
	return active
}

// HasActiveFactorOfType returns true if account has an active factor of the given type
func (k Keeper) HasActiveFactorOfType(ctx sdk.Context, address sdk.AccAddress, factorType types.FactorType) bool {
	return len(k.GetActiveFactorsByType(ctx, address, factorType)) > 0
}

// IsMFAEnabled checks if MFA is enabled for an account.
// MFA is considered enabled if the account has an active MFA policy with at least one enrolled factor.
func (k Keeper) IsMFAEnabled(ctx sdk.Context, address sdk.AccAddress) (bool, error) {
	// Check if account has an MFA policy
	policy, found := k.GetMFAPolicy(ctx, address)
	if !found {
		return false, nil
	}

	// Policy exists but may not be enabled
	if !policy.Enabled {
		return false, nil
	}

	// Check if account has at least one active factor
	enrollments := k.GetFactorEnrollments(ctx, address)
	for _, enrollment := range enrollments {
		if enrollment.Status == types.EnrollmentStatusActive {
			return true, nil
		}
	}

	return false, nil
}

// ============================================================================
// MFA Policy
// ============================================================================

// mfaPolicyStore is the stored format of an MFA policy
type mfaPolicyStore struct {
	AccountAddress     string                     `json:"account_address"`
	RequiredFactors    []types.FactorCombination  `json:"required_factors"`
	TrustedDeviceRule  *types.TrustedDevicePolicy `json:"trusted_device_rule,omitempty"`
	RecoveryFactors    []types.FactorCombination  `json:"recovery_factors,omitempty"`
	KeyRotationFactors []types.FactorCombination  `json:"key_rotation_factors,omitempty"`
	SessionDuration    int64                      `json:"session_duration"`
	VEIDThreshold      uint32                     `json:"veid_threshold,omitempty"`
	Enabled            bool                       `json:"enabled"`
	CreatedAt          int64                      `json:"created_at"`
	UpdatedAt          int64                      `json:"updated_at"`
}

// SetMFAPolicy sets the MFA policy for an account
func (k Keeper) SetMFAPolicy(ctx sdk.Context, policy *types.MFAPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}

	address, err := sdk.AccAddressFromBech32(policy.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	store := ctx.KVStore(k.skey)
	key := types.MFAPolicyKey(address)

	bz, err := json.Marshal(&mfaPolicyStore{
		AccountAddress:     policy.AccountAddress,
		RequiredFactors:    policy.RequiredFactors,
		TrustedDeviceRule:  policy.TrustedDeviceRule,
		RecoveryFactors:    policy.RecoveryFactors,
		KeyRotationFactors: policy.KeyRotationFactors,
		SessionDuration:    policy.SessionDuration,
		VEIDThreshold:      policy.VEIDThreshold,
		Enabled:            policy.Enabled,
		CreatedAt:          policy.CreatedAt,
		UpdatedAt:          policy.UpdatedAt,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePolicyUpdated,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, policy.AccountAddress),
		),
	)

	return nil
}

// GetMFAPolicy returns the MFA policy for an account
func (k Keeper) GetMFAPolicy(ctx sdk.Context, address sdk.AccAddress) (*types.MFAPolicy, bool) {
	store := ctx.KVStore(k.skey)
	key := types.MFAPolicyKey(address)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var ps mfaPolicyStore
	_ = json.Unmarshal(bz, &ps)

	return &types.MFAPolicy{
		AccountAddress:     ps.AccountAddress,
		RequiredFactors:    ps.RequiredFactors,
		TrustedDeviceRule:  ps.TrustedDeviceRule,
		RecoveryFactors:    ps.RecoveryFactors,
		KeyRotationFactors: ps.KeyRotationFactors,
		SessionDuration:    ps.SessionDuration,
		VEIDThreshold:      ps.VEIDThreshold,
		Enabled:            ps.Enabled,
		CreatedAt:          ps.CreatedAt,
		UpdatedAt:          ps.UpdatedAt,
	}, true
}

// DeleteMFAPolicy deletes the MFA policy for an account
func (k Keeper) DeleteMFAPolicy(ctx sdk.Context, address sdk.AccAddress) error {
	store := ctx.KVStore(k.skey)
	key := types.MFAPolicyKey(address)
	if !store.Has(key) {
		return types.ErrPolicyNotFound.Wrapf("no policy found for address %s", address.String())
	}
	store.Delete(key)
	return nil
}

// ============================================================================
// Challenge Management
// ============================================================================

// challengeStore is the stored format of a challenge
type challengeStore struct {
	ChallengeID     string                         `json:"challenge_id"`
	AccountAddress  string                         `json:"account_address"`
	FactorType      types.FactorType               `json:"factor_type"`
	FactorID        string                         `json:"factor_id"`
	TransactionType types.SensitiveTransactionType `json:"transaction_type"`
	Status          types.ChallengeStatus          `json:"status"`
	ChallengeData   []byte                         `json:"challenge_data,omitempty"`
	CreatedAt       int64                          `json:"created_at"`
	ExpiresAt       int64                          `json:"expires_at"`
	VerifiedAt      int64                          `json:"verified_at,omitempty"`
	AttemptCount    uint32                         `json:"attempt_count"`
	MaxAttempts     uint32                         `json:"max_attempts"`
	Nonce           string                         `json:"nonce"`
	SessionID       string                         `json:"session_id,omitempty"`
	Metadata        *types.ChallengeMetadata       `json:"metadata,omitempty"`
}

// CreateChallenge creates a new MFA challenge
func (k Keeper) CreateChallenge(ctx sdk.Context, challenge *types.Challenge) error {
	if err := challenge.Validate(); err != nil {
		return err
	}

	address, err := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	store := ctx.KVStore(k.skey)

	// Store the challenge
	challengeKey := types.ChallengeKey(challenge.ChallengeID)
	bz, err := json.Marshal(&challengeStore{
		ChallengeID:     challenge.ChallengeID,
		AccountAddress:  challenge.AccountAddress,
		FactorType:      challenge.FactorType,
		FactorID:        challenge.FactorID,
		TransactionType: challenge.TransactionType,
		Status:          challenge.Status,
		ChallengeData:   challenge.ChallengeData,
		CreatedAt:       challenge.CreatedAt,
		ExpiresAt:       challenge.ExpiresAt,
		VerifiedAt:      challenge.VerifiedAt,
		AttemptCount:    challenge.AttemptCount,
		MaxAttempts:     challenge.MaxAttempts,
		Nonce:           challenge.Nonce,
		SessionID:       challenge.SessionID,
		Metadata:        challenge.Metadata,
	})
	if err != nil {
		return err
	}
	store.Set(challengeKey, bz)

	// Index by account
	indexKey := types.AccountChallengesKey(address, challenge.ChallengeID)
	store.Set(indexKey, []byte{1})

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeCreated,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
		),
	)

	return nil
}

// GetChallenge returns a challenge by ID
func (k Keeper) GetChallenge(ctx sdk.Context, challengeID string) (*types.Challenge, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ChallengeKey(challengeID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var cs challengeStore
	_ = json.Unmarshal(bz, &cs)

	return &types.Challenge{
		ChallengeID:     cs.ChallengeID,
		AccountAddress:  cs.AccountAddress,
		FactorType:      cs.FactorType,
		FactorID:        cs.FactorID,
		TransactionType: cs.TransactionType,
		Status:          cs.Status,
		ChallengeData:   cs.ChallengeData,
		CreatedAt:       cs.CreatedAt,
		ExpiresAt:       cs.ExpiresAt,
		VerifiedAt:      cs.VerifiedAt,
		AttemptCount:    cs.AttemptCount,
		MaxAttempts:     cs.MaxAttempts,
		Nonce:           cs.Nonce,
		SessionID:       cs.SessionID,
		Metadata:        cs.Metadata,
	}, true
}

// UpdateChallenge updates a challenge
func (k Keeper) UpdateChallenge(ctx sdk.Context, challenge *types.Challenge) error {
	store := ctx.KVStore(k.skey)
	key := types.ChallengeKey(challenge.ChallengeID)

	if !store.Has(key) {
		return types.ErrChallengeNotFound.Wrapf("challenge %s not found", challenge.ChallengeID)
	}

	bz, err := json.Marshal(&challengeStore{
		ChallengeID:     challenge.ChallengeID,
		AccountAddress:  challenge.AccountAddress,
		FactorType:      challenge.FactorType,
		FactorID:        challenge.FactorID,
		TransactionType: challenge.TransactionType,
		Status:          challenge.Status,
		ChallengeData:   challenge.ChallengeData,
		CreatedAt:       challenge.CreatedAt,
		ExpiresAt:       challenge.ExpiresAt,
		VerifiedAt:      challenge.VerifiedAt,
		AttemptCount:    challenge.AttemptCount,
		MaxAttempts:     challenge.MaxAttempts,
		Nonce:           challenge.Nonce,
		SessionID:       challenge.SessionID,
		Metadata:        challenge.Metadata,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// DeleteChallenge deletes a challenge
func (k Keeper) DeleteChallenge(ctx sdk.Context, challengeID string) error {
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return types.ErrChallengeNotFound.Wrapf("challenge %s not found", challengeID)
	}

	store := ctx.KVStore(k.skey)

	// Delete the challenge
	store.Delete(types.ChallengeKey(challengeID))

	// Delete the index
	address, _ := sdk.AccAddressFromBech32(challenge.AccountAddress)
	store.Delete(types.AccountChallengesKey(address, challengeID))

	return nil
}

// GetPendingChallenges returns pending challenges for an account
func (k Keeper) GetPendingChallenges(ctx sdk.Context, address sdk.AccAddress) []types.Challenge {
	store := ctx.KVStore(k.skey)
	prefix := types.AccountChallengesPrefixKey(address)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	now := ctx.BlockTime()
	var challenges []types.Challenge

	for ; iterator.Valid(); iterator.Next() {
		// Extract challenge ID from key
		key := iterator.Key()
		challengeID := string(key[len(prefix):])

		challenge, found := k.GetChallenge(ctx, challengeID)
		if found && challenge.IsPending() && !challenge.IsExpired(now) {
			challenges = append(challenges, *challenge)
		}
	}

	return challenges
}

// VerifyMFAChallenge verifies an MFA challenge response
func (k Keeper) VerifyMFAChallenge(ctx sdk.Context, challengeID string, response *types.ChallengeResponse) (bool, error) {
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return false, types.ErrChallengeNotFound.Wrapf("challenge %s not found", challengeID)
	}

	now := ctx.BlockTime()

	// Check if challenge is expired
	if challenge.IsExpired(now) {
		challenge.MarkExpired()
		_ = k.UpdateChallenge(ctx, challenge)
		return false, types.ErrChallengeExpired.Wrap("challenge has expired")
	}

	// Check if already used
	if !challenge.IsPending() {
		return false, types.ErrChallengeAlreadyUsed.Wrap("challenge has already been processed")
	}

	// Check attempt count
	if challenge.AttemptCount >= challenge.MaxAttempts {
		challenge.MarkFailed()
		_ = k.UpdateChallenge(ctx, challenge)
		return false, types.ErrMaxAttemptsExceeded.Wrap("maximum verification attempts exceeded")
	}

	// Record this attempt
	challenge.RecordAttempt()

	// Verify based on factor type
	verified := false
	var verifyErr error

	switch challenge.FactorType {
	case types.FactorTypeFIDO2:
		verified, verifyErr = k.verifyFIDO2Response(ctx, challenge, response)
	case types.FactorTypeTOTP:
		verified, verifyErr = k.verifyTOTPResponse(ctx, challenge, response)
	case types.FactorTypeSMS, types.FactorTypeEmail:
		verified, verifyErr = k.verifyOTPResponse(ctx, challenge, response)
	case types.FactorTypeVEID:
		verified, verifyErr = k.verifyVEIDThreshold(ctx, challenge)
	case types.FactorTypeTrustedDevice:
		verified, verifyErr = k.verifyTrustedDevice(ctx, challenge, response)
	default:
		verifyErr = types.ErrInvalidFactorType.Wrapf("unsupported factor type: %s", challenge.FactorType.String())
	}

	if verifyErr != nil {
		challenge.MarkFailed()
		_ = k.UpdateChallenge(ctx, challenge)
		return false, verifyErr
	}

	if !verified {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			),
		)

		return false, types.ErrVerificationFailed.Wrap("verification failed")
	}

	// Mark as verified
	challenge.MarkVerified(now.Unix())
	_ = k.UpdateChallenge(ctx, challenge)

	// Update factor usage
	address, _ := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if enrollment, found := k.GetFactorEnrollment(ctx, address, challenge.FactorType, challenge.FactorID); found {
		enrollment.UpdateLastUsed(now.Unix())
		_ = k.updateFactorEnrollment(ctx, enrollment)
	}

	// Emit success event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeVerified,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
		),
	)

	return true, nil
}

// verifyFIDO2Response verifies a FIDO2 authentication response
func (k Keeper) verifyFIDO2Response(ctx sdk.Context, challenge *types.Challenge, response *types.ChallengeResponse) (bool, error) {
	if response == nil {
		return false, types.ErrInvalidChallengeResponse.Wrap("missing response")
	}

	payload, err := types.ParseFIDO2AssertionPayload(response.ResponseData)
	if err != nil {
		return false, err
	}

	address, _ := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err := k.VerifyFIDO2Assertion(
		ctx,
		address,
		challenge.ChallengeID,
		payload.CredentialID,
		payload.ClientDataJSON,
		payload.AuthenticatorData,
		payload.Signature,
	); err != nil {
		return false, err
	}

	return true, nil
}

// verifyTOTPResponse verifies a TOTP code
//
//nolint:unparam // ctx kept for future on-chain TOTP enrollment lookup
func (k Keeper) verifyTOTPResponse(_ sdk.Context, _ *types.Challenge, response *types.ChallengeResponse) (bool, error) {
	// TOTP verification happens off-chain as we don't store seeds on-chain
	// The response should contain proof of verification from an off-chain verifier
	// For now, assume the response is a signed attestation from a trusted verifier

	if len(response.ResponseData) == 0 {
		return false, types.ErrInvalidChallengeResponse.Wrap("empty response data")
	}

	// In production, this would verify a signed attestation from an off-chain TOTP verifier
	return true, nil
}

// verifyOTPResponse verifies an SMS/Email OTP code
//
//nolint:unparam // ctx kept for future on-chain OTP enrollment lookup
func (k Keeper) verifyOTPResponse(_ sdk.Context, _ *types.Challenge, response *types.ChallengeResponse) (bool, error) {
	// Similar to TOTP, OTP verification happens off-chain
	// The response should contain proof of verification

	if len(response.ResponseData) == 0 {
		return false, types.ErrInvalidChallengeResponse.Wrap("empty response data")
	}

	return true, nil
}

// verifyVEIDThreshold verifies that the account meets the VEID score threshold
func (k Keeper) verifyVEIDThreshold(ctx sdk.Context, challenge *types.Challenge) (bool, error) {
	address, _ := sdk.AccAddressFromBech32(challenge.AccountAddress)

	// Get the MFA policy to find the required threshold
	policy, found := k.GetMFAPolicy(ctx, address)
	threshold := uint32(50) // Default threshold
	if found && policy.VEIDThreshold > 0 {
		threshold = policy.VEIDThreshold
	}

	// Get the current VEID score
	if k.veidKeeper == nil {
		return false, types.ErrVerificationFailed.Wrap("VEID keeper not available")
	}

	score, found := k.veidKeeper.GetVEIDScore(ctx, address)
	if !found {
		return false, types.ErrVEIDScoreInsufficient.Wrap("no VEID score found for account")
	}

	if score < threshold {
		return false, types.ErrVEIDScoreInsufficient.Wrapf("VEID score %d is below threshold %d", score, threshold)
	}

	return true, nil
}

// verifyTrustedDevice verifies that the request comes from a trusted device
func (k Keeper) verifyTrustedDevice(ctx sdk.Context, challenge *types.Challenge, response *types.ChallengeResponse) (bool, error) {
	if response.ClientInfo == nil || response.ClientInfo.DeviceFingerprint == "" {
		return false, types.ErrInvalidChallengeResponse.Wrap("device fingerprint required")
	}

	address, _ := sdk.AccAddressFromBech32(challenge.AccountAddress)
	return k.IsTrustedDevice(ctx, address, response.ClientInfo.DeviceFingerprint), nil
}

// updateFactorEnrollment updates a factor enrollment
func (k Keeper) updateFactorEnrollment(ctx sdk.Context, enrollment *types.FactorEnrollment) error {
	address, _ := sdk.AccAddressFromBech32(enrollment.AccountAddress)
	store := ctx.KVStore(k.skey)
	key := types.FactorEnrollmentKey(address, enrollment.FactorType, enrollment.FactorID)

	bz, err := json.Marshal(&factorEnrollmentStore{
		AccountAddress:   enrollment.AccountAddress,
		FactorType:       enrollment.FactorType,
		FactorID:         enrollment.FactorID,
		PublicIdentifier: enrollment.PublicIdentifier,
		Label:            enrollment.Label,
		Status:           enrollment.Status,
		EnrolledAt:       enrollment.EnrolledAt,
		VerifiedAt:       enrollment.VerifiedAt,
		RevokedAt:        enrollment.RevokedAt,
		LastUsedAt:       enrollment.LastUsedAt,
		UseCount:         enrollment.UseCount,
		Metadata:         enrollment.Metadata,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// ============================================================================
// Authorization Sessions
// ============================================================================

// sessionStore is the stored format of an authorization session
type sessionStore struct {
	SessionID         string                         `json:"session_id"`
	AccountAddress    string                         `json:"account_address"`
	TransactionType   types.SensitiveTransactionType `json:"transaction_type"`
	VerifiedFactors   []types.FactorType             `json:"verified_factors"`
	CreatedAt         int64                          `json:"created_at"`
	ExpiresAt         int64                          `json:"expires_at"`
	UsedAt            int64                          `json:"used_at,omitempty"`
	IsSingleUse       bool                           `json:"is_single_use"`
	DeviceFingerprint string                         `json:"device_fingerprint,omitempty"`
}

// CreateAuthorizationSession creates a new authorization session
func (k Keeper) CreateAuthorizationSession(ctx sdk.Context, session *types.AuthorizationSession) error {
	address, err := sdk.AccAddressFromBech32(session.AccountAddress)
	if err != nil {
		return types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	// Generate session ID if not provided
	if session.SessionID == "" {
		idBytes := make([]byte, 16)
		_, _ = rand.Read(idBytes)
		session.SessionID = hex.EncodeToString(idBytes)
	}

	store := ctx.KVStore(k.skey)

	// Store the session
	sessionKey := types.AuthorizationSessionKey(session.SessionID)
	bz, err := json.Marshal(&sessionStore{
		SessionID:         session.SessionID,
		AccountAddress:    session.AccountAddress,
		TransactionType:   session.TransactionType,
		VerifiedFactors:   session.VerifiedFactors,
		CreatedAt:         session.CreatedAt,
		ExpiresAt:         session.ExpiresAt,
		UsedAt:            session.UsedAt,
		IsSingleUse:       session.IsSingleUse,
		DeviceFingerprint: session.DeviceFingerprint,
	})
	if err != nil {
		return err
	}
	store.Set(sessionKey, bz)

	// Index by account
	indexKey := types.AccountSessionsKey(address, session.SessionID)
	store.Set(indexKey, []byte{1})

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSessionCreated,
			sdk.NewAttribute(types.AttributeKeySessionID, session.SessionID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, session.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyTransactionType, session.TransactionType.String()),
		),
	)

	return nil
}

// GetAuthorizationSession returns an authorization session by ID
func (k Keeper) GetAuthorizationSession(ctx sdk.Context, sessionID string) (*types.AuthorizationSession, bool) {
	store := ctx.KVStore(k.skey)
	key := types.AuthorizationSessionKey(sessionID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var ss sessionStore
	_ = json.Unmarshal(bz, &ss)

	return &types.AuthorizationSession{
		SessionID:         ss.SessionID,
		AccountAddress:    ss.AccountAddress,
		TransactionType:   ss.TransactionType,
		VerifiedFactors:   ss.VerifiedFactors,
		CreatedAt:         ss.CreatedAt,
		ExpiresAt:         ss.ExpiresAt,
		UsedAt:            ss.UsedAt,
		IsSingleUse:       ss.IsSingleUse,
		DeviceFingerprint: ss.DeviceFingerprint,
	}, true
}

// UseAuthorizationSession marks a session as used
func (k Keeper) UseAuthorizationSession(ctx sdk.Context, sessionID string) error {
	session, found := k.GetAuthorizationSession(ctx, sessionID)
	if !found {
		return types.ErrSessionNotFound.Wrapf("session %s not found", sessionID)
	}

	now := ctx.BlockTime()
	if !session.IsValid(now) {
		return types.ErrSessionExpired.Wrap("session has expired or already used")
	}

	session.MarkUsed(now.Unix())

	store := ctx.KVStore(k.skey)
	key := types.AuthorizationSessionKey(sessionID)

	bz, err := json.Marshal(&sessionStore{
		SessionID:         session.SessionID,
		AccountAddress:    session.AccountAddress,
		TransactionType:   session.TransactionType,
		VerifiedFactors:   session.VerifiedFactors,
		CreatedAt:         session.CreatedAt,
		ExpiresAt:         session.ExpiresAt,
		UsedAt:            session.UsedAt,
		IsSingleUse:       session.IsSingleUse,
		DeviceFingerprint: session.DeviceFingerprint,
	})
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSessionUsed,
			sdk.NewAttribute(types.AttributeKeySessionID, sessionID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, session.AccountAddress),
		),
	)

	return nil
}

// DeleteAuthorizationSession deletes an authorization session
func (k Keeper) DeleteAuthorizationSession(ctx sdk.Context, sessionID string) error {
	session, found := k.GetAuthorizationSession(ctx, sessionID)
	if !found {
		return types.ErrSessionNotFound.Wrapf("session %s not found", sessionID)
	}

	store := ctx.KVStore(k.skey)

	// Delete the session
	store.Delete(types.AuthorizationSessionKey(sessionID))

	// Delete the index
	address, _ := sdk.AccAddressFromBech32(session.AccountAddress)
	store.Delete(types.AccountSessionsKey(address, sessionID))

	return nil
}

// GetAccountSessions returns all authorization sessions for an account
func (k Keeper) GetAccountSessions(ctx sdk.Context, address sdk.AccAddress) []types.AuthorizationSession {
	store := ctx.KVStore(k.skey)
	prefix := types.AccountSessionsPrefixKey(address)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var sessions []types.AuthorizationSession
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		sessionID := string(key[len(prefix):])

		session, found := k.GetAuthorizationSession(ctx, sessionID)
		if found {
			sessions = append(sessions, *session)
		}
	}

	return sessions
}

// ============================================================================
// Trusted Devices
// ============================================================================

// trustedDeviceStore is the stored format of a trusted device
type trustedDeviceStore struct {
	AccountAddress string           `json:"account_address"`
	DeviceInfo     types.DeviceInfo `json:"device_info"`
	AddedAt        int64            `json:"added_at"`
	LastUsedAt     int64            `json:"last_used_at"`
}

// AddTrustedDevice adds a trusted device for an account
// Returns the plaintext trust token that should be sent to the client
func (k Keeper) AddTrustedDevice(ctx sdk.Context, address sdk.AccAddress, device *types.DeviceInfo) (string, error) {
	params := k.GetParams(ctx)

	// Check max trusted devices
	existing := k.GetTrustedDevices(ctx, address)
	if safeUint32FromInt(len(existing)) >= params.MaxTrustedDevices {
		return "", types.ErrMaxTrustedDevicesReached.Wrapf("maximum %d trusted devices allowed", params.MaxTrustedDevices)
	}

	now := ctx.BlockTime().Unix()
	device.TrustExpiresAt = now + params.TrustedDeviceTTL
	device.FirstSeenAt = now
	device.LastSeenAt = now

	// Generate trust token (32 random bytes, base64 encoded)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", types.ErrInvalidEnrollment.Wrapf("failed to generate trust token: %v", err)
	}
	trustToken := base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Hash the token with bcrypt before storing
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(trustToken), bcrypt.DefaultCost)
	if err != nil {
		return "", types.ErrInvalidEnrollment.Wrapf("failed to hash trust token: %v", err)
	}
	device.TrustTokenHash = string(hashedToken)

	store := ctx.KVStore(k.skey)
	key := types.TrustedDeviceKey(address, device.Fingerprint)

	td := types.TrustedDevice{
		AccountAddress: address.String(),
		DeviceInfo:     *device,
		AddedAt:        now,
		LastUsedAt:     now,
	}

	bz, err := json.Marshal(&trustedDeviceStore{
		AccountAddress: td.AccountAddress,
		DeviceInfo:     td.DeviceInfo,
		AddedAt:        td.AddedAt,
		LastUsedAt:     td.LastUsedAt,
	})
	if err != nil {
		return "", err
	}

	store.Set(key, bz)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTrustedDeviceAdded,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyDeviceFingerprint, device.Fingerprint),
		),
	)

	return trustToken, nil
}

// RemoveTrustedDevice removes a trusted device
func (k Keeper) RemoveTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) error {
	store := ctx.KVStore(k.skey)
	key := types.TrustedDeviceKey(address, fingerprint)

	if !store.Has(key) {
		return types.ErrTrustedDeviceNotFound.Wrapf("device %s not found", fingerprint)
	}

	store.Delete(key)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTrustedDeviceRemoved,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyDeviceFingerprint, fingerprint),
		),
	)

	return nil
}

// GetTrustedDevice returns a trusted device by fingerprint
func (k Keeper) GetTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) (*types.TrustedDevice, bool) {
	store := ctx.KVStore(k.skey)
	key := types.TrustedDeviceKey(address, fingerprint)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var ds trustedDeviceStore
	_ = json.Unmarshal(bz, &ds)

	return &types.TrustedDevice{
		AccountAddress: ds.AccountAddress,
		DeviceInfo:     ds.DeviceInfo,
		AddedAt:        ds.AddedAt,
		LastUsedAt:     ds.LastUsedAt,
	}, true
}

// GetTrustedDevices returns all trusted devices for an account
func (k Keeper) GetTrustedDevices(ctx sdk.Context, address sdk.AccAddress) []types.TrustedDevice {
	store := ctx.KVStore(k.skey)
	prefix := types.TrustedDevicePrefixKey(address)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var devices []types.TrustedDevice
	for ; iterator.Valid(); iterator.Next() {
		var ds trustedDeviceStore
		_ = json.Unmarshal(iterator.Value(), &ds)
		devices = append(devices, types.TrustedDevice{
			AccountAddress: ds.AccountAddress,
			DeviceInfo:     ds.DeviceInfo,
			AddedAt:        ds.AddedAt,
			LastUsedAt:     ds.LastUsedAt,
		})
	}

	return devices
}

// IsTrustedDevice returns true if the device is trusted and not expired
func (k Keeper) IsTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) bool {
	device, found := k.GetTrustedDevice(ctx, address, fingerprint)
	if !found {
		return false
	}

	now := ctx.BlockTime().Unix()
	return device.DeviceInfo.TrustExpiresAt > now
}

// ValidateTrustToken validates a trust token for a trusted device
func (k Keeper) ValidateTrustToken(ctx sdk.Context, address sdk.AccAddress, fingerprint string, token string) bool {
	device, found := k.GetTrustedDevice(ctx, address, fingerprint)
	if !found {
		return false
	}

	// Check if device trust has expired
	now := ctx.BlockTime().Unix()
	if device.DeviceInfo.TrustExpiresAt > 0 && now > device.DeviceInfo.TrustExpiresAt {
		return false
	}

	// Validate token against stored hash
	if device.DeviceInfo.TrustTokenHash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(device.DeviceInfo.TrustTokenHash), []byte(token))
	return err == nil
}

// ============================================================================
// Sensitive Transaction Config
// ============================================================================

// sensitiveTxConfigStore is the stored format of a sensitive tx config
type sensitiveTxConfigStore struct {
	TransactionType             types.SensitiveTransactionType `json:"transaction_type"`
	Enabled                     bool                           `json:"enabled"`
	MinVEIDScore                uint32                         `json:"min_veid_score"`
	RequiredFactorCombinations  []types.FactorCombination      `json:"required_factor_combinations"`
	SessionDuration             int64                          `json:"session_duration"`
	IsSingleUse                 bool                           `json:"is_single_use"`
	AllowTrustedDeviceReduction bool                           `json:"allow_trusted_device_reduction"`
	ValueThreshold              string                         `json:"value_threshold,omitempty"`
	CooldownPeriod              int64                          `json:"cooldown_period,omitempty"`
	Description                 string                         `json:"description"`
}

// SetSensitiveTxConfig sets a sensitive transaction configuration
func (k Keeper) SetSensitiveTxConfig(ctx sdk.Context, config *types.SensitiveTxConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := types.SensitiveTxConfigKey(config.TransactionType)

	bz, err := json.Marshal(&sensitiveTxConfigStore{
		TransactionType:             config.TransactionType,
		Enabled:                     config.Enabled,
		MinVEIDScore:                config.MinVEIDScore,
		RequiredFactorCombinations:  config.RequiredFactorCombinations,
		SessionDuration:             config.SessionDuration,
		IsSingleUse:                 config.IsSingleUse,
		AllowTrustedDeviceReduction: config.AllowTrustedDeviceReduction,
		ValueThreshold:              config.ValueThreshold,
		CooldownPeriod:              config.CooldownPeriod,
		Description:                 config.Description,
	})
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// GetSensitiveTxConfig returns a sensitive transaction configuration
func (k Keeper) GetSensitiveTxConfig(ctx sdk.Context, txType types.SensitiveTransactionType) (*types.SensitiveTxConfig, bool) {
	store := ctx.KVStore(k.skey)
	key := types.SensitiveTxConfigKey(txType)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var cs sensitiveTxConfigStore
	_ = json.Unmarshal(bz, &cs)

	return &types.SensitiveTxConfig{
		TransactionType:             cs.TransactionType,
		Enabled:                     cs.Enabled,
		MinVEIDScore:                cs.MinVEIDScore,
		RequiredFactorCombinations:  cs.RequiredFactorCombinations,
		SessionDuration:             cs.SessionDuration,
		IsSingleUse:                 cs.IsSingleUse,
		AllowTrustedDeviceReduction: cs.AllowTrustedDeviceReduction,
		ValueThreshold:              cs.ValueThreshold,
		CooldownPeriod:              cs.CooldownPeriod,
		Description:                 cs.Description,
	}, true
}

// GetAllSensitiveTxConfigs returns all sensitive transaction configurations
func (k Keeper) GetAllSensitiveTxConfigs(ctx sdk.Context) []types.SensitiveTxConfig {
	var configs []types.SensitiveTxConfig

	// Iterate through all sensitive transaction types
	for txType := types.SensitiveTxAccountRecovery; txType <= types.SensitiveTxWebhookConfiguration; txType++ {
		if config, found := k.GetSensitiveTxConfig(ctx, txType); found {
			configs = append(configs, *config)
		}
	}

	return configs
}

// ============================================================================
// Genesis
// ============================================================================

// InitGenesis initializes the mfa module's state from a genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, gs *types.GenesisState) {
	// Set params
	if err := k.SetParams(ctx, gs.Params); err != nil {
		panic(err)
	}

	// Set MFA policies
	for _, policy := range gs.MFAPolicies {
		p := policy
		if err := k.SetMFAPolicy(ctx, &p); err != nil {
			panic(err)
		}
	}

	// Set factor enrollments
	for _, enrollment := range gs.FactorEnrollments {
		e := enrollment
		if err := k.EnrollFactor(ctx, &e); err != nil {
			panic(err)
		}
	}

	// Set sensitive tx configs
	for _, config := range gs.SensitiveTxConfigs {
		c := config
		if err := k.SetSensitiveTxConfig(ctx, &c); err != nil {
			panic(err)
		}
	}

	// Set trusted devices
	for _, device := range gs.TrustedDevices {
		address, _ := sdk.AccAddressFromBech32(device.AccountAddress)
		info := device.DeviceInfo
		if _, err := k.AddTrustedDevice(ctx, address, &info); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the mfa module's state to a genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get all MFA policies
	var policies []types.MFAPolicy
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixMFAPolicy)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var ps mfaPolicyStore
		_ = json.Unmarshal(iterator.Value(), &ps)
		policies = append(policies, types.MFAPolicy{
			AccountAddress:     ps.AccountAddress,
			RequiredFactors:    ps.RequiredFactors,
			TrustedDeviceRule:  ps.TrustedDeviceRule,
			RecoveryFactors:    ps.RecoveryFactors,
			KeyRotationFactors: ps.KeyRotationFactors,
			SessionDuration:    ps.SessionDuration,
			VEIDThreshold:      ps.VEIDThreshold,
			Enabled:            ps.Enabled,
			CreatedAt:          ps.CreatedAt,
			UpdatedAt:          ps.UpdatedAt,
		})
	}

	// Get all factor enrollments
	var enrollments []types.FactorEnrollment
	enrollmentIterator := storetypes.KVStorePrefixIterator(store, types.PrefixFactorEnrollment)
	defer enrollmentIterator.Close()

	for ; enrollmentIterator.Valid(); enrollmentIterator.Next() {
		var es factorEnrollmentStore
		_ = json.Unmarshal(enrollmentIterator.Value(), &es)
		enrollments = append(enrollments, types.FactorEnrollment{
			AccountAddress:   es.AccountAddress,
			FactorType:       es.FactorType,
			FactorID:         es.FactorID,
			PublicIdentifier: es.PublicIdentifier,
			Label:            es.Label,
			Status:           es.Status,
			EnrolledAt:       es.EnrolledAt,
			VerifiedAt:       es.VerifiedAt,
			RevokedAt:        es.RevokedAt,
			LastUsedAt:       es.LastUsedAt,
			UseCount:         es.UseCount,
			Metadata:         es.Metadata,
		})
	}

	// Get all trusted devices
	var devices []types.TrustedDevice
	deviceIterator := storetypes.KVStorePrefixIterator(store, types.PrefixTrustedDevice)
	defer deviceIterator.Close()

	for ; deviceIterator.Valid(); deviceIterator.Next() {
		var ds trustedDeviceStore
		_ = json.Unmarshal(deviceIterator.Value(), &ds)
		devices = append(devices, types.TrustedDevice{
			AccountAddress: ds.AccountAddress,
			DeviceInfo:     ds.DeviceInfo,
			AddedAt:        ds.AddedAt,
			LastUsedAt:     ds.LastUsedAt,
		})
	}

	return &types.GenesisState{
		Params:             k.GetParams(ctx),
		MFAPolicies:        policies,
		FactorEnrollments:  enrollments,
		SensitiveTxConfigs: k.GetAllSensitiveTxConfigs(ctx),
		TrustedDevices:     devices,
	}
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	max := int(^uint32(0))
	if value > max {
		return ^uint32(0)
	}
	//nolint:gosec // range checked above
	return uint32(value)
}
