package settlement_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	settlement "github.com/virtengine/virtengine/x/settlement"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func TestSettlementGenesisRoundTripPreservesNextSequencesAndPayouts(t *testing.T) {
	keeper, ctx := newGenesisTestKeeper(t)

	keeper.SetNextEscrowSequence(ctx, 101)
	keeper.SetNextSettlementSequence(ctx, 102)
	keeper.SetNextUsageSequence(ctx, 103)
	keeper.SetNextDistributionSequence(ctx, 104)
	keeper.SetNextPayoutSequence(ctx, 105)
	keeper.SetNextFiatConversionSequence(ctx, 106)
	keeper.ActivateFinancialCases(ctx)

	provider := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()
	payout := types.NewPayoutRecord(
		"payout-1700000000-77", "invoice-genesis", "settlement-genesis", "escrow-genesis",
		"order-genesis", "lease-genesis", provider, customer,
		sdk.NewCoins(sdk.NewInt64Coin("uve", 10)), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(),
		ctx.BlockTime(), ctx.BlockHeight(),
	)
	require.NoError(t, keeper.SetPayout(ctx, *payout))
	conversion := types.NewFiatConversionRecord("conv-1700000000-78", types.FiatConversionRequest{
		InvoiceID: payout.InvoiceID, SettlementID: payout.SettlementID, PayoutID: payout.PayoutID,
		Provider: provider, Customer: customer, RequestedBy: provider,
		CryptoAmount: sdk.NewInt64Coin("uve", 10), FiatCurrency: "USD", PaymentMethod: "bank_transfer",
		DestinationHash: types.HashDestination("opaque"), SlippageTolerance: 0.01,
		CryptoToken: types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken: types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
	}, sdk.NewInt64Coin("uve", 10), ctx.BlockTime())
	conversion.ProtocolVersion = 1
	conversion.SlippageToleranceExact = "0.010000000000000000"
	conversion.RequestDigest = bytes.Repeat([]byte{3}, 32)
	conversion.DailyBucket = "20231114"
	conversion.DailyQuotaReserved = true
	conversion.LegacyQuarantined = true
	conversion.QuarantineReason = "genesis_fixture"
	payout.FiatConversionID = conversion.ConversionID
	require.NoError(t, keeper.SetPayout(ctx, *payout))
	require.NoError(t, keeper.ImportFiatConversion(ctx, *conversion))

	exported := settlement.ExportGenesis(ctx, keeper)
	require.Equal(t, uint64(101), exported.EscrowSequence)
	require.Equal(t, uint64(102), exported.SettlementSequence)
	require.Equal(t, uint64(103), exported.UsageSequence)
	require.Equal(t, uint64(104), exported.DistributionSequence)
	require.Equal(t, uint64(105), exported.PayoutSequence)
	require.Equal(t, uint64(106), exported.FiatConversionSequence)
	require.Len(t, exported.PayoutRecords, 1)
	require.Equal(t, payout.PayoutID, exported.PayoutRecords[0].PayoutID)
	require.True(t, exported.FinancialCasesActive)

	restoredKeeper, restoredCtx := newGenesisTestKeeper(t)
	settlement.InitGenesis(restoredCtx, restoredKeeper, exported)
	restored := settlement.ExportGenesis(restoredCtx, restoredKeeper)
	require.Equal(t, exported.EscrowSequence, restored.EscrowSequence)
	require.Equal(t, exported.PayoutSequence, restored.PayoutSequence)
	require.Equal(t, exported.FiatConversionSequence, restored.FiatConversionSequence)
	require.Len(t, restored.PayoutRecords, 1)
	require.Equal(t, payout.PayoutID, restored.PayoutRecords[0].PayoutID)
	require.Len(t, restored.FiatConversionRecords, 1)
	restoredConversion, found := restoredKeeper.GetFiatConversion(restoredCtx, conversion.ConversionID)
	require.True(t, found)
	require.Equal(t, conversion.RequestDigest, restoredConversion.RequestDigest)
	require.Equal(t, conversion.ConversionID, string(restoredCtx.KVStore(restoredKeeper.StoreKey()).Get(types.FiatConversionIdempotencyKey(conversion.IdempotencyKey))))
	require.Equal(t, "10", string(restoredCtx.KVStore(restoredKeeper.StoreKey()).Get(types.FiatDailyTotalKey(provider, conversion.DailyBucket))))
}

