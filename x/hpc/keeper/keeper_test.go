package keeper_test

import (
	"testing"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
)

// KeeperTestSuite defines the test suite for HPC keeper
type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	keeper        *keeper.Keeper
	cdc           codec.Codec
	storeService  store.KVStoreService
}

// SetupTest initializes the test suite
func (s *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Initialize keeper with mock dependencies
	// Note: In production tests, use proper mock store service
	s.keeper = keeper.NewKeeper(s.cdc, nil, nil, nil)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestClusterValidation tests HPC cluster validation
func TestClusterValidation(t *testing.T) {
	testCases := []struct {
		name        string
		cluster     types.HPCCluster
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid cluster",
			cluster: types.HPCCluster{
				ID:              1,
				Owner:           "cosmos1abc123",
				ProviderAddress: "provider1",
				Name:            "test-cluster",
				Status:          types.ClusterStatusActive,
				TotalCPUCores:   100,
				TotalMemoryMB:   1024000,
				TotalGPUs:       8,
				Partitions: []types.Partition{
					{
						Name:         "default",
						TotalNodes:   10,
						AvailableNodes: 10,
						MaxTimeMinutes: 1440,
						GPUsPerNode:  1,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing owner",
			cluster: types.HPCCluster{
				ID:              1,
				Owner:           "",
				ProviderAddress: "provider1",
				Name:            "test-cluster",
				Status:          types.ClusterStatusActive,
			},
			expectError: true,
			errorMsg:    "owner is required",
		},
		{
			name: "missing name",
			cluster: types.HPCCluster{
				ID:              1,
				Owner:           "cosmos1abc123",
				ProviderAddress: "provider1",
				Name:            "",
				Status:          types.ClusterStatusActive,
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "invalid status",
			cluster: types.HPCCluster{
				ID:              1,
				Owner:           "cosmos1abc123",
				ProviderAddress: "provider1",
				Name:            "test-cluster",
				Status:          999,
			},
			expectError: true,
			errorMsg:    "invalid status",
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
	testCases := []struct {
		name        string
		offering    types.HPCOffering
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid offering",
			offering: types.HPCOffering{
				ID:         1,
				ClusterID:  1,
				Owner:      "cosmos1abc123",
				Name:       "standard-compute",
				IsActive:   true,
				MinCPUCores: 1,
				MaxCPUCores: 64,
				MinMemoryMB: 1024,
				MaxMemoryMB: 256000,
				Pricing: types.HPCPricing{
					PricePerCPUHour:    sdk.NewInt(100),
					PricePerGBHour:     sdk.NewInt(10),
					PricePerGPUHour:    sdk.NewInt(1000),
					MinimumChargeHours: 1,
				},
				QueueOptions: []types.QueueOption{
					{
						Name:           "standard",
						Priority:       50,
						MaxWallTimeMin: 1440,
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing cluster ID",
			offering: types.HPCOffering{
				ID:         1,
				ClusterID:  0,
				Owner:      "cosmos1abc123",
				Name:       "test",
				MinCPUCores: 1,
				MaxCPUCores: 64,
			},
			expectError: true,
			errorMsg:    "cluster ID is required",
		},
		{
			name: "invalid CPU range",
			offering: types.HPCOffering{
				ID:         1,
				ClusterID:  1,
				Owner:      "cosmos1abc123",
				Name:       "test",
				MinCPUCores: 64,
				MaxCPUCores: 1,
			},
			expectError: true,
			errorMsg:    "min CPU cores cannot exceed max",
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

// TestJobStateTransitions tests valid state transitions for HPC jobs
func TestJobStateTransitions(t *testing.T) {
	testCases := []struct {
		name        string
		fromState   types.JobState
		toState     types.JobState
		valid       bool
	}{
		{"pending to queued", types.JobStatePending, types.JobStateQueued, true},
		{"queued to running", types.JobStateQueued, types.JobStateRunning, true},
		{"running to completed", types.JobStateRunning, types.JobStateCompleted, true},
		{"running to failed", types.JobStateRunning, types.JobStateFailed, true},
		{"pending to cancelled", types.JobStatePending, types.JobStateCancelled, true},
		{"queued to cancelled", types.JobStateQueued, types.JobStateCancelled, true},
		{"running to cancelled", types.JobStateRunning, types.JobStateCancelled, true},
		{"completed to running", types.JobStateCompleted, types.JobStateRunning, false},
		{"failed to running", types.JobStateFailed, types.JobStateRunning, false},
		{"cancelled to running", types.JobStateCancelled, types.JobStateRunning, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidStateTransition(tc.fromState, tc.toState)
			require.Equal(t, tc.valid, result)
		})
	}
}

// TestJobValidation tests HPC job validation
func TestJobValidation(t *testing.T) {
	testCases := []struct {
		name        string
		job         types.HPCJob
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid job",
			job: types.HPCJob{
				ID:         1,
				ClusterID:  1,
				OfferingID: 1,
				Owner:      "cosmos1abc123",
				State:      types.JobStatePending,
				Workload: types.JobWorkloadSpec{
					Type:       "batch",
					Command:    "echo hello",
					Partition:  "default",
					NumTasks:   1,
					CPUsPerTask: 1,
				},
				Resources: types.JobResources{
					CPUCores:   4,
					MemoryMB:   8192,
					GPUs:       0,
					MaxTimeMin: 60,
				},
				MaxBudget: sdk.NewInt(10000),
			},
			expectError: false,
		},
		{
			name: "missing owner",
			job: types.HPCJob{
				ID:         1,
				ClusterID:  1,
				OfferingID: 1,
				Owner:      "",
				State:      types.JobStatePending,
			},
			expectError: true,
			errorMsg:    "owner is required",
		},
		{
			name: "zero resources",
			job: types.HPCJob{
				ID:         1,
				ClusterID:  1,
				OfferingID: 1,
				Owner:      "cosmos1abc123",
				State:      types.JobStatePending,
				Resources: types.JobResources{
					CPUCores: 0,
				},
			},
			expectError: true,
			errorMsg:    "CPU cores must be positive",
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

// TestGenesisParams tests default genesis parameters
func TestGenesisParams(t *testing.T) {
	params := types.DefaultParams()

	require.Equal(t, int64(50000), params.PlatformFeeRate)   // 5% in fixed-point
	require.Equal(t, int64(700000), params.ProviderRewardRate) // 70% in fixed-point
	require.Equal(t, int64(250000), params.NodeRewardRate)    // 25% in fixed-point
	require.Equal(t, int64(500000), params.LatencyWeightFactor) // 50% weight
	require.Equal(t, int64(500000), params.CapacityWeightFactor) // 50% weight
	require.Equal(t, uint64(100), params.MaxLatencyMs)
	require.Equal(t, uint64(10), params.MinClusterNodes)

	// Verify rates sum to 100%
	totalRate := params.PlatformFeeRate + params.ProviderRewardRate + params.NodeRewardRate
	require.Equal(t, int64(1000000), totalRate, "rates should sum to 100%")
}

// TestParamsValidation tests parameter validation
func TestParamsValidation(t *testing.T) {
	testCases := []struct {
		name        string
		params      types.HPCParams
		expectError bool
		errorMsg    string
	}{
		{
			name:        "default params valid",
			params:      types.DefaultParams(),
			expectError: false,
		},
		{
			name: "rates exceed 100%",
			params: types.HPCParams{
				PlatformFeeRate:    500000,
				ProviderRewardRate: 500000,
				NodeRewardRate:     500000, // Total 150%
			},
			expectError: true,
			errorMsg:    "rates must sum to 100%",
		},
		{
			name: "negative platform fee",
			params: types.HPCParams{
				PlatformFeeRate:    -1,
				ProviderRewardRate: 700000,
				NodeRewardRate:     300001,
			},
			expectError: true,
			errorMsg:    "platform fee rate must be non-negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
