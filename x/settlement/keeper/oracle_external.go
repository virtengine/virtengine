package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/virtengine/virtengine/pkg/pricefeed"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// ExternalPriceProvider defines the external aggregator contract used by settlement pricing.
type ExternalPriceProvider interface {
	GetPrice(ctx context.Context, baseAsset, quoteAsset string) (pricefeed.AggregatedPrice, error)
}

// ExternalSourcePriceProvider resolves prices from a named external source.
type ExternalSourcePriceProvider interface {
	GetPriceFromSource(ctx context.Context, source, baseAsset, quoteAsset string) (pricefeed.PriceData, error)
}

// ExternalOraclePriceFeed adapts external pricefeed aggregators to settlement PriceFeed.
type ExternalOraclePriceFeed struct {
	provider     ExternalPriceProvider
	sourceType   types.OracleSourceType
	providerName string
}

// NewExternalOraclePriceFeed creates an adapter for external FX/crypto pricing providers.
func NewExternalOraclePriceFeed(provider ExternalPriceProvider, sourceType types.OracleSourceType) ExternalOraclePriceFeed {
	return ExternalOraclePriceFeed{provider: provider, sourceType: sourceType}
}

// NewExternalOracleSourcePriceFeed creates an adapter pinned to a specific external source.
func NewExternalOracleSourcePriceFeed(provider ExternalPriceProvider, sourceType types.OracleSourceType, providerName string) ExternalOraclePriceFeed {
	return ExternalOraclePriceFeed{
		provider:     provider,
		sourceType:   sourceType,
		providerName: strings.TrimSpace(providerName),
	}
}

func (e ExternalOraclePriceFeed) GetPrice(ctx context.Context, base, quote string) (types.Price, error) {
	if e.provider == nil {
		return types.Price{}, types.ErrOracleUnavailable.Wrap("external price provider not configured")
	}

	normalizedBase := normalizeExternalBase(base)
	normalizedQuote := normalizeExternalQuote(quote)

	if e.providerName != "" {
		if namedProvider, ok := e.provider.(ExternalSourcePriceProvider); ok {
			priceData, err := namedProvider.GetPriceFromSource(ctx, e.providerName, normalizedBase, normalizedQuote)
			if err != nil {
				return types.Price{}, err
			}
			return e.toSettlementPrice(ctx, base, quote, priceData.Price, priceData.Timestamp, priceData.Source), nil
		}
	}

	aggregated, err := e.provider.GetPrice(ctx, normalizedBase, normalizedQuote)
	if err != nil {
		return types.Price{}, err
	}

	return e.toSettlementPrice(ctx, base, quote, aggregated.Price, aggregated.Timestamp, aggregated.Source), nil
}

func (e ExternalOraclePriceFeed) toSettlementPrice(ctx context.Context, base, quote string, rate sdkmath.LegacyDec, ts time.Time, source string) types.Price {
	timestamp := externalPriceTimestamp(ctx, ts)
	source = strings.TrimSpace(source)
	if source == "" {
		source = string(e.sourceType)
	}

	return types.Price{
		Base:      strings.ToUpper(strings.TrimSpace(base)),
		Quote:     strings.ToUpper(strings.TrimSpace(quote)),
		Rate:      rate,
		Timestamp: timestamp,
		Source:    source,
	}
}

func (e ExternalOraclePriceFeed) GetPrices(ctx context.Context, pairs []types.CurrencyPair) ([]types.Price, error) {
	results := make([]types.Price, 0, len(pairs))
	for _, pair := range pairs {
		price, err := e.GetPrice(ctx, pair.Base, pair.Quote)
		if err != nil {
			return nil, err
		}
		results = append(results, price)
	}
	return results, nil
}

func (e ExternalOraclePriceFeed) SubscribePrices(ctx context.Context, pairs []types.CurrencyPair) (<-chan types.PriceUpdate, error) {
	return nil, fmt.Errorf("external oracle feed does not support subscriptions")
}

func normalizeExternalBase(base string) string {
	normalized := strings.ToLower(strings.TrimSpace(base))
	switch normalized {
	case "vrt":
		return "uve"
	default:
		return normalized
	}
}

func normalizeExternalQuote(quote string) string {
	return strings.ToLower(strings.TrimSpace(quote))
}

func externalPriceTimestamp(ctx context.Context, ts time.Time) time.Time {
	if !ts.IsZero() {
		return ts
	}
	sdkCtx, err := unwrapSDKContext(ctx)
	if err == nil {
		return sdkCtx.BlockTime()
	}
	return time.Now().UTC()
}
