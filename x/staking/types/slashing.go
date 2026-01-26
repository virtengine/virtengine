// Package types contains types for the staking module.
//
// VE-921: Slashing types for validator misbehavior handling
package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashReason indicates the reason for slashing
type SlashReason string

const (
	// SlashReasonDoubleSigning is for double signing infractions
	SlashReasonDoubleSigning SlashReason = "double_signing"

	// SlashReasonDowntime is for excessive downtime
	SlashReasonDowntime SlashReason = "downtime"

	// SlashReasonInvalidVEIDAttestation is for invalid VEID attestations
	SlashReasonInvalidVEIDAttestation SlashReason = "invalid_veid_attestation"

	// SlashReasonMissedRecomputation is for missing VEID recomputation
	SlashReasonMissedRecomputation SlashReason = "missed_recomputation"

	// SlashReasonInconsistentScore is for scores differing from consensus
	SlashReasonInconsistentScore SlashReason = "inconsistent_score"

	// SlashReasonExpiredAttestation is for expired attestation
	SlashReasonExpiredAttestation SlashReason = "expired_attestation"

	// SlashReasonDebugModeEnabled is for enclave debug mode
	SlashReasonDebugModeEnabled SlashReason = "debug_mode_enabled"

	// SlashReasonNonAllowlistedMeasurement is for non-allowlisted enclave measurement
	SlashReasonNonAllowlistedMeasurement SlashReason = "non_allowlisted_measurement"
)

// SlashRecord represents a slashing record
type SlashRecord struct {
	// SlashID is the unique identifier for this slashing event
	SlashID string `json:"slash_id"`

	// ValidatorAddress is the validator being slashed
	ValidatorAddress string `json:"validator_address"`

	// Reason is the reason for slashing
	Reason SlashReason `json:"reason"`

	// Amount is the amount slashed
	Amount sdk.Coins `json:"amount"`

	// SlashPercent is the percentage slashed (fixed-point, 1e6 scale)
	SlashPercent int64 `json:"slash_percent"`

	// InfractionHeight is the block height of the infraction
	InfractionHeight int64 `json:"infraction_height"`

	// SlashHeight is the block height when slash was executed
	SlashHeight int64 `json:"slash_height"`

	// SlashTime is when the slash was executed
	SlashTime time.Time `json:"slash_time"`

	// Jailed indicates if the validator was jailed
	Jailed bool `json:"jailed"`

	// JailDuration is how long the validator is jailed
	JailDuration int64 `json:"jail_duration"` // seconds

	// JailedUntil is when the jail period ends
	JailedUntil time.Time `json:"jailed_until,omitempty"`

	// Tombstoned indicates if validator is permanently banned
	Tombstoned bool `json:"tombstoned"`

	// Evidence contains the infraction evidence
	Evidence string `json:"evidence,omitempty"`

	// EvidenceHash is the hash of the evidence
	EvidenceHash string `json:"evidence_hash,omitempty"`

	// ReporterAddress is who reported the infraction (if any)
	ReporterAddress string `json:"reporter_address,omitempty"`
}

// NewSlashRecord creates a new slash record
func NewSlashRecord(
	slashID string,
	validatorAddr string,
	reason SlashReason,
	amount sdk.Coins,
	slashPercent int64,
	infractionHeight int64,
	slashHeight int64,
	slashTime time.Time,
) *SlashRecord {
	return &SlashRecord{
		SlashID:          slashID,
		ValidatorAddress: validatorAddr,
		Reason:           reason,
		Amount:           amount,
		SlashPercent:     slashPercent,
		InfractionHeight: infractionHeight,
		SlashHeight:      slashHeight,
		SlashTime:        slashTime,
	}
}

// Validate validates the slash record
func (s *SlashRecord) Validate() error {
	if s.SlashID == "" {
		return fmt.Errorf("slash_id cannot be empty")
	}

	if s.ValidatorAddress == "" {
		return fmt.Errorf("validator_address cannot be empty")
	}

	if !IsValidSlashReason(s.Reason) {
		return fmt.Errorf("invalid slash reason: %s", s.Reason)
	}

	if s.SlashPercent < 0 || s.SlashPercent > FixedPointScale {
		return fmt.Errorf("slash_percent must be between 0 and %d", FixedPointScale)
	}

	if s.InfractionHeight < 0 {
		return fmt.Errorf("infraction_height cannot be negative")
	}

	if s.SlashHeight < 0 {
		return fmt.Errorf("slash_height cannot be negative")
	}

	return nil
}

// IsValidSlashReason checks if the reason is valid
func IsValidSlashReason(reason SlashReason) bool {
	switch reason {
	case SlashReasonDoubleSigning,
		SlashReasonDowntime,
		SlashReasonInvalidVEIDAttestation,
		SlashReasonMissedRecomputation,
		SlashReasonInconsistentScore,
		SlashReasonExpiredAttestation,
		SlashReasonDebugModeEnabled,
		SlashReasonNonAllowlistedMeasurement:
		return true
	default:
		return false
	}
}

