package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

const currentReservationVersion uint32 = 1

func (k Keeper) ActivateCanonicalReservations(ctx sdk.Context) {
	ctx.KVStore(k.skey).Set(types.CanonicalReservationsActivationKey(), []byte{1})
}

func (k Keeper) IsCanonicalReservationsActive(ctx sdk.Context) bool {
	return ctx.KVStore(k.skey).Has(types.CanonicalReservationsActivationKey())
}

type reservationReplay struct {
	ReservationID string `json:"reservation_id"`
	PayloadHash   []byte `json:"payload_hash"`
}

// Reserve atomically and idempotently claims authoritative inventory capacity.
func (k Keeper) Reserve(ctx sdk.Context, request types.ReservationRequest) (*types.Reservation, error) {
	if err := validateReservationRequest(request); err != nil {
		return nil, err
	}
	payloadHash, err := reservationRequestHash(request)
	if err != nil {
		return nil, err
	}
	store := ctx.KVStore(k.skey)
	if existingBytes := store.Get(types.ReservationIdempotencyKey(request.IdempotencyKey)); existingBytes != nil {
		var replay reservationReplay
		if err := json.Unmarshal(existingBytes, &replay); err != nil || !bytes.Equal(replay.PayloadHash, payloadHash) {
			return nil, types.ErrReservationConflict
		}
		existing, found := k.GetReservation(ctx, replay.ReservationID)
		if !found {
			return nil, types.ErrCapacityInvariant.Wrap("idempotency index references missing reservation")
		}
		return &existing, nil
	}

	inventory, err := k.selectReservationInventory(ctx, request)
	if err != nil {
		return nil, err
	}
	if err := k.checkReservationEligibility(ctx, inventory, request); err != nil {
		return nil, err
	}
	available, err := subtractCapacityChecked(inventory.Available, request.Capacity)
	if err != nil {
		return nil, types.ErrNoEligibleInventory.Wrap(err.Error())
	}

	reservationID := request.IdempotencyKey
	if reservationID == "" {
		reservationID = fmt.Sprintf("reservation-%d", k.nextSequence(ctx, types.SequenceKey(types.ReservationSequenceKeyPrefix, "reservation")))
	}
	if _, found := k.GetReservation(ctx, reservationID); found {
		return nil, types.ErrReservationConflict.Wrap("reservation ID already exists")
	}

	now := ctx.BlockTime()
	expiresAt := request.ExpiresAt
	if expiresAt == nil && request.ExpiresHeight == 0 {
		value := now.Add(secondsToDuration(k.GetParams(ctx).ReservationTimeoutSeconds))
		expiresAt = &value
	}
	version := request.Version
	if version == 0 {
		version = currentReservationVersion
	}
	reservation := types.Reservation{
		ReservationId:    reservationID,
		IdempotencyKey:   request.IdempotencyKey,
		PayloadHash:      append([]byte(nil), payloadHash...),
		RequestId:        request.RequestId,
		RequesterAddress: request.RequesterAddress,
		ProviderAddress:  inventory.ProviderAddress,
		InventoryId:      inventory.InventoryId,
		ResourceClass:    request.ResourceClass,
		Capacity:         request.Capacity,
		State:            types.ReservationStatePending,
		ConsumerType:     request.ConsumerType,
		ConsumerId:       request.ConsumerId,
		MarketOrderId:    request.MarketOrderId,
		MarketBidId:      request.MarketBidId,
		MarketLeaseId:    request.MarketLeaseId,
		HpcJobId:         request.HpcJobId,
		EscrowId:         request.EscrowId,
		CollateralId:     request.CollateralId,
		Version:          version,
		LegacySource:     request.LegacySource,
		LegacyReference:  request.LegacyReference,
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        expiresAt,
		ExpiresHeight:    request.ExpiresHeight,
		CreatedHeight:    ctx.BlockHeight(),
	}

	inventory.Available = available
	inventory.UpdatedAt = now
	if err := k.SetInventory(ctx, inventory); err != nil {
		return nil, err
	}
	if err := k.SetReservation(ctx, reservation); err != nil {
		return nil, err
	}
	replayBytes, err := json.Marshal(reservationReplay{ReservationID: reservationID, PayloadHash: payloadHash})
	if err != nil {
		return nil, err
	}
	store.Set(types.ReservationIdempotencyKey(request.IdempotencyKey), replayBytes)
	k.setReservationIndexes(ctx, reservation)
	if err := k.recordReservationEvent(ctx, reservation, types.ReservationStateUnspecified, types.ReservationStatePending, "reserved"); err != nil {
		return nil, err
	}
	return &reservation, nil
}

