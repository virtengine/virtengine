package core

import (
	"context"
	"errors"
	"math"
	"math/big"
	"math/rand" //nolint:gosec // G404: deterministic simulation uses weak random for reproducibility, not security
	"strconv"
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
	"github.com/virtengine/virtengine/sim/agents"
	economicsim "github.com/virtengine/virtengine/sim/economics"
	"github.com/virtengine/virtengine/sim/markets"
	"github.com/virtengine/virtengine/sim/model"
)

// Engine runs economic simulations.
type Engine struct {
	config Config
	clock  *Clock
	events *EventQueue
	stats  *MetricsCollector

	rng *rand.Rand

	state model.State

	tokenModel      *economicsim.TokenModel
	inflationModel  economicsim.InflationModel
	escrowModel     *economicsim.EscrowModel
	settlementModel economicsim.SettlementModel
	stakingMarket   markets.StakingMarket

	agents []agents.Agent
}

// NewEngine creates a simulation engine.
func NewEngine(config Config) *Engine {
	return &Engine{config: config}
}

// Initialize prepares the simulation engine.
func (e *Engine) Initialize(ctx context.Context) error {
	if e.config.StartTime.IsZero() {
		e.config.StartTime = time.Now().UTC()
	}
	if e.config.EndTime.IsZero() {
		e.config.EndTime = e.config.StartTime.Add(365 * 24 * time.Hour)
	}
	if e.config.TimeStep == 0 {
		e.config.TimeStep = 24 * time.Hour
	}
	if e.config.Market.ComputeBasePrice == 0 {
		e.config.Market = markets.DefaultMarketParams()
	}
	if e.config.Tokenomics.InitialSupply == nil {
		e.config.Tokenomics = economics.DefaultTokenomicsParams()
	}
	if e.config.TokenPriceUSD == 0 {
		e.config.TokenPriceUSD = 1.0
	}

	e.clock = NewClock(e.config.StartTime, e.config.TimeStep)
	e.events = NewEventQueue()
	e.stats = NewMetricsCollector()
	e.rng = newDeterministicRNG(e.config.Seed)

	tokenSupply := new(big.Int).Set(e.config.Tokenomics.CurrentSupply)
	circulating := new(big.Int).Set(e.config.Tokenomics.CirculatingSupply)
	burned := new(big.Int).Set(e.config.Tokenomics.BurnedSupply)

	e.tokenModel = economicsim.NewTokenModel(tokenSupply, circulating, burned)
	e.escrowModel = economicsim.NewEscrowModel()
	e.inflationModel = economicsim.InflationModel{
		TargetInflationBPS: e.config.Tokenomics.TargetInflationBPS,
		MinInflationBPS:    e.config.Tokenomics.MinInflationBPS,
		MaxInflationBPS:    e.config.Tokenomics.MaxInflationBPS,
		TargetStakingBPS:   e.config.Tokenomics.TargetStakingRatioBPS,
	}
	e.settlementModel = economicsim.SettlementModel{FailureRate: 0.02}
	e.stakingMarket = markets.StakingMarket{TargetRatioBPS: e.config.Tokenomics.TargetStakingRatioBPS, AdjustmentBPS: 25}

	e.state = model.State{
		Time:          e.config.StartTime,
		BlockHeight:   0,
		TokenSupply:   new(big.Int).Set(tokenSupply),
		Staked:        new(big.Int).Set(e.config.Tokenomics.TotalStaked),
		Escrowed:      big.NewInt(0),
		Burned:        new(big.Int).Set(burned),
		Circulating:   new(big.Int).Set(circulating),
		TokenPriceUSD: e.config.TokenPriceUSD,
		Market:        markets.NewMarketState(e.config.Market),
	}

	return e.createAgents()
}

