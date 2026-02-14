package testutil

import (
	"math/rand" //nolint:gosec // G404: test helpers use weak random for non-security data
	"testing"

	types "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
)

func RandRangeInt(minVal, maxVal int) int {
	return rand.Intn(maxVal-minVal) + minVal //nolint:gosec // G404: test helper randomness is non-security-sensitive
}

func RandRangeUint(minVal, maxVal uint) uint {
	val := rand.Uint64() //nolint:gosec // G404: test helper randomness is non-security-sensitive
	val %= uint64(maxVal - minVal)
	val += uint64(minVal)
	return uint(val)
}

func RandRangeUint64(minVal, maxVal uint64) uint64 {
	val := rand.Uint64() //nolint:gosec // G404: test helper randomness is non-security-sensitive
	val %= maxVal - minVal
	val += minVal
	return val
}

func ResourceUnits(_ testing.TB) types.Resources {
	return types.Resources{
		ID: 1,
		CPU: &types.CPU{
			Units: types.NewResourceValue(uint64(RandCPUUnits())),
		},
		Memory: &types.Memory{
			Quantity: types.NewResourceValue(RandMemoryQuantity()),
		},
		GPU: &types.GPU{
			Units: types.NewResourceValue(uint64(RandGPUUnits())),
		},
		Storage: types.Volumes{
			types.Storage{
				Quantity: types.NewResourceValue(RandStorageQuantity()),
			},
		},
	}
}