// ActivateReservation binds the one executable consumer and activates capacity.
func (k Keeper) ActivateReservation(ctx sdk.Context, reservationID string, link types.ReservationLink) (*types.Reservation, error) {
	reservation, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == types.ReservationStateActive && reservation.ConsumerType == link.ConsumerType && reservation.ConsumerId == link.ConsumerId {
		return &reservation, nil
	}
	if !types.CanTransitionReservation(reservation.State, types.ReservationStateActive) {
		return nil, types.ErrInvalidReservationTransition
	}
	if link.ConsumerType == "" || link.ConsumerId == "" {
		return nil, types.ErrLineageConflict.Wrap("consumer type and ID required")
	}
	if reservation.ConsumerId != "" && (reservation.ConsumerId != link.ConsumerId || reservation.ConsumerType != link.ConsumerType) {
		return nil, types.ErrLineageConflict.Wrap("reservation already names a different consumer")
	}
	if err := applyReservationLink(&reservation, link); err != nil {
		return nil, err
	}
	from := reservation.State
	now := ctx.BlockTime()
	reservation.State = types.ReservationStateActive
	reservation.ActivatedAt = &now
	reservation.ActivatedHeight = ctx.BlockHeight()
	reservation.UpdatedAt = now
	if err := k.SetReservation(ctx, reservation); err != nil {
		return nil, err
	}
	k.clearReservationExpiryIndexes(ctx, reservation)
	k.setReservationIndexes(ctx, reservation)
	if err := k.recordReservationEvent(ctx, reservation, from, reservation.State, "activated"); err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (k Keeper) ConsumeReservation(ctx sdk.Context, reservationID string, amount types.ResourceCapacity, reason string) (*types.Reservation, error) {
	reservation, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == types.ReservationStateConsumed && capacityEqual(reservation.Consumed, amount) {
		return &reservation, nil
	}
	if !types.CanTransitionReservation(reservation.State, types.ReservationStateConsumed) {
		return nil, types.ErrInvalidReservationTransition
	}
	if err := validateCapacity(amount, false); err != nil || !capacitySatisfies(reservation.Capacity, amount) {
		return nil, types.ErrInvalidRequest.Wrap("consumed capacity exceeds reservation")
	}
	from := reservation.State
	now := ctx.BlockTime()
	reservation.State = types.ReservationStateConsumed
	reservation.Consumed = amount
	reservation.ConsumedAt = &now
	reservation.ConsumedHeight = ctx.BlockHeight()
	reservation.UpdatedAt = now
	reservation.Reason = reason
	if err := k.SetReservation(ctx, reservation); err != nil {
		return nil, err
	}
	if err := k.recordReservationEvent(ctx, reservation, from, reservation.State, reason); err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (k Keeper) ReleaseReservation(ctx sdk.Context, reservationID, reason string) (*types.Reservation, error) {
	return k.transitionReservation(ctx, reservationID, types.ReservationStateReleased, reason, true)
}

func (k Keeper) QuarantineReservation(ctx sdk.Context, reservationID, reason string) (*types.Reservation, error) {
	return k.transitionReservation(ctx, reservationID, types.ReservationStateQuarantined, reason, false)
}

func (k Keeper) SlashReservation(ctx sdk.Context, reservationID, reason string) (*types.Reservation, error) {
	reservation, err := k.transitionReservation(ctx, reservationID, types.ReservationStateSlashed, reason, true)
	if err == nil {
		k.recordReservationSlashing(ctx, *reservation, reason)
	}
	return reservation, err
}

func (k Keeper) transitionReservation(ctx sdk.Context, reservationID string, target types.ReservationState, reason string, restore bool) (*types.Reservation, error) {
	reservation, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == target {
		return &reservation, nil
	}
	if types.IsTerminalReservationState(reservation.State) || !types.CanTransitionReservation(reservation.State, target) {
		return nil, types.ErrInvalidReservationTransition
	}
	from := reservation.State
	now := ctx.BlockTime()
	reservation.State = target
	reservation.Reason = reason
	reservation.UpdatedAt = now
	switch target {
	case types.ReservationStateReleased:
		reservation.ReleasedAt = &now
		reservation.ReleasedHeight = ctx.BlockHeight()
	case types.ReservationStateExpired:
		reservation.ExpiredAt = &now
	case types.ReservationStateQuarantined:
		reservation.QuarantinedAt = &now
		reservation.QuarantinedHeight = ctx.BlockHeight()
	case types.ReservationStateSlashed:
		reservation.SlashedAt = &now
		reservation.SlashedHeight = ctx.BlockHeight()
	}
	if restore {
		if err := k.restoreReservationCapacity(ctx, reservation); err != nil {
			return nil, err
		}
	}
	if err := k.SetReservation(ctx, reservation); err != nil {
		return nil, err
	}
	k.clearReservationExpiryIndexes(ctx, reservation)
	if err := k.recordReservationEvent(ctx, reservation, from, target, reason); err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (k Keeper) SetReservation(ctx sdk.Context, reservation types.Reservation) error {
	bz, err := json.Marshal(reservation)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ReservationKey(reservation.ReservationId), bz)
	return nil
}

func (k Keeper) GetReservation(ctx sdk.Context, reservationID string) (types.Reservation, bool) {
	bz := ctx.KVStore(k.skey).Get(types.ReservationKey(reservationID))
	if bz == nil {
		return types.Reservation{}, false
	}
	var reservation types.Reservation
	if err := json.Unmarshal(bz, &reservation); err != nil {
		return types.Reservation{}, false
	}
	return reservation, true
}

func (k Keeper) WithReservations(ctx sdk.Context, fn func(types.Reservation) bool) {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var reservation types.Reservation
		if err := json.Unmarshal(iter.Value(), &reservation); err == nil && fn(reservation) {
			return
		}
	}
}

func (k Keeper) GetReservationByLineage(ctx sdk.Context, kind, id string) (types.Reservation, bool) {
	var key []byte
	switch kind {
	case "order":
		key = types.ReservationLineageKey(types.ReservationOrderKeyPrefix, id)
	case "bid":
		key = types.ReservationLineageKey(types.ReservationBidKeyPrefix, id)
	case "lease":
		key = types.ReservationLineageKey(types.ReservationLeaseKeyPrefix, id)
	case "job":
		key = types.ReservationLineageKey(types.ReservationJobKeyPrefix, id)
	case "consumer":
		// Consumer IDs are unique across the canonical market/HPC namespaces.
		var result types.Reservation
		var found bool
		k.WithReservations(ctx, func(reservation types.Reservation) bool {
			if reservation.ConsumerId == id {
				result, found = reservation, true
				return true
			}
			return false
		})
		return result, found
	default:
		return types.Reservation{}, false
	}
	reservationID := ctx.KVStore(k.skey).Get(key)
	if reservationID == nil {
		return types.Reservation{}, false
	}
	return k.GetReservation(ctx, string(reservationID))
}

func (k Keeper) GetReservationByConsumer(ctx sdk.Context, consumerType, consumerID string) (types.Reservation, bool) {
	reservationID := ctx.KVStore(k.skey).Get(types.ReservationConsumerKey(consumerType, consumerID))
	if reservationID == nil {
		return types.Reservation{}, false
	}
	return k.GetReservation(ctx, string(reservationID))
}

func (k Keeper) GetReservationsByProvider(ctx sdk.Context, provider string) []types.Reservation {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationProviderPrefix(provider))
	defer iter.Close()
	result := make([]types.Reservation, 0)
	for ; iter.Valid(); iter.Next() {
		id := string(iter.Key()[len(types.ReservationProviderPrefix(provider)):])
		if reservation, found := k.GetReservation(ctx, id); found {
			result = append(result, reservation)
		}
	}
	return result
}

