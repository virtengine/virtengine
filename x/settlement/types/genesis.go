package types

import "fmt"

// GenesisState is the genesis state for the settlement module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// EscrowAccounts are the initial escrow accounts
	EscrowAccounts []EscrowAccount `json:"escrow_accounts"`

	// SettlementRecords are the initial settlement records
	SettlementRecords []SettlementRecord `json:"settlement_records"`

	// RewardDistributions are the initial reward distributions
	RewardDistributions []RewardDistribution `json:"reward_distributions"`

	// UsageRecords are the initial usage records
	UsageRecords []UsageRecord `json:"usage_records"`

	// ClaimableRewards are the initial claimable rewards
	ClaimableRewards []ClaimableRewards `json:"claimable_rewards"`

	// PayoutRecords are the initial payout records
	PayoutRecords []PayoutRecord `json:"payout_records"`

	// EscrowSequence is the next escrow sequence number
	EscrowSequence uint64 `json:"escrow_sequence"`

	// SettlementSequence is the next settlement sequence number
	SettlementSequence uint64 `json:"settlement_sequence"`

	// DistributionSequence is the next distribution sequence number
	DistributionSequence uint64 `json:"distribution_sequence"`

	// UsageSequence is the next usage sequence number
	UsageSequence uint64 `json:"usage_sequence"`

	// PayoutSequence is the next payout sequence number
	PayoutSequence uint64 `json:"payout_sequence"`
}

