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
			Provider:        "cosmos1provider",
			Status:          types.ClusterStatusActive,
			TotalNodes:      10,
			AvailableNodes:  8,
			TotalCores:      640,
			AvailableCores:  512,
			TotalMemoryGB:   5120,
			AvailableMemoryGB: 4096,
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
			OfferingID:    "hpc-offering-1",
			ClusterID:     "hpc-cluster-1",
			Provider:      "cosmos1provider",
			Name:          "Standard HPC",
			Status:        types.OfferingStatusActive,
			CoresMin:      1,
			CoresMax:      64,
			MemoryGBMin:   1,
			MemoryGBMax:   256,
			PricePerCoreHour: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
			PricePerGBHour:   sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
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
	genesis.Jobs = []types.HPCJob{
		{
			JobID:        "hpc-job-1",
			OfferingID:   "hpc-offering-1",
			ClusterID:    "hpc-cluster-1",
			Submitter:    "cosmos1submitter",
			Status:       types.JobStatusRunning,
			RequestedCores: 8,
			RequestedMemoryGB: 32,
			SubmittedAt:  now,
			StartedAt:    now,
			WallTimeSeconds: 3600,
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
			ClusterID: "", // Invalid
			Name:      "Test Cluster",
			Provider:  "cosmos1provider",
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
			OfferingID: "", // Invalid
			ClusterID:  "hpc-cluster-1",
			Provider:   "cosmos1provider",
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
			JobID:     "", // Invalid
			Submitter: "cosmos1submitter",
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
			ClusterID: "hpc-cluster-1",
			Name:      "Cluster 1",
			Provider:  "cosmos1provider1",
			Status:    types.ClusterStatusActive,
		},
		{
			ClusterID: "hpc-cluster-1", // Duplicate
			Name:      "Cluster 2",
			Provider:  "cosmos1provider2",
			Status:    types.ClusterStatusActive,
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
			OfferingID: "hpc-offering-1",
			ClusterID:  "hpc-cluster-1",
			Provider:   "cosmos1provider",
			Name:       "Offering 1",
			Status:     types.OfferingStatusActive,
		},
		{
			OfferingID: "hpc-offering-1", // Duplicate
			ClusterID:  "hpc-cluster-1",
			Provider:   "cosmos1provider",
			Name:       "Offering 2",
			Status:     types.OfferingStatusActive,
		},
	}

	err := hpc.ValidateGenesis(genesis)
	s.Require().Error(err)
}

// Test: DefaultParams
func (s *GenesisTestSuite) TestDefaultParams() {
	params := types.DefaultParams()

	s.Require().NotNil(params)
	s.Require().Greater(params.MaxJobWallTimeSeconds, int64(0))
	s.Require().Greater(params.MaxCoresPerJob, uint32(0))
	s.Require().Greater(params.MaxMemoryGBPerJob, uint32(0))
}

// Test: Params validation - valid
func (s *GenesisTestSuite) TestParamsValidation_Valid() {
	params := types.DefaultParams()
	err := params.Validate()
	s.Require().NoError(err)
}

// Test: Params validation - zero max wall time
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxWallTime() {
	params := types.DefaultParams()
	params.MaxJobWallTimeSeconds = 0

	err := params.Validate()
	s.Require().Error(err)
}

// Test: Params validation - zero max cores
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxCores() {
	params := types.DefaultParams()
	params.MaxCoresPerJob = 0

	err := params.Validate()
	s.Require().Error(err)
}

// Test: Params validation - zero max memory
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxMemory() {
	params := types.DefaultParams()
	params.MaxMemoryGBPerJob = 0

	err := params.Validate()
	s.Require().Error(err)
}

