// Package types contains types for the staking module.
//
// VE-921: Genesis state and parameters
package types

import (
	"fmt"
)

// GenesisState is the genesis state for the staking module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// ValidatorPerformances are the initial validator performances
	ValidatorPerformances []ValidatorPerformance `json:"validator_performances"`

	// SlashRecords are the initial slashing records
	SlashRecords []SlashRecord `json:"slash_records"`

	// RewardEpochs are the initial reward epochs
	RewardEpochs []RewardEpoch `json:"reward_epochs"`

	// ValidatorRewards are the initial validator rewards
	ValidatorRewards []ValidatorReward `json:"validator_rewards"`

	// ValidatorSigningInfos are the initial signing infos
	ValidatorSigningInfos []ValidatorSigningInfo `json:"validator_signing_infos"`

	// CurrentEpoch is the current epoch number
	CurrentEpoch uint64 `json:"current_epoch"`

	// SlashSequence is the next slashing sequence number
	SlashSequence uint64 `json:"slash_sequence"`
}

// Params defines the parameters for the staking module
type Params struct {
	// EpochLength is the number of blocks per reward epoch
	EpochLength uint64 `json:"epoch_length"`

	// BaseRewardPerBlock is the base reward per block in smallest unit
	BaseRewardPerBlock int64 `json:"base_reward_per_block"`

	// VEIDRewardPool is the VEID verification reward pool per epoch
	VEIDRewardPool int64 `json:"veid_reward_pool"`

	// IdentityNetworkRewardPool is the identity network reward pool per epoch
	IdentityNetworkRewardPool int64 `json:"identity_network_reward_pool"`

	// DowntimeThreshold is the consecutive missed blocks before slashing
	DowntimeThreshold int64 `json:"downtime_threshold"`

	// SignedBlocksWindow is the window size for tracking missed blocks
	SignedBlocksWindow int64 `json:"signed_blocks_window"`

	// MinSignedPerWindow is the minimum percentage that must be signed (fixed-point)
	MinSignedPerWindow int64 `json:"min_signed_per_window"`

	// SlashFractionDoubleSign is the slash fraction for double signing (fixed-point)
	SlashFractionDoubleSign int64 `json:"slash_fraction_double_sign"`

	// SlashFractionDowntime is the slash fraction for downtime (fixed-point)
	SlashFractionDowntime int64 `json:"slash_fraction_downtime"`

	// SlashFractionInvalidAttestation is the slash fraction for invalid VEID attestation (fixed-point)
	SlashFractionInvalidAttestation int64 `json:"slash_fraction_invalid_attestation"`

	// JailDurationDowntime is the jail duration for downtime (seconds)
	JailDurationDowntime int64 `json:"jail_duration_downtime"`

	// JailDurationDoubleSign is the jail duration for double signing (seconds)
	JailDurationDoubleSign int64 `json:"jail_duration_double_sign"`

	// JailDurationInvalidAttestation is the jail duration for invalid attestation (seconds)
	JailDurationInvalidAttestation int64 `json:"jail_duration_invalid_attestation"`

	// ScoreTolerance is the allowed score difference from consensus (fixed-point)
	ScoreTolerance int64 `json:"score_tolerance"`

	// MaxMissedVEIDRecomputations is max missed recomputations before slash
	MaxMissedVEIDRecomputations int64 `json:"max_missed_veid_recomputations"`

	// RewardDenom is the denomination for rewards
	RewardDenom string `json:"reward_denom"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                DefaultParams(),
		ValidatorPerformances: []ValidatorPerformance{},
		SlashRecords:          []SlashRecord{},
		RewardEpochs:          []RewardEpoch{},
		ValidatorRewards:      []ValidatorReward{},
		ValidatorSigningInfos: []ValidatorSigningInfo{},
		CurrentEpoch:          1,
		SlashSequence:         1,
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		EpochLength:                     100,       // 100 blocks per epoch
		BaseRewardPerBlock:              1000000,   // 1 token per block (in smallest unit)
		VEIDRewardPool:                  100000000, // 100 tokens per epoch
		IdentityNetworkRewardPool:       50000000,  // 50 tokens per epoch
		DowntimeThreshold:               100,       // 100 consecutive missed blocks
		SignedBlocksWindow:              10000,     // 10000 block window
		MinSignedPerWindow:              500000,    // 50% minimum (fixed-point 1e6)
		SlashFractionDoubleSign:         50000,     // 5% (fixed-point 1e6)
		SlashFractionDowntime:           1000,      // 0.1% (fixed-point 1e6)
		SlashFractionInvalidAttestation: 50000,     // 5% (fixed-point 1e6)
		JailDurationDowntime:            600,       // 10 minutes
		JailDurationDoubleSign:          604800,    // 1 week
		JailDurationInvalidAttestation:  604800,    // 1 week
		ScoreTolerance:                  100,       // 1% tolerance
		MaxMissedVEIDRecomputations:     10,        // 10 missed before slash
		RewardDenom:                     "uve",
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate validator performances
	seenPerf := make(map[string]bool)
	for _, perf := range gs.ValidatorPerformances {
		if err := perf.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s-%d", perf.ValidatorAddress, perf.EpochNumber)
		if seenPerf[key] {
			return fmt.Errorf("duplicate validator performance: %s", key)
		}
		seenPerf[key] = true
	}

	// Validate slash records
	seenSlash := make(map[string]bool)
	for _, slash := range gs.SlashRecords {
		if err := slash.Validate(); err != nil {
			return err
		}
		if seenSlash[slash.SlashID] {
			return fmt.Errorf("duplicate slash_id: %s", slash.SlashID)
		}
		seenSlash[slash.SlashID] = true
	}

	// Validate reward epochs
	seenEpoch := make(map[uint64]bool)
	for _, epoch := range gs.RewardEpochs {
		if err := epoch.Validate(); err != nil {
			return err
		}
		if seenEpoch[epoch.EpochNumber] {
			return fmt.Errorf("duplicate epoch_number: %d", epoch.EpochNumber)
		}
		seenEpoch[epoch.EpochNumber] = true
	}

	// Validate validator rewards
	seenReward := make(map[string]bool)
	for _, reward := range gs.ValidatorRewards {
		if err := reward.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s-%d", reward.ValidatorAddress, reward.EpochNumber)
		if seenReward[key] {
			return fmt.Errorf("duplicate validator reward: %s", key)
		}
		seenReward[key] = true
	}

	return nil
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.EpochLength == 0 {
		return fmt.Errorf("epoch_length must be greater than zero")
	}

	if p.BaseRewardPerBlock < 0 {
		return fmt.Errorf("base_reward_per_block cannot be negative")
	}

	if p.DowntimeThreshold <= 0 {
		return fmt.Errorf("downtime_threshold must be greater than zero")
	}

	if p.SignedBlocksWindow <= 0 {
		return fmt.Errorf("signed_blocks_window must be greater than zero")
	}

	if p.MinSignedPerWindow < 0 || p.MinSignedPerWindow > FixedPointScale {
		return fmt.Errorf("min_signed_per_window must be between 0 and %d", FixedPointScale)
	}

	if p.SlashFractionDoubleSign < 0 || p.SlashFractionDoubleSign > FixedPointScale {
		return fmt.Errorf("slash_fraction_double_sign must be between 0 and %d", FixedPointScale)
	}

	if p.SlashFractionDowntime < 0 || p.SlashFractionDowntime > FixedPointScale {
		return fmt.Errorf("slash_fraction_downtime must be between 0 and %d", FixedPointScale)
	}

	if p.SlashFractionInvalidAttestation < 0 || p.SlashFractionInvalidAttestation > FixedPointScale {
		return fmt.Errorf("slash_fraction_invalid_attestation must be between 0 and %d", FixedPointScale)
	}

	if p.RewardDenom == "" {
		return fmt.Errorf("reward_denom cannot be empty")
	}

	return nil
}
