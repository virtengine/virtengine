package analysis

import (
	"math/big"

	"github.com/virtengine/virtengine/pkg/economics"
)

const (
	severityLow      = "low"
	severityMedium   = "medium"
	severityHigh     = "high"
	severityCritical = "critical"
)

// AttackAnalyzer analyzes attack costs and vulnerabilities.
type AttackAnalyzer struct {
	params economics.TokenomicsParams
}

// NewAttackAnalyzer creates a new attack analyzer.
func NewAttackAnalyzer(params economics.TokenomicsParams) *AttackAnalyzer {
	return &AttackAnalyzer{params: params}
}

// AttackScenario defines an attack scenario for analysis.
type AttackScenario struct {
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	TokenPriceUSD   float64 `json:"token_price_usd"`
	StakingRatioBPS int64   `json:"staking_ratio_bps"`
}

// Analyze51Attack analyzes the cost and feasibility of a 51% attack.
func (a *AttackAnalyzer) Analyze51Attack(
	state economics.NetworkState,
	validators []economics.ValidatorState,
	tokenPriceUSD float64,
) economics.AttackAnalysis {
	totalStaked := state.TotalStaked
	if totalStaked == nil || totalStaked.Sign() == 0 {
		totalStaked = big.NewInt(0)
	}

	// Calculate tokens needed for 51% attack
	// Attacker needs to control 51% of staked tokens
	tokensNeeded := new(big.Int).Mul(totalStaked, big.NewInt(51))
	tokensNeeded.Div(tokensNeeded, big.NewInt(49)) // Need 51/(100-51) = 51/49 more than existing

	// Convert to float for USD calculation
	tokensFloat := float64(tokensNeeded.Int64())
	costUSD := tokensFloat * tokenPriceUSD / 1000000 // Assuming 6 decimals

	// Calculate percentage of total supply
	supplyPercentage := float64(0)
	if state.TotalSupply != nil && state.TotalSupply.Sign() > 0 {
		pct := new(big.Int).Mul(tokensNeeded, big.NewInt(10000))
		pct.Div(pct, state.TotalSupply)
		supplyPercentage = float64(pct.Int64()) / 100
	}

	// Estimate time to prepare (based on market liquidity, simplified)
	timeToPrepare := a.estimateAcquisitionTime(tokensNeeded, state.TotalSupply)

	// Detection difficulty based on stake concentration
	detectionDifficulty := a.assessDetectionDifficulty(validators, tokensNeeded)

	// Risk level based on cost and other factors
	riskLevel := a.assess51AttackRisk(costUSD, supplyPercentage, int64(len(validators)))

	return economics.AttackAnalysis{
		AttackType:          "51_percent_attack",
		CostEstimateUSD:     costUSD,
		TokensRequired:      tokensNeeded,
		PercentageOfSupply:  supplyPercentage,
		TimeToPrepare:       timeToPrepare,
		DetectionDifficulty: detectionDifficulty,
		MitigationStrategy:  "Increase staking ratio, implement slashing for equivocation, use social consensus for major forks",
		RiskLevel:           riskLevel,
	}
}

