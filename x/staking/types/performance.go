// Package types contains types for the staking module.
//
// VE-921: Performance types for validator performance tracking
// This file provides utility methods for ValidatorPerformance (generated proto type).
package types

import (
	"fmt"
	"time"

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
)

// FixedPointScale is the scale factor for fixed-point arithmetic (1e6)
const FixedPointScale int64 = 1000000

// MaxPerformanceScore is the maximum performance score (10000 = 100%)
const MaxPerformanceScore int64 = 10000

// PerformanceWeight constants for reward calculation
const (
	WeightBlockProposal    int64 = 3000  // 30%
	WeightVEIDVerification int64 = 4000  // 40%
	WeightUptime           int64 = 3000  // 30%
	TotalWeight            int64 = 10000
)

// NewValidatorPerformance creates a new validator performance record
func NewValidatorPerformance(validatorAddr string, epochNumber uint64) *stakingv1.ValidatorPerformance {
	return &stakingv1.ValidatorPerformance{
		ValidatorAddress:       validatorAddr,
		EpochNumber:            epochNumber,
		VEIDVerificationScore:  MaxPerformanceScore, // Start with full score
		OverallScore:           0,
	}
}

// ValidateValidatorPerformance validates the validator performance record
func ValidateValidatorPerformance(vp *stakingv1.ValidatorPerformance) error {
	if vp.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	if vp.BlocksMissed < 0 {
		return fmt.Errorf("blocks_missed cannot be negative")
	}

	if vp.BlocksProposed < 0 {
		return fmt.Errorf("blocks_proposed cannot be negative")
	}

	if vp.OverallScore < 0 || vp.OverallScore > MaxPerformanceScore {
		return fmt.Errorf("overall_score must be between 0 and %d", MaxPerformanceScore)
	}

	return nil
}

// ComputeOverallScore computes the overall performance score deterministically
// This uses integer arithmetic to ensure determinism across all nodes
func ComputeOverallScore(vp *stakingv1.ValidatorPerformance) int64 {
	// Block proposal score (0-10000)
	var blockScore int64
	if vp.BlocksExpected > 0 {
		blockScore = (vp.BlocksProposed * MaxPerformanceScore) / vp.BlocksExpected
		if blockScore > MaxPerformanceScore {
			blockScore = MaxPerformanceScore
		}
	} else if vp.BlocksProposed > 0 {
		blockScore = MaxPerformanceScore
	}

	// VEID verification score (0-10000)
	veidScore := vp.VEIDVerificationScore
	if vp.VEIDVerificationsExpected > 0 {
		completionRate := (vp.VEIDVerificationsCompleted * MaxPerformanceScore) / vp.VEIDVerificationsExpected
		if completionRate > MaxPerformanceScore {
			completionRate = MaxPerformanceScore
		}
		// Combine completion rate with quality score
		veidScore = (completionRate + vp.VEIDVerificationScore) / 2
	}

	// Uptime score (0-10000)
	var uptimeScore int64
	totalTime := vp.UptimeSeconds + vp.DowntimeSeconds
	if totalTime > 0 {
		uptimeScore = (vp.UptimeSeconds * MaxPerformanceScore) / totalTime
	} else {
		uptimeScore = MaxPerformanceScore // Assume full uptime if no data
	}

	// Weighted average
	overallScore := (blockScore*WeightBlockProposal +
		veidScore*WeightVEIDVerification +
		uptimeScore*WeightUptime) / TotalWeight

	// Store the computed score in the struct
	vp.OverallScore = overallScore

	return overallScore
}

// GetUptimePercent returns the uptime percentage in fixed-point (1e6 scale)
func GetUptimePercent(vp *stakingv1.ValidatorPerformance) int64 {
	totalTime := vp.UptimeSeconds + vp.DowntimeSeconds
	if totalTime == 0 {
		return FixedPointScale // 100%
	}
	return (vp.UptimeSeconds * FixedPointScale) / totalTime
}

// ShouldSlashForDowntime checks if the validator should be slashed for downtime
func ShouldSlashForDowntime(vp *stakingv1.ValidatorPerformance, threshold int64) bool {
	return vp.ConsecutiveMissedBlocks >= threshold
}

// NewValidatorSigningInfo creates a new signing info record
func NewValidatorSigningInfo(validatorAddr string, startHeight int64) *stakingv1.ValidatorSigningInfo {
	return &stakingv1.ValidatorSigningInfo{
		ValidatorAddress: validatorAddr,
		StartHeight:      startHeight,
	}
}

// IsTombstoned returns true if the validator is tombstoned
func IsTombstoned(vsi *stakingv1.ValidatorSigningInfo) bool {
	return vsi.Tombstoned
}

// IsJailed returns true if the validator is currently jailed
func IsJailed(vsi *stakingv1.ValidatorSigningInfo, currentTime time.Time) bool {
	if vsi.JailedUntil == nil {
		return false
	}
	return !vsi.JailedUntil.IsZero() && currentTime.Before(*vsi.JailedUntil)
}

// IncrementInfractionCount increments the infraction count
func IncrementInfractionCount(vsi *stakingv1.ValidatorSigningInfo) {
	vsi.InfractionCount++
}
