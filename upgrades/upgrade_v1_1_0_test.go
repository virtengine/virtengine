package upgrades_test

import (
	"strings"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dv1 "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	dv1beta "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	emodule "github.com/virtengine/virtengine/sdk/go/node/escrow/module"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	"github.com/virtengine/virtengine/sdk/go/testutil"

	apptypes "github.com/virtengine/virtengine/app/types"
	"github.com/virtengine/virtengine/testutil/state"
	utypes "github.com/virtengine/virtengine/upgrades/types"
	"github.com/virtengine/virtengine/x/escrow/keeper"
)

// v1.1.0 is the heaviest upgrade handler in the repo: ~330 lines of pure state
// surgery that runs ONCE per chain upgrade and rewrites escrow balances, escrow
// payment balances, and - through the deployment and market hooks - orders, bids
// and leases. Before this file nothing in the tree called UpgradeHandler() for
// it at all: upgrades/software/v1.1.0/upgrade_test.go only asserts the handler is
// REGISTERED and that an e2e harness case exists, so deleting the entire
// closeOverdrawnEscrowAccounts call site left the build fully green.
//
// These tests drive the handler over a real app and pin:
//   - the ibc/170C denom filter - the guard deciding whose funds get touched,
//   - the owed = totalRate * (blockHeight + SettledAt) arithmetic,
//   - the Overdrawn/Closed state flip driven by deposits <= owed,
//   - the payment rewrite (zeroed Balance, Unsettled backfilled from Rate),
//   - idempotence of the escrow, market and deployment stores after a re-run,
//   - and the unguarded Funds[0] index at upgrade.go:102, which today PANICS the
//     handler instead of returning an error an operator could act on.
const upgradeV110 = "v1.1.0"

// upgradeUSDTIBCDenom is the only denom the handler will rewrite: the ibc/170C
// hash of veUSDt. An account whose first fund is anything else is skipped by the
// continue at upgrades/software/v1.1.0/upgrade.go:102-104.
const upgradeUSDTIBCDenom = "ibc/170C677610AC31DF0904FFE09CD3B5C657492170E7E52372E48756B71E56F2F1"

func upgradeV110Handler(t *testing.T, app *apptypes.App) func(ctx sdk.Context) error {
	t.Helper()

	upgradeInit, ok := utypes.GetUpgradesList()[upgradeV110]
	require.True(t, ok, "upgrade %s not registered", upgradeV110)

	up, err := upgradeInit(log.NewNopLogger(), app)
	require.NoError(t, err)

	handler := up.UpgradeHandler()
	return func(ctx sdk.Context) error {
		_, err := handler(ctx, upgradetypes.Plan{Name: upgradeV110}, currentVersionMap(app))
		return err
	}
}

// seedV110Deployment writes a real ACTIVE deployment with one OPEN group using the
// keeper's own write path, so the group keys are the ones GetGroups and
// ValidateClosable expect.
//
// gseq is always 1 today; it stays a parameter so a future multi-group fixture
// does not have to reshape every call site, and //nolint:unparam documents that
// honestly rather than letting the linter decide the signature is wrong.
func seedV110Deployment(t *testing.T, ctx sdk.Context, app *apptypes.App, owner sdk.AccAddress, gseq uint32) (dv1.DeploymentID, dv1.GroupID) { //nolint:unparam // see doc comment
	t.Helper()

	did := testutil.DeploymentIDForAccount(t, owner)
	group := testutil.DeploymentGroup(t, did, gseq)

	deployment := dv1.Deployment{
		ID:    did,
		State: dv1.DeploymentActive,
		Hash:  testutil.DefaultDeploymentHash[:],
	}

	require.NoError(t, app.Keepers.VirtEngine.Deployment.Create(ctx, deployment, []dv1beta.Group{group}))

	return did, group.ID
}

