package simulation

import (
	"math/big"
	"sort"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
)

// StakingSimulator simulates staking reward dynamics.
type StakingSimulator struct {
	params economics.TokenomicsParams
}

// NewStakingSimulator creates a new staking simulator.
func NewStakingSimulator(params economics.TokenomicsParams) *StakingSimulator {
	return &StakingSimulator{params: params}
}

// SimulateRewardDistribution simulates reward distribution for an epoch.
func (s *StakingSimulator) SimulateRewardDistribution(
	validators []economics.ValidatorState,
	epochBlocks int64,
) []ValidatorRewardResult {
	results := make([]ValidatorRewardResult, len(validators))

	// Calculate total stake
	totalStake := big.NewInt(0)
	for _, v := range validators {
		totalStake.Add(totalStake, v.TotalStake)
	}

	if totalStake.Sign() == 0 {
		return results
	}

	// Calculate epoch reward pool
	epochRewardPool := s.params.BaseRewardPerBlock * epochBlocks

	for i, validator := range validators {
		results[i] = s.calculateValidatorReward(validator, totalStake, epochRewardPool)
	}

	return results
}

// ValidatorRewardResult contains calculated rewards for a validator.
type ValidatorRewardResult struct {
	Address             string   `json:"address"`
	BlockProposalReward *big.Int `json:"block_proposal_reward"`
	VEIDReward          *big.Int `json:"veid_reward"`
	UptimeReward        *big.Int `json:"uptime_reward"`
	TotalReward         *big.Int `json:"total_reward"`
	CommissionEarned    *big.Int `json:"commission_earned"`
	DelegatorRewards    *big.Int `json:"delegator_rewards"`
	EffectiveAPR        int64    `json:"effective_apr_bps"`
}

// calculateValidatorReward calculates rewards for a single validator.
func (s *StakingSimulator) calculateValidatorReward(
	validator economics.ValidatorState,
	totalStake *big.Int,
	epochRewardPool int64,
) ValidatorRewardResult {
	result := ValidatorRewardResult{
		Address:             validator.Address,
		BlockProposalReward: big.NewInt(0),
		VEIDReward:          big.NewInt(0),
		UptimeReward:        big.NewInt(0),
		TotalReward:         big.NewInt(0),
		CommissionEarned:    big.NewInt(0),
		DelegatorRewards:    big.NewInt(0),
	}

	if totalStake.Sign() == 0 || validator.TotalStake.Sign() == 0 {
		return result
	}

	// Stake weight (scaled by 1e6 for precision)
	stakeWeight := new(big.Int).Mul(validator.TotalStake, big.NewInt(1000000))
	stakeWeight.Div(stakeWeight, totalStake)

	// Base reward proportional to stake
	baseReward := new(big.Int).Mul(big.NewInt(epochRewardPool), stakeWeight)
	baseReward.Div(baseReward, big.NewInt(1000000))

	// Performance multiplier (0.5 to 1.5x based on uptime score 0-10000)
	// Score 0 = 50% multiplier, Score 10000 = 150% multiplier
	performanceMultiplier := 5000 + validator.UptimeScore
	adjustedReward := new(big.Int).Mul(baseReward, big.NewInt(performanceMultiplier))
	adjustedReward.Div(adjustedReward, big.NewInt(10000))

	// Reward weights (must sum to 10000)
	const (
		weightBlockProposal = 5000 // 50%
		weightVEID          = 2000 // 20%
		weightUptime        = 3000 // 30%
	)

	// Block proposal reward
	result.BlockProposalReward = new(big.Int).Mul(adjustedReward, big.NewInt(weightBlockProposal))
	result.BlockProposalReward.Div(result.BlockProposalReward, big.NewInt(10000))

	// VEID verification reward
	result.VEIDReward = new(big.Int).Mul(adjustedReward, big.NewInt(weightVEID))
	result.VEIDReward.Div(result.VEIDReward, big.NewInt(10000))

	// Bonus for high VEID verification count
	if validator.VEIDVerifications > 100 {
		bonus := new(big.Int).Mul(result.VEIDReward, big.NewInt(1000)) // 10% bonus
		bonus.Div(bonus, big.NewInt(10000))
		result.VEIDReward.Add(result.VEIDReward, bonus)
	}

	// Uptime reward
	result.UptimeReward = new(big.Int).Mul(adjustedReward, big.NewInt(weightUptime))
	result.UptimeReward.Div(result.UptimeReward, big.NewInt(10000))

	// Total reward
	result.TotalReward = new(big.Int).Add(result.BlockProposalReward, result.VEIDReward)
	result.TotalReward.Add(result.TotalReward, result.UptimeReward)

	// Commission split
	result.CommissionEarned = new(big.Int).Mul(result.TotalReward, big.NewInt(validator.Commission))
	result.CommissionEarned.Div(result.CommissionEarned, big.NewInt(10000))

	result.DelegatorRewards = new(big.Int).Sub(result.TotalReward, result.CommissionEarned)

	// Effective APR (annualized, assuming 1 epoch = 1 day for simplicity)
	if validator.TotalStake.Sign() > 0 {
		annualizedReward := new(big.Int).Mul(result.TotalReward, big.NewInt(365))
		apr := new(big.Int).Mul(annualizedReward, big.NewInt(10000))
		apr.Div(apr, validator.TotalStake)
		result.EffectiveAPR = apr.Int64()
	}

	return result
}

