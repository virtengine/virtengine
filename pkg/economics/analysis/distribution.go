package analysis

import (
	"math"
	"math/big"
	"sort"

	"github.com/virtengine/virtengine/pkg/economics"
)

// DistributionAnalyzer analyzes token distribution fairness.
type DistributionAnalyzer struct{}

// NewDistributionAnalyzer creates a new distribution analyzer.
func NewDistributionAnalyzer() *DistributionAnalyzer {
	return &DistributionAnalyzer{}
}

// Holding represents a token holder's balance.
type Holding struct {
	Address string   `json:"address"`
	Balance *big.Int `json:"balance"`
	IsStaked bool    `json:"is_staked"`
}

// AnalyzeDistribution performs comprehensive distribution analysis.
func (a *DistributionAnalyzer) AnalyzeDistribution(
	holdings []Holding,
	totalSupply *big.Int,
) economics.DistributionMetrics {
	if len(holdings) == 0 || totalSupply.Sign() == 0 {
		return economics.DistributionMetrics{}
	}

	// Sort holdings by balance (descending)
	sorted := make([]Holding, len(holdings))
	copy(sorted, holdings)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Balance.Cmp(sorted[j].Balance) > 0
	})

	// Calculate Gini coefficient
	gini := a.calculateGini(sorted)

	// Calculate Lorenz curve
	lorenz := a.calculateLorenzCurve(sorted, totalSupply)

	// Calculate top holdings
	top10 := a.calculateTopNHoldings(sorted, 10, totalSupply)
	top100 := a.calculateTopNHoldings(sorted, 100, totalSupply)

	// Calculate Nakamoto coefficient
	nakamoto := a.calculateNakamotoCoefficient(sorted, totalSupply)

	// Calculate Herfindahl-Hirschman Index
	hhi := a.calculateHHI(sorted, totalSupply)

	// Calculate minimum validators for 51% attack
	minFor51 := a.calculateMinFor51(sorted, totalSupply)

	return economics.DistributionMetrics{
		GiniCoefficient:     gini,
		LorenzCurve:         lorenz,
		Top10HoldingsBPS:    top10,
		Top100HoldingsBPS:   top100,
		NakamotoCoefficient: nakamoto,
		HerfindahlIndex:     hhi,
		MinValidatorsFor51:  minFor51,
	}
}

// calculateGini calculates the Gini coefficient (0 = perfect equality, 1 = perfect inequality).
func (a *DistributionAnalyzer) calculateGini(sortedHoldings []Holding) float64 {
	n := len(sortedHoldings)
	if n == 0 {
		return 0
	}

	// Calculate cumulative wealth
	var totalWealth float64
	for _, h := range sortedHoldings {
		totalWealth += float64(h.Balance.Int64())
	}

	if totalWealth == 0 {
		return 0
	}

	// Gini formula: G = (2 * Σ(i * x_i)) / (n * Σ x_i) - (n + 1) / n
	// Where holdings are sorted ascending
	var sumIndexedWealth float64
	for i, h := range sortedHoldings {
		// Reverse index since we're sorted descending
		reverseIdx := n - i
		sumIndexedWealth += float64(reverseIdx) * float64(h.Balance.Int64())
	}

	gini := (2*sumIndexedWealth)/(float64(n)*totalWealth) - float64(n+1)/float64(n)
	
	// Ensure Gini is in valid range [0, 1]
	if gini < 0 {
		gini = -gini
	}
	if gini > 1 {
		gini = 1
	}

	return gini
}

// calculateLorenzCurve calculates the Lorenz curve (cumulative % of wealth vs cumulative % of population).
func (a *DistributionAnalyzer) calculateLorenzCurve(sortedHoldings []Holding, totalSupply *big.Int) []float64 {
	n := len(sortedHoldings)
	if n == 0 {
		return []float64{}
	}

	// Create 10 points on the Lorenz curve (deciles)
	curve := make([]float64, 10)

	totalSupplyFloat := float64(totalSupply.Int64())
	if totalSupplyFloat == 0 {
		return curve
	}

	// Sort ascending for Lorenz curve
	ascending := make([]Holding, len(sortedHoldings))
	copy(ascending, sortedHoldings)
	sort.Slice(ascending, func(i, j int) bool {
		return ascending[i].Balance.Cmp(ascending[j].Balance) < 0
	})

	cumulativeWealth := float64(0)
	decileSize := n / 10
	if decileSize == 0 {
		decileSize = 1
	}

	for i, h := range ascending {
		cumulativeWealth += float64(h.Balance.Int64())
		
		decile := i / decileSize
		if decile >= 10 {
			decile = 9
		}
		curve[decile] = cumulativeWealth / totalSupplyFloat
	}

	return curve
}