// seedV110EscrowAccount writes an OPEN deployment-scope escrow account directly
// into the store in the shape the handler expects, with the ibc/170C denom first.
//
// It deliberately does NOT go through EscrowKeeper.AccountCreate: the public
// write path mints real bank coins and validates deposits against the
// deployment's min-deposit params, which is not the legacy on-chain shape this
// handler was written against. Direct store writes are the same fixture style
// the v1.3.0 suite uses, for the same reason.
func seedV110EscrowAccount(t *testing.T, ctx sdk.Context, app *apptypes.App, id escrowid.Account, owner sdk.AccAddress, funds, deposits sdkmath.LegacyDec) {
	t.Helper()

	acct := etypes.AccountState{
		Owner:     owner.String(),
		State:     etypes.StateOpen,
		SettledAt: 0,
		Funds: []etypes.Balance{{
			Denom:  upgradeUSDTIBCDenom,
			Amount: funds,
		}},
		Deposits: []etypes.Depositor{{
			Owner:   owner.String(),
			Height:  1,
			Balance: sdk.NewDecCoinFromDec(upgradeUSDTIBCDenom, deposits),
		}},
	}

	store := ctx.KVStore(app.GetKey(emodule.StoreKey))
	store.Set(keeper.BuildAccountsKey(etypes.StateOpen, &id), app.GetCodec().MustMarshal(&acct))
}

// seedV110Payment writes an OPEN escrow payment against a deployment account.
// The payment id is built from a real LeaseID via ToEscrowPaymentID rather than
// hand-assembled, because the handler's OnEscrowPaymentClosed path resolves the
// lease through LeaseIDFromPaymentID and silently does nothing if the shape is
// wrong - which would make a cascade assertion vacuous.
func seedV110Payment(t *testing.T, ctx sdk.Context, app *apptypes.App, leaseID mv1.LeaseID, st etypes.State, rate, balance sdkmath.LegacyDec) escrowid.Payment {
	t.Helper()

	pmtID := leaseID.ToEscrowPaymentID()

	pmt := etypes.PaymentState{
		Owner:   leaseID.Owner,
		State:   st,
		Rate:    sdk.NewDecCoinFromDec(upgradeUSDTIBCDenom, rate),
		Balance: sdk.NewDecCoinFromDec(upgradeUSDTIBCDenom, balance),
	}

	store := ctx.KVStore(app.GetKey(emodule.StoreKey))
	store.Set(keeper.BuildPaymentsKey(st, &pmtID), app.GetCodec().MustMarshal(&pmt))

	return pmtID
}

func readV110Account(t *testing.T, ctx sdk.Context, app *apptypes.App, st etypes.State, id escrowid.Account) etypes.AccountState {
	t.Helper()

	bz := ctx.KVStore(app.GetKey(emodule.StoreKey)).Get(keeper.BuildAccountsKey(st, &id))
	require.NotNil(t, bz, "escrow account %s in state %s must exist", id.Key(), st)

	var acct etypes.AccountState
	require.NoError(t, app.GetCodec().Unmarshal(bz, &acct))
	return acct
}

// hasV110Account reports whether an escrow account exists under a state prefix.
// The handler DELETEs the old key and SETs the new one, so asserting both
// "old prefix gone" and "new prefix present" is what proves the record MOVED
// rather than being overwritten in place under one key.
func hasV110Account(t *testing.T, ctx sdk.Context, app *apptypes.App, st etypes.State, id escrowid.Account) bool {
	t.Helper()
	return ctx.KVStore(app.GetKey(emodule.StoreKey)).Has(keeper.BuildAccountsKey(st, &id))
}

func readV110Payment(t *testing.T, ctx sdk.Context, app *apptypes.App, st etypes.State, id escrowid.Payment) etypes.PaymentState {
	t.Helper()

	bz := ctx.KVStore(app.GetKey(emodule.StoreKey)).Get(keeper.BuildPaymentsKey(st, &id))
	require.NotNil(t, bz, "escrow payment %s in state %s must exist", id.XID, st)

	var pmt etypes.PaymentState
	require.NoError(t, app.GetCodec().Unmarshal(bz, &pmt))
	return pmt
}

