package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

const fixedPointScale int64 = 1000000

// SetAllocation stores a resource allocation.
func (k Keeper) SetAllocation(ctx sdk.Context, allocation types.ResourceAllocation) error {
	store := ctx.KVStore(k.skey)
	key := types.AllocationKey(allocation.AllocationId)
	bz, err := json.Marshal(allocation)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetAllocation retrieves a resource allocation.
func (k Keeper) GetAllocation(ctx sdk.Context, allocationID string) (types.ResourceAllocation, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.AllocationKey(allocationID))
	if bz == nil {
		return types.ResourceAllocation{}, false
	}
	var allocation types.ResourceAllocation
	if err := json.Unmarshal(bz, &allocation); err != nil {
		return types.ResourceAllocation{}, false
	}
	return allocation, true
}

// WithAllocations iterates allocations.
func (k Keeper) WithAllocations(ctx sdk.Context, fn func(types.ResourceAllocation) bool) {
	store := prefix.NewStore(ctx.KVStore(k.skey), types.AllocationKeyPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var allocation types.ResourceAllocation
		if err := json.Unmarshal(iter.Value(), &allocation); err != nil {
			continue
		}
		if fn(allocation) {
			return
		}
	}
}

// RecordAllocationEvent stores lifecycle events.
func (k Keeper) RecordAllocationEvent(ctx sdk.Context, allocation types.ResourceAllocation, reason string) error {
	seqKey := types.SequenceKey(types.AllocationEventSeqKeyPrefix, allocation.AllocationId)
	seq := k.nextSequence(ctx, seqKey)
	event := types.AllocationEvent{
		AllocationId:    allocation.AllocationId,
		Sequence:        seq,
		State:           allocation.State,
		Reason:          reason,
		ProviderAddress: allocation.ProviderAddress,
		CreatedAt:       ctx.BlockTime(),
	}

	bz, err := json.Marshal(event)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.AllocationEventKey(allocation.AllocationId, seq), bz)
	return nil
}

// AllocateResources selects a provider and creates a pending allocation.
func (k Keeper) AllocateResources(ctx sdk.Context, request types.ResourceRequest) (*types.ResourceAllocation, error) {
	params := k.GetParams(ctx)
	limit := uint64ToInt(params.MaxCandidates)
	if request.MaxCandidates > 0 {
		requestLimit := uint64ToInt(request.MaxCandidates)
		if requestLimit < limit {
			limit = requestLimit
		}
	}
	candidates := k.selectInventoryCandidates(ctx, request, limit)
	if len(candidates) == 0 {
		return nil, types.ErrNoEligibleInventory
	}

	selected := candidates[0]
	inventory := selected.inventory
	if !capacitySatisfies(inventory.Available, request.Required) {
		return nil, types.ErrNoEligibleInventory
	}

	inventory.Available = subtractCapacity(inventory.Available, request.Required)
	inventory.UpdatedAt = ctx.BlockTime()
	if err := k.SetInventory(ctx, inventory); err != nil {
		return nil, err
	}

	allocationID := fmt.Sprintf("res-alloc-%d", k.nextSequence(ctx, types.SequenceKey(types.AllocationSequenceKeyPrefix, "allocation")))
	allocation := types.ResourceAllocation{
		AllocationId:     allocationID,
		RequestId:        request.RequestId,
		RequesterAddress: request.RequesterAddress,
		ProviderAddress:  inventory.ProviderAddress,
		ResourceClass:    request.ResourceClass,
		Required:         request.Required,
		Assigned:         request.Required,
		State:            types.AllocationStatePending,
		Score:            strconv.FormatInt(selected.combinedScore, 10),
		Locality:         inventory.Locality,
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
		BlockHeight:      ctx.BlockHeight(),
	}

	expiresAt := ctx.BlockTime().Add(secondsToDuration(params.ReservationTimeoutSeconds))
	allocation.ExpiresAt = &expiresAt

	if err := k.SetAllocation(ctx, allocation); err != nil {
		return nil, err
	}

	ctx.KVStore(k.skey).Set(types.AllocationProviderKey(inventory.ProviderAddress, allocationID), []byte{0x01})
	ctx.KVStore(k.skey).Set(types.PendingAllocationKey(unixToUint64(expiresAt.Unix()), allocationID), []byte{0x01})

	if err := k.RecordAllocationEvent(ctx, allocation, "allocation_pending"); err != nil {
		k.Logger(ctx).Error("failed to record allocation event", "error", err)
	}

	return &allocation, nil
}