// Params defines the parameters for the settlement module
type Params struct {
	// PlatformFeeRate is the platform fee rate (e.g., 0.05 for 5%)
	PlatformFeeRate string `json:"platform_fee_rate"`

	// ValidatorFeeRate is the validator fee rate (e.g., 0.01 for 1%)
	ValidatorFeeRate string `json:"validator_fee_rate"`

	// MinEscrowDuration is the minimum escrow duration in seconds
	MinEscrowDuration uint64 `json:"min_escrow_duration"`

	// MaxEscrowDuration is the maximum escrow duration in seconds
	MaxEscrowDuration uint64 `json:"max_escrow_duration"`

	// SettlementPeriod is the default settlement period in seconds
	SettlementPeriod uint64 `json:"settlement_period"`

	// RewardClaimExpiry is how long rewards can be claimed (in seconds)
	RewardClaimExpiry uint64 `json:"reward_claim_expiry"`

	// MinSettlementAmount is the minimum amount for a settlement
	MinSettlementAmount string `json:"min_settlement_amount"`

	// UsageGracePeriod is the grace period for usage disputes (in seconds)
	UsageGracePeriod uint64 `json:"usage_grace_period"`

	// StakingRewardEpochLength is the length of staking reward epochs in blocks
	StakingRewardEpochLength uint64 `json:"staking_reward_epoch_length"`

	// VerificationRewardAmount is the base reward for identity verifications
	VerificationRewardAmount string `json:"verification_reward_amount"`

	// PayoutHoldbackRate is the holdback rate for payouts (e.g., 0.0 for no holdback)
	PayoutHoldbackRate string `json:"payout_holdback_rate"`

	// MaxPayoutRetries is the maximum number of retry attempts for failed payouts
	MaxPayoutRetries uint32 `json:"max_payout_retries"`

	// DisputeWindowDuration is the dispute window duration in seconds
	DisputeWindowDuration uint64 `json:"dispute_window_duration"`

	// UsageRewardRateBps is the base reward rate for usage rewards (basis points)
	UsageRewardRateBps uint32 `json:"usage_reward_rate_bps"`

	// UsageRewardCPUMultiplierBps is the CPU usage reward multiplier in basis points
	UsageRewardCPUMultiplierBps uint32 `json:"usage_reward_cpu_multiplier_bps"`

	// UsageRewardMemoryMultiplierBps is the memory usage reward multiplier in basis points
	UsageRewardMemoryMultiplierBps uint32 `json:"usage_reward_memory_multiplier_bps"`

	// UsageRewardStorageMultiplierBps is the storage usage reward multiplier in basis points
	UsageRewardStorageMultiplierBps uint32 `json:"usage_reward_storage_multiplier_bps"`

	// UsageRewardGPUMultiplierBps is the GPU usage reward multiplier in basis points
	UsageRewardGPUMultiplierBps uint32 `json:"usage_reward_gpu_multiplier_bps"`

	// UsageRewardNetworkMultiplierBps is the network usage reward multiplier in basis points
	UsageRewardNetworkMultiplierBps uint32 `json:"usage_reward_network_multiplier_bps"`

	// UsageRewardSLAOnTimeMultiplierBps is the on-time reporting SLA multiplier in basis points
	UsageRewardSLAOnTimeMultiplierBps uint32 `json:"usage_reward_sla_ontime_multiplier_bps"`

	// UsageRewardSLALateMultiplierBps is the late reporting SLA multiplier in basis points
	UsageRewardSLALateMultiplierBps uint32 `json:"usage_reward_sla_late_multiplier_bps"`

	// UsageRewardAcknowledgedMultiplierBps is the customer-acknowledged quality multiplier in basis points
	UsageRewardAcknowledgedMultiplierBps uint32 `json:"usage_reward_ack_multiplier_bps"`

	// UsageRewardUnacknowledgedMultiplierBps is the unacknowledged quality multiplier in basis points
	UsageRewardUnacknowledgedMultiplierBps uint32 `json:"usage_reward_unack_multiplier_bps"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:               DefaultParams(),
		EscrowAccounts:       []EscrowAccount{},
		SettlementRecords:    []SettlementRecord{},
		RewardDistributions:  []RewardDistribution{},
		UsageRecords:         []UsageRecord{},
		ClaimableRewards:     []ClaimableRewards{},
		PayoutRecords:        []PayoutRecord{},
		EscrowSequence:       1,
		SettlementSequence:   1,
		DistributionSequence: 1,
		UsageSequence:        1,
		PayoutSequence:       1,
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		PlatformFeeRate:                        "0.05",   // 5%
		ValidatorFeeRate:                       "0.01",   // 1%
		MinEscrowDuration:                      3600,     // 1 hour
		MaxEscrowDuration:                      31536000, // 1 year
		SettlementPeriod:                       86400,    // 1 day
		RewardClaimExpiry:                      2592000,  // 30 days
		MinSettlementAmount:                    "1000",   // Minimum tokens for settlement
		UsageGracePeriod:                       86400,    // 1 day grace period
		StakingRewardEpochLength:               100,      // 100 blocks per epoch
		VerificationRewardAmount:               "100",    // Base reward for verification
		PayoutHoldbackRate:                     "0.0",    // No holdback by default
		MaxPayoutRetries:                       3,        // 3 retry attempts
		DisputeWindowDuration:                  604800,   // 7 days
		UsageRewardRateBps:                     1000,     // 10% base reward on usage value
		UsageRewardCPUMultiplierBps:            10000,    // 1.0x
		UsageRewardMemoryMultiplierBps:         10000,    // 1.0x
		UsageRewardStorageMultiplierBps:        10000,    // 1.0x
		UsageRewardGPUMultiplierBps:            12000,    // 1.2x
		UsageRewardNetworkMultiplierBps:        9000,     // 0.9x
		UsageRewardSLAOnTimeMultiplierBps:      10000,    // 1.0x
		UsageRewardSLALateMultiplierBps:        8000,     // 0.8x
		UsageRewardAcknowledgedMultiplierBps:   10000,    // 1.0x
		UsageRewardUnacknowledgedMultiplierBps: 9000,     // 0.9x
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate escrow accounts
	seenEscrows := make(map[string]bool)
	for _, escrow := range gs.EscrowAccounts {
		if err := escrow.Validate(); err != nil {
			return err
		}
		if seenEscrows[escrow.EscrowID] {
			return ErrEscrowExists.Wrapf("duplicate escrow_id: %s", escrow.EscrowID)
		}
		seenEscrows[escrow.EscrowID] = true
	}

	// Validate settlement records
	seenSettlements := make(map[string]bool)
	for _, settlement := range gs.SettlementRecords {
		if err := settlement.Validate(); err != nil {
			return err
		}
		if seenSettlements[settlement.SettlementID] {
			return ErrSettlementExists.Wrapf("duplicate settlement_id: %s", settlement.SettlementID)
		}
		seenSettlements[settlement.SettlementID] = true
	}

	// Validate reward distributions
	seenDistributions := make(map[string]bool)
	for _, dist := range gs.RewardDistributions {
		if err := dist.Validate(); err != nil {
			return err
		}
		if seenDistributions[dist.DistributionID] {
			return ErrInvalidReward.Wrapf("duplicate distribution_id: %s", dist.DistributionID)
		}
		seenDistributions[dist.DistributionID] = true
	}

	// Validate usage records
	seenUsage := make(map[string]bool)
	for _, usage := range gs.UsageRecords {
		if err := usage.Validate(); err != nil {
			return err
		}
		if seenUsage[usage.UsageID] {
			return ErrUsageRecordExists.Wrapf("duplicate usage_id: %s", usage.UsageID)
		}
		seenUsage[usage.UsageID] = true
	}

	// Validate payout records
	seenPayouts := make(map[string]bool)
	for _, payout := range gs.PayoutRecords {
		if err := payout.Validate(); err != nil {
			return err
		}
		if seenPayouts[payout.PayoutID] {
			return ErrPayoutExists.Wrapf("duplicate payout_id: %s", payout.PayoutID)
		}
		seenPayouts[payout.PayoutID] = true
	}

	return nil
}

// Validate validates the parameters
func (p Params) Validate() error {
	// Validate fee rates are between 0 and 1
	// We'll do basic validation here; more sophisticated parsing would be needed in production

	if p.MinEscrowDuration == 0 {
		return ErrInvalidParams.Wrap("min_escrow_duration must be greater than zero")
	}

	if p.MaxEscrowDuration <= p.MinEscrowDuration {
		return ErrInvalidParams.Wrap("max_escrow_duration must be greater than min_escrow_duration")
	}

	if p.SettlementPeriod == 0 {
		return ErrInvalidParams.Wrap("settlement_period must be greater than zero")
	}

	if p.StakingRewardEpochLength == 0 {
		return ErrInvalidParams.Wrap("staking_reward_epoch_length must be greater than zero")
	}

	if p.UsageRewardRateBps > 10000 {
		return ErrInvalidParams.Wrap("usage_reward_rate_bps cannot exceed 10000")
	}

	if err := validateRewardMultiplierBps(p.UsageRewardCPUMultiplierBps, "usage_reward_cpu_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardMemoryMultiplierBps, "usage_reward_memory_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardStorageMultiplierBps, "usage_reward_storage_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardGPUMultiplierBps, "usage_reward_gpu_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardNetworkMultiplierBps, "usage_reward_network_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardSLAOnTimeMultiplierBps, "usage_reward_sla_ontime_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardSLALateMultiplierBps, "usage_reward_sla_late_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardAcknowledgedMultiplierBps, "usage_reward_ack_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardUnacknowledgedMultiplierBps, "usage_reward_unack_multiplier_bps"); err != nil {
		return err
	}

	return nil
}

func validateRewardMultiplierBps(value uint32, name string) error {
	if value == 0 {
		return ErrInvalidParams.Wrapf("%s must be greater than zero", name)
	}
	if value > 20000 {
		return ErrInvalidParams.Wrapf("%s cannot exceed 20000", name)
	}
	return nil
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string {
	return fmt.Sprintf("%+v", *gs)
}
