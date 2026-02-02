package audit

import (
	"math/big"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
	"github.com/virtengine/virtengine/pkg/economics/analysis"
	"github.com/virtengine/virtengine/pkg/economics/simulation"
)

const severityHigh = "high"

// EconomicAuditor performs comprehensive economic security audits.
type EconomicAuditor struct {
	params           economics.TokenomicsParams
	inflationSim     *simulation.InflationSimulator
	stakingSim       *simulation.StakingSimulator
	feeMarketSim     *simulation.FeeMarketSimulator
	distributionAnal *analysis.DistributionAnalyzer
	attackAnal       *analysis.AttackAnalyzer
	gameTheoryAnal   *analysis.GameTheoryAnalyzer
}

// NewEconomicAuditor creates a new economic auditor.
func NewEconomicAuditor(params economics.TokenomicsParams) *EconomicAuditor {
	return &EconomicAuditor{
		params:           params,
		inflationSim:     simulation.NewInflationSimulator(params),
		stakingSim:       simulation.NewStakingSimulator(params),
		feeMarketSim:     simulation.NewFeeMarketSimulator(params),
		distributionAnal: analysis.NewDistributionAnalyzer(),
		attackAnal:       analysis.NewAttackAnalyzer(params),
		gameTheoryAnal:   analysis.NewGameTheoryAnalyzer(params),
	}
}

// AuditInput contains all data needed for a comprehensive audit.
type AuditInput struct {
	NetworkState       economics.NetworkState     `json:"network_state"`
	Validators         []economics.ValidatorState `json:"validators"`
	Holdings           []analysis.Holding         `json:"holdings"`
	HistoricalFees     []int64                    `json:"historical_fees"`
	TokenPriceUSD      float64                    `json:"token_price_usd"`
	SlashingEnabled    bool                       `json:"slashing_enabled"`
	SlashingPenaltyBPS int64                      `json:"slashing_penalty_bps"`
}

// PerformAudit performs a comprehensive economic security audit.
func (a *EconomicAuditor) PerformAudit(input AuditInput) economics.EconomicSecurityAudit {
	audit := economics.EconomicSecurityAudit{
		Timestamp:       time.Now(),
		Vulnerabilities: make([]economics.Vulnerability, 0),
		Recommendations: make([]economics.Recommendation, 0),
	}

	// Inflation analysis
	audit.InflationAnalysis = a.inflationSim.AnalyzeInflationDynamics(input.NetworkState)

	// Staking analysis
	audit.StakingAnalysis = a.stakingSim.AnalyzeStakingDynamics(input.NetworkState, input.Validators)

	// Fee market analysis
	audit.FeeMarketAnalysis = a.feeMarketSim.AnalyzeFeeMarket(input.NetworkState, input.HistoricalFees)

	// Distribution analysis
	audit.DistributionMetrics = a.distributionAnal.AnalyzeDistribution(input.Holdings, input.NetworkState.TotalSupply)

	// Attack analyses
	audit.AttackAnalyses = a.attackAnal.ComprehensiveAttackAnalysis(
		input.NetworkState,
		input.Validators,
		input.TokenPriceUSD,
		input.SlashingEnabled,
		input.SlashingPenaltyBPS,
	)

	// Game theory analyses
	avgStake := a.calculateAvgStake(input.Validators)
	avgCommission := a.calculateAvgCommission(input.Validators)
	avgAPR := audit.StakingAnalysis.CurrentAPR

	audit.GameTheoryAnalyses = a.gameTheoryAnal.ComprehensiveGameTheoryAnalysis(
		int64(len(input.Validators)),
		avgStake,
		avgCommission,
		input.SlashingPenaltyBPS,
		input.NetworkState.TotalDelegators,
		a.calculateAvgDelegation(input.Validators),
		avgAPR,
		a.params.UnbondingPeriodDays,
	)

	// Identify vulnerabilities
	audit.Vulnerabilities = a.identifyVulnerabilities(audit)

	// Generate recommendations
	audit.Recommendations = a.generateRecommendations(audit)

	// Calculate overall score
	audit.OverallScore = a.calculateOverallScore(audit)

	return audit
}

