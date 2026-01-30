package v1_1_0

import (
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkTestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	"github.com/virtengine/virtengine/testutil/state"
)

const upgradeTestIBCDenom = "ibc/170C677610AC31DF0904FFE09CD3B5C657492170E7E52372E48756B71E56F2F1"

func TestCloseOverdrawnEscrowAccounts_Overdraft(t *testing.T) {
	suite := state.SetupTestSuite(t)
	ctx := suite.Context()

	owner := sdkTestutil.AccAddress(t)
	provider := sdkTestutil.AccAddress(t)
	did := sdkTestutil.DeploymentIDForAccount(t, owner)
	lid := sdkTestutil.LeaseIDForAccount(t, owner, provider)

	aid := did.ToEscrowAccountID()
	pid := lid.ToEscrowPaymentID()

	depositCoin := sdk.NewCoin(upgradeTestIBCDenom, sdkmath.NewInt(50))
	rateCoin := sdk.NewCoin(upgradeTestIBCDenom, sdkmath.NewInt(10))

	suite.BankKeeper().
		On("SendCoinsFromAccountToModule", mock.Anything, owner, emodule.ModuleName, sdk.NewCoins(depositCoin)).
		Return(nil).Once()

	require.NoError(t, suite.EscrowKeeper().AccountCreate(ctx, aid, owner, []etypes.Depositor{{
		Owner:   owner.String(),
		Height: ctx.BlockHeight(),
		Balance: sdk.NewDecCoinFromCoin(depositCoin),
	}}))

	require.NoError(t, suite.EscrowKeeper().PaymentCreate(ctx, pid, provider, sdk.NewDecCoinFromCoin(rateCoin)))

	ctx = ctx.WithBlockHeight(10)

	up, err := initUpgrade(log.NewNopLogger(), suite.App().App)
	require.NoError(t, err)

	require.NoError(t, up.(*upgrade).closeOverdrawnEscrowAccounts(ctx))

	account, err := suite.EscrowKeeper().GetAccount(ctx, aid)
	require.NoError(t, err)
	require.Equal(t, etypes.StateOverdrawn, account.State.State)
	require.Empty(t, account.State.Deposits)

	payment, err := suite.EscrowKeeper().GetPayment(ctx, pid)
	require.NoError(t, err)
	require.Equal(t, etypes.StateOverdrawn, payment.State.State)
}

func TestCloseOverdrawnEscrowAccounts_Closed(t *testing.T) {
	suite := state.SetupTestSuite(t)
	ctx := suite.Context()

	owner := sdkTestutil.AccAddress(t)
	provider := sdkTestutil.AccAddress(t)
	did := sdkTestutil.DeploymentIDForAccount(t, owner)
	lid := sdkTestutil.LeaseIDForAccount(t, owner, provider)

	aid := did.ToEscrowAccountID()
	pid := lid.ToEscrowPaymentID()

	depositCoin := sdk.NewCoin(upgradeTestIBCDenom, sdkmath.NewInt(200))
	rateCoin := sdk.NewCoin(upgradeTestIBCDenom, sdkmath.NewInt(10))

	suite.BankKeeper().
		On("SendCoinsFromAccountToModule", mock.Anything, owner, emodule.ModuleName, sdk.NewCoins(depositCoin)).
		Return(nil).Once()

	require.NoError(t, suite.EscrowKeeper().AccountCreate(ctx, aid, owner, []etypes.Depositor{{
		Owner:   owner.String(),
		Height: ctx.BlockHeight(),
		Balance: sdk.NewDecCoinFromCoin(depositCoin),
	}}))

	require.NoError(t, suite.EscrowKeeper().PaymentCreate(ctx, pid, provider, sdk.NewDecCoinFromCoin(rateCoin)))

	ctx = ctx.WithBlockHeight(10)

	up, err := initUpgrade(log.NewNopLogger(), suite.App().App)
	require.NoError(t, err)

	require.NoError(t, up.(*upgrade).closeOverdrawnEscrowAccounts(ctx))

	account, err := suite.EscrowKeeper().GetAccount(ctx, aid)
	require.NoError(t, err)
	require.Equal(t, etypes.StateClosed, account.State.State)
	require.Empty(t, account.State.Deposits)

	payment, err := suite.EscrowKeeper().GetPayment(ctx, pid)
	require.NoError(t, err)
	require.Equal(t, etypes.StateClosed, payment.State.State)
}
