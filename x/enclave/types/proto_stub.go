// Package types contains proto.Message stub implementations for the enclave module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgRegisterEnclaveIdentity
func (m *MsgRegisterEnclaveIdentity) ProtoMessage()  {}
func (m *MsgRegisterEnclaveIdentity) Reset()         { *m = MsgRegisterEnclaveIdentity{} }
func (m *MsgRegisterEnclaveIdentity) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRotateEnclaveIdentity
func (m *MsgRotateEnclaveIdentity) ProtoMessage()  {}
func (m *MsgRotateEnclaveIdentity) Reset()         { *m = MsgRotateEnclaveIdentity{} }
func (m *MsgRotateEnclaveIdentity) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgProposeMeasurement
func (m *MsgProposeMeasurement) ProtoMessage()  {}
func (m *MsgProposeMeasurement) Reset()         { *m = MsgProposeMeasurement{} }
func (m *MsgProposeMeasurement) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeMeasurement
func (m *MsgRevokeMeasurement) ProtoMessage()  {}
func (m *MsgRevokeMeasurement) Reset()         { *m = MsgRevokeMeasurement{} }
func (m *MsgRevokeMeasurement) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs - add proto.Message methods to existing types (defined in server.go)

// Proto.Message interface stubs for MsgRegisterEnclaveIdentityResponse
func (m *MsgRegisterEnclaveIdentityResponse) ProtoMessage()  {}
func (m *MsgRegisterEnclaveIdentityResponse) Reset()         { *m = MsgRegisterEnclaveIdentityResponse{} }
func (m *MsgRegisterEnclaveIdentityResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRotateEnclaveIdentityResponse
func (m *MsgRotateEnclaveIdentityResponse) ProtoMessage()  {}
func (m *MsgRotateEnclaveIdentityResponse) Reset()         { *m = MsgRotateEnclaveIdentityResponse{} }
func (m *MsgRotateEnclaveIdentityResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgProposeMeasurementResponse
func (m *MsgProposeMeasurementResponse) ProtoMessage()  {}
func (m *MsgProposeMeasurementResponse) Reset()         { *m = MsgProposeMeasurementResponse{} }
func (m *MsgProposeMeasurementResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeMeasurementResponse
func (m *MsgRevokeMeasurementResponse) ProtoMessage()  {}
func (m *MsgRevokeMeasurementResponse) Reset()         { *m = MsgRevokeMeasurementResponse{} }
func (m *MsgRevokeMeasurementResponse) String() string { return fmt.Sprintf("%+v", *m) }
