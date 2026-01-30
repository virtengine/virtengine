// Package types contains types for the HPC module.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// Type aliases to generated protobuf types
type (
	MsgRegisterCluster           = hpcv1.MsgRegisterCluster
	MsgRegisterClusterResponse   = hpcv1.MsgRegisterClusterResponse
	MsgUpdateCluster             = hpcv1.MsgUpdateCluster
	MsgUpdateClusterResponse     = hpcv1.MsgUpdateClusterResponse
	MsgDeregisterCluster         = hpcv1.MsgDeregisterCluster
	MsgDeregisterClusterResponse = hpcv1.MsgDeregisterClusterResponse
	MsgCreateOffering            = hpcv1.MsgCreateOffering
	MsgCreateOfferingResponse    = hpcv1.MsgCreateOfferingResponse
	MsgUpdateOffering            = hpcv1.MsgUpdateOffering
	MsgUpdateOfferingResponse    = hpcv1.MsgUpdateOfferingResponse
	MsgSubmitJob                 = hpcv1.MsgSubmitJob
	MsgSubmitJobResponse         = hpcv1.MsgSubmitJobResponse
	MsgCancelJob                 = hpcv1.MsgCancelJob
	MsgCancelJobResponse         = hpcv1.MsgCancelJobResponse
	MsgReportJobStatus           = hpcv1.MsgReportJobStatus
	MsgReportJobStatusResponse   = hpcv1.MsgReportJobStatusResponse
	MsgUpdateNodeMetadata        = hpcv1.MsgUpdateNodeMetadata
	MsgUpdateNodeMetadataResponse = hpcv1.MsgUpdateNodeMetadataResponse
	MsgFlagDispute               = hpcv1.MsgFlagDispute
	MsgFlagDisputeResponse       = hpcv1.MsgFlagDisputeResponse
	MsgResolveDispute            = hpcv1.MsgResolveDispute
	MsgResolveDisputeResponse    = hpcv1.MsgResolveDisputeResponse
)

// Message type constants
const (
	TypeMsgRegisterCluster   = "register_cluster"
	TypeMsgUpdateCluster     = "update_cluster"
	TypeMsgDeregisterCluster = "deregister_cluster"
	TypeMsgCreateOffering    = "create_offering"
	TypeMsgUpdateOffering    = "update_offering"
	TypeMsgSubmitJob         = "submit_job"
	TypeMsgCancelJob         = "cancel_job"
	TypeMsgReportJobStatus   = "report_job_status"
	TypeMsgUpdateNodeMetadata = "update_node_metadata"
	TypeMsgFlagDispute       = "flag_dispute"
	TypeMsgResolveDispute    = "resolve_dispute"
)

var (
	_ sdk.Msg = &MsgRegisterCluster{}
	_ sdk.Msg = &MsgUpdateCluster{}
	_ sdk.Msg = &MsgDeregisterCluster{}
	_ sdk.Msg = &MsgCreateOffering{}
	_ sdk.Msg = &MsgUpdateOffering{}
	_ sdk.Msg = &MsgSubmitJob{}
	_ sdk.Msg = &MsgCancelJob{}
	_ sdk.Msg = &MsgReportJobStatus{}
	_ sdk.Msg = &MsgUpdateNodeMetadata{}
	_ sdk.Msg = &MsgFlagDispute{}
	_ sdk.Msg = &MsgResolveDispute{}
)

// NewMsgRegisterCluster creates a new MsgRegisterCluster
func NewMsgRegisterCluster(owner, name, clusterType, region, endpoint string, totalNodes, totalGpus uint64) *MsgRegisterCluster {
	return &MsgRegisterCluster{
		Owner:       owner,
		Name:        name,
		ClusterType: clusterType,
		Region:      region,
		Endpoint:    endpoint,
		TotalNodes:  totalNodes,
		TotalGpus:   totalGpus,
	}
}

