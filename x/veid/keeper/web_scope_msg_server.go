package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// SubmitSSOVerificationProof submits an SSO verification proof and records it on-chain.
func (ms msgServer) SubmitSSOVerificationProof(goCtx context.Context, msg *types.MsgSubmitSSOVerificationProof) (*types.MsgSubmitSSOVerificationProofResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, types.ErrInvalidSSO.Wrap("empty request")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	if msg.LinkageId == "" {
		return nil, types.ErrInvalidSSO.Wrap("linkage_id cannot be empty")
	}
	if len(msg.AttestationData) == 0 {
		return nil, types.ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}

	var att types.SSOAttestation
	if err := json.Unmarshal(msg.AttestationData, &att); err != nil {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation_data: %v", err)
	}
	if err := att.Validate(); err != nil {
		return nil, err
	}
	if att.LinkedAccountAddress != msg.AccountAddress {
		return nil, types.ErrInvalidAttestation.Wrap("attestation account address mismatch")
	}

	if err := ms.keeper.ValidateSSOAttestationSubmission(ctx, &att, nil); err != nil {
		return nil, err
	}

	if existing := ms.keeper.GetSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, att.ProviderType); existing != "" {
		return nil, types.ErrDuplicateLinkage.Wrapf("SSO linkage already exists: %s", existing)
	}

	evidenceHash := hashEvidence(msg.AttestationData)
	if msg.EvidenceHash != "" && msg.EvidenceHash != evidenceHash {
		return nil, types.ErrInvalidSSO.Wrap("evidence_hash does not match attestation_data")
	}

	linkage := att.ToLinkageMetadata(msg.LinkageId)
	linkage.EvidenceHash = evidenceHash
	linkage.EvidenceStorageBackend = msg.EvidenceStorageBackend
	linkage.EvidenceStorageRef = msg.EvidenceStorageRef
	linkage.EvidenceMetadata = msg.EvidenceMetadata

	if err := linkage.Validate(); err != nil {
		return nil, err
	}

	if err := ms.keeper.SetSSOLinkage(ctx, linkage); err != nil {
		return nil, err
	}
	ms.keeper.SetSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, att.ProviderType, msg.LinkageId)

	scoreContribution := types.GetSSOScoringWeight(att.ProviderType)
	if err := ms.keeper.applyWebScopeScore(ctx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "sso_" + string(att.ProviderType),
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitSSOVerificationProofResponse{
		LinkageId:         msg.LinkageId,
		Status:            string(types.SSOStatusVerified),
		ScoreContribution: scoreContribution,
		VerifiedAt:        ctx.BlockTime().Unix(),
	}, nil
}

