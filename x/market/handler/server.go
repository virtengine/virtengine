package handler

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	atypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	dbeta "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	v1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	attributesv1 "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

const (
	marketCPUUnitsPerCore = uint64(1000)
	marketBytesPerGiB     = uint64(1024 * 1024 * 1024)
)

type msgServer struct {
	keepers Keepers
}

// NewServer returns an implementation of the market MsgServer interface
// for the provided Keeper.
func NewServer(k Keepers) types.MsgServer {
	return &msgServer{keepers: k}
}

var _ types.MsgServer = msgServer{}

func (ms msgServer) CreateBid(goCtx context.Context, msg *types.MsgCreateBid) (*types.MsgCreateBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := ms.keepers.Market.GetParams(ctx)

	minDeposit := params.BidMinDeposit
	if msg.Deposit.Amount.Denom != minDeposit.Denom {
		return nil, fmt.Errorf("%w: mininum:%v received:%v", v1.ErrInvalidDeposit, minDeposit, msg.Deposit)
	}
	if minDeposit.Amount.GT(msg.Deposit.Amount.Amount) {
		return nil, fmt.Errorf("%w: mininum:%v received:%v", v1.ErrInvalidDeposit, minDeposit, msg.Deposit)
	}

	if ms.keepers.Market.BidCountForOrder(ctx, msg.ID.OrderID()) > params.OrderMaxBids {
		return nil, fmt.Errorf("%w: too many existing bids (%v)", v1.ErrInvalidBid, params.OrderMaxBids)
	}

	if msg.ID.BSeq != 0 {
		return nil, v1.ErrInvalidBid
	}

	order, found := ms.keepers.Market.GetOrder(ctx, msg.ID.OrderID())
	if !found {
		return nil, v1.ErrOrderNotFound
	}

	if err := order.ValidateCanBid(); err != nil {
		return nil, err
	}

	if !msg.Price.IsValid() {
		return nil, v1.ErrBidInvalidPrice
	}

	if order.Price().IsLT(msg.Price) {
		return nil, v1.ErrBidOverOrder
	}

	if !msg.ResourcesOffer.MatchGSpec(order.Spec) {
		return nil, v1.ErrCapabilitiesMismatch
	}

	provider, err := sdk.AccAddressFromBech32(msg.ID.Provider)
	if err != nil {
		return nil, v1.ErrEmptyProvider
	}

	var prov ptypes.Provider
	if prov, found = ms.keepers.Provider.Get(ctx, provider); !found {
		return nil, v1.ErrUnknownProvider
	}

	provAttr, _ := ms.keepers.Audit.GetProviderAttributes(ctx, provider)

	provAttr = append([]atypes.AuditedProvider{{
		Owner:      msg.ID.Provider,
		Attributes: prov.Attributes,
	}}, provAttr...)

	if !order.MatchRequirements(provAttr) {
		return nil, v1.ErrAttributeMismatch
	}

	if !order.MatchResourcesRequirements(prov.Attributes) {
		return nil, v1.ErrCapabilitiesMismatch
	}

	deposits, err := ms.keepers.Escrow.AuthorizeDeposits(ctx, msg)
	if err != nil {
		return nil, err
	}

	bid, err := ms.keepers.Market.CreateBid(ctx, msg.ID, msg.Price, msg.ResourcesOffer)
	if err != nil {
		return nil, err
	}

	// create an escrow account for this bid
	err = ms.keepers.Escrow.AccountCreate(ctx, bid.ID.ToEscrowAccountID(), provider, deposits)
	if err != nil {
		return &types.MsgCreateBidResponse{}, err
	}

	telemetry.IncrCounter(1.0, "ve.bids")
	return &types.MsgCreateBidResponse{}, nil
}

