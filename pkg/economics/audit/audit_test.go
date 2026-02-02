package audit

import (
	"math/big"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
	"github.com/virtengine/virtengine/pkg/economics/analysis"
)

const severityCritical = "critical"

func TestEconomicAuditor_PerformAudit(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	auditor := NewEconomicAuditor(params)

	input := AuditInput{
		NetworkState: economics.NetworkState{
			BlockHeight:       1000000,
			Timestamp:         time.Now(),
			TotalSupply:       params.CurrentSupply,
			TotalStaked:       params.TotalStaked,
			TotalDelegators:   1000,
			TransactionVolume: big.NewInt(10000),
		},
		Validators: []economics.ValidatorState{
			{
				Address:        "val1",
				SelfStake:      big.NewInt(100_000_000_000),
				DelegatedStake: big.NewInt(900_000_000_000),
				TotalStake:     big.NewInt(1_000_000_000_000),
				Commission:     1000,
				UptimeScore:    9500,
			},
			{
				Address:        "val2",
				SelfStake:      big.NewInt(50_000_000_000),
				DelegatedStake: big.NewInt(450_000_000_000),
				TotalStake:     big.NewInt(500_000_000_000),
				Commission:     500,
				UptimeScore:    8000,
			},
		},
		Holdings: []analysis.Holding{
			{Address: "addr1", Balance: big.NewInt(100_000_000_000), IsStaked: true},
			{Address: "addr2", Balance: big.NewInt(50_000_000_000), IsStaked: true},
			{Address: "addr3", Balance: big.NewInt(30_000_000_000), IsStaked: false},
		},
		HistoricalFees:     []int64{100, 150, 200, 180, 220},
		TokenPriceUSD:      1.0,
		SlashingEnabled:    true,
		SlashingPenaltyBPS: 500,
	}

	audit := auditor.PerformAudit(input)

	// Verify audit has required fields
	if audit.Timestamp.IsZero() {
		t.Error("Audit should have timestamp")
	}

	if audit.OverallScore < 0 || audit.OverallScore > 100 {
		t.Errorf("Overall score %d should be between 0 and 100", audit.OverallScore)
	}

	// Verify inflation analysis
	if audit.InflationAnalysis.CurrentRateBPS == 0 {
		t.Error("Inflation analysis should have current rate")
	}

	// Verify staking analysis
	if audit.StakingAnalysis.ValidatorCount != 2 {
		t.Errorf("Expected 2 validators, got %d", audit.StakingAnalysis.ValidatorCount)
	}

	// Verify attack analyses exist
	if len(audit.AttackAnalyses) == 0 {
		t.Error("Should have attack analyses")
	}

	// Verify game theory analyses exist
	if len(audit.GameTheoryAnalyses) == 0 {
		t.Error("Should have game theory analyses")
	}
}

func TestEconomicAuditor_GenerateAuditReport(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	auditor := NewEconomicAuditor(params)

	// Create a mock audit
	mockAudit := economics.EconomicSecurityAudit{
		Timestamp:    time.Now(),
		OverallScore: 75,
		InflationAnalysis: economics.InflationAnalysis{
			CurrentRateBPS:      700,
			SustainabilityScore: 85,
			Trend:               "stable",
		},
		StakingAnalysis: economics.StakingAnalysis{
			CurrentRatioBPS:   6700,
			ValidatorCount:    100,
			ConcentrationRisk: "low",
		},
		FeeMarketAnalysis: economics.FeeMarketAnalysis{
			SpamResistance: 80,
		},
		DistributionMetrics: economics.DistributionMetrics{
			GiniCoefficient:     0.5,
			NakamotoCoefficient: 25,
		},
		AttackAnalyses: []economics.AttackAnalysis{
			{AttackType: "51_percent_attack", RiskLevel: "low", CostEstimateUSD: 100000000},
		},
		Vulnerabilities: []economics.Vulnerability{},
		Recommendations: []economics.Recommendation{},
	}

	report := auditor.GenerateAuditReport(mockAudit)

	if report.Title == "" {
		t.Error("Report should have title")
	}

	if report.OverallScore != mockAudit.OverallScore {
		t.Error("Report score should match audit score")
	}

	if report.Summary == "" {
		t.Error("Report should have summary")
	}

	if len(report.Sections) == 0 {
		t.Error("Report should have sections")
	}
}

