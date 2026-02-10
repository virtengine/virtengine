package simulation

import (
	"math/big"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
)

func TestInflationSimulator_SimulateYear(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewInflationSimulator(params)

	state := economics.NetworkState{
		BlockHeight: 1000000,
		Timestamp:   time.Now(),
		TotalSupply: params.CurrentSupply,
		TotalStaked: params.TotalStaked,
	}

	result := sim.SimulateYear(state)

	// Verify result structure
	if result.Scenario != "annual_inflation" {
		t.Errorf("Expected scenario 'annual_inflation', got '%s'", result.Scenario)
	}

	if len(result.Snapshots) != 12 {
		t.Errorf("Expected 12 monthly snapshots, got %d", len(result.Snapshots))
	}

	// Supply should increase
	if result.FinalState.TotalSupply.Cmp(state.TotalSupply) <= 0 {
		t.Error("Expected supply to increase after one year of inflation")
	}

	// Verify metrics are calculated
	if result.Metrics.AvgInflationBPS == 0 {
		t.Error("Expected non-zero average inflation")
	}

	if result.Metrics.SupplyGrowthBPS <= 0 {
		t.Error("Expected positive supply growth")
	}
}

func TestInflationSimulator_AdjustInflation(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewInflationSimulator(params)

	testCases := []struct {
		name              string
		stakingRatioBPS   int64
		expectedInflation int64
		withinRange       bool
	}{
		{
			name:            "at target staking ratio",
			stakingRatioBPS: params.TargetStakingRatioBPS,
			withinRange:     true,
		},
		{
			name:            "below target - higher inflation",
			stakingRatioBPS: params.TargetStakingRatioBPS - 2000,
			withinRange:     true,
		},
		{
			name:            "above target - lower inflation",
			stakingRatioBPS: params.TargetStakingRatioBPS + 2000,
			withinRange:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inflation := sim.adjustInflation(tc.stakingRatioBPS)

			if inflation < params.MinInflationBPS || inflation > params.MaxInflationBPS {
				t.Errorf("Inflation %d should be within [%d, %d]",
					inflation, params.MinInflationBPS, params.MaxInflationBPS)
			}
		})
	}
}

func TestInflationSimulator_CalculateAPR(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewInflationSimulator(params)

	testCases := []struct {
		name            string
		inflationBPS    int64
		stakingRatioBPS int64
		expectPositive  bool
	}{
		{
			name:            "normal conditions",
			inflationBPS:    700,
			stakingRatioBPS: 6700,
			expectPositive:  true,
		},
		{
			name:            "low staking ratio - high APR",
			inflationBPS:    700,
			stakingRatioBPS: 3000,
			expectPositive:  true,
		},
		{
			name:            "zero staking ratio",
			inflationBPS:    700,
			stakingRatioBPS: 0,
			expectPositive:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apr := sim.calculateAPR(tc.inflationBPS, tc.stakingRatioBPS)

			if tc.expectPositive && apr <= 0 {
				t.Errorf("Expected positive APR, got %d", apr)
			}

			if !tc.expectPositive && apr != 0 {
				t.Errorf("Expected zero APR, got %d", apr)
			}
		})
	}
}

func TestInflationSimulator_AnalyzeInflationDynamics(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewInflationSimulator(params)

	state := economics.NetworkState{
		TotalSupply: params.CurrentSupply,
		TotalStaked: params.TotalStaked,
	}

	analysis := sim.AnalyzeInflationDynamics(state)

	// Verify analysis fields
	if analysis.CurrentRateBPS == 0 {
		t.Error("Expected non-zero current rate")
	}

	if analysis.SustainabilityScore < 0 || analysis.SustainabilityScore > 100 {
		t.Errorf("Sustainability score %d should be between 0 and 100", analysis.SustainabilityScore)
	}

	validTrends := map[string]bool{"inflationary": true, "deflationary": true, "stable": true}
	if !validTrends[analysis.Trend] {
		t.Errorf("Invalid trend: %s", analysis.Trend)
	}
}

func TestStakingSimulator_SimulateRewardDistribution(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewStakingSimulator(params)

	validators := []economics.ValidatorState{
		{
			Address:           "validator1",
			SelfStake:         big.NewInt(100_000_000_000),
			DelegatedStake:    big.NewInt(900_000_000_000),
			TotalStake:        big.NewInt(1_000_000_000_000),
			Commission:        1000, // 10%
			UptimeScore:       9500,
			VEIDVerifications: 50,
		},
		{
			Address:           "validator2",
			SelfStake:         big.NewInt(50_000_000_000),
			DelegatedStake:    big.NewInt(450_000_000_000),
			TotalStake:        big.NewInt(500_000_000_000),
			Commission:        500, // 5%
			UptimeScore:       8000,
			VEIDVerifications: 20,
		},
	}

	epochBlocks := int64(10000)
	results := sim.SimulateRewardDistribution(validators, epochBlocks)

	if len(results) != len(validators) {
		t.Errorf("Expected %d results, got %d", len(validators), len(results))
	}

	// Verify rewards are distributed proportionally
	for i, result := range results {
		if result.TotalReward.Sign() <= 0 {
			t.Errorf("Validator %d should have positive rewards", i)
		}

		if result.CommissionEarned.Sign() < 0 {
			t.Error("Commission should be non-negative")
		}

		if result.DelegatorRewards.Sign() < 0 {
			t.Error("Delegator rewards should be non-negative")
		}
	}

	// First validator has more stake, should have more rewards
	if results[0].TotalReward.Cmp(results[1].TotalReward) <= 0 {
		t.Error("Validator with more stake should earn more rewards")
	}
}

