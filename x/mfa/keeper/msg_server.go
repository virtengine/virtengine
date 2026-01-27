package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// msgServer implements the MsgServer interface for the MFA module
type msgServer struct {
	Keeper
}

// NewMsgServerWithContext returns an implementation of the MsgServer interface
func NewMsgServerWithContext(k Keeper) types.MsgServer {
	return &msgServer{Keeper: k}
}

// EnrollFactor enrolls a new MFA factor for the sender's account
func (m *msgServer) EnrollFactor(goCtx context.Context, msg *types.MsgEnrollFactor) (*types.MsgEnrollFactorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	// Check if account is operational
	if m.rolesKeeper != nil && !m.rolesKeeper.IsAccountOperational(ctx, address) {
		return nil, types.ErrUnauthorized.Wrap("account is not operational")
	}

	params := m.GetParams(ctx)
	if !params.IsFactorTypeAllowed(msg.FactorType) {
		return nil, types.ErrInvalidFactorType.Wrapf("factor type %s is not allowed", msg.FactorType.String())
	}

	now := ctx.BlockTime().Unix()

	// Generate factor ID based on factor type
	factorID := generateFactorID(msg.FactorType, msg.PublicIdentifier, msg.Label)

	enrollment := &types.FactorEnrollment{
		AccountAddress:   msg.Sender,
		FactorType:       msg.FactorType,
		FactorID:         factorID,
		PublicIdentifier: msg.PublicIdentifier,
		Label:            msg.Label,
		Status:           types.EnrollmentStatusActive, // Assume verified for now
		EnrolledAt:       now,
		VerifiedAt:       now,
		UseCount:         0,
		Metadata:         msg.Metadata,
	}

	if err := m.Keeper.EnrollFactor(ctx, enrollment); err != nil {
		return nil, err
	}

	return &types.MsgEnrollFactorResponse{
		FactorID: factorID,
		Status:   types.EnrollmentStatusActive,
	}, nil
}

// RevokeFactor revokes an enrolled MFA factor
func (m *msgServer) RevokeFactor(goCtx context.Context, msg *types.MsgRevokeFactor) (*types.MsgRevokeFactorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	// Check if MFA is enabled and proof is required
	policy, found := m.GetMFAPolicy(ctx, address)
	if found && policy.Enabled {
		if msg.MFAProof == nil {
			return nil, types.ErrMFARequired.Wrap("MFA proof required to revoke factor")
		}

		// Validate the MFA proof
		if err := m.validateMFAProof(ctx, address, msg.MFAProof); err != nil {
			return nil, err
		}
	}

	if err := m.Keeper.RevokeFactor(ctx, address, msg.FactorType, msg.FactorID); err != nil {
		return nil, err
	}

	return &types.MsgRevokeFactorResponse{
		Success: true,
	}, nil
}

// SetMFAPolicy sets the MFA policy for the sender's account
func (m *msgServer) SetMFAPolicy(goCtx context.Context, msg *types.MsgSetMFAPolicy) (*types.MsgSetMFAPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	// Check if MFA is already enabled and requires proof to modify
	existingPolicy, found := m.GetMFAPolicy(ctx, address)
	if found && existingPolicy.Enabled {
		if msg.MFAProof == nil {
			return nil, types.ErrMFARequired.Wrap("MFA proof required to modify policy")
		}

		if err := m.validateMFAProof(ctx, address, msg.MFAProof); err != nil {
			return nil, err
		}
	}

	// Ensure the policy is for the sender's account
	if msg.Policy.AccountAddress != msg.Sender {
		return nil, types.ErrUnauthorized.Wrap("can only set policy for own account")
	}

	// If enabling MFA, verify at least one factor is enrolled
	if msg.Policy.Enabled {
		params := m.GetParams(ctx)
		if params.RequireAtLeastOneFactor {
			enrollments := m.GetFactorEnrollments(ctx, address)
			hasActiveFactor := false
			for _, e := range enrollments {
				if e.IsActive() {
					hasActiveFactor = true
					break
				}
			}
			if !hasActiveFactor {
				return nil, types.ErrNoActiveFactors.Wrap("must enroll at least one factor before enabling MFA")
			}
		}
	}

	now := ctx.BlockTime().Unix()
	msg.Policy.UpdatedAt = now
	if !found {
		msg.Policy.CreatedAt = now
	} else {
		msg.Policy.CreatedAt = existingPolicy.CreatedAt
	}

	if err := m.Keeper.SetMFAPolicy(ctx, &msg.Policy); err != nil {
		return nil, err
	}

	return &types.MsgSetMFAPolicyResponse{
		Success: true,
	}, nil
}

