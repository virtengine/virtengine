package hpc

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type integrationBankKeeper struct {
	transfers []sdk.Coins
}

func (b *integrationBankKeeper) SendCoins(_ context.Context, _ sdk.AccAddress, _ sdk.AccAddress, amt sdk.Coins) error {
	b.transfers = append(b.transfers, amt)
	return nil
}

func (b *integrationBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, amt sdk.Coins) error {
	b.transfers = append(b.transfers, amt)
	return nil
}

func (b *integrationBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, amt sdk.Coins) error {
	b.transfers = append(b.transfers, amt)
	return nil
}

func (b *integrationBankKeeper) SpendableCoins(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

type mockSettlementKeeper struct {
	mu               sync.Mutex
	escrowByOrder    map[string]settlementtypes.EscrowAccount
	escrowByID       map[string]settlementtypes.EscrowAccount
	usageRecords     map[string]settlementtypes.UsageRecord
	usageByOrder     map[string][]string
	financialCases   map[string]settlementtypes.FinancialCase
	usageCounter     int
	settlementNumber int
	bank             *integrationBankKeeper
}

func newMockSettlementKeeper(escrow settlementtypes.EscrowAccount, bank *integrationBankKeeper) *mockSettlementKeeper {
	return &mockSettlementKeeper{
		escrowByOrder: map[string]settlementtypes.EscrowAccount{
			escrow.OrderID: escrow,
		},
		escrowByID: map[string]settlementtypes.EscrowAccount{
			escrow.EscrowID: escrow,
		},
		usageRecords:   make(map[string]settlementtypes.UsageRecord),
		usageByOrder:   make(map[string][]string),
		financialCases: make(map[string]settlementtypes.FinancialCase),
		bank:           bank,
	}
}

func (m *mockSettlementKeeper) OpenFinancialCase(_ sdk.Context, _ settlementkeeper.FinancialCaseOpenRequest) (*settlementtypes.FinancialCase, *settlementtypes.FinancialClaim, bool, error) {
	return nil, nil, false, fmt.Errorf("financial cases are not enabled in this lifecycle fixture")
}

func (m *mockSettlementKeeper) AddFinancialClaim(_ sdk.Context, _ string, _ settlementtypes.FinancialClaim) (*settlementtypes.FinancialCase, *settlementtypes.FinancialClaim, bool, error) {
	return nil, nil, false, fmt.Errorf("financial cases are not enabled in this lifecycle fixture")
}

func (m *mockSettlementKeeper) EscalateFinancialCase(_ sdk.Context, _ string, _ string, _ []byte) error {
	return fmt.Errorf("financial cases are not enabled in this lifecycle fixture")
}

func (m *mockSettlementKeeper) GetFinancialCase(_ sdk.Context, caseID string) (settlementtypes.FinancialCase, bool) {
	financialCase, found := m.financialCases[caseID]
	return financialCase, found
}

func (m *mockSettlementKeeper) GetFinancialCaseBySubject(_ sdk.Context, _ settlementtypes.FinancialSubject) (settlementtypes.FinancialCase, bool) {
	return settlementtypes.FinancialCase{}, false
}

func (m *mockSettlementKeeper) IsFinancialCasesActive(_ sdk.Context) bool { return false }

func (m *mockSettlementKeeper) RecordUsage(_ sdk.Context, record *settlementtypes.UsageRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if record.UsageID == "" {
		m.usageCounter++
		record.UsageID = fmt.Sprintf("usage-%d", m.usageCounter)
	}

	m.usageRecords[record.UsageID] = *record
	m.usageByOrder[record.OrderID] = append(m.usageByOrder[record.OrderID], record.UsageID)
	return nil
}

func (m *mockSettlementKeeper) SettleOrder(
	ctx sdk.Context,
	orderID string,
	usageRecordIDs []string,
	isFinal bool,
) (*settlementtypes.SettlementRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	escrow, ok := m.escrowByOrder[orderID]
	if !ok {
		return nil, fmt.Errorf("escrow not found for order %s", orderID)
	}

	records := usageRecordIDs
	if len(records) == 0 {
		records = m.usageByOrder[orderID]
	}

	total := sdk.NewCoins()
	var provider string
	var customer string
	var totalUnits uint64
	for _, id := range records {
		if record, ok := m.usageRecords[id]; ok {
			total = total.Add(record.TotalCost...)
			totalUnits += record.UsageUnits
			if provider == "" {
				provider = record.Provider
			}
			if customer == "" {
				customer = record.Customer
			}
		}
	}

	if total.IsZero() {
		total = sdk.NewCoins(sdk.NewInt64Coin("uve", 1000))
	}
	if provider == "" {
		provider = escrow.Recipient
	}
	if customer == "" {
		customer = escrow.Depositor
	}

	m.settlementNumber++
	settlementID := fmt.Sprintf("settlement-%d", m.settlementNumber)

	record := settlementtypes.NewSettlementRecord(
		settlementID,
		escrow.EscrowID,
		orderID,
		escrow.LeaseID,
		provider,
		customer,
		total,
		total,
		sdk.NewCoins(),
		sdk.NewCoins(),
		records,
		totalUnits,
		ctx.BlockTime(),
		ctx.BlockTime(),
		settlementtypes.SettlementTypeFinal,
		isFinal,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	if m.bank != nil && !total.IsZero() {
		if recipient, err := sdk.AccAddressFromBech32(provider); err == nil {
			_ = m.bank.SendCoinsFromModuleToAccount(ctx, "settlement", recipient, total)
		}
	}

	return record, nil
}

func (m *mockSettlementKeeper) GetEscrowByOrder(_ sdk.Context, orderID string) (settlementtypes.EscrowAccount, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	escrow, ok := m.escrowByOrder[orderID]
	return escrow, ok
}

func (m *mockSettlementKeeper) GetEscrow(_ sdk.Context, escrowID string) (settlementtypes.EscrowAccount, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	escrow, ok := m.escrowByID[escrowID]
	return escrow, ok
}

func (m *mockSettlementKeeper) GetUsageRecord(_ sdk.Context, usageID string) (settlementtypes.UsageRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.usageRecords[usageID]
	return record, ok
}

func setupIntegrationKeeper(t testing.TB) (sdk.Context, keeper.Keeper, *integrationBankKeeper) {
	t.Helper()

	cfg := testutilmod.MakeTestEncodingConfig()
	cdc := cfg.Codec

	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(0, 0)}, false, testutil.Logger(t))
	bank := &integrationBankKeeper{}

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	k := keeper.NewKeeper(cdc, key, bank, authority)
	return ctx, k, bank
}

