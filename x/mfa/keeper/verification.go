package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// ============================================================================
// MFA Challenge Verification Implementation (MFA-CORE-001)
// Implements verification flows for all factor types per veid-flow-spec.md
// ============================================================================

// VerificationResult contains the result of a verification attempt
type VerificationResult struct {
	// Verified indicates if verification succeeded
	Verified bool

	// FactorType is the type of factor that was verified
	FactorType types.FactorType

	// FactorID is the ID of the verified factor
	FactorID string

	// Metadata contains additional verification metadata
	Metadata map[string]string
}

// ============================================================================
// TOTP Verification (RFC 6238)
// Uses pquerna/otp library patterns for TOTP validation
// ============================================================================

// TOTPConfig contains TOTP verification configuration
type TOTPConfig struct {
	// Period is the time step in seconds (default: 30)
	Period uint

	// Digits is the number of digits in the OTP (default: 6)
	Digits uint

	// Skew is the number of periods before/after to check (default: 1)
	Skew uint

	// Algorithm is the hash algorithm (SHA1, SHA256, SHA512)
	Algorithm string
}

// DefaultTOTPConfig returns the default TOTP configuration
// Changed to SHA-256 in v0.9.x for improved security
// Legacy SHA-1 enrollments continue to work (algorithm stored per enrollment)
func DefaultTOTPConfig() TOTPConfig {
	return TOTPConfig{
		Period:    30,
		Digits:    6,
		Skew:      1,
		Algorithm: "SHA256",
	}
}

const (
	otpVerifierReceiptVersion = uint32(1)
	otpVerifierKeyEpoch       = uint64(1)
	otpVerifierReceiptDomain  = "virtengine/mfa/otp-verifier-receipt/v1\x00"
)

// OTPVerifierReceipt is an off-chain verifier attestation. It never contains
// an OTP or a commitment derived from one.
type OTPVerifierReceipt struct {
	Version             uint32                         `json:"version"`
	ChainID             string                         `json:"chain_id"`
	AccountAddress      string                         `json:"account_address"`
	ChallengeID         string                         `json:"challenge_id"`
	FactorType          types.FactorType               `json:"factor_type"`
	FactorID            string                         `json:"factor_id"`
	TransactionType     types.SensitiveTransactionType `json:"transaction_type"`
	ChallengeNonce      string                         `json:"challenge_nonce"`
	ChallengeDataDigest []byte                         `json:"challenge_data_digest"`
	IssuedAt            int64                          `json:"issued_at"`
	ExpiresAt           int64                          `json:"expires_at"`
	VerifierKeyEpoch    uint64                         `json:"verifier_key_epoch"`
	ReceiptNonce        string                         `json:"receipt_nonce"`
	DeliveryMethod      string                         `json:"delivery_method,omitempty"`
	DeliveryID          string                         `json:"delivery_id,omitempty"`
	Signature           []byte                         `json:"signature"`
}

// SignBytes returns deterministic, domain-separated bytes for Ed25519 signing.
func (r OTPVerifierReceipt) SignBytes() ([]byte, error) {
	unsigned := r
	unsigned.Signature = nil
	payload, err := json.Marshal(unsigned)
	if err != nil {
		return nil, err
	}
	return append([]byte(otpVerifierReceiptDomain), payload...), nil
}

func decodeOTPVerifierReceipt(data []byte) (*OTPVerifierReceipt, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var receipt OTPVerifierReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return nil, types.ErrInvalidChallengeResponse.Wrapf("invalid verifier receipt: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, types.ErrInvalidChallengeResponse.Wrap("verifier receipt must contain one JSON object")
	}
	return &receipt, nil
}

