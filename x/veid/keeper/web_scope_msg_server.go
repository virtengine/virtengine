package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
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
	if err := validateSSOWebEvidenceSubjectBinding(msg.AccountAddress, &att); err != nil {
		return nil, err
	}

	evidenceHash := hashEvidence(msg.AttestationData)
	if msg.EvidenceHash != "" && msg.EvidenceHash != evidenceHash {
		return nil, types.ErrInvalidSSO.Wrap("evidence_hash does not match attestation_data")
	}

	evidence, err := buildSSOWebEvidenceContext(ctx, msg, &att)
	if err != nil {
		return nil, err
	}
	validation, err := ms.keeper.validateWebEvidenceSubmission(ctx, &att.VerificationAttestation, evidence, att.LinkageSignature)
	if err != nil {
		return nil, err
	}

	linkage := att.ToLinkageMetadata(msg.LinkageId)
	linkage.EvidenceHash = evidenceHash
	linkage.EvidenceStorageBackend = msg.EvidenceStorageBackend
	linkage.EvidenceStorageRef = msg.EvidenceStorageRef
	linkage.EvidenceMetadata = msg.EvidenceMetadata

	if err := linkage.Validate(); err != nil {
		return nil, err
	}

	if existing := ms.keeper.GetSSOLinkageByAccountAndProvider(ctx, msg.AccountAddress, att.ProviderType); existing != "" {
		existingLinkage, found := ms.keeper.GetSSOLinkage(ctx, existing)
		if validation.ExactReplay &&
			found &&
			existing == msg.LinkageId &&
			existingLinkage.EvidenceHash == evidenceHash &&
			webEvidenceStorageMatches(
				existingLinkage.EvidenceStorageBackend,
				existingLinkage.EvidenceStorageRef,
				existingLinkage.EvidenceMetadata,
				msg.EvidenceStorageBackend,
				msg.EvidenceStorageRef,
				msg.EvidenceMetadata,
			) {
			return &types.MsgSubmitSSOVerificationProofResponse{
				LinkageId:         msg.LinkageId,
				Status:            string(existingLinkage.Status),
				ScoreContribution: types.GetSSOScoringWeight(att.ProviderType),
				VerifiedAt:        existingLinkage.VerifiedAt.Unix(),
			}, nil
		}
		return nil, types.ErrDuplicateLinkage.Wrapf("SSO linkage already exists: %s", existing)
	}
	if validation.ExactReplay {
		return nil, types.ErrNonceAlreadyUsed.Wrap("web evidence replay target missing")
	}
	if _, found := ms.keeper.GetSSOLinkage(ctx, msg.LinkageId); found {
		return nil, types.ErrDuplicateLinkage.Wrapf("SSO linkage already exists: %s", msg.LinkageId)
	}
	if ms.keeper.IsSSONonceUsed(ctx, hashNonce(att.OIDCNonce)) {
		return nil, types.ErrNonceAlreadyUsed.Wrap("OIDC nonce already used")
	}
	applyCtx, write := ctx.CacheContext()
	if err := ms.keeper.recordWebEvidenceReplay(applyCtx, validation); err != nil {
		return nil, err
	}
	if err := ms.keeper.recordSSOAttestationSubmission(applyCtx, &att, validation.SignerKey); err != nil {
		return nil, err
	}

	if err := ms.keeper.SetSSOLinkage(applyCtx, linkage); err != nil {
		return nil, err
	}
	ms.keeper.SetSSOLinkageByAccountAndProvider(applyCtx, msg.AccountAddress, att.ProviderType, msg.LinkageId)

	scoreContribution := types.GetSSOScoringWeight(att.ProviderType)
	if err := ms.keeper.applyWebScopeScore(applyCtx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "sso_" + string(att.ProviderType),
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}
	write()

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

	evidence, err := buildEmailWebEvidenceContext(ctx, msg, attestation)
	if err != nil {
		return nil, err
	}
	validation, err := ms.keeper.validateWebEvidenceSubmission(ctx, attestation, evidence, msg.AccountSignature)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime()
	verifiedAt := attestation.IssuedAt
	expiresAt := &attestation.ExpiresAt

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

	if err := record.Validate(); err != nil {
		return nil, err
	}
	if existing, found := ms.keeper.GetEmailVerificationRecord(ctx, msg.VerificationId); found {
		if validation.ExactReplay &&
			existing.AccountAddress == msg.AccountAddress &&
			existing.EmailHash == msg.EmailHash &&
			existing.EvidenceHash == evidenceHash &&
			bytes.Equal(existing.AccountSignature, msg.AccountSignature) &&
			webEvidenceStorageMatches(
				existing.EvidenceStorageBackend,
				existing.EvidenceStorageRef,
				existing.EvidenceMetadata,
				msg.EvidenceStorageBackend,
				msg.EvidenceStorageRef,
				msg.EvidenceMetadata,
			) {
			scoreContribution := types.CalculateEmailScore(existing, types.DefaultEmailScoringWeight(), ctx.BlockTime())
			return &types.MsgSubmitEmailVerificationProofResponse{
				VerificationId:    msg.VerificationId,
				Status:            string(existing.Status),
				ScoreContribution: scoreContribution,
				VerifiedAt:        existing.VerifiedAt.Unix(),
			}, nil
		}
		return nil, types.ErrInvalidEmail.Wrap("verification_id already exists")
	}
	if validation.ExactReplay {
		return nil, types.ErrNonceAlreadyUsed.Wrap("web evidence replay target missing")
	}
	nonceHash := hashEvidence([]byte(msg.Nonce))
	if ms.keeper.IsEmailNonceUsed(ctx, nonceHash) {
		return nil, types.ErrNonceAlreadyUsed.Wrap("email nonce already used")
	}
	applyCtx, write := ctx.CacheContext()
	if err := ms.keeper.recordWebEvidenceReplay(applyCtx, validation); err != nil {
		return nil, err
	}
	if err := ms.keeper.SetEmailVerificationRecord(applyCtx, record); err != nil {
		return nil, err
	}

	nonceRecord := types.NewUsedNonceRecord(msg.Nonce, verifiedAt, msg.AccountAddress, msg.VerificationId, 365)
	if err := ms.keeper.SetEmailUsedNonce(applyCtx, nonceRecord); err != nil {
		return nil, err
	}

	scoreContribution := types.CalculateEmailScore(record, types.DefaultEmailScoringWeight(), applyCtx.BlockTime())
	if err := ms.keeper.applyWebScopeScore(applyCtx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "email_verification",
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}
	write()

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

	evidence, err := buildSMSWebEvidenceContext(ctx, msg, attestation)
	if err != nil {
		return nil, err
	}
	validation, err := ms.keeper.validateWebEvidenceSubmission(ctx, attestation, evidence, msg.AccountSignature)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime()
	verifiedAt := attestation.IssuedAt
	expiresAt := &attestation.ExpiresAt

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

	if err := record.Validate(); err != nil {
		return nil, err
	}
	if existing, found := ms.keeper.GetSMSVerificationRecord(ctx, msg.VerificationId); found {
		if validation.ExactReplay &&
			existing.AccountAddress == msg.AccountAddress &&
			existing.PhoneHash.Hash == msg.PhoneHash &&
			existing.EvidenceHash == evidenceHash &&
			bytes.Equal(existing.AccountSignature, msg.AccountSignature) &&
			webEvidenceStorageMatches(
				existing.EvidenceStorageBackend,
				existing.EvidenceStorageRef,
				existing.EvidenceMetadata,
				msg.EvidenceStorageBackend,
				msg.EvidenceStorageRef,
				msg.EvidenceMetadata,
			) {
			scoreContribution := types.CalculateSMSScore(existing, types.DefaultSMSScoringWeight(), ctx.BlockTime())
			return &types.MsgSubmitSMSVerificationProofResponse{
				VerificationId:    msg.VerificationId,
				Status:            string(existing.Status),
				ScoreContribution: scoreContribution,
				VerifiedAt:        existing.VerifiedAt.Unix(),
			}, nil
		}
		return nil, types.ErrInvalidPhone.Wrap("verification_id already exists")
	}
	if validation.ExactReplay {
		return nil, types.ErrNonceAlreadyUsed.Wrap("web evidence replay target missing")
	}
	applyCtx, write := ctx.CacheContext()
	if err := ms.keeper.recordWebEvidenceReplay(applyCtx, validation); err != nil {
		return nil, err
	}
	if err := ms.keeper.SetSMSVerificationRecord(applyCtx, record); err != nil {
		return nil, err
	}

	scoreContribution := types.CalculateSMSScore(record, types.DefaultSMSScoringWeight(), applyCtx.BlockTime())
	if err := ms.keeper.applyWebScopeScore(applyCtx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "sms_verification",
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}
	write()

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

	evidence, err := buildSocialWebEvidenceContext(ctx, msg, attestation, provider, ageDays)
	if err != nil {
		return nil, err
	}
	validation, err := ms.keeper.validateWebEvidenceSubmission(ctx, attestation, evidence, msg.AccountSignature)
	if err != nil {
		return nil, err
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

	if existing, found := ms.keeper.GetSocialMediaScope(ctx, msg.ScopeId); found {
		if validation.ExactReplay &&
			existing.AccountAddress == msg.AccountAddress &&
			existing.Provider == provider &&
			existing.EvidenceHash == evidenceHash &&
			webEvidenceStorageMatches(
				existing.EvidenceStorageBackend,
				existing.EvidenceStorageRef,
				existing.EvidenceMetadata,
				msg.EvidenceStorageBackend,
				msg.EvidenceStorageRef,
				msg.EvidenceMetadata,
			) {
			scoreContribution := types.CalculateSocialMediaScore(existing, webScopeSocialNameMatch(ctx, ms.keeper, msg.AccountAddress, existing.ProfileNameHash), now)
			return &types.MsgSubmitSocialMediaScopeResponse{
				ScopeId:           msg.ScopeId,
				Status:            string(existing.Status),
				ScoreContribution: scoreContribution,
				VerifiedAt:        existing.UpdatedAt.Unix(),
			}, nil
		}
		return nil, types.ErrInvalidScope.Wrap("scope_id already exists")
	}
	if validation.ExactReplay {
		return nil, types.ErrNonceAlreadyUsed.Wrap("web evidence replay target missing")
	}
	applyCtx, write := ctx.CacheContext()
	if err := ms.keeper.recordWebEvidenceReplay(applyCtx, validation); err != nil {
		return nil, err
	}
	if err := ms.keeper.SetSocialMediaScope(applyCtx, scope); err != nil {
		return nil, err
	}

	nameMatch := webScopeSocialNameMatch(applyCtx, ms.keeper, msg.AccountAddress, msg.ProfileNameHash)
	scoreContribution := types.CalculateSocialMediaScore(scope, nameMatch, now)
	if err := ms.keeper.applyWebScopeScore(applyCtx, msg.AccountAddress, []webScopeContribution{
		{
			FeatureName:  "social_media_" + string(provider),
			ScoreBasisPt: scoreContribution,
			EvidenceHash: evidenceHash,
		},
	}); err != nil {
		return nil, err
	}
	write()

	return &types.MsgSubmitSocialMediaScopeResponse{
		ScopeId:           msg.ScopeId,
		Status:            string(types.SocialMediaStatusVerified),
		ScoreContribution: scoreContribution,
		VerifiedAt:        now.Unix(),
	}, nil
}

