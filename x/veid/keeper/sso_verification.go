// Package keeper provides SSO attestation validation for the VEID module.
//
// This file implements validation for SSO attestation submissions, including signature
// verification, replay protection, and event/audit emission consistent with keeper patterns.
//
// NOTE: This file assumes SSO attestation types are defined in x/veid/types/sso_attestation.go
// and related status/types in x/veid/types/sso_verification.go.

package keeper

import (
	"crypto/ed25519"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ValidateSSOAttestationSubmission validates an SSO attestation submission.
// - Verifies attestation signature using signer key info (if available)
// - Enforces replay protection (OIDC nonce)
// - Emits events and audit logs consistent with keeper patterns
func (k Keeper) ValidateSSOAttestationSubmission(
	ctx sdk.Context,
	att *types.SSOAttestation,
	signerKeyInfo *types.SignerKeyInfo, // may be nil if not available
) error {
	// 1. Replay protection: check OIDC nonce has not been used
	if att.OIDCNonce == "" {
		return types.ErrInvalidAttestation.Wrap("missing OIDC nonce for replay protection")
	}
	nonceHash := hashNonce(att.OIDCNonce)
	if k.IsSSONonceUsed(ctx, nonceHash) {
		return types.ErrNonceAlreadyUsed.Wrap("OIDC nonce already used")
	}

	// 2. Signature validation (if signerKeyInfo provided)
	if signerKeyInfo != nil {
		if !signerKeyInfo.State.CanVerify() {
			return types.ErrInvalidSignerKey.Wrapf("signer key not valid for verification: %s", signerKeyInfo.State)
		}
		if signerKeyInfo.Fingerprint != att.Issuer.KeyFingerprint {
			return types.ErrInvalidSignerKey.Wrap("signer key fingerprint does not match attestation issuer")
		}
		if signerKeyInfo.Algorithm != types.ProofTypeEd25519 {
			return types.ErrInvalidSignerKey.Wrapf("unsupported signer key algorithm: %s", signerKeyInfo.Algorithm)
		}

		canonicalBytes, err := att.CanonicalBytes()
		if err != nil {
			return types.ErrInvalidAttestation.Wrapf("failed to build canonical bytes: %v", err)
		}
		signatureBytes, err := att.GetProofBytes()
		if err != nil {
			return types.ErrInvalidAttestation.Wrapf("failed to decode signature: %v", err)
		}
		if !ed25519.Verify(signerKeyInfo.PublicKey, canonicalBytes, signatureBytes) {
			return types.ErrAttestationSignatureInvalid.Wrap("attestation signature invalid")
		}
	}

	// 3. Mark nonce as used (prevent replay)
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

	// 4. Emit event (pattern: EventVerificationSubmitted)
	if err := k.EmitVerificationSubmittedEvent(
		ctx,
		att.LinkedAccountAddress,
		att.OIDCIssuer+":"+att.SubjectHash,
		string(att.ProviderType),
		att.OIDCNonce,
	); err != nil {
		return err
	}

	// 5. Write audit log (pattern: SetAuditEntry)
	details := map[string]interface{}{
		"oidc_issuer": att.OIDCIssuer,
		"subject_hash": att.SubjectHash,
		"provider_type": string(att.ProviderType),
		"nonce_hash": nonceHash,
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