// AnalyzeSpamAttack analyzes spam attack costs and impact.
func (a *AttackAnalyzer) AnalyzeSpamAttack(
	state economics.NetworkState,
	minGasPrice int64,
	targetTxPerSecond int64,
	durationSeconds int64,
	tokenPriceUSD float64,
) economics.AttackAnalysis {
	// Calculate total transactions needed
	totalTxs := targetTxPerSecond * durationSeconds

	// Calculate cost per transaction (minimum gas * min gas price)
	const minGasPerTx = 21000
	costPerTx := minGasPerTx * minGasPrice

	// Total tokens needed
	tokensNeeded := big.NewInt(costPerTx * totalTxs)

	// Convert to USD
	costUSD := float64(tokensNeeded.Int64()) * tokenPriceUSD / 1000000

	// Calculate percentage of supply
	supplyPercentage := float64(0)
	if state.TotalSupply != nil && state.TotalSupply.Sign() > 0 {
		pct := new(big.Int).Mul(tokensNeeded, big.NewInt(10000))
		pct.Div(pct, state.TotalSupply)
		supplyPercentage = float64(pct.Int64()) / 100
	}

	// Risk assessment
	riskLevel := a.assessSpamRisk(minGasPrice, targetTxPerSecond, costUSD)

	return economics.AttackAnalysis{
		AttackType:          "spam_attack",
		CostEstimateUSD:     costUSD,
		TokensRequired:      tokensNeeded,
		PercentageOfSupply:  supplyPercentage,
		TimeToPrepare:       "immediate",
		DetectionDifficulty: "low",
		MitigationStrategy:  "Increase minimum gas price, implement mempool prioritization, rate limiting per account",
		RiskLevel:           riskLevel,
	}
}

// AnalyzeLongRangeAttack analyzes long-range attack vulnerability.
func (a *AttackAnalyzer) AnalyzeLongRangeAttack(
	state economics.NetworkState,
	unbondingPeriodDays int64,
	tokenPriceUSD float64,
) economics.AttackAnalysis {
	// Long-range attack requires acquiring 2/3+ of stake at some historical point
	// and maintaining it through unbonding period

	tokensNeeded := new(big.Int).Mul(state.TotalStaked, big.NewInt(67))
	tokensNeeded.Div(tokensNeeded, big.NewInt(33))

	costUSD := float64(tokensNeeded.Int64()) * tokenPriceUSD / 1000000

	supplyPercentage := float64(0)
	if state.TotalSupply != nil && state.TotalSupply.Sign() > 0 {
		pct := new(big.Int).Mul(tokensNeeded, big.NewInt(10000))
		pct.Div(pct, state.TotalSupply)
		supplyPercentage = float64(pct.Int64()) / 100
	}

	// Time to prepare is at least unbonding period
	timeToPrepare := formatDays(unbondingPeriodDays) + " (minimum unbonding period)"

	// Risk level
	riskLevel := severityLow
	if unbondingPeriodDays < 7 {
		riskLevel = severityHigh
	} else if unbondingPeriodDays < 14 {
		riskLevel = severityMedium
	}

	return economics.AttackAnalysis{
		AttackType:          "long_range_attack",
		CostEstimateUSD:     costUSD,
		TokensRequired:      tokensNeeded,
		PercentageOfSupply:  supplyPercentage,
		TimeToPrepare:       timeToPrepare,
		DetectionDifficulty: "medium",
		MitigationStrategy:  "Implement weak subjectivity checkpoints, social consensus for deep reorgs, longer unbonding period",
		RiskLevel:           riskLevel,
	}
}

// AnalyzeCartellization analyzes validator cartel formation risk.
func (a *AttackAnalyzer) AnalyzeCartellization(
	validators []economics.ValidatorState,
	tokenPriceUSD float64,
) economics.AttackAnalysis {
	if len(validators) == 0 {
		return economics.AttackAnalysis{
			AttackType: "cartel_formation",
			RiskLevel:  "unknown",
		}
	}

	// Calculate minimum validators needed for cartel (33% for liveness, 67% for safety)
	totalStake := big.NewInt(0)
	for _, v := range validators {
		totalStake.Add(totalStake, v.TotalStake)
	}

	// Sort by stake
	sorted := make([]economics.ValidatorState, len(validators))
	copy(sorted, validators)
	sortValidatorsByStake(sorted)

	// Find minimum validators for 33% and 67%
	minFor33 := findMinForPercentage(sorted, totalStake, 33)
	minFor67 := findMinForPercentage(sorted, totalStake, 67)

	// Calculate stake needed for cartel
	cartelStake := new(big.Int).Mul(totalStake, big.NewInt(33))
	cartelStake.Div(cartelStake, big.NewInt(100))

	costUSD := float64(cartelStake.Int64()) * tokenPriceUSD / 1000000

	// Risk assessment
	var riskLevel string
	if minFor33 <= 3 {
		riskLevel = severityCritical
	} else if minFor33 <= 5 {
		riskLevel = severityHigh
	} else if minFor33 <= 10 {
		riskLevel = severityMedium
	} else {
		riskLevel = severityLow
	}

	return economics.AttackAnalysis{
		AttackType:          "cartel_formation",
		CostEstimateUSD:     costUSD,
		TokensRequired:      cartelStake,
		PercentageOfSupply:  33.0,
		TimeToPrepare:       "variable (coordination required)",
		DetectionDifficulty: "high",
		MitigationStrategy:  formatCartelMitigation(minFor33, minFor67),
		RiskLevel:           riskLevel,
	}
}