func TestSettlementGenesisValidatesAndPreservesFiatCustodyBalance(t *testing.T) {
	amount := sdk.NewInt64Coin("uve", 10)
	bank := &genesisBankKeeper{balances: map[string]sdk.Coins{
		types.FiatConversionCustodyAccountName: sdk.NewCoins(amount),
	}}
	keeper, ctx := newGenesisTestKeeperWithBank(t, bank)
	provider := sdk.AccAddress(bytes.Repeat([]byte{3}, 20)).String()
	customer := sdk.AccAddress(bytes.Repeat([]byte{4}, 20)).String()
	payout := types.NewPayoutRecord(
		"payout-custody-genesis", "invoice-custody-genesis", "settlement-custody-genesis", "escrow-custody-genesis",
		"order-custody-genesis", "lease-custody-genesis", provider, customer,
		sdk.NewCoins(amount), sdk.NewCoins(), sdk.NewCoins(), sdk.NewCoins(), ctx.BlockTime(), ctx.BlockHeight(),
	)
	payout.FiatConversionID = "conversion-custody-genesis"
	payout.State = types.PayoutStateCompleted
	payout.ExternalFinalityHash = bytes.Repeat([]byte{5}, 32)
	payout.ValueMovementApplied = true
	payout.CompletedAt = timePtr(ctx.BlockTime())

	conversion := types.NewFiatConversionRecord("conversion-custody-genesis", types.FiatConversionRequest{
		InvoiceID: payout.InvoiceID, SettlementID: payout.SettlementID, PayoutID: payout.PayoutID,
		Provider: provider, Customer: customer, RequestedBy: provider,
		CryptoAmount: amount, FiatCurrency: "USD", PaymentMethod: "bank_transfer",
		DestinationHash: types.HashDestination("opaque"), SlippageToleranceExact: "0.010000000000000000",
		CryptoToken: types.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6},
		StableToken: types.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6},
	}, amount, ctx.BlockTime())
	conversion.ProtocolVersion = 1
	conversion.RequestDigest = bytes.Repeat([]byte{7}, 32)
	conversion.DailyBucket = "20231114"
	conversion.DailyQuotaReserved = true
	conversion.PayoutFinalityHash = append([]byte(nil), payout.ExternalFinalityHash...)
	conversion.ValueMovementApplied = true
	conversion.CustodySinkAmount = amount
	effectHash := genesisFiatCustodyEffectHash(*conversion, *payout)
	conversion.CustodySinkEffectHash = effectHash
	payout.ValueMovementEffectHash = append([]byte(nil), effectHash...)
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStateSwapPending, "fixture_swap_pending", "", nil, ctx.BlockTime()))
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStateSwapSubmitted, "fixture_swap_submitted", "", nil, ctx.BlockTime()))
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStateSwapSettled, "fixture_swap_settled", "", nil, ctx.BlockTime()))
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStateOffRampPending, "fixture_offramp_pending", "", nil, ctx.BlockTime()))
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStatePayoutPending, "fixture_payout_pending", "", nil, ctx.BlockTime()))
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStatePayoutSubmitted, "fixture_payout_submitted", "", nil, ctx.BlockTime()))
	require.NoError(t, conversion.TransitionTo(types.FiatConversionStatePayoutCompleted, "fixture_payout_completed", "", nil, ctx.BlockTime()))

	genesis := types.DefaultGenesisState()
	genesis.PayoutRecords = []types.PayoutRecord{*payout}
	genesis.FiatConversionRecords = []types.FiatConversionRecord{*conversion}
	genesis.FiatConversionCustodyBalance = sdk.NewCoins(amount)
	require.NoError(t, genesis.Validate())
	settlement.InitGenesis(ctx, keeper, genesis)

	exported := settlement.ExportGenesis(ctx, keeper)
	require.Equal(t, sdk.NewCoins(amount), exported.FiatConversionCustodyBalance)
	require.NoError(t, exported.Validate())

	wrongDeclared := *genesis
	wrongDeclared.FiatConversionCustodyBalance = sdk.NewCoins()
	require.Error(t, wrongDeclared.Validate())

	wrongBank := &genesisBankKeeper{balances: map[string]sdk.Coins{}}
	wrongKeeper, wrongCtx := newGenesisTestKeeperWithBank(t, wrongBank)
	require.Panics(t, func() { settlement.InitGenesis(wrongCtx, wrongKeeper, genesis) })
}

func newGenesisTestKeeper(t *testing.T) (settlementkeeper.Keeper, sdk.Context) {
	return newGenesisTestKeeperWithBank(t, &genesisBankKeeper{})
}

func newGenesisTestKeeperWithBank(t *testing.T, bank *genesisBankKeeper) (settlementkeeper.Keeper, sdk.Context) {
	t.Helper()
	database := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(database, log.NewNopLogger(), metrics.NewNoOpMetrics())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, database)
	require.NoError(t, stateStore.LoadLatestVersion())
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 10, Time: time.Unix(1_700_000_000, 0).UTC()}, false, log.NewNopLogger())
	c := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	return settlementkeeper.NewKeeper(c, storeKey, bank, nil, sdk.AccAddress(bytes.Repeat([]byte{9}, 20)).String(), nil), ctx
}

func timePtr(value time.Time) *time.Time { return &value }

func genesisFiatCustodyEffectHash(conversion types.FiatConversionRecord, payout types.PayoutRecord) []byte {
	hash := sha256.New()
	for _, value := range []string{
		"virtengine/settlement/fiat-custody-sink/v1", conversion.ConversionID, payout.PayoutID,
		conversion.CryptoAmount.Denom, conversion.CryptoAmount.Amount.String(), types.ModuleAccountName,
		types.FiatConversionCustodyAccountName,
	} {
		writeGenesisCanonicalBytes(hash, []byte(value))
	}
	writeGenesisCanonicalBytes(hash, conversion.PayoutFinalityHash)
	return hash.Sum(nil)
}

func writeGenesisCanonicalBytes(hash interface{ Write([]byte) (int, error) }, value []byte) {
	length := []byte{byte(len(value) >> 24), byte(len(value) >> 16), byte(len(value) >> 8), byte(len(value))}
	_, _ = hash.Write(length)
	_, _ = hash.Write(value)
}

type genesisBankKeeper struct {
	balances map[string]sdk.Coins
}

func (*genesisBankKeeper) SendCoins(context.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
	return nil
}
func (*genesisBankKeeper) SendCoinsFromModuleToModule(context.Context, string, string, sdk.Coins) error {
	return nil
}
func (*genesisBankKeeper) SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error {
	return nil
}
func (*genesisBankKeeper) SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error {
	return nil
}
func (k *genesisBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	if addr.Equals(authtypes.NewModuleAddress(types.FiatConversionCustodyAccountName)) {
		return k.balances[types.FiatConversionCustodyAccountName]
	}
	return sdk.NewCoins()
}
func (k *genesisBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, k.SpendableCoins(ctx, addr).AmountOf(denom))
}
