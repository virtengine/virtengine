//go:build e2e.integration

package hpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	hpckeeper "github.com/virtengine/virtengine/x/hpc/keeper"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// TestNodeRegistrationAndPruning validates on-chain registration and stale pruning.
func TestNodeRegistrationAndPruning(t *testing.T) {
	baseTime := time.Unix(1_700_000_000, 0).UTC()
	ctx, keeper, _ := setupIntegrationKeeper(t)
	ctx = ctx.WithBlockTime(baseTime).WithBlockHeight(1)

	msgServer := hpckeeper.NewMsgServerImpl(keeper)

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
		Capacity:           realNodeCapacityFixture(),
		Health:             realNodeHealthFixture(hpcv1.HealthStatusHealthy, "mixed"),
		Hardware:           realNodeHardwareFixture(),
		Locality:           realNodeLocalityFixture(),
	}

	_, err = msgServer.UpdateNodeMetadata(ctx, update)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(2)

	node, found := keeper.GetNodeMetadata(ctx, "node-001")
	require.True(t, found)
	require.Equal(t, hpctypes.NodeStateActive, node.State)
	require.True(t, node.Active)
	require.NotNil(t, node.Capacity)
	require.Equal(t, int32(20), node.Capacity.CPUCoresAllocated)
	require.Equal(t, int32(2), node.Capacity.GPUsAllocated)
	require.Equal(t, int32(1250), node.Capacity.StorageGBAllocated)
	require.NotNil(t, node.Health)
	require.Equal(t, int32(81), node.Health.GPUUtilizationPercent)
	require.Equal(t, int32(77), node.Health.GPUMemoryUtilizationPercent)
	require.Equal(t, int32(43), node.Health.DiskIOUtilizationPercent)
	require.Equal(t, int32(36), node.Health.NetworkUtilizationPercent)
	require.Equal(t, "mixed", node.Health.SLURMState)
	require.NotNil(t, node.Hardware)
	require.Equal(t, "NVIDIA H100", node.Hardware.GPUModel)
	require.NotNil(t, node.Locality)
	require.Equal(t, "rack-7", node.Locality.Rack)

	// Advance block time beyond offline + deregistration thresholds.
	staleCtx := ctx.WithBlockHeight(3).WithBlockTime(baseTime.Add(2 * time.Hour))

	require.NoError(t, keeper.CheckStaleNodes(staleCtx))

	node, found = keeper.GetNodeMetadata(staleCtx, "node-001")
	require.True(t, found)
	require.Equal(t, hpctypes.NodeStateOffline, node.State)
	require.False(t, node.Active)
	require.Equal(t, hpctypes.HealthStatusOffline, node.HealthStatus)

	activeNodes := keeper.GetActiveNodesByCluster(staleCtx, clusterResp.ClusterId)
	require.Len(t, activeNodes, 0)

	// Run a second pass to trigger deregistration.
	require.NoError(t, keeper.CheckStaleNodes(staleCtx))

	node, found = keeper.GetNodeMetadata(staleCtx, "node-001")
	require.True(t, found)
	require.Equal(t, hpctypes.NodeStateDeregistered, node.State)
	require.False(t, node.Active)
	require.NotNil(t, node.DeregisteredAt)
}