func (k Keeper) ReservationEvents(ctx sdk.Context, reservationID string) []types.ReservationEvent {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationEventPrefix(reservationID))
	defer iter.Close()
	result := make([]types.ReservationEvent, 0)
	for ; iter.Valid(); iter.Next() {
		var event types.ReservationEvent
		if err := json.Unmarshal(iter.Value(), &event); err == nil {
			result = append(result, event)
		}
	}
	return result
}

// ExpireReservations processes ordered time/height indexes and never scans reservations.
func (k Keeper) ExpireReservations(ctx sdk.Context, limit uint64) (uint64, error) {
	if limit == 0 {
		limit = 1000
	}
	processed := uint64(0)
	if err := k.expireReservationIndex(ctx, types.ReservationExpiryTimeKeyPrefix, unixToUint64(ctx.BlockTime().Unix()), limit, &processed); err != nil {
		return processed, err
	}
	if processed < limit {
		height := positiveHeight(ctx.BlockHeight())
		if err := k.expireReservationIndex(ctx, types.ReservationExpiryHeightKeyPrefix, height, limit, &processed); err != nil {
			return processed, err
		}
	}
	return processed, nil
}

func positiveHeight(height int64) uint64 {
	if height <= 0 {
		return 0
	}
	return uint64(height) //nolint:gosec // positive int64 always fits uint64
}

