package roles

import (
	"github.com/virtengine/virtengine/x/roles/keeper"
	"github.com/virtengine/virtengine/x/roles/types"
)

// Module aliases for types
const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

// Role aliases
const (
	RoleUnspecified     = types.RoleUnspecified
	RoleGenesisAccount  = types.RoleGenesisAccount
	RoleAdministrator   = types.RoleAdministrator
	RoleModerator       = types.RoleModerator
	RoleValidator       = types.RoleValidator
	RoleServiceProvider = types.RoleServiceProvider
	RoleCustomer        = types.RoleCustomer
	RoleSupportAgent    = types.RoleSupportAgent
)

// Account state aliases
const (
	AccountStateUnspecified = types.AccountStateUnspecified
	AccountStateActive      = types.AccountStateActive
	AccountStateSuspended   = types.AccountStateSuspended
	AccountStateTerminated  = types.AccountStateTerminated
)

// Type aliases
type (
	Role               = types.Role
	AccountState       = types.AccountState
	RoleAssignment     = types.RoleAssignment
	AccountStateRecord = types.AccountStateRecord
	GenesisState       = types.GenesisState
	Params             = types.Params

	// Messages
	MsgAssignRole      = types.MsgAssignRole
	MsgRevokeRole      = types.MsgRevokeRole
	MsgSetAccountState = types.MsgSetAccountState
	MsgNominateAdmin   = types.MsgNominateAdmin

	// Responses
	MsgAssignRoleResponse      = types.MsgAssignRoleResponse
	MsgRevokeRoleResponse      = types.MsgRevokeRoleResponse
	MsgSetAccountStateResponse = types.MsgSetAccountStateResponse
	MsgNominateAdminResponse   = types.MsgNominateAdminResponse

	// Events
	EventRoleAssigned        = types.EventRoleAssigned
	EventRoleRevoked         = types.EventRoleRevoked
	EventAccountStateChanged = types.EventAccountStateChanged
	EventAdminNominated      = types.EventAdminNominated

	// Keeper
	Keeper = keeper.Keeper
)

// Function aliases
var (
	NewKeeper           = keeper.NewKeeper
	NewMsgServerImpl    = keeper.NewMsgServerImpl
	RoleFromString      = types.RoleFromString
	AccountStateFromString = types.AccountStateFromString
	AllRoles            = types.AllRoles
	AllAccountStates    = types.AllAccountStates
	DefaultParams       = types.DefaultParams
)

// Error aliases
var (
	ErrInvalidAddress           = types.ErrInvalidAddress
	ErrInvalidRole              = types.ErrInvalidRole
	ErrInvalidAccountState      = types.ErrInvalidAccountState
	ErrUnauthorized             = types.ErrUnauthorized
	ErrRoleNotFound             = types.ErrRoleNotFound
	ErrRoleAlreadyAssigned      = types.ErrRoleAlreadyAssigned
	ErrAccountStateNotFound     = types.ErrAccountStateNotFound
	ErrInvalidStateTransition   = types.ErrInvalidStateTransition
	ErrCannotModifyGenesisAccount = types.ErrCannotModifyGenesisAccount
	ErrAccountTerminated        = types.ErrAccountTerminated
	ErrAccountSuspended         = types.ErrAccountSuspended
	ErrNotGenesisAccount        = types.ErrNotGenesisAccount
	ErrCannotRevokeOwnRole      = types.ErrCannotRevokeOwnRole
	ErrCannotSuspendSelf        = types.ErrCannotSuspendSelf
)
