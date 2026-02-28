package keeper

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/x/support/types"
)

func newRetentionSupportRequest(
	t *testing.T,
	submitter sdk.AccAddress,
	sequence uint64,
	ticketSequence uint64,
	now time.Time,
	archiveAfter time.Duration,
	purgeAfter time.Duration,
) *types.SupportRequest {
	t.Helper()

	request := types.NewSupportRequest(
		types.SupportRequestID{
			SubmitterAddress: submitter.String(),
			Sequence:         sequence,
		},
		fmt.Sprintf("SUP-%06d", ticketSequence),
		submitter.String(),
		types.SupportCategoryTechnical,
		types.SupportPriorityNormal,
		types.EncryptedSupportPayload{
			Envelope: makeTestEnvelope(t, []string{"submitter-key", "support-key"}),
		},
		now.UTC(),
	)
	request.RetentionPolicy = (&types.RetentionPolicy{
		Version:             types.RetentionPolicyVersion,
		ArchiveAfterSeconds: int64(archiveAfter / time.Second),
		PurgeAfterSeconds:   int64(purgeAfter / time.Second),
		CreatedAt:           now.UTC(),
		CreatedAtBlock:      1,
	}).CopyWithTimestamps(now.UTC(), 1)
	return request
}

func overwriteStoredSupportRequest(t *testing.T, keeper Keeper, ctx sdk.Context, request types.SupportRequest) {
	t.Helper()

	bz, err := json.Marshal(&request)
	require.NoError(t, err)
	ctx.KVStore(keeper.skey).Set(types.SupportRequestKey(request.ID.String()), bz)
}

func requireEventAttribute(t *testing.T, events sdk.Events, eventType, key, expectedValue string) {
	t.Helper()

	for _, event := range events {
		if event.Type != eventType {
			continue
		}
		for _, attribute := range event.Attributes {
			if attribute.Key == key && attribute.Value == expectedValue {
				return
			}
		}
	}

	require.Failf(t, "missing event attribute", "event %s missing %s=%s", eventType, key, expectedValue)
}

func TestRetentionQueueReschedulesOnUpdate(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	now := time.Date(2026, 4, 11, 2, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(now).WithBlockHeight(1)

	request := newRetentionSupportRequest(t, sdk.AccAddress("retention-submit"), 1, 1, now, time.Hour, 2*time.Hour)
	require.NoError(t, keeper.CreateSupportRequest(ctx, request))

	archiveEntry, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionArchive, request.ID.String())
	require.True(t, found)
	require.Equal(t, now.Add(time.Hour).Unix(), archiveEntry.DueAtUnix)

	purgeEntry, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionPurge, request.ID.String())
	require.True(t, found)
	require.Equal(t, now.Add(2*time.Hour).Unix(), purgeEntry.DueAtUnix)

	updateTime := now.Add(15 * time.Minute)
	request.RetentionPolicy = (&types.RetentionPolicy{
		Version:             types.RetentionPolicyVersion,
		ArchiveAfterSeconds: int64((3 * time.Hour) / time.Second),
		PurgeAfterSeconds:   int64((4 * time.Hour) / time.Second),
		CreatedAt:           updateTime.UTC(),
		CreatedAtBlock:      2,
	}).CopyWithTimestamps(updateTime.UTC(), 2)
	request.UpdatedAt = updateTime.UTC()

	ctx = ctx.WithBlockTime(updateTime).WithBlockHeight(2)
	require.NoError(t, keeper.UpdateSupportRequest(ctx, request))

	archiveEntry, found = keeper.getRetentionQueueEntryByRequest(ctx, retentionActionArchive, request.ID.String())
	require.True(t, found)
	require.Equal(t, updateTime.Add(3*time.Hour).Unix(), archiveEntry.DueAtUnix)

	purgeEntry, found = keeper.getRetentionQueueEntryByRequest(ctx, retentionActionPurge, request.ID.String())
	require.True(t, found)
	require.Equal(t, updateTime.Add(4*time.Hour).Unix(), purgeEntry.DueAtUnix)

	require.False(t, ctx.KVStore(keeper.skey).Has(retentionActionArchive.queueKey(now.Add(time.Hour).Unix(), request.ID.String())))
	require.False(t, ctx.KVStore(keeper.skey).Has(retentionActionPurge.queueKey(now.Add(2*time.Hour).Unix(), request.ID.String())))
	require.Len(t, keeper.listRetentionQueueEntries(ctx, retentionActionArchive), 1)
	require.Len(t, keeper.listRetentionQueueEntries(ctx, retentionActionPurge), 1)
}

