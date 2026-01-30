package upgrade

import (
	"encoding/json"
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/testutil/state"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

type testCase struct {
	Modules struct {
		Added   []string `json:"added"`
		Removed []string `json:"removed"`
		Renamed struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"renamed"`
	} `json:"modules"`
	Migrations map[string][]struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"migrations"`
}

func TestUpgradeTestCasesMatchStoreUpgrades(t *testing.T) {
	cases := loadUpgradeTestCases(t)
	require.NotEmpty(t, cases)

	suite := state.SetupTestSuite(t)
	app := suite.App()

	upgrades := utypes.GetUpgradesList()
	require.NotEmpty(t, upgrades)

	for name, initFn := range upgrades {
		tc, ok := cases[name]
		require.True(t, ok, "missing test case for upgrade %s", name)

		up, err := initFn(log.NewNopLogger(), app.App)
		require.NoError(t, err)

		var added []string
		var removed []string
		if store := up.StoreLoader(); store != nil {
			added = append([]string{}, store.Added...)
			removed = append([]string{}, store.Deleted...)
		}

		require.ElementsMatch(t, tc.Modules.Added, added, "added modules mismatch for %s", name)
		require.ElementsMatch(t, tc.Modules.Removed, removed, "removed modules mismatch for %s", name)
	}

	for name := range cases {
		if _, ok := upgrades[name]; !ok {
			t.Fatalf("test case defined for unknown upgrade %s", name)
		}
	}
}

func loadUpgradeTestCases(t *testing.T) map[string]testCase {
	t.Helper()

	data, err := os.ReadFile("test-cases.json")
	require.NoError(t, err)

	cases := make(map[string]testCase)
	require.NoError(t, json.Unmarshal(data, &cases))
	return cases
}
