package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// ============================================================================
// Authorization Session Management
//
// Implements session duration per action type as defined in MFA-CORE-002:
// - Critical (AccountRecovery, KeyRotation, AccountDeletion, TwoFactorDisable): Single use
// - High (ProviderReg, LargeWithdrawal, ValidatorReg, RoleAssignment): 15 minutes
// - Medium (HighValueOrder, GovernanceProposal, GovernanceVote): 30 minutes
// - Low (MediumWithdrawal, TransferToNewAddress, APIKeyGeneration): 60 minutes
// ============================================================================

// HasValidAuthSession checks if an account has a valid authorization session for the given action.
// Returns true if:
// - A valid (non-expired) session exists for the account and action
// - For single-use sessions, it has not been used yet
// - Device fingerprint matches if session is device-bound
func (k Keeper) HasValidAuthSession(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType) bool {
	return k.hasValidAuthSessionWithDevice(ctx, address, action, "")
}

// HasValidAuthSessionWithDevice checks if an account has a valid authorization session
// for the given action, optionally validating the device fingerprint.
func (k Keeper) HasValidAuthSessionWithDevice(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) bool {
	return k.hasValidAuthSessionWithDevice(ctx, address, action, deviceFingerprint)
}

// hasValidAuthSessionWithDevice is the internal implementation for session validation.
func (k Keeper) hasValidAuthSessionWithDevice(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) bool {
	sessions := k.GetAccountSessions(ctx, address)
	now := ctx.BlockTime()

	for _, session := range sessions {
		// Check if session is for the correct action type
		if session.TransactionType != action {
			continue
		}

		// Check if session is valid (not expired, not used if single-use)
		if !session.IsValid(now) {
			continue
		}

		// If device fingerprint validation is requested and session is device-bound
		if deviceFingerprint != "" && session.DeviceFingerprint != "" {
			if session.DeviceFingerprint != deviceFingerprint {
				continue
			}
		}

		return true
	}

	return false
}

// ConsumeAuthSession consumes a single-use authorization session for the given action.
// For single-use sessions (Critical tier), this marks the session as used.
// For multi-use sessions, this is a no-op (session remains valid until expiry).
// Returns an error if no valid session exists.
func (k Keeper) ConsumeAuthSession(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType) error {
	return k.consumeAuthSessionWithDevice(ctx, address, action, "")
}

// ConsumeAuthSessionWithDevice consumes a session with device validation.
func (k Keeper) ConsumeAuthSessionWithDevice(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) error {
	return k.consumeAuthSessionWithDevice(ctx, address, action, deviceFingerprint)
}

// consumeAuthSessionWithDevice is the internal implementation for session consumption.
func (k Keeper) consumeAuthSessionWithDevice(ctx sdk.Context, address sdk.AccAddress, action types.SensitiveTransactionType, deviceFingerprint string) error {
	sessions := k.GetAccountSessions(ctx, address)
	now := ctx.BlockTime()

	for _, session := range sessions {
		// Check if session is for the correct action type
		if session.TransactionType != action {
			continue
		}

		// Check if session is valid
		if !session.IsValid(now) {
			continue
		}

		// Validate device fingerprint if provided
		if deviceFingerprint != "" && session.DeviceFingerprint != "" {
			if session.DeviceFingerprint != deviceFingerprint {
				return types.ErrDeviceMismatch.Wrap("device fingerprint does not match session")
			}
		}

		// For single-use sessions, mark as used
		if session.IsSingleUse {
			return k.UseAuthorizationSession(ctx, session.SessionID)
		}

		// For multi-use sessions, just emit event (session remains valid)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSessionUsed,
				sdk.NewAttribute(types.AttributeKeySessionID, session.SessionID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
				sdk.NewAttribute(types.AttributeKeyTransactionType, action.String()),
			),
		)

		return nil
	}

	return types.ErrSessionNotFound.Wrapf("no valid authorization session found for action %s", action.String())
}

