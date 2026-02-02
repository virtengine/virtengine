// Package keeper provides the keeper for the marketplace module.
//
// VE-300 to VE-304: Marketplace on-chain module
// This file implements the marketplace keeper for managing offerings, orders,
// allocations, identity gating, MFA gating, and Waldur bridge operations.
package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// VEIDKeeper defines the interface for the VEID keeper
type VEIDKeeper interface {
	// GetIdentityScore returns the identity score for an account
	GetIdentityScore(ctx sdk.Context, address sdk.AccAddress) (uint32, bool)
	// GetIdentityStatus returns the identity status for an account
	GetIdentityStatus(ctx sdk.Context, address sdk.AccAddress) (string, bool)
	// IsEmailVerified returns whether the account's email is verified
	IsEmailVerified(ctx sdk.Context, address sdk.AccAddress) bool
	// IsDomainVerified returns whether the account's domain is verified
	IsDomainVerified(ctx sdk.Context, address sdk.AccAddress) bool
}

// MFAKeeper defines the interface for the MFA keeper
type MFAKeeper interface {
	// HasActiveFactors returns whether the account has active MFA factors
	HasActiveFactors(ctx sdk.Context, address sdk.AccAddress) bool
	// GetLastMFAVerification returns the last MFA verification time
	GetLastMFAVerification(ctx sdk.Context, address sdk.AccAddress) (*time.Time, bool)
	// IsTrustedDevice returns whether the device is trusted
	IsTrustedDevice(ctx sdk.Context, address sdk.AccAddress, fingerprint string) bool
	// CreateChallenge creates a new MFA challenge
	CreateChallenge(ctx sdk.Context, address sdk.AccAddress, actionType string) (string, error)
	// VerifyChallenge verifies an MFA challenge
	VerifyChallenge(ctx sdk.Context, challengeID string, response interface{}) (bool, error)
}

// ProviderKeeper defines the interface for the provider keeper
type ProviderKeeper interface {
	// IsProvider returns whether the account is a registered provider
	IsProvider(ctx sdk.Context, address sdk.AccAddress) bool
	// GetProvider returns provider information
	GetProvider(ctx sdk.Context, address sdk.AccAddress) (interface{}, bool)
}

// IKeeper defines the interface for the marketplace keeper
type IKeeper interface {
	// Params
	GetParams(ctx sdk.Context) marketplace.Params
	SetParams(ctx sdk.Context, params marketplace.Params) error

	// Offerings
	CreateOffering(ctx sdk.Context, offering *marketplace.Offering) error
	GetOffering(ctx sdk.Context, id marketplace.OfferingID) (*marketplace.Offering, bool)
	UpdateOffering(ctx sdk.Context, offering *marketplace.Offering) error
	TerminateOffering(ctx sdk.Context, id marketplace.OfferingID, reason string) error
	WithOfferings(ctx sdk.Context, fn func(marketplace.Offering) bool)
	GetOfferingsByProvider(ctx sdk.Context, providerAddress string) []marketplace.Offering

	// Orders
	CreateOrder(ctx sdk.Context, order *marketplace.Order) error
	GetOrder(ctx sdk.Context, id marketplace.OrderID) (*marketplace.Order, bool)
	UpdateOrder(ctx sdk.Context, order *marketplace.Order) error
	WithOrders(ctx sdk.Context, fn func(marketplace.Order) bool)
	GetOrdersByCustomer(ctx sdk.Context, customerAddress string) []marketplace.Order
	GetOrdersByOffering(ctx sdk.Context, offeringID marketplace.OfferingID) []marketplace.Order

	// Bids
	CreateBid(ctx sdk.Context, bid *marketplace.MarketplaceBid) error
	GetBid(ctx sdk.Context, id marketplace.BidID) (*marketplace.MarketplaceBid, bool)
	AcceptBid(ctx sdk.Context, id marketplace.BidID) (*marketplace.Allocation, error)
	WithBidsForOrder(ctx sdk.Context, orderID marketplace.OrderID, fn func(marketplace.MarketplaceBid) bool)
	WithBids(ctx sdk.Context, fn func(marketplace.MarketplaceBid) bool)

	// Allocations
	CreateAllocation(ctx sdk.Context, allocation *marketplace.Allocation) error
	GetAllocation(ctx sdk.Context, id marketplace.AllocationID) (*marketplace.Allocation, bool)
	UpdateAllocation(ctx sdk.Context, allocation *marketplace.Allocation) error
	WithAllocations(ctx sdk.Context, fn func(marketplace.Allocation) bool)

	// Identity Gating (VE-301)
	CheckIdentityGating(ctx sdk.Context, offering *marketplace.Offering, customerAddress sdk.AccAddress) error
	GetProviderIdentitySettings(ctx sdk.Context, providerAddress string) (*marketplace.ProviderIdentitySettings, bool)
	SetProviderIdentitySettings(ctx sdk.Context, providerAddress string, settings *marketplace.ProviderIdentitySettings) error
	WithProviderSettings(ctx sdk.Context, fn func(address string, settings marketplace.ProviderIdentitySettings) bool)

	// MFA Gating (VE-302)
	CheckMFAGating(ctx sdk.Context, actionType marketplace.MarketplaceActionType, accountAddress sdk.AccAddress, value uint64, deviceFingerprint string) (*marketplace.MFAGatingResult, error)
	RecordMFAAudit(ctx sdk.Context, record *marketplace.MFAAuditRecord) error
	GetMFAActionConfig(ctx sdk.Context, actionType marketplace.MarketplaceActionType) (*marketplace.MFAActionConfig, bool)
	SetMFAActionConfig(ctx sdk.Context, config *marketplace.MFAActionConfig) error
	WithMFAActionConfigs(ctx sdk.Context, fn func(config marketplace.MFAActionConfig) bool)

	// Waldur Bridge (VE-303)
	GetWaldurSyncRecord(ctx sdk.Context, entityType marketplace.WaldurSyncType, entityID string) (*marketplace.WaldurSyncRecord, bool)
	SetWaldurSyncRecord(ctx sdk.Context, record *marketplace.WaldurSyncRecord) error
	ProcessWaldurCallback(ctx sdk.Context, callback *marketplace.WaldurCallback) error
	IsNonceProcessed(ctx sdk.Context, nonce string) bool
	MarkNonceProcessed(ctx sdk.Context, nonce string) error

	// Events (VE-304)
	EmitMarketplaceEvent(ctx sdk.Context, event marketplace.MarketplaceEvent) error
	GetEventSequence(ctx sdk.Context) uint64
	IncrementEventSequence(ctx sdk.Context) uint64
	GetEventCheckpoint(ctx sdk.Context, subscriberID string) (*marketplace.EventCheckpoint, bool)
	SetEventCheckpoint(ctx sdk.Context, checkpoint *marketplace.EventCheckpoint) error
	RequestUsageUpdate(ctx sdk.Context, allocationID marketplace.AllocationID, requestType string) error

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
}

