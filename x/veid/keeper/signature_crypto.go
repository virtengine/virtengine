// Package keeper provides the VEID module keeper.
//
// This file implements cryptographic signature verification for VEID scopes.
// Client signatures use Ed25519 (approved capture apps).
// User signatures use ECDSA secp256k1 (Cosmos wallet signatures).
//
// Task Reference: VE-3022 - Cryptographic Signature Verification
package keeper

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Supported Signature Algorithms
// ============================================================================

const (
	// AlgorithmEd25519 is the Ed25519 signature algorithm for client signatures
	AlgorithmEd25519 = "ed25519"

	// AlgorithmSecp256k1 is the secp256k1 signature algorithm for user (Cosmos) signatures
	AlgorithmSecp256k1 = "secp256k1"

	// Ed25519PublicKeySize is the size of an Ed25519 public key in bytes
	Ed25519PublicKeySize = ed25519.PublicKeySize // 32 bytes

	// Ed25519SignatureSize is the size of an Ed25519 signature in bytes
	Ed25519SignatureSize = ed25519.SignatureSize // 64 bytes

	// Secp256k1PublicKeySize is the size of a compressed secp256k1 public key
	Secp256k1PublicKeySize = 33 // compressed public key

	// Secp256k1SignatureSize is the size of a secp256k1 signature
	Secp256k1SignatureSize = 64 // R + S values

	// SaltBindingMaxAge is the maximum age of a salt binding timestamp (5 minutes)
	SaltBindingMaxAge = 5 * time.Minute

	// SaltBindingMaxFuture is the maximum time in the future for a timestamp (1 minute)
	SaltBindingMaxFuture = 1 * time.Minute

	// errFmtByteLength is a common error format string for byte length mismatches
	errFmtByteLength = "expected %d bytes, got %d bytes"
)

// ============================================================================
// Ed25519 Client Signature Verification
// ============================================================================

// VerifyEd25519Signature verifies an Ed25519 signature from an approved client.
// This is used for client (capture app) signatures.
//
// Parameters:
//   - pubKey: The Ed25519 public key (32 bytes)
//   - message: The message that was signed
//   - signature: The Ed25519 signature (64 bytes)
//
// Returns an error if verification fails.
func VerifyEd25519Signature(pubKey []byte, message []byte, signature []byte) error {
	// Validate public key length
	if len(pubKey) != Ed25519PublicKeySize {
		return types.ErrInvalidPublicKeyLength.Wrapf(
			errFmtByteLength,
			Ed25519PublicKeySize, len(pubKey),
		)
	}

	// Validate signature length
	if len(signature) != Ed25519SignatureSize {
		return types.ErrInvalidSignatureLength.Wrapf(
			errFmtByteLength,
			Ed25519SignatureSize, len(signature),
		)
	}

	// Verify the signature
	if !ed25519.Verify(pubKey, message, signature) {
		return types.ErrSignatureVerificationFailed.Wrap("ed25519 signature verification failed")
	}

	return nil
}

// ============================================================================
// Secp256k1 User Signature Verification
// ============================================================================

// VerifySecp256k1Signature verifies a Cosmos-style ECDSA secp256k1 signature.
// This is used for user wallet signatures.
//
// Parameters:
//   - pubKey: The secp256k1 public key (Cosmos SDK type)
//   - message: The message that was signed (will be hashed with SHA256)
//   - signature: The secp256k1 signature (64 bytes: R || S)
//
// Returns an error if verification fails.
func VerifySecp256k1Signature(pubKey *secp256k1.PubKey, message []byte, signature []byte) error {
	// Validate public key
	if pubKey == nil {
		return types.ErrInvalidPublicKeyLength.Wrap("public key is nil")
	}

	// Validate signature length
	if len(signature) != Secp256k1SignatureSize {
		return types.ErrInvalidSignatureLength.Wrapf(
			errFmtByteLength,
			Secp256k1SignatureSize, len(signature),
		)
	}

	// Hash the message (Cosmos SDK standard is to sign the SHA256 hash)
	hash := sha256.Sum256(message)

	// Verify the signature using Cosmos SDK's secp256k1 implementation
	if !pubKey.VerifySignature(hash[:], signature) {
		return types.ErrSignatureVerificationFailed.Wrap("secp256k1 signature verification failed")
	}

	return nil
}

