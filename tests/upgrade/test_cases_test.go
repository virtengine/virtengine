package upgrade

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	apptypes "github.com/virtengine/virtengine/app/types"
	_ "github.com/virtengine/virtengine/upgrades"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

type upgradeMatrixCase struct {
	Modules struct {
		Added   []string `json:"added"`
		Removed []string `json:"removed"`
		Renamed struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"renamed"`
	} `json:"modules"`
	Migrations map[string][]upgradeMigration `json:"migrations"`
}

type upgradeMigration struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func TestUpgradeTestCasesMatchStoreUpgrades(t *testing.T) {
	cases := loadUpgradeTestCases(t)
	require.NotEmpty(t, cases)

	upgrades := utypes.GetUpgradesList()
	require.NotEmpty(t, upgrades)

	for name, initFn := range upgrades {
		tc, ok := cases[name]
		require.True(t, ok, "missing test case for upgrade %s", name)

		up, err := initFn(log.NewNopLogger(), &apptypes.App{})
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

func TestUpgradeTestCasesMigrationMatrix(t *testing.T) {
	cases := loadUpgradeTestCases(t)
	require.NotEmpty(t, cases)
	require.NoError(t, validateUpgradeMigrationMatrix(cases))
}

func TestUpgradeTestCasesMigrationMatrixRejectsInvalidDeclarations(t *testing.T) {
	migrationCase := func(module string, migrations ...upgradeMigration) upgradeMatrixCase {
		return upgradeMatrixCase{Migrations: map[string][]upgradeMigration{module: migrations}}
	}

	tests := []struct {
		name    string
		cases   map[string]upgradeMatrixCase
		message string
	}{
		{
			name:    "zero transitions",
			cases:   map[string]upgradeMatrixCase{"v1.0.0": {}},
			message: "selected zero transitions",
		},
		{
			name: "non incremental transition",
			cases: map[string]upgradeMatrixCase{
				"v1.0.0": migrationCase("market", upgradeMigration{From: "1", To: "3"}),
			},
			message: "must advance exactly one consensus version",
		},
		{
			name: "duplicate transition",
			cases: map[string]upgradeMatrixCase{
				"v1.0.0": migrationCase("market", upgradeMigration{From: "1", To: "2"}),
				"v1.1.0": migrationCase("market", upgradeMigration{From: "1", To: "2"}),
			},
			message: "duplicate migration",
		},
		{
			name: "discontinuous chain",
			cases: map[string]upgradeMatrixCase{
				"v1.0.0": migrationCase("market", upgradeMigration{From: "1", To: "2"}),
				"v1.1.0": migrationCase("market", upgradeMigration{From: "3", To: "4"}),
			},
			message: "is discontinuous",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorContains(t, validateUpgradeMigrationMatrix(tc.cases), tc.message)
		})
	}
}

func validateUpgradeMigrationMatrix(cases map[string]upgradeMatrixCase) error {

	type transition struct {
		upgrade string
		from    uint64
		to      uint64
	}

	byModule := make(map[string][]transition)
	seen := make(map[string]string)
	transitionCount := 0

	for upgrade, tc := range cases {
		for module, migrations := range tc.Migrations {
			if module == "" {
				return fmt.Errorf("upgrade %s declares a migration with an empty module name", upgrade)
			}
			for _, migration := range migrations {
				from, err := strconv.ParseUint(migration.From, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid from version for %s/%s: %w", upgrade, module, err)
				}
				to, err := strconv.ParseUint(migration.To, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid to version for %s/%s: %w", upgrade, module, err)
				}
				if to != from+1 {
					return fmt.Errorf("migration %s/%s must advance exactly one consensus version", upgrade, module)
				}

				key := module + "/" + migration.From + "-" + migration.To
				if previousUpgrade, ok := seen[key]; ok {
					return fmt.Errorf("duplicate migration %s declared by %s and %s", key, previousUpgrade, upgrade)
				}
				seen[key] = upgrade
				byModule[module] = append(byModule[module], transition{upgrade: upgrade, from: from, to: to})
				transitionCount++
			}
		}
	}

	if transitionCount == 0 {
		return fmt.Errorf("upgrade migration matrix selected zero transitions")
	}
	for module, transitions := range byModule {
		sort.Slice(transitions, func(i, j int) bool {
			return transitions[i].from < transitions[j].from
		})
		for i := 1; i < len(transitions); i++ {
			if transitions[i].from != transitions[i-1].to {
				return fmt.Errorf("migration chain for %s is discontinuous between %s and %s", module, transitions[i-1].upgrade, transitions[i].upgrade)
			}
		}
	}

	return nil
}

func loadUpgradeTestCases(t *testing.T) map[string]upgradeMatrixCase {
	t.Helper()

	data, err := os.ReadFile("test-cases.json")
	require.NoError(t, err)

	cases := make(map[string]upgradeMatrixCase)
	require.NoError(t, json.Unmarshal(data, &cases))
	return cases
}