func webScopeSocialNameMatch(ctx sdk.Context, k Keeper, accountAddress string, profileNameHash string) bool {
	address := sdk.MustAccAddressFromBech32(accountAddress)
	wallet, found := k.GetWallet(ctx, address)
	if !found {
		return false
	}
	docHash, ok := wallet.DerivedFeatures.DocFieldHashes[types.DocFieldNameHash]
	if !ok {
		return false
	}
	nameHashBytes, err := hex.DecodeString(profileNameHash)
	return err == nil && bytes.Equal(docHash, nameHashBytes)
}

func hashEvidence(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func validateSSOWebEvidenceSubjectBinding(accountAddress string, att *types.SSOAttestation) error {
	if att == nil {
		return types.ErrInvalidAttestation.Wrap("SSO attestation cannot be nil")
	}
	if att.LinkedAccountAddress != accountAddress {
		return types.ErrInvalidAttestation.Wrap("SSO linked account address mismatch")
	}
	if att.Subject.AccountAddress != accountAddress {
		return types.ErrInvalidAttestation.Wrap("SSO subject account address mismatch")
	}
	expectedSubjectID := "did:virtengine:" + accountAddress
	if att.Subject.ID != expectedSubjectID {
		return types.ErrInvalidAttestation.Wrap("SSO subject ID does not match account address")
	}
	return nil
}

func buildSSOWebEvidenceContext(ctx sdk.Context, msg *types.MsgSubmitSSOVerificationProof, att *types.SSOAttestation) (types.WebEvidenceContext, error) {
	canonical, err := att.CanonicalBytes()
	if err != nil {
		return types.WebEvidenceContext{}, types.ErrInvalidAttestation.Wrapf("failed to build SSO canonical bytes: %v", err)
	}
	serviceHash, err := webEvidenceServiceMetadataHash(msg.EvidenceMetadata)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	return types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             ctx.ChainID(),
		AccountAddress:      msg.AccountAddress,
		EvidenceType:        types.AttestationTypeSSOVerification,
		Action:              types.WebEvidenceActionSubmitSSO,
		ScopeID:             msg.LinkageId,
		AttestationDigest:   types.WebEvidenceDigestHex(canonical),
		Issuer:              att.Issuer,
		IssuerAlgorithm:     att.Proof.Type,
		Nonce:               att.Nonce,
		Challenge:           att.OIDCNonce,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: serviceHash,
		CallerFields: map[string]string{
			"linkage_id":             msg.LinkageId,
			"provider":               string(att.ProviderType),
			"oidc_issuer":            att.OIDCIssuer,
			"subject_hash":           att.SubjectHash,
			"email_hash":             att.EmailHash,
			"email_domain_hash":      att.EmailDomainHash,
			"tenant_id_hash":         att.TenantIDHash,
			"oidc_nonce":             att.OIDCNonce,
			"email_verified":         strconv.FormatBool(att.EmailVerified),
			"linked_account_address": att.LinkedAccountAddress,
		},
	}), nil
}

