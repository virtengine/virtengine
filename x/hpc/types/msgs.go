// Package types contains types for the HPC module.
package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// Type aliases to generated protobuf types
type (
	MsgRegisterCluster            = hpcv1.MsgRegisterCluster
	MsgRegisterClusterResponse    = hpcv1.MsgRegisterClusterResponse
	MsgUpdateCluster              = hpcv1.MsgUpdateCluster
	MsgUpdateClusterResponse      = hpcv1.MsgUpdateClusterResponse
	MsgDeregisterCluster          = hpcv1.MsgDeregisterCluster
	MsgDeregisterClusterResponse  = hpcv1.MsgDeregisterClusterResponse
	MsgCreateOffering             = hpcv1.MsgCreateOffering
	MsgCreateOfferingResponse     = hpcv1.MsgCreateOfferingResponse
	MsgUpdateOffering             = hpcv1.MsgUpdateOffering
	MsgUpdateOfferingResponse     = hpcv1.MsgUpdateOfferingResponse
	MsgSubmitJob                  = hpcv1.MsgSubmitJob
	MsgSubmitJobResponse          = hpcv1.MsgSubmitJobResponse
	MsgCancelJob                  = hpcv1.MsgCancelJob
	MsgCancelJobResponse          = hpcv1.MsgCancelJobResponse
	MsgReportJobStatus            = hpcv1.MsgReportJobStatus
	MsgReportJobStatusResponse    = hpcv1.MsgReportJobStatusResponse
	MsgUpdateNodeMetadata         = hpcv1.MsgUpdateNodeMetadata
	MsgUpdateNodeMetadataResponse = hpcv1.MsgUpdateNodeMetadataResponse
	MsgFlagDispute                = hpcv1.MsgFlagDispute
	MsgFlagDisputeResponse        = hpcv1.MsgFlagDisputeResponse
	MsgResolveDispute             = hpcv1.MsgResolveDispute
	MsgResolveDisputeResponse     = hpcv1.MsgResolveDisputeResponse
)

// Message type constants
const (
	TypeMsgRegisterCluster    = "register_cluster"
	TypeMsgUpdateCluster      = "update_cluster"
	TypeMsgDeregisterCluster  = "deregister_cluster"
	TypeMsgCreateOffering     = "create_offering"
	TypeMsgUpdateOffering     = "update_offering"
	TypeMsgSubmitJob          = "submit_job"
	TypeMsgCancelJob          = "cancel_job"
	TypeMsgReportJobStatus    = "report_job_status"
	TypeMsgUpdateNodeMetadata = "update_node_metadata"
	TypeMsgFlagDispute        = "flag_dispute"
	TypeMsgResolveDispute     = "resolve_dispute"
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
	description := strings.TrimSpace(clusterType)
	if endpoint != "" {
		if description != "" {
			description = fmt.Sprintf("%s (endpoint=%s)", description, endpoint)
		} else {
			description = fmt.Sprintf("endpoint=%s", endpoint)
		}
	}

	return &MsgRegisterCluster{
		ProviderAddress: owner,
		Name:            name,
		Description:     description,
		Region:          region,
		Partitions:      []hpcv1.Partition{},
		TotalNodes:      clampUint64ToInt32(totalNodes),
		ClusterMetadata: hpcv1.ClusterMetadata{TotalGpus: int64(totalGpus)},
	}
}

// NewMsgUpdateCluster creates a new MsgUpdateCluster
func NewMsgUpdateCluster(owner, clusterID, endpoint string, totalNodes, totalGpus uint64, active bool) *MsgUpdateCluster {
	state := hpcv1.ClusterStateOffline
	if active {
		state = hpcv1.ClusterStateActive
	}

	description := strings.TrimSpace(endpoint)

	return &MsgUpdateCluster{
		ProviderAddress: owner,
		ClusterId:       clusterID,
		Description:     description,
		State:           state,
		Partitions:      []hpcv1.Partition{},
		TotalNodes:      clampUint64ToInt32(totalNodes),
		ClusterMetadata: hpcv1.ClusterMetadata{TotalGpus: int64(totalGpus)},
	}
}

// NewMsgDeregisterCluster creates a new MsgDeregisterCluster
func NewMsgDeregisterCluster(owner, clusterID string) *MsgDeregisterCluster {
	return &MsgDeregisterCluster{
		ProviderAddress: owner,
		ClusterId:       clusterID,
	}
}

// NewMsgCreateOffering creates a new MsgCreateOffering
func NewMsgCreateOffering(provider, clusterID, name, resourceType, pricePerHour string, minDuration, maxDuration uint64) *MsgCreateOffering {
	pricing := hpcv1.HPCPricing{
		BaseNodeHourPrice: pricePerHour,
		CpuCoreHourPrice:  pricePerHour,
		MemoryGbHourPrice: pricePerHour,
		StorageGbPrice:    pricePerHour,
		NetworkGbPrice:    pricePerHour,
	}

	if decCoin, err := sdk.ParseDecCoin(pricePerHour); err == nil {
		pricing.Currency = decCoin.Denom
	}

	return &MsgCreateOffering{
		ProviderAddress:           provider,
		ClusterId:                 clusterID,
		Name:                      name,
		Description:               resourceType,
		QueueOptions:              []hpcv1.QueueOption{},
		Pricing:                   pricing,
		RequiredIdentityThreshold: 0,
		MaxRuntimeSeconds:         int64(maxDuration),
		PreconfiguredWorkloads:    []hpcv1.PreconfiguredWorkload{},
		SupportsCustomWorkloads:   true,
	}
}

