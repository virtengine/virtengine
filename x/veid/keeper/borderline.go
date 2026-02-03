package keeper

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// MFA Keeper Interface for Borderline Fallback
// ============================================================================

// MFAKeeper defines the interface for the MFA keeper that veid needs
type MFAKeeper interface {
	// CreateChallenge creates a new MFA challenge
	CreateChallenge(ctx sdk.Context, challenge *mfatypes.Challenge) error

	// GetChallenge retrieves a challenge by ID
	GetChallenge(ctx sdk.Context, challengeID string) (*mfatypes.Challenge, bool)

	// GetFactorEnrollments returns all factor enrollments for an account
	GetFactorEnrollments(ctx sdk.Context, address sdk.AccAddress) []mfatypes.FactorEnrollment

	// HasActiveFactorOfType checks if account has an active factor of the given type
	HasActiveFactorOfType(ctx sdk.Context, address sdk.AccAddress, factorType mfatypes.FactorType) bool

	// VerifyMFAChallenge verifies an MFA challenge response
	VerifyMFAChallenge(ctx sdk.Context, challengeID string, response *mfatypes.ChallengeResponse) (bool, error)

	// GetParams returns MFA module parameters
	GetParams(ctx sdk.Context) mfatypes.Params
}

// SetMFAKeeper sets the MFA keeper reference for borderline fallback operations
func (k *Keeper) SetMFAKeeper(mfaKeeper MFAKeeper) {
	k.mfaKeeper = mfaKeeper
}

// ============================================================================
// Borderline Parameters Management
// ============================================================================

// borderlineParamsStore is the stored format of borderline params (matches proto structure)
type borderlineParamsStore struct {
	LowerThreshold   uint32 `json:"lower_threshold"`
	UpperThreshold   uint32 `json:"upper_threshold"`
	MfaTimeoutBlocks int64  `json:"mfa_timeout_blocks"`
	RequiredFactors  uint32 `json:"required_factors"`
}

// SetBorderlineParams sets the borderline parameters
func (k Keeper) SetBorderlineParams(ctx sdk.Context, params types.BorderlineParams) error {
	if err := types.ValidateBorderlineParams(params); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&borderlineParamsStore{
		LowerThreshold:   params.LowerThreshold,
		UpperThreshold:   params.UpperThreshold,
		MfaTimeoutBlocks: params.MfaTimeoutBlocks,
		RequiredFactors:  params.RequiredFactors,
	})
	if err != nil {
		return err
	}

	store.Set(types.BorderlineParamsKey(), bz)
	return nil
}

// GetBorderlineParams returns the borderline parameters
func (k Keeper) GetBorderlineParams(ctx sdk.Context) types.BorderlineParams {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.BorderlineParamsKey())
	if bz == nil {
		return types.DefaultBorderlineParams()
	}

	var ps borderlineParamsStore
	if err := json.Unmarshal(bz, &ps); err != nil {
		return types.DefaultBorderlineParams()
	}

	return types.BorderlineParams{
		LowerThreshold:   ps.LowerThreshold,
		UpperThreshold:   ps.UpperThreshold,
		MfaTimeoutBlocks: ps.MfaTimeoutBlocks,
		RequiredFactors:  ps.RequiredFactors,
	}
}

// ============================================================================
// Borderline Detection and Fallback Triggering
// ============================================================================

// CheckBorderlineAndTriggerFallback evaluates a facial verification score and
// determines the appropriate action based on borderline thresholds.
// Returns the resulting verification status and any error.
func (k Keeper) CheckBorderlineAndTriggerFallback(
	ctx sdk.Context,
	accountAddr string,
	score uint32,
) (types.VerificationStatus, error) {
	params := k.GetBorderlineParams(ctx)

	// If score is at or above upper threshold, verified directly
	if types.IsScoreAboveUpperThreshold(params, score) {
		k.Logger(ctx).Info("score above upper threshold, verified directly",
			"account", accountAddr,
			"score", score,
			"upper_threshold", params.UpperThreshold,
		)
		return types.VerificationStatusVerified, nil
	}

	// If score is in borderline band, trigger MFA fallback
	if types.IsScoreInBorderlineBand(params, score) {
		k.Logger(ctx).Info("score in borderline band, triggering fallback",
			"account", accountAddr,
			"score", score,
			"lower", params.LowerThreshold,
			"upper", params.UpperThreshold,
		)
		return k.TriggerBorderlineFallback(ctx, accountAddr, score)
	}

	// Score is below lower threshold, reject
	k.Logger(ctx).Info("score below lower threshold, rejected",
		"account", accountAddr,
		"score", score,
		"lower_threshold", params.LowerThreshold,
	)
	return types.VerificationStatusRejected, nil
}

