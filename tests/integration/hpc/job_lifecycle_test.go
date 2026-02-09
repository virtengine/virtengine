package hpc

import (
	"bytes"
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
)

type integrationBankKeeper struct {
	transfers []sdk.Coins
}

func (b *integrationBankKeeper) SendCoins(_ context.Context, _ sdk.AccAddress, _ sdk.AccAddress, amt sdk.Coins) error {
	b.transfers = append(b.transfers, amt)
	return nil
}

func (b *integrationBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, amt sdk.Coins) error {
	b.transfers = append(b.transfers, amt)
	return nil
}

func (b *integrationBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, amt sdk.Coins) error {
	b.transfers = append(b.transfers, amt)
	return nil
}

func (b *integrationBankKeeper) SpendableCoins(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

func setupIntegrationKeeper(t testing.TB) (sdk.Context, keeper.Keeper, *integrationBankKeeper) {
	t.Helper()

	cfg := testutilmod.MakeTestEncodingConfig()
	cdc := cfg.Codec

	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(0, 0)}, false, testutil.Logger(t))
	bank := &integrationBankKeeper{}

	k := keeper.NewKeeper(cdc, key, bank, "authority")
	return ctx, k, bank
}

func TestJobLifecycleSubmitScheduleRunCompleteSettle(t *testing.T) {
	ctx, k, bank := setupIntegrationKeeper(t)

	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	cluster := types.HPCCluster{
		ClusterID:       "cluster-lifecycle",
		ProviderAddress: providerAddr,
		Name:            "Lifecycle Cluster",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  8,
		Region:          "us-west-2",
	}
	require.NoError(t, k.SetCluster(ctx, cluster))

	mustSetNode := func(nodeID string, latency int64) {
		node := types.NodeMetadata{
			NodeID:          nodeID,
			ClusterID:       cluster.ClusterID,
			ProviderAddress: providerAddr,
			Region:          cluster.Region,
			AvgLatencyMs:    latency,
			Active:          true,
		}
		require.NoError(t, k.SetNodeMetadata(ctx, node))
	}
	mustSetNode("node-life-1", 12)

	offering := types.HPCOffering{
		OfferingID:        "offering-lifecycle",
		ClusterID:         cluster.ClusterID,
		ProviderAddress:   providerAddr,
		Name:              "Lifecycle Offering",
		MaxRuntimeSeconds: 7200,
		Active:            true,
	}
	require.NoError(t, k.SetOffering(ctx, offering))

	job := types.HPCJob{
		JobID:           "job-lifecycle",
		OfferingID:      offering.OfferingID,
		CustomerAddress: customerAddr,
		QueueName:       "default",
		WorkloadSpec: types.JobWorkloadSpec{
			ContainerImage: "docker.io/library/alpine:latest",
			Command:        "echo",
			Arguments:      []string{"hello"},
		},
		Resources: types.JobResources{
			Nodes:           2,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 16,
		},
		MaxRuntimeSeconds: 3600,
	}
	require.NoError(t, k.SubmitJob(ctx, &job))

	decision, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)
	require.Equal(t, cluster.ClusterID, decision.SelectedClusterID)

	ctx = ctx.WithBlockTime(time.Unix(1_700_000_100, 0))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateQueued, "queued", 0, nil))
	ctx = ctx.WithBlockTime(time.Unix(1_700_000_200, 0))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateRunning, "running", 0, nil))
	ctx = ctx.WithBlockTime(time.Unix(1_700_000_800, 0))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateCompleted, "completed", 0, nil))

	result, err := k.ProcessJobSettlement(ctx, job.JobID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotEmpty(t, result.SettlementID)
	require.NotEmpty(t, bank.transfers)
}
