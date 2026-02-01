// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

// Package slo provides SLO (Service Level Objective) definitions and verification
// for chaos experiments in VirtEngine.
package slo

import (
	"time"

	"github.com/virtengine/virtengine/pkg/chaos"
)

// ResilienceBaseline defines target metrics for system resilience.
type ResilienceBaseline struct {
	// Name is the identifier for this baseline.
	Name string `json:"name"`

	// Description provides context for this baseline.
	Description string `json:"description"`

	// MTTR is the target Mean Time To Recovery.
	MTTR time.Duration `json:"mttr"`

	// MTTD is the target Mean Time To Detect.
	MTTD time.Duration `json:"mttd"`

	// RecoverySuccessRate is the minimum acceptable recovery success rate (0.0-1.0).
	RecoverySuccessRate float64 `json:"recovery_success_rate"`

	// AvailabilityTarget is the minimum acceptable availability (0.0-1.0).
	AvailabilityTarget float64 `json:"availability_target"`

	// ErrorBudgetMinutes is the maximum allowed downtime per month.
	ErrorBudgetMinutes float64 `json:"error_budget_minutes"`

	// BlastRadiusLimit is the maximum percentage of resources that should be affected.
	BlastRadiusLimit float64 `json:"blast_radius_limit"`
}

// ChainResilienceBaseline returns the baseline for blockchain operations.
func ChainResilienceBaseline() ResilienceBaseline {
	return ResilienceBaseline{
		Name:                "chain-resilience",
		Description:         "Resilience baseline for blockchain consensus and block production",
		MTTR:                30 * time.Second,
		MTTD:                5 * time.Second,
		RecoverySuccessRate: 0.99,
		AvailabilityTarget:  0.999,
		ErrorBudgetMinutes:  43.2, // 99.9% availability = 43.2 min/month
		BlastRadiusLimit:    0.33, // Max 1/3 of validators affected
	}
}

// IdentityScoringResilienceBaseline returns the baseline for VEID scoring.
func IdentityScoringResilienceBaseline() ResilienceBaseline {
	return ResilienceBaseline{
		Name:                "identity-scoring-resilience",
		Description:         "Resilience baseline for VEID identity scoring service",
		MTTR:                2 * time.Minute,
		MTTD:                30 * time.Second,
		RecoverySuccessRate: 0.95,
		AvailabilityTarget:  0.995,
		ErrorBudgetMinutes:  216, // 99.5% availability = 3.6 hrs/month
		BlastRadiusLimit:    0.50,
	}
}

// MarketplaceResilienceBaseline returns the baseline for marketplace operations.
func MarketplaceResilienceBaseline() ResilienceBaseline {
	return ResilienceBaseline{
		Name:                "marketplace-resilience",
		Description:         "Resilience baseline for marketplace order fulfillment",
		MTTR:                5 * time.Minute,
		MTTD:                1 * time.Minute,
		RecoverySuccessRate: 0.95,
		AvailabilityTarget:  0.995,
		ErrorBudgetMinutes:  216,
		BlastRadiusLimit:    0.50,
	}
}

// HPCResilienceBaseline returns the baseline for HPC job scheduling.
func HPCResilienceBaseline() ResilienceBaseline {
	return ResilienceBaseline{
		Name:                "hpc-resilience",
		Description:         "Resilience baseline for HPC job scheduling and execution",
		MTTR:                10 * time.Minute,
		MTTD:                2 * time.Minute,
		RecoverySuccessRate: 0.90,
		AvailabilityTarget:  0.99,
		ErrorBudgetMinutes:  432, // 99.0% availability = 7.2 hrs/month
		BlastRadiusLimit:    0.50,
	}
}

// ProviderDaemonResilienceBaseline returns the baseline for provider daemon.
func ProviderDaemonResilienceBaseline() ResilienceBaseline {
	return ResilienceBaseline{
		Name:                "provider-daemon-resilience",
		Description:         "Resilience baseline for provider daemon health and bidding",
		MTTR:                1 * time.Minute,
		MTTD:                30 * time.Second,
		RecoverySuccessRate: 0.95,
		AvailabilityTarget:  0.995,
		ErrorBudgetMinutes:  216,
		BlastRadiusLimit:    0.50,
	}
}

// AllResilienceBaselines returns all defined resilience baselines.
func AllResilienceBaselines() []ResilienceBaseline {
	return []ResilienceBaseline{
		ChainResilienceBaseline(),
		IdentityScoringResilienceBaseline(),
		MarketplaceResilienceBaseline(),
		HPCResilienceBaseline(),
		ProviderDaemonResilienceBaseline(),
	}
}

// SLODefinition defines an SLO with probes for verification.
type SLODefinition struct {
	// Name is the identifier for this SLO.
	Name string `json:"name"`

	// Description provides context for this SLO.
	Description string `json:"description"`

	// Target is the target value for this SLO (e.g., 0.999 for 99.9%).
	Target float64 `json:"target"`

	// Window is the measurement window for this SLO.
	Window time.Duration `json:"window"`

	// Probes are the checks to verify this SLO.
	Probes []chaos.Probe `json:"probes"`
}

