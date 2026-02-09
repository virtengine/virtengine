package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/review/types"
)

func TestGRPCQueryReview(t *testing.T) {
	keeper, ctx, market, _ := setupKeeper(t)
	reviewer := bech32Addr(t)
	provider := bech32Addr(t)

	market.AddCompletedOrder("order-1", reviewer, provider)
	review := createTestReview(t, reviewer, provider, "order-1", 5)
	require.NoError(t, keeper.SubmitReview(ctx, review))

	querier := GRPCQuerier{Keeper: keeper}

	resp, err := querier.Review(ctx, &reviewv1.QueryReviewRequest{
		ReviewId: review.ID.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, review.ID.String(), resp.Review.ReviewId)
	require.Equal(t, review.ReviewerAddress, resp.Review.Reviewer)
}

func TestGRPCQueryReviewsByUser(t *testing.T) {
	keeper, ctx, market, _ := setupKeeper(t)
	reviewer := bech32Addr(t)
	provider := bech32Addr(t)

	market.AddCompletedOrder("order-1", reviewer, provider)
	market.AddCompletedOrder("order-2", reviewer, provider)

	review1 := createTestReview(t, reviewer, provider, "order-1", 4)
	review2 := createTestReview(t, reviewer, provider, "order-2", 5)

	require.NoError(t, keeper.SubmitReview(ctx, review1))
	require.NoError(t, keeper.SubmitReview(ctx, review2))

	querier := GRPCQuerier{Keeper: keeper}

	resp, err := querier.ReviewsByUser(ctx, &reviewv1.QueryReviewsByUserRequest{
		Reviewer: reviewer,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Reviews, 2)
}

func bech32Addr(t *testing.T) string {
	t.Helper()
	return sdk.MustBech32ifyAddressBytes(sdkutil.Bech32PrefixAccAddr, testutil.AccAddress(t))
}

func TestGRPCQueryReviewParams(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	customParams := types.Params{
		MinReviewInterval:     10,
		MaxCommentLength:      100,
		RequireCompletedOrder: false,
		ReviewWindow:          20,
		MinRating:             1,
		MaxRating:             5,
	}
	require.NoError(t, keeper.SetParams(ctx, customParams))

	querier := GRPCQuerier{Keeper: keeper}

	resp, err := querier.Params(ctx, &reviewv1.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, customParams, resp.Params)
}
