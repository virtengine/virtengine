package keeper_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
)

// VE-503: Proximity-based mini-supercomputer clustering tests

// TestSchedulingDecisionScoring tests the scoring algorithm for cluster selection
func TestSchedulingDecisionScoring(t *testing.T) {
	// Test fixed-point arithmetic for determinism
	// Scale: 1,000,000 = 100%

	testCases := []struct {
		name             string
		latencyMs        uint64
		maxLatencyMs     uint64
		availableNodes   uint64
		totalNodes       uint64
		latencyWeight    int64
		capacityWeight   int64
		expectedMinScore int64
		expectedMaxScore int64
	}{
		{
			name:             "perfect latency, full capacity",
			latencyMs:        10,
			maxLatencyMs:     100,
			availableNodes:   10,
			totalNodes:       10,
			latencyWeight:    500000, // 50%
			capacityWeight:   500000, // 50%
			expectedMinScore: 900000, // High score
			expectedMaxScore: 1000000,
		},
		{
			name:             "max latency, half capacity",
			latencyMs:        100,
			maxLatencyMs:     100,
			availableNodes:   5,
			totalNodes:       10,
			latencyWeight:    500000,
			capacityWeight:   500000,
			expectedMinScore: 200000,
			expectedMaxScore: 350000,
		},
		{
			name:             "zero latency, zero capacity",
			latencyMs:        0,
			maxLatencyMs:     100,
			availableNodes:   0,
			totalNodes:       10,
			latencyWeight:    500000,
			capacityWeight:   500000,
			expectedMinScore: 400000,
			expectedMaxScore: 600000,
		},
		{
			name:             "latency-only weight",
			latencyMs:        50,
			maxLatencyMs:     100,
			availableNodes:   10,
			totalNodes:       10,
			latencyWeight:    1000000, // 100%
			capacityWeight:   0,       // 0%
			expectedMinScore: 400000,
			expectedMaxScore: 600000,
		},
		{
			name:             "capacity-only weight",
			latencyMs:        100,
			maxLatencyMs:     100,
			availableNodes:   10,
			totalNodes:       10,
			latencyWeight:    0,       // 0%
			capacityWeight:   1000000, // 100%
			expectedMinScore: 900000,
			expectedMaxScore: 1000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := calculateClusterScore(
				tc.latencyMs,
				tc.maxLatencyMs,
				tc.availableNodes,
				tc.totalNodes,
				tc.latencyWeight,
				tc.capacityWeight,
			)

			require.GreaterOrEqual(t, score, tc.expectedMinScore,
				"score should be >= %d, got %d", tc.expectedMinScore, score)
			require.LessOrEqual(t, score, tc.expectedMaxScore,
				"score should be <= %d, got %d", tc.expectedMaxScore, score)
		})
	}
}

// calculateClusterScore mirrors the keeper's scoring logic for testing
//
//nolint:gosec // G115: integer overflow conversion is acceptable in test code with controlled inputs
func calculateClusterScore(
	latencyMs, maxLatencyMs uint64,
	availableNodes, totalNodes uint64,
	latencyWeight, capacityWeight int64,
) int64 {
	const scale = int64(1000000)

	// Latency score: lower is better
	// latencyScore = (maxLatency - actualLatency) / maxLatency
	var latencyScore int64
	if maxLatencyMs > 0 && latencyMs <= maxLatencyMs {
		latencyScore = (int64(maxLatencyMs) - int64(latencyMs)) * scale / int64(maxLatencyMs) //nolint:gosec // test code
	}

	// Capacity score: higher availability is better
	// capacityScore = availableNodes / totalNodes
	var capacityScore int64
	if totalNodes > 0 {
		capacityScore = int64(availableNodes) * scale / int64(totalNodes) //nolint:gosec // test code
	}

	// Combined score with weights (fixed-point multiplication)
	totalScore := (latencyScore*latencyWeight + capacityScore*capacityWeight) / scale

	return totalScore
}

