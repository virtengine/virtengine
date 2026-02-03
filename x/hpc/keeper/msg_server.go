// Package keeper implements the HPC module keeper.
//
// VE-2019: MsgServer implementation for HPC module
package keeper

import (
	"context"
	"fmt"
	"math"

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

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid owner address")
	}

	// Generate cluster ID
	seq := ms.keeper.GetNextClusterSequence(ctx)
	clusterID := fmt.Sprintf("HPC-%d", seq)
	totalNodes, err := safeInt32FromUint64(msg.TotalNodes)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap(err.Error())
	}

	// Create the cluster
	cluster := &types.HPCCluster{
		ClusterID:       clusterID,
		ProviderAddress: ownerAddr.String(),
		Name:            msg.Name,
		Region:          msg.Region,
		TotalNodes:      totalNodes,
		AvailableNodes:  totalNodes,
		State:           types.ClusterStateActive,
	}

	// Register the cluster
	if err := ms.keeper.RegisterCluster(ctx, cluster); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster registered",
		"cluster_id", clusterID,
		"owner", msg.Owner,
		"name", msg.Name,
	)

	return &types.MsgRegisterClusterResponse{ClusterId: clusterID}, nil
}

// UpdateCluster handles updating an HPC cluster
func (ms msgServer) UpdateCluster(goCtx context.Context, msg *types.MsgUpdateCluster) (*types.MsgUpdateClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
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
	if msg.TotalNodes > 0 {
		totalNodes, err := safeInt32FromUint64(msg.TotalNodes)
		if err != nil {
			return nil, types.ErrInvalidCluster.Wrap(err.Error())
		}
		cluster.TotalNodes = totalNodes
	}

	// Update the cluster
	if err := ms.keeper.UpdateCluster(ctx, &cluster); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster updated",
		"cluster_id", msg.ClusterId,
		"owner", msg.Owner,
	)

	return &types.MsgUpdateClusterResponse{}, nil
}

// DeregisterCluster handles deregistering an HPC cluster
func (ms msgServer) DeregisterCluster(goCtx context.Context, msg *types.MsgDeregisterCluster) (*types.MsgDeregisterClusterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, types.ErrInvalidCluster.Wrap("invalid owner address")
	}

	// Deregister the cluster
	if err := ms.keeper.DeregisterCluster(ctx, msg.ClusterId, ownerAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC cluster deregistered",
		"cluster_id", msg.ClusterId,
		"owner", msg.Owner,
	)

	return &types.MsgDeregisterClusterResponse{}, nil
}

// CreateOffering handles creating a new HPC offering
func (ms msgServer) CreateOffering(goCtx context.Context, msg *types.MsgCreateOffering) (*types.MsgCreateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
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
		OfferingID:      offeringID,
		ClusterID:       msg.ClusterId,
		ProviderAddress: msg.Provider,
		Name:            msg.Name,
		Active:          true,
		CreatedAt:       ctx.BlockTime(),
	}

	// Create the offering
	if err := ms.keeper.CreateOffering(ctx, offering); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC offering created",
		"offering_id", offeringID,
		"cluster_id", msg.ClusterId,
		"provider", msg.Provider,
	)

	return &types.MsgCreateOfferingResponse{OfferingId: offeringID}, nil
}

func safeInt32FromUint64(value uint64) (int32, error) {
	maxInt32 := uint64(^uint32(0) >> 1)
	if value > maxInt32 {
		return 0, fmt.Errorf("value exceeds int32: %d", value)
	}
	return int32(value), nil
}

// UpdateOffering handles updating an HPC offering
func (ms msgServer) UpdateOffering(goCtx context.Context, msg *types.MsgUpdateOffering) (*types.MsgUpdateOfferingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate provider address
	providerAddr, err := sdk.AccAddressFromBech32(msg.Provider)
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
	offering.Active = msg.Active
	offering.UpdatedAt = ctx.BlockTime()

	// Update the offering
	if err := ms.keeper.UpdateOffering(ctx, &offering); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC offering updated",
		"offering_id", msg.OfferingId,
		"provider", msg.Provider,
	)

	return &types.MsgUpdateOfferingResponse{}, nil
}

