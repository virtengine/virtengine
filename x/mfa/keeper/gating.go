package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.akt.dev/node/x/mfa/types"
)

// MFAGatingHooks provides hooks for transaction gating based on MFA requirements
type MFAGatingHooks struct {
	keeper Keeper
}

// NewMFAGatingHooks creates a new MFA gating hooks instance
func NewMFAGatingHooks(k Keeper) MFAGatingHooks {
	return MFAGatingHooks{keeper: k}
}

// IsSensitiveTransaction checks if a message type URL is considered sensitive
func (h MFAGatingHooks) IsSensitiveTransaction(msgTypeURL string) bool {
	_, isSensitive := types.GetSensitiveTransactionType(msgTypeURL)
	return isSensitive
}

// GetSensitiveTransactionType returns the sensitive transaction type for a message type URL
func (h MFAGatingHooks) GetSensitiveTransactionType(msgTypeURL string) (types.SensitiveTransactionType, bool) {
	return types.GetSensitiveTransactionType(msgTypeURL)
}

// RequiresMFA checks if MFA is required for a given account and transaction type
// Returns the policy and whether MFA is required
func (h MFAGatingHooks) RequiresMFA(
	ctx sdk.Context,
	account sdk.AccAddress,
	txType types.SensitiveTransactionType,
) (*types.MFAPolicy, bool, []types.FactorCombination) {
	// Check if there's a sensitive tx config for this type
	config, found := h.keeper.GetSensitiveTxConfig(ctx, txType)
	if !found || !config.Enabled {
		return nil, false, nil
	}

	// Check if account has MFA policy
	policy, found := h.keeper.GetMFAPolicy(ctx, account)
	if !found {
		// Use global config if no account-specific policy
		return nil, true, config.RequiredFactorCombinations
	}

	if !policy.Enabled {
		// MFA disabled for this account, still use global requirements
		return policy, true, config.RequiredFactorCombinations
	}

	// Get the required factors for this action from the account policy
	requiredFactors := policy.GetRequiredFactorsForAction(txType)
	if len(requiredFactors) == 0 {
		// Fall back to global config
		requiredFactors = config.RequiredFactorCombinations
	}

	return policy, true, requiredFactors
}

// ValidateMFAProof validates an MFA proof for a given account and transaction
func (h MFAGatingHooks) ValidateMFAProof(
	ctx sdk.Context,
	account sdk.AccAddress,
	txType types.SensitiveTransactionType,
	proof *types.MFAProof,
	deviceFingerprint string,
) error {
	if proof == nil {
		return types.ErrMFARequired.Wrap("MFA proof is required for this transaction")
	}

	// Validate proof structure
	if err := proof.Validate(); err != nil {
		return err
	}

	// Get the authorization session
	session, found := h.keeper.GetAuthorizationSession(ctx, proof.SessionID)
	if !found {
		return types.ErrSessionNotFound.Wrapf("session %s not found", proof.SessionID)
	}

	// Verify session belongs to this account
	if session.AccountAddress != account.String() {
		return types.ErrUnauthorized.Wrap("session belongs to different account")
	}

	// Verify session is for the correct transaction type
	if session.TransactionType != txType {
		return types.ErrUnauthorized.Wrapf("session authorized for %s, not %s",
			session.TransactionType.String(), txType.String())
	}

	// Verify session is still valid
	now := ctx.BlockTime()
	if !session.IsValid(now) {
		return types.ErrSessionExpired.Wrap("authorization session has expired or already used")
	}

	// Verify device fingerprint matches if session is bound to a device
	if session.DeviceFingerprint != "" && deviceFingerprint != "" {
		if session.DeviceFingerprint != deviceFingerprint {
			return types.ErrDeviceMismatch.Wrap("request from different device than session")
		}
	}

	// Get the policy to verify factors are sufficient
	policy, found, requiredCombinations := h.RequiresMFA(ctx, account, txType)
	if !found && len(requiredCombinations) == 0 {
		// No MFA required (shouldn't happen if we got here)
		return nil
	}

	// Check if verified factors satisfy any combination
	match := checkFactorCombinations(requiredCombinations, session.VerifiedFactors)
	if !match {
		return types.ErrInsufficientFactors.Wrap("verified factors do not satisfy policy requirements")
	}

	// Check VEID threshold if required
	if hasVEIDFactor(requiredCombinations) {
		threshold := uint32(50)
		if policy != nil && policy.VEIDThreshold > 0 {
			threshold = policy.VEIDThreshold
		}

		if h.keeper.veidKeeper != nil {
			score, found := h.keeper.veidKeeper.GetVEIDScore(ctx, account)
			if !found || score < threshold {
				return types.ErrVEIDScoreInsufficient.Wrapf("VEID score %d below threshold %d", score, threshold)
			}
		}
	}

	// If single-use session, mark it as used
	if session.IsSingleUse {
		if err := h.keeper.UseAuthorizationSession(ctx, proof.SessionID); err != nil {
			return err
		}
	}

	// Emit event for successful MFA validation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSessionUsed,
			sdk.NewAttribute(types.AttributeKeySessionID, session.SessionID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, account.String()),
			sdk.NewAttribute(types.AttributeKeyTransactionType, txType.String()),
		),
	)

	return nil
}

