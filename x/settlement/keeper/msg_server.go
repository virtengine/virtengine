package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	if ms.keeper.IsFinancialCasesActive(ctx) {
		evidenceHash := sha256.Sum256([]byte(msg.Evidence))
		respondent := escrow.Recipient
		if sender.String() == escrow.Recipient {
			respondent = escrow.Depositor
		}
		idempotencyHash := sha256.Sum256([]byte("settlement/escrow/financial-case/v1\x00" + msg.EscrowId + "\x00" + sender.String() + "\x00" + hex.EncodeToString(evidenceHash[:])))
		idempotency := idempotencyHash[:]
		_, _, _, err := ms.keeper.OpenFinancialCase(ctx, FinancialCaseOpenRequest{
			Subject:  types.FinancialSubject{Type: types.FinancialSubjectTypeOrder, PrimaryId: escrow.OrderID, OrderId: escrow.OrderID, EscrowId: escrow.EscrowID, LeaseId: escrow.LeaseID},
			Claimant: sender.String(), Respondent: respondent, IdempotencyKey: idempotency,
			Claim: types.FinancialClaim{ClaimType: types.FinancialClaimTypeBilling, Claimant: sender.String(), SourceModule: "settlement", SourceReference: msg.EscrowId, EvidenceHash: evidenceHash[:], EncryptedReference: "settlement://escrow/" + msg.EscrowId + "/evidence-hash/" + hex.EncodeToString(evidenceHash[:]), IdempotencyKey: idempotency},
		})
		if err != nil {
			return nil, err
		}
		return &types.MsgDisputeEscrowResponse{DisputedAt: ctx.BlockTime().Unix()}, nil
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
	record.AllocationID = msg.AllocationId
	record.ChainID = msg.ChainId
	if msg.RawMetrics != nil {
		record.Metrics = types.RawUsageMetrics{
			CPUMilliSeconds:    msg.RawMetrics.CpuMilliSeconds,
			MemoryByteSeconds:  msg.RawMetrics.MemoryByteSeconds,
			StorageByteSeconds: msg.RawMetrics.StorageByteSeconds,
			NetworkBytesIn:     msg.RawMetrics.NetworkBytesIn,
			NetworkBytesOut:    msg.RawMetrics.NetworkBytesOut,
			GPUSeconds:         msg.RawMetrics.GpuSeconds,
		}
	}
	record.PricingVersion = msg.PricingVersion
	record.FormulaVersion = msg.FormulaVersion
	record.ModelVersion = msg.ModelVersion
	record.Sequence = msg.StreamSequence
	record.Nonce = append([]byte(nil), msg.Nonce...)
	record.IdempotencyKey = append([]byte(nil), msg.IdempotencyKey...)
	record.ProviderKeyEpoch = msg.ProviderKeyEpoch
	record.ProviderKeyID = msg.ProviderKeyId
	record.IssuedAtHeight = msg.IssuedAtHeight
	record.ExpiresAtHeight = msg.ExpiresAtHeight
	record.IssuedAtUnix = msg.IssuedAtUnix
	record.ExpiresAtUnix = msg.ExpiresAtUnix
	record.SignatureVersion = msg.SignatureVersion

	if err := ms.keeper.RecordUsage(ctx, record); err != nil {
		return nil, err
	}

	return &types.MsgRecordUsageResponse{
		UsageId:              record.UsageID,
		TotalCost:            record.TotalCost.String(),
		RecordedAt:           ctx.BlockTime().Unix(),
		UsageDigest:          append([]byte(nil), record.UsageDigest...),
		AuthenticationStatus: record.AuthenticationStatus,
		ExactDuplicate:       record.ExactDuplicate,
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

	if msg.SignatureVersion == 0 && !ms.keeper.IsUsageAuthenticationActive(ctx) {
		if err := ms.keeper.AcknowledgeUsage(ctx, msg.UsageId, msg.Signature); err != nil {
			return nil, err
		}
		return &types.MsgAcknowledgeUsageResponse{AcknowledgedAt: ctx.BlockTime().Unix()}, nil
	}

	exactDuplicate := usage.CustomerAcknowledged && bytes.Equal(usage.CustomerAckReplayKey, msg.ReplayKey)
	proof := types.UsageAcknowledgmentProof{
		Signature:        append([]byte(nil), msg.Signature...),
		UsageDigest:      append([]byte(nil), msg.UsageDigest...),
		ReplayKey:        append([]byte(nil), msg.ReplayKey...),
		IssuedAtHeight:   msg.IssuedAtHeight,
		ExpiresAtHeight:  msg.ExpiresAtHeight,
		IssuedAtUnix:     msg.IssuedAtUnix,
		ExpiresAtUnix:    msg.ExpiresAtUnix,
		SignatureVersion: msg.SignatureVersion,
	}
	if err := ms.keeper.AcknowledgeUsageAuthenticated(ctx, msg.UsageId, proof); err != nil {
		return nil, err
	}

	return &types.MsgAcknowledgeUsageResponse{
		AcknowledgedAt: ctx.BlockTime().Unix(),
		ExactDuplicate: exactDuplicate,
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

func (ms msgServer) OpenFinancialCase(goCtx context.Context, msg *types.MsgOpenFinancialCase) (*types.MsgOpenFinancialCaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	claim := types.FinancialClaim{ClaimType: msg.ClaimType, Claimant: msg.Sender, SourceModule: financialSourceSettlement, SourceReference: msg.SourceReference, EvidenceHash: append([]byte(nil), msg.EvidenceHash...), EncryptedReference: msg.EncryptedReference, IdempotencyKey: append([]byte(nil), msg.IdempotencyKey...)}
	financialCase, added, duplicate, err := ms.keeper.OpenFinancialCase(ctx, FinancialCaseOpenRequest{Subject: msg.Subject, Claimant: msg.Sender, Respondent: msg.Respondent, Claim: claim, IdempotencyKey: msg.IdempotencyKey})
	if err != nil {
		return nil, err
	}
	return &types.MsgOpenFinancialCaseResponse{CaseId: financialCase.CaseId, ClaimId: added.ClaimId, ExactDuplicate: duplicate, Status: financialCase.Status}, nil
}

func (ms msgServer) AddFinancialClaim(goCtx context.Context, msg *types.MsgAddFinancialClaim) (*types.MsgAddFinancialClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	claim := types.FinancialClaim{ClaimType: msg.ClaimType, Claimant: msg.Sender, SourceModule: financialSourceSettlement, SourceReference: msg.SourceReference, EvidenceHash: append([]byte(nil), msg.EvidenceHash...), EncryptedReference: msg.EncryptedReference, IdempotencyKey: append([]byte(nil), msg.IdempotencyKey...), Recommendation: msg.Recommendation}
	financialCase, added, duplicate, err := ms.keeper.AddFinancialClaim(ctx, msg.CaseId, claim)
	if err != nil {
		return nil, err
	}
	return &types.MsgAddFinancialClaimResponse{CaseId: financialCase.CaseId, ClaimId: added.ClaimId, ExactDuplicate: duplicate, Status: financialCase.Status}, nil
}

func (ms msgServer) SubmitFinancialCaseForReview(goCtx context.Context, msg *types.MsgSubmitFinancialCaseForReview) (*types.MsgSubmitFinancialCaseForReviewResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.keeper.SubmitFinancialCaseForReview(ctx, msg.CaseId, msg.Sender); err != nil {
		return nil, err
	}
	financialCase, _ := ms.keeper.GetFinancialCase(ctx, msg.CaseId)
	return &types.MsgSubmitFinancialCaseForReviewResponse{Status: financialCase.Status}, nil
}

func (ms msgServer) EscalateFinancialCase(goCtx context.Context, msg *types.MsgEscalateFinancialCase) (*types.MsgEscalateFinancialCaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.keeper.EscalateFinancialCase(ctx, msg.CaseId, msg.Sender, msg.ReasonHash); err != nil {
		return nil, err
	}
	financialCase, _ := ms.keeper.GetFinancialCase(ctx, msg.CaseId)
	return &types.MsgEscalateFinancialCaseResponse{Status: financialCase.Status}, nil
}

func (ms msgServer) ResolveFinancialCase(goCtx context.Context, msg *types.MsgResolveFinancialCase) (*types.MsgResolveFinancialCaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.keeper.ResolveFinancialCase(ctx, msg.CaseId, msg.Resolver, msg.Allocation); err != nil {
		return nil, err
	}
	financialCase, _ := ms.keeper.GetFinancialCase(ctx, msg.CaseId)
	return &types.MsgResolveFinancialCaseResponse{Status: financialCase.Status, AppealDeadlineHeight: financialCase.AppealDeadlineHeight, AppealDeadlineTime: financialCase.AppealDeadlineTime}, nil
}

func (ms msgServer) AppealFinancialCase(goCtx context.Context, msg *types.MsgAppealFinancialCase) (*types.MsgAppealFinancialCaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	appeal, _, err := ms.keeper.AppealFinancialCase(ctx, msg.CaseId, msg.Appellant, msg.EvidenceHash, msg.EncryptedReference, msg.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	financialCase, _ := ms.keeper.GetFinancialCase(ctx, msg.CaseId)
	return &types.MsgAppealFinancialCaseResponse{AppealId: appeal.AppealId, Status: financialCase.Status}, nil
}

func (ms msgServer) CancelFinancialCase(goCtx context.Context, msg *types.MsgCancelFinancialCase) (*types.MsgCancelFinancialCaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.keeper.CancelFinancialCase(ctx, msg.CaseId, msg.Sender, msg.ReasonHash); err != nil {
		return nil, err
	}
	financialCase, _ := ms.keeper.GetFinancialCase(ctx, msg.CaseId)
	return &types.MsgCancelFinancialCaseResponse{Status: financialCase.Status}, nil
}

func (ms msgServer) FinalizeFinancialCase(goCtx context.Context, msg *types.MsgFinalizeFinancialCase) (*types.MsgFinalizeFinancialCaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	financialCase, err := ms.keeper.FinalizeFinancialCase(ctx, msg.CaseId, msg.Sender)
	if err != nil {
		return nil, err
	}
	return &types.MsgFinalizeFinancialCaseResponse{Status: financialCase.Status, Effects: financialCase.Effects}, nil
}

// RecordFiatConversionObservation records one provider-signed external result.
func (ms msgServer) RecordFiatConversionObservation(goCtx context.Context, msg *types.MsgRecordFiatConversionObservation) (*types.MsgRecordFiatConversionObservationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	result, err := ms.keeper.RecordFiatConversionObservation(ctx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRecordFiatConversionObservationResponse{
		ConversionId: result.Conversion.ConversionID, ObservationSequence: result.Conversion.ObservationSequence,
		Stage: msg.Stage, State: string(result.Conversion.State), ExactDuplicate: result.ExactDuplicate,
		ObservationDigest: append([]byte(nil), result.ObservationDigest...),
	}, nil
}

// UpdateParams applies a complete validated parameter set from x/gov only.
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized.Wrap("invalid settlement parameter authority")
	}
	params := paramsFromProto(msg.Params, ms.keeper.GetParams(ctx))
	if err := ms.keeper.SetParams(ctx, params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
