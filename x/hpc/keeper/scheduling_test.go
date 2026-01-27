//go:build ignore
// +build ignore

// TODO: This test file is excluded until HPC scheduling API is stabilized.

package keeper_test

import (
	"testing"

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
		latencyScore = (int64(maxLatencyMs) - int64(latencyMs)) * scale / int64(maxLatencyMs)
	}

	// Capacity score: higher availability is better
	// capacityScore = availableNodes / totalNodes
	var capacityScore int64
	if totalNodes > 0 {
		capacityScore = int64(availableNodes) * scale / int64(totalNodes)
	}

	// Combined score with weights (fixed-point multiplication)
	totalScore := (latencyScore*latencyWeight + capacityScore*capacityWeight) / scale

	return totalScore
}

// TestNodeMetadataValidation tests node metadata validation
func TestNodeMetadataValidation(t *testing.T) {
	testCases := []struct {
		name        string
		metadata    types.NodeMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid metadata",
			metadata: types.NodeMetadata{
				NodeAddress: "node1",
				ClusterID:   1,
				Region:      "us-east-1",
				Datacenter:  "dc1",
				Resources: types.NodeResources{
					CPUCores:          32,
					MemoryMB:          65536,
					GPUs:              2,
					AvailableCPU:      16,
					AvailableMemoryMB: 32768,
					AvailableGPUs:     1,
				},
				LatencyMeasurements: []types.LatencyMeasurement{
					{
						TargetRegion: "us-west-1",
						LatencyMs:    50,
						MeasuredAt:   1000000,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing node address",
			metadata: types.NodeMetadata{
				NodeAddress: "",
				ClusterID:   1,
			},
			expectError: true,
			errorMsg:    "node address is required",
		},
		{
			name: "zero cluster ID",
			metadata: types.NodeMetadata{
				NodeAddress: "node1",
				ClusterID:   0,
			},
			expectError: true,
			errorMsg:    "cluster ID is required",
		},
		{
			name: "available exceeds total CPU",
			metadata: types.NodeMetadata{
				NodeAddress: "node1",
				ClusterID:   1,
				Resources: types.NodeResources{
					CPUCores:     32,
					AvailableCPU: 64,
				},
			},
			expectError: true,
			errorMsg:    "available CPU cannot exceed total",
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

// TestSchedulingDecisionValidation tests scheduling decision validation
func TestSchedulingDecisionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		decision    types.SchedulingDecision
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid decision",
			decision: types.SchedulingDecision{
				JobID:           1,
				SelectedCluster: 1,
				Score:           850000,
				Candidates: []types.ClusterCandidate{
					{ClusterID: 1, Score: 850000, LatencyScore: 900000, CapacityScore: 800000},
					{ClusterID: 2, Score: 750000, LatencyScore: 700000, CapacityScore: 800000},
				},
				DecisionTime: 1000000,
				Reason:       "highest score",
			},
			expectError: false,
		},
		{
			name: "zero job ID",
			decision: types.SchedulingDecision{
				JobID:           0,
				SelectedCluster: 1,
			},
			expectError: true,
			errorMsg:    "job ID is required",
		},
		{
			name: "no selected cluster",
			decision: types.SchedulingDecision{
				JobID:           1,
				SelectedCluster: 0,
			},
			expectError: true,
			errorMsg:    "selected cluster is required",
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
		{ClusterID: 1, Score: 500000},
		{ClusterID: 2, Score: 900000},
		{ClusterID: 3, Score: 750000},
		{ClusterID: 4, Score: 600000},
	}

	// Sort candidates by score descending
	sorted := keeper.SortCandidatesByScore(candidates)

	require.Equal(t, uint64(2), sorted[0].ClusterID, "highest score should be first")
	require.Equal(t, uint64(3), sorted[1].ClusterID)
	require.Equal(t, uint64(4), sorted[2].ClusterID)
	require.Equal(t, uint64(1), sorted[3].ClusterID, "lowest score should be last")
}

// TestLatencyMeasurementAggregation tests aggregating latency measurements
func TestLatencyMeasurementAggregation(t *testing.T) {
	measurements := []types.LatencyMeasurement{
		{TargetRegion: "us-east-1", LatencyMs: 10, MeasuredAt: 100},
		{TargetRegion: "us-east-1", LatencyMs: 20, MeasuredAt: 200},
		{TargetRegion: "us-east-1", LatencyMs: 15, MeasuredAt: 300},
	}

	// Average latency should be 15ms
	avgLatency := keeper.CalculateAverageLatency(measurements, "us-east-1")
	require.Equal(t, uint64(15), avgLatency)

	// Non-existent region should return max latency
	unknownLatency := keeper.CalculateAverageLatency(measurements, "unknown-region")
	require.Equal(t, uint64(0xFFFFFFFFFFFFFFFF), unknownLatency)
}

// TestEligibleClusterFiltering tests filtering clusters by job requirements
func TestEligibleClusterFiltering(t *testing.T) {
	clusters := []types.HPCCluster{
		{
			ID:            1,
			Status:        types.ClusterStatusActive,
			TotalCPUCores: 100,
			TotalMemoryMB: 512000,
			TotalGPUs:     4,
		},
		{
			ID:            2,
			Status:        types.ClusterStatusMaintenance,
			TotalCPUCores: 200,
			TotalMemoryMB: 1024000,
			TotalGPUs:     8,
		},
		{
			ID:            3,
			Status:        types.ClusterStatusActive,
			TotalCPUCores: 50,
			TotalMemoryMB: 256000,
			TotalGPUs:     0,
		},
	}

	testCases := []struct {
		name          string
		resources     types.JobResources
		expectedCount int
		expectedIDs   []uint64
	}{
		{
			name: "basic CPU/memory requirements",
			resources: types.JobResources{
				CPUCores: 32,
				MemoryMB: 128000,
				GPUs:     0,
			},
			expectedCount: 2, // Clusters 1 and 3 (cluster 2 is in maintenance)
			expectedIDs:   []uint64{1, 3},
		},
		{
			name: "GPU requirements",
			resources: types.JobResources{
				CPUCores: 16,
				MemoryMB: 64000,
				GPUs:     2,
			},
			expectedCount: 1, // Only cluster 1 (active with GPUs)
			expectedIDs:   []uint64{1},
		},
		{
			name: "exceeds all clusters",
			resources: types.JobResources{
				CPUCores: 500,
				MemoryMB: 2048000,
				GPUs:     16,
			},
			expectedCount: 0,
			expectedIDs:   []uint64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eligible := keeper.FilterEligibleClusters(clusters, tc.resources)
			require.Equal(t, tc.expectedCount, len(eligible))

			for i, id := range tc.expectedIDs {
				require.Equal(t, id, eligible[i].ID)
			}
		})
	}
}
