package types

import (
	"cosmossdk.io/errors"
)

// Enclave module sentinel errors
var (
	// ErrInvalidEnclaveIdentity is returned when an enclave identity is invalid
	ErrInvalidEnclaveIdentity = errors.Register(ModuleName, 1900, "invalid enclave identity")

	// ErrEnclaveIdentityNotFound is returned when an enclave identity is not found
	ErrEnclaveIdentityNotFound = errors.Register(ModuleName, 1901, "enclave identity not found")

	// ErrEnclaveIdentityExists is returned when trying to register an already existing identity
	ErrEnclaveIdentityExists = errors.Register(ModuleName, 1902, "enclave identity already exists")

	// ErrInvalidMeasurement is returned when a measurement is invalid
	ErrInvalidMeasurement = errors.Register(ModuleName, 1903, "invalid enclave measurement")

	// ErrMeasurementNotAllowlisted is returned when a measurement is not in the allowlist
	ErrMeasurementNotAllowlisted = errors.Register(ModuleName, 1904, "enclave measurement not allowlisted")

	// ErrMeasurementRevoked is returned when a measurement has been revoked
	ErrMeasurementRevoked = errors.Register(ModuleName, 1905, "enclave measurement revoked")

	// ErrMeasurementExpired is returned when a measurement has expired
	ErrMeasurementExpired = errors.Register(ModuleName, 1906, "enclave measurement expired")

	// ErrAttestationInvalid is returned when attestation verification fails
	ErrAttestationInvalid = errors.Register(ModuleName, 1907, "invalid attestation")

	// ErrAttestationExpired is returned when an attestation has expired
	ErrAttestationExpired = errors.Register(ModuleName, 1908, "attestation expired")

	// ErrDebugModeEnabled is returned when debug mode is enabled on a production enclave
	ErrDebugModeEnabled = errors.Register(ModuleName, 1909, "debug mode must be disabled for production")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errors.Register(ModuleName, 1910, "unauthorized")

	// ErrKeyRotationInProgress is returned when a key rotation is already in progress
	ErrKeyRotationInProgress = errors.Register(ModuleName, 1911, "key rotation already in progress")

	// ErrNoActiveRotation is returned when no active key rotation exists
	ErrNoActiveRotation = errors.Register(ModuleName, 1912, "no active key rotation")

	// ErrInvalidAttestedResult is returned when an attested result is invalid
	ErrInvalidAttestedResult = errors.Register(ModuleName, 1913, "invalid attested result")

	// ErrEnclaveSignatureInvalid is returned when enclave signature verification fails
	ErrEnclaveSignatureInvalid = errors.Register(ModuleName, 1914, "invalid enclave signature")

	// ErrScoreMismatch is returned when recomputed score doesn't match proposer's score
	ErrScoreMismatch = errors.Register(ModuleName, 1915, "score mismatch during consensus verification")

	// ErrEnclaveUnavailable is returned when the enclave runtime is unavailable
	ErrEnclaveUnavailable = errors.Register(ModuleName, 1916, "enclave runtime unavailable")

	// ErrISVSVNTooLow is returned when the enclave security version is too low
	ErrISVSVNTooLow = errors.Register(ModuleName, 1917, "enclave security version too low")

	// ErrInvalidQuoteVersion is returned when the quote version is not supported
	ErrInvalidQuoteVersion = errors.Register(ModuleName, 1918, "unsupported quote version")
)
