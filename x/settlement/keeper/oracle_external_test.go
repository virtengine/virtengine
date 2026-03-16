package keeper

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/pricefeed"
	"github.com/virtengine/virtengine/x/settlement/types"
)

type mockExternalPriceProvider struct {
	price        pricefeed.AggregatedPrice
	err          error
	lastBase     string
	lastQuote    string
	requestCount int
}

func (m *mockExternalPriceProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (pricefeed.AggregatedPrice, error) {
	m.lastBase = baseAsset
	m.lastQuote = quoteAsset
	m.requestCount++
	if m.err != nil {
		return pricefeed.AggregatedPrice{}, m.err
	}
	return m.price, nil
}

func TestExternalOraclePriceFeedGetPriceNormalizesBaseAndQuote(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	provider := &mockExternalPriceProvider{
		price: pricefeed.AggregatedPrice{
			PriceData: pricefeed.PriceData{
				Price:     sdkmath.LegacyMustNewDecFromStr("1.15"),
				Timestamp: now,
				Source:    "coingecko-primary",
			},
		},
	}
	feed := NewExternalOraclePriceFeed(provider, types.OracleSourceTypeBandIBC)

	price, err := feed.GetPrice(context.Background(), "VRT", "USD")
	require.NoError(t, err)
	require.Equal(t, "uve", provider.lastBase)
	require.Equal(t, "usd", provider.lastQuote)
	require.Equal(t, "VRT", price.Base)
	require.Equal(t, "USD", price.Quote)
	require.True(t, price.Rate.Equal(sdkmath.LegacyMustNewDecFromStr("1.15")))
	require.Equal(t, now, price.Timestamp)
	require.Equal(t, "coingecko-primary", price.Source)
}

func TestExternalOraclePriceFeedFallsBackToSourceType(t *testing.T) {
	provider := &mockExternalPriceProvider{
		price: pricefeed.AggregatedPrice{
			PriceData: pricefeed.PriceData{
				Price:  sdkmath.LegacyMustNewDecFromStr("0.99"),
				Source: "",
			},
		},
	}
	feed := NewExternalOraclePriceFeed(provider, types.OracleSourceTypeChainlinkIBC)

	price, err := feed.GetPrice(context.Background(), "btc", "usd")
	require.NoError(t, err)
	require.Equal(t, string(types.OracleSourceTypeChainlinkIBC), price.Source)
	require.False(t, price.Timestamp.IsZero())
}

func TestExternalOraclePriceFeedPropagatesProviderErrors(t *testing.T) {
	provider := &mockExternalPriceProvider{err: errors.New("provider unavailable")}
	feed := NewExternalOraclePriceFeed(provider, types.OracleSourceTypeBandIBC)

	_, err := feed.GetPrice(context.Background(), "VRT", "USD")
	require.ErrorContains(t, err, "provider unavailable")
}

func TestExternalOraclePriceFeedRejectsNilProvider(t *testing.T) {
	feed := NewExternalOraclePriceFeed(nil, types.OracleSourceTypeBandIBC)

	_, err := feed.GetPrice(context.Background(), "VRT", "USD")
	require.ErrorContains(t, err, "external price provider not configured")
}

func TestExternalOraclePriceFeedGetPrices(t *testing.T) {
	provider := &mockExternalPriceProvider{
		price: pricefeed.AggregatedPrice{
			PriceData: pricefeed.PriceData{
				Price:     sdkmath.LegacyMustNewDecFromStr("42.5"),
				Timestamp: time.Date(2026, 3, 15, 12, 30, 0, 0, time.UTC),
				Source:    "external-pricing",
			},
		},
	}
	feed := NewExternalOraclePriceFeed(provider, types.OracleSourceTypeBandIBC)

	prices, err := feed.GetPrices(context.Background(), []types.CurrencyPair{
		{Base: "VRT", Quote: "USD"},
		{Base: "BTC", Quote: "USD"},
	})
	require.NoError(t, err)
	require.Len(t, prices, 2)
	require.Equal(t, 2, provider.requestCount)
}

func TestExternalOraclePriceFeedGetPricesStopsOnProviderError(t *testing.T) {
	provider := &mockExternalPriceProvider{err: errors.New("provider unavailable")}
	feed := NewExternalOraclePriceFeed(provider, types.OracleSourceTypeBandIBC)

	_, err := feed.GetPrices(context.Background(), []types.CurrencyPair{{Base: "VRT", Quote: "USD"}})
	require.ErrorContains(t, err, "provider unavailable")
	require.Equal(t, 1, provider.requestCount)
}

func TestExternalOraclePriceFeedSubscribePricesUnsupported(t *testing.T) {
	provider := &mockExternalPriceProvider{}
	feed := NewExternalOraclePriceFeed(provider, types.OracleSourceTypeBandIBC)

	updates, err := feed.SubscribePrices(context.Background(), []types.CurrencyPair{{Base: "VRT", Quote: "USD"}})
	require.Nil(t, updates)
	require.ErrorContains(t, err, "does not support subscriptions")
}
