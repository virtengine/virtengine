// Copyright (c) VirtEngine Inc.
// SPDX-License-Identifier: BUSL-1.1

package keeper

import (
	"context"
	"sort"
	"time"

	stdmath "math"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	otypes "github.com/virtengine/virtengine/x/oracle/types"
)

// IKeeper defines the expected interface for the Oracle module keeper.
type IKeeper interface {
	Codec() codec.BinaryCodec
	StoreKey() storetypes.StoreKey
	GetAuthority() string

	// Params
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error

	// Price operations
	AddPriceEntry(ctx sdk.Context, source uint32, denom, baseDenom string, price types.PriceDataState) error
	GetLatestPrice(ctx sdk.Context, source uint32, denom, baseDenom string) (*types.PriceData, bool)
	GetAggregatedPrice(ctx sdk.Context, denom, baseDenom string) (*types.AggregatedPrice, *types.PriceHealth, error)
	GetPrices(ctx sdk.Context, filters types.PricesFilter) ([]types.PriceData, error)
	GetPriceFeedConfig(ctx sdk.Context, denom string) (*types.QueryPriceFeedConfigResponse, error)

	// Query server
	NewQuerier() Querier
}

// Keeper implements the Oracle module keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// The address capable of executing a MsgUpdateParams message.
	// This should be the x/gov module account.
	authority string
}

// NewKeeper creates a new Oracle Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
) IKeeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// Codec returns the keeper's codec.
func (k *Keeper) Codec() codec.BinaryCodec {
	return k.cdc
}