// TestSchedulingNodeMetadataValidation tests node metadata validation for scheduling
func TestSchedulingNodeMetadataValidation(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()

	testCases := []struct {
		name        string
		metadata    types.NodeMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid metadata",
			metadata: types.NodeMetadata{
				NodeID:          "node-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Region:          "us-east-1",
				Datacenter:      "dc1",
				Resources: types.NodeResources{
					CPUCores:  32,
					MemoryGB:  64,
					GPUs:      2,
					StorageGB: 1000,
				},
			},
			expectError: false,
		},
		{
			name: "missing node ID",
			metadata: types.NodeMetadata{
				NodeID:          "",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Region:          "us-east-1",
			},
			expectError: true,
			errorMsg:    "node_id",
		},
		{
			name: "missing cluster ID",
			metadata: types.NodeMetadata{
				NodeID:          "node-001",
				ClusterID:       "",
				ProviderAddress: validAddr,
				Region:          "us-east-1",
			},
			expectError: true,
			errorMsg:    "cluster_id",
		},
		{
			name: "missing region",
			metadata: types.NodeMetadata{
				NodeID:          "node-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Region:          "",
			},
			expectError: true,
			errorMsg:    "region",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.metadata.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSchedulingDecisionValidationFields tests scheduling decision validation
func TestSchedulingDecisionValidationFields(t *testing.T) {
	testCases := []struct {
		name        string
		decision    types.SchedulingDecision
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid decision",
			decision: types.SchedulingDecision{
				DecisionID:        "decision-001",
				JobID:             "job-001",
				SelectedClusterID: "cluster-001",
				CandidateClusters: []types.ClusterCandidate{
					{ClusterID: "cluster-001", CombinedScore: "850000", LatencyScore: "900000", CapacityScore: "800000"},
					{ClusterID: "cluster-002", CombinedScore: "750000", LatencyScore: "700000", CapacityScore: "800000"},
				},
				DecisionReason: "highest score",
			},
			expectError: false,
		},
		{
			name: "missing job ID",
			decision: types.SchedulingDecision{
				DecisionID:        "decision-001",
				JobID:             "",
				SelectedClusterID: "cluster-001",
				DecisionReason:    "test",
			},
			expectError: true,
			errorMsg:    "job_id",
		},
		{
			name: "no selected cluster",
			decision: types.SchedulingDecision{
				DecisionID:        "decision-001",
				JobID:             "job-001",
				SelectedClusterID: "",
				DecisionReason:    "test",
			},
			expectError: true,
			errorMsg:    "selected_cluster_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.decision.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestClusterCandidateSorting tests that clusters are sorted by score correctly
func TestClusterCandidateSorting(t *testing.T) {
	candidates := []types.ClusterCandidate{
		{ClusterID: "cluster-001", CombinedScore: "500000"},
		{ClusterID: "cluster-002", CombinedScore: "900000"},
		{ClusterID: "cluster-003", CombinedScore: "750000"},
		{ClusterID: "cluster-004", CombinedScore: "600000"},
	}

	// Sort candidates by score descending
	sorted := keeper.SortCandidatesByScore(candidates)

	require.Equal(t, "cluster-002", sorted[0].ClusterID, "highest score should be first")
	require.Equal(t, "cluster-003", sorted[1].ClusterID)
	require.Equal(t, "cluster-004", sorted[2].ClusterID)
	require.Equal(t, "cluster-001", sorted[3].ClusterID, "lowest score should be last")
}

// TestLatencyMeasurementAggregation tests aggregating latency measurements
func TestLatencyMeasurementAggregation(t *testing.T) {
	measurements := []types.LatencyMeasurement{
		{TargetNodeID: "node-001", LatencyMs: 10},
		{TargetNodeID: "node-002", LatencyMs: 20},
		{TargetNodeID: "node-003", LatencyMs: 15},
	}

	// Average latency should be 15ms
	avgLatency := keeper.CalculateAverageLatency(measurements)
	require.Equal(t, int64(15), avgLatency)

	// Empty measurements should return 0
	emptyAvg := keeper.CalculateAverageLatency([]types.LatencyMeasurement{})
	require.Equal(t, int64(0), emptyAvg)
}

// TestEligibleClusterFiltering tests filtering clusters by job requirements
func TestEligibleClusterFiltering(t *testing.T) {
	clusters := []types.HPCCluster{
		{
			ClusterID:      "cluster-001",
			State:          types.ClusterStateActive,
			TotalNodes:     10,
			AvailableNodes: 8,
		},
		{
			ClusterID:      "cluster-002",
			State:          types.ClusterStateDraining, // Not active
			TotalNodes:     20,
			AvailableNodes: 15,
		},
		{
			ClusterID:      "cluster-003",
			State:          types.ClusterStateActive,
			TotalNodes:     5,
			AvailableNodes: 5,
		},
	}

	testCases := []struct {
		name          string
		resources     types.JobResources
		expectedCount int
		expectedIDs   []string
	}{
		{
			name: "basic node requirements",
			resources: types.JobResources{
				Nodes: 4,
			},
			expectedCount: 2, // Clusters 1 and 3 (cluster 2 is draining)
			expectedIDs:   []string{"cluster-001", "cluster-003"},
		},
		{
			name: "high node requirements",
			resources: types.JobResources{
				Nodes: 6,
			},
			expectedCount: 1, // Only cluster 1 has enough nodes
			expectedIDs:   []string{"cluster-001"},
		},
		{
			name: "exceeds all clusters",
			resources: types.JobResources{
				Nodes: 50,
			},
			expectedCount: 0,
			expectedIDs:   []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eligible := keeper.FilterEligibleClusters(clusters, tc.resources)
			require.Equal(t, tc.expectedCount, len(eligible))

			for i, id := range tc.expectedIDs {
				require.Equal(t, id, eligible[i].ClusterID)
			}
		})
	}
}

// TestScheduleJobPriorityOrdering validates that higher score clusters are selected first.
func TestScheduleJobPriorityOrdering(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	clusterA := types.HPCCluster{
		ClusterID:       "cluster-a",
		ProviderAddress: providerAddr,
		Name:            "cluster-a",
		State:           types.ClusterStateActive,
		TotalNodes:      10,
		AvailableNodes:  8,
		Region:          "us-east-1",
	}
	clusterB := types.HPCCluster{
		ClusterID:       "cluster-b",
		ProviderAddress: providerAddr,
		Name:            "cluster-b",
		State:           types.ClusterStateActive,
		TotalNodes:      6,
		AvailableNodes:  4,
		Region:          "us-east-1",
	}

	mustSetCluster(t, ctx, k, clusterA)
	mustSetCluster(t, ctx, k, clusterB)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-a-1",
		ClusterID:       "cluster-a",
		ProviderAddress: providerAddr,
		Region:          "us-east-1",
		AvgLatencyMs:    10,
		Active:          true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-b-1",
		ClusterID:       "cluster-b",
		ProviderAddress: providerAddr,
		Region:          "us-east-1",
		AvgLatencyMs:    40,
		Active:          true,
	})

	offering := types.HPCOffering{
		OfferingID:        "offering-1",
		ClusterID:         "cluster-a",
		ProviderAddress:   providerAddr,
		Name:              "standard",
		MaxRuntimeSeconds: 3600,
		Active:            true,
	}
	mustSetOffering(t, ctx, k, offering)

	job := types.HPCJob{
		JobID:           "job-1",
		OfferingID:      "offering-1",
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes: 2,
		},
	}

	decision, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)
	require.Equal(t, "cluster-a", decision.SelectedClusterID)
	require.GreaterOrEqual(t, len(decision.CandidateClusters), 2)
	require.Equal(t, "cluster-a", decision.CandidateClusters[0].ClusterID)
}

