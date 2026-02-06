package scenarios

import "github.com/virtengine/virtengine/sim/core"

// BlackSwanConfig returns an extreme stress scenario.
func BlackSwanConfig() core.Config {
	cfg := baseConfig()
	cfg.ScenarioName = "black_swan"
	cfg.UserDemandMean = 25
	cfg.UserDemandStdDev = 12
	cfg.NumUsers = 600
	cfg.NumAttackers = 8
	cfg.TokenPriceUSD = 0.4
	cfg.Market.PriceAdjustment = 0.3
	cfg.Market.MaxPriceMove = 0.6
	return cfg
}
