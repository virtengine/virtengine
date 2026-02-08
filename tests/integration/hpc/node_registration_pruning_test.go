//go:build e2e.integration

package hpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/virtengine/virtengine/app"
	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	hpckeeper "github.com/virtengine/virtengine/x/hpc/keeper"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// TestNodeRegistrationAndPruning validates on-chain registration and stale pruning.
func TestNodeRegistrationAndPruning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	app := app.Setup(
		app.WithChainID("virtengine-hpc-integration-1"),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return app.GenesisStateWithValSet(cdc)
		}),
	)

	baseTime := time.Unix(1_700_000_000, 0).UTC()
	ctx := app.NewUncachedContext(false, cmtproto.Header{
		Height: 1,
		Time:   baseTime,
	})

	msgServer := hpckeeper.NewMsgServerImpl(app.Keepers.VirtEngine.HPC)

	provider := sdktestutil.AccAddress(t)
	registerCluster := hpctypes.NewMsgRegisterCluster(
		provider.String(),
		"Test Cluster",
		"slurm",
		"us-east-1",
		"http://cluster.local",
		10,
		2,
	)

	clusterResp, err := msgServer.RegisterCluster(ctx, registerCluster)
	require.NoError(t, err)

	update := &hpctypes.MsgUpdateNodeMetadata{
		ProviderAddress:    provider.String(),
		NodeId:             "node-001",
		ClusterId:          clusterResp.ClusterId,
		Region:             "us-east-1",
		Datacenter:         "dc1",
		Active:             true,
		State:              hpcv1.NodeStateActive,
		HealthStatus:       hpcv1.HealthStatusHealthy,
		LastSequenceNumber: 1,
		Capacity: &hpcv1.NodeCapacity{
			CpuCoresTotal:      64,
			CpuCoresAvailable:  64,
			MemoryGbTotal:      512,
			MemoryGbAvailable:  512,
			GpusTotal:          8,
			GpusAvailable:      8,
			GpuType:            "NVIDIA A100",
			StorageGbTotal:     4000,
			StorageGbAvailable: 4000,
		},
		Health: &hpcv1.NodeHealth{
			Status:                   hpcv1.HealthStatusHealthy,
			UptimeSeconds:            120,
			LoadAverage_1M:           "0.15",
			CpuUtilizationPercent:    5,
			MemoryUtilizationPercent: 8,
			SlurmState:               "idle",
		},
		Hardware: &hpcv1.NodeHardware{
			CpuModel:    "AMD EPYC 9654",
			CpuVendor:   "AMD",
			CpuArch:     "x86_64",
			GpuModel:    "NVIDIA A100",
			StorageType: "NVMe",
		},
		Locality: &hpcv1.NodeLocality{
			Region:     "us-east-1",
			Datacenter: "dc1",
			Zone:       "a",
			Rack:       "rack-7",
			Row:        "row-2",
			Position:   "u14",
		},
	}

	_, err = msgServer.UpdateNodeMetadata(ctx, update)
	require.NoError(t, err)

	app.Commit()

	ctx = app.NewUncachedContext(false, cmtproto.Header{
		Height: 2,
		Time:   baseTime,
	})

	node, found := app.Keepers.VirtEngine.HPC.GetNodeMetadata(ctx, "node-001")
	require.True(t, found)
	require.Equal(t, hpctypes.NodeStateActive, node.State)
	require.True(t, node.Active)

	// Advance block time beyond offline + deregistration thresholds.
	staleCtx := app.NewUncachedContext(false, cmtproto.Header{
		Height: 3,
		Time:   baseTime.Add(2 * time.Hour),
	})

	require.NoError(t, app.Keepers.VirtEngine.HPC.CheckStaleNodes(staleCtx))

	node, found = app.Keepers.VirtEngine.HPC.GetNodeMetadata(staleCtx, "node-001")
	require.True(t, found)
	require.Equal(t, hpctypes.NodeStateOffline, node.State)
	require.False(t, node.Active)
	require.Equal(t, hpctypes.HealthStatusOffline, node.HealthStatus)

	activeNodes := app.Keepers.VirtEngine.HPC.GetActiveNodesByCluster(staleCtx, clusterResp.ClusterId)
	require.Len(t, activeNodes, 0)

	// Run a second pass to trigger deregistration.
	require.NoError(t, app.Keepers.VirtEngine.HPC.CheckStaleNodes(staleCtx))

	node, found = app.Keepers.VirtEngine.HPC.GetNodeMetadata(staleCtx, "node-001")
	require.True(t, found)
	require.Equal(t, hpctypes.NodeStateDeregistered, node.State)
	require.False(t, node.Active)
	require.NotNil(t, node.DeregisteredAt)
}