// TriggerBorderlineFallback initiates the MFA fallback process for a borderline score.
// It creates an MFA challenge and records the borderline event.
func (k Keeper) TriggerBorderlineFallback(
	ctx sdk.Context,
	accountAddr string,
	borderlineScore uint32,
) (types.VerificationStatus, error) {
	params := k.GetBorderlineParams(ctx)

	// Parse account address
	address, err := sdk.AccAddressFromBech32(accountAddr)
	if err != nil {
		return types.VerificationStatusRejected, types.ErrInvalidAddress.Wrap(err.Error())
	}

	// Check if account has any enrolled MFA factors
	if k.mfaKeeper == nil {
		return types.VerificationStatusRejected, types.ErrNoEnrolledFactors.Wrap("MFA keeper not configured")
	}

	// Get available factors based on required factor count
	availableFactors := k.getAvailableFactorsForFallbackCount(ctx, address, params.RequiredFactors)
	if len(availableFactors) == 0 {
		k.Logger(ctx).Warn("no enrolled factors for borderline fallback",
			"account", accountAddr,
			"required_factors", params.RequiredFactors,
		)
		return types.VerificationStatusRejected, types.ErrNoEnrolledFactors
	}

	// Generate fallback ID
	fallbackID, err := generateFallbackID()
	if err != nil {
		return types.VerificationStatusRejected, err
	}

	// Create MFA challenge for borderline verification
	challengeID, err := k.createBorderlineMFAChallenge(ctx, address, availableFactors, params)
	if err != nil {
		k.Logger(ctx).Error("failed to create MFA challenge for borderline",
			"account", accountAddr,
			"error", err,
		)
		return types.VerificationStatusRejected, err
	}

	// Create and store fallback record
	now := ctx.BlockTime().Unix()
	// Convert MFA timeout blocks to seconds (approximately 6 seconds per block)
	challengeTimeoutSeconds := params.MfaTimeoutBlocks * 6
	expiresAt := now + challengeTimeoutSeconds

	// Convert RequiredFactors count to factor type list
	requiredFactorsList := k.getDefaultRequiredFactorsList()

	fallbackRecord := types.NewBorderlineFallbackRecord(
		fallbackID,
		accountAddr,
		borderlineScore,
		challengeID,
		requiredFactorsList,
		now,
		expiresAt,
		ctx.BlockHeight(),
	)

	if err := k.setBorderlineFallbackRecord(ctx, fallbackRecord); err != nil {
		return types.VerificationStatusRejected, err
	}

	// Add to pending queue for expiry tracking
	k.addToPendingFallbackQueue(ctx, fallbackRecord)

	// Emit borderline fallback triggered event
	k.emitBorderlineFallbackTriggeredEvent(ctx, fallbackRecord, requiredFactorsList)

	k.Logger(ctx).Info("borderline fallback triggered",
		"account", accountAddr,
		"fallback_id", fallbackID,
		"challenge_id", challengeID,
		"score", borderlineScore,
		"expires_at", expiresAt,
	)

	return types.VerificationStatusNeedsAdditionalFactor, nil
}

// getDefaultRequiredFactorsList returns the default list of acceptable MFA factor types
func (k Keeper) getDefaultRequiredFactorsList() []string {
	return []string{"totp", "fido2", "email", "sms"}
}

// getAvailableFactorsForFallbackCount returns factor types based on required count
func (k Keeper) getAvailableFactorsForFallbackCount(
	ctx sdk.Context,
	address sdk.AccAddress,
	_ uint32,
) []mfatypes.FactorType {
	var available []mfatypes.FactorType

	// Check all default factor types
	for _, factorName := range k.getDefaultRequiredFactorsList() {
		factorType, err := mfatypes.FactorTypeFromString(factorName)
		if err != nil {
			continue
		}

		if k.mfaKeeper.HasActiveFactorOfType(ctx, address, factorType) {
			available = append(available, factorType)
		}
	}

	return available
}

// getAvailableFactorsForFallback returns the factor types that are both required
// by borderline params and enrolled by the account
//
//nolint:unused // reserved for configurable factor selection
func (k Keeper) getAvailableFactorsForFallback(
	ctx sdk.Context,
	address sdk.AccAddress,
	requiredFactors []string,
) []mfatypes.FactorType {
	var available []mfatypes.FactorType

	for _, factorName := range requiredFactors {
		factorType, err := mfatypes.FactorTypeFromString(factorName)
		if err != nil {
			continue
		}

		if k.mfaKeeper.HasActiveFactorOfType(ctx, address, factorType) {
			available = append(available, factorType)
		}
	}

	return available
}

