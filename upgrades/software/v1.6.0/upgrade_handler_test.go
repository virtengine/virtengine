package v1_6_0_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	"github.com/virtengine/virtengine/upgrades/software/v1.6.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	resourcetypes "github.com/virtengine/virtengine/x/resources/types"
)

func TestUpgradeHandlerRunsRegisteredMigrationsAndIsIdempotent(t *testing.T) {
	application := app.Setup(app.WithChainID("task84c-upgrade-handler"))
	ctx := application.NewContext(false)
	ctx.KVStore(application.GetKey(marketplacetypes.StoreKey)).Delete(marketplacetypes.CanonicalLifecycleActivationKey())

	fromVM := application.MM.GetVersionMap()
	fromVM[marketv1.ModuleName] = 7
	fromVM[marketplacetypes.ModuleName] = 1
	fromVM[resourcetypes.ModuleName] = 1
	fromVM[hpctypes.ModuleName] = 1

	constructor, found := utypes.GetUpgradesList()[v1_6_0.UpgradeName]
	require.True(t, found)
	upgrade, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	toVM, err := upgrade.UpgradeHandler()(ctx, upgradetypes.Plan{Name: v1_6_0.UpgradeName}, fromVM)
	require.NoError(t, err)
	require.Equal(t, uint64(8), toVM[marketv1.ModuleName])
	require.Equal(t, uint64(2), toVM[marketplacetypes.ModuleName])
	require.Equal(t, uint64(2), toVM[resourcetypes.ModuleName])
	require.Equal(t, uint64(2), toVM[hpctypes.ModuleName])
	require.True(t, application.Keepers.VirtEngine.Marketplace.IsCanonicalLifecycleActive(ctx))
	require.True(t, application.Keepers.VirtEngine.Resources.IsCanonicalReservationsActive(ctx))

	retried, err := upgrade.UpgradeHandler()(ctx, upgradetypes.Plan{Name: v1_6_0.UpgradeName}, toVM)
	require.NoError(t, err)
	require.Equal(t, module.VersionMap(toVM), retried)

	retriedFromOriginal, err := upgrade.UpgradeHandler()(ctx, upgradetypes.Plan{Name: v1_6_0.UpgradeName}, fromVM)
	require.NoError(t, err)
	require.Equal(t, module.VersionMap(toVM), retriedFromOriginal)
}

func TestUpgradeHandlerRejectsPrematureNonOwnerActivation(t *testing.T) {
	application := app.Setup(app.WithChainID("task84c-upgrade-precondition"))
	ctx := application.NewContext(false)
	ctx.KVStore(application.GetKey(resourcetypes.StoreKey)).Delete(resourcetypes.CanonicalReservationsActivationKey())
	application.Keepers.VirtEngine.Marketplace.ActivateCanonicalLifecycle(ctx)
	fromVM := application.MM.GetVersionMap()
	fromVM[marketv1.ModuleName] = 7
	fromVM[marketplacetypes.ModuleName] = 1
	fromVM[resourcetypes.ModuleName] = 1
	fromVM[hpctypes.ModuleName] = 1

	constructor := utypes.GetUpgradesList()[v1_6_0.UpgradeName]
	upgrade, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)
	_, err = upgrade.UpgradeHandler()(ctx, upgradetypes.Plan{Name: v1_6_0.UpgradeName}, fromVM)
	require.ErrorContains(t, err, "canonical activation is partial")
}
