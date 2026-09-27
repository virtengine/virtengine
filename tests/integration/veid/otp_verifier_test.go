package veid

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

// The MFA msg server deliberately refuses to mint TOTP/SMS/email challenges
// (msgServer.CreateChallenge) and requires a verifier-signed receipt instead
// (verifyOTPVerifierReceipt). These helpers reproduce the sanctioned flow so
// the integration tests exercise the real production path rather than a
// shortcut the chain would reject.

const (
	otpVerifierReceiptVersion = uint32(1)
	otpVerifierKeyEpoch       = uint64(1)
	otpVerifierReceiptDomain  = "virtengine/mfa/otp-verifier-receipt/v1\x00"
)

func deterministicOTPVerifierKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := sha256.Sum256([]byte("virtengine-mfa-otp-verifier-test-key"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	return privateKey.Public().(ed25519.PublicKey), privateKey
}

// enrollOTPVerifier registers an email/SMS/TOTP verifier enrollment whose
// PublicIdentifier is a real 32-byte Ed25519 key, which is what the production
// verification path requires of an OTP verifier.
func enrollOTPVerifier(
	t *testing.T,
	ctx sdk.Context,
	k mfakeeper.Keeper,
	address sdk.AccAddress,
	factorType mfatypes.FactorType,
	factorID string,
) {
	t.Helper()

	publicKey, _ := deterministicOTPVerifierKey()
	require.NoError(t, k.EnrollFactor(ctx, &mfatypes.FactorEnrollment{
		AccountAddress:   address.String(),
		FactorType:       factorType,
		FactorID:         factorID,
		PublicIdentifier: publicKey,
		Status:           mfatypes.EnrollmentStatusActive,
		EnrolledAt:       ctx.BlockTime().Unix(),
		VerifiedAt:       ctx.BlockTime().Unix(),
	}))
}

// signedOTPChallengeResponse builds the verifier-signed receipt the chain
// expects for an OTP challenge response.
func signedOTPChallengeResponse(t *testing.T, ctx sdk.Context, challenge *mfatypes.Challenge) *mfatypes.ChallengeResponse {
	t.Helper()

	digest := sha256.Sum256(challenge.ChallengeData)
	receipt := mfakeeper.OTPVerifierReceipt{
		Version:             otpVerifierReceiptVersion,
		ChainID:             ctx.ChainID(),
		AccountAddress:      challenge.AccountAddress,
		ChallengeID:         challenge.ChallengeID,
		FactorType:          challenge.FactorType,
		FactorID:            challenge.FactorID,
		TransactionType:     challenge.TransactionType,
		ChallengeNonce:      challenge.Nonce,
		ChallengeDataDigest: digest[:],
		IssuedAt:            challenge.CreatedAt,
		ExpiresAt:           challenge.ExpiresAt,
		VerifierKeyEpoch:    otpVerifierKeyEpoch,
		ReceiptNonce:        "receipt-" + challenge.ChallengeID,
	}
	if challenge.Metadata != nil && challenge.Metadata.OTPInfo != nil {
		receipt.DeliveryMethod = challenge.Metadata.OTPInfo.DeliveryMethod
		receipt.DeliveryID = challenge.Metadata.OTPInfo.DeliveryDestinationMasked
	}

	signBytes, err := receipt.SignBytes()
	require.NoError(t, err)

	_, privateKey := deterministicOTPVerifierKey()
	receipt.Signature = ed25519.Sign(privateKey, signBytes)

	responseData, err := json.Marshal(receipt)
	require.NoError(t, err)

	return &mfatypes.ChallengeResponse{
		ChallengeID:  challenge.ChallengeID,
		FactorType:   challenge.FactorType,
		ResponseData: responseData,
		Timestamp:    receipt.IssuedAt,
	}
}
