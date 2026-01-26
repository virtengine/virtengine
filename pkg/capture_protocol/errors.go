package capture_protocol

import (
	"fmt"
)

// Error codes for the capture protocol
const (
	// Salt errors (1001-1099)
	ErrCodeSaltEmpty              = "SALT_EMPTY"
	ErrCodeSaltTooShort           = "SALT_TOO_SHORT"
	ErrCodeSaltTooLong            = "SALT_TOO_LONG"
	ErrCodeSaltWeak               = "SALT_WEAK"
	ErrCodeSaltExpired            = "SALT_EXPIRED"
	ErrCodeSaltFromFuture         = "SALT_FROM_FUTURE"
	ErrCodeSaltReplayed           = "SALT_REPLAYED"
	ErrCodeSaltMismatch           = "SALT_MISMATCH"

	// Binding errors (1101-1199)
	ErrCodeBindingHashMissing     = "BINDING_HASH_MISSING"
	ErrCodeBindingHashMismatch    = "BINDING_HASH_MISMATCH"
	ErrCodeBindingSaltMismatch    = "BINDING_SALT_MISMATCH"
	ErrCodeBindingDeviceIDMissing = "BINDING_DEVICE_ID_MISSING"
	ErrCodeBindingSessionIDMissing = "BINDING_SESSION_ID_MISSING"

	// Client signature errors (1201-1299)
	ErrCodeClientSignatureMissing = "CLIENT_SIGNATURE_MISSING"
	ErrCodeClientSignatureInvalid = "CLIENT_SIGNATURE_INVALID"
	ErrCodeClientIDMissing        = "CLIENT_ID_MISSING"
	ErrCodeClientNotApproved      = "CLIENT_NOT_APPROVED"
	ErrCodeClientNotActive        = "CLIENT_NOT_ACTIVE"
	ErrCodeClientKeyMismatch      = "CLIENT_KEY_MISMATCH"

	// User signature errors (1301-1399)
	ErrCodeUserSignatureMissing   = "USER_SIGNATURE_MISSING"
	ErrCodeUserSignatureInvalid   = "USER_SIGNATURE_INVALID"
	ErrCodeUserPublicKeyMissing   = "USER_PUBLIC_KEY_MISSING"
	ErrCodeUserAddressMismatch    = "USER_ADDRESS_MISMATCH"

	// Signature chain errors (1401-1499)
	ErrCodeAlgorithmMismatch      = "ALGORITHM_MISMATCH"
	ErrCodeSignedDataMismatch     = "SIGNED_DATA_MISMATCH"
	ErrCodeSignatureChainBroken   = "SIGNATURE_CHAIN_BROKEN"

	// Version errors (1501-1599)
	ErrCodeVersionMissing         = "VERSION_MISSING"
	ErrCodeVersionUnsupported     = "VERSION_UNSUPPORTED"

	// Metadata errors (1601-1699)
	ErrCodeMetadataClientIDMissing        = "METADATA_CLIENT_ID_MISSING"
	ErrCodeMetadataClientIDMismatch       = "METADATA_CLIENT_ID_MISMATCH"
	ErrCodeMetadataDeviceFingerprintMissing = "METADATA_DEVICE_FINGERPRINT_MISSING"
	ErrCodeMetadataSessionIDMissing       = "METADATA_SESSION_ID_MISSING"
	ErrCodeMetadataSessionIDMismatch      = "METADATA_SESSION_ID_MISMATCH"

	// Replay/timing errors (1701-1799)
	ErrCodePayloadExpired         = "PAYLOAD_EXPIRED"
	ErrCodePayloadFromFuture      = "PAYLOAD_FROM_FUTURE"
)

// ProtocolError represents an error in the capture protocol
type ProtocolError struct {
	Code    string
	Message string
	Field   string
	Details map[string]interface{}
	Wrapped error
}

// Error implements the error interface
func (e *ProtocolError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *ProtocolError) Unwrap() error {
	return e.Wrapped
}

