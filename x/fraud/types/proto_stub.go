// Package types contains proto.Message stub implementations for the fraud module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgSubmitFraudReport
func (m *MsgSubmitFraudReport) ProtoMessage()  {}
func (m *MsgSubmitFraudReport) Reset()         { *m = MsgSubmitFraudReport{} }
func (m *MsgSubmitFraudReport) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAssignModerator
func (m *MsgAssignModerator) ProtoMessage()  {}
func (m *MsgAssignModerator) Reset()         { *m = MsgAssignModerator{} }
func (m *MsgAssignModerator) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateReportStatus
func (m *MsgUpdateReportStatus) ProtoMessage()  {}
func (m *MsgUpdateReportStatus) Reset()         { *m = MsgUpdateReportStatus{} }
func (m *MsgUpdateReportStatus) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgResolveFraudReport
func (m *MsgResolveFraudReport) ProtoMessage()  {}
func (m *MsgResolveFraudReport) Reset()         { *m = MsgResolveFraudReport{} }
func (m *MsgResolveFraudReport) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRejectFraudReport
func (m *MsgRejectFraudReport) ProtoMessage()  {}
func (m *MsgRejectFraudReport) Reset()         { *m = MsgRejectFraudReport{} }
func (m *MsgRejectFraudReport) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgEscalateFraudReport
func (m *MsgEscalateFraudReport) ProtoMessage()  {}
func (m *MsgEscalateFraudReport) Reset()         { *m = MsgEscalateFraudReport{} }
func (m *MsgEscalateFraudReport) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateParams
func (m *MsgUpdateParams) ProtoMessage()  {}
func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs - add proto.Message methods to existing types (defined in codec.go)

// Proto.Message interface stubs for MsgSubmitFraudReportResponse
func (m *MsgSubmitFraudReportResponse) ProtoMessage()  {}
func (m *MsgSubmitFraudReportResponse) Reset()         { *m = MsgSubmitFraudReportResponse{} }
func (m *MsgSubmitFraudReportResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAssignModeratorResponse
func (m *MsgAssignModeratorResponse) ProtoMessage()  {}
func (m *MsgAssignModeratorResponse) Reset()         { *m = MsgAssignModeratorResponse{} }
func (m *MsgAssignModeratorResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateReportStatusResponse
func (m *MsgUpdateReportStatusResponse) ProtoMessage()  {}
func (m *MsgUpdateReportStatusResponse) Reset()         { *m = MsgUpdateReportStatusResponse{} }
func (m *MsgUpdateReportStatusResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgResolveFraudReportResponse
func (m *MsgResolveFraudReportResponse) ProtoMessage()  {}
func (m *MsgResolveFraudReportResponse) Reset()         { *m = MsgResolveFraudReportResponse{} }
func (m *MsgResolveFraudReportResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRejectFraudReportResponse
func (m *MsgRejectFraudReportResponse) ProtoMessage()  {}
func (m *MsgRejectFraudReportResponse) Reset()         { *m = MsgRejectFraudReportResponse{} }
func (m *MsgRejectFraudReportResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgEscalateFraudReportResponse
func (m *MsgEscalateFraudReportResponse) ProtoMessage()  {}
func (m *MsgEscalateFraudReportResponse) Reset()         { *m = MsgEscalateFraudReportResponse{} }
func (m *MsgEscalateFraudReportResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateParamsResponse
func (m *MsgUpdateParamsResponse) ProtoMessage()  {}
func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return fmt.Sprintf("%+v", *m) }
