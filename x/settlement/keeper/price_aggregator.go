package keeper

import (
	"encoding/json"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

type oracleSource struct {
	config types.OracleSourceConfig
	feed   PriceFeed
}

type sourcePrice struct {
	price      types.Price
	priority   uint32
	sourceID   string
	sourceType types.OracleSourceType
}

// AggregatePrice returns the aggregated price for a single pair.
func (k Keeper) AggregatePrice(ctx sdk.Context, pair types.CurrencyPair) (types.Price, error) {
	prices, err := k.AggregatePrices(ctx, []types.CurrencyPair{pair})
	if err != nil {
		return types.Price{}, err
	}
	if len(prices) == 0 {
		return types.Price{}, types.ErrOracleUnavailable.Wrap("oracle returned no prices")
	}
	return prices[0], nil
}

// AggregatePrices returns aggregated prices for multiple pairs.
func (k Keeper) AggregatePrices(ctx sdk.Context, pairs []types.CurrencyPair) ([]types.Price, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	for _, pair := range pairs {
		if err := pair.Validate(); err != nil {
			return nil, err
		}
	}

	sources := k.enabledOracleSources(ctx)
	if len(sources) == 0 {
		return nil, types.ErrOracleUnavailable.Wrap("no oracle sources configured")
	}

	now := ctx.BlockTime()
	staleness := k.oracleStalenessThreshold(ctx)
	minSources := k.oracleMinSources(ctx)

	results := make([]types.Price, 0, len(pairs))
	for _, pair := range pairs {
		aggregated, err := k.aggregatePair(ctx, pair, sources, now, staleness, minSources)
		if err != nil {
			return nil, err
		}
		results = append(results, aggregated)
	}
	return results, nil
}

func (k Keeper) enabledOracleSources(ctx sdk.Context) []oracleSource {
	configs := k.oracleSources(ctx)
	sources := make([]oracleSource, 0, len(configs))
	for _, cfg := range configs {
		feed := k.priceFeedForSource(cfg.Type)
		if feed == nil {
			continue
		}
		sources = append(sources, oracleSource{config: cfg, feed: feed})
	}
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].config.Priority == sources[j].config.Priority {
			return sources[i].config.ID < sources[j].config.ID
		}
		return sources[i].config.Priority < sources[j].config.Priority
	})
	return sources
}

func (k Keeper) aggregatePair(
	ctx sdk.Context,
	pair types.CurrencyPair,
	sources []oracleSource,
	now time.Time,
	staleness time.Duration,
	minSources int,
) (types.Price, error) {
	valid := make([]sourcePrice, 0, len(sources))
	staleRejected := false
	for _, source := range sources {
		price, err := source.feed.GetPrice(ctx, pair.Base, pair.Quote)
		if err != nil {
			continue
		}
		if price.Timestamp.IsZero() {
			price.Timestamp = now
		}
		if now.Sub(price.Timestamp) > staleness {
			staleRejected = true
			continue
		}
		if err := price.Validate(); err != nil {
			continue
		}
		if price.Source == "" {
			price.Source = source.config.ID
		}
		valid = append(valid, sourcePrice{
			price:      price,
			priority:   source.config.Priority,
			sourceID:   source.config.ID,
			sourceType: source.config.Type,
		})
	}

	if len(valid) == 0 {
		if staleRejected {
			return types.Price{}, types.ErrOracleStalePrice.Wrap("oracle prices stale")
		}
		return types.Price{}, types.ErrOracleUnavailable.Wrap("no valid oracle prices")
	}

	if len(valid) < minSources {
		if manual, ok := selectManualOverride(valid); ok {
			return k.fallbackManualPrice(ctx, manual, now)
		}
		return types.Price{}, types.ErrOracleInsufficientSources.Wrap("not enough oracle sources")
	}

	aggregated := medianPrice(valid, pair, now)
	k.recordPriceHistory(ctx, aggregated)
	k.maybeAlertPriceDeviation(ctx, aggregated)
	k.recordLatestPrice(ctx, aggregated)
	for _, entry := range valid {
		k.recordPriceHistory(ctx, entry.price)
	}

	return aggregated, nil
}