func buildEmailWebEvidenceContext(ctx sdk.Context, msg *types.MsgSubmitEmailVerificationProof, att *types.VerificationAttestation) (types.WebEvidenceContext, error) {
	if msg.Nonce != att.Nonce {
		return types.WebEvidenceContext{}, types.ErrInvalidAttestation.Wrap("email nonce does not match attestation nonce")
	}
	if err := requireUnixMatch("verified_at", msg.VerifiedAt, att.IssuedAt); err != nil {
		return types.WebEvidenceContext{}, err
	}
	if err := requireUnixMatch("expires_at", msg.ExpiresAt, att.ExpiresAt); err != nil {
		return types.WebEvidenceContext{}, err
	}
	digest, err := types.WebEvidenceAttestationDigestHex(att)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	serviceHash, err := webEvidenceServiceMetadataHash(msg.EvidenceMetadata)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	return types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             ctx.ChainID(),
		AccountAddress:      msg.AccountAddress,
		EvidenceType:        types.AttestationTypeEmailVerification,
		Action:              types.WebEvidenceActionSubmitEmail,
		ScopeID:             msg.VerificationId,
		AttestationDigest:   digest,
		Issuer:              att.Issuer,
		IssuerAlgorithm:     att.Proof.Type,
		Nonce:               att.Nonce,
		Challenge:           msg.VerificationId,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: serviceHash,
		CallerFields: map[string]string{
			"verification_id":   msg.VerificationId,
			"email_hash":        msg.EmailHash,
			"domain_hash":       msg.DomainHash,
			"nonce":             msg.Nonce,
			"is_organizational": strconv.FormatBool(msg.IsOrganizational),
			"verified_at_unix":  strconv.FormatInt(att.IssuedAt.Unix(), 10),
			"expires_at_unix":   strconv.FormatInt(att.ExpiresAt.Unix(), 10),
		},
	}), nil
}

