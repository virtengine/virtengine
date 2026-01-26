package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the roles module
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1, "invalid address")

	// ErrInvalidRole is returned when a role is invalid
	ErrInvalidRole = errorsmod.Register(ModuleName, 2, "invalid role")

	// ErrInvalidAccountState is returned when an account state is invalid
	ErrInvalidAccountState = errorsmod.Register(ModuleName, 3, "invalid account state")

	// ErrUnauthorized is returned when the sender is not authorized to perform an action
	ErrUnauthorized = errorsmod.Register(ModuleName, 4, "unauthorized")

	// ErrRoleNotFound is returned when a role assignment is not found
	ErrRoleNotFound = errorsmod.Register(ModuleName, 5, "role assignment not found")

	// ErrRoleAlreadyAssigned is returned when a role is already assigned to an account
	ErrRoleAlreadyAssigned = errorsmod.Register(ModuleName, 6, "role already assigned")

	// ErrAccountStateNotFound is returned when an account state is not found
	ErrAccountStateNotFound = errorsmod.Register(ModuleName, 7, "account state not found")

	// ErrInvalidStateTransition is returned when an account state transition is not allowed
	ErrInvalidStateTransition = errorsmod.Register(ModuleName, 8, "invalid state transition")

	// ErrCannotModifyGenesisAccount is returned when trying to modify a genesis account inappropriately
	ErrCannotModifyGenesisAccount = errorsmod.Register(ModuleName, 9, "cannot modify genesis account")

	// ErrAccountTerminated is returned when trying to perform operations on a terminated account
	ErrAccountTerminated = errorsmod.Register(ModuleName, 10, "account is terminated")

	// ErrAccountSuspended is returned when trying to perform operations on a suspended account
	ErrAccountSuspended = errorsmod.Register(ModuleName, 11, "account is suspended")

	// ErrNotGenesisAccount is returned when only genesis accounts can perform an action
	ErrNotGenesisAccount = errorsmod.Register(ModuleName, 12, "only genesis accounts can perform this action")

	// ErrCannotRevokeOwnRole is returned when trying to revoke own role
	ErrCannotRevokeOwnRole = errorsmod.Register(ModuleName, 13, "cannot revoke own role")

	// ErrCannotSuspendSelf is returned when trying to suspend own account
	ErrCannotSuspendSelf = errorsmod.Register(ModuleName, 14, "cannot suspend own account")
)
