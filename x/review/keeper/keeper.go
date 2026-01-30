// Package keeper implements the Review module keeper.
//
// VE-911: Provider public reviews - Keeper implementation
// This keeper manages reviews with star ratings (1-5), verified order links,
// on-chain content hashes for integrity, and rating aggregation per provider.
package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/review/types"
)

// IKeeper defines the interface for the Review keeper
type IKeeper interface {
	// Reviews
	SubmitReview(ctx sdk.Context, review *types.Review) error
	GetReview(ctx sdk.Context, reviewID string) (types.Review, bool)
	GetReviewsByProvider(ctx sdk.Context, providerAddr string) []types.Review
	GetReviewsByReviewer(ctx sdk.Context, reviewerAddr string) []types.Review
	GetReviewByOrder(ctx sdk.Context, orderID string) (types.Review, bool)
	SetReview(ctx sdk.Context, review types.Review) error
	DeleteReview(ctx sdk.Context, reviewID string, moderatorAddr, reason string) error

	// Aggregations
	GetProviderAggregation(ctx sdk.Context, providerAddr string) (types.ProviderAggregation, bool)
	SetProviderAggregation(ctx sdk.Context, agg types.ProviderAggregation) error
	UpdateProviderAggregation(ctx sdk.Context, providerAddr string, rating uint8, isAdd bool) error
	GetTopProviders(ctx sdk.Context, limit int) []types.ProviderAggregation

	// Order verification
	VerifyOrderCompleted(ctx sdk.Context, orderID, customerAddr, providerAddr string) (*types.OrderReference, error)

	// Parameters
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Iterators
	WithReviews(ctx sdk.Context, fn func(types.Review) bool)
	WithProviderAggregations(ctx sdk.Context, fn func(types.ProviderAggregation) bool)

	// Genesis
	SetNextReviewSequence(ctx sdk.Context, seq uint64)
	GetNextReviewSequence(ctx sdk.Context) uint64

	// Codec and store
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
	Logger(ctx sdk.Context) log.Logger
}

// MarketKeeper defines the expected market keeper interface for order verification
type MarketKeeper interface {
	// GetOrder returns an order by ID
	GetOrderByID(ctx sdk.Context, orderID string) (interface{}, bool)
	// IsOrderCompleted checks if an order is in a completed state
	IsOrderCompleted(ctx sdk.Context, orderID string) bool
	// GetOrderCustomer returns the customer address for an order
	GetOrderCustomer(ctx sdk.Context, orderID string) string
	// GetOrderProvider returns the provider address for an order
	GetOrderProvider(ctx sdk.Context, orderID string) string
	// GetOrderCompletedAt returns when the order was completed
	GetOrderCompletedAt(ctx sdk.Context, orderID string) time.Time
	// GetOrderHash returns a hash of the order for verification
	GetOrderHash(ctx sdk.Context, orderID string) string
}

// RolesKeeper defines the expected roles keeper interface
type RolesKeeper interface {
	IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool
	IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool
}

// Keeper implements the Review module keeper
type Keeper struct {
	skey         storetypes.StoreKey
	cdc          codec.BinaryCodec
	marketKeeper MarketKeeper
	rolesKeeper  RolesKeeper
	authority    string
}

// NewKeeper creates and returns an instance for Review keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	skey storetypes.StoreKey,
	marketKeeper MarketKeeper,
	rolesKeeper RolesKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		skey:         skey,
		marketKeeper: marketKeeper,
		rolesKeeper:  rolesKeeper,
		authority:    authority,
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

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetNextReviewSequence sets the next review sequence number
func (k Keeper) SetNextReviewSequence(ctx sdk.Context, seq uint64) {
	store := ctx.KVStore(k.skey)
	store.Set(types.SequenceKeyReview, []byte(strconv.FormatUint(seq, 10)))
}

// GetNextReviewSequence returns and increments the next review sequence
func (k Keeper) GetNextReviewSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.SequenceKeyReview)
	if bz == nil {
		k.SetNextReviewSequence(ctx, 2)
		return 1
	}

	seq, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		k.SetNextReviewSequence(ctx, 2)
		return 1
	}

	k.SetNextReviewSequence(ctx, seq+1)
	return seq
}

