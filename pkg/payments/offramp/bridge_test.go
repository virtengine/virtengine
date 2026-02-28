package offramp

import (
	"context"
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func testQuoteRequest() QuoteRequest {
	return QuoteRequest{
		CryptoSymbol:  "USDC",
		CryptoDenom:   "uusdc",
		CryptoAmount:  sdkmath.NewInt(1_000_000),
		FiatCurrency:  "USD",
		PaymentMethod: "bank_transfer",
		Sender:        "provider-1",
		Destination:   "destination-ref",
	}
}

func TestBridgeGetQuoteSelectsBestAvailableProvider(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)

	alpha := newContractProvider("alpha", clock.Now, Quote{
		ID:           "alpha-quote",
		FiatAmount:   sdkmath.LegacyNewDec(100),
		ExchangeRate: sdkmath.LegacyNewDec(1),
		Fee:          sdkmath.NewInt(10),
	})
	bravo := newContractProvider("bravo", clock.Now, Quote{
		ID:           "bravo-quote",
		FiatAmount:   sdkmath.LegacyNewDec(110),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(11, 1),
		Fee:          sdkmath.NewInt(5),
	})

	require.NoError(t, bridge.RegisterAdapter(alpha))
	require.NoError(t, bridge.RegisterAdapter(bravo))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)
	require.Equal(t, "bravo", quote.Provider)
	require.Equal(t, "2", quote.AuditFields["bridge_candidate_count"])
	require.Equal(t, "1", quote.AuditFields["bridge_selected_rank"])
	require.Equal(t, "1", quote.AuditFields["bridge_candidate_rank"])
}

func TestBridgeInitiatePayoutIdempotentByMetadata(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 5, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)
	provider := newContractProvider("contract", clock.Now, Quote{
		ID:           "contract-quote",
		FiatAmount:   sdkmath.LegacyNewDec(105),
		ExchangeRate: sdkmath.LegacyNewDec(1),
		Fee:          sdkmath.NewInt(2),
	})
	provider.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "contract-payout-1",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(105),
			},
		},
	}

	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)

	metadata := map[string]string{
		"conversion_id":   "conv-1",
		"idempotency_key": "idem-1",
	}

	first, err := bridge.InitiatePayout(context.Background(), quote, "swap-tx-1", "destination-ref", metadata)
	require.NoError(t, err)

	second, err := bridge.InitiatePayout(context.Background(), quote, "swap-tx-1", "destination-ref", metadata)
	require.NoError(t, err)

	require.Equal(t, first.ID, second.ID)
	require.Equal(t, 1, provider.initCalls)
}

func TestBridgeStoresFinanceReconciliationEvidence(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 7, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)
	provider := newContractProvider("finance", clock.Now, Quote{
		ID:           "finance-quote",
		FiatAmount:   sdkmath.LegacyNewDec(107),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(107, 2),
		Fee:          sdkmath.NewInt(2),
	})
	provider.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "finance-payout-1",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(107),
				Reference:  "finance-ref-1",
			},
		},
	}
	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)

	metadata := map[string]string{
		"conversion_id":   "conv-finance",
		"idempotency_key": "idem-finance",
		"invoice_id":      "invoice-finance",
		"settlement_id":   "settlement-finance",
		"payout_id":       "payout-finance",
	}
	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-finance", "destination-ref", metadata)
	require.NoError(t, err)

	require.Equal(t, metadata, result.Metadata)
	require.Equal(t, quote.ID, result.QuoteID)
	require.Equal(t, "1", result.AuditFields["bridge_attempt"])
	require.Equal(t, "finance", result.AuditFields["bridge_selected_provider"])
	require.Equal(t, "finance", result.AuditFields["bridge_quote_provider"])
	require.False(t, result.StatusUpdatedAt.IsZero())

	resolved, err := bridge.FindPayoutByMetadata(context.Background(), "finance", metadata)
	require.NoError(t, err)
	require.Equal(t, result.ID, resolved.ID)
	require.Equal(t, result.Reference, resolved.Reference)
	require.Equal(t, metadata, resolved.Metadata)
}

