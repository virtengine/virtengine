// Package keeper implements the HPC module keeper.
//
// VE-2019: MsgServer implementation for HPC module
package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	"github.com/virtengine/virtengine/x/hpc/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the HPC MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// RegisterCluster handles registering a new HPC cluster
func (ms msgServer) RegisterCluster(goCtx context.Context, msg *types.MsgRegisterCluster) (*types.MsgRegisterClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid owner address")
	}

	// Generate cluster ID
	seq := ms.keeper.GetNextClusterSequence(ctx)
	clusterID := fmt.Sprintf("HPC-%d", seq)
	// Create the cluster
	cluster := &types.HPCCluster{
		ClusterID:       clusterID,
		ProviderAddress: ownerAddr.String(),
		Name:            msg.Name,
		Description:     msg.Description,
		Region:          msg.Region,
		Partitions:      partitionsFromProto(msg.Partitions),
		TotalNodes:      msg.TotalNodes,
		AvailableNodes:  msg.TotalNodes,
		State:           types.ClusterStateActive,
		ClusterMetadata: clusterMetadataFromProto(msg.ClusterMetadata),
		SLURMVersion:    msg.SlurmVersion,
	}

	// Register the cluster
	if err := ms.keeper.RegisterCluster(ctx, cluster); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster registered",
		"cluster_id", clusterID,
		"provider", msg.ProviderAddress,
		"name", msg.Name,
	)

	return &types.MsgRegisterClusterResponse{ClusterId: clusterID}, nil
}

// UpdateCluster handles updating an HPC cluster
func (ms msgServer) UpdateCluster(goCtx context.Context, msg *types.MsgUpdateCluster) (*types.MsgUpdateClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid owner address")
	}

	// Get existing cluster
	cluster, found := ms.keeper.GetCluster(ctx, msg.ClusterId)
	if !found {
		return nil, types.ErrClusterNotFound.Wrapf("cluster %s not found", msg.ClusterId)
	}

	// Verify ownership
	if cluster.ProviderAddress != ownerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the cluster owner")
	}

	// Apply updates
	if msg.Name != "" {
		cluster.Name = msg.Name
	}
	if msg.Description != "" {
		cluster.Description = msg.Description
	}
	if msg.State != hpcv1.ClusterStateUnspecified {
		cluster.State = clusterStateFromProto(msg.State)
	}
	if len(msg.Partitions) > 0 {
		cluster.Partitions = partitionsFromProto(msg.Partitions)
	}
	if msg.TotalNodes > 0 {
		cluster.TotalNodes = msg.TotalNodes
	}
	if msg.AvailableNodes != 0 {
		cluster.AvailableNodes = msg.AvailableNodes
	}
	if msg.ClusterMetadata != nil {
		cluster.ClusterMetadata = clusterMetadataFromProto(*msg.ClusterMetadata)
	}

	// Update the cluster
	if err := ms.keeper.UpdateCluster(ctx, &cluster); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster updated",
		"cluster_id", msg.ClusterId,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgUpdateClusterResponse{}, nil
}

// DeregisterCluster handles deregistering an HPC cluster
func (ms msgServer) DeregisterCluster(goCtx context.Context, msg *types.MsgDeregisterCluster) (*types.MsgDeregisterClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid owner address")
	}

	// Deregister the cluster
	if err := ms.keeper.DeregisterCluster(ctx, msg.ClusterId, ownerAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster deregistered",
		"cluster_id", msg.ClusterId,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgDeregisterClusterResponse{}, nil
}