func TestEconomicAuditor_IdentifyVulnerabilities(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	auditor := NewEconomicAuditor(params)

	// Create audit with known issues
	audit := economics.EconomicSecurityAudit{
		InflationAnalysis: economics.InflationAnalysis{
			CurrentRateBPS:      2500, // High inflation
			SustainabilityScore: 40,   // Low sustainability
		},
		StakingAnalysis: economics.StakingAnalysis{
			CurrentRatioBPS:   4000, // Low staking ratio
			ConcentrationRisk: "high",
		},
		FeeMarketAnalysis: economics.FeeMarketAnalysis{
			SpamResistance: 30, // Low spam resistance
		},
		DistributionMetrics: economics.DistributionMetrics{
			NakamotoCoefficient: 5, // Very low
			GiniCoefficient:     0.85,
		},
		AttackAnalyses: []economics.AttackAnalysis{
			{RiskLevel: severityCritical, AttackType: "51_attack", CostEstimateUSD: 100000},
		},
		GameTheoryAnalyses: []economics.GameTheoryAnalysis{
			{Scenario: "test", IncentiveAlignment: "misaligned", NashEquilibrium: "bad"},
		},
	}

	vulnerabilities := auditor.identifyVulnerabilities(audit)

	// Should identify multiple vulnerabilities
	if len(vulnerabilities) < 3 {
		t.Errorf("Expected at least 3 vulnerabilities, got %d", len(vulnerabilities))
	}

	// Should have critical vulnerabilities
	hasCritical := false
	for _, v := range vulnerabilities {
		if v.Severity == severityCritical {
			hasCritical = true
			break
		}
	}
	if !hasCritical {
		t.Error("Should identify critical vulnerabilities")
	}
}

func TestEconomicAuditor_CalculateOverallScore(t *testing.T) {
	params := economics.DefaultTokenomicsParams()
	auditor := NewEconomicAuditor(params)

	testCases := []struct {
		name        string
		audit       economics.EconomicSecurityAudit
		expectRange [2]int64
	}{
		{
			name: "healthy network",
			audit: economics.EconomicSecurityAudit{
				InflationAnalysis: economics.InflationAnalysis{
					SustainabilityScore: 90,
				},
				StakingAnalysis: economics.StakingAnalysis{
					CurrentRatioBPS: 6700,
				},
				FeeMarketAnalysis: economics.FeeMarketAnalysis{
					SpamResistance: 90,
				},
				DistributionMetrics: economics.DistributionMetrics{
					GiniCoefficient:     0.4,
					NakamotoCoefficient: 50,
				},
				AttackAnalyses: []economics.AttackAnalysis{
					{RiskLevel: "low"},
				},
			},
			expectRange: [2]int64{70, 100},
		},
		{
			name: "struggling network",
			audit: economics.EconomicSecurityAudit{
				InflationAnalysis: economics.InflationAnalysis{
					SustainabilityScore: 30,
				},
				StakingAnalysis: economics.StakingAnalysis{
					CurrentRatioBPS: 4000,
				},
				FeeMarketAnalysis: economics.FeeMarketAnalysis{
					SpamResistance: 40,
				},
				DistributionMetrics: economics.DistributionMetrics{
					GiniCoefficient:     0.9,
					NakamotoCoefficient: 5,
				},
				AttackAnalyses: []economics.AttackAnalysis{
					{RiskLevel: severityCritical},
					{RiskLevel: "high"},
				},
			},
			expectRange: [2]int64{0, 50},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := auditor.calculateOverallScore(tc.audit)

			if score < tc.expectRange[0] || score > tc.expectRange[1] {
				t.Errorf("Score %d should be in range [%d, %d]",
					score, tc.expectRange[0], tc.expectRange[1])
			}
		})
	}
}

func TestFormatFunctions(t *testing.T) {
	// Test formatBPS
	bps := formatBPS(6700)
	if bps == "" {
		t.Error("formatBPS should return non-empty string")
	}

	// Test formatFloat
	f := formatFloat(123.456)
	if f == "" {
		t.Error("formatFloat should return non-empty string")
	}

	// Test formatInt64
	i := formatInt64(12345)
	if i != "12345" {
		t.Errorf("Expected '12345', got '%s'", i)
	}

	// Test joinStrings
	joined := joinStrings([]string{"a", "b", "c"})
	if joined != "a; b; c" {
		t.Errorf("Expected 'a; b; c', got '%s'", joined)
	}
}
