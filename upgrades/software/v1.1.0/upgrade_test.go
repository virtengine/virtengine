//go:build upgrade_test

package v1_1_0_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	v110 "github.com/virtengine/virtengine/upgrades/software/v1.1.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

func TestUpgradeHarnessCoverageRegistered(t *testing.T) {
	data, err := os.ReadFile(filepath.Clean("../../../tests/upgrade/test-cases.json"))
	require.NoError(t, err)

	var cases map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &cases))

	_, hasCase := cases[v110.UpgradeName]
	require.True(t, hasCase, "upgrade %s should have an e2e harness test case", v110.UpgradeName)

	_, isRegistered := utypes.GetUpgradesList()[v110.UpgradeName]
	require.True(t, isRegistered, "upgrade %s should be registered", v110.UpgradeName)
}
