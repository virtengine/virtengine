package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the roles module
// NOTE: Error codes start at 100 to avoid conflicts with Cosmos SDK core, IBC-Go, and CosmWasm modules
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1400, "invalid address")

	// ErrInvalidRole is returned when a role is invalid
	ErrInvalidRole = errorsmod.Register(ModuleName, 1401, "invalid role")

	// ErrInvalidAccountState is returned when an account state is invalid
	ErrInvalidAccountState = errorsmod.Register(ModuleName, 1402, "invalid account state")

	// ErrUnauthorized is returned when the sender is not authorized to perform an action
	ErrUnauthorized = errorsmod.Register(ModuleName, 1403, "unauthorized")

	// ErrRoleNotFound is returned when a role assignment is not found
	ErrRoleNotFound = errorsmod.Register(ModuleName, 1404, "role assignment not found")

	// ErrRoleAlreadyAssigned is returned when a role is already assigned to an account
	ErrRoleAlreadyAssigned = errorsmod.Register(ModuleName, 1405, "role already assigned")

	// ErrAccountStateNotFound is returned when an account state is not found
	ErrAccountStateNotFound = errorsmod.Register(ModuleName, 1406, "account state not found")

	// ErrInvalidStateTransition is returned when an account state transition is not allowed
	ErrInvalidStateTransition = errorsmod.Register(ModuleName, 1407, "invalid state transition")

	// ErrCannotModifyGenesisAccount is returned when trying to modify a genesis account inappropriately
	ErrCannotModifyGenesisAccount = errorsmod.Register(ModuleName, 1408, "cannot modify genesis account")

	// ErrAccountTerminated is returned when trying to perform operations on a terminated account
	ErrAccountTerminated = errorsmod.Register(ModuleName, 1409, "account is terminated")

	// ErrAccountSuspended is returned when trying to perform operations on a suspended account
	ErrAccountSuspended = errorsmod.Register(ModuleName, 1410, "account is suspended")

	// ErrNotGenesisAccount is returned when only genesis accounts can perform an action
	ErrNotGenesisAccount = errorsmod.Register(ModuleName, 1411, "only genesis accounts can perform this action")

	// ErrCannotRevokeOwnRole is returned when trying to revoke own role
	ErrCannotRevokeOwnRole = errorsmod.Register(ModuleName, 1412, "cannot revoke own role")

	// ErrCannotSuspendSelf is returned when trying to suspend own account
	ErrCannotSuspendSelf = errorsmod.Register(ModuleName, 1413, "cannot suspend own account")
)