// ActivateAllocation transitions allocation to active.
func (k Keeper) ActivateAllocation(ctx sdk.Context, allocationID, provider string) (*types.ResourceAllocation, error) {
	allocation, found := k.GetAllocation(ctx, allocationID)
	if !found {
		return nil, types.ErrAllocationNotFound
	}
	if allocation.ProviderAddress != provider {
		return nil, types.ErrUnauthorized
	}
	if allocation.State != types.AllocationStatePending {
		return nil, types.ErrInvalidState
	}
	if allocation.ExpiresAt != nil && ctx.BlockTime().After(*allocation.ExpiresAt) {
		return nil, types.ErrInvalidState
	}

	allocation.State = types.AllocationStateActive
	now := ctx.BlockTime()
	allocation.ActivatedAt = &now
	allocation.UpdatedAt = now

	if err := k.SetAllocation(ctx, allocation); err != nil {
		return nil, err
	}
	if allocation.ExpiresAt != nil {
		ctx.KVStore(k.skey).Delete(types.PendingAllocationKey(unixToUint64(allocation.ExpiresAt.Unix()), allocation.AllocationId))
	}

	if err := k.RecordAllocationEvent(ctx, allocation, "allocation_active"); err != nil {
		k.Logger(ctx).Error("failed to record allocation event", "error", err)
	}

	return &allocation, nil
}

// ReleaseAllocation releases an allocation and returns capacity.
func (k Keeper) ReleaseAllocation(ctx sdk.Context, allocationID, requester, reason string) (*types.ResourceAllocation, error) {
	allocation, found := k.GetAllocation(ctx, allocationID)
	if !found {
		return nil, types.ErrAllocationNotFound
	}
	if allocation.RequesterAddress != requester {
		return nil, types.ErrUnauthorized
	}

	if allocation.State == types.AllocationStateReleased || allocation.State == types.AllocationStateExpired {
		return &allocation, nil
	}

	allocation.State = types.AllocationStateReleased
	allocation.UpdatedAt = ctx.BlockTime()

	if err := k.SetAllocation(ctx, allocation); err != nil {
		return nil, err
	}

	k.restoreInventoryCapacity(ctx, allocation)

	if err := k.RecordAllocationEvent(ctx, allocation, reason); err != nil {
		k.Logger(ctx).Error("failed to record allocation event", "error", err)
	}

	return &allocation, nil
}

// ExpirePendingAllocations marks pending allocations expired and applies slashing.
func (k Keeper) ExpirePendingAllocations(ctx sdk.Context) {
	params := k.GetParams(ctx)
	cutoff := ctx.BlockTime().Add(-secondsToDuration(params.SlashingGraceSeconds))

	store := prefix.NewStore(ctx.KVStore(k.skey), types.PendingAllocationKeyPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) < 9 {
			continue
		}
		expiry := uint64ToInt64(binaryToUint64(key[:8]))
		expiryTime := time.Unix(expiry, 0)
		if expiryTime.After(ctx.BlockTime()) {
			break
		}

		allocationID := string(key[9:])
		allocation, found := k.GetAllocation(ctx, allocationID)
		if !found {
			store.Delete(key)
			continue
		}
		if allocation.State != types.AllocationStatePending {
			store.Delete(key)
			continue
		}

		allocation.State = types.AllocationStateExpired
		allocation.UpdatedAt = ctx.BlockTime()
		if err := k.SetAllocation(ctx, allocation); err != nil {
			k.Logger(ctx).Error("failed to expire allocation", "allocation", allocation.AllocationId, "error", err)
			continue
		}

		store.Delete(key)
		k.restoreInventoryCapacity(ctx, allocation)

		slashReason := "allocation_expired"
		if allocation.ExpiresAt != nil && allocation.ExpiresAt.Before(cutoff) {
			slashReason = "allocation_expired_slash"
			k.recordSlashing(ctx, allocation, params.SlashingPenalty)
		}

		if err := k.RecordAllocationEvent(ctx, allocation, slashReason); err != nil {
			k.Logger(ctx).Error("failed to record allocation event", "error", err)
		}
	}
}