// AnalyzeNothingAtStake analyzes nothing-at-stake attack vulnerability.
func (a *AttackAnalyzer) AnalyzeNothingAtStake(
	slashingEnabled bool,
	slashingPercentBPS int64,
) economics.AttackAnalysis {
	riskLevel := severityLow
	mitigation := "Slashing for equivocation is enabled with " + formatBPS(slashingPercentBPS) + " penalty"

	if !slashingEnabled {
		riskLevel = severityCritical
		mitigation = "CRITICAL: Enable slashing for equivocation immediately"
	} else if slashingPercentBPS < 100 {
		riskLevel = severityHigh
		mitigation = "Increase slashing penalty from " + formatBPS(slashingPercentBPS) + " to at least 1%"
	} else if slashingPercentBPS < 500 {
		riskLevel = severityMedium
		mitigation = "Consider increasing slashing penalty from " + formatBPS(slashingPercentBPS) + " to 5%"
	}

	return economics.AttackAnalysis{
		AttackType:          "nothing_at_stake",
		CostEstimateUSD:     0, // Cost is risk of slashing
		TokensRequired:      big.NewInt(0),
		PercentageOfSupply:  0,
		TimeToPrepare:       "immediate",
		DetectionDifficulty: "low",
		MitigationStrategy:  mitigation,
		RiskLevel:           riskLevel,
	}
}

// ComprehensiveAttackAnalysis performs all attack analyses.
func (a *AttackAnalyzer) ComprehensiveAttackAnalysis(
	state economics.NetworkState,
	validators []economics.ValidatorState,
	tokenPriceUSD float64,
	slashingEnabled bool,
	slashingPercentBPS int64,
) []economics.AttackAnalysis {
	analyses := make([]economics.AttackAnalysis, 0, 5)

	// 51% attack
	analyses = append(analyses, a.Analyze51Attack(state, validators, tokenPriceUSD))

	// Spam attack (100 TPS for 1 hour)
	analyses = append(analyses, a.AnalyzeSpamAttack(
		state,
		a.params.MinGasPrice,
		100,
		3600,
		tokenPriceUSD,
	))

	// Long-range attack
	analyses = append(analyses, a.AnalyzeLongRangeAttack(
		state,
		a.params.UnbondingPeriodDays,
		tokenPriceUSD,
	))

	// Cartel formation
	analyses = append(analyses, a.AnalyzeCartellization(validators, tokenPriceUSD))

	// Nothing-at-stake
	analyses = append(analyses, a.AnalyzeNothingAtStake(slashingEnabled, slashingPercentBPS))

	return analyses
}

// GenerateSecurityRecommendations generates security recommendations from attack analyses.
func (a *AttackAnalyzer) GenerateSecurityRecommendations(
	analyses []economics.AttackAnalysis,
) []economics.Recommendation {
	var recommendations []economics.Recommendation

	for _, analysis := range analyses {
		if analysis.RiskLevel == severityCritical || analysis.RiskLevel == severityHigh {
			recommendations = append(recommendations, economics.Recommendation{
				Category:    "security",
				Priority:    analysis.RiskLevel,
				Title:       "Address " + analysis.AttackType + " Vulnerability",
				Description: "Attack cost: $" + formatFloat(analysis.CostEstimateUSD),
				Impact:      "Network security could be compromised",
				Action:      analysis.MitigationStrategy,
			})
		}
	}

	return recommendations
}