// NewMsgUpdateOffering creates a new MsgUpdateOffering
func NewMsgUpdateOffering(provider, offeringID, pricePerHour string, active bool) *MsgUpdateOffering {
	pricing := hpcv1.HPCPricing{
		BaseNodeHourPrice: pricePerHour,
		CpuCoreHourPrice:  pricePerHour,
		MemoryGbHourPrice: pricePerHour,
		StorageGbPrice:    pricePerHour,
		NetworkGbPrice:    pricePerHour,
	}

	if decCoin, err := sdk.ParseDecCoin(pricePerHour); err == nil {
		pricing.Currency = decCoin.Denom
	}

	return &MsgUpdateOffering{
		ProviderAddress:           provider,
		OfferingId:                offeringID,
		QueueOptions:              []hpcv1.QueueOption{},
		Pricing:                   pricing,
		RequiredIdentityThreshold: 0,
		Active:                    active,
	}
}

// NewMsgSubmitJob creates a new MsgSubmitJob
func NewMsgSubmitJob(submitter, offeringID, jobScript string, requestedNodes, requestedGpus, maxDuration uint64, maxBudget string) *MsgSubmitJob {
	maxPrice, err := sdk.ParseCoinsNormalized(maxBudget)
	if err != nil {
		maxPrice = sdk.NewCoins()
	}

	return &MsgSubmitJob{
		CustomerAddress: submitter,
		OfferingId:      offeringID,
		WorkloadSpec: hpcv1.JobWorkloadSpec{
			Command: jobScript,
		},
		Resources: hpcv1.JobResources{
			Nodes:       clampUint64ToInt32(requestedNodes),
			GpusPerNode: clampUint64ToInt32(requestedGpus),
		},
		MaxRuntimeSeconds: int64(maxDuration),
		MaxPrice:          maxPrice,
	}
}

// NewMsgCancelJob creates a new MsgCancelJob
func NewMsgCancelJob(sender, jobID, reason string) *MsgCancelJob {
	return &MsgCancelJob{
		RequesterAddress: sender,
		JobId:            jobID,
		Reason:           reason,
	}
}

// NewMsgReportJobStatus creates a new MsgReportJobStatus
func NewMsgReportJobStatus(reporter, jobID, status string, progressPercent uint64, outputLocation, errorMessage string) *MsgReportJobStatus {
	message := strings.TrimSpace(status)
	if outputLocation != "" {
		message = fmt.Sprintf("%s output=%s", message, outputLocation)
	}
	if errorMessage != "" {
		message = fmt.Sprintf("%s error=%s", message, errorMessage)
	}

	return &MsgReportJobStatus{
		ProviderAddress: reporter,
		JobId:           jobID,
		State:           parseJobState(status),
		StatusMessage:   strings.TrimSpace(message),
	}
}

// NewMsgUpdateNodeMetadata creates a new MsgUpdateNodeMetadata
func NewMsgUpdateNodeMetadata(owner, clusterID, nodeID, gpuModel string, gpuMemoryGb uint64, cpuModel string, memoryGb uint64) *MsgUpdateNodeMetadata {
	return &MsgUpdateNodeMetadata{
		ProviderAddress: owner,
		NodeId:          nodeID,
		ClusterId:       clusterID,
		Resources: hpcv1.NodeResources{
			CpuCores: 0,
			MemoryGb: clampUint64ToInt32(memoryGb),
			GpuType:  gpuModel,
		},
	}
}

// NewMsgFlagDispute creates a new MsgFlagDispute
func NewMsgFlagDispute(sender, jobID, reason, evidence string) *MsgFlagDispute {
	return &MsgFlagDispute{
		DisputerAddress: sender,
		JobId:           jobID,
		DisputeType:     "usage",
		Reason:          reason,
		Evidence:        evidence,
	}
}

// NewMsgResolveDispute creates a new MsgResolveDispute
func NewMsgResolveDispute(authority, disputeID, resolution, refundAmount string) *MsgResolveDispute {
	resolutionText := resolution
	if refundAmount != "" {
		resolutionText = fmt.Sprintf("%s (refund=%s)", resolution, refundAmount)
	}

	return &MsgResolveDispute{
		ResolverAddress: authority,
		DisputeId:       disputeID,
		Status:          hpcv1.DisputeStatusResolved,
		Resolution:      resolutionText,
	}
}

const maxInt32 = int32(^uint32(0) >> 1)

func clampUint64ToInt32(value uint64) int32 {
	if value > uint64(maxInt32) {
		return maxInt32
	}
	return int32(value)
}

func parseJobState(value string) hpcv1.JobState {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return hpcv1.JobStateUnspecified
	}
	if !strings.HasPrefix(normalized, "JOB_STATE_") {
		normalized = "JOB_STATE_" + normalized
	}
	if state, ok := hpcv1.JobState_value[normalized]; ok {
		return hpcv1.JobState(state)
	}
	return hpcv1.JobStateUnspecified
}