// TestScheduleJobResourceAllocation ensures resource requests route to eligible clusters.
func TestScheduleJobResourceAllocation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{3}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{4}, 20)).String()

	clusterSmall := types.HPCCluster{
		ClusterID:       "cluster-small",
		ProviderAddress: providerAddr,
		Name:            "cluster-small",
		State:           types.ClusterStateActive,
		TotalNodes:      3,
		AvailableNodes:  2,
		Region:          "us-west-2",
	}
	clusterLarge := types.HPCCluster{
		ClusterID:       "cluster-large",
		ProviderAddress: providerAddr,
		Name:            "cluster-large",
		State:           types.ClusterStateActive,
		TotalNodes:      12,
		AvailableNodes:  10,
		Region:          "us-west-2",
	}

	mustSetCluster(t, ctx, k, clusterSmall)
	mustSetCluster(t, ctx, k, clusterLarge)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-large-1",
		ClusterID:       "cluster-large",
		ProviderAddress: providerAddr,
		Region:          "us-west-2",
		AvgLatencyMs:    15,
		Active:          true,
	})

	offering := types.HPCOffering{
		OfferingID:        "offering-2",
		ClusterID:         "cluster-large",
		ProviderAddress:   providerAddr,
		Name:              "gpu-standard",
		MaxRuntimeSeconds: 7200,
		Active:            true,
	}
	mustSetOffering(t, ctx, k, offering)

	job := types.HPCJob{
		JobID:           "job-2",
		OfferingID:      "offering-2",
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes: 4,
		},
	}

	decision, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)
	require.Equal(t, "cluster-large", decision.SelectedClusterID)
}

