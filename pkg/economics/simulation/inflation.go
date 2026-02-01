package simulation

import (
	"math/big"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
)

// InflationSimulator simulates inflation/deflation dynamics.
type InflationSimulator struct {
	params economics.TokenomicsParams
}

// NewInflationSimulator creates a new inflation simulator.
func NewInflationSimulator(params economics.TokenomicsParams) *InflationSimulator {
	return &InflationSimulator{params: params}
}

// SimulateYear simulates one year of inflation dynamics.
func (s *InflationSimulator) SimulateYear(state economics.NetworkState) economics.SimulationResult {
	result := economics.SimulationResult{
		Scenario:     "annual_inflation",
		Duration:     365 * 24 * time.Hour,
		InitialState: state,
		Snapshots:    make([]economics.NetworkStateSnapshot, 0, 12),
	}

	currentSupply := new(big.Int).Set(state.TotalSupply)
	currentStaked := new(big.Int).Set(state.TotalStaked)
	blocksPerMonth := s.params.BlocksPerYear / 12

	var totalInflation int64
	var totalStakingRatio int64

	for month := 0; month < 12; month++ {
		// Calculate staking ratio (in basis points)
		stakingRatioBPS := s.calculateStakingRatioBPS(currentSupply, currentStaked)

		// Adjust inflation based on staking ratio
		inflationBPS := s.adjustInflation(stakingRatioBPS)

		// Calculate monthly mint
		monthlyMint := s.calculateMonthlyMint(currentSupply, inflationBPS)

		// Update supply
		currentSupply = new(big.Int).Add(currentSupply, monthlyMint)

		// Simulate staking changes (simplified model)
		stakingChange := s.simulateStakingChange(currentStaked, inflationBPS, stakingRatioBPS)
		currentStaked = new(big.Int).Add(currentStaked, stakingChange)

		// Ensure staked doesn't exceed supply
		if currentStaked.Cmp(currentSupply) > 0 {
			currentStaked = new(big.Int).Set(currentSupply)
		}

		// Calculate APR
		apr := s.calculateAPR(inflationBPS, stakingRatioBPS)

		// Record snapshot
		result.Snapshots = append(result.Snapshots, economics.NetworkStateSnapshot{
			BlockHeight:     state.BlockHeight + int64(month+1)*blocksPerMonth,
			TotalSupply:     new(big.Int).Set(currentSupply),
			StakingRatioBPS: stakingRatioBPS,
			InflationBPS:    inflationBPS,
			APR:             apr,
		})

		totalInflation += inflationBPS
		totalStakingRatio += stakingRatioBPS
	}

	result.FinalState = economics.NetworkState{
		BlockHeight: state.BlockHeight + s.params.BlocksPerYear,
		Timestamp:   state.Timestamp.Add(365 * 24 * time.Hour),
		TotalSupply: currentSupply,
		TotalStaked: currentStaked,
	}

	// Calculate supply growth
	supplyGrowth := new(big.Int).Sub(currentSupply, state.TotalSupply)
	supplyGrowthBPS := new(big.Int).Mul(supplyGrowth, big.NewInt(10000))
	supplyGrowthBPS.Div(supplyGrowthBPS, state.TotalSupply)

	result.Metrics = economics.SimulationMetrics{
		AvgInflationBPS:    totalInflation / 12,
		AvgStakingRatioBPS: totalStakingRatio / 12,
		SupplyGrowthBPS:    supplyGrowthBPS.Int64(),
	}

	// Generate recommendations
	result.Recommendations = s.generateRecommendations(result.Metrics)

	return result
}

// SimulateMultiYear runs multi-year simulation with Monte Carlo variations.
func (s *InflationSimulator) SimulateMultiYear(state economics.NetworkState, years int, scenarios int) []economics.SimulationResult {
	results := make([]economics.SimulationResult, scenarios)

	for i := 0; i < scenarios; i++ {
		currentState := state
		var yearlySnapshots []economics.NetworkStateSnapshot

		for year := 0; year < years; year++ {
			yearResult := s.SimulateYear(currentState)
			yearlySnapshots = append(yearlySnapshots, yearResult.Snapshots...)
			currentState = yearResult.FinalState
		}

		results[i] = economics.SimulationResult{
			Scenario:     "multi_year_simulation",
			Duration:     time.Duration(years) * 365 * 24 * time.Hour,
			InitialState: state,
			FinalState:   currentState,
			Snapshots:    yearlySnapshots,
		}
	}

	return results
}

// calculateStakingRatioBPS calculates staking ratio in basis points.
func (s *InflationSimulator) calculateStakingRatioBPS(supply, staked *big.Int) int64 {
	if supply.Sign() == 0 {
		return 0
	}
	ratio := new(big.Int).Mul(staked, big.NewInt(10000))
	ratio.Div(ratio, supply)
	return ratio.Int64()
}

// adjustInflation adjusts inflation based on staking ratio.
// Uses a bonding curve: higher staking = lower inflation (and vice versa).
func (s *InflationSimulator) adjustInflation(stakingRatioBPS int64) int64 {
	// If staking ratio equals target, use target inflation
	if stakingRatioBPS == s.params.TargetStakingRatioBPS {
		return s.params.TargetInflationBPS
	}

	// If staking is below target, increase inflation to incentivize staking
	// If staking is above target, decrease inflation
	deviation := s.params.TargetStakingRatioBPS - stakingRatioBPS

	// Adjustment factor: 1 BPS change in staking ratio = 0.1 BPS change in inflation
	adjustment := deviation / 10

	newInflation := s.params.TargetInflationBPS + adjustment

	// Clamp to min/max
	if newInflation < s.params.MinInflationBPS {
		return s.params.MinInflationBPS
	}
	if newInflation > s.params.MaxInflationBPS {
		return s.params.MaxInflationBPS
	}

	return newInflation
}

