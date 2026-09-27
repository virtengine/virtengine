package types

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
)

func TestCapacityForOrderScalesByQuantity(t *testing.T) {
	order := marketplacetypes.Order{
		RequestedQuantity: 3,
		ResourceUnits:     map[string]uint64{"cpu": 2, "ram": 4},
	}
	candidate := marketplacetypes.Candidate{Capacity: 3}

	capacity, err := capacityForOrder(order, candidate)
	require.NoError(t, err)
	require.Equal(t, int64(6), capacity.CpuCores)
	require.Equal(t, int64(12), capacity.MemoryGb)
}

func TestCapacityForOrderRequiresUnits(t *testing.T) {
	order := marketplacetypes.Order{RequestedQuantity: 1}
	_, err := capacityForOrder(order, marketplacetypes.Candidate{Capacity: 1})
	require.Error(t, err)
}

func TestCapacityForOrderRequiresGPUType(t *testing.T) {
	order := marketplacetypes.Order{
		RequestedQuantity: 1,
		ResourceUnits:     map[string]uint64{"gpu": 1},
	}
	_, err := capacityForOrder(order, marketplacetypes.Candidate{Capacity: 1})
	require.Error(t, err, "gpu capacity without a type must be rejected")

	typed := marketplacetypes.Candidate{Capacity: 1, GPUType: "a100"}
	capacity, err := capacityForOrder(order, typed)
	require.NoError(t, err)
	require.Equal(t, int64(1), capacity.Gpus)
	require.Equal(t, "a100", capacity.GpuType)
}

func TestCapacityForOrderSaturates(t *testing.T) {
	order := marketplacetypes.Order{
		RequestedQuantity: 1,
		ResourceUnits:     map[string]uint64{"cpu": math.MaxUint64},
	}
	capacity, err := capacityForOrder(order, marketplacetypes.Candidate{Capacity: math.MaxUint64})
	require.NoError(t, err)
	require.Equal(t, int64(math.MaxInt64), capacity.CpuCores)
}

func TestResourceClassForCandidate(t *testing.T) {
	require.Equal(t, resourcestypes.ResourceClassStorage, resourceClassForCandidate(marketplacetypes.Candidate{ResourceClass: "storage"}))
	require.Equal(t, resourcestypes.ResourceClassNetwork, resourceClassForCandidate(marketplacetypes.Candidate{ResourceClass: "network"}))
	require.Equal(t, resourcestypes.ResourceClassCompute, resourceClassForCandidate(marketplacetypes.Candidate{}))
}