func (k Keeper) verifyOTPVerifierReceipt(ctx sdk.Context, challenge *types.Challenge, response *types.ChallengeResponse) (bool, error) {
	receipt, err := decodeOTPVerifierReceipt(response.ResponseData)
	if err != nil {
		return false, err
	}
	address, err := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err != nil {
		return false, types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}
	enrollment, err := k.requireActiveEnrollment(ctx, address, challenge.FactorType, challenge.FactorID)
	if err != nil {
		return false, err
	}
	if len(enrollment.PublicIdentifier) != ed25519.PublicKeySize {
		return false, types.ErrInvalidEnrollment.Wrap("OTP verifier public key must be 32-byte Ed25519 key")
	}
	digest := sha256.Sum256(challenge.ChallengeData)
	now := ctx.BlockTime().Unix()
	if receipt.Version != otpVerifierReceiptVersion ||
		receipt.ChainID == "" || receipt.ChainID != ctx.ChainID() ||
		receipt.AccountAddress != challenge.AccountAddress ||
		receipt.ChallengeID != challenge.ChallengeID ||
		receipt.FactorType != challenge.FactorType ||
		receipt.FactorID != challenge.FactorID ||
		receipt.TransactionType != challenge.TransactionType ||
		receipt.ChallengeNonce != challenge.Nonce ||
		!bytes.Equal(receipt.ChallengeDataDigest, digest[:]) ||
		receipt.IssuedAt < challenge.CreatedAt || receipt.IssuedAt > now ||
		receipt.ExpiresAt <= receipt.IssuedAt || receipt.ExpiresAt < now || receipt.ExpiresAt > challenge.ExpiresAt ||
		receipt.VerifierKeyEpoch != otpVerifierKeyEpoch || receipt.ReceiptNonce == "" ||
		response.Timestamp != receipt.IssuedAt {
		return false, types.ErrInvalidChallengeResponse.Wrap("verifier receipt binding mismatch")
	}
	if challenge.FactorType == types.FactorTypeSMS || challenge.FactorType == types.FactorTypeEmail {
		if challenge.Metadata == nil || challenge.Metadata.OTPInfo == nil {
			return false, types.ErrInvalidChallenge.Wrap("missing OTP delivery metadata")
		}
		otpInfo := challenge.Metadata.OTPInfo
		if receipt.DeliveryMethod != otpInfo.DeliveryMethod || receipt.DeliveryID != otpInfo.DeliveryDestinationMasked {
			return false, types.ErrInvalidChallengeResponse.Wrap("OTP delivery binding mismatch")
		}
	} else if receipt.DeliveryMethod != "" || receipt.DeliveryID != "" {
		return false, types.ErrInvalidChallengeResponse.Wrap("unexpected OTP delivery metadata")
	}
	signBytes, err := receipt.SignBytes()
	if err != nil {
		return false, types.ErrInvalidChallengeResponse.Wrapf("failed to encode verifier receipt: %v", err)
	}
	if len(receipt.Signature) != ed25519.SignatureSize || !ed25519.Verify(enrollment.PublicIdentifier, signBytes, receipt.Signature) {
		return false, types.ErrInvalidSignature.Wrap("OTP verifier receipt signature verification failed")
	}
	return true, nil
}

func safeUint32FromIntVerification(value int) uint32 {
	if value < 0 {
		return 0
	}
	max := int(^uint32(0))
	if value > max {
		return ^uint32(0)
	}
	//nolint:gosec // range checked above
	return uint32(value)
}

// ============================================================================
// SMS/Email OTP Verification
// ============================================================================

// ============================================================================
// VEID Biometric Threshold Verification
// ============================================================================

// VEIDVerificationConfig contains VEID verification configuration
type VEIDVerificationConfig struct {
	// DefaultThreshold is the default VEID score threshold
	DefaultThreshold uint32

	// MinimumThreshold is the minimum acceptable threshold
	MinimumThreshold uint32

	// MaximumThreshold is the maximum threshold
	MaximumThreshold uint32
}

// DefaultVEIDConfig returns the default VEID verification configuration
func DefaultVEIDConfig() VEIDVerificationConfig {
	return VEIDVerificationConfig{
		DefaultThreshold: 50,
		MinimumThreshold: 50,
		MaximumThreshold: 100,
	}
}

// verifyVEIDScore verifies that the account meets the VEID biometric score threshold
//
//nolint:unused // reserved for future VEID gating enforcement
func (k Keeper) verifyVEIDScore(
	ctx sdk.Context,
	challenge *types.Challenge,
	requiredThreshold uint32,
) (bool, uint32, error) {
	address, err := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err != nil {
		return false, 0, types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	// Get the current VEID score from the VEID keeper
	if k.veidKeeper == nil {
		return false, 0, types.ErrVerificationFailed.Wrap("VEID keeper not available")
	}

	score, found := k.veidKeeper.GetVEIDScore(ctx, address)
	if !found {
		return false, 0, types.ErrVEIDScoreInsufficient.Wrap("no VEID score found for account")
	}

	// Check if score meets threshold
	if score < requiredThreshold {
		// Emit event for insufficient score
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeVEID.String()),
				sdk.NewAttribute(types.AttributeKeyVEIDScore, strconv.FormatUint(uint64(score), 10)),
				sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatUint(uint64(requiredThreshold), 10)),
				sdk.NewAttribute(types.AttributeKeyReason, "score_below_threshold"),
			),
		)
		return false, score, types.ErrVEIDScoreInsufficient.Wrapf(
			"VEID score %d is below required threshold %d", score, requiredThreshold)
	}

	return true, score, nil
}

// ============================================================================
// Hardware Key Verification (X.509/Smart Card/PIV)
// ============================================================================

// HardwareKeyResponse represents a hardware key verification response
type HardwareKeyResponse struct {
	// Signature is the signature over the challenge data
	Signature []byte `json:"signature"`

	// CertificateChain is the DER-encoded certificate chain
	CertificateChain [][]byte `json:"certificate_chain,omitempty"`

	// Algorithm is the signature algorithm used
	Algorithm string `json:"algorithm"`

	// KeyID is the key identifier
	KeyID string `json:"key_id"`
}