// CanBypassMFA checks if MFA can be bypassed for a trusted device
func (h MFAGatingHooks) CanBypassMFA(
	ctx sdk.Context,
	account sdk.AccAddress,
	txType types.SensitiveTransactionType,
	deviceFingerprint string,
) (bool, *types.FactorCombination) {
	// Check if there's a sensitive tx config that allows trusted device bypass
	config, found := h.keeper.GetSensitiveTxConfig(ctx, txType)
	if !found || !config.AllowTrustedDeviceReduction {
		return false, nil
	}

	// Check if device is trusted
	if deviceFingerprint == "" || !h.keeper.IsTrustedDevice(ctx, account, deviceFingerprint) {
		return false, nil
	}

	// Get account policy to check trusted device rule
	policy, found := h.keeper.GetMFAPolicy(ctx, account)
	if !found || !policy.Enabled {
		return false, nil
	}

	// Check if trusted device can reduce factors for this action
	if !policy.CanUseTrustedDevice(txType, true) {
		return false, nil
	}

	// Return the reduced factors (if any)
	reducedFactors := policy.GetReducedFactors()
	if reducedFactors == nil {
		// Complete bypass allowed
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMFABypassed,
				sdk.NewAttribute(types.AttributeKeyAccountAddress, account.String()),
				sdk.NewAttribute(types.AttributeKeyTransactionType, txType.String()),
				sdk.NewAttribute(types.AttributeKeyDeviceFingerprint, deviceFingerprint),
			),
		)
		return true, nil
	}

	return false, reducedFactors
}

// CheckMFARequired is a convenience method that combines RequiresMFA and CanBypassMFA
// Returns: (mfaRequired bool, bypassAllowed bool, requiredFactors []FactorCombination)
func (h MFAGatingHooks) CheckMFARequired(
	ctx sdk.Context,
	account sdk.AccAddress,
	msgTypeURL string,
	deviceFingerprint string,
) (bool, bool, []types.FactorCombination) {
	// First check if this is a sensitive transaction
	txType, isSensitive := h.GetSensitiveTransactionType(msgTypeURL)
	if !isSensitive {
		return false, false, nil
	}

	// Check if MFA is required
	_, required, requiredCombinations := h.RequiresMFA(ctx, account, txType)
	if !required {
		return false, false, nil
	}

	// Check if bypass is allowed
	canBypass, reducedFactors := h.CanBypassMFA(ctx, account, txType, deviceFingerprint)
	if canBypass && reducedFactors == nil {
		return true, true, nil
	}

	// If there are reduced factors, use those instead
	if reducedFactors != nil {
		return true, false, []types.FactorCombination{*reducedFactors}
	}

	return true, false, requiredCombinations
}

