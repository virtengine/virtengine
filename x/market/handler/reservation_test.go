package handler

import (
	"bytes"
	"errors"
	"math"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dv1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	dbeta "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	escrowtypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	markettypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	attributesv1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
	rv1beta4 "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
	resourcetypes "github.com/virtengine/virtengine/x/resources/types"
)

func TestMarketResourceCapacityUsesGeneratedUnitsAndRoundsUp(t *testing.T) {
	offers := markettypes.ResourcesOffer{{
		Count: 2,
		Resources: rv1beta4.Resources{
			CPU:     &rv1beta4.CPU{Units: rv1beta4.NewResourceValue(1500)},
			Memory:  &rv1beta4.Memory{Quantity: rv1beta4.NewResourceValue(1536 * 1024 * 1024)},
			Storage: rv1beta4.Volumes{{Quantity: rv1beta4.NewResourceValue(3*1024*1024*1024 + 1)}},
		},
	}}

	capacity, err := marketResourceCapacity(offers)
	require.NoError(t, err)
	require.Equal(t, int64(3), capacity.CpuCores)
	require.Equal(t, int64(3), capacity.MemoryGb)
	require.Equal(t, int64(7), capacity.StorageGb)
}

func TestMarketResourceCapacityRejectsOverflowAndMissingGPUType(t *testing.T) {
	_, err := marketResourceCapacity(markettypes.ResourcesOffer{{Count: 2, Resources: rv1beta4.Resources{CPU: &rv1beta4.CPU{Units: rv1beta4.NewResourceValue(math.MaxUint64)}}}})
	require.ErrorContains(t, err, "overflow")

	_, err = marketResourceCapacity(markettypes.ResourcesOffer{{Count: 1, Resources: rv1beta4.Resources{GPU: &rv1beta4.GPU{Units: rv1beta4.NewResourceValue(1)}}}})
	require.ErrorContains(t, err, "GPU type")
}

func TestMarketResourceCapacityExtractsGPUType(t *testing.T) {
	capacity, err := marketResourceCapacity(markettypes.ResourcesOffer{{Count: 2, Resources: rv1beta4.Resources{GPU: &rv1beta4.GPU{Units: rv1beta4.NewResourceValue(1), Attributes: attributesv1.Attributes{{Key: "gpu_type", Value: "nvidia-a100"}}}}}})
	require.NoError(t, err)
	require.Equal(t, int64(2), capacity.Gpus)
	require.Equal(t, "nvidia-a100", capacity.GpuType)
}

func TestCreateLeaseReservationFailureIsAtomic(t *testing.T) {
	ctx := reservationTestContext(t)
	market := newReservationMarketStub()
	resources := &reservationKeeperStub{reserveErr: resourcetypes.ErrNoEligibleInventory}
	escrow := &reservationEscrowStub{}
	server := &msgServer{keepers: Keepers{Market: market, Deployment: reservationDeploymentStub{group: market.group}, Escrow: escrow, Resources: resources}}

	_, err := server.CreateLease(ctx, &markettypes.MsgCreateLease{BidID: market.bid.ID})
	require.Error(t, err)
	require.Zero(t, escrow.paymentCreates)
	require.Zero(t, market.leaseCreates)
	require.Zero(t, market.matches)
}

func TestCreateAndCloseLeaseReservationLifecycle(t *testing.T) {
	ctx := reservationTestContext(t)
	market := newReservationMarketStub()
	resources := &reservationKeeperStub{}
	escrow := &reservationEscrowStub{}
	server := &msgServer{keepers: Keepers{Market: market, Deployment: reservationDeploymentStub{group: market.group}, Escrow: escrow, Resources: resources}}

	response, err := server.CreateLease(ctx, &markettypes.MsgCreateLease{BidID: market.bid.ID})
	require.NoError(t, err)
	require.Equal(t, "reservation-market", response.ReservationId)
	require.Equal(t, 1, resources.reserveCalls)
	require.Equal(t, 1, resources.activateCalls)
	require.Equal(t, 1, escrow.paymentCreates)
	require.Equal(t, 1, market.leaseCreates)

	_, err = server.CloseLease(ctx, &markettypes.MsgCloseLease{ID: market.bid.ID.LeaseID(), Reason: mv1.LeaseClosedReasonOwner})
	require.NoError(t, err)
	// Keeper-level close hooks own exact-once release; the handler must not
	// duplicate that cross-module mutation.
	require.Zero(t, resources.releaseCalls)
}

