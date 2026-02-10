package keeper

import (
	types "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"
)

func NewQuerier(_ Keeper) types.QueryServer {
	return nil
}
