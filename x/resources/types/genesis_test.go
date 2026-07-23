package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenesisValidateCapacityConservation(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	state := *DefaultGenesisState()
	state.Inventories = []ResourceInventory{{
		InventoryId: "inventory", ProviderAddress: "provider", ResourceClass: ResourceClassCompute,
		Total: ResourceCapacity{CpuCores: 4}, Available: ResourceCapacity{CpuCores: 3},
	}}
	state.Reservations = []Reservation{{
		ReservationId: "reservation", State: ReservationStateActive,
		ProviderAddress: "provider", InventoryId: "inventory", ResourceClass: ResourceClassCompute,
		Capacity: ResourceCapacity{CpuCores: 1}, ConsumerType: "hpc_job", ConsumerId: "job",
		CreatedAt: now, UpdatedAt: now,
	}}
	require.NoError(t, state.Validate())

	state.Inventories[0].Available.CpuCores = 4
	require.ErrorContains(t, state.Validate(), "does not conserve capacity")
}

func TestGenesisValidateRejectsDuplicateLineage(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	state := *DefaultGenesisState()
	state.Inventories = []ResourceInventory{{
		InventoryId: "inventory", ProviderAddress: "provider", ResourceClass: ResourceClassCompute,
		Total: ResourceCapacity{CpuCores: 2}, Available: ResourceCapacity{},
	}}
	state.Reservations = []Reservation{
		{ReservationId: "one", State: ReservationStateActive, ProviderAddress: "provider", InventoryId: "inventory", ResourceClass: ResourceClassCompute, Capacity: ResourceCapacity{CpuCores: 1}, ConsumerType: "hpc_job", ConsumerId: "job-one", HpcJobId: "job-shared", CreatedAt: now, UpdatedAt: now},
		{ReservationId: "two", State: ReservationStateActive, ProviderAddress: "provider", InventoryId: "inventory", ResourceClass: ResourceClassCompute, Capacity: ResourceCapacity{CpuCores: 1}, ConsumerType: "hpc_job", ConsumerId: "job-two", HpcJobId: "job-shared", CreatedAt: now, UpdatedAt: now},
	}
	require.ErrorContains(t, state.Validate(), "lineage")
}

func TestDefaultGenesisActivatesCurrentProtocolForNewChains(t *testing.T) {
	require.True(t, DefaultGenesisState().CanonicalReservationsActive)
}
