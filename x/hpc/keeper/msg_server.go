// Package keeper implements the HPC module keeper.
//
// VE-2019: MsgServer implementation for HPC module
package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid provider address")
	}

	// Generate cluster ID
	seq := ms.keeper.GetNextClusterSequence(ctx)
	clusterID := fmt.Sprintf("HPC-%d", seq)

	// Create the cluster
	cluster := &types.HPCCluster{
		ClusterID:       clusterID,
		ProviderAddress: providerAddr.String(),
		Name:            msg.Name,
		Description:     msg.Description,
		Region:          msg.Region,
		Partitions:      msg.Partitions,
		TotalNodes:      msg.TotalNodes,
		AvailableNodes:  msg.TotalNodes,
		ClusterMetadata: msg.ClusterMetadata,
		SLURMVersion:    msg.SLURMVersion,
		State:           types.ClusterStateActive,
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

	return &types.MsgRegisterClusterResponse{ClusterID: clusterID}, nil
}

// UpdateCluster handles updating an HPC cluster
func (ms msgServer) UpdateCluster(goCtx context.Context, msg *types.MsgUpdateCluster) (*types.MsgUpdateClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid provider address")
	}

	// Get existing cluster
	cluster, found := ms.keeper.GetCluster(ctx, msg.ClusterID)
	if !found {
		return nil, types.ErrClusterNotFound.Wrapf("cluster %s not found", msg.ClusterID)
	}

	// Verify ownership
	if cluster.ProviderAddress != providerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the cluster owner")
	}

	// Apply updates
	if msg.Name != "" {
		cluster.Name = msg.Name
	}
	if msg.Description != "" {
		cluster.Description = msg.Description
	}
	if msg.State != "" {
		cluster.State = msg.State
	}
	if len(msg.Partitions) > 0 {
		cluster.Partitions = msg.Partitions
	}
	if msg.TotalNodes > 0 {
		cluster.TotalNodes = msg.TotalNodes
	}
	if msg.AvailableNodes >= 0 {
		cluster.AvailableNodes = msg.AvailableNodes
	}
	if msg.ClusterMetadata.TotalCPUCores > 0 {
		cluster.ClusterMetadata = msg.ClusterMetadata
	}

	// Update the cluster
	if err := ms.keeper.UpdateCluster(ctx, &cluster); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster updated",
		"cluster_id", msg.ClusterID,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgUpdateClusterResponse{}, nil
}