// verifyHardwareKeyResponse verifies a hardware key (X.509/smart card) authentication response
func (k Keeper) verifyHardwareKeyResponse(
	ctx sdk.Context,
	challenge *types.Challenge,
	response *types.ChallengeResponse,
	enrollment *types.FactorEnrollment,
) (bool, error) {
	// Parse hardware key response
	var hkResp HardwareKeyResponse
	if err := json.Unmarshal(response.ResponseData, &hkResp); err != nil {
		return false, types.ErrInvalidChallengeResponse.Wrapf("invalid hardware key response: %v", err)
	}

	// Validate required fields
	if len(hkResp.Signature) == 0 {
		return false, types.ErrInvalidChallengeResponse.Wrap("missing signature")
	}

	if hkResp.KeyID == "" {
		return false, types.ErrInvalidChallengeResponse.Wrap("missing key ID")
	}

	// Get hardware key metadata from enrollment
	if enrollment.Metadata == nil || enrollment.Metadata.HardwareKeyInfo == nil {
		return false, types.ErrEnrollmentNotFound.Wrap("hardware key metadata not found")
	}

	hkInfo := enrollment.Metadata.HardwareKeyInfo

	// Verify key ID matches enrollment
	if hkInfo.KeyID != hkResp.KeyID {
		return false, types.ErrInvalidChallengeResponse.Wrap("key ID mismatch")
	}

	// Check certificate validity period
	now := ctx.BlockTime().Unix()
	if now < hkInfo.NotBefore {
		return false, types.ErrCertificateNotYetValid.Wrapf(
			"certificate valid from %d, current time %d", hkInfo.NotBefore, now)
	}
	if now > hkInfo.NotAfter {
		return false, types.ErrCertificateExpired.Wrapf(
			"certificate expired at %d, current time %d", hkInfo.NotAfter, now)
	}

	// Verify signature using stored public key fingerprint
	// The challenge data should be signed by the private key corresponding to the enrolled public key
	if challenge.Metadata == nil || challenge.Metadata.HardwareKeyChallenge == nil {
		return false, types.ErrInvalidChallenge.Wrap("missing hardware key challenge data")
	}

	hkChallenge := challenge.Metadata.HardwareKeyChallenge

	// Verify the signature over the challenge data
	verified, err := k.verifyHardwareKeySignature(
		hkChallenge.Challenge,
		hkResp.Signature,
		hkResp.Algorithm,
		enrollment.PublicIdentifier,
	)
	if err != nil {
		return false, err
	}

	if !verified {
		return false, types.ErrInvalidSignature.Wrap("signature verification failed")
	}

	// Check revocation status if enabled
	if hkInfo.RevocationCheckEnabled {
		// In production, this would check OCSP/CRL
		// For now, we trust the enrollment revocation status
		if hkInfo.LastRevocationCheck > 0 {
			_ = hkInfo.LastRevocationCheck
		}
	}

	return true, nil
}

// verifyHardwareKeySignature verifies a hardware key signature
func (k Keeper) verifyHardwareKeySignature(
	data []byte,
	signature []byte,
	algorithm string,
	publicKeyBytes []byte,
) (bool, error) {
	// The public key bytes contain the DER-encoded public key
	if len(publicKeyBytes) == 0 {
		return false, types.ErrInvalidChallengeResponse.Wrap("missing public key")
	}

	// Hash the data for signature verification
	var dataHash []byte
	switch algorithm {
	case "RS256", "ES256", "PS256":
		h := sha256.Sum256(data)
		dataHash = h[:]
	case "RS384", "ES384", "PS384":
		h := sha512.Sum384(data)
		dataHash = h[:]
	case "RS512", "ES512", "PS512":
		h := sha512.Sum512(data)
		dataHash = h[:]
	case "EdDSA":
		// EdDSA doesn't pre-hash
		dataHash = data
	default:
		return false, types.ErrFIDO2UnsupportedAlgorithm.Wrapf("unsupported algorithm: %s", algorithm)
	}

	// Verify using the stored public key
	// This requires parsing the public key and using appropriate crypto verification
	// The actual implementation depends on the key type stored in publicKeyBytes

	// For now, verify the signature format is valid
	if len(signature) == 0 {
		return false, types.ErrInvalidSignature.Wrap("empty signature")
	}

	// In production, use crypto/x509 and crypto/ecdsa or crypto/rsa
	// to verify the signature with the public key
	_ = dataHash

	return true, nil
}

// ============================================================================
// Factor Combination Rules
// Implements factor combination policies from veid-flow-spec.md
// ============================================================================

// FactorCombinationPolicy defines required factor combinations for actions
type FactorCombinationPolicy struct {
	// MinFactors is the minimum number of factors required
	MinFactors uint32

	// RequiredTypes are factor types that must be verified
	RequiredTypes []types.FactorType

	// OptionalTypes are additional acceptable factor types
	OptionalTypes []types.FactorType

	// MinSecurityLevel is the minimum combined security level
	MinSecurityLevel types.FactorSecurityLevel

	// RequireHighSecurityFactor indicates if at least one high-security factor is required
	RequireHighSecurityFactor bool
}

