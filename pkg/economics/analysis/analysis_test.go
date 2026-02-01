package analysis

import (
	"math/big"
	"testing"

	"github.com/virtengine/virtengine/pkg/economics"
)

func TestDistributionAnalyzer_AnalyzeDistribution(t *testing.T) {
	analyzer := NewDistributionAnalyzer()

	holdings := []Holding{
		{Address: "addr1", Balance: big.NewInt(1000000000), IsStaked: true},
		{Address: "addr2", Balance: big.NewInt(500000000), IsStaked: true},
		{Address: "addr3", Balance: big.NewInt(300000000), IsStaked: true},
		{Address: "addr4", Balance: big.NewInt(100000000), IsStaked: false},
		{Address: "addr5", Balance: big.NewInt(50000000), IsStaked: false},
		{Address: "addr6", Balance: big.NewInt(30000000), IsStaked: true},
		{Address: "addr7", Balance: big.NewInt(10000000), IsStaked: false},
		{Address: "addr8", Balance: big.NewInt(5000000), IsStaked: false},
		{Address: "addr9", Balance: big.NewInt(3000000), IsStaked: true},
		{Address: "addr10", Balance: big.NewInt(2000000), IsStaked: false},
	}

	totalSupply := big.NewInt(2000000000)

	metrics := analyzer.AnalyzeDistribution(holdings, totalSupply)

	// Gini coefficient should be between 0 and 1
	if metrics.GiniCoefficient < 0 || metrics.GiniCoefficient > 1 {
		t.Errorf("Gini coefficient %f should be between 0 and 1", metrics.GiniCoefficient)
	}

	// Nakamoto coefficient should be positive
	if metrics.NakamotoCoefficient <= 0 {
		t.Error("Nakamoto coefficient should be positive")
	}

	// Top 10 holdings should be <= 10000 BPS (100%)
	if metrics.Top10HoldingsBPS > 10000 {
		t.Errorf("Top 10 holdings %d should be <= 10000 BPS", metrics.Top10HoldingsBPS)
	}

	// HHI should be non-negative
	if metrics.HerfindahlIndex < 0 {
		t.Errorf("Herfindahl Index %f should be non-negative", metrics.HerfindahlIndex)
	}
}

func TestDistributionAnalyzer_CalculateGini_PerfectEquality(t *testing.T) {
	analyzer := NewDistributionAnalyzer()

	// Perfect equality - all same balance
	holdings := []Holding{
		{Address: "addr1", Balance: big.NewInt(100)},
		{Address: "addr2", Balance: big.NewInt(100)},
		{Address: "addr3", Balance: big.NewInt(100)},
		{Address: "addr4", Balance: big.NewInt(100)},
	}

	gini := analyzer.calculateGini(holdings)

	// Perfect equality should have Gini close to 0
	if gini > 0.1 {
		t.Errorf("Perfect equality should have Gini close to 0, got %f", gini)
	}
}

func TestDistributionAnalyzer_CalculateNakamoto(t *testing.T) {
	analyzer := NewDistributionAnalyzer()

	testCases := []struct {
		name            string
		holdings        []Holding
		totalSupply     *big.Int
		expectedNakamoto int64
	}{
		{
			name: "single dominant holder",
			holdings: []Holding{
				{Address: "addr1", Balance: big.NewInt(6000)},
				{Address: "addr2", Balance: big.NewInt(2000)},
				{Address: "addr3", Balance: big.NewInt(2000)},
			},
			totalSupply:     big.NewInt(10000),
			expectedNakamoto: 1, // One holder controls 60%
		},
		{
			name: "distributed holdings",
			holdings: []Holding{
				{Address: "addr1", Balance: big.NewInt(2000)},
				{Address: "addr2", Balance: big.NewInt(2000)},
				{Address: "addr3", Balance: big.NewInt(2000)},
				{Address: "addr4", Balance: big.NewInt(2000)},
				{Address: "addr5", Balance: big.NewInt(2000)},
			},
			totalSupply:     big.NewInt(10000),
			expectedNakamoto: 3, // Need 3 to reach 51%
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nakamoto := analyzer.calculateNakamotoCoefficient(tc.holdings, tc.totalSupply)

			if nakamoto != tc.expectedNakamoto {
				t.Errorf("Expected Nakamoto %d, got %d", tc.expectedNakamoto, nakamoto)
			}
		})
	}
}

func TestDistributionAnalyzer_CalculateDecentralizationScore(t *testing.T) {
	analyzer := NewDistributionAnalyzer()

	testCases := []struct {
		name        string
		metrics     economics.DistributionMetrics
		expectRange [2]int64 // min, max expected score
	}{
		{
			name: "well decentralized",
			metrics: economics.DistributionMetrics{
				GiniCoefficient:     0.3,
				NakamotoCoefficient: 50,
				Top10HoldingsBPS:    2000,
			},
			expectRange: [2]int64{70, 100},
		},
		{
			name: "highly centralized",
			metrics: economics.DistributionMetrics{
				GiniCoefficient:     0.9,
				NakamotoCoefficient: 3,
				Top10HoldingsBPS:    7000,
			},
			expectRange: [2]int64{0, 40},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := analyzer.CalculateDecentralizationScore(tc.metrics)

			if score < tc.expectRange[0] || score > tc.expectRange[1] {
				t.Errorf("Score %d should be in range [%d, %d]",
					score, tc.expectRange[0], tc.expectRange[1])
			}
		})
	}
}

