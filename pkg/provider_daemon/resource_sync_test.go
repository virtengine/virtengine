package provider_daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

type resourceSyncChainStub struct {
	reservations []resourcesv1.Reservation
	queryErr     error
	heartbeats   []*resourcesv1.MsgProviderHeartbeat
}

func (s *resourceSyncChainStub) SubmitResourceHeartbeat(_ context.Context, heartbeat *resourcesv1.MsgProviderHeartbeat) error {
	s.heartbeats = append(s.heartbeats, heartbeat)
	return nil
}

func (s *resourceSyncChainStub) GetProviderReservations(context.Context, string) ([]resourcesv1.Reservation, error) {
	return s.reservations, s.queryErr
}

func TestResourceSyncSubtractsCanonicalNonterminalReservations(t *testing.T) {
	chain := &resourceSyncChainStub{reservations: []resourcesv1.Reservation{
		{ReservationId: "pending", State: resourcesv1.ReservationState_RESERVATION_STATE_PENDING, Capacity: resourcesv1.ResourceCapacity{CpuCores: 1, MemoryGb: 2}},
		{ReservationId: "consumed", State: resourcesv1.ReservationState_RESERVATION_STATE_CONSUMED, Capacity: resourcesv1.ResourceCapacity{CpuCores: 2, MemoryGb: 3}},
		{ReservationId: "released", State: resourcesv1.ReservationState_RESERVATION_STATE_RELEASED, Capacity: resourcesv1.ResourceCapacity{CpuCores: 4, MemoryGb: 4}},
	}}
	snapshot := NewStaticResourceSnapshotProvider(CapacityConfig{TotalCPUCores: 8, TotalMemoryGB: 16}, "", 1000, 0)
	syncer, err := NewResourceAvailabilitySync(ResourceSyncConfig{ProviderAddress: "provider", InventoryID: "inventory"}, chain, snapshot)
	require.NoError(t, err)

	require.NoError(t, syncer.syncOnce(context.Background()))
	require.Len(t, chain.heartbeats, 1)
	require.Equal(t, int64(5), chain.heartbeats[0].Available.CpuCores)
	require.Equal(t, int64(11), chain.heartbeats[0].Available.MemoryGb)
}

func TestResourceSyncFailsClosedWhenReservationQueryFails(t *testing.T) {
	chain := &resourceSyncChainStub{queryErr: errors.New("query unavailable")}
	snapshot := NewStaticResourceSnapshotProvider(CapacityConfig{TotalCPUCores: 8}, "", 0, 0)
	syncer, err := NewResourceAvailabilitySync(ResourceSyncConfig{ProviderAddress: "provider", InventoryID: "inventory"}, chain, snapshot)
	require.NoError(t, err)

	err = syncer.syncOnce(context.Background())
	require.ErrorContains(t, err, "query provider reservations")
	require.Empty(t, chain.heartbeats)
}

func TestResourceSyncRejectsReservationsBeyondPhysicalCapacity(t *testing.T) {
	chain := &resourceSyncChainStub{reservations: []resourcesv1.Reservation{{
		ReservationId: "oversubscribed", State: resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE,
		Capacity: resourcesv1.ResourceCapacity{CpuCores: 9},
	}}}
	snapshot := NewStaticResourceSnapshotProvider(CapacityConfig{TotalCPUCores: 8}, "", 0, 0)
	syncer, err := NewResourceAvailabilitySync(ResourceSyncConfig{ProviderAddress: "provider", InventoryID: "inventory"}, chain, snapshot)
	require.NoError(t, err)

	err = syncer.syncOnce(context.Background())
	require.ErrorContains(t, err, "exceed physical capacity")
	require.Empty(t, chain.heartbeats)
}