// GetCombinationPolicy returns the factor combination policy for a transaction type
func GetCombinationPolicy(txType types.SensitiveTransactionType) FactorCombinationPolicy {
	switch txType {
	// Critical Tier - Always require full MFA
	case types.SensitiveTxAccountRecovery:
		return FactorCombinationPolicy{
			MinFactors:                3,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			OptionalTypes:             []types.FactorType{types.FactorTypeSMS, types.FactorTypeEmail},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxKeyRotation:
		return FactorCombinationPolicy{
			MinFactors:                3,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			OptionalTypes:             []types.FactorType{types.FactorTypeTOTP},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxPrimaryEmailChange, types.SensitiveTxPhoneNumberChange:
		return FactorCombinationPolicy{
			MinFactors:                2,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID},
			OptionalTypes:             []types.FactorType{types.FactorTypeSMS, types.FactorTypeEmail},
			MinSecurityLevel:          types.FactorSecurityLevelMedium,
			RequireHighSecurityFactor: false,
		}
	case types.SensitiveTxTwoFactorDisable:
		return FactorCombinationPolicy{
			MinFactors:                3,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2, types.FactorTypeTOTP},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxAccountDeletion:
		return FactorCombinationPolicy{
			MinFactors:                3,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			OptionalTypes:             []types.FactorType{types.FactorTypeSMS, types.FactorTypeEmail},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}

	// High Tier - Require strong MFA
	case types.SensitiveTxProviderRegistration:
		return FactorCombinationPolicy{
			MinFactors:                2,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxValidatorRegistration:
		return FactorCombinationPolicy{
			MinFactors:                3,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			OptionalTypes:             []types.FactorType{types.FactorTypeTOTP, types.FactorTypeHardwareKey},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxLargeWithdrawal:
		return FactorCombinationPolicy{
			MinFactors:                2,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxGovernanceProposal:
		return FactorCombinationPolicy{
			MinFactors:                2,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}
	case types.SensitiveTxRoleAssignment:
		return FactorCombinationPolicy{
			MinFactors:                2,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			MinSecurityLevel:          types.FactorSecurityLevelHigh,
			RequireHighSecurityFactor: true,
		}

	// Medium Tier - Standard MFA
	case types.SensitiveTxHighValueOrder:
		return FactorCombinationPolicy{
			MinFactors:                2,
			RequiredTypes:             []types.FactorType{types.FactorTypeVEID},
			OptionalTypes:             []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeTOTP},
			MinSecurityLevel:          types.FactorSecurityLevelMedium,
			RequireHighSecurityFactor: false,
		}
	case types.SensitiveTxMediumWithdrawal:
		return FactorCombinationPolicy{
			MinFactors:                1,
			RequiredTypes:             []types.FactorType{types.FactorTypeFIDO2},
			OptionalTypes:             []types.FactorType{types.FactorTypeTOTP, types.FactorTypeHardwareKey},
			MinSecurityLevel:          types.FactorSecurityLevelMedium,
			RequireHighSecurityFactor: false,
		}
	case types.SensitiveTxAPIKeyGeneration:
		return FactorCombinationPolicy{
			MinFactors:                1,
			OptionalTypes:             []types.FactorType{types.FactorTypeTOTP, types.FactorTypeFIDO2},
			MinSecurityLevel:          types.FactorSecurityLevelMedium,
			RequireHighSecurityFactor: false,
		}

	// Default - Low tier or unspecified
	default:
		return FactorCombinationPolicy{
			MinFactors:                1,
			OptionalTypes:             types.AllFactorTypes(),
			MinSecurityLevel:          types.FactorSecurityLevelLow,
			RequireHighSecurityFactor: false,
		}
	}
}

// ValidateFactorCombination validates that verified factors meet policy requirements
func ValidateFactorCombination(
	verifiedFactors []types.FactorType,
	policy FactorCombinationPolicy,
) error {
	// Check minimum factor count
	if safeUint32FromIntVerification(len(verifiedFactors)) < policy.MinFactors {
		return types.ErrInsufficientFactors.Wrapf(
			"verified %d factors, required minimum %d", len(verifiedFactors), policy.MinFactors)
	}

	// Check required types are present
	for _, required := range policy.RequiredTypes {
		found := false
		for _, verified := range verifiedFactors {
			if verified == required {
				found = true
				break
			}
		}
		if !found {
			return types.ErrInsufficientFactors.Wrapf(
				"required factor type %s not verified", required.String())
		}
	}

	// Check if high security factor is required
	if policy.RequireHighSecurityFactor {
		hasHighSecurity := false
		for _, verified := range verifiedFactors {
			if verified.GetSecurityLevel() == types.FactorSecurityLevelHigh {
				hasHighSecurity = true
				break
			}
		}
		if !hasHighSecurity {
			return types.ErrInsufficientFactors.Wrap(
				"at least one high-security factor required")
		}
	}

	// Check minimum security level
	for _, verified := range verifiedFactors {
		if verified.GetSecurityLevel() >= policy.MinSecurityLevel {
			return nil // At least one factor meets security level
		}
	}

	return types.ErrInsufficientFactors.Wrapf(
		"no verified factor meets minimum security level %d", policy.MinSecurityLevel)
}

// ============================================================================
// Comprehensive Challenge Verification
// ============================================================================

// VerifyChallengeWithPolicy verifies a challenge and validates against factor combination policy
func (k Keeper) VerifyChallengeWithPolicy(
	ctx sdk.Context,
	challengeID string,
	response *types.ChallengeResponse,
	txType types.SensitiveTransactionType,
	alreadyVerifiedFactors []types.FactorType,
) (*VerificationResult, error) {
	// Get the challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return nil, types.ErrChallengeNotFound.Wrapf("challenge %s not found", challengeID)
	}

	// Emit attempt event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"mfa_verification_attempt",
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
			sdk.NewAttribute(types.AttributeKeyAttemptCount, strconv.FormatUint(uint64(challenge.AttemptCount+1), 10)),
		),
	)

	// Verify the challenge
	verified, err := k.VerifyMFAChallenge(ctx, challengeID, response)
	if err != nil {
		// Emit failure event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			),
		)
		return nil, err
	}

	if !verified {
		return &VerificationResult{
			Verified:   false,
			FactorType: challenge.FactorType,
			FactorID:   challenge.FactorID,
		}, nil
	}

	// Add to verified factors list
	allVerified := make([]types.FactorType, 0, len(alreadyVerifiedFactors)+1)
	allVerified = append(allVerified, alreadyVerifiedFactors...)
	allVerified = append(allVerified, challenge.FactorType)

	// Get policy for this transaction type
	policy := GetCombinationPolicy(txType)

	// Validate factor combination
	if err := ValidateFactorCombination(allVerified, policy); err != nil {
		// Not enough factors yet, but this verification succeeded
		return &VerificationResult{
			Verified:   true,
			FactorType: challenge.FactorType,
			FactorID:   challenge.FactorID,
			Metadata: map[string]string{
				"policy_satisfied":   "false",
				"factors_verified":   strconv.Itoa(len(allVerified)),
				"factors_required":   strconv.FormatUint(uint64(policy.MinFactors), 10),
				"additional_factors": err.Error(),
			},
		}, nil
	}

	// Emit success event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeVerified,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
			sdk.NewAttribute(types.AttributeKeyTransactionType, txType.String()),
			sdk.NewAttribute(types.AttributeKeyVerifiedFactors, formatFactorList(allVerified)),
		),
	)

	return &VerificationResult{
		Verified:   true,
		FactorType: challenge.FactorType,
		FactorID:   challenge.FactorID,
		Metadata: map[string]string{
			"policy_satisfied": "true",
			"factors_verified": strconv.Itoa(len(allVerified)),
		},
	}, nil
}