// TestScheduleJobPreemptionLogic simulates a higher-priority job arriving when capacity is exhausted.
func TestScheduleJobPreemptionLogic(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{6}, 20)).String()

	clusterPrimary := types.HPCCluster{
		ClusterID:       "cluster-primary",
		ProviderAddress: providerAddr,
		Name:            "cluster-primary",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  4,
		Region:          "eu-central-1",
	}
	clusterFallback := types.HPCCluster{
		ClusterID:       "cluster-fallback",
		ProviderAddress: providerAddr,
		Name:            "cluster-fallback",
		State:           types.ClusterStateActive,
		TotalNodes:      12,
		AvailableNodes:  12,
		Region:          "eu-central-1",
	}

	mustSetCluster(t, ctx, k, clusterPrimary)
	mustSetCluster(t, ctx, k, clusterFallback)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-primary-1",
		ClusterID:       "cluster-primary",
		ProviderAddress: providerAddr,
		Region:          "eu-central-1",
		AvgLatencyMs:    10,
		Active:          true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-fallback-1",
		ClusterID:       "cluster-fallback",
		ProviderAddress: providerAddr,
		Region:          "eu-central-1",
		AvgLatencyMs:    80,
		Active:          true,
	})

	offering := types.HPCOffering{
		OfferingID:        "offering-3",
		ClusterID:         "cluster-primary",
		ProviderAddress:   providerAddr,
		Name:              "priority-queue",
		MaxRuntimeSeconds: 3600,
		Active:            true,
	}
	mustSetOffering(t, ctx, k, offering)

	initialJob := types.HPCJob{
		JobID:           "job-initial",
		OfferingID:      "offering-3",
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes: 4,
		},
	}

	initialDecision, err := k.ScheduleJob(ctx, &initialJob)
	require.NoError(t, err)
	require.Equal(t, "cluster-primary", initialDecision.SelectedClusterID)

	clusterPrimary.AvailableNodes = 0
	mustSetCluster(t, ctx, k, clusterPrimary)

	urgentJob := types.HPCJob{
		JobID:           "job-urgent",
		OfferingID:      "offering-3",
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes: 4,
		},
	}

	decision, err := k.ScheduleJob(ctx, &urgentJob)
	require.NoError(t, err)
	require.Equal(t, "cluster-fallback", decision.SelectedClusterID)
	require.True(t, decision.IsFallback)
}

// TestSchedulingFairnessAcrossUsers ensures identical jobs from different users are treated consistently.
func TestSchedulingFairnessAcrossUsers(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String()
	customerA := sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String()
	customerB := sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String()

	cluster := types.HPCCluster{
		ClusterID:       "cluster-fair",
		ProviderAddress: providerAddr,
		Name:            "cluster-fair",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  8,
		Region:          "ap-south-1",
	}
	mustSetCluster(t, ctx, k, cluster)
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-fair-1",
		ClusterID:       "cluster-fair",
		ProviderAddress: providerAddr,
		Region:          "ap-south-1",
		AvgLatencyMs:    20,
		Active:          true,
	})

	offering := types.HPCOffering{
		OfferingID:        "offering-4",
		ClusterID:         "cluster-fair",
		ProviderAddress:   providerAddr,
		Name:              "fair-share",
		MaxRuntimeSeconds: 3600,
		Active:            true,
	}
	mustSetOffering(t, ctx, k, offering)

	jobA := types.HPCJob{
		JobID:           "job-fair-a",
		OfferingID:      "offering-4",
		CustomerAddress: customerA,
		Resources: types.JobResources{
			Nodes: 2,
		},
	}
	jobB := types.HPCJob{
		JobID:           "job-fair-b",
		OfferingID:      "offering-4",
		CustomerAddress: customerB,
		Resources: types.JobResources{
			Nodes: 2,
		},
	}

	decisionA, err := k.ScheduleJob(ctx, &jobA)
	require.NoError(t, err)
	decisionB, err := k.ScheduleJob(ctx, &jobB)
	require.NoError(t, err)

	require.Equal(t, decisionA.SelectedClusterID, decisionB.SelectedClusterID)
}