// OptimizeRewardParameters finds optimal reward parameters.
func (s *StakingSimulator) OptimizeRewardParameters(
	state economics.NetworkState,
	validators []economics.ValidatorState,
) RewardOptimizationResult {
	result := RewardOptimizationResult{
		CurrentParams:   s.params,
		Recommendations: make([]economics.Recommendation, 0),
	}

	// Analyze current state
	currentStakingRatio := s.calculateStakingRatio(state)
	result.CurrentStakingRatioBPS = currentStakingRatio

	// Simulate with current params
	currentAPR := s.estimateNetworkAPR(validators)
	result.CurrentAPR = currentAPR

	// Find optimal parameters through parameter sweep
	bestParams := s.params
	bestScore := s.evaluateParameters(s.params, state, validators)

	// Try different base reward values
	for multiplier := int64(50); multiplier <= 200; multiplier += 10 {
		testParams := s.params
		testParams.BaseRewardPerBlock = (s.params.BaseRewardPerBlock * multiplier) / 100
		score := s.evaluateParameters(testParams, state, validators)
		if score > bestScore {
			bestScore = score
			bestParams = testParams
		}
	}

	// Try different VEID reward pool sizes
	for multiplier := int64(50); multiplier <= 200; multiplier += 10 {
		testParams := s.params
		testParams.VEIDRewardPool = (s.params.VEIDRewardPool * multiplier) / 100
		score := s.evaluateParameters(testParams, state, validators)
		if score > bestScore {
			bestScore = score
			bestParams = testParams
		}
	}

	result.OptimalParams = bestParams
	result.OptimalScore = bestScore

	// Calculate improvement
	if bestParams.BaseRewardPerBlock != s.params.BaseRewardPerBlock {
		change := (bestParams.BaseRewardPerBlock * 100 / s.params.BaseRewardPerBlock) - 100
		result.Recommendations = append(result.Recommendations, economics.Recommendation{
			Category:    "staking",
			Priority:    "medium",
			Title:       "Adjust Base Reward",
			Description: "Optimal base reward per block differs from current setting.",
			Impact:      "Improved staking equilibrium and validator incentives.",
			Action:      formatRecommendation("base_reward_per_block", s.params.BaseRewardPerBlock, bestParams.BaseRewardPerBlock, change),
		})
	}

	if bestParams.VEIDRewardPool != s.params.VEIDRewardPool {
		change := (bestParams.VEIDRewardPool * 100 / s.params.VEIDRewardPool) - 100
		result.Recommendations = append(result.Recommendations, economics.Recommendation{
			Category:    "veid",
			Priority:    "medium",
			Title:       "Adjust VEID Reward Pool",
			Description: "Optimal VEID reward pool differs from current setting.",
			Impact:      "Better incentive alignment for identity verification work.",
			Action:      formatRecommendation("veid_reward_pool", s.params.VEIDRewardPool, bestParams.VEIDRewardPool, change),
		})
	}

	return result
}

