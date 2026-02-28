package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

func TestBuildUsageComponents_ExcludesQueuePenaltyAsCharge(t *testing.T) {
	denom := "uvirt"
	metrics := types.HPCDetailedMetrics{
		CPUCoreSeconds:   3600,
		MemoryGBSeconds:  7200,
		QueueTimeSeconds: 7200,
	}
	breakdown := types.BillableBreakdown{
		CPUCost:      sdk.NewCoin(denom, sdkmath.NewInt(100)),
		MemoryCost:   sdk.NewCoin(denom, sdkmath.NewInt(50)),
		GPUCost:      sdk.NewCoin(denom, sdkmath.ZeroInt()),
		StorageCost:  sdk.NewCoin(denom, sdkmath.ZeroInt()),
		NetworkCost:  sdk.NewCoin(denom, sdkmath.ZeroInt()),
		NodeCost:     sdk.NewCoin(denom, sdkmath.ZeroInt()),
		QueuePenalty: sdk.NewCoin(denom, sdkmath.NewInt(10)),
	}

	components := buildUsageComponents(metrics, breakdown)
	require.Len(t, components, 2)
	for _, component := range components {
		require.NotEqual(t, "queue_penalty", component.usageType)
	}

	scaled := scaleComponentsToTarget(sdk.NewCoin(denom, sdkmath.NewInt(140)), components)
	total := sdkmath.ZeroInt()
	for _, component := range scaled {
		total = total.Add(component.cost.Amount)
	}
	require.Equal(t, int64(140), total.Int64())
}