// Keeper implements the marketplace keeper
type Keeper struct {
	skey storetypes.StoreKey
	cdc  codec.BinaryCodec

	// Authority is the address capable of executing governance operations
	authority string

	// External keepers
	veidKeeper     VEIDKeeper
	mfaKeeper      MFAKeeper
	providerKeeper ProviderKeeper
}

// NewKeeper creates a new marketplace keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	authority string,
	veidKeeper VEIDKeeper,
	mfaKeeper MFAKeeper,
	providerKeeper ProviderKeeper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		skey:           skey,
		authority:      authority,
		veidKeeper:     veidKeeper,
		mfaKeeper:      mfaKeeper,
		providerKeeper: providerKeeper,
	}
}

// Codec returns keeper codec
func (k Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns store key
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.skey
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ============================================================================
// Parameters
// ============================================================================

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params marketplace.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(marketplace.ParamsKey(), bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) marketplace.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(marketplace.ParamsKey())
	if bz == nil {
		return marketplace.DefaultParams()
	}

	var params marketplace.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return marketplace.DefaultParams()
	}
	return params
}

// ============================================================================

// CreateOffering creates a new offering
func (k Keeper) CreateOffering(ctx sdk.Context, offering *marketplace.Offering) error {
	if err := offering.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := marketplace.OfferingKey(offering.ID)

	if store.Has(key) {
		return marketplace.ErrOfferingExists
	}

	bz, err := json.Marshal(offering)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Emit event
	event := &marketplace.BaseMarketplaceEvent{
		EventType:   marketplace.EventOfferingCreated,
		EventID:     fmt.Sprintf("evt_offering_created_%s_%d", offering.ID.String(), k.IncrementEventSequence(ctx)),
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().UTC(),
		Sequence:    k.GetEventSequence(ctx),
	}
	return k.EmitMarketplaceEvent(ctx, event)
}

// GetOffering returns an offering by ID
func (k Keeper) GetOffering(ctx sdk.Context, id marketplace.OfferingID) (*marketplace.Offering, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.OfferingKey(id)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var offering marketplace.Offering
	if err := json.Unmarshal(bz, &offering); err != nil {
		return nil, false
	}
	return &offering, true
}

