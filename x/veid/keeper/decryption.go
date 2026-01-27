package keeper

import (
	"crypto/sha256"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Decrypted Scope
// ============================================================================

// DecryptedScope contains the decrypted content of an identity scope
type DecryptedScope struct {
	// ScopeID is the scope identifier
	ScopeID string

	// ScopeType is the type of scope
	ScopeType types.ScopeType

	// Plaintext is the decrypted scope content
	Plaintext []byte

	// ContentHash is the SHA256 hash of the plaintext for audit/consensus
	ContentHash []byte
}

// NewDecryptedScope creates a new decrypted scope
func NewDecryptedScope(scopeID string, scopeType types.ScopeType, plaintext []byte) *DecryptedScope {
	h := sha256.Sum256(plaintext)
	return &DecryptedScope{
		ScopeID:     scopeID,
		ScopeType:   scopeType,
		Plaintext:   plaintext,
		ContentHash: h[:],
	}
}

// ============================================================================
// Decryption Configuration
// ============================================================================

// DecryptionConfig holds configuration for scope decryption
type DecryptionConfig struct {
	// ValidatorKeyPath is the file path to the validator's identity key
	ValidatorKeyPath string

	// KeyPassword is the password for the encrypted key file (if encrypted)
	KeyPassword string

	// MaxDecryptionTime is the maximum time allowed for decryption in milliseconds
	MaxDecryptionTime int64
}

// DefaultDecryptionConfig returns default decryption configuration
func DefaultDecryptionConfig() DecryptionConfig {
	return DecryptionConfig{
		ValidatorKeyPath:  "",
		KeyPassword:       "",
		MaxDecryptionTime: 500, // 500ms default
	}
}

// ValidatorKeyProvider is an interface for obtaining validator identity keys
// This allows for different key storage backends (file, HSM, vault, etc.)
type ValidatorKeyProvider interface {
	// GetPrivateKey returns the validator's X25519 private key for decryption
	GetPrivateKey() ([]byte, error)

	// GetKeyFingerprint returns the public key fingerprint for this validator
	GetKeyFingerprint() string

	// Close releases any resources held by the provider
	Close() error
}

// ============================================================================
// In-Memory Key Provider (for testing/development)
// ============================================================================

// InMemoryKeyProvider holds a key pair in memory
type InMemoryKeyProvider struct {
	keyPair *encryptioncrypto.KeyPair
}

// NewInMemoryKeyProvider creates a new in-memory key provider with the given key pair
func NewInMemoryKeyProvider(keyPair *encryptioncrypto.KeyPair) *InMemoryKeyProvider {
	return &InMemoryKeyProvider{
		keyPair: keyPair,
	}
}

// GetPrivateKey returns the private key bytes
func (p *InMemoryKeyProvider) GetPrivateKey() ([]byte, error) {
	if p.keyPair == nil {
		return nil, types.ErrValidatorKeyNotFound.Wrap("no key pair configured")
	}
	return p.keyPair.PrivateKey[:], nil
}

// GetKeyFingerprint returns the public key fingerprint
func (p *InMemoryKeyProvider) GetKeyFingerprint() string {
	if p.keyPair == nil {
		return ""
	}
	return p.keyPair.Fingerprint()
}

// Close is a no-op for in-memory provider
func (p *InMemoryKeyProvider) Close() error {
	return nil
}

// ============================================================================
// Scope Decryption
// ============================================================================

// DecryptScope decrypts an identity scope using the validator's key
func (k Keeper) DecryptScope(
	ctx sdk.Context,
	scope types.IdentityScope,
	keyProvider ValidatorKeyProvider,
) (*DecryptedScope, error) {
	// Validate the scope
	if err := scope.Validate(); err != nil {
		return nil, types.ErrDecryptionFailed.Wrap(err.Error())
	}

	// Check if scope is in a valid state for decryption
	if scope.Revoked {
		return nil, types.ErrScopeRevoked.Wrapf("scope %s is revoked", scope.ScopeID)
	}

	if scope.Status == types.VerificationStatusVerified {
		// Already verified, no need to decrypt again
		k.Logger(ctx).Debug("scope already verified", "scope_id", scope.ScopeID)
	}

	// Get the validator's private key
	privateKey, err := keyProvider.GetPrivateKey()
	if err != nil {
		return nil, types.ErrValidatorKeyNotFound.Wrap(err.Error())
	}

	// Check if this envelope is encrypted for this validator
	validatorFingerprint := keyProvider.GetKeyFingerprint()
	recipientIndex := scope.EncryptedPayload.GetRecipientIndex(validatorFingerprint)
	if recipientIndex < 0 {
		return nil, types.ErrDecryptionFailed.Wrapf(
			"envelope not encrypted for validator %s", validatorFingerprint)
	}

	// Decrypt the envelope
	plaintext, err := encryptioncrypto.OpenEnvelope(&scope.EncryptedPayload, privateKey)
	if err != nil {
		return nil, types.ErrDecryptionFailed.Wrap(err.Error())
	}

	return NewDecryptedScope(scope.ScopeID, scope.ScopeType, plaintext), nil
}

// DecryptScopesForVerification decrypts multiple scopes for verification
func (k Keeper) DecryptScopesForVerification(
	ctx sdk.Context,
	address sdk.AccAddress,
	scopeIDs []string,
	keyProvider ValidatorKeyProvider,
) ([]DecryptedScope, []types.ScopeVerificationResult, error) {
	decrypted := make([]DecryptedScope, 0, len(scopeIDs))
	results := make([]types.ScopeVerificationResult, 0, len(scopeIDs))

	for _, scopeID := range scopeIDs {
		// Get the scope
		scope, found := k.GetScope(ctx, address, scopeID)
		if !found {
			result := types.NewScopeVerificationResult(scopeID, "")
			result.SetFailure(types.ReasonCodeScopeNotFound)
			result.Details = fmt.Sprintf("scope %s not found", scopeID)
			results = append(results, *result)
			continue
		}

		// Check if scope is revoked
		if scope.Revoked {
			result := types.NewScopeVerificationResult(scopeID, scope.ScopeType)
			result.SetFailure(types.ReasonCodeScopeRevoked)
			result.Details = scope.RevokedReason
			results = append(results, *result)
			continue
		}

		// Check if scope is expired
		if scope.ExpiresAt != nil && scope.ExpiresAt.Before(ctx.BlockTime()) {
			result := types.NewScopeVerificationResult(scopeID, scope.ScopeType)
			result.SetFailure(types.ReasonCodeScopeExpired)
			results = append(results, *result)
			continue
		}

		// Attempt decryption
		decryptedScope, err := k.DecryptScope(ctx, scope, keyProvider)
		if err != nil {
			result := types.NewScopeVerificationResult(scopeID, scope.ScopeType)
			result.SetFailure(types.ReasonCodeDecryptError)
			result.Details = err.Error()
			results = append(results, *result)
			continue
		}

		decrypted = append(decrypted, *decryptedScope)

		// Create success result (score will be set later by ML)
		result := types.NewScopeVerificationResult(scopeID, scope.ScopeType)
		results = append(results, *result)
	}

	return decrypted, results, nil
}

// ValidateDecryptedPayload validates the structure of a decrypted payload
// Returns true if the payload appears valid for its scope type
func (k Keeper) ValidateDecryptedPayload(
	ctx sdk.Context,
	decrypted DecryptedScope,
) (bool, string) {
	// Basic validation: payload must have content
	if len(decrypted.Plaintext) == 0 {
		return false, "empty payload"
	}

	// Scope type-specific validation
	switch decrypted.ScopeType {
	case types.ScopeTypeIDDocument:
		return k.validateIDDocumentPayload(decrypted.Plaintext)
	case types.ScopeTypeSelfie:
		return k.validateSelfiePayload(decrypted.Plaintext)
	case types.ScopeTypeFaceVideo:
		return k.validateFaceVideoPayload(decrypted.Plaintext)
	case types.ScopeTypeBiometric:
		return k.validateBiometricPayload(decrypted.Plaintext)
	case types.ScopeTypeSSOMetadata:
		return k.validateSSOMetadataPayload(decrypted.Plaintext)
	case types.ScopeTypeEmailProof:
		return k.validateEmailProofPayload(decrypted.Plaintext)
	case types.ScopeTypeSMSProof:
		return k.validateSMSProofPayload(decrypted.Plaintext)
	case types.ScopeTypeDomainVerify:
		return k.validateDomainVerifyPayload(decrypted.Plaintext)
	default:
		return false, fmt.Sprintf("unknown scope type: %s", decrypted.ScopeType)
	}
}

// ============================================================================
// Payload Validation Helpers
// ============================================================================

// validateIDDocumentPayload validates an ID document payload structure
func (k Keeper) validateIDDocumentPayload(payload []byte) (bool, string) {
	// Minimum size check (expecting image data)
	if len(payload) < 1024 {
		return false, "ID document payload too small"
	}

	// Check for common image headers (JPEG, PNG)
	if !hasValidImageHeader(payload) {
		return false, "ID document does not appear to be a valid image"
	}

	return true, ""
}

// validateSelfiePayload validates a selfie payload structure
func (k Keeper) validateSelfiePayload(payload []byte) (bool, string) {
	if len(payload) < 1024 {
		return false, "selfie payload too small"
	}

	if !hasValidImageHeader(payload) {
		return false, "selfie does not appear to be a valid image"
	}

	return true, ""
}

// validateFaceVideoPayload validates a face video payload structure
func (k Keeper) validateFaceVideoPayload(payload []byte) (bool, string) {
	if len(payload) < 10240 {
		return false, "face video payload too small"
	}

	// Check for common video container headers
	if !hasValidVideoHeader(payload) {
		return false, "face video does not appear to be a valid video"
	}

	return true, ""
}

// validateBiometricPayload validates a biometric data payload
func (k Keeper) validateBiometricPayload(payload []byte) (bool, string) {
	if len(payload) < 256 {
		return false, "biometric payload too small"
	}
	return true, ""
}

// validateSSOMetadataPayload validates SSO metadata payload
func (k Keeper) validateSSOMetadataPayload(payload []byte) (bool, string) {
	if len(payload) < 32 {
		return false, "SSO metadata payload too small"
	}
	// SSO metadata is typically JSON
	if payload[0] != '{' {
		return false, "SSO metadata does not appear to be valid JSON"
	}
	return true, ""
}

// validateEmailProofPayload validates email proof payload
func (k Keeper) validateEmailProofPayload(payload []byte) (bool, string) {
	if len(payload) < 32 {
		return false, "email proof payload too small"
	}
	return true, ""
}

// validateSMSProofPayload validates SMS proof payload
func (k Keeper) validateSMSProofPayload(payload []byte) (bool, string) {
	if len(payload) < 16 {
		return false, "SMS proof payload too small"
	}
	return true, ""
}

// validateDomainVerifyPayload validates domain verification payload
func (k Keeper) validateDomainVerifyPayload(payload []byte) (bool, string) {
	if len(payload) < 32 {
		return false, "domain verification payload too small"
	}
	return true, ""
}

// hasValidImageHeader checks if payload starts with JPEG or PNG header
func hasValidImageHeader(payload []byte) bool {
	if len(payload) < 8 {
		return false
	}

	// JPEG: starts with FF D8 FF
	if payload[0] == 0xFF && payload[1] == 0xD8 && payload[2] == 0xFF {
		return true
	}

	// PNG: starts with 89 50 4E 47 0D 0A 1A 0A
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngHeader {
		if payload[i] != b {
			break
		}
		if i == len(pngHeader)-1 {
			return true
		}
	}

	// WebP: starts with RIFF....WEBP
	if len(payload) >= 12 &&
		payload[0] == 'R' && payload[1] == 'I' && payload[2] == 'F' && payload[3] == 'F' &&
		payload[8] == 'W' && payload[9] == 'E' && payload[10] == 'B' && payload[11] == 'P' {
		return true
	}

	return false
}

// hasValidVideoHeader checks if payload starts with common video container headers
func hasValidVideoHeader(payload []byte) bool {
	if len(payload) < 12 {
		return false
	}

	// MP4/MOV: ftyp box
	if len(payload) >= 8 && string(payload[4:8]) == "ftyp" {
		return true
	}

	// WebM: EBML header
	if payload[0] == 0x1A && payload[1] == 0x45 && payload[2] == 0xDF && payload[3] == 0xA3 {
		return true
	}

	// AVI: RIFF....AVI
	if payload[0] == 'R' && payload[1] == 'I' && payload[2] == 'F' && payload[3] == 'F' &&
		payload[8] == 'A' && payload[9] == 'V' && payload[10] == 'I' {
		return true
	}

	return false
}