func buildSMSWebEvidenceContext(ctx sdk.Context, msg *types.MsgSubmitSMSVerificationProof, att *types.VerificationAttestation) (types.WebEvidenceContext, error) {
	if err := requireUnixMatch("verified_at", msg.VerifiedAt, att.IssuedAt); err != nil {
		return types.WebEvidenceContext{}, err
	}
	if err := requireUnixMatch("expires_at", msg.ExpiresAt, att.ExpiresAt); err != nil {
		return types.WebEvidenceContext{}, err
	}
	digest, err := types.WebEvidenceAttestationDigestHex(att)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	serviceHash, err := webEvidenceServiceMetadataHash(msg.EvidenceMetadata)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	return types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             ctx.ChainID(),
		AccountAddress:      msg.AccountAddress,
		EvidenceType:        types.AttestationTypeSMSVerification,
		Action:              types.WebEvidenceActionSubmitSMS,
		ScopeID:             msg.VerificationId,
		AttestationDigest:   digest,
		Issuer:              att.Issuer,
		IssuerAlgorithm:     att.Proof.Type,
		Nonce:               att.Nonce,
		Challenge:           msg.VerificationId,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: serviceHash,
		CallerFields: map[string]string{
			"verification_id":   msg.VerificationId,
			"phone_hash":        msg.PhoneHash,
			"phone_hash_salt":   msg.PhoneHashSalt,
			"country_code_hash": msg.CountryCodeHash,
			"is_voip":           strconv.FormatBool(msg.IsVoip),
			"carrier_type":      msg.CarrierType,
			"validator_address": msg.ValidatorAddress,
			"verified_at_unix":  strconv.FormatInt(att.IssuedAt.Unix(), 10),
			"expires_at_unix":   strconv.FormatInt(att.ExpiresAt.Unix(), 10),
		},
	}), nil
}

