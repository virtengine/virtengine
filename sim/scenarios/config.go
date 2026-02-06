package scenarios

import (
	"time"

	"github.com/virtengine/virtengine/pkg/economics"
	"github.com/virtengine/virtengine/sim/core"
	"github.com/virtengine/virtengine/sim/markets"
)

func baseConfig() core.Config {
	return core.Config{
		ScenarioName:       "baseline",
		StartTime:          time.Now().UTC(),
		EndTime:            time.Now().UTC().Add(365 * 24 * time.Hour),
		TimeStep:           24 * time.Hour,
		Seed:               42,
		NumUsers:           200,
		NumProviders:       40,
		NumValidators:      60,
		NumAttackers:       2,
		Tokenomics:         economics.DefaultTokenomicsParams(),
		Market:             markets.DefaultMarketParams(),
		UserDemandMean:     10,
		UserDemandStdDev:   2.5,
		ProviderCapacity:   12,
		TokenPriceUSD:      1.25,
		SlashingEnabled:    true,
		SlashingPenaltyBPS: 500,
	}
}
