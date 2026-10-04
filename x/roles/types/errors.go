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

	// ErrInvalidSanction is returned when a sanction record is malformed
	ErrInvalidSanction = errorsmod.Register(ModuleName, 1414, "invalid sanction")

	// ErrInvalidSanctionScope is returned when a sanction scope is unknown
	ErrInvalidSanctionScope = errorsmod.Register(ModuleName, 1415, "invalid sanction scope")

	// ErrInvalidSanctionKind is returned when a sanction kind is unknown
	ErrInvalidSanctionKind = errorsmod.Register(ModuleName, 1416, "invalid sanction kind")

	// ErrInvalidSanctionReason is returned when a sanction reason code is unknown
	ErrInvalidSanctionReason = errorsmod.Register(ModuleName, 1417, "invalid sanction reason code")

	// ErrInvalidSanctionJustification is returned when the justification is missing or out of range
	ErrInvalidSanctionJustification = errorsmod.Register(ModuleName, 1418, "invalid sanction justification")

	// ErrInvalidSanctionStatus is returned when a sanction status is unknown
	ErrInvalidSanctionStatus = errorsmod.Register(ModuleName, 1419, "invalid sanction status")

	// ErrSanctionNoticeRequired is returned when a sanction requiring notice omits it
	ErrSanctionNoticeRequired = errorsmod.Register(ModuleName, 1420, "sanction notice is required")

	// ErrSecondReviewerRequired is returned when suspension/termination is attempted without a distinct second reviewer
	ErrSecondReviewerRequired = errorsmod.Register(ModuleName, 1421, "suspension and termination require a distinct second reviewer")

	// ErrSecondReviewerMustDiffer is returned when the second reviewer is the same identity as the imposer
	ErrSecondReviewerMustDiffer = errorsmod.Register(ModuleName, 1422, "second reviewer must be a distinct identity from the imposing reviewer")

	// ErrSecondReviewerUnauthorized is returned when the second reviewer lacks moderator-or-above authority
	ErrSecondReviewerUnauthorized = errorsmod.Register(ModuleName, 1423, "second reviewer is not a moderator")

	// ErrSanctionNotFound is returned when a sanction record does not exist
	ErrSanctionNotFound = errorsmod.Register(ModuleName, 1424, "sanction not found")

	// ErrInvalidSanctionTransition is returned when a sanction status transition is not allowed
	ErrInvalidSanctionTransition = errorsmod.Register(ModuleName, 1425, "invalid sanction status transition")

	// ErrSanctionNotAppealable is returned when an appeal is opened against a record that cannot be appealed
	ErrSanctionNotAppealable = errorsmod.Register(ModuleName, 1426, "sanction is not appealable")

	// ErrAppealAlreadyOpen is returned when an appeal is already open for the sanction
	ErrAppealAlreadyOpen = errorsmod.Register(ModuleName, 1427, "an appeal is already open for this sanction")

	// ErrAppealNotOpen is returned when resolving an appeal that is not open
	ErrAppealNotOpen = errorsmod.Register(ModuleName, 1428, "appeal is not open")

	// ErrEmergencyHoldExpired is returned when an emergency hold lapses before it is reviewed
	ErrEmergencyHoldExpired = errorsmod.Register(ModuleName, 1429, "emergency hold expired before review")

	// ErrEmergencyHoldTooLong is returned when an emergency hold exceeds the maximum single-handed window
	ErrEmergencyHoldTooLong = errorsmod.Register(ModuleName, 1430, "emergency hold exceeds the maximum unreviewed window")

	// ErrSanctionSubjectMismatch is returned when an appeal or review targets a different subject
	ErrSanctionSubjectMismatch = errorsmod.Register(ModuleName, 1431, "sanction subject mismatch")

	// ErrSanctionExpired is returned when acting on a sanction that has already lapsed
	ErrSanctionExpired = errorsmod.Register(ModuleName, 1432, "sanction has expired")

	// ErrSanctionRequired is returned when the bare administrative account-state
	// path is used to reach a punitive state (Suspended or Terminated). Those
	// states strip an account's access and must be reachable only through a
	// sanction record, which carries a scope, a reason code, a duration, notice
	// and a second, distinct reviewer.
	ErrSanctionRequired = errorsmod.Register(ModuleName, 1433,
		"suspension and termination must be imposed through a sanction with a scope, reason, duration, notice and a second reviewer")
)