// TestUpgradeV110HandlerDebitsOwedFromEscrowFunds pins the core money movement:
// owed = sum(payment rates) * (blockHeight + SettledAt), subtracted from Funds[0],
// with deposits cleared. This is the assertion that fails if the owed
// computation, the heightDelta or the Funds[0] debit is mutated.
func TestUpgradeV110HandlerDebitsOwedFromEscrowFunds(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()

	// Height 100; two payments at 0.5 and 0.25 per block; SettledAt 0. The height
	// must be set BEFORE Context() is captured, because the handler reads
	// ctx.BlockHeight() into heightDelta and an SDK context is immutable.
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	seedV110EscrowAccount(t, ctx, app.App, id, owner,
		sdkmath.LegacyNewDec(1000), sdkmath.LegacyNewDec(900))

	seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 1, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOpen, sdkmath.LegacyNewDecWithPrec(5, 1), sdkmath.LegacyNewDec(200))

	seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 2, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOpen, sdkmath.LegacyNewDecWithPrec(25, 2), sdkmath.LegacyNewDec(100))

	require.NoError(t, upgradeV110Handler(t, app.App)(ctx))

	// totalRate = 0.75, heightDelta = 100 + 0 = 100, owed = 75.
	// deposits 900 > owed 75, so NOT an overdraft: the account closes.
	after := readV110Account(t, ctx, app.App, etypes.StateClosed, id)

	require.Equal(t, sdkmath.LegacyNewDec(925), after.Funds[0].Amount,
		"Funds[0] must be debited by exactly owed = totalRate * heightDelta")
	require.Equal(t, upgradeUSDTIBCDenom, after.Funds[0].Denom, "the denom must not be rewritten")
	require.Empty(t, after.Deposits, "deposits must be cleared once the account is settled")
	require.False(t, hasV110Account(t, ctx, app.App, etypes.StateOverdrawn, id),
		"an account that covers its debt must not also land in OVERDRAWN")
	require.False(t, hasV110Account(t, ctx, app.App, etypes.StateOpen, id),
		"the OPEN account key must be deleted, not overwritten in place")
}

// TestUpgradeV110HandlerFlipsToOverdrawnOnOverdrawnPayment covers the only way
// an account lands on OVERDRAWN: at least one of its payments is ALREADY
// overdrawn, which the loop at upgrade.go:123-125 latches into val.State.State
// before the overdraft decision is made.
//
// The deposits<=owed term alone does NOT do it. That was measured, not assumed:
// with deposits 10 and owed 75 the account is still written under the CLOSED
// prefix, because the `if !overdraft` at upgrade.go:142-144 is the only
// statement that assigns StateClosed - when overdraft is true the account keeps
// whatever state the loop left it in, which for an all-OPEN payment set is
// StateOpen. The record therefore moves key but stays semantically "open" while
// its deposits have just been zeroed. That behaviour is pinned separately in
// TestUpgradeV110HandlerLeavesAccountOpenWhenDepositsCannotCoverOwed rather than
// papered over, because it is a money-path anomaly a chain operator needs to see.
func TestUpgradeV110HandlerFlipsToOverdrawnOnOverdrawnPayment(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	// Deposits ABOVE owed, so the only term that can drive the branch is the
	// already-overdrawn payment.
	seedV110EscrowAccount(t, ctx, app.App, id, owner,
		sdkmath.LegacyNewDec(1000), sdkmath.LegacyNewDec(900))

	pmtID := seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 1, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOverdrawn, sdkmath.LegacyNewDecWithPrec(75, 2), sdkmath.LegacyNewDec(50))

	require.NoError(t, upgradeV110Handler(t, app.App)(ctx))

	require.True(t, hasV110Account(t, ctx, app.App, etypes.StateOverdrawn, id),
		"an account with an already-overdrawn payment must flip to OVERDRAWN")
	require.False(t, hasV110Account(t, ctx, app.App, etypes.StateClosed, id),
		"an overdrawn account must not also be written under the CLOSED prefix")
	require.False(t, hasV110Account(t, ctx, app.App, etypes.StateOpen, id))

	after := readV110Account(t, ctx, app.App, etypes.StateOverdrawn, id)
	require.Equal(t, sdkmath.LegacyNewDec(925), after.Funds[0].Amount,
		"the owed debit must happen on the overdrawn branch too")

	pmt := readV110Payment(t, ctx, app.App, etypes.StateOverdrawn, pmtID)
	require.Equal(t, sdkmath.LegacyZeroDec(), pmt.Balance.Amount,
		"an overdrawn payment must have its balance zeroed")
	// MEASURED, and it is a second real defect. Line upgrade.go:168 reads
	//   payments[i].State.Unsettled.Amount.Set(payments[i].State.Rate.Amount.MulInt64Mut(heightDelta))
	// MulInt64Mut mutates its RECEIVER, and the receiver here is the payment's own
	// Rate.Amount. So the rate the migration persists is not 0.75 per block, it is
	// 0.75 * 100 = 75 - the same total that was just written to Unsettled. The
	// per-block rate is destroyed on every payment the handler rewrites.
	//
	// This is pinned as observed rather than asserted as intended: a later fix must
	// change this expectation to 0.75 AND add a non-mutating copy before
	// MulInt64Mut. See the card filed alongside this suite.
	require.Equal(t, sdkmath.LegacyNewDec(75), pmt.Rate.Amount,
		"the persisted Rate is aliased onto rate*heightDelta by MulInt64Mut at upgrade.go:168")
	require.Equal(t, sdkmath.LegacyNewDec(75), pmt.Unsettled.Amount,
		"Unsettled is rate * heightDelta and equals the corrupted Rate")
}

