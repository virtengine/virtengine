//go:build e2e.integration

package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/x/support/types"
)

func setupKeeperOnDB(t *testing.T, db dbm.DB, now time.Time, height int64) (Keeper, sdk.Context, storetypes.CommitMultiStore) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	keeper := NewKeeper(cdc, storeKey, "authority", nil, nil)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   now.UTC(),
		Height: height,
	}, false, log.NewNopLogger())
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	return keeper, ctx, stateStore
}

func TestRetentionQueueSurvivesRestartAndCompletes(t *testing.T) {
	start := time.Date(2026, 4, 11, 6, 0, 0, 0, time.UTC)
	db := dbm.NewMemDB()

	keeper, ctx, stateStore := setupKeeperOnDB(t, db, start, 1)
	request := newRetentionSupportRequest(t, sdk.AccAddress("restart-submit"), 1, 1, start, time.Minute, 2*time.Minute)
	require.NoError(t, keeper.CreateSupportRequest(ctx, request))

	archiveEntry, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionArchive, request.ID.String())
	require.True(t, found)
	require.Equal(t, start.Add(time.Minute).Unix(), archiveEntry.DueAtUnix)
	purgeEntry, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionPurge, request.ID.String())
	require.True(t, found)
	require.Equal(t, start.Add(2*time.Minute).Unix(), purgeEntry.DueAtUnix)

	stateStore.Commit()

	keeper, ctx, stateStore = setupKeeperOnDB(t, db, time.Unix(archiveEntry.DueAtUnix, 0).Add(time.Second), 2)
	archiveEntryAfterRestart, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionArchive, request.ID.String())
	require.True(t, found)
	require.Equal(t, archiveEntry.DueAtUnix, archiveEntryAfterRestart.DueAtUnix)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	archived, purged := keeper.ProcessRetentionPolicies(ctx)
	require.Equal(t, 1, archived)
	require.Zero(t, purged)

	stored, found := keeper.GetSupportRequest(ctx, request.ID)
	require.True(t, found)
	require.True(t, stored.Archived)
	require.False(t, stored.Purged)
	require.False(t, ctx.KVStore(keeper.skey).Has(retentionActionArchive.queueKey(archiveEntry.DueAtUnix, request.ID.String())))

	stateStore.Commit()

	keeper, ctx, _ = setupKeeperOnDB(t, db, time.Unix(purgeEntry.DueAtUnix, 0).Add(time.Second), 3)
	purgeEntryAfterRestart, found := keeper.getRetentionQueueEntryByRequest(ctx, retentionActionPurge, request.ID.String())
	require.True(t, found)
	require.Equal(t, purgeEntry.DueAtUnix, purgeEntryAfterRestart.DueAtUnix)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	archived, purged = keeper.ProcessRetentionPolicies(ctx)
	require.Zero(t, archived)
	require.Equal(t, 1, purged)

	stored, found = keeper.GetSupportRequest(ctx, request.ID)
	require.True(t, found)
	require.True(t, stored.Archived)
	require.True(t, stored.Purged)
	require.Nil(t, stored.Payload.Envelope)
	require.NotEmpty(t, stored.Payload.EnvelopeHash)
	_, found = keeper.getRetentionQueueEntryByRequest(ctx, retentionActionPurge, request.ID.String())
	require.False(t, found)
}
