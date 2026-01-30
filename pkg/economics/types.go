package economics

import (
	"math/big"
	"time"
)

// TokenomicsParams contains core economic parameters for the network.
type TokenomicsParams struct {
	// Token Supply
	InitialSupply      *big.Int `json:"initial_supply"`
	MaxSupply          *big.Int `json:"max_supply"`
	CurrentSupply      *big.Int `json:"current_supply"`
	CirculatingSupply  *big.Int `json:"circulating_supply"`
	LockedSupply       *big.Int `json:"locked_supply"`
	BurnedSupply       *big.Int `json:"burned_supply"`

	// Inflation/Deflation
	InflationRateBPS   int64 `json:"inflation_rate_bps"`   // basis points (1 = 0.01%)
	TargetInflationBPS int64 `json:"target_inflation_bps"`
	MinInflationBPS    int64 `json:"min_inflation_bps"`
	MaxInflationBPS    int64 `json:"max_inflation_bps"`

	// Staking
	StakingRatioBPS       int64    `json:"staking_ratio_bps"`
	TargetStakingRatioBPS int64    `json:"target_staking_ratio_bps"`
	TotalStaked           *big.Int `json:"total_staked"`
	UnbondingPeriodDays   int64    `json:"unbonding_period_days"`

	// Rewards
	BaseRewardPerBlock    int64 `json:"base_reward_per_block"`
	VEIDRewardPool        int64 `json:"veid_reward_pool"`
	IdentityNetworkPool   int64 `json:"identity_network_pool"`
	BlocksPerYear         int64 `json:"blocks_per_year"`

	// Fees
	DefaultTakeRateBPS    int64              `json:"default_take_rate_bps"`
	DenomTakeRates        map[string]int64   `json:"denom_take_rates"`
	MinGasPrice           int64              `json:"min_gas_price"`

	// Governance
	ProposalDepositMin    *big.Int `json:"proposal_deposit_min"`
	VotingPeriodDays      int64    `json:"voting_period_days"`
	QuorumBPS             int64    `json:"quorum_bps"`
}

// DefaultTokenomicsParams returns sensible default parameters.
func DefaultTokenomicsParams() TokenomicsParams {
	return TokenomicsParams{
		InitialSupply:         big.NewInt(1_000_000_000_000_000), // 1B tokens with 6 decimals
		MaxSupply:             big.NewInt(10_000_000_000_000_000), // 10B max
		CurrentSupply:         big.NewInt(1_000_000_000_000_000),
		CirculatingSupply:     big.NewInt(500_000_000_000_000),
		LockedSupply:          big.NewInt(500_000_000_000_000),
		BurnedSupply:          big.NewInt(0),
		InflationRateBPS:      700,  // 7%
		TargetInflationBPS:    700,
		MinInflationBPS:       100,  // 1%
		MaxInflationBPS:       2000, // 20%
		StakingRatioBPS:       6700, // 67%
		TargetStakingRatioBPS: 6700,
		TotalStaked:           big.NewInt(335_000_000_000_000),
		UnbondingPeriodDays:   21,
		BaseRewardPerBlock:    1000000,      // 1 token per block
		VEIDRewardPool:        10000000000,  // 10k tokens
		IdentityNetworkPool:   5000000000,   // 5k tokens
		BlocksPerYear:         6_311_520,    // ~5s blocks
		DefaultTakeRateBPS:    400,          // 4%
		DenomTakeRates:        map[string]int64{"uvirt": 400},
		MinGasPrice:           100,
		ProposalDepositMin:    big.NewInt(10_000_000_000), // 10k tokens
		VotingPeriodDays:      14,
		QuorumBPS:             3340, // 33.4%
	}
}

// NetworkState represents the current state of the network for simulation.
type NetworkState struct {
	BlockHeight       int64     `json:"block_height"`
	Timestamp         time.Time `json:"timestamp"`
	TotalSupply       *big.Int  `json:"total_supply"`
	TotalStaked       *big.Int  `json:"total_staked"`
	TotalBurned       *big.Int  `json:"total_burned"`
	ActiveValidators  int64     `json:"active_validators"`
	TotalDelegators   int64     `json:"total_delegators"`
	TransactionVolume *big.Int  `json:"transaction_volume"`
	FeesCollected     *big.Int  `json:"fees_collected"`
	RewardsDistributed *big.Int `json:"rewards_distributed"`
}

// ValidatorState represents a validator's economic state.
type ValidatorState struct {
	Address           string   `json:"address"`
	SelfStake         *big.Int `json:"self_stake"`
	DelegatedStake    *big.Int `json:"delegated_stake"`
	TotalStake        *big.Int `json:"total_stake"`
	Commission        int64    `json:"commission_bps"`
	VotingPower       int64    `json:"voting_power_bps"`
	UptimeScore       int64    `json:"uptime_score"`
	VEIDVerifications int64    `json:"veid_verifications"`
}

// SimulationResult contains results from an economic simulation.
type SimulationResult struct {
	Scenario          string                 `json:"scenario"`
	Duration          time.Duration          `json:"duration"`
	InitialState      NetworkState           `json:"initial_state"`
	FinalState        NetworkState           `json:"final_state"`
	Snapshots         []NetworkStateSnapshot `json:"snapshots"`
	Metrics           SimulationMetrics      `json:"metrics"`
	Recommendations   []Recommendation       `json:"recommendations"`
}