// createBorderlineMFAChallenge creates an MFA challenge for borderline verification
func (k Keeper) createBorderlineMFAChallenge(
	ctx sdk.Context,
	address sdk.AccAddress,
	availableFactors []mfatypes.FactorType,
	params types.BorderlineParams,
) (string, error) {
	if len(availableFactors) == 0 {
		return "", types.ErrNoEnrolledFactors
	}

	// Use the first available factor type for the challenge
	// In a more sophisticated implementation, this could be user-selectable
	factorType := availableFactors[0]

	// Get an active enrollment of this type
	enrollments := k.mfaKeeper.GetFactorEnrollments(ctx, address)
	var factorID string
	for _, e := range enrollments {
		if e.FactorType == factorType && e.IsActive() {
			factorID = e.FactorID
			break
		}
	}

	if factorID == "" {
		return "", types.ErrNoEnrolledFactors.Wrapf("no active %s factor found", factorType.String())
	}

	// Create the MFA challenge
	mfaParams := k.mfaKeeper.GetParams(ctx)
	// Convert MFA timeout blocks to seconds
	challengeTimeoutSeconds := params.MfaTimeoutBlocks * 6
	challenge, err := mfatypes.NewChallenge(
		address.String(),
		factorType,
		factorID,
		mfatypes.SensitiveTxUnspecified, // Borderline verification isn't a standard sensitive tx
		challengeTimeoutSeconds,
		mfaParams.MaxChallengeAttempts,
	)
	if err != nil {
		return "", err
	}

	// Store the challenge via MFA keeper
	if err := k.mfaKeeper.CreateChallenge(ctx, challenge); err != nil {
		return "", err
	}

	return challenge.ChallengeID, nil
}

// generateFallbackID generates a unique fallback ID
func generateFallbackID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate fallback ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ============================================================================
// Borderline Fallback Record Storage
// ============================================================================

// borderlineFallbackStore is the stored format of a borderline fallback record
type borderlineFallbackStore struct {
	FallbackID              string   `json:"fallback_id"`
	AccountAddress          string   `json:"account_address"`
	BorderlineScore         uint32   `json:"borderline_score"`
	ChallengeID             string   `json:"challenge_id"`
	Status                  string   `json:"status"`
	RequiredFactors         []string `json:"required_factors"`
	SatisfiedFactors        []string `json:"satisfied_factors,omitempty"`
	CreatedAt               int64    `json:"created_at"`
	ExpiresAt               int64    `json:"expires_at"`
	CompletedAt             int64    `json:"completed_at,omitempty"`
	BlockHeight             int64    `json:"block_height"`
	FinalVerificationStatus string   `json:"final_verification_status,omitempty"`
}

// setBorderlineFallbackRecord stores a borderline fallback record
func (k Keeper) setBorderlineFallbackRecord(ctx sdk.Context, record *types.BorderlineFallbackRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(&borderlineFallbackStore{
		FallbackID:              record.FallbackID,
		AccountAddress:          record.AccountAddress,
		BorderlineScore:         record.BorderlineScore,
		ChallengeID:             record.ChallengeID,
		Status:                  string(record.Status),
		RequiredFactors:         record.RequiredFactors,
		SatisfiedFactors:        record.SatisfiedFactors,
		CreatedAt:               record.CreatedAt,
		ExpiresAt:               record.ExpiresAt,
		CompletedAt:             record.CompletedAt,
		BlockHeight:             record.BlockHeight,
		FinalVerificationStatus: string(record.FinalVerificationStatus),
	})
	if err != nil {
		return err
	}

	store.Set(types.BorderlineFallbackKey(record.FallbackID), bz)

	// Update account index
	k.updateAccountFallbackIndex(ctx, record.AccountAddress, record.FallbackID)

	return nil
}

// GetBorderlineFallbackRecord retrieves a borderline fallback record by ID
func (k Keeper) GetBorderlineFallbackRecord(ctx sdk.Context, fallbackID string) (*types.BorderlineFallbackRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.BorderlineFallbackKey(fallbackID))
	if bz == nil {
		return nil, false
	}

	var fs borderlineFallbackStore
	if err := json.Unmarshal(bz, &fs); err != nil {
		k.Logger(ctx).Error("failed to unmarshal borderline fallback record", "error", err)
		return nil, false
	}

	return &types.BorderlineFallbackRecord{
		FallbackID:              fs.FallbackID,
		AccountAddress:          fs.AccountAddress,
		BorderlineScore:         fs.BorderlineScore,
		ChallengeID:             fs.ChallengeID,
		Status:                  types.BorderlineFallbackStatus(fs.Status),
		RequiredFactors:         fs.RequiredFactors,
		SatisfiedFactors:        fs.SatisfiedFactors,
		CreatedAt:               fs.CreatedAt,
		ExpiresAt:               fs.ExpiresAt,
		CompletedAt:             fs.CompletedAt,
		BlockHeight:             fs.BlockHeight,
		FinalVerificationStatus: types.VerificationStatus(fs.FinalVerificationStatus),
	}, true
}