// DeregisterCluster handles deregistering an HPC cluster
func (ms msgServer) DeregisterCluster(goCtx context.Context, msg *types.MsgDeregisterCluster) (*types.MsgDeregisterClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid provider address")
	}

	// Deregister the cluster
	if err := ms.keeper.DeregisterCluster(ctx, msg.ClusterID, providerAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster deregistered",
		"cluster_id", msg.ClusterID,
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
	cluster, found := ms.keeper.GetCluster(ctx, msg.ClusterID)
	if !found {
		return nil, types.ErrClusterNotFound.Wrapf("cluster %s not found", msg.ClusterID)
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
		ClusterID:                 msg.ClusterID,
		ProviderAddress:           msg.ProviderAddress,
		Name:                      msg.Name,
		Description:               msg.Description,
		QueueOptions:              msg.QueueOptions,
		Pricing:                   msg.Pricing,
		RequiredIdentityThreshold: msg.RequiredIdentityThreshold,
		MaxRuntimeSeconds:         msg.MaxRuntimeSeconds,
		PreconfiguredWorkloads:    msg.PreconfiguredWorkloads,
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
		"cluster_id", msg.ClusterID,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgCreateOfferingResponse{OfferingID: offeringID}, nil
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
	offering, found := ms.keeper.GetOffering(ctx, msg.OfferingID)
	if !found {
		return nil, types.ErrOfferingNotFound.Wrapf("offering %s not found", msg.OfferingID)
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
		offering.QueueOptions = msg.QueueOptions
	}
	if msg.Pricing.BaseNodeHourPrice != "" {
		offering.Pricing = msg.Pricing
	}
	if msg.RequiredIdentityThreshold > 0 {
		offering.RequiredIdentityThreshold = msg.RequiredIdentityThreshold
	}
	if msg.MaxRuntimeSeconds > 0 {
		offering.MaxRuntimeSeconds = msg.MaxRuntimeSeconds
	}
	if len(msg.PreconfiguredWorkloads) > 0 {
		offering.PreconfiguredWorkloads = msg.PreconfiguredWorkloads
	}
	if msg.SupportsCustomWorkloads != nil {
		offering.SupportsCustomWorkloads = *msg.SupportsCustomWorkloads
	}
	if msg.Active != nil {
		offering.Active = *msg.Active
	}
	offering.UpdatedAt = ctx.BlockTime()

	// Update the offering
	if err := ms.keeper.UpdateOffering(ctx, &offering); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC offering updated",
		"offering_id", msg.OfferingID,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgUpdateOfferingResponse{}, nil
}

// SubmitJob handles submitting a new HPC job
func (ms msgServer) SubmitJob(goCtx context.Context, msg *types.MsgSubmitJob) (*types.MsgSubmitJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate customer address
	customerAddr, err := sdk.AccAddressFromBech32(msg.CustomerAddress)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid customer address")
	}

	// Verify offering exists
	offering, found := ms.keeper.GetOffering(ctx, msg.OfferingID)
	if !found {
		return nil, types.ErrOfferingNotFound.Wrapf("offering %s not found", msg.OfferingID)
	}
	if !offering.Active {
		return nil, types.ErrInvalidOffering.Wrapf("offering %s is not active", msg.OfferingID)
	}

	// Generate job ID
	seq := ms.keeper.GetNextJobSequence(ctx)
	jobID := fmt.Sprintf("JOB-%d", seq)

	// Create the job
	job := &types.HPCJob{
		JobID:                   jobID,
		CustomerAddress:         customerAddr.String(),
		OfferingID:              msg.OfferingID,
		ClusterID:               offering.ClusterID,
		ProviderAddress:         offering.ProviderAddress,
		QueueName:               msg.QueueName,
		WorkloadSpec:            msg.WorkloadSpec,
		Resources:               msg.Resources,
		DataReferences:          msg.DataReferences,
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
		"customer", msg.CustomerAddress,
		"offering_id", msg.OfferingID,
	)

	return &types.MsgSubmitJobResponse{JobID: jobID}, nil
}

// CancelJob handles cancelling an HPC job
func (ms msgServer) CancelJob(goCtx context.Context, msg *types.MsgCancelJob) (*types.MsgCancelJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate requester address
	requesterAddr, err := sdk.AccAddressFromBech32(msg.RequesterAddress)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid requester address")
	}

	// Cancel the job
	if err := ms.keeper.CancelJob(ctx, msg.JobID, requesterAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC job cancelled",
		"job_id", msg.JobID,
		"requester", msg.RequesterAddress,
	)

	return &types.MsgCancelJobResponse{}, nil
}

// ReportJobStatus handles provider reporting job status
func (ms msgServer) ReportJobStatus(goCtx context.Context, msg *types.MsgReportJobStatus) (*types.MsgReportJobStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid provider address")
	}

	// Get the job to verify ownership
	job, found := ms.keeper.GetJob(ctx, msg.JobID)
	if !found {
		return nil, types.ErrJobNotFound.Wrapf("job %s not found", msg.JobID)
	}

	// Verify provider owns this job
	if job.ProviderAddress != providerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the job provider")
	}

	// Update job status
	if err := ms.keeper.UpdateJobStatus(ctx, msg.JobID, msg.State, msg.StatusMessage, msg.ExitCode, &msg.UsageMetrics); err != nil {
		return nil, err
	}

	// If job completed, distribute rewards
	if msg.State == types.JobStateCompleted || msg.State == types.JobStateFailed {
		if _, err := ms.keeper.DistributeJobRewards(ctx, msg.JobID); err != nil {
			ms.keeper.Logger(ctx).Error("failed to distribute job rewards", "job_id", msg.JobID, "error", err)
		}
	}

	ms.keeper.Logger(ctx).Info("HPC job status reported",
		"job_id", msg.JobID,
		"state", msg.State,
		"provider", msg.ProviderAddress,
	)

	return &types.MsgReportJobStatusResponse{}, nil
}

