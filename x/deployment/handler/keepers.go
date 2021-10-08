package handler

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/virtengine/virtengine/x/deployment/types"
	etypes "github.com/virtengine/virtengine/x/escrow/types"
	mtypes "github.com/virtengine/virtengine/x/market/types"
)

// MarketKeeper Interface includes market methods
type MarketKeeper interface {
	CreateOrder(ctx sdk.Context, id types.GroupID, spec types.GroupSpec) (mtypes.Order, error)
	OnGroupClosed(ctx sdk.Context, id types.GroupID)
}

type EscrowKeeper interface {
	AccountCreate(ctx sdk.Context, id etypes.AccountID, owner sdk.AccAddress, deposit sdk.Coin) error
	AccountDeposit(ctx sdk.Context, id etypes.AccountID, amount sdk.Coin) error
	AccountClose(ctx sdk.Context, id etypes.AccountID) error
}