// VerifyOrderCompleted verifies that an order exists, is completed, and belongs to the reviewer
func (k Keeper) VerifyOrderCompleted(ctx sdk.Context, orderID, customerAddr, providerAddr string) (*types.OrderReference, error) {
	// If no market keeper is set (testing), return a mock verification
	if k.marketKeeper == nil {
		return &types.OrderReference{
			OrderID:         orderID,
			CustomerAddress: customerAddr,
			ProviderAddress: providerAddr,
			// BUGFIX-001: Use ctx.BlockTime() for consensus safety, even in test mock
			CompletedAt:     ctx.BlockTime().UTC().Add(-24 * time.Hour),
			OrderHash:       fmt.Sprintf("mock-hash-%s", orderID),
		}, nil
	}

	// Check if order exists
	_, exists := k.marketKeeper.GetOrderByID(ctx, orderID)
	if !exists {
		return nil, types.ErrOrderNotFound.Wrapf("order %s not found", orderID)
	}

	// Check if order is completed
	if !k.marketKeeper.IsOrderCompleted(ctx, orderID) {
		return nil, types.ErrOrderNotCompleted.Wrapf("order %s is not completed", orderID)
	}

	// Verify the reviewer is the order customer
	orderCustomer := k.marketKeeper.GetOrderCustomer(ctx, orderID)
	if orderCustomer != customerAddr {
		return nil, types.ErrUnauthorizedReviewer.Wrapf(
			"reviewer %s is not the order customer %s", customerAddr, orderCustomer)
	}

	// Verify the provider matches
	orderProvider := k.marketKeeper.GetOrderProvider(ctx, orderID)
	if orderProvider != providerAddr {
		return nil, types.ErrProviderNotFound.Wrapf(
			"order provider %s does not match review provider %s", orderProvider, providerAddr)
	}

	return &types.OrderReference{
		OrderID:         orderID,
		CustomerAddress: customerAddr,
		ProviderAddress: providerAddr,
		CompletedAt:     k.marketKeeper.GetOrderCompletedAt(ctx, orderID),
		OrderHash:       k.marketKeeper.GetOrderHash(ctx, orderID),
	}, nil
}

// SubmitReview submits a new review for a provider
func (k Keeper) SubmitReview(ctx sdk.Context, review *types.Review) error {
	params := k.GetParams(ctx)

	// Validate review
	if err := review.Validate(); err != nil {
		return err
	}

	// Check if order was already reviewed
	if _, exists := k.GetReviewByOrder(ctx, review.OrderRef.OrderID); exists {
		return types.ErrDuplicateReview.Wrapf("order %s has already been reviewed", review.OrderRef.OrderID)
	}

	// Verify order is completed (if required by params)
	if params.RequireCompletedOrder {
		orderRef, err := k.VerifyOrderCompleted(ctx, review.OrderRef.OrderID,
			review.ReviewerAddress, review.ProviderAddress)
		if err != nil {
			return err
		}
		review.OrderRef = *orderRef
	}

	// Assign review ID with sequence
	seq := k.GetNextReviewSequence(ctx)
	review.ID = types.ReviewID{
		ProviderAddress: review.ProviderAddress,
		Sequence:        seq,
	}

	// Set block height
	review.BlockHeight = ctx.BlockHeight()

	// Compute and verify content hash
	review.ContentHash = review.ComputeContentHash()

	// Store the review
	if err := k.SetReview(ctx, *review); err != nil {
		return err
	}

	// Update provider aggregation
	if err := k.UpdateProviderAggregation(ctx, review.ProviderAddress, review.Rating, true); err != nil {
		return fmt.Errorf("failed to update aggregation: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeReviewSubmitted,
			sdk.NewAttribute(types.AttributeKeyReviewID, review.ID.String()),
			sdk.NewAttribute(types.AttributeKeyOrderID, review.OrderRef.OrderID),
			sdk.NewAttribute(types.AttributeKeyProviderAddress, review.ProviderAddress),
			sdk.NewAttribute(types.AttributeKeyReviewerAddress, review.ReviewerAddress),
			sdk.NewAttribute(types.AttributeKeyRating, strconv.Itoa(int(review.Rating))),
			sdk.NewAttribute(types.AttributeKeyContentHash, review.ContentHash),
		),
	})

	k.Logger(ctx).Info("review submitted",
		"review_id", review.ID.String(),
		"provider", review.ProviderAddress,
		"reviewer", review.ReviewerAddress,
		"rating", review.Rating,
	)

	return nil
}

