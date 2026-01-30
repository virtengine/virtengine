// Package types contains types for the delegation module.
//
// VE-922: Genesis state and parameters for delegated staking
package types

import (
	"fmt"

	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
)

// Type alias for Params from generated proto
type Params = delegationv1.Params

// DefaultUnbondingPeriod is the default unbonding period in seconds (21 days)
const DefaultUnbondingPeriod uint64 = 21 * 24 * 60 * 60

// DefaultMaxValidators is the default max validators per delegator
const DefaultMaxValidators uint64 = 10

// DefaultMinDelegation is the default minimum delegation amount (1 token)
const DefaultMinDelegation = "1000000"

// DefaultRedelegationCooldown is the default redelegation cooldown in seconds (7 days)
const DefaultRedelegationCooldown uint64 = 7 * 24 * 60 * 60

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
		UnbondingPeriod:      DefaultUnbondingPeriod,
		MaxValidators:        DefaultMaxValidators,
		MinDelegation:        DefaultMinDelegation,
		RedelegationCooldown: DefaultRedelegationCooldown,
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	if err := ValidateParams(&gs.Params); err != nil {
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

// ValidateParams validates the parameters
func ValidateParams(p *Params) error {
	if p.UnbondingPeriod == 0 {
		return fmt.Errorf("unbonding_period must be positive")
	}

	if p.MaxValidators == 0 {
		return fmt.Errorf("max_validators must be positive")
	}

	if p.MinDelegation == "" {
		return fmt.Errorf("min_delegation cannot be empty")
	}

	return nil
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string { return fmt.Sprintf("%+v", *gs) }
