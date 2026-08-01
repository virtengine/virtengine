package v1_6_0

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"

	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	resourcetypes "github.com/virtengine/virtengine/x/resources/types"
)

func TestCompletedRetryVersionMap(t *testing.T) {
	fromVM := module.VersionMap{
		marketv1.ModuleName:         7,
		marketplacetypes.ModuleName: 1,
		resourcetypes.ModuleName:    1,
		hpctypes.ModuleName:         1,
	}

	retried, completed, err := completedRetryVersionMap(fromVM, true, true)
	require.NoError(t, err)
	require.True(t, completed)
	require.Equal(t, uint64(8), retried[marketv1.ModuleName])
	require.Equal(t, uint64(2), retried[marketplacetypes.ModuleName])
	require.Equal(t, uint64(2), retried[resourcetypes.ModuleName])
	require.Equal(t, uint64(2), retried[hpctypes.ModuleName])
	require.Equal(t, uint64(7), fromVM[marketv1.ModuleName], "source version map must not be mutated")
}

func TestCompletedRetryVersionMapRejectsPartialActivation(t *testing.T) {
	_, completed, err := completedRetryVersionMap(module.VersionMap{}, true, false)
	require.ErrorContains(t, err, "canonical activation is partial")
	require.False(t, completed)
}

func TestCompletedRetryVersionMapAllowsInitialMigration(t *testing.T) {
	for _, resourcesActive := range []bool{false, true} {
		retried, completed, err := completedRetryVersionMap(module.VersionMap{}, false, resourcesActive)
		require.NoError(t, err)
		require.False(t, completed)
		require.Nil(t, retried)
	}
}
