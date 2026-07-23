package keeper_test

import (
	"bytes"
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	"github.com/virtengine/virtengine/x/hpc/types"
)

func TestHPCSubmitRequiresAndActivatesReservation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	resources := newHPCReservationStub()
	k.SetResourcesKeeper(resources)
	seedHPCOffering(t, ctx, k)

	job := validReservationJob(sdk.AccAddress(bytes.Repeat([]byte{3}, 20)).String())
	require.NoError(t, k.SubmitJob(ctx, &job))
	require.Equal(t, "reservation-hpc", job.ReservationID)
	require.Equal(t, 1, resources.reserveCalls)
	require.Equal(t, 1, resources.activateCalls)

	_, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)

	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateQueued, "queued", 0, nil))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateRunning, "running", 0, nil))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateCompleted, "done", 0, nil))
	require.Equal(t, 1, resources.releaseCalls)
}

func TestHPCExecutableStateRejectsMissingReservation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	resources := newHPCReservationStub()
	k.SetResourcesKeeper(resources)
	job := validReservationJob(sdk.AccAddress(bytes.Repeat([]byte{4}, 20)).String())
	job.ReservationID = ""
	require.NoError(t, k.SetJob(ctx, job))

	err := k.UpdateJobStatus(ctx, job.JobID, types.JobStateQueued, "queued", 0, nil)
	require.Error(t, err)
}

func TestHPCReservationActivationFailureIsAtomic(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	resources := newHPCReservationStub()
	resources.activateErr = errors.New("activation failed")
	k.SetResourcesKeeper(resources)
	seedHPCOffering(t, ctx, k)

	job := validReservationJob(sdk.AccAddress(bytes.Repeat([]byte{6}, 20)).String())
	err := k.SubmitJob(ctx, &job)
	require.ErrorContains(t, err, "activation failed")
	require.Empty(t, job.ReservationID)
	_, found := k.GetJob(ctx, job.JobID)
	require.False(t, found)
}

func TestHPCReusesCanonicalMarketReservation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	resources := newHPCReservationStub()
	resources.reservation.ConsumerType = "market_lease"
	resources.reservation.ProviderAddress = sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String()
	k.SetResourcesKeeper(resources)
	seedHPCOffering(t, ctx, k)

	job := validReservationJob(sdk.AccAddress(bytes.Repeat([]byte{7}, 20)).String())
	resources.reservation.RequesterAddress = job.CustomerAddress
	resources.reservation.Capacity = resourcesv1.ResourceCapacity{CpuCores: 2, MemoryGb: 4, StorageGb: 10}
	leaseID := marketv1.LeaseID{Owner: job.CustomerAddress, DSeq: 1, GSeq: 1, OSeq: 1, Provider: resources.reservation.ProviderAddress}
	resources.reservation.MarketOrderId = leaseID.OrderID().String()
	resources.reservation.MarketBidId = leaseID.BidID().String()
	resources.reservation.MarketLeaseId = leaseID.String()
	resources.reservation.ConsumerId = leaseID.String()
	k.SetMarketKeeper(&hpcMarketStub{lease: marketv1.Lease{ID: leaseID, State: marketv1.LeaseActive, ReservationId: resources.reservation.ReservationId}})
	job.ReservationID = resources.reservation.ReservationId
	job.MarketOrderID = resources.reservation.MarketOrderId
	job.MarketBidID = resources.reservation.MarketBidId
	job.MarketLeaseID = resources.reservation.MarketLeaseId
	require.NoError(t, k.SubmitJob(ctx, &job))
	require.Zero(t, resources.reserveCalls)
	require.Zero(t, resources.activateCalls)
	require.Equal(t, resources.reservation.ReservationId, job.AllocationID)
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateQueued, "queued", 0, nil))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateRunning, "running", 0, nil))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateCompleted, "done", 0, nil))
	require.Zero(t, resources.releaseCalls, "canonical market lease owns final capacity release")
}

func TestHPCReusesConsumedCanonicalMarketReservation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	resources := newHPCReservationStub()
	resources.reservation.State = resourcesv1.ReservationState_RESERVATION_STATE_CONSUMED
	resources.reservation.ConsumerType = "market_lease"
	resources.reservation.ConsumerId = "lease-consumed"
	resources.reservation.MarketOrderId = "order-consumed"
	resources.reservation.MarketBidId = "bid-consumed"
	resources.reservation.MarketLeaseId = "lease-consumed"
	resources.reservation.ProviderAddress = sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String()
	resources.reservation.Capacity = resourcesv1.ResourceCapacity{CpuCores: 2, MemoryGb: 4, StorageGb: 10}
	k.SetResourcesKeeper(resources)
	seedHPCOffering(t, ctx, k)

	job := validReservationJob(sdk.AccAddress(bytes.Repeat([]byte{8}, 20)).String())
	leaseID := marketv1.LeaseID{Owner: job.CustomerAddress, DSeq: 2, GSeq: 1, OSeq: 1, Provider: resources.reservation.ProviderAddress}
	resources.reservation.MarketOrderId = leaseID.OrderID().String()
	resources.reservation.MarketBidId = leaseID.BidID().String()
	resources.reservation.MarketLeaseId = leaseID.String()
	k.SetMarketKeeper(&hpcMarketStub{lease: marketv1.Lease{ID: leaseID, State: marketv1.LeaseActive, ReservationId: resources.reservation.ReservationId}})
	resources.reservation.RequesterAddress = job.CustomerAddress
	job.ReservationID = resources.reservation.ReservationId
	job.MarketOrderID = resources.reservation.MarketOrderId
	job.MarketBidID = resources.reservation.MarketBidId
	job.MarketLeaseID = resources.reservation.MarketLeaseId
	require.NoError(t, k.SubmitJob(ctx, &job))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateQueued, "queued", 0, nil))
}

