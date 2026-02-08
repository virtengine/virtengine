package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

// SetInventory stores a resource inventory snapshot.
func (k Keeper) SetInventory(ctx sdk.Context, inventory types.ResourceInventory) error {
	store := ctx.KVStore(k.skey)
	key := types.InventoryKey(inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
	bz, err := json.Marshal(inventory)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetInventory retrieves an inventory snapshot.
func (k Keeper) GetInventory(ctx sdk.Context, provider string, class types.ResourceClass, inventoryID string) (types.ResourceInventory, bool) {
	store := ctx.KVStore(k.skey)
	key := types.InventoryKey(provider, class, inventoryID)
	bz := store.Get(key)
	if bz == nil {
		return types.ResourceInventory{}, false
	}
	var inventory types.ResourceInventory
	if err := json.Unmarshal(bz, &inventory); err != nil {
		return types.ResourceInventory{}, false
	}
	return inventory, true
}

// WithInventories iterates over all inventories.
func (k Keeper) WithInventories(ctx sdk.Context, fn func(types.ResourceInventory) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.InventoryKeyPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var inventory types.ResourceInventory
		if err := json.Unmarshal(iter.Value(), &inventory); err != nil {
			continue
		}
		if fn(inventory) {
			return
		}
	}
}

// UpdateInventoryFromHeartbeat updates inventory from a heartbeat.
func (k Keeper) UpdateInventoryFromHeartbeat(ctx sdk.Context, msg *types.MsgProviderHeartbeat) (*types.ResourceInventory, error) {
	now := ctx.BlockTime()
	inventoryID := msg.InventoryId
	if inventoryID == "" {
		inventoryID = fmt.Sprintf("%s-%d", msg.ProviderAddress, msg.ResourceClass)
	}

	inventory, found := k.GetInventory(ctx, msg.ProviderAddress, msg.ResourceClass, inventoryID)
	if found {
		if msg.Sequence <= inventory.HeartbeatSequence {
			return nil, types.ErrStaleHeartbeat
		}
	} else {
		inventory = types.ResourceInventory{
			InventoryId:     inventoryID,
			ProviderAddress: msg.ProviderAddress,
			ResourceClass:   msg.ResourceClass,
		}
	}

	inventory.Total = msg.Total
	inventory.Available = msg.Available
	inventory.Locality = msg.Locality
	inventory.HeartbeatSequence = msg.Sequence
	inventory.LastHeartbeat = now
	inventory.UpdatedAt = now
	inventory.Active = true

	if err := k.SetInventory(ctx, inventory); err != nil {
		return nil, err
	}

	return &inventory, nil
}

// PruneStaleInventories marks inventories stale based on heartbeat timeout.
func (k Keeper) PruneStaleInventories(ctx sdk.Context) {
	params := k.GetParams(ctx)
	cutoff := ctx.BlockTime().Add(-secondsToDuration(params.HeartbeatTimeoutSeconds))

	k.WithInventories(ctx, func(inv types.ResourceInventory) bool {
		if inv.LastHeartbeat.Before(cutoff) {
			inv.Active = false
			inv.UpdatedAt = ctx.BlockTime()
			if err := k.SetInventory(ctx, inv); err != nil {
				k.Logger(ctx).Error("failed to mark stale inventory", "provider", inv.ProviderAddress, "error", err)
			}
		}
		return false
	})
}