// calculateTopNHoldings calculates the percentage of supply held by top N holders.
func (a *DistributionAnalyzer) calculateTopNHoldings(sortedHoldings []Holding, n int, totalSupply *big.Int) int64 {
	if len(sortedHoldings) == 0 || totalSupply.Sign() == 0 {
		return 0
	}

	if n > len(sortedHoldings) {
		n = len(sortedHoldings)
	}

	topSum := big.NewInt(0)
	for i := 0; i < n; i++ {
		topSum.Add(topSum, sortedHoldings[i].Balance)
	}

	// Return in basis points
	bps := new(big.Int).Mul(topSum, big.NewInt(10000))
	bps.Div(bps, totalSupply)
	return bps.Int64()
}

// calculateNakamotoCoefficient calculates the minimum number of entities to control 51%.
func (a *DistributionAnalyzer) calculateNakamotoCoefficient(sortedHoldings []Holding, totalSupply *big.Int) int64 {
	if len(sortedHoldings) == 0 || totalSupply.Sign() == 0 {
		return 0
	}

	target := new(big.Int).Mul(totalSupply, big.NewInt(51))
	target.Div(target, big.NewInt(100))

	cumulative := big.NewInt(0)
	for i, h := range sortedHoldings {
		cumulative.Add(cumulative, h.Balance)
		if cumulative.Cmp(target) >= 0 {
			return int64(i + 1)
		}
	}

	return int64(len(sortedHoldings))
}

// calculateHHI calculates the Herfindahl-Hirschman Index (measure of concentration).
// HHI ranges from 0 (perfect competition) to 10000 (monopoly).
func (a *DistributionAnalyzer) calculateHHI(sortedHoldings []Holding, totalSupply *big.Int) float64 {
	if len(sortedHoldings) == 0 || totalSupply.Sign() == 0 {
		return 0
	}

	totalSupplyFloat := float64(totalSupply.Int64())
	if totalSupplyFloat == 0 {
		return 0
	}

	var hhi float64
	for _, h := range sortedHoldings {
		share := float64(h.Balance.Int64()) / totalSupplyFloat * 100 // percentage
		hhi += share * share
	}

	return hhi
}

// calculateMinFor51 calculates minimum validators needed for 51% attack.
func (a *DistributionAnalyzer) calculateMinFor51(sortedHoldings []Holding, totalSupply *big.Int) int64 {
	// Same as Nakamoto coefficient for staked holdings
	stakedHoldings := make([]Holding, 0)
	for _, h := range sortedHoldings {
		if h.IsStaked {
			stakedHoldings = append(stakedHoldings, h)
		}
	}

	if len(stakedHoldings) == 0 {
		return a.calculateNakamotoCoefficient(sortedHoldings, totalSupply)
	}

	// Sort staked holdings by balance
	sort.Slice(stakedHoldings, func(i, j int) bool {
		return stakedHoldings[i].Balance.Cmp(stakedHoldings[j].Balance) > 0
	})

	// Calculate total staked
	totalStaked := big.NewInt(0)
	for _, h := range stakedHoldings {
		totalStaked.Add(totalStaked, h.Balance)
	}

	return a.calculateNakamotoCoefficient(stakedHoldings, totalStaked)
}

// GenerateDistributionReport generates a detailed distribution report.
func (a *DistributionAnalyzer) GenerateDistributionReport(
	holdings []Holding,
	totalSupply *big.Int,
) DistributionReport {
	metrics := a.AnalyzeDistribution(holdings, totalSupply)

	report := DistributionReport{
		Metrics:         metrics,
		Recommendations: a.generateRecommendations(metrics),
		RiskLevel:       a.assessRiskLevel(metrics),
	}

	return report
}

// DistributionReport contains the full distribution analysis report.
type DistributionReport struct {
	Metrics         economics.DistributionMetrics `json:"metrics"`
	Recommendations []economics.Recommendation    `json:"recommendations"`
	RiskLevel       string                        `json:"risk_level"`
}

