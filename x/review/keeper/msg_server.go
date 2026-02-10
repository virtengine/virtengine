package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/virtengine/virtengine/x/review/types"
)

// MsgServer implements the Review module MsgServer interface.
type MsgServer struct {
	keeper Keeper
}

// NewMsgServer returns an implementation of the Review MsgServer interface.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &MsgServer{keeper: keeper}
}

// SubmitReview handles submitting a new review.
func (m *MsgServer) SubmitReview(ctx context.Context, msg *types.MsgSubmitReview) (*types.MsgSubmitReviewResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	orderRef := types.OrderReference{
		OrderID:         msg.OrderId,
		CustomerAddress: msg.Reviewer,
		ProviderAddress: msg.SubjectAddress,
		CompletedAt:     sdkCtx.BlockTime().UTC(),
	}

	rating, err := toUint8Rating(msg.Rating)
	if err != nil {
		return nil, err
	}

	review := types.Review{
		ID:              types.ReviewID{ProviderAddress: msg.SubjectAddress, Sequence: 1},
		ReviewerAddress: msg.Reviewer,
		ProviderAddress: msg.SubjectAddress,
		OrderRef:        orderRef,
		Rating:          rating,
		Text:            msg.Comment,
		State:           types.ReviewStateActive,
		CreatedAt:       sdkCtx.BlockTime().UTC(),
		UpdatedAt:       sdkCtx.BlockTime().UTC(),
		BlockHeight:     sdkCtx.BlockHeight(),
	}
	review.ContentHash = review.ComputeContentHash()

	if err := m.keeper.SubmitReview(sdkCtx, &review); err != nil {
		return nil, err
	}

	return &types.MsgSubmitReviewResponse{
		ReviewId:    review.ID.String(),
		SubmittedAt: sdkCtx.BlockTime().Unix(),
	}, nil
}

// DeleteReview handles deleting a review.
func (m *MsgServer) DeleteReview(ctx context.Context, msg *types.MsgDeleteReview) (*types.MsgDeleteReviewResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := m.keeper.DeleteReview(sdkCtx, msg.ReviewId, msg.Authority, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgDeleteReviewResponse{}, nil
}

// UpdateParams handles updating module parameters.
func (m *MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if m.keeper.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	if err := types.ValidateParams(&msg.Params); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := m.keeper.SetParams(sdkCtx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

var _ types.MsgServer = (*MsgServer)(nil)

func toUint8Rating(value uint32) (uint8, error) {
	if value < uint32(types.MinRating) || value > uint32(types.MaxRating) {
		return 0, types.ErrInvalidRating.Wrapf("rating %d is outside range [%d, %d]", value, types.MinRating, types.MaxRating)
	}

	return uint8(value), nil
}
