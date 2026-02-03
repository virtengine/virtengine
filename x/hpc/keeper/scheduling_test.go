package keeper_test

import (
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
