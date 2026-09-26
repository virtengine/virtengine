package keeper

import (
	"encoding/json"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

// SetMilestoneDispute stores an order-linked milestone dispute.
//
// Disputes are keyed by dispute id and stored independently of the order, so a
// hold on one milestone never rewrites the order's lifecycle state.
func (k Keeper) SetMilestoneDispute(ctx sdk.Context, dispute *marketplace.MilestoneDispute) error {
	if dispute == nil {
		return marketplace.ErrDisputeNotFound.Wrap("nil dispute")
	}
	if err := dispute.Validate(); err != nil {
		return err
	}

	bz, err := json.Marshal(dispute)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(marketplace.MilestoneDisputeKey(dispute.DisputeID), bz)
	return nil
}

// GetMilestoneDispute returns a milestone dispute by id.
func (k Keeper) GetMilestoneDispute(ctx sdk.Context, disputeID string) (*marketplace.MilestoneDispute, bool) {
	bz := ctx.KVStore(k.skey).Get(marketplace.MilestoneDisputeKey(disputeID))
	if bz == nil {
		return nil, false
	}

	var dispute marketplace.MilestoneDispute
	if err := json.Unmarshal(bz, &dispute); err != nil {
		return nil, false
	}
	return &dispute, true
}

// WithMilestoneDisputes iterates over milestone disputes in key order, which is
// deterministic for a given store.
func (k Keeper) WithMilestoneDisputes(ctx sdk.Context, fn func(marketplace.MilestoneDispute) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, marketplace.MilestoneDisputeKeyPrefix)
	defer func() { _ = iter.Close() }()

	for ; iter.Valid(); iter.Next() {
		var dispute marketplace.MilestoneDispute
		if err := json.Unmarshal(iter.Value(), &dispute); err != nil {
			continue
		}
		if fn(dispute) {
			return
		}
	}
}
