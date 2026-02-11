package keeper

import (
	"context"
	"encoding/json"

	"cosmossdk.io/store/prefix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	"github.com/virtengine/virtengine/x/hpc/types"
)

// Querier implements the gRPC QueryServer for the HPC module.
type Querier struct {
	Keeper
	hpcv1.UnimplementedQueryServer
}

// NewQuerier returns a new Querier.
func NewQuerier(k Keeper) *Querier {
	return &Querier{Keeper: k}
}

var _ hpcv1.QueryServer = (*Querier)(nil)

// NodeMetadata returns metadata for a specific node.
func (q *Querier) NodeMetadata(ctx context.Context, req *hpcv1.QueryNodeMetadataRequest) (*hpcv1.QueryNodeMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	node, found := q.GetNodeMetadata(sdkCtx, req.NodeId)
	if !found {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	return &hpcv1.QueryNodeMetadataResponse{
		Node: nodeMetadataToProto(node),
	}, nil
}

// NodesByCluster returns nodes belonging to a cluster.
func (q *Querier) NodesByCluster(ctx context.Context, req *hpcv1.QueryNodesByClusterRequest) (*hpcv1.QueryNodesByClusterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := prefix.NewStore(sdkCtx.KVStore(q.skey), types.NodeMetadataPrefix)

	nodes := make([]hpcv1.NodeMetadata, 0)
	pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var node types.NodeMetadata
		if err := json.Unmarshal(value, &node); err != nil {
			return false, err
		}

		match := node.ClusterID == req.ClusterId
		if accumulate && match {
			nodes = append(nodes, nodeMetadataToProto(node))
		}

		return match, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &hpcv1.QueryNodesByClusterResponse{
		Nodes:      nodes,
		Pagination: pageRes,
	}, nil
}

// SchedulingDecision returns a scheduling decision by ID.
func (q *Querier) SchedulingDecision(ctx context.Context, req *hpcv1.QuerySchedulingDecisionRequest) (*hpcv1.QuerySchedulingDecisionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.DecisionId == "" {
		return nil, status.Error(codes.InvalidArgument, "decision_id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	decision, found := q.GetSchedulingDecision(sdkCtx, req.DecisionId)
	if !found {
		return nil, status.Error(codes.NotFound, "scheduling decision not found")
	}

	return &hpcv1.QuerySchedulingDecisionResponse{
		Decision: schedulingDecisionToProto(decision),
	}, nil
}

// SchedulingDecisionByJob returns the scheduling decision for a job.
func (q *Querier) SchedulingDecisionByJob(ctx context.Context, req *hpcv1.QuerySchedulingDecisionByJobRequest) (*hpcv1.QuerySchedulingDecisionByJobResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	job, found := q.GetJob(sdkCtx, req.JobId)
	if !found || job.SchedulingDecisionID == "" {
		return nil, status.Error(codes.NotFound, "scheduling decision not found")
	}

	decision, found := q.GetSchedulingDecision(sdkCtx, job.SchedulingDecisionID)
	if !found {
		return nil, status.Error(codes.NotFound, "scheduling decision not found")
	}

	return &hpcv1.QuerySchedulingDecisionByJobResponse{
		Decision: schedulingDecisionToProto(decision),
	}, nil
}

// Params returns module parameters.
func (q *Querier) Params(ctx context.Context, _ *hpcv1.QueryParamsRequest) (*hpcv1.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.GetParams(sdkCtx)
	return &hpcv1.QueryParamsResponse{
		Params: paramsToProto(params),
	}, nil
}

func nodeMetadataToProto(node types.NodeMetadata) hpcv1.NodeMetadata {
	locality := node.Locality
	if locality == nil && (node.Region != "" || node.Datacenter != "") {
		locality = &types.NodeLocality{
			Region:     node.Region,
			Datacenter: node.Datacenter,
		}
	}

	out := hpcv1.NodeMetadata{
		NodeId:               node.NodeID,
		ClusterId:            node.ClusterID,
		ProviderAddress:      node.ProviderAddress,
		Region:               node.Region,
		Datacenter:           node.Datacenter,
		LatencyMeasurements:  latencyMeasurementsToProto(node.LatencyMeasurements),
		AvgLatencyMs:         node.AvgLatencyMs,
		NetworkBandwidthMbps: node.NetworkBandwidthMbps,
		Resources:            nodeResourcesToProto(node.Resources),
		Active:               node.Active,
		LastHeartbeat:        node.LastHeartbeat,
		JoinedAt:             node.JoinedAt,
		UpdatedAt:            node.UpdatedAt,
		BlockHeight:          node.BlockHeight,
		State:                nodeStateToProto(node.State),
		HealthStatus:         healthStatusToProto(node.HealthStatus),
		AgentPubkey:          node.AgentPubkey,
		HardwareFingerprint:  node.HardwareFingerprint,
		AgentVersion:         node.AgentVersion,
		LastSequenceNumber:   node.LastSequenceNumber,
	}

	if node.Capacity != nil {
		out.Capacity = nodeCapacityToProto(*node.Capacity)
	}
	if node.Health != nil {
		out.Health = nodeHealthToProto(*node.Health)
	}
	if node.Hardware != nil {
		out.Hardware = nodeHardwareToProto(*node.Hardware)
	}
	if node.Topology != nil {
		out.Topology = nodeTopologyToProto(*node.Topology)
	}
	if locality != nil {
		out.Locality = nodeLocalityToProto(*locality)
	}

	return out
}

func latencyMeasurementsToProto(measurements []types.LatencyMeasurement) []hpcv1.LatencyMeasurement {
	if len(measurements) == 0 {
		return nil
	}
	out := make([]hpcv1.LatencyMeasurement, 0, len(measurements))
	for _, measurement := range measurements {
		out = append(out, hpcv1.LatencyMeasurement{
			TargetNodeId: measurement.TargetNodeID,
			LatencyMs:    measurement.LatencyMs,
			MeasuredAt:   measurement.MeasuredAt,
		})
	}
	return out
}

func schedulingDecisionToProto(decision types.SchedulingDecision) hpcv1.SchedulingDecision {
	candidates := make([]hpcv1.ClusterCandidate, 0, len(decision.CandidateClusters))
	for _, candidate := range decision.CandidateClusters {
		candidates = append(candidates, clusterCandidateToProto(candidate))
	}

	return hpcv1.SchedulingDecision{
		DecisionId:        decision.DecisionID,
		JobId:             decision.JobID,
		SelectedClusterId: decision.SelectedClusterID,
		CandidateClusters: candidates,
		DecisionReason:    decision.DecisionReason,
		IsFallback:        decision.IsFallback,
		FallbackReason:    decision.FallbackReason,
		LatencyScore:      decision.LatencyScore,
		CapacityScore:     decision.CapacityScore,
		CombinedScore:     decision.CombinedScore,
		CreatedAt:         decision.CreatedAt,
		BlockHeight:       decision.BlockHeight,
	}
}

func clusterCandidateToProto(candidate types.ClusterCandidate) hpcv1.ClusterCandidate {
	return hpcv1.ClusterCandidate{
		ClusterId:           candidate.ClusterID,
		Region:              candidate.Region,
		AvgLatencyMs:        candidate.AvgLatencyMs,
		AvailableNodes:      candidate.AvailableNodes,
		LatencyScore:        candidate.LatencyScore,
		CapacityScore:       candidate.CapacityScore,
		CombinedScore:       candidate.CombinedScore,
		Eligible:            candidate.Eligible,
		IneligibilityReason: candidate.IneligibilityReason,
	}
}

func nodeResourcesToProto(resources types.NodeResources) hpcv1.NodeResources {
	return hpcv1.NodeResources{
		CpuCores:  resources.CPUCores,
		MemoryGb:  resources.MemoryGB,
		Gpus:      resources.GPUs,
		GpuType:   resources.GPUType,
		StorageGb: resources.StorageGB,
	}
}

func nodeCapacityToProto(capacity types.NodeCapacity) *hpcv1.NodeCapacity {
	return &hpcv1.NodeCapacity{
		CpuCoresTotal:      capacity.CPUCoresTotal,
		CpuCoresAvailable:  capacity.CPUCoresAvailable,
		CpuCoresAllocated:  capacity.CPUCoresAllocated,
		MemoryGbTotal:      capacity.MemoryGBTotal,
		MemoryGbAvailable:  capacity.MemoryGBAvailable,
		MemoryGbAllocated:  capacity.MemoryGBAllocated,
		GpusTotal:          capacity.GPUsTotal,
		GpusAvailable:      capacity.GPUsAvailable,
		GpusAllocated:      capacity.GPUsAllocated,
		GpuType:            capacity.GPUType,
		StorageGbTotal:     capacity.StorageGBTotal,
		StorageGbAvailable: capacity.StorageGBAvailable,
		StorageGbAllocated: capacity.StorageGBAllocated,
	}
}

func nodeHealthToProto(health types.NodeHealth) *hpcv1.NodeHealth {
	return &hpcv1.NodeHealth{
		Status:                      healthStatusToProto(health.Status),
		UptimeSeconds:               health.UptimeSeconds,
		LoadAverage_1M:              health.LoadAverage1m,
		LoadAverage_5M:              health.LoadAverage5m,
		LoadAverage_15M:             health.LoadAverage15m,
		CpuUtilizationPercent:       health.CPUUtilizationPercent,
		MemoryUtilizationPercent:    health.MemoryUtilizationPercent,
		GpuUtilizationPercent:       health.GPUUtilizationPercent,
		GpuMemoryUtilizationPercent: health.GPUMemoryUtilizationPercent,
		DiskIoUtilizationPercent:    health.DiskIOUtilizationPercent,
		NetworkUtilizationPercent:   health.NetworkUtilizationPercent,
		TemperatureCelsius:          health.TemperatureCelsius,
		GpuTemperatureCelsius:       health.GPUTemperatureCelsius,
		ErrorCount_24H:              health.ErrorCount24h,
		WarningCount_24H:            health.WarningCount24h,
		LastErrorMessage:            health.LastErrorMessage,
		SlurmState:                  health.SLURMState,
	}
}

func nodeHardwareToProto(hardware types.NodeHardware) *hpcv1.NodeHardware {
	return &hpcv1.NodeHardware{
		CpuModel:       hardware.CPUModel,
		CpuVendor:      hardware.CPUVendor,
		CpuArch:        hardware.CPUArch,
		Sockets:        hardware.Sockets,
		CoresPerSocket: hardware.CoresPerSocket,
		ThreadsPerCore: hardware.ThreadsPerCore,
		MemoryType:     hardware.MemoryType,
		MemorySpeedMhz: hardware.MemorySpeedMHz,
		GpuModel:       hardware.GPUModel,
		GpuMemoryGb:    hardware.GPUMemoryGB,
		StorageType:    hardware.StorageType,
		Features:       hardware.Features,
	}
}

func nodeTopologyToProto(topology types.NodeTopology) *hpcv1.NodeTopology {
	return &hpcv1.NodeTopology{
		NumaNodes:     topology.NUMANodes,
		NumaMemoryGb:  topology.NUMAMemoryGB,
		Interconnect:  topology.Interconnect,
		NetworkFabric: topology.NetworkFabric,
		TopologyHint:  topology.TopologyHint,
	}
}

func nodeLocalityToProto(locality types.NodeLocality) *hpcv1.NodeLocality {
	return &hpcv1.NodeLocality{
		Region:     locality.Region,
		Datacenter: locality.Datacenter,
		Zone:       locality.Zone,
		Rack:       locality.Rack,
		Row:        locality.Row,
		Position:   locality.Position,
	}
}

func nodeStateToProto(state types.NodeState) hpcv1.NodeState {
	switch state {
	case types.NodeStateUnknown:
		return hpcv1.NodeStateUnknown
	case types.NodeStatePending:
		return hpcv1.NodeStatePending
	case types.NodeStateActive:
		return hpcv1.NodeStateActive
	case types.NodeStateStale:
		return hpcv1.NodeStateStale
	case types.NodeStateDraining:
		return hpcv1.NodeStateDraining
	case types.NodeStateDrained:
		return hpcv1.NodeStateDrained
	case types.NodeStateOffline:
		return hpcv1.NodeStateOffline
	case types.NodeStateDeregistered:
		return hpcv1.NodeStateDeregistered
	default:
		return hpcv1.NodeStateUnspecified
	}
}

func healthStatusToProto(status types.HealthStatus) hpcv1.HealthStatus {
	switch status {
	case types.HealthStatusHealthy:
		return hpcv1.HealthStatusHealthy
	case types.HealthStatusDegraded:
		return hpcv1.HealthStatusDegraded
	case types.HealthStatusUnhealthy:
		return hpcv1.HealthStatusUnhealthy
	case types.HealthStatusDraining:
		return hpcv1.HealthStatusDraining
	case types.HealthStatusOffline:
		return hpcv1.HealthStatusOffline
	default:
		return hpcv1.HealthStatusUnspecified
	}
}

func paramsToProto(params types.Params) hpcv1.Params {
	return hpcv1.Params{
		PlatformFeeRate:           params.PlatformFeeRate,
		ProviderRewardRate:        params.ProviderRewardRate,
		NodeRewardRate:            params.NodeRewardRate,
		MinJobDurationSeconds:     params.MinJobDurationSeconds,
		MaxJobDurationSeconds:     params.MaxJobDurationSeconds,
		DefaultIdentityThreshold:  params.DefaultIdentityThreshold,
		ClusterHeartbeatTimeout:   params.ClusterHeartbeatTimeout,
		NodeHeartbeatTimeout:      params.NodeHeartbeatTimeout,
		LatencyWeightFactor:       params.LatencyWeightFactor,
		CapacityWeightFactor:      params.CapacityWeightFactor,
		MaxLatencyMs:              params.MaxLatencyMs,
		DisputeResolutionPeriod:   params.DisputeResolutionPeriod,
		RewardFormulaVersion:      params.RewardFormulaVersion,
		EnableProximityClustering: params.EnableProximityClustering,
	}
}
