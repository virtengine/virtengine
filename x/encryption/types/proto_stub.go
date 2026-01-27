// Package types contains proto.Message stub implementations for the encryption module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgRegisterRecipientKey
func (m *MsgRegisterRecipientKey) ProtoMessage()  {}
func (m *MsgRegisterRecipientKey) Reset()         { *m = MsgRegisterRecipientKey{} }
func (m *MsgRegisterRecipientKey) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeRecipientKey
func (m *MsgRevokeRecipientKey) ProtoMessage()  {}
func (m *MsgRevokeRecipientKey) Reset()         { *m = MsgRevokeRecipientKey{} }
func (m *MsgRevokeRecipientKey) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateKeyLabel
func (m *MsgUpdateKeyLabel) ProtoMessage()  {}
func (m *MsgUpdateKeyLabel) Reset()         { *m = MsgUpdateKeyLabel{} }
func (m *MsgUpdateKeyLabel) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs - add proto.Message methods to existing types

// Proto.Message interface stubs for MsgRegisterRecipientKeyResponse
func (m *MsgRegisterRecipientKeyResponse) ProtoMessage()  {}
func (m *MsgRegisterRecipientKeyResponse) Reset()         { *m = MsgRegisterRecipientKeyResponse{} }
func (m *MsgRegisterRecipientKeyResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeRecipientKeyResponse
func (m *MsgRevokeRecipientKeyResponse) ProtoMessage()  {}
func (m *MsgRevokeRecipientKeyResponse) Reset()         { *m = MsgRevokeRecipientKeyResponse{} }
func (m *MsgRevokeRecipientKeyResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateKeyLabelResponse
func (m *MsgUpdateKeyLabelResponse) ProtoMessage()  {}
func (m *MsgUpdateKeyLabelResponse) Reset()         { *m = MsgUpdateKeyLabelResponse{} }
func (m *MsgUpdateKeyLabelResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Event type stubs

// Proto.Message interface stubs for EventKeyRegistered
func (m *EventKeyRegistered) ProtoMessage()  {}
func (m *EventKeyRegistered) Reset()         { *m = EventKeyRegistered{} }
func (m *EventKeyRegistered) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventKeyRevoked
func (m *EventKeyRevoked) ProtoMessage()  {}
func (m *EventKeyRevoked) Reset()         { *m = EventKeyRevoked{} }
func (m *EventKeyRevoked) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for EventKeyUpdated
func (m *EventKeyUpdated) ProtoMessage()  {}
func (m *EventKeyUpdated) Reset()         { *m = EventKeyUpdated{} }
func (m *EventKeyUpdated) String() string { return fmt.Sprintf("%+v", *m) }

// Genesis state stubs

// Proto.Message interface stubs for GenesisState
func (m *GenesisState) ProtoMessage()  {}
func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for Params
func (m *Params) ProtoMessage()  {}
func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for RecipientKeyRecord
func (m *RecipientKeyRecord) ProtoMessage()  {}
func (m *RecipientKeyRecord) Reset()         { *m = RecipientKeyRecord{} }
func (m *RecipientKeyRecord) String() string { return fmt.Sprintf("%+v", *m) }
