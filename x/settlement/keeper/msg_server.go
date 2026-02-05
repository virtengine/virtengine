package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

type msgServer struct {
	keeper IKeeper
}

// NewMsgServerImpl returns an implementation of the settlement MsgServer interface
func NewMsgServerImpl(k IKeeper) settlementv1.MsgServer {
	return &msgServer{keeper: k}
}

var _ settlementv1.MsgServer = msgServer{}

// CreateEscrow handles creating a new escrow account
func (ms msgServer) CreateEscrow(goCtx context.Context, msg *types.MsgCreateEscrow) (*types.MsgCreateEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	expiresIn, err := durationFromSeconds(msg.ExpiresIn)
	if err != nil {
		return nil, types.ErrInvalidParams.Wrap(err.Error())
	}
	amount := sdk.NewCoins(msg.Amount...)

	escrowID, err := ms.keeper.CreateEscrow(ctx, msg.OrderId, sender, amount, expiresIn, nil)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateEscrowResponse{
		EscrowId:  escrowID,
		CreatedAt: ctx.BlockTime().Unix(),
	}, nil
}

func durationFromSeconds(seconds uint64) (time.Duration, error) {
	maxSeconds := uint64(^uint64(0)>>1) / uint64(time.Second)
	if seconds > maxSeconds {
		return 0, fmt.Errorf("duration out of range: %d seconds", seconds)
	}
	return time.Duration(seconds) * time.Second, nil
}

// ActivateEscrow handles activating an escrow
func (ms msgServer) ActivateEscrow(goCtx context.Context, msg *types.MsgActivateEscrow) (*types.MsgActivateEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate sender has authority (typically the market module or governance)
	// In production, you would check if sender is authorized to activate escrows

	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid recipient address")
	}

	if err := ms.keeper.ActivateEscrow(ctx, msg.EscrowId, msg.LeaseId, recipient); err != nil {
		return nil, err
	}

	return &types.MsgActivateEscrowResponse{
		ActivatedAt: ctx.BlockTime().Unix(),
	}, nil
}

// ReleaseEscrow handles releasing an escrow
func (ms msgServer) ReleaseEscrow(goCtx context.Context, msg *types.MsgReleaseEscrow) (*types.MsgReleaseEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is authorized (depositor or governance)
	escrow, found := ms.keeper.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", msg.EscrowId)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	if !sender.Equals(depositor) && sender.String() != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("only depositor or governance can release escrow")
	}

	balanceBefore := escrow.Balance

	if err := ms.keeper.ReleaseEscrow(ctx, msg.EscrowId, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgReleaseEscrowResponse{
		ReleasedAmount: balanceBefore.String(),
		ReleasedAt:     ctx.BlockTime().Unix(),
	}, nil
}

// RefundEscrow handles refunding an escrow
func (ms msgServer) RefundEscrow(goCtx context.Context, msg *types.MsgRefundEscrow) (*types.MsgRefundEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is authorized
	escrow, found := ms.keeper.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", msg.EscrowId)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	// Only depositor, recipient (provider), or governance can request refund
	if !sender.Equals(depositor) && !sender.Equals(recipient) && sender.String() != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("not authorized to refund escrow")
	}

	balanceBefore := escrow.Balance

	if err := ms.keeper.RefundEscrow(ctx, msg.EscrowId, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgRefundEscrowResponse{
		RefundedAmount: balanceBefore.String(),
		RefundedAt:     ctx.BlockTime().Unix(),
	}, nil
}

// DisputeEscrow handles disputing an escrow
func (ms msgServer) DisputeEscrow(goCtx context.Context, msg *types.MsgDisputeEscrow) (*types.MsgDisputeEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is party to the escrow
	escrow, found := ms.keeper.GetEscrow(ctx, msg.EscrowId)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("escrow %s not found", msg.EscrowId)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	if !sender.Equals(depositor) && !sender.Equals(recipient) {
		return nil, types.ErrUnauthorized.Wrap("only parties to escrow can file dispute")
	}

	if err := ms.keeper.DisputeEscrow(ctx, msg.EscrowId, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgDisputeEscrowResponse{
		DisputedAt: ctx.BlockTime().Unix(),
	}, nil
}

// SettleOrder handles settling an order
func (ms msgServer) SettleOrder(goCtx context.Context, msg *types.MsgSettleOrder) (*types.MsgSettleOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Validate sender is authorized (provider, customer, or governance)
	escrow, found := ms.keeper.GetEscrowByOrder(ctx, msg.OrderId)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("no escrow found for order %s", msg.OrderId)
	}

	depositor, _ := sdk.AccAddressFromBech32(escrow.Depositor)
	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	if !sender.Equals(depositor) && !sender.Equals(recipient) && sender.String() != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("not authorized to settle order")
	}

	settlement, err := ms.keeper.SettleOrder(ctx, msg.OrderId, msg.UsageRecordIds, msg.IsFinal)
	if err != nil {
		return nil, err
	}

	return &types.MsgSettleOrderResponse{
		SettlementId:  settlement.SettlementID,
		TotalAmount:   settlement.TotalAmount.String(),
		ProviderShare: settlement.ProviderShare.String(),
		PlatformFee:   settlement.PlatformFee.String(),
		SettledAt:     settlement.SettledAt.Unix(),
	}, nil
}

// RecordUsage handles recording usage from a provider
func (ms msgServer) RecordUsage(goCtx context.Context, msg *types.MsgRecordUsage) (*types.MsgRecordUsageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get escrow to validate provider
	escrow, found := ms.keeper.GetEscrowByOrder(ctx, msg.OrderId)
	if !found {
		return nil, types.ErrEscrowNotFound.Wrapf("no escrow found for order %s", msg.OrderId)
	}

	recipient, _ := sdk.AccAddressFromBech32(escrow.Recipient)
	if !sender.Equals(recipient) {
		return nil, types.ErrUnauthorized.Wrap("only the provider can record usage")
	}

	// Create usage record
	record := types.NewUsageRecord(
		"", // ID will be generated
		msg.OrderId,
		msg.LeaseId,
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

	if err := ms.keeper.RecordUsage(ctx, record); err != nil {
		return nil, err
	}

	return &types.MsgRecordUsageResponse{
		UsageId:    record.UsageID,
		TotalCost:  record.TotalCost.String(),
		RecordedAt: ctx.BlockTime().Unix(),
	}, nil
}

// AcknowledgeUsage handles customer acknowledgment of usage
func (ms msgServer) AcknowledgeUsage(goCtx context.Context, msg *types.MsgAcknowledgeUsage) (*types.MsgAcknowledgeUsageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrInvalidAddress.Wrap("invalid sender address")
	}

	// Get usage record to validate customer
	usage, found := ms.keeper.GetUsageRecord(ctx, msg.UsageId)
	if !found {
		return nil, types.ErrUsageRecordNotFound.Wrapf("usage record %s not found", msg.UsageId)
	}

	customer, _ := sdk.AccAddressFromBech32(usage.Customer)
	if !sender.Equals(customer) {
		return nil, types.ErrUnauthorized.Wrap("only the customer can acknowledge usage")
	}

	if err := ms.keeper.AcknowledgeUsage(ctx, msg.UsageId, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgAcknowledgeUsageResponse{
		AcknowledgedAt: ctx.BlockTime().Unix(),
	}, nil
}

// ClaimRewards handles claiming accumulated rewards
func (ms msgServer) ClaimRewards(goCtx context.Context, msg *types.MsgClaimRewards) (*types.MsgClaimRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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
