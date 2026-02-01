package util

import (
	"math"

	sdkmath "cosmossdk.io/math"

	rtypes "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
)

func ComputeCommittedResources(factor float64, rv rtypes.ResourceValue) rtypes.ResourceValue {
	// If the value is less than 1, commit the original value. There is no concept of under-commit
	if factor <= 1.0 {
		return rv
	}

	v := rv.Val.Uint64()
	fraction := 1.0 / factor
	committedValue := math.Round(float64(v) * fraction)

	// Don't return a value of zero, since this is used as a resource request
	if committedValue <= 0 {
		committedValue = 1
	}

	result := rtypes.ResourceValue{
		Val: sdkmath.NewInt(int64(committedValue)),
	}

	return result
}

