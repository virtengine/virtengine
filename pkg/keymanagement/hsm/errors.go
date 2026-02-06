package hsm

import "errors"

// Sentinel errors for HSM operations.
var (
	// ErrNotConnected is returned when operations are attempted before Connect.
	ErrNotConnected = errors.New("hsm: not connected")

	// ErrSessionClosed is returned when the HSM session has been closed.
	ErrSessionClosed = errors.New("hsm: session closed")

	// ErrKeyNotFound is returned when a requested key label does not exist.
	ErrKeyNotFound = errors.New("hsm: key not found")

	// ErrKeyExists is returned when a key with the given label already exists.
	ErrKeyExists = errors.New("hsm: key already exists")

	// ErrOperationFailed is returned when an HSM cryptographic operation fails.
	ErrOperationFailed = errors.New("hsm: operation failed")

	// ErrAuthFailed is returned when HSM authentication (PIN/password) fails.
	ErrAuthFailed = errors.New("hsm: authentication failed")

	// ErrUnsupportedKeyType is returned for an unrecognised key type.
	ErrUnsupportedKeyType = errors.New("hsm: unsupported key type")

	// ErrDeviceNotFound is returned when no HSM device is detected.
	ErrDeviceNotFound = errors.New("hsm: device not found")
)
