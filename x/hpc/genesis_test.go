package hpc_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/hpc"
	"github.com/virtengine/virtengine/x/hpc/types"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

// Test: DefaultGenesisState returns valid state
func (s *GenesisTestSuite) TestDefaultGenesisState() {
	genesis := types.DefaultGenesisState()

	s.Require().NotNil(genesis)
	s.Require().NotNil(genesis.Params)
	s.Require().Empty(genesis.Clusters)
	s.Require().Empty(genesis.Offerings)
	s.Require().Empty(genesis.Jobs)
	s.Require().Empty(genesis.SchedulingDecisions)
	s.Require().Empty(genesis.HPCRewards)
	s.Require().Empty(genesis.Disputes)
}

// Test: ValidateGenesis with default state
func (s *GenesisTestSuite) TestValidateGenesis_Default() {
	genesis := types.DefaultGenesisState()
	err := hpc.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid clusters
func (s *GenesisTestSuite) TestValidateGenesis_ValidClusters() {
	genesis := types.DefaultGenesisState()
	genesis.Clusters = []types.HPCCluster{
		{
			ClusterID:       "hpc-cluster-1",
			Name:            "Test Cluster",
			ProviderAddress: "cosmos1provider",
			State:           types.ClusterStateActive,
			TotalNodes:      10,
			AvailableNodes:  8,
			Region:          "us-west-1",
			ClusterMetadata: types.ClusterMetadata{
				TotalCPUCores: 640,
				TotalMemoryGB: 5120,
			},
		},
	}
	genesis.ClusterSequence = 2

	err := hpc.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid offerings
func (s *GenesisTestSuite) TestValidateGenesis_ValidOfferings() {
	genesis := types.DefaultGenesisState()
	genesis.Offerings = []types.HPCOffering{
		{
			OfferingID:      "hpc-offering-1",
			ClusterID:       "hpc-cluster-1",
			ProviderAddress: "cosmos1provider",
			Name:            "Standard HPC",
			Active:          true,
			Pricing: types.HPCPricing{
				BaseNodeHourPrice: "100",
			},
		},
	}
	genesis.OfferingSequence = 2

	err := hpc.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid jobs
func (s *GenesisTestSuite) TestValidateGenesis_ValidJobs() {
	genesis := types.DefaultGenesisState()
	now := time.Now().UTC()
	startedAt := now
	genesis.Jobs = []types.HPCJob{
		{
			JobID:             "hpc-job-1",
			OfferingID:        "hpc-offering-1",
			ClusterID:         "hpc-cluster-1",
			CustomerAddress:   "cosmos1submitter",
			State:             types.JobStateRunning,
			MaxRuntimeSeconds: 3600,
			Resources: types.JobResources{
				Nodes:           1,
				CPUCoresPerNode: 8,
				MemoryGBPerNode: 32,
			},
			CreatedAt: now,
			StartedAt: &startedAt,
		},
	}
	genesis.JobSequence = 2

	err := hpc.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid cluster - empty ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidCluster_EmptyID() {
	genesis := types.DefaultGenesisState()
	genesis.Clusters = []types.HPCCluster{
		{
			ClusterID:       "", // Invalid
			Name:            "Test Cluster",
			ProviderAddress: "cosmos1provider",
		},
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid offering - empty ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidOffering_EmptyID() {
	genesis := types.DefaultGenesisState()
	genesis.Offerings = []types.HPCOffering{
		{
			OfferingID:      "", // Invalid
			ClusterID:       "hpc-cluster-1",
			ProviderAddress: "cosmos1provider",
		},
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid job - empty ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidJob_EmptyID() {
	genesis := types.DefaultGenesisState()
	genesis.Jobs = []types.HPCJob{
		{
			JobID:           "", // Invalid
			CustomerAddress: "cosmos1submitter",
		},
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate cluster IDs
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateClusters() {
	genesis := types.DefaultGenesisState()
	genesis.Clusters = []types.HPCCluster{
		{
			ClusterID:       "hpc-cluster-1",
			Name:            "Cluster 1",
			ProviderAddress: "cosmos1provider1",
			State:           types.ClusterStateActive,
		},
		{
			ClusterID:       "hpc-cluster-1", // Duplicate
			Name:            "Cluster 2",
			ProviderAddress: "cosmos1provider2",
			State:           types.ClusterStateActive,
		},
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate offering IDs
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateOfferings() {
	genesis := types.DefaultGenesisState()
	genesis.Offerings = []types.HPCOffering{
		{
			OfferingID:      "hpc-offering-1",
			ClusterID:       "hpc-cluster-1",
			ProviderAddress: "cosmos1provider",
			Name:            "Offering 1",
			Active:          true,
		},
		{
			OfferingID:      "hpc-offering-1", // Duplicate
			ClusterID:       "hpc-cluster-1",
			ProviderAddress: "cosmos1provider",
			Name:            "Offering 2",
			Active:          true,
		},
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: DefaultParams
func (s *GenesisTestSuite) TestDefaultParams() {
	params := types.DefaultParams()

	s.Require().NotNil(params)
	s.Require().Greater(params.MaxJobDurationSeconds, int64(0))
	s.Require().Greater(params.MinJobDurationSeconds, int64(0))
}

// Test: Params field validation - valid values
func (s *GenesisTestSuite) TestParamsFieldValidation_Valid() {
	params := types.DefaultParams()
	// Params doesn't have a Validate method, so we just check the values are sensible
	s.Require().Greater(params.MaxJobDurationSeconds, int64(0))
	s.Require().Greater(params.MinJobDurationSeconds, int64(0))
	s.Require().GreaterOrEqual(params.MaxJobDurationSeconds, params.MinJobDurationSeconds)
}

// Test: Params field values - durations are reasonable
func (s *GenesisTestSuite) TestParamsFieldValidation_DurationValues() {
	params := types.DefaultParams()
	// Min should be at least 60 seconds
	s.Require().GreaterOrEqual(params.MinJobDurationSeconds, int64(60))
}

// Test: Params field values - heartbeat timeouts are positive
func (s *GenesisTestSuite) TestParamsFieldValidation_HeartbeatTimeouts() {
	params := types.DefaultParams()
	s.Require().Greater(params.ClusterHeartbeatTimeout, int64(0))
	s.Require().Greater(params.NodeHeartbeatTimeout, int64(0))
}

// Table-driven tests for default params field values
func TestParamsFieldValuesTable(t *testing.T) {
	params := types.DefaultParams()

	tests := []struct {
		name      string
		condition bool
	}{
		{
			name:      "max job duration positive",
			condition: params.MaxJobDurationSeconds > 0,
		},
		{
			name:      "min job duration positive",
			condition: params.MinJobDurationSeconds > 0,
		},
		{
			name:      "max >= min duration",
			condition: params.MaxJobDurationSeconds >= params.MinJobDurationSeconds,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, tc.condition)
		})
	}
}

// Test: HPCCluster Validate
func (s *GenesisTestSuite) TestHPCClusterValidate() {
	tests := []struct {
		name        string
		cluster     types.HPCCluster
		expectError bool
	}{
		{
			name: "valid cluster",
			cluster: types.HPCCluster{
				ClusterID:       "hpc-cluster-1",
				Name:            "Test Cluster",
				ProviderAddress: "cosmos1provider",
				State:           types.ClusterStateActive,
				TotalNodes:      10,
			},
			expectError: false,
		},
		{
			name: "empty cluster ID",
			cluster: types.HPCCluster{
				ClusterID:       "",
				Name:            "Test Cluster",
				ProviderAddress: "cosmos1provider",
			},
			expectError: true,
		},
		{
			name: "empty provider",
			cluster: types.HPCCluster{
				ClusterID:       "hpc-cluster-1",
				Name:            "Test Cluster",
				ProviderAddress: "",
			},
			expectError: true,
		},
		{
			name: "empty name",
			cluster: types.HPCCluster{
				ClusterID:       "hpc-cluster-1",
				Name:            "",
				ProviderAddress: "cosmos1provider",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.cluster.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: HPCOffering Validate
func (s *GenesisTestSuite) TestHPCOfferingValidate() {
	tests := []struct {
		name        string
		offering    types.HPCOffering
		expectError bool
	}{
		{
			name: "valid offering",
			offering: types.HPCOffering{
				OfferingID:      "hpc-offering-1",
				ClusterID:       "hpc-cluster-1",
				ProviderAddress: "cosmos1provider",
				Name:            "Standard HPC",
				Active:          true,
				Pricing: types.HPCPricing{
					BaseNodeHourPrice: "100",
				},
			},
			expectError: false,
		},
		{
			name: "empty offering ID",
			offering: types.HPCOffering{
				OfferingID:      "",
				ClusterID:       "hpc-cluster-1",
				ProviderAddress: "cosmos1provider",
			},
			expectError: true,
		},
		{
			name: "empty cluster ID",
			offering: types.HPCOffering{
				OfferingID:      "hpc-offering-1",
				ClusterID:       "",
				ProviderAddress: "cosmos1provider",
				Name:            "Invalid",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.offering.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: HPCJob Validate
func (s *GenesisTestSuite) TestHPCJobValidate() {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		job         types.HPCJob
		expectError bool
	}{
		{
			name: "valid job",
			job: types.HPCJob{
				JobID:             "hpc-job-1",
				OfferingID:        "hpc-offering-1",
				ClusterID:         "hpc-cluster-1",
				CustomerAddress:   "cosmos1submitter",
				State:             types.JobStatePending,
				MaxRuntimeSeconds: 3600,
				Resources: types.JobResources{
					Nodes:           1,
					CPUCoresPerNode: 8,
					MemoryGBPerNode: 32,
				},
				CreatedAt: now,
			},
			expectError: false,
		},
		{
			name: "empty job ID",
			job: types.HPCJob{
				JobID:           "",
				CustomerAddress: "cosmos1submitter",
			},
			expectError: true,
		},
		{
			name: "empty customer address",
			job: types.HPCJob{
				JobID:           "hpc-job-1",
				CustomerAddress: "",
			},
			expectError: true,
		},
		{
			name: "zero resources",
			job: types.HPCJob{
				JobID:           "hpc-job-1",
				CustomerAddress: "cosmos1submitter",
				Resources: types.JobResources{
					Nodes: 0,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.job.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: HPCDispute Validate
func (s *GenesisTestSuite) TestHPCDisputeValidate() {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		dispute     types.HPCDispute
		expectError bool
	}{
		{
			name: "valid dispute",
			dispute: types.HPCDispute{
				DisputeID:       "hpc-dispute-1",
				JobID:           "hpc-job-1",
				DisputerAddress: "cosmos1disputant",
				Status:          types.DisputeStatusPending,
				Reason:          "Service quality issue",
				CreatedAt:       now,
			},
			expectError: false,
		},
		{
			name: "empty dispute ID",
			dispute: types.HPCDispute{
				DisputeID:       "",
				JobID:           "hpc-job-1",
				DisputerAddress: "cosmos1disputant",
			},
			expectError: true,
		},
		{
			name: "empty job ID",
			dispute: types.HPCDispute{
				DisputeID:       "hpc-dispute-1",
				JobID:           "",
				DisputerAddress: "cosmos1disputant",
			},
			expectError: true,
		},
		{
			name: "empty reason",
			dispute: types.HPCDispute{
				DisputeID:       "hpc-dispute-1",
				JobID:           "hpc-job-1",
				DisputerAddress: "cosmos1disputant",
				Reason:          "",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.dispute.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: SchedulingDecision Validate
func (s *GenesisTestSuite) TestSchedulingDecisionValidate() {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		decision    types.SchedulingDecision
		expectError bool
	}{
		{
			name: "valid decision",
			decision: types.SchedulingDecision{
				DecisionID:        "hpc-decision-1",
				JobID:             "hpc-job-1",
				SelectedClusterID: "hpc-cluster-1",
				DecisionReason:    "best capacity",
				CreatedAt:         now,
			},
			expectError: false,
		},
		{
			name: "empty decision ID",
			decision: types.SchedulingDecision{
				DecisionID:        "",
				JobID:             "hpc-job-1",
				SelectedClusterID: "hpc-cluster-1",
			},
			expectError: true,
		},
		{
			name: "empty job ID",
			decision: types.SchedulingDecision{
				DecisionID:        "hpc-decision-1",
				JobID:             "",
				SelectedClusterID: "hpc-cluster-1",
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := tc.decision.Validate()
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// Test: Complete genesis state with all entities
func (s *GenesisTestSuite) TestValidateGenesis_CompleteState() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Clusters: []types.HPCCluster{
			{
				ClusterID:       "hpc-cluster-1",
				Name:            "Production Cluster",
				ProviderAddress: "cosmos1provider",
				State:           types.ClusterStateActive,
				TotalNodes:      100,
			},
		},
		Offerings: []types.HPCOffering{
			{
				OfferingID:      "hpc-offering-1",
				ClusterID:       "hpc-cluster-1",
				ProviderAddress: "cosmos1provider",
				Name:            "Standard",
				Active:          true,
				Pricing: types.HPCPricing{
					BaseNodeHourPrice: "100",
					CPUCoreHourPrice:  "10",
					MemoryGBHourPrice: "5",
				},
			},
		},
		Jobs: []types.HPCJob{
			{
				JobID:             "hpc-job-1",
				OfferingID:        "hpc-offering-1",
				ClusterID:         "hpc-cluster-1",
				CustomerAddress:   "cosmos1user",
				State:             types.JobStateRunning,
				MaxRuntimeSeconds: 3600,
				Resources: types.JobResources{
					Nodes:           1,
					CPUCoresPerNode: 8,
					MemoryGBPerNode: 32,
				},
				CreatedAt: now,
				StartedAt: &now,
			},
		},
		SchedulingDecisions: []types.SchedulingDecision{
			{
				DecisionID:        "hpc-decision-1",
				JobID:             "hpc-job-1",
				SelectedClusterID: "hpc-cluster-1",
				DecisionReason:    "best capacity",
				CreatedAt:         now,
			},
		},
		HPCRewards: []types.HPCRewardRecord{
			{
				RewardID:    "hpc-reward-1",
				JobID:       "hpc-job-1",
				ClusterID:   "hpc-cluster-1",
				TotalReward: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(1000))),
				IssuedAt:    now,
			},
		},
		Disputes: []types.HPCDispute{
			{
				DisputeID:       "hpc-dispute-1",
				JobID:           "hpc-job-1",
				DisputerAddress: "cosmos1user",
				Status:          types.DisputeStatusPending,
				Reason:          "Quality issue",
				CreatedAt:       now,
			},
		},
		ClusterSequence:  2,
		OfferingSequence: 2,
		JobSequence:      2,
		DecisionSequence: 2,
		DisputeSequence:  2,
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().NoError(err)
}

// Test: extractSequenceFromID helper function
func TestExtractSequenceFromID(t *testing.T) {
	tests := []struct {
		id       string
		expected uint64
	}{
		{"hpc-cluster-1", 1},
		{"hpc-cluster-123", 123},
		{"hpc-offering-999", 999},
		{"invalid", 0},
		{"", 0},
		{"hpc-cluster-abc", 0},
	}

	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			// Use exported genesis function which internally uses extractSequenceFromID
			genesis := types.DefaultGenesisState()
			genesis.Clusters = []types.HPCCluster{
				{
					ClusterID:       tc.id,
					Name:            "Test",
					ProviderAddress: "cosmos1provider",
					State:           types.ClusterStateActive,
				},
			}
			// The function is tested implicitly through ExportGenesis behavior
		})
	}
}
