package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/deployment/keeper"
)

type marketStub struct{}

func (marketStub) CreateOrder(_ sdk.Context, _ v1.GroupID, _ types.GroupSpec) (mtypes.Order, error) {
	return mtypes.Order{}, nil
}

func (marketStub) OnGroupClosed(_ sdk.Context, _ v1.GroupID) error {
	return nil
}

type escrowStub struct {
	closed []escrowid.Account
}

func (e *escrowStub) AccountCreate(_ sdk.Context, _ escrowid.Account, _ sdk.AccAddress, _ []etypes.Depositor) error {
	return nil
}

func (e *escrowStub) AccountDeposit(_ sdk.Context, _ escrowid.Account, _ []etypes.Depositor) error {
	return nil
}

func (e *escrowStub) AccountClose(_ sdk.Context, id escrowid.Account) error {
	e.closed = append(e.closed, id)
	return nil
}

func (e *escrowStub) AuthorizeDeposits(_ sdk.Context, _ sdk.Msg) ([]etypes.Depositor, error) {
	return nil, nil
}

func TestMsgServerUpdateDeployment(t *testing.T) {
	ctx, depKeeper := setupKeeper(t)

	deployment := testutil.Deployment(t)
	groups := testutil.DeploymentGroups(t, deployment.ID, 0)
	require.NoError(t, depKeeper.Create(ctx, deployment, groups))

	newHash := []byte{9, 9, 9}
	msgServer := keeper.NewMsgServer(depKeeper, marketStub{}, &escrowStub{})
	_, err := msgServer.UpdateDeployment(ctx, &types.MsgUpdateDeployment{
		ID:   deployment.ID,
		Hash: newHash,
	})
	require.NoError(t, err)
}

func TestMsgServerCloseDeployment(t *testing.T) {
	ctx, depKeeper := setupKeeper(t)

	deployment := testutil.Deployment(t)
	groups := testutil.DeploymentGroups(t, deployment.ID, 0)
	require.NoError(t, depKeeper.Create(ctx, deployment, groups))

	escrow := &escrowStub{}
	msgServer := keeper.NewMsgServer(depKeeper, marketStub{}, escrow)
	_, err := msgServer.CloseDeployment(ctx, &types.MsgCloseDeployment{
		ID: deployment.ID,
	})
	require.NoError(t, err)
	require.Len(t, escrow.closed, 1)
}
