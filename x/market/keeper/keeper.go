package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	dtypesBeta "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"

	"github.com/virtengine/virtengine/x/market/keeper/keys"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type IKeeper interface {
	NewQuerier() Querier
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
	CreateOrder(ctx sdk.Context, gid dtypes.GroupID, spec dtypesBeta.GroupSpec) (types.Order, error)
	CreateBid(ctx sdk.Context, id mv1.BidID, price sdk.DecCoin, roffer types.ResourcesOffer) (types.Bid, error)
	CreateLease(ctx sdk.Context, bid types.Bid) error
	OnOrderMatched(ctx sdk.Context, order types.Order)
	OnBidMatched(ctx sdk.Context, bid types.Bid)
	OnBidLost(ctx sdk.Context, bid types.Bid)
	OnBidClosed(ctx sdk.Context, bid types.Bid) error
	OnOrderClosed(ctx sdk.Context, order types.Order) error
	OnLeaseClosed(ctx sdk.Context, lease mv1.Lease, state mv1.Lease_State, reason mv1.LeaseClosedReason) error
	OnGroupClosed(ctx sdk.Context, id dtypes.GroupID) error
	GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool)
	GetBid(ctx sdk.Context, id mv1.BidID) (types.Bid, bool)
	GetLease(ctx sdk.Context, id mv1.LeaseID) (mv1.Lease, bool)
	LeaseForOrder(ctx sdk.Context, bs types.Bid_State, oid mv1.OrderID) (mv1.Lease, bool)
	WithOrders(ctx sdk.Context, fn func(types.Order) bool)
	WithBids(ctx sdk.Context, fn func(types.Bid) bool)
	WithBidsForOrder(ctx sdk.Context, id mv1.OrderID, state types.Bid_State, fn func(types.Bid) bool)
	WithLeases(ctx sdk.Context, fn func(mv1.Lease) bool)
	WithOrdersForGroup(ctx sdk.Context, id dtypes.GroupID, state types.Order_State, fn func(types.Order) bool)
	BidCountForOrder(ctx sdk.Context, id mv1.OrderID) uint32
	GetParams(ctx sdk.Context) (params types.Params)
	SetParams(ctx sdk.Context, params types.Params) error
	GetAuthority() string
	// Additional methods for cross-module integration
	GetOrderByID(ctx sdk.Context, orderID string) (interface{}, bool)
	IsOrderCompleted(ctx sdk.Context, orderID string) bool
	GetOrderCustomer(ctx sdk.Context, orderID string) string
	GetOrderProvider(ctx sdk.Context, orderID string) string
	GetOrderCompletedAt(ctx sdk.Context, orderID string) time.Time
	GetOrderHash(ctx sdk.Context, orderID string) string
	// ProviderHasActiveLeases checks if a provider has any active leases
	ProviderHasActiveLeases(ctx sdk.Context, provider sdk.AccAddress) bool
	// VEID Gating (VE-301: Marketplace gating)
	CheckVEIDGating(ctx sdk.Context, customerAddr sdk.AccAddress, requirements VEIDGatingRequirements) (*VEIDGatingResult, error)
	SetVEIDKeeper(veidKeeper VEIDKeeper)
}

// Keeper of the market store
type Keeper struct {
	cdc     codec.BinaryCodec
	skey    storetypes.StoreKey
	ekeeper EscrowKeeper
	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
	// veidKeeper is the VEID module keeper for identity gating checks.
	// This is optional and can be set after initialization via SetVEIDKeeper.
	// VE-301: Marketplace gating
	veidKeeper VEIDKeeper
}

// NewKeeper creates and returns an instance for Market keeper
func NewKeeper(cdc codec.BinaryCodec, skey storetypes.StoreKey, ekeeper EscrowKeeper, authority string) IKeeper {
	return &Keeper{
		skey:      skey,
		cdc:       cdc,
		ekeeper:   ekeeper,
		authority: authority,
	}
}