func (e *Engine) createAgents() error {
	rng := e.rng
	if rng == nil {
		return errors.New("rng not initialized")
	}

	e.agents = make([]agents.Agent, 0, e.config.NumUsers+e.config.NumProviders+e.config.NumValidators+e.config.NumAttackers)

	for i := 0; i < e.config.NumUsers; i++ {
		base := agents.NewBaseAgent("user-"+strconv.Itoa(i), agents.UserAgent, newDeterministicRNG(rng.Int63()))
		e.agents = append(e.agents, agents.NewUser(base.ID(), e.config.UserDemandMean, e.config.UserDemandStdDev, base))
	}

	for i := 0; i < e.config.NumProviders; i++ {
		capacity := e.config.ProviderCapacity * (0.5 + rng.Float64())
		base := agents.NewBaseAgent("provider-"+strconv.Itoa(i), agents.ProviderAgent, newDeterministicRNG(rng.Int63()))
		e.agents = append(e.agents, agents.NewProvider(base.ID(), capacity, base))
	}

	if e.config.NumValidators == 0 {
		e.config.NumValidators = 10
	}
	stakePerValidator := new(big.Int)
	if e.state.Staked.Sign() > 0 {
		stakePerValidator.Div(e.state.Staked, big.NewInt(int64(e.config.NumValidators)))
	}

	for i := 0; i < e.config.NumValidators; i++ {
		base := agents.NewBaseAgent("validator-"+strconv.Itoa(i), agents.ValidatorAgent, newDeterministicRNG(rng.Int63()))
		e.agents = append(e.agents, agents.NewValidator(base.ID(), stakePerValidator, base))
	}

	for i := 0; i < e.config.NumAttackers; i++ {
		budget := 100000 + rng.Float64()*250000
		base := agents.NewBaseAgent("attacker-"+strconv.Itoa(i), agents.AttackerAgent, newDeterministicRNG(rng.Int63()))
		e.agents = append(e.agents, agents.NewAttacker(base.ID(), budget, base))
	}

	return nil
}

// Run executes the simulation.
func (e *Engine) Run(ctx context.Context) (SimulationResult, error) {
	if e.clock == nil {
		if err := e.Initialize(ctx); err != nil {
			return SimulationResult{}, err
		}
	}

	initial := e.copyState()
	snapshots := make([]Snapshot, 0)
	steps := int64(0)
	periodsPerYear := int64(365 * 24 * time.Hour / e.config.TimeStep)
	if periodsPerYear <= 0 {
		periodsPerYear = 365
	}

	for e.state.Time.Before(e.config.EndTime) || e.state.Time.Equal(e.config.EndTime) {
		select {
		case <-ctx.Done():
			return SimulationResult{}, ctx.Err()
		default:
		}

		steps++
		e.collectEvents(ctx)
		e.processEvents()
		e.updateEconomics(periodsPerYear)
		e.updateMarkets()
		e.updateDerivedMetrics()

		snapshots = append(snapshots, e.buildSnapshot())
		e.stats.RecordStep(e.state)

		e.state.Time = e.clock.Step()
		e.state.BlockHeight += e.blocksPerStep()
	}

	final := e.copyState()

	result := SimulationResult{
		Scenario:  e.config.ScenarioName,
		Duration:  e.config.EndTime.Sub(e.config.StartTime),
		Start:     e.config.StartTime,
		End:       e.config.EndTime,
		Initial:   initial,
		Final:     final,
		Snapshots: snapshots,
		Metrics:   e.stats.Finalize(initial, final),
	}

	return result, nil
}

func (e *Engine) collectEvents(ctx context.Context) {
	for _, agent := range e.agents {
		events, err := agent.Step(ctx, &e.state)
		if err != nil {
			continue
		}
		for _, ev := range events {
			e.events.Push(ev)
		}
	}
}

