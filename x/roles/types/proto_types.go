// Package types provides type aliases and extensions for the roles module.
//
// This file imports generated protobuf types from sdk/go/node/roles/v1 and creates
// type aliases for use throughout x/roles. This approach ensures:
// - Generated types are the source of truth for on-chain data structures
// - Additional methods (Validate, GetSigners, etc.) can be added via extension functions
// - Backward compatibility with existing keeper code
package types

import (
	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
)

// ============================================================================
// Proto Type Aliases - Enums from types.pb.go
// ============================================================================

// RolePB is the protobuf-generated enum for roles
type RolePB = rolesv1.Role

// Proto enum constants for Role
const (
	RolePBUnspecified     = rolesv1.RoleUnspecified
	RolePBGenesisAccount  = rolesv1.RoleGenesisAccount
	RolePBAdministrator   = rolesv1.RoleAdministrator
	RolePBModerator       = rolesv1.RoleModerator
	RolePBValidator       = rolesv1.RoleValidator
	RolePBServiceProvider = rolesv1.RoleServiceProvider
	RolePBCustomer        = rolesv1.RoleCustomer
	RolePBSupportAgent    = rolesv1.RoleSupportAgent
)

// AccountStatePB is the protobuf-generated enum for account state
type AccountStatePB = rolesv1.AccountState

// Proto enum constants for AccountState
const (
	AccountStatePBUnspecified = rolesv1.AccountStateUnspecified
	AccountStatePBActive      = rolesv1.AccountStateActive
	AccountStatePBSuspended   = rolesv1.AccountStateSuspended
	AccountStatePBTerminated  = rolesv1.AccountStateTerminated
)

// ============================================================================
// Proto Type Aliases - Data Types from types.pb.go
// ============================================================================

// RoleAssignmentPB is the generated proto type for role assignments
type RoleAssignmentPB = rolesv1.RoleAssignment

// AccountStateRecordPB is the generated proto type for account state records
type AccountStateRecordPB = rolesv1.AccountStateRecord

// ParamsPB is the generated proto type for module params
type ParamsPB = rolesv1.Params

// ============================================================================
// Proto Type Aliases - Event Types from types.pb.go
// ============================================================================

// EventRoleAssignedPB is the generated proto type for role assignment events
type EventRoleAssignedPB = rolesv1.EventRoleAssigned

// EventRoleRevokedPB is the generated proto type for role revocation events
type EventRoleRevokedPB = rolesv1.EventRoleRevoked

// EventAccountStateChangedPB is the generated proto type for account state change events
type EventAccountStateChangedPB = rolesv1.EventAccountStateChanged

// EventAdminNominatedPB is the generated proto type for admin nomination events
type EventAdminNominatedPB = rolesv1.EventAdminNominated

// ============================================================================
// Proto Type Aliases - Message Types from tx.pb.go
// ============================================================================

// MsgAssignRolePB is the generated proto type for role assignment message
type MsgAssignRolePB = rolesv1.MsgAssignRole

// MsgAssignRoleResponsePB is the generated proto response type
type MsgAssignRoleResponsePB = rolesv1.MsgAssignRoleResponse

// MsgRevokeRolePB is the generated proto type for role revocation message
type MsgRevokeRolePB = rolesv1.MsgRevokeRole

// MsgRevokeRoleResponsePB is the generated proto response type
type MsgRevokeRoleResponsePB = rolesv1.MsgRevokeRoleResponse

// MsgSetAccountStatePB is the generated proto type for setting account state
type MsgSetAccountStatePB = rolesv1.MsgSetAccountState

// MsgSetAccountStateResponsePB is the generated proto response type
type MsgSetAccountStateResponsePB = rolesv1.MsgSetAccountStateResponse

// MsgNominateAdminPB is the generated proto type for nominating an admin
type MsgNominateAdminPB = rolesv1.MsgNominateAdmin

// MsgNominateAdminResponsePB is the generated proto response type
type MsgNominateAdminResponsePB = rolesv1.MsgNominateAdminResponse

// MsgUpdateParamsPB is the generated proto type for updating params
type MsgUpdateParamsPB = rolesv1.MsgUpdateParams

// MsgUpdateParamsResponsePB is the generated proto response type
type MsgUpdateParamsResponsePB = rolesv1.MsgUpdateParamsResponse

// ============================================================================
// Proto Type Aliases - Genesis from genesis.pb.go
// ============================================================================

// GenesisStatePB is the generated proto type for genesis state
type GenesisStatePB = rolesv1.GenesisState

// ============================================================================
// Proto Service Descriptors
// ============================================================================

// Msg_serviceDesc is the gRPC service descriptor for the Msg service
var Msg_serviceDesc = rolesv1.Msg_serviceDesc

// MsgServerPB is the generated proto server interface
type MsgServerPB = rolesv1.MsgServer

// RegisterMsgServerPB registers the protobuf-generated MsgServer
var RegisterMsgServerPB = rolesv1.RegisterMsgServer
