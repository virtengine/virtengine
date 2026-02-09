package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/take/v1"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	emocks "github.com/virtengine/virtengine/testutil/cosmos/mocks"
	ekeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	tkeeper "github.com/virtengine/virtengine/x/take/keeper"
)

func setupSettlementKeepers(t *testing.T) (sdk.Context, tkeeper.IKeeper, ekeeper.Keeper, *emocks.BankKeeper) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	takeStoreKey := storetypes.NewKVStoreKey(types.StoreKey)
	escrowStoreKey := storetypes.NewKVStoreKey(emodule.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(takeStoreKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(escrowStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 1,
	}, false, log.NewNopLogger())

	bankKeeper := &emocks.BankKeeper{}
	bankKeeper.
		On("SpendableCoin", mock.Anything, mock.Anything, mock.Anything).
		Return(sdk.NewInt64Coin("uve", 10000000)).
		Maybe()

	authzKeeper := &emocks.AuthzKeeper{}

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	takeKeeper := tkeeper.NewKeeper(cdc, takeStoreKey, authority)

	storeService := runtime.NewKVStoreService(escrowStoreKey)
	sb := collections.NewSchemaBuilder(storeService)
	feepool := collections.NewItem(sb, distrtypes.FeePoolKey, "fee_pool", codec.CollValue[distrtypes.FeePool](cdc))

	escrowKeeper := ekeeper.NewKeeper(cdc, escrowStoreKey, bankKeeper, takeKeeper, authzKeeper, feepool)
	require.NoError(t, feepool.Set(ctx, distrtypes.FeePool{CommunityPool: sdk.NewDecCoins()}))

	return ctx, takeKeeper, escrowKeeper, bankKeeper
}

func TestSettlementTakeFeeIntegration(t *testing.T) {
	ctx, takeKeeper, escrowKeeper, bankKeeper := setupSettlementKeepers(t)

	params := types.Params{
		DefaultTakeRate: 20,
		DenomTakeRates: types.DenomTakeRates{
			{Denom: "uve", Rate: 5},
		},
	}
	require.NoError(t, takeKeeper.SetParams(ctx, params))

	lid := testutil.LeaseID(t)
	did := lid.DeploymentID()
	aid := did.ToEscrowAccountID()
	pid := lid.ToEscrowPaymentID()

	aowner := testutil.AccAddress(t)
	powner := testutil.AccAddress(t)
	amt := testutil.VECoin(t, 1000)
	rate := testutil.VECoin(t, 10)

	bankKeeper.
		On("SendCoinsFromAccountToModule", ctx, aowner, emodule.ModuleName, sdk.NewCoins(amt)).
		Return(nil).Once()

	require.NoError(t, escrowKeeper.AccountCreate(ctx, aid, aowner, []etypes.Depositor{{
		Owner:   aowner.String(),
		Height:  ctx.BlockHeight(),
		Balance: sdk.NewDecCoinFromCoin(amt),
	}}))

	require.NoError(t, escrowKeeper.PaymentCreate(ctx, pid, powner, sdk.NewDecCoinFromCoin(rate)))

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 10)
	paymentGross := rate.Amount.Int64() * 10
	feeAmount := paymentGross * 5 / 100

	bankKeeper.
		On("SendCoinsFromModuleToModule", ctx, emodule.ModuleName, distrtypes.ModuleName, sdk.NewCoins(testutil.VECoin(t, feeAmount))).
		Return(nil).Once().
		On("SendCoinsFromModuleToAccount", mock.Anything, emodule.ModuleName, powner, sdk.NewCoins(testutil.VECoin(t, paymentGross-feeAmount))).
		Return(nil).Once().
		On("SendCoinsFromModuleToAccount", mock.Anything, emodule.ModuleName, aowner, sdk.NewCoins(testutil.VECoin(t, amt.Amount.Int64()-paymentGross))).
		Return(nil).Once()

	require.NoError(t, escrowKeeper.AccountClose(ctx, aid))

	account, err := escrowKeeper.GetAccount(ctx, aid)
	require.NoError(t, err)
	require.Equal(t, etypes.StateClosed, account.State.State)
	bankKeeper.AssertExpectations(t)
}
