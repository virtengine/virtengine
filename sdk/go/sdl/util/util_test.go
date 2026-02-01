package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	rtypes "github.com/virtengine/virtengine/sdk/go/node/types/resources/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/sdl/util"
)

func TestComputeCommittedResources(t *testing.T) {
	rv := rtypes.NewResourceValue(100)
	// Negative factor returns original value
	require.Equal(t, uint64(100), util.ComputeCommittedResources(-1.0, rv).Val.Uint64())

	// Zero factor returns original value
	require.Equal(t, uint64(100), util.ComputeCommittedResources(0.0, rv).Val.Uint64())

	// Factor of one returns the original value
	require.Equal(t, uint64(100), util.ComputeCommittedResources(1.0, rv).Val.Uint64())

	require.Equal(t, uint64(50), util.ComputeCommittedResources(2.0, rv).Val.Uint64())

	require.Equal(t, uint64(33), util.ComputeCommittedResources(3.0, rv).Val.Uint64())

	// Even for huge overcommit values, zero is not returned
	require.Equal(t, uint64(1), util.ComputeCommittedResources(10000.0, rv).Val.Uint64())
}

