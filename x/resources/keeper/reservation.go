package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/resources/types"
)

const currentReservationVersion uint32 = 1

const (
	maxReservationIdentifierLength = 512
	maxReservationReasonLength     = 1024
)

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
	cacheCtx, write := ctx.CacheContext()
	reservation, err := k.reserve(cacheCtx, request)
	if err != nil {
		return nil, err
	}
	write()
	return reservation, nil
}

func (k Keeper) reserve(ctx sdk.Context, request types.ReservationRequest) (*types.Reservation, error) {
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

	reservationID := "reservation/" + hex.EncodeToString(payloadHash)
	if _, found := k.GetReservation(ctx, reservationID); found {
		return nil, types.ErrReservationConflict.Wrap("reservation ID already exists")
	}

	now := ctx.BlockTime()
	expiresAt := request.ExpiresAt
	if expiresAt == nil && request.ExpiresHeight == 0 {
		value := now.Add(secondsToDuration(k.GetParams(ctx).ReservationTimeoutSeconds))
		expiresAt = &value
	}
	if expiresAt != nil && !expiresAt.After(now) {
		return nil, types.ErrInvalidRequest.Wrap("expires_at must be after block time")
	}
	if request.ExpiresHeight != 0 && request.ExpiresHeight <= ctx.BlockHeight() {
		return nil, types.ErrInvalidRequest.Wrap("expires_height must be after block height")
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
	if err := k.validateReservationIndexOwnership(ctx, reservation); err != nil {
		return nil, err
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
	if err := k.setReservationIndexes(ctx, reservation); err != nil {
		return nil, err
	}
	if err := k.recordReservationEvent(ctx, reservation, types.ReservationStateUnspecified, types.ReservationStatePending, "reserved"); err != nil {
		return nil, err
	}
	return &reservation, nil
}

// ActivateReservation binds the one executable consumer and activates capacity.
func (k Keeper) ActivateReservation(ctx sdk.Context, reservationID string, link types.ReservationLink) (*types.Reservation, error) {
	cacheCtx, write := ctx.CacheContext()
	reservation, err := k.activateReservation(cacheCtx, reservationID, link)
	if err != nil {
		return nil, err
	}
	write()
	return reservation, nil
}

func (k Keeper) activateReservation(ctx sdk.Context, reservationID string, link types.ReservationLink) (*types.Reservation, error) {
	reservation, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == types.ReservationStateActive {
		if reservationLinkEqual(reservation, link) {
			return &reservation, nil
		}
		return nil, types.ErrLineageConflict.Wrap("active reservation retry payload differs")
	}
	if !types.CanTransitionReservation(reservation.State, types.ReservationStateActive) {
		return nil, types.ErrInvalidReservationTransition
	}
	if (reservation.ExpiresAt != nil && !ctx.BlockTime().Before(*reservation.ExpiresAt)) ||
		(reservation.ExpiresHeight > 0 && ctx.BlockHeight() >= reservation.ExpiresHeight) {
		return nil, types.ErrInvalidReservationTransition.Wrap("reservation has expired")
	}
	inventory, found := k.GetInventory(ctx, reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
	cutoff := ctx.BlockTime().Add(-secondsToDuration(k.GetParams(ctx).HeartbeatTimeoutSeconds))
	if !found || !inventory.Active || inventory.LastHeartbeat.Before(cutoff) {
		return nil, types.ErrNoEligibleInventory.Wrap("reservation inventory is unavailable")
	}
	if err := k.checkReservationEligibility(ctx, inventory, types.ReservationRequest{}); err != nil {
		return nil, err
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
	if err := k.validateReservationIndexOwnership(ctx, reservation); err != nil {
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
	if err := k.setReservationIndexes(ctx, reservation); err != nil {
		return nil, err
	}
	if err := k.recordReservationEvent(ctx, reservation, from, reservation.State, "activated"); err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (k Keeper) ConsumeReservation(ctx sdk.Context, reservationID string, amount types.ResourceCapacity, reason string) (*types.Reservation, error) {
	if err := validateReservationReason(reason); err != nil {
		return nil, err
	}
	cacheCtx, write := ctx.CacheContext()
	reservation, err := k.consumeReservation(cacheCtx, reservationID, amount, reason)
	if err != nil {
		return nil, err
	}
	write()
	return reservation, nil
}

func (k Keeper) consumeReservation(ctx sdk.Context, reservationID string, amount types.ResourceCapacity, reason string) (*types.Reservation, error) {
	reservation, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == types.ReservationStateConsumed {
		if capacityEqual(reservation.Consumed, amount) && reservation.Reason == reason {
			return &reservation, nil
		}
		return nil, types.ErrReservationConflict.Wrap("consumption retry payload differs")
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
	existing, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if existing.State == types.ReservationStateSlashed {
		if existing.Reason == reason {
			return &existing, nil
		}
		return nil, types.ErrReservationConflict.Wrap("slashing retry reason differs")
	}
	cacheCtx, write := ctx.CacheContext()
	reservation, err := k.transitionReservationCached(cacheCtx, reservationID, types.ReservationStateSlashed, reason, true)
	if err != nil {
		return nil, err
	}
	if err := k.recordReservationSlashing(cacheCtx, *reservation, reason); err != nil {
		return nil, err
	}
	write()
	return reservation, err
}

// HoldReservationForFinancialCase retains capacity exactly once while the
// canonical settlement case is active.
func (k Keeper) HoldReservationForFinancialCase(ctx sdk.Context, reservationID, caseID string) (*types.Reservation, error) {
	if caseID == "" {
		return nil, types.ErrInvalidRequest.Wrap("financial case ID required")
	}
	cacheCtx, write := ctx.CacheContext()
	reservation, found := k.GetReservation(cacheCtx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == types.ReservationStateDisputed {
		if reservation.FinancialCaseId == caseID {
			return &reservation, nil
		}
		return nil, types.ErrReservationConflict.Wrap("reservation held by another financial case")
	}
	if types.IsTerminalReservationState(reservation.State) {
		return nil, types.ErrInvalidReservationTransition
	}
	if !types.CanTransitionReservation(reservation.State, types.ReservationStateDisputed) {
		return nil, types.ErrInvalidReservationTransition
	}
	from := reservation.State
	now := cacheCtx.BlockTime()
	reservation.PreDisputeState = int32(from)
	reservation.State = types.ReservationStateDisputed
	reservation.FinancialCaseId = caseID
	reservation.DisputedAt = &now
	reservation.DisputedHeight = cacheCtx.BlockHeight()
	reservation.UpdatedAt = now
	reservation.Reason = "canonical_financial_case"
	if err := k.SetReservation(cacheCtx, reservation); err != nil {
		return nil, err
	}
	k.clearReservationExpiryIndexes(cacheCtx, reservation)
	if err := k.recordReservationEvent(cacheCtx, reservation, from, reservation.State, "canonical_financial_case"); err != nil {
		return nil, err
	}
	write()
	return &reservation, nil
}

// ReleaseReservationFinancialCaseHold restores the exact pre-dispute state.
func (k Keeper) ReleaseReservationFinancialCaseHold(ctx sdk.Context, reservationID, caseID string) (*types.Reservation, error) {
	cacheCtx, write := ctx.CacheContext()
	reservation, found := k.GetReservation(cacheCtx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State != types.ReservationStateDisputed {
		if reservation.FinancialCaseId == "" {
			return &reservation, nil
		}
		return nil, types.ErrReservationConflict.Wrap("reservation is not disputed")
	}
	if reservation.FinancialCaseId != caseID {
		return nil, types.ErrReservationConflict.Wrap("financial case ID mismatch")
	}
	target := types.ReservationState(reservation.PreDisputeState)
	if target != types.ReservationStatePending && target != types.ReservationStateActive && target != types.ReservationStateConsumed {
		return nil, types.ErrInvalidReservationTransition.Wrap("invalid pre-dispute state")
	}
	from := reservation.State
	reservation.State = target
	reservation.FinancialCaseId = ""
	reservation.PreDisputeState = 0
	reservation.DisputedAt = nil
	reservation.DisputedHeight = 0
	reservation.UpdatedAt = cacheCtx.BlockTime()
	reservation.Reason = "financial_case_hold_released"
	if err := k.SetReservation(cacheCtx, reservation); err != nil {
		return nil, err
	}
	if err := k.setReservationIndexes(cacheCtx, reservation); err != nil {
		return nil, err
	}
	if err := k.recordReservationEvent(cacheCtx, reservation, from, target, "financial_case_hold_released"); err != nil {
		return nil, err
	}
	write()
	return &reservation, nil
}

// FinalizeReservationFinancialCase applies the terminal capacity effect once.
func (k Keeper) FinalizeReservationFinancialCase(ctx sdk.Context, reservationID, caseID string, slash bool) (*types.Reservation, error) {
	cacheCtx, write := ctx.CacheContext()
	reservation, found := k.GetReservation(cacheCtx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	target := types.ReservationStateReleased
	reason := "financial_case_finalized"
	if slash {
		target, reason = types.ReservationStateSlashed, "financial_case_fraud_confirmed"
	}
	if reservation.State == target && reservation.Reason == reason {
		return &reservation, nil
	}
	if reservation.State != types.ReservationStateDisputed || reservation.FinancialCaseId != caseID {
		return nil, types.ErrReservationConflict.Wrap("financial case reservation hold missing")
	}
	updated, err := k.transitionReservationCached(cacheCtx, reservationID, target, reason, true)
	if err != nil {
		return nil, err
	}
	updated.FinancialCaseId = ""
	updated.PreDisputeState = 0
	updated.DisputedAt = nil
	updated.DisputedHeight = 0
	if err := k.SetReservation(cacheCtx, *updated); err != nil {
		return nil, err
	}
	if slash {
		if err := k.recordReservationSlashing(cacheCtx, *updated, reason); err != nil {
			return nil, err
		}
	}
	write()
	return updated, nil
}

func (k Keeper) transitionReservation(ctx sdk.Context, reservationID string, target types.ReservationState, reason string, restore bool) (*types.Reservation, error) {
	if err := validateReservationReason(reason); err != nil {
		return nil, err
	}
	cacheCtx, write := ctx.CacheContext()
	reservation, err := k.transitionReservationCached(cacheCtx, reservationID, target, reason, restore)
	if err != nil {
		return nil, err
	}
	write()
	return reservation, nil
}

func (k Keeper) transitionReservationCached(ctx sdk.Context, reservationID string, target types.ReservationState, reason string, restore bool) (*types.Reservation, error) {
	reservation, found := k.GetReservation(ctx, reservationID)
	if !found {
		return nil, types.ErrReservationNotFound
	}
	if reservation.State == target {
		if reservation.Reason == reason {
			return &reservation, nil
		}
		return nil, types.ErrReservationConflict.Wrap("transition retry reason differs")
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
	if reservation.ReservationId == "" {
		return types.ErrInvalidRequest.Wrap("reservation ID required")
	}
	allowZero := reservation.LegacySource != "" && (reservation.State == types.ReservationStateQuarantined || types.IsTerminalReservationState(reservation.State))
	if err := validateCapacity(reservation.Capacity, allowZero); err != nil {
		return err
	}
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

// ValidateReservationStore rejects malformed records that iterator helpers
// deliberately skip so a corrupt reservation cannot disappear from invariants.
func (k Keeper) ValidateReservationStore(ctx sdk.Context) error {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var reservation types.Reservation
		if err := json.Unmarshal(iter.Value(), &reservation); err != nil {
			return types.ErrCapacityInvariant.Wrapf("malformed reservation record at %x", iter.Key())
		}
		if reservation.ReservationId == "" || !bytes.Equal(iter.Key(), types.ReservationKey(reservation.ReservationId)) {
			return types.ErrCapacityInvariant.Wrapf("reservation key does not match record at %x", iter.Key())
		}
		if err := k.validateReservationIndexOwnership(ctx, reservation); err != nil {
			return types.ErrCapacityInvariant.Wrapf("reservation %s index: %v", reservation.ReservationId, err)
		}
	}
	return nil
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
	reservations, _ := k.reservationsByProvider(ctx, provider)
	return reservations
}

func (k Keeper) reservationsByProvider(ctx sdk.Context, provider string) ([]types.Reservation, error) {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationProviderPrefix(provider))
	defer iter.Close()
	result := make([]types.Reservation, 0)
	for ; iter.Valid(); iter.Next() {
		id := string(iter.Key()[len(types.ReservationProviderPrefix(provider)):])
		if reservation, found := k.GetReservation(ctx, id); found {
			result = append(result, reservation)
		} else {
			return nil, types.ErrCapacityInvariant.Wrapf("provider index references missing reservation %s", id)
		}
	}
	return result, nil
}

func (k Keeper) ReservationEvents(ctx sdk.Context, reservationID string) ([]types.ReservationEvent, error) {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationEventPrefix(reservationID))
	defer iter.Close()
	result := make([]types.ReservationEvent, 0)
	for ; iter.Valid(); iter.Next() {
		var event types.ReservationEvent
		if err := json.Unmarshal(iter.Value(), &event); err != nil {
			return nil, types.ErrCapacityInvariant.Wrapf("malformed reservation event at %x", iter.Key())
		}
		result = append(result, event)
	}
	return result, nil
}

// SetReservationEvent restores a validated transition event and advances its sequence.
func (k Keeper) SetReservationEvent(ctx sdk.Context, event types.ReservationEvent) error {
	if event.ReservationId == "" || event.Sequence == 0 {
		return types.ErrInvalidRequest.Wrap("reservation event identity required")
	}
	if _, found := k.GetReservation(ctx, event.ReservationId); !found {
		return types.ErrReservationNotFound
	}
	key := types.ReservationEventKey(event.ReservationId, event.Sequence)
	store := ctx.KVStore(k.skey)
	if store.Has(key) {
		return types.ErrReservationConflict.Wrap("reservation event already exists")
	}
	bz, err := json.Marshal(event)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	sequenceKey := types.SequenceKey(types.ReservationEventSeqKeyPrefix, event.ReservationId)
	current := uint64(0)
	if sequenceBytes := store.Get(sequenceKey); len(sequenceBytes) == 8 {
		current = binary.BigEndian.Uint64(sequenceBytes)
	}
	if event.Sequence > current {
		sequenceBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceBytes, event.Sequence)
		store.Set(sequenceKey, sequenceBytes)
	}
	return nil
}

// WithReservationEvents iterates every stored event in deterministic key order.
func (k Keeper) WithReservationEvents(ctx sdk.Context, fn func(types.ReservationEvent) bool) error {
	iter := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.ReservationEventKeyPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var event types.ReservationEvent
		if err := json.Unmarshal(iter.Value(), &event); err != nil {
			return types.ErrCapacityInvariant.Wrapf("malformed reservation event at %x", iter.Key())
		}
		if fn(event) {
			return nil
		}
	}
	return nil
}

// ExpireReservations processes ordered time/height indexes and never scans reservations.
func (k Keeper) ExpireReservations(ctx sdk.Context, limit uint64) (uint64, error) {
	if limit == 0 {
		limit = 1000
	}
	processed := uint64(0)
	seen := make(map[string]struct{})
	if err := k.expireReservationIndex(ctx, types.ReservationExpiryTimeKeyPrefix, unixToUint64(ctx.BlockTime().Unix()), limit, &processed, seen); err != nil {
		return processed, err
	}
	if processed < limit {
		height := positiveHeight(ctx.BlockHeight())
		if err := k.expireReservationIndex(ctx, types.ReservationExpiryHeightKeyPrefix, height, limit, &processed, seen); err != nil {
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

func (k Keeper) expireReservationIndex(ctx sdk.Context, prefix []byte, cutoff, limit uint64, processed *uint64, seen map[string]struct{}) error {
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
		if _, duplicate := seen[reservationID]; duplicate {
			store.Delete(key)
			continue
		}
		if _, err := k.transitionReservation(ctx, reservationID, types.ReservationStateExpired, "reservation_expired", true); err != nil {
			return err
		}
		seen[reservationID] = struct{}{}
		(*processed)++
	}
	return nil
}

func (k Keeper) ValidateCapacityConservation(ctx sdk.Context) error {
	if err := k.ValidateReservationStore(ctx); err != nil {
		return err
	}
	reservationsByInventory := make(map[string]types.ResourceCapacity)
	inventories := make(map[string]struct{})
	consumerOwners := make(map[string]string)
	var invariantErr error
	k.WithReservations(ctx, func(reservation types.Reservation) bool {
		allowZero := reservation.State == types.ReservationStateQuarantined && reservation.LegacySource != ""
		if err := validateCapacity(reservation.Capacity, allowZero); err != nil {
			invariantErr = types.ErrCapacityInvariant.Wrapf("reservation %s: %v", reservation.ReservationId, err)
			return true
		}
		if reservation.State < types.ReservationStatePending || reservation.State > types.ReservationStateDisputed || reservation.ExpiresHeight < 0 || reservation.CreatedHeight < 0 || reservation.ActivatedHeight < 0 || reservation.ConsumedHeight < 0 || reservation.ReleasedHeight < 0 || reservation.QuarantinedHeight < 0 || reservation.SlashedHeight < 0 || reservation.DisputedHeight < 0 {
			invariantErr = types.ErrCapacityInvariant.Wrapf("reservation %s has invalid state or lifecycle height", reservation.ReservationId)
			return true
		}
		if reservation.State == types.ReservationStateDisputed && (reservation.FinancialCaseId == "" || reservation.PreDisputeState == 0) {
			invariantErr = types.ErrCapacityInvariant.Wrapf("disputed reservation %s has no financial case or prior state", reservation.ReservationId)
			return true
		}
		if reservation.State != types.ReservationStateDisputed && (reservation.FinancialCaseId != "" || reservation.PreDisputeState != 0 || reservation.DisputedAt != nil || reservation.DisputedHeight != 0) {
			invariantErr = types.ErrCapacityInvariant.Wrapf("reservation %s retains a financial-case binding outside disputed state", reservation.ReservationId)
			return true
		}
		if reservation.State == types.ReservationStatePending && reservation.ExpiresAt == nil && reservation.ExpiresHeight == 0 {
			invariantErr = types.ErrCapacityInvariant.Wrapf("pending reservation %s has no expiry", reservation.ReservationId)
			return true
		}
		if reservation.ActivatedAt != nil && reservation.ActivatedAt.Before(reservation.CreatedAt) || reservation.ConsumedAt != nil && reservation.ConsumedAt.Before(reservation.CreatedAt) || reservation.ReleasedAt != nil && reservation.ReleasedAt.Before(reservation.CreatedAt) || reservation.ExpiredAt != nil && reservation.ExpiredAt.Before(reservation.CreatedAt) || reservation.QuarantinedAt != nil && reservation.QuarantinedAt.Before(reservation.CreatedAt) || reservation.SlashedAt != nil && reservation.SlashedAt.Before(reservation.CreatedAt) {
			invariantErr = types.ErrCapacityInvariant.Wrapf("reservation %s lifecycle time precedes creation", reservation.ReservationId)
			return true
		}
		if reservation.State == types.ReservationStatePending || reservation.State == types.ReservationStateActive || reservation.State == types.ReservationStateConsumed || reservation.State == types.ReservationStateQuarantined || reservation.State == types.ReservationStateDisputed {
			key := inventoryIdentity(reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
			total, err := addCapacityChecked(reservationsByInventory[key], reservation.Capacity)
			if err != nil {
				invariantErr = err
				return true
			}
			reservationsByInventory[key] = total
		}
		if reservation.State == types.ReservationStateActive || reservation.State == types.ReservationStateConsumed || reservation.State == types.ReservationStateQuarantined || reservation.State == types.ReservationStateDisputed {
			if reservation.ConsumerType == "" || reservation.ConsumerId == "" {
				invariantErr = types.ErrCapacityInvariant.Wrapf("reservation %s has no active consumer lineage", reservation.ReservationId)
				return true
			}
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
		inventories[key] = struct{}{}
		sum, err := addCapacityChecked(inventory.Available, reservationsByInventory[key])
		if err != nil || !capacityEqual(sum, inventory.Total) {
			invariantErr = types.ErrCapacityInvariant.Wrapf("inventory %s total does not equal available plus reservations", inventory.InventoryId)
			return true
		}
		return false
	})
	if invariantErr != nil {
		return invariantErr
	}
	providers := make(map[string]struct{}, len(inventories))
	for key := range inventories {
		providers[strings.SplitN(key, "\x00", 2)[0]] = struct{}{}
	}
	providerNames := make([]string, 0, len(providers))
	for provider := range providers {
		providerNames = append(providerNames, provider)
	}
	sort.Strings(providerNames)
	for _, provider := range providerNames {
		if _, err := k.reservationsByProvider(ctx, provider); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(reservationsByInventory))
	for key := range reservationsByInventory {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		capacity := reservationsByInventory[key]
		if _, found := inventories[key]; !found && !capacityIsZero(capacity) {
			return types.ErrCapacityInvariant.Wrapf("nonterminal reservation references missing inventory %q", key)
		}
	}
	return invariantErr
}

func (k Keeper) reservedCapacityForInventory(ctx sdk.Context, provider string, class types.ResourceClass, inventoryID string) (types.ResourceCapacity, error) {
	total := types.ResourceCapacity{}
	reservations, err := k.reservationsByProvider(ctx, provider)
	if err != nil {
		return types.ResourceCapacity{}, err
	}
	for _, reservation := range reservations {
		if reservation.InventoryId != inventoryID || reservation.ResourceClass != class {
			continue
		}
		if reservation.State != types.ReservationStatePending && reservation.State != types.ReservationStateActive && reservation.State != types.ReservationStateConsumed && reservation.State != types.ReservationStateQuarantined && reservation.State != types.ReservationStateDisputed {
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
	reservations, err := k.reservationsByProvider(ctx, inventory.ProviderAddress)
	if err != nil {
		return err
	}
	for _, reservation := range reservations {
		if reservation.InventoryId != inventory.InventoryId || reservation.ResourceClass != inventory.ResourceClass {
			continue
		}
		switch reservation.State {
		case types.ReservationStatePending:
			if _, err := k.transitionReservationCached(ctx, reservation.ReservationId, types.ReservationStateExpired, reason, true); err != nil {
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

func (k Keeper) setReservationIndexes(ctx sdk.Context, reservation types.Reservation) error {
	if err := k.validateReservationIndexOwnership(ctx, reservation); err != nil {
		return err
	}
	store := ctx.KVStore(k.skey)
	if reservation.ProviderAddress != "" {
		store.Set(types.ReservationProviderKey(reservation.ProviderAddress, reservation.ReservationId), []byte{1})
	}
	indexes := []struct {
		prefix []byte
		id     string
	}{
		{types.ReservationOrderKeyPrefix, reservation.MarketOrderId}, {types.ReservationBidKeyPrefix, reservation.MarketBidId},
		{types.ReservationLeaseKeyPrefix, reservation.MarketLeaseId}, {types.ReservationJobKeyPrefix, reservation.HpcJobId},
	}
	for _, index := range indexes {
		if index.id != "" {
			key := types.ReservationLineageKey(index.prefix, index.id)
			if existing := store.Get(key); existing != nil && string(existing) != reservation.ReservationId {
				return types.ErrLineageConflict.Wrapf("lineage %q already belongs to reservation %s", index.id, string(existing))
			}
			store.Set(key, []byte(reservation.ReservationId))
		}
	}
	if reservation.ConsumerType != "" && reservation.ConsumerId != "" {
		key := types.ReservationConsumerKey(reservation.ConsumerType, reservation.ConsumerId)
		if existing := store.Get(key); existing != nil && string(existing) != reservation.ReservationId {
			return types.ErrLineageConflict.Wrapf("consumer already belongs to reservation %s", string(existing))
		}
		store.Set(key, []byte(reservation.ReservationId))
	}
	if reservation.State == types.ReservationStatePending {
		if reservation.ExpiresAt != nil {
			store.Set(types.ReservationExpiryTimeKey(unixToUint64(reservation.ExpiresAt.Unix()), reservation.ReservationId), []byte{1})
		}
		if reservation.ExpiresHeight > 0 {
			store.Set(types.ReservationExpiryHeightKey(uint64(reservation.ExpiresHeight), reservation.ReservationId), []byte{1})
		}
	}
	return nil
}

func (k Keeper) validateReservationIndexOwnership(ctx sdk.Context, reservation types.Reservation) error {
	store := ctx.KVStore(k.skey)
	indexes := []struct {
		prefix []byte
		id     string
	}{{types.ReservationOrderKeyPrefix, reservation.MarketOrderId}, {types.ReservationBidKeyPrefix, reservation.MarketBidId}, {types.ReservationLeaseKeyPrefix, reservation.MarketLeaseId}, {types.ReservationJobKeyPrefix, reservation.HpcJobId}}
	for _, index := range indexes {
		if index.id == "" {
			continue
		}
		if existing := store.Get(types.ReservationLineageKey(index.prefix, index.id)); existing != nil && string(existing) != reservation.ReservationId {
			return types.ErrLineageConflict.Wrapf("lineage %q already belongs to reservation %s", index.id, string(existing))
		}
	}
	if reservation.ConsumerType != "" && reservation.ConsumerId != "" {
		if existing := store.Get(types.ReservationConsumerKey(reservation.ConsumerType, reservation.ConsumerId)); existing != nil && string(existing) != reservation.ReservationId {
			return types.ErrLineageConflict.Wrapf("consumer already belongs to reservation %s", string(existing))
		}
	}
	return nil
}

// RebuildReservationIndexes deterministically restores derived genesis/migration indexes.
func (k Keeper) RebuildReservationIndexes(ctx sdk.Context, reservation types.Reservation) error {
	if err := k.setReservationIndexes(ctx, reservation); err != nil {
		return err
	}
	if reservation.IdempotencyKey == "" || len(reservation.PayloadHash) == 0 {
		return nil
	}
	replayBytes, err := json.Marshal(reservationReplay{ReservationID: reservation.ReservationId, PayloadHash: reservation.PayloadHash})
	if err != nil {
		return err
	}
	key := types.ReservationIdempotencyKey(reservation.IdempotencyKey)
	store := ctx.KVStore(k.skey)
	if existing := store.Get(key); existing != nil && !bytes.Equal(existing, replayBytes) {
		return types.ErrReservationConflict.Wrapf("idempotency key for reservation %s is already owned", reservation.ReservationId)
	}
	store.Set(key, replayBytes)
	return nil
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
	if reservation.InventoryId == "" || capacityEqual(reservation.Capacity, types.ResourceCapacity{}) {
		if reservation.LegacySource != "" {
			return nil
		}
		return types.ErrCapacityInvariant.Wrap("reservation has no inventory capacity to restore")
	}
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

func (k Keeper) recordReservationSlashing(ctx sdk.Context, reservation types.Reservation, reason string) error {
	entry := types.SlashingEvent{ProviderAddress: reservation.ProviderAddress, AllocationId: reservation.ReservationId, Reason: reason, Penalty: k.GetParams(ctx).SlashingPenalty, CreatedAt: ctx.BlockTime()}
	bz, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	sequence := k.nextSequence(ctx, types.SequenceKey(types.SlashingEventKeyPrefix, reservation.ReservationId))
	ctx.KVStore(k.skey).Set(types.SlashingEventKey(reservation.ReservationId, sequence), bz)
	return nil
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
		return types.ErrEligibilityUnavailable
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
	if request.ExpiresHeight < 0 {
		return types.ErrInvalidRequest.Wrap("expires_height cannot be negative")
	}
	if request.ExpiresAt != nil && request.ExpiresAt.Unix() <= 0 {
		return types.ErrInvalidRequest.Wrap("expires_at must be a positive Unix time")
	}
	identifiers := []string{request.IdempotencyKey, request.RequestId, request.RequesterAddress, request.ProviderAddress, request.InventoryId, request.ConsumerType, request.ConsumerId, request.MarketOrderId, request.MarketBidId, request.MarketLeaseId, request.HpcJobId, request.EscrowId, request.CollateralId, request.LegacySource, request.LegacyReference}
	for _, value := range identifiers {
		if len(value) > maxReservationIdentifierLength || strings.ContainsRune(value, '\x00') {
			return types.ErrInvalidRequest.Wrap("reservation identifier is too long or contains NUL")
		}
	}
	return validateCapacity(request.Capacity, false)
}

func validateReservationReason(reason string) error {
	if len(reason) > maxReservationReasonLength || strings.ContainsRune(reason, '\x00') {
		return types.ErrInvalidRequest.Wrap("reservation reason is too long or contains NUL")
	}
	return nil
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

func reservationLinkEqual(reservation types.Reservation, link types.ReservationLink) bool {
	return reservation.ConsumerType == link.ConsumerType &&
		reservation.ConsumerId == link.ConsumerId &&
		reservation.MarketOrderId == link.MarketOrderId &&
		reservation.MarketBidId == link.MarketBidId &&
		reservation.MarketLeaseId == link.MarketLeaseId &&
		reservation.HpcJobId == link.HpcJobId &&
		reservation.EscrowId == link.EscrowId &&
		reservation.CollateralId == link.CollateralId
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
	if capacity.Gpus > 0 && capacity.GpuType == "" {
		return types.ErrInvalidRequest.Wrap("gpu_type required when GPU capacity is nonzero")
	}
	if capacity.Gpus == 0 && capacity.GpuType != "" {
		return types.ErrInvalidRequest.Wrap("gpu_type requires nonzero GPU capacity")
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

func capacityIsZero(capacity types.ResourceCapacity) bool {
	return capacity.CpuCores == 0 && capacity.MemoryGb == 0 && capacity.StorageGb == 0 && capacity.NetworkMbps == 0 && capacity.Gpus == 0
}

func inventoryIdentity(provider string, class types.ResourceClass, inventoryID string) string {
	return fmt.Sprintf("%s\x00%d\x00%s", provider, class, inventoryID)
}