// NetworkStateSnapshot is a point-in-time state snapshot.
type NetworkStateSnapshot struct {
	BlockHeight     int64    `json:"block_height"`
	TotalSupply     *big.Int `json:"total_supply"`
	StakingRatioBPS int64    `json:"staking_ratio_bps"`
	InflationBPS    int64    `json:"inflation_bps"`
	APR             int64    `json:"apr_bps"`
}

// SimulationMetrics contains aggregated metrics from simulation.
type SimulationMetrics struct {
	AvgInflationBPS      int64   `json:"avg_inflation_bps"`
	AvgStakingRatioBPS   int64   `json:"avg_staking_ratio_bps"`
	AvgAPR               int64   `json:"avg_apr_bps"`
	SupplyGrowthBPS      int64   `json:"supply_growth_bps"`
	GiniCoefficient      float64 `json:"gini_coefficient"`
	NakamotoCoefficient  int64   `json:"nakamoto_coefficient"`
	AttackCostUSD        float64 `json:"attack_cost_usd"`
	SecurityScore        int64   `json:"security_score"`
}

// Recommendation represents an optimization recommendation.
type Recommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Action      string `json:"action"`
}

// DistributionMetrics contains token distribution analysis.
type DistributionMetrics struct {
	GiniCoefficient     float64   `json:"gini_coefficient"`
	LorenzCurve         []float64 `json:"lorenz_curve"`
	Top10HoldingsBPS    int64     `json:"top10_holdings_bps"`
	Top100HoldingsBPS   int64     `json:"top100_holdings_bps"`
	NakamotoCoefficient int64     `json:"nakamoto_coefficient"`
	HerfindahlIndex     float64   `json:"herfindahl_index"`
	MinValidatorsFor51  int64     `json:"min_validators_for_51"`
}

// AttackAnalysis contains attack cost analysis results.
type AttackAnalysis struct {
	AttackType          string   `json:"attack_type"`
	CostEstimateUSD     float64  `json:"cost_estimate_usd"`
	TokensRequired      *big.Int `json:"tokens_required"`
	PercentageOfSupply  float64  `json:"percentage_of_supply"`
	TimeToPrepare       string   `json:"time_to_prepare"`
	DetectionDifficulty string   `json:"detection_difficulty"`
	MitigationStrategy  string   `json:"mitigation_strategy"`
	RiskLevel           string   `json:"risk_level"`
}

// GameTheoryAnalysis contains game theory analysis results.
type GameTheoryAnalysis struct {
	Scenario           string             `json:"scenario"`
	Players            []string           `json:"players"`
	Strategies         map[string][]string `json:"strategies"`
	NashEquilibrium    string             `json:"nash_equilibrium"`
	PayoffMatrix       [][]float64        `json:"payoff_matrix"`
	DominantStrategy   string             `json:"dominant_strategy"`
	IncentiveAlignment string             `json:"incentive_alignment"`
	Recommendations    []string           `json:"recommendations"`
}

// EconomicSecurityAudit contains results from economic security audit.
type EconomicSecurityAudit struct {
	Timestamp           time.Time               `json:"timestamp"`
	OverallScore        int64                   `json:"overall_score"`
	InflationAnalysis   InflationAnalysis       `json:"inflation_analysis"`
	StakingAnalysis     StakingAnalysis         `json:"staking_analysis"`
	FeeMarketAnalysis   FeeMarketAnalysis       `json:"fee_market_analysis"`
	DistributionMetrics DistributionMetrics     `json:"distribution_metrics"`
	AttackAnalyses      []AttackAnalysis        `json:"attack_analyses"`
	GameTheoryAnalyses  []GameTheoryAnalysis    `json:"game_theory_analyses"`
	Vulnerabilities     []Vulnerability         `json:"vulnerabilities"`
	Recommendations     []Recommendation        `json:"recommendations"`
}

// InflationAnalysis contains inflation-specific analysis.
type InflationAnalysis struct {
	CurrentRateBPS      int64   `json:"current_rate_bps"`
	ProjectedRateBPS    int64   `json:"projected_rate_bps"`
	YearlyMintedTokens  *big.Int `json:"yearly_minted_tokens"`
	SustainabilityScore int64   `json:"sustainability_score"`
	Trend               string  `json:"trend"` // "inflationary", "deflationary", "stable"
}

// StakingAnalysis contains staking-specific analysis.
type StakingAnalysis struct {
	CurrentRatioBPS     int64    `json:"current_ratio_bps"`
	OptimalRatioBPS     int64    `json:"optimal_ratio_bps"`
	CurrentAPR          int64    `json:"current_apr_bps"`
	OptimalAPR          int64    `json:"optimal_apr_bps"`
	ValidatorCount      int64    `json:"validator_count"`
	DelegatorCount      int64    `json:"delegator_count"`
	ConcentrationRisk   string   `json:"concentration_risk"`
	UnbondingPressure   float64  `json:"unbonding_pressure"`
}

// FeeMarketAnalysis contains fee market analysis.
type FeeMarketAnalysis struct {
	AverageFeeBPS       int64   `json:"average_fee_bps"`
	MedianFeeBPS        int64   `json:"median_fee_bps"`
	FeeVolatility       float64 `json:"fee_volatility"`
	RevenueEstimateYear *big.Int `json:"revenue_estimate_year"`
	MarketEfficiency    float64 `json:"market_efficiency"`
	SpamResistance      int64   `json:"spam_resistance_score"`
}

// Vulnerability represents an identified economic vulnerability.
type Vulnerability struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Mitigation  string `json:"mitigation"`
	Status      string `json:"status"` // "open", "mitigated", "accepted"
}