// CreateAuthSessionForAction creates an authorization session for the given action type
// using the default duration and single-use settings based on the action's risk level.
func (k Keeper) CreateAuthSessionForAction(
	ctx sdk.Context,
	address sdk.AccAddress,
	action types.SensitiveTransactionType,
	verifiedFactors []types.FactorType,
	deviceFingerprint string,
) (*types.AuthorizationSession, error) {
	now := ctx.BlockTime().Unix()

	// Get session duration from sensitive tx config if available, otherwise use defaults
	duration := action.GetDefaultSessionDuration()
	isSingleUse := action.IsSingleUse()

	// Check if there's a custom config for this action
	if config, found := k.GetSensitiveTxConfig(ctx, action); found {
		duration = config.SessionDuration
		isSingleUse = config.IsSingleUse
	}

	// Calculate expiry time
	var expiresAt int64
	if isSingleUse {
		// Single-use sessions get a short window (5 minutes) to complete the transaction
		expiresAt = now + 5*60
	} else if duration > 0 {
		expiresAt = now + duration
	} else {
		// Default to 15 minutes if no duration configured
		expiresAt = now + 15*60
	}

	session := &types.AuthorizationSession{
		AccountAddress:    address.String(),
		TransactionType:   action,
		VerifiedFactors:   verifiedFactors,
		CreatedAt:         now,
		ExpiresAt:         expiresAt,
		IsSingleUse:       isSingleUse,
		DeviceFingerprint: deviceFingerprint,
	}

	if err := k.CreateAuthorizationSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetValidSessionsForAccount returns all valid (non-expired, non-consumed) sessions for an account.
func (k Keeper) GetValidSessionsForAccount(ctx sdk.Context, address sdk.AccAddress) []types.AuthorizationSession {
	sessions := k.GetAccountSessions(ctx, address)
	now := ctx.BlockTime()

	var validSessions []types.AuthorizationSession
	for _, session := range sessions {
		if session.IsValid(now) {
			validSessions = append(validSessions, session)
		}
	}

	return validSessions
}

// CleanupExpiredSessions removes expired sessions for an account.
// This is called during EndBlock or can be triggered manually.
func (k Keeper) CleanupExpiredSessions(ctx sdk.Context, address sdk.AccAddress) int {
	sessions := k.GetAccountSessions(ctx, address)
	now := ctx.BlockTime()
	deleted := 0

	for _, session := range sessions {
		if !session.IsValid(now) {
			// Emit expiry event for sessions that expired (not just used single-use sessions)
			if now.Unix() > session.ExpiresAt && session.UsedAt == 0 {
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeSessionExpired,
						sdk.NewAttribute(types.AttributeKeySessionID, session.SessionID),
						sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
					),
				)
			}

			_ = k.DeleteAuthorizationSession(ctx, session.SessionID)
			deleted++
		}
	}

	return deleted
}

// ValidateSessionForTransaction validates that a session is appropriate for the given transaction.
// This performs comprehensive checks including:
// - Session exists and is valid
// - Session is for the correct action type
// - Device fingerprint matches (if session is device-bound)
// - Session has not been consumed (for single-use)
func (k Keeper) ValidateSessionForTransaction(
	ctx sdk.Context,
	sessionID string,
	address sdk.AccAddress,
	action types.SensitiveTransactionType,
	deviceFingerprint string,
) error {
	session, found := k.GetAuthorizationSession(ctx, sessionID)
	if !found {
		return types.ErrSessionNotFound.Wrapf("session %s not found", sessionID)
	}

	// Verify the session belongs to this account
	if session.AccountAddress != address.String() {
		return types.ErrUnauthorized.Wrap("session does not belong to this account")
	}

	// Verify the session is for this action type
	if session.TransactionType != action {
		return types.ErrUnauthorized.Wrapf("session is for action %s, not %s",
			session.TransactionType.String(), action.String())
	}

	// Check if session is valid
	now := ctx.BlockTime()
	if !session.IsValid(now) {
		if session.IsSingleUse && session.UsedAt > 0 {
			return types.ErrSessionAlreadyUsed.Wrap("single-use session has already been consumed")
		}
		return types.ErrSessionExpired.Wrap("session has expired")
	}

	// Validate device fingerprint if session is device-bound
	if session.DeviceFingerprint != "" && deviceFingerprint != "" {
		if session.DeviceFingerprint != deviceFingerprint {
			return types.ErrDeviceMismatch.Wrap("device fingerprint does not match session")
		}
	}

	return nil
}

// GetSessionDurationForAction returns the session duration in seconds for a given action type.
func (k Keeper) GetSessionDurationForAction(ctx sdk.Context, action types.SensitiveTransactionType) int64 {
	// Check for custom config first
	if config, found := k.GetSensitiveTxConfig(ctx, action); found {
		return config.SessionDuration
	}

	// Fall back to defaults based on risk level
	return action.GetDefaultSessionDuration()
}

// IsActionSingleUse returns whether an action type requires single-use authorization.
func (k Keeper) IsActionSingleUse(ctx sdk.Context, action types.SensitiveTransactionType) bool {
	// Check for custom config first
	if config, found := k.GetSensitiveTxConfig(ctx, action); found {
		return config.IsSingleUse
	}

	// Fall back to type defaults
	return action.IsSingleUse()
}
