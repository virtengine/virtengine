package v1_9_0_test

import (
	"encoding/json"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	v1_9_0 "github.com/virtengine/virtengine/upgrades/software/v1.9.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	marketplacetypes "github.com/virtengine/virtengine/x/market/types/marketplace"
	resourcestypes "github.com/virtengine/virtengine/x/resources/types"
)

func TestUpgradeHandlerEnablesAutoResolution(t *testing.T) {
	application := app.Setup(app.WithChainID("task-adr010-upgrade-handler"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	ctx = ctx.WithBlockHeight(17)

	constructor, found := utypes.GetUpgradesList()[v1_9_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	require.True(t, application.Keepers.VirtEngine.Resources.IsCanonicalReservationsActive(ctx))
	require.True(t, application.Keepers.VirtEngine.Marketplace.IsCanonicalLifecycleActive(ctx))
	require.False(t, application.Keepers.VirtEngine.Marketplace.GetParams(ctx).EnableAutoResolution)

	handler := up.UpgradeHandler()
	to, err := handler(ctx, upgradetypes.Plan{Name: v1_9_0.UpgradeName, Height: 17}, from)
	require.NoError(t, err)
	require.Equal(t, from, to, "no module version changes are required")
	require.True(t, application.Keepers.VirtEngine.Marketplace.GetParams(ctx).EnableAutoResolution)

	retry, err := handler(ctx, upgradetypes.Plan{Name: v1_9_0.UpgradeName, Height: 17}, to)
	require.NoError(t, err)
	require.Equal(t, to, retry)
}

func TestUpgradeHandlerRejectsInactiveReservations(t *testing.T) {
	application := app.Setup(app.WithChainID("task-adr010-upgrade-guard"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	ctx = ctx.WithBlockHeight(17)

	// Simulate a chain that has not activated canonical reservations.
	ctx.KVStore(application.GetKey(resourcestypes.ModuleName)).Delete(resourcestypes.CanonicalReservationsActivationKey())
	require.False(t, application.Keepers.VirtEngine.Resources.IsCanonicalReservationsActive(ctx))

	constructor, found := utypes.GetUpgradesList()[v1_9_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	handler := up.UpgradeHandler()
	_, err = handler(ctx, upgradetypes.Plan{Name: v1_9_0.UpgradeName, Height: 17}, from)
	require.ErrorContains(t, err, "canonical reservations are not active")
	require.False(t, application.Keepers.VirtEngine.Marketplace.GetParams(ctx).EnableAutoResolution)
}

func TestUpgradeHandlerRejectsActiveLegacyLifecycle(t *testing.T) {
	application := app.Setup(app.WithChainID("task-adr010-upgrade-legacy"))
	ctx := application.NewContext(false)
	from := application.MM.GetVersionMap()
	ctx = ctx.WithBlockHeight(17)

	// Seed a non-terminal legacy order directly in state. The canonical
	// write fence would reject this through public methods by design.
	order := marketplacetypes.Order{
		ID:                marketplacetypes.OrderID{CustomerAddress: "ve1customer", Sequence: 1},
		State:             marketplacetypes.OrderStateOpen,
		RequestedQuantity: 1,
		MaxBidPrice:       1,
	}
	bz, err := json.Marshal(order)
	require.NoError(t, err)
	ctx.KVStore(application.GetKey(marketplacetypes.StoreKey)).Set(marketplacetypes.OrderKey(order.ID), bz)

	constructor, found := utypes.GetUpgradesList()[v1_9_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)

	handler := up.UpgradeHandler()
	_, err = handler(ctx, upgradetypes.Plan{Name: v1_9_0.UpgradeName, Height: 17}, from)
	require.ErrorContains(t, err, "non-terminal legacy order")
	require.False(t, application.Keepers.VirtEngine.Marketplace.GetParams(ctx).EnableAutoResolution)
}
