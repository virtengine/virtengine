package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/resources/types"
)

func TestMigrateReservationsGoldenFixture(t *testing.T) {
	k, ctx := setupKeeper(t)
	provider := seedInventory(t, k, ctx, "migration-inventory", "virtengine1providermigration", 10)
	inventory, found := k.GetInventory(ctx, provider, types.ResourceClassCompute, "migration-inventory")
	require.True(t, found)
	inventory.Available.CpuCores = 8
	require.NoError(t, k.SetInventory(ctx, inventory))

	fixtures := []types.ResourceAllocation{
		{AllocationId: "linked-active", RequestId: "request-active", RequesterAddress: "requester", ProviderAddress: provider, ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 2}, Assigned: types.ResourceCapacity{CpuCores: 2}, State: types.AllocationStateActive, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
		{AllocationId: "orphan-pending", RequestId: "request-orphan", RequesterAddress: "requester", ProviderAddress: "virtengine1providerorphan", ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1}, Assigned: types.ResourceCapacity{CpuCores: 1}, State: types.AllocationStatePending, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
		{AllocationId: "terminal-released", RequestId: "request-terminal", RequesterAddress: "requester", ProviderAddress: provider, ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1}, Assigned: types.ResourceCapacity{CpuCores: 1}, State: types.AllocationStateReleased, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
	}
	for _, fixture := range fixtures {
		require.NoError(t, k.SetAllocation(ctx, fixture))
	}

	report, err := k.MigrateReservations(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), report.InventoriesScanned)
	require.Equal(t, uint64(3), report.AllocationsScanned)
	require.Equal(t, uint64(3), report.ReservationsCreated)
	require.Equal(t, uint64(1), report.TerminalPreserved)
	require.Equal(t, uint64(1), report.Quarantined)

	active, found := k.GetReservation(ctx, "legacy/allocation/linked-active")
	require.True(t, found)
	require.Equal(t, types.ReservationStateActive, active.State)
	orphan, found := k.GetReservation(ctx, "legacy/allocation/orphan-pending")
	require.True(t, found)
	require.Equal(t, types.ReservationStateQuarantined, orphan.State)
	terminal, found := k.GetReservation(ctx, "legacy/allocation/terminal-released")
	require.True(t, found)
	require.Equal(t, types.ReservationStateReleased, terminal.State)

	second, err := k.MigrateReservations(ctx)
	require.NoError(t, err)
	require.Zero(t, second.ReservationsCreated, "migration retry must not duplicate reservations")
	require.Equal(t, uint64(3), second.AlreadyLinked)
}

func TestMigrateReservationsQuarantinesDanglingAndInconsistentLinks(t *testing.T) {
	k, ctx := setupKeeper(t)
	provider := seedInventory(t, k, ctx, "migration-linked-inventory", "ignored", 4)
	inventory, found := k.GetInventory(ctx, provider, types.ResourceClassCompute, "migration-linked-inventory")
	require.True(t, found)
	inventory.Available.CpuCores = 3
	require.NoError(t, k.SetInventory(ctx, inventory))

	matching := types.Reservation{
		ReservationId: "existing/matching", IdempotencyKey: "existing/matching", RequestId: "matching-request",
		RequesterAddress: "requester", ProviderAddress: provider, InventoryId: "migration-linked-inventory",
		ResourceClass: types.ResourceClassCompute, Capacity: types.ResourceCapacity{CpuCores: 1},
		State: types.ReservationStateActive, ConsumerType: "legacy_allocation", ConsumerId: "matching",
		Version: 1, LegacySource: "resources_allocation", LegacyReference: "matching",
		CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime(),
	}
	require.NoError(t, k.SetReservation(ctx, matching))
	require.NoError(t, k.RebuildReservationIndexes(ctx, matching))

	fixtures := []types.ResourceAllocation{
		{AllocationId: "matching", ReservationId: matching.ReservationId, RequestId: "matching-request", RequesterAddress: "requester", ProviderAddress: provider, ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1}, Assigned: types.ResourceCapacity{CpuCores: 1}, State: types.AllocationStateActive, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
		{AllocationId: "dangling", ReservationId: "missing/reservation", RequestId: "dangling-request", RequesterAddress: "requester", ProviderAddress: provider, ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1}, Assigned: types.ResourceCapacity{CpuCores: 1}, State: types.AllocationStateActive, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
	}
	for _, fixture := range fixtures {
		require.NoError(t, k.SetAllocation(ctx, fixture))
	}

	report, err := k.MigrateReservations(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), report.AlreadyLinked)
	require.Equal(t, uint64(1), report.Quarantined)
	dangling, found := k.GetAllocation(ctx, "dangling")
	require.True(t, found)
	require.NotEqual(t, "missing/reservation", dangling.ReservationId)
	quarantine, found := k.GetReservation(ctx, dangling.ReservationId)
	require.True(t, found)
	require.Equal(t, types.ReservationStateQuarantined, quarantine.State)
	require.Equal(t, types.ResourceCapacity{}, quarantine.Capacity)

	retry, err := k.MigrateReservations(ctx)
	require.NoError(t, err)
	require.Zero(t, retry.ReservationsCreated)
	require.Equal(t, uint64(2), retry.AlreadyLinked)
}