// Wrap wraps another error
func (e *ProtocolError) Wrap(err error) *ProtocolError {
	return &ProtocolError{
		Code:    e.Code,
		Message: e.Message,
		Field:   e.Field,
		Details: e.Details,
		Wrapped: err,
	}
}

// WithDetails adds details to the error
func (e *ProtocolError) WithDetails(keyvals ...interface{}) *ProtocolError {
	details := make(map[string]interface{})
	for k, v := range e.Details {
		details[k] = v
	}
	for i := 0; i < len(keyvals)-1; i += 2 {
		if key, ok := keyvals[i].(string); ok {
			details[key] = keyvals[i+1]
		}
	}
	return &ProtocolError{
		Code:    e.Code,
		Message: e.Message,
		Field:   e.Field,
		Details: details,
		Wrapped: e.Wrapped,
	}
}

// WithField sets the field name
func (e *ProtocolError) WithField(field string) *ProtocolError {
	return &ProtocolError{
		Code:    e.Code,
		Message: e.Message,
		Field:   field,
		Details: e.Details,
		Wrapped: e.Wrapped,
	}
}

// newError creates a new ProtocolError
func newError(code, message string) *ProtocolError {
	return &ProtocolError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Salt errors
var (
	ErrSaltEmpty = newError(
		ErrCodeSaltEmpty,
		"salt cannot be empty",
	)

	ErrSaltTooShort = newError(
		ErrCodeSaltTooShort,
		"salt is too short",
	)

	ErrSaltTooLong = newError(
		ErrCodeSaltTooLong,
		"salt exceeds maximum length",
	)

	ErrSaltWeak = newError(
		ErrCodeSaltWeak,
		"salt appears to be weak (low entropy)",
	)

	ErrSaltExpired = newError(
		ErrCodeSaltExpired,
		"salt timestamp is too old",
	)

	ErrSaltFromFuture = newError(
		ErrCodeSaltFromFuture,
		"salt timestamp is in the future",
	)

	ErrSaltReplayed = newError(
		ErrCodeSaltReplayed,
		"salt has already been used (replay detected)",
	)

	ErrSaltMismatch = newError(
		ErrCodeSaltMismatch,
		"salt does not match salt in binding",
	)
)

// Binding errors
var (
	ErrBindingHashMissing = newError(
		ErrCodeBindingHashMissing,
		"binding hash is missing",
	)

	ErrBindingHashMismatch = newError(
		ErrCodeBindingHashMismatch,
		"binding hash does not match computed hash",
	)

	ErrBindingSaltMismatch = newError(
		ErrCodeBindingSaltMismatch,
		"salt in binding does not match payload salt",
	)

	ErrBindingDeviceIDMissing = newError(
		ErrCodeBindingDeviceIDMissing,
		"device ID is missing from binding",
	)

	ErrBindingSessionIDMissing = newError(
		ErrCodeBindingSessionIDMissing,
		"session ID is missing from binding",
	)
)

// Client signature errors
var (
	ErrClientSignatureMissing = newError(
		ErrCodeClientSignatureMissing,
		"client signature is required but missing",
	)

	ErrClientSignatureInvalid = newError(
		ErrCodeClientSignatureInvalid,
		"client signature verification failed",
	)

	ErrClientIDMissing = newError(
		ErrCodeClientIDMissing,
		"client ID is missing from signature",
	)

	ErrClientNotApproved = newError(
		ErrCodeClientNotApproved,
		"client is not in the approved client list",
	)

	ErrClientNotActive = newError(
		ErrCodeClientNotActive,
		"client is registered but not currently active",
	)

	ErrClientKeyMismatch = newError(
		ErrCodeClientKeyMismatch,
		"client public key does not match registered key",
	)
)

// User signature errors
var (
	ErrUserSignatureMissing = newError(
		ErrCodeUserSignatureMissing,
		"user signature is required but missing",
	)

	ErrUserSignatureInvalid = newError(
		ErrCodeUserSignatureInvalid,
		"user signature verification failed",
	)

	ErrUserPublicKeyMissing = newError(
		ErrCodeUserPublicKeyMissing,
		"user public key is missing from signature",
	)

	ErrUserAddressMismatch = newError(
		ErrCodeUserAddressMismatch,
		"user address does not match expected account",
	)
)

// Signature chain errors
var (
	ErrAlgorithmMismatch = newError(
		ErrCodeAlgorithmMismatch,
		"signature algorithm does not match expected algorithm",
	)

	ErrSignedDataMismatch = newError(
		ErrCodeSignedDataMismatch,
		"signed data does not match expected data",
	)

	ErrSignatureChainBroken = newError(
		ErrCodeSignatureChainBroken,
		"signature chain is broken (user signature does not include client signature)",
	)
)

// Version errors
var (
	ErrVersionMissing = newError(
		ErrCodeVersionMissing,
		"protocol version is missing",
	)

	ErrVersionUnsupported = newError(
		ErrCodeVersionUnsupported,
		"protocol version is not supported",
	)
)

// Metadata errors
var (
	ErrMetadataClientIDMissing = newError(
		ErrCodeMetadataClientIDMissing,
		"client ID is missing from metadata",
	)

	ErrMetadataClientIDMismatch = newError(
		ErrCodeMetadataClientIDMismatch,
		"client ID in metadata does not match signature client ID",
	)

	ErrMetadataDeviceFingerprintMissing = newError(
		ErrCodeMetadataDeviceFingerprintMissing,
		"device fingerprint is missing from metadata",
	)

	ErrMetadataSessionIDMissing = newError(
		ErrCodeMetadataSessionIDMissing,
		"session ID is missing from metadata",
	)

	ErrMetadataSessionIDMismatch = newError(
		ErrCodeMetadataSessionIDMismatch,
		"session ID in metadata does not match binding session ID",
	)
)

// Replay/timing errors
var (
	ErrPayloadExpired = newError(
		ErrCodePayloadExpired,
		"payload has expired (timestamp too old)",
	)

	ErrPayloadFromFuture = newError(
		ErrCodePayloadFromFuture,
		"payload timestamp is in the future",
	)
)

// IsProtocolError checks if an error is a ProtocolError
func IsProtocolError(err error) bool {
	_, ok := err.(*ProtocolError)
	return ok
}

// GetErrorCode returns the error code for a ProtocolError, or empty string
func GetErrorCode(err error) string {
	if protocolErr, ok := err.(*ProtocolError); ok {
		return protocolErr.Code
	}
	return ""
}

// IsSaltError checks if an error is related to salt validation
func IsSaltError(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case ErrCodeSaltEmpty, ErrCodeSaltTooShort, ErrCodeSaltTooLong,
		ErrCodeSaltWeak, ErrCodeSaltExpired, ErrCodeSaltFromFuture,
		ErrCodeSaltReplayed, ErrCodeSaltMismatch:
		return true
	}
	return false
}

// IsSignatureError checks if an error is related to signature validation
func IsSignatureError(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case ErrCodeClientSignatureMissing, ErrCodeClientSignatureInvalid,
		ErrCodeClientIDMissing, ErrCodeClientNotApproved,
		ErrCodeClientNotActive, ErrCodeClientKeyMismatch,
		ErrCodeUserSignatureMissing, ErrCodeUserSignatureInvalid,
		ErrCodeUserPublicKeyMissing, ErrCodeUserAddressMismatch,
		ErrCodeAlgorithmMismatch, ErrCodeSignedDataMismatch,
		ErrCodeSignatureChainBroken:
		return true
	}
	return false
}

// IsReplayError checks if an error indicates a replay attack
func IsReplayError(err error) bool {
	code := GetErrorCode(err)
	return code == ErrCodeSaltReplayed
}

// IsClientError checks if an error is related to client validation
func IsClientError(err error) bool {
	code := GetErrorCode(err)
	switch code {
	case ErrCodeClientSignatureMissing, ErrCodeClientSignatureInvalid,
		ErrCodeClientIDMissing, ErrCodeClientNotApproved,
		ErrCodeClientNotActive, ErrCodeClientKeyMismatch:
		return true
	}
	return false
}