// generateRecommendations generates recommendations based on distribution metrics.
func (a *DistributionAnalyzer) generateRecommendations(metrics economics.DistributionMetrics) []economics.Recommendation {
	var recommendations []economics.Recommendation

	// Gini coefficient recommendations
	if metrics.GiniCoefficient > 0.8 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "distribution",
			Priority:    "high",
			Title:       "Extreme Wealth Inequality",
			Description: "Gini coefficient exceeds 0.8, indicating severe wealth concentration.",
			Impact:      "Centralization risks, potential for governance capture.",
			Action:      "Consider progressive staking rewards or delegation incentives for smaller holders.",
		})
	} else if metrics.GiniCoefficient > 0.6 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "distribution",
			Priority:    "medium",
			Title:       "High Wealth Inequality",
			Description: "Gini coefficient exceeds 0.6, indicating significant wealth concentration.",
			Impact:      "Moderate centralization risk.",
			Action:      "Monitor distribution trends and consider incentive adjustments.",
		})
	}

	// Nakamoto coefficient recommendations
	if metrics.NakamotoCoefficient < 10 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "security",
			Priority:    "critical",
			Title:       "Low Nakamoto Coefficient",
			Description: "Fewer than 10 entities could control 51% of the network.",
			Impact:      "High vulnerability to collusion attacks.",
			Action:      "Implement validator caps, encourage delegation to smaller validators.",
		})
	} else if metrics.NakamotoCoefficient < 20 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "security",
			Priority:    "high",
			Title:       "Moderate Nakamoto Coefficient",
			Description: "Fewer than 20 entities could control 51% of the network.",
			Impact:      "Moderate vulnerability to collusion.",
			Action:      "Continue monitoring and encourage decentralization.",
		})
	}

	// Top 10 concentration recommendations
	if metrics.Top10HoldingsBPS > 5000 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "distribution",
			Priority:    "high",
			Title:       "High Top-10 Concentration",
			Description: "Top 10 holders control more than 50% of supply.",
			Impact:      "Significant influence by small group of holders.",
			Action:      "Encourage broader token distribution and participation.",
		})
	}

	return recommendations
}

// assessRiskLevel assesses overall distribution risk level.
func (a *DistributionAnalyzer) assessRiskLevel(metrics economics.DistributionMetrics) string {
	riskScore := 0

	if metrics.GiniCoefficient > 0.8 {
		riskScore += 3
	} else if metrics.GiniCoefficient > 0.6 {
		riskScore += 2
	} else if metrics.GiniCoefficient > 0.4 {
		riskScore += 1
	}

	if metrics.NakamotoCoefficient < 10 {
		riskScore += 3
	} else if metrics.NakamotoCoefficient < 20 {
		riskScore += 2
	} else if metrics.NakamotoCoefficient < 30 {
		riskScore += 1
	}

	if metrics.Top10HoldingsBPS > 5000 {
		riskScore += 2
	} else if metrics.Top10HoldingsBPS > 3000 {
		riskScore += 1
	}

	switch {
	case riskScore >= 6:
		return "critical"
	case riskScore >= 4:
		return "high"
	case riskScore >= 2:
		return "medium"
	default:
		return "low"
	}
}

// SimulateRedistribution simulates the effect of a redistribution policy.
func (a *DistributionAnalyzer) SimulateRedistribution(
	holdings []Holding,
	totalSupply *big.Int,
	policy RedistributionPolicy,
) RedistributionResult {
	// Deep copy holdings
	newHoldings := make([]Holding, len(holdings))
	for i, h := range holdings {
		newHoldings[i] = Holding{
			Address:  h.Address,
			Balance:  new(big.Int).Set(h.Balance),
			IsStaked: h.IsStaked,
		}
	}

	// Apply policy
	switch policy.Type {
	case "progressive_rewards":
		newHoldings = a.applyProgressiveRewards(newHoldings, totalSupply, policy)
	case "validator_cap":
		newHoldings = a.applyValidatorCap(newHoldings, totalSupply, policy)
	case "delegation_bonus":
		newHoldings = a.applyDelegationBonus(newHoldings, totalSupply, policy)
	}

	beforeMetrics := a.AnalyzeDistribution(holdings, totalSupply)
	afterMetrics := a.AnalyzeDistribution(newHoldings, totalSupply)

	return RedistributionResult{
		Policy:        policy,
		BeforeMetrics: beforeMetrics,
		AfterMetrics:  afterMetrics,
		GiniChange:    afterMetrics.GiniCoefficient - beforeMetrics.GiniCoefficient,
		NakamotoChange: afterMetrics.NakamotoCoefficient - beforeMetrics.NakamotoCoefficient,
	}
}

// RedistributionPolicy defines a redistribution policy.
type RedistributionPolicy struct {
	Type       string  `json:"type"`
	Parameters map[string]int64 `json:"parameters"`
}