func TestStakingSimulator_AnalyzeStakingDynamics(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewStakingSimulator(params)

	state := economics.NetworkState{
		TotalSupply:     params.CurrentSupply,
		TotalStaked:     params.TotalStaked,
		TotalDelegators: 1000,
	}

	validators := []economics.ValidatorState{
		{Address: "v1", TotalStake: big.NewInt(100_000_000_000)},
		{Address: "v2", TotalStake: big.NewInt(80_000_000_000)},
		{Address: "v3", TotalStake: big.NewInt(60_000_000_000)},
	}

	analysis := sim.AnalyzeStakingDynamics(state, validators)

	if analysis.CurrentRatioBPS <= 0 {
		t.Error("Expected positive staking ratio")
	}

	if analysis.ValidatorCount != 3 {
		t.Errorf("Expected 3 validators, got %d", analysis.ValidatorCount)
	}

	validRisks := map[string]bool{"low": true, "moderate": true, "high": true, "insufficient_data": true}
	if !validRisks[analysis.ConcentrationRisk] {
		t.Errorf("Invalid concentration risk: %s", analysis.ConcentrationRisk)
	}
}

func TestFeeMarketSimulator_SimulateFeeMarket(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewFeeMarketSimulator(params)

	state := economics.NetworkState{
		BlockHeight: 1000,
	}

	// Generate sample transactions per block
	txsPerBlock := make([][]Transaction, 10)
	for i := range txsPerBlock {
		txsPerBlock[i] = GenerateSampleTransactions(100, 200, 0.05)
	}

	result := sim.SimulateFeeMarket(state, txsPerBlock, 400)

	if len(result.Snapshots) != 10 {
		t.Errorf("Expected 10 block snapshots, got %d", len(result.Snapshots))
	}

	if result.TotalFeesCollected.Sign() <= 0 {
		t.Error("Expected positive total fees")
	}

	if result.ProtocolRevenue.Sign() <= 0 {
		t.Error("Expected positive protocol revenue")
	}

	if result.ValidatorRevenue.Sign() <= 0 {
		t.Error("Expected positive validator revenue")
	}

	// Protocol + Validator should equal total
	sumRevenue := new(big.Int).Add(result.ProtocolRevenue, result.ValidatorRevenue)
	if sumRevenue.Cmp(result.TotalFeesCollected) != 0 {
		t.Error("Protocol + Validator revenue should equal total fees")
	}

	if result.SpamResistanceScore < 0 || result.SpamResistanceScore > 100 {
		t.Errorf("Spam resistance score %d should be between 0 and 100", result.SpamResistanceScore)
	}
}

func TestFeeMarketSimulator_AnalyzeFeeMarket(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	sim := NewFeeMarketSimulator(params)

	state := economics.NetworkState{
		TransactionVolume: big.NewInt(10000),
	}

	historicalFees := []int64{100, 150, 200, 180, 220, 190, 210, 200, 195, 205}

	analysis := sim.AnalyzeFeeMarket(state, historicalFees)

	if analysis.AverageFeeBPS <= 0 {
		t.Error("Expected positive average fee")
	}

	if analysis.MedianFeeBPS <= 0 {
		t.Error("Expected positive median fee")
	}

	if analysis.MarketEfficiency < 0 || analysis.MarketEfficiency > 1 {
		t.Errorf("Market efficiency %f should be between 0 and 1", analysis.MarketEfficiency)
	}
}

func TestGenerateSampleTransactions(t *testing.T) {
	txs := GenerateSampleTransactions(100, 200, 0.1)

	if len(txs) != 100 {
		t.Errorf("Expected 100 transactions, got %d", len(txs))
	}

	spamCount := 0
	for _, tx := range txs {
		if tx.IsSpam {
			spamCount++
		}
		if tx.GasUsed <= 0 {
			t.Error("Gas used should be positive")
		}
		if tx.GasPrice <= 0 {
			t.Error("Gas price should be positive")
		}
		if tx.FeePaid.Sign() <= 0 {
			t.Error("Fee paid should be positive")
		}
	}

	// Spam ratio should be approximately 10%
	spamRatio := float64(spamCount) / 100
	if spamRatio < 0.05 || spamRatio > 0.15 {
		t.Errorf("Spam ratio %f should be approximately 0.1", spamRatio)
	}
}
