package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/review/types"
)

func TestMsgServerSubmitReview(t *testing.T) {
	keeper, ctx, market, _ := setupKeeper(t)
	reviewer := bech32AddrMsg(t)
	provider := bech32AddrMsg(t)
	market.AddCompletedOrder("order-msg-1", reviewer, provider)

	msgServer := NewMsgServer(keeper)
	resp, err := msgServer.SubmitReview(ctx, &types.MsgSubmitReview{
		Reviewer:       reviewer,
		SubjectAddress: provider,
		SubjectType:    "provider",
		OrderId:        "order-msg-1",
		Rating:         5,
		Comment:        "This review has enough length to be valid.",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ReviewId)
}

func TestMsgServerDeleteReview(t *testing.T) {
	keeper, ctx, market, roles := setupKeeper(t)
	reviewer := bech32AddrMsg(t)
	provider := bech32AddrMsg(t)
	market.AddCompletedOrder("order-msg-2", reviewer, provider)

	review := createTestReview(t, reviewer, provider, "order-msg-2", 4)
	require.NoError(t, keeper.SubmitReview(ctx, review))

	roles.AddModerator(reviewer)
	msgServer := NewMsgServer(keeper)
	_, err := msgServer.DeleteReview(ctx, &types.MsgDeleteReview{
		Authority: reviewer,
		ReviewId:  review.ID.String(),
		Reason:    "policy violation",
	})
	require.NoError(t, err)
}

func bech32AddrMsg(t *testing.T) string {
	t.Helper()
	return sdk.MustBech32ifyAddressBytes(sdkutil.Bech32PrefixAccAddr, testutil.AccAddress(t))
}

func TestMsgServerUpdateParams(t *testing.T) {
	keeper, ctx, _, _ := setupKeeper(t)

	msgServer := NewMsgServer(keeper)
	customParams := types.Params{
		MinReviewInterval:     5,
		MaxCommentLength:      500,
		RequireCompletedOrder: true,
		ReviewWindow:          50,
		MinRating:             1,
		MaxRating:             5,
	}

	_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: "authority",
		Params:    customParams,
	})
	require.NoError(t, err)

	storedParams := keeper.GetParams(ctx)
	require.Equal(t, customParams, storedParams)
}
