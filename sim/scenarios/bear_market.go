package scenarios

import "github.com/virtengine/virtengine/sim/core"

// BearMarketConfig returns a low-demand scenario.
func BearMarketConfig() core.Config {
	cfg := baseConfig()
	cfg.ScenarioName = "bear_market"
	cfg.UserDemandMean = 5
	cfg.UserDemandStdDev = 1.8
	cfg.NumUsers = 120
	cfg.TokenPriceUSD = 0.65
	cfg.NumAttackers = 3
	cfg.Market.ComputeBasePrice *= 0.7
	cfg.Market.StorageBasePrice *= 0.8
	return cfg
}