// SubmitJob handles submitting a new HPC job
func (ms msgServer) SubmitJob(goCtx context.Context, msg *types.MsgSubmitJob) (*types.MsgSubmitJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate submitter address
	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
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
	maxInt32 := uint64(math.MaxInt32)
	if msg.RequestedNodes > maxInt32 {
		return nil, types.ErrInvalidJob.Wrap("requested_nodes exceeds int32")
	}
	maxInt64 := uint64(math.MaxInt64)
	if msg.MaxDuration > maxInt64 {
		return nil, types.ErrInvalidJob.Wrap("max_duration exceeds int64")
	}

	// Create the job
	job := &types.HPCJob{
		JobID:           jobID,
		CustomerAddress: submitterAddr.String(),
		OfferingID:      msg.OfferingId,
		ClusterID:       offering.ClusterID,
		ProviderAddress: offering.ProviderAddress,
		Resources: types.JobResources{
			//nolint:gosec // bounds checked above
			Nodes: int32(msg.RequestedNodes),
		},
		MaxRuntimeSeconds: safeInt64FromUint64(msg.MaxDuration),
		State:             types.JobStatePending,
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
		"submitter", msg.Submitter,
		"offering_id", msg.OfferingId,
	)

	return &types.MsgSubmitJobResponse{JobId: jobID}, nil
}

// CancelJob handles cancelling an HPC job
func (ms msgServer) CancelJob(goCtx context.Context, msg *types.MsgCancelJob) (*types.MsgCancelJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender address
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidJob.Wrap("invalid sender address")
	}

	// Cancel the job
	if err := ms.keeper.CancelJob(ctx, msg.JobId, senderAddr); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC job cancelled",
		"job_id", msg.JobId,
		"sender", msg.Sender,
	)

	return &types.MsgCancelJobResponse{}, nil
}

func safeInt64FromUint64(value uint64) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(value)
}

// ReportJobStatus handles provider reporting job status
func (ms msgServer) ReportJobStatus(goCtx context.Context, msg *types.MsgReportJobStatus) (*types.MsgReportJobStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate reporter address
	reporterAddr, err := sdk.AccAddressFromBech32(msg.Reporter)
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
	var jobState types.JobState
	switch msg.Status {
	case "running":
		jobState = types.JobStateRunning
	case "completed":
		jobState = types.JobStateCompleted
	case "failed":
		jobState = types.JobStateFailed
	case "cancelled":
		jobState = types.JobStateCancelled
	default:
		jobState = types.JobState(msg.Status)
	}

	// Update job status
	if err := ms.keeper.UpdateJobStatus(ctx, msg.JobId, jobState, msg.ErrorMessage, 0, nil); err != nil {
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
		"status", msg.Status,
		"reporter", msg.Reporter,
	)

	return &types.MsgReportJobStatusResponse{}, nil
}

// UpdateNodeMetadata handles updating node metadata
func (ms msgServer) UpdateNodeMetadata(goCtx context.Context, msg *types.MsgUpdateNodeMetadata) (*types.MsgUpdateNodeMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
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
		NodeID:          msg.NodeId,
		ClusterID:       msg.ClusterId,
		ProviderAddress: ownerAddr.String(),
		LastHeartbeat:   ctx.BlockTime(),
		UpdatedAt:       ctx.BlockTime(),
		Active:          true,
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
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
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
		DisputeType:     "general",
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
		"sender", msg.Sender,
	)

	return &types.MsgFlagDisputeResponse{DisputeId: disputeID}, nil
}

// ResolveDispute handles resolving a dispute
func (ms msgServer) ResolveDispute(goCtx context.Context, msg *types.MsgResolveDispute) (*types.MsgResolveDisputeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address (must be module authority)
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrapf("expected %s, got %s", ms.keeper.GetAuthority(), msg.Authority)
	}

	// Determine dispute status from resolution
	var disputeStatus types.DisputeStatus
	switch msg.Resolution {
	case "resolved":
		disputeStatus = types.DisputeStatusResolved
	case "rejected":
		disputeStatus = types.DisputeStatusRejected
	default:
		disputeStatus = types.DisputeStatusResolved
	}

	// Resolve the dispute
	if err := ms.keeper.ResolveDispute(ctx, msg.DisputeId, disputeStatus, msg.Resolution, sdk.MustAccAddressFromBech32(msg.Authority)); err != nil {
		return nil, err
	}

	ms.keeper.Logger(ctx).Info("HPC dispute resolved",
		"dispute_id", msg.DisputeId,
		"resolution", msg.Resolution,
		"authority", msg.Authority,
	)

	return &types.MsgResolveDisputeResponse{}, nil
}
