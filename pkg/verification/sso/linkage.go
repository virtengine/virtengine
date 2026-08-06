package sso

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func (s *DefaultService) verifyChallengeLinkageSignature(ctx context.Context, challenge *Challenge, signature []byte) error {
	if !s.config.RequireLinkageSignature {
		return nil
	}

	if s.chainClient == nil {
		return fmt.Errorf("%w: linkage signature verification requires chain client", ErrChainClientNotConfigured)
	}

	binding, err := s.chainClient.GetWalletBinding(ctx, challenge.AccountAddress)
	if err != nil {
		return fmt.Errorf("%w: failed to load wallet binding: %v", ErrInvalidLinkageSignature, err)
	}
	if binding == nil || len(binding.BindingPubKey) == 0 {
		return fmt.Errorf("%w: wallet binding not found for account %s", ErrInvalidLinkageSignature, challenge.AccountAddress)
	}

	return verifyDetachedWalletSignature(binding.BindingPubKey, []byte(challenge.LinkageMessage), signature)
}

func verifyDetachedWalletSignature(pubKey, message, signature []byte) error {
	if len(pubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: invalid binding public key size", ErrInvalidLinkageSignature)
	}
	if len(signature) != ed25519.SignatureSize {
		return fmt.Errorf("%w: invalid linkage signature size", ErrInvalidLinkageSignature)
	}

	hashedMessage := hashLinkageMessage(message)
	if !ed25519.Verify(pubKey, hashedMessage, signature) {
		return fmt.Errorf("%w: detached signature verification failed", ErrInvalidLinkageSignature)
	}

	return nil
}

func hashLinkageMessage(message []byte) []byte {
	if len(message) == 32 {
		hashed := make([]byte, 32)
		copy(hashed, message)
		return hashed
	}

	sum := sha256.Sum256(message)
	return sum[:]
}

func linkageStatusFromRecord(accountAddress string, linkage *veidtypes.SSOLinkageMetadata) *LinkageStatus {
	status := &LinkageStatus{
		Exists:            true,
		LinkageID:         linkage.LinkageID,
		AccountAddress:    accountAddress,
		ProviderType:      linkage.Provider,
		Issuer:            linkage.Issuer,
		Status:            linkage.Status,
		ExpiresAt:         linkage.ExpiresAt,
		ScoreContribution: veidtypes.GetSSOScoringWeight(linkage.Provider),
	}

	if !linkage.VerifiedAt.IsZero() {
		verifiedAt := linkage.VerifiedAt
		status.VerifiedAt = &verifiedAt
	}

	return status
}