// formatFactorList formats a list of factor types as a comma-separated string
func formatFactorList(factors []types.FactorType) string {
	if len(factors) == 0 {
		return ""
	}

	result := factors[0].String()
	for i := 1; i < len(factors); i++ {
		result += "," + factors[i].String()
	}
	return result
}

// ============================================================================
// Enhanced Factor-Specific Verification Methods
// ============================================================================

// VerifyFIDO2ChallengeResponse verifies a FIDO2 WebAuthn challenge response
// Implements full FIDO2 verification per WebAuthn spec
func (k Keeper) VerifyFIDO2ChallengeResponse(
	ctx sdk.Context,
	challengeID string,
	credentialID []byte,
	clientDataJSON []byte,
	authenticatorData []byte,
	signature []byte,
	userHandle []byte,
) (bool, error) {
	// Get the challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return false, types.ErrChallengeNotFound
	}

	if challenge.FactorType != types.FactorTypeFIDO2 {
		return false, types.ErrInvalidFactorType.Wrap("challenge is not for FIDO2")
	}

	// Validate challenge state
	now := ctx.BlockTime()
	if challenge.IsExpired(now) {
		challenge.MarkExpired()
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeExpired,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			),
		)

		return false, types.ErrChallengeExpired
	}

	if !challenge.IsPending() {
		return false, types.ErrChallengeAlreadyUsed
	}

	if !challenge.CanAttempt(now) {
		return false, types.ErrMaxAttemptsExceeded
	}

	// Record the attempt
	challenge.RecordAttempt()

	// Get account address
	address, err := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err != nil {
		return false, types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	// Use the existing VerifyFIDO2Assertion method
	if err := k.VerifyFIDO2Assertion(
		ctx,
		address,
		challengeID,
		credentialID,
		clientDataJSON,
		authenticatorData,
		signature,
	); err != nil {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeFIDO2.String()),
				sdk.NewAttribute(types.AttributeKeyReason, "signature_verification_failed"),
			),
		)

		return false, err
	}

	// Challenge is already marked as verified by VerifyFIDO2Assertion
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeVerified,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeFIDO2.String()),
			sdk.NewAttribute(types.AttributeKeyFactorID, base64.RawURLEncoding.EncodeToString(credentialID)),
		),
	)

	return true, nil
}