func TestJobLifecycleSubmitScheduleRunCompleteSettle(t *testing.T) {
	ctx, k, bank := setupIntegrationKeeper(t)

	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	customerAddr := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	cluster := types.HPCCluster{
		ClusterID:       "cluster-lifecycle",
		ProviderAddress: providerAddr,
		Name:            "Lifecycle Cluster",
		State:           types.ClusterStateActive,
		TotalNodes:      8,
		AvailableNodes:  8,
		Region:          "us-west-2",
	}
	require.NoError(t, k.SetCluster(ctx, cluster))

	mustSetNode := func(nodeID string, latency int64) {
		node := types.NodeMetadata{
			NodeID:          nodeID,
			ClusterID:       cluster.ClusterID,
			ProviderAddress: providerAddr,
			Region:          cluster.Region,
			AvgLatencyMs:    latency,
			Active:          true,
		}
		require.NoError(t, k.SetNodeMetadata(ctx, node))
	}
	mustSetNode("node-life-1", 12)

	offering := types.HPCOffering{
		OfferingID:        "offering-lifecycle",
		ClusterID:         cluster.ClusterID,
		ProviderAddress:   providerAddr,
		Name:              "Lifecycle Offering",
		MaxRuntimeSeconds: 7200,
		Active:            true,
	}
	require.NoError(t, k.SetOffering(ctx, offering))

	job := types.HPCJob{
		JobID:           "job-lifecycle",
		OfferingID:      offering.OfferingID,
		CustomerAddress: customerAddr,
		EscrowID:        "escrow-job-lifecycle",
		QueueName:       "default",
		WorkloadSpec: types.JobWorkloadSpec{
			ContainerImage: "docker.io/library/alpine:latest",
			Command:        "echo",
			Arguments:      []string{"hello"},
		},
		Resources: types.JobResources{
			Nodes:           2,
			CPUCoresPerNode: 4,
			MemoryGBPerNode: 16,
		},
		MaxRuntimeSeconds: 3600,
	}
	// This historical lifecycle fixture isolates HPC accounting/settlement and
	// intentionally predates the Task 84C reservation activation exercised by
	// the marketplace integration suite.
	job.ProviderAddress = providerAddr
	job.ClusterID = cluster.ClusterID
	job.State = types.JobStatePending
	job.CreatedAt = ctx.BlockTime()
	job.BlockHeight = ctx.BlockHeight()
	require.NoError(t, k.SetJob(ctx, job))

	escrow := settlementtypes.EscrowAccount{
		EscrowID:     job.EscrowID,
		OrderID:      job.JobID,
		Depositor:    customerAddr,
		Recipient:    providerAddr,
		Amount:       sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
		Balance:      sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)),
		State:        settlementtypes.EscrowStateActive,
		CreatedAt:    ctx.BlockTime(),
		ExpiresAt:    ctx.BlockTime().Add(24 * time.Hour),
		TotalSettled: sdk.NewCoins(),
		BlockHeight:  ctx.BlockHeight(),
	}

	k.SetSettlementKeeper(newMockSettlementKeeper(escrow, bank))

	decision, err := k.ScheduleJob(ctx, &job)
	require.NoError(t, err)
	require.Equal(t, cluster.ClusterID, decision.SelectedClusterID)

	ctx = ctx.WithBlockTime(time.Unix(1_700_000_100, 0))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateQueued, "queued", 0, nil))
	ctx = ctx.WithBlockTime(time.Unix(1_700_000_200, 0))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateRunning, "running", 0, nil))
	ctx = ctx.WithBlockTime(time.Unix(1_700_000_800, 0))
	require.NoError(t, k.UpdateJobStatus(ctx, job.JobID, types.JobStateCompleted, "completed", 0, nil))

	usageMetrics := realSchedulerMetricsFixture(10*time.Minute, 2, 4, 16, 0, 40)
	require.Greater(t, usageMetrics.StorageGBHours, int64(0))
	require.Greater(t, usageMetrics.NetworkBytesIn, int64(0))
	require.Greater(t, usageMetrics.EnergyJoules, int64(0))
	require.Equal(t, "mixed", usageMetrics.SchedulerSpecific["slurm_state"])
	usage := settlementtypes.UsageRecord{
		UsageID: "usage-lifecycle-1", OrderID: job.JobID, LeaseID: "lease-lifecycle",
		Provider: providerAddr, Customer: customerAddr, UsageUnits: 1, UsageType: "hpc",
		PeriodStart: time.Unix(1_700_000_200, 0), PeriodEnd: time.Unix(1_700_000_800, 0),
		TotalCost: sdk.NewCoins(sdk.NewInt64Coin("uve", 1000)), ProviderSignature: []byte("provider-proof"),
		SignatureVersion: settlementtypes.SignatureVersionV1, SignatureVerified: true,
		UsageDigest:          bytes.Repeat([]byte{84}, settlementtypes.DigestSize),
		CustomerAcknowledged: true, AuthenticationStatus: settlementtypes.UsageAuthenticationStatusVerified,
	}
	settlementKeeper := newMockSettlementKeeper(escrow, bank)
	require.NoError(t, settlementKeeper.RecordUsage(ctx, &usage))
	k.SetSettlementKeeper(settlementKeeper)
	detailedMetrics := types.HPCDetailedMetrics{
		WallClockSeconds: 600, CPUCoreSeconds: 4_800, MemoryGBSeconds: 19_200,
		StorageGBHours: usageMetrics.StorageGBHours, NetworkBytesIn: usageMetrics.NetworkBytesIn,
		NetworkBytesOut: usageMetrics.NetworkBytesOut, NodesUsed: 2, EnergyJoules: usageMetrics.EnergyJoules,
		SubmitTime: time.Unix(1_700_000_100, 0),
	}
	require.NoError(t, k.CreateUsageSnapshot(ctx, &types.HPCUsageSnapshot{
		SnapshotID: "usage-lifecycle-1", JobID: job.JobID, ClusterID: cluster.ClusterID,
		SchedulerType: "SLURM", SnapshotType: types.SnapshotTypeFinal, SequenceNumber: 1,
		ProviderAddress: providerAddr, CustomerAddress: customerAddr,
		Metrics: detailedMetrics, CumulativeMetrics: detailedMetrics, JobState: types.JobStateCompleted,
		SnapshotTime: ctx.BlockTime(), ProviderSignature: "authenticated-provider-proof",
	}))

	result, err := k.ProcessJobSettlement(ctx, job.JobID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotEmpty(t, result.SettlementID)
	require.NotEmpty(t, bank.transfers)
}
