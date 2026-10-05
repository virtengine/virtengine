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
	v1100 "github.com/virtengine/virtengine/upgrades/software/v1.10.0"
	v120 "github.com/virtengine/virtengine/upgrades/software/v1.2.0"
	v130 "github.com/virtengine/virtengine/upgrades/software/v1.3.0"
	v140 "github.com/virtengine/virtengine/upgrades/software/v1.4.0"
	v150 "github.com/virtengine/virtengine/upgrades/software/v1.5.0"
	v160 "github.com/virtengine/virtengine/upgrades/software/v1.6.0"
	v170 "github.com/virtengine/virtengine/upgrades/software/v1.7.0"
	v180 "github.com/virtengine/virtengine/upgrades/software/v1.8.0"
	v190 "github.com/virtengine/virtengine/upgrades/software/v1.9.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

func TestUpgradeRegistryIncludesAllExpectedVersions(t *testing.T) {
	expected := expectedUpgradeNames()

	actual := make([]string, 0, len(utypes.GetUpgradesList()))
	for name := range utypes.GetUpgradesList() {
		actual = append(actual, name)
	}

	// Sort with the SAME comparator TestUpgradeRegistryIsSemverSorted asserts,
	// not sort.Strings. Upgrade names are semver, and semver ordering diverges
	// from lexicographic ordering as soon as a version has two digits: "v1.10.0"
	// sorts before "v1.2.0" as a string but after "v1.9.0" as a version. Using
	// sort.Strings here made this test demand a lexicographic list while its
	// sibling demanded a semver one, so no single expected list could satisfy
	// both from v1.10.0 onward.
	sortBySemver(actual)
	require.Equal(t, expected, actual)
}

func TestUpgradeRegistryIsSemverSorted(t *testing.T) {
	expected := expectedUpgradeNames()
	actual := append([]string(nil), expected...)
	sortBySemver(actual)

	require.Equal(t, expected, actual)
}

// sortBySemver orders upgrade names by version. golang.org/x/mod/semver requires
// a leading "v", which every UpgradeName already carries.
func sortBySemver(names []string) {
	sort.Slice(names, func(i, j int) bool {
		return semver.Compare(names[i], names[j]) < 0
	})
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
		v140.UpgradeName,
		v150.UpgradeName,
		v160.UpgradeName,
		v170.UpgradeName,
		v180.UpgradeName,
		v190.UpgradeName,
		v1100.UpgradeName,
	}
}
