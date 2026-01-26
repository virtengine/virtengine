package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the config module
var (
	// ErrInvalidClientID is returned when a client ID is invalid
	ErrInvalidClientID = errorsmod.Register(ModuleName, 1, "invalid client ID")

	// ErrInvalidClientName is returned when a client name is invalid
	ErrInvalidClientName = errorsmod.Register(ModuleName, 2, "invalid client name")

	// ErrInvalidClientDescription is returned when a client description is invalid
	ErrInvalidClientDescription = errorsmod.Register(ModuleName, 3, "invalid client description")

	// ErrInvalidPublicKey is returned when a public key is invalid
	ErrInvalidPublicKey = errorsmod.Register(ModuleName, 4, "invalid public key")

	// ErrInvalidKeyType is returned when a key type is invalid
	ErrInvalidKeyType = errorsmod.Register(ModuleName, 5, "invalid key type")

	// ErrInvalidVersionConstraint is returned when a version constraint is invalid
	ErrInvalidVersionConstraint = errorsmod.Register(ModuleName, 6, "invalid version constraint")

	// ErrInvalidClientStatus is returned when a client status is invalid
	ErrInvalidClientStatus = errorsmod.Register(ModuleName, 7, "invalid client status")

	// ErrInvalidRegisteredBy is returned when registered_by is invalid
	ErrInvalidRegisteredBy = errorsmod.Register(ModuleName, 8, "invalid registered_by address")

	// ErrInvalidTimestamp is returned when a timestamp is invalid
	ErrInvalidTimestamp = errorsmod.Register(ModuleName, 9, "invalid timestamp")

	// ErrClientNotFound is returned when a client is not found
	ErrClientNotFound = errorsmod.Register(ModuleName, 10, "client not found")

	// ErrClientAlreadyExists is returned when a client already exists
	ErrClientAlreadyExists = errorsmod.Register(ModuleName, 11, "client already exists")

	// ErrInvalidStatusTransition is returned when a status transition is invalid
	ErrInvalidStatusTransition = errorsmod.Register(ModuleName, 12, "invalid status transition")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 13, "unauthorized")

	// ErrClientNotApproved is returned when a client is not approved for an action
	ErrClientNotApproved = errorsmod.Register(ModuleName, 14, "client not approved")

	// ErrClientSuspended is returned when trying to use a suspended client
	ErrClientSuspended = errorsmod.Register(ModuleName, 15, "client is suspended")

	// ErrClientRevoked is returned when trying to use a revoked client
	ErrClientRevoked = errorsmod.Register(ModuleName, 16, "client is revoked")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 17, "invalid signature")

	// ErrVersionNotAllowed is returned when a version is outside allowed constraints
	ErrVersionNotAllowed = errorsmod.Register(ModuleName, 18, "version not allowed")

	// ErrScopeNotAllowed is returned when a client cannot submit a scope type
	ErrScopeNotAllowed = errorsmod.Register(ModuleName, 19, "scope type not allowed for this client")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 20, "invalid address")

	// ErrInvalidProposal is returned when a governance proposal is invalid
	ErrInvalidProposal = errorsmod.Register(ModuleName, 21, "invalid proposal")

	// ErrSignatureVerificationFailed is returned when signature verification fails
	ErrSignatureVerificationFailed = errorsmod.Register(ModuleName, 22, "signature verification failed")

	// ErrInvalidPayloadHash is returned when a payload hash is invalid
	ErrInvalidPayloadHash = errorsmod.Register(ModuleName, 23, "invalid payload hash")

	// ErrSaltBindingFailed is returned when salt binding validation fails
	ErrSaltBindingFailed = errorsmod.Register(ModuleName, 24, "salt binding validation failed")
)
