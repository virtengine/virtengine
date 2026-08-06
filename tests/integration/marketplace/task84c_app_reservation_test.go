//go:build e2e.integration

package marketplace_test

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/app"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	resourcetypes "github.com/virtengine/virtengine/x/resources/types"
)

func TestTask84CRealAppMarketAndHPCCompeteForFinalCapacity(t *testing.T) {
	appInstance := app.Setup(app.WithChainID("task84c-app-reservation"))
	ctx := appInstance.NewContext(false).WithBlockHeight(10).WithBlockTime(time.Unix(1_750_000_000, 0).UTC())
	provider := sdk.AccAddress(bytes.Repeat([]byte{0x41}, 20)).String()
	require.NoError(t, appInstance.Keepers.VirtEngine.Provider.Create(ctx, providerv1beta4.Provider{Owner: provider, HostURI: "https://provider.example"}))
	customer := sdk.AccAddress(bytes.Repeat([]byte{0x42}, 20)).String()

	require.NoError(t, appInstance.Keepers.VirtEngine.Resources.SetInventory(ctx, resourcetypes.ResourceInventory{
		InventoryId: "last-unit", ProviderAddress: provider, ResourceClass: resourcetypes.ResourceClassCompute,
		Total: resourcetypes.ResourceCapacity{CpuCores: 1}, Available: resourcetypes.ResourceCapacity{CpuCores: 1},
		Active: true, HeartbeatSequence: 1, LastHeartbeat: ctx.BlockTime(), UpdatedAt: ctx.BlockTime(),
	}))

	market, err := appInstance.Keepers.VirtEngine.Resources.Reserve(ctx, resourcesv1.ReservationRequest{
		IdempotencyKey: "market-final", RequestId: "market-order", RequesterAddress: customer,
		ResourceClass: resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE, Capacity: resourcesv1.ResourceCapacity{CpuCores: 1},
		ConsumerType: "market_lease", ConsumerId: "lease-final", Version: 1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, market.ReservationId)

	hpc, err := appInstance.Keepers.VirtEngine.Resources.Reserve(ctx, resourcesv1.ReservationRequest{
		IdempotencyKey: "hpc-final", RequestId: "hpc-job", RequesterAddress: customer,
		ResourceClass: resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE, Capacity: resourcesv1.ResourceCapacity{CpuCores: 1},
		ConsumerType: "hpc_job", ConsumerId: "job-final", Version: 1,
	})
	require.ErrorIs(t, err, resourcetypes.ErrNoEligibleInventory)
	require.Nil(t, hpc)
	require.NoError(t, appInstance.Keepers.VirtEngine.Resources.ValidateCapacityConservation(ctx))
}