// UpdateNodeMetadata handles updating node metadata
func (ms msgServer) UpdateNodeMetadata(goCtx context.Context, msg *types.MsgUpdateNodeMetadata) (*types.MsgUpdateNodeMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.ProviderAddress)
	if err != nil {
		return nil, types.ErrInvalidNodeMetadata.Wrap("invalid provider address")
	}

	// Verify cluster ownership
	cluster, found := ms.keeper.GetCluster(ctx, msg.ClusterID)
	if !found {
		return nil, types.ErrClusterNotFound.Wrapf("cluster %s not found", msg.ClusterID)
	}
	if cluster.ProviderAddress != providerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("not the cluster owner")
	}

	// Create/update node metadata
	node := &types.NodeMetadata{
		NodeID:               msg.NodeID,
		ClusterID:            msg.ClusterID,
		ProviderAddress:      providerAddr.String(),
		Region:               msg.Region,
		Datacenter:           msg.Datacenter,
		LatencyMeasurements:  msg.LatencyMeasurements,
		NetworkBandwidthMbps: msg.NetworkBandwidthMbps,
		Resources:            msg.Resources,
		LastHeartbeat:        ctx.BlockTime(),
		UpdatedAt:            ctx.BlockTime(),
	}
	if msg.Active != nil {
		node.Active = *msg.Active
	}

	if err := ms.keeper.UpdateNodeMetadata(ctx, node); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC node metadata updated",
		"node_id", msg.NodeID,
		"cluster_id", msg.ClusterID,
	)

	return &types.MsgUpdateNodeMetadataResponse{}, nil
}

// FlagDispute handles flagging a dispute for an HPC job
func (ms msgServer) FlagDispute(goCtx context.Context, msg *types.MsgFlagDispute) (*types.MsgFlagDisputeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate disputer address
	disputerAddr, err := sdk.AccAddressFromBech32(msg.DisputerAddress)
	if err != nil {
		return nil, types.ErrInvalidDispute.Wrap("invalid disputer address")
	}

	// Get the job to verify relationship
	job, found := ms.keeper.GetJob(ctx, msg.JobID)
	if !found {
		return nil, types.ErrJobNotFound.Wrapf("job %s not found", msg.JobID)
	}

	// Verify disputer is either customer or provider
	if job.CustomerAddress != disputerAddr.String() && job.ProviderAddress != disputerAddr.String() {
		return nil, types.ErrUnauthorized.Wrap("disputer must be customer or provider")
	}

	// Generate dispute ID
	seq := ms.keeper.GetNextDisputeSequence(ctx)
	disputeID := fmt.Sprintf("DSP-%d", seq)

	// Create the dispute
	dispute := &types.HPCDispute{
		DisputeID:       disputeID,
		JobID:           msg.JobID,
		RewardID:        msg.RewardID,
		DisputerAddress: disputerAddr.String(),
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
		"job_id", msg.JobID,
		"disputer", msg.DisputerAddress,
	)

	return &types.MsgFlagDisputeResponse{DisputeID: disputeID}, nil
}

// ResolveDispute handles resolving a dispute
func (ms msgServer) ResolveDispute(goCtx context.Context, msg *types.MsgResolveDispute) (*types.MsgResolveDisputeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate resolver address (must be module authority)
	if msg.ResolverAddress != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("expected %s, got %s", ms.keeper.GetAuthority(), msg.ResolverAddress)
	}

	// Resolve the dispute
	if err := ms.keeper.ResolveDispute(ctx, msg.DisputeID, msg.Status, msg.Resolution, sdk.MustAccAddressFromBech32(msg.ResolverAddress)); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC dispute resolved",
		"dispute_id", msg.DisputeID,
		"status", msg.Status,
		"resolver", msg.ResolverAddress,
	)

	return &types.MsgResolveDisputeResponse{}, nil
}
