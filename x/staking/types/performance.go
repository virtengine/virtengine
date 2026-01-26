// Package types contains types for the staking module.
//
// VE-921: Performance types for validator performance tracking
package types

import (
	"fmt"
	"time"
)

// FixedPointScale is the scale factor for fixed-point arithmetic (1e6)
const FixedPointScale int64 = 1000000

// MaxPerformanceScore is the maximum performance score (10000 = 100%)
const MaxPerformanceScore int64 = 10000

// PerformanceWeight constants for reward calculation
const (
	WeightBlockProposal       int64 = 3000 // 30%
	WeightVEIDVerification    int64 = 4000 // 40%
	WeightUptime              int64 = 3000 // 30%
	TotalWeight               int64 = 10000
)

// ValidatorPerformance represents a validator's performance metrics
type ValidatorPerformance struct {
	// ValidatorAddress is the validator's blockchain address
	ValidatorAddress string `json:"validator_address"`

	// BlocksProposed is the number of blocks proposed in the current epoch
	BlocksProposed int64 `json:"blocks_proposed"`

	// BlocksExpected is the expected number of blocks based on stake weight
	BlocksExpected int64 `json:"blocks_expected"`

	// BlocksMissed is the number of missed blocks (when expected to sign)
	BlocksMissed int64 `json:"blocks_missed"`

	// TotalSignatures is the total number of blocks signed
	TotalSignatures int64 `json:"total_signatures"`

	// VEIDVerificationsCompleted is the number of VEID verifications completed
	VEIDVerificationsCompleted int64 `json:"veid_verifications_completed"`

	// VEIDVerificationsExpected is the expected VEID verifications based on committee selection
	VEIDVerificationsExpected int64 `json:"veid_verifications_expected"`

	// VEIDVerificationScore is the quality score for VEID verifications (0-10000)
	VEIDVerificationScore int64 `json:"veid_verification_score"`

	// UptimeSeconds is the total uptime in seconds
	UptimeSeconds int64 `json:"uptime_seconds"`

	// DowntimeSeconds is the total downtime in seconds
	DowntimeSeconds int64 `json:"downtime_seconds"`

	// ConsecutiveMissedBlocks is the current streak of missed blocks
	ConsecutiveMissedBlocks int64 `json:"consecutiVIRTENGINE_missed_blocks"`

	// LastProposedHeight is the last height where this validator proposed a block
	LastProposedHeight int64 `json:"last_proposed_height"`

	// LastSignedHeight is the last height where this validator signed
	LastSignedHeight int64 `json:"last_signed_height"`

	// EpochNumber is the epoch this performance record belongs to
	EpochNumber uint64 `json:"epoch_number"`

	// UpdatedAt is when this record was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// OverallScore is the computed overall performance score (0-10000)
	OverallScore int64 `json:"overall_score"`
}

// NewValidatorPerformance creates a new validator performance record
func NewValidatorPerformance(validatorAddr string, epochNumber uint64) *ValidatorPerformance {
	return &ValidatorPerformance{
		ValidatorAddress:       validatorAddr,
		EpochNumber:            epochNumber,
		VEIDVerificationScore:  MaxPerformanceScore, // Start with full score
		OverallScore:           0,
	}
}

// Validate validates the validator performance record
func (vp *ValidatorPerformance) Validate() error {
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
func (vp *ValidatorPerformance) ComputeOverallScore() int64 {
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

	vp.OverallScore = overallScore
	return overallScore
}

// GetUptimePercent returns the uptime percentage in fixed-point (1e6 scale)
func (vp *ValidatorPerformance) GetUptimePercent() int64 {
	totalTime := vp.UptimeSeconds + vp.DowntimeSeconds
	if totalTime == 0 {
		return FixedPointScale // 100%
	}
	return (vp.UptimeSeconds * FixedPointScale) / totalTime
}

// ShouldSlashForDowntime checks if the validator should be slashed for downtime
func (vp *ValidatorPerformance) ShouldSlashForDowntime(threshold int64) bool {
	return vp.ConsecutiveMissedBlocks >= threshold
}

// ValidatorSigningInfo contains validator signing information for slashing
type ValidatorSigningInfo struct {
	// ValidatorAddress is the validator's blockchain address
	ValidatorAddress string `json:"validator_address"`

	// StartHeight is the height at which validator started signing
	StartHeight int64 `json:"start_height"`

	// IndexOffset is the current index offset into the signed blocks window
	IndexOffset int64 `json:"index_offset"`

	// JailedUntil is the time until which the validator is jailed
	JailedUntil time.Time `json:"jailed_until"`

	// Tombstoned indicates if the validator has been tombstoned (permanently banned)
	Tombstoned bool `json:"tombstoned"`

	// MissedBlocksCounter is the counter for missed blocks in the current window
	MissedBlocksCounter int64 `json:"missed_blocks_counter"`

	// InfractionCount is the total number of infractions
	InfractionCount int64 `json:"infraction_count"`
}

// NewValidatorSigningInfo creates a new signing info record
func NewValidatorSigningInfo(validatorAddr string, startHeight int64) *ValidatorSigningInfo {
	return &ValidatorSigningInfo{
		ValidatorAddress: validatorAddr,
		StartHeight:      startHeight,
	}
}

// IsTombstoned returns true if the validator is tombstoned
func (vsi *ValidatorSigningInfo) IsTombstoned() bool {
	return vsi.Tombstoned
}

// IsJailed returns true if the validator is currently jailed
func (vsi *ValidatorSigningInfo) IsJailed(currentTime time.Time) bool {
	return !vsi.JailedUntil.IsZero() && currentTime.Before(vsi.JailedUntil)
}

// IncrementInfractionCount increments the infraction count
func (vsi *ValidatorSigningInfo) IncrementInfractionCount() {
	vsi.InfractionCount++
}