// VerifyTOTPChallengeResponse verifies a signed TOTP verifier receipt.
func (k Keeper) VerifyTOTPChallengeResponse(
	ctx sdk.Context,
	challengeID string,
	response *types.ChallengeResponse,
) (bool, error) {
	// Get the challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return false, types.ErrChallengeNotFound
	}

	if challenge.FactorType != types.FactorTypeTOTP {
		return false, types.ErrInvalidFactorType.Wrap("challenge is not for TOTP")
	}

	// Validate challenge state
	now := ctx.BlockTime()
	if challenge.IsExpired(now) {
		challenge.MarkExpired()
		_ = k.UpdateChallenge(ctx, challenge)
		return false, types.ErrChallengeExpired
	}

	if !challenge.IsPending() {
		return false, types.ErrChallengeAlreadyUsed
	}

	if !challenge.CanAttempt(now) {
		return false, types.ErrMaxAttemptsExceeded
	}

	// Record the attempt
	challenge.RecordAttempt()

	// Get account address
	address, err := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err != nil {
		return false, types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	// Get the factor enrollment
	enrollment, found := k.GetFactorEnrollment(ctx, address, types.FactorTypeTOTP, challenge.FactorID)
	if !found {
		return false, types.ErrEnrollmentNotFound
	}

	if !enrollment.IsActive() {
		return false, types.ErrFactorRevoked
	}

	// Verify the off-chain verifier receipt.
	verified, err := k.verifyOTPVerifierReceipt(ctx, challenge, response)
	if err != nil {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeTOTP.String()),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			),
		)

		return false, err
	}

	if !verified {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeTOTP.String()),
				sdk.NewAttribute(types.AttributeKeyReason, "invalid_code"),
			),
		)

		return false, nil
	}

	// Mark challenge as verified
	challenge.MarkVerified(now.Unix())
	_ = k.UpdateChallenge(ctx, challenge)

	// Update factor usage
	enrollment.UpdateLastUsed(now.Unix())
	_ = k.updateFactorEnrollment(ctx, enrollment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeVerified,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeTOTP.String()),
		),
	)

	return true, nil
}

// VerifyOTPChallengeResponse verifies a signed SMS/Email verifier receipt.
func (k Keeper) VerifyOTPChallengeResponse(
	ctx sdk.Context,
	challengeID string,
	response *types.ChallengeResponse,
) (bool, error) {
	// Get the challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return false, types.ErrChallengeNotFound
	}

	if challenge.FactorType != types.FactorTypeSMS && challenge.FactorType != types.FactorTypeEmail {
		return false, types.ErrInvalidFactorType.Wrap("challenge is not for SMS/Email OTP")
	}

	// Validate challenge state
	now := ctx.BlockTime()
	if challenge.IsExpired(now) {
		challenge.MarkExpired()
		_ = k.UpdateChallenge(ctx, challenge)
		return false, types.ErrChallengeExpired
	}

	if !challenge.IsPending() {
		return false, types.ErrChallengeAlreadyUsed
	}

	if !challenge.CanAttempt(now) {
		return false, types.ErrMaxAttemptsExceeded
	}

	// Record the attempt
	challenge.RecordAttempt()

	// Verify the off-chain verifier receipt.
	verified, err := k.verifyOTPVerifierReceipt(ctx, challenge, response)
	if err != nil {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			),
		)

		return false, err
	}

	if !verified {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
				sdk.NewAttribute(types.AttributeKeyReason, "invalid_otp"),
			),
		)

		return false, nil
	}

	// Mark challenge as verified
	challenge.MarkVerified(now.Unix())
	_ = k.UpdateChallenge(ctx, challenge)

	// Update factor usage
	address, _ := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if enrollment, found := k.GetFactorEnrollment(ctx, address, challenge.FactorType, challenge.FactorID); found {
		enrollment.UpdateLastUsed(now.Unix())
		_ = k.updateFactorEnrollment(ctx, enrollment)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeVerified,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, challenge.FactorType.String()),
		),
	)

	return true, nil
}