func (e *Engine) processEvents() {
	events := e.events.Drain()
	if len(events) == 0 {
		return
	}

	for _, ev := range events {
		switch ev.Type {
		case model.EventDemand:
			data, ok := ev.Data.(model.DemandEvent)
			if ok {
				e.state.Market.ComputeDemand += data.Action.ComputeDemand
				e.state.Market.StorageDemand += data.Action.StorageDemand
				e.state.Market.GPUDemand += data.Action.GPUDemand
				e.state.Market.GasDemand += data.Action.GasDemand
			}
		case model.EventSupply:
			data, ok := ev.Data.(model.SupplyEvent)
			if ok {
				e.state.Market.ComputeSupply += data.Action.ComputeSupply
				e.state.Market.StorageSupply += data.Action.StorageSupply
				e.state.Market.GPUSupply += data.Action.GPUSupply
			}
		case model.EventStake:
			data, ok := ev.Data.(model.StakeEvent)
			if ok {
				e.state.Staked.Add(e.state.Staked, data.Amount)
				e.tokenModel.Lock(data.Amount)
			}
		case model.EventUnstake:
			data, ok := ev.Data.(model.StakeEvent)
			if ok {
				e.state.Staked.Sub(e.state.Staked, data.Amount)
				if e.state.Staked.Sign() < 0 {
					e.state.Staked.SetInt64(0)
				}
				e.tokenModel.Unlock(data.Amount)
			}
		case model.EventEscrowLock:
			data, ok := ev.Data.(model.EscrowEvent)
			if ok {
				e.escrowModel.Lock(data.Amount)
				e.tokenModel.Lock(data.Amount)
				e.state.Escrowed.Add(e.state.Escrowed, data.Amount)
			}
		case model.EventEscrowRelease:
			data, ok := ev.Data.(model.EscrowEvent)
			if ok {
				e.escrowModel.Release(data.Amount)
				e.tokenModel.Unlock(data.Amount)
				e.state.Escrowed.Sub(e.state.Escrowed, data.Amount)
				if e.state.Escrowed.Sign() < 0 {
					e.state.Escrowed.SetInt64(0)
				}
			}
		case model.EventSettlement:
			data, ok := ev.Data.(model.SettlementEvent)
			if ok {
				if !data.Success {
					e.stats.RecordSettlementFailure()
				}
			}
		case model.EventSlash:
			data, ok := ev.Data.(model.SlashEvent)
			if ok {
				e.state.Staked.Sub(e.state.Staked, data.Amount)
				e.tokenModel.Burn(data.Amount)
				e.state.Burned.Add(e.state.Burned, data.Amount)
			}
		case model.EventPriceManipulate:
			data, ok := ev.Data.(model.AttackEvent)
			if ok {
				e.stats.RecordAttack(data.CostUSD, 0, 0.1, 0.4, 0)
				e.state.Market.ComputePrice *= 1.05
				e.state.Market.GPUPrice *= 1.08
			}
		case model.EventSybil:
			data, ok := ev.Data.(model.AttackEvent)
			if ok {
				e.stats.RecordAttack(data.CostUSD, 0.6, 0.1, 0.1, 0)
			}
		case model.EventProviderExit:
			data, ok := ev.Data.(model.ProviderExitEvent)
			if ok {
				e.state.Market.ComputeSupply -= data.Capacity
				e.state.Market.StorageSupply -= data.Capacity * 0.8
				e.state.Market.GPUSupply -= data.Capacity * 0.4
				e.stats.RecordProviderExit()
			}
		case model.EventMEV:
			data, ok := ev.Data.(model.MEVEvent)
			if ok {
				e.stats.RecordAttack(float64(data.ExtractedFees.Int64())*e.state.TokenPriceUSD, 0, 0, 0.1, 0.5)
			}
		}
	}
}

func (e *Engine) updateEconomics(periodsPerYear int64) {
	// Inflation adjustment.
	e.state.StakingRatio = e.calculateStakingRatio()
	e.state.InflationBPS = e.inflationModel.AdjustInflation(e.state.StakingRatio)

	mint := e.inflationModel.MintForPeriod(e.tokenModel.Supply, e.state.InflationBPS, periodsPerYear)
	e.tokenModel.Mint(mint)

	e.state.TokenSupply = new(big.Int).Set(e.tokenModel.Supply)
	e.state.Circulating = new(big.Int).Set(e.tokenModel.Circulating)
	e.state.Burned = new(big.Int).Set(e.tokenModel.Burned)

	// APR estimate.
	e.state.APR = e.calculateAPR()
	// adjust staking ratio tendency
	e.state.StakingRatio = e.stakingMarket.UpdateRatio(e.state.StakingRatio, e.state.APR)
}

