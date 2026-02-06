// Package payment provides payment gateway integration for Visa/Mastercard.
//
// Price feed integration helpers for fiat-to-crypto conversion.
package payment

import (
	"context"
	"sort"
	"time"

	"github.com/virtengine/virtengine/pkg/pricefeed"
)

// PriceFeedHealth summarizes price feed health for monitoring.
type PriceFeedHealth struct {
	Sources    []pricefeed.SourceHealth `json:"sources"`
	CacheStats pricefeed.CacheStats     `json:"cache_stats,omitempty"`
	Metrics    pricefeed.MetricsSummary `json:"metrics,omitempty"`
}

// PriceFeedMonitor exposes price feed health data for monitoring.
type PriceFeedMonitor interface {
	GetPriceFeedHealth(ctx context.Context) (PriceFeedHealth, error)
}

func buildPriceFeedConfig(cfg Config) pricefeed.Config {
	pfCfg := pricefeed.DefaultConfig()
	convCfg := cfg.ConversionConfig

	// Configure providers based on price feed source
	switch convCfg.PriceFeedSource {
	case "coingecko":
		pfCfg.Strategy = pricefeed.StrategyPrimary
		// CoinGecko is already first priority in default config
	case "chainlink":
		pfCfg.Strategy = pricefeed.StrategyPrimary
		// Move Chainlink to first priority
		for i := range pfCfg.Providers {
			if pfCfg.Providers[i].Type == pricefeed.SourceTypeChainlink {
				pfCfg.Providers[i].Priority = 1
			} else {
				pfCfg.Providers[i].Priority++
			}
		}
	case "pyth":
		pfCfg.Strategy = pricefeed.StrategyPrimary
		// Move Pyth to first priority
		for i := range pfCfg.Providers {
			if pfCfg.Providers[i].Type == pricefeed.SourceTypePyth {
				pfCfg.Providers[i].Priority = 1
			} else {
				pfCfg.Providers[i].Priority++
			}
		}
	case "median":
		pfCfg.Strategy = pricefeed.StrategyMedian
	case "weighted":
		pfCfg.Strategy = pricefeed.StrategyWeighted
	}

	applyCacheConfig(&pfCfg, convCfg)
	applyRetryConfig(&pfCfg, cfg)
	applyProviderOverrides(&pfCfg, convCfg, cfg)

	if convCfg.MaxPriceDeviation > 0 {
		pfCfg.MaxPriceDeviation = convCfg.MaxPriceDeviation
	}

	pfCfg.EnableMetrics = true
	pfCfg.EnableLogging = cfg.EnableLogging

	return pfCfg
}

func applyCacheConfig(cfg *pricefeed.Config, convCfg ConversionConfig) {
	if convCfg.CacheTTLSeconds > 0 {
		cfg.CacheConfig.TTL = time.Duration(convCfg.CacheTTLSeconds) * time.Second
	}
	if cfg.CacheConfig.TTL <= 0 {
		cfg.CacheConfig.TTL = 30 * time.Second
	}
	cfg.CacheConfig.Enabled = true
}

func applyRetryConfig(cfg *pricefeed.Config, svcCfg Config) {
	if svcCfg.RetryMaxAttempts > 0 {
		cfg.RetryConfig.MaxRetries = svcCfg.RetryMaxAttempts
	}
	if svcCfg.RetryInitialDelay > 0 {
		cfg.RetryConfig.InitialDelay = svcCfg.RetryInitialDelay
	}
	if svcCfg.RetryMaxDelay > 0 {
		cfg.RetryConfig.MaxDelay = svcCfg.RetryMaxDelay
	}
	if svcCfg.RetryBackoffFactor > 0 {
		cfg.RetryConfig.BackoffFactor = svcCfg.RetryBackoffFactor
	}
}

func applyProviderOverrides(cfg *pricefeed.Config, convCfg ConversionConfig, svcCfg Config) {
	for i := range cfg.Providers {
		provider := &cfg.Providers[i]
		if svcCfg.RequestTimeout > 0 {
			provider.RequestTimeout = svcCfg.RequestTimeout
		}

		switch provider.Type {
		case pricefeed.SourceTypeCoinGecko:
			applyCoinGeckoConfig(provider, convCfg)
		case pricefeed.SourceTypeChainlink:
			if provider.ChainlinkConfig != nil && convCfg.ChainlinkRPCURL != "" {
				provider.ChainlinkConfig.RPCURL = convCfg.ChainlinkRPCURL
			}
		case pricefeed.SourceTypePyth:
			if provider.PythConfig != nil && convCfg.PythHermesURL != "" {
				provider.PythConfig.HermesURL = convCfg.PythHermesURL
			}
		}
	}
}

func buildRateAttribution(price pricefeed.AggregatedPrice) []RateSourceAttribution {
	if len(price.SourcePrices) == 0 {
		return nil
	}

	attributions := make([]RateSourceAttribution, 0, len(price.SourcePrices))
	for _, sourcePrice := range price.SourcePrices {
		attributions = append(attributions, RateSourceAttribution{
			Source:     sourcePrice.Source,
			BaseAsset:  sourcePrice.BaseAsset,
			QuoteAsset: sourcePrice.QuoteAsset,
			Price:      sourcePrice.Price,
			Timestamp:  sourcePrice.Timestamp,
			Confidence: sourcePrice.Confidence,
		})
	}

	sort.Slice(attributions, func(i, j int) bool {
		if attributions[i].Source == attributions[j].Source {
			if attributions[i].BaseAsset == attributions[j].BaseAsset {
				return attributions[i].QuoteAsset < attributions[j].QuoteAsset
			}
			return attributions[i].BaseAsset < attributions[j].BaseAsset
		}
		return attributions[i].Source < attributions[j].Source
	})

	return attributions
}

func (s *paymentService) GetPriceFeedHealth(ctx context.Context) (PriceFeedHealth, error) {
	if s.priceFeed == nil {
		return PriceFeedHealth{}, pricefeed.ErrPriceFeedUnavailable
	}

	health := PriceFeedHealth{
		Sources: s.priceFeed.ListSources(),
	}

	if agg, ok := s.priceFeed.(*pricefeed.PriceFeedAggregator); ok {
		health.CacheStats = agg.CacheStats()
	}
	if s.priceFeedMetrics != nil {
		health.Metrics = s.priceFeedMetrics.Summary()
	}

	return health, nil
}
