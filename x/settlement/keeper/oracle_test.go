package keeper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/settlement/types"
)

type mockPriceFeed struct {
	prices map[string]types.Price
	err    error
}

func (m mockPriceFeed) GetPrice(ctx context.Context, base, quote string) (types.Price, error) {
	if m.err != nil {
		return types.Price{}, m.err
	}
	price, ok := m.prices[pairKey(base, quote)]
	if !ok {
		return types.Price{}, types.ErrOracleUnavailable.Wrap("price not found")
	}
	return price, nil
}

func (m mockPriceFeed) GetPrices(ctx context.Context, pairs []types.CurrencyPair) ([]types.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	results := make([]types.Price, 0, len(pairs))
	for _, pair := range pairs {
		if price, ok := m.prices[pairKey(pair.Base, pair.Quote)]; ok {
			results = append(results, price)
		}
	}
	if len(results) == 0 {
		return nil, types.ErrOracleUnavailable.Wrap("prices not found")
	}
	return results, nil
}

func (m mockPriceFeed) SubscribePrices(ctx context.Context, pairs []types.CurrencyPair) (<-chan types.PriceUpdate, error) {
	return nil, errors.New("subscription not supported")
}

func pairKey(base, quote string) string {
	return base + "/" + quote
}

func priceFor(quote, rate string, ts time.Time, source string) types.Price {
	dec, _ := sdkmath.LegacyNewDecFromStr(rate)
	return types.Price{
		Base:      "VRT",
		Quote:     quote,
		Rate:      dec,
		Timestamp: ts,
		Source:    source,
	}
}

func (s *KeeperTestSuite) setOracleParams(sources []types.OracleSourceConfig, minSources uint32, stalenessSec uint64) {
	t := s.T()
	params := s.keeper.GetParams(s.ctx)
	params.OracleSources = sources
	params.OracleMinSources = minSources
	params.OracleStalenessThresholdSeconds = stalenessSec
	params.OracleDeviationThresholdBps = 500
	params.OracleDeviationWindowSeconds = 60
	require.NoError(t, s.keeper.SetParams(s.ctx, params))
}

func (s *KeeperTestSuite) oraclePriceFetchWithMockProviders() {
	t := s.T()
	now := s.ctx.BlockTime()

	feed := mockPriceFeed{
		prices: map[string]types.Price{
			pairKey("VRT", "USD"): priceFor("USD", "1.05", now, "cosmos"),
		},
	}

	s.keeper.SetPriceFeed(types.OracleSourceTypeCosmosOracle, feed)
	s.setOracleParams([]types.OracleSourceConfig{
		{ID: "cosmos", Type: types.OracleSourceTypeCosmosOracle, Enabled: true, Priority: 1},
	}, 1, 300)

	price, err := s.keeper.AggregatePrice(s.ctx, types.CurrencyPair{Base: "VRT", Quote: "USD"})
	require.NoError(t, err)
	require.Equal(t, "VRT", price.Base)
	require.Equal(t, "USD", price.Quote)
	require.True(t, price.Rate.GT(sdkmath.LegacyZeroDec()))
}

func (s *KeeperTestSuite) oraclePriceAggregationMedianOutlier() {
	t := s.T()
	now := s.ctx.BlockTime()

	s.keeper.SetPriceFeed(types.OracleSourceTypeCosmosOracle, mockPriceFeed{
		prices: map[string]types.Price{pairKey("VRT", "USD"): priceFor("USD", "1.00", now, "cosmos")},
	})
	s.keeper.SetPriceFeed(types.OracleSourceTypeBandIBC, mockPriceFeed{
		prices: map[string]types.Price{pairKey("VRT", "USD"): priceFor("USD", "1.10", now, "band")},
	})
	s.keeper.SetPriceFeed(types.OracleSourceTypeChainlinkIBC, mockPriceFeed{
		prices: map[string]types.Price{pairKey("VRT", "USD"): priceFor("USD", "100.00", now, "chainlink")},
	})

	s.setOracleParams([]types.OracleSourceConfig{
		{ID: "cosmos", Type: types.OracleSourceTypeCosmosOracle, Enabled: true, Priority: 1},
		{ID: "band", Type: types.OracleSourceTypeBandIBC, Enabled: true, Priority: 2},
		{ID: "chainlink", Type: types.OracleSourceTypeChainlinkIBC, Enabled: true, Priority: 3},
	}, 3, 300)

	price, err := s.keeper.AggregatePrice(s.ctx, types.CurrencyPair{Base: "VRT", Quote: "USD"})
	require.NoError(t, err)
	expected := sdkmath.LegacyMustNewDecFromStr("1.10")
	require.True(t, price.Rate.Equal(expected))
}

func (s *KeeperTestSuite) oraclePriceStalenessRejection() {
	t := s.T()
	staleTime := s.ctx.BlockTime().Add(-10 * time.Minute)

	s.keeper.SetPriceFeed(types.OracleSourceTypeCosmosOracle, mockPriceFeed{
		prices: map[string]types.Price{pairKey("VRT", "USD"): priceFor("USD", "1.05", staleTime, "cosmos")},
	})

	s.setOracleParams([]types.OracleSourceConfig{
		{ID: "cosmos", Type: types.OracleSourceTypeCosmosOracle, Enabled: true, Priority: 1},
	}, 1, 60)

	_, err := s.keeper.AggregatePrice(s.ctx, types.CurrencyPair{Base: "VRT", Quote: "USD"})
	require.ErrorIs(t, err, types.ErrOracleStalePrice)
}