// RewardOptimizationResult contains optimization analysis results.
type RewardOptimizationResult struct {
	CurrentParams          economics.TokenomicsParams `json:"current_params"`
	OptimalParams          economics.TokenomicsParams `json:"optimal_params"`
	CurrentStakingRatioBPS int64                      `json:"current_staking_ratio_bps"`
	CurrentAPR             int64                      `json:"current_apr_bps"`
	OptimalScore           float64                    `json:"optimal_score"`
	Recommendations        []economics.Recommendation `json:"recommendations"`
}

// calculateStakingRatio calculates staking ratio in basis points.
func (s *StakingSimulator) calculateStakingRatio(state economics.NetworkState) int64 {
	if state.TotalSupply.Sign() == 0 {
		return 0
	}
	ratio := new(big.Int).Mul(state.TotalStaked, big.NewInt(10000))
	ratio.Div(ratio, state.TotalSupply)
	return ratio.Int64()
}

// estimateNetworkAPR estimates average network APR.
func (s *StakingSimulator) estimateNetworkAPR(validators []economics.ValidatorState) int64 {
	if len(validators) == 0 {
		return 0
	}

	totalStake := big.NewInt(0)
	for _, v := range validators {
		totalStake.Add(totalStake, v.TotalStake)
	}

	if totalStake.Sign() == 0 {
		return 0
	}

	// Annual rewards
	annualRewards := s.params.BaseRewardPerBlock * s.params.BlocksPerYear

	// APR = (Annual Rewards / Total Staked) * 10000
	apr := new(big.Int).Mul(big.NewInt(annualRewards), big.NewInt(10000))
	apr.Div(apr, totalStake)
	return apr.Int64()
}

// evaluateParameters scores a parameter set.
func (s *StakingSimulator) evaluateParameters(
	params economics.TokenomicsParams,
	state economics.NetworkState,
	validators []economics.ValidatorState,
) float64 {
	score := float64(100)

	// Calculate resulting APR
	tempSim := NewStakingSimulator(params)
	apr := tempSim.estimateNetworkAPR(validators)

	// Optimal APR is 8-12% (800-1200 BPS)
	if apr < 800 {
		score -= float64(800-apr) / 10
	} else if apr > 1200 {
		score -= float64(apr-1200) / 10
	}

	// Calculate resulting staking ratio tendency
	stakingRatio := s.calculateStakingRatio(state)

	// Optimal staking ratio is 60-70%
	if stakingRatio < 6000 {
		score -= float64(6000-stakingRatio) / 100
	} else if stakingRatio > 7000 {
		score -= float64(stakingRatio-7000) / 100
	}

	// Penalize extreme parameters
	if params.BaseRewardPerBlock < 100000 {
		score -= 20
	}
	if params.BaseRewardPerBlock > 10000000 {
		score -= 20
	}

	return score
}

func formatRecommendation(param string, current, optimal, changePercent int64) string {
	return "Update " + param + " from " + formatAmount(current) + " to " + formatAmount(optimal) + " (" + formatPercent(changePercent) + ")"
}

func formatAmount(amount int64) string {
	return big.NewInt(amount).String()
}

func formatPercent(bps int64) string {
	if bps >= 0 {
		return "+" + big.NewInt(bps).String() + "%"
	}
	return big.NewInt(bps).String() + "%"
}