func TestCloseLeaseReservationFailureIsAtomic(t *testing.T) {
	ctx := reservationTestContext(t)
	market := newReservationMarketStub()
	market.order.State = markettypes.OrderActive
	market.bid.State = markettypes.BidActive
	market.closeErr = errors.New("reservation release failed")
	escrow := &reservationEscrowStub{}
	server := &msgServer{keepers: Keepers{Market: market, Deployment: reservationDeploymentStub{group: market.group}, Escrow: escrow, Resources: &reservationKeeperStub{}}}

	_, err := server.CloseLease(ctx, &markettypes.MsgCloseLease{ID: market.bid.ID.LeaseID(), Reason: mv1.LeaseClosedReasonOwner})
	require.ErrorContains(t, err, "reservation release failed")
	require.Equal(t, mv1.LeaseActive, market.lease.State)
	require.Equal(t, markettypes.BidActive, market.bid.State)
	require.Equal(t, markettypes.OrderActive, market.order.State)
	require.Zero(t, escrow.paymentCloses)
}

type reservationKeeperStub struct {
	reserveErr                                error
	reserveCalls, activateCalls, releaseCalls int
}

func (s *reservationKeeperStub) Reserve(_ sdk.Context, request resourcesv1.ReservationRequest) (*resourcesv1.Reservation, error) {
	s.reserveCalls++
	if s.reserveErr != nil {
		return nil, s.reserveErr
	}
	return &resourcesv1.Reservation{ReservationId: "reservation-market", ProviderAddress: request.ProviderAddress, ConsumerType: request.ConsumerType, ConsumerId: request.ConsumerId, Capacity: request.Capacity, State: resourcesv1.ReservationState_RESERVATION_STATE_PENDING}, nil
}
func (s *reservationKeeperStub) ActivateReservation(_ sdk.Context, id string, _ resourcesv1.ReservationLink) (*resourcesv1.Reservation, error) {
	s.activateCalls++
	return &resourcesv1.Reservation{ReservationId: id, State: resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE}, nil
}
func (s *reservationKeeperStub) ReleaseReservation(_ sdk.Context, id, _ string) (*resourcesv1.Reservation, error) {
	s.releaseCalls++
	return &resourcesv1.Reservation{ReservationId: id, State: resourcesv1.ReservationState_RESERVATION_STATE_RELEASED}, nil
}

type reservationMarketStub struct {
	order                 markettypes.Order
	bid                   markettypes.Bid
	lease                 mv1.Lease
	group                 dbeta.Group
	leaseCreates, matches int
	closeErr              error
}