// SetVEIDKeeper sets the VEID keeper for identity gating checks.
// This is called during app initialization to wire in the VEID module.
// VE-301: Marketplace gating
func (k *Keeper) SetVEIDKeeper(veidKeeper VEIDKeeper) {
	k.veidKeeper = veidKeeper
}

func (k Keeper) NewQuerier() Querier {
	return Querier{k}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the x/mint module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the x/market module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(mv1.ParamsPrefix(), bz)

	return nil
}

// GetParams returns the current x/market module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(mv1.ParamsPrefix())
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// CreateOrder creates a new order with given group id and specifications. It returns created order.
// VE-301: This function now performs VEID gating checks if a VEIDKeeper is configured.
func (k Keeper) CreateOrder(ctx sdk.Context, gid dtypes.GroupID, spec dtypesBeta.GroupSpec) (types.Order, error) {
	store := ctx.KVStore(k.skey)

	// VE-301: Check VEID gating requirements before creating order
	// The owner (customer) address is extracted from the GroupID
	customerAddr, err := sdk.AccAddressFromBech32(gid.Owner)
	if err != nil {
		return types.Order{}, fmt.Errorf("invalid owner address in group ID: %w", err)
	}

	// Get VEID gating requirements from params or use defaults
	requirements := k.getVEIDGatingRequirementsForOrder(ctx)

	// Perform VEID gating check if requirements are non-zero
	if k.veidKeeper != nil && !isZeroRequirements(requirements) {
		result, gatingErr := k.CheckVEIDGating(ctx, customerAddr, requirements)
		if gatingErr != nil {
			ctx.Logger().Info("VEID gating failed for order creation",
				"owner", gid.Owner,
				"score", result.CustomerScore,
				"tier", result.CustomerTier,
				"failures", len(result.FailureReasons),
			)
			return types.Order{}, gatingErr
		}
		ctx.Logger().Debug("VEID gating passed for order creation",
			"owner", gid.Owner,
			"score", result.CustomerScore,
			"tier", result.CustomerTier,
		)
	}

	oseq := uint32(1)

	k.WithOrdersForGroup(ctx, gid, types.OrderActive, func(_ types.Order) bool {
		err = mv1.ErrOrderActive
		return true
	})

	k.WithOrdersForGroup(ctx, gid, types.OrderOpen, func(_ types.Order) bool {
		err = mv1.ErrOrderActive
		return true
	})

	k.WithOrdersForGroup(ctx, gid, types.OrderClosed, func(_ types.Order) bool {
		oseq++
		return false
	})

	if err != nil {
		return types.Order{}, fmt.Errorf("%w: create order: active order exists", err)
	}

	orderID := mv1.MakeOrderID(gid, oseq)

	if res := k.findOrder(ctx, orderID); len(res) > 0 {
		return types.Order{}, mv1.ErrOrderExists
	}

	order := types.Order{
		ID:        mv1.MakeOrderID(gid, oseq),
		Spec:      spec,
		State:     types.OrderOpen,
		CreatedAt: ctx.BlockHeight(),
	}

	key := keys.MustOrderKey(keys.OrderStateOpenPrefix, order.ID)
	store.Set(key, k.cdc.MustMarshal(&order))

	ctx.Logger().Info("created order", "order", order.ID)

	err = ctx.EventManager().EmitTypedEvent(
		&mv1.EventOrderCreated{ID: order.ID},
	)
	if err != nil {
		return types.Order{}, err
	}

	return order, nil
}

// getVEIDGatingRequirementsForOrder returns the VEID gating requirements for order creation.
// This reads from market params if available, otherwise returns defaults.
// VE-301: Marketplace gating
func (k Keeper) getVEIDGatingRequirementsForOrder(ctx sdk.Context) VEIDGatingRequirements {
	// For now, return defaults. In a future iteration, this could read from
	// market params to allow governance to set minimum requirements.
	// params := k.GetParams(ctx)
	// if params.VEIDGatingEnabled {
	//     return VEIDGatingRequirements{
	//         MinCustomerScore:        params.MinCustomerScore,
	//         MinCustomerTier:         params.MinCustomerTier,
	//         RequireVerifiedStatus:   params.RequireVerifiedStatus,
	//         RequireUnlockedIdentity: true,
	//     }
	// }
	return DefaultVEIDGatingRequirements()
}

