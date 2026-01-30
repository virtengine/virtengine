package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

// Type aliases to generated protobuf types
type (
	MsgCreateEscrow             = settlementv1.MsgCreateEscrow
	MsgCreateEscrowResponse     = settlementv1.MsgCreateEscrowResponse
	MsgActivateEscrow           = settlementv1.MsgActivateEscrow
	MsgActivateEscrowResponse   = settlementv1.MsgActivateEscrowResponse
	MsgReleaseEscrow            = settlementv1.MsgReleaseEscrow
	MsgReleaseEscrowResponse    = settlementv1.MsgReleaseEscrowResponse
	MsgRefundEscrow             = settlementv1.MsgRefundEscrow
	MsgRefundEscrowResponse     = settlementv1.MsgRefundEscrowResponse
	MsgDisputeEscrow            = settlementv1.MsgDisputeEscrow
	MsgDisputeEscrowResponse    = settlementv1.MsgDisputeEscrowResponse
	MsgSettleOrder              = settlementv1.MsgSettleOrder
	MsgSettleOrderResponse      = settlementv1.MsgSettleOrderResponse
	MsgRecordUsage              = settlementv1.MsgRecordUsage
	MsgRecordUsageResponse      = settlementv1.MsgRecordUsageResponse
	MsgAcknowledgeUsage         = settlementv1.MsgAcknowledgeUsage
	MsgAcknowledgeUsageResponse = settlementv1.MsgAcknowledgeUsageResponse
	MsgClaimRewards             = settlementv1.MsgClaimRewards
	MsgClaimRewardsResponse     = settlementv1.MsgClaimRewardsResponse
)

// Message type constants
const (
	TypeMsgCreateEscrow      = "create_escrow"
	TypeMsgActivateEscrow    = "activate_escrow"
	TypeMsgReleaseEscrow     = "release_escrow"
	TypeMsgRefundEscrow      = "refund_escrow"
	TypeMsgDisputeEscrow     = "dispute_escrow"
	TypeMsgSettleOrder       = "settle_order"
	TypeMsgRecordUsage       = "record_usage"
	TypeMsgAcknowledgeUsage  = "acknowledge_usage"
	TypeMsgClaimRewards      = "claim_rewards"
	TypeMsgDistributeRewards = "distribute_rewards"
)

var (
	_ sdk.Msg = &MsgCreateEscrow{}
	_ sdk.Msg = &MsgActivateEscrow{}
	_ sdk.Msg = &MsgReleaseEscrow{}
	_ sdk.Msg = &MsgRefundEscrow{}
	_ sdk.Msg = &MsgDisputeEscrow{}
	_ sdk.Msg = &MsgSettleOrder{}
	_ sdk.Msg = &MsgRecordUsage{}
	_ sdk.Msg = &MsgAcknowledgeUsage{}
	_ sdk.Msg = &MsgClaimRewards{}
)

// NewMsgCreateEscrow creates a new MsgCreateEscrow
func NewMsgCreateEscrow(sender, orderID string, amount sdk.Coins, expiresIn uint64) *MsgCreateEscrow {
	return &MsgCreateEscrow{
		Sender:    sender,
		OrderId:   orderID,
		Amount:    amount,
		ExpiresIn: expiresIn,
	}
}

// NewMsgActivateEscrow creates a new MsgActivateEscrow
func NewMsgActivateEscrow(sender, escrowID, leaseID, recipient string) *MsgActivateEscrow {
	return &MsgActivateEscrow{
		Sender:    sender,
		EscrowId:  escrowID,
		LeaseId:   leaseID,
		Recipient: recipient,
	}
}

// NewMsgReleaseEscrow creates a new MsgReleaseEscrow
func NewMsgReleaseEscrow(sender, escrowID string, amount sdk.Coins, reason string) *MsgReleaseEscrow {
	return &MsgReleaseEscrow{
		Sender:   sender,
		EscrowId: escrowID,
		Amount:   amount,
		Reason:   reason,
	}
}

// NewMsgRefundEscrow creates a new MsgRefundEscrow
func NewMsgRefundEscrow(sender, escrowID, reason string) *MsgRefundEscrow {
	return &MsgRefundEscrow{
		Sender:   sender,
		EscrowId: escrowID,
		Reason:   reason,
	}
}

// NewMsgDisputeEscrow creates a new MsgDisputeEscrow
func NewMsgDisputeEscrow(sender, escrowID, reason, evidence string) *MsgDisputeEscrow {
	return &MsgDisputeEscrow{
		Sender:   sender,
		EscrowId: escrowID,
		Reason:   reason,
		Evidence: evidence,
	}
}

// NewMsgSettleOrder creates a new MsgSettleOrder
func NewMsgSettleOrder(sender, orderID string, usageRecordIDs []string, isFinal bool) *MsgSettleOrder {
	return &MsgSettleOrder{
		Sender:         sender,
		OrderId:        orderID,
		UsageRecordIds: usageRecordIDs,
		IsFinal:        isFinal,
	}
}

// NewMsgRecordUsage creates a new MsgRecordUsage
func NewMsgRecordUsage(
	sender, orderID, leaseID string,
	usageUnits uint64,
	usageType string,
	periodStart, periodEnd int64,
	unitPrice sdk.DecCoin,
	signature []byte,
) *MsgRecordUsage {
	return &MsgRecordUsage{
		Sender:      sender,
		OrderId:     orderID,
		LeaseId:     leaseID,
		UsageUnits:  usageUnits,
		UsageType:   usageType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		UnitPrice:   unitPrice,
		Signature:   signature,
	}
}

// NewMsgAcknowledgeUsage creates a new MsgAcknowledgeUsage
func NewMsgAcknowledgeUsage(sender, usageID string, signature []byte) *MsgAcknowledgeUsage {
	return &MsgAcknowledgeUsage{
		Sender:    sender,
		UsageId:   usageID,
		Signature: signature,
	}
}

// NewMsgClaimRewards creates a new MsgClaimRewards
func NewMsgClaimRewards(sender, source string) *MsgClaimRewards {
	return &MsgClaimRewards{
		Sender: sender,
		Source: source,
	}
}
