package keeper

import (
	"math"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

func (k Keeper) oracleSources(ctx sdk.Context) []types.OracleSourceConfig {
	params := k.GetParams(ctx)
	sources := make([]types.OracleSourceConfig, 0, len(params.OracleSources))
	for _, source := range params.OracleSources {
		if source.Enabled {
			sources = append(sources, source)
		}
	}
	sort.Slice(sources, func(i, j int) bool {
		if sources[i].Priority == sources[j].Priority {
			return sources[i].ID < sources[j].ID
		}
		return sources[i].Priority < sources[j].Priority
	})
	return sources
}

func (k Keeper) oracleStalenessThreshold(ctx sdk.Context) time.Duration {
	params := k.GetParams(ctx)
	return safeOracleDurationFromSeconds(params.OracleStalenessThresholdSeconds, 5*time.Minute)
}

func (k Keeper) oracleMinSources(ctx sdk.Context) int {
	params := k.GetParams(ctx)
	if params.OracleMinSources == 0 {
		return 1
	}
	return int(params.OracleMinSources)
}

func (k Keeper) oracleDeviationThresholdBps(ctx sdk.Context) uint32 {
	params := k.GetParams(ctx)
	if params.OracleDeviationThresholdBps == 0 {
		return 500
	}
	return params.OracleDeviationThresholdBps
}

func (k Keeper) oracleDeviationWindow(ctx sdk.Context) time.Duration {
	params := k.GetParams(ctx)
	return safeOracleDurationFromSeconds(params.OracleDeviationWindowSeconds, time.Minute)
}

func (k Keeper) fiatConversionSpreadBps(ctx sdk.Context) uint32 {
	params := k.GetParams(ctx)
	return params.FiatConversionSpreadBps
}

func (k Keeper) oracleManualPrices(ctx sdk.Context) []types.ManualPriceOverride {
	params := k.GetParams(ctx)
	return params.OracleManualPrices
}

func safeOracleDurationFromSeconds(seconds uint64, fallback time.Duration) time.Duration {
	if seconds == 0 {
		return fallback
	}
	maxSeconds := uint64(math.MaxInt64 / int64(time.Second))
	if seconds > maxSeconds {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