// Helper functions

func (a *AttackAnalyzer) estimateAcquisitionTime(tokensNeeded, totalSupply *big.Int) string {
	if totalSupply == nil || totalSupply.Sign() == 0 {
		return "unknown"
	}

	pct := new(big.Int).Mul(tokensNeeded, big.NewInt(100))
	pct.Div(pct, totalSupply)

	percentage := pct.Int64()

	if percentage > 50 {
		return "6+ months (likely impossible without market manipulation)"
	} else if percentage > 20 {
		return "3-6 months"
	} else if percentage > 10 {
		return "1-3 months"
	} else if percentage > 5 {
		return "2-4 weeks"
	}
	return "days to weeks"
}

func (a *AttackAnalyzer) assessDetectionDifficulty(validators []economics.ValidatorState, tokensNeeded *big.Int) string {
	if len(validators) == 0 {
		return "unknown"
	}

	// If attacker needs to acquire significant stake, it's easier to detect
	// through unusual staking patterns
	if tokensNeeded.Cmp(big.NewInt(1_000_000_000_000)) > 0 { // 1M tokens
		return "low (large stake movements are visible)"
	}
	return "medium"
}

//nolint:unparam // supplyPercentage kept for future validator concentration analysis
func (a *AttackAnalyzer) assess51AttackRisk(costUSD, _ float64, _ int64) string {
	// Very low cost = critical risk
	if costUSD < 1_000_000 {
		return severityCritical
	}
	if costUSD < 10_000_000 {
		return severityHigh
	}
	if costUSD < 100_000_000 {
		return severityMedium
	}
	return severityLow
}

//nolint:unparam // minGasPrice kept for future TPS-based cost scaling
func (a *AttackAnalyzer) assessSpamRisk(_, _ int64, costUSD float64) string {
	// If spam is cheap, risk is high
	costPerDay := costUSD * (86400 / 3600) // Scale to daily cost
	if costPerDay < 100 {
		return severityCritical
	}
	if costPerDay < 1000 {
		return severityHigh
	}
	if costPerDay < 10000 {
		return severityMedium
	}
	return severityLow
}

func sortValidatorsByStake(validators []economics.ValidatorState) {
	for i := 0; i < len(validators)-1; i++ {
		for j := i + 1; j < len(validators); j++ {
			if validators[i].TotalStake.Cmp(validators[j].TotalStake) < 0 {
				validators[i], validators[j] = validators[j], validators[i]
			}
		}
	}
}

func findMinForPercentage(sortedValidators []economics.ValidatorState, totalStake *big.Int, percentage int64) int64 {
	target := new(big.Int).Mul(totalStake, big.NewInt(percentage))
	target.Div(target, big.NewInt(100))

	cumulative := big.NewInt(0)
	for i, v := range sortedValidators {
		cumulative.Add(cumulative, v.TotalStake)
		if cumulative.Cmp(target) >= 0 {
			return int64(i + 1)
		}
	}
	return int64(len(sortedValidators))
}

func formatDays(days int64) string {
	return big.NewInt(days).String() + " days"
}

func formatBPS(bps int64) string {
	return big.NewInt(bps).String() + " BPS (" + formatFloat(float64(bps)/100) + "%)"
}

func formatFloat(f float64) string {
	return big.NewFloat(f).Text('f', 2)
}

func formatCartelMitigation(minFor33, minFor67 int64) string {
	return "Currently requires " + big.NewInt(minFor33).String() + " validators for 33% (liveness attack) and " +
		big.NewInt(minFor67).String() + " validators for 67% (safety attack). " +
		"Encourage stake distribution and implement validator caps."
}
