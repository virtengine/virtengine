// Package types contains types for the Benchmark module.
//
// VE-602: Reliability score types
package types

import (
	"fmt"
	"time"
)

// ScoreVersion is the current version of the scoring algorithm
const ScoreVersion = "1.0.0"

// FixedPointScale is the scale factor for fixed-point arithmetic (1e6)
const FixedPointScale int64 = 1000000

// ReliabilityScoreInputs contains the inputs for reliability score computation
type ReliabilityScoreInputs struct {
	// BenchmarkSummary is the average benchmark summary score (0-10000)
	BenchmarkSummary int64 `json:"benchmark_summary"`

	// ProvisioningSuccessRate is success rate (0-1000000, fixed-point)
	ProvisioningSuccessRate int64 `json:"provisioning_success_rate"`

	// ProvisioningAttempts is total provisioning attempts
	ProvisioningAttempts int64 `json:"provisioning_attempts"`

	// ProvisioningSuccesses is successful provisioning count
	ProvisioningSuccesses int64 `json:"provisioning_successes"`

	// MeanTimeToProvision is average provision time in seconds
	MeanTimeToProvision int64 `json:"mean_time_to_provision"`

	// MeanTimeBetweenFailures is MTBF in seconds (0 means no failures)
	MeanTimeBetweenFailures int64 `json:"mean_time_between_failures"`

	// TotalUptimeSeconds is total uptime
	TotalUptimeSeconds int64 `json:"total_uptime_seconds"`

	// TotalDowntimeSeconds is total downtime
	TotalDowntimeSeconds int64 `json:"total_downtime_seconds"`

	// DisputeCount is number of disputes filed
	DisputeCount int64 `json:"dispute_count"`

	// DisputesResolved is disputes resolved in provider's favor
	DisputesResolved int64 `json:"disputes_resolved"`

	// DisputesLost is disputes resolved against provider
	DisputesLost int64 `json:"disputes_lost"`

	// AnomalyFlagCount is number of anomaly flags
	AnomalyFlagCount int64 `json:"anomaly_flag_count"`
}

// ReliabilityScore represents a provider's reliability score
type ReliabilityScore struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// Score is the overall reliability score (0-10000)
	Score int64 `json:"score"`

	// ScoreVersion is the version of the scoring algorithm used
	ScoreVersion string `json:"score_version"`

	// Inputs contains the score computation inputs
	Inputs ReliabilityScoreInputs `json:"inputs"`

	// ComponentScores contains individual component scores
	ComponentScores ComponentScores `json:"component_scores"`

	// UpdatedAt is when the score was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// BlockHeight is when the score was last updated on-chain
	BlockHeight int64 `json:"block_height"`
}

// ComponentScores contains individual component scores
type ComponentScores struct {
	// PerformanceScore from benchmarks (0-10000)
	PerformanceScore int64 `json:"performance_score"`

	// UptimeScore from uptime/MTBF (0-10000)
	UptimeScore int64 `json:"uptime_score"`

	// ProvisioningScore from success rate (0-10000)
	ProvisioningScore int64 `json:"provisioning_score"`

	// TrustScore from disputes/anomalies (0-10000)
	TrustScore int64 `json:"trust_score"`
}

// Validate validates the reliability score
func (s *ReliabilityScore) Validate() error {
	if s.ProviderAddress == "" {
		return fmt.Errorf("provider_address cannot be empty")
	}

	if s.Score < 0 || s.Score > 10000 {
		return fmt.Errorf("score out of bounds (0-10000)")
	}

	if s.ScoreVersion == "" {
		return fmt.Errorf("score_version cannot be empty")
	}

	return nil
}

