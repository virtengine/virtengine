# Task 30D: Tokenomics Simulation & Economic Validation

**vibe-kanban ID:** `21cc7d01-e500-48cb-9914-42dc83f626de`

## Problem Statement

Before mainnet launch, the VirtEngine economic model requires:
- Validation that token economics are sustainable
- Simulation of various market conditions
- Stress testing of escrow and settlement mechanisms
- Analysis of attack vectors (economic attacks, gaming)
- Validation of staking reward curves
- Fee market analysis

Without rigorous simulation:
- Token value could collapse due to poor economics
- Inflationary/deflationary pressures could destabilize network
- Providers could game the system for unfair advantage
- Users could face unpredictable costs

## Acceptance Criteria

### AC-1: Simulation Framework
- [ ] Agent-based simulation engine
- [ ] Market dynamics simulator
- [ ] Monte Carlo analysis support
- [ ] Parameter sweep automation
- [ ] Results visualization dashboard

### AC-2: Token Supply Model
- [ ] Genesis distribution simulation
- [ ] Inflation/deflation curves
- [ ] Vesting schedule impacts
- [ ] Staking ratio equilibrium
- [ ] Token velocity analysis

### AC-3: Fee Market Analysis
- [ ] Gas price dynamics
- [ ] Compute pricing equilibrium
- [ ] Storage pricing model
- [ ] GPU pricing model
- [ ] Fee burn mechanisms

### AC-4: Escrow & Settlement Validation
- [ ] Settlement accuracy under load
- [ ] Escrow underfunding scenarios
- [ ] Dispute resolution economics
- [ ] Slashing impact analysis
- [ ] Provider exit scenarios

### AC-5: Attack Vector Analysis
- [ ] Sybil attack costs
- [ ] Provider collusion scenarios
- [ ] Market manipulation attacks
- [ ] Griefing attack economics
- [ ] MEV extraction analysis

### AC-6: Economic Reports
- [ ] Mainnet parameter recommendations
- [ ] Risk assessment document
- [ ] Sensitivity analysis
- [ ] Long-term sustainability report
- [ ] Investor-facing economics paper

## Technical Requirements

### Simulation Framework Architecture

```
sim/
├── core/
│   ├── engine.go           # Simulation engine
│   ├── clock.go            # Time management
│   ├── events.go           # Event queue
│   └── metrics.go          # Metrics collection
├── agents/
│   ├── types.go            # Agent interfaces
│   ├── user.go             # User agent
│   ├── provider.go         # Provider agent
│   ├── validator.go        # Validator agent
│   ├── arbitrageur.go      # Market arbitrageur
│   └── attacker.go         # Adversarial agents
├── markets/
│   ├── compute.go          # Compute market model
│   ├── storage.go          # Storage market model
│   ├── staking.go          # Staking market
│   └── fee.go              # Fee market
├── economics/
│   ├── token.go            # Token supply model
│   ├── inflation.go        # Inflation curves
│   ├── escrow.go           # Escrow model
│   └── settlement.go       # Settlement model
├── scenarios/
│   ├── baseline.go         # Normal operation
│   ├── bull_market.go      # High demand
│   ├── bear_market.go      # Low demand
│   ├── attack.go           # Attack scenarios
│   └── black_swan.go       # Extreme events
├── analysis/
│   ├── monte_carlo.go      # Monte Carlo analysis
│   ├── sensitivity.go      # Sensitivity analysis
│   └── visualization.go    # Chart generation
└── cmd/
    ├── simulate/           # Simulation CLI
    └── analyze/            # Analysis CLI
```

### Simulation Engine Core

