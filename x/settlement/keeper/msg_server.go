package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the settlement MsgServer interface
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

var _ types.MsgServer = msgServer{}

// CreateEscrow handles creating a new escrow account
func (ms msgServer) CreateEscrow(ctx sdk.Context, msg *types.MsgCreateEscrow) (*types.MsgCreateEscrowResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	expiresIn := time.Duration(msg.ExpiresIn) * time.Second

	escrowID, err := ms.keeper.CreateEscrow(ctx, msg.OrderID, sender, msg.Amount, expiresIn, msg.Conditions)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateEscrowResponse{
		EscrowID:  escrowID,
		CreatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// ActivateEscrow handles activating an escrow
func (ms msgServer) ActivateEscrow(ctx sdk.Context, msg *types.MsgActivateEscrow) (*types.MsgActivateEscrowResponse, error) {
	// Validate sender has authority (typically the market module or governance)
	// In production, you would check if sender is authorized to activate escrows

	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid recipient address")
	}

	if err := ms.keeper.ActivateEscrow(ctx, msg.EscrowID, msg.LeaseID, recipient); err != nil {
		return nil, err
	}

	return &types.MsgActivateEscrowResponse{
		ActivatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// ReleaseEscrow handles releasing an escrow
func (ms msgServer) ReleaseEscrow(ctx sdk.Context, msg *types.MsgReleaseEscrow) (*types.MsgReleaseEscrowResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is authorized (depositor or governance)
	escrow, found := ms.keeper.GetEscrow(ctx, msg.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", msg.EscrowID)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	if !sender.Equals(depositor) && sender.String() != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("only depositor or governance can release escrow")
	}

	balanceBefore := escrow.Balance

	if err := ms.keeper.ReleaseEscrow(ctx, msg.EscrowID, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgReleaseEscrowResponse{
		ReleasedAmount: balanceBefore.String(),
		ReleasedAt:     ctx.BlockTime().Unix(),
	}, nil
}

// RefundEscrow handles refunding an escrow
func (ms msgServer) RefundEscrow(ctx sdk.Context, msg *types.MsgRefundEscrow) (*types.MsgRefundEscrowResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is authorized
	escrow, found := ms.keeper.GetEscrow(ctx, msg.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", msg.EscrowID)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	// Only depositor, recipient (provider), or governance can request refund
	if !sender.Equals(depositor) && !sender.Equals(recipient) && sender.String() != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("not authorized to refund escrow")
	}

	balanceBefore := escrow.Balance

	if err := ms.keeper.RefundEscrow(ctx, msg.EscrowID, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgRefundEscrowResponse{
		RefundedAmount: balanceBefore.String(),
		RefundedAt:     ctx.BlockTime().Unix(),
	}, nil
}

// DisputeEscrow handles disputing an escrow
func (ms msgServer) DisputeEscrow(ctx sdk.Context, msg *types.MsgDisputeEscrow) (*types.MsgDisputeEscrowResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is party to the escrow
	escrow, found := ms.keeper.GetEscrow(ctx, msg.EscrowID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", msg.EscrowID)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	if !sender.Equals(depositor) && !sender.Equals(recipient) {
		return nil, types.ErrUnauthorized.Wrap("only parties to escrow can file dispute")
	}

	if err := ms.keeper.DisputeEscrow(ctx, msg.EscrowID, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgDisputeEscrowResponse{
		DisputedAt: ctx.BlockTime().Unix(),
	}, nil
}

// SettleOrder handles settling an order
func (ms msgServer) SettleOrder(ctx sdk.Context, msg *types.MsgSettleOrder) (*types.MsgSettleOrderResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is authorized (provider, customer, or governance)
	escrow, found := ms.keeper.GetEscrowByOrder(ctx, msg.OrderID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("no escrow found for order %s", msg.OrderID)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	if !sender.Equals(depositor) && !sender.Equals(recipient) && sender.String() != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("not authorized to settle order")
	}

	settlement, err := ms.keeper.SettleOrder(ctx, msg.OrderID, msg.UsageRecordIDs, msg.IsFinal)
	if err != nil {
		return nil, err
	}

	return &types.MsgSettleOrderResponse{
		SettlementID:  settlement.SettlementID,
		TotalAmount:   settlement.TotalAmount.String(),
		ProviderShare: settlement.ProviderShare.String(),
		PlatformFee:   settlement.PlatformFee.String(),
		SettledAt:     settlement.SettledAt.Unix(),
	}, nil
}

// RecordUsage handles recording usage from a provider
func (ms msgServer) RecordUsage(ctx sdk.Context, msg *types.MsgRecordUsage) (*types.MsgRecordUsageResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get escrow to validate provider
	escrow, found := ms.keeper.GetEscrowByOrder(ctx, msg.OrderID)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("no escrow found for order %s", msg.OrderID)
	}

	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)
	if !sender.Equals(recipient) {
		return nil, types.ErrUnauthorized.Wrap("only the provider can record usage")
	}

	// Create usage record
	record := types.NewUsageRecord(
		"", // ID will be generated
		msg.OrderID,
		msg.LeaseID,
		msg.Sender,
		escrow.Depositor,
		msg.UsageUnits,
		msg.UsageType,
		time.Unix(msg.PeriodStart, 0),
		time.Unix(msg.PeriodEnd, 0),
		msg.UnitPrice,
		msg.Signature,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	if msg.Metadata != nil {
		record.Metadata = msg.Metadata
	}

	if err := ms.keeper.RecordUsage(ctx, record); err != nil {
		return nil, err
	}

	return &types.MsgRecordUsageResponse{
		UsageID:    record.UsageID,
		TotalCost:  record.TotalCost.String(),
		RecordedAt: ctx.BlockTime().Unix(),
	}, nil
}

// AcknowledgeUsage handles customer acknowledgment of usage
func (ms msgServer) AcknowledgeUsage(ctx sdk.Context, msg *types.MsgAcknowledgeUsage) (*types.MsgAcknowledgeUsageResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get usage record to validate customer
	usage, found := ms.keeper.GetUsageRecord(ctx, msg.UsageID)
	if !found {
		return nil, types.ErrUsageRecordNotFound.Wrapf("usage record %s not found", msg.UsageID)
	}

	customer, _ := sdk.AccAddressFromBech32(usage.Customer)
	if !sender.Equals(customer) {
		return nil, types.ErrUnauthorized.Wrap("only the customer can acknowledge usage")
	}

	if err := ms.keeper.AcknowledgeUsage(ctx, msg.UsageID, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgAcknowledgeUsageResponse{
		AcknowledgedAt: ctx.BlockTime().Unix(),
	}, nil
}

// ClaimRewards handles claiming accumulated rewards
func (ms msgServer) ClaimRewards(ctx sdk.Context, msg *types.MsgClaimRewards) (*types.MsgClaimRewardsResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	claimed, err := ms.keeper.ClaimRewards(ctx, sender, msg.Source)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimRewardsResponse{
		ClaimedAmount: claimed.String(),
		ClaimedAt:     ctx.BlockTime().Unix(),
	}, nil
}
