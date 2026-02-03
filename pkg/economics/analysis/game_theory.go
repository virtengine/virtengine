package analysis

import (
	"math/big"

	"github.com/virtengine/virtengine/pkg/economics"
)

const (
	alignmentAligned            = "aligned"
	alignmentMisaligned         = "misaligned"
	alignmentUnknown            = "unknown"
	strategyHonestParticipation = "honest_participation"
)

// GameTheoryAnalyzer analyzes game-theoretic properties of the economic system.
type GameTheoryAnalyzer struct {
	params economics.TokenomicsParams
}

// NewGameTheoryAnalyzer creates a new game theory analyzer.
func NewGameTheoryAnalyzer(params economics.TokenomicsParams) *GameTheoryAnalyzer {
	return &GameTheoryAnalyzer{params: params}
}

// Player represents an economic actor in the game.
type Player struct {
	Type     string  // "validator", "delegator", "user", "attacker"
	Strategy string  // Current strategy
	Payoff   float64 // Expected payoff
}

// AnalyzeValidatorIncentives analyzes validator incentive alignment.
func (g *GameTheoryAnalyzer) AnalyzeValidatorIncentives(
	validatorCount int64,
	avgStake int64,
	avgCommission int64,
	slashingPenaltyBPS int64,
) economics.GameTheoryAnalysis {
	strategies := map[string][]string{
		"validator": {
			strategyHonestParticipation,
			"double_signing",
			"selective_censorship",
			"collusion",
			"free_riding",
		},
	}

	// Calculate payoffs for each strategy (simplified model)
	payoffs := g.calculateValidatorPayoffs(avgStake, avgCommission, slashingPenaltyBPS)

	// Determine Nash equilibrium
	nashEquilibrium := g.findNashEquilibrium(payoffs)

	// Determine dominant strategy
	dominantStrategy := g.findDominantStrategy(payoffs)

	// Assess incentive alignment
	alignment := g.assessIncentiveAlignment(nashEquilibrium, dominantStrategy)

	return economics.GameTheoryAnalysis{
		Scenario:           "validator_incentives",
		Players:            []string{"validator"},
		Strategies:         strategies,
		NashEquilibrium:    nashEquilibrium,
		PayoffMatrix:       payoffs,
		DominantStrategy:   dominantStrategy,
		IncentiveAlignment: alignment,
		Recommendations:    g.generateValidatorRecommendations(nashEquilibrium, alignment),
	}
}

// AnalyzeDelegatorIncentives analyzes delegator incentive alignment.
func (g *GameTheoryAnalyzer) AnalyzeDelegatorIncentives(
	delegatorCount int64,
	avgDelegation int64,
	avgAPR int64,
	unbondingDays int64,
) economics.GameTheoryAnalysis {
	strategies := map[string][]string{
		"delegator": {
			"delegate_to_top_validator",
			"delegate_to_small_validator",
			"split_delegation",
			"redelegation_chasing",
			"hold_liquid",
		},
	}

	// Calculate payoffs
	payoffs := g.calculateDelegatorPayoffs(avgDelegation, avgAPR, unbondingDays)

	nashEquilibrium := g.findDelegatorNashEquilibrium(payoffs)
	dominantStrategy := g.findDelegatorDominantStrategy(payoffs)
	alignment := g.assessDelegatorAlignment(nashEquilibrium)

	return economics.GameTheoryAnalysis{
		Scenario:           "delegator_incentives",
		Players:            []string{"delegator"},
		Strategies:         strategies,
		NashEquilibrium:    nashEquilibrium,
		PayoffMatrix:       payoffs,
		DominantStrategy:   dominantStrategy,
		IncentiveAlignment: alignment,
		Recommendations:    g.generateDelegatorRecommendations(alignment),
	}
}

// AnalyzeVEIDVerifierIncentives analyzes VEID verifier incentive alignment.
func (g *GameTheoryAnalyzer) AnalyzeVEIDVerifierIncentives(
	verificationReward int64,
	verificationCost int64,
	penaltyForFalsePositive int64,
) economics.GameTheoryAnalysis {
	strategies := map[string][]string{
		"verifier": {
			"thorough_verification",
			"quick_approval",
			"random_rejection",
			"collusive_approval",
		},
	}

	// Calculate payoffs
	payoffs := g.calculateVEIDPayoffs(verificationReward, verificationCost, penaltyForFalsePositive)

	nashEquilibrium := "thorough_verification"
	if verificationReward < verificationCost*2 {
		nashEquilibrium = "quick_approval" // Insufficient incentive for thorough work
	}

	dominantStrategy := nashEquilibrium

	alignment := alignmentAligned
	if nashEquilibrium != "thorough_verification" {
		alignment = alignmentMisaligned
	}

	return economics.GameTheoryAnalysis{
		Scenario:           "veid_verifier_incentives",
		Players:            []string{"verifier"},
		Strategies:         strategies,
		NashEquilibrium:    nashEquilibrium,
		PayoffMatrix:       payoffs,
		DominantStrategy:   dominantStrategy,
		IncentiveAlignment: alignment,
		Recommendations:    g.generateVEIDRecommendations(alignment, verificationReward, verificationCost),
	}
}

