// Package types contains proto.Message stub implementations for the hpc module.
//
// These are temporary stub implementations until proper protobuf generation is set up.
// They implement the proto.Message interface required by Cosmos SDK.
package types

import "fmt"

// Proto.Message interface stubs for MsgRegisterCluster
func (m *MsgRegisterCluster) ProtoMessage()  {}
func (m *MsgRegisterCluster) Reset()         { *m = MsgRegisterCluster{} }
func (m *MsgRegisterCluster) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateCluster
func (m *MsgUpdateCluster) ProtoMessage()  {}
func (m *MsgUpdateCluster) Reset()         { *m = MsgUpdateCluster{} }
func (m *MsgUpdateCluster) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgDeregisterCluster
func (m *MsgDeregisterCluster) ProtoMessage()  {}
func (m *MsgDeregisterCluster) Reset()         { *m = MsgDeregisterCluster{} }
func (m *MsgDeregisterCluster) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgCreateOffering
func (m *MsgCreateOffering) ProtoMessage()  {}
func (m *MsgCreateOffering) Reset()         { *m = MsgCreateOffering{} }
func (m *MsgCreateOffering) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateOffering
func (m *MsgUpdateOffering) ProtoMessage()  {}
func (m *MsgUpdateOffering) Reset()         { *m = MsgUpdateOffering{} }
func (m *MsgUpdateOffering) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgSubmitJob
func (m *MsgSubmitJob) ProtoMessage()  {}
func (m *MsgSubmitJob) Reset()         { *m = MsgSubmitJob{} }
func (m *MsgSubmitJob) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgCancelJob
func (m *MsgCancelJob) ProtoMessage()  {}
func (m *MsgCancelJob) Reset()         { *m = MsgCancelJob{} }
func (m *MsgCancelJob) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgReportJobStatus
func (m *MsgReportJobStatus) ProtoMessage()  {}
func (m *MsgReportJobStatus) Reset()         { *m = MsgReportJobStatus{} }
func (m *MsgReportJobStatus) String() string { return fmt.Sprintf("%+v", *m) }

// Response type stubs

// MsgRegisterClusterResponse is the response for MsgRegisterCluster
type MsgRegisterClusterResponse struct {
	ClusterID string `json:"cluster_id"`
}

func (m *MsgRegisterClusterResponse) ProtoMessage()  {}
func (m *MsgRegisterClusterResponse) Reset()         { *m = MsgRegisterClusterResponse{} }
func (m *MsgRegisterClusterResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgUpdateClusterResponse is the response for MsgUpdateCluster
type MsgUpdateClusterResponse struct{}

func (m *MsgUpdateClusterResponse) ProtoMessage()  {}
func (m *MsgUpdateClusterResponse) Reset()         { *m = MsgUpdateClusterResponse{} }
func (m *MsgUpdateClusterResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgDeregisterClusterResponse is the response for MsgDeregisterCluster
type MsgDeregisterClusterResponse struct{}

func (m *MsgDeregisterClusterResponse) ProtoMessage()  {}
func (m *MsgDeregisterClusterResponse) Reset()         { *m = MsgDeregisterClusterResponse{} }
func (m *MsgDeregisterClusterResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgCreateOfferingResponse is the response for MsgCreateOffering
type MsgCreateOfferingResponse struct {
	OfferingID string `json:"offering_id"`
}

func (m *MsgCreateOfferingResponse) ProtoMessage()  {}
func (m *MsgCreateOfferingResponse) Reset()         { *m = MsgCreateOfferingResponse{} }
func (m *MsgCreateOfferingResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgUpdateOfferingResponse is the response for MsgUpdateOffering
type MsgUpdateOfferingResponse struct{}

func (m *MsgUpdateOfferingResponse) ProtoMessage()  {}
func (m *MsgUpdateOfferingResponse) Reset()         { *m = MsgUpdateOfferingResponse{} }
func (m *MsgUpdateOfferingResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgSubmitJobResponse is the response for MsgSubmitJob
type MsgSubmitJobResponse struct {
	JobID string `json:"job_id"`
}

func (m *MsgSubmitJobResponse) ProtoMessage()  {}
func (m *MsgSubmitJobResponse) Reset()         { *m = MsgSubmitJobResponse{} }
func (m *MsgSubmitJobResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgCancelJobResponse is the response for MsgCancelJob
type MsgCancelJobResponse struct{}

func (m *MsgCancelJobResponse) ProtoMessage()  {}
func (m *MsgCancelJobResponse) Reset()         { *m = MsgCancelJobResponse{} }
func (m *MsgCancelJobResponse) String() string { return fmt.Sprintf("%+v", *m) }

// MsgReportJobStatusResponse is the response for MsgReportJobStatus
type MsgReportJobStatusResponse struct{}

func (m *MsgReportJobStatusResponse) ProtoMessage()  {}
func (m *MsgReportJobStatusResponse) Reset()         { *m = MsgReportJobStatusResponse{} }
func (m *MsgReportJobStatusResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgUpdateNodeMetadata
func (m *MsgUpdateNodeMetadata) ProtoMessage()  {}
func (m *MsgUpdateNodeMetadata) Reset()         { *m = MsgUpdateNodeMetadata{} }
func (m *MsgUpdateNodeMetadata) String() string { return fmt.Sprintf("%+v", *m) }

// MsgUpdateNodeMetadataResponse is the response for MsgUpdateNodeMetadata
type MsgUpdateNodeMetadataResponse struct{}

func (m *MsgUpdateNodeMetadataResponse) ProtoMessage()  {}
func (m *MsgUpdateNodeMetadataResponse) Reset()         { *m = MsgUpdateNodeMetadataResponse{} }
func (m *MsgUpdateNodeMetadataResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgFlagDispute
func (m *MsgFlagDispute) ProtoMessage()  {}
func (m *MsgFlagDispute) Reset()         { *m = MsgFlagDispute{} }
func (m *MsgFlagDispute) String() string { return fmt.Sprintf("%+v", *m) }

// MsgFlagDisputeResponse is the response for MsgFlagDispute
type MsgFlagDisputeResponse struct {
	DisputeID string `json:"dispute_id"`
}

func (m *MsgFlagDisputeResponse) ProtoMessage()  {}
func (m *MsgFlagDisputeResponse) Reset()         { *m = MsgFlagDisputeResponse{} }
func (m *MsgFlagDisputeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for MsgResolveDispute
func (m *MsgResolveDispute) ProtoMessage()  {}
func (m *MsgResolveDispute) Reset()         { *m = MsgResolveDispute{} }
func (m *MsgResolveDispute) String() string { return fmt.Sprintf("%+v", *m) }

// MsgResolveDisputeResponse is the response for MsgResolveDispute
type MsgResolveDisputeResponse struct{}

func (m *MsgResolveDisputeResponse) ProtoMessage()  {}
func (m *MsgResolveDisputeResponse) Reset()         { *m = MsgResolveDisputeResponse{} }
func (m *MsgResolveDisputeResponse) String() string { return fmt.Sprintf("%+v", *m) }

// Proto.Message interface stubs for GenesisState
func (m *GenesisState) ProtoMessage()  {}
func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return fmt.Sprintf("%+v", *m) }