func (k Keeper) expireReservationIndex(ctx sdk.Context, prefix []byte, cutoff, limit uint64, processed *uint64) error {
	store := ctx.KVStore(k.skey)
	end := append(append([]byte{}, prefix...), make([]byte, 8)...)
	binary.BigEndian.PutUint64(end[len(prefix):], cutoff)
	end = append(end, 0x01)
	iter := store.Iterator(prefix, end)
	defer iter.Close()
	keys := make([][]byte, 0)
	for ; iter.Valid() && *processed+uint64(len(keys)) < limit; iter.Next() {
		keys = append(keys, append([]byte(nil), iter.Key()...))
	}
	for _, key := range keys {
		if len(key) < len(prefix)+9 {
			store.Delete(key)
			continue
		}
		reservationID := string(key[len(prefix)+9:])
		reservation, found := k.GetReservation(ctx, reservationID)
		if !found || reservation.State != types.ReservationStatePending {
			store.Delete(key)
			continue
		}
		if _, err := k.transitionReservation(ctx, reservationID, types.ReservationStateExpired, "reservation_expired", true); err != nil {
			return err
		}
		(*processed)++
	}
	return nil
}

func (k Keeper) ValidateCapacityConservation(ctx sdk.Context) error {
	reservationsByInventory := make(map[string]types.ResourceCapacity)
	consumerOwners := make(map[string]string)
	var invariantErr error
	k.WithReservations(ctx, func(reservation types.Reservation) bool {
		allowZero := reservation.State == types.ReservationStateQuarantined && reservation.LegacySource != ""
		if err := validateCapacity(reservation.Capacity, allowZero); err != nil {
			invariantErr = types.ErrCapacityInvariant.Wrapf("reservation %s: %v", reservation.ReservationId, err)
			return true
		}
		if reservation.State == types.ReservationStatePending || reservation.State == types.ReservationStateActive || reservation.State == types.ReservationStateConsumed || reservation.State == types.ReservationStateQuarantined {
			key := inventoryIdentity(reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
			total, err := addCapacityChecked(reservationsByInventory[key], reservation.Capacity)
			if err != nil {
				invariantErr = err
				return true
			}
			reservationsByInventory[key] = total
		}
		if reservation.State == types.ReservationStateActive || reservation.State == types.ReservationStateConsumed || reservation.State == types.ReservationStateQuarantined {
			consumerKey := reservation.ConsumerType + "\x00" + reservation.ConsumerId
			if owner, exists := consumerOwners[consumerKey]; exists && owner != reservation.ReservationId {
				invariantErr = types.ErrCapacityInvariant.Wrapf("consumer %s has reservations %s and %s", consumerKey, owner, reservation.ReservationId)
				return true
			}
			consumerOwners[consumerKey] = reservation.ReservationId
		}
		return false
	})
	if invariantErr != nil {
		return invariantErr
	}
	k.WithInventories(ctx, func(inventory types.ResourceInventory) bool {
		if err := validateCapacity(inventory.Total, true); err != nil {
			invariantErr = types.ErrCapacityInvariant.Wrapf("inventory %s total: %v", inventory.InventoryId, err)
			return true
		}
		if err := validateCapacity(inventory.Available, true); err != nil {
			invariantErr = types.ErrCapacityInvariant.Wrapf("inventory %s available: %v", inventory.InventoryId, err)
			return true
		}
		key := inventoryIdentity(inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
		sum, err := addCapacityChecked(inventory.Available, reservationsByInventory[key])
		if err != nil || !capacityEqual(sum, inventory.Total) {
			invariantErr = types.ErrCapacityInvariant.Wrapf("inventory %s total does not equal available plus reservations", inventory.InventoryId)
			return true
		}
		return false
	})
	return invariantErr
}

func (k Keeper) reservedCapacityForInventory(ctx sdk.Context, provider string, class types.ResourceClass, inventoryID string) (types.ResourceCapacity, error) {
	total := types.ResourceCapacity{}
	for _, reservation := range k.GetReservationsByProvider(ctx, provider) {
		if reservation.InventoryId != inventoryID || reservation.ResourceClass != class {
			continue
		}
		if reservation.State != types.ReservationStatePending && reservation.State != types.ReservationStateActive && reservation.State != types.ReservationStateConsumed && reservation.State != types.ReservationStateQuarantined {
			continue
		}
		var err error
		total, err = addCapacityChecked(total, reservation.Capacity)
		if err != nil {
			return types.ResourceCapacity{}, err
		}
	}
	return total, nil
}

func (k Keeper) handleInventoryUnavailable(ctx sdk.Context, inventory types.ResourceInventory, reason string) error {
	reservations := k.GetReservationsByProvider(ctx, inventory.ProviderAddress)
	for _, reservation := range reservations {
		if reservation.InventoryId != inventory.InventoryId || reservation.ResourceClass != inventory.ResourceClass {
			continue
		}
		switch reservation.State {
		case types.ReservationStatePending:
			if _, err := k.transitionReservation(ctx, reservation.ReservationId, types.ReservationStateExpired, reason, true); err != nil {
				return err
			}
		case types.ReservationStateActive, types.ReservationStateConsumed:
			if _, err := k.QuarantineReservation(ctx, reservation.ReservationId, reason); err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) setReservationIndexes(ctx sdk.Context, reservation types.Reservation) {
	store := ctx.KVStore(k.skey)
	store.Set(types.ReservationProviderKey(reservation.ProviderAddress, reservation.ReservationId), []byte{1})
	indexes := []struct {
		prefix []byte
		id     string
	}{
		{types.ReservationOrderKeyPrefix, reservation.MarketOrderId}, {types.ReservationBidKeyPrefix, reservation.MarketBidId},
		{types.ReservationLeaseKeyPrefix, reservation.MarketLeaseId}, {types.ReservationJobKeyPrefix, reservation.HpcJobId},
	}
	for _, index := range indexes {
		if index.id != "" {
			store.Set(types.ReservationLineageKey(index.prefix, index.id), []byte(reservation.ReservationId))
		}
	}
	if reservation.ConsumerType != "" && reservation.ConsumerId != "" {
		store.Set(types.ReservationConsumerKey(reservation.ConsumerType, reservation.ConsumerId), []byte(reservation.ReservationId))
	}
	if reservation.State == types.ReservationStatePending {
		if reservation.ExpiresAt != nil {
			store.Set(types.ReservationExpiryTimeKey(unixToUint64(reservation.ExpiresAt.Unix()), reservation.ReservationId), []byte{1})
		}
		if reservation.ExpiresHeight > 0 {
			store.Set(types.ReservationExpiryHeightKey(uint64(reservation.ExpiresHeight), reservation.ReservationId), []byte{1})
		}
	}
}

// RebuildReservationIndexes deterministically restores derived genesis/migration indexes.
func (k Keeper) RebuildReservationIndexes(ctx sdk.Context, reservation types.Reservation) {
	k.setReservationIndexes(ctx, reservation)
}

func (k Keeper) clearReservationExpiryIndexes(ctx sdk.Context, reservation types.Reservation) {
	store := ctx.KVStore(k.skey)
	if reservation.ExpiresAt != nil {
		store.Delete(types.ReservationExpiryTimeKey(unixToUint64(reservation.ExpiresAt.Unix()), reservation.ReservationId))
	}
	if reservation.ExpiresHeight > 0 {
		store.Delete(types.ReservationExpiryHeightKey(uint64(reservation.ExpiresHeight), reservation.ReservationId))
	}
}

func (k Keeper) recordReservationEvent(ctx sdk.Context, reservation types.Reservation, from, to types.ReservationState, reason string) error {
	sequence := k.nextSequence(ctx, types.SequenceKey(types.ReservationEventSeqKeyPrefix, reservation.ReservationId))
	event := types.ReservationEvent{ReservationId: reservation.ReservationId, Sequence: sequence, FromState: from, ToState: to, Reason: reason, ProviderAddress: reservation.ProviderAddress, ConsumerType: reservation.ConsumerType, CreatedAt: ctx.BlockTime(), BlockHeight: ctx.BlockHeight()}
	bz, err := json.Marshal(event)
	if err != nil {
		return err
	}
	ctx.KVStore(k.skey).Set(types.ReservationEventKey(reservation.ReservationId, sequence), bz)
	ctx.EventManager().EmitEvent(sdk.NewEvent("resource_reservation_transition",
		sdk.NewAttribute("reservation_id", reservation.ReservationId), sdk.NewAttribute("provider", reservation.ProviderAddress),
		sdk.NewAttribute("consumer_type", reservation.ConsumerType), sdk.NewAttribute("from_state", from.String()), sdk.NewAttribute("to_state", to.String())))
	return nil
}

func (k Keeper) restoreReservationCapacity(ctx sdk.Context, reservation types.Reservation) error {
	inventory, found := k.GetInventory(ctx, reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
	if !found {
		return types.ErrInventoryNotFound.Wrapf("%s/%s", reservation.ProviderAddress, reservation.InventoryId)
	}
	available, err := addCapacityChecked(inventory.Available, reservation.Capacity)
	if err != nil || !capacitySatisfies(inventory.Total, available) {
		return types.ErrCapacityInvariant.Wrap("release would exceed declared capacity")
	}
	inventory.Available = available
	inventory.UpdatedAt = ctx.BlockTime()
	return k.SetInventory(ctx, inventory)
}

func (k Keeper) recordReservationSlashing(ctx sdk.Context, reservation types.Reservation, reason string) {
	entry := types.SlashingEvent{ProviderAddress: reservation.ProviderAddress, AllocationId: reservation.ReservationId, Reason: reason, Penalty: k.GetParams(ctx).SlashingPenalty, CreatedAt: ctx.BlockTime()}
	bz, err := json.Marshal(entry)
	if err != nil {
		return
	}
	sequence := k.nextSequence(ctx, types.SequenceKey(types.SlashingEventKeyPrefix, reservation.ReservationId))
	ctx.KVStore(k.skey).Set(types.SlashingEventKey(reservation.ReservationId, sequence), bz)
}

func (k Keeper) selectReservationInventory(ctx sdk.Context, request types.ReservationRequest) (types.ResourceInventory, error) {
	if request.ProviderAddress != "" && request.InventoryId != "" {
		inventory, found := k.GetInventory(ctx, request.ProviderAddress, request.ResourceClass, request.InventoryId)
		if !found || !inventory.Active || !capacitySatisfies(inventory.Available, request.Capacity) {
			return types.ResourceInventory{}, types.ErrNoEligibleInventory
		}
		return inventory, nil
	}
	candidates := k.selectInventoryCandidates(ctx, types.ResourceRequest{RequestId: request.RequestId, RequesterAddress: request.RequesterAddress, ResourceClass: request.ResourceClass, Required: request.Capacity, Locality: request.Locality}, 0)
	for _, candidate := range candidates {
		if request.ProviderAddress == "" || candidate.inventory.ProviderAddress == request.ProviderAddress {
			return candidate.inventory, nil
		}
	}
	return types.ResourceInventory{}, types.ErrNoEligibleInventory
}

func (k Keeper) checkReservationEligibility(ctx sdk.Context, inventory types.ResourceInventory, request types.ReservationRequest) error {
	if k.eligibilityKeeper == nil {
		if request.RequireBenchmark || request.RequireAttestation || request.RequireCollateral {
			return types.ErrEligibilityUnavailable
		}
		return nil
	}
	provider, err := sdk.AccAddressFromBech32(inventory.ProviderAddress)
	if err != nil || !k.eligibilityKeeper.IsProviderEligible(ctx, provider) {
		return types.ErrProviderIneligible
	}
	if request.RequireBenchmark && !k.eligibilityKeeper.HasCurrentBenchmark(ctx, inventory.ProviderAddress) {
		return types.ErrProviderIneligible.Wrap("benchmark requirement not met")
	}
	if request.RequireAttestation && !k.eligibilityKeeper.HasCurrentAttestation(ctx, inventory.ProviderAddress) {
		return types.ErrProviderIneligible.Wrap("attestation requirement not met")
	}
	if request.RequireCollateral && !k.eligibilityKeeper.HasSufficientCollateral(ctx, inventory.ProviderAddress, request.CollateralId) {
		return types.ErrProviderIneligible.Wrap("collateral requirement not met")
	}
	return nil
}

func validateReservationRequest(request types.ReservationRequest) error {
	if request.IdempotencyKey == "" || request.RequestId == "" || request.RequesterAddress == "" || request.ConsumerType == "" || request.ConsumerId == "" {
		return types.ErrInvalidRequest.Wrap("idempotency, request, requester, and consumer fields are required")
	}
	if request.ResourceClass == types.ResourceClassUnspecified {
		return types.ErrInvalidRequest.Wrap("resource class required")
	}
	if request.Version > currentReservationVersion {
		return types.ErrInvalidRequest.Wrap("unsupported reservation version")
	}
	return validateCapacity(request.Capacity, false)
}

func reservationRequestHash(request types.ReservationRequest) ([]byte, error) {
	// No maps are present; protobuf JSON field order is not used. This internal
	// canonical struct has declaration-order encoding/json output.
	bz, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(bz)
	return digest[:], nil
}

func applyReservationLink(reservation *types.Reservation, link types.ReservationLink) error {
	pairs := [][2]*string{{&reservation.ConsumerType, &link.ConsumerType}, {&reservation.ConsumerId, &link.ConsumerId}, {&reservation.MarketOrderId, &link.MarketOrderId}, {&reservation.MarketBidId, &link.MarketBidId}, {&reservation.MarketLeaseId, &link.MarketLeaseId}, {&reservation.HpcJobId, &link.HpcJobId}, {&reservation.EscrowId, &link.EscrowId}, {&reservation.CollateralId, &link.CollateralId}}
	for _, pair := range pairs {
		if *pair[0] != "" && *pair[1] != "" && *pair[0] != *pair[1] {
			return types.ErrLineageConflict
		}
		if *pair[0] == "" {
			*pair[0] = *pair[1]
		}
	}
	return nil
}

func validateCapacity(capacity types.ResourceCapacity, allowZero bool) error {
	values := []int64{capacity.CpuCores, capacity.MemoryGb, capacity.StorageGb, capacity.NetworkMbps, capacity.Gpus}
	allZero := true
	for _, value := range values {
		if value < 0 {
			return types.ErrInvalidRequest.Wrap("negative capacity dimension")
		}
		allZero = allZero && value == 0
	}
	if !allowZero && allZero {
		return types.ErrInvalidRequest.Wrap("capacity is empty")
	}
	return nil
}

func addCapacityChecked(a, b types.ResourceCapacity) (types.ResourceCapacity, error) {
	if err := validateCapacity(a, true); err != nil {
		return types.ResourceCapacity{}, err
	}
	if err := validateCapacity(b, true); err != nil {
		return types.ResourceCapacity{}, err
	}
	valuesA := []*int64{&a.CpuCores, &a.MemoryGb, &a.StorageGb, &a.NetworkMbps, &a.Gpus}
	valuesB := []int64{b.CpuCores, b.MemoryGb, b.StorageGb, b.NetworkMbps, b.Gpus}
	for i, value := range valuesB {
		if value > math.MaxInt64-*valuesA[i] {
			return types.ResourceCapacity{}, types.ErrCapacityOverflow
		}
		*valuesA[i] += value
	}
	if a.GpuType == "" {
		a.GpuType = b.GpuType
	}
	if a.GpuType != "" && b.GpuType != "" && a.GpuType != b.GpuType {
		return types.ResourceCapacity{}, types.ErrCapacityInvariant.Wrap("GPU type mismatch")
	}
	return a, nil
}

func subtractCapacityChecked(a, b types.ResourceCapacity) (types.ResourceCapacity, error) {
	if err := validateCapacity(a, true); err != nil {
		return types.ResourceCapacity{}, err
	}
	if err := validateCapacity(b, true); err != nil {
		return types.ResourceCapacity{}, err
	}
	if !capacitySatisfies(a, b) {
		return types.ResourceCapacity{}, types.ErrNoEligibleInventory
	}
	a.CpuCores -= b.CpuCores
	a.MemoryGb -= b.MemoryGb
	a.StorageGb -= b.StorageGb
	a.NetworkMbps -= b.NetworkMbps
	a.Gpus -= b.Gpus
	return a, nil
}

func capacityEqual(a, b types.ResourceCapacity) bool {
	return a.CpuCores == b.CpuCores && a.MemoryGb == b.MemoryGb && a.StorageGb == b.StorageGb && a.NetworkMbps == b.NetworkMbps && a.Gpus == b.Gpus && a.GpuType == b.GpuType
}

func inventoryIdentity(provider string, class types.ResourceClass, inventoryID string) string {
	return fmt.Sprintf("%s\x00%d\x00%s", provider, class, inventoryID)
}
