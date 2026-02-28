package upgrade

import (
	"sort"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"

	apptypes "github.com/virtengine/virtengine/app/types"
	_ "github.com/virtengine/virtengine/upgrades"
	v100 "github.com/virtengine/virtengine/upgrades/software/v1.0.0"
	v110 "github.com/virtengine/virtengine/upgrades/software/v1.1.0"
	v120 "github.com/virtengine/virtengine/upgrades/software/v1.2.0"
	v130 "github.com/virtengine/virtengine/upgrades/software/v1.3.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

func TestUpgradeRegistryIncludesAllExpectedVersions(t *testing.T) {
	expected := expectedUpgradeNames()

	var actual []string
	for name := range utypes.GetUpgradesList() {
		actual = append(actual, name)
	}

	sort.Strings(actual)
	require.Equal(t, expected, actual)
}

func TestUpgradeRegistryIsSemverSorted(t *testing.T) {
	expected := expectedUpgradeNames()
	actual := append([]string(nil), expected...)
	sort.Slice(actual, func(i, j int) bool {
		return semver.Compare(actual[i], actual[j]) < 0
	})

	require.Equal(t, expected, actual)
}

func TestUpgradeConstructorsReturnRegisteredUpgrades(t *testing.T) {
	upgrades := utypes.GetUpgradesList()
	require.NotEmpty(t, upgrades)

	for _, name := range expectedUpgradeNames() {
		initFn, ok := upgrades[name]
		require.True(t, ok, "upgrade %s should be registered", name)

		upgrade, err := initFn(log.NewNopLogger(), &apptypes.App{})
		require.NoError(t, err, name)
		require.NotNil(t, upgrade, name)
		require.NotNil(t, upgrade.StoreLoader(), name)
	}
}

func expectedUpgradeNames() []string {
	return []string{
		v100.UpgradeName,
		v110.UpgradeName,
		v120.UpgradeName,
		v130.UpgradeName,
	}
}
