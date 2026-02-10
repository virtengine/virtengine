# Price Feed Failure Fallback Behavior

This document describes the failure fallback behavior for the price feed system used in fiat-to-crypto conversions.

## Overview

The price feed system (`pkg/pricefeed`) implements multiple layers of resilience to handle price source failures gracefully while maintaining system integrity.

## Fallback Hierarchy

When fetching a price, the system uses the following fallback hierarchy:

```
1. Primary Source → 2. Secondary Sources → 3. Cached Price → 4. Stale Cache → 5. Error
```

### 1. Primary Source (Strategy: Primary)

The system queries configured sources in priority order. When using `StrategyPrimary`:

- Sources are ordered by `Priority` field (lower = higher priority)
- First healthy source to return a valid price wins
- Unhealthy sources (circuit open, rate limited) are skipped

### 2. Secondary Sources (Automatic Fallback)

If the primary source fails:
- **Rate limited**: Tries next source immediately
- **Timeout/Network error**: Retries with exponential backoff, then tries next source
- **Invalid data**: Skips to next source

### 3. Cached Price

If all live sources fail but cache is enabled:
- Returns cached price if within TTL (default: 30 seconds)
- Cache is populated on every successful price fetch

### 4. Stale Cache (Optional)

If `AllowStale: true` in cache config:
- Returns stale price if within `StaleMaxAge` (default: 5 minutes)
- Stale prices are marked with increased age
- Useful for brief outages

### 5. Error Response

If all fallbacks fail:
- Returns `ErrAllSourcesFailed` or `ErrPriceFeedUnavailable`
- Payment service rejects the conversion request
- User sees appropriate error message

## Circuit Breaker

Each price source has an independent circuit breaker:

| State | Behavior |
|-------|----------|
| **Closed** | Normal operation, requests allowed |
| **Open** | Source failed repeatedly, requests blocked for timeout period (30s) |
| **Half-Open** | After timeout, allows limited requests to test recovery |

Configuration:
- `FailureThreshold`: 5 consecutive failures opens circuit
- `SuccessThreshold`: 2 consecutive successes closes circuit
- `Timeout`: 30 seconds before half-open transition

## Rate Limiting

### CoinGecko
- Free tier: 10-30 requests/minute
- Pro tier: Higher limits with API key
- Rate limit errors trigger fallback to next source

### Chainlink
- No explicit rate limits (on-chain reads)
- Subject to RPC provider limits
- Recommended: Use dedicated RPC endpoint

### Pyth
- High throughput (Hermes API)
- No aggressive rate limiting
- Sub-second price updates available

## Retry Configuration

```go
RetryConfig{
    MaxRetries:    3,           // Maximum retry attempts
    InitialDelay:  100ms,       // First retry delay
    MaxDelay:      5s,          // Maximum retry delay
    BackoffFactor: 2.0,         // Exponential backoff multiplier
    RetryableErrors: []string{  // Errors that trigger retry
        "timeout",
        "connection refused",
        "rate limit",
        "503", "504",
    },
}
```

## Aggregation Strategies

### StrategyPrimary (Default)
- Uses first healthy source in priority order
- Fastest response time
- Best for low-latency requirements

### StrategyMedian
- Queries all sources in parallel
- Returns median price
- Protects against single-source manipulation
- Rejects if deviation > `MaxPriceDeviation` (default: 5%)

### StrategyWeighted
- Queries all sources in parallel
- Weighted average by source confidence/volume
- Best accuracy for large transactions

## Monitoring

### Metrics Exposed

| Metric | Description |
|--------|-------------|
| `pricefeed_fetch_total` | Total price fetch requests by source |
| `pricefeed_fetch_errors` | Failed requests by source |
| `pricefeed_fetch_latency` | Request latency by source |
| `pricefeed_cache_hits` | Cache hit count |
| `pricefeed_cache_misses` | Cache miss count |
| `pricefeed_price_deviation` | Price deviation between sources |
| `pricefeed_source_healthy` | Source health status |

### Health Checks

```go
sources := aggregator.ListSources()
for _, source := range sources {
    fmt.Printf("%s: healthy=%v, error_rate=%.2f%%\n",
        source.Source, source.Healthy, source.ErrorRate())
}
```

### Alerts (Recommended)

1. **All sources unhealthy**: Critical - conversion service degraded
2. **Price deviation > 5%**: Warning - potential price manipulation
3. **Cache hit rate < 50%**: Info - high API load
4. **Single source error rate > 20%**: Warning - source degraded

## Configuration Example

```go
conversionConfig := ConversionConfig{
    Enabled:              true,
    CryptoDenom:          "uve",
    PriceFeedSource:      "coingecko",  // or "chainlink", "pyth", "median", "weighted"
    ConversionFeePercent: 1.5,
    QuoteValiditySeconds: 60,
    MinSlippagePercent:   0.5,
    
    // Optional advanced settings
    CoinGeckoAPIKey:   "your-api-key",        // For Pro tier
    ChainlinkRPCURL:   "https://eth.rpc.url", // For Chainlink
    PythHermesURL:     "https://hermes.pyth.network",
    CacheTTLSeconds:   30,
    MaxPriceDeviation: 0.05,  // 5%
}
```

## Failure Scenarios

### Scenario 1: CoinGecko Rate Limited

```
Request → CoinGecko (429) → Chainlink (success) → Return price
```

### Scenario 2: All Sources Down

```
Request → CoinGecko (timeout) → Chainlink (timeout) → Pyth (timeout)
        → Check cache (hit, age=20s) → Return cached price
```

### Scenario 3: Stale Cache Only

```
Request → All sources fail → Cache miss → Stale cache (age=2min)
        → AllowStale=true → Return stale price with warning
```

### Scenario 4: Complete Failure

```
Request → All sources fail → Cache miss → Stale cache expired
        → Return ErrAllSourcesFailed → Reject conversion
```

## Best Practices

1. **Configure multiple sources**: Enable at least 2 sources for redundancy
2. **Use appropriate strategy**: Primary for speed, Median for accuracy
3. **Set reasonable cache TTL**: 30s is good balance for crypto volatility
4. **Monitor metrics**: Alert on source health and deviation
5. **Test fallbacks**: Periodically simulate source failures
6. **Use API keys**: Unlock higher rate limits for production

## Troubleshooting

### "price feed unavailable" Error

1. Check if price feed is initialized: `cfg.ConversionConfig.Enabled`
2. Verify network connectivity to price sources
3. Check for rate limiting (too many requests)
4. Verify RPC endpoints for Chainlink/Pyth

### Stale Prices Being Returned

1. Check cache configuration: `CacheTTLSeconds`
2. Verify source health: `aggregator.ListSources()`
3. Check for rate limiting issues

### High Price Deviation

1. Normal during volatile markets
2. If persistent, check for source misconfiguration
3. Consider switching to single trusted source temporarily