// UpdateOffering updates an offering
func (k Keeper) UpdateOffering(ctx sdk.Context, offering *marketplace.Offering) error {
	if err := offering.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := marketplace.OfferingKey(offering.ID)

	if !store.Has(key) {
		return marketplace.ErrOfferingNotFound
	}

	offering.UpdatedAt = ctx.BlockTime().UTC()
	bz, err := json.Marshal(offering)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	seq := k.IncrementEventSequence(ctx)
	event := &marketplace.BaseMarketplaceEvent{
		EventType:   marketplace.EventOfferingUpdated,
		EventID:     fmt.Sprintf("evt_offering_updated_%s_%d", offering.ID.String(), seq),
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().UTC(),
		Sequence:    seq,
	}
	return k.EmitMarketplaceEvent(ctx, event)
}

// TerminateOffering terminates an offering
func (k Keeper) TerminateOffering(ctx sdk.Context, id marketplace.OfferingID, reason string) error {
	offering, found := k.GetOffering(ctx, id)
	if !found {
		return marketplace.ErrOfferingNotFound
	}

	now := ctx.BlockTime().UTC()
	offering.State = marketplace.OfferingStateTerminated
	offering.TerminatedAt = &now
	offering.UpdatedAt = now

	if err := k.UpdateOffering(ctx, offering); err != nil {
		return err
	}

	seq := k.IncrementEventSequence(ctx)
	event := &marketplace.BaseMarketplaceEvent{
		EventType:   marketplace.EventOfferingTerminated,
		EventID:     fmt.Sprintf("evt_offering_terminated_%s_%d", offering.ID.String(), seq),
		BlockHeight: ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().UTC(),
		Sequence:    seq,
	}
	return k.EmitMarketplaceEvent(ctx, event)
}

// WithOfferings iterates over all offerings
func (k Keeper) WithOfferings(ctx sdk.Context, fn func(marketplace.Offering) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.OfferingKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var offering marketplace.Offering
		if err := json.Unmarshal(iter.Value(), &offering); err != nil {
			continue
		}
		if fn(offering) {
			break
		}
	}
}

// GetOfferingsByProvider returns offerings for a provider
func (k Keeper) GetOfferingsByProvider(ctx sdk.Context, providerAddress string) []marketplace.Offering {
	var result []marketplace.Offering
	k.WithOfferings(ctx, func(offering marketplace.Offering) bool {
		if offering.ID.ProviderAddress == providerAddress {
			result = append(result, offering)
		}
		return false
	})
	return result
}

// ============================================================================
// Orders
// ============================================================================

// CreateOrder creates a new order with identity and MFA gating checks
func (k Keeper) CreateOrder(ctx sdk.Context, order *marketplace.Order) error {
	if err := order.Validate(); err != nil {
		return err
	}

	// Get the offering
	offering, found := k.GetOffering(ctx, order.OfferingID)
	if !found {
		return marketplace.ErrOfferingNotFound
	}

	// Check if offering can accept orders
	if err := offering.CanAcceptOrder(); err != nil {
		return err
	}

	// Parse customer address
	customerAddr, err := sdk.AccAddressFromBech32(order.ID.CustomerAddress)
	if err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	params := k.GetParams(ctx)

	// VE-301: Identity gating check
	if params.EnableIdentityGating {
		if err := k.CheckIdentityGating(ctx, offering, customerAddr); err != nil {
			return err
		}
	}

	// Store the order
	store := ctx.KVStore(k.skey)
	key := marketplace.OrderKey(order.ID)

	if store.Has(key) {
		return marketplace.ErrOrderExists
	}

	bz, err := json.Marshal(order)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Update offering order count
	offering.TotalOrderCount++
	if order.State.IsActive() {
		offering.ActiveOrderCount++
	}
	if err := k.UpdateOffering(ctx, offering); err != nil {
		return err
	}

	// Emit event
	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewOrderCreatedEventAt(order, ctx.BlockHeight(), seq, ctx.BlockTime())
	return k.EmitMarketplaceEvent(ctx, event)
}

// GetOrder returns an order by ID
func (k Keeper) GetOrder(ctx sdk.Context, id marketplace.OrderID) (*marketplace.Order, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.OrderKey(id)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var order marketplace.Order
	if err := json.Unmarshal(bz, &order); err != nil {
		return nil, false
	}
	return &order, true
}

// UpdateOrder updates an order
func (k Keeper) UpdateOrder(ctx sdk.Context, order *marketplace.Order) error {
	if err := order.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := marketplace.OrderKey(order.ID)

	if !store.Has(key) {
		return marketplace.ErrOrderNotFound
	}

	order.UpdatedAt = ctx.BlockTime().UTC()
	bz, err := json.Marshal(order)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// WithOrders iterates over all orders
func (k Keeper) WithOrders(ctx sdk.Context, fn func(marketplace.Order) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.OrderKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var order marketplace.Order
		if err := json.Unmarshal(iter.Value(), &order); err != nil {
			continue
		}
		if fn(order) {
			break
		}
	}
}