```go
// sim/core/engine.go
package core

import (
    "context"
    "sync"
    "time"

    "github.com/virtengine/virtengine/sim/agents"
    "github.com/virtengine/virtengine/sim/markets"
)

// Config holds simulation configuration
type Config struct {
    // Time settings
    StartTime    time.Time
    EndTime      time.Time
    TimeStep     time.Duration
    
    // Agent counts
    NumUsers     int
    NumProviders int
    NumValidators int
    
    // Economic parameters
    InitialSupply     sdk.Int
    InflationRate     sdk.Dec
    StakingRatio      sdk.Dec
    MinGasPrice       sdk.Dec
    
    // Market parameters
    ComputePriceMin   sdk.Dec
    ComputePriceMax   sdk.Dec
    StoragePriceMin   sdk.Dec
    StoragePriceMax   sdk.Dec
    
    // Behavioral parameters
    UserDemandMean    float64
    UserDemandStdDev  float64
    ProviderCapacity  float64
    
    // Random seed for reproducibility
    Seed int64
}

// Engine runs economic simulations
type Engine struct {
    config     Config
    clock      *Clock
    events     *EventQueue
    metrics    *MetricsCollector
    
    // State
    tokenSupply    sdk.Int
    totalStaked    sdk.Int
    totalEscrowed  sdk.Int
    
    // Agents
    users      []agents.User
    providers  []agents.Provider
    validators []agents.Validator
    attackers  []agents.Attacker
    
    // Markets
    computeMarket *markets.ComputeMarket
    storageMarket *markets.StorageMarket
    stakingMarket *markets.StakingMarket
    feeMarket     *markets.FeeMarket
    
    mu sync.RWMutex
}

// NewEngine creates a simulation engine
func NewEngine(config Config) *Engine {
    return &Engine{
        config:  config,
        clock:   NewClock(config.StartTime, config.TimeStep),
        events:  NewEventQueue(),
        metrics: NewMetricsCollector(),
    }
}

// Initialize sets up the simulation
func (e *Engine) Initialize(ctx context.Context) error {
    // Initialize token supply
    e.tokenSupply = e.config.InitialSupply
    
    // Create agents
    if err := e.createAgents(ctx); err != nil {
        return err
    }
    
    // Initialize markets
    e.computeMarket = markets.NewComputeMarket(e.config)
    e.storageMarket = markets.NewStorageMarket(e.config)
    e.stakingMarket = markets.NewStakingMarket(e.config)
    e.feeMarket = markets.NewFeeMarket(e.config)
    
    return nil
}

// Run executes the simulation
func (e *Engine) Run(ctx context.Context) (*SimulationResult, error) {
    for e.clock.Now().Before(e.config.EndTime) {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        // Process events
        events := e.events.PopEventsUntil(e.clock.Now())
        for _, event := range events {
            if err := e.processEvent(event); err != nil {
                return nil, err
            }
        }
        
        // Agent decisions
        if err := e.runAgentDecisions(ctx); err != nil {
            return nil, err
        }
        
        // Market clearing
        e.computeMarket.Clear(e.clock.Now())
        e.storageMarket.Clear(e.clock.Now())
        
        // Token economics
        e.applyInflation()
        e.processSettlements()
        
        // Collect metrics
        e.collectMetrics()
        
        // Advance time
        e.clock.Advance()
    }
    
    return e.buildResult(), nil
}

func (e *Engine) runAgentDecisions(ctx context.Context) error {
    // Users decide on deployments
    for _, user := range e.users {
        action := user.Decide(e.getMarketState())
        e.events.Push(action)
    }
    
    // Providers decide on bids and capacity
    for _, provider := range e.providers {
        actions := provider.Decide(e.getMarketState())
        for _, action := range actions {
            e.events.Push(action)
        }
    }
    
    // Validators decide on staking
    for _, validator := range e.validators {
        actions := validator.Decide(e.getMarketState())
        for _, action := range actions {
            e.events.Push(action)
        }
    }
    
    // Attackers execute strategies
    for _, attacker := range e.attackers {
        actions := attacker.Attack(e.getMarketState())
        for _, action := range actions {
            e.events.Push(action)
        }
    }
    
    return nil
}

func (e *Engine) applyInflation() {
    // Calculate inflation for this time step
    stepInflation := e.config.InflationRate.
        MulInt64(int64(e.config.TimeStep)).
        QuoInt64(int64(365 * 24 * time.Hour))
    
    inflationAmount := e.tokenSupply.ToDec().Mul(stepInflation).TruncateInt()
    
    // Distribute to stakers
    if !e.totalStaked.IsZero() {
        e.distributeToStakers(inflationAmount)
    }
    
    e.tokenSupply = e.tokenSupply.Add(inflationAmount)
    
    e.metrics.Record("token_supply", e.tokenSupply.Int64())
    e.metrics.Record("inflation_issued", inflationAmount.Int64())
}
```

### Agent Models