// isZeroRequirements checks if the requirements are effectively empty (no gating).
func isZeroRequirements(req VEIDGatingRequirements) bool {
	return req.MinCustomerScore == 0 &&
		req.MinCustomerTier == 0 &&
		len(req.RequiredScopes) == 0 &&
		!req.RequireVerifiedStatus
}

// CreateBid creates a bid for a order with given orderID, price for bid and provider
func (k Keeper) CreateBid(ctx sdk.Context, id mv1.BidID, price sdk.DecCoin, roffer types.ResourcesOffer) (types.Bid, error) {
	store := ctx.KVStore(k.skey)

	if key := k.findBid(ctx, id); len(key) > 0 {
		return types.Bid{}, mv1.ErrBidExists
	}

	bid := types.Bid{
		ID:             id,
		State:          types.BidOpen,
		Price:          price,
		CreatedAt:      ctx.BlockHeight(),
		ResourcesOffer: roffer,
	}

	data := k.cdc.MustMarshal(&bid)

	key := keys.MustBidKey(keys.BidStateToPrefix(bid.State), id)
	revKey := keys.MustBidStateRevereKey(bid.State, id)

	store.Set(key, data)

	if len(revKey) > 0 {
		store.Set(revKey, data)
	}

	err := ctx.EventManager().EmitTypedEvent(
		&mv1.EventBidCreated{
			ID:    bid.ID,
			Price: price,
		},
	)
	if err != nil {
		return types.Bid{}, err
	}

	return bid, nil
}

