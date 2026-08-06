package v1_8_0_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	v1_8_0 "github.com/virtengine/virtengine/upgrades/software/v1.8.0"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

func TestUpgradeHandlerActivatesAuthenticatedFiatProtocolIdempotently(t *testing.T) {
	application := app.Setup(app.WithChainID("task85b-upgrade-handler"))
	ctx := application.NewContext(false).WithBlockHeight(18)
	from := application.MM.GetVersionMap()
	from[settlementtypes.ModuleName] = 3

	constructor, found := utypes.GetUpgradesList()[v1_8_0.UpgradeName]
	require.True(t, found)
	up, err := constructor(application.Logger(), application.App)
	require.NoError(t, err)
	handler := up.UpgradeHandler()
	to, err := handler(ctx, upgradetypes.Plan{Name: v1_8_0.UpgradeName, Height: 18}, from)
	require.NoError(t, err)
	require.Equal(t, uint64(4), to[settlementtypes.ModuleName])
	params := application.Keepers.VirtEngine.Settlement.GetParams(ctx)
	require.False(t, params.FiatConversionEnabled)
	require.Equal(t, settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED, params.FiatConversionDEXProfileState)
	require.Equal(t, settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED, params.FiatConversionPayoutProfileState)
	require.Empty(t, application.Keepers.VirtEngine.Settlement.ValidateFiatConversionInvariants(ctx))

	retry, err := handler(ctx, upgradetypes.Plan{Name: v1_8_0.UpgradeName, Height: 18}, to)
	require.NoError(t, err)
	require.Equal(t, to, retry)
}