```go
// sim/agents/user.go
package agents

import (
    "math/rand"

    "github.com/virtengine/virtengine/sim/core"
)

// UserBehavior defines user decision patterns
type UserBehavior struct {
    PriceSensitivity float64 // How much price affects demand
    QualityWeight    float64 // Weight given to provider reputation
    LoyaltyFactor    float64 // Tendency to stay with same provider
    ChurnProbability float64 // Probability of leaving the network
}

// User represents a compute consumer
type User struct {
    ID       string
    Balance  sdk.Int
    Behavior UserBehavior
    
    // State
    activeLeases  []*Lease
    demandProfile DemandProfile
    rng           *rand.Rand
}

// Decide makes deployment decisions based on market state
func (u *User) Decide(state MarketState) *core.Event {
    // Check if user churns
    if u.rng.Float64() < u.Behavior.ChurnProbability {
        return u.exitMarket()
    }
    
    // Calculate demand
    demand := u.calculateDemand(state)
    if demand.IsZero() {
        return nil
    }
    
    // Find best provider
    provider := u.selectProvider(state, demand)
    if provider == nil {
        return nil // No suitable provider
    }
    
    // Check if price acceptable
    price := state.ComputePrice
    maxPrice := u.maxAcceptablePrice(demand)
    if price.GT(maxPrice) {
        return nil // Too expensive
    }
    
    // Create deployment order
    return &core.Event{
        Type:      core.EventCreateOrder,
        Timestamp: state.Now,
        Data: OrderEvent{
            UserID:     u.ID,
            ProviderID: provider.ID,
            Resources:  demand,
            MaxPrice:   maxPrice,
        },
    }
}

func (u *User) selectProvider(state MarketState, demand ResourceSpec) *ProviderInfo {
    var bestProvider *ProviderInfo
    var bestScore float64 = -1
    
    for _, p := range state.ActiveProviders {
        if !p.HasCapacity(demand) {
            continue
        }
        
        // Score based on price, quality, and loyalty
        priceScore := 1.0 - (p.Price.MustFloat64() / state.MaxPrice.MustFloat64())
        qualityScore := p.Reputation
        loyaltyScore := 0.0
        if u.hasLeasedFrom(p.ID) {
            loyaltyScore = u.Behavior.LoyaltyFactor
        }
        
        score := u.Behavior.PriceSensitivity*priceScore +
            u.Behavior.QualityWeight*qualityScore +
            loyaltyScore
        
        if score > bestScore {
            bestScore = score
            bestProvider = &p
        }
    }
    
    return bestProvider
}

// sim/agents/provider.go
package agents

// ProviderStrategy defines provider behavior
type ProviderStrategy struct {
    PricingModel      string  // "cost_plus", "market", "aggressive"
    CapacityTarget    float64 // Target utilization (0-1)
    ReputationWeight  float64 // Investment in quality
    RiskTolerance     float64 // Willingness to take risky jobs
}

// Provider represents a compute provider
type Provider struct {
    ID         string
    Capacity   ResourceSpec
    FixedCosts sdk.Dec
    VarCosts   sdk.Dec
    Strategy   ProviderStrategy
    
    // State
    utilization float64
    reputation  float64
    revenue     sdk.Int
    stake       sdk.Int
}

// Decide makes bidding and capacity decisions
func (p *Provider) Decide(state MarketState) []*core.Event {
    var events []*core.Event
    
    // Adjust pricing
    newPrice := p.calculatePrice(state)
    if !newPrice.Equal(p.currentPrice) {
        events = append(events, &core.Event{
            Type:      core.EventUpdatePrice,
            Timestamp: state.Now,
            Data: PriceUpdateEvent{
                ProviderID: p.ID,
                NewPrice:   newPrice,
            },
        })
    }
    
    // Bid on open orders
    for _, order := range state.OpenOrders {
        if p.shouldBid(order, state) {
            events = append(events, p.createBid(order, state))
        }
    }
    
    // Adjust capacity (scale up/down)
    capacityChange := p.calculateCapacityAdjustment(state)
    if !capacityChange.IsZero() {
        events = append(events, &core.Event{
            Type:      core.EventAdjustCapacity,
            Timestamp: state.Now,
            Data: CapacityEvent{
                ProviderID: p.ID,
                Delta:      capacityChange,
            },
        })
    }
    
    return events
}

func (p *Provider) calculatePrice(state MarketState) sdk.Dec {
    switch p.Strategy.PricingModel {
    case "cost_plus":
        // Cost plus margin
        margin := sdk.NewDecWithPrec(20, 2) // 20%
        return p.VarCosts.Mul(sdk.OneDec().Add(margin))
        
    case "market":
        // Follow market price
        return state.ComputePrice
        
    case "aggressive":
        // Undercut market to gain share
        discount := sdk.NewDecWithPrec(10, 2) // 10% below market
        return state.ComputePrice.Mul(sdk.OneDec().Sub(discount))
        
    default:
        return state.ComputePrice
    }
}

// sim/agents/attacker.go
package agents

// AttackType defines different attack strategies
type AttackType string

const (
    AttackSybil       AttackType = "sybil"
    AttackCollusion   AttackType = "collusion"
    AttackGriefing    AttackType = "griefing"
    AttackManipulation AttackType = "manipulation"
)

// Attacker simulates adversarial behavior
type Attacker struct {
    ID          string
    Type        AttackType
    Budget      sdk.Int
    Identities  []string // For sybil attacks
    Colluders   []string // For collusion
}

// Attack executes attack strategies
func (a *Attacker) Attack(state MarketState) []*core.Event {
    switch a.Type {
    case AttackSybil:
        return a.sybilAttack(state)
    case AttackCollusion:
        return a.collusionAttack(state)
    case AttackGriefing:
        return a.griefingAttack(state)
    case AttackManipulation:
        return a.manipulationAttack(state)
    default:
        return nil
    }
}

func (a *Attacker) sybilAttack(state MarketState) []*core.Event {
    var events []*core.Event
    
    // Create fake identities if budget allows
    identityCost := state.VerificationCost
    numNewIdentities := a.Budget.Quo(identityCost).Int64()
    
    for i := int64(0); i < numNewIdentities; i++ {
        events = append(events, &core.Event{
            Type: core.EventCreateIdentity,
            Data: SybilIdentityEvent{
                AttackerID: a.ID,
                IdentityID: fmt.Sprintf("%s-sybil-%d", a.ID, i),
                Cost:       identityCost,
            },
        })
    }
    
    return events
}

func (a *Attacker) collusionAttack(state MarketState) []*core.Event {
    // Coordinate prices among colluding providers
    var events []*core.Event
    
    collusionPrice := state.ComputePrice.MulInt64(2) // Double the price
    
    for _, colluder := range a.Colluders {
        events = append(events, &core.Event{
            Type: core.EventCollusionSignal,
            Data: CollusionEvent{
                LeaderID:    a.ID,
                FollowerID:  colluder,
                TargetPrice: collusionPrice,
            },
        })
    }
    
    return events
}
```

