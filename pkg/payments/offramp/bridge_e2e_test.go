//go:build e2e.integration

package offramp

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestBridgeRealProviderLifecycleE2E(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 11, 15, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)

	realProvider := newContractProvider("real-partner", clock.Now, Quote{
		ID:           "real-quote",
		FiatAmount:   sdkmath.LegacyNewDec(125),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(125, 2),
		Fee:          sdkmath.NewInt(4),
	})
	realProvider.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:           "real-payout-1",
				Status:       StatusProcessing,
				Provider:     "real-partner",
				Reference:    "partner-ref-1",
				FiatAmount:   sdkmath.LegacyNewDec(125),
				CryptoAmount: sdkmath.NewInt(1_000_000),
				Metadata: map[string]string{
					"conversion_id":   "conv-e2e-1",
					"idempotency_key": "idem-e2e-1",
				},
				InitiatedAt:     clock.Now(),
				StatusUpdatedAt: clock.Now(),
			},
			err: errors.New("timeout while waiting for partner acknowledgement"),
		},
	}
	realProvider.statusSteps["real-payout-1"] = []PayoutResult{
		{
			ID:              "real-payout-1",
			Status:          StatusProcessing,
			Provider:        "real-partner",
			Reference:       "partner-ref-1",
			FiatAmount:      sdkmath.LegacyNewDec(125),
			CryptoAmount:    sdkmath.NewInt(1_000_000),
			StatusUpdatedAt: clock.Now(),
		},
		{
			ID:              "real-payout-1",
			Status:          StatusCompleted,
			Provider:        "real-partner",
			Reference:       "partner-ref-1",
			FiatAmount:      sdkmath.LegacyNewDec(125),
			CryptoAmount:    sdkmath.NewInt(1_000_000),
			StatusUpdatedAt: clock.Now().Add(time.Minute),
		},
	}
	mockFallback := newMockProviderWithClock("mock-fallback", []string{"USD"}, []string{"bank_transfer"}, clock.Now)
	mockFallback.rateByFiat["USD"] = sdkmath.LegacyNewDecWithPrec(1, 6)

	require.NoError(t, bridge.RegisterAdapter(realProvider))
	require.NoError(t, bridge.RegisterAdapter(mockFallback))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)
	require.Equal(t, "real-partner", quote.Provider)

	metadata := map[string]string{
		"conversion_id":   "conv-e2e-1",
		"idempotency_key": "idem-e2e-1",
	}
	started, err := bridge.InitiatePayout(context.Background(), quote, "swap-e2e-1", "destination-ref", metadata)
	require.NoError(t, err)
	require.Equal(t, "real-payout-1", started.ID)
	require.Equal(t, "real-partner", started.Provider)
	require.Equal(t, 1, realProvider.initCalls)

	replayed, err := bridge.InitiatePayout(context.Background(), quote, "swap-e2e-1", "destination-ref", metadata)
	require.NoError(t, err)
	require.Equal(t, started.ID, replayed.ID)
	require.Equal(t, 1, realProvider.initCalls)
	require.Zero(t, mockFallback.payoutSeq)

	firstStatus, err := bridge.GetStatus(context.Background(), started.ID)
	require.NoError(t, err)
	require.Equal(t, StatusProcessing, firstStatus.Status)

	clock.Advance(time.Minute)
	finalStatus, err := bridge.GetStatus(context.Background(), started.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, finalStatus.Status)
	require.NotNil(t, finalStatus.CompletedAt)

	found, err := bridge.FindPayoutByMetadata(context.Background(), "", metadata)
	require.NoError(t, err)
	require.Equal(t, started.ID, found.ID)
	require.Equal(t, StatusCompleted, found.Status)
}