// calculateAvgStake calculates average validator stake.
func (a *EconomicAuditor) calculateAvgStake(validators []economics.ValidatorState) int64 {
	if len(validators) == 0 {
		return 0
	}
	total := big.NewInt(0)
	for _, v := range validators {
		total.Add(total, v.TotalStake)
	}
	avg := new(big.Int).Div(total, big.NewInt(int64(len(validators))))
	return avg.Int64()
}

// calculateAvgCommission calculates average validator commission.
func (a *EconomicAuditor) calculateAvgCommission(validators []economics.ValidatorState) int64 {
	if len(validators) == 0 {
		return 0
	}
	total := int64(0)
	for _, v := range validators {
		total += v.Commission
	}
	return total / int64(len(validators))
}

// calculateAvgDelegation calculates average delegation.
func (a *EconomicAuditor) calculateAvgDelegation(validators []economics.ValidatorState) int64 {
	if len(validators) == 0 {
		return 0
	}
	total := big.NewInt(0)
	count := int64(0)
	for _, v := range validators {
		total.Add(total, v.DelegatedStake)
		count++
	}
	if count == 0 {
		return 0
	}
	avg := new(big.Int).Div(total, big.NewInt(count))
	return avg.Int64()
}

// identifyVulnerabilities identifies economic vulnerabilities.
func (a *EconomicAuditor) identifyVulnerabilities(audit economics.EconomicSecurityAudit) []economics.Vulnerability {
	vulnerabilities := make([]economics.Vulnerability, 0)
	vulnID := 1

	// Inflation vulnerabilities
	if audit.InflationAnalysis.CurrentRateBPS > 2000 {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    severityHigh,
			Category:    "inflation",
			Title:       "Excessive Inflation Rate",
			Description: "Current inflation rate exceeds 20%, risking token value dilution.",
			Impact:      "Token holders may experience significant value loss.",
			Mitigation:  "Reduce base reward per block or adjust inflation curve.",
			Status:      "open",
		})
		vulnID++
	}

	if audit.InflationAnalysis.SustainabilityScore < 50 {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    "medium",
			Category:    "inflation",
			Title:       "Unsustainable Inflation Model",
			Description: "Inflation model shows low sustainability score.",
			Impact:      "Long-term economic instability.",
			Mitigation:  "Review and adjust inflation parameters.",
			Status:      "open",
		})
		vulnID++
	}

	// Staking vulnerabilities
	if audit.StakingAnalysis.CurrentRatioBPS < 5000 {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    "critical",
			Category:    "staking",
			Title:       "Low Staking Ratio",
			Description: "Staking ratio below 50% significantly increases attack vulnerability.",
			Impact:      "Network security is compromised.",
			Mitigation:  "Increase staking incentives through higher APR.",
			Status:      "open",
		})
		vulnID++
	}

	if audit.StakingAnalysis.ConcentrationRisk == "high" {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    severityHigh,
			Category:    "staking",
			Title:       "High Stake Concentration",
			Description: "Stake is highly concentrated among few validators.",
			Impact:      "Centralization risk and potential for collusion.",
			Mitigation:  "Implement validator caps and delegation incentives.",
			Status:      "open",
		})
		vulnID++
	}

	// Fee market vulnerabilities
	if audit.FeeMarketAnalysis.SpamResistance < 50 {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    severityHigh,
			Category:    "fee_market",
			Title:       "Low Spam Resistance",
			Description: "Network is vulnerable to spam attacks.",
			Impact:      "Network congestion and degraded performance.",
			Mitigation:  "Increase minimum gas price.",
			Status:      "open",
		})
		vulnID++
	}

	// Distribution vulnerabilities
	if audit.DistributionMetrics.NakamotoCoefficient < 10 {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    "critical",
			Category:    "distribution",
			Title:       "Low Nakamoto Coefficient",
			Description: "Fewer than 10 entities could control the network.",
			Impact:      "High vulnerability to collusion attacks.",
			Mitigation:  "Implement validator caps and encourage delegation to smaller validators.",
			Status:      "open",
		})
		vulnID++
	}

	if audit.DistributionMetrics.GiniCoefficient > 0.8 {
		vulnerabilities = append(vulnerabilities, economics.Vulnerability{
			ID:          formatVulnID(vulnID),
			Severity:    severityHigh,
			Category:    "distribution",
			Title:       "Extreme Wealth Inequality",
			Description: "Gini coefficient indicates severe wealth concentration.",
			Impact:      "Governance capture risk and reduced decentralization.",
			Mitigation:  "Implement progressive staking rewards.",
			Status:      "open",
		})
		vulnID++
	}

	// Attack vulnerabilities
	for _, attack := range audit.AttackAnalyses {
		if attack.RiskLevel == "critical" {
			vulnerabilities = append(vulnerabilities, economics.Vulnerability{
				ID:          formatVulnID(vulnID),
				Severity:    "critical",
				Category:    "attack",
				Title:       attack.AttackType + " Vulnerability",
				Description: "Attack cost: $" + formatFloat(attack.CostEstimateUSD),
				Impact:      "Network could be compromised.",
				Mitigation:  attack.MitigationStrategy,
				Status:      "open",
			})
			vulnID++
		} else if attack.RiskLevel == severityHigh {
			vulnerabilities = append(vulnerabilities, economics.Vulnerability{
				ID:          formatVulnID(vulnID),
				Severity:    severityHigh,
				Category:    "attack",
				Title:       attack.AttackType + " Risk",
				Description: "Attack cost: $" + formatFloat(attack.CostEstimateUSD),
				Impact:      "Potential for network disruption.",
				Mitigation:  attack.MitigationStrategy,
				Status:      "open",
			})
			vulnID++
		}
	}

	// Game theory vulnerabilities
	for _, gt := range audit.GameTheoryAnalyses {
		if gt.IncentiveAlignment == "misaligned" {
			vulnerabilities = append(vulnerabilities, economics.Vulnerability{
				ID:          formatVulnID(vulnID),
				Severity:    "medium",
				Category:    "incentives",
				Title:       "Misaligned Incentives: " + gt.Scenario,
				Description: "Nash equilibrium: " + gt.NashEquilibrium,
				Impact:      "Actors may behave contrary to network interests.",
				Mitigation:  joinStrings(gt.Recommendations),
				Status:      "open",
			})
			vulnID++
		}
	}

	return vulnerabilities
}

