package handler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	atypes "github.com/virtengine/virtengine/x/audit/types"
	dtypes "github.com/virtengine/virtengine/x/deployment/types"
	etypes "github.com/virtengine/virtengine/x/escrow/types"
	"github.com/virtengine/virtengine/x/market/keeper"
	ptypes "github.com/virtengine/virtengine/x/provider/types"
)

type EscrowKeeper interface {
	AccountCreate(ctx sdk.Context, id etypes.AccountID, owner sdk.AccAddress, deposit sdk.Coin) error
	AccountDeposit(ctx sdk.Context, id etypes.AccountID, amount sdk.Coin) error
	AccountClose(ctx sdk.Context, id etypes.AccountID) error
	PaymentCreate(ctx sdk.Context, id etypes.AccountID, pid string, owner sdk.AccAddress, rate sdk.Coin) error
	PaymentWithdraw(ctx sdk.Context, id etypes.AccountID, pid string) error
	PaymentClose(ctx sdk.Context, id etypes.AccountID, pid string) error
}

// ProviderKeeper Interface includes provider methods
type ProviderKeeper interface {
	Get(ctx sdk.Context, id sdk.Address) (ptypes.Provider, bool)
	WithProviders(ctx sdk.Context, fn func(ptypes.Provider) bool)
}

type AuditKeeper interface {
	GetProviderAttributes(ctx sdk.Context, id sdk.Address) (atypes.Providers, bool)
}

// DeploymentKeeper Interface includes deployment methods
type DeploymentKeeper interface {
	GetGroup(ctx sdk.Context, id dtypes.GroupID) (dtypes.Group, bool)
	OnBidClosed(ctx sdk.Context, id dtypes.GroupID) error
	OnLeaseClosed(ctx sdk.Context, id dtypes.GroupID) (dtypes.Group, error)
}

// Keepers include all modules keepers
type Keepers struct {
	Escrow     EscrowKeeper
	Market     keeper.IKeeper
	Deployment DeploymentKeeper
	Provider   ProviderKeeper
	Audit      AuditKeeper
	Bank       bankkeeper.Keeper
}