// CreateOffering handles creating a new HPC offering
func (ms msgServer) CreateOffering(goCtx context.Context, msg *types.MsgCreateOffering) (*types.MsgCreateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidOffering.Wrap("invalid provider address")
	}

	// Verify cluster ownership
	cluster, found := ms.keeper.GetCluster(ctx, msg.ClusterId)
	if !found {
		return nil, types.ErrClusterNotFound.Wrapf("cluster %s not found", msg.ClusterId)
	}
	if cluster.ProviderAddress != providerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the cluster owner")
	}

	// Generate offering ID
	seq := ms.keeper.GetNextOfferingSequence(ctx)
	offeringID := fmt.Sprintf("OFF-%d", seq)

	// Create the offering
	offering := &types.HPCOffering{
		OfferingID:                offeringID,
		ClusterID:                 msg.ClusterId,
		ProviderAddress:           msg.ProviderAddress,
		Name:                      msg.Name,
		Description:               msg.Description,
		QueueOptions:              queueOptionsFromProto(msg.QueueOptions),
		Pricing:                   pricingFromProto(msg.Pricing),
		RequiredIdentityThreshold: msg.RequiredIdentityThreshold,
		MaxRuntimeSeconds:         msg.MaxRuntimeSeconds,
		PreconfiguredWorkloads:    preconfiguredWorkloadsFromProto(msg.PreconfiguredWorkloads),
		SupportsCustomWorkloads:   msg.SupportsCustomWorkloads,
		Active:                    true,
		CreatedAt:                 ctx.BlockTime(),
	}

	// Create the offering
	if err := ms.keeper.CreateOffering(ctx, offering); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC offering created",
		"offering_id", offeringID,
		"cluster_id", msg.ClusterId,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgCreateOfferingResponse{OfferingId: offeringID}, nil
}

// UpdateOffering handles updating an HPC offering
func (ms msgServer) UpdateOffering(goCtx context.Context, msg *types.MsgUpdateOffering) (*types.MsgUpdateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidOffering.Wrap("invalid provider address")
	}

	// Get existing offering
	offering, found := ms.keeper.GetOffering(ctx, msg.OfferingId)
	if !found {
		return nil, types.ErrOfferingNotFound.Wrapf("offering %s not found", msg.OfferingId)
	}

	// Verify ownership
	if offering.ProviderAddress != providerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the offering owner")
	}

	// Apply updates
	if msg.Name != "" {
		offering.Name = msg.Name
	}
	if msg.Description != "" {
		offering.Description = msg.Description
	}
	if len(msg.QueueOptions) > 0 {
		offering.QueueOptions = queueOptionsFromProto(msg.QueueOptions)
	}
	if msg.Pricing != nil {
		offering.Pricing = pricingFromProto(*msg.Pricing)
	}
	if msg.RequiredIdentityThreshold != 0 {
		offering.RequiredIdentityThreshold = msg.RequiredIdentityThreshold
	}
	if msg.MaxRuntimeSeconds != 0 {
		offering.MaxRuntimeSeconds = msg.MaxRuntimeSeconds
	}
	if len(msg.PreconfiguredWorkloads) > 0 {
		offering.PreconfiguredWorkloads = preconfiguredWorkloadsFromProto(msg.PreconfiguredWorkloads)
	}
	if msg.SupportsCustomWorkloads {
		offering.SupportsCustomWorkloads = true
	}
	offering.Active = msg.Active
	offering.UpdatedAt = ctx.BlockTime()

	// Update the offering
	if err := ms.keeper.UpdateOffering(ctx, &offering); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC offering updated",
		"offering_id", msg.OfferingId,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgUpdateOfferingResponse{}, nil
}