// SubmitEmailVerificationProof submits an email verification proof and records it on-chain.
func (ms msgServer) SubmitEmailVerificationProof(goCtx context.Context, msg *types.MsgSubmitEmailVerificationProof) (*types.MsgSubmitEmailVerificationProofResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, types.ErrInvalidEmail.Wrap("empty request")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	if msg.VerificationId == "" {
		return nil, types.ErrInvalidEmail.Wrap("verification_id cannot be empty")
	}
	if msg.EmailHash == "" {
		return nil, types.ErrInvalidEmail.Wrap("email_hash cannot be empty")
	}
	if msg.Nonce == "" {
		return nil, types.ErrInvalidEmail.Wrap("nonce cannot be empty")
	}
	if len(msg.AttestationData) == 0 {
		return nil, types.ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}

	if _, found := ms.keeper.GetEmailVerificationRecord(ctx, msg.VerificationId); found {
		return nil, types.ErrInvalidEmail.Wrap("verification_id already exists")
	}

	nonceHash := hashEvidence([]byte(msg.Nonce))
	if ms.keeper.IsEmailNonceUsed(ctx, nonceHash) {
		return nil, types.ErrNonceAlreadyUsed.Wrap("email nonce already used")
	}

	attestation, err := types.AttestationFromJSON(msg.AttestationData)
	if err != nil {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation_data: %v", err)
	}
	if err := attestation.Validate(); err != nil {
		return nil, err
	}
	if attestation.Type != types.AttestationTypeEmailVerification {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation type: %s", attestation.Type)
	}
	if attestation.Subject.AccountAddress != msg.AccountAddress {
		return nil, types.ErrInvalidAttestation.Wrap("attestation account address mismatch")
	}

	evidenceHash := hashEvidence(msg.AttestationData)
	if msg.EvidenceHash != "" && msg.EvidenceHash != evidenceHash {
		return nil, types.ErrInvalidEmail.Wrap("evidence_hash does not match attestation_data")
	}

	now := ctx.BlockTime()
	verifiedAt := now
	if msg.VerifiedAt > 0 {
		verifiedAt = time.Unix(msg.VerifiedAt, 0)
	}
	var expiresAt *time.Time
	if msg.ExpiresAt > 0 {
		ts := time.Unix(msg.ExpiresAt, 0)
		expiresAt = &ts
	}

	record := &types.EmailVerificationRecord{
		Version:                types.EmailVerificationVersion,
		VerificationID:         msg.VerificationId,
		AccountAddress:         msg.AccountAddress,
		EmailHash:              msg.EmailHash,
		DomainHash:             msg.DomainHash,
		Nonce:                  msg.Nonce,
		NonceUsedAt:            &verifiedAt,
		Status:                 types.EmailStatusVerified,
		VerifiedAt:             &verifiedAt,
		ExpiresAt:              expiresAt,
		CreatedAt:              now,
		UpdatedAt:              now,
		AccountSignature:       msg.AccountSignature,
		IsOrganizational:       msg.IsOrganizational,
		EvidenceHash:           evidenceHash,
		EvidenceStorageBackend: msg.EvidenceStorageBackend,
		EvidenceStorageRef:     msg.EvidenceStorageRef,
		EvidenceMetadata:       msg.EvidenceMetadata,
	}

	if err := ms.keeper.SetEmailVerificationRecord(ctx, record); err != nil {
		return nil, err
	}

	nonceRecord := types.NewUsedNonceRecord(msg.Nonce, verifiedAt, msg.AccountAddress, msg.VerificationId, 365)
	if err := ms.keeper.SetEmailUsedNonce(ctx, nonceRecord); err != nil {
		return nil, err
	}

	scoreContribution := types.CalculateEmailScore(record, types.DefaultEmailScoringWeight(), ctx.BlockTime())
	if err := ms.keeper.applyWebScopeScore(ctx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "email_verification",
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitEmailVerificationProofResponse{
		VerificationId:    msg.VerificationId,
		Status:            string(types.EmailStatusVerified),
		ScoreContribution: scoreContribution,
		VerifiedAt:        verifiedAt.Unix(),
	}, nil
}

// SubmitSMSVerificationProof submits an SMS verification proof and records it on-chain.
func (ms msgServer) SubmitSMSVerificationProof(goCtx context.Context, msg *types.MsgSubmitSMSVerificationProof) (*types.MsgSubmitSMSVerificationProofResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, types.ErrInvalidPhone.Wrap("empty request")
	}

	_, err := sdk.AccAddressFromBech32(msg.AccountAddress)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	if msg.VerificationId == "" {
		return nil, types.ErrInvalidPhone.Wrap("verification_id cannot be empty")
	}
	if msg.PhoneHash == "" {
		return nil, types.ErrInvalidPhone.Wrap("phone_hash cannot be empty")
	}
	if msg.PhoneHashSalt == "" {
		return nil, types.ErrInvalidPhone.Wrap("phone_hash_salt cannot be empty")
	}
	if len(msg.AttestationData) == 0 {
		return nil, types.ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}

	if _, found := ms.keeper.GetSMSVerificationRecord(ctx, msg.VerificationId); found {
		return nil, types.ErrInvalidPhone.Wrap("verification_id already exists")
	}

	attestation, err := types.AttestationFromJSON(msg.AttestationData)
	if err != nil {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation_data: %v", err)
	}
	if err := attestation.Validate(); err != nil {
		return nil, err
	}
	if attestation.Type != types.AttestationTypeSMSVerification {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation type: %s", attestation.Type)
	}
	if attestation.Subject.AccountAddress != msg.AccountAddress {
		return nil, types.ErrInvalidAttestation.Wrap("attestation account address mismatch")
	}

	evidenceHash := hashEvidence(msg.AttestationData)
	if msg.EvidenceHash != "" && msg.EvidenceHash != evidenceHash {
		return nil, types.ErrInvalidPhone.Wrap("evidence_hash does not match attestation_data")
	}

	now := ctx.BlockTime()
	verifiedAt := now
	if msg.VerifiedAt > 0 {
		verifiedAt = time.Unix(msg.VerifiedAt, 0)
	}
	var expiresAt *time.Time
	if msg.ExpiresAt > 0 {
		ts := time.Unix(msg.ExpiresAt, 0)
		expiresAt = &ts
	}

	record := &types.SMSVerificationRecord{
		Version:        types.SMSVerificationVersion,
		VerificationID: msg.VerificationId,
		AccountAddress: msg.AccountAddress,
		PhoneHash: types.PhoneNumberHash{
			Hash:            msg.PhoneHash,
			Salt:            msg.PhoneHashSalt,
			CountryCodeHash: msg.CountryCodeHash,
			CreatedAt:       now,
		},
		Status:                 types.SMSStatusVerified,
		VerifiedAt:             &verifiedAt,
		ExpiresAt:              expiresAt,
		CreatedAt:              now,
		UpdatedAt:              now,
		IsVoIP:                 msg.IsVoip,
		CarrierType:            msg.CarrierType,
		ValidatorAddress:       msg.ValidatorAddress,
		AccountSignature:       msg.AccountSignature,
		EvidenceHash:           evidenceHash,
		EvidenceStorageBackend: msg.EvidenceStorageBackend,
		EvidenceStorageRef:     msg.EvidenceStorageRef,
		EvidenceMetadata:       msg.EvidenceMetadata,
	}

	if err := ms.keeper.SetSMSVerificationRecord(ctx, record); err != nil {
		return nil, err
	}

	scoreContribution := types.CalculateSMSScore(record, types.DefaultSMSScoringWeight(), ctx.BlockTime())
	if err := ms.keeper.applyWebScopeScore(ctx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "sms_verification",
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitSMSVerificationProofResponse{
		VerificationId:    msg.VerificationId,
		Status:            string(types.SMSStatusVerified),
		ScoreContribution: scoreContribution,
		VerifiedAt:        verifiedAt.Unix(),
	}, nil
}

// SubmitSocialMediaScope submits a social media scope and records it on-chain.
func (ms msgServer) SubmitSocialMediaScope(goCtx context.Context, msg *types.MsgSubmitSocialMediaScope) (*types.MsgSubmitSocialMediaScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, types.ErrInvalidScope.Wrap("empty request")
	}

	if _, err := sdk.AccAddressFromBech32(msg.AccountAddress); err != nil {
		return nil, types.ErrInvalidAddress.Wrap(errMsgInvalidAccountAddr)
	}

	if msg.ScopeId == "" {
		return nil, types.ErrInvalidScope.Wrap("scope_id cannot be empty")
	}

	provider := types.SocialMediaProviderFromProto(msg.Provider)
	if provider == "" {
		return nil, types.ErrInvalidScope.Wrap("invalid provider")
	}

	if len(msg.AttestationData) == 0 {
		return nil, types.ErrInvalidAttestation.Wrap("attestation_data cannot be empty")
	}

	if _, found := ms.keeper.GetSocialMediaScope(ctx, msg.ScopeId); found {
		return nil, types.ErrInvalidScope.Wrap("scope_id already exists")
	}

	attestation, err := types.AttestationFromJSON(msg.AttestationData)
	if err != nil {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation_data: %v", err)
	}
	if err := attestation.Validate(); err != nil {
		return nil, err
	}
	if attestation.Type != types.AttestationTypeSocialMediaVerification {
		return nil, types.ErrInvalidAttestation.Wrapf("invalid attestation type: %s", attestation.Type)
	}
	if attestation.Subject.AccountAddress != msg.AccountAddress {
		return nil, types.ErrInvalidAttestation.Wrap("attestation account address mismatch")
	}

	evidenceHash := hashEvidence(msg.AttestationData)
	if msg.EvidenceHash != "" && msg.EvidenceHash != evidenceHash {
		return nil, types.ErrInvalidScope.Wrap("evidence_hash does not match attestation_data")
	}

	now := ctx.BlockTime()
	var accountCreatedAt *time.Time
	if msg.AccountCreatedAt > 0 {
		if msg.AccountCreatedAt > now.Unix() {
			return nil, types.ErrInvalidScope.Wrap("account_created_at cannot be in the future")
		}
		ts := time.Unix(msg.AccountCreatedAt, 0)
		accountCreatedAt = &ts
	}

	ageDays := msg.AccountAgeDays
	if ageDays == 0 && accountCreatedAt != nil {
		ageDays = uint32(now.Sub(*accountCreatedAt).Hours() / 24)
	}

	scope := &types.SocialMediaScope{
		Version:                types.SocialMediaScopeVersion,
		ScopeID:                msg.ScopeId,
		AccountAddress:         msg.AccountAddress,
		Provider:               provider,
		ProfileNameHash:        msg.ProfileNameHash,
		EmailHash:              msg.EmailHash,
		UsernameHash:           msg.UsernameHash,
		OrgHash:                msg.OrgHash,
		AccountCreatedAt:       accountCreatedAt,
		AccountAgeDays:         ageDays,
		IsVerified:             msg.IsVerified,
		FriendCountRange:       msg.FriendCountRange,
		Status:                 types.SocialMediaStatusVerified,
		CreatedAt:              now,
		UpdatedAt:              now,
		EncryptedPayload:       encryptedPayloadFromProto(&msg.EncryptedPayload),
		EvidenceHash:           evidenceHash,
		EvidenceStorageBackend: msg.EvidenceStorageBackend,
		EvidenceStorageRef:     msg.EvidenceStorageRef,
		EvidenceMetadata:       msg.EvidenceMetadata,
	}

	if err := scope.Validate(); err != nil {
		return nil, err
	}

	if err := ms.keeper.SetSocialMediaScope(ctx, scope); err != nil {
		return nil, err
	}

	nameMatch := false
	address := sdk.MustAccAddressFromBech32(msg.AccountAddress)
	if wallet, found := ms.keeper.GetWallet(ctx, address); found {
		if docHash, ok := wallet.DerivedFeatures.DocFieldHashes[types.DocFieldNameHash]; ok {
			if nameHashBytes, err := hex.DecodeString(msg.ProfileNameHash); err == nil {
				nameMatch = bytes.Equal(docHash, nameHashBytes)
			}
		}
	}

	scoreContribution := types.CalculateSocialMediaScore(scope, nameMatch, now)
	if err := ms.keeper.applyWebScopeScore(ctx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "social_media_" + string(provider),
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}

	return &types.MsgSubmitSocialMediaScopeResponse{
		ScopeId:           msg.ScopeId,
		Status:            string(types.SocialMediaStatusVerified),
		ScoreContribution: scoreContribution,
		VerifiedAt:        now.Unix(),
	}, nil
}

func hashEvidence(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
