package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oraclev1 "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// PriceFeed defines a unified interface for oracle price sources.
type PriceFeed interface {
	GetPrice(ctx context.Context, base, quote string) (types.Price, error)
	GetPrices(ctx context.Context, pairs []types.CurrencyPair) ([]types.Price, error)
	SubscribePrices(ctx context.Context, pairs []types.CurrencyPair) (<-chan types.PriceUpdate, error)
}

// OracleKeeper defines the subset of x/oracle functionality needed by settlement.
type OracleKeeper interface {
	GetAggregatedPrice(ctx sdk.Context, denom, baseDenom string) (*oraclev1.AggregatedPrice, *oraclev1.PriceHealth, error)
}

// CosmosOraclePriceFeed adapts x/oracle aggregated prices to PriceFeed.
type CosmosOraclePriceFeed struct {
	keeper OracleKeeper
}

// NewCosmosOraclePriceFeed builds a Cosmos oracle adapter.
func NewCosmosOraclePriceFeed(keeper OracleKeeper) CosmosOraclePriceFeed {
	return CosmosOraclePriceFeed{keeper: keeper}
}

// GetPrice returns the aggregated price for a pair from x/oracle.
func (c CosmosOraclePriceFeed) GetPrice(ctx context.Context, base, quote string) (types.Price, error) {
	sdkCtx, err := unwrapSDKContext(ctx)
	if err != nil {
		return types.Price{}, err
	}
	agg, health, err := c.keeper.GetAggregatedPrice(sdkCtx, base, quote)
	if err != nil {
		return types.Price{}, err
	}
	if agg == nil {
		return types.Price{}, types.ErrOracleUnavailable.Wrap("oracle aggregated price missing")
	}
	if health != nil && !health.IsHealthy {
		return types.Price{}, types.ErrOracleUnavailable.Wrapf("oracle unhealthy: %v", health.FailureReason)
	}
	return types.Price{
		Base:      base,
		Quote:     quote,
		Rate:      agg.MedianPrice,
		Timestamp: agg.Timestamp,
		Source:    string(types.OracleSourceTypeCosmosOracle),
	}, nil
}

// GetPrices fetches aggregated prices for multiple pairs.
func (c CosmosOraclePriceFeed) GetPrices(ctx context.Context, pairs []types.CurrencyPair) ([]types.Price, error) {
	results := make([]types.Price, 0, len(pairs))
	for _, pair := range pairs {
		price, err := c.GetPrice(ctx, pair.Base, pair.Quote)
		if err != nil {
			return nil, err
		}
		results = append(results, price)
	}
	return results, nil
}

// SubscribePrices is not supported for the Cosmos oracle adapter yet.
func (c CosmosOraclePriceFeed) SubscribePrices(ctx context.Context, pairs []types.CurrencyPair) (<-chan types.PriceUpdate, error) {
	return nil, fmt.Errorf("cosmos oracle feed does not support subscriptions")
}

// manualPriceFeed exposes governance-set emergency prices.
type manualPriceFeed struct {
	getOverrides func(ctx sdk.Context) []types.ManualPriceOverride
}

func (m manualPriceFeed) GetPrice(ctx context.Context, base, quote string) (types.Price, error) {
	sdkCtx, err := unwrapSDKContext(ctx)
	if err != nil {
		return types.Price{}, err
	}
	now := sdkCtx.BlockTime()
	for _, override := range m.getOverrides(sdkCtx) {
		if !override.IsExpired(now) && strings.EqualFold(override.Base, base) && strings.EqualFold(override.Quote, quote) {
			return types.Price{
				Base:      base,
				Quote:     quote,
				Rate:      override.Rate,
				Timestamp: now,
				Source:    string(types.OracleSourceTypeManual),
			}, nil
		}
	}
	return types.Price{}, types.ErrOracleUnavailable.Wrap("manual price override not available")
}

func (m manualPriceFeed) GetPrices(ctx context.Context, pairs []types.CurrencyPair) ([]types.Price, error) {
	results := make([]types.Price, 0, len(pairs))
	for _, pair := range pairs {
		price, err := m.GetPrice(ctx, pair.Base, pair.Quote)
		if err != nil {
			continue
		}
		results = append(results, price)
	}
	if len(results) == 0 {
		return nil, types.ErrOracleUnavailable.Wrap("manual price override not available")
	}
	return results, nil
}

func (m manualPriceFeed) SubscribePrices(ctx context.Context, pairs []types.CurrencyPair) (<-chan types.PriceUpdate, error) {
	return nil, fmt.Errorf("manual price feed does not support subscriptions")
}

func (k Keeper) priceFeedForSource(sourceType types.OracleSourceType) PriceFeed {
	if sourceType == types.OracleSourceTypeManual {
		return manualPriceFeed{getOverrides: k.oracleManualPrices}
	}
	if k.priceFeeds == nil {
		return nil
	}
	return k.priceFeeds[sourceType]
}

func unwrapSDKContext(ctx context.Context) (sdkCtx sdk.Context, err error) {
	if sdkCtx, ok := ctx.(sdk.Context); ok {
		return sdkCtx, nil
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to unwrap sdk context: %v", r)
		}
	}()
	return sdk.UnwrapSDKContext(ctx), nil
}
