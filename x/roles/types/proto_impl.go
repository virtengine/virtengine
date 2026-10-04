// Package types provides proto.Message interface implementations for the roles module.
//
// These implementations allow the local message types to be used with Cosmos SDK's
// proto codec for JSON serialization. For gRPC and binary serialization, the generated
// proto types from sdk/go/node/roles/v1 are used via the adapters in grpc_handlers.go.
package types

import "fmt"

// ============================================================================
// Proto.Message Interface Implementations - Message Types
// ============================================================================

// Proto.Message interface for MsgAssignRole
func (m *MsgAssignRole) ProtoMessage()  {}
func (m *MsgAssignRole) Reset()         { *m = MsgAssignRole{} }
func (m *MsgAssignRole) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgRevokeRole
func (m *MsgRevokeRole) ProtoMessage()  {}
func (m *MsgRevokeRole) Reset()         { *m = MsgRevokeRole{} }
func (m *MsgRevokeRole) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgSetAccountState
func (m *MsgSetAccountState) ProtoMessage()  {}
func (m *MsgSetAccountState) Reset()         { *m = MsgSetAccountState{} }
func (m *MsgSetAccountState) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgNominateAdmin
func (m *MsgNominateAdmin) ProtoMessage()  {}
func (m *MsgNominateAdmin) Reset()         { *m = MsgNominateAdmin{} }
func (m *MsgNominateAdmin) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgUpdateParams
func (m *MsgUpdateParams) ProtoMessage()  {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgImposeSanction
func (m *MsgImposeSanction) ProtoMessage()  {}
func (m *MsgImposeSanction) Reset()         { *m = MsgImposeSanction{} }
func (m *MsgImposeSanction) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgConfirmSanction
func (m *MsgConfirmSanction) ProtoMessage()  {}
func (m *MsgConfirmSanction) Reset()         { *m = MsgConfirmSanction{} }
func (m *MsgConfirmSanction) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgRevokeSanction
func (m *MsgRevokeSanction) ProtoMessage()  {}
func (m *MsgRevokeSanction) Reset()         { *m = MsgRevokeSanction{} }
func (m *MsgRevokeSanction) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgOpenSanctionAppeal
func (m *MsgOpenSanctionAppeal) ProtoMessage()  {}
func (m *MsgOpenSanctionAppeal) Reset()         { *m = MsgOpenSanctionAppeal{} }
func (m *MsgOpenSanctionAppeal) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgResolveSanctionAppeal
func (m *MsgResolveSanctionAppeal) ProtoMessage()  {}
func (m *MsgResolveSanctionAppeal) Reset()         { *m = MsgResolveSanctionAppeal{} }
func (m *MsgResolveSanctionAppeal) String() string { return fmt.Sprintf("%+v", *m) }

// ============================================================================
// Proto.Message Interface Implementations - Response Types
// ============================================================================

// Proto.Message interface for MsgAssignRoleResponse
func (m *MsgAssignRoleResponse) ProtoMessage()  {}
func (m *MsgAssignRoleResponse) Reset()         { *m = MsgAssignRoleResponse{} }
func (m *MsgAssignRoleResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgRevokeRoleResponse
func (m *MsgRevokeRoleResponse) ProtoMessage()  {}
func (m *MsgRevokeRoleResponse) Reset()         { *m = MsgRevokeRoleResponse{} }
func (m *MsgRevokeRoleResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgSetAccountStateResponse
func (m *MsgSetAccountStateResponse) ProtoMessage()  {}
func (m *MsgSetAccountStateResponse) Reset()         { *m = MsgSetAccountStateResponse{} }
func (m *MsgSetAccountStateResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgNominateAdminResponse
func (m *MsgNominateAdminResponse) ProtoMessage()  {}
func (m *MsgNominateAdminResponse) Reset()         { *m = MsgNominateAdminResponse{} }
func (m *MsgNominateAdminResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgUpdateParamsResponse
func (m *MsgUpdateParamsResponse) ProtoMessage()  {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgImposeSanctionResponse
func (m *MsgImposeSanctionResponse) ProtoMessage()  {}
func (m *MsgImposeSanctionResponse) Reset()         { *m = MsgImposeSanctionResponse{} }
func (m *MsgImposeSanctionResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgConfirmSanctionResponse
func (m *MsgConfirmSanctionResponse) ProtoMessage()  {}
func (m *MsgConfirmSanctionResponse) Reset()         { *m = MsgConfirmSanctionResponse{} }
func (m *MsgConfirmSanctionResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgRevokeSanctionResponse
func (m *MsgRevokeSanctionResponse) ProtoMessage()  {}
func (m *MsgRevokeSanctionResponse) Reset()         { *m = MsgRevokeSanctionResponse{} }
func (m *MsgRevokeSanctionResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgOpenSanctionAppealResponse
func (m *MsgOpenSanctionAppealResponse) ProtoMessage()  {}
func (m *MsgOpenSanctionAppealResponse) Reset()         { *m = MsgOpenSanctionAppealResponse{} }
func (m *MsgOpenSanctionAppealResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for MsgResolveSanctionAppealResponse
func (m *MsgResolveSanctionAppealResponse) ProtoMessage()  {}
func (m *MsgResolveSanctionAppealResponse) Reset()         { *m = MsgResolveSanctionAppealResponse{} }
func (m *MsgResolveSanctionAppealResponse) String() string { return fmt.Sprintf("%+v", *m) }

// ============================================================================
// Proto.Message Interface Implementations - Event Types
// ============================================================================

// Proto.Message interface for EventRoleAssigned
func (m *EventRoleAssigned) ProtoMessage()  {}
func (m *EventRoleAssigned) Reset()         { *m = EventRoleAssigned{} }
func (m *EventRoleAssigned) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for EventRoleRevoked
func (m *EventRoleRevoked) ProtoMessage()  {}
func (m *EventRoleRevoked) Reset()         { *m = EventRoleRevoked{} }
func (m *EventRoleRevoked) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for EventAccountStateChanged
func (m *EventAccountStateChanged) ProtoMessage()  {}
func (m *EventAccountStateChanged) Reset()         { *m = EventAccountStateChanged{} }
func (m *EventAccountStateChanged) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface for EventAdminNominated
func (m *EventAdminNominated) ProtoMessage()  {}
func (m *EventAdminNominated) Reset()         { *m = EventAdminNominated{} }
func (m *EventAdminNominated) String() string { return fmt.Sprintf("%+v", *m) }
