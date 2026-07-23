package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/resources/types"
)

func TestMigrateReservationsGoldenFixture(t *testing.T) {
	k, ctx := setupKeeper(t)
	seedInventory(t, k, ctx, "migration-inventory", "virtengine1providermigration", 10)

	fixtures := []types.ResourceAllocation{
		{AllocationId: "linked-active", RequestId: "request-active", RequesterAddress: "requester", ProviderAddress: "virtengine1providermigration", ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 2}, Assigned: types.ResourceCapacity{CpuCores: 2}, State: types.AllocationStateActive, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
		{AllocationId: "orphan-pending", RequestId: "request-orphan", RequesterAddress: "requester", ProviderAddress: "virtengine1providerorphan", ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1}, Assigned: types.ResourceCapacity{CpuCores: 1}, State: types.AllocationStatePending, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
		{AllocationId: "terminal-released", RequestId: "request-terminal", RequesterAddress: "requester", ProviderAddress: "virtengine1providermigration", ResourceClass: types.ResourceClassCompute, Required: types.ResourceCapacity{CpuCores: 1}, Assigned: types.ResourceCapacity{CpuCores: 1}, State: types.AllocationStateReleased, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()},
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
	require.Equal(t, uint64(2), report.Quarantined)

	active, found := k.GetReservation(ctx, "legacy/allocation/linked-active")
	require.True(t, found)
	require.Equal(t, types.ReservationStateQuarantined, active.State)
	orphan, found := k.GetReservation(ctx, "legacy/allocation/orphan-pending")
	require.True(t, found)
	require.Equal(t, types.ReservationStateQuarantined, orphan.State)
	terminal, found := k.GetReservation(ctx, "legacy/allocation/terminal-released")
	require.True(t, found)
	require.Equal(t, types.ReservationStateReleased, terminal.State)

	second, err := k.MigrateReservations(ctx)
	require.NoError(t, err)
	require.Zero(t, second.ReservationsCreated, "migration retry must not duplicate reservations")
}
