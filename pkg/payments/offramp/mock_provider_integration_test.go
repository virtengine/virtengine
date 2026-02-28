//go:build integration

package offramp

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestMockProviderContractIntegration(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC))
	provider := newMockProviderWithClock("mock", []string{"USD"}, []string{"bank_transfer"}, clock.Now)

	quote, err := provider.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)
	require.Equal(t, "mock-quote-001", quote.ID)
	require.Equal(t, "mock", quote.Provider)
	require.Equal(t, "mock", quote.AuditFields["provider_type"])

	metadata := map[string]string{
		"conversion_id":   "conv-int-1",
		"idempotency_key": "idem-int-1",
	}
	first, err := provider.InitiatePayout(context.Background(), PayoutRequest{
		Quote:       quote,
		CryptoTxRef: "swap-int-1",
		Destination: "destination-ref",
		Metadata:    metadata,
	})
	require.NoError(t, err)
	require.Equal(t, "mock-payout-001", first.ID)

	second, err := provider.InitiatePayout(context.Background(), PayoutRequest{
		Quote:       quote,
		CryptoTxRef: "swap-int-1",
		Destination: "destination-ref",
		Metadata:    metadata,
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)

	found, err := provider.FindPayoutByMetadata(context.Background(), metadata)
	require.NoError(t, err)
	require.Equal(t, first.ID, found.ID)

	provider.mu.Lock()
	pending := clonePayoutResult(first)
	pending.Status = StatusProcessing
	pending.CompletedAt = nil
	pending.StatusUpdatedAt = clock.Now()
	provider.payouts[first.ID] = pending
	provider.mu.Unlock()

	require.NoError(t, provider.Cancel(context.Background(), first.ID))

	status, err := provider.GetStatus(context.Background(), first.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCancelled, status.Status)
}

func TestBridgeWithMockProviderIntegration(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 11, 5, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)
	provider := newMockProviderWithClock("mock", []string{"USD"}, []string{"bank_transfer"}, clock.Now)
	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)
	require.Equal(t, sdkmath.LegacyNewDec(1), quote.ExchangeRate)

	metadata := map[string]string{
		"conversion_id":   "conv-int-2",
		"idempotency_key": "idem-int-2",
	}
	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-int-2", "destination-ref", metadata)
	require.NoError(t, err)
	require.Equal(t, "mock", result.Provider)

	resolved, err := bridge.FindPayoutByMetadata(context.Background(), "", metadata)
	require.NoError(t, err)
	require.Equal(t, result.ID, resolved.ID)

	status, err := bridge.GetStatus(context.Background(), result.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, status.Status)
}
