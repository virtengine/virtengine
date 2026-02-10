package analysis

import (
	"context"
	"math"
	"math/rand" //nolint:gosec // G404: Monte Carlo simulation uses weak random for reproducibility, not security
	"sort"
	"sync"

	"github.com/virtengine/virtengine/sim/core"
)

// ParameterRange defines Monte Carlo variation range.
type ParameterRange struct {
	Min  float64
	Max  float64
	Dist string // uniform, normal, lognormal
}

// MonteCarloConfig configures Monte Carlo runs.
type MonteCarloConfig struct {
	Runs         int
	Parallelism  int
	Confidence   float64
	ParamRanges  map[string]ParameterRange
	BaseConfig   core.Config
	ScenarioName string
}

// MonteCarloResult summarizes distribution statistics.
type MonteCarloResult struct {
	Metric          string
	Mean            float64
	StdDev          float64
	Median          float64
	ConfidenceLower float64
	ConfidenceUpper float64
	Percentile5     float64
	Percentile95    float64
	Distribution    []float64
}

// MonteCarloAnalyzer runs multiple simulations.
type MonteCarloAnalyzer struct {
	config MonteCarloConfig
}

// NewMonteCarloAnalyzer builds an analyzer.
func NewMonteCarloAnalyzer(config MonteCarloConfig) *MonteCarloAnalyzer {
	if config.Parallelism <= 0 {
		config.Parallelism = 4
	}
	if config.Confidence <= 0 {
		config.Confidence = 0.95
	}
	return &MonteCarloAnalyzer{config: config}
}

// Run executes Monte Carlo simulations.
func (a *MonteCarloAnalyzer) Run(ctx context.Context) (map[string]MonteCarloResult, error) {
	runs := make([]map[string]float64, a.config.Runs)
	sem := make(chan struct{}, a.config.Parallelism)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i := 0; i < a.config.Runs; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}

			config := a.perturbConfig(idx)
			engine := core.NewEngine(config)
			if err := engine.Initialize(ctx); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			result, err := engine.Run(ctx)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			runs[idx] = extractMetrics(result)
		}()
	}

	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	return aggregateMetrics(runs, a.config.Confidence), nil
}

func (a *MonteCarloAnalyzer) perturbConfig(seed int) core.Config {
	config := a.config.BaseConfig
	config.Seed = int64(seed + 100)
	config.ScenarioName = a.config.ScenarioName

	// #nosec G404 -- deterministic simulation rng
	rng := rand.New(rand.NewSource(int64(seed) + 99))

	for param, prange := range a.config.ParamRanges {
		value := sampleRange(rng, prange)
		applyParam(&config, param, value)
	}

	return config
}

func sampleRange(rng *rand.Rand, prange ParameterRange) float64 {
	switch prange.Dist {
	case "normal":
		mean := (prange.Min + prange.Max) / 2
		std := (prange.Max - prange.Min) / 4
		value := rng.NormFloat64()*std + mean
		if value < prange.Min {
			return prange.Min
		}
		if value > prange.Max {
			return prange.Max
		}
		return value
	case "lognormal":
		logMean := math.Log((prange.Min + prange.Max) / 2)
		logStd := math.Log(prange.Max/prange.Min) / 4
		value := math.Exp(rng.NormFloat64()*logStd + logMean)
		if value < prange.Min {
			return prange.Min
		}
		if value > prange.Max {
			return prange.Max
		}
		return value
	default:
		return prange.Min + rng.Float64()*(prange.Max-prange.Min)
	}
}

func applyParam(config *core.Config, param string, value float64) {
	switch param {
	case "inflation_target_bps":
		config.Tokenomics.TargetInflationBPS = int64(value)
	case "staking_target_bps":
		config.Tokenomics.TargetStakingRatioBPS = int64(value)
	case "base_compute_price":
		config.Market.ComputeBasePrice = value
	case "base_storage_price":
		config.Market.StorageBasePrice = value
	case "base_gpu_price":
		config.Market.GPUBasePrice = value
	case "base_gas_price":
		config.Market.GasBasePrice = value
	case "user_demand_mean":
		config.UserDemandMean = value
	case "user_demand_stddev":
		config.UserDemandStdDev = value
	case "token_price_usd":
		config.TokenPriceUSD = value
	}
}

func extractMetrics(result core.SimulationResult) map[string]float64 {
	metrics := make(map[string]float64)
	metrics["avg_inflation_bps"] = float64(result.Metrics.AverageInflationBPS)
	metrics["avg_staking_bps"] = float64(result.Metrics.AverageStakingBPS)
	metrics["avg_apr_bps"] = float64(result.Metrics.AverageAPR)
	metrics["supply_growth_bps"] = float64(result.Metrics.SupplyGrowthBPS)
	metrics["avg_velocity"] = result.Metrics.AverageVelocity
	metrics["avg_compute_price"] = result.Metrics.AvgComputePrice
	metrics["avg_storage_price"] = result.Metrics.AvgStoragePrice
	metrics["avg_gpu_price"] = result.Metrics.AvgGPUPrice
	metrics["avg_gas_price"] = result.Metrics.AvgGasPrice
	metrics["attack_cost_usd"] = result.Metrics.AttackCostUSD
	metrics["sybil_risk"] = result.Metrics.SybilRiskScore
	metrics["collusion_risk"] = result.Metrics.CollusionRisk
	metrics["manipulation_risk"] = result.Metrics.ManipulationRisk
	metrics["mev_risk"] = result.Metrics.MEVRisk
	metrics["settlement_failures"] = float64(result.Metrics.SettlementFailures)
	metrics["escrow_underfunded"] = float64(result.Metrics.EscrowUnderfunded)
	return metrics
}

func aggregateMetrics(runs []map[string]float64, confidence float64) map[string]MonteCarloResult {
	metricNames := make(map[string]struct{})
	for _, run := range runs {
		for name := range run {
			metricNames[name] = struct{}{}
		}
	}

	results := make(map[string]MonteCarloResult)
	alpha := 1 - confidence

	for name := range metricNames {
		values := make([]float64, 0, len(runs))
		for _, run := range runs {
			if value, ok := run[name]; ok {
				values = append(values, value)
			}
		}
		if len(values) == 0 {
			continue
		}
		sorted := make([]float64, len(values))
		copy(sorted, values)
		sort.Float64s(sorted)
		mean := mean(values)
		std := stddev(values, mean)
		median := percentile(sorted, 0.5)

		result := MonteCarloResult{
			Metric:          name,
			Mean:            mean,
			StdDev:          std,
			Median:          median,
			ConfidenceLower: percentile(sorted, alpha/2),
			ConfidenceUpper: percentile(sorted, 1-alpha/2),
			Percentile5:     percentile(sorted, 0.05),
			Percentile95:    percentile(sorted, 0.95),
			Distribution:    values,
		}
		results[name] = result
	}

	return results
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stddev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	varSum := 0.0
	for _, v := range values {
		delta := v - mean
		varSum += delta * delta
	}
	return math.Sqrt(varSum / float64(len(values)))
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := p * float64(len(sorted)-1)
	low := int(math.Floor(idx))
	high := int(math.Ceil(idx))
	if low == high {
		return sorted[low]
	}
	weight := idx - float64(low)
	return sorted[low]*(1-weight) + sorted[high]*weight
}
