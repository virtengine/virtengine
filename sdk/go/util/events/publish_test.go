package events

// import (
// 	"testing"
//
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/assert"
//
// 	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
// 	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
// 	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
// 	"github.com/virtengine/virtengine/sdk/go/sdkutil"
//
// 	"github.com/virtengine/virtengine/testutil"
// )
//
// func Test_processEvent(t *testing.T) {
// 	tests := []sdkutil.ModuleEvent{
// 		// x/deployment events
// 		dtypes.NewEventDeploymentCreated(testutil.DeploymentID(t), testutil.DeploymentVersion(t)),
// 		dtypes.NewEventDeploymentUpdated(testutil.DeploymentID(t), testutil.DeploymentVersion(t)),
// 		dtypes.NewEventDeploymentClosed(testutil.DeploymentID(t)),
// 		dtypes.NewEventGroupClosed(testutil.GroupID(t)),
//
// 		// x/market events
// 		mtypes.NewEventOrderCreated(testutil.OrderID(t)),
// 		mtypes.NewEventOrderClosed(testutil.OrderID(t)),
// 		mtypes.NewEventBidCreated(testutil.BidID(t), testutil.DecCoin(t)),
// 		mtypes.NewEventBidClosed(testutil.BidID(t), testutil.DecCoin(t)),
// 		mtypes.NewEventLeaseCreated(testutil.LeaseID(t), testutil.DecCoin(t)),
// 		mtypes.NewEventLeaseClosed(testutil.LeaseID(t), testutil.DecCoin(t)),
//
// 		// x/provider events
// 		ptypes.NewEventProviderCreated(testutil.AccAddress(t)),
// 		ptypes.NewEventProviderUpdated(testutil.AccAddress(t)),
// 		ptypes.NewEventProviderDeleted(testutil.AccAddress(t)),
// 	}
//
// 	for _, test := range tests {
// 		sdkevs := sdk.Events{
// 			test.ToSDKEvent(),
// 		}.ToABCIEvents()
//
// 		sdkev := sdkevs[0]
//
// 		ev, ok := processEvent(sdkev)
// 		assert.True(t, ok, test)
// 		assert.Equal(t, test, ev, test)
// 	}
// }