// SetReview stores a review in the store
func (k Keeper) SetReview(ctx sdk.Context, review types.Review) error {
	store := ctx.KVStore(k.skey)

	bz, err := json.Marshal(review)
	if err != nil {
		return fmt.Errorf("failed to marshal review: %w", err)
	}

	// Store review by ID
	store.Set(types.GetReviewKey(review.ID.String()), bz)

	// Create indexes
	// Provider index
	k.addToIndex(store, types.GetProviderIndexKey(review.ProviderAddress), review.ID.String())

	// Reviewer index
	k.addToIndex(store, types.GetReviewerIndexKey(review.ReviewerAddress), review.ID.String())

	// Order index (one-to-one mapping)
	store.Set(types.GetOrderIndexKey(review.OrderRef.OrderID), []byte(review.ID.String()))

	return nil
}

// GetReview retrieves a review by ID
func (k Keeper) GetReview(ctx sdk.Context, reviewID string) (types.Review, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetReviewKey(reviewID))
	if bz == nil {
		return types.Review{}, false
	}

	var review types.Review
	if err := json.Unmarshal(bz, &review); err != nil {
		return types.Review{}, false
	}
	return review, true
}

// GetReviewByOrder retrieves a review by order ID
func (k Keeper) GetReviewByOrder(ctx sdk.Context, orderID string) (types.Review, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetOrderIndexKey(orderID))
	if bz == nil {
		return types.Review{}, false
	}

	reviewID := string(bz)
	return k.GetReview(ctx, reviewID)
}

// GetReviewsByProvider retrieves all reviews for a provider
func (k Keeper) GetReviewsByProvider(ctx sdk.Context, providerAddr string) []types.Review {
	store := ctx.KVStore(k.skey)
	reviewIDs := k.getFromIndex(store, types.GetProviderIndexKey(providerAddr))

	reviews := make([]types.Review, 0, len(reviewIDs))
	for _, id := range reviewIDs {
		if review, exists := k.GetReview(ctx, id); exists {
			reviews = append(reviews, review)
		}
	}
	return reviews
}

// GetReviewsByReviewer retrieves all reviews by a reviewer
func (k Keeper) GetReviewsByReviewer(ctx sdk.Context, reviewerAddr string) []types.Review {
	store := ctx.KVStore(k.skey)
	reviewIDs := k.getFromIndex(store, types.GetReviewerIndexKey(reviewerAddr))

	reviews := make([]types.Review, 0, len(reviewIDs))
	for _, id := range reviewIDs {
		if review, exists := k.GetReview(ctx, id); exists {
			reviews = append(reviews, review)
		}
	}
	return reviews
}

// DeleteReview marks a review as deleted (moderator action)
func (k Keeper) DeleteReview(ctx sdk.Context, reviewID string, moderatorAddr, reason string) error {
	review, exists := k.GetReview(ctx, reviewID)
	if !exists {
		return types.ErrReviewNotFound.Wrapf("review %s not found", reviewID)
	}

	// Check moderator permissions
	if k.rolesKeeper != nil {
		modAddr, err := sdk.AccAddressFromBech32(moderatorAddr)
		if err != nil {
			return types.ErrInvalidAddress.Wrapf("invalid moderator address: %v", err)
		}
		if !k.rolesKeeper.IsModerator(ctx, modAddr) && !k.rolesKeeper.IsAdmin(ctx, modAddr) {
			return fmt.Errorf("unauthorized: only moderators can delete reviews")
		}
	}

	// Mark as deleted
	if err := review.Delete(moderatorAddr, reason); err != nil {
		return err
	}

	// Update aggregation (remove the rating)
	if err := k.UpdateProviderAggregation(ctx, review.ProviderAddress, review.Rating, false); err != nil {
		k.Logger(ctx).Error("failed to update aggregation after delete", "error", err)
	}

	// Store updated review
	if err := k.SetReview(ctx, review); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeReviewDeleted,
			sdk.NewAttribute(types.AttributeKeyReviewID, reviewID),
			sdk.NewAttribute(types.AttributeKeyModeratorAddress, moderatorAddr),
			sdk.NewAttribute(types.AttributeKeyDeleteReason, reason),
		),
	})

	return nil
}

// GetProviderAggregation retrieves the rating aggregation for a provider
func (k Keeper) GetProviderAggregation(ctx sdk.Context, providerAddr string) (types.ProviderAggregation, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetProviderAggregationKey(providerAddr))
	if bz == nil {
		return types.ProviderAggregation{}, false
	}

	var agg types.ProviderAggregation
	if err := json.Unmarshal(bz, &agg); err != nil {
		return types.ProviderAggregation{}, false
	}
	return agg, true
}