// GetOrdersByCustomer returns orders for a customer
func (k Keeper) GetOrdersByCustomer(ctx sdk.Context, customerAddress string) []marketplace.Order {
	var result []marketplace.Order
	k.WithOrders(ctx, func(order marketplace.Order) bool {
		if order.ID.CustomerAddress == customerAddress {
			result = append(result, order)
		}
		return false
	})
	return result
}

// GetOrdersByOffering returns orders for an offering
func (k Keeper) GetOrdersByOffering(ctx sdk.Context, offeringID marketplace.OfferingID) []marketplace.Order {
	var result []marketplace.Order
	k.WithOrders(ctx, func(order marketplace.Order) bool {
		if order.OfferingID == offeringID {
			result = append(result, order)
		}
		return false
	})
	return result
}

// ============================================================================
// Bids
// ============================================================================

// CreateBid creates a new bid
func (k Keeper) CreateBid(ctx sdk.Context, bid *marketplace.MarketplaceBid) error {
	if err := bid.ID.Validate(); err != nil {
		return err
	}

	// Get the order
	order, found := k.GetOrder(ctx, bid.ID.OrderID)
	if !found {
		return marketplace.ErrOrderNotFound
	}

	// Check if order can accept bids
	if err := order.CanAcceptBidAt(ctx.BlockTime()); err != nil {
		return err
	}

	// Check bid price against order max
	if bid.Price > order.MaxBidPrice {
		return marketplace.ErrBidPriceTooHigh
	}

	store := ctx.KVStore(k.skey)
	key := marketplace.BidKey(bid.ID)

	if store.Has(key) {
		return marketplace.ErrBidExists
	}

	bid.CreatedAt = ctx.BlockTime().UTC()
	bid.UpdatedAt = bid.CreatedAt
	bid.State = marketplace.BidStateOpen

	bz, err := json.Marshal(bid)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	// Update order bid count
	order.BidCount++
	if err := k.UpdateOrder(ctx, order); err != nil {
		return err
	}

	// Emit event
	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewBidPlacedEventAt(bid, ctx.BlockHeight(), seq, ctx.BlockTime())
	return k.EmitMarketplaceEvent(ctx, event)
}

// GetBid returns a bid by ID
func (k Keeper) GetBid(ctx sdk.Context, id marketplace.BidID) (*marketplace.MarketplaceBid, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.BidKey(id)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var bid marketplace.MarketplaceBid
	if err := json.Unmarshal(bz, &bid); err != nil {
		return nil, false
	}
	return &bid, true
}

// AcceptBid accepts a bid and creates an allocation
func (k Keeper) AcceptBid(ctx sdk.Context, id marketplace.BidID) (*marketplace.Allocation, error) {
	bid, found := k.GetBid(ctx, id)
	if !found {
		return nil, marketplace.ErrBidNotFound
	}

	if bid.State != marketplace.BidStateOpen {
		return nil, marketplace.ErrBidNotOpen
	}

	order, found := k.GetOrder(ctx, bid.ID.OrderID)
	if !found {
		return nil, marketplace.ErrOrderNotFound
	}

	// Create allocation
	allocationID := marketplace.AllocationID{
		OrderID:  order.ID,
		Sequence: 1, // First allocation for this order
	}

	allocation := marketplace.NewAllocationAt(
		allocationID,
		order.OfferingID,
		bid.ID.ProviderAddress,
		bid.ID,
		bid.Price,
		ctx.BlockTime(),
	)

	if err := k.CreateAllocation(ctx, allocation); err != nil {
		return nil, err
	}

	// Update bid state
	bid.State = marketplace.BidStateAccepted
	bid.UpdatedAt = ctx.BlockTime().UTC()
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(bid)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bid: %w", err)
	}
	store.Set(marketplace.BidKey(bid.ID), bz)

	// Update order state
	if err := order.SetStateAt(marketplace.OrderStateMatched, "bid accepted", ctx.BlockTime()); err != nil {
		return nil, err
	}
	order.AllocatedProviderAddress = bid.ID.ProviderAddress
	order.AcceptedPrice = bid.Price
	if err := k.UpdateOrder(ctx, order); err != nil {
		return nil, err
	}

	// Emit allocation created event
	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewAllocationCreatedEvent(allocation, order.ID.CustomerAddress, ctx.BlockHeight(), seq)
	if err := k.EmitMarketplaceEvent(ctx, event); err != nil {
		return nil, err
	}

	return allocation, nil
}

