package keeper_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type integrationBankKeeper struct {
	balances  map[string]sdk.Coins
	transfers []BankTransfer
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func newIntegrationBankKeeper() *integrationBankKeeper {
	return &integrationBankKeeper{
		balances:  make(map[string]sdk.Coins),
		transfers: []BankTransfer{},
	}
}

func (m *integrationBankKeeper) SetBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.balances[addr.String()] = coins
}

func (m *integrationBankKeeper) SendCoins(_ context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	m.transfers = append(m.transfers, BankTransfer{
		Method: "send",
		From:   fromAddr.String(),
		To:     toAddr.String(),
		Amount: amt,
	})
	return nil
}

func (m *integrationBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.transfers = append(m.transfers, BankTransfer{
		Method: "module_to_account",
		From:   senderModule,
		To:     recipientAddr.String(),
		Amount: amt,
	})
	return nil
}

func (m *integrationBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	m.transfers = append(m.transfers, BankTransfer{
		Method: "account_to_module",
		From:   senderAddr.String(),
		To:     recipientModule,
		Amount: amt,
	})
	balance := m.balances[senderAddr.String()]
	m.balances[senderAddr.String()] = balance.Sub(amt...)
	return nil
}

func (m *integrationBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	if coins, ok := m.balances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

func (m *integrationBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.SpendableCoins(context.Background(), addr)
	return sdk.NewCoin(denom, coins.AmountOf(denom))
}

func TestHPCSettlementFlowsThroughSettlementModule(t *testing.T) {
	cfg := testutilmod.MakeTestEncodingConfig()
	cdc := cfg.Codec

	hpcKey := storetypes.NewKVStoreKey(types.StoreKey)
	settlementKey := storetypes.NewKVStoreKey(settlementtypes.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(hpcKey, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(settlementKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, cmtproto.Header{Time: time.Now().UTC()}, false, log.NewNopLogger())
	bank := newIntegrationBankKeeper()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	hpcKeeper := keeper.NewKeeper(cdc, hpcKey, bank, authority)
	settlementKeeper := settlementkeeper.NewKeeper(cdc, settlementKey, bank, authority, nil)
	hpcKeeper.SetSettlementKeeper(settlementKeeper)

	depositor := sdk.AccAddress([]byte("hpc-depositor-addr"))
	provider := sdk.AccAddress([]byte("hpc-provider-addr"))

	bank.SetBalance(depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100000))))

	escrowID, err := settlementKeeper.CreateEscrow(ctx, "job-1", depositor, sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(5000))), time.Hour, nil)
	require.NoError(t, err)
	require.NoError(t, settlementKeeper.ActivateEscrow(ctx, escrowID, "job-1", provider))

	job := types.HPCJob{
		JobID:             "job-1",
		OfferingID:        "off-1",
		ClusterID:         "cluster-1",
		ProviderAddress:   provider.String(),
		CustomerAddress:   depositor.String(),
		State:             types.JobStateCompleted,
		QueueName:         "default",
		WorkloadSpec:      types.JobWorkloadSpec{ContainerImage: "image"},
		Resources:         types.JobResources{Nodes: 1},
		MaxRuntimeSeconds: 3600,
		AgreedPrice:       sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(5000))),
		EscrowID:          escrowID,
		CreatedAt:         ctx.BlockTime(),
		BlockHeight:       ctx.BlockHeight(),
	}
	require.NoError(t, hpcKeeper.SetJob(ctx, job))

	record := &types.HPCAccountingRecord{
		JobID:           "job-1",
		ClusterID:       "cluster-1",
		ProviderAddress: provider.String(),
		CustomerAddress: depositor.String(),
		OfferingID:      "off-1",
		SchedulerType:   "slurm",
		UsageMetrics: types.HPCDetailedMetrics{
			CPUCoreSeconds:  3600,
			MemoryGBSeconds: 3600,
			StorageGBHours:  10,
			NetworkBytesIn:  1024 * 1024 * 1024,
			NetworkBytesOut: 1024 * 1024 * 1024,
			NodeHours:       sdkmath.LegacyOneDec(),
			SubmitTime:      ctx.BlockTime(),
			StartTime:       ptrTime(ctx.BlockTime()),
			EndTime:         ptrTime(ctx.BlockTime()),
		},
		BillableAmount: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(5000))),
		BillableBreakdown: types.BillableBreakdown{
			CPUCost:     sdk.NewCoin("uve", sdkmath.NewInt(2000)),
			MemoryCost:  sdk.NewCoin("uve", sdkmath.NewInt(1000)),
			StorageCost: sdk.NewCoin("uve", sdkmath.NewInt(500)),
			NetworkCost: sdk.NewCoin("uve", sdkmath.NewInt(500)),
			NodeCost:    sdk.NewCoin("uve", sdkmath.NewInt(1000)),
		},
		ProviderReward: sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(4500))),
		PlatformFee:    sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(500))),
		Status:         types.AccountingStatusPending,
		PeriodStart:    ctx.BlockTime().Add(-time.Hour),
		PeriodEnd:      ctx.BlockTime(),
		FormulaVersion: "v1",
	}

	require.NoError(t, hpcKeeper.CreateAccountingRecord(ctx, record))

	result, err := hpcKeeper.ProcessJobSettlement(ctx, job.JobID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotEmpty(t, result.SettlementID)

	settlementRecord, found := settlementKeeper.GetSettlement(ctx, result.SettlementID)
	require.True(t, found)
	require.Equal(t, job.JobID, settlementRecord.OrderID)
	require.False(t, settlementRecord.TotalAmount.IsZero())

	escrow, found := settlementKeeper.GetEscrow(ctx, escrowID)
	require.True(t, found)
	require.True(t, escrow.TotalSettled.IsAllGTE(settlementRecord.TotalAmount))
}
