// Package keystorage provides secure key storage backends for the signer service.
package keystorage

import (
	"cosmossdk.io/errors"
)

// Error codes for the keystorage package
var (
	// ErrKeyNotFound indicates the requested key was not found
	ErrKeyNotFound = errors.Register("verification/keystorage", 1, "key not found")

	// ErrKeyExists indicates the key already exists
	ErrKeyExists = errors.Register("verification/keystorage", 2, "key already exists")

	// ErrStorageFull indicates the storage is full
	ErrStorageFull = errors.Register("verification/keystorage", 3, "storage full")

	// ErrStorageError indicates a storage operation failed
	ErrStorageError = errors.Register("verification/keystorage", 4, "storage error")

	// ErrEncryptionError indicates an encryption operation failed
	ErrEncryptionError = errors.Register("verification/keystorage", 5, "encryption error")

	// ErrDecryptionError indicates a decryption operation failed
	ErrDecryptionError = errors.Register("verification/keystorage", 6, "decryption error")

	// ErrConnectionError indicates a connection to the backend failed
	ErrConnectionError = errors.Register("verification/keystorage", 7, "connection error")

	// ErrAuthenticationError indicates authentication failed
	ErrAuthenticationError = errors.Register("verification/keystorage", 8, "authentication error")

	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig = errors.Register("verification/keystorage", 9, "invalid configuration")

	// ErrUnsupportedOperation indicates the operation is not supported
	ErrUnsupportedOperation = errors.Register("verification/keystorage", 10, "unsupported operation")
)
