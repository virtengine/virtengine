// Package types contains types for the staking module.
//
// VE-921: Message types for staking module
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Error message constants to avoid duplication
const (
	errMsgInvalidAuthority = "invalid authority address: %v"
	errMsgInvalidValidator = "invalid validator address: %v"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSlashValidator{}
	_ sdk.Msg = &MsgUnjailValidator{}
	_ sdk.Msg = &MsgRecordPerformance{}
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
		return ErrInvalidValidator.Wrapf(errMsgInvalidAuthority, err)
	}
	return m.Params.Validate()
}

// GetSigners returns the expected signers
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// MsgSlashValidator is the message for slashing a validator
type MsgSlashValidator struct {
	// Authority is the address that controls the module
	Authority string `json:"authority"`

	// ValidatorAddress is the validator to slash
	ValidatorAddress string `json:"validator_address"`

	// Reason is the slashing reason
	Reason SlashReason `json:"reason"`

	// InfractionHeight is when the infraction occurred
	InfractionHeight int64 `json:"infraction_height"`

	// Evidence is the evidence supporting the slash
	Evidence string `json:"evidence,omitempty"`
}

// NewMsgSlashValidator creates a new MsgSlashValidator
func NewMsgSlashValidator(authority, validatorAddr string, reason SlashReason, infractionHeight int64, evidence string) *MsgSlashValidator {
	return &MsgSlashValidator{
		Authority:        authority,
		ValidatorAddress: validatorAddr,
		Reason:           reason,
		InfractionHeight: infractionHeight,
		Evidence:         evidence,
	}
}

// ValidateBasic performs basic validation
func (m *MsgSlashValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidAuthority, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}

	if !IsValidSlashReason(m.Reason) {
		return ErrInvalidSlashReason.Wrapf("invalid slash reason: %s", m.Reason)
	}

	if m.InfractionHeight < 0 {
		return ErrInvalidParams.Wrap("infraction_height cannot be negative")
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgSlashValidator) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// MsgUnjailValidator is the message for unjailing a validator
type MsgUnjailValidator struct {
	// ValidatorAddress is the validator to unjail
	ValidatorAddress string `json:"validator_address"`
}

// NewMsgUnjailValidator creates a new MsgUnjailValidator
func NewMsgUnjailValidator(validatorAddr string) *MsgUnjailValidator {
	return &MsgUnjailValidator{
		ValidatorAddress: validatorAddr,
	}
}

// ValidateBasic performs basic validation
func (m *MsgUnjailValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidValidator.Wrapf(errMsgInvalidValidator, err)
	}
	return nil
}

// GetSigners returns the expected signers
func (m *MsgUnjailValidator) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.ValidatorAddress)
	return []sdk.AccAddress{addr}
}

// MsgRecordPerformance is the message for recording validator performance
type MsgRecordPerformance struct {
	// Authority is the address that controls the module
	Authority string `json:"authority"`

	// ValidatorAddress is the validator
	ValidatorAddress string `json:"validator_address"`

	// BlocksProposed is the number of blocks proposed
	BlocksProposed int64 `json:"blocks_proposed"`

	// BlocksSigned is the number of blocks signed
	BlocksSigned int64 `json:"blocks_signed"`

	// VEIDVerificationsCompleted is VEID verifications completed
	VEIDVerificationsCompleted int64 `json:"veid_verifications_completed"`

	// VEIDVerificationScore is the VEID verification quality score
	VEIDVerificationScore int64 `json:"veid_verification_score"`
}

// NewMsgRecordPerformance creates a new MsgRecordPerformance
func NewMsgRecordPerformance(
	authority string,
	validatorAddr string,
	blocksProposed int64,
	blocksSigned int64,
	veidCompleted int64,
	veidScore int64,
) *MsgRecordPerformance {
	return &MsgRecordPerformance{
		Authority:                  authority,
		ValidatorAddress:           validatorAddr,
		BlocksProposed:             blocksProposed,
		BlocksSigned:               blocksSigned,
		VEIDVerificationsCompleted: veidCompleted,
		VEIDVerificationScore:      veidScore,
	}
}

// ValidateBasic performs basic validation
func (m *MsgRecordPerformance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidValidator.Wrapf("invalid authority address: %v", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.ValidatorAddress); err != nil {
		return ErrInvalidValidator.Wrapf("invalid validator address: %v", err)
	}

	if m.BlocksProposed < 0 {
		return ErrInvalidPerformanceMetric.Wrap("blocks_proposed cannot be negative")
	}

	if m.BlocksSigned < 0 {
		return ErrInvalidPerformanceMetric.Wrap("blocks_signed cannot be negative")
	}

	if m.VEIDVerificationsCompleted < 0 {
		return ErrInvalidPerformanceMetric.Wrap("veid_verifications_completed cannot be negative")
	}

	if m.VEIDVerificationScore < 0 || m.VEIDVerificationScore > MaxPerformanceScore {
		return ErrInvalidPerformanceMetric.Wrapf("veid_verification_score must be between 0 and %d", MaxPerformanceScore)
	}

	return nil
}

// GetSigners returns the expected signers
func (m *MsgRecordPerformance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
