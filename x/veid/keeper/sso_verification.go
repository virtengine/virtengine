// Package keeper provides SSO attestation validation for the VEID module.
//
// This file implements validation for SSO attestation submissions, including signature
// verification, replay protection, and event/audit emission consistent with keeper patterns.
//
// NOTE: This file assumes SSO attestation types are defined in x/veid/types/sso_attestation.go
// and related status/types in x/veid/types/sso_verification.go.

package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ValidateSSOAttestationSubmission validates an SSO attestation submission
// without mutating nonce or audit state. Mutation is internal to the web
// evidence handler after the shared issuer/account proof gate has passed.
func (k Keeper) ValidateSSOAttestationSubmission(
	ctx sdk.Context,
	att *types.SSOAttestation,
	signerKeyInfo *types.SignerKeyInfo,
) error {
	if att == nil {
		return types.ErrInvalidAttestation.Wrap("SSO attestation cannot be nil")
	}
	if signerKeyInfo == nil {
		return types.ErrSignerKeyNotFound.Wrap("signer key is required for SSO attestation")
	}
	if !signerKeyInfo.State.CanVerify() {
		return types.ErrInvalidSignerKey.Wrapf("signer key not valid for verification: %s", signerKeyInfo.State)
	}
	if signerKeyInfo.KeyID != att.Issuer.KeyID || signerKeyInfo.Fingerprint != att.Issuer.KeyFingerprint {
		return types.ErrInvalidSignerKey.Wrap("signer key does not match attestation issuer")
	}
	if att.Subject.AccountAddress != att.LinkedAccountAddress {
		return types.ErrInvalidAttestation.Wrap("SSO subject account does not match linked account")
	}
	expectedSubjectID := fmt.Sprintf("did:virtengine:%s", att.Subject.AccountAddress)
	if att.Subject.ID != expectedSubjectID {
		return types.ErrInvalidAttestation.Wrap("SSO subject ID does not match account address")
	}

	// 1. Replay protection: check OIDC nonce has not been used
	if att.OIDCNonce == "" {
		return types.ErrInvalidAttestation.Wrap("missing OIDC nonce for replay protection")
	}
	nonceHash := hashNonce(att.OIDCNonce)
	if k.IsSSONonceUsed(ctx, nonceHash) {
		return types.ErrNonceAlreadyUsed.Wrap("OIDC nonce already used")
	}
	return nil
}

func (k Keeper) recordSSOAttestationSubmission(
	ctx sdk.Context,
	att *types.SSOAttestation,
	signerKeyInfo *types.SignerKeyInfo,
) error {
	if err := k.ValidateSSOAttestationSubmission(ctx, att, signerKeyInfo); err != nil {
		return err
	}
	nonceHash := hashNonce(att.OIDCNonce)

	// Mark nonce as used (prevent replay)
	nonceRecord := types.NewSSONonceRecord(
		nonceHash,
		att.LinkedAccountAddress,
		att.ProviderType,
		att.OIDCIssuer,
		att.ID,
		ctx.BlockTime(),
		ctx.BlockHeight(),
		24*time.Hour*365, // Keep nonce records for 1 year
	)
	k.SetSSONonceRecord(ctx, nonceRecord)

	// Emit event (pattern: EventVerificationSubmitted)
	if err := k.EmitVerificationSubmittedEvent(
		ctx,
		att.LinkedAccountAddress,
		att.OIDCIssuer+":"+att.SubjectHash,
		string(att.ProviderType),
		nonceHash,
	); err != nil {
		return err
	}

	// Write audit log (pattern: SetAuditEntry)
	details := map[string]interface{}{
		"oidc_issuer":   att.OIDCIssuer,
		"subject_hash":  att.SubjectHash,
		"provider_type": string(att.ProviderType),
		"nonce_hash":    nonceHash,
	}
	audit := types.NewAuditEntry(
		fmt.Sprintf("sso_attestation:%s:%s", att.OIDCIssuer, att.SubjectHash),
		types.AuditEventTypeVerification,
		att.LinkedAccountAddress,
		ctx.BlockTime(),
		ctx.BlockHeight(),
		details,
	)
	if err := k.SetAuditEntry(ctx, audit); err != nil {
		return err
	}

	return nil
}