// CreateLease creates lease for bid with given bidID.
// Should only be called by the EndBlock handler or unit tests.
func (k Keeper) CreateLease(ctx sdk.Context, bid types.Bid) error {
	store := ctx.KVStore(k.skey)

	lease := mv1.Lease{
		ID:        mv1.LeaseID(bid.ID),
		State:     mv1.LeaseActive,
		Price:     bid.Price,
		CreatedAt: ctx.BlockHeight(),
	}

	data := k.cdc.MustMarshal(&lease)

	// create (active) lease in store
	key := keys.MustLeaseKey(keys.LeaseStateToPrefix(lease.State), lease.ID)
	revKey := keys.MustLeaseStateReverseKey(lease.State, lease.ID)

	store.Set(key, data)
	if len(revKey) > 0 {
		store.Set(revKey, data)
	}

	err := ctx.EventManager().EmitTypedEvent(
		&mv1.EventLeaseCreated{
			ID:    lease.ID,
			Price: lease.Price,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// OnOrderMatched updates order state to matched
func (k Keeper) OnOrderMatched(ctx sdk.Context, order types.Order) {
	currState := order.State
	order.State = types.OrderActive
	k.updateOrder(ctx, order, currState)
}

// OnBidMatched updates bid state to matched
func (k Keeper) OnBidMatched(ctx sdk.Context, bid types.Bid) {
	currState := bid.State
	bid.State = types.BidActive
	k.updateBid(ctx, bid, currState)
}

// OnBidLost updates bid state to bid lost
func (k Keeper) OnBidLost(ctx sdk.Context, bid types.Bid) {
	currState := bid.State
	bid.State = types.BidLost
	k.updateBid(ctx, bid, currState)
}

// OnBidClosed updates bid state to closed
func (k Keeper) OnBidClosed(ctx sdk.Context, bid types.Bid) error {
	switch bid.State {
	case types.BidClosed, types.BidLost:
		return nil
	}

	currState := bid.State
	bid.State = types.BidClosed
	k.updateBid(ctx, bid, currState)

	_ = k.ekeeper.AccountClose(ctx, bid.ID.ToEscrowAccountID())

	err := ctx.EventManager().EmitTypedEvent(
		&mv1.EventBidClosed{
			ID: bid.ID,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// OnOrderClosed updates order state to closed
func (k Keeper) OnOrderClosed(ctx sdk.Context, order types.Order) error {
	if order.State == types.OrderClosed {
		return nil
	}

	currState := order.State

	order.State = types.OrderClosed

	k.updateOrder(ctx, order, currState)

	err := ctx.EventManager().EmitTypedEvent(
		&mv1.EventOrderClosed{
			ID: order.ID,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// OnLeaseClosed updates lease state to closed
func (k Keeper) OnLeaseClosed(ctx sdk.Context, lease mv1.Lease, state mv1.Lease_State, reason mv1.LeaseClosedReason) error {
	switch lease.State {
	case mv1.LeaseClosed, mv1.LeaseInsufficientFunds:
		return nil
	}

	currState := lease.State

	lease.State = state
	lease.ClosedOn = ctx.BlockHeight()

	store := ctx.KVStore(k.skey)

	key := keys.MustLeaseKey(keys.LeaseStateToPrefix(currState), lease.ID)
	revKey := keys.MustLeaseStateReverseKey(currState, lease.ID)

	store.Delete(key)
	if len(revKey) > 0 {
		store.Delete(revKey)
	}

	key = keys.MustLeaseKey(keys.LeaseStateToPrefix(lease.State), lease.ID)
	store.Set(key, k.cdc.MustMarshal(&lease))

	err := ctx.EventManager().EmitTypedEvent(
		&mv1.EventLeaseClosed{
			ID:     lease.ID,
			Reason: reason,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// OnGroupClosed updates state of all orders, bids and leases in group to closed
func (k Keeper) OnGroupClosed(ctx sdk.Context, id dtypes.GroupID) error {
	processClose := func(ctx sdk.Context, bid types.Bid) error {
		err := k.OnBidClosed(ctx, bid)
		if err != nil {
			return err
		}

		if lease, ok := k.GetLease(ctx, bid.ID.LeaseID()); ok {
			// OnGroupClosed is callable by x/deployment only so only reason is owner
			err = k.OnLeaseClosed(ctx, lease, mv1.LeaseClosed, mv1.LeaseClosedReasonOwner)
			if err != nil {
				return err
			}
			if err := k.ekeeper.PaymentClose(ctx, lease.ID.ToEscrowPaymentID()); err != nil {
				ctx.Logger().With("err", err).Info("error closing payment")
			}
			if err != nil {
				return err
			}
		}

		return nil
	}

	var err error
	k.WithOrdersForGroup(ctx, id, types.OrderActive, func(order types.Order) bool {
		err = k.OnOrderClosed(ctx, order)
		if err != nil {
			return true
		}

		k.WithBidsForOrder(ctx, order.ID, types.BidOpen, func(bid types.Bid) bool {
			err = processClose(ctx, bid)
			return err != nil
		})

		if err != nil {
			return true
		}

		k.WithBidsForOrder(ctx, order.ID, types.BidActive, func(bid types.Bid) bool {
			err = processClose(ctx, bid)
			return err != nil
		})

		return err != nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) findOrder(ctx sdk.Context, id mv1.OrderID) []byte {
	store := ctx.KVStore(k.skey)

	aKey := keys.MustOrderKey(keys.OrderStateActivePrefix, id)
	oKey := keys.MustOrderKey(keys.OrderStateOpenPrefix, id)
	cKey := keys.MustOrderKey(keys.OrderStateClosedPrefix, id)

	var key []byte

	// nolint: gocritic
	if store.Has(aKey) {
		key = aKey
	} else if store.Has(oKey) {
		key = oKey
	} else if store.Has(cKey) {
		key = cKey
	}

	return key
}

// GetOrder returns order with given orderID from market store
func (k Keeper) GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool) {
	key := k.findOrder(ctx, id)

	if len(key) == 0 {
		return types.Order{}, false
	}

	store := ctx.KVStore(k.skey)

	buf := store.Get(key)

	var val types.Order
	k.cdc.MustUnmarshal(buf, &val)

	return val, true
}

func (k Keeper) findBid(ctx sdk.Context, id mv1.BidID) []byte {
	store := ctx.KVStore(k.skey)

	aKey := keys.MustBidKey(keys.BidStateActivePrefix, id)
	oKey := keys.MustBidKey(keys.BidStateOpenPrefix, id)
	lKey := keys.MustBidKey(keys.BidStateLostPrefix, id)
	cKey := keys.MustBidKey(keys.BidStateClosedPrefix, id)

	var key []byte

	// nolint: gocritic
	if store.Has(aKey) {
		key = aKey
	} else if store.Has(oKey) {
		key = oKey
	} else if store.Has(lKey) {
		key = lKey
	} else if store.Has(cKey) {
		key = cKey
	}

	return key
}

// GetBid returns bid with given bidID from market store
func (k Keeper) GetBid(ctx sdk.Context, id mv1.BidID) (types.Bid, bool) {
	store := ctx.KVStore(k.skey)

	key := k.findBid(ctx, id)

	if len(key) == 0 {
		return types.Bid{}, false
	}

	buf := store.Get(key)

	var val types.Bid
	k.cdc.MustUnmarshal(buf, &val)

	return val, true
}

func (k Keeper) findLease(ctx sdk.Context, id mv1.LeaseID) []byte {
	store := ctx.KVStore(k.skey)

	aKey := keys.MustLeaseKey(keys.LeaseStateActivePrefix, id)
	iKey := keys.MustLeaseKey(keys.LeaseStateInsufficientFundsPrefix, id)
	cKey := keys.MustLeaseKey(keys.LeaseStateClosedPrefix, id)

	var key []byte

	// nolint: gocritic
	if store.Has(aKey) {
		key = aKey
	} else if store.Has(iKey) {
		key = iKey
	} else if store.Has(cKey) {
		key = cKey
	}

	return key
}

// GetLease returns lease with given leaseID from market store
func (k Keeper) GetLease(ctx sdk.Context, id mv1.LeaseID) (mv1.Lease, bool) {
	store := ctx.KVStore(k.skey)
	key := k.findLease(ctx, id)

	if len(key) == 0 {
		return mv1.Lease{}, false
	}

	buf := store.Get(key)

	var val mv1.Lease
	k.cdc.MustUnmarshal(buf, &val)

	return val, true
}

// LeaseForOrder returns lease for order with given ID and lease found status
func (k Keeper) LeaseForOrder(ctx sdk.Context, bs types.Bid_State, oid mv1.OrderID) (mv1.Lease, bool) {
	var value mv1.Lease
	var found bool

	k.WithBidsForOrder(ctx, oid, bs, func(item types.Bid) bool {
		value, found = k.GetLease(ctx, mv1.LeaseID(item.ID))
		return true
	})

	return value, found
}

// WithOrders iterates all orders in market
func (k Keeper) WithOrders(ctx sdk.Context, fn func(types.Order) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, keys.OrderPrefix)
	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var val types.Order
		k.cdc.MustUnmarshal(iter.Value(), &val)
		if stop := fn(val); stop {
			break
		}
	}
}

// WithBids iterates all bids in market
func (k Keeper) WithBids(ctx sdk.Context, fn func(types.Bid) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, keys.BidPrefix)

	defer func() {
		_ = iter.Close()
	}()

	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var val types.Bid
		k.cdc.MustUnmarshal(iter.Value(), &val)
		if stop := fn(val); stop {
			break
		}
	}
}

// WithLeases iterates all leases in market
func (k Keeper) WithLeases(ctx sdk.Context, fn func(mv1.Lease) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, keys.LeasePrefix)

	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var val mv1.Lease
		k.cdc.MustUnmarshal(iter.Value(), &val)
		if stop := fn(val); stop {
			break
		}
	}
}

// WithOrdersForGroup iterates all orders of a group in market with given GroupID
func (k Keeper) WithOrdersForGroup(ctx sdk.Context, id dtypes.GroupID, state types.Order_State, fn func(types.Order) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, keys.OrdersForGroupPrefix(keys.OrderStateToPrefix(state), id))

	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var val types.Order
		k.cdc.MustUnmarshal(iter.Value(), &val)
		if stop := fn(val); stop {
			break
		}
	}
}

// WithBidsForOrder iterates all bids of an order in market with given OrderID
func (k Keeper) WithBidsForOrder(ctx sdk.Context, id mv1.OrderID, state types.Bid_State, fn func(types.Bid) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, keys.BidsForOrderPrefix(keys.BidStateToPrefix(state), id))

	defer func() {
		_ = iter.Close()
	}()

	for ; iter.Valid(); iter.Next() {
		var val types.Bid
		k.cdc.MustUnmarshal(iter.Value(), &val)
		if stop := fn(val); stop {
			break
		}
	}
}

func (k Keeper) BidCountForOrder(ctx sdk.Context, id mv1.OrderID) uint32 {
	store := ctx.KVStore(k.skey)
	oiter := storetypes.KVStorePrefixIterator(store, keys.BidsForOrderPrefix(keys.BidStateOpenPrefix, id))
	aiter := storetypes.KVStorePrefixIterator(store, keys.BidsForOrderPrefix(keys.BidStateActivePrefix, id))
	citer := storetypes.KVStorePrefixIterator(store, keys.BidsForOrderPrefix(keys.BidStateClosedPrefix, id))

	defer func() {
		_ = oiter.Close()
		_ = aiter.Close()
		_ = citer.Close()
	}()

	count := uint32(0)
	for ; oiter.Valid(); oiter.Next() {
		count++
	}

	for ; aiter.Valid(); aiter.Next() {
		count++
	}

	for ; citer.Valid(); citer.Next() {
		count++
	}

	return count
}

func (k Keeper) updateOrder(ctx sdk.Context, order types.Order, currState types.Order_State) {
	store := ctx.KVStore(k.skey)

	switch currState {
	case types.OrderOpen:
	case types.OrderActive:
	default:
		panic(fmt.Sprintf("unexpected current state of the order: %d", currState))
	}

	key := keys.MustOrderKey(keys.OrderStateToPrefix(currState), order.ID)
	store.Delete(key)

	switch order.State {
	case types.OrderActive:
	case types.OrderClosed:
	default:
		panic(fmt.Sprintf("unexpected new state of the order: %d", order.State))
	}

	data := k.cdc.MustMarshal(&order)

	key = keys.MustOrderKey(keys.OrderStateToPrefix(order.State), order.ID)
	store.Set(key, data)
}

func (k Keeper) updateBid(ctx sdk.Context, bid types.Bid, currState types.Bid_State) {
	store := ctx.KVStore(k.skey)

	switch currState {
	case types.BidOpen:
	case types.BidActive:
	default:
		panic(fmt.Sprintf("unexpected current state of the bid: %d", currState))
	}

	key := keys.MustBidKey(keys.BidStateToPrefix(currState), bid.ID)
	revKey := keys.MustBidStateRevereKey(currState, bid.ID)
	store.Delete(key)
	if revKey != nil {
		store.Delete(revKey)
	}

	switch bid.State {
	case types.BidActive:
	case types.BidLost:
	case types.BidClosed:
	default:
		panic(fmt.Sprintf("unexpected new state of the bid: %d", bid.State))
	}

	data := k.cdc.MustMarshal(&bid)

	key = keys.MustBidKey(keys.BidStateToPrefix(bid.State), bid.ID)
	revKey = keys.MustBidStateRevereKey(bid.State, bid.ID)

	store.Set(key, data)
	if len(revKey) > 0 {
		store.Set(revKey, data)
	}
}

// GetOrderByID returns an order by its string ID.
// The ID format is "customer/sequence" (e.g., "virtengine1abc.../123").
func (k Keeper) GetOrderByID(ctx sdk.Context, orderID string) (interface{}, bool) {
	// Try to parse the order ID using the marketplace parser
	parsed, err := marketplace.ParseOrderID(orderID)
	if err == nil {
		// Check for overflow when converting uint64 to uint32
		if parsed.Sequence > math.MaxUint32 {
			// Sequence too large for mv1.OrderID format, fall through to linear search
		} else {
			// Convert marketplace.OrderID to mv1.OrderID
			id := mv1.OrderID{
				Owner: parsed.CustomerAddress,
				OSeq:  uint32(parsed.Sequence),
			}
			order, found := k.GetOrder(ctx, id)
			if found {
				return order, true
			}
		}
	}

	// Fallback: search through all orders if parsing failed or order not found
	var foundOrder types.Order
	found := false
	k.WithOrders(ctx, func(order types.Order) bool {
		if order.ID.String() == orderID {
			foundOrder = order
			found = true
			return true // stop iteration
		}
		return false
	})
	if found {
		return foundOrder, true
	}
	return nil, false
}

// IsOrderCompleted checks if an order is in a completed state
func (k Keeper) IsOrderCompleted(ctx sdk.Context, orderID string) bool {
	orderIface, found := k.GetOrderByID(ctx, orderID)
	if !found {
		return false
	}
	order, ok := orderIface.(types.Order)
	if !ok {
		return false
	}
	return order.State == types.OrderClosed
}

// GetOrderCustomer returns the customer address for an order
func (k Keeper) GetOrderCustomer(ctx sdk.Context, orderID string) string {
	orderIface, found := k.GetOrderByID(ctx, orderID)
	if !found {
		return ""
	}
	order, ok := orderIface.(types.Order)
	if !ok {
		return ""
	}
	return order.ID.Owner
}

// GetOrderProvider returns the provider address for an order
// Note: This requires looking up the lease for the order
func (k Keeper) GetOrderProvider(ctx sdk.Context, orderID string) string {
	orderIface, found := k.GetOrderByID(ctx, orderID)
	if !found {
		return ""
	}
	order, ok := orderIface.(types.Order)
	if !ok {
		return ""
	}
	// Find the active lease for this order
	lease, found := k.LeaseForOrder(ctx, types.BidActive, order.ID)
	if !found {
		return ""
	}
	return lease.ID.Provider
}

// GetOrderCompletedAt returns when the order was completed.
// Note: Orders don't track completion timestamp directly. For closed orders,
// this returns the current block time as an approximation. For accurate
// completion timestamps, callers should monitor order state change events.
func (k Keeper) GetOrderCompletedAt(ctx sdk.Context, orderID string) time.Time {
	orderIface, found := k.GetOrderByID(ctx, orderID)
	if !found {
		return time.Time{}
	}
	order, ok := orderIface.(types.Order)
	if !ok {
		return time.Time{}
	}
	if order.State == types.OrderClosed {
		// Orders don't have a ClosedAt field. Return current block time.
		// Callers needing accurate completion time should track state change events.
		return ctx.BlockTime()
	}
	return time.Time{}
}

// GetOrderHash returns a hash of the order for verification
func (k Keeper) GetOrderHash(ctx sdk.Context, orderID string) string {
	orderIface, found := k.GetOrderByID(ctx, orderID)
	if !found {
		return ""
	}
	order, ok := orderIface.(types.Order)
	if !ok {
		return ""
	}
	// Hash the order ID as a simple implementation
	hash := sha256.Sum256([]byte(order.ID.String()))
	return hex.EncodeToString(hash[:])
}

// ProviderHasActiveLeases checks if a provider has any active leases.
// This is used to prevent provider deletion while leases are active.
// Uses the reverse index (provider → leases) for O(1) lookup instead of full scan.
func (k Keeper) ProviderHasActiveLeases(ctx sdk.Context, provider sdk.AccAddress) bool {
	store := ctx.KVStore(k.skey)

	// Use the reverse index to efficiently check for active leases by provider
	prefix := keys.LeasesByProviderPrefix(keys.LeaseStateActivePrefix, provider)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer func() {
		_ = iter.Close()
	}()

	// If there's at least one entry, provider has active leases
	return iter.Valid()
}
