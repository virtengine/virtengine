package simulation

import (
	"math/big"
	"sort"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
)

const (
	blockRewardShareBPS  int64 = 8000
	uptimeRewardShareBPS int64 = 2000
	warningUptimeBPS     int64 = 9900
	minorSlashUptimeBPS  int64 = 9800
	minorSlashPenaltyBPS int64 = 50
	majorSlashPenaltyBPS int64 = 100
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
	for i := range results {
		results[i] = newValidatorRewardResult(validators[i].Address)
	}

	if epochBlocks <= 0 {
		return results
	}

	// Calculate total stake
	totalStake := big.NewInt(0)
	for _, v := range validators {
		totalStake.Add(totalStake, safeBigInt(v.TotalStake))
	}

	if totalStake.Sign() == 0 {
		return results
	}

	consensusPool := new(big.Int).Mul(big.NewInt(s.params.BaseRewardPerBlock), big.NewInt(epochBlocks))
	blockRewardPool := bpsMul(consensusPool, blockRewardShareBPS)
	uptimeRewardPool := bpsMul(consensusPool, uptimeRewardShareBPS)
	veidRewardPool := s.scaleAnnualPoolToEpoch(s.params.VEIDRewardPool, epochBlocks)

	blockRewards := distributePool(validators, blockRewardPool, func(v economics.ValidatorState) *big.Int {
		return new(big.Int).Set(safeBigInt(v.TotalStake))
	})
	uptimeRewards := distributePool(validators, uptimeRewardPool, func(v economics.ValidatorState) *big.Int {
		weight := new(big.Int).Set(safeBigInt(v.TotalStake))
		if weight.Sign() == 0 {
			return weight
		}
		weight.Mul(weight, big.NewInt(clampBPS(v.UptimeScore)))
		return weight
	})
	veidRewards := distributePool(validators, veidRewardPool, func(v economics.ValidatorState) *big.Int {
		if v.VEIDVerifications <= 0 {
			return big.NewInt(0)
		}
		return big.NewInt(v.VEIDVerifications)
	})

	for i, validator := range validators {
		results[i] = s.calculateValidatorReward(validator, epochBlocks, blockRewards[i], uptimeRewards[i], veidRewards[i])
	}

	return results
}

// ValidatorRewardResult contains calculated rewards for a validator.
type ValidatorRewardResult struct {
	Address             string   `json:"address"`
	GrossReward         *big.Int `json:"gross_reward"`
	BlockProposalReward *big.Int `json:"block_proposal_reward"`
	VEIDReward          *big.Int `json:"veid_reward"`
	UptimeReward        *big.Int `json:"uptime_reward"`
	SlashPenalty        *big.Int `json:"slash_penalty"`
	TotalReward         *big.Int `json:"total_reward"`
	CommissionEarned    *big.Int `json:"commission_earned"`
	DelegatorRewards    *big.Int `json:"delegator_rewards"`
	EffectiveAPR        int64    `json:"effective_apr_bps"`
}