// CreateChallenge creates a new MFA challenge
func (m *msgServer) CreateChallenge(goCtx context.Context, msg *types.MsgCreateChallenge) (*types.MsgCreateChallengeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	params := m.GetParams(ctx)

	// Find an active enrollment for this factor type
	enrollments := m.GetActiveFactorsByType(ctx, address, msg.FactorType)
	if len(enrollments) == 0 {
		return nil, types.ErrEnrollmentNotFound.Wrapf("no active %s factor found", msg.FactorType.String())
	}

	// Use specific factor if provided, otherwise use first active
	var enrollment types.FactorEnrollment
	if msg.FactorID != "" {
		found := false
		for _, e := range enrollments {
			if e.FactorID == msg.FactorID {
				enrollment = e
				found = true
				break
			}
		}
		if !found {
			return nil, types.ErrEnrollmentNotFound.Wrapf("factor %s not found", msg.FactorID)
		}
	} else {
		enrollment = enrollments[0]
	}

	// Create the challenge
	challenge, err := types.NewChallenge(
		msg.Sender,
		msg.FactorType,
		enrollment.FactorID,
		msg.TransactionType,
		params.ChallengeTTL,
		params.MaxChallengeAttempts,
	)
	if err != nil {
		return nil, err
	}

	challenge.Metadata = &types.ChallengeMetadata{
		ClientInfo: msg.ClientInfo,
	}

	// Add factor-specific challenge data
	if msg.FactorType == types.FactorTypeFIDO2 {
		challenge.Metadata.FIDO2Challenge = &types.FIDO2ChallengeData{
			Challenge:                   challenge.ChallengeData,
			RelyingPartyID:              "virtengine.com", // Configure from params
			AllowedCredentials:          [][]byte{enrollment.Metadata.FIDO2Info.CredentialID},
			UserVerificationRequirement: "preferred",
		}
	}

	if err := m.Keeper.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	return &types.MsgCreateChallengeResponse{
		ChallengeID:   challenge.ChallengeID,
		ChallengeData: challenge.ChallengeData,
		ExpiresAt:     challenge.ExpiresAt,
	}, nil
}

// VerifyChallenge verifies an MFA challenge response
func (m *msgServer) VerifyChallenge(goCtx context.Context, msg *types.MsgVerifyChallenge) (*types.MsgVerifyChallengeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	// Verify the challenge
	verified, err := m.Keeper.VerifyMFAChallenge(ctx, msg.ChallengeID, msg.Response)
	if err != nil {
		return &types.MsgVerifyChallengeResponse{
			Verified: false,
		}, err
	}

	if !verified {
		return &types.MsgVerifyChallengeResponse{
			Verified: false,
		}, nil
	}

	// Get the challenge to determine transaction type and create session
	challenge, _ := m.GetChallenge(ctx, msg.ChallengeID)

	// Get the policy to determine session duration
	policy, found := m.GetMFAPolicy(ctx, address)
	sessionDuration := m.GetParams(ctx).DefaultSessionDuration
	if found && policy.SessionDuration > 0 {
		sessionDuration = policy.SessionDuration
	}

	now := ctx.BlockTime().Unix()
	session := &types.AuthorizationSession{
		AccountAddress:  msg.Sender,
		TransactionType: challenge.TransactionType,
		VerifiedFactors: []types.FactorType{challenge.FactorType},
		CreatedAt:       now,
		ExpiresAt:       now + sessionDuration,
		IsSingleUse:     challenge.TransactionType.IsSingleUse(),
	}

	if msg.Response.ClientInfo != nil {
		session.DeviceFingerprint = msg.Response.ClientInfo.DeviceFingerprint
	}

	if err := m.Keeper.CreateAuthorizationSession(ctx, session); err != nil {
		return nil, err
	}

	// Link session to challenge
	challenge.SessionID = session.SessionID
	m.UpdateChallenge(ctx, challenge)

	// Check if more factors are required
	var remainingFactors []types.FactorType
	if found {
		requiredCombinations := policy.GetRequiredFactorsForAction(challenge.TransactionType)
		match := checkFactorsSatisfied(requiredCombinations, session.VerifiedFactors)
		if !match.Matched {
			remainingFactors = match.MissingFactors
		}
	}

	return &types.MsgVerifyChallengeResponse{
		Verified:         true,
		SessionID:        session.SessionID,
		SessionExpiresAt: session.ExpiresAt,
		RemainingFactors: remainingFactors,
	}, nil
}

