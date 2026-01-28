package v1

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the veid module (range: 100-199)
// NOTE: Error codes start at 100 to avoid conflicts with Cosmos SDK core, IBC-Go, and CosmWasm modules
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1000, "invalid address")

	// ErrInvalidScope is returned when a scope is malformed
	ErrInvalidScope = errorsmod.Register(ModuleName, 1001, "invalid scope")

	// ErrInvalidScopeType is returned when a scope type is invalid
	ErrInvalidScopeType = errorsmod.Register(ModuleName, 1002, "invalid scope type")

	// ErrInvalidScopeVersion is returned when a scope version is invalid
	ErrInvalidScopeVersion = errorsmod.Register(ModuleName, 1003, "invalid scope version")

	// ErrInvalidPayload is returned when an encrypted payload is invalid
	ErrInvalidPayload = errorsmod.Register(ModuleName, 1004, "invalid encrypted payload")

	// ErrInvalidSalt is returned when a salt is invalid
	ErrInvalidSalt = errorsmod.Register(ModuleName, 1005, "invalid salt")

	// ErrSaltAlreadyUsed is returned when a salt has already been used
	ErrSaltAlreadyUsed = errorsmod.Register(ModuleName, 1006, "salt already used")

	// ErrInvalidDeviceInfo is returned when device information is invalid
	ErrInvalidDeviceInfo = errorsmod.Register(ModuleName, 1007, "invalid device info")

	// ErrInvalidClientID is returned when a client ID is invalid
	ErrInvalidClientID = errorsmod.Register(ModuleName, 1008, "invalid client ID")

	// ErrClientNotApproved is returned when a client is not approved
	ErrClientNotApproved = errorsmod.Register(ModuleName, 1009, "client not approved")

	// ErrInvalidClientSignature is returned when a client signature is invalid
	ErrInvalidClientSignature = errorsmod.Register(ModuleName, 1010, "invalid client signature")

	// ErrInvalidUserSignature is returned when a user signature is invalid
	ErrInvalidUserSignature = errorsmod.Register(ModuleName, 1011, "invalid user signature")

	// ErrInvalidPayloadHash is returned when a payload hash is invalid
	ErrInvalidPayloadHash = errorsmod.Register(ModuleName, 1012, "invalid payload hash")

	// ErrInvalidVerificationStatus is returned when a verification status is invalid
	ErrInvalidVerificationStatus = errorsmod.Register(ModuleName, 1013, "invalid verification status")

	// ErrInvalidVerificationEvent is returned when a verification event is invalid
	ErrInvalidVerificationEvent = errorsmod.Register(ModuleName, 1014, "invalid verification event")

	// ErrInvalidScore is returned when an identity score is invalid
	ErrInvalidScore = errorsmod.Register(ModuleName, 1015, "invalid score")

	// ErrInvalidTier is returned when an identity tier is invalid
	ErrInvalidTier = errorsmod.Register(ModuleName, 1016, "invalid tier")

	// ErrInvalidIdentityRecord is returned when an identity record is invalid
	ErrInvalidIdentityRecord = errorsmod.Register(ModuleName, 1017, "invalid identity record")

	// ErrInvalidWallet is returned when an identity wallet is invalid
	ErrInvalidWallet = errorsmod.Register(ModuleName, 1018, "invalid identity wallet")

	// ErrScopeNotFound is returned when a scope is not found
	ErrScopeNotFound = errorsmod.Register(ModuleName, 1019, "scope not found")

	// ErrIdentityRecordNotFound is returned when an identity record is not found
	ErrIdentityRecordNotFound = errorsmod.Register(ModuleName, 1020, "identity record not found")

	// ErrScopeAlreadyExists is returned when a scope already exists
	ErrScopeAlreadyExists = errorsmod.Register(ModuleName, 1021, "scope already exists")

	// ErrScopeRevoked is returned when trying to use a revoked scope
	ErrScopeRevoked = errorsmod.Register(ModuleName, 1022, "scope has been revoked")

	// ErrScopeExpired is returned when trying to use an expired scope
	ErrScopeExpired = errorsmod.Register(ModuleName, 1023, "scope has expired")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 1024, "unauthorized")

	// ErrInvalidStatusTransition is returned when an invalid status transition is attempted
	ErrInvalidStatusTransition = errorsmod.Register(ModuleName, 1025, "invalid status transition")

	// ErrIdentityLocked is returned when trying to modify a locked identity
	ErrIdentityLocked = errorsmod.Register(ModuleName, 1026, "identity is locked")

	// ErrMaxScopesExceeded is returned when the maximum number of scopes is exceeded
	ErrMaxScopesExceeded = errorsmod.Register(ModuleName, 1027, "maximum scopes exceeded")

	// ErrVerificationInProgress is returned when verification is already in progress
	ErrVerificationInProgress = errorsmod.Register(ModuleName, 1028, "verification already in progress")

	// ErrValidatorOnly is returned when a non-validator attempts a validator-only action
	ErrValidatorOnly = errorsmod.Register(ModuleName, 1029, "action restricted to validators")
)