func selectManualOverride(valid []sourcePrice) (sourcePrice, bool) {
	for _, entry := range valid {
		if entry.sourceType == types.OracleSourceTypeManual {
			return entry, true
		}
	}
	return sourcePrice{}, false
}

func (k Keeper) fallbackManualPrice(ctx sdk.Context, manual sourcePrice, now time.Time) (types.Price, error) {
	fallback := manual.price
	if fallback.Timestamp.IsZero() {
		fallback.Timestamp = now
	}
	fallback.Source = "manual-override"
	k.recordPriceHistory(ctx, fallback)
	k.maybeAlertPriceDeviation(ctx, fallback)
	k.recordLatestPrice(ctx, fallback)
	return fallback, nil
}

func medianPrice(prices []sourcePrice, pair types.CurrencyPair, now time.Time) types.Price {
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].price.Rate.LT(prices[j].price.Rate)
	})
	mid := len(prices) / 2
	median := prices[mid].price.Rate
	if len(prices)%2 == 0 {
		median = prices[mid-1].price.Rate.Add(prices[mid].price.Rate).QuoInt64(2)
	}
	return types.Price{
		Base:      pair.Base,
		Quote:     pair.Quote,
		Rate:      median,
		Timestamp: now,
		Source:    "aggregated",
	}
}

func (k Keeper) recordPriceHistory(ctx sdk.Context, price types.Price) {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&price)
	if err != nil {
		return
	}
	key := types.OraclePriceHistoryKey(price.Base, price.Quote, timestampToUint64(price.Timestamp))
	store.Set(key, bz)
}

func (k Keeper) recordLatestPrice(ctx sdk.Context, price types.Price) {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&price)
	if err != nil {
		return
	}
	key := types.OracleLatestPriceKey(price.Base, price.Quote)
	store.Set(key, bz)
}

// GetLatestOraclePrice retrieves the latest aggregated price for a pair.
func (k Keeper) GetLatestOraclePrice(ctx sdk.Context, pair types.CurrencyPair) (types.Price, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.OracleLatestPriceKey(pair.Base, pair.Quote))
	if bz == nil {
		return types.Price{}, false
	}
	var price types.Price
	if err := json.Unmarshal(bz, &price); err != nil {
		return types.Price{}, false
	}
	return price, true
}

func (k Keeper) maybeAlertPriceDeviation(ctx sdk.Context, price types.Price) {
	last, found := k.GetLatestOraclePrice(ctx, price.Pair())
	if !found {
		return
	}
	window := k.oracleDeviationWindow(ctx)
	if price.Timestamp.Sub(last.Timestamp) > window {
		return
	}
	if !last.Rate.IsPositive() {
		return
	}
	delta := price.Rate.Sub(last.Rate)
	if delta.IsNegative() {
		delta = delta.Neg()
	}
	changePct := delta.Quo(last.Rate)
	threshold := sdkmath.LegacyNewDec(int64(k.oracleDeviationThresholdBps(ctx))).QuoInt64(10000)
	if changePct.LT(threshold) {
		return
	}
	alert := types.PriceAlert{
		Base:         price.Base,
		Quote:        price.Quote,
		OldRate:      last.Rate,
		NewRate:      price.Rate,
		ChangePct:    changePct,
		OccurredAt:   price.Timestamp,
		WindowSec:    uint64(window.Seconds()),
		ThresholdBps: k.oracleDeviationThresholdBps(ctx),
		Source:       price.Source,
	}
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(&alert)
	if err != nil {
		return
	}
	store.Set(types.OraclePriceAlertKey(price.Base, price.Quote, timestampToUint64(price.Timestamp)), bz)
}

func timestampToUint64(ts time.Time) uint64 {
	unixNano := ts.UnixNano()
	if unixNano < 0 {
		return 0
	}
	return uint64(unixNano)
}
