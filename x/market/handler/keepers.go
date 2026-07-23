package handler

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	atypes "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
	dtypes "github.com/virtengine/virtengine/sdk/go/node/deployment/v1"
	dbeta "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4"
	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
	mv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	types "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	ptypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

type EscrowKeeper interface {
	AccountCreate(ctx sdk.Context, id escrowid.Account, owner sdk.AccAddress, deposits []etypes.Depositor) error
	AccountDeposit(ctx sdk.Context, id escrowid.Account, deposits []etypes.Depositor) error
	AccountClose(ctx sdk.Context, id escrowid.Account) error
	PaymentCreate(ctx sdk.Context, id escrowid.Payment, provider sdk.AccAddress, rate sdk.DecCoin) error
	PaymentWithdraw(ctx sdk.Context, id escrowid.Payment) error
	PaymentClose(ctx sdk.Context, id escrowid.Payment) error
	AuthorizeDeposits(sctx sdk.Context, msg sdk.Msg) ([]etypes.Depositor, error)
}

// ProviderKeeper Interface includes provider methods
type ProviderKeeper interface {
	Get(ctx sdk.Context, id sdk.Address) (ptypes.Provider, bool)
	WithProviders(ctx sdk.Context, fn func(ptypes.Provider) bool)
}

type AuditKeeper interface {
	GetProviderAttributes(ctx sdk.Context, id sdk.Address) (atypes.AuditedProviders, bool)
}

// DeploymentKeeper Interface includes deployment methods
type DeploymentKeeper interface {
	GetGroup(ctx sdk.Context, id dtypes.GroupID) (dbeta.Group, bool)
	OnBidClosed(ctx sdk.Context, id dtypes.GroupID) error
	OnLeaseClosed(ctx sdk.Context, id dtypes.GroupID) (dbeta.Group, error)
}

type AuthzKeeper interface {
	DeleteGrant(ctx context.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error
	GetAuthorization(ctx context.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	SaveGrant(ctx context.Context, grantee sdk.AccAddress, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error
	GetGranteeGrantsByMsgType(ctx context.Context, grantee sdk.AccAddress, msgType string, onGrant authzkeeper.OnGrantFn)
}

type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// MarketLifecycleKeeper is the narrow mutable market surface used by MsgServer.
type MarketLifecycleKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	GetAuthority() string
	SetParams(ctx sdk.Context, params types.Params) error
	BidCountForOrder(ctx sdk.Context, id mv1.OrderID) uint32
	GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool)
	GetBid(ctx sdk.Context, id mv1.BidID) (types.Bid, bool)
	GetLease(ctx sdk.Context, id mv1.LeaseID) (mv1.Lease, bool)
	CreateBid(ctx sdk.Context, id mv1.BidID, price sdk.DecCoin, offer types.ResourcesOffer) (types.Bid, error)
	CreateLease(ctx sdk.Context, bid types.Bid) error
	CreateOrder(ctx sdk.Context, id dtypes.GroupID, spec dbeta.GroupSpec) (types.Order, error)
	OnOrderMatched(ctx sdk.Context, order types.Order)
	OnBidMatched(ctx sdk.Context, bid types.Bid)
	OnBidLost(ctx sdk.Context, bid types.Bid)
	OnBidClosed(ctx sdk.Context, bid types.Bid) error
	OnOrderClosed(ctx sdk.Context, order types.Order) error
	OnLeaseClosed(ctx sdk.Context, lease mv1.Lease, state mv1.Lease_State, reason mv1.LeaseClosedReason) error
	WithBidsForOrder(ctx sdk.Context, id mv1.OrderID, state types.Bid_State, fn func(types.Bid) bool)
	WithOrders(ctx sdk.Context, fn func(types.Order) bool)
	WithBids(ctx sdk.Context, fn func(types.Bid) bool)
}

// ResourcesKeeper is the authoritative capacity mutation surface.
type ResourcesKeeper interface {
	Reserve(ctx sdk.Context, request resourcesv1.ReservationRequest) (*resourcesv1.Reservation, error)
	ActivateReservation(ctx sdk.Context, reservationID string, link resourcesv1.ReservationLink) (*resourcesv1.Reservation, error)
	ReleaseReservation(ctx sdk.Context, reservationID, reason string) (*resourcesv1.Reservation, error)
}

// Keepers include all modules keepers
type Keepers struct {
	Escrow     EscrowKeeper
	Market     MarketLifecycleKeeper
	Resources  ResourcesKeeper
	Deployment DeploymentKeeper
	Provider   ProviderKeeper
	Audit      AuditKeeper
	Account    govtypes.AccountKeeper
	Authz      AuthzKeeper
	Bank       BankKeeper
}
