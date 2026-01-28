// Package types contains proto.Message stub implementations for the mfa module.
//
// These implement the proto.Message interface for the local message types
// to support Amino JSON serialization. The actual proto types are in
// sdk/go/node/mfa/v1/*.pb.go.
package types

import "fmt"

// Proto.Message interface stubs for MsgEnrollFactor
func (m *MsgEnrollFactor) ProtoMessage()  {}
func (m *MsgEnrollFactor) Reset()         { *m = MsgEnrollFactor{} }
func (m *MsgEnrollFactor) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgEnrollFactorResponse
func (m *MsgEnrollFactorResponse) ProtoMessage()  {}
func (m *MsgEnrollFactorResponse) Reset()         { *m = MsgEnrollFactorResponse{} }
func (m *MsgEnrollFactorResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeFactor
func (m *MsgRevokeFactor) ProtoMessage()  {}
func (m *MsgRevokeFactor) Reset()         { *m = MsgRevokeFactor{} }
func (m *MsgRevokeFactor) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRevokeFactorResponse
func (m *MsgRevokeFactorResponse) ProtoMessage()  {}
func (m *MsgRevokeFactorResponse) Reset()         { *m = MsgRevokeFactorResponse{} }
func (m *MsgRevokeFactorResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSetMFAPolicy
func (m *MsgSetMFAPolicy) ProtoMessage()  {}
func (m *MsgSetMFAPolicy) Reset()         { *m = MsgSetMFAPolicy{} }
func (m *MsgSetMFAPolicy) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSetMFAPolicyResponse
func (m *MsgSetMFAPolicyResponse) ProtoMessage()  {}
func (m *MsgSetMFAPolicyResponse) Reset()         { *m = MsgSetMFAPolicyResponse{} }
func (m *MsgSetMFAPolicyResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgCreateChallenge
func (m *MsgCreateChallenge) ProtoMessage()  {}
func (m *MsgCreateChallenge) Reset()         { *m = MsgCreateChallenge{} }
func (m *MsgCreateChallenge) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgCreateChallengeResponse
func (m *MsgCreateChallengeResponse) ProtoMessage()  {}
func (m *MsgCreateChallengeResponse) Reset()         { *m = MsgCreateChallengeResponse{} }
func (m *MsgCreateChallengeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgVerifyChallenge
func (m *MsgVerifyChallenge) ProtoMessage()  {}
func (m *MsgVerifyChallenge) Reset()         { *m = MsgVerifyChallenge{} }
func (m *MsgVerifyChallenge) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgVerifyChallengeResponse
func (m *MsgVerifyChallengeResponse) ProtoMessage()  {}
func (m *MsgVerifyChallengeResponse) Reset()         { *m = MsgVerifyChallengeResponse{} }
func (m *MsgVerifyChallengeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAddTrustedDevice
func (m *MsgAddTrustedDevice) ProtoMessage()  {}
func (m *MsgAddTrustedDevice) Reset()         { *m = MsgAddTrustedDevice{} }
func (m *MsgAddTrustedDevice) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAddTrustedDeviceResponse
func (m *MsgAddTrustedDeviceResponse) ProtoMessage()  {}
func (m *MsgAddTrustedDeviceResponse) Reset()         { *m = MsgAddTrustedDeviceResponse{} }
func (m *MsgAddTrustedDeviceResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRemoveTrustedDevice
func (m *MsgRemoveTrustedDevice) ProtoMessage()  {}
func (m *MsgRemoveTrustedDevice) Reset()         { *m = MsgRemoveTrustedDevice{} }
func (m *MsgRemoveTrustedDevice) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRemoveTrustedDeviceResponse
func (m *MsgRemoveTrustedDeviceResponse) ProtoMessage()  {}
func (m *MsgRemoveTrustedDeviceResponse) Reset()         { *m = MsgRemoveTrustedDeviceResponse{} }
func (m *MsgRemoveTrustedDeviceResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateSensitiveTxConfig
func (m *MsgUpdateSensitiveTxConfig) ProtoMessage()  {}
func (m *MsgUpdateSensitiveTxConfig) Reset()         { *m = MsgUpdateSensitiveTxConfig{} }
func (m *MsgUpdateSensitiveTxConfig) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateSensitiveTxConfigResponse
func (m *MsgUpdateSensitiveTxConfigResponse) ProtoMessage()  {}
func (m *MsgUpdateSensitiveTxConfigResponse) Reset()         { *m = MsgUpdateSensitiveTxConfigResponse{} }
func (m *MsgUpdateSensitiveTxConfigResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for GenesisState
func (m *GenesisState) ProtoMessage()  {}
func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return fmt.Sprintf("%+v", *m) }
