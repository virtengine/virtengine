package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
	"github.com/virtengine/virtengine/x/review/types"
)

// GRPCQuerier implements the gRPC query interface for the review module.
type GRPCQuerier struct {
	Keeper
}

var _ reviewv1.QueryServer = GRPCQuerier{}

// Review returns a review by ID.
func (q GRPCQuerier) Review(c context.Context, req *reviewv1.QueryReviewRequest) (*reviewv1.QueryReviewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	review, found := q.GetReview(ctx, req.ReviewId)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrReviewNotFound.Error())
	}

	return &reviewv1.QueryReviewResponse{
		Review: toProtoReview(review),
	}, nil
}

// ReviewsByUser returns reviews by reviewer address.
func (q GRPCQuerier) ReviewsByUser(c context.Context, req *reviewv1.QueryReviewsByUserRequest) (*reviewv1.QueryReviewsByUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if _, err := sdk.AccAddressFromBech32(req.Reviewer); err != nil {
		return nil, types.ErrInvalidAddress.Wrap(err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	reviews := q.GetReviewsByReviewer(ctx, req.Reviewer)
	resp := make([]reviewv1.Review, 0, len(reviews))
	for _, review := range reviews {
		resp = append(resp, toProtoReview(review))
	}

	return &reviewv1.QueryReviewsByUserResponse{
		Reviews: resp,
	}, nil
}

// Params returns the module parameters.
func (q GRPCQuerier) Params(c context.Context, req *reviewv1.QueryParamsRequest) (*reviewv1.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := q.GetParams(ctx)

	return &reviewv1.QueryParamsResponse{
		Params: params,
	}, nil
}

func toProtoReview(review types.Review) reviewv1.Review {
	return reviewv1.Review{
		ReviewId:       review.ID.String(),
		Reviewer:       review.ReviewerAddress,
		SubjectAddress: review.ProviderAddress,
		SubjectType:    "provider",
		OrderId:        review.OrderRef.OrderID,
		LeaseId:        "",
		Rating:         uint32(review.Rating),
		Comment:        review.Text,
		SubmittedAt:    review.CreatedAt.Unix(),
		BlockHeight:    review.BlockHeight,
	}
}