type hpcMarketStub struct{ lease marketv1.Lease }

func (s *hpcMarketStub) GetLease(sdk.Context, marketv1.LeaseID) (marketv1.Lease, bool) {
	return s.lease, true
}

func (s *hpcMarketStub) OnDisputeOpened(sdk.Context, marketv1.LeaseID, string) error { return nil }

func TestHPCRejectsMarketLineageWithoutCanonicalReservation(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	resources := newHPCReservationStub()
	k.SetResourcesKeeper(resources)
	seedHPCOffering(t, ctx, k)

	job := validReservationJob(sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String())
	job.MarketLeaseID = "lease-orphan"
	err := k.SubmitJob(ctx, &job)
	require.ErrorContains(t, err, "market-backed jobs require")
	require.Zero(t, resources.reserveCalls)
}

type hpcReservationStub struct {
	reservation                               resourcesv1.Reservation
	reserveCalls, activateCalls, releaseCalls int
	activateErr                               error
}

func (*hpcReservationStub) IsCanonicalReservationsActive(sdk.Context) bool { return true }

func newHPCReservationStub() *hpcReservationStub {
	return &hpcReservationStub{reservation: resourcesv1.Reservation{ReservationId: "reservation-hpc", State: resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE}}
}
func (s *hpcReservationStub) Reserve(_ sdk.Context, request resourcesv1.ReservationRequest) (*resourcesv1.Reservation, error) {
	s.reserveCalls++
	s.reservation.ProviderAddress = request.ProviderAddress
	s.reservation.RequesterAddress = request.RequesterAddress
	s.reservation.ConsumerType = request.ConsumerType
	s.reservation.ConsumerId = request.ConsumerId
	s.reservation.Capacity = request.Capacity
	copy := s.reservation
	copy.State = resourcesv1.ReservationState_RESERVATION_STATE_PENDING
	return &copy, nil
}
func (s *hpcReservationStub) ActivateReservation(_ sdk.Context, _ string, link resourcesv1.ReservationLink) (*resourcesv1.Reservation, error) {
	s.activateCalls++
	if s.activateErr != nil {
		return nil, s.activateErr
	}
	s.reservation.ConsumerType = link.ConsumerType
	s.reservation.ConsumerId = link.ConsumerId
	return &s.reservation, nil
}
func (s *hpcReservationStub) ReleaseReservation(_ sdk.Context, _, _ string) (*resourcesv1.Reservation, error) {
	s.releaseCalls++
	s.reservation.State = resourcesv1.ReservationState_RESERVATION_STATE_RELEASED
	return &s.reservation, nil
}
func (s *hpcReservationStub) QuarantineReservation(_ sdk.Context, _, _ string) (*resourcesv1.Reservation, error) {
	s.reservation.State = resourcesv1.ReservationState_RESERVATION_STATE_QUARANTINED
	return &s.reservation, nil
}
func (s *hpcReservationStub) GetReservation(_ sdk.Context, _ string) (resourcesv1.Reservation, bool) {
	return s.reservation, true
}

func seedHPCOffering(t *testing.T, ctx sdk.Context, k interface {
	SetCluster(sdk.Context, types.HPCCluster) error
	SetOffering(sdk.Context, types.HPCOffering) error
}) {
	t.Helper()
	provider := sdk.AccAddress(bytes.Repeat([]byte{5}, 20)).String()
	require.NoError(t, k.SetCluster(ctx, types.HPCCluster{ClusterID: "cluster-reservation", ProviderAddress: provider, Name: "cluster", Region: "region", State: types.ClusterStateActive, TotalNodes: 2, AvailableNodes: 2, Partitions: []types.Partition{{Name: "default", Nodes: 2}}, ClusterMetadata: types.ClusterMetadata{TotalCPUCores: 8, TotalMemoryGB: 32}, SLURMVersion: "1", CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()}))
	require.NoError(t, k.SetOffering(ctx, types.HPCOffering{OfferingID: "offering-reservation", ClusterID: "cluster-reservation", ProviderAddress: provider, Name: "offering", QueueOptions: []types.QueueOption{{PartitionName: "default", MaxNodes: 2}}, MaxRuntimeSeconds: 3600, SupportsCustomWorkloads: true, Active: true, CreatedAt: ctx.BlockTime(), UpdatedAt: ctx.BlockTime()}))
}

func validReservationJob(customer string) types.HPCJob {
	return types.HPCJob{JobID: "job-reservation", OfferingID: "offering-reservation", ClusterID: "cluster-reservation", CustomerAddress: customer, QueueName: "default", WorkloadSpec: types.JobWorkloadSpec{ContainerImage: "image"}, Resources: types.JobResources{Nodes: 1, CPUCoresPerNode: 2, MemoryGBPerNode: 4, StorageGB: 10}, MaxRuntimeSeconds: 600, State: types.JobStatePending}
}