// TestClusterCapacityChecksBeforeScheduling verifies no scheduling happens when capacity is insufficient.
func TestClusterCapacityChecksBeforeScheduling(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{10}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{11}, 20)).String()

	cluster := types.HPCCluster{
		ClusterID:       "cluster-capacity",
		ProviderAddress: providerAddr,
		Name:            "cluster-capacity",
		State:           types.ClusterStateActive,
		TotalNodes:      4,
		AvailableNodes:  1,
		Region:          "us-central-1",
	}
	mustSetCluster(t, ctx, k, cluster)

	offering := types.HPCOffering{
		OfferingID:        "offering-5",
		ClusterID:         "cluster-capacity",
		ProviderAddress:   providerAddr,
		Name:              "capacity-check",
		MaxRuntimeSeconds: 3600,
		Active:            true,
	}
	mustSetOffering(t, ctx, k, offering)

	job := types.HPCJob{
		JobID:           "job-capacity",
		OfferingID:      "offering-5",
		CustomerAddress: customerAddr,
		Resources: types.JobResources{
			Nodes: 2,
		},
	}

	_, err := k.ScheduleJob(ctx, &job)
	require.ErrorIs(t, err, types.ErrNoAvailableCluster)
}

// TestSchedulingWithHeterogeneousResources verifies resource-aware filtering with GPU-heavy jobs.
func TestSchedulingWithHeterogeneousResources(t *testing.T) {
	clusters := []types.HPCCluster{
		{
			ClusterID:      "cluster-cpu",
			State:          types.ClusterStateActive,
			TotalNodes:     8,
			AvailableNodes: 8,
			ClusterMetadata: types.ClusterMetadata{
				TotalCPUCores: 256,
				TotalMemoryGB: 512,
				TotalGPUs:     0,
			},
		},
		{
			ClusterID:      "cluster-gpu",
			State:          types.ClusterStateActive,
			TotalNodes:     4,
			AvailableNodes: 4,
			ClusterMetadata: types.ClusterMetadata{
				TotalCPUCores: 128,
				TotalMemoryGB: 256,
				TotalGPUs:     16,
			},
		},
		{
			ClusterID:      "cluster-mixed",
			State:          types.ClusterStateActive,
			TotalNodes:     2,
			AvailableNodes: 2,
			ClusterMetadata: types.ClusterMetadata{
				TotalCPUCores: 64,
				TotalMemoryGB: 128,
				TotalGPUs:     4,
			},
		},
	}

	resources := types.JobResources{
		Nodes:           2,
		CPUCoresPerNode: 16,
		MemoryGBPerNode: 64,
		GPUsPerNode:     4,
	}

	eligible := keeper.FilterEligibleClusters(clusters, resources)
	require.Len(t, eligible, 1)
	require.Equal(t, "cluster-gpu", eligible[0].ClusterID)
}

// TestSchedulingFairShareBias ensures tenants with heavy usage are deprioritized.
func TestSchedulingFairShareBias(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{12}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{13}, 20)).String()

	clusterA := types.HPCCluster{
		ClusterID:       "cluster-usage-heavy",
		ProviderAddress: providerAddr,
		Name:            "cluster-usage-heavy",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  8,
		Region:          "us-west-1",
		Partitions: []types.Partition{
			{Name: "default", Priority: 500},
		},
	}
	clusterB := types.HPCCluster{
		ClusterID:       "cluster-usage-light",
		ProviderAddress: providerAddr,
		Name:            "cluster-usage-light",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  8,
		Region:          "us-west-1",
		Partitions: []types.Partition{
			{Name: "default", Priority: 500},
		},
	}

	mustSetCluster(t, ctx, k, clusterA)
	mustSetCluster(t, ctx, k, clusterB)

	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-a",
		ClusterID:       "cluster-usage-heavy",
		ProviderAddress: providerAddr,
		Region:          "us-west-1",
		AvgLatencyMs:    20,
		Active:          true,
	})
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-b",
		ClusterID:       "cluster-usage-light",
		ProviderAddress: providerAddr,
		Region:          "us-west-1",
		AvgLatencyMs:    20,
		Active:          true,
	})

	offering := types.HPCOffering{
		OfferingID:        "offering-fair-share",
		ClusterID:         "cluster-usage-heavy",
		ProviderAddress:   providerAddr,
		Name:              "fair-share",
		MaxRuntimeSeconds: 3600,
		QueueOptions: []types.QueueOption{
			{PartitionName: "default", DisplayName: "default"},
		},
		Active: true,
	}
	mustSetOffering(t, ctx, k, offering)

	// Heavy usage on cluster A
	mustSetJob(t, ctx, k, types.HPCJob{
		JobID:           "job-running-heavy",
		OfferingID:      offering.OfferingID,
		ClusterID:       "cluster-usage-heavy",
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		QueueName:       "default",
		State:           types.JobStateRunning,
		Resources: types.JobResources{
			Nodes: 2,
		},
	})

	job := types.HPCJob{
		JobID:           "job-fairness",
		OfferingID:      offering.OfferingID,
		CustomerAddress: customerAddr,
		QueueName:       "default",
		Resources: types.JobResources{
			Nodes: 1,
		},
	}

	decision, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)
	require.Equal(t, "cluster-usage-light", decision.SelectedClusterID)
}