func (e *Engine) updateMarkets() {
	e.state.Market = markets.UpdateCompute(e.state.Market, e.config.Market)
	e.state.Market = markets.UpdateStorage(e.state.Market, e.config.Market)
	e.state.Market = markets.UpdateGPU(e.state.Market, e.config.Market)
	e.state.Market = markets.UpdateGas(e.state.Market, e.config.Market)

	e.state.Market.Utilization = utilization(e.state.Market)

	var feeResult markets.FeeResult
	e.state.Market, feeResult = markets.ApplyFees(e.state.Market, e.config.Market)
	feeBurn := big.NewInt(int64(math.Round(feeResult.Burned)))
	e.stats.RecordFeeBurned(feeBurn)
	e.tokenModel.Burn(feeBurn)

	// escrow settlement based on usage costs.
	escrowCost := int64(math.Round((e.state.Market.ComputeDemand*e.state.Market.ComputePrice +
		e.state.Market.StorageDemand*e.state.Market.StoragePrice +
		e.state.Market.GPUDemand*e.state.Market.GPUPrice)))
	if escrowCost > 0 {
		amount := big.NewInt(escrowCost)
		e.escrowModel.Lock(amount)
		e.tokenModel.Lock(amount)
		e.state.Escrowed.Add(e.state.Escrowed, amount)

		success, settled := e.settlementModel.Settle(amount, e.escrowModel)
		e.events.Push(model.Event{Type: model.EventSettlement, Timestamp: e.state.Time, Data: model.SettlementEvent{Amount: settled, Success: success}})
		if !success {
			e.stats.RecordEscrowUnderfunded()
		}
		if success {
			e.escrowModel.Release(amount)
			e.tokenModel.Unlock(amount)
			e.state.Escrowed.Sub(e.state.Escrowed, amount)
		}
	}
}

func (e *Engine) updateDerivedMetrics() {
	e.state.TokenVelocity = e.calculateVelocity()
}

func (e *Engine) calculateStakingRatio() int64 {
	if e.state.TokenSupply.Sign() == 0 {
		return 0
	}
	ratio := new(big.Int).Mul(e.state.Staked, big.NewInt(10000))
	ratio.Div(ratio, e.state.TokenSupply)
	return ratio.Int64()
}

func (e *Engine) calculateAPR() int64 {
	if e.state.StakingRatio == 0 {
		return 0
	}
	return (e.state.InflationBPS * 10000) / e.state.StakingRatio
}

func (e *Engine) calculateVelocity() float64 {
	if e.state.Circulating.Sign() == 0 {
		return 0
	}
	volume := e.state.Market.GasDemand + e.state.Market.ComputeDemand + e.state.Market.StorageDemand
	return volume / float64(e.state.Circulating.Int64())
}

func (e *Engine) buildSnapshot() Snapshot {
	return Snapshot{
		Time:          e.state.Time,
		BlockHeight:   e.state.BlockHeight,
		TokenSupply:   new(big.Int).Set(e.state.TokenSupply),
		Staked:        new(big.Int).Set(e.state.Staked),
		Escrowed:      new(big.Int).Set(e.state.Escrowed),
		Burned:        new(big.Int).Set(e.state.Burned),
		StakingRatio:  e.state.StakingRatio,
		InflationBPS:  e.state.InflationBPS,
		APR:           e.state.APR,
		TokenVelocity: e.state.TokenVelocity,
		Market:        e.state.Market,
	}
}

func (e *Engine) blocksPerStep() int64 {
	if e.config.Tokenomics.BlocksPerYear == 0 {
		return int64(e.config.TimeStep.Seconds() / 5)
	}
	yearSeconds := float64(365 * 24 * time.Hour / time.Second)
	secondsPerBlock := yearSeconds / float64(e.config.Tokenomics.BlocksPerYear)
	blocks := int64(math.Round(e.config.TimeStep.Seconds() / secondsPerBlock))
	if blocks < 1 {
		blocks = 1
	}
	return blocks
}

func (e *Engine) copyState() model.State {
	copyState := e.state
	copyState.TokenSupply = new(big.Int).Set(e.state.TokenSupply)
	copyState.Staked = new(big.Int).Set(e.state.Staked)
	copyState.Escrowed = new(big.Int).Set(e.state.Escrowed)
	copyState.Burned = new(big.Int).Set(e.state.Burned)
	copyState.Circulating = new(big.Int).Set(e.state.Circulating)
	return copyState
}

func utilization(state markets.MarketState) float64 {
	if state.ComputeSupply <= 0 {
		return 0
	}
	return math.Min(1, state.ComputeDemand/state.ComputeSupply)
}

func newDeterministicRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed)) // #nosec G404 -- deterministic simulation RNG for reproducibility
}