// VerifySecp256k1SignatureRaw verifies a secp256k1 signature using raw public key bytes.
// This is a convenience function when you have the raw key bytes.
//
// Parameters:
//   - pubKeyBytes: The compressed secp256k1 public key (33 bytes)
//   - message: The message that was signed
//   - signature: The secp256k1 signature (64 bytes)
//
// Returns an error if verification fails.
func VerifySecp256k1SignatureRaw(pubKeyBytes []byte, message []byte, signature []byte) error {
	// Validate public key length
	if len(pubKeyBytes) != Secp256k1PublicKeySize {
		return types.ErrInvalidPublicKeyLength.Wrapf(
			errFmtByteLength,
			Secp256k1PublicKeySize, len(pubKeyBytes),
		)
	}

	// Create Cosmos SDK public key from bytes
	pubKey := &secp256k1.PubKey{Key: pubKeyBytes}

	return VerifySecp256k1Signature(pubKey, message, signature)
}

// ============================================================================
// Salt Binding Verification
// ============================================================================

// SaltBindingData represents the data that is signed for salt binding verification.
// This prevents replay attacks by cryptographically binding the salt to:
// - The user's address
// - The scope ID
// - A timestamp
type SaltBindingData struct {
	Salt      []byte
	Address   sdk.AccAddress
	ScopeID   string
	Timestamp time.Time
}

// SaltBindingPayload constructs the canonical byte representation for salt binding.
// Format: Hash(salt || address || scope_id || timestamp_unix)
func (s *SaltBindingData) Payload() []byte {
	h := sha256.New()

	// Write salt
	h.Write(s.Salt)

	// Write address bytes
	h.Write(s.Address.Bytes())

	// Write scope ID
	h.Write([]byte(s.ScopeID))

	// Write timestamp as 8-byte big-endian
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, safeUint64FromInt64(s.Timestamp.Unix()))
	h.Write(ts)

	return h.Sum(nil)
}

// SaltBindingVerifyParams contains parameters for salt binding verification.
type SaltBindingVerifyParams struct {
	// Binding data
	BindingData *SaltBindingData
	// Signature over the binding
	Signature []byte
	// Public key to verify against
	PubKey []byte
	// Signature algorithm ("ed25519" or "secp256k1")
	Algorithm string
	// Current time for timestamp validation
	CurrentTime time.Time
}

// VerifySaltBindingWithParams verifies the cryptographic salt binding using a params struct.
// This ensures the salt was properly bound to the user, scope, and timestamp.
func VerifySaltBindingWithParams(params *SaltBindingVerifyParams) error {
	bindingData := params.BindingData
	if bindingData == nil {
		return types.ErrInvalidSaltBindingPayload.Wrap("binding data is nil")
	}

	// Validate inputs
	if len(bindingData.Salt) == 0 {
		return types.ErrInvalidSaltBindingPayload.Wrap("salt cannot be empty")
	}
	if len(bindingData.Address) == 0 {
		return types.ErrInvalidSaltBindingPayload.Wrap("address cannot be empty")
	}
	if bindingData.ScopeID == "" {
		return types.ErrInvalidSaltBindingPayload.Wrap("scope ID cannot be empty")
	}

	// Validate timestamp is within acceptable range
	age := params.CurrentTime.Sub(bindingData.Timestamp)
	if age > SaltBindingMaxAge {
		return types.ErrTimestampOutOfRange.Wrapf(
			"timestamp is too old: %v > %v",
			age, SaltBindingMaxAge,
		)
	}
	if age < -SaltBindingMaxFuture {
		return types.ErrTimestampOutOfRange.Wrapf(
			"timestamp is in the future: %v",
			-age,
		)
	}

	// Construct payload
	payload := bindingData.Payload()

	// Verify signature based on algorithm
	switch params.Algorithm {
	case AlgorithmEd25519:
		return VerifyEd25519Signature(params.PubKey, payload, params.Signature)
	case AlgorithmSecp256k1:
		return VerifySecp256k1SignatureRaw(params.PubKey, payload, params.Signature)
	default:
		return types.ErrUnsupportedSignatureAlgorithm.Wrapf("algorithm: %s", params.Algorithm)
	}
}

// VerifySaltBinding verifies the cryptographic salt binding.
// This ensures the salt was properly bound to the user, scope, and timestamp.
//
// Deprecated: Use VerifySaltBindingWithParams for new code.
//
//nolint:revive // Keeping for backward compatibility
func VerifySaltBinding(
	salt []byte,
	address sdk.AccAddress,
	scopeID string,
	timestamp time.Time,
	signature []byte,
	pubKey []byte,
	algorithm string,
	currentTime time.Time,
) error {
	return VerifySaltBindingWithParams(&SaltBindingVerifyParams{
		BindingData: &SaltBindingData{
			Salt:      salt,
			Address:   address,
			ScopeID:   scopeID,
			Timestamp: timestamp,
		},
		Signature:   signature,
		PubKey:      pubKey,
		Algorithm:   algorithm,
		CurrentTime: currentTime,
	})
}