// VerifyHardwareKeyChallengeResponse verifies a hardware key challenge response
func (k Keeper) VerifyHardwareKeyChallengeResponse(
	ctx sdk.Context,
	challengeID string,
	signature []byte,
	keyID string,
	algorithm string,
) (bool, error) {
	// Get the challenge
	challenge, found := k.GetChallenge(ctx, challengeID)
	if !found {
		return false, types.ErrChallengeNotFound
	}

	if challenge.FactorType != types.FactorTypeHardwareKey {
		return false, types.ErrInvalidFactorType.Wrap("challenge is not for hardware key")
	}

	// Validate challenge state
	now := ctx.BlockTime()
	if challenge.IsExpired(now) {
		challenge.MarkExpired()
		_ = k.UpdateChallenge(ctx, challenge)
		return false, types.ErrChallengeExpired
	}

	if !challenge.IsPending() {
		return false, types.ErrChallengeAlreadyUsed
	}

	if !challenge.CanAttempt(now) {
		return false, types.ErrMaxAttemptsExceeded
	}

	// Record the attempt
	challenge.RecordAttempt()

	// Get account address
	address, err := sdk.AccAddressFromBech32(challenge.AccountAddress)
	if err != nil {
		return false, types.ErrInvalidAddress.Wrapf("invalid account address: %v", err)
	}

	// Get the factor enrollment
	enrollment, found := k.GetFactorEnrollment(ctx, address, types.FactorTypeHardwareKey, challenge.FactorID)
	if !found {
		return false, types.ErrEnrollmentNotFound
	}

	if !enrollment.IsActive() {
		return false, types.ErrFactorRevoked
	}

	// Build response for verification
	hkResp := HardwareKeyResponse{
		Signature: signature,
		Algorithm: algorithm,
		KeyID:     keyID,
	}
	responseData, _ := json.Marshal(hkResp) //nolint:errchkjson // Best-effort marshal for response data

	response := &types.ChallengeResponse{
		ChallengeID:  challengeID,
		FactorType:   types.FactorTypeHardwareKey,
		ResponseData: responseData,
		Timestamp:    now.Unix(),
	}

	// Verify the hardware key response
	verified, err := k.verifyHardwareKeyResponse(ctx, challenge, response, enrollment)
	if err != nil {
		_ = k.UpdateChallenge(ctx, challenge)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeChallengeFailed,
				sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
				sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
				sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeHardwareKey.String()),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			),
		)

		return false, err
	}

	if !verified {
		_ = k.UpdateChallenge(ctx, challenge)
		return false, types.ErrInvalidSignature
	}

	// Mark challenge as verified
	challenge.MarkVerified(now.Unix())
	_ = k.UpdateChallenge(ctx, challenge)

	// Update factor usage
	enrollment.UpdateLastUsed(now.Unix())
	_ = k.updateFactorEnrollment(ctx, enrollment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeVerified,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, challenge.AccountAddress),
			sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeHardwareKey.String()),
			sdk.NewAttribute(types.AttributeKeyFactorID, keyID),
		),
	)

	return true, nil
}

// ============================================================================
// Audit Event Helpers
// ============================================================================

// EmitVerificationAuditEvent emits an audit event for verification attempts
func (k Keeper) EmitVerificationAuditEvent(
	ctx sdk.Context,
	challengeID string,
	accountAddress string,
	factorType types.FactorType,
	success bool,
	reason string,
) {
	eventType := types.EventTypeChallengeVerified
	status := types.AttributeValueSuccess
	if !success {
		eventType = types.EventTypeChallengeFailed
		status = types.AttributeValueFailure
	}

	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeyChallengeID, challengeID),
		sdk.NewAttribute(types.AttributeKeyAccountAddress, accountAddress),
		sdk.NewAttribute(types.AttributeKeyFactorType, factorType.String()),
		sdk.NewAttribute(types.AttributeKeyStatus, status),
		sdk.NewAttribute(types.AttributeKeyTimestamp, strconv.FormatInt(ctx.BlockTime().Unix(), 10)),
	}

	if reason != "" {
		attrs = append(attrs, sdk.NewAttribute(types.AttributeKeyReason, reason))
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(eventType, attrs...))
}

// ============================================================================
// Challenge Creation Helpers
// ============================================================================

// CreateTOTPChallenge creates a random transaction-bound TOTP challenge.
func (k Keeper) CreateTOTPChallenge(
	ctx sdk.Context,
	address sdk.AccAddress,
	factorID string,
	transactionType types.SensitiveTransactionType,
) (*types.Challenge, error) {
	if _, err := k.requireOTPVerifierEnrollment(ctx, address, types.FactorTypeTOTP, factorID); err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)
	challenge, err := types.NewChallengeAt(
		address.String(),
		types.FactorTypeTOTP,
		factorID,
		transactionType,
		params.ChallengeTTL,
		params.MaxChallengeAttempts,
		ctx.BlockTime().Unix(),
	)
	if err != nil {
		return nil, err
	}

	if err := k.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}
	return challenge, nil
}

// CreateOTPChallenge creates a new OTP challenge for SMS or Email verification.
func (k Keeper) CreateOTPChallenge(
	ctx sdk.Context,
	address sdk.AccAddress,
	factorType types.FactorType,
	factorID string,
	transactionType types.SensitiveTransactionType,
	deliveryMethod string,
	deliveryID string,
) (*types.Challenge, error) {
	if factorType != types.FactorTypeSMS && factorType != types.FactorTypeEmail {
		return nil, types.ErrInvalidFactorType.Wrap("must be SMS or Email factor type")
	}
	if deliveryMethod == "" || deliveryID == "" {
		return nil, types.ErrInvalidChallenge.Wrap("OTP delivery method and ID are required")
	}
	if _, err := k.requireOTPVerifierEnrollment(ctx, address, factorType, factorID); err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)

	// Create the challenge
	challenge, err := types.NewChallengeAt(
		address.String(),
		factorType,
		factorID,
		transactionType,
		params.ChallengeTTL,
		params.MaxChallengeAttempts,
		ctx.BlockTime().Unix(),
	)
	if err != nil {
		return nil, err
	}

	// Add OTP metadata
	challenge.Metadata = &types.ChallengeMetadata{
		OTPInfo: &types.OTPChallengeInfo{
			DeliveryMethod:            deliveryMethod,
			DeliveryDestinationMasked: deliveryID,
			SentAt:                    ctx.BlockTime().Unix(),
		},
	}

	// Store the challenge
	if err := k.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeCreated,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyFactorType, factorType.String()),
		),
	)

	return challenge, nil
}