### Monte Carlo Analysis

```go
// sim/analysis/monte_carlo.go
package analysis

import (
    "context"
    "math"
    "sync"

    "github.com/virtengine/virtengine/sim/core"
    "gonum.org/v1/gonum/stat"
)

// MonteCarloConfig holds analysis configuration
type MonteCarloConfig struct {
    NumRuns           int
    ConfidenceLevel   float64 // e.g., 0.95 for 95% CI
    Parallelism       int
    ParameterRanges   map[string]ParameterRange
}

// ParameterRange defines parameter variation
type ParameterRange struct {
    Min  float64
    Max  float64
    Dist string // "uniform", "normal", "lognormal"
}

// MonteCarloResult holds analysis results
type MonteCarloResult struct {
    Metric           string
    Mean             float64
    StdDev           float64
    Median           float64
    ConfidenceLower  float64
    ConfidenceUpper  float64
    Percentile5      float64
    Percentile95     float64
    Distribution     []float64
}

// MonteCarloAnalyzer runs Monte Carlo simulations
type MonteCarloAnalyzer struct {
    baseConfig core.Config
    mcConfig   MonteCarloConfig
}

// NewMonteCarloAnalyzer creates an analyzer
func NewMonteCarloAnalyzer(base core.Config, mc MonteCarloConfig) *MonteCarloAnalyzer {
    return &MonteCarloAnalyzer{
        baseConfig: base,
        mcConfig:   mc,
    }
}

// Run executes Monte Carlo analysis
func (a *MonteCarloAnalyzer) Run(ctx context.Context) (map[string]*MonteCarloResult, error) {
    results := make([]map[string]float64, a.mcConfig.NumRuns)
    
    // Run simulations in parallel
    var wg sync.WaitGroup
    sem := make(chan struct{}, a.mcConfig.Parallelism)
    errChan := make(chan error, 1)
    
    for i := 0; i < a.mcConfig.NumRuns; i++ {
        wg.Add(1)
        go func(runIndex int) {
            defer wg.Done()
            
            select {
            case <-ctx.Done():
                return
            case sem <- struct{}{}:
                defer func() { <-sem }()
            }
            
            // Generate perturbed config
            config := a.perturbConfig(runIndex)
            
            // Run simulation
            engine := core.NewEngine(config)
            if err := engine.Initialize(ctx); err != nil {
                select {
                case errChan <- err:
                default:
                }
                return
            }
            
            result, err := engine.Run(ctx)
            if err != nil {
                select {
                case errChan <- err:
                default:
                }
                return
            }
            
            results[runIndex] = result.Metrics
        }(i)
    }
    
    wg.Wait()
    
    select {
    case err := <-errChan:
        return nil, err
    default:
    }
    
    // Aggregate results
    return a.aggregateResults(results), nil
}

func (a *MonteCarloAnalyzer) aggregateResults(runs []map[string]float64) map[string]*MonteCarloResult {
    aggregated := make(map[string]*MonteCarloResult)
    
    // Collect all metric names
    metrics := make(map[string]bool)
    for _, run := range runs {
        for metric := range run {
            metrics[metric] = true
        }
    }
    
    // Compute statistics for each metric
    for metric := range metrics {
        values := make([]float64, 0, len(runs))
        for _, run := range runs {
            if v, ok := run[metric]; ok {
                values = append(values, v)
            }
        }
        
        if len(values) == 0 {
            continue
        }
        
        mean := stat.Mean(values, nil)
        stddev := stat.StdDev(values, nil)
        
        sorted := make([]float64, len(values))
        copy(sorted, values)
        sort.Float64s(sorted)
        
        n := len(sorted)
        alpha := 1 - a.mcConfig.ConfidenceLevel
        
        aggregated[metric] = &MonteCarloResult{
            Metric:           metric,
            Mean:             mean,
            StdDev:           stddev,
            Median:           sorted[n/2],
            ConfidenceLower:  sorted[int(float64(n)*alpha/2)],
            ConfidenceUpper:  sorted[int(float64(n)*(1-alpha/2))],
            Percentile5:      sorted[int(float64(n)*0.05)],
            Percentile95:     sorted[int(float64(n)*0.95)],
            Distribution:     values,
        }
    }
    
    return aggregated
}

func (a *MonteCarloAnalyzer) perturbConfig(seed int) core.Config {
    config := a.baseConfig
    config.Seed = int64(seed)
    
    rng := rand.New(rand.NewSource(int64(seed)))
    
    for param, prange := range a.mcConfig.ParameterRanges {
        var value float64
        
        switch prange.Dist {
        case "uniform":
            value = prange.Min + rng.Float64()*(prange.Max-prange.Min)
        case "normal":
            mean := (prange.Min + prange.Max) / 2
            stddev := (prange.Max - prange.Min) / 4
            value = rng.NormFloat64()*stddev + mean
            value = math.Max(prange.Min, math.Min(prange.Max, value))
        case "lognormal":
            logMean := math.Log((prange.Min + prange.Max) / 2)
            logStd := math.Log(prange.Max/prange.Min) / 4
            value = math.Exp(rng.NormFloat64()*logStd + logMean)
        }
        
        config = a.setParameter(config, param, value)
    }
    
    return config
}
```