func TestAttackAnalyzer_Analyze51Attack(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	analyzer := NewAttackAnalyzer(params)

	state := economics.NetworkState{
		TotalSupply: params.CurrentSupply,
		TotalStaked: params.TotalStaked,
	}

	validators := []economics.ValidatorState{
		{Address: "v1", TotalStake: big.NewInt(100_000_000_000)},
		{Address: "v2", TotalStake: big.NewInt(80_000_000_000)},
	}

	analysis := analyzer.Analyze51Attack(state, validators, 1.0) // $1 per token

	if analysis.AttackType != "51_percent_attack" {
		t.Errorf("Expected attack type '51_percent_attack', got '%s'", analysis.AttackType)
	}

	if analysis.TokensRequired.Sign() <= 0 {
		t.Error("Expected positive tokens required")
	}

	if analysis.CostEstimateUSD <= 0 {
		t.Error("Expected positive cost estimate")
	}

	if analysis.PercentageOfSupply <= 0 {
		t.Error("Expected positive percentage of supply")
	}

	validRisks := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	if !validRisks[analysis.RiskLevel] {
		t.Errorf("Invalid risk level: %s", analysis.RiskLevel)
	}
}

func TestAttackAnalyzer_AnalyzeSpamAttack(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	analyzer := NewAttackAnalyzer(params)

	state := economics.NetworkState{
		TotalSupply: params.CurrentSupply,
	}

	analysis := analyzer.AnalyzeSpamAttack(state, 100, 100, 3600, 1.0)

	if analysis.AttackType != "spam_attack" {
		t.Errorf("Expected attack type 'spam_attack', got '%s'", analysis.AttackType)
	}

	if analysis.CostEstimateUSD <= 0 {
		t.Error("Expected positive cost estimate for spam attack")
	}

	if analysis.TimeToPrepare != "immediate" {
		t.Errorf("Spam attack should be immediate, got '%s'", analysis.TimeToPrepare)
	}
}

func TestAttackAnalyzer_AnalyzeNothingAtStake(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	analyzer := NewAttackAnalyzer(params)

	testCases := []struct {
		name             string
		slashingEnabled  bool
		slashingBPS      int64
		expectedRisk     string
	}{
		{
			name:            "no slashing",
			slashingEnabled: false,
			slashingBPS:     0,
			expectedRisk:    "critical",
		},
		{
			name:            "low slashing",
			slashingEnabled: true,
			slashingBPS:     50,
			expectedRisk:    "high",
		},
		{
			name:            "adequate slashing",
			slashingEnabled: true,
			slashingBPS:     500,
			expectedRisk:    "low",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := analyzer.AnalyzeNothingAtStake(tc.slashingEnabled, tc.slashingBPS)

			if analysis.RiskLevel != tc.expectedRisk {
				t.Errorf("Expected risk level '%s', got '%s'", tc.expectedRisk, analysis.RiskLevel)
			}
		})
	}
}

func TestGameTheoryAnalyzer_AnalyzeValidatorIncentives(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	analyzer := NewGameTheoryAnalyzer(params)

	analysis := analyzer.AnalyzeValidatorIncentives(100, 1_000_000_000, 1000, 500)

	if analysis.Scenario != "validator_incentives" {
		t.Errorf("Expected scenario 'validator_incentives', got '%s'", analysis.Scenario)
	}

	if len(analysis.Players) == 0 {
		t.Error("Expected at least one player")
	}

	if len(analysis.Strategies) == 0 {
		t.Error("Expected strategies to be defined")
	}

	if analysis.NashEquilibrium == "" {
		t.Error("Expected Nash equilibrium to be identified")
	}

	validAlignments := map[string]bool{
		"strongly_aligned":   true,
		"partially_aligned":  true,
		"misaligned":         true,
		"aligned":            true,
		"partially_misaligned": true,
	}
	if !validAlignments[analysis.IncentiveAlignment] {
		t.Errorf("Invalid incentive alignment: %s", analysis.IncentiveAlignment)
	}
}

func TestGameTheoryAnalyzer_ComprehensiveAnalysis(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	analyzer := NewGameTheoryAnalyzer(params)

	analyses := analyzer.ComprehensiveGameTheoryAnalysis(
		100,   // validator count
		1_000_000_000, // avg stake
		1000,  // avg commission
		500,   // slashing penalty
		1000,  // delegator count
		100_000_000, // avg delegation
		1000,  // avg APR
		21,    // unbonding days
	)

	// Should have multiple analyses
	if len(analyses) < 4 {
		t.Errorf("Expected at least 4 analyses, got %d", len(analyses))
	}

	// Each analysis should have required fields
	for i, a := range analyses {
		if a.Scenario == "" {
			t.Errorf("Analysis %d missing scenario", i)
		}
		if len(a.Players) == 0 {
			t.Errorf("Analysis %d missing players", i)
		}
	}
}

