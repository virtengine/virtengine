package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the encryption module
// NOTE: Error codes start at 100 to avoid conflicts with Cosmos SDK core, IBC-Go, and CosmWasm modules
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1300, "invalid address")

	// ErrInvalidPublicKey is returned when a public key is invalid
	ErrInvalidPublicKey = errorsmod.Register(ModuleName, 1301, "invalid public key")

	// ErrKeyNotFound is returned when a recipient key is not found
	ErrKeyNotFound = errorsmod.Register(ModuleName, 1302, "recipient key not found")

	// ErrKeyAlreadyExists is returned when trying to register a key that already exists
	ErrKeyAlreadyExists = errorsmod.Register(ModuleName, 1303, "key already exists")

	// ErrKeyRevoked is returned when trying to use a revoked key
	ErrKeyRevoked = errorsmod.Register(ModuleName, 1304, "key has been revoked")

	// ErrKeyDeprecated is returned when trying to use a deprecated key for new encryption
	ErrKeyDeprecated = errorsmod.Register(ModuleName, 1323, "key is deprecated")

	// ErrInvalidEnvelope is returned when an envelope is malformed
	ErrInvalidEnvelope = errorsmod.Register(ModuleName, 1305, "invalid envelope")

	// ErrUnsupportedAlgorithm is returned when an encryption algorithm is not supported
	ErrUnsupportedAlgorithm = errorsmod.Register(ModuleName, 1306, "unsupported encryption algorithm")

	// ErrUnsupportedVersion is returned when an envelope version is not supported
	ErrUnsupportedVersion = errorsmod.Register(ModuleName, 1307, "unsupported envelope version")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 1308, "invalid signature")

	// ErrEncryptionFailed is returned when encryption fails
	ErrEncryptionFailed = errorsmod.Register(ModuleName, 1309, "encryption failed")

	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errorsmod.Register(ModuleName, 1310, "decryption failed")

	// ErrInvalidNonce is returned when a nonce is invalid
	ErrInvalidNonce = errorsmod.Register(ModuleName, 1311, "invalid nonce")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 1312, "unauthorized")

	// ErrNotRecipient is returned when trying to decrypt with a non-recipient key
	ErrNotRecipient = errorsmod.Register(ModuleName, 1313, "not a recipient of this envelope")

	// ErrInvalidKeyFingerprint is returned when a key fingerprint is invalid
	ErrInvalidKeyFingerprint = errorsmod.Register(ModuleName, 1314, "invalid key fingerprint")

	// ErrMaxRecipientsExceeded is returned when too many recipients are specified
	ErrMaxRecipientsExceeded = errorsmod.Register(ModuleName, 1315, "maximum recipients exceeded")

	// ErrKeyExpired is returned when a key is expired
	ErrKeyExpired = errorsmod.Register(ModuleName, 1324, "key has expired")

	// ErrReencryptionJobFailed is returned when reencryption job processing fails
	ErrReencryptionJobFailed = errorsmod.Register(ModuleName, 1325, "reencryption job failed")

	// ============================================================================
	// Cryptography Agility Errors (VE-227)
	// ============================================================================

	// ErrCryptoAgility is returned for crypto agility related errors
	ErrCryptoAgility = errorsmod.Register(ModuleName, 1316, "cryptography agility error")

	// ErrAlgorithmNotFound is returned when an algorithm is not found in registry
	ErrAlgorithmNotFound = errorsmod.Register(ModuleName, 1317, "algorithm not found")

	// ErrAlgorithmDeprecated is returned when trying to use a deprecated algorithm
	ErrAlgorithmDeprecated = errorsmod.Register(ModuleName, 1318, "algorithm is deprecated")

	// ErrAlgorithmDisabled is returned when trying to use a disabled algorithm
	ErrAlgorithmDisabled = errorsmod.Register(ModuleName, 1319, "algorithm is disabled")

	// ErrKeyRotationInProgress is returned when a key rotation is already in progress
	ErrKeyRotationInProgress = errorsmod.Register(ModuleName, 1320, "key rotation already in progress")

	// ErrKeyRotationNotFound is returned when a key rotation record is not found
	ErrKeyRotationNotFound = errorsmod.Register(ModuleName, 1321, "key rotation not found")

	// ErrMigrationFailed is returned when algorithm migration fails
	ErrMigrationFailed = errorsmod.Register(ModuleName, 1322, "algorithm migration failed")
)
