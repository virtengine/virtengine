package core

import (
	"math/big"

	"github.com/virtengine/virtengine/sim/model"
)

// MetricsCollector aggregates simulation metrics over time.
type MetricsCollector struct {
	steps int64

	inflationSum int64
	stakingSum   int64
	aprSum       int64
	velocitySum  float64

	computePriceSum float64
	storagePriceSum float64
	gpuPriceSum     float64
	gasPriceSum     float64

	feeBurned *big.Int

	settlementFailures int64
	escrowUnderfunded  int64
	providerExits      int64

	attackCostUSD    float64
	sybilRiskSum     float64
	collusionRiskSum float64
	manipRiskSum     float64
	mevRiskSum       float64
}

// NewMetricsCollector creates a metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{feeBurned: big.NewInt(0)}
}

// RecordStep records metrics for a simulation step.
func (m *MetricsCollector) RecordStep(state model.State) {
	m.steps++
	m.inflationSum += state.InflationBPS
	m.stakingSum += state.StakingRatio
	m.aprSum += state.APR
	m.velocitySum += state.TokenVelocity

	m.computePriceSum += state.Market.ComputePrice
	m.storagePriceSum += state.Market.StoragePrice
	m.gpuPriceSum += state.Market.GPUPrice
	m.gasPriceSum += state.Market.GasPrice
}

// RecordFeeBurned accumulates fee burn totals.
func (m *MetricsCollector) RecordFeeBurned(amount *big.Int) {
	if amount == nil {
		return
	}
	m.feeBurned.Add(m.feeBurned, amount)
}

// RecordAttack aggregates attack risks.
func (m *MetricsCollector) RecordAttack(costUSD, sybil, collusion, manip, mev float64) {
	m.attackCostUSD += costUSD
	m.sybilRiskSum += sybil
	m.collusionRiskSum += collusion
	m.manipRiskSum += manip
	m.mevRiskSum += mev
}

// RecordSettlementFailure tracks settlement issues.
func (m *MetricsCollector) RecordSettlementFailure() {
	m.settlementFailures++
}

// RecordEscrowUnderfunded tracks escrow underfunding incidents.
func (m *MetricsCollector) RecordEscrowUnderfunded() {
	m.escrowUnderfunded++
}

// RecordProviderExit tracks provider exits.
func (m *MetricsCollector) RecordProviderExit() {
	m.providerExits++
}

// Finalize builds aggregated metrics.
func (m *MetricsCollector) Finalize(initial, final model.State) Metrics {
	metrics := Metrics{FeeBurned: new(big.Int).Set(m.feeBurned)}

	if m.steps == 0 {
		return metrics
	}
	metrics.AverageInflationBPS = m.inflationSum / m.steps
	metrics.AverageStakingBPS = m.stakingSum / m.steps
	metrics.AverageAPR = m.aprSum / m.steps
	metrics.AverageVelocity = m.velocitySum / float64(m.steps)

	metrics.AvgComputePrice = m.computePriceSum / float64(m.steps)
	metrics.AvgStoragePrice = m.storagePriceSum / float64(m.steps)
	metrics.AvgGPUPrice = m.gpuPriceSum / float64(m.steps)
	metrics.AvgGasPrice = m.gasPriceSum / float64(m.steps)

	metrics.SettlementFailures = m.settlementFailures
	metrics.EscrowUnderfunded = m.escrowUnderfunded
	metrics.ProviderExits = m.providerExits

	metrics.AttackCostUSD = m.attackCostUSD
	metrics.SybilRiskScore = m.sybilRiskSum / float64(m.steps)
	metrics.CollusionRisk = m.collusionRiskSum / float64(m.steps)
	metrics.ManipulationRisk = m.manipRiskSum / float64(m.steps)
	metrics.MEVRisk = m.mevRiskSum / float64(m.steps)

	if initial.TokenSupply != nil && final.TokenSupply != nil && initial.TokenSupply.Sign() > 0 {
		delta := new(big.Int).Sub(final.TokenSupply, initial.TokenSupply)
		deltaBPS := new(big.Int).Mul(delta, big.NewInt(10000))
		deltaBPS.Div(deltaBPS, initial.TokenSupply)
		metrics.SupplyGrowthBPS = deltaBPS.Int64()
	}

	return metrics
}
