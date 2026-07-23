package v1_6_0

import (
	"github.com/stretchr/testify/require"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	"testing"
)

func TestUpgradeRegistered(t *testing.T) {
	require.Equal(t, "v1.6.0", UpgradeName)
	_, found := utypes.GetUpgradesList()[UpgradeName]
	require.True(t, found)
}
