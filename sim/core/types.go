package core

import (
	"math/big"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
	"github.com/virtengine/virtengine/sim/markets"
	"github.com/virtengine/virtengine/sim/model"
)

// Config holds the simulation configuration.
type Config struct {
	ScenarioName string

	StartTime time.Time
	EndTime   time.Time
	TimeStep  time.Duration
	Seed      int64

	NumUsers      int
	NumProviders  int
	NumValidators int
	NumAttackers  int

	Tokenomics economics.TokenomicsParams
	Market     markets.MarketParams

	UserDemandMean   float64
	UserDemandStdDev float64
	ProviderCapacity float64
	TokenPriceUSD    float64

	SlashingEnabled    bool
	SlashingPenaltyBPS int64
}

// Snapshot captures a point-in-time state.
type Snapshot struct {
	Time          time.Time
	BlockHeight   int64
	TokenSupply   *big.Int
	Staked        *big.Int
	Escrowed      *big.Int
	Burned        *big.Int
	StakingRatio  int64
	InflationBPS  int64
	APR           int64
	TokenVelocity float64
	Market        markets.MarketState
}

// SimulationResult contains the outcome of a run.
type SimulationResult struct {
	Scenario string
	Duration time.Duration
	Start    time.Time
	End      time.Time

	Initial model.State
	Final   model.State

	Snapshots []Snapshot
	Metrics   Metrics
}

// Metrics contains aggregated statistics for a run.
type Metrics struct {
	AverageInflationBPS int64
	AverageStakingBPS   int64
	AverageAPR          int64
	SupplyGrowthBPS     int64
	AverageVelocity     float64

	AvgComputePrice float64
	AvgStoragePrice float64
	AvgGPUPrice     float64
	AvgGasPrice     float64
	FeeBurned       *big.Int

	SettlementFailures int64
	EscrowUnderfunded  int64
	ProviderExits      int64

	AttackCostUSD    float64
	SybilRiskScore   float64
	CollusionRisk    float64
	ManipulationRisk float64
	MEVRisk          float64
}
