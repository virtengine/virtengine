// Package types contains types for the staking module.
//
// VE-921: Genesis state and parameters
// Functions for genesis state and parameter validation.
package types

import (
	"fmt"

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
)

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                DefaultParams(),
		ValidatorPerformances: []stakingv1.ValidatorPerformance{},
		SlashRecords:          []stakingv1.SlashRecord{},
		RewardEpochs:          []stakingv1.RewardEpoch{},
		ValidatorRewards:      []stakingv1.ValidatorReward{},
		ValidatorSigningInfos: []stakingv1.ValidatorSigningInfo{},
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

// ValidateGenesis validates the genesis state
func ValidateGenesis(gs *GenesisState) error {
	if err := ValidateParams(gs.Params); err != nil {
		return err
	}

	// Validate validator performances
	seenPerf := make(map[string]bool)
	for _, perf := range gs.ValidatorPerformances {
		key := fmt.Sprintf("%s-%d", perf.ValidatorAddress, perf.EpochNumber)
		if seenPerf[key] {
			return fmt.Errorf("duplicate validator performance: %s", key)
		}
		seenPerf[key] = true
	}

	// Validate slash records
	seenSlash := make(map[string]bool)
	for _, slash := range gs.SlashRecords {
		if seenSlash[slash.SlashId] {
			return fmt.Errorf("duplicate slash_id: %s", slash.SlashId)
		}
		seenSlash[slash.SlashId] = true
	}

	// Validate reward epochs
	seenEpoch := make(map[uint64]bool)
	for _, epoch := range gs.RewardEpochs {
		if seenEpoch[epoch.EpochNumber] {
			return fmt.Errorf("duplicate epoch_number: %d", epoch.EpochNumber)
		}
		seenEpoch[epoch.EpochNumber] = true
	}

	// Validate validator rewards
	seenReward := make(map[string]bool)
	for _, reward := range gs.ValidatorRewards {
		key := fmt.Sprintf("%s-%d", reward.ValidatorAddress, reward.EpochNumber)
		if seenReward[key] {
			return fmt.Errorf("duplicate validator reward: %s", key)
		}
		seenReward[key] = true
	}

	return nil
}

// ValidateParams validates the parameters
func ValidateParams(p Params) error {
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