func buildSocialWebEvidenceContext(
	ctx sdk.Context,
	msg *types.MsgSubmitSocialMediaScope,
	att *types.VerificationAttestation,
	provider types.SocialMediaProviderType,
	ageDays uint32,
) (types.WebEvidenceContext, error) {
	if msg.AccountCreatedAt > 0 && msg.AccountCreatedAt > ctx.BlockTime().Unix() {
		return types.WebEvidenceContext{}, types.ErrInvalidScope.Wrap("account_created_at cannot be in the future")
	}
	digest, err := types.WebEvidenceAttestationDigestHex(att)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	serviceHash, err := webEvidenceServiceMetadataHash(msg.EvidenceMetadata)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	payloadDigest, err := webEvidenceEncryptedPayloadDigest(&msg.EncryptedPayload)
	if err != nil {
		return types.WebEvidenceContext{}, err
	}
	return types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             ctx.ChainID(),
		AccountAddress:      msg.AccountAddress,
		EvidenceType:        types.AttestationTypeSocialMediaVerification,
		Action:              types.WebEvidenceActionSubmitSocial,
		ScopeID:             msg.ScopeId,
		AttestationDigest:   digest,
		Issuer:              att.Issuer,
		IssuerAlgorithm:     att.Proof.Type,
		Nonce:               att.Nonce,
		Challenge:           msg.ScopeId,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: serviceHash,
		CallerFields: map[string]string{
			"scope_id":                 msg.ScopeId,
			"provider":                 string(provider),
			"profile_name_hash":        msg.ProfileNameHash,
			"email_hash":               msg.EmailHash,
			"username_hash":            msg.UsernameHash,
			"org_hash":                 msg.OrgHash,
			"account_created_at_unix":  strconv.FormatInt(msg.AccountCreatedAt, 10),
			"account_age_days":         strconv.FormatUint(uint64(ageDays), 10),
			"is_verified":              strconv.FormatBool(msg.IsVerified),
			"friend_count_range":       msg.FriendCountRange,
			"encrypted_payload_digest": payloadDigest,
		},
	}), nil
}

func requireUnixMatch(field string, msgValue int64, expected time.Time) error {
	if msgValue > 0 && msgValue != expected.Unix() {
		return types.ErrInvalidAttestation.Wrapf("%s does not match attestation timestamp", field)
	}
	return nil
}

func webEvidenceServiceMetadataHash(metadata map[string]string) (string, error) {
	return types.WebEvidenceServiceMetadataHash(metadata)
}

func webEvidenceStorageMatches(
	storedBackend string,
	storedRef string,
	storedMetadata map[string]string,
	msgBackend string,
	msgRef string,
	msgMetadata map[string]string,
) bool {
	if storedBackend != msgBackend || storedRef != msgRef {
		return false
	}
	if len(storedMetadata) != len(msgMetadata) {
		return false
	}
	for key, storedValue := range storedMetadata {
		if msgValue, ok := msgMetadata[key]; !ok || msgValue != storedValue {
			return false
		}
	}
	return true
}

func webEvidenceEncryptedPayloadDigest(payload *types.EncryptedPayloadEnvelopePB) (string, error) {
	converted := encryptedPayloadFromProto(payload)
	bz, err := json.Marshal(converted)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(bz)
	return hex.EncodeToString(hash[:]), nil
}