// GetBorderlineFallbackByChallenge finds a fallback record by challenge ID
func (k Keeper) GetBorderlineFallbackByChallenge(ctx sdk.Context, challengeID string) (*types.BorderlineFallbackRecord, bool) {
	store := ctx.KVStore(k.skey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PrefixBorderlineFallback)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var fs borderlineFallbackStore
		if err := json.Unmarshal(iterator.Value(), &fs); err != nil {
			continue
		}

		if fs.ChallengeID == challengeID {
			return &types.BorderlineFallbackRecord{
				FallbackID:              fs.FallbackID,
				AccountAddress:          fs.AccountAddress,
				BorderlineScore:         fs.BorderlineScore,
				ChallengeID:             fs.ChallengeID,
				Status:                  types.BorderlineFallbackStatus(fs.Status),
				RequiredFactors:         fs.RequiredFactors,
				SatisfiedFactors:        fs.SatisfiedFactors,
				CreatedAt:               fs.CreatedAt,
				ExpiresAt:               fs.ExpiresAt,
				CompletedAt:             fs.CompletedAt,
				BlockHeight:             fs.BlockHeight,
				FinalVerificationStatus: types.VerificationStatus(fs.FinalVerificationStatus),
			}, true
		}
	}

	return nil, false
}

// GetPendingFallbacksForAccount returns pending fallbacks for an account
func (k Keeper) GetPendingFallbacksForAccount(ctx sdk.Context, accountAddress string) []types.BorderlineFallbackRecord {
	address, err := sdk.AccAddressFromBech32(accountAddress)
	if err != nil {
		return nil
	}

	store := ctx.KVStore(k.skey)
	bz := store.Get(types.BorderlineFallbackByAccountKey(address.Bytes()))
	if bz == nil {
		return nil
	}

	fallbackIDs := make([]string, 0, 1)
	if err := json.Unmarshal(bz, &fallbackIDs); err != nil {
		return nil
	}

	var pending []types.BorderlineFallbackRecord
	for _, id := range fallbackIDs {
		record, found := k.GetBorderlineFallbackRecord(ctx, id)
		if found && record.IsPending() {
			pending = append(pending, *record)
		}
	}

	return pending
}

// updateAccountFallbackIndex updates the index of fallbacks by account
func (k Keeper) updateAccountFallbackIndex(ctx sdk.Context, accountAddress string, fallbackID string) {
	address, err := sdk.AccAddressFromBech32(accountAddress)
	if err != nil {
		return
	}

	store := ctx.KVStore(k.skey)
	key := types.BorderlineFallbackByAccountKey(address.Bytes())

	fallbackIDs := make([]string, 0, 1)
	if bz := store.Get(key); bz != nil {
		_ = json.Unmarshal(bz, &fallbackIDs)
	}

	// Add if not already present
	for _, id := range fallbackIDs {
		if id == fallbackID {
			return
		}
	}

	fallbackIDs = append(fallbackIDs, fallbackID)
	bz, _ := json.Marshal(fallbackIDs) //nolint:errchkjson // string slice cannot fail to marshal
	store.Set(key, bz)
}

// addToPendingFallbackQueue adds a fallback to the pending queue for expiry tracking
func (k Keeper) addToPendingFallbackQueue(ctx sdk.Context, record *types.BorderlineFallbackRecord) {
	store := ctx.KVStore(k.skey)
	key := types.PendingBorderlineFallbackKey(record.ExpiresAt, record.FallbackID)
	store.Set(key, []byte{1})
}

// removeFromPendingFallbackQueue removes a fallback from the pending queue
func (k Keeper) removeFromPendingFallbackQueue(ctx sdk.Context, record *types.BorderlineFallbackRecord) {
	store := ctx.KVStore(k.skey)
	key := types.PendingBorderlineFallbackKey(record.ExpiresAt, record.FallbackID)
	store.Delete(key)
}

// ============================================================================
// Event Emission
// ============================================================================

// emitBorderlineFallbackTriggeredEvent emits an event when borderline fallback is triggered
func (k Keeper) emitBorderlineFallbackTriggeredEvent(
	ctx sdk.Context,
	record *types.BorderlineFallbackRecord,
	requiredFactors []string,
) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorderlineFallbackTriggered,
			sdk.NewAttribute(types.AttributeKeyAccountAddress, record.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFallbackID, record.FallbackID),
			sdk.NewAttribute(types.AttributeKeyBorderlineScore, fmt.Sprintf("%d", record.BorderlineScore)),
			sdk.NewAttribute(types.AttributeKeyChallengeID, record.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyRequiredFactors, strings.Join(requiredFactors, ",")),
			sdk.NewAttribute(types.AttributeKeyExpiresAt, fmt.Sprintf("%d", record.ExpiresAt)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", record.BlockHeight)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", record.CreatedAt)),
		),
	)
}