// generateRecommendations generates prioritized recommendations.
func (a *EconomicAuditor) generateRecommendations(audit economics.EconomicSecurityAudit) []economics.Recommendation {
	recommendations := make([]economics.Recommendation, 0)

	// Add recommendations from simulations
	infResult := a.inflationSim.SimulateYear(economics.NetworkState{
		TotalSupply: a.params.CurrentSupply,
		TotalStaked: a.params.TotalStaked,
	})
	recommendations = append(recommendations, infResult.Recommendations...)

	// Add recommendations from attack analysis
	recommendations = append(recommendations, a.attackAnal.GenerateSecurityRecommendations(audit.AttackAnalyses)...)

	// Add distribution recommendations
	distReport := a.distributionAnal.GenerateDistributionReport(nil, nil)
	if len(distReport.Recommendations) > 0 {
		recommendations = append(recommendations, distReport.Recommendations...)
	}

	// Add custom recommendations based on overall audit
	if audit.OverallScore < 50 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "overall",
			Priority:    "critical",
			Title:       "Comprehensive Economic Review Required",
			Description: "Overall economic security score is below 50.",
			Impact:      "Multiple economic vulnerabilities exist.",
			Action:      "Conduct in-depth review and address critical vulnerabilities immediately.",
		})
	}

	// Sort by priority
	sortRecommendations(recommendations)

	return recommendations
}

