package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	eid "github.com/virtengine/virtengine/sdk/go/node/escrow/id/v1"

	etypes "github.com/virtengine/virtengine/sdk/go/node/escrow/types/v1"
)

type EscrowKeeper interface {
	GetAccount(ctx sdk.Context, id eid.Account) (etypes.Account, error)
}
