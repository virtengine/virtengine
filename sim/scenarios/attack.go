package scenarios

import "github.com/virtengine/virtengine/sim/core"

// AttackConfig returns a scenario focused on adversarial behavior.
func AttackConfig() core.Config {
	cfg := baseConfig()
	cfg.ScenarioName = "attack_scenario"
	cfg.UserDemandMean = 9
	cfg.NumAttackers = 6
	cfg.TokenPriceUSD = 1.1
	cfg.Market.FeeBurnBPS = 1000
	return cfg
}