// WithBidsForOrder iterates over bids for an order
func (k Keeper) WithBidsForOrder(ctx sdk.Context, orderID marketplace.OrderID, fn func(marketplace.MarketplaceBid) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.BidKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var bid marketplace.MarketplaceBid
		if err := json.Unmarshal(iter.Value(), &bid); err != nil {
			continue
		}
		if bid.ID.OrderID == orderID {
			if fn(bid) {
				break
			}
		}
	}
}

// ============================================================================
// Allocations
// ============================================================================

// CreateAllocation creates a new allocation
func (k Keeper) CreateAllocation(ctx sdk.Context, allocation *marketplace.Allocation) error {
	if err := allocation.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := marketplace.AllocationKey(allocation.ID)

	if store.Has(key) {
		return marketplace.ErrAllocationExists
	}

	bz, err := json.Marshal(allocation)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetAllocation returns an allocation by ID
func (k Keeper) GetAllocation(ctx sdk.Context, id marketplace.AllocationID) (*marketplace.Allocation, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.AllocationKey(id)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var allocation marketplace.Allocation
	if err := json.Unmarshal(bz, &allocation); err != nil {
		return nil, false
	}
	return &allocation, true
}

// UpdateAllocation updates an allocation
func (k Keeper) UpdateAllocation(ctx sdk.Context, allocation *marketplace.Allocation) error {
	if err := allocation.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := marketplace.AllocationKey(allocation.ID)

	if !store.Has(key) {
		return marketplace.ErrAllocationNotFound
	}

	allocation.UpdatedAt = ctx.BlockTime().UTC()
	bz, err := json.Marshal(allocation)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// WithAllocations iterates over all allocations
func (k Keeper) WithAllocations(ctx sdk.Context, fn func(marketplace.Allocation) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.AllocationKeyPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var allocation marketplace.Allocation
		if err := json.Unmarshal(iter.Value(), &allocation); err != nil {
			continue
		}
		if fn(allocation) {
			break
		}
	}
}

// ============================================================================
// Identity Gating (VE-301)
// ============================================================================

// CheckIdentityGating checks identity requirements for an order
func (k Keeper) CheckIdentityGating(ctx sdk.Context, offering *marketplace.Offering, customerAddress sdk.AccAddress) error {
	// Get customer identity info
	score, _ := k.veidKeeper.GetIdentityScore(ctx, customerAddress)
	status, _ := k.veidKeeper.GetIdentityStatus(ctx, customerAddress)
	emailVerified := k.veidKeeper.IsEmailVerified(ctx, customerAddress)
	domainVerified := k.veidKeeper.IsDomainVerified(ctx, customerAddress)
	mfaEnabled := k.mfaKeeper.HasActiveFactors(ctx, customerAddress)

	customerInfo := &marketplace.CustomerIdentityInfo{
		Score:          score,
		Status:         status,
		EmailVerified:  emailVerified,
		DomainVerified: domainVerified,
		MFAEnabled:     mfaEnabled,
	}

	// Get provider settings
	var providerSettings *marketplace.ProviderIdentitySettings
	if settings, found := k.GetProviderIdentitySettings(ctx, offering.ID.ProviderAddress); found {
		providerSettings = settings
	}

	// Validate
	return marketplace.ValidateOrderCreation(offering, customerInfo, providerSettings)
}

// GetProviderIdentitySettings returns provider identity settings
func (k Keeper) GetProviderIdentitySettings(ctx sdk.Context, providerAddress string) (*marketplace.ProviderIdentitySettings, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.ProviderSettingsKey(providerAddress)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var settings marketplace.ProviderIdentitySettings
	if err := json.Unmarshal(bz, &settings); err != nil {
		return nil, false
	}
	return &settings, true
}

// SetProviderIdentitySettings sets provider identity settings
func (k Keeper) SetProviderIdentitySettings(ctx sdk.Context, providerAddress string, settings *marketplace.ProviderIdentitySettings) error {
	store := ctx.KVStore(k.skey)
	key := marketplace.ProviderSettingsKey(providerAddress)

	bz, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// ============================================================================
// MFA Gating (VE-302)
// ============================================================================

// CheckMFAGating checks MFA requirements for an action
func (k Keeper) CheckMFAGating(ctx sdk.Context, actionType marketplace.MarketplaceActionType, accountAddress sdk.AccAddress, value uint64, deviceFingerprint string) (*marketplace.MFAGatingResult, error) {
	params := k.GetParams(ctx)
	if !params.EnableMFAGating {
		return &marketplace.MFAGatingResult{
			Required:  false,
			Satisfied: true,
			Reason:    "MFA gating disabled",
		}, nil
	}

	// Build MFA context
	lastMFA, _ := k.mfaKeeper.GetLastMFAVerification(ctx, accountAddress)
	isTrusted := k.mfaKeeper.IsTrustedDevice(ctx, accountAddress, deviceFingerprint)
	hasFactors := k.mfaKeeper.HasActiveFactors(ctx, accountAddress)

	mfaContext := &marketplace.MFAGatingContext{
		ActionType:        actionType,
		AccountAddress:    accountAddress.String(),
		TransactionValue:  value,
		IsTrustedDevice:   isTrusted,
		DeviceFingerprint: deviceFingerprint,
		LastMFAVerifiedAt: lastMFA,
		AccountMFAPolicy: &marketplace.AccountMFAPolicy{
			Enabled:             hasFactors,
			RequireForHighValue: true,
			HighValueThreshold:  1000000,
		},
	}

	checker := marketplace.NewMFAGatingChecker()

	// Load custom configs
	for _, config := range params.MFAConfigs {
		c := config
		checker.SetConfig(&c)
	}

	result := checker.Check(mfaContext)
	return result, nil
}

// RecordMFAAudit records an MFA audit entry
func (k Keeper) RecordMFAAudit(ctx sdk.Context, record *marketplace.MFAAuditRecord) error {
	store := ctx.KVStore(k.skey)
	key := marketplace.MFAAuditKey(record.ChallengeID)

	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetMFAActionConfig returns MFA action config
func (k Keeper) GetMFAActionConfig(ctx sdk.Context, actionType marketplace.MarketplaceActionType) (*marketplace.MFAActionConfig, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.MFAConfigKey(actionType)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var config marketplace.MFAActionConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return nil, false
	}
	return &config, true
}

// SetMFAActionConfig sets MFA action config
func (k Keeper) SetMFAActionConfig(ctx sdk.Context, config *marketplace.MFAActionConfig) error {
	store := ctx.KVStore(k.skey)
	key := marketplace.MFAConfigKey(config.ActionType)

	bz, err := json.Marshal(config)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// ============================================================================
// Waldur Bridge (VE-303)
// ============================================================================

// GetWaldurSyncRecord returns a Waldur sync record
func (k Keeper) GetWaldurSyncRecord(ctx sdk.Context, entityType marketplace.WaldurSyncType, entityID string) (*marketplace.WaldurSyncRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.SyncRecordKey(entityType, entityID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var record marketplace.WaldurSyncRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}
	return &record, true
}

// SetWaldurSyncRecord sets a Waldur sync record
func (k Keeper) SetWaldurSyncRecord(ctx sdk.Context, record *marketplace.WaldurSyncRecord) error {
	store := ctx.KVStore(k.skey)
	key := marketplace.SyncRecordKey(record.EntityType, record.EntityID)

	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// ProcessWaldurCallback processes a Waldur callback
func (k Keeper) ProcessWaldurCallback(ctx sdk.Context, callback *marketplace.WaldurCallback) error {
	// Validate callback
	if err := callback.ValidateAt(ctx.BlockTime()); err != nil {
		return marketplace.ErrWaldurCallbackInvalid.Wrap(err.Error())
	}

	// Check for replay
	if k.IsNonceProcessed(ctx, callback.Nonce) {
		return marketplace.ErrWaldurNonceReplayed
	}

	// Mark nonce as processed
	if err := k.MarkNonceProcessed(ctx, callback.Nonce); err != nil {
		return err
	}

	// Process based on action type
	switch callback.ActionType {
	case marketplace.ActionTypeProvision:
		return k.processProvisionCallback(ctx, callback)
	case marketplace.ActionTypeTerminate:
		return k.processTerminateCallback(ctx, callback)
	default:
		// Other action types can be added here
		return nil
	}
}

// processProvisionCallback handles provision callbacks
func (k Keeper) processProvisionCallback(ctx sdk.Context, callback *marketplace.WaldurCallback) error {
	if callback.ChainEntityType != marketplace.SyncTypeAllocation {
		return marketplace.ErrWaldurCallbackInvalid.Wrapf("unexpected entity type: %s", callback.ChainEntityType)
	}

	allocationID, err := marketplace.ParseAllocationID(callback.ChainEntityID)
	if err != nil {
		return marketplace.ErrWaldurCallbackInvalid.Wrap(err.Error())
	}

	allocation, found := k.GetAllocation(ctx, allocationID)
	if !found {
		return marketplace.ErrAllocationNotFound.Wrapf("allocation %s not found", callback.ChainEntityID)
	}

	if allocation.State.IsTerminal() {
		return marketplace.ErrInvalidStateTransition.Wrapf("allocation %s already terminal", allocationID.String())
	}

	order, found := k.GetOrder(ctx, allocationID.OrderID)
	if !found {
		return marketplace.ErrOrderNotFound.Wrapf("order %s not found", allocationID.OrderID.String())
	}

	reason := "provision requested by waldur"
	if value, ok := callback.Payload["reason"]; ok && value != "" {
		reason = value
	}

	if allocation.State != marketplace.AllocationStateProvisioning && allocation.State != marketplace.AllocationStateActive {
		if err := allocation.SetStateAt(marketplace.AllocationStateProvisioning, reason, ctx.BlockTime()); err != nil {
			return err
		}
		if err := k.UpdateAllocation(ctx, allocation); err != nil {
			return err
		}
	}

	if order.State != marketplace.OrderStateProvisioning && order.State != marketplace.OrderStateActive {
		if err := order.SetStateAt(marketplace.OrderStateProvisioning, reason, ctx.BlockTime()); err != nil {
			return marketplace.ErrInvalidStateTransition.Wrap(err.Error())
		}
		if err := k.UpdateOrder(ctx, order); err != nil {
			return err
		}
	}

	encryptedConfigRef := ""
	if value, ok := callback.Payload["encrypted_config_ref"]; ok && value != "" {
		encryptedConfigRef = value
	} else if order.EncryptedConfig != nil {
		encryptedConfigRef = order.EncryptedConfig.EnvelopeRef
	}

	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewProvisionRequestedEventAt(allocation, encryptedConfigRef, ctx.BlockHeight(), seq, ctx.BlockTime())
	return k.EmitMarketplaceEvent(ctx, event)
}

// processTerminateCallback handles terminate callbacks
func (k Keeper) processTerminateCallback(ctx sdk.Context, callback *marketplace.WaldurCallback) error {
	if callback.ChainEntityType != marketplace.SyncTypeAllocation {
		return marketplace.ErrWaldurCallbackInvalid.Wrapf("unexpected entity type: %s", callback.ChainEntityType)
	}

	allocationID, err := marketplace.ParseAllocationID(callback.ChainEntityID)
	if err != nil {
		return marketplace.ErrWaldurCallbackInvalid.Wrap(err.Error())
	}

	allocation, found := k.GetAllocation(ctx, allocationID)
	if !found {
		return marketplace.ErrAllocationNotFound.Wrapf("allocation %s not found", callback.ChainEntityID)
	}

	if allocation.State.IsTerminal() {
		return marketplace.ErrInvalidStateTransition.Wrapf("allocation %s already terminal", allocationID.String())
	}

	order, found := k.GetOrder(ctx, allocationID.OrderID)
	if !found {
		return marketplace.ErrOrderNotFound.Wrapf("order %s not found", allocationID.OrderID.String())
	}

	reason := "termination requested by waldur"
	if value, ok := callback.Payload["reason"]; ok && value != "" {
		reason = value
	}

	immediate := false
	if value, ok := callback.Payload["immediate"]; ok {
		immediate = value == "true" || value == "1"
	}

	targetAllocationState := marketplace.AllocationStateTerminating
	if immediate {
		targetAllocationState = marketplace.AllocationStateTerminated
	}

	if err := allocation.SetStateAt(targetAllocationState, reason, ctx.BlockTime()); err != nil {
		return err
	}
	if err := k.UpdateAllocation(ctx, allocation); err != nil {
		return err
	}

	switch {
	case immediate && order.State == marketplace.OrderStatePendingTermination:
		if err := order.SetStateAt(marketplace.OrderStateTerminated, reason, ctx.BlockTime()); err != nil {
			return marketplace.ErrInvalidStateTransition.Wrap(err.Error())
		}
	case order.State != marketplace.OrderStatePendingTermination && order.State != marketplace.OrderStateTerminated:
		if err := order.SetStateAt(marketplace.OrderStatePendingTermination, reason, ctx.BlockTime()); err != nil {
			return marketplace.ErrInvalidStateTransition.Wrap(err.Error())
		}
	}

	if err := k.UpdateOrder(ctx, order); err != nil {
		return err
	}

	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewTerminateRequestedEventAt(
		allocation.ID.String(),
		order.ID.String(),
		allocation.ProviderAddress,
		callback.SignerID,
		reason,
		immediate,
		ctx.BlockHeight(),
		seq,
		ctx.BlockTime(),
	)
	return k.EmitMarketplaceEvent(ctx, event)
}

// IsNonceProcessed checks if a nonce has been processed
func (k Keeper) IsNonceProcessed(ctx sdk.Context, nonce string) bool {
	store := ctx.KVStore(k.skey)
	key := marketplace.ProcessedNonceKey(nonce)
	return store.Has(key)
}

// MarkNonceProcessed marks a nonce as processed
func (k Keeper) MarkNonceProcessed(ctx sdk.Context, nonce string) error {
	store := ctx.KVStore(k.skey)
	key := marketplace.ProcessedNonceKey(nonce)

	// Store with expiry timestamp
	expiry := ctx.BlockTime().Add(2 * time.Hour)
	bz, err := json.Marshal(expiry)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// ============================================================================
// Events (VE-304)
// ============================================================================

// EmitMarketplaceEvent emits a marketplace event
func (k Keeper) EmitMarketplaceEvent(ctx sdk.Context, event marketplace.MarketplaceEvent) error {
	payloadJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	attributes := []sdk.Attribute{
		sdk.NewAttribute("event_type", string(event.GetEventType())),
		sdk.NewAttribute("event_id", event.GetEventID()),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", event.GetBlockHeight())),
		sdk.NewAttribute("sequence", fmt.Sprintf("%d", event.GetSequence())),
		sdk.NewAttribute("payload_json", string(payloadJSON)),
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("marketplace_event", attributes...),
	)

	// Emit SDK event for subscription (legacy typed event)
	return ctx.EventManager().EmitTypedEvent(&MarketplaceEventWrapper{
		EventType:   string(event.GetEventType()),
		EventID:     event.GetEventID(),
		BlockHeight: event.GetBlockHeight(),
		Sequence:    event.GetSequence(),
	})
}

// MarketplaceEventWrapper is a wrapper for typed event emission
type MarketplaceEventWrapper struct {
	EventType   string `json:"event_type"`
	EventID     string `json:"event_id"`
	BlockHeight int64  `json:"block_height"`
	Sequence    uint64 `json:"sequence"`
}

// Proto.Message interface stubs for MarketplaceEventWrapper
func (m *MarketplaceEventWrapper) ProtoMessage()  {}
func (m *MarketplaceEventWrapper) Reset()         { *m = MarketplaceEventWrapper{} }
func (m *MarketplaceEventWrapper) String() string { return fmt.Sprintf("%+v", *m) }

// GetEventSequence returns the current event sequence
func (k Keeper) GetEventSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(marketplace.EventSequenceKey())
	if bz == nil {
		return 0
	}

	var seq uint64
	if err := json.Unmarshal(bz, &seq); err != nil {
		return 0
	}
	return seq
}

// IncrementEventSequence increments and returns the event sequence
func (k Keeper) IncrementEventSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	seq := k.GetEventSequence(ctx) + 1

	bz, _ := json.Marshal(seq) //nolint:errchkjson // uint64 cannot fail to marshal
	store.Set(marketplace.EventSequenceKey(), bz)
	return seq
}

// GetEventCheckpoint returns an event checkpoint
func (k Keeper) GetEventCheckpoint(ctx sdk.Context, subscriberID string) (*marketplace.EventCheckpoint, bool) {
	store := ctx.KVStore(k.skey)
	key := marketplace.EventCheckpointKey(subscriberID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var checkpoint marketplace.EventCheckpoint
	if err := json.Unmarshal(bz, &checkpoint); err != nil {
		return nil, false
	}
	return &checkpoint, true
}

// SetEventCheckpoint sets an event checkpoint
func (k Keeper) SetEventCheckpoint(ctx sdk.Context, checkpoint *marketplace.EventCheckpoint) error {
	store := ctx.KVStore(k.skey)
	key := marketplace.EventCheckpointKey(checkpoint.SubscriberID)

	checkpoint.UpdatedAt = ctx.BlockTime().UTC()
	bz, err := json.Marshal(checkpoint)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// RequestUsageUpdate requests a usage update for an allocation and emits UsageUpdateRequested event
func (k Keeper) RequestUsageUpdate(ctx sdk.Context, allocationID marketplace.AllocationID, requestType string) error {
	allocation, found := k.GetAllocation(ctx, allocationID)
	if !found {
		return marketplace.ErrAllocationNotFound
	}

	if allocation.State.IsTerminal() {
		return marketplace.ErrInvalidStateTransition.Wrapf("allocation %s is in terminal state", allocationID.String())
	}

	seq := k.IncrementEventSequence(ctx)
	event := marketplace.NewUsageUpdateRequestedEventAt(
		allocation.ID.String(),
		allocation.ProviderAddress,
		requestType,
		ctx.BlockHeight(),
		seq,
		ctx.BlockTime(),
	)
	return k.EmitMarketplaceEvent(ctx, event)
}
