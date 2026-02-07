package data_vault

import (
	"errors"
	"fmt"
)

var (
	// ErrBlobNotFound indicates the requested blob does not exist
	ErrBlobNotFound = errors.New("blob not found")

	// ErrUnauthorized indicates the requester lacks permission
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrInvalidScope indicates an invalid scope value
	ErrInvalidScope = errors.New("invalid scope")

	// ErrInvalidKey indicates an invalid encryption key
	ErrInvalidKey = errors.New("invalid encryption key")

	// ErrKeyRotationInProgress indicates a rotation is already active
	ErrKeyRotationInProgress = errors.New("key rotation in progress")

	// ErrDecryptionFailed indicates decryption failed
	ErrDecryptionFailed = errors.New("decryption failed")

	// ErrEncryptionFailed indicates encryption failed
	ErrEncryptionFailed = errors.New("encryption failed")

	// ErrStorageBackend indicates a storage backend error
	ErrStorageBackend = errors.New("storage backend error")

	// ErrBlobExpired indicates the blob has expired
	ErrBlobExpired = errors.New("blob expired")

	// ErrInvalidRequest indicates an invalid request
	ErrInvalidRequest = errors.New("invalid request")
)

// VaultError wraps an error with additional context
type VaultError struct {
	Op      string // Operation that failed
	BlobID  BlobID // Blob ID if applicable
	Scope   Scope  // Scope if applicable
	Err     error  // Underlying error
	Message string // Additional context
}

func (e *VaultError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *VaultError) Unwrap() error {
	return e.Err
}

// NewVaultError creates a new VaultError
func NewVaultError(op string, err error, msg string) *VaultError {
	return &VaultError{
		Op:      op,
		Err:     err,
		Message: msg,
	}
}