// StoreKey returns the keeper's store key.
func (k *Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetAuthority returns the x/oracle module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the current Oracle module parameters.
func (k *Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(otypes.ParamsPrefix())
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the Oracle module parameters.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.ValidateBasic(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(otypes.ParamsPrefix(), bz)

	return nil
}

// AddPriceEntry adds a new price entry for a source and denomination pair.
func (k *Keeper) AddPriceEntry(ctx sdk.Context, source uint32, denom, baseDenom string, priceState types.PriceDataState) error {
	store := ctx.KVStore(k.storeKey)
	height := ctx.BlockHeight()

	// Create price data record
	priceData := types.PriceData{
		ID: types.PriceDataRecordID{
			Source:    source,
			Denom:     denom,
			BaseDenom: baseDenom,
			Height:    height,
		},
		State: priceState,
	}

	// Store the price data
	key := otypes.PriceDataKey(source, denom, baseDenom, height)
	bz := k.cdc.MustMarshal(&priceData)
	store.Set(key, bz)

	// Update latest price data ID
	latestID := types.PriceDataID{
		Source:    source,
		Denom:     denom,
		BaseDenom: baseDenom,
	}
	latestKey := otypes.LatestPriceDataKey(source, denom, baseDenom)
	latestBz := k.cdc.MustMarshal(&latestID)
	store.Set(latestKey, latestBz)

	return nil
}

// GetLatestPrice returns the latest price for a source and denomination pair.
func (k *Keeper) GetLatestPrice(ctx sdk.Context, source uint32, denom, baseDenom string) (*types.PriceData, bool) {
	store := ctx.KVStore(k.storeKey)

	// Get iterator for all prices for this source/pair, reverse to get latest
	prefix := otypes.PriceDataPrefixByPair(source, denom, baseDenom)
	iter := storetypes.KVStoreReversePrefixIterator(store, prefix)
	defer iter.Close()

	if !iter.Valid() {
		return nil, false
	}

	var priceData types.PriceData
	if err := k.cdc.Unmarshal(iter.Value(), &priceData); err != nil {
		return nil, false
	}

	return &priceData, true
}

// GetAggregatedPrice calculates the aggregated price from all sources.
func (k *Keeper) GetAggregatedPrice(ctx sdk.Context, denom, baseDenom string) (*types.AggregatedPrice, *types.PriceHealth, error) {
	params := k.GetParams(ctx)
	currentHeight := ctx.BlockHeight()
	maxStaleness := params.MaxPriceStalenessBlocks

	// Collect prices from all sources
	prices := make([]sdkmath.LegacyDec, 0, len(params.Sources))
	var totalSources uint32
	var healthySources uint32
	var failureReasons []string

	// Iterate through sources (0 to len(params.Sources))
	for sourceIdx := range params.Sources {
		source := safeUint32FromInt(sourceIdx)
		totalSources++

		priceData, found := k.GetLatestPrice(ctx, source, denom, baseDenom)
		if !found {
			failureReasons = append(failureReasons, "source has no price data")
			continue
		}

		// Check staleness
		if currentHeight-priceData.ID.Height > maxStaleness {
			failureReasons = append(failureReasons, "price data is stale")
			continue
		}

		healthySources++
		prices = append(prices, priceData.State.Price)
	}

	// Check minimum sources requirement
	hasMinSources := healthySources >= uint32(params.MinPriceSources)
	if !hasMinSources {
		failureReasons = append(failureReasons, "insufficient price sources")
	}

	health := &types.PriceHealth{
		Denom:               denom,
		IsHealthy:           hasMinSources && len(failureReasons) == 0,
		HasMinSources:       hasMinSources,
		DeviationOk:         true, // Will be updated below
		TotalSources:        totalSources,
		TotalHealthySources: healthySources,
		FailureReason:       failureReasons,
	}

	if len(prices) == 0 {
		// Return empty aggregated price with unhealthy status
		return &types.AggregatedPrice{
			Denom:       denom,
			TWAP:        sdkmath.LegacyZeroDec(),
			MedianPrice: sdkmath.LegacyZeroDec(),
			MinPrice:    sdkmath.LegacyZeroDec(),
			MaxPrice:    sdkmath.LegacyZeroDec(),
			Timestamp:   ctx.BlockTime(),
			NumSources:  0,
		}, health, nil
	}

	// Sort prices for median calculation
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].LT(prices[j])
	})

	// Calculate median
	var median sdkmath.LegacyDec
	n := len(prices)
	if n%2 == 0 {
		median = prices[n/2-1].Add(prices[n/2]).Quo(sdkmath.LegacyNewDec(2))
	} else {
		median = prices[n/2]
	}

	// Calculate min and max
	minPrice := prices[0]
	maxPrice := prices[n-1]

	// Calculate deviation in basis points
	var deviationBps uint64
	if !minPrice.IsZero() {
		deviation := maxPrice.Sub(minPrice).Quo(minPrice).Mul(sdkmath.LegacyNewDec(10000))
		deviationBps = deviation.TruncateInt().Uint64()
	}

	// Check deviation against max allowed
	if deviationBps > params.MaxPriceDeviationBps {
		health.DeviationOk = false
		health.IsHealthy = false
		health.FailureReason = append(health.FailureReason, "price deviation exceeds maximum")
	}

	// Calculate TWAP (simplified - using median for now, full TWAP requires historical data)
	twap := k.calculateTWAP(ctx, denom, baseDenom, params.TwapWindow)
	if twap.IsZero() {
		twap = median // Fall back to median if TWAP calculation fails
	}

	aggregated := &types.AggregatedPrice{
		Denom:        denom,
		TWAP:         twap,
		MedianPrice:  median,
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		Timestamp:    ctx.BlockTime(),
		NumSources:   healthySources,
		DeviationBps: deviationBps,
	}

	return aggregated, health, nil
}

