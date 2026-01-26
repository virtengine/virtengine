// Package types contains types for the staking module.
//
// VE-921: Staking rewards events
package types

const (
	// EventTypeValidatorSlashed is emitted when a validator is slashed
	EventTypeValidatorSlashed = "validator_slashed"

	// EventTypeValidatorJailed is emitted when a validator is jailed
	EventTypeValidatorJailed = "validator_jailed"

	// EventTypeValidatorUnjailed is emitted when a validator is unjailed
	EventTypeValidatorUnjailed = "validator_unjailed"

	// EventTypeRewardsDistributed is emitted when staking rewards are distributed
	EventTypeRewardsDistributed = "staking_rewards_distributed"

	// EventTypeVEIDRewardsDistributed is emitted when VEID verification rewards are distributed
	EventTypeVEIDRewardsDistributed = "veid_rewards_distributed"

	// EventTypePerformanceRecorded is emitted when validator performance is recorded
	EventTypePerformanceRecorded = "performance_recorded"

	// EventTypeDoubleSignDetected is emitted when double signing is detected
	EventTypeDoubleSignDetected = "double_sign_detected"

	// EventTypeInvalidAttestationDetected is emitted when invalid VEID attestation is detected
	EventTypeInvalidAttestationDetected = "invalid_attestation_detected"

	// EventTypeDowntimeDetected is emitted when excessive downtime is detected
	EventTypeDowntimeDetected = "downtime_detected"

	// Attribute keys
	AttributeKeyValidatorAddress = "validator_address"
	AttributeKeySlashReason      = "slash_reason"
	AttributeKeySlashAmount      = "slash_amount"
	AttributeKeySlashPercent     = "slash_percent"
	AttributeKeyJailDuration     = "jail_duration"
	AttributeKeyEpochNumber      = "epoch_number"
	AttributeKeyTotalRewards     = "total_rewards"
	AttributeKeyRecipientCount   = "recipient_count"
	AttributeKeyPerformanceScore = "performance_score"
	AttributeKeyUptimePercent    = "uptime_percent"
	AttributeKeyBlocksProposed   = "blocks_proposed"
	AttributeKeyBlocksMissed     = "blocks_missed"
	AttributeKeyVEIDScore        = "veid_score"
)

// EventValidatorSlashed is emitted when a validator is slashed
type EventValidatorSlashed struct {
	ValidatorAddress string `json:"validator_address"`
	SlashReason      string `json:"slash_reason"`
	SlashAmount      string `json:"slash_amount"`
	SlashPercent     string `json:"slash_percent"`
	InfractionHeight int64  `json:"infraction_height"`
	SlashTime        int64  `json:"slash_time"`
}

// EventValidatorJailed is emitted when a validator is jailed
type EventValidatorJailed struct {
	ValidatorAddress string `json:"validator_address"`
	Reason           string `json:"reason"`
	JailDuration     int64  `json:"jail_duration"`
	JailedUntil      int64  `json:"jailed_until"`
}

// EventValidatorUnjailed is emitted when a validator is unjailed
type EventValidatorUnjailed struct {
	ValidatorAddress string `json:"validator_address"`
	UnjailedAt       int64  `json:"unjailed_at"`
}

// EventRewardsDistributed is emitted when rewards are distributed
type EventRewardsDistributed struct {
	EpochNumber    uint64 `json:"epoch_number"`
	TotalRewards   string `json:"total_rewards"`
	RecipientCount uint32 `json:"recipient_count"`
	RewardType     string `json:"reward_type"`
	DistributedAt  int64  `json:"distributed_at"`
}

// EventPerformanceRecorded is emitted when validator performance is recorded
type EventPerformanceRecorded struct {
	ValidatorAddress string `json:"validator_address"`
	PerformanceScore int64  `json:"performance_score"`
	UptimePercent    int64  `json:"uptime_percent"`
	BlocksProposed   int64  `json:"blocks_proposed"`
	VEIDScore        int64  `json:"veid_score"`
	BlockHeight      int64  `json:"block_height"`
}

// EventDoubleSignDetected is emitted when double signing is detected
type EventDoubleSignDetected struct {
	ValidatorAddress string `json:"validator_address"`
	Height1          int64  `json:"height_1"`
	Height2          int64  `json:"height_2"`
	DetectedAt       int64  `json:"detected_at"`
}

// EventInvalidAttestationDetected is emitted when invalid VEID attestation is detected
type EventInvalidAttestationDetected struct {
	ValidatorAddress string `json:"validator_address"`
	AttestationID    string `json:"attestation_id"`
	Reason           string `json:"reason"`
	DetectedAt       int64  `json:"detected_at"`
}

// EventDowntimeDetected is emitted when excessive downtime is detected
type EventDowntimeDetected struct {
	ValidatorAddress string `json:"validator_address"`
	MissedBlocks     int64  `json:"missed_blocks"`
	WindowSize       int64  `json:"window_size"`
	Threshold        int64  `json:"threshold"`
	DetectedAt       int64  `json:"detected_at"`
}