// TestUpgradeV110HandlerLeavesAccountOpenWhenDepositsCannotCoverOwed pins the
// anomaly measured in TestUpgradeV110HandlerFlipsToOverdrawnOnOverdrawnPayment's
// doc comment, in isolation: deposits (10) strictly below owed (75) with an
// all-OPEN payment set produces an account written under the CLOSED prefix, with
// its funds driven to exactly zero.
//
// funds 1000 - owed 75 = 925 is NOT what happens: measured output is 0. The
// zeroing is the second loop at upgrade.go:180-235, which re-iterates the SAME
// search prefix, finds the account still stored under the OPEN key the first loop
// never wrote over, sees its deployment is now closed, and sets Funds[0] to zero.
// So the account is processed twice in one handler run, and the second pass wins.
//
// This is recorded as observed behaviour, not endorsed. If the intent is for
// deposits<=owed to mean OVERDRAWN, both the missing assignment and the
// double-processing at upgrade.go:180 are defects in a one-shot migration over
// real balances - see the card filed alongside this suite.
func TestUpgradeV110HandlerLeavesAccountOpenWhenDepositsCannotCoverOwed(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	// deposits 10, far below owed = 75.
	seedV110EscrowAccount(t, ctx, app.App, id, owner,
		sdkmath.LegacyNewDec(1000), sdkmath.LegacyNewDec(10))

	seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 1, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOpen, sdkmath.LegacyNewDecWithPrec(75, 2), sdkmath.LegacyNewDec(50))

	require.NoError(t, upgradeV110Handler(t, app.App)(ctx))

	// Measured, not derived: the account ends up CLOSED with funds at zero.
	require.True(t, hasV110Account(t, ctx, app.App, etypes.StateClosed, id),
		"the account is written under the CLOSED prefix")
	require.False(t, hasV110Account(t, ctx, app.App, etypes.StateOverdrawn, id),
		"deposits<=owed alone does NOT produce OVERDRAWN")

	after := readV110Account(t, ctx, app.App, etypes.StateClosed, id)
	require.Equal(t, sdkmath.LegacyZeroDec(), after.Funds[0].Amount,
		"the second pass zeroes Funds[0] once the deployment is seen closed")
	require.Empty(t, after.Deposits)
}

// TestUpgradeV110HandlerBackfillsUnsettledFromRate pins the payment's Unsettled
// field: it must become rate * heightDelta, the amount the account still owes the
// provider. Here Unsettled = 0.75 * 100 = 75.
func TestUpgradeV110HandlerBackfillsUnsettledFromRate(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	seedV110EscrowAccount(t, ctx, app.App, id, owner,
		sdkmath.LegacyNewDec(1000), sdkmath.LegacyNewDec(900))

	pmtID := seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 1, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOpen, sdkmath.LegacyNewDecWithPrec(75, 2), sdkmath.LegacyNewDec(50))

	require.NoError(t, upgradeV110Handler(t, app.App)(ctx))

	pmt := readV110Payment(t, ctx, app.App, etypes.StateClosed, pmtID)
	require.Equal(t, sdkmath.LegacyZeroDec(), pmt.Balance.Amount,
		"the payment balance must be zeroed")
	require.Equal(t, sdkmath.LegacyNewDec(75), pmt.Unsettled.Amount,
		"Unsettled must be backfilled to rate * heightDelta = 0.75 * 100")
}

// TestUpgradeV110HandlerIsIdempotent is the retried-proposal check. A second run
// of the handler must not rewrite a single byte of the escrow, market or
// deployment store: a node that retried the upgrade and a node that did not must
// not diverge, or the chain halts instead of retrying.
func TestUpgradeV110HandlerIsIdempotent(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	seedV110EscrowAccount(t, ctx, app.App, id, owner,
		sdkmath.LegacyNewDec(1000), sdkmath.LegacyNewDec(900))
	seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 1, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOpen, sdkmath.LegacyNewDecWithPrec(5, 1), sdkmath.LegacyNewDec(200))

	handler := upgradeV110Handler(t, app.App)

	require.NoError(t, handler(ctx))
	afterFirst := digestV110Stores(t, ctx, app.App)

	require.NoError(t, handler(ctx))
	require.Equal(t, afterFirst, digestV110Stores(t, ctx, app.App),
		"re-running the v1.1.0 upgrade handler must leave escrow, market and deployment byte-identical")
}