// calculateTWAP calculates the time-weighted average price over a window.
func (k *Keeper) calculateTWAP(ctx sdk.Context, denom, baseDenom string, windowBlocks int64) sdkmath.LegacyDec {
	store := ctx.KVStore(k.storeKey)
	currentHeight := ctx.BlockHeight()
	startHeight := currentHeight - windowBlocks
	if startHeight < 0 {
		startHeight = 0
	}

	params := k.GetParams(ctx)
	totalWeightedPrice := sdkmath.LegacyZeroDec()
	var totalWeight int64

	// Collect prices from all sources within the window
	for sourceIdx := range params.Sources {
		source := safeUint32FromInt(sourceIdx)
		prefix := otypes.PriceDataPrefixByPair(source, denom, baseDenom)
		iter := storetypes.KVStoreReversePrefixIterator(store, prefix)
		defer iter.Close()

		prevHeight := currentHeight
		for ; iter.Valid(); iter.Next() {
			var priceData types.PriceData
			if err := k.cdc.Unmarshal(iter.Value(), &priceData); err != nil {
				continue
			}

			if priceData.ID.Height < startHeight {
				break
			}

			// Weight is the number of blocks this price was valid
			weight := prevHeight - priceData.ID.Height
			if weight > 0 {
				totalWeightedPrice = totalWeightedPrice.Add(priceData.State.Price.MulInt64(weight))
				totalWeight += weight
			}
			prevHeight = priceData.ID.Height
		}
	}

	if totalWeight == 0 {
		return sdkmath.LegacyZeroDec()
	}

	return totalWeightedPrice.QuoInt64(totalWeight)
}

// GetPrices returns price data matching the given filters.
func (k *Keeper) GetPrices(ctx sdk.Context, filters types.PricesFilter) ([]types.PriceData, error) {
	store := ctx.KVStore(k.storeKey)
	params := k.GetParams(ctx)

	// If both asset and base denom are specified, use targeted query
	if filters.AssetDenom != "" && filters.BaseDenom != "" {
		return k.getPricesByPair(store, params, filters)
	}

	// Otherwise, do a full scan with filtering
	return k.getAllPricesFiltered(store, filters)
}

// getPricesByPair retrieves prices for a specific asset/base denomination pair.
func (k *Keeper) getPricesByPair(store storetypes.KVStore, params types.Params, filters types.PricesFilter) ([]types.PriceData, error) {
	results := make([]types.PriceData, 0, len(params.Sources))

	for sourceIdx := range params.Sources {
		source := safeUint32FromInt(sourceIdx)
		prices := k.getPricesForSource(store, source, filters)
		results = append(results, prices...)
	}

	return results, nil
}

// getPricesForSource retrieves prices for a specific source.
func (k *Keeper) getPricesForSource(store storetypes.KVStore, source uint32, filters types.PricesFilter) []types.PriceData {
	var results []types.PriceData

	if filters.Height > 0 {
		// Get specific price at height
		priceData, found := k.getPriceAtHeight(store, source, filters.AssetDenom, filters.BaseDenom, filters.Height)
		if found {
			results = append(results, priceData)
		}
	} else {
		// Get all prices for this pair
		results = k.getAllPricesForPair(store, source, filters.AssetDenom, filters.BaseDenom)
	}

	return results
}

func safeUint32FromInt(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > stdmath.MaxInt32 {
		return stdmath.MaxUint32
	}
	//nolint:gosec // range checked above
	return uint32(value)
}

// getPriceAtHeight retrieves a specific price at a given block height.
func (k *Keeper) getPriceAtHeight(store storetypes.KVStore, source uint32, denom, baseDenom string, height int64) (types.PriceData, bool) {
	key := otypes.PriceDataKey(source, denom, baseDenom, height)
	bz := store.Get(key)
	if bz == nil {
		return types.PriceData{}, false
	}

	var priceData types.PriceData
	if err := k.cdc.Unmarshal(bz, &priceData); err != nil {
		return types.PriceData{}, false
	}

	return priceData, true
}

// getAllPricesForPair retrieves all prices for a specific source and pair.
func (k *Keeper) getAllPricesForPair(store storetypes.KVStore, source uint32, denom, baseDenom string) []types.PriceData {
	var results []types.PriceData
	prefix := otypes.PriceDataPrefixByPair(source, denom, baseDenom)
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var priceData types.PriceData
		if err := k.cdc.Unmarshal(iter.Value(), &priceData); err == nil {
			results = append(results, priceData)
		}
	}

	return results
}

