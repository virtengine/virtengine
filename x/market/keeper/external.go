package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	escrowid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"
	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
)

type EscrowKeeper interface {
	GetAccount(ctx sdk.Context, id escrowid.Account) (etypes.Account, error)
	GetPayment(ctx sdk.Context, id escrowid.Payment) (etypes.Payment, error)
	AccountClose(ctx sdk.Context, id escrowid.Account) error
	PaymentClose(ctx sdk.Context, id escrowid.Payment) error
}