func (s *KeeperTestSuite) oracleFallbackChain() {
	t := s.T()
	now := s.ctx.BlockTime()

	s.keeper.SetPriceFeed(types.OracleSourceTypeCosmosOracle, mockPriceFeed{err: errors.New("oracle down")})
	s.keeper.SetPriceFeed(types.OracleSourceTypeBandIBC, mockPriceFeed{
		prices: map[string]types.Price{pairKey("VRT", "USD"): priceFor("USD", "1.02", now, "band")},
	})
	s.keeper.SetPriceFeed(types.OracleSourceTypeChainlinkIBC, mockPriceFeed{
		prices: map[string]types.Price{pairKey("VRT", "USD"): priceFor("USD", "1.04", now, "chainlink")},
	})

	s.setOracleParams([]types.OracleSourceConfig{
		{ID: "cosmos", Type: types.OracleSourceTypeCosmosOracle, Enabled: true, Priority: 1},
		{ID: "band", Type: types.OracleSourceTypeBandIBC, Enabled: true, Priority: 2},
		{ID: "chainlink", Type: types.OracleSourceTypeChainlinkIBC, Enabled: true, Priority: 3},
	}, 2, 300)

	price, err := s.keeper.AggregatePrice(s.ctx, types.CurrencyPair{Base: "VRT", Quote: "USD"})
	require.NoError(t, err)
	require.True(t, price.Rate.GT(sdkmath.LegacyZeroDec()))
}

func (s *KeeperTestSuite) oracleManualOverride() {
	t := s.T()
	now := s.ctx.BlockTime()

	params := s.keeper.GetParams(s.ctx)
	params.OracleSources = []types.OracleSourceConfig{
		{ID: "manual", Type: types.OracleSourceTypeManual, Enabled: true, Priority: 1},
	}
	params.OracleManualPrices = []types.ManualPriceOverride{
		{
			Base:      "VRT",
			Quote:     "USD",
			Rate:      sdkmath.LegacyMustNewDecFromStr("1.25"),
			UpdatedAt: now,
			ExpiresAt: now.Add(10 * time.Minute),
		},
	}
	params.OracleMinSources = 1
	params.OracleStalenessThresholdSeconds = 300
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	price, err := s.keeper.AggregatePrice(s.ctx, types.CurrencyPair{Base: "VRT", Quote: "USD"})
	require.NoError(t, err)
	require.True(t, price.Rate.Equal(sdkmath.LegacyMustNewDecFromStr("1.25")))
}

func (s *KeeperTestSuite) settlementRateCalculatorLocksLiveRates() {
	t := s.T()
	now := s.ctx.BlockTime()

	s.keeper.SetPriceFeed(types.OracleSourceTypeCosmosOracle, mockPriceFeed{
		prices: map[string]types.Price{
			pairKey("VRT", "USD"): priceFor("USD", "1.00", now, "cosmos"),
			pairKey("VRT", "EUR"): priceFor("EUR", "0.90", now, "cosmos"),
			pairKey("VRT", "GBP"): priceFor("GBP", "0.80", now, "cosmos"),
		},
	})

	params := s.keeper.GetParams(s.ctx)
	params.OracleSources = []types.OracleSourceConfig{
		{ID: "cosmos", Type: types.OracleSourceTypeCosmosOracle, Enabled: true, Priority: 1},
	}
	params.OracleMinSources = 1
	params.OracleStalenessThresholdSeconds = 300
	params.FiatConversionSpreadBps = 100
	require.NoError(t, s.keeper.SetParams(s.ctx, params))

	lock, err := s.keeper.LockSettlementRates(s.ctx, "settlement-1", "invoice-1")
	require.NoError(t, err)
	require.Equal(t, types.SettlementRateStatusLocked, lock.Status)
	require.Len(t, lock.Rates, 3)
	for _, rate := range lock.Rates {
		require.True(t, rate.FinalRate.GT(rate.RawRate))
	}

	stored, found := s.keeper.GetSettlementRateLock(s.ctx, "settlement-1")
	require.True(t, found)
	require.Equal(t, lock.Status, stored.Status)
}

func TestOraclePriceFetchWithMockProviders(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	suite.oraclePriceFetchWithMockProviders()
}

func TestOraclePriceAggregationMedianOutlier(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	suite.oraclePriceAggregationMedianOutlier()
}

func TestOraclePriceStalenessRejection(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	suite.oraclePriceStalenessRejection()
}

func TestOracleFallbackChain(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	suite.oracleFallbackChain()
}

func TestOracleManualOverride(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	suite.oracleManualOverride()
}

func TestOracleSettlementRateCalculatorLocksLiveRates(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	suite.settlementRateCalculatorLocksLiveRates()
}