// SubmitJob handles submitting a new HPC job
func (ms msgServer) SubmitJob(goCtx context.Context, msg *types.MsgSubmitJob) (*types.MsgSubmitJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate submitter address
	submitterAddr, err := sdk.AccAddressFromBech32(msg.CustomerAddress)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid submitter address")
	}

	// Verify offering exists
	offering, found := ms.keeper.GetOffering(ctx, msg.OfferingId)
	if !found {
		return nil, types.ErrOfferingNotFound.Wrapf("offering %s not found", msg.OfferingId)
	}
	if !offering.Active {
		return nil, types.ErrInvalidOffering.Wrapf("offering %s is not active", msg.OfferingId)
	}

	// Generate job ID
	seq := ms.keeper.GetNextJobSequence(ctx)
	jobID := fmt.Sprintf("JOB-%d", seq)
	// Create the job
	job := &types.HPCJob{
		JobID:                   jobID,
		CustomerAddress:         submitterAddr.String(),
		OfferingID:              msg.OfferingId,
		ClusterID:               offering.ClusterID,
		ProviderAddress:         offering.ProviderAddress,
		QueueName:               msg.QueueName,
		WorkloadSpec:            workloadSpecFromProto(msg.WorkloadSpec),
		Resources:               jobResourcesFromProto(msg.Resources),
		DataReferences:          dataReferencesFromProto(msg.DataReferences),
		EncryptedInputsPointer:  msg.EncryptedInputsPointer,
		EncryptedOutputsPointer: msg.EncryptedOutputsPointer,
		MaxRuntimeSeconds:       msg.MaxRuntimeSeconds,
		AgreedPrice:             msg.MaxPrice,
		State:                   types.JobStatePending,
	}

	// Submit the job
	if err := ms.keeper.SubmitJob(ctx, job); err != nil {
		return nil, err
	}

	// Try to schedule the job
	if _, err := ms.keeper.ScheduleJob(ctx, job); err != nil {
		ms.keeper.Logger(ctx).Warn("job scheduling failed, will retry", "job_id", jobID, "error", err)
	}

	ms.keeper.Logger(ctx).Info("HPC job submitted",
		"job_id", jobID,
		"submitter", msg.CustomerAddress,
		"offering_id", msg.OfferingId,
	)

	return &types.MsgSubmitJobResponse{JobId: jobID}, nil
}

// CancelJob handles cancelling an HPC job
func (ms msgServer) CancelJob(goCtx context.Context, msg *types.MsgCancelJob) (*types.MsgCancelJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.RequesterAddress)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid sender address")
	}

	// Cancel the job
	if err := ms.keeper.CancelJob(ctx, msg.JobId, senderAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC job cancelled",
		"job_id", msg.JobId,
		"sender", msg.RequesterAddress,
	)

	return &types.MsgCancelJobResponse{}, nil
}

// ReportJobStatus handles provider reporting job status
func (ms msgServer) ReportJobStatus(goCtx context.Context, msg *types.MsgReportJobStatus) (*types.MsgReportJobStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate reporter address
	reporterAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid reporter address")
	}

	// Get the job to verify ownership
	job, found := ms.keeper.GetJob(ctx, msg.JobId)
	if !found {
		return nil, types.ErrJobNotFound.Wrapf("job %s not found", msg.JobId)
	}

	// Verify reporter owns this job (provider)
	if job.ProviderAddress != reporterAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the job provider")
	}

	// Map status string to JobState
	// Update job status
	jobState := jobStateFromProto(msg.State)
	if err := ms.keeper.UpdateJobStatus(ctx, msg.JobId, jobState, msg.StatusMessage, msg.ExitCode, usageMetricsFromProto(msg.UsageMetrics)); err != nil {
		return nil, err
	}

	// If job completed, distribute rewards
	if jobState == types.JobStateCompleted || jobState == types.JobStateFailed {
		if _, err := ms.keeper.DistributeJobRewards(ctx, msg.JobId); err != nil {
			ms.keeper.Logger(ctx).Error("failed to distribute job rewards", "job_id", msg.JobId, "error", err)
		}
	}

	ms.keeper.Logger(ctx).Info("HPC job status reported",
		"job_id", msg.JobId,
		"status", msg.State,
		"reporter", msg.ProviderAddress,
	)

	return &types.MsgReportJobStatusResponse{}, nil
}

// UpdateNodeMetadata handles updating node metadata
func (ms msgServer) UpdateNodeMetadata(goCtx context.Context, msg *types.MsgUpdateNodeMetadata) (*types.MsgUpdateNodeMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidNodeMetadata.Wrap("invalid owner address")
	}

	// Verify cluster ownership
	cluster, found := ms.keeper.GetCluster(ctx, msg.ClusterId)
	if !found {
		return nil, types.ErrClusterNotFound.Wrapf("cluster %s not found", msg.ClusterId)
	}
	if cluster.ProviderAddress != ownerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the cluster owner")
	}

	// Create/update node metadata
	node := &types.NodeMetadata{
		NodeID:               msg.NodeId,
		ClusterID:            msg.ClusterId,
		ProviderAddress:      ownerAddr.String(),
		Region:               msg.Region,
		Datacenter:           msg.Datacenter,
		LatencyMeasurements:  latencyMeasurementsFromProto(msg.LatencyMeasurements),
		NetworkBandwidthMbps: msg.NetworkBandwidthMbps,
		Resources:            nodeResourcesFromProto(msg.Resources),
		LastHeartbeat:        ctx.BlockTime(),
		UpdatedAt:            ctx.BlockTime(),
		Active:               msg.Active,
	}

	if err := ms.keeper.UpdateNodeMetadata(ctx, node); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC node metadata updated",
		"node_id", msg.NodeId,
		"cluster_id", msg.ClusterId,
	)

	return &types.MsgUpdateNodeMetadataResponse{}, nil
}

