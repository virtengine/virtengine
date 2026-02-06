// Package payment provides payment gateway integration for Visa/Mastercard.
//
// Deterministic quote helpers for conversion pricing.
package payment

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
)

func conversionQuoteExpiry(rateTimestamp time.Time, validitySeconds int) time.Time {
	base := rateTimestamp
	if base.IsZero() {
		base = time.Now().UTC()
	}
	if validitySeconds <= 0 {
		validitySeconds = 60
	}
	return base.Add(time.Duration(validitySeconds) * time.Second)
}

func deterministicQuoteID(
	req ConversionQuoteRequest,
	rate ConversionRate,
	fee Amount,
	cryptoAmount sdkmath.Int,
	expiresAt time.Time,
) string {
	components := []string{
		string(req.FiatAmount.Currency),
		formatInt(req.FiatAmount.Value),
		strings.ToLower(req.CryptoDenom),
		strings.ToLower(req.DestinationAddress),
		rate.Rate.String(),
		rate.Timestamp.UTC().Format(time.RFC3339Nano),
		rate.Source,
		rate.Strategy,
		formatInt(fee.Value),
		cryptoAmount.String(),
		expiresAt.UTC().Format(time.RFC3339Nano),
	}

	if len(rate.SourceAttribution) > 0 {
		attribution := append([]RateSourceAttribution(nil), rate.SourceAttribution...)
		sort.Slice(attribution, func(i, j int) bool {
			if attribution[i].Source == attribution[j].Source {
				if attribution[i].BaseAsset == attribution[j].BaseAsset {
					return attribution[i].QuoteAsset < attribution[j].QuoteAsset
				}
				return attribution[i].BaseAsset < attribution[j].BaseAsset
			}
			return attribution[i].Source < attribution[j].Source
		})
		for _, source := range attribution {
			components = append(components,
				source.Source,
				source.BaseAsset,
				source.QuoteAsset,
				source.Price.String(),
				source.Timestamp.UTC().Format(time.RFC3339Nano),
				formatFloat(source.Confidence),
			)
		}
	}

	sum := sha256.Sum256([]byte(strings.Join(components, "|")))
	return "quote_" + hex.EncodeToString(sum[:])
}

func formatInt(value int64) string {
	return sdkmath.NewInt(value).String()
}

func formatFloat(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 6, 64)
	return strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
}