### Visualization Dashboard

```go
// sim/analysis/visualization.go
package analysis

import (
    "encoding/json"
    "html/template"
    "net/http"
)

// Dashboard serves the visualization interface
type Dashboard struct {
    results   map[string]*MonteCarloResult
    scenarios map[string]*ScenarioResult
    port      int
}

// Serve starts the dashboard server
func (d *Dashboard) Serve() error {
    http.HandleFunc("/", d.handleIndex)
    http.HandleFunc("/api/results", d.handleResults)
    http.HandleFunc("/api/scenarios", d.handleScenarios)
    http.HandleFunc("/api/sensitivity", d.handleSensitivity)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    return http.ListenAndServe(fmt.Sprintf(":%d", d.port), nil)
}

func (d *Dashboard) handleResults(w http.ResponseWriter, r *http.Request) {
    // Format for charting
    chartData := make(map[string]interface{})
    
    for metric, result := range d.results {
        chartData[metric] = map[string]interface{}{
            "mean":       result.Mean,
            "stddev":     result.StdDev,
            "ci_lower":   result.ConfidenceLower,
            "ci_upper":   result.ConfidenceUpper,
            "histogram":  buildHistogram(result.Distribution, 50),
        }
    }
    
    json.NewEncoder(w).Encode(chartData)
}
```