// TestSchedulingQuotaEnforcement ensures tenant quotas are enforced.
func TestSchedulingQuotaEnforcement(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{14}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{15}, 20)).String()

	cluster := types.HPCCluster{
		ClusterID:       "cluster-quota",
		ProviderAddress: providerAddr,
		Name:            "cluster-quota",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  8,
		Region:          "us-east-2",
		Partitions: []types.Partition{
			{Name: "default", Priority: 400},
		},
	}
	mustSetCluster(t, ctx, k, cluster)

	offering := types.HPCOffering{
		OfferingID:        "offering-quota",
		ClusterID:         "cluster-quota",
		ProviderAddress:   providerAddr,
		Name:              "quota",
		MaxRuntimeSeconds: 3600,
		QueueOptions: []types.QueueOption{
			{PartitionName: "default", DisplayName: "default"},
		},
		Active: true,
	}
	mustSetOffering(t, ctx, k, offering)

	// Existing running jobs already consume quota/burst
	mustSetJob(t, ctx, k, types.HPCJob{
		JobID:           "job-running-quota",
		OfferingID:      offering.OfferingID,
		ClusterID:       cluster.ClusterID,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		QueueName:       "default",
		State:           types.JobStateRunning,
		Resources: types.JobResources{
			Nodes: 4,
		},
	})

	job := types.HPCJob{
		JobID:           "job-quota-new",
		OfferingID:      offering.OfferingID,
		CustomerAddress: customerAddr,
		QueueName:       "default",
		Resources: types.JobResources{
			Nodes: 2,
		},
	}

	_, err := k.ScheduleJob(ctx, &job)
	require.ErrorIs(t, err, types.ErrTenantQuotaExceeded)
}

// TestSchedulingPreemptionPlan ensures preemption is planned deterministically.
func TestSchedulingPreemptionPlan(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{16}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{17}, 20)).String()

	cluster := types.HPCCluster{
		ClusterID:       "cluster-preempt",
		ProviderAddress: providerAddr,
		Name:            "cluster-preempt",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  0,
		Region:          "eu-west-1",
		Partitions: []types.Partition{
			{Name: "low", Priority: 100},
			{Name: "high", Priority: 900},
		},
	}
	mustSetCluster(t, ctx, k, cluster)
	mustSetNode(t, ctx, k, types.NodeMetadata{
		NodeID:          "node-preempt-1",
		ClusterID:       cluster.ClusterID,
		ProviderAddress: providerAddr,
		Region:          "eu-west-1",
		AvgLatencyMs:    15,
		Active:          true,
	})

	offering := types.HPCOffering{
		OfferingID:        "offering-preempt",
		ClusterID:         cluster.ClusterID,
		ProviderAddress:   providerAddr,
		Name:              "preempt",
		MaxRuntimeSeconds: 7200,
		QueueOptions: []types.QueueOption{
			{PartitionName: "low", DisplayName: "low"},
			{PartitionName: "high", DisplayName: "high"},
		},
		Active: true,
	}
	mustSetOffering(t, ctx, k, offering)

	// Low priority running job
	mustSetJob(t, ctx, k, types.HPCJob{
		JobID:           "job-preempted",
		OfferingID:      offering.OfferingID,
		ClusterID:       cluster.ClusterID,
		ProviderAddress: providerAddr,
		CustomerAddress: customerAddr,
		QueueName:       "low",
		State:           types.JobStateRunning,
		Resources: types.JobResources{
			Nodes: 4,
		},
	})

	job := types.HPCJob{
		JobID:           "job-preemptor",
		OfferingID:      offering.OfferingID,
		CustomerAddress: customerAddr,
		QueueName:       "high",
		Resources: types.JobResources{
			Nodes: 4,
		},
	}

	decision, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)
	require.True(t, decision.PreemptionPlanned)
	require.Contains(t, decision.PreemptedJobIDs, "job-preempted")
	require.Equal(t, cluster.ClusterID, decision.SelectedClusterID)
}
