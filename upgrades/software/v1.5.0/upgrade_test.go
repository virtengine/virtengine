// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package v1_5_0

import (
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	apptypes "github.com/virtengine/virtengine/app/types"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

func TestUpgradeRegisteredAndDoesNotChangeStores(t *testing.T) {
	registered, ok := utypes.GetUpgradesList()[UpgradeName]
	require.True(t, ok)
	require.NotNil(t, registered)

	up, err := initUpgrade(log.NewNopLogger(), &apptypes.App{})
	require.NoError(t, err)
	require.Empty(t, up.StoreLoader().Added)
	require.Empty(t, up.StoreLoader().Renamed)
	require.Empty(t, up.StoreLoader().Deleted)
	require.Equal(t, "v1.5.0", UpgradeName)
}
