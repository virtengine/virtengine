// Package types contains proto.Message stub implementations for the staking module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgUpdateParams
func (m *MsgUpdateParams) ProtoMessage()  {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSlashValidator
func (m *MsgSlashValidator) ProtoMessage()  {}
func (m *MsgSlashValidator) Reset()         { *m = MsgSlashValidator{} }
func (m *MsgSlashValidator) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUnjailValidator
func (m *MsgUnjailValidator) ProtoMessage()  {}
func (m *MsgUnjailValidator) Reset()         { *m = MsgUnjailValidator{} }
func (m *MsgUnjailValidator) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRecordPerformance
func (m *MsgRecordPerformance) ProtoMessage()  {}
func (m *MsgRecordPerformance) Reset()         { *m = MsgRecordPerformance{} }
func (m *MsgRecordPerformance) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// MsgUpdateParamsResponse is the response for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

func (m *MsgUpdateParamsResponse) ProtoMessage()  {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgSlashValidatorResponse is the response for MsgSlashValidator
type MsgSlashValidatorResponse struct{}

func (m *MsgSlashValidatorResponse) ProtoMessage()  {}
func (m *MsgSlashValidatorResponse) Reset()         { *m = MsgSlashValidatorResponse{} }
func (m *MsgSlashValidatorResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgUnjailValidatorResponse is the response for MsgUnjailValidator
type MsgUnjailValidatorResponse struct{}

func (m *MsgUnjailValidatorResponse) ProtoMessage()  {}
func (m *MsgUnjailValidatorResponse) Reset()         { *m = MsgUnjailValidatorResponse{} }
func (m *MsgUnjailValidatorResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgRecordPerformanceResponse is the response for MsgRecordPerformance
type MsgRecordPerformanceResponse struct{}

func (m *MsgRecordPerformanceResponse) ProtoMessage()  {}
func (m *MsgRecordPerformanceResponse) Reset()         { *m = MsgRecordPerformanceResponse{} }
func (m *MsgRecordPerformanceResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Genesis state stubs

// Proto.Message interface stubs for GenesisState
func (m *GenesisState) ProtoMessage()  {}
func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return fmt.Sprintf("%+v", *m) }
