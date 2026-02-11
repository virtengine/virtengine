package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/resources"
	"github.com/virtengine/virtengine/x/resources/types"
)

func TestEndBlockerExpiresPendingAllocations(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	params.ReservationTimeoutSeconds = 1
	params.SlashingGraceSeconds = 1
	require.NoError(t, k.SetParams(ctx, params))

	inventory := types.ResourceInventory{
		InventoryId:       "inv-endblock",
		ProviderAddress:   "virtengine1providerendblock",
		ResourceClass:     types.ResourceClassCompute,
		Total:             types.ResourceCapacity{CpuCores: 8, MemoryGb: 16},
		Available:         types.ResourceCapacity{CpuCores: 8, MemoryGb: 16},
		Locality:          types.Locality{Region: "us-west"},
		Active:            true,
		HeartbeatSequence: 1,
		LastHeartbeat:     ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}
	require.NoError(t, k.SetInventory(ctx, inventory))

	request := types.ResourceRequest{
		RequestId:        "req-endblock",
		RequesterAddress: "virtengine1requesterendblock",
		ResourceClass:    types.ResourceClassCompute,
		Required:         types.ResourceCapacity{CpuCores: 4, MemoryGb: 8},
		Locality:         types.Locality{Region: "us-west"},
	}

	allocation, err := k.AllocateResources(ctx, request)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))

	module := resources.NewAppModule(nil, k)
	err = module.EndBlock(ctx)
	require.NoError(t, err)

	after, found := k.GetAllocation(ctx, allocation.AllocationId)
	require.True(t, found)
	require.Equal(t, types.AllocationStateExpired, after.State)

	invAfter, found := k.GetInventory(ctx, inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
	require.True(t, found)
	require.Equal(t, int64(8), invAfter.Available.CpuCores)
}
