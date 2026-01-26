// Package types contains types for the delegation module.
//
// VE-922: Delegation module messages
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Error message constants
const (
	errMsgInvalidAuthority  = "invalid authority address: %v"
	errMsgInvalidDelegator  = "invalid delegator address: %v"
	errMsgInvalidValidator  = "invalid validator address: %v"
	errMsgInvalidAmount     = "invalid amount: %v"
	errMsgSameValidator     = "source and destination validators cannot be the same"
	errMsgNegativeAmount    = "amount must be positive"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgUndelegate{}
	_ sdk.Msg = &MsgRedelegate{}
	_ sdk.Msg = &MsgClaimRewards{}
	_ sdk.Msg = &MsgClaimAllRewards{}
)

// MsgUpdateParams is the message for updating module parameters
type MsgUpdateParams struct {
	// Authority is the address that controls the module
	Authority string `json:"authority"`

	// Params are the new parameters
	Params Params `json:"params"`
}

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic performs basic validation
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidDelegator.Wrapf(errMsgInvalidAuthority, err)
	}
	return m.Params.Validate()
}

// GetSigners returns the expected signers
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// MsgDelegate is the message for delegating tokens to a validator
type MsgDelegate struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// Amount is the amount to delegate
	Amount sdk.Coin `json:"amount"`
}

// NewMsgDelegate creates a new MsgDelegate
func NewMsgDelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           amount,
	}
}

// ValidateBasic performs basic validation
func (m *MsgDelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.DelegatorAddress); err != nil {
		return ErrInvalidDelegator.Wrapf(errMsgInvalidDelegator, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}

	if !m.Amount.IsValid() {
		return ErrInvalidAmount.Wrapf(errMsgInvalidAmount, m.Amount)
	}

	if !m.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap(errMsgNegativeAmount)
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgDelegate) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.DelegatorAddress)
	return []sdk.AccAddress{addr}
}

// MsgUndelegate is the message for undelegating tokens from a validator
type MsgUndelegate struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`

	// Amount is the amount to undelegate
	Amount sdk.Coin `json:"amount"`
}

// NewMsgUndelegate creates a new MsgUndelegate
func NewMsgUndelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Amount:           amount,
	}
}

// ValidateBasic performs basic validation
func (m *MsgUndelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.DelegatorAddress); err != nil {
		return ErrInvalidDelegator.Wrapf(errMsgInvalidDelegator, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}

	if !m.Amount.IsValid() {
		return ErrInvalidAmount.Wrapf(errMsgInvalidAmount, m.Amount)
	}

	if !m.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap(errMsgNegativeAmount)
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgUndelegate) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.DelegatorAddress)
	return []sdk.AccAddress{addr}
}

// MsgRedelegate is the message for redelegating tokens between validators
type MsgRedelegate struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorSrcAddress is the source validator's address
	ValidatorSrcAddress string `json:"validator_src_address"`

	// ValidatorDstAddress is the destination validator's address
	ValidatorDstAddress string `json:"validator_dst_address"`

	// Amount is the amount to redelegate
	Amount sdk.Coin `json:"amount"`
}

// NewMsgRedelegate creates a new MsgRedelegate
func NewMsgRedelegate(delegatorAddr, srcValidatorAddr, dstValidatorAddr string, amount sdk.Coin) *MsgRedelegate {
	return &MsgRedelegate{
		DelegatorAddress:    delegatorAddr,
		ValidatorSrcAddress: srcValidatorAddr,
		ValidatorDstAddress: dstValidatorAddr,
		Amount:              amount,
	}
}

// ValidateBasic performs basic validation
func (m *MsgRedelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.DelegatorAddress); err != nil {
		return ErrInvalidDelegator.Wrapf(errMsgInvalidDelegator, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorSrcAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorDstAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}

	if m.ValidatorSrcAddress == m.ValidatorDstAddress {
		return ErrSelfRedelegation.Wrap(errMsgSameValidator)
	}

	if !m.Amount.IsValid() {
		return ErrInvalidAmount.Wrapf(errMsgInvalidAmount, m.Amount)
	}

	if !m.Amount.IsPositive() {
		return ErrInvalidAmount.Wrap(errMsgNegativeAmount)
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgRedelegate) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.DelegatorAddress)
	return []sdk.AccAddress{addr}
}

// MsgClaimRewards is the message for claiming delegation rewards from a validator
type MsgClaimRewards struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`

	// ValidatorAddress is the validator's address
	ValidatorAddress string `json:"validator_address"`
}

// NewMsgClaimRewards creates a new MsgClaimRewards
func NewMsgClaimRewards(delegatorAddr, validatorAddr string) *MsgClaimRewards {
	return &MsgClaimRewards{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
	}
}

// ValidateBasic performs basic validation
func (m *MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.DelegatorAddress); err != nil {
		return ErrInvalidDelegator.Wrapf(errMsgInvalidDelegator, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgClaimRewards) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.DelegatorAddress)
	return []sdk.AccAddress{addr}
}

// MsgClaimAllRewards is the message for claiming all delegation rewards
type MsgClaimAllRewards struct {
	// DelegatorAddress is the delegator's address
	DelegatorAddress string `json:"delegator_address"`
}

// NewMsgClaimAllRewards creates a new MsgClaimAllRewards
func NewMsgClaimAllRewards(delegatorAddr string) *MsgClaimAllRewards {
	return &MsgClaimAllRewards{
		DelegatorAddress: delegatorAddr,
	}
}

// ValidateBasic performs basic validation
func (m *MsgClaimAllRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.DelegatorAddress); err != nil {
		return ErrInvalidDelegator.Wrapf(errMsgInvalidDelegator, err)
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgClaimAllRewards) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.DelegatorAddress)
	return []sdk.AccAddress{addr}
}