// DoubleSignEvidence represents evidence of double signing
type DoubleSignEvidence struct {
	// EvidenceID is the unique identifier
	EvidenceID string `json:"evidence_id"`

	// ValidatorAddress is the validator who double signed
	ValidatorAddress string `json:"validator_address"`

	// Height1 is the first height of the double sign
	Height1 int64 `json:"height_1"`

	// Height2 is the second height of the double sign
	Height2 int64 `json:"height_2"`

	// VoteHash1 is the hash of the first vote
	VoteHash1 string `json:"vote_hash_1"`

	// VoteHash2 is the hash of the second vote
	VoteHash2 string `json:"vote_hash_2"`

	// DetectedAt is when the double sign was detected
	DetectedAt time.Time `json:"detected_at"`

	// DetectedHeight is the block height when detected
	DetectedHeight int64 `json:"detected_height"`

	// Processed indicates if this evidence has been processed
	Processed bool `json:"processed"`
}

// InvalidVEIDAttestation represents evidence of invalid VEID attestation
type InvalidVEIDAttestation struct {
	// RecordID is the unique identifier
	RecordID string `json:"record_id"`

	// ValidatorAddress is the validator with invalid attestation
	ValidatorAddress string `json:"validator_address"`

	// AttestationID is the ID of the invalid attestation
	AttestationID string `json:"attestation_id"`

	// Reason is why the attestation is invalid
	Reason string `json:"reason"`

	// ExpectedScore is the expected score from consensus
	ExpectedScore int64 `json:"expected_score"`

	// ActualScore is the score the validator reported
	ActualScore int64 `json:"actual_score"`

	// ScoreDifference is the difference between expected and actual
	ScoreDifference int64 `json:"score_difference"`

	// DetectedAt is when the issue was detected
	DetectedAt time.Time `json:"detected_at"`

	// DetectedHeight is the block height when detected
	DetectedHeight int64 `json:"detected_height"`

	// Processed indicates if this evidence has been processed
	Processed bool `json:"processed"`
}

// SlashConfig defines the slashing configuration for a reason
type SlashConfig struct {
	// Reason is the slash reason
	Reason SlashReason `json:"reason"`

	// SlashPercent is the base slash percentage (fixed-point, 1e6 scale)
	SlashPercent int64 `json:"slash_percent"`

	// JailDuration is the jail duration in seconds
	JailDuration int64 `json:"jail_duration"`

	// IsTombstone indicates if this should tombstone the validator
	IsTombstone bool `json:"is_tombstone"`

	// EscalationMultiplier is the multiplier for repeat offenses
	EscalationMultiplier int64 `json:"escalation_multiplier"`
}

// DefaultSlashConfigs returns the default slashing configurations
func DefaultSlashConfigs() map[SlashReason]SlashConfig {
	return map[SlashReason]SlashConfig{
		SlashReasonDoubleSigning: {
			Reason:               SlashReasonDoubleSigning,
			SlashPercent:         50000,  // 5%
			JailDuration:         604800, // 1 week
			IsTombstone:          true,
			EscalationMultiplier: 2,
		},
		SlashReasonDowntime: {
			Reason:               SlashReasonDowntime,
			SlashPercent:         1000,  // 0.1%
			JailDuration:         600,   // 10 minutes
			IsTombstone:          false,
			EscalationMultiplier: 2,
		},
		SlashReasonInvalidVEIDAttestation: {
			Reason:               SlashReasonInvalidVEIDAttestation,
			SlashPercent:         50000,  // 5%
			JailDuration:         604800, // 1 week
			IsTombstone:          false,
			EscalationMultiplier: 2,
		},
		SlashReasonMissedRecomputation: {
			Reason:               SlashReasonMissedRecomputation,
			SlashPercent:         10000,  // 1% per missed block
			JailDuration:         3600,   // 1 hour
			IsTombstone:          false,
			EscalationMultiplier: 1,
		},
		SlashReasonInconsistentScore: {
			Reason:               SlashReasonInconsistentScore,
			SlashPercent:         20000,  // 2%
			JailDuration:         86400,  // 1 day
			IsTombstone:          false,
			EscalationMultiplier: 2,
		},
		SlashReasonExpiredAttestation: {
			Reason:               SlashReasonExpiredAttestation,
			SlashPercent:         10000,  // 1%
			JailDuration:         3600,   // 1 hour (first offense)
			IsTombstone:          false,
			EscalationMultiplier: 2,
		},
		SlashReasonDebugModeEnabled: {
			Reason:               SlashReasonDebugModeEnabled,
			SlashPercent:         200000, // 20%
			JailDuration:         0,      // Permanent
			IsTombstone:          true,
			EscalationMultiplier: 1,
		},
		SlashReasonNonAllowlistedMeasurement: {
			Reason:               SlashReasonNonAllowlistedMeasurement,
			SlashPercent:         100000, // 10%
			JailDuration:         0,      // Indefinite
			IsTombstone:          true,
			EscalationMultiplier: 1,
		},
	}
}

// GetSlashConfig returns the slash config for a reason
func GetSlashConfig(reason SlashReason) SlashConfig {
	configs := DefaultSlashConfigs()
	if config, ok := configs[reason]; ok {
		return config
	}
	// Default config for unknown reasons
	return SlashConfig{
		Reason:               reason,
		SlashPercent:         10000, // 1%
		JailDuration:         3600,
		IsTombstone:          false,
		EscalationMultiplier: 1,
	}
}
