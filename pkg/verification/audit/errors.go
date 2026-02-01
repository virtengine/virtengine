// Package audit provides audit logging for verification services.
package audit

import (
	"cosmossdk.io/errors"
)

// Error codes for the audit package
var (
	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig = errors.Register("verification/audit", 1, "invalid configuration")

	// ErrStorageError indicates a storage operation failed
	ErrStorageError = errors.Register("verification/audit", 2, "storage error")

	// ErrUnsupportedOperation indicates the operation is not supported
	ErrUnsupportedOperation = errors.Register("verification/audit", 3, "unsupported operation")

	// ErrConnectionError indicates a connection error
	ErrConnectionError = errors.Register("verification/audit", 4, "connection error")
)