// AddTrustedDevice adds a trusted device for the sender's account
func (m *msgServer) AddTrustedDevice(goCtx context.Context, msg *types.MsgAddTrustedDevice) (*types.MsgAddTrustedDeviceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	// Validate MFA proof
	if err := m.validateMFAProof(ctx, address, msg.MFAProof); err != nil {
		return nil, err
	}

	params := m.GetParams(ctx)
	now := ctx.BlockTime().Unix()

	deviceInfo := msg.DeviceInfo
	deviceInfo.TrustExpiresAt = now + params.TrustedDeviceTTL

	if err := m.Keeper.AddTrustedDevice(ctx, address, &deviceInfo); err != nil {
		return nil, err
	}

	return &types.MsgAddTrustedDeviceResponse{
		Success:        true,
		TrustExpiresAt: deviceInfo.TrustExpiresAt,
	}, nil
}

// RemoveTrustedDevice removes a trusted device
func (m *msgServer) RemoveTrustedDevice(goCtx context.Context, msg *types.MsgRemoveTrustedDevice) (*types.MsgRemoveTrustedDeviceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrapf("invalid sender address: %v", err)
	}

	// MFA proof may be optional if removing from the trusted device itself
	if msg.MFAProof != nil {
		if err := m.validateMFAProof(ctx, address, msg.MFAProof); err != nil {
			return nil, err
		}
	}

	if err := m.Keeper.RemoveTrustedDevice(ctx, address, msg.DeviceFingerprint); err != nil {
		return nil, err
	}

	return &types.MsgRemoveTrustedDeviceResponse{
		Success: true,
	}, nil
}

// UpdateSensitiveTxConfig updates sensitive transaction configuration (governance only)
func (m *msgServer) UpdateSensitiveTxConfig(goCtx context.Context, msg *types.MsgUpdateSensitiveTxConfig) (*types.MsgUpdateSensitiveTxConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != m.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := m.Keeper.SetSensitiveTxConfig(ctx, &msg.Config); err != nil {
		return nil, err
	}

	return &types.MsgUpdateSensitiveTxConfigResponse{
		Success: true,
	}, nil
}

// Helper functions

// validateMFAProof validates an MFA proof
func (m *msgServer) validateMFAProof(ctx sdk.Context, address sdk.AccAddress, proof *types.MFAProof) error {
	if proof == nil {
		return types.ErrMFARequired.Wrap("MFA proof is required")
	}

	if err := proof.Validate(); err != nil {
		return err
	}

	// Verify the session exists and is valid
	session, found := m.GetAuthorizationSession(ctx, proof.SessionID)
	if !found {
		return types.ErrSessionNotFound.Wrapf("session %s not found", proof.SessionID)
	}

	// Verify session belongs to this account
	if session.AccountAddress != address.String() {
		return types.ErrUnauthorized.Wrap("session belongs to different account")
	}

	// Verify session is still valid
	now := ctx.BlockTime()
	if !session.IsValid(now) {
		return types.ErrSessionExpired.Wrap("authorization session has expired")
	}

	return nil
}

// generateFactorID generates a unique factor ID
func generateFactorID(factorType types.FactorType, publicIdentifier []byte, label string) string {
	// In production, this would use a proper hash function
	// For now, use a combination of type and identifier
	if len(publicIdentifier) > 0 {
		return types.ComputeFactorFingerprint(factorType, publicIdentifier)
	}
	return types.ComputeFactorFingerprint(factorType, []byte(label+time.Now().String()))
}

// checkFactorsSatisfied checks if the verified factors satisfy any combination
func checkFactorsSatisfied(combinations []types.FactorCombination, verified []types.FactorType) types.PolicyMatch {
	verifiedSet := make(map[types.FactorType]bool)
	for _, ft := range verified {
		verifiedSet[ft] = true
	}

	result := types.PolicyMatch{
		Matched:          false,
		AvailableOptions: combinations,
	}

	for i, combo := range combinations {
		allPresent := true
		var missing []types.FactorType

		for _, requiredFactor := range combo.Factors {
			if !verifiedSet[requiredFactor] {
				allPresent = false
				missing = append(missing, requiredFactor)
			}
		}

		if allPresent {
			result.Matched = true
			result.MatchedCombination = &combinations[i]
			result.MissingFactors = nil
			return result
		}

		if result.MissingFactors == nil || len(missing) < len(result.MissingFactors) {
			result.MissingFactors = missing
		}
	}

	return result
}