func TestProcessRetentionPoliciesRespectsBackpressure(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	now := time.Date(2026, 4, 11, 3, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(now).WithBlockHeight(1)

	totalRequests := retentionProcessLimitPerBlock + 3
	for sequence := 1; sequence <= totalRequests; sequence++ {
		request := newRetentionSupportRequest(
			t,
			sdk.AccAddress(fmt.Sprintf("submitter-%02d", sequence)),
			uint64(sequence),
			uint64(sequence),
			now,
			time.Minute,
			0,
		)
		require.NoError(t, keeper.CreateSupportRequest(ctx, request))
	}

	processCtx := ctx.WithBlockTime(now.Add(2 * time.Minute)).WithBlockHeight(2).WithEventManager(sdk.NewEventManager())
	archived, purged := keeper.ProcessRetentionPolicies(processCtx)
	require.Equal(t, retentionProcessLimitPerBlock, archived)
	require.Zero(t, purged)
	require.Len(t, keeper.listRetentionQueueEntries(processCtx, retentionActionArchive), totalRequests-retentionProcessLimitPerBlock)
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_summary", "archive_backpressured", "true")
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_summary", "archive_completed", fmt.Sprintf("%d", retentionProcessLimitPerBlock))

	processCtx = processCtx.WithBlockTime(now.Add(3 * time.Minute)).WithBlockHeight(3).WithEventManager(sdk.NewEventManager())
	archived, purged = keeper.ProcessRetentionPolicies(processCtx)
	require.Equal(t, totalRequests-retentionProcessLimitPerBlock, archived)
	require.Zero(t, purged)
	require.Len(t, keeper.listRetentionQueueEntries(processCtx, retentionActionArchive), 0)
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_summary", "archive_backpressured", "false")
}

func TestProcessRetentionPoliciesRetriesFailedEntries(t *testing.T) {
	keeper, ctx := setupKeeper(t)
	now := time.Date(2026, 4, 11, 4, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(now).WithBlockHeight(1)

	request := newRetentionSupportRequest(t, sdk.AccAddress("retry-submit"), 1, 1, now, time.Minute, 0)
	require.NoError(t, keeper.CreateSupportRequest(ctx, request))

	corrupted, found := keeper.GetSupportRequest(ctx, request.ID)
	require.True(t, found)
	corrupted.SubmitterAddress = ""
	overwriteStoredSupportRequest(t, keeper, ctx, corrupted)

	processCtx := ctx.WithBlockTime(now.Add(2 * time.Minute)).WithBlockHeight(2).WithEventManager(sdk.NewEventManager())
	archived, purged := keeper.ProcessRetentionPolicies(processCtx)
	require.Zero(t, archived)
	require.Zero(t, purged)

	archiveEntry, found := keeper.getRetentionQueueEntryByRequest(processCtx, retentionActionArchive, request.ID.String())
	require.True(t, found)
	require.Equal(t, uint32(1), archiveEntry.Attempt)
	require.Equal(t, now.Add(2*time.Minute).Add(retentionRetryBaseBackoff).Unix(), archiveEntry.DueAtUnix)
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_summary", "archive_retried", "1")
	requireEventAttribute(t, processCtx.EventManager().Events(), "support_retention_queue", "status", "retry_scheduled")

	repaired := corrupted
	repaired.SubmitterAddress = request.SubmitterAddress
	overwriteStoredSupportRequest(t, keeper, processCtx, repaired)

	retryCtx := processCtx.WithBlockTime(time.Unix(archiveEntry.DueAtUnix, 0).Add(time.Second)).WithBlockHeight(3).WithEventManager(sdk.NewEventManager())
	archived, purged = keeper.ProcessRetentionPolicies(retryCtx)
	require.Equal(t, 1, archived)
	require.Zero(t, purged)
	require.False(t, retryCtx.KVStore(keeper.skey).Has(retentionActionArchive.queueKey(archiveEntry.DueAtUnix, request.ID.String())))

	stored, found := keeper.GetSupportRequest(retryCtx, request.ID)
	require.True(t, found)
	require.True(t, stored.Archived)
	_, found = keeper.getRetentionQueueEntryByRequest(retryCtx, retentionActionArchive, request.ID.String())
	require.False(t, found)
	requireEventAttribute(t, retryCtx.EventManager().Events(), "support_retention_queue", "status", "completed")
}
