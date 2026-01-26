package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the encryption module
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1, "invalid address")

	// ErrInvalidPublicKey is returned when a public key is invalid
	ErrInvalidPublicKey = errorsmod.Register(ModuleName, 2, "invalid public key")

	// ErrKeyNotFound is returned when a recipient key is not found
	ErrKeyNotFound = errorsmod.Register(ModuleName, 3, "recipient key not found")

	// ErrKeyAlreadyExists is returned when trying to register a key that already exists
	ErrKeyAlreadyExists = errorsmod.Register(ModuleName, 4, "key already exists")

	// ErrKeyRevoked is returned when trying to use a revoked key
	ErrKeyRevoked = errorsmod.Register(ModuleName, 5, "key has been revoked")

	// ErrInvalidEnvelope is returned when an envelope is malformed
	ErrInvalidEnvelope = errorsmod.Register(ModuleName, 6, "invalid envelope")

	// ErrUnsupportedAlgorithm is returned when an encryption algorithm is not supported
	ErrUnsupportedAlgorithm = errorsmod.Register(ModuleName, 7, "unsupported encryption algorithm")

	// ErrUnsupportedVersion is returned when an envelope version is not supported
	ErrUnsupportedVersion = errorsmod.Register(ModuleName, 8, "unsupported envelope version")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 9, "invalid signature")

	// ErrEncryptionFailed is returned when encryption fails
	ErrEncryptionFailed = errorsmod.Register(ModuleName, 10, "encryption failed")

	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errorsmod.Register(ModuleName, 11, "decryption failed")

	// ErrInvalidNonce is returned when a nonce is invalid
	ErrInvalidNonce = errorsmod.Register(ModuleName, 12, "invalid nonce")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 13, "unauthorized")

	// ErrNotRecipient is returned when trying to decrypt with a non-recipient key
	ErrNotRecipient = errorsmod.Register(ModuleName, 14, "not a recipient of this envelope")

	// ErrInvalidKeyFingerprint is returned when a key fingerprint is invalid
	ErrInvalidKeyFingerprint = errorsmod.Register(ModuleName, 15, "invalid key fingerprint")

	// ErrMaxRecipientsExceeded is returned when too many recipients are specified
	ErrMaxRecipientsExceeded = errorsmod.Register(ModuleName, 16, "maximum recipients exceeded")

	// ============================================================================
	// Cryptography Agility Errors (VE-227)
	// ============================================================================

	// ErrCryptoAgility is returned for crypto agility related errors
	ErrCryptoAgility = errorsmod.Register(ModuleName, 17, "cryptography agility error")

	// ErrAlgorithmNotFound is returned when an algorithm is not found in registry
	ErrAlgorithmNotFound = errorsmod.Register(ModuleName, 18, "algorithm not found")

	// ErrAlgorithmDeprecated is returned when trying to use a deprecated algorithm
	ErrAlgorithmDeprecated = errorsmod.Register(ModuleName, 19, "algorithm is deprecated")

	// ErrAlgorithmDisabled is returned when trying to use a disabled algorithm
	ErrAlgorithmDisabled = errorsmod.Register(ModuleName, 20, "algorithm is disabled")

	// ErrKeyRotationInProgress is returned when a key rotation is already in progress
	ErrKeyRotationInProgress = errorsmod.Register(ModuleName, 21, "key rotation already in progress")

	// ErrKeyRotationNotFound is returned when a key rotation record is not found
	ErrKeyRotationNotFound = errorsmod.Register(ModuleName, 22, "key rotation not found")

	// ErrMigrationFailed is returned when algorithm migration fails
	ErrMigrationFailed = errorsmod.Register(ModuleName, 23, "algorithm migration failed")
)