func (k Keeper) recordSlashing(ctx sdk.Context, allocation types.ResourceAllocation, penalty string) {
	entry := types.SlashingEvent{
		ProviderAddress: allocation.ProviderAddress,
		AllocationId:    allocation.AllocationId,
		Reason:          "non_fulfillment",
		Penalty:         penalty,
		CreatedAt:       ctx.BlockTime(),
	}
	bz, err := json.Marshal(entry)
	if err != nil {
		return
	}
	key := types.SequenceKey(types.SlashingEventKeyPrefix, allocation.AllocationId)
	seq := k.nextSequence(ctx, key)
	ctx.KVStore(k.skey).Set(types.SlashingEventKey(allocation.AllocationId, seq), bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"resource_slash",
			sdk.NewAttribute("provider", allocation.ProviderAddress),
			sdk.NewAttribute("allocation_id", allocation.AllocationId),
			sdk.NewAttribute("penalty", penalty),
		),
	)
}

func (k Keeper) restoreInventoryCapacity(ctx sdk.Context, allocation types.ResourceAllocation) {
	var restored bool
	k.WithInventories(ctx, func(inv types.ResourceInventory) bool {
		if inv.ProviderAddress != allocation.ProviderAddress || inv.ResourceClass != allocation.ResourceClass {
			return false
		}
		inv.Available = addCapacity(inv.Available, allocation.Assigned)
		inv.UpdatedAt = ctx.BlockTime()
		if err := k.SetInventory(ctx, inv); err != nil {
			k.Logger(ctx).Error("failed to restore inventory capacity", "error", err)
		}
		restored = true
		return true
	})
	if !restored {
		k.Logger(ctx).Debug("no inventory found to restore", "provider", allocation.ProviderAddress, "allocation", allocation.AllocationId)
	}
}

type inventoryCandidate struct {
	inventory     types.ResourceInventory
	localityScore int64
	capacityScore int64
	combinedScore int64
}

func (k Keeper) selectInventoryCandidates(ctx sdk.Context, request types.ResourceRequest, limit int) []inventoryCandidate {
	params := k.GetParams(ctx)
	cutoff := ctx.BlockTime().Add(-secondsToDuration(params.HeartbeatTimeoutSeconds))
	localityWeight := parseFixedPoint(params.LocalityWeight)
	capacityWeight := parseFixedPoint(params.CapacityWeight)

	candidates := make([]inventoryCandidate, 0)
	k.WithInventories(ctx, func(inv types.ResourceInventory) bool {
		if !inv.Active || inv.ResourceClass != request.ResourceClass {
			return false
		}
		if inv.LastHeartbeat.Before(cutoff) {
			return false
		}
		if !capacitySatisfies(inv.Available, request.Required) {
			return false
		}

		localityScore := computeLocalityScore(inv.Locality, request.Locality)
		capacityScore := computeCapacityScore(inv.Available, request.Required)
		combined := (localityWeight*localityScore + capacityWeight*capacityScore) / fixedPointScale

		candidates = append(candidates, inventoryCandidate{
			inventory:     inv,
			localityScore: localityScore,
			capacityScore: capacityScore,
			combinedScore: combined,
		})
		return false
	})

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].combinedScore == candidates[j].combinedScore {
			if candidates[i].inventory.ProviderAddress == candidates[j].inventory.ProviderAddress {
				return candidates[i].inventory.InventoryId < candidates[j].inventory.InventoryId
			}
			return candidates[i].inventory.ProviderAddress < candidates[j].inventory.ProviderAddress
		}
		return candidates[i].combinedScore > candidates[j].combinedScore
	})

	if limit > 0 && len(candidates) > limit {
		return candidates[:limit]
	}
	return candidates
}