// NewMsgUpdateCluster creates a new MsgUpdateCluster
func NewMsgUpdateCluster(owner, clusterID, endpoint string, totalNodes, totalGpus uint64, active bool) *MsgUpdateCluster {
	return &MsgUpdateCluster{
		Owner:      owner,
		ClusterId:  clusterID,
		Endpoint:   endpoint,
		TotalNodes: totalNodes,
		TotalGpus:  totalGpus,
		Active:     active,
	}
}

// NewMsgDeregisterCluster creates a new MsgDeregisterCluster
func NewMsgDeregisterCluster(owner, clusterID string) *MsgDeregisterCluster {
	return &MsgDeregisterCluster{
		Owner:     owner,
		ClusterId: clusterID,
	}
}

// NewMsgCreateOffering creates a new MsgCreateOffering
func NewMsgCreateOffering(provider, clusterID, name, resourceType, pricePerHour string, minDuration, maxDuration uint64) *MsgCreateOffering {
	return &MsgCreateOffering{
		Provider:     provider,
		ClusterId:    clusterID,
		Name:         name,
		ResourceType: resourceType,
		PricePerHour: pricePerHour,
		MinDuration:  minDuration,
		MaxDuration:  maxDuration,
	}
}

// NewMsgUpdateOffering creates a new MsgUpdateOffering
func NewMsgUpdateOffering(provider, offeringID, pricePerHour string, active bool) *MsgUpdateOffering {
	return &MsgUpdateOffering{
		Provider:     provider,
		OfferingId:   offeringID,
		PricePerHour: pricePerHour,
		Active:       active,
	}
}

// NewMsgSubmitJob creates a new MsgSubmitJob
func NewMsgSubmitJob(submitter, offeringID, jobScript string, requestedNodes, requestedGpus, maxDuration uint64, maxBudget string) *MsgSubmitJob {
	return &MsgSubmitJob{
		Submitter:      submitter,
		OfferingId:     offeringID,
		JobScript:      jobScript,
		RequestedNodes: requestedNodes,
		RequestedGpus:  requestedGpus,
		MaxDuration:    maxDuration,
		MaxBudget:      maxBudget,
	}
}

// NewMsgCancelJob creates a new MsgCancelJob
func NewMsgCancelJob(sender, jobID, reason string) *MsgCancelJob {
	return &MsgCancelJob{
		Sender: sender,
		JobId:  jobID,
		Reason: reason,
	}
}

// NewMsgReportJobStatus creates a new MsgReportJobStatus
func NewMsgReportJobStatus(reporter, jobID, status string, progressPercent uint64, outputLocation, errorMessage string) *MsgReportJobStatus {
	return &MsgReportJobStatus{
		Reporter:        reporter,
		JobId:           jobID,
		Status:          status,
		ProgressPercent: progressPercent,
		OutputLocation:  outputLocation,
		ErrorMessage:    errorMessage,
	}
}

// NewMsgUpdateNodeMetadata creates a new MsgUpdateNodeMetadata
func NewMsgUpdateNodeMetadata(owner, clusterID, nodeID, gpuModel string, gpuMemoryGb uint64, cpuModel string, memoryGb uint64) *MsgUpdateNodeMetadata {
	return &MsgUpdateNodeMetadata{
		Owner:       owner,
		ClusterId:   clusterID,
		NodeId:      nodeID,
		GpuModel:    gpuModel,
		GpuMemoryGb: gpuMemoryGb,
		CpuModel:    cpuModel,
		MemoryGb:    memoryGb,
	}
}

// NewMsgFlagDispute creates a new MsgFlagDispute
func NewMsgFlagDispute(sender, jobID, reason, evidence string) *MsgFlagDispute {
	return &MsgFlagDispute{
		Sender:   sender,
		JobId:    jobID,
		Reason:   reason,
		Evidence: evidence,
	}
}

// NewMsgResolveDispute creates a new MsgResolveDispute
func NewMsgResolveDispute(authority, disputeID, resolution, refundAmount string) *MsgResolveDispute {
	return &MsgResolveDispute{
		Authority:    authority,
		DisputeId:    disputeID,
		Resolution:   resolution,
		RefundAmount: refundAmount,
	}
}