// AnalyzeStakingDynamics provides detailed staking analysis.
func (s *StakingSimulator) AnalyzeStakingDynamics(
	state economics.NetworkState,
	validators []economics.ValidatorState,
) economics.StakingAnalysis {
	currentRatio := s.calculateStakingRatio(state)
	currentAPR := s.estimateNetworkAPR(validators)

	// Optimal ratio based on security vs liquidity tradeoff
	optimalRatio := int64(6700) // 67%

	// Optimal APR based on inflation and market conditions
	optimalAPR := s.calculateOptimalAPR(currentRatio)

	// Concentration risk analysis
	concentrationRisk := s.analyzeConcentration(validators)

	// Unbonding pressure (simplified)
	unbondingPressure := s.estimateUnbondingPressure(currentAPR)

	return economics.StakingAnalysis{
		CurrentRatioBPS:   currentRatio,
		OptimalRatioBPS:   optimalRatio,
		CurrentAPR:        currentAPR,
		OptimalAPR:        optimalAPR,
		ValidatorCount:    int64(len(validators)),
		DelegatorCount:    state.TotalDelegators,
		ConcentrationRisk: concentrationRisk,
		UnbondingPressure: unbondingPressure,
	}
}

// calculateOptimalAPR calculates optimal APR given staking ratio.
func (s *StakingSimulator) calculateOptimalAPR(stakingRatioBPS int64) int64 {
	// Higher staking ratio needs lower APR to maintain equilibrium
	// Lower staking ratio needs higher APR to attract stakers

	// Base optimal APR
	baseAPR := int64(1000) // 10%

	// Adjustment based on deviation from target
	deviation := s.params.TargetStakingRatioBPS - stakingRatioBPS
	adjustment := deviation / 50 // 0.2% APR adjustment per 1% staking deviation

	optimalAPR := baseAPR + adjustment

	// Clamp to reasonable range
	if optimalAPR < 400 {
		return 400 // Min 4%
	}
	if optimalAPR > 2000 {
		return 2000 // Max 20%
	}
	return optimalAPR
}

// analyzeConcentration analyzes stake concentration among validators.
func (s *StakingSimulator) analyzeConcentration(validators []economics.ValidatorState) string {
	if len(validators) == 0 {
		return "insufficient_data"
	}

	// Sort by stake
	sorted := make([]economics.ValidatorState, len(validators))
	copy(sorted, validators)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalStake.Cmp(sorted[j].TotalStake) > 0
	})

	// Calculate total stake
	totalStake := big.NewInt(0)
	for _, v := range sorted {
		totalStake.Add(totalStake, v.TotalStake)
	}

	if totalStake.Sign() == 0 {
		return "insufficient_data"
	}

	// Calculate top 10 concentration
	top10Stake := big.NewInt(0)
	for i := 0; i < 10 && i < len(sorted); i++ {
		top10Stake.Add(top10Stake, sorted[i].TotalStake)
	}

	top10Ratio := new(big.Int).Mul(top10Stake, big.NewInt(10000))
	top10Ratio.Div(top10Ratio, totalStake)

	if top10Ratio.Int64() > 5000 {
		return "high" // Top 10 control >50%
	} else if top10Ratio.Int64() > 3300 {
		return "moderate" // Top 10 control >33%
	}
	return "low"
}

// estimateUnbondingPressure estimates unbonding pressure based on APR.
func (s *StakingSimulator) estimateUnbondingPressure(currentAPR int64) float64 {
	// Lower APR = higher unbonding pressure
	if currentAPR < 400 {
		return 0.9 // Very high pressure
	} else if currentAPR < 600 {
		return 0.6 // High pressure
	} else if currentAPR < 800 {
		return 0.3 // Moderate pressure
	}
	return 0.1 // Low pressure
}