// calculateOverallScore calculates overall economic security score (0-100).
func (a *EconomicAuditor) calculateOverallScore(audit economics.EconomicSecurityAudit) int64 {
	score := int64(100)

	// Inflation score (max -20)
	if audit.InflationAnalysis.SustainabilityScore < 100 {
		score -= (100 - audit.InflationAnalysis.SustainabilityScore) / 5
	}

	// Staking score (max -25)
	if audit.StakingAnalysis.CurrentRatioBPS < 6700 {
		diff := (6700 - audit.StakingAnalysis.CurrentRatioBPS) / 100
		if diff > 25 {
			diff = 25
		}
		score -= diff
	}

	// Fee market score (max -15)
	if audit.FeeMarketAnalysis.SpamResistance < 100 {
		score -= (100 - audit.FeeMarketAnalysis.SpamResistance) / 7
	}

	// Distribution score (max -20)
	decentralizationScore := a.distributionAnal.CalculateDecentralizationScore(audit.DistributionMetrics)
	score -= (100 - decentralizationScore) / 5

	// Attack vulnerability score (max -20)
	for _, attack := range audit.AttackAnalyses {
		switch attack.RiskLevel {
		case "critical":
			score -= 10
		case "high":
			score -= 5
		case "medium":
			score -= 2
		}
	}

	// Clamp to 0-100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// GenerateAuditReport generates a formatted audit report.
func (a *EconomicAuditor) GenerateAuditReport(audit economics.EconomicSecurityAudit) AuditReport {
	report := AuditReport{
		Title:        "VirtEngine Economic Security Audit",
		Timestamp:    audit.Timestamp,
		OverallScore: audit.OverallScore,
		Summary:      a.generateSummary(audit),
		Sections:     make([]AuditSection, 0),
	}

	// Inflation section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Inflation Analysis",
		Score:   audit.InflationAnalysis.SustainabilityScore,
		Content: formatInflationAnalysis(audit.InflationAnalysis),
	})

	// Staking section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Staking Analysis",
		Score:   a.calculateStakingScore(audit.StakingAnalysis),
		Content: formatStakingAnalysis(audit.StakingAnalysis),
	})

	// Fee Market section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Fee Market Analysis",
		Score:   audit.FeeMarketAnalysis.SpamResistance,
		Content: formatFeeMarketAnalysis(audit.FeeMarketAnalysis),
	})

	// Distribution section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Token Distribution Analysis",
		Score:   a.distributionAnal.CalculateDecentralizationScore(audit.DistributionMetrics),
		Content: formatDistributionAnalysis(audit.DistributionMetrics),
	})

	// Security section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Security Analysis",
		Score:   a.calculateSecurityScore(audit.AttackAnalyses),
		Content: formatSecurityAnalysis(audit.AttackAnalyses),
	})

	// Vulnerabilities section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Identified Vulnerabilities",
		Content: formatVulnerabilities(audit.Vulnerabilities),
	})

	// Recommendations section
	report.Sections = append(report.Sections, AuditSection{
		Title:   "Recommendations",
		Content: formatRecommendations(audit.Recommendations),
	})

	return report
}

// AuditReport is the formatted audit report.
type AuditReport struct {
	Title        string         `json:"title"`
	Timestamp    time.Time      `json:"timestamp"`
	OverallScore int64          `json:"overall_score"`
	Summary      string         `json:"summary"`
	Sections     []AuditSection `json:"sections"`
}

// AuditSection is a section of the audit report.
type AuditSection struct {
	Title   string `json:"title"`
	Score   int64  `json:"score,omitempty"`
	Content string `json:"content"`
}

// generateSummary generates executive summary.
func (a *EconomicAuditor) generateSummary(audit economics.EconomicSecurityAudit) string {
	riskLevel := "LOW"
	if audit.OverallScore < 50 {
		riskLevel = "CRITICAL"
	} else if audit.OverallScore < 70 {
		riskLevel = "HIGH"
	} else if audit.OverallScore < 85 {
		riskLevel = "MEDIUM"
	}

	criticalVulns := 0
	highVulns := 0
	for _, v := range audit.Vulnerabilities {
		if v.Severity == "critical" {
			criticalVulns++
		} else if v.Severity == severityHigh {
			highVulns++
		}
	}

	return "Economic Security Audit completed with overall score: " + formatInt64(audit.OverallScore) + "/100. " +
		"Risk Level: " + riskLevel + ". " +
		"Found " + formatInt64(int64(criticalVulns)) + " critical and " + formatInt64(int64(highVulns)) + " high severity vulnerabilities. " +
		"Inflation trend: " + audit.InflationAnalysis.Trend + ". " +
		"Staking ratio: " + formatBPS(audit.StakingAnalysis.CurrentRatioBPS) + ". " +
		"Nakamoto coefficient: " + formatInt64(audit.DistributionMetrics.NakamotoCoefficient) + "."
}

// calculateStakingScore calculates staking-specific score.
func (a *EconomicAuditor) calculateStakingScore(analysis economics.StakingAnalysis) int64 {
	score := int64(100)

	// Staking ratio score
	if analysis.CurrentRatioBPS < 6700 {
		score -= (6700 - analysis.CurrentRatioBPS) / 100
	}

	// Concentration risk
	switch analysis.ConcentrationRisk {
	case "high":
		score -= 20
	case "moderate":
		score -= 10
	}

	// Unbonding pressure
	if analysis.UnbondingPressure > 0.5 {
		score -= int64(analysis.UnbondingPressure * 20)
	}

	if score < 0 {
		score = 0
	}

	return score
}