// ChainAvailabilitySLO returns the SLO for chain availability.
func ChainAvailabilitySLO() SLODefinition {
	return SLODefinition{
		Name:        "SLO-CHAIN-001",
		Description: "Chain produces blocks 99.9% of the time",
		Target:      0.999,
		Window:      30 * 24 * time.Hour, // 30-day rolling
		Probes: []chaos.Probe{
			{
				Type:            chaos.ProbeTypePrometheus,
				Name:            "block-production-rate",
				Query:           "1 - (sum(rate(tendermint_consensus_rounds_total{result=\"timeout\"}[5m])) / sum(rate(tendermint_consensus_rounds_total[5m])))",
				Interval:        30 * time.Second,
				Timeout:         10 * time.Second,
				SuccessCriteria: "value >= 0.999",
			},
		},
	}
}

// BlockTimeSLO returns the SLO for block time P99.
func BlockTimeSLO() SLODefinition {
	return SLODefinition{
		Name:        "SLO-CHAIN-002",
		Description: "99th percentile block time under 6 seconds",
		Target:      6.0,
		Window:      1 * time.Hour,
		Probes: []chaos.Probe{
			{
				Type:            chaos.ProbeTypePrometheus,
				Name:            "block-time-p99",
				Query:           "histogram_quantile(0.99, sum(rate(tendermint_consensus_block_time_seconds_bucket[1h])) by (le))",
				Interval:        1 * time.Minute,
				Timeout:         10 * time.Second,
				SuccessCriteria: "value <= 6.0",
			},
		},
	}
}

// VEIDScoringSLO returns the SLO for identity scoring processing time.
func VEIDScoringSLO() SLODefinition {
	return SLODefinition{
		Name:        "SLO-VEID-001",
		Description: "Identity scores computed within 5 minutes (P95)",
		Target:      300.0, // 5 minutes in seconds
		Window:      24 * time.Hour,
		Probes: []chaos.Probe{
			{
				Type:            chaos.ProbeTypePrometheus,
				Name:            "veid-scoring-latency-p95",
				Query:           "histogram_quantile(0.95, sum(rate(veid_scoring_duration_seconds_bucket[24h])) by (le))",
				Interval:        5 * time.Minute,
				Timeout:         10 * time.Second,
				SuccessCriteria: "value <= 300",
			},
		},
	}
}

// MarketOrderFulfillmentSLO returns the SLO for marketplace order fulfillment.
func MarketOrderFulfillmentSLO() SLODefinition {
	return SLODefinition{
		Name:        "SLO-MARKET-001",
		Description: "Orders receive allocation within 10 minutes (P95)",
		Target:      600.0, // 10 minutes in seconds
		Window:      24 * time.Hour,
		Probes: []chaos.Probe{
			{
				Type:            chaos.ProbeTypePrometheus,
				Name:            "market-order-fulfillment-p95",
				Query:           "histogram_quantile(0.95, sum(rate(market_order_fulfillment_seconds_bucket[24h])) by (le))",
				Interval:        5 * time.Minute,
				Timeout:         10 * time.Second,
				SuccessCriteria: "value <= 600",
			},
		},
	}
}

// HPCSchedulingSLO returns the SLO for HPC job scheduling.
func HPCSchedulingSLO() SLODefinition {
	return SLODefinition{
		Name:        "SLO-HPC-001",
		Description: "Jobs scheduled within 15 minutes (P95)",
		Target:      900.0, // 15 minutes in seconds
		Window:      24 * time.Hour,
		Probes: []chaos.Probe{
			{
				Type:            chaos.ProbeTypePrometheus,
				Name:            "hpc-scheduling-latency-p95",
				Query:           "histogram_quantile(0.95, sum(rate(hpc_job_scheduling_seconds_bucket[24h])) by (le))",
				Interval:        5 * time.Minute,
				Timeout:         10 * time.Second,
				SuccessCriteria: "value <= 900",
			},
		},
	}
}

// AllSLODefinitions returns all defined SLOs.
func AllSLODefinitions() []SLODefinition {
	return []SLODefinition{
		ChainAvailabilitySLO(),
		BlockTimeSLO(),
		VEIDScoringSLO(),
		MarketOrderFulfillmentSLO(),
		HPCSchedulingSLO(),
	}
}

// BuildSteadyStateHypothesis creates a SteadyStateHypothesis from an SLO definition.
func (s *SLODefinition) BuildSteadyStateHypothesis() *chaos.SteadyStateHypothesis {
	return &chaos.SteadyStateHypothesis{
		Name:      s.Name,
		Title:     s.Description,
		Probes:    s.Probes,
		Tolerance: 0.05, // 5% default tolerance
	}
}