// calculateMonthlyMint calculates tokens to mint in a month.
func (s *InflationSimulator) calculateMonthlyMint(supply *big.Int, inflationBPS int64) *big.Int {
	// Monthly = (Supply * InflationBPS) / 10000 / 12
	mint := new(big.Int).Mul(supply, big.NewInt(inflationBPS))
	mint.Div(mint, big.NewInt(10000))
	mint.Div(mint, big.NewInt(12))
	return mint
}

// simulateStakingChange simulates how staking changes based on APR.
func (s *InflationSimulator) simulateStakingChange(staked *big.Int, inflationBPS, stakingRatioBPS int64) *big.Int {
	apr := s.calculateAPR(inflationBPS, stakingRatioBPS)

	// Higher APR attracts more staking (simplified model)
	// If APR > 10% (1000 BPS), net inflow; if < 5% (500 BPS), net outflow
	var changeRateBPS int64
	if apr > 1000 {
		changeRateBPS = (apr - 1000) / 10 // 0.1% inflow per 1% above 10% APR
	} else if apr < 500 {
		changeRateBPS = -(500 - apr) / 10 // 0.1% outflow per 1% below 5% APR
	}

	// Monthly change
	change := new(big.Int).Mul(staked, big.NewInt(changeRateBPS))
	change.Div(change, big.NewInt(10000))
	change.Div(change, big.NewInt(12))

	return change
}

// calculateAPR calculates staking APR in basis points.
func (s *InflationSimulator) calculateAPR(inflationBPS, stakingRatioBPS int64) int64 {
	if stakingRatioBPS == 0 {
		return 0
	}
	// APR = (Inflation * 10000) / StakingRatio
	// This gives the return to stakers
	apr := (inflationBPS * 10000) / stakingRatioBPS
	return apr
}

// generateRecommendations generates recommendations based on simulation results.
func (s *InflationSimulator) generateRecommendations(metrics economics.SimulationMetrics) []economics.Recommendation {
	var recommendations []economics.Recommendation

	// Inflation recommendations
	if metrics.AvgInflationBPS > 1500 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "inflation",
			Priority:    "high",
			Title:       "High Inflation Rate",
			Description: "Average inflation exceeds 15%, which may dilute token value.",
			Impact:      "Token holders may see significant value dilution.",
			Action:      "Consider reducing base reward per block or adjusting inflation curve.",
		})
	}

	if metrics.AvgInflationBPS < 200 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "inflation",
			Priority:    "medium",
			Title:       "Low Inflation Rate",
			Description: "Average inflation below 2% may not provide sufficient staking incentives.",
			Impact:      "Reduced staking participation and network security.",
			Action:      "Consider increasing minimum inflation floor.",
		})
	}

	// Staking recommendations
	if metrics.AvgStakingRatioBPS < 5000 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "staking",
			Priority:    "critical",
			Title:       "Low Staking Ratio",
			Description: "Staking ratio below 50% increases attack vulnerability.",
			Impact:      "Network security is compromised; 51% attacks become cheaper.",
			Action:      "Increase staking incentives through higher APR or shorter unbonding.",
		})
	}

	if metrics.AvgStakingRatioBPS > 9000 {
		recommendations = append(recommendations, economics.Recommendation{
			Category:    "staking",
			Priority:    "medium",
			Title:       "Very High Staking Ratio",
			Description: "Staking ratio above 90% may reduce liquidity.",
			Impact:      "Market liquidity constraints, difficulty for new users to acquire tokens.",
			Action:      "Consider adjusting inflation curve to allow more token circulation.",
		})
	}

	return recommendations
}

// AnalyzeInflationDynamics provides detailed inflation analysis.
func (s *InflationSimulator) AnalyzeInflationDynamics(state economics.NetworkState) economics.InflationAnalysis {
	stakingRatioBPS := s.calculateStakingRatioBPS(state.TotalSupply, state.TotalStaked)
	currentInflation := s.adjustInflation(stakingRatioBPS)

	// Project future inflation (simplified 1-year projection)
	yearResult := s.SimulateYear(state)
	projectedInflation := yearResult.Metrics.AvgInflationBPS

	// Calculate yearly minted tokens
	yearlyMint := s.calculateMonthlyMint(state.TotalSupply, currentInflation)
	yearlyMint.Mul(yearlyMint, big.NewInt(12))

	// Determine trend
	var trend string
	if projectedInflation > currentInflation+50 {
		trend = "inflationary"
	} else if projectedInflation < currentInflation-50 {
		trend = "deflationary"
	} else {
		trend = "stable"
	}

	// Sustainability score (0-100)
	// Based on how close inflation is to target and stability
	sustainabilityScore := int64(100)
	deviation := abs(currentInflation - s.params.TargetInflationBPS)
	sustainabilityScore -= deviation / 20 // -5 points per 1% deviation
	if sustainabilityScore < 0 {
		sustainabilityScore = 0
	}

	return economics.InflationAnalysis{
		CurrentRateBPS:      currentInflation,
		ProjectedRateBPS:    projectedInflation,
		YearlyMintedTokens:  yearlyMint,
		SustainabilityScore: sustainabilityScore,
		Trend:               trend,
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

