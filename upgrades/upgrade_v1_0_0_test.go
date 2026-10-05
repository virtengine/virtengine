package upgrades_test

import (
	"testing"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/testutil/state"
	utypes "github.com/virtengine/virtengine/upgrades/types"
)

const upgradeV100 = "v1.0.0"

// TestUpgradeV100HandlerPanicsOnOccupiedBaseappSubspace pins a REAL defect.
//
// upgrade.go:152 runs OUTSIDE the returned closure, at UpgradeHandler() call
// time:
//
//	baseAppLegacySS := up.Keepers.Cosmos.Params.Subspace(baseapp.Paramspace).
//	    WithKeyTable(paramstypes.ConsensusParamsKeyTable())
//
// ParamsKeeper.Subspace panics with "subspace already occupied" when the name is
// already registered, and "baseapp" IS already registered on every app this
// handler can run against. Measured on a freshly constructed test app, before
// this suite touches anything, GetSubspaces() lists 19 subspaces and index 15 is
// name="baseapp" - so this is app construction, not cross-test contamination.
//
// Consequence: the v1.0.0 handler cannot execute at all on a current binary. It
// panics before the returned func is ever called, which is an unrecoverable halt
// at exactly the moment the chain needs the upgrade to succeed. The SDK's own
// guidance for this migration (baseapp/params_legacy.go:8-33) is to use an
// existing subspace or set ConsensusParamsKeeper explicitly, not to re-register.
//
// This test asserts the CURRENT panicking behaviour so the defect is visible in
// CI. Remediation belongs in the handler (use GetSubspace, or drop the legacy
// consensus-param migration entirely and set the params explicitly); once it is
// fixed this test fails and is replaced by one that asserts the migration runs.
// The card forbids weakening x/params usage to make a test pass, and this test
// is deliberately NOT that: it does not modify a single line of handler code.
func TestUpgradeV100HandlerPanicsOnOccupiedBaseappSubspace(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	ctx := suite.Context()

	upgradeInit, ok := utypes.GetUpgradesList()[upgradeV100]
	require.True(t, ok, "upgrade %s not registered", upgradeV100)

	up, err := upgradeInit(log.NewNopLogger(), app.App)
	require.NoError(t, err)

	// The occupancy is a property of the app, not of the handler call, so prove
	// it first: a test that only asserted the panic could be passing because of
	// some unrelated earlier Subspace() call.
	_, found := app.Keepers.Cosmos.Params.GetSubspace(baseapp.Paramspace)
	require.True(t, found,
		"the %q params subspace must already be registered by app construction", baseapp.Paramspace)

	require.PanicsWithValue(t, "subspace already occupied", func() {
		_ = up.UpgradeHandler()
	}, "upgrade.go:152 re-registers the already-occupied %q subspace, which panics "+
		"before the returned closure is invoked", baseapp.Paramspace)

	// The panic happens at UpgradeHandler() time, so the returned func is never
	// reached. This asserts the panic is not merely deferred into the handler.
	var handlerPanicked bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				handlerPanicked = true
			}
		}()
		if _, herr := up.UpgradeHandler()(ctx, upgradetypes.Plan{Name: upgradeV100}, currentVersionMap(app.App)); herr != nil {
			t.Logf("PROBE_V100 handler returned error=%v", herr)
		}
	}()
	require.True(t, handlerPanicked,
		"calling the handler closure must still panic: UpgradeHandler() itself never returns")
}