func (ms msgServer) CloseBid(goCtx context.Context, msg *types.MsgCloseBid) (*types.MsgCloseBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.closeBid(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms msgServer) closeBid(ctx sdk.Context, msg *types.MsgCloseBid) (*types.MsgCloseBidResponse, error) {

	bid, found := ms.keepers.Market.GetBid(ctx, msg.ID)
	if !found {
		return nil, v1.ErrUnknownBid
	}

	order, found := ms.keepers.Market.GetOrder(ctx, msg.ID.OrderID())
	if !found {
		return nil, v1.ErrUnknownOrderForBid
	}

	if bid.State == types.BidOpen {
		if err := ms.keepers.Market.OnBidClosed(ctx, bid); err != nil {
			return nil, err
		}
		return &types.MsgCloseBidResponse{}, nil
	}

	lease, found := ms.keepers.Market.GetLease(ctx, v1.LeaseID(msg.ID))
	if !found {
		return nil, v1.ErrUnknownLeaseForBid
	}

	if lease.State != v1.LeaseActive {
		return nil, v1.ErrLeaseNotActive
	}

	if bid.State != types.BidActive {
		return nil, v1.ErrBidNotActive
	}

	if err := ms.keepers.Deployment.OnBidClosed(ctx, order.ID.GroupID()); err != nil {
		return nil, err
	}

	if err := ms.keepers.Market.OnLeaseClosed(ctx, lease, v1.LeaseClosed, msg.Reason); err != nil {
		return nil, err
	}
	if err := ms.keepers.Market.OnBidClosed(ctx, bid); err != nil {
		return nil, err
	}
	if err := ms.keepers.Market.OnOrderClosed(ctx, order); err != nil {
		return nil, err
	}

	if err := ms.keepers.Escrow.PaymentClose(ctx, lease.ID.ToEscrowPaymentID()); err != nil {
		return nil, err
	}

	telemetry.IncrCounter(1.0, "ve.order_closed")

	return &types.MsgCloseBidResponse{}, nil
}

func (ms msgServer) WithdrawLease(goCtx context.Context, msg *types.MsgWithdrawLease) (*types.MsgWithdrawLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := ms.keepers.Market.GetLease(ctx, msg.ID)
	if !found {
		return nil, v1.ErrUnknownLease
	}

	if err := ms.keepers.Escrow.PaymentWithdraw(ctx, msg.ID.ToEscrowPaymentID()); err != nil {
		return &types.MsgWithdrawLeaseResponse{}, err
	}

	return &types.MsgWithdrawLeaseResponse{}, nil
}

func (ms msgServer) CreateLease(goCtx context.Context, msg *types.MsgCreateLease) (*types.MsgCreateLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.createLease(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms msgServer) createLease(ctx sdk.Context, msg *types.MsgCreateLease) (*types.MsgCreateLeaseResponse, error) {

	bid, found := ms.keepers.Market.GetBid(ctx, msg.BidID)
	if !found {
		return &types.MsgCreateLeaseResponse{}, v1.ErrBidNotFound
	}

	if bid.State != types.BidOpen {
		return &types.MsgCreateLeaseResponse{}, v1.ErrBidNotOpen
	}

	order, found := ms.keepers.Market.GetOrder(ctx, msg.BidID.OrderID())
	if !found {
		return &types.MsgCreateLeaseResponse{}, v1.ErrOrderNotFound
	}

	if order.State != types.OrderOpen {
		return &types.MsgCreateLeaseResponse{}, v1.ErrOrderNotOpen
	}

	group, found := ms.keepers.Deployment.GetGroup(ctx, order.ID.GroupID())
	if !found {
		return &types.MsgCreateLeaseResponse{}, v1.ErrGroupNotFound
	}

	if group.State != dbeta.GroupOpen {
		return &types.MsgCreateLeaseResponse{}, v1.ErrGroupNotOpen
	}

	provider, err := sdk.AccAddressFromBech32(msg.BidID.Provider)
	if err != nil {
		return &types.MsgCreateLeaseResponse{}, err
	}

	if ms.keepers.Resources == nil {
		return nil, fmt.Errorf("resources keeper is required for lease creation")
	}
	capacity, err := marketResourceCapacity(bid.ResourcesOffer)
	if err != nil {
		return nil, err
	}
	leaseID := msg.BidID.LeaseID()
	reservationRequest := resourcesv1.ReservationRequest{
		IdempotencyKey: "market/lease/" + leaseID.String(), RequestId: "market/order/" + order.ID.String(),
		RequesterAddress: order.ID.Owner, ProviderAddress: msg.BidID.Provider,
		ResourceClass: resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE, Capacity: capacity,
		ConsumerType: "market_lease", ConsumerId: leaseID.String(), MarketOrderId: order.ID.String(),
		MarketBidId: bid.ID.String(), MarketLeaseId: leaseID.String(), EscrowId: "market/payment/" + leaseID.String(), Version: 1,
	}
	reservation, err := ms.keepers.Resources.Reserve(ctx, reservationRequest)
	if err != nil {
		return nil, err
	}
	bid.ReservationId = reservation.ReservationId
	order.ReservationId = reservation.ReservationId

	err = ms.keepers.Escrow.PaymentCreate(ctx, leaseID.ToEscrowPaymentID(), provider, bid.Price)
	if err != nil {
		return &types.MsgCreateLeaseResponse{}, err
	}

	err = ms.keepers.Market.CreateLease(ctx, bid)
	if err != nil {
		return &types.MsgCreateLeaseResponse{}, err
	}

	ms.keepers.Market.OnOrderMatched(ctx, order)
	ms.keepers.Market.OnBidMatched(ctx, bid)
	if _, err = ms.keepers.Resources.ActivateReservation(ctx, reservation.ReservationId, resourcesv1.ReservationLink{
		ConsumerType: "market_lease", ConsumerId: leaseID.String(), MarketOrderId: order.ID.String(),
		MarketBidId: bid.ID.String(), MarketLeaseId: leaseID.String(), EscrowId: "market/payment/" + leaseID.String(),
	}); err != nil {
		return nil, err
	}

	// close losing bids
	ms.keepers.Market.WithBidsForOrder(ctx, msg.BidID.OrderID(), types.BidOpen, func(cbid types.Bid) bool {
		ms.keepers.Market.OnBidLost(ctx, cbid)

		if err = ms.keepers.Escrow.AccountClose(ctx, cbid.ID.ToEscrowAccountID()); err != nil {
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateLeaseResponse{ReservationId: reservation.ReservationId}, nil
}

func (ms msgServer) CloseLease(goCtx context.Context, msg *types.MsgCloseLease) (*types.MsgCloseLeaseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	cacheCtx, write := ctx.CacheContext()
	response, err := ms.closeLease(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	write()
	return response, nil
}

func (ms msgServer) closeLease(ctx sdk.Context, msg *types.MsgCloseLease) (*types.MsgCloseLeaseResponse, error) {

	order, found := ms.keepers.Market.GetOrder(ctx, msg.ID.OrderID())
	if !found {
		return nil, v1.ErrOrderNotFound
	}

	if order.State != types.OrderActive {
		return &types.MsgCloseLeaseResponse{}, v1.ErrOrderClosed
	}

	bid, found := ms.keepers.Market.GetBid(ctx, msg.ID.BidID())
	if !found {
		return &types.MsgCloseLeaseResponse{}, v1.ErrBidNotFound
	}
	if bid.State != types.BidActive {
		return &types.MsgCloseLeaseResponse{}, v1.ErrBidNotActive
	}

	lease, found := ms.keepers.Market.GetLease(ctx, msg.ID)
	if !found {
		return &types.MsgCloseLeaseResponse{}, v1.ErrLeaseNotFound
	}
	if lease.State != v1.LeaseActive {
		return &types.MsgCloseLeaseResponse{}, v1.ErrOrderClosed
	}

	if err := ms.keepers.Market.OnLeaseClosed(ctx, lease, v1.LeaseClosed, v1.LeaseClosedReasonOwner); err != nil {
		return nil, err
	}
	if err := ms.keepers.Market.OnBidClosed(ctx, bid); err != nil {
		return nil, err
	}
	if err := ms.keepers.Market.OnOrderClosed(ctx, order); err != nil {
		return nil, err
	}
	err := ms.keepers.Escrow.PaymentClose(ctx, lease.ID.ToEscrowPaymentID())
	if err != nil {
		return &types.MsgCloseLeaseResponse{}, err
	}

	group, err := ms.keepers.Deployment.OnLeaseClosed(ctx, msg.ID.GroupID())
	if err != nil {
		return &types.MsgCloseLeaseResponse{}, err
	}

	if group.State != dbeta.GroupOpen {
		return &types.MsgCloseLeaseResponse{}, nil
	}

	if _, err := ms.keepers.Market.CreateOrder(ctx, group.ID, group.GroupSpec); err != nil {
		return &types.MsgCloseLeaseResponse{}, err
	}

	return &types.MsgCloseLeaseResponse{}, nil
}

func marketResourceCapacity(offers types.ResourcesOffer) (resourcesv1.ResourceCapacity, error) {
	capacity := resourcesv1.ResourceCapacity{}
	for _, offer := range offers {
		count := int64(offer.Count)
		if count <= 0 {
			return capacity, fmt.Errorf("resource offer count must be positive")
		}
		if offer.Resources.CPU != nil {
			units, err := checkedMarketProductUint(offer.Resources.CPU.Units.Value(), uint64(offer.Count))
			if err != nil {
				return capacity, err
			}
			value, err := checkedMarketUintToInt64(ceilMarketUnits(units, marketCPUUnitsPerCore))
			if err != nil {
				return capacity, err
			}
			capacity.CpuCores, err = checkedMarketAdd(capacity.CpuCores, value)
			if err != nil {
				return capacity, err
			}
		}
		if offer.Resources.Memory != nil {
			bytes, err := checkedMarketProductUint(offer.Resources.Memory.Quantity.Value(), uint64(offer.Count))
			if err != nil {
				return capacity, err
			}
			value, err := checkedMarketUintToInt64(ceilMarketUnits(bytes, marketBytesPerGiB))
			if err != nil {
				return capacity, err
			}
			capacity.MemoryGb, err = checkedMarketAdd(capacity.MemoryGb, value)
			if err != nil {
				return capacity, err
			}
		}
		for _, storage := range offer.Resources.Storage {
			bytes, err := checkedMarketProductUint(storage.Quantity.Value(), uint64(offer.Count))
			if err != nil {
				return capacity, err
			}
			value, err := checkedMarketUintToInt64(ceilMarketUnits(bytes, marketBytesPerGiB))
			if err != nil {
				return capacity, err
			}
			capacity.StorageGb, err = checkedMarketAdd(capacity.StorageGb, value)
			if err != nil {
				return capacity, err
			}
		}
		if offer.Resources.GPU != nil {
			value, err := checkedMarketProduct(offer.Resources.GPU.Units.Value(), uint64(offer.Count))
			if err != nil {
				return capacity, err
			}
			capacity.Gpus, err = checkedMarketAdd(capacity.Gpus, value)
			if err != nil {
				return capacity, err
			}
			gpuType, err := marketGPUType(offer.Resources.GPU.Attributes)
			if err != nil {
				return capacity, err
			}
			if value > 0 && gpuType == "" {
				return capacity, fmt.Errorf("GPU type is required when GPU capacity is nonzero")
			}
			if capacity.GpuType != "" && gpuType != "" && capacity.GpuType != gpuType {
				return capacity, fmt.Errorf("resource offer contains multiple GPU types")
			}
			if capacity.GpuType == "" {
				capacity.GpuType = gpuType
			}
		}
	}
	if capacity.CpuCores == 0 && capacity.MemoryGb == 0 && capacity.StorageGb == 0 && capacity.Gpus == 0 {
		return capacity, fmt.Errorf("bid contains no reservable capacity")
	}
	return capacity, nil
}

func checkedMarketProduct(value, count uint64) (int64, error) {
	if count > 0 && value > uint64(math.MaxInt64)/count {
		return 0, fmt.Errorf("resource capacity overflow")
	}
	return int64(value * count), nil //nolint:gosec // bounded above
}

func checkedMarketProductUint(value, count uint64) (uint64, error) {
	if count > 0 && value > math.MaxUint64/count {
		return 0, fmt.Errorf("resource capacity overflow")
	}
	return value * count, nil
}

func checkedMarketUintToInt64(value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("resource capacity overflow")
	}
	return int64(value), nil //nolint:gosec // bounded above
}

func ceilMarketUnits(value, unit uint64) uint64 {
	if value == 0 {
		return 0
	}
	return 1 + (value-1)/unit
}

func marketGPUType(attributes attributesv1.Attributes) (string, error) {
	var gpuType string
	for _, attribute := range attributes {
		key := strings.ToLower(strings.TrimSpace(attribute.Key))
		value := strings.TrimSpace(attribute.Value)
		candidate := ""
		switch {
		case key == "gpu_type" || strings.HasSuffix(key, "/gpu_type"):
			candidate = value
		case strings.HasPrefix(key, "vendor/nvidia/model/"):
			candidate = strings.Split(strings.TrimPrefix(key, "vendor/nvidia/model/"), "/")[0]
		}
		if candidate == "" || candidate == "*" {
			continue
		}
		if gpuType != "" && gpuType != candidate {
			return "", fmt.Errorf("resource offer contains conflicting GPU attributes")
		}
		gpuType = candidate
	}
	return gpuType, nil
}

func checkedMarketAdd(current, value int64) (int64, error) {
	if value < 0 || current > math.MaxInt64-value {
		return 0, fmt.Errorf("resource capacity overflow")
	}
	return current + value, nil
}

func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.keepers.Market.GetAuthority() != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", ms.keepers.Market.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.keepers.Market.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
