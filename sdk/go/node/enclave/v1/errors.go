package v1

import (
	"cosmossdk.io/errors"
)

var (
	// ErrInvalidEnclaveIdentity indicates an invalid enclave identity
	ErrInvalidEnclaveIdentity = errors.Register(ModuleName, 1, "invalid enclave identity")

	// ErrInvalidMeasurement indicates an invalid measurement
	ErrInvalidMeasurement = errors.Register(ModuleName, 2, "invalid measurement")

	// ErrEnclaveIdentityNotFound indicates the enclave identity was not found
	ErrEnclaveIdentityNotFound = errors.Register(ModuleName, 3, "enclave identity not found")

	// ErrEnclaveIdentityExists indicates the enclave identity already exists
	ErrEnclaveIdentityExists = errors.Register(ModuleName, 4, "enclave identity already exists")

	// ErrMeasurementNotAllowlisted indicates the measurement is not allowlisted
	ErrMeasurementNotAllowlisted = errors.Register(ModuleName, 5, "measurement not allowlisted")

	// ErrKeyRotationInProgress indicates a key rotation is already in progress
	ErrKeyRotationInProgress = errors.Register(ModuleName, 6, "key rotation already in progress")

	// ErrNoActiveRotation indicates no active key rotation was found
	ErrNoActiveRotation = errors.Register(ModuleName, 7, "no active key rotation")

	// ErrUnauthorized indicates an unauthorized action
	ErrUnauthorized = errors.Register(ModuleName, 8, "unauthorized")
)

