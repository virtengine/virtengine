package scenarios

import "github.com/virtengine/virtengine/sim/core"

// BullMarketConfig returns a high-demand scenario.
func BullMarketConfig() core.Config {
	cfg := baseConfig()
	cfg.ScenarioName = "bull_market"
	cfg.UserDemandMean = 18
	cfg.UserDemandStdDev = 5
	cfg.NumUsers = 350
	cfg.TokenPriceUSD = 2.75
	cfg.NumAttackers = 1
	cfg.Market.ComputeBasePrice *= 1.4
	cfg.Market.GPUBasePrice *= 1.5
	return cfg
}