func TestBridgeInitiatePayoutFailsOverOnRetryableError(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 10, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)

	primary := newContractProvider("primary", clock.Now, Quote{
		ID:           "primary-quote",
		FiatAmount:   sdkmath.LegacyNewDec(110),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(11, 1),
		Fee:          sdkmath.NewInt(4),
	})
	primary.initSteps = []providerStep{
		{err: errors.New("503 service unavailable")},
	}

	secondary := newContractProvider("secondary", clock.Now, Quote{
		ID:           "secondary-quote",
		FiatAmount:   sdkmath.LegacyNewDec(108),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(108, 2),
		Fee:          sdkmath.NewInt(3),
	})
	secondary.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "secondary-payout-1",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(108),
			},
		},
	}

	require.NoError(t, bridge.RegisterAdapter(primary))
	require.NoError(t, bridge.RegisterAdapter(secondary))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)
	require.Equal(t, "primary", quote.Provider)

	metadata := map[string]string{
		"conversion_id":   "conv-2",
		"idempotency_key": "idem-2",
	}
	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-tx-2", "destination-ref", metadata)
	require.NoError(t, err)

	require.Equal(t, "secondary", result.Provider)
	require.Equal(t, 1, primary.initCalls)
	require.Equal(t, 1, secondary.initCalls)
	require.Equal(t, "2", result.AuditFields["bridge_attempt"])
}

func TestBridgeInitiatePayoutRecoversAmbiguousProviderResult(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 15, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)

	primary := newContractProvider("primary", clock.Now, Quote{
		ID:           "primary-quote",
		FiatAmount:   sdkmath.LegacyNewDec(120),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(12, 1),
		Fee:          sdkmath.NewInt(4),
	})
	primary.initSteps = []providerStep{
		{err: errors.New("timeout waiting for partner")},
	}

	lookupMetadata := map[string]string{
		"conversion_id":   "conv-3",
		"idempotency_key": "idem-3",
	}
	primary.lookup[metadataLookupKey("primary", lookupMetadata)] = PayoutResult{
		ID:              "primary-payout-existing",
		Status:          StatusProcessing,
		Provider:        "primary",
		Reference:       "primary-ref",
		FiatAmount:      sdkmath.LegacyNewDec(120),
		CryptoAmount:    sdkmath.NewInt(1_000_000),
		Metadata:        cloneStringMap(lookupMetadata),
		InitiatedAt:     clock.Now(),
		StatusUpdatedAt: clock.Now(),
	}

	secondary := newContractProvider("secondary", clock.Now, Quote{
		ID:           "secondary-quote",
		FiatAmount:   sdkmath.LegacyNewDec(118),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(118, 2),
		Fee:          sdkmath.NewInt(2),
	})
	secondary.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "secondary-payout-should-not-run",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(118),
			},
		},
	}

	require.NoError(t, bridge.RegisterAdapter(primary))
	require.NoError(t, bridge.RegisterAdapter(secondary))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)

	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-tx-3", "destination-ref", lookupMetadata)
	require.NoError(t, err)

	require.Equal(t, "primary-payout-existing", result.ID)
	require.Equal(t, "primary", result.Provider)
	require.Equal(t, 1, primary.lookupCalls)
	require.Zero(t, secondary.initCalls)
	require.Equal(t, "metadata_lookup", result.AuditFields["bridge_recovery_reason"])
}

func TestBridgeFindPayoutByMetadataUsesCachedOperation(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 20, 0, 0, time.UTC))
	ctx := context.Background()
	bridge := newBridgeWithClock(clock.Now)
	provider := newMockProviderWithClock("mock", []string{"USD"}, []string{"bank_transfer"}, clock.Now)
	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(ctx, testQuoteRequest())
	require.NoError(t, err)

	metadata := map[string]string{
		"conversion_id":   "conv-4",
		"idempotency_key": "idem-4",
	}
	created, err := bridge.InitiatePayout(ctx, quote, "swap-tx-4", "destination-ref", metadata)
	require.NoError(t, err)

	resolved, err := bridge.FindPayoutByMetadata(ctx, "mock", metadata)
	require.NoError(t, err)
	require.Equal(t, created.ID, resolved.ID)
	require.Equal(t, metadata, resolved.Metadata)
}

func TestBridgeFindPayoutByMetadataFallsBackToProviderLookup(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 25, 0, 0, time.UTC))
	ctx := context.Background()
	bridge := newBridgeWithClock(clock.Now)
	provider := newMockProviderWithClock("mock", []string{"USD"}, []string{"bank_transfer"}, clock.Now)
	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := provider.GetQuote(ctx, testQuoteRequest())
	require.NoError(t, err)

	metadata := map[string]string{
		"conversion_id":   "conv-5",
		"idempotency_key": "idem-5",
	}
	created, err := provider.InitiatePayout(ctx, PayoutRequest{
		Quote:       quote,
		CryptoTxRef: "swap-tx-5",
		Destination: "destination-ref",
		Metadata:    metadata,
	})
	require.NoError(t, err)

	resolved, err := bridge.FindPayoutByMetadata(ctx, "", metadata)
	require.NoError(t, err)
	require.Equal(t, created.ID, resolved.ID)
}

