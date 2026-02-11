package keeper_test

import (
	"context"
	"fmt"
	"math"
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

	"github.com/virtengine/virtengine/sdk/go/testutil"
	billing "github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

type BankTransfer struct {
	Method string
	From   string
	To     string
	Amount sdk.Coins
}

type MockBankKeeper struct {
	mu         sync.Mutex
	spendable  map[string]sdk.Coins
	transfers  []BankTransfer
	failSends  bool
	errorToUse error
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		spendable: make(map[string]sdk.Coins),
		transfers: []BankTransfer{},
	}
}

func (m *MockBankKeeper) SetSpendable(addr sdk.AccAddress, coins sdk.Coins) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.spendable[addr.String()] = coins
}

func (m *MockBankKeeper) FailTransfers(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failSends = err != nil
	m.errorToUse = err
}

func (m *MockBankKeeper) Transfers() []BankTransfer {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]BankTransfer, len(m.transfers))
	copy(out, m.transfers)
	return out
}

func (m *MockBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	m.mu.Lock()
	defer m.mu.Unlock()
	if coins, ok := m.spendable[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

func (m *MockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.SpendableCoins(context.Background(), addr)
	return sdk.NewCoin(denom, coins.AmountOf(denom))
}

func (m *MockBankKeeper) SendCoins(_ context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return m.recordTransfer("send", fromAddr.String(), toAddr.String(), amt)
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return m.recordTransfer("module_to_account", senderModule, recipientAddr.String(), amt)
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return m.recordTransfer("account_to_module", senderAddr.String(), recipientModule, amt)
}

func (m *MockBankKeeper) recordTransfer(method, from, to string, amt sdk.Coins) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failSends {
		return m.errorToUse
	}
	m.transfers = append(m.transfers, BankTransfer{
		Method: method,
		From:   from,
		To:     to,
		Amount: amt,
	})
	return nil
}

type MockBillingKeeper struct {
	mu              sync.Mutex
	usageRecords    []*billing.UsageRecord
	createdInvoices []*billing.Invoice
	ledgerRecords   []*billing.InvoiceLedgerRecord
	statusUpdates   []*billing.InvoiceLedgerEntry
	paymentEntries  []*billing.InvoiceLedgerEntry
	sequence        uint64
}

func NewMockBillingKeeper() *MockBillingKeeper {
	return &MockBillingKeeper{sequence: 1000}
}

func (m *MockBillingKeeper) SaveUsageRecord(_ sdk.Context, record *billing.UsageRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usageRecords = append(m.usageRecords, record)
	return nil
}

func (m *MockBillingKeeper) CreateInvoice(_ sdk.Context, invoice *billing.Invoice, _ string) (*billing.InvoiceLedgerRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.createdInvoices = append(m.createdInvoices, invoice)

	lineItemCount := uint32(0)
	if len(invoice.LineItems) > math.MaxUint32 {
		lineItemCount = math.MaxUint32
	} else {
		lineItemCount = uint32(len(invoice.LineItems)) //nolint:gosec // bounded by MaxUint32 check above
	}

	record := &billing.InvoiceLedgerRecord{
		InvoiceID:          invoice.InvoiceID,
		InvoiceNumber:      invoice.InvoiceNumber,
		EscrowID:           invoice.EscrowID,
		OrderID:            invoice.OrderID,
		LeaseID:            invoice.LeaseID,
		Provider:           invoice.Provider,
		Customer:           invoice.Customer,
		Status:             invoice.Status,
		Currency:           invoice.Currency,
		Subtotal:           invoice.Subtotal,
		DiscountTotal:      invoice.DiscountTotal,
		TaxTotal:           invoice.TaxTotal,
		Total:              invoice.Total,
		AmountPaid:         invoice.AmountPaid,
		AmountDue:          invoice.AmountDue,
		LineItemCount:      lineItemCount,
		BillingPeriodStart: invoice.BillingPeriod.StartTime,
		BillingPeriodEnd:   invoice.BillingPeriod.EndTime,
		DueDate:            invoice.DueDate,
		IssuedAt:           invoice.IssuedAt,
	}

	m.ledgerRecords = append(m.ledgerRecords, record)
	return record, nil
}

func (m *MockBillingKeeper) UpdateInvoiceStatus(_ sdk.Context, invoiceID string, newStatus billing.InvoiceStatus, initiator string) (*billing.InvoiceLedgerEntry, error) {
	entry := &billing.InvoiceLedgerEntry{
		EntryID:           fmt.Sprintf("entry-%d", time.Now().UnixNano()),
		InvoiceID:         invoiceID,
		EntryType:         billing.LedgerEntryTypeIssued,
		NewStatus:         newStatus,
		Initiator:         initiator,
		SequenceNumber:    1,
		Timestamp:         time.Now().UTC(),
		PreviousEntryHash: "",
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusUpdates = append(m.statusUpdates, entry)
	return entry, nil
}

func (m *MockBillingKeeper) RecordPayment(_ sdk.Context, invoiceID string, amount sdk.Coins, initiator string) (*billing.InvoiceLedgerEntry, error) {
	entry := &billing.InvoiceLedgerEntry{
		EntryID:           fmt.Sprintf("payment-%d", time.Now().UnixNano()),
		InvoiceID:         invoiceID,
		EntryType:         billing.LedgerEntryTypePayment,
		NewStatus:         billing.InvoiceStatusPaid,
		Amount:            amount,
		Initiator:         initiator,
		SequenceNumber:    2,
		Timestamp:         time.Now().UTC(),
		PreviousEntryHash: "",
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.paymentEntries = append(m.paymentEntries, entry)
	return entry, nil
}

func (m *MockBillingKeeper) GetInvoiceSequence(_ sdk.Context) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sequence++
	return m.sequence
}

func (m *MockBillingKeeper) UsageRecords() []*billing.UsageRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*billing.UsageRecord, len(m.usageRecords))
	copy(out, m.usageRecords)
	return out
}

