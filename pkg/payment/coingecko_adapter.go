// Package payment provides payment gateway integration for Visa/Mastercard.
//
// CoinGecko adapter overrides for price feed configuration.
package payment

import "github.com/virtengine/virtengine/pkg/pricefeed"

func applyCoinGeckoConfig(provider *pricefeed.ProviderConfig, convCfg ConversionConfig) {
	if provider.CoinGeckoConfig == nil {
		return
	}

	if convCfg.CoinGeckoAPIKey != "" {
		provider.CoinGeckoConfig.APIKey = convCfg.CoinGeckoAPIKey
		provider.CoinGeckoConfig.UsePro = true
	}
}