// getAllPricesFiltered retrieves all prices with optional filtering.
func (k *Keeper) getAllPricesFiltered(store storetypes.KVStore, filters types.PricesFilter) ([]types.PriceData, error) {
	var results []types.PriceData
	iter := storetypes.KVStorePrefixIterator(store, otypes.PriceDataPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var priceData types.PriceData
		if err := k.cdc.Unmarshal(iter.Value(), &priceData); err != nil {
			continue
		}

		if k.matchesFilters(priceData, filters) {
			results = append(results, priceData)
		}
	}

	return results, nil
}

// matchesFilters checks if a price data record matches the given filters.
func (k *Keeper) matchesFilters(priceData types.PriceData, filters types.PricesFilter) bool {
	if filters.AssetDenom != "" && priceData.ID.Denom != filters.AssetDenom {
		return false
	}
	if filters.BaseDenom != "" && priceData.ID.BaseDenom != filters.BaseDenom {
		return false
	}
	if filters.Height > 0 && priceData.ID.Height != filters.Height {
		return false
	}
	return true
}

// GetPriceFeedConfig returns the price feed configuration for a denomination.
func (k *Keeper) GetPriceFeedConfig(ctx sdk.Context, denom string) (*types.QueryPriceFeedConfigResponse, error) {
	params := k.GetParams(ctx)

	// Check if any source is configured for this denom
	enabled := len(params.Sources) > 0

	return &types.QueryPriceFeedConfigResponse{
		PriceFeedId:         "", // Would be set from FeedContractsParams
		PythContractAddress: "", // Would be set from FeedContractsParams
		Enabled:             enabled,
	}, nil
}

// NewQuerier creates a new Querier instance.
func (k *Keeper) NewQuerier() Querier {
	return Querier{keeper: k}
}

// Querier implements the grpc query service for the Oracle module.
type Querier struct {
	keeper IKeeper
}

// Params implements the Query/Params gRPC method.
func (q Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := q.keeper.GetParams(sdkCtx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// AggregatedPrice implements the Query/AggregatedPrice gRPC method.
func (q Querier) AggregatedPrice(ctx context.Context, req *types.QueryAggregatedPriceRequest) (*types.QueryAggregatedPriceResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Default base denom to "usd" if not specified in request
	baseDenom := "usd"

	aggregatedPrice, priceHealth, err := q.keeper.GetAggregatedPrice(sdkCtx, req.Denom, baseDenom)
	if err != nil {
		return nil, err
	}

	return &types.QueryAggregatedPriceResponse{
		AggregatedPrice: *aggregatedPrice,
		PriceHealth:     *priceHealth,
	}, nil
}

// PriceFeedConfig implements the Query/PriceFeedConfig gRPC method.
func (q Querier) PriceFeedConfig(ctx context.Context, req *types.QueryPriceFeedConfigRequest) (*types.QueryPriceFeedConfigResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return q.keeper.GetPriceFeedConfig(sdkCtx, req.Denom)
}

// Prices implements the Query/Prices gRPC method.
func (q Querier) Prices(ctx context.Context, req *types.QueryPricesRequest) (*types.QueryPricesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	prices, err := q.keeper.GetPrices(sdkCtx, req.Filters)
	if err != nil {
		return nil, err
	}

	return &types.QueryPricesResponse{
		Prices: prices,
	}, nil
}

// Ensure Querier implements the QueryServer interface
var _ types.QueryServer = Querier{}

// getDefaultPriceData returns default price data when not found
//nolint:unused // reserved for default price fallback during maintenance
func getDefaultPriceData() types.PriceData {
	return types.PriceData{
		ID: types.PriceDataRecordID{},
		State: types.PriceDataState{
			Price:     sdkmath.LegacyZeroDec(),
			Timestamp: time.Time{},
		},
	}
}