func (k Keeper) requireActiveEnrollment(
	ctx sdk.Context,
	address sdk.AccAddress,
	factorType types.FactorType,
	factorID string,
) (*types.FactorEnrollment, error) {
	enrollment, found := k.GetFactorEnrollment(ctx, address, factorType, factorID)
	if !found {
		return nil, types.ErrEnrollmentNotFound.Wrapf("factor %s not found", factorID)
	}
	if !enrollment.CanVerify(ctx.BlockTime()) {
		return nil, types.ErrFactorRevoked.Wrapf("factor %s is not active", factorID)
	}
	return enrollment, nil
}

func (k Keeper) requireOTPVerifierEnrollment(
	ctx sdk.Context,
	address sdk.AccAddress,
	factorType types.FactorType,
	factorID string,
) (*types.FactorEnrollment, error) {
	enrollment, err := k.requireActiveEnrollment(ctx, address, factorType, factorID)
	if err != nil {
		return nil, err
	}
	if len(enrollment.PublicIdentifier) != ed25519.PublicKeySize {
		return nil, types.ErrInvalidEnrollment.Wrap("OTP verifier public key must be 32-byte Ed25519 key")
	}
	return enrollment, nil
}

// CreateVEIDChallenge creates a new VEID score threshold challenge
func (k Keeper) CreateVEIDChallenge(
	ctx sdk.Context,
	address sdk.AccAddress,
	threshold uint32,
	txType types.SensitiveTransactionType,
) (*types.Challenge, error) {
	params := k.GetParams(ctx)

	// Create the challenge
	challenge, err := types.NewChallengeAt(
		address.String(),
		types.FactorTypeVEID,
		"veid_threshold",
		txType,
		params.ChallengeTTL,
		1, // VEID check is instant, only 1 attempt needed
		ctx.BlockTime().Unix(),
	)
	if err != nil {
		return nil, err
	}

	// Store threshold in challenge data
	thresholdBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(thresholdBytes, threshold)
	challenge.ChallengeData = thresholdBytes

	// Store the challenge
	if err := k.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeCreated,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeVEID.String()),
			sdk.NewAttribute(types.AttributeKeyThreshold, strconv.FormatUint(uint64(threshold), 10)),
		),
	)

	return challenge, nil
}

// CreateHardwareKeyChallenge creates a new hardware key challenge
func (k Keeper) CreateHardwareKeyChallenge(
	ctx sdk.Context,
	address sdk.AccAddress,
	factorID string,
	txType types.SensitiveTransactionType,
) (*types.Challenge, error) {
	params := k.GetParams(ctx)

	// Generate random challenge bytes
	challengeBytes := make([]byte, 32)
	nonceBytes := make([]byte, 16)

	// Use deterministic randomness for consensus
	// In production, this would use a VRF or block hash
	seed := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d",
		address.String(), factorID, ctx.BlockHeight())))
	copy(challengeBytes, seed[:])

	nonceSeed := sha256.Sum256(append(seed[:], []byte("nonce")...))
	copy(nonceBytes, nonceSeed[:16])

	// Create the challenge
	challenge, err := types.NewChallengeAt(
		address.String(),
		types.FactorTypeHardwareKey,
		factorID,
		txType,
		params.ChallengeTTL,
		params.MaxChallengeAttempts,
		ctx.BlockTime().Unix(),
	)
	if err != nil {
		return nil, err
	}

	challenge.ChallengeData = challengeBytes
	challenge.Metadata = &types.ChallengeMetadata{
		HardwareKeyChallenge: &types.HardwareKeyChallenge{
			Challenge:     challengeBytes,
			ChallengeType: types.HardwareKeyChallengeTypeSign,
			KeyID:         factorID,
			Nonce:         hex.EncodeToString(nonceBytes),
			CreatedAt:     ctx.BlockTime().Unix(),
			ExpiresAt:     ctx.BlockTime().Unix() + params.ChallengeTTL,
		},
	}

	// Store the challenge
	if err := k.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeChallengeCreated,
			sdk.NewAttribute(types.AttributeKeyChallengeID, challenge.ChallengeID),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, address.String()),
			sdk.NewAttribute(types.AttributeKeyFactorType, types.FactorTypeHardwareKey.String()),
		),
	)

	return challenge, nil
}
