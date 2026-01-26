// Package types contains types for the delegation module.
//
// VE-922: Genesis state and parameters for delegated staking
package types

import (
	"fmt"
)

// DefaultUnbondingPeriod is the default unbonding period in seconds (21 days)
const DefaultUnbondingPeriod int64 = 21 * 24 * 60 * 60

// DefaultMaxValidatorsPerDelegator is the default max validators per delegator
const DefaultMaxValidatorsPerDelegator int64 = 10

// DefaultMinDelegationAmount is the default minimum delegation amount (1 token = 1e6 utoken)
const DefaultMinDelegationAmount int64 = 1000000

// DefaultValidatorCommissionRate is the default validator commission rate (10% = 1000 basis points)
const DefaultValidatorCommissionRate int64 = 1000

// DefaultMaxRedelegations is the default maximum simultaneous redelegations
const DefaultMaxRedelegations int64 = 7

// BasisPointsMax is 100% in basis points
const BasisPointsMax int64 = 10000

// GenesisState is the genesis state for the delegation module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// Delegations are the initial delegations
	Delegations []Delegation `json:"delegations"`

	// UnbondingDelegations are the initial unbonding delegations
	UnbondingDelegations []UnbondingDelegation `json:"unbonding_delegations"`

	// Redelegations are the initial redelegations
	Redelegations []Redelegation `json:"redelegations"`

	// ValidatorShares are the initial validator shares
	ValidatorShares []ValidatorShares `json:"validator_shares"`

	// DelegatorRewards are the initial delegator rewards
	DelegatorRewards []DelegatorReward `json:"delegator_rewards"`

	// DelegationSequence is the next delegation sequence number
	DelegationSequence uint64 `json:"delegation_sequence"`

	// UnbondingSequence is the next unbonding sequence number
	UnbondingSequence uint64 `json:"unbonding_sequence"`

	// RedelegationSequence is the next redelegation sequence number
	RedelegationSequence uint64 `json:"redelegation_sequence"`
}

// Params defines the parameters for the delegation module
type Params struct {
	// UnbondingPeriod is the duration for unbonding in seconds
	UnbondingPeriod int64 `json:"unbonding_period"`

	// MaxValidatorsPerDelegator is the maximum number of validators a delegator can delegate to
	MaxValidatorsPerDelegator int64 `json:"max_validators_per_delegator"`

	// MinDelegationAmount is the minimum delegation amount in base units
	MinDelegationAmount int64 `json:"min_delegation_amount"`

	// MaxRedelegations is the maximum number of simultaneous redelegations
	MaxRedelegations int64 `json:"max_redelegations"`

	// ValidatorCommissionRate is the validator commission rate in basis points (e.g., 1000 = 10%)
	ValidatorCommissionRate int64 `json:"validator_commission_rate"`

	// RewardDenom is the denomination for rewards
	RewardDenom string `json:"reward_denom"`

	// StakeDenom is the denomination for staking
	StakeDenom string `json:"stake_denom"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:               DefaultParams(),
		Delegations:          []Delegation{},
		UnbondingDelegations: []UnbondingDelegation{},
		Redelegations:        []Redelegation{},
		ValidatorShares:      []ValidatorShares{},
		DelegatorRewards:     []DelegatorReward{},
		DelegationSequence:   1,
		UnbondingSequence:    1,
		RedelegationSequence: 1,
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		UnbondingPeriod:           DefaultUnbondingPeriod,
		MaxValidatorsPerDelegator: DefaultMaxValidatorsPerDelegator,
		MinDelegationAmount:       DefaultMinDelegationAmount,
		MaxRedelegations:          DefaultMaxRedelegations,
		ValidatorCommissionRate:   DefaultValidatorCommissionRate,
		RewardDenom:               "uve",
		StakeDenom:                "uve",
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate delegations
	for i, del := range gs.Delegations {
		if err := del.Validate(); err != nil {
			return fmt.Errorf("invalid delegation at index %d: %w", i, err)
		}
	}

	// Validate unbonding delegations
	for i, ubd := range gs.UnbondingDelegations {
		if err := ubd.Validate(); err != nil {
			return fmt.Errorf("invalid unbonding delegation at index %d: %w", i, err)
		}
	}

	// Validate redelegations
	for i, red := range gs.Redelegations {
		if err := red.Validate(); err != nil {
			return fmt.Errorf("invalid redelegation at index %d: %w", i, err)
		}
	}

	// Validate validator shares
	for i, vs := range gs.ValidatorShares {
		if err := vs.Validate(); err != nil {
			return fmt.Errorf("invalid validator shares at index %d: %w", i, err)
		}
	}

	// Validate delegator rewards
	for i, dr := range gs.DelegatorRewards {
		if err := dr.Validate(); err != nil {
			return fmt.Errorf("invalid delegator reward at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.UnbondingPeriod <= 0 {
		return fmt.Errorf("unbonding_period must be positive: %d", p.UnbondingPeriod)
	}

	if p.MaxValidatorsPerDelegator <= 0 {
		return fmt.Errorf("max_validators_per_delegator must be positive: %d", p.MaxValidatorsPerDelegator)
	}

	if p.MinDelegationAmount <= 0 {
		return fmt.Errorf("min_delegation_amount must be positive: %d", p.MinDelegationAmount)
	}

	if p.MaxRedelegations <= 0 {
		return fmt.Errorf("max_redelegations must be positive: %d", p.MaxRedelegations)
	}

	if p.ValidatorCommissionRate < 0 || p.ValidatorCommissionRate > BasisPointsMax {
		return fmt.Errorf("validator_commission_rate must be between 0 and %d: %d", BasisPointsMax, p.ValidatorCommissionRate)
	}

	if p.RewardDenom == "" {
		return fmt.Errorf("reward_denom cannot be empty")
	}

	if p.StakeDenom == "" {
		return fmt.Errorf("stake_denom cannot be empty")
	}

	return nil
}
