// Package types contains proto.Message stub implementations for the settlement module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgCreateEscrow
func (m *MsgCreateEscrow) ProtoMessage()  {}
func (m *MsgCreateEscrow) Reset()         { *m = MsgCreateEscrow{} }
func (m *MsgCreateEscrow) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgActivateEscrow
func (m *MsgActivateEscrow) ProtoMessage()  {}
func (m *MsgActivateEscrow) Reset()         { *m = MsgActivateEscrow{} }
func (m *MsgActivateEscrow) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgReleaseEscrow
func (m *MsgReleaseEscrow) ProtoMessage()  {}
func (m *MsgReleaseEscrow) Reset()         { *m = MsgReleaseEscrow{} }
func (m *MsgReleaseEscrow) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRefundEscrow
func (m *MsgRefundEscrow) ProtoMessage()  {}
func (m *MsgRefundEscrow) Reset()         { *m = MsgRefundEscrow{} }
func (m *MsgRefundEscrow) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgDisputeEscrow
func (m *MsgDisputeEscrow) ProtoMessage()  {}
func (m *MsgDisputeEscrow) Reset()         { *m = MsgDisputeEscrow{} }
func (m *MsgDisputeEscrow) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSettleOrder
func (m *MsgSettleOrder) ProtoMessage()  {}
func (m *MsgSettleOrder) Reset()         { *m = MsgSettleOrder{} }
func (m *MsgSettleOrder) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRecordUsage
func (m *MsgRecordUsage) ProtoMessage()  {}
func (m *MsgRecordUsage) Reset()         { *m = MsgRecordUsage{} }
func (m *MsgRecordUsage) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAcknowledgeUsage
func (m *MsgAcknowledgeUsage) ProtoMessage()  {}
func (m *MsgAcknowledgeUsage) Reset()         { *m = MsgAcknowledgeUsage{} }
func (m *MsgAcknowledgeUsage) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgClaimRewards
func (m *MsgClaimRewards) ProtoMessage()  {}
func (m *MsgClaimRewards) Reset()         { *m = MsgClaimRewards{} }
func (m *MsgClaimRewards) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// Proto.Message interface stubs for MsgCreateEscrowResponse
func (m *MsgCreateEscrowResponse) ProtoMessage()  {}
func (m *MsgCreateEscrowResponse) Reset()         { *m = MsgCreateEscrowResponse{} }
func (m *MsgCreateEscrowResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgActivateEscrowResponse
func (m *MsgActivateEscrowResponse) ProtoMessage()  {}
func (m *MsgActivateEscrowResponse) Reset()         { *m = MsgActivateEscrowResponse{} }
func (m *MsgActivateEscrowResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgReleaseEscrowResponse
func (m *MsgReleaseEscrowResponse) ProtoMessage()  {}
func (m *MsgReleaseEscrowResponse) Reset()         { *m = MsgReleaseEscrowResponse{} }
func (m *MsgReleaseEscrowResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRefundEscrowResponse
func (m *MsgRefundEscrowResponse) ProtoMessage()  {}
func (m *MsgRefundEscrowResponse) Reset()         { *m = MsgRefundEscrowResponse{} }
func (m *MsgRefundEscrowResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgDisputeEscrowResponse
func (m *MsgDisputeEscrowResponse) ProtoMessage()  {}
func (m *MsgDisputeEscrowResponse) Reset()         { *m = MsgDisputeEscrowResponse{} }
func (m *MsgDisputeEscrowResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSettleOrderResponse
func (m *MsgSettleOrderResponse) ProtoMessage()  {}
func (m *MsgSettleOrderResponse) Reset()         { *m = MsgSettleOrderResponse{} }
func (m *MsgSettleOrderResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgRecordUsageResponse
func (m *MsgRecordUsageResponse) ProtoMessage()  {}
func (m *MsgRecordUsageResponse) Reset()         { *m = MsgRecordUsageResponse{} }
func (m *MsgRecordUsageResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgAcknowledgeUsageResponse
func (m *MsgAcknowledgeUsageResponse) ProtoMessage()  {}
func (m *MsgAcknowledgeUsageResponse) Reset()         { *m = MsgAcknowledgeUsageResponse{} }
func (m *MsgAcknowledgeUsageResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgClaimRewardsResponse
func (m *MsgClaimRewardsResponse) ProtoMessage()  {}
func (m *MsgClaimRewardsResponse) Reset()         { *m = MsgClaimRewardsResponse{} }
func (m *MsgClaimRewardsResponse) String() string { return fmt.Sprintf("%+v", *m) }
