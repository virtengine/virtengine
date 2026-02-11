package keeper

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/x/support/types"
)

func TestRetentionQueue_ArchiveAndPurge(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	submitter := sdk.AccAddress("submitter-retention")
	reqID := types.SupportRequestID{SubmitterAddress: submitter.String(), Sequence: 1}

	payload := types.EncryptedSupportPayload{Envelope: makeTestEnvelope(t, []string{"key-1"})}
	request := types.NewSupportRequest(
		reqID,
		"SUP-000001",
		submitter.String(),
		types.SupportCategoryTechnical,
		types.SupportPriorityNormal,
		payload,
		ctx.BlockTime(),
	)
	request.RetentionPolicy = (&types.RetentionPolicy{
		Version:             types.RetentionPolicyVersion,
		ArchiveAfterSeconds: 1,
		PurgeAfterSeconds:   2,
		CreatedAt:           ctx.BlockTime().UTC(),
		CreatedAtBlock:      ctx.BlockHeight(),
	}).CopyWithTimestamps(ctx.BlockTime(), ctx.BlockHeight())

	require.NoError(t, keeper.CreateSupportRequest(ctx, request))

	_, found := keeper.GetRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionArchive)
	require.True(t, found)
	_, found = keeper.GetRetentionQueueEntry(ctx, request.ID.String(), types.RetentionActionPurge)
	require.True(t, found)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))
	archived, purged := keeper.ProcessRetentionPolicies(ctx)
	require.Equal(t, 1, archived)
	require.Equal(t, 1, purged)

	stored, ok := keeper.GetSupportRequest(ctx, reqID)
	require.True(t, ok)
	require.True(t, stored.Archived)
	require.True(t, stored.Purged)
}

func TestRetentionQueue_RetryOnFailure(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	submitter := sdk.AccAddress("submitter-retry")
	reqID := types.SupportRequestID{SubmitterAddress: submitter.String(), Sequence: 1}

	payload := types.EncryptedSupportPayload{Envelope: makeTestEnvelope(t, []string{"key-2"})}
	badRequest := types.NewSupportRequest(
		reqID,
		"SUP-000002",
		submitter.String(),
		types.SupportCategory("invalid"),
		types.SupportPriorityNormal,
		payload,
		ctx.BlockTime(),
	)
	badRequest.RetentionPolicy = (&types.RetentionPolicy{
		Version:             types.RetentionPolicyVersion,
		ArchiveAfterSeconds: 1,
		PurgeAfterSeconds:   0,
		CreatedAt:           ctx.BlockTime().UTC(),
		CreatedAtBlock:      ctx.BlockHeight(),
	}).CopyWithTimestamps(ctx.BlockTime(), ctx.BlockHeight())

	store := ctx.KVStore(keeper.StoreKey())
	bz, err := json.Marshal(badRequest)
	require.NoError(t, err)
	store.Set(types.SupportRequestKey(reqID.String()), bz)

	entry := types.RetentionQueueEntry{
		RequestID:   reqID.String(),
		Action:      types.RetentionActionArchive,
		ScheduledAt: ctx.BlockTime().UTC(),
	}
	require.NoError(t, keeper.setRetentionQueueEntry(ctx, entry))

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))
	archived, purged := keeper.ProcessRetentionPolicies(ctx)
	require.Equal(t, 1, archived)
	require.Equal(t, 0, purged)

	entry, found := keeper.GetRetentionQueueEntry(ctx, reqID.String(), types.RetentionActionArchive)
	require.True(t, found)
	require.Equal(t, uint32(1), entry.Attempts)
	require.True(t, entry.ScheduledAt.After(ctx.BlockTime()))
}
