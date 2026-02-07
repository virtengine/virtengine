package payment

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
)

func TestDeterministicQuoteID(t *testing.T) {
	rateTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	rate := ConversionRate{
		FromCurrency: CurrencyUSD,
		ToCrypto:     "uve",
		Rate:         sdkmath.LegacyMustNewDecFromStr("2.5"),
		Timestamp:    rateTime,
		Source:       "coingecko-primary",
		Strategy:     "primary",
		SourceAttribution: []RateSourceAttribution{
			{
				Source:     "coingecko-primary",
				BaseAsset:  "uve",
				QuoteAsset: "usd",
				Price:      sdkmath.LegacyMustNewDecFromStr("0.4"),
				Timestamp:  rateTime,
				Confidence: 1.0,
			},
		},
	}

	req := ConversionQuoteRequest{
		FiatAmount:         NewAmount(10000, CurrencyUSD),
		CryptoDenom:        "uve",
		DestinationAddress: "virtengine1xyz",
	}

	fee := Amount{Value: 100, Currency: CurrencyUSD}
	cryptoAmount := sdkmath.NewInt(1000)
	expiresAt := rateTime.Add(60 * time.Second)

	first := deterministicQuoteID(req, rate, fee, cryptoAmount, expiresAt)
	second := deterministicQuoteID(req, rate, fee, cryptoAmount, expiresAt)

	if first != second {
		t.Fatalf("expected deterministic quote id, got %s and %s", first, second)
	}

	rate.Source = "chainlink-primary"
	changed := deterministicQuoteID(req, rate, fee, cryptoAmount, expiresAt)
	if first == changed {
		t.Fatalf("expected quote id to change when source changes")
	}
}
