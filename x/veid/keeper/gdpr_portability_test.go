package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	markettypes "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	delegationtypes "github.com/virtengine/virtengine/x/delegation/types"
	"github.com/virtengine/virtengine/x/veid/types"
)

type stubMarketKeeper struct {
	orders []markettypes.Order
	bids   []markettypes.Bid
	leases []mv1.Lease
}

func (s stubMarketKeeper) WithOrders(_ sdk.Context, fn func(markettypes.Order) bool) {
	for _, order := range s.orders {
		if fn(order) {
			return
		}
	}
}

func (s stubMarketKeeper) WithBids(_ sdk.Context, fn func(markettypes.Bid) bool) {
	for _, bid := range s.bids {
		if fn(bid) {
			return
		}
	}
}

func (s stubMarketKeeper) WithLeases(_ sdk.Context, fn func(mv1.Lease) bool) {
	for _, lease := range s.leases {
		if fn(lease) {
			return
		}
	}
}

type stubEscrowKeeper struct {
	accounts []etypes.Account
	payments []etypes.Payment
}

func (s stubEscrowKeeper) WithAccounts(_ sdk.Context, fn func(etypes.Account) bool) {
	for _, account := range s.accounts {
		if fn(account) {
			return
		}
	}
}

func (s stubEscrowKeeper) WithPayments(_ sdk.Context, fn func(etypes.Payment) bool) {
	for _, payment := range s.payments {
		if fn(payment) {
			return
		}
	}
}

type stubDelegationKeeper struct {
	delegations    []delegationtypes.Delegation
	unbondings     []delegationtypes.UnbondingDelegation
	redelegations  []delegationtypes.Redelegation
	rewards        []delegationtypes.DelegatorReward
	slashingEvents []delegationtypes.DelegatorSlashingEvent
}

func (s stubDelegationKeeper) GetDelegatorDelegations(_ sdk.Context, delegatorAddr string) []delegationtypes.Delegation {
	var out []delegationtypes.Delegation
	for _, del := range s.delegations {
		if del.DelegatorAddress == delegatorAddr {
			out = append(out, del)
		}
	}
	return out
}

func (s stubDelegationKeeper) GetDelegatorUnbondingDelegations(_ sdk.Context, delegatorAddr string) []delegationtypes.UnbondingDelegation {
	var out []delegationtypes.UnbondingDelegation
	for _, del := range s.unbondings {
		if del.DelegatorAddress == delegatorAddr {
			out = append(out, del)
		}
	}
	return out
}

func (s stubDelegationKeeper) GetDelegatorRedelegations(_ sdk.Context, delegatorAddr string) []delegationtypes.Redelegation {
	var out []delegationtypes.Redelegation
	for _, del := range s.redelegations {
		if del.DelegatorAddress == delegatorAddr {
			out = append(out, del)
		}
	}
	return out
}

func (s stubDelegationKeeper) GetDelegatorUnclaimedRewards(_ sdk.Context, delegatorAddr string) []delegationtypes.DelegatorReward {
	var out []delegationtypes.DelegatorReward
	for _, reward := range s.rewards {
		if reward.DelegatorAddress == delegatorAddr {
			out = append(out, reward)
		}
	}
	return out
}

func (s stubDelegationKeeper) GetDelegatorSlashingEvents(_ sdk.Context, delegatorAddr string) []delegationtypes.DelegatorSlashingEvent {
	var out []delegationtypes.DelegatorSlashingEvent
	for _, event := range s.slashingEvents {
		if event.DelegatorAddress == delegatorAddr {
			out = append(out, event)
		}
	}
	return out
}