// SetProviderAggregation stores a provider aggregation
func (k Keeper) SetProviderAggregation(ctx sdk.Context, agg types.ProviderAggregation) error {
	store := ctx.KVStore(k.skey)
	agg.BlockHeight = ctx.BlockHeight()

	bz, err := json.Marshal(agg)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregation: %w", err)
	}

	store.Set(types.GetProviderAggregationKey(agg.ProviderAddress), bz)
	return nil
}

// UpdateProviderAggregation updates the aggregation for a provider
func (k Keeper) UpdateProviderAggregation(ctx sdk.Context, providerAddr string, rating uint8, isAdd bool) error {
	agg, exists := k.GetProviderAggregation(ctx, providerAddr)
	if !exists {
		agg = *types.NewProviderAggregation(providerAddr)
	}

	blockTime := ctx.BlockTime().UTC()
	if isAdd {
		// BUGFIX-001: Use ctx.BlockTime() for consensus-safe time
		if err := agg.AddReview(rating, blockTime); err != nil {
			return err
		}
	} else {
		// BUGFIX-001: Use ctx.BlockTime() for consensus-safe time
		if err := agg.RemoveReview(rating, blockTime); err != nil {
			return err
		}
	}

	if err := k.SetProviderAggregation(ctx, agg); err != nil {
		return err
	}

	// Emit aggregation updated event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAggregationUpdated,
			sdk.NewAttribute(types.AttributeKeyProviderAddress, providerAddr),
			sdk.NewAttribute(types.AttributeKeyAverageRating, agg.GetAverageRatingDisplay()),
			sdk.NewAttribute(types.AttributeKeyTotalReviews, strconv.FormatUint(agg.TotalReviews, 10)),
		),
	})

	return nil
}

// GetTopProviders returns providers sorted by rating
func (k Keeper) GetTopProviders(ctx sdk.Context, limit int) []types.ProviderAggregation {
	var aggs types.ProviderAggregations
	k.WithProviderAggregations(ctx, func(agg types.ProviderAggregation) bool {
		aggs = append(aggs, agg)
		return false
	})
	return aggs.TopByRating(limit)
}

// WithReviews iterates over all reviews
func (k Keeper) WithReviews(ctx sdk.Context, fn func(types.Review) bool) {
	store := ctx.KVStore(k.skey)
	iter := prefix.NewStore(store, types.ReviewPrefix).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var review types.Review
		if err := json.Unmarshal(iter.Value(), &review); err != nil {
			continue
		}
		if fn(review) {
			break
		}
	}
}

// WithProviderAggregations iterates over all provider aggregations
func (k Keeper) WithProviderAggregations(ctx sdk.Context, fn func(types.ProviderAggregation) bool) {
	store := ctx.KVStore(k.skey)
	iter := prefix.NewStore(store, types.ProviderAggregationPrefix).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var agg types.ProviderAggregation
		if err := json.Unmarshal(iter.Value(), &agg); err != nil {
			continue
		}
		if fn(agg) {
			break
		}
	}
}

// Index helper functions
func (k Keeper) addToIndex(store storetypes.KVStore, key []byte, reviewID string) {
	bz := store.Get(key)
	var ids []string
	if bz != nil {
		_ = json.Unmarshal(bz, &ids)
	}

	// Check if already exists
	for _, id := range ids {
		if id == reviewID {
			return
		}
	}

	ids = append(ids, reviewID)
	newBz, _ := json.Marshal(ids)
	store.Set(key, newBz)
}

func (k Keeper) getFromIndex(store storetypes.KVStore, key []byte) []string {
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(bz, &ids); err != nil {
		return nil
	}
	return ids
}

func (k Keeper) removeFromIndex(store storetypes.KVStore, key []byte, reviewID string) {
	bz := store.Get(key)
	if bz == nil {
		return
	}

	var ids []string
	if err := json.Unmarshal(bz, &ids); err != nil {
		return
	}

	newIds := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != reviewID {
			newIds = append(newIds, id)
		}
	}

	if len(newIds) == 0 {
		store.Delete(key)
	} else {
		newBz, _ := json.Marshal(newIds)
		store.Set(key, newBz)
	}
}