// FlagDispute handles flagging a dispute for an HPC job
func (ms msgServer) FlagDispute(goCtx context.Context, msg *types.MsgFlagDispute) (*types.MsgFlagDisputeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.DisputerAddress)
	if err != nil {
		return nil, types.ErrInvalidDispute.Wrap("invalid sender address")
	}

	// Get the job to verify relationship
	job, found := ms.keeper.GetJob(ctx, msg.JobId)
	if !found {
		return nil, types.ErrJobNotFound.Wrapf("job %s not found", msg.JobId)
	}

	// Verify sender is either customer or provider
	if job.CustomerAddress != senderAddr.String() && job.ProviderAddress != senderAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("sender must be customer or provider")
	}

	// Generate dispute ID
	seq := ms.keeper.GetNextDisputeSequence(ctx)
	disputeID := fmt.Sprintf("DSP-%d", seq)

	// Create the dispute
	dispute := &types.HPCDispute{
		DisputeID:       disputeID,
		JobID:           msg.JobId,
		DisputerAddress: senderAddr.String(),
		DisputeType:     msg.DisputeType,
		Reason:          msg.Reason,
		Evidence:        msg.Evidence,
		Status:          types.DisputeStatusPending,
		CreatedAt:       ctx.BlockTime(),
	}

	if err := ms.keeper.FlagDispute(ctx, dispute); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC dispute flagged",
		"dispute_id", disputeID,
		"job_id", msg.JobId,
		"sender", msg.DisputerAddress,
	)

	return &types.MsgFlagDisputeResponse{DisputeId: disputeID}, nil
}

// ResolveDispute handles resolving a dispute
func (ms msgServer) ResolveDispute(goCtx context.Context, msg *types.MsgResolveDispute) (*types.MsgResolveDisputeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate resolver address (must be module authority)
	if msg.ResolverAddress != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("expected %s, got %s", ms.keeper.GetAuthority(), msg.ResolverAddress)
	}

	// Resolve the dispute
	resolverAddr, err := sdk.AccAddressFromBech32(msg.ResolverAddress)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrap("invalid resolver address")
	}
	if err := ms.keeper.ResolveDispute(ctx, msg.DisputeId, disputeStatusFromProto(msg.Status), msg.Resolution, resolverAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC dispute resolved",
		"dispute_id", msg.DisputeId,
		"resolution", msg.Resolution,
		"resolver", msg.ResolverAddress,
	)

	return &types.MsgResolveDisputeResponse{}, nil
}