// TestUpgradeV110HandlerSkipsForeignDenoms is the scoping test. The handler must
// leave an account whose first fund is NOT the ibc/170C denom completely alone -
// it is other people's money this migration has no business touching.
func TestUpgradeV110HandlerSkipsForeignDenoms(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	seedV110EscrowAccount(t, ctx, app.App, id, owner,
		sdkmath.LegacyNewDec(1000), sdkmath.LegacyNewDec(900))
	seedV110Payment(t, ctx, app.App, mv1.LeaseID{
		Owner: owner.String(), DSeq: did.DSeq, GSeq: 1, OSeq: 1, Provider: owner.String(),
	}, etypes.StateOpen, sdkmath.LegacyNewDecWithPrec(5, 1), sdkmath.LegacyNewDec(200))

	// Re-write the seeded account so Funds[0] carries a different denom, which is
	// the exact condition the guard at upgrade.go:102 screens on.
	store := ctx.KVStore(app.GetKey(emodule.StoreKey))
	var acct etypes.AccountState
	require.NoError(t, app.App.GetCodec().Unmarshal(
		store.Get(keeper.BuildAccountsKey(etypes.StateOpen, &id)), &acct))
	acct.Funds = []etypes.Balance{{Denom: "uve", Amount: sdkmath.LegacyNewDec(1000)}}
	store.Set(keeper.BuildAccountsKey(etypes.StateOpen, &id), app.App.GetCodec().MustMarshal(&acct))

	before := digestV110Stores(t, ctx, app.App)

	require.NoError(t, upgradeV110Handler(t, app.App)(ctx))

	require.Equal(t, before, digestV110Stores(t, ctx, app.App),
		"an escrow account not holding ibc/170C must be left untouched by the upgrade")
}

// TestUpgradeV110HandlerPanicsOnEmptyFunds pins a REAL defect as it behaves
// today. upgrade.go:102 reads val.State.Funds[0].Denom with no length check, so
// an account whose Funds slice is empty - what an interrupted earlier attempt or
// a hand-edited genesis leaves behind - panics and takes the chain down
// mid-upgrade instead of returning an error an operator could act on.
//
// This asserts the CURRENT panicking behaviour so the hazard is visible in CI
// rather than latent. When the Funds[0] guard is added this test fails, and the
// remediation is to flip it to require.NoError and drop the Panic assertion -
// the card asks for one behaviour or the other, never a test that accepts both.
func TestUpgradeV110HandlerPanicsOnEmptyFunds(t *testing.T) {
	suite := state.SetupTestSuite(t)
	app := suite.App()
	suite.SetBlockHeight(100)
	ctx := suite.Context()

	owner := testutil.AccAddress(t)
	did, _ := seedV110Deployment(t, ctx, app.App, owner, 1)
	id := did.ToEscrowAccountID()

	// An OPEN account with NO funds at all.
	acct := etypes.AccountState{
		Owner: owner.String(),
		State: etypes.StateOpen,
		Funds: []etypes.Balance{},
	}
	store := ctx.KVStore(app.GetKey(emodule.StoreKey))
	store.Set(keeper.BuildAccountsKey(etypes.StateOpen, &id), app.App.GetCodec().MustMarshal(&acct))

	require.Panics(t, func() {
		_ = upgradeV110Handler(t, app.App)(ctx)
	}, "Funds[0] is indexed with no length guard at upgrade.go:102 - an empty Funds "+
		"slice panics the handler and halts the chain instead of returning an error")
}

// digestV110Stores fingerprints escrow, market and deployment together: the
// handler writes across all three, so an idempotence claim has to cover all of
// them rather than the escrow store alone.
func digestV110Stores(t *testing.T, ctx sdk.Context, app *apptypes.App) string {
	t.Helper()

	keys := map[string]*storetypes.KVStoreKey{
		"escrow":     app.GetKey(emodule.StoreKey),
		"market":     app.GetKey(mv1.StoreKey),
		"deployment": app.GetKey(dv1.StoreKey),
	}

	digests := make([]string, 0, len(keys))
	for _, name := range []string{"deployment", "escrow", "market"} {
		digests = append(digests, name+"="+storeDigest(t, ctx.KVStore(keys[name])))
	}
	return strings.Join(digests, "|")
}