func TestGDPRPortabilityExportPackageIncludesCrossModuleData(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	defer closeStoreIfNeeded(stateStore)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	k := NewKeeper(cdc, storeKey, "authority")
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	subject := sdk.AccAddress([]byte("gdpr-subject-addr"))
	other := sdk.AccAddress([]byte("gdpr-other-addr"))

	wallet := types.NewIdentityWallet("wallet-1", subject.String(), ctx.BlockTime(), []byte("sig"), []byte("pk"))
	wallet.Tier = types.IdentityTierStandard
	wallet.CurrentScore = 80
	wallet.ScoreStatus = types.AccountStatusVerified
	require.NoError(t, k.SetWallet(ctx, wallet))

	scope := &types.IdentityScope{
		ScopeID:    "scope-1",
		ScopeType:  types.ScopeTypeIDDocument,
		Status:     types.VerificationStatusVerified,
		UploadedAt: ctx.BlockTime(),
	}
	require.NoError(t, k.setScope(ctx, subject, scope))

	orderID := mv1.OrderID{Owner: subject.String(), DSeq: 1, GSeq: 1, OSeq: 1}
	otherOrderID := mv1.OrderID{Owner: other.String(), DSeq: 2, GSeq: 1, OSeq: 1}
	bidID := mv1.BidID{Owner: subject.String(), DSeq: 1, GSeq: 1, OSeq: 1, Provider: subject.String()}
	otherBidID := mv1.BidID{Owner: other.String(), DSeq: 2, GSeq: 1, OSeq: 1, Provider: other.String()}
	leaseID := mv1.LeaseID{Owner: subject.String(), DSeq: 1, GSeq: 1, OSeq: 1, Provider: subject.String()}
	otherLeaseID := mv1.LeaseID{Owner: other.String(), DSeq: 2, GSeq: 1, OSeq: 1, Provider: other.String()}

	marketKeeper := stubMarketKeeper{
		orders: []markettypes.Order{
			{ID: orderID, State: markettypes.OrderActive, CreatedAt: 10},
			{ID: otherOrderID, State: markettypes.OrderActive, CreatedAt: 11},
		},
		bids: []markettypes.Bid{
			{ID: bidID, State: markettypes.BidOpen, CreatedAt: 12},
			{ID: otherBidID, State: markettypes.BidOpen, CreatedAt: 13},
		},
		leases: []mv1.Lease{
			{ID: leaseID, State: mv1.LeaseActive, CreatedAt: 14},
			{ID: otherLeaseID, State: mv1.LeaseActive, CreatedAt: 15},
		},
	}

	escrowKeeper := stubEscrowKeeper{
		accounts: []etypes.Account{
			{ID: escrowid.Account{Scope: escrowid.ScopeDeployment, XID: "scope-1"}, State: etypes.AccountState{Owner: subject.String(), State: etypes.StateOpen}},
			{ID: escrowid.Account{Scope: escrowid.ScopeDeployment, XID: "scope-2"}, State: etypes.AccountState{Owner: other.String(), State: etypes.StateOpen}},
		},
		payments: []etypes.Payment{
			{ID: escrowid.Payment{AID: escrowid.Account{Scope: escrowid.ScopeDeployment, XID: "scope-1"}, XID: "pay-1"}, State: etypes.PaymentState{Owner: subject.String(), State: etypes.StateOpen}},
			{ID: escrowid.Payment{AID: escrowid.Account{Scope: escrowid.ScopeDeployment, XID: "scope-2"}, XID: "pay-2"}, State: etypes.PaymentState{Owner: other.String(), State: etypes.StateOpen}},
		},
	}

	delegationKeeper := stubDelegationKeeper{
		delegations: []delegationtypes.Delegation{
			{DelegatorAddress: subject.String(), ValidatorAddress: "val-1", Shares: "100", CreatedAt: ctx.BlockTime()},
			{DelegatorAddress: other.String(), ValidatorAddress: "val-2", Shares: "50", CreatedAt: ctx.BlockTime()},
		},
		rewards: []delegationtypes.DelegatorReward{
			{DelegatorAddress: subject.String(), ValidatorAddress: "val-1", EpochNumber: 7, Reward: "1000", Claimed: false},
		},
		slashingEvents: []delegationtypes.DelegatorSlashingEvent{
			{DelegatorAddress: subject.String(), ValidatorAddress: "val-1", BlockHeight: 55, SlashAmount: "0.1"},
		},
	}

	k.SetMarketKeeper(marketKeeper)
	k.SetEscrowKeeper(escrowKeeper)
	k.SetDelegationKeeper(delegationKeeper)

	request, err := k.SubmitExportRequest(ctx, subject, []types.ExportCategory{
		types.ExportCategoryIdentity,
		types.ExportCategoryMarketplace,
		types.ExportCategoryEscrow,
		types.ExportCategoryDelegations,
	}, types.ExportFormatJSON)
	require.NoError(t, err)

	pkg, err := k.ProcessExportRequest(ctx, request.RequestID)
	require.NoError(t, err)
	require.NotNil(t, pkg)

	require.NotNil(t, pkg.Identity)
	require.Len(t, pkg.Identity.ActiveScopes, 1)
	require.Equal(t, subject.String(), pkg.Identity.WalletAddress)

	require.NotNil(t, pkg.Marketplace)
	require.Equal(t, 1, pkg.Marketplace.TotalOrders)
	require.Equal(t, 1, pkg.Marketplace.TotalBids)
	require.Equal(t, 1, pkg.Marketplace.TotalLeases)

	require.NotNil(t, pkg.Escrow)
	require.Equal(t, 1, pkg.Escrow.TotalAccounts)
	require.Equal(t, 1, pkg.Escrow.TotalPayments)

	require.NotNil(t, pkg.Delegations)
	require.Len(t, pkg.Delegations.StakingDelegations, 1)
	require.Len(t, pkg.Delegations.Rewards, 1)
	require.Len(t, pkg.Delegations.SlashingEvents, 1)

	audits, total, err := k.ListAuditEntries(ctx, 0, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(1), total)
	require.Len(t, audits, 1)
	require.Equal(t, types.AuditEventTypeGDPRPortability, audits[0].EventType)
}