func newReservationMarketStub() *reservationMarketStub {
	id := mv1.BidID{Owner: sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String(), DSeq: 1, GSeq: 1, OSeq: 1, Provider: sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String(), BSeq: 0}
	return &reservationMarketStub{order: markettypes.Order{ID: id.OrderID(), State: markettypes.OrderOpen}, bid: markettypes.Bid{ID: id, State: markettypes.BidOpen, ResourcesOffer: markettypes.ResourcesOffer{{Resources: rv1beta4.Resources{CPU: &rv1beta4.CPU{Units: rv1beta4.NewResourceValue(1)}}, Count: 1}}}, lease: mv1.Lease{ID: id.LeaseID(), State: mv1.LeaseActive, ReservationId: "reservation-market"}, group: dbeta.Group{ID: id.GroupID(), State: dbeta.GroupOpen}}
}
func (s *reservationMarketStub) GetParams(sdk.Context) markettypes.Params {
	return markettypes.DefaultParams()
}
func (s *reservationMarketStub) GetAuthority() string                             { return "virtengine1authority" }
func (s *reservationMarketStub) SetParams(sdk.Context, markettypes.Params) error  { return nil }
func (s *reservationMarketStub) BidCountForOrder(sdk.Context, mv1.OrderID) uint32 { return 0 }
func (s *reservationMarketStub) GetBid(sdk.Context, mv1.BidID) (markettypes.Bid, bool) {
	return s.bid, true
}
func (s *reservationMarketStub) GetOrder(sdk.Context, mv1.OrderID) (markettypes.Order, bool) {
	return s.order, true
}
func (s *reservationMarketStub) GetLease(sdk.Context, mv1.LeaseID) (mv1.Lease, bool) {
	return s.lease, true
}
func (s *reservationMarketStub) CreateBid(sdk.Context, mv1.BidID, sdk.DecCoin, markettypes.ResourcesOffer) (markettypes.Bid, error) {
	return s.bid, nil
}
func (s *reservationMarketStub) CreateLease(sdk.Context, markettypes.Bid) error {
	s.leaseCreates++
	return nil
}
func (s *reservationMarketStub) CreateOrder(sdk.Context, dv1.GroupID, dbeta.GroupSpec) (markettypes.Order, error) {
	return s.order, nil
}
func (s *reservationMarketStub) OnOrderMatched(sdk.Context, markettypes.Order) {
	s.matches++
	s.order.State = markettypes.OrderActive
}
func (s *reservationMarketStub) OnBidMatched(sdk.Context, markettypes.Bid) {
	s.matches++
	s.bid.State = markettypes.BidActive
}
func (s *reservationMarketStub) OnBidLost(sdk.Context, markettypes.Bid) {}
func (s *reservationMarketStub) OnBidClosed(_ sdk.Context, _ markettypes.Bid) error {
	s.bid.State = markettypes.BidClosed
	return nil
}
func (s *reservationMarketStub) OnOrderClosed(_ sdk.Context, _ markettypes.Order) error {
	s.order.State = markettypes.OrderClosed
	return nil
}
func (s *reservationMarketStub) OnLeaseClosed(_ sdk.Context, lease mv1.Lease, _ mv1.Lease_State, _ mv1.LeaseClosedReason) error {
	if s.closeErr != nil {
		return s.closeErr
	}
	lease.State = mv1.LeaseClosed
	s.lease = lease
	return nil
}
func (s *reservationMarketStub) WithBidsForOrder(sdk.Context, mv1.OrderID, markettypes.Bid_State, func(markettypes.Bid) bool) {
}
func (s *reservationMarketStub) WithOrders(sdk.Context, func(markettypes.Order) bool) {}
func (s *reservationMarketStub) WithBids(sdk.Context, func(markettypes.Bid) bool)     {}

type reservationDeploymentStub struct{ group dbeta.Group }

func (s reservationDeploymentStub) GetGroup(sdk.Context, dv1.GroupID) (dbeta.Group, bool) {
	return s.group, true
}
func (s reservationDeploymentStub) OnBidClosed(sdk.Context, dv1.GroupID) error { return nil }
func (s reservationDeploymentStub) OnLeaseClosed(sdk.Context, dv1.GroupID) (dbeta.Group, error) {
	return s.group, nil
}

type reservationEscrowStub struct{ paymentCreates, paymentCloses int }

func (s *reservationEscrowStub) AccountCreate(sdk.Context, escrowid.Account, sdk.AccAddress, []escrowtypes.Depositor) error {
	return nil
}
func (s *reservationEscrowStub) AccountDeposit(sdk.Context, escrowid.Account, []escrowtypes.Depositor) error {
	return nil
}
func (s *reservationEscrowStub) AccountClose(sdk.Context, escrowid.Account) error { return nil }
func (s *reservationEscrowStub) PaymentCreate(sdk.Context, escrowid.Payment, sdk.AccAddress, sdk.DecCoin) error {
	s.paymentCreates++
	return nil
}
func (s *reservationEscrowStub) PaymentWithdraw(sdk.Context, escrowid.Payment) error { return nil }
func (s *reservationEscrowStub) PaymentClose(sdk.Context, escrowid.Payment) error {
	s.paymentCloses++
	return nil
}
func (s *reservationEscrowStub) AuthorizeDeposits(sdk.Context, sdk.Msg) ([]escrowtypes.Depositor, error) {
	return nil, nil
}

func reservationTestContext(t *testing.T) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	key := storetypes.NewKVStoreKey("market-reservation-test")
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	return sdk.NewContext(stateStore, cmtproto.Header{Height: 1, Time: time.Unix(1, 0).UTC()}, false, log.NewNopLogger())
}