### CLI Commands

```go
// sim/cmd/simulate/main.go
package main

import (
    "context"
    "os"

    "github.com/spf13/cobra"
    "github.com/virtengine/virtengine/sim/core"
    "github.com/virtengine/virtengine/sim/analysis"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "ve-sim",
        Short: "VirtEngine economic simulation",
    }
    
    rootCmd.AddCommand(runCmd())
    rootCmd.AddCommand(analyzeCmd())
    rootCmd.AddCommand(dashboardCmd())
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func runCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "run [scenario]",
        Short: "Run a simulation scenario",
        Args:  cobra.ExactArgs(1),
        RunE:  runSimulation,
    }
    
    cmd.Flags().String("config", "", "Config file path")
    cmd.Flags().Int("seed", 42, "Random seed")
    cmd.Flags().Duration("duration", 365*24*time.Hour, "Simulation duration")
    cmd.Flags().String("output", "results.json", "Output file")
    
    return cmd
}

func analyzeCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "analyze",
        Short: "Run Monte Carlo analysis",
        RunE:  runAnalysis,
    }
    
    cmd.Flags().Int("runs", 1000, "Number of Monte Carlo runs")
    cmd.Flags().Int("parallelism", 8, "Parallel simulations")
    cmd.Flags().Float64("confidence", 0.95, "Confidence interval")
    
    return cmd
}

func dashboardCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "dashboard",
        Short: "Start visualization dashboard",
        RunE: func(cmd *cobra.Command, args []string) error {
            dashboard := analysis.NewDashboard(8080)
            return dashboard.Serve()
        },
    }
}
```

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `sim/core/engine.go` | Simulation engine | 600 |
| `sim/core/clock.go` | Time management | 100 |
| `sim/core/events.go` | Event queue | 200 |
| `sim/core/metrics.go` | Metrics collection | 250 |
| `sim/agents/user.go` | User agent | 400 |
| `sim/agents/provider.go` | Provider agent | 400 |
| `sim/agents/validator.go` | Validator agent | 300 |
| `sim/agents/attacker.go` | Adversarial agents | 350 |
| `sim/markets/*.go` | Market models | 600 |
| `sim/economics/*.go` | Economic models | 500 |
| `sim/scenarios/*.go` | Scenario definitions | 400 |
| `sim/analysis/monte_carlo.go` | Monte Carlo analysis | 400 |
| `sim/analysis/sensitivity.go` | Sensitivity analysis | 300 |
| `sim/analysis/visualization.go` | Dashboard | 400 |
| `sim/cmd/simulate/main.go` | CLI | 300 |
| `*_test.go` | Test files | 800 |
| `_docs/tokenomics-report.md` | Economic report | 1000 |

**Total Estimated:** 6,300 lines

## Validation Checklist

- [ ] Baseline scenario runs without errors
- [ ] Monte Carlo produces valid confidence intervals
- [ ] Token supply model matches expected curves
- [ ] Fee market reaches equilibrium
- [ ] Attack scenarios show economic cost
- [ ] Escrow underfunding is detected
- [ ] Dashboard visualizes all key metrics
- [ ] Sensitivity analysis identifies critical parameters
- [ ] Report includes mainnet recommendations
- [ ] Simulation is reproducible (same seed = same result)

## Dependencies

- None (independent analysis)

## Key Metrics to Track

1. **Token Economics**
   - Total supply over time
   - Inflation rate
   - Token velocity
   - Staking ratio

2. **Market Health**
   - Price stability
   - Provider utilization
   - User growth/churn
   - Order fill rate

3. **Attack Costs**
   - Sybil attack ROI
   - Collusion profitability
   - Griefing costs
   - Market manipulation impact

4. **System Stability**
   - Settlement accuracy
   - Escrow adequacy
   - Slashing events
   - Provider exits
