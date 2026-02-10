// Package types contains types for the staking module.
//
// VE-921: Message types for staking module
// This file provides constructor functions and SDK interface methods for staking messages.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingv1 "github.com/virtengine/virtengine/sdk/go/node/staking/v1"
)

// Verify generated types implement sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSlashValidator{}
	_ sdk.Msg = &MsgUnjailValidator{}
	_ sdk.Msg = &MsgRecordPerformance{}
)

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params stakingv1.Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// NewMsgSlashValidator creates a new MsgSlashValidator
func NewMsgSlashValidator(authority, validatorAddr string, reason stakingv1.SlashReason, infractionHeight int64, evidence string) *MsgSlashValidator {
	return &MsgSlashValidator{
		Authority:        authority,
		ValidatorAddress: validatorAddr,
		Reason:           reason,
		InfractionHeight: infractionHeight,
		Evidence:         evidence,
	}
}

// NewMsgUnjailValidator creates a new MsgUnjailValidator
func NewMsgUnjailValidator(validatorAddr string) *MsgUnjailValidator {
	return &MsgUnjailValidator{
		ValidatorAddress: validatorAddr,
	}
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