// GetAccountMFAStatus returns a summary of MFA status for an account
func (h MFAGatingHooks) GetAccountMFAStatus(ctx sdk.Context, account sdk.AccAddress) AccountMFAStatus {
	status := AccountMFAStatus{
		Address:         account.String(),
		MFAEnabled:      false,
		FactorCount:     0,
		ActiveFactors:   []types.FactorType{},
		TrustedDevices:  0,
		PendingChallenges: 0,
	}

	// Get policy
	policy, found := h.keeper.GetMFAPolicy(ctx, account)
	if found {
		status.MFAEnabled = policy.Enabled
		status.Policy = policy
	}

	// Get enrollments
	enrollments := h.keeper.GetFactorEnrollments(ctx, account)
	now := ctx.BlockTime()
	seenTypes := make(map[types.FactorType]bool)
	
	for _, e := range enrollments {
		if e.CanVerify(now) {
			status.FactorCount++
			if !seenTypes[e.FactorType] {
				status.ActiveFactors = append(status.ActiveFactors, e.FactorType)
				seenTypes[e.FactorType] = true
			}
		}
	}

	// Get trusted devices
	devices := h.keeper.GetTrustedDevices(ctx, account)
	nowUnix := now.Unix()
	for _, d := range devices {
		if d.DeviceInfo.TrustExpiresAt > nowUnix {
			status.TrustedDevices++
		}
	}

	// Get pending challenges
	challenges := h.keeper.GetPendingChallenges(ctx, account)
	status.PendingChallenges = len(challenges)

	return status
}

// AccountMFAStatus represents the MFA status for an account
type AccountMFAStatus struct {
	Address           string
	MFAEnabled        bool
	Policy            *types.MFAPolicy
	FactorCount       int
	ActiveFactors     []types.FactorType
	TrustedDevices    int
	PendingChallenges int
}

// Helper functions

// checkFactorCombinations checks if verified factors satisfy any combination
func checkFactorCombinations(combinations []types.FactorCombination, verified []types.FactorType) bool {
	if len(combinations) == 0 {
		return true
	}

	verifiedSet := make(map[types.FactorType]bool)
	for _, ft := range verified {
		verifiedSet[ft] = true
	}

	// Check each combination (OR logic)
	for _, combo := range combinations {
		allPresent := true
		for _, requiredFactor := range combo.Factors {
			if !verifiedSet[requiredFactor] {
				allPresent = false
				break
			}
		}
		if allPresent {
			return true
		}
	}

	return false
}

// hasVEIDFactor checks if any combination requires VEID factor
func hasVEIDFactor(combinations []types.FactorCombination) bool {
	for _, combo := range combinations {
		for _, ft := range combo.Factors {
			if ft == types.FactorTypeVEID {
				return true
			}
		}
	}
	return false
}

// CleanupExpiredData cleans up expired challenges and sessions
// Should be called in EndBlock or periodically
func (h MFAGatingHooks) CleanupExpiredData(ctx sdk.Context) {
	now := ctx.BlockTime()
	h.cleanupExpiredChallenges(ctx, now)
	h.cleanupExpiredSessions(ctx, now)
}

// cleanupExpiredChallenges removes expired challenges
func (h MFAGatingHooks) cleanupExpiredChallenges(ctx sdk.Context, now time.Time) {
	// Implementation would iterate through challenges and delete expired ones
	// For efficiency, this could be done with a secondary index by expiration time
}

// cleanupExpiredSessions removes expired sessions
func (h MFAGatingHooks) cleanupExpiredSessions(ctx sdk.Context, now time.Time) {
	// Implementation would iterate through sessions and delete expired ones
}
