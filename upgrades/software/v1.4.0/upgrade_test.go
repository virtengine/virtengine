// Package v1_4_0 tests the Task 84A activation-only upgrade.
//
//nolint:revive
package v1_4_0

import (
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

func TestUpgradeRegisteredAndDoesNotChangeStores(t *testing.T) {
	t.Parallel()

	registered, ok := utypes.GetUpgradesList()[UpgradeName]
	require.True(t, ok)
	require.NotNil(t, registered)

	up, err := initUpgrade(log.NewNopLogger(), &apptypes.App{})
	require.NoError(t, err)
	require.Empty(t, up.StoreLoader().Added)
	require.Empty(t, up.StoreLoader().Renamed)
	require.Empty(t, up.StoreLoader().Deleted)
}

func TestVoteExtensionActivationHeightIsHPlusOne(t *testing.T) {
	t.Parallel()

	height, err := voteExtensionActivationHeight(0, 100)
	require.NoError(t, err)
	require.Equal(t, int64(101), height)

	height, err = voteExtensionActivationHeight(101, 100)
	require.NoError(t, err)
	require.Equal(t, int64(101), height)

	_, err = voteExtensionActivationHeight(102, 100)
	require.Error(t, err)
}