// calculateSecurityScore calculates security-specific score.
func (a *EconomicAuditor) calculateSecurityScore(attacks []economics.AttackAnalysis) int64 {
	score := int64(100)

	for _, attack := range attacks {
		switch attack.RiskLevel {
		case "critical":
			score -= 25
		case "high":
			score -= 15
		case "medium":
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// Helper functions

func formatVulnID(id int) string {
	return "VULN-" + formatInt64(int64(id))
}

func formatFloat(f float64) string {
	return big.NewFloat(f).Text('f', 2)
}

func formatInt64(i int64) string {
	return big.NewInt(i).String()
}

func formatBPS(bps int64) string {
	return big.NewInt(bps/100).String() + "." + big.NewInt(bps%100).String() + "%"
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += "; "
		}
		result += s
	}
	return result
}

func sortRecommendations(recommendations []economics.Recommendation) {
	// Priority order: critical, high, medium, low
	priorityOrder := map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3}

	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			pi := priorityOrder[recommendations[i].Priority]
			pj := priorityOrder[recommendations[j].Priority]
			if pi > pj {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}
}

func formatInflationAnalysis(a economics.InflationAnalysis) string {
	return "Current Rate: " + formatBPS(a.CurrentRateBPS) + "\n" +
		"Projected Rate: " + formatBPS(a.ProjectedRateBPS) + "\n" +
		"Sustainability Score: " + formatInt64(a.SustainabilityScore) + "/100\n" +
		"Trend: " + a.Trend
}

func formatStakingAnalysis(a economics.StakingAnalysis) string {
	return "Current Staking Ratio: " + formatBPS(a.CurrentRatioBPS) + "\n" +
		"Optimal Staking Ratio: " + formatBPS(a.OptimalRatioBPS) + "\n" +
		"Current APR: " + formatBPS(a.CurrentAPR) + "\n" +
		"Validator Count: " + formatInt64(a.ValidatorCount) + "\n" +
		"Concentration Risk: " + a.ConcentrationRisk
}

func formatFeeMarketAnalysis(a economics.FeeMarketAnalysis) string {
	return "Average Fee: " + formatInt64(a.AverageFeeBPS) + "\n" +
		"Median Fee: " + formatInt64(a.MedianFeeBPS) + "\n" +
		"Fee Volatility: " + formatFloat(a.FeeVolatility) + "\n" +
		"Market Efficiency: " + formatFloat(a.MarketEfficiency) + "\n" +
		"Spam Resistance: " + formatInt64(a.SpamResistance) + "/100"
}

func formatDistributionAnalysis(a economics.DistributionMetrics) string {
	return "Gini Coefficient: " + formatFloat(a.GiniCoefficient) + "\n" +
		"Nakamoto Coefficient: " + formatInt64(a.NakamotoCoefficient) + "\n" +
		"Top 10 Holdings: " + formatBPS(a.Top10HoldingsBPS) + "\n" +
		"Top 100 Holdings: " + formatBPS(a.Top100HoldingsBPS) + "\n" +
		"Herfindahl Index: " + formatFloat(a.HerfindahlIndex)
}

func formatSecurityAnalysis(attacks []economics.AttackAnalysis) string {
	result := ""
	for _, a := range attacks {
		result += a.AttackType + ": Risk=" + a.RiskLevel + ", Cost=$" + formatFloat(a.CostEstimateUSD) + "\n"
	}
	return result
}

func formatVulnerabilities(vulns []economics.Vulnerability) string {
	result := ""
	for _, v := range vulns {
		result += "[" + v.Severity + "] " + v.Title + ": " + v.Description + "\n"
	}
	if result == "" {
		result = "No vulnerabilities identified."
	}
	return result
}

func formatRecommendations(recs []economics.Recommendation) string {
	result := ""
	for _, r := range recs {
		result += "[" + r.Priority + "] " + r.Title + ": " + r.Action + "\n"
	}
	if result == "" {
		result = "No recommendations at this time."
	}
	return result
}