func (m *MockBillingKeeper) CreatedInvoices() []*billing.Invoice {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*billing.Invoice, len(m.createdInvoices))
	copy(out, m.createdInvoices)
	return out
}

func (m *MockBillingKeeper) LedgerRecords() []*billing.InvoiceLedgerRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*billing.InvoiceLedgerRecord, len(m.ledgerRecords))
	copy(out, m.ledgerRecords)
	return out
}

func setupHPCKeeper(t testing.TB) (sdk.Context, keeper.Keeper, *MockBankKeeper) {
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
	bank := NewMockBankKeeper()
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	k := keeper.NewKeeper(cdc, key, bank, authority)
	return ctx, k, bank
}

func setupHPCKeeperWithSettlement(t testing.TB) (sdk.Context, keeper.Keeper, settlementkeeper.Keeper, *MockBankKeeper) {
	t.Helper()

	cfg := testutilmod.MakeTestEncodingConfig()
	cdc := cfg.Codec

	hpcKey := storetypes.NewKVStoreKey(types.StoreKey)
	settlementKey := storetypes.NewKVStoreKey(settlementtypes.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(hpcKey, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(settlementKey, storetypes.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(0, 0)}, false, testutil.Logger(t))
	bank := NewMockBankKeeper()
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	hpcKeeper := keeper.NewKeeper(cdc, hpcKey, bank, authority)
	settlementKeeper := settlementkeeper.NewKeeper(cdc, settlementKey, bank, authority, nil)
	hpcKeeper.SetSettlementKeeper(settlementKeeper)

	return ctx, hpcKeeper, settlementKeeper, bank
}

func mustSetCluster(t testing.TB, ctx sdk.Context, k keeper.Keeper, cluster types.HPCCluster) {
	t.Helper()
	if err := k.SetCluster(ctx, cluster); err != nil {
		t.Fatalf("failed to set cluster: %v", err)
	}
}

func mustSetOffering(t testing.TB, ctx sdk.Context, k keeper.Keeper, offering types.HPCOffering) {
	t.Helper()
	if err := k.SetOffering(ctx, offering); err != nil {
		t.Fatalf("failed to set offering: %v", err)
	}
}

func mustSetJob(t testing.TB, ctx sdk.Context, k keeper.Keeper, job types.HPCJob) {
	t.Helper()
	if err := k.SetJob(ctx, job); err != nil {
		t.Fatalf("failed to set job: %v", err)
	}
}

func mustSetNode(t testing.TB, ctx sdk.Context, k keeper.Keeper, node types.NodeMetadata) {
	t.Helper()
	if err := k.SetNodeMetadata(ctx, node); err != nil {
		t.Fatalf("failed to set node: %v", err)
	}
}
