// Package types contains proto.Message stub implementations for the roles module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgAssignRole
func (m *MsgAssignRole) ProtoMessage()  {}
func (m *MsgAssignRole) Reset()         { *m = MsgAssignRole{} }
func (m *MsgAssignRole) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeRole
func (m *MsgRevokeRole) ProtoMessage()  {}
func (m *MsgRevokeRole) Reset()         { *m = MsgRevokeRole{} }
func (m *MsgRevokeRole) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSetAccountState
func (m *MsgSetAccountState) ProtoMessage()  {}
func (m *MsgSetAccountState) Reset()         { *m = MsgSetAccountState{} }
func (m *MsgSetAccountState) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgNominateAdmin
func (m *MsgNominateAdmin) ProtoMessage()  {}
func (m *MsgNominateAdmin) Reset()         { *m = MsgNominateAdmin{} }
func (m *MsgNominateAdmin) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// Proto.Message interface stubs for MsgAssignRoleResponse
func (m *MsgAssignRoleResponse) ProtoMessage()  {}
func (m *MsgAssignRoleResponse) Reset()         { *m = MsgAssignRoleResponse{} }
func (m *MsgAssignRoleResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeRoleResponse
func (m *MsgRevokeRoleResponse) ProtoMessage()  {}
func (m *MsgRevokeRoleResponse) Reset()         { *m = MsgRevokeRoleResponse{} }
func (m *MsgRevokeRoleResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSetAccountStateResponse
func (m *MsgSetAccountStateResponse) ProtoMessage()  {}
func (m *MsgSetAccountStateResponse) Reset()         { *m = MsgSetAccountStateResponse{} }
func (m *MsgSetAccountStateResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgNominateAdminResponse
func (m *MsgNominateAdminResponse) ProtoMessage()  {}
func (m *MsgNominateAdminResponse) Reset()         { *m = MsgNominateAdminResponse{} }
func (m *MsgNominateAdminResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Event type stubs

// Proto.Message interface stubs for EventRoleAssigned
func (m *EventRoleAssigned) ProtoMessage()  {}
func (m *EventRoleAssigned) Reset()         { *m = EventRoleAssigned{} }
func (m *EventRoleAssigned) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventRoleRevoked
func (m *EventRoleRevoked) ProtoMessage()  {}
func (m *EventRoleRevoked) Reset()         { *m = EventRoleRevoked{} }
func (m *EventRoleRevoked) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventAccountStateChanged
func (m *EventAccountStateChanged) ProtoMessage()  {}
func (m *EventAccountStateChanged) Reset()         { *m = EventAccountStateChanged{} }
func (m *EventAccountStateChanged) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventAdminNominated
func (m *EventAdminNominated) ProtoMessage()  {}
func (m *EventAdminNominated) Reset()         { *m = EventAdminNominated{} }
func (m *EventAdminNominated) String() string { return fmt.Sprintf("%+v", *m) }
