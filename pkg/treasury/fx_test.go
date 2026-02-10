package treasury

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/pricefeed"
)

type mockFXProvider struct {
	price     sdkmath.LegacyDec
	timestamp time.Time
}

func (m mockFXProvider) GetPrice(ctx context.Context, baseAsset, quoteAsset string) (pricefeed.PriceData, error) {
	return pricefeed.PriceData{
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
		Price:      m.price,
		Timestamp:  m.timestamp,
		Source:     "mock",
		Confidence: 1.0,
	}, nil
}

func TestFXSnapshots(t *testing.T) {
	store := NewInMemorySnapshotStore()

	t1 := time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(2 * time.Hour)

	svc := NewFXService(mockFXProvider{price: sdkmath.LegacyMustNewDecFromStr("1.05"), timestamp: t1}, store)
	snap1, err := svc.GetRate(context.Background(), "UVE", "USD")
	require.NoError(t, err)
	require.Equal(t, t1, snap1.Timestamp)

	svc.provider = mockFXProvider{price: sdkmath.LegacyMustNewDecFromStr("1.10"), timestamp: t2}
	snap2, err := svc.GetRate(context.Background(), "UVE", "USD")
	require.NoError(t, err)
	require.Equal(t, t2, snap2.Timestamp)

	historical, ok := svc.GetHistoricalRate("UVE", "USD", t1.Add(30*time.Minute))
	require.True(t, ok)
	require.Equal(t, snap1.Rate, historical.Rate)

	historical, ok = svc.GetHistoricalRate("UVE", "USD", t2.Add(30*time.Minute))
	require.True(t, ok)
	require.Equal(t, snap2.Rate, historical.Rate)
}
