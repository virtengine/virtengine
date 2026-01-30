// Package types contains types for the delegation module.
//
// VE-922: Delegation module messages
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	delegationv1 "github.com/virtengine/virtengine/sdk/go/node/delegation/v1"
)

// Type aliases to generated protobuf types
type (
	MsgUpdateParams            = delegationv1.MsgUpdateParams
	MsgUpdateParamsResponse    = delegationv1.MsgUpdateParamsResponse
	MsgDelegate                = delegationv1.MsgDelegate
	MsgDelegateResponse        = delegationv1.MsgDelegateResponse
	MsgUndelegate              = delegationv1.MsgUndelegate
	MsgUndelegateResponse      = delegationv1.MsgUndelegateResponse
	MsgRedelegate              = delegationv1.MsgRedelegate
	MsgRedelegateResponse      = delegationv1.MsgRedelegateResponse
	MsgClaimRewards            = delegationv1.MsgClaimRewards
	MsgClaimRewardsResponse    = delegationv1.MsgClaimRewardsResponse
	MsgClaimAllRewards         = delegationv1.MsgClaimAllRewards
	MsgClaimAllRewardsResponse = delegationv1.MsgClaimAllRewardsResponse
)

// Message type constants
const (
	TypeMsgUpdateParams   = "update_params"
	TypeMsgDelegate       = "delegate"
	TypeMsgUndelegate     = "undelegate"
	TypeMsgRedelegate     = "redelegate"
	TypeMsgClaimRewards   = "claim_rewards"
	TypeMsgClaimAllRewards = "claim_all_rewards"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgUndelegate{}
	_ sdk.Msg = &MsgRedelegate{}
	_ sdk.Msg = &MsgClaimRewards{}
	_ sdk.Msg = &MsgClaimAllRewards{}
)

// NewMsgUpdateParams creates a new MsgUpdateParams
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// NewMsgDelegate creates a new MsgDelegate
func NewMsgDelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		Delegator: delegatorAddr,
		Validator: validatorAddr,
		Amount:    amount,
	}
}

// NewMsgUndelegate creates a new MsgUndelegate
func NewMsgUndelegate(delegatorAddr, validatorAddr string, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		Delegator: delegatorAddr,
		Validator: validatorAddr,
		Amount:    amount,
	}
}

// NewMsgRedelegate creates a new MsgRedelegate
func NewMsgRedelegate(delegatorAddr, srcValidatorAddr, dstValidatorAddr string, amount sdk.Coin) *MsgRedelegate {
	return &MsgRedelegate{
		Delegator:    delegatorAddr,
		SrcValidator: srcValidatorAddr,
		DstValidator: dstValidatorAddr,
		Amount:       amount,
	}
}

// NewMsgClaimRewards creates a new MsgClaimRewards
func NewMsgClaimRewards(delegatorAddr, validatorAddr string) *MsgClaimRewards {
	return &MsgClaimRewards{
		Delegator: delegatorAddr,
		Validator: validatorAddr,
	}
}

// NewMsgClaimAllRewards creates a new MsgClaimAllRewards
func NewMsgClaimAllRewards(delegatorAddr string) *MsgClaimAllRewards {
	return &MsgClaimAllRewards{
		Delegator: delegatorAddr,
	}
}
