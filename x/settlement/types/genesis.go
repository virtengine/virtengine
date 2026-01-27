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

	// EscrowSequence is the next escrow sequence number
	EscrowSequence uint64 `json:"escrow_sequence"`

	// SettlementSequence is the next settlement sequence number
	SettlementSequence uint64 `json:"settlement_sequence"`

	// DistributionSequence is the next distribution sequence number
	DistributionSequence uint64 `json:"distribution_sequence"`

	// UsageSequence is the next usage sequence number
	UsageSequence uint64 `json:"usage_sequence"`
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
		EscrowSequence:       1,
		SettlementSequence:   1,
		DistributionSequence: 1,
		UsageSequence:        1,
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		PlatformFeeRate:          "0.05",  // 5%
		ValidatorFeeRate:         "0.01",  // 1%
		MinEscrowDuration:        3600,    // 1 hour
		MaxEscrowDuration:        31536000, // 1 year
		SettlementPeriod:         86400,   // 1 day
		RewardClaimExpiry:        2592000, // 30 days
		MinSettlementAmount:      "1000",  // Minimum tokens for settlement
		UsageGracePeriod:         86400,   // 1 day grace period
		StakingRewardEpochLength: 100,     // 100 blocks per epoch
		VerificationRewardAmount: "100",   // Base reward for verification
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