// AnalyzeProviderIncentives analyzes compute provider incentive alignment.
func (g *GameTheoryAnalyzer) AnalyzeProviderIncentives(
	takeRateBPS int64,
	avgLeaseValue int64,
	reputationWeight int64,
) economics.GameTheoryAnalysis {
	strategies := map[string][]string{
		"provider": {
			"honest_service",
			"overselling_capacity",
			"underbidding_then_abandon",
			"quality_degradation",
		},
	}

	payoffs := g.calculateProviderPayoffs(takeRateBPS, avgLeaseValue, reputationWeight)

	nashEquilibrium := "honest_service"
	if reputationWeight < 2000 { // Less than 20% weight on reputation
		nashEquilibrium = "quality_degradation"
	}

	alignment := alignmentAligned
	if nashEquilibrium != "honest_service" {
		alignment = "partially_misaligned"
	}

	return economics.GameTheoryAnalysis{
		Scenario:           "provider_incentives",
		Players:            []string{"provider"},
		Strategies:         strategies,
		NashEquilibrium:    nashEquilibrium,
		PayoffMatrix:       payoffs,
		DominantStrategy:   nashEquilibrium,
		IncentiveAlignment: alignment,
		Recommendations:    g.generateProviderRecommendations(alignment, reputationWeight),
	}
}

// AnalyzeMultiPlayerGame analyzes a multi-player game scenario.
func (g *GameTheoryAnalyzer) AnalyzeMultiPlayerGame() []economics.GameTheoryAnalysis {
	analyses := make([]economics.GameTheoryAnalysis, 0, 3)

	// Validator vs Delegator game
	analyses = append(analyses, g.analyzeValidatorDelegatorGame())

	// Provider vs Tenant game
	analyses = append(analyses, g.analyzeProviderTenantGame())

	// Verifier vs User game
	analyses = append(analyses, g.analyzeVerifierUserGame())

	return analyses
}

// Helper functions for payoff calculations

func (g *GameTheoryAnalyzer) calculateValidatorPayoffs(avgStake, avgCommission, slashingBPS int64) [][]float64 {
	// Simplified payoff matrix
	// Rows: validator strategies, Cols: network states
	// honest_participation, double_signing, selective_censorship, collusion, free_riding

	baseReward := float64(avgStake) * float64(g.params.InflationRateBPS) / 10000
	slashCost := float64(avgStake) * float64(slashingBPS) / 10000
	commissionGain := baseReward * float64(avgCommission) / 10000

	return [][]float64{
		{baseReward + commissionGain, baseReward + commissionGain}, // honest
		{-slashCost, -slashCost},                                   // double_signing (always detected)
		{baseReward*0.9 + commissionGain*1.1, -slashCost * 0.5},    // censorship (sometimes detected)
		{baseReward*1.2 + commissionGain*1.5, -slashCost * 0.3},    // collusion (hard to detect)
		{baseReward * 0.7, baseReward * 0.5},                       // free_riding
	}
}

func (g *GameTheoryAnalyzer) calculateDelegatorPayoffs(avgDelegation, avgAPR, unbondingDays int64) [][]float64 {
	baseReward := float64(avgDelegation) * float64(avgAPR) / 10000
	opportunityCost := baseReward * float64(unbondingDays) / 365 * 0.1 // 10% annual opportunity cost

	return [][]float64{
		{baseReward * 1.0, baseReward * 0.9},                                  // top_validator (stable, slightly less reward)
		{baseReward * 1.1, baseReward * 0.8},                                  // small_validator (higher reward, more risk)
		{baseReward * 1.0, baseReward * 0.95},                                 // split_delegation (diversified)
		{baseReward*1.05 - opportunityCost, baseReward*0.9 - opportunityCost}, // redelegation_chasing
		{0, 0}, // hold_liquid
	}
}