func TestBridgeGetStatusKeepsCachedPayoutOnRefreshFailure(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)
	provider := newContractProvider("status-provider", clock.Now, Quote{
		ID:           "status-quote",
		FiatAmount:   sdkmath.LegacyNewDec(100),
		ExchangeRate: sdkmath.LegacyOneDec(),
		Fee:          sdkmath.NewInt(1),
	})
	provider.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "status-payout",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(100),
			},
		},
	}
	provider.statusSteps["status-payout"] = []PayoutResult{}

	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)

	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-status", "destination-ref", map[string]string{
		"conversion_id":   "conv-6",
		"idempotency_key": "idem-6",
	})
	require.NoError(t, err)
	require.Equal(t, StatusProcessing, result.Status)

	provider.mu.Lock()
	provider.statusErr = errors.New("temporary network failure")
	provider.mu.Unlock()

	status, err := bridge.GetStatus(context.Background(), result.ID)
	require.NoError(t, err)
	require.Equal(t, StatusProcessing, status.Status)
	require.Equal(t, "stale", status.AuditFields["bridge_status_refresh"])
}

func TestBridgeGetStatusPublishesCompletionEvidence(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 32, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)
	provider := newContractProvider("complete-provider", clock.Now, Quote{
		ID:           "complete-quote",
		FiatAmount:   sdkmath.LegacyNewDec(111),
		ExchangeRate: sdkmath.LegacyNewDecWithPrec(111, 2),
		Fee:          sdkmath.NewInt(1),
	})
	provider.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "complete-payout",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(111),
				Reference:  "complete-ref",
			},
		},
	}
	provider.statusSteps["complete-payout"] = []PayoutResult{
		{
			ID:         "complete-payout",
			Status:     StatusCompleted,
			Provider:   "complete-provider",
			Reference:  "complete-ref",
			FiatAmount: sdkmath.LegacyNewDec(111),
		},
	}
	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)

	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-complete", "destination-ref", map[string]string{
		"conversion_id":   "conv-complete",
		"idempotency_key": "idem-complete",
	})
	require.NoError(t, err)

	clock.Advance(time.Minute)
	status, err := bridge.GetStatus(context.Background(), result.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, status.Status)
	require.NotNil(t, status.CompletedAt)
	require.Equal(t, "provider", status.AuditFields["bridge_status_source"])
	require.False(t, status.StatusUpdatedAt.IsZero())
}

func TestBridgeCancelMarksPendingPayoutCancelled(t *testing.T) {
	t.Parallel()

	clock := newTestClock(time.Date(2026, 4, 11, 10, 35, 0, 0, time.UTC))
	bridge := newBridgeWithClock(clock.Now)
	provider := newContractProvider("cancel-provider", clock.Now, Quote{
		ID:           "cancel-quote",
		FiatAmount:   sdkmath.LegacyNewDec(101),
		ExchangeRate: sdkmath.LegacyOneDec(),
		Fee:          sdkmath.NewInt(1),
	})
	provider.initSteps = []providerStep{
		{
			result: PayoutResult{
				ID:         "cancel-payout",
				Status:     StatusProcessing,
				FiatAmount: sdkmath.LegacyNewDec(101),
			},
		},
	}

	require.NoError(t, bridge.RegisterAdapter(provider))

	quote, err := bridge.GetQuote(context.Background(), testQuoteRequest())
	require.NoError(t, err)
	result, err := bridge.InitiatePayout(context.Background(), quote, "swap-cancel", "destination-ref", map[string]string{
		"conversion_id":   "conv-7",
		"idempotency_key": "idem-7",
	})
	require.NoError(t, err)

	require.NoError(t, bridge.Cancel(context.Background(), result.ID))

	updated, err := bridge.GetStatus(context.Background(), result.ID)
	require.NoError(t, err)
	require.Equal(t, StatusCancelled, updated.Status)
	require.NoError(t, bridge.Cancel(context.Background(), result.ID))
}
