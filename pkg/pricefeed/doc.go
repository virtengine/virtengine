// Package pricefeed provides real-time price feed integration for cryptocurrency
// price discovery.
//
// # Overview
//
// This package implements a multi-source price feed system with support for:
//   - CoinGecko API (centralized, free tier available)
//   - Chainlink price feeds (decentralized oracles)
//   - Pyth Network (high-frequency decentralized oracles)
//
// # Architecture
//
// The package uses an aggregator pattern where multiple price sources are queried
// and their results are aggregated with configurable strategies:
//   - Primary: Use first healthy source in priority order
//   - Median: Use median price across all sources
//   - Weighted: Weight by source confidence/liquidity
//
// # Caching
//
// All price feeds are cached with configurable TTL to reduce API calls and improve
// response times. Cache staleness is tracked for monitoring.
//
// # Retry Logic
//
// Failed requests are retried with exponential backoff. Circuit breakers prevent
// cascading failures when a source is unavailable.
//
// # Failure Fallback
//
// When all sources fail, the aggregator returns an error. Callers should:
//  1. Check cache for stale-but-valid prices (if acceptable for use case)
//  2. Use last known good price with increased slippage tolerance
//  3. Reject the conversion request with appropriate user messaging
//
// # Monitoring
//
// The package exposes Prometheus metrics for:
//   - Price fetch latency by source
//   - Cache hit/miss rates
//   - Error rates and types
//   - Price deviation between sources
//
// # Usage
//
//	cfg := pricefeed.DefaultConfig()
//	agg, err := pricefeed.NewAggregator(cfg)
//	if err != nil {
//	    return err
//	}
//	defer agg.Close()
//
//	price, err := agg.GetPrice(ctx, "virtengine", "usd")
//	if err != nil {
//	    // Handle failure - see Failure Fallback section
//	    return err
//	}
//
//	// Use price.Rate, price.Timestamp, price.Source
//
// PAY-001: Real price feed integration for payment conversions
package pricefeed