// ============================================================================
// Address Derivation and Validation
// ============================================================================

// VerifyAddressMatchesPubKey verifies that a secp256k1 public key derives to the expected address.
// This is used to ensure the user signature is from the account owner.
func VerifyAddressMatchesPubKey(pubKey *secp256k1.PubKey, expectedAddr sdk.AccAddress) error {
	if pubKey == nil {
		return types.ErrInvalidPublicKeyLength.Wrap("public key is nil")
	}

	// Derive address from public key
	derivedAddr := sdk.AccAddress(pubKey.Address())

	// Compare addresses
	if !bytes.Equal(derivedAddr, expectedAddr) {
		return types.ErrPublicKeyMismatch.Wrapf(
			"derived address %s does not match expected address %s",
			derivedAddr.String(), expectedAddr.String(),
		)
	}

	return nil
}

// ============================================================================
// Composite Signature Verification
// ============================================================================

// ClientSignatureVerificationResult contains the result of client signature verification
type ClientSignatureVerificationResult struct {
	ClientID  string
	Algorithm string
	Verified  bool
	Error     error
}

// UserSignatureVerificationResult contains the result of user signature verification
type UserSignatureVerificationResult struct {
	Address   sdk.AccAddress
	Algorithm string
	Verified  bool
	Error     error
}

// CompositeSignatureParams contains parameters for composite signature verification.
type CompositeSignatureParams struct {
	// Client signature parameters
	ClientPubKey    []byte
	ClientAlgorithm string
	ClientSignature []byte
	ClientPayload   []byte

	// User signature parameters
	UserPubKey    *secp256k1.PubKey
	UserSignature []byte
	UserPayload   []byte
	UserAddress   sdk.AccAddress
}

// VerifyCompositeSignatures verifies both client and user signatures.
// This is the composite verification used during scope uploads.
func VerifyCompositeSignatures(params *CompositeSignatureParams) (*ClientSignatureVerificationResult, *UserSignatureVerificationResult) {
	clientResult := &ClientSignatureVerificationResult{
		Algorithm: params.ClientAlgorithm,
	}
	userResult := &UserSignatureVerificationResult{
		Address:   params.UserAddress,
		Algorithm: AlgorithmSecp256k1,
	}

	// Verify client signature
	switch params.ClientAlgorithm {
	case AlgorithmEd25519:
		clientResult.Error = VerifyEd25519Signature(params.ClientPubKey, params.ClientPayload, params.ClientSignature)
	case AlgorithmSecp256k1:
		clientResult.Error = VerifySecp256k1SignatureRaw(params.ClientPubKey, params.ClientPayload, params.ClientSignature)
	default:
		clientResult.Error = types.ErrUnsupportedSignatureAlgorithm.Wrapf("client algorithm: %s", params.ClientAlgorithm)
	}
	clientResult.Verified = clientResult.Error == nil

	// Verify user signature
	if params.UserPubKey != nil {
		// Verify the public key derives to the expected address
		if err := VerifyAddressMatchesPubKey(params.UserPubKey, params.UserAddress); err != nil {
			userResult.Error = err
		} else {
			userResult.Error = VerifySecp256k1Signature(params.UserPubKey, params.UserPayload, params.UserSignature)
		}
	} else {
		userResult.Error = types.ErrInvalidPublicKeyLength.Wrap("user public key is nil")
	}
	userResult.Verified = userResult.Error == nil

	return clientResult, userResult
}

// VerifyClientAndUserSignatures verifies both client and user signatures.
//
// Deprecated: Use VerifyCompositeSignatures for new code.
//
//nolint:revive // Keeping for backward compatibility
func VerifyClientAndUserSignatures(
	clientPubKey []byte,
	clientAlgorithm string,
	clientSignature []byte,
	clientPayload []byte,
	userPubKey *secp256k1.PubKey,
	userSignature []byte,
	userPayload []byte,
	userAddress sdk.AccAddress,
) (*ClientSignatureVerificationResult, *UserSignatureVerificationResult) {
	return VerifyCompositeSignatures(&CompositeSignatureParams{
		ClientPubKey:    clientPubKey,
		ClientAlgorithm: clientAlgorithm,
		ClientSignature: clientSignature,
		ClientPayload:   clientPayload,
		UserPubKey:      userPubKey,
		UserSignature:   userSignature,
		UserPayload:     userPayload,
		UserAddress:     userAddress,
	})
}