// RedistributionResult contains redistribution simulation results.
type RedistributionResult struct {
	Policy         RedistributionPolicy          `json:"policy"`
	BeforeMetrics  economics.DistributionMetrics `json:"before_metrics"`
	AfterMetrics   economics.DistributionMetrics `json:"after_metrics"`
	GiniChange     float64                       `json:"gini_change"`
	NakamotoChange int64                         `json:"nakamoto_change"`
}

// applyProgressiveRewards applies progressive rewards (smaller holders get bonus).
func (a *DistributionAnalyzer) applyProgressiveRewards(
	holdings []Holding,
	totalSupply *big.Int,
	policy RedistributionPolicy,
) []Holding {
	bonusBPS := policy.Parameters["bonus_bps"]
	if bonusBPS == 0 {
		bonusBPS = 100 // 1% default bonus
	}

	threshold := policy.Parameters["threshold_bps"]
	if threshold == 0 {
		threshold = 100 // 1% of supply threshold
	}

	thresholdAmount := new(big.Int).Mul(totalSupply, big.NewInt(threshold))
	thresholdAmount.Div(thresholdAmount, big.NewInt(10000))

	for i := range holdings {
		if holdings[i].Balance.Cmp(thresholdAmount) < 0 {
			// Apply bonus
			bonus := new(big.Int).Mul(holdings[i].Balance, big.NewInt(bonusBPS))
			bonus.Div(bonus, big.NewInt(10000))
			holdings[i].Balance.Add(holdings[i].Balance, bonus)
		}
	}

	return holdings
}

// applyValidatorCap applies a cap on validator stake.
func (a *DistributionAnalyzer) applyValidatorCap(
	holdings []Holding,
	totalSupply *big.Int,
	policy RedistributionPolicy,
) []Holding {
	capBPS := policy.Parameters["cap_bps"]
	if capBPS == 0 {
		capBPS = 500 // 5% cap
	}

	capAmount := new(big.Int).Mul(totalSupply, big.NewInt(capBPS))
	capAmount.Div(capAmount, big.NewInt(10000))

	for i := range holdings {
		if holdings[i].IsStaked && holdings[i].Balance.Cmp(capAmount) > 0 {
			holdings[i].Balance.Set(capAmount)
		}
	}

	return holdings
}

// applyDelegationBonus applies bonus for delegation to smaller validators.
func (a *DistributionAnalyzer) applyDelegationBonus(
	holdings []Holding,
	totalSupply *big.Int,
	policy RedistributionPolicy,
) []Holding {
	bonusBPS := policy.Parameters["bonus_bps"]
	if bonusBPS == 0 {
		bonusBPS = 50 // 0.5% bonus
	}

	// Calculate median stake
	stakes := make([]*big.Int, 0)
	for _, h := range holdings {
		if h.IsStaked {
			stakes = append(stakes, h.Balance)
		}
	}

	if len(stakes) == 0 {
		return holdings
	}

	sort.Slice(stakes, func(i, j int) bool {
		return stakes[i].Cmp(stakes[j]) < 0
	})
	medianStake := stakes[len(stakes)/2]

	for i := range holdings {
		if holdings[i].IsStaked && holdings[i].Balance.Cmp(medianStake) < 0 {
			bonus := new(big.Int).Mul(holdings[i].Balance, big.NewInt(bonusBPS))
			bonus.Div(bonus, big.NewInt(10000))
			holdings[i].Balance.Add(holdings[i].Balance, bonus)
		}
	}

	return holdings
}

// CalculateDecentralizationScore calculates an overall decentralization score (0-100).
func (a *DistributionAnalyzer) CalculateDecentralizationScore(metrics economics.DistributionMetrics) int64 {
	score := int64(100)

	// Gini penalty (max -30 points)
	giniPenalty := int64(math.Min(30, metrics.GiniCoefficient*30))
	score -= giniPenalty

	// Nakamoto bonus/penalty
	if metrics.NakamotoCoefficient >= 100 {
		// Excellent decentralization
	} else if metrics.NakamotoCoefficient >= 50 {
		score -= 5
	} else if metrics.NakamotoCoefficient >= 20 {
		score -= 15
	} else if metrics.NakamotoCoefficient >= 10 {
		score -= 25
	} else {
		score -= 35
	}

	// Top 10 concentration penalty
	if metrics.Top10HoldingsBPS > 5000 {
		score -= 20
	} else if metrics.Top10HoldingsBPS > 3000 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}