// SimulateUnbonding simulates unbonding dynamics.
func (s *StakingSimulator) SimulateUnbonding(
	state economics.NetworkState,
	unbondingAmount *big.Int,
	durationDays int64,
) UnbondingSimulationResult {
	result := UnbondingSimulationResult{
		UnbondingAmount: unbondingAmount,
		UnbondingDays:   durationDays,
		DailySnapshots:  make([]UnbondingSnapshot, durationDays),
	}

	currentStaked := new(big.Int).Set(state.TotalStaked)
	dailyUnbond := new(big.Int).Div(unbondingAmount, big.NewInt(durationDays))

	for day := int64(0); day < durationDays; day++ {
		currentStaked.Sub(currentStaked, dailyUnbond)

		stakingRatio := int64(0)
		if state.TotalSupply.Sign() > 0 {
			ratio := new(big.Int).Mul(currentStaked, big.NewInt(10000))
			ratio.Div(ratio, state.TotalSupply)
			stakingRatio = ratio.Int64()
		}

		result.DailySnapshots[day] = UnbondingSnapshot{
			Day:             day + 1,
			RemainingStaked: new(big.Int).Set(currentStaked),
			StakingRatioBPS: stakingRatio,
		}
	}

	result.FinalStaked = currentStaked
	result.FinalRatioBPS = result.DailySnapshots[durationDays-1].StakingRatioBPS

	return result
}

// UnbondingSimulationResult contains unbonding simulation results.
type UnbondingSimulationResult struct {
	UnbondingAmount *big.Int            `json:"unbonding_amount"`
	UnbondingDays   int64               `json:"unbonding_days"`
	DailySnapshots  []UnbondingSnapshot `json:"daily_snapshots"`
	FinalStaked     *big.Int            `json:"final_staked"`
	FinalRatioBPS   int64               `json:"final_ratio_bps"`
}

// UnbondingSnapshot is a daily snapshot during unbonding.
type UnbondingSnapshot struct {
	Day             int64    `json:"day"`
	RemainingStaked *big.Int `json:"remaining_staked"`
	StakingRatioBPS int64    `json:"staking_ratio_bps"`
}

// SimulateEpoch simulates a complete epoch of staking rewards.
func (s *StakingSimulator) SimulateEpoch(
	state economics.NetworkState,
	validators []economics.ValidatorState,
	epochDuration time.Duration,
) EpochSimulationResult {
	epochBlocks := int64(epochDuration.Seconds()) / 5 // Assuming 5s blocks

	validatorRewards := s.SimulateRewardDistribution(validators, epochBlocks)

	totalDistributed := big.NewInt(0)
	for _, r := range validatorRewards {
		totalDistributed.Add(totalDistributed, r.TotalReward)
	}

	newSupply := new(big.Int).Add(state.TotalSupply, totalDistributed)

	return EpochSimulationResult{
		EpochBlocks:        epochBlocks,
		TotalDistributed:   totalDistributed,
		ValidatorRewards:   validatorRewards,
		NewTotalSupply:     newSupply,
		InflationThisEpoch: s.calculateEpochInflation(state.TotalSupply, totalDistributed),
	}
}

// EpochSimulationResult contains epoch simulation results.
type EpochSimulationResult struct {
	EpochBlocks        int64                   `json:"epoch_blocks"`
	TotalDistributed   *big.Int                `json:"total_distributed"`
	ValidatorRewards   []ValidatorRewardResult `json:"validator_rewards"`
	NewTotalSupply     *big.Int                `json:"new_total_supply"`
	InflationThisEpoch int64                   `json:"inflation_this_epoch_bps"`
}

// calculateEpochInflation calculates inflation for an epoch in basis points.
func (s *StakingSimulator) calculateEpochInflation(supply, minted *big.Int) int64 {
	if supply.Sign() == 0 {
		return 0
	}
	inflation := new(big.Int).Mul(minted, big.NewInt(10000))
	inflation.Div(inflation, supply)
	return inflation.Int64()
}
