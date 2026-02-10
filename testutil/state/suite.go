package state

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"cosmossdk.io/collections"
	"cosmossdk.io/store"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	atypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	mtypes "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	ttypes "github.com/virtengine/virtengine/sdk/go/node/take/v1"

	"github.com/virtengine/virtengine/app"
	emocks "github.com/virtengine/virtengine/testutil/cosmos/mocks"
	akeeper "github.com/virtengine/virtengine/x/audit/keeper"
	dkeeper "github.com/virtengine/virtengine/x/deployment/keeper"
	ekeeper "github.com/virtengine/virtengine/x/escrow/keeper"
	mhooks "github.com/virtengine/virtengine/x/market/hooks"
	mkeeper "github.com/virtengine/virtengine/x/market/keeper"
	pkeeper "github.com/virtengine/virtengine/x/provider/keeper"
	tkeeper "github.com/virtengine/virtengine/x/take/keeper"
)

// TestSuite encapsulates a functional VirtEngine nodes data stores for
// ephemeral testing.
type TestSuite struct {
	t       testing.TB
	ms      store.CommitMultiStore
	ctx     sdk.Context
	app     *app.VirtEngineApp
	keepers Keepers
}

type Keepers struct {
	Take       tkeeper.IKeeper
	Escrow     ekeeper.Keeper
	Audit      akeeper.IKeeper
	Market     mkeeper.IKeeper
	Deployment dkeeper.IKeeper
	Provider   pkeeper.IKeeper
	Bank       *emocks.BankKeeper
	Authz      *emocks.AuthzKeeper
}

// SetupTestSuite provides toolkit for accessing stores and keepers
// for complex data interactions.
func SetupTestSuite(t testing.TB) *TestSuite {
	return SetupTestSuiteWithKeepers(t, Keepers{})
}

func SetupTestSuiteWithKeepers(t testing.TB, keepers Keepers) *TestSuite {
	dir, err := os.MkdirTemp("", "virtengined-test-home")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	if keepers.Bank == nil {
		bkeeper := &emocks.BankKeeper{}
		// do not set bank mock during suite setup, each test must set them manually
		// to make sure escrow balance values are tracked correctly
		bkeeper.
			On("SpendableCoin", mock.Anything, mock.Anything, mock.Anything).
			Return(sdk.NewInt64Coin("uve", 10000000))

		keepers.Bank = bkeeper
	}

	if keepers.Authz == nil {
		keeper := &emocks.AuthzKeeper{}

		keepers.Authz = keeper
	}

	app := app.Setup(
		app.WithCheckTx(false),
		app.WithHome(dir),
		app.WithGenesis(app.GenesisStateWithValSet),
	)

	ctx := app.NewContext(false)

	cdc := app.AppCodec()

	vals, err := app.Keepers.Cosmos.Staking.GetAllValidators(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Manually set validator signing info, otherwise we panic
	for _, val := range vals {
		consAddr, _ := val.GetConsAddr()
		signingInfo := slashingtypes.NewValidatorSigningInfo(
			consAddr,
			0,
			ctx.BlockHeight(),
			time.Unix(0, 0),
			false,
			0,
		)
		err = app.Keepers.Cosmos.Slashing.SetValidatorSigningInfo(ctx, consAddr, signingInfo)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	if keepers.Audit == nil {
		keepers.Audit = akeeper.NewKeeper(cdc, app.GetKey(atypes.StoreKey))
	}

	if keepers.Take == nil {
		keepers.Take = tkeeper.NewKeeper(cdc, app.GetKey(ttypes.StoreKey), authtypes.NewModuleAddress(govtypes.ModuleName).String())
	}

	if keepers.Escrow == nil {
		storeService := runtime.NewKVStoreService(app.GetKey(types.StoreKey))
		sb := collections.NewSchemaBuilder(storeService)

		feepool := collections.NewItem(sb, types.FeePoolKey, "fee_pool", codec.CollValue[types.FeePool](cdc))
		keepers.Escrow = ekeeper.NewKeeper(cdc, app.GetKey(emodule.StoreKey), keepers.Bank, keepers.Take, keepers.Authz, feepool)
	}
	if keepers.Market == nil {
		keepers.Market = mkeeper.NewKeeper(cdc, app.GetKey(mtypes.StoreKey), keepers.Escrow, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	}
	if keepers.Deployment == nil {
		keepers.Deployment = dkeeper.NewKeeper(cdc, app.GetKey(dtypes.StoreKey), keepers.Escrow, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	}
	if keepers.Provider == nil {
		keepers.Provider = pkeeper.NewKeeper(cdc, app.GetKey(ptypes.StoreKey))
	}

	hook := mhooks.New(keepers.Deployment, keepers.Market)

	keepers.Escrow.AddOnAccountClosedHook(hook.OnEscrowAccountClosed)
	keepers.Escrow.AddOnPaymentClosedHook(hook.OnEscrowPaymentClosed)

	return &TestSuite{
		t:       t,
		app:     app,
		ctx:     ctx,
		keepers: keepers,
	}
}

func (ts *TestSuite) PrepareMocks(fn func(ts *TestSuite)) {
	fn(ts)
}

func (ts *TestSuite) App() *app.VirtEngineApp {
	return ts.app
}

// SetBlockHeight provides arbitrarily setting the chain's block height.
func (ts *TestSuite) SetBlockHeight(height int64) {
	ts.ctx = ts.ctx.WithBlockHeight(height)
}

// Store provides access to the underlying KVStore
func (ts *TestSuite) Store() store.CommitMultiStore {
	return ts.ms
}

// Context of the current mempool
func (ts *TestSuite) Context() sdk.Context {
	return ts.ctx
}

// AuditKeeper key store
func (ts *TestSuite) AuditKeeper() akeeper.IKeeper {
	return ts.keepers.Audit
}

// EscrowKeeper key store
func (ts *TestSuite) EscrowKeeper() ekeeper.Keeper {
	return ts.keepers.Escrow
}

// MarketKeeper key store
func (ts *TestSuite) MarketKeeper() mkeeper.IKeeper {
	return ts.keepers.Market
}

// DeploymentKeeper key store
func (ts *TestSuite) DeploymentKeeper() dkeeper.IKeeper {
	return ts.keepers.Deployment
}

// ProviderKeeper key store
func (ts *TestSuite) ProviderKeeper() pkeeper.IKeeper {
	return ts.keepers.Provider
}

// BankKeeper key store
func (ts *TestSuite) BankKeeper() *emocks.BankKeeper {
	return ts.keepers.Bank
}

// AuthzKeeper key store
func (ts *TestSuite) AuthzKeeper() *emocks.AuthzKeeper {
	return ts.keepers.Authz
}