func (g *GameTheoryAnalyzer) calculateVEIDPayoffs(reward, cost, penalty int64) [][]float64 {
	r := float64(reward)
	c := float64(cost)
	p := float64(penalty)

	return [][]float64{
		{r - c, r - c},         // thorough (always correct)
		{r - c*0.3, r*0.5 - p}, // quick_approval (sometimes wrong)
		{-p, -p},               // random_rejection (penalized)
		{r*2 - c*0.1, -p * 5},  // collusive (high penalty if caught)
	}
}

func (g *GameTheoryAnalyzer) calculateProviderPayoffs(takeRateBPS, leaseValue, reputationWeight int64) [][]float64 {
	revenue := float64(leaseValue) * float64(10000-takeRateBPS) / 10000
	reputationBonus := revenue * float64(reputationWeight) / 10000

	return [][]float64{
		{revenue + reputationBonus, revenue + reputationBonus},             // honest
		{revenue * 1.5, -reputationBonus * 2},                              // overselling
		{revenue * 0.5, -reputationBonus * 3},                              // underbidding
		{revenue*0.9 + reputationBonus*0.3, revenue*0.7 - reputationBonus}, // degradation
	}
}

func (g *GameTheoryAnalyzer) findNashEquilibrium(payoffs [][]float64) string {
	if len(payoffs) == 0 {
		return alignmentUnknown
	}

	// Find strategy with highest minimum payoff (maximin)
	strategies := []string{strategyHonestParticipation, "double_signing", "selective_censorship", "collusion", "free_riding"}

	bestStrategy := 0
	bestMinPayoff := float64(-1e18)

	for i, row := range payoffs {
		if i >= len(strategies) {
			break
		}
		minPayoff := row[0]
		for _, p := range row {
			if p < minPayoff {
				minPayoff = p
			}
		}
		if minPayoff > bestMinPayoff {
			bestMinPayoff = minPayoff
			bestStrategy = i
		}
	}

	if bestStrategy < len(strategies) {
		return strategies[bestStrategy]
	}
	return "unknown"
}

func (g *GameTheoryAnalyzer) findDominantStrategy(payoffs [][]float64) string {
	// Same as Nash equilibrium for single-player games
	return g.findNashEquilibrium(payoffs)
}

func (g *GameTheoryAnalyzer) findDelegatorNashEquilibrium(payoffs [][]float64) string {
	strategies := []string{"delegate_to_top_validator", "delegate_to_small_validator", "split_delegation", "redelegation_chasing", "hold_liquid"}

	bestStrategy := 0
	bestMinPayoff := float64(-1e18)

	for i, row := range payoffs {
		if i >= len(strategies) {
			break
		}
		minPayoff := row[0]
		for _, p := range row {
			if p < minPayoff {
				minPayoff = p
			}
		}
		if minPayoff > bestMinPayoff {
			bestMinPayoff = minPayoff
			bestStrategy = i
		}
	}

	if bestStrategy < len(strategies) {
		return strategies[bestStrategy]
	}
	return alignmentUnknown
}

func (g *GameTheoryAnalyzer) findDelegatorDominantStrategy(payoffs [][]float64) string {
	return g.findDelegatorNashEquilibrium(payoffs)
}

func (g *GameTheoryAnalyzer) assessIncentiveAlignment(nash, dominant string) string {
	if nash == strategyHonestParticipation && dominant == strategyHonestParticipation {
		return "strongly_aligned"
	}
	if nash == strategyHonestParticipation || dominant == strategyHonestParticipation {
		return "partially_aligned"
	}
	return alignmentMisaligned
}

func (g *GameTheoryAnalyzer) assessDelegatorAlignment(nash string) string {
	if nash == "split_delegation" || nash == "delegate_to_small_validator" {
		return alignmentAligned // Encourages decentralization
	}
	if nash == "delegate_to_top_validator" {
		return "partially_aligned" // Not ideal but not harmful
	}
	return alignmentMisaligned
}

func (g *GameTheoryAnalyzer) generateValidatorRecommendations(nash, alignment string) []string {
	var recommendations []string

	if alignment == alignmentMisaligned {
		recommendations = append(recommendations, "Increase slashing penalties for misbehavior")
		recommendations = append(recommendations, "Implement stronger detection mechanisms for collusion")
	}

	if nash == "free_riding" {
		recommendations = append(recommendations, "Penalize validators with low uptime or participation")
	}

	if alignment == "strongly_aligned" {
		recommendations = append(recommendations, "Current incentive structure is optimal")
	}

	return recommendations
}

func (g *GameTheoryAnalyzer) generateDelegatorRecommendations(alignment string) []string {
	var recommendations []string

	if alignment == alignmentMisaligned {
		recommendations = append(recommendations, "Increase rewards for delegating to smaller validators")
		recommendations = append(recommendations, "Reduce unbonding period to lower switching costs")
	}

	recommendations = append(recommendations, "Consider implementing delegation incentives for decentralization")

	return recommendations
}