// calculateValidatorReward calculates rewards for a single validator.
func (s *StakingSimulator) calculateValidatorReward(
	validator economics.ValidatorState,
	epochBlocks int64,
	blockReward *big.Int,
	uptimeReward *big.Int,
	veidReward *big.Int,
) ValidatorRewardResult {
	result := newValidatorRewardResult(validator.Address)
	validatorStake := safeBigInt(validator.TotalStake)

	if validatorStake.Sign() == 0 || epochBlocks <= 0 {
		return result
	}

	result.BlockProposalReward = new(big.Int).Set(blockReward)
	result.UptimeReward = new(big.Int).Set(uptimeReward)
	result.VEIDReward = new(big.Int).Set(veidReward)
	result.GrossReward = sumBigInts(result.BlockProposalReward, result.UptimeReward, result.VEIDReward)
	result.SlashPenalty = s.calculateEpochSlashPenalty(validatorStake, slashPenaltyBPS(clampBPS(validator.UptimeScore)), epochBlocks)

	result.TotalReward = new(big.Int).Sub(new(big.Int).Set(result.GrossReward), result.SlashPenalty)
	if result.TotalReward.Sign() < 0 {
		result.TotalReward = big.NewInt(0)
	}

	// Commission split
	result.CommissionEarned = new(big.Int).Mul(result.TotalReward, big.NewInt(clampBPS(validator.Commission)))
	result.CommissionEarned.Div(result.CommissionEarned, big.NewInt(10000))

	result.DelegatorRewards = new(big.Int).Sub(result.TotalReward, result.CommissionEarned)

	// Effective APR annualized to chain blocks-per-year.
	if s.params.BlocksPerYear > 0 {
		annualizedReward := new(big.Int).Mul(result.TotalReward, big.NewInt(s.params.BlocksPerYear))
		annualizedReward.Div(annualizedReward, big.NewInt(epochBlocks))
		apr := new(big.Int).Mul(annualizedReward, big.NewInt(10000))
		apr.Div(apr, validatorStake)
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
		totalStake.Add(totalStake, safeBigInt(v.TotalStake))
	}

	if totalStake.Sign() == 0 {
		return 0
	}

	annualRewards := big.NewInt(0)
	annualRewards.Add(annualRewards, new(big.Int).Mul(big.NewInt(s.params.BaseRewardPerBlock), big.NewInt(s.params.BlocksPerYear)))
	if s.params.VEIDRewardPool > 0 {
		annualRewards.Add(annualRewards, big.NewInt(s.params.VEIDRewardPool))
	}

	// APR = (Annual Rewards / Total Staked) * 10000
	apr := new(big.Int).Mul(annualRewards, big.NewInt(10000))
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

func newValidatorRewardResult(address string) ValidatorRewardResult {
	return ValidatorRewardResult{
		Address:             address,
		GrossReward:         big.NewInt(0),
		BlockProposalReward: big.NewInt(0),
		VEIDReward:          big.NewInt(0),
		UptimeReward:        big.NewInt(0),
		SlashPenalty:        big.NewInt(0),
		TotalReward:         big.NewInt(0),
		CommissionEarned:    big.NewInt(0),
		DelegatorRewards:    big.NewInt(0),
	}
}

func distributePool(
	validators []economics.ValidatorState,
	pool *big.Int,
	weightFn func(economics.ValidatorState) *big.Int,
) []*big.Int {
	shares := make([]*big.Int, len(validators))
	weights := make([]*big.Int, len(validators))
	totalWeight := big.NewInt(0)
	lastEligible := -1

	for i, validator := range validators {
		shares[i] = big.NewInt(0)
		weight := weightFn(validator)
		if weight == nil || weight.Sign() <= 0 {
			weights[i] = big.NewInt(0)
			continue
		}
		weights[i] = new(big.Int).Set(weight)
		totalWeight.Add(totalWeight, weight)
		lastEligible = i
	}

	if pool == nil || pool.Sign() <= 0 || totalWeight.Sign() == 0 || lastEligible == -1 {
		return shares
	}

	distributed := big.NewInt(0)
	for i, weight := range weights {
		if weight.Sign() == 0 {
			continue
		}

		if i == lastEligible {
			shares[i] = new(big.Int).Sub(pool, distributed)
			continue
		}

		share := new(big.Int).Mul(pool, weight)
		share.Div(share, totalWeight)
		shares[i] = share
		distributed.Add(distributed, share)
	}

	return shares
}

func (s *StakingSimulator) scaleAnnualPoolToEpoch(pool, epochBlocks int64) *big.Int {
	if pool <= 0 || epochBlocks <= 0 || s.params.BlocksPerYear <= 0 {
		return big.NewInt(0)
	}
	scaled := new(big.Int).Mul(big.NewInt(pool), big.NewInt(epochBlocks))
	scaled.Div(scaled, big.NewInt(s.params.BlocksPerYear))
	return scaled
}

func safeBigInt(value *big.Int) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}
	return value
}

func clampBPS(value int64) int64 {
	if value < 0 {
		return 0
	}
	if value > 10000 {
		return 10000
	}
	return value
}

func slashPenaltyBPS(uptimeBPS int64) int64 {
	if uptimeBPS >= warningUptimeBPS {
		return 0
	}
	if uptimeBPS >= minorSlashUptimeBPS {
		return minorSlashPenaltyBPS
	}
	return majorSlashPenaltyBPS
}

func (s *StakingSimulator) calculateEpochSlashPenalty(stake *big.Int, slashBPS, epochBlocks int64) *big.Int {
	if stake == nil || stake.Sign() == 0 || slashBPS <= 0 || epochBlocks <= 0 || s.params.BlocksPerYear <= 0 {
		return big.NewInt(0)
	}
	penalty := new(big.Int).Mul(stake, big.NewInt(slashBPS))
	penalty.Mul(penalty, big.NewInt(epochBlocks))
	penalty.Div(penalty, big.NewInt(10000))
	penalty.Div(penalty, big.NewInt(s.params.BlocksPerYear))
	return penalty
}

func bpsMul(value *big.Int, bps int64) *big.Int {
	if value == nil || value.Sign() == 0 || bps <= 0 {
		return big.NewInt(0)
	}
	result := new(big.Int).Mul(value, big.NewInt(bps))
	result.Div(result, big.NewInt(10000))
	return result
}

func sumBigInts(values ...*big.Int) *big.Int {
	total := big.NewInt(0)
	for _, value := range values {
		total.Add(total, safeBigInt(value))
	}
	return total
}