func parseFixedPoint(input string) int64 {
	if input == "" {
		return fixedPointScale
	}
	val, err := strconv.ParseInt(input, 10, 64)
	if err != nil || val <= 0 {
		return fixedPointScale
	}
	return val
}

func computeLocalityScore(inv types.Locality, req types.Locality) int64 {
	score := fixedPointScale
	if req.Region != "" {
		if inv.Region != req.Region {
			score = fixedPointScale / 3
		} else if req.Zone != "" && inv.Zone != req.Zone {
			score = fixedPointScale / 2
		}
	}
	return score
}

func computeCapacityScore(available, required types.ResourceCapacity) int64 {
	var ratios []int64
	if required.CpuCores > 0 {
		ratios = append(ratios, ratioScore(available.CpuCores, required.CpuCores))
	}
	if required.MemoryGb > 0 {
		ratios = append(ratios, ratioScore(available.MemoryGb, required.MemoryGb))
	}
	if required.StorageGb > 0 {
		ratios = append(ratios, ratioScore(available.StorageGb, required.StorageGb))
	}
	if required.NetworkMbps > 0 {
		ratios = append(ratios, ratioScore(available.NetworkMbps, required.NetworkMbps))
	}
	if required.Gpus > 0 {
		ratios = append(ratios, ratioScore(available.Gpus, required.Gpus))
	}
	if len(ratios) == 0 {
		return fixedPointScale
	}
	min := ratios[0]
	for _, r := range ratios[1:] {
		if r < min {
			min = r
		}
	}
	return min
}

func ratioScore(available, required int64) int64 {
	if required <= 0 {
		return fixedPointScale
	}
	if available <= 0 {
		return 0
	}
	score := available * fixedPointScale / required
	if score > fixedPointScale {
		score = fixedPointScale
	}
	return score
}

func capacitySatisfies(available, required types.ResourceCapacity) bool {
	if required.CpuCores > 0 && available.CpuCores < required.CpuCores {
		return false
	}
	if required.MemoryGb > 0 && available.MemoryGb < required.MemoryGb {
		return false
	}
	if required.StorageGb > 0 && available.StorageGb < required.StorageGb {
		return false
	}
	if required.NetworkMbps > 0 && available.NetworkMbps < required.NetworkMbps {
		return false
	}
	if required.Gpus > 0 && available.Gpus < required.Gpus {
		return false
	}
	if required.GpuType != "" && available.GpuType != "" && required.GpuType != available.GpuType {
		return false
	}
	return true
}

func subtractCapacity(available, required types.ResourceCapacity) types.ResourceCapacity {
	available.CpuCores = nonNegativeInt64(available.CpuCores - required.CpuCores)
	available.MemoryGb = nonNegativeInt64(available.MemoryGb - required.MemoryGb)
	available.StorageGb = nonNegativeInt64(available.StorageGb - required.StorageGb)
	available.NetworkMbps = nonNegativeInt64(available.NetworkMbps - required.NetworkMbps)
	available.Gpus = nonNegativeInt64(available.Gpus - required.Gpus)
	return available
}

func addCapacity(available, delta types.ResourceCapacity) types.ResourceCapacity {
	available.CpuCores += delta.CpuCores
	available.MemoryGb += delta.MemoryGb
	available.StorageGb += delta.StorageGb
	available.NetworkMbps += delta.NetworkMbps
	available.Gpus += delta.Gpus
	return available
}

func nonNegativeInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func uint64ToInt(value uint64) int {
	maxInt := int(^uint(0) >> 1)
	if value > uint64(maxInt) {
		return maxInt
	}
	return int(value)
}

func uint64ToInt64(value uint64) int64 {
	maxInt64 := int64(^uint64(0) >> 1)
	if value > uint64(maxInt64) {
		return maxInt64
	}
	return int64(value)
}

func unixToUint64(value int64) uint64 {
	if value <= 0 {
		return 0
	}
	return uint64(value)
}

func secondsToDuration(seconds uint64) time.Duration {
	maxInt64 := int64(^uint64(0) >> 1)
	maxSeconds := maxInt64 / int64(time.Second)
	if seconds > uint64(maxSeconds) {
		return time.Duration(maxSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func binaryToUint64(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}