func (g *GameTheoryAnalyzer) generateVEIDRecommendations(alignment string, reward, cost int64) []string {
	var recommendations []string

	if alignment == alignmentMisaligned {
		recommendations = append(recommendations, "Increase verification rewards to at least 2x the estimated cost")
		recommendations = append(recommendations, "Implement quality scoring for verifications")
	}

	if reward < cost*2 {
		recommendations = append(recommendations, "Current reward ("+formatInt(reward)+") should be increased relative to cost ("+formatInt(cost)+")")
	}

	return recommendations
}

func (g *GameTheoryAnalyzer) generateProviderRecommendations(alignment string, reputationWeight int64) []string {
	var recommendations []string

	if alignment != alignmentAligned {
		recommendations = append(recommendations, "Increase reputation weight in provider selection")
	}

	if reputationWeight < 2000 {
		recommendations = append(recommendations, "Reputation weight ("+formatInt(reputationWeight)+" BPS) should be at least 2000 BPS (20%)")
	}

	return recommendations
}

func (g *GameTheoryAnalyzer) analyzeValidatorDelegatorGame() economics.GameTheoryAnalysis {
	return economics.GameTheoryAnalysis{
		Scenario: "validator_delegator_interaction",
		Players:  []string{"validator", "delegator"},
		Strategies: map[string][]string{
			"validator": {"low_commission", "high_commission", "competitive"},
			"delegator": {"delegate", "withdraw", "redelegate"},
		},
		NashEquilibrium:    "competitive_commission + delegate",
		IncentiveAlignment: alignmentAligned,
		Recommendations: []string{
			"Commission market is competitive when there are enough validators",
			"Delegators benefit from shopping around for best risk-adjusted returns",
		},
	}
}

func (g *GameTheoryAnalyzer) analyzeProviderTenantGame() economics.GameTheoryAnalysis {
	return economics.GameTheoryAnalysis{
		Scenario: "provider_tenant_interaction",
		Players:  []string{"provider", "tenant"},
		Strategies: map[string][]string{
			"provider": {"honest_fulfillment", "underdeliver", "overcharge"},
			"tenant":   {"verify_usage", "trust", "dispute"},
		},
		NashEquilibrium:    "honest_fulfillment + trust",
		IncentiveAlignment: alignmentAligned,
		Recommendations: []string{
			"Escrow mechanism ensures providers are incentivized to deliver",
			"Usage verification and reputation systems reduce information asymmetry",
		},
	}
}

func (g *GameTheoryAnalyzer) analyzeVerifierUserGame() economics.GameTheoryAnalysis {
	return economics.GameTheoryAnalysis{
		Scenario: "verifier_user_interaction",
		Players:  []string{"verifier", "user"},
		Strategies: map[string][]string{
			"verifier": {"thorough", "quick", "biased"},
			"user":     {"honest_submission", "fraudulent_submission"},
		},
		NashEquilibrium:    "thorough + honest_submission",
		IncentiveAlignment: alignmentAligned,
		Recommendations: []string{
			"Multi-verifier consensus reduces bias risk",
			"Penalty for false submissions deters fraud",
		},
	}
}

func formatInt(i int64) string {
	return big.NewInt(i).String()
}

// ComprehensiveGameTheoryAnalysis performs all game theory analyses.
func (g *GameTheoryAnalyzer) ComprehensiveGameTheoryAnalysis(
	validatorCount int64,
	avgStake int64,
	avgCommission int64,
	slashingPenaltyBPS int64,
	delegatorCount int64,
	avgDelegation int64,
	avgAPR int64,
	unbondingDays int64,
) []economics.GameTheoryAnalysis {
	analyses := make([]economics.GameTheoryAnalysis, 0, 7)

	analyses = append(analyses, g.AnalyzeValidatorIncentives(validatorCount, avgStake, avgCommission, slashingPenaltyBPS))
	analyses = append(analyses, g.AnalyzeDelegatorIncentives(delegatorCount, avgDelegation, avgAPR, unbondingDays))
	analyses = append(analyses, g.AnalyzeVEIDVerifierIncentives(g.params.VEIDRewardPool/1000, g.params.VEIDRewardPool/5000, g.params.VEIDRewardPool/100))
	analyses = append(analyses, g.AnalyzeProviderIncentives(g.params.DefaultTakeRateBPS, 1000000000, 3000))

	multiPlayer := g.AnalyzeMultiPlayerGame()
	analyses = append(analyses, multiPlayer...)

	return analyses
}