// UpdateParams updates module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	current := ms.keeper.GetParams(ctx)
	params := current
	params.PlatformFeeRate = msg.Params.PlatformFeeRate
	params.ProviderRewardRate = msg.Params.ProviderRewardRate
	params.NodeRewardRate = msg.Params.NodeRewardRate
	params.MinJobDurationSeconds = msg.Params.MinJobDurationSeconds
	params.MaxJobDurationSeconds = msg.Params.MaxJobDurationSeconds
	params.DefaultIdentityThreshold = msg.Params.DefaultIdentityThreshold
	params.ClusterHeartbeatTimeout = msg.Params.ClusterHeartbeatTimeout
	params.NodeHeartbeatTimeout = msg.Params.NodeHeartbeatTimeout
	params.LatencyWeightFactor = msg.Params.LatencyWeightFactor
	params.CapacityWeightFactor = msg.Params.CapacityWeightFactor
	params.MaxLatencyMs = msg.Params.MaxLatencyMs
	params.DisputeResolutionPeriod = msg.Params.DisputeResolutionPeriod
	params.RewardFormulaVersion = msg.Params.RewardFormulaVersion
	params.EnableProximityClustering = msg.Params.EnableProximityClustering

	if err := ms.keeper.SetParams(ctx, params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func clusterStateFromProto(state hpcv1.ClusterState) types.ClusterState {
	switch state {
	case hpcv1.ClusterStatePending:
		return types.ClusterStatePending
	case hpcv1.ClusterStateActive:
		return types.ClusterStateActive
	case hpcv1.ClusterStateDraining:
		return types.ClusterStateDraining
	case hpcv1.ClusterStateOffline:
		return types.ClusterStateOffline
	case hpcv1.ClusterStateDeregistered:
		return types.ClusterStateDeregistered
	default:
		return types.ClusterStatePending
	}
}

func jobStateFromProto(state hpcv1.JobState) types.JobState {
	switch state {
	case hpcv1.JobStatePending:
		return types.JobStatePending
	case hpcv1.JobStateQueued:
		return types.JobStateQueued
	case hpcv1.JobStateRunning:
		return types.JobStateRunning
	case hpcv1.JobStateCompleted:
		return types.JobStateCompleted
	case hpcv1.JobStateFailed:
		return types.JobStateFailed
	case hpcv1.JobStateCancelled:
		return types.JobStateCancelled
	case hpcv1.JobStateTimeout:
		return types.JobStateTimeout
	default:
		return types.JobStatePending
	}
}

func disputeStatusFromProto(status hpcv1.DisputeStatus) types.DisputeStatus {
	switch status {
	case hpcv1.DisputeStatusUnderReview:
		return types.DisputeStatusUnderReview
	case hpcv1.DisputeStatusResolved:
		return types.DisputeStatusResolved
	case hpcv1.DisputeStatusRejected:
		return types.DisputeStatusRejected
	default:
		return types.DisputeStatusPending
	}
}

func clusterMetadataFromProto(meta hpcv1.ClusterMetadata) types.ClusterMetadata {
	return types.ClusterMetadata{
		TotalCPUCores:    meta.TotalCpuCores,
		TotalMemoryGB:    meta.TotalMemoryGb,
		TotalGPUs:        meta.TotalGpus,
		GPUTypes:         meta.GpuTypes,
		InterconnectType: meta.InterconnectType,
		StorageType:      meta.StorageType,
		TotalStorageGB:   meta.TotalStorageGb,
	}
}

func partitionsFromProto(partitions []hpcv1.Partition) []types.Partition {
	if len(partitions) == 0 {
		return nil
	}
	out := make([]types.Partition, 0, len(partitions))
	for _, partition := range partitions {
		out = append(out, types.Partition{
			Name:           partition.Name,
			Nodes:          partition.Nodes,
			MaxRuntime:     partition.MaxRuntime,
			DefaultRuntime: partition.DefaultRuntime,
			MaxNodes:       partition.MaxNodes,
			Features:       partition.Features,
			Priority:       partition.Priority,
			State:          partition.State,
		})
	}
	return out
}

func queueOptionsFromProto(options []hpcv1.QueueOption) []types.QueueOption {
	if len(options) == 0 {
		return nil
	}
	out := make([]types.QueueOption, 0, len(options))
	for _, option := range options {
		out = append(out, types.QueueOption{
			PartitionName:   option.PartitionName,
			DisplayName:     option.DisplayName,
			MaxNodes:        option.MaxNodes,
			MaxRuntime:      option.MaxRuntime,
			Features:        option.Features,
			PriceMultiplier: option.PriceMultiplier,
		})
	}
	return out
}

func pricingFromProto(pricing hpcv1.HPCPricing) types.HPCPricing {
	return types.HPCPricing{
		BaseNodeHourPrice: pricing.BaseNodeHourPrice,
		CPUCoreHourPrice:  pricing.CpuCoreHourPrice,
		GPUHourPrice:      pricing.GpuHourPrice,
		MemoryGBHourPrice: pricing.MemoryGbHourPrice,
		StorageGBPrice:    pricing.StorageGbPrice,
		NetworkGBPrice:    pricing.NetworkGbPrice,
		Currency:          pricing.Currency,
		MinimumCharge:     pricing.MinimumCharge,
	}
}

func preconfiguredWorkloadsFromProto(workloads []hpcv1.PreconfiguredWorkload) []types.PreconfiguredWorkload {
	if len(workloads) == 0 {
		return nil
	}
	out := make([]types.PreconfiguredWorkload, 0, len(workloads))
	for _, workload := range workloads {
		out = append(out, types.PreconfiguredWorkload{
			WorkloadID:        workload.WorkloadId,
			Name:              workload.Name,
			Description:       workload.Description,
			ContainerImage:    workload.ContainerImage,
			DefaultCommand:    workload.DefaultCommand,
			RequiredResources: jobResourcesFromProto(workload.RequiredResources),
			Category:          workload.Category,
			Version:           workload.Version,
		})
	}
	return out
}

func workloadSpecFromProto(spec hpcv1.JobWorkloadSpec) types.JobWorkloadSpec {
	return types.JobWorkloadSpec{
		ContainerImage:          spec.ContainerImage,
		Command:                 spec.Command,
		Arguments:               spec.Arguments,
		Environment:             spec.Environment,
		WorkingDirectory:        spec.WorkingDirectory,
		PreconfiguredWorkloadID: spec.PreconfiguredWorkloadId,
		IsPreconfigured:         spec.IsPreconfigured,
	}
}

func jobResourcesFromProto(resources hpcv1.JobResources) types.JobResources {
	return types.JobResources{
		Nodes:           resources.Nodes,
		CPUCoresPerNode: resources.CpuCoresPerNode,
		MemoryGBPerNode: resources.MemoryGbPerNode,
		GPUsPerNode:     resources.GpusPerNode,
		StorageGB:       resources.StorageGb,
		GPUType:         resources.GpuType,
	}
}

func dataReferencesFromProto(references []hpcv1.DataReference) []types.DataReference {
	if len(references) == 0 {
		return nil
	}
	out := make([]types.DataReference, 0, len(references))
	for _, reference := range references {
		out = append(out, types.DataReference{
			ReferenceID: reference.ReferenceId,
			Type:        reference.Type,
			URI:         reference.Uri,
			Encrypted:   reference.Encrypted,
			Checksum:    reference.Checksum,
			SizeBytes:   reference.SizeBytes,
		})
	}
	return out
}

func usageMetricsFromProto(metrics *hpcv1.HPCUsageMetrics) *types.HPCUsageMetrics {
	if metrics == nil {
		return nil
	}
	return &types.HPCUsageMetrics{
		WallClockSeconds: metrics.WallClockSeconds,
		CPUCoreSeconds:   metrics.CpuCoreSeconds,
		MemoryGBSeconds:  metrics.MemoryGbSeconds,
		GPUSeconds:       metrics.GpuSeconds,
		StorageGBHours:   metrics.StorageGbHours,
		NetworkBytesIn:   metrics.NetworkBytesIn,
		NetworkBytesOut:  metrics.NetworkBytesOut,
		NodeHours:        metrics.NodeHours,
		NodesUsed:        metrics.NodesUsed,
	}
}

func nodeResourcesFromProto(resources *hpcv1.NodeResources) types.NodeResources {
	if resources == nil {
		return types.NodeResources{}
	}
	return types.NodeResources{
		CPUCores:  resources.CpuCores,
		MemoryGB:  resources.MemoryGb,
		GPUs:      resources.Gpus,
		GPUType:   resources.GpuType,
		StorageGB: resources.StorageGb,
	}
}

func latencyMeasurementsFromProto(measurements []hpcv1.LatencyMeasurement) []types.LatencyMeasurement {
	if len(measurements) == 0 {
		return nil
	}
	out := make([]types.LatencyMeasurement, 0, len(measurements))
	for _, measurement := range measurements {
		out = append(out, types.LatencyMeasurement{
			TargetNodeID: measurement.TargetNodeId,
			LatencyMs:    measurement.LatencyMs,
			MeasuredAt:   measurement.MeasuredAt,
		})
	}
	return out
}
