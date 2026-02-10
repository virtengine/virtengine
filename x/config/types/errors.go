package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the config module
// NOTE: Error codes start at 100 to avoid conflicts with Cosmos SDK core, IBC-Go, and CosmWasm modules
var (
	// ErrInvalidClientID is returned when a client ID is invalid
	ErrInvalidClientID = errorsmod.Register(ModuleName, 1600, "invalid client ID")

	// ErrInvalidClientName is returned when a client name is invalid
	ErrInvalidClientName = errorsmod.Register(ModuleName, 1601, "invalid client name")

	// ErrInvalidClientDescription is returned when a client description is invalid
	ErrInvalidClientDescription = errorsmod.Register(ModuleName, 1602, "invalid client description")

	// ErrInvalidPublicKey is returned when a public key is invalid
	ErrInvalidPublicKey = errorsmod.Register(ModuleName, 1603, "invalid public key")

	// ErrInvalidKeyType is returned when a key type is invalid
	ErrInvalidKeyType = errorsmod.Register(ModuleName, 1604, "invalid key type")

	// ErrInvalidVersionConstraint is returned when a version constraint is invalid
	ErrInvalidVersionConstraint = errorsmod.Register(ModuleName, 1605, "invalid version constraint")

	// ErrInvalidClientStatus is returned when a client status is invalid
	ErrInvalidClientStatus = errorsmod.Register(ModuleName, 1606, "invalid client status")

	// ErrInvalidRegisteredBy is returned when registered_by is invalid
	ErrInvalidRegisteredBy = errorsmod.Register(ModuleName, 1607, "invalid registered_by address")

	// ErrInvalidTimestamp is returned when a timestamp is invalid
	ErrInvalidTimestamp = errorsmod.Register(ModuleName, 1608, "invalid timestamp")

	// ErrClientNotFound is returned when a client is not found
	ErrClientNotFound = errorsmod.Register(ModuleName, 1609, "client not found")

	// ErrClientAlreadyExists is returned when a client already exists
	ErrClientAlreadyExists = errorsmod.Register(ModuleName, 1610, "client already exists")

	// ErrInvalidStatusTransition is returned when a status transition is invalid
	ErrInvalidStatusTransition = errorsmod.Register(ModuleName, 1611, "invalid status transition")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 1612, "unauthorized")

	// ErrClientNotApproved is returned when a client is not approved for an action
	ErrClientNotApproved = errorsmod.Register(ModuleName, 1613, "client not approved")

	// ErrClientSuspended is returned when trying to use a suspended client
	ErrClientSuspended = errorsmod.Register(ModuleName, 1614, "client is suspended")

	// ErrClientRevoked is returned when trying to use a revoked client
	ErrClientRevoked = errorsmod.Register(ModuleName, 1615, "client is revoked")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 1616, "invalid signature")

	// ErrVersionNotAllowed is returned when a version is outside allowed constraints
	ErrVersionNotAllowed = errorsmod.Register(ModuleName, 1617, "version not allowed")

	// ErrScopeNotAllowed is returned when a client cannot submit a scope type
	ErrScopeNotAllowed = errorsmod.Register(ModuleName, 1618, "scope type not allowed for this client")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1619, "invalid address")

	// ErrInvalidProposal is returned when a governance proposal is invalid
	ErrInvalidProposal = errorsmod.Register(ModuleName, 1620, "invalid proposal")

	// ErrSignatureVerificationFailed is returned when signature verification fails
	ErrSignatureVerificationFailed = errorsmod.Register(ModuleName, 1621, "signature verification failed")

	// ErrInvalidPayloadHash is returned when a payload hash is invalid
	ErrInvalidPayloadHash = errorsmod.Register(ModuleName, 1622, "invalid payload hash")

	// ErrSaltBindingFailed is returned when salt binding validation fails
	ErrSaltBindingFailed = errorsmod.Register(ModuleName, 1623, "salt binding validation failed")
)
