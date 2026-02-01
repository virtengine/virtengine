package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// TestClusterValidation tests HPC cluster validation
func TestClusterValidation(t *testing.T) {
	// Create a valid bech32 address for tests
	validAddr := sdk.AccAddress(make([]byte, 20)).String()

	testCases := []struct {
		name        string
		cluster     types.HPCCluster
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid cluster",
			cluster: types.HPCCluster{
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Name:            "test-cluster",
				State:           types.ClusterStateActive,
				TotalNodes:      10,
				AvailableNodes:  8,
				Region:          "us-west-2",
				Partitions: []types.Partition{
					{
						Name:           "default",
						Nodes:          10,
						MaxRuntime:     86400,
						DefaultRuntime: 3600,
						MaxNodes:       10,
						Priority:       1,
						State:          "up",
					},
				},
				ClusterMetadata: types.ClusterMetadata{
					TotalCPUCores:    100,
					TotalMemoryGB:    1024,
					TotalGPUs:        8,
					InterconnectType: "infiniband",
					StorageType:      "lustre",
					TotalStorageGB:   10000,
				},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				BlockHeight: 100,
			},
			expectError: false,
		},
		{
			name: "missing cluster_id",
			cluster: types.HPCCluster{
				ClusterID:       "",
				ProviderAddress: validAddr,
				Name:            "test-cluster",
				State:           types.ClusterStateActive,
				TotalNodes:      10,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "cluster_id cannot be empty",
		},
		{
			name: "missing name",
			cluster: types.HPCCluster{
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Name:            "",
				State:           types.ClusterStateActive,
				TotalNodes:      10,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name: "invalid provider address",
			cluster: types.HPCCluster{
				ClusterID:       "cluster-001",
				ProviderAddress: "invalid-address",
				Name:            "test-cluster",
				State:           types.ClusterStateActive,
				TotalNodes:      10,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "invalid provider address",
		},
		{
			name: "invalid state",
			cluster: types.HPCCluster{
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Name:            "test-cluster",
				State:           types.ClusterState("invalid"),
				TotalNodes:      10,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "invalid cluster state",
		},
		{
			name: "missing region",
			cluster: types.HPCCluster{
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Name:            "test-cluster",
				State:           types.ClusterStateActive,
				TotalNodes:      10,
				Region:          "",
			},
			expectError: true,
			errorMsg:    "region cannot be empty",
		},
		{
			name: "zero total nodes",
			cluster: types.HPCCluster{
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Name:            "test-cluster",
				State:           types.ClusterStateActive,
				TotalNodes:      0,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "total_nodes must be at least 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cluster.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestOfferingValidation tests HPC offering validation
func TestOfferingValidation(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()

	testCases := []struct {
		name        string
		offering    types.HPCOffering
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid offering",
			offering: types.HPCOffering{
				OfferingID:                "offering-001",
				ClusterID:                 "cluster-001",
				ProviderAddress:           validAddr,
				Name:                      "standard-compute",
				Active:                    true,
				RequiredIdentityThreshold: 50,
				MaxRuntimeSeconds:         86400,
				SupportsCustomWorkloads:   true,
				Pricing: types.HPCPricing{
					BaseNodeHourPrice: "100",
					CPUCoreHourPrice:  "10",
					GPUHourPrice:      "1000",
					MemoryGBHourPrice: "5",
					StorageGBPrice:    "1",
					NetworkGBPrice:    "2",
					Currency:          "uvirt",
					MinimumCharge:     "100",
				},
				QueueOptions: []types.QueueOption{
					{
						PartitionName:   "default",
						DisplayName:     "Default Queue",
						MaxNodes:        10,
						MaxRuntime:      86400,
						PriceMultiplier: "1.0",
					},
				},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				BlockHeight: 100,
			},
			expectError: false,
		},
		{
			name: "missing offering_id",
			offering: types.HPCOffering{
				OfferingID:                "",
				ClusterID:                 "cluster-001",
				ProviderAddress:           validAddr,
				Name:                      "test",
				RequiredIdentityThreshold: 50,
				MaxRuntimeSeconds:         86400,
				QueueOptions: []types.QueueOption{
					{PartitionName: "default", DisplayName: "Default", MaxNodes: 10, MaxRuntime: 86400, PriceMultiplier: "1.0"},
				},
			},
			expectError: true,
			errorMsg:    "offering_id cannot be empty",
		},
		{
			name: "missing cluster_id",
			offering: types.HPCOffering{
				OfferingID:                "offering-001",
				ClusterID:                 "",
				ProviderAddress:           validAddr,
				Name:                      "test",
				RequiredIdentityThreshold: 50,
				MaxRuntimeSeconds:         86400,
				QueueOptions: []types.QueueOption{
					{PartitionName: "default", DisplayName: "Default", MaxNodes: 10, MaxRuntime: 86400, PriceMultiplier: "1.0"},
				},
			},
			expectError: true,
			errorMsg:    "cluster_id cannot be empty",
		},
		{
			name: "invalid identity threshold",
			offering: types.HPCOffering{
				OfferingID:                "offering-001",
				ClusterID:                 "cluster-001",
				ProviderAddress:           validAddr,
				Name:                      "test",
				RequiredIdentityThreshold: 150, // Invalid: > 100
				MaxRuntimeSeconds:         86400,
				QueueOptions: []types.QueueOption{
					{PartitionName: "default", DisplayName: "Default", MaxNodes: 10, MaxRuntime: 86400, PriceMultiplier: "1.0"},
				},
			},
			expectError: true,
			errorMsg:    "required_identity_threshold must be between 0 and 100",
		},
		{
			name: "max runtime too short",
			offering: types.HPCOffering{
				OfferingID:                "offering-001",
				ClusterID:                 "cluster-001",
				ProviderAddress:           validAddr,
				Name:                      "test",
				RequiredIdentityThreshold: 50,
				MaxRuntimeSeconds:         30, // Invalid: < 60
				QueueOptions: []types.QueueOption{
					{PartitionName: "default", DisplayName: "Default", MaxNodes: 10, MaxRuntime: 86400, PriceMultiplier: "1.0"},
				},
			},
			expectError: true,
			errorMsg:    "max_runtime_seconds must be at least 60",
		},
		{
			name: "no queue options",
			offering: types.HPCOffering{
				OfferingID:                "offering-001",
				ClusterID:                 "cluster-001",
				ProviderAddress:           validAddr,
				Name:                      "test",
				RequiredIdentityThreshold: 50,
				MaxRuntimeSeconds:         86400,
				QueueOptions:              []types.QueueOption{}, // Empty
			},
			expectError: true,
			errorMsg:    "at least one queue option is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.offering.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestJobStateValidation tests valid job states
func TestJobStateValidation(t *testing.T) {
	testCases := []struct {
		name  string
		state types.JobState
		valid bool
	}{
		{"pending valid", types.JobStatePending, true},
		{"queued valid", types.JobStateQueued, true},
		{"running valid", types.JobStateRunning, true},
		{"completed valid", types.JobStateCompleted, true},
		{"failed valid", types.JobStateFailed, true},
		{"cancelled valid", types.JobStateCancelled, true},
		{"timeout valid", types.JobStateTimeout, true},
		{"invalid state", types.JobState("invalid"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidJobState(tc.state)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestTerminalJobStates tests terminal job state detection
func TestTerminalJobStates(t *testing.T) {
	testCases := []struct {
		name     string
		state    types.JobState
		terminal bool
	}{
		{"pending not terminal", types.JobStatePending, false},
		{"queued not terminal", types.JobStateQueued, false},
		{"running not terminal", types.JobStateRunning, false},
		{"completed is terminal", types.JobStateCompleted, true},
		{"failed is terminal", types.JobStateFailed, true},
		{"cancelled is terminal", types.JobStateCancelled, true},
		{"timeout is terminal", types.JobStateTimeout, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsTerminalJobState(tc.state)
			require.Equal(t, tc.terminal, result)
		})
	}
}

// TestJobValidation tests HPC job validation
func TestJobValidation(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()

	testCases := []struct {
		name        string
		job         types.HPCJob
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid job",
			job: types.HPCJob{
				JobID:           "job-001",
				OfferingID:      "offering-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
				State:           types.JobStatePending,
				QueueName:       "default",
				WorkloadSpec: types.JobWorkloadSpec{
					ContainerImage: "ubuntu:latest",
					Command:        "echo hello",
					IsPreconfigured: false,
				},
				Resources: types.JobResources{
					Nodes:           1,
					CPUCoresPerNode: 4,
					MemoryGBPerNode: 8,
					StorageGB:       10,
				},
				MaxRuntimeSeconds: 3600,
				AgreedPrice:       sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(10000))),
				CreatedAt:         time.Now(),
				BlockHeight:       100,
			},
			expectError: false,
		},
		{
			name: "missing job_id",
			job: types.HPCJob{
				JobID:           "",
				OfferingID:      "offering-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
				State:           types.JobStatePending,
				QueueName:       "default",
				Resources: types.JobResources{
					Nodes: 1,
				},
				MaxRuntimeSeconds: 3600,
			},
			expectError: true,
			errorMsg:    "job_id cannot be empty",
		},
		{
			name: "invalid customer address",
			job: types.HPCJob{
				JobID:           "job-001",
				OfferingID:      "offering-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: "invalid-address",
				State:           types.JobStatePending,
				QueueName:       "default",
				Resources: types.JobResources{
					Nodes: 1,
				},
				MaxRuntimeSeconds: 3600,
			},
			expectError: true,
			errorMsg:    "invalid customer address",
		},
		{
			name: "zero nodes",
			job: types.HPCJob{
				JobID:           "job-001",
				OfferingID:      "offering-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
				State:           types.JobStatePending,
				QueueName:       "default",
				WorkloadSpec: types.JobWorkloadSpec{
					ContainerImage: "ubuntu:latest",
					Command:        "echo hello",
				},
				Resources: types.JobResources{
					Nodes: 0, // Invalid
				},
				MaxRuntimeSeconds: 3600,
			},
			expectError: true,
			errorMsg:    "nodes must be at least 1",
		},
		{
			name: "max runtime too short",
			job: types.HPCJob{
				JobID:           "job-001",
				OfferingID:      "offering-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				CustomerAddress: validAddr,
				State:           types.JobStatePending,
				QueueName:       "default",
				WorkloadSpec: types.JobWorkloadSpec{
					ContainerImage: "ubuntu:latest",
					Command:        "echo hello",
				},
				Resources: types.JobResources{
					Nodes: 1,
				},
				MaxRuntimeSeconds: 30, // Invalid: < 60
			},
			expectError: true,
			errorMsg:    "max_runtime_seconds must be at least 60",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.job.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestDefaultParams tests default genesis parameters
func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, "50000", params.PlatformFeeRate)       // 5% in fixed-point (50000/1000000)
	require.Equal(t, "800000", params.ProviderRewardRate)   // 80% in fixed-point (800000/1000000)
	require.Equal(t, "150000", params.NodeRewardRate)       // 15% in fixed-point (150000/1000000)
	require.Equal(t, "600000", params.LatencyWeightFactor)  // 0.6 weight
	require.Equal(t, "400000", params.CapacityWeightFactor) // 0.4 weight
	require.Equal(t, int64(50), params.MaxLatencyMs)
	require.Equal(t, int64(60), params.MinJobDurationSeconds)
	require.Equal(t, int64(604800), params.MaxJobDurationSeconds)
	require.Equal(t, int32(50), params.DefaultIdentityThreshold)
	require.Equal(t, "uvirt", params.DefaultDenom)
	require.True(t, params.EnableProximityClustering)
}

// TestDefaultGenesisState tests default genesis state
func TestDefaultGenesisState(t *testing.T) {
	gs := types.DefaultGenesisState()

	require.NotNil(t, gs)
	require.NotNil(t, gs.Params)
	require.Empty(t, gs.Clusters)
	require.Empty(t, gs.Offerings)
	require.Empty(t, gs.Jobs)
	require.Empty(t, gs.JobAccountings)
	require.Empty(t, gs.NodeMetadatas)
	require.Empty(t, gs.SchedulingDecisions)
	require.Empty(t, gs.HPCRewards)
	require.Empty(t, gs.Disputes)
	require.Equal(t, uint64(1), gs.ClusterSequence)
	require.Equal(t, uint64(1), gs.OfferingSequence)
	require.Equal(t, uint64(1), gs.JobSequence)
	require.Equal(t, uint64(1), gs.DecisionSequence)
	require.Equal(t, uint64(1), gs.DisputeSequence)
}

// TestClusterStateValidation tests valid cluster states
func TestClusterStateValidation(t *testing.T) {
	testCases := []struct {
		name  string
		state types.ClusterState
		valid bool
	}{
		{"pending valid", types.ClusterStatePending, true},
		{"active valid", types.ClusterStateActive, true},
		{"draining valid", types.ClusterStateDraining, true},
		{"offline valid", types.ClusterStateOffline, true},
		{"deregistered valid", types.ClusterStateDeregistered, true},
		{"invalid state", types.ClusterState("invalid"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidClusterState(tc.state)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestNodeMetadataValidation tests node metadata validation
func TestNodeMetadataValidation(t *testing.T) {
	validAddr := sdk.AccAddress(make([]byte, 20)).String()

	testCases := []struct {
		name        string
		node        types.NodeMetadata
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid node metadata",
			node: types.NodeMetadata{
				NodeID:          "node-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Region:          "us-west-2",
				Active:          true,
				State:           types.NodeStateActive,
				HealthStatus:    types.HealthStatusHealthy,
				Resources: types.NodeResources{
					CPUCores:  32,
					MemoryGB:  128,
					GPUs:      4,
					StorageGB: 1000,
				},
			},
			expectError: false,
		},
		{
			name: "missing node_id",
			node: types.NodeMetadata{
				NodeID:          "",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "node_id cannot be empty",
		},
		{
			name: "missing cluster_id",
			node: types.NodeMetadata{
				NodeID:          "node-001",
				ClusterID:       "",
				ProviderAddress: validAddr,
				Region:          "us-west-2",
			},
			expectError: true,
			errorMsg:    "cluster_id cannot be empty",
		},
		{
			name: "missing region",
			node: types.NodeMetadata{
				NodeID:          "node-001",
				ClusterID:       "cluster-001",
				ProviderAddress: validAddr,
				Region:          "",
			},
			expectError: true,
			errorMsg:    "region cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.node.Validate()
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
				DecisionID:        "decision-001",
				JobID:             "job-001",
				SelectedClusterID: "cluster-001",
				DecisionReason:    "Lowest latency and sufficient capacity",
				LatencyScore:      "850000",
				CapacityScore:     "750000",
				CombinedScore:     "810000",
				CreatedAt:         time.Now(),
				BlockHeight:       100,
			},
			expectError: false,
		},
		{
			name: "missing decision_id",
			decision: types.SchedulingDecision{
				DecisionID:        "",
				JobID:             "job-001",
				SelectedClusterID: "cluster-001",
				DecisionReason:    "Test",
			},
			expectError: true,
			errorMsg:    "decision_id cannot be empty",
		},
		{
			name: "missing job_id",
			decision: types.SchedulingDecision{
				DecisionID:        "decision-001",
				JobID:             "",
				SelectedClusterID: "cluster-001",
				DecisionReason:    "Test",
			},
			expectError: true,
			errorMsg:    "job_id cannot be empty",
		},
		{
			name: "missing decision_reason",
			decision: types.SchedulingDecision{
				DecisionID:        "decision-001",
				JobID:             "job-001",
				SelectedClusterID: "cluster-001",
				DecisionReason:    "",
			},
			expectError: true,
			errorMsg:    "decision_reason cannot be empty",
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

// TestHPCRewardSourceValidation tests reward source validation
func TestHPCRewardSourceValidation(t *testing.T) {
	testCases := []struct {
		name   string
		source types.HPCRewardSource
		valid  bool
	}{
		{"job_completion valid", types.HPCRewardSourceJobCompletion, true},
		{"usage valid", types.HPCRewardSourceUsage, true},
		{"bonus valid", types.HPCRewardSourceBonus, true},
		{"invalid source", types.HPCRewardSource("invalid"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidHPCRewardSource(tc.source)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestDisputeStatusValidation tests dispute status validation
func TestDisputeStatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status types.DisputeStatus
		valid  bool
	}{
		{"pending valid", types.DisputeStatusPending, true},
		{"under_review valid", types.DisputeStatusUnderReview, true},
		{"resolved valid", types.DisputeStatusResolved, true},
		{"rejected valid", types.DisputeStatusRejected, true},
		{"invalid status", types.DisputeStatus("invalid"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidDisputeStatus(tc.status)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestNodeStateValidation tests node state validation
func TestNodeStateValidation(t *testing.T) {
	testCases := []struct {
		name  string
		state types.NodeState
		valid bool
	}{
		{"unknown valid", types.NodeStateUnknown, true},
		{"pending valid", types.NodeStatePending, true},
		{"active valid", types.NodeStateActive, true},
		{"stale valid", types.NodeStateStale, true},
		{"draining valid", types.NodeStateDraining, true},
		{"drained valid", types.NodeStateDrained, true},
		{"offline valid", types.NodeStateOffline, true},
		{"deregistered valid", types.NodeStateDeregistered, true},
		{"invalid state", types.NodeState("invalid"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidNodeState(tc.state)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestNodeStateTransitions tests valid node state transitions
func TestNodeStateTransitions(t *testing.T) {
	testCases := []struct {
		name      string
		fromState types.NodeState
		toState   types.NodeState
		valid     bool
	}{
		{"pending to active", types.NodeStatePending, types.NodeStateActive, true},
		{"pending to offline", types.NodeStatePending, types.NodeStateOffline, true},
		{"active to stale", types.NodeStateActive, types.NodeStateStale, true},
		{"active to draining", types.NodeStateActive, types.NodeStateDraining, true},
		{"stale to active", types.NodeStateStale, types.NodeStateActive, true},
		{"draining to drained", types.NodeStateDraining, types.NodeStateDrained, true},
		{"offline to active", types.NodeStateOffline, types.NodeStateActive, true},
		{"deregistered to active", types.NodeStateDeregistered, types.NodeStateActive, false}, // Terminal
		{"active to pending", types.NodeStateActive, types.NodeStatePending, false},           // Invalid
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidNodeStateTransition(tc.fromState, tc.toState)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestTerminalNodeState tests terminal node state detection
func TestTerminalNodeState(t *testing.T) {
	testCases := []struct {
		name     string
		state    types.NodeState
		terminal bool
	}{
		{"pending not terminal", types.NodeStatePending, false},
		{"active not terminal", types.NodeStateActive, false},
		{"stale not terminal", types.NodeStateStale, false},
		{"draining not terminal", types.NodeStateDraining, false},
		{"offline not terminal", types.NodeStateOffline, false},
		{"deregistered is terminal", types.NodeStateDeregistered, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsTerminalNodeState(tc.state)
			require.Equal(t, tc.terminal, result)
		})
	}
}

// TestHealthStatusValidation tests health status validation
func TestHealthStatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status types.HealthStatus
		valid  bool
	}{
		{"healthy valid", types.HealthStatusHealthy, true},
		{"degraded valid", types.HealthStatusDegraded, true},
		{"unhealthy valid", types.HealthStatusUnhealthy, true},
		{"draining valid", types.HealthStatusDraining, true},
		{"offline valid", types.HealthStatusOffline, true},
		{"invalid status", types.HealthStatus("invalid"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidHealthStatus(tc.status)
			require.Equal(t, tc.valid, result)
		})
	}
}