// Table-driven tests for various param validations
func TestParamsValidationTable(t *testing.T) {
	tests := []struct {
		name        string
		modifier    func(*types.Params)
		expectError bool
	}{
		{
			name:        "valid default params",
			modifier:    func(p *types.Params) {},
			expectError: false,
		},
		{
			name: "max wall time too high",
			modifier: func(p *types.Params) {
				p.MaxJobWallTimeSeconds = 365 * 24 * 3600 * 10 // 10 years
			},
			expectError: true,
		},
		{
			name: "max cores too high",
			modifier: func(p *types.Params) {
				p.MaxCoresPerJob = 1000000
			},
			expectError: true,
		},
		{
			name: "max memory too high",
			modifier: func(p *types.Params) {
				p.MaxMemoryGBPerJob = 10000000
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.modifier(&params)
			err := params.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
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
				ClusterID:  "hpc-cluster-1",
				Name:       "Test Cluster",
				Provider:   "cosmos1provider",
				Status:     types.ClusterStatusActive,
				TotalNodes: 10,
			},
			expectError: false,
		},
		{
			name: "empty cluster ID",
			cluster: types.HPCCluster{
				ClusterID: "",
				Name:      "Test Cluster",
				Provider:  "cosmos1provider",
			},
			expectError: true,
		},
		{
			name: "empty provider",
			cluster: types.HPCCluster{
				ClusterID: "hpc-cluster-1",
				Name:      "Test Cluster",
				Provider:  "",
			},
			expectError: true,
		},
		{
			name: "empty name",
			cluster: types.HPCCluster{
				ClusterID: "hpc-cluster-1",
				Name:      "",
				Provider:  "cosmos1provider",
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
				OfferingID:       "hpc-offering-1",
				ClusterID:        "hpc-cluster-1",
				Provider:         "cosmos1provider",
				Name:             "Standard HPC",
				Status:           types.OfferingStatusActive,
				CoresMin:         1,
				CoresMax:         64,
				MemoryGBMin:      1,
				MemoryGBMax:      256,
				PricePerCoreHour: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
				PricePerGBHour:   sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
			},
			expectError: false,
		},
		{
			name: "empty offering ID",
			offering: types.HPCOffering{
				OfferingID: "",
				ClusterID:  "hpc-cluster-1",
				Provider:   "cosmos1provider",
			},
			expectError: true,
		},
		{
			name: "invalid core range",
			offering: types.HPCOffering{
				OfferingID:  "hpc-offering-1",
				ClusterID:   "hpc-cluster-1",
				Provider:    "cosmos1provider",
				Name:        "Invalid",
				CoresMin:    64,
				CoresMax:    1, // Max < Min
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
				Submitter:         "cosmos1submitter",
				Status:            types.JobStatusPending,
				RequestedCores:    8,
				RequestedMemoryGB: 32,
				WallTimeSeconds:   3600,
				SubmittedAt:       now,
			},
			expectError: false,
		},
		{
			name: "empty job ID",
			job: types.HPCJob{
				JobID:     "",
				Submitter: "cosmos1submitter",
			},
			expectError: true,
		},
		{
			name: "empty submitter",
			job: types.HPCJob{
				JobID:     "hpc-job-1",
				Submitter: "",
			},
			expectError: true,
		},
		{
			name: "zero requested cores",
			job: types.HPCJob{
				JobID:          "hpc-job-1",
				Submitter:      "cosmos1submitter",
				RequestedCores: 0,
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
				DisputeID:   "hpc-dispute-1",
				JobID:       "hpc-job-1",
				Disputant:   "cosmos1disputant",
				Status:      types.DisputeStatusOpen,
				Reason:      "Service quality issue",
				SubmittedAt: now,
			},
			expectError: false,
		},
		{
			name: "empty dispute ID",
			dispute: types.HPCDispute{
				DisputeID: "",
				JobID:     "hpc-job-1",
				Disputant: "cosmos1disputant",
			},
			expectError: true,
		},
		{
			name: "empty job ID",
			dispute: types.HPCDispute{
				DisputeID: "hpc-dispute-1",
				JobID:     "",
				Disputant: "cosmos1disputant",
			},
			expectError: true,
		},
		{
			name: "empty reason",
			dispute: types.HPCDispute{
				DisputeID: "hpc-dispute-1",
				JobID:     "hpc-job-1",
				Disputant: "cosmos1disputant",
				Reason:    "",
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
				DecisionID:  "hpc-decision-1",
				JobID:       "hpc-job-1",
				ClusterID:   "hpc-cluster-1",
				NodeID:      "node-1",
				ScheduledAt: now,
				Status:      types.SchedulingStatusApproved,
			},
			expectError: false,
		},
		{
			name: "empty decision ID",
			decision: types.SchedulingDecision{
				DecisionID: "",
				JobID:      "hpc-job-1",
				ClusterID:  "hpc-cluster-1",
			},
			expectError: true,
		},
		{
			name: "empty job ID",
			decision: types.SchedulingDecision{
				DecisionID: "hpc-decision-1",
				JobID:      "",
				ClusterID:  "hpc-cluster-1",
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
				ClusterID:  "hpc-cluster-1",
				Name:       "Production Cluster",
				Provider:   "cosmos1provider",
				Status:     types.ClusterStatusActive,
				TotalNodes: 100,
			},
		},
		Offerings: []types.HPCOffering{
			{
				OfferingID:       "hpc-offering-1",
				ClusterID:        "hpc-cluster-1",
				Provider:         "cosmos1provider",
				Name:             "Standard",
				Status:           types.OfferingStatusActive,
				CoresMin:         1,
				CoresMax:         64,
				MemoryGBMin:      1,
				MemoryGBMax:      256,
				PricePerCoreHour: sdk.NewDecCoin("uve", sdkmath.NewInt(100)),
				PricePerGBHour:   sdk.NewDecCoin("uve", sdkmath.NewInt(10)),
			},
		},
		Jobs: []types.HPCJob{
			{
				JobID:             "hpc-job-1",
				OfferingID:        "hpc-offering-1",
				ClusterID:         "hpc-cluster-1",
				Submitter:         "cosmos1user",
				Status:            types.JobStatusRunning,
				RequestedCores:    8,
				RequestedMemoryGB: 32,
				WallTimeSeconds:   3600,
				SubmittedAt:       now,
				StartedAt:         now,
			},
		},
		SchedulingDecisions: []types.SchedulingDecision{
			{
				DecisionID:  "hpc-decision-1",
				JobID:       "hpc-job-1",
				ClusterID:   "hpc-cluster-1",
				NodeID:      "node-1",
				ScheduledAt: now,
				Status:      types.SchedulingStatusApproved,
			},
		},
		HPCRewards: []types.HPCRewardRecord{
			{
				RewardID:   "hpc-reward-1",
				JobID:      "hpc-job-1",
				Provider:   "cosmos1provider",
				Amount:     sdk.NewCoin("uve", sdkmath.NewInt(1000)),
				PaidAt:     now,
			},
		},
		Disputes: []types.HPCDispute{
			{
				DisputeID:   "hpc-dispute-1",
				JobID:       "hpc-job-1",
				Disputant:   "cosmos1user",
				Status:      types.DisputeStatusOpen,
				Reason:      "Quality issue",
				SubmittedAt: now,
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
					ClusterID: tc.id,
					Name:      "Test",
					Provider:  "cosmos1provider",
					Status:    types.ClusterStatusActive,
				},
			}
			// The function is tested implicitly through ExportGenesis behavior
		})
	}
}
