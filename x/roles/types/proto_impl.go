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
