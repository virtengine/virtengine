package v1_7_0_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	escrowmodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	v1_7_0 "github.com/virtengine/virtengine/upgrades/software/v1.7.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	fraudtypes "github.com/virtengine/virtengine/x/fraud/types"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
	reviewtypes "github.com/virtengine/virtengine/x/review/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

func TestUpgradeHandlerActivatesCanonicalFinancialCasesIdempotently(t *testing.T) {
	application := app.Setup(app.WithChainID("task84d-upgrade-handler"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	from[settlementtypes.ModuleName] = 2
	from[fraudtypes.ModuleName] = 1
	from[hpctypes.ModuleName] = 2
	from[reviewtypes.ModuleName] = 1
	from[escrowmodule.ModuleName] = 3
	from[resourcestypes.ModuleName] = 2
	ctx = ctx.WithBlockHeight(17)

	constructor, found := utypes.GetUpgradesList()[v1_7_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)
	handler := up.UpgradeHandler()
	to, err := handler(ctx, upgradetypes.Plan{Name: v1_7_0.UpgradeName, Height: 17}, from)
	require.NoError(t, err)
	require.Equal(t, uint64(3), to[settlementtypes.ModuleName])
	require.Equal(t, uint64(2), to[fraudtypes.ModuleName])
	require.Equal(t, uint64(3), to[hpctypes.ModuleName])
	require.Equal(t, uint64(2), to[reviewtypes.ModuleName])
	require.Equal(t, uint64(4), to[escrowmodule.ModuleName])
	require.Equal(t, uint64(3), to[resourcestypes.ModuleName])
	require.True(t, application.Keepers.VirtEngine.Settlement.IsFinancialCasesActive(ctx))
	require.Empty(t, application.Keepers.VirtEngine.Settlement.ValidateFinancialCaseInvariants(ctx))

	retry, err := handler(ctx, upgradetypes.Plan{Name: v1_7_0.UpgradeName, Height: 17}, to)
	require.NoError(t, err)
	require.Equal(t, to, retry)
}
