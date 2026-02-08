package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/resources/types"
)

func TestAllocateResourcesSelectsLocality(t *testing.T) {
	k, ctx := setupKeeper(t)

	inventoryA := types.ResourceInventory{
		InventoryId:       "inv-a",
		ProviderAddress:   "virtengine1provideraaaa",
		ResourceClass:     types.ResourceClassCompute,
		Total:             types.ResourceCapacity{CpuCores: 32, MemoryGb: 64, StorageGb: 1000, NetworkMbps: 1000},
		Available:         types.ResourceCapacity{CpuCores: 32, MemoryGb: 64, StorageGb: 1000, NetworkMbps: 1000},
		Locality:          types.Locality{Region: "us-west", Zone: "us-west-1"},
		Active:            true,
		HeartbeatSequence: 1,
		LastHeartbeat:     ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}
	inventoryB := types.ResourceInventory{
		InventoryId:       "inv-b",
		ProviderAddress:   "virtengine1providerbbbb",
		ResourceClass:     types.ResourceClassCompute,
		Total:             types.ResourceCapacity{CpuCores: 32, MemoryGb: 64, StorageGb: 1000, NetworkMbps: 1000},
		Available:         types.ResourceCapacity{CpuCores: 32, MemoryGb: 64, StorageGb: 1000, NetworkMbps: 1000},
		Locality:          types.Locality{Region: "us-east", Zone: "us-east-1"},
		Active:            true,
		HeartbeatSequence: 1,
		LastHeartbeat:     ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}

	require.NoError(t, k.SetInventory(ctx, inventoryA))
	require.NoError(t, k.SetInventory(ctx, inventoryB))

	request := types.ResourceRequest{
		RequestId:        "req-1",
		RequesterAddress: "virtengine1requester",
		ResourceClass:    types.ResourceClassCompute,
		Required:         types.ResourceCapacity{CpuCores: 4, MemoryGb: 8},
		Locality:         types.Locality{Region: "us-west"},
	}

	allocation, err := k.AllocateResources(ctx, request)
	require.NoError(t, err)
	require.Equal(t, "virtengine1provideraaaa", allocation.ProviderAddress)
}

func TestAllocationLifecycleExpiry(t *testing.T) {
	k, ctx := setupKeeper(t)

	params := types.DefaultParams()
	params.ReservationTimeoutSeconds = 1
	params.SlashingGraceSeconds = 1
	require.NoError(t, k.SetParams(ctx, params))

	inventory := types.ResourceInventory{
		InventoryId:       "inv-exp",
		ProviderAddress:   "virtengine1providercccc",
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
		RequestId:        "req-exp",
		RequesterAddress: "virtengine1requester",
		ResourceClass:    types.ResourceClassCompute,
		Required:         types.ResourceCapacity{CpuCores: 4, MemoryGb: 8},
		Locality:         types.Locality{Region: "us-west"},
	}

	allocation, err := k.AllocateResources(ctx, request)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))
	k.ExpirePendingAllocations(ctx)

	after, found := k.GetAllocation(ctx, allocation.AllocationId)
	require.True(t, found)
	require.Equal(t, types.AllocationStateExpired, after.State)

	invAfter, found := k.GetInventory(ctx, inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
	require.True(t, found)
	require.Equal(t, int64(8), invAfter.Available.CpuCores)
}

func TestActivateAllocation(t *testing.T) {
	k, ctx := setupKeeper(t)

	inventory := types.ResourceInventory{
		InventoryId:       "inv-act",
		ProviderAddress:   "virtengine1providerdddd",
		ResourceClass:     types.ResourceClassCompute,
		Total:             types.ResourceCapacity{CpuCores: 16, MemoryGb: 32},
		Available:         types.ResourceCapacity{CpuCores: 16, MemoryGb: 32},
		Locality:          types.Locality{Region: "us-west"},
		Active:            true,
		HeartbeatSequence: 1,
		LastHeartbeat:     ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}
	require.NoError(t, k.SetInventory(ctx, inventory))

	request := types.ResourceRequest{
		RequestId:        "req-act",
		RequesterAddress: "virtengine1requester",
		ResourceClass:    types.ResourceClassCompute,
		Required:         types.ResourceCapacity{CpuCores: 2, MemoryGb: 4},
		Locality:         types.Locality{Region: "us-west"},
	}

	allocation, err := k.AllocateResources(ctx, request)
	require.NoError(t, err)

	active, err := k.ActivateAllocation(ctx, allocation.AllocationId, inventory.ProviderAddress)
	require.NoError(t, err)
	require.Equal(t, types.AllocationStateActive, active.State)
	require.NotNil(t, active.ActivatedAt)
}