// ComputeReliabilityScore computes the reliability score from inputs
// Uses deterministic fixed-point arithmetic for consensus compatibility
func ComputeReliabilityScore(inputs ReliabilityScoreInputs) (int64, ComponentScores) {
	// Weights for each component (must sum to 1000000)
	const (
		performanceWeight  int64 = 300000 // 30%
		uptimeWeight       int64 = 250000 // 25%
		provisioningWeight int64 = 250000 // 25%
		trustWeight        int64 = 200000 // 20%
	)

	// Calculate performance score (from benchmark summary)
	performanceScore := inputs.BenchmarkSummary
	if performanceScore < 0 {
		performanceScore = 0
	} else if performanceScore > 10000 {
		performanceScore = 10000
	}

	// Calculate uptime score
	uptimeScore := computeUptimeScore(inputs)

	// Calculate provisioning score
	provisioningScore := computeProvisioningScore(inputs)

	// Calculate trust score
	trustScore := computeTrustScore(inputs)

	// Compute weighted average using fixed-point math
	weightedSum := (performanceScore * performanceWeight) +
		(uptimeScore * uptimeWeight) +
		(provisioningScore * provisioningWeight) +
		(trustScore * trustWeight)

	// Divide by total weight (1000000)
	score := weightedSum / FixedPointScale

	// Clamp to valid range
	if score < 0 {
		score = 0
	} else if score > 10000 {
		score = 10000
	}

	components := ComponentScores{
		PerformanceScore:  performanceScore,
		UptimeScore:       uptimeScore,
		ProvisioningScore: provisioningScore,
		TrustScore:        trustScore,
	}

	return score, components
}

// computeUptimeScore calculates the uptime component score
func computeUptimeScore(inputs ReliabilityScoreInputs) int64 {
	totalTime := inputs.TotalUptimeSeconds + inputs.TotalDowntimeSeconds
	if totalTime == 0 {
		// No data, return neutral score
		return 5000
	}

	// Calculate uptime percentage using fixed-point
	uptimePercent := (inputs.TotalUptimeSeconds * FixedPointScale) / totalTime

	// Scale to 0-10000 range
	uptimeScore := (uptimePercent * 10000) / FixedPointScale

	// Bonus for high MTBF (cap at 1000 bonus points)
	mtbfBonus := int64(0)
	if inputs.MeanTimeBetweenFailures > 0 {
		// 1 point per hour of MTBF, capped at 1000
		mtbfBonus = inputs.MeanTimeBetweenFailures / 3600
		if mtbfBonus > 1000 {
			mtbfBonus = 1000
		}
	}

	score := uptimeScore + mtbfBonus
	if score > 10000 {
		score = 10000
	}

	return score
}

// computeProvisioningScore calculates the provisioning component score
func computeProvisioningScore(inputs ReliabilityScoreInputs) int64 {
	if inputs.ProvisioningAttempts == 0 {
		// No data, return neutral score
		return 5000
	}

	// Use provided success rate if available, otherwise compute
	successRate := inputs.ProvisioningSuccessRate
	if successRate == 0 && inputs.ProvisioningAttempts > 0 {
		successRate = (inputs.ProvisioningSuccesses * FixedPointScale) / inputs.ProvisioningAttempts
	}

	// Scale success rate (0-1000000) to score (0-10000)
	baseScore := (successRate * 10000) / FixedPointScale

	// Penalty for slow provisioning (target: under 300 seconds)
	speedPenalty := int64(0)
	if inputs.MeanTimeToProvision > 300 {
		// 1 point penalty per 10 seconds over 300, capped at 2000
		speedPenalty = (inputs.MeanTimeToProvision - 300) / 10
		if speedPenalty > 2000 {
			speedPenalty = 2000
		}
	}

	score := baseScore - speedPenalty
	if score < 0 {
		score = 0
	}

	return score
}

// computeTrustScore calculates the trust component score
func computeTrustScore(inputs ReliabilityScoreInputs) int64 {
	// Start with full trust score
	score := int64(10000)

	// Penalty for disputes
	if inputs.DisputeCount > 0 {
		// Weight disputes lost more heavily
		disputePenalty := (inputs.DisputesLost * 500) + (inputs.DisputeCount * 100)
		score -= disputePenalty
	}

	// Penalty for anomaly flags
	if inputs.AnomalyFlagCount > 0 {
		anomalyPenalty := inputs.AnomalyFlagCount * 200
		score -= anomalyPenalty
	}

	if score < 0 {
		score = 0
	}

	return score
}
