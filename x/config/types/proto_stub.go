// Package types contains proto.Message stub implementations for the config module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgRegisterApprovedClient
func (m *MsgRegisterApprovedClient) ProtoMessage()  {}
func (m *MsgRegisterApprovedClient) Reset()         { *m = MsgRegisterApprovedClient{} }
func (m *MsgRegisterApprovedClient) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateApprovedClient
func (m *MsgUpdateApprovedClient) ProtoMessage()  {}
func (m *MsgUpdateApprovedClient) Reset()         { *m = MsgUpdateApprovedClient{} }
func (m *MsgUpdateApprovedClient) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSuspendApprovedClient
func (m *MsgSuspendApprovedClient) ProtoMessage()  {}
func (m *MsgSuspendApprovedClient) Reset()         { *m = MsgSuspendApprovedClient{} }
func (m *MsgSuspendApprovedClient) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeApprovedClient
func (m *MsgRevokeApprovedClient) ProtoMessage()  {}
func (m *MsgRevokeApprovedClient) Reset()         { *m = MsgRevokeApprovedClient{} }
func (m *MsgRevokeApprovedClient) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgReactivateApprovedClient
func (m *MsgReactivateApprovedClient) ProtoMessage()  {}
func (m *MsgReactivateApprovedClient) Reset()         { *m = MsgReactivateApprovedClient{} }
func (m *MsgReactivateApprovedClient) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateParams
func (m *MsgUpdateParams) ProtoMessage()  {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// Proto.Message interface stubs for MsgRegisterApprovedClientResponse
func (m *MsgRegisterApprovedClientResponse) ProtoMessage()  {}
func (m *MsgRegisterApprovedClientResponse) Reset()         { *m = MsgRegisterApprovedClientResponse{} }
func (m *MsgRegisterApprovedClientResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateApprovedClientResponse
func (m *MsgUpdateApprovedClientResponse) ProtoMessage()  {}
func (m *MsgUpdateApprovedClientResponse) Reset()         { *m = MsgUpdateApprovedClientResponse{} }
func (m *MsgUpdateApprovedClientResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSuspendApprovedClientResponse
func (m *MsgSuspendApprovedClientResponse) ProtoMessage()  {}
func (m *MsgSuspendApprovedClientResponse) Reset()         { *m = MsgSuspendApprovedClientResponse{} }
func (m *MsgSuspendApprovedClientResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeApprovedClientResponse
func (m *MsgRevokeApprovedClientResponse) ProtoMessage()  {}
func (m *MsgRevokeApprovedClientResponse) Reset()         { *m = MsgRevokeApprovedClientResponse{} }
func (m *MsgRevokeApprovedClientResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgReactivateApprovedClientResponse
func (m *MsgReactivateApprovedClientResponse) ProtoMessage()  {}
func (m *MsgReactivateApprovedClientResponse) Reset()         { *m = MsgReactivateApprovedClientResponse{} }
func (m *MsgReactivateApprovedClientResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateParamsResponse
func (m *MsgUpdateParamsResponse) ProtoMessage()  {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Event type stubs

// Proto.Message interface stubs for EventClientRegistered
func (m *EventClientRegistered) ProtoMessage()  {}
func (m *EventClientRegistered) Reset()         { *m = EventClientRegistered{} }
func (m *EventClientRegistered) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventClientUpdated
func (m *EventClientUpdated) ProtoMessage()  {}
func (m *EventClientUpdated) Reset()         { *m = EventClientUpdated{} }
func (m *EventClientUpdated) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventClientSuspended
func (m *EventClientSuspended) ProtoMessage()  {}
func (m *EventClientSuspended) Reset()         { *m = EventClientSuspended{} }
func (m *EventClientSuspended) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventClientRevoked
func (m *EventClientRevoked) ProtoMessage()  {}
func (m *EventClientRevoked) Reset()         { *m = EventClientRevoked{} }
func (m *EventClientRevoked) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventClientReactivated
func (m *EventClientReactivated) ProtoMessage()  {}
func (m *EventClientReactivated) Reset()         { *m = EventClientReactivated{} }
func (m *EventClientReactivated) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventSignatureVerified
func (m *EventSignatureVerified) ProtoMessage()  {}
func (m *EventSignatureVerified) Reset()         { *m = EventSignatureVerified{} }
func (m *EventSignatureVerified) String() string { return fmt.Sprintf("%+v", *m) }

// Genesis state stubs

// Proto.Message interface stubs for GenesisState
func (m *GenesisState) ProtoMessage()  {}
func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return fmt.Sprintf("%+v", *m) }
