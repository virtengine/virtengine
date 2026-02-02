// Package v1 provides additional methods for generated settlement types.
package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdk.Msg interface methods for MsgCreateEscrow

func (msg *MsgCreateEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderId == "" {
		return ErrInvalidEscrow.Wrap("order_id cannot be empty")
	}

	// Convert to sdk.Coins for validation
	amount := sdk.NewCoins(msg.Amount...)
	if !amount.IsValid() || amount.IsZero() {
		return ErrInvalidAmount.Wrap("amount must be valid and non-zero")
	}

	if msg.ExpiresIn == 0 {
		return ErrInvalidEscrow.Wrap("expires_in must be greater than zero")
	}

	return nil
}

func (msg *MsgCreateEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgActivateEscrow

func (msg *MsgActivateEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.LeaseId == "" {
		return ErrInvalidEscrow.Wrap("lease_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return ErrInvalidAddress.Wrap("invalid recipient address")
	}

	return nil
}

func (msg *MsgActivateEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgReleaseEscrow

func (msg *MsgReleaseEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if len(msg.Amount) > 0 {
		amount := sdk.NewCoins(msg.Amount...)
		if !amount.IsValid() {
			return ErrInvalidAmount.Wrap("amount must be valid")
		}
	}

	return nil
}

func (msg *MsgReleaseEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgRefundEscrow

func (msg *MsgRefundEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidEscrow.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg *MsgRefundEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgDisputeEscrow

func (msg *MsgDisputeEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.EscrowId == "" {
		return ErrInvalidEscrow.Wrap("escrow_id cannot be empty")
	}

	if msg.Reason == "" {
		return ErrInvalidEscrow.Wrap("reason cannot be empty")
	}

	return nil
}

func (msg *MsgDisputeEscrow) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgSettleOrder

func (msg *MsgSettleOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderId == "" {
		return ErrInvalidSettlement.Wrap("order_id cannot be empty")
	}

	return nil
}

func (msg *MsgSettleOrder) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgRecordUsage

func (msg *MsgRecordUsage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.OrderId == "" {
		return ErrInvalidUsageRecord.Wrap("order_id cannot be empty")
	}

	if msg.LeaseId == "" {
		return ErrInvalidUsageRecord.Wrap("lease_id cannot be empty")
	}

	if msg.UsageUnits == 0 {
		return ErrInvalidUsageRecord.Wrap("usage_units must be greater than zero")
	}

	if msg.UsageType == "" {
		return ErrInvalidUsageRecord.Wrap("usage_type cannot be empty")
	}

	if msg.PeriodEnd <= msg.PeriodStart {
		return ErrInvalidUsageRecord.Wrap("period_end must be after period_start")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidSignature.Wrap("signature cannot be empty")
	}

	return nil
}

func (msg *MsgRecordUsage) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgAcknowledgeUsage

func (msg *MsgAcknowledgeUsage) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	if msg.UsageId == "" {
		return ErrInvalidUsageRecord.Wrap("usage_id cannot be empty")
	}

	if len(msg.Signature) == 0 {
		return ErrInvalidSignature.Wrap("signature cannot be empty")
	}

	return nil
}

func (msg *MsgAcknowledgeUsage) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// sdk.Msg interface methods for MsgClaimRewards

func (msg *MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Source is optional, no validation needed if empty
	return nil
}

func (msg *MsgClaimRewards) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}
