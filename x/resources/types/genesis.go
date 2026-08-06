package types

import (
	"fmt"
	"math"
	"sort"
)

// GenesisState defines the module genesis state.
type GenesisState struct {
	Params                      Params
	Inventories                 []ResourceInventory
	Allocations                 []ResourceAllocation
	Reservations                []Reservation
	CanonicalReservationsActive bool
	ReservationEvents           []ReservationEvent
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                      DefaultParams(),
		Inventories:                 []ResourceInventory{},
		Allocations:                 []ResourceAllocation{},
		Reservations:                []Reservation{},
		CanonicalReservationsActive: true,
		ReservationEvents:           []ReservationEvent{},
	}
}

// Validate validates genesis state.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}
	inventories := make(map[string]ResourceInventory, len(gs.Inventories))
	for _, inventory := range gs.Inventories {
		if inventory.InventoryId == "" || inventory.ProviderAddress == "" || inventory.ResourceClass == ResourceClassUnspecified {
			return fmt.Errorf("inventory identity is incomplete")
		}
		key := fmt.Sprintf("%s\x00%d\x00%s", inventory.ProviderAddress, inventory.ResourceClass, inventory.InventoryId)
		if _, exists := inventories[key]; exists {
			return fmt.Errorf("duplicate inventory %s", inventory.InventoryId)
		}
		if err := validateGenesisCapacity(inventory.Total, true); err != nil {
			return fmt.Errorf("inventory %s total: %w", inventory.InventoryId, err)
		}
		if err := validateGenesisCapacity(inventory.Available, true); err != nil || !genesisCapacitySatisfies(inventory.Total, inventory.Available) {
			return fmt.Errorf("inventory %s available capacity is invalid", inventory.InventoryId)
		}
		inventories[key] = inventory
	}
	seen := make(map[string]struct{}, len(gs.Reservations))
	consumers := make(map[string]string, len(gs.Reservations))
	lineages := make(map[string]string, len(gs.Reservations)*4)
	reserved := make(map[string]ResourceCapacity, len(gs.Inventories))
	for _, reservation := range gs.Reservations {
		if reservation.ReservationId == "" {
			return fmt.Errorf("reservation ID cannot be empty")
		}
		if _, exists := seen[reservation.ReservationId]; exists {
			return fmt.Errorf("duplicate reservation ID %s", reservation.ReservationId)
		}
		if reservation.State < ReservationStatePending || reservation.State > ReservationStateSlashed {
			return fmt.Errorf("reservation %s has unspecified state", reservation.ReservationId)
		}
		allowZero := reservation.LegacySource != "" && (reservation.State == ReservationStateQuarantined || IsTerminalReservationState(reservation.State))
		if err := validateGenesisCapacity(reservation.Capacity, allowZero); err != nil {
			return fmt.Errorf("reservation %s capacity: %w", reservation.ReservationId, err)
		}
		if err := validateGenesisCapacity(reservation.Consumed, true); err != nil || !genesisCapacitySatisfies(reservation.Capacity, reservation.Consumed) {
			return fmt.Errorf("reservation %s consumed capacity is invalid", reservation.ReservationId)
		}
		if reservation.State == ReservationStatePending && reservation.ExpiresAt == nil && reservation.ExpiresHeight == 0 {
			return fmt.Errorf("pending reservation %s has no expiry", reservation.ReservationId)
		}
		if reservation.ExpiresHeight < 0 || reservation.CreatedHeight < 0 || reservation.ActivatedHeight < 0 || reservation.ConsumedHeight < 0 || reservation.ReleasedHeight < 0 || reservation.QuarantinedHeight < 0 || reservation.SlashedHeight < 0 {
			return fmt.Errorf("reservation %s has a negative lifecycle height", reservation.ReservationId)
		}
		if reservation.ActivatedAt != nil && reservation.ActivatedAt.Before(reservation.CreatedAt) || reservation.ConsumedAt != nil && reservation.ConsumedAt.Before(reservation.CreatedAt) || reservation.ReleasedAt != nil && reservation.ReleasedAt.Before(reservation.CreatedAt) || reservation.ExpiredAt != nil && reservation.ExpiredAt.Before(reservation.CreatedAt) || reservation.QuarantinedAt != nil && reservation.QuarantinedAt.Before(reservation.CreatedAt) || reservation.SlashedAt != nil && reservation.SlashedAt.Before(reservation.CreatedAt) {
			return fmt.Errorf("reservation %s lifecycle time precedes creation", reservation.ReservationId)
		}
		legacyQuarantine := reservation.State == ReservationStateQuarantined && reservation.LegacySource != "" && reservation.InventoryId == ""
		if !legacyQuarantine && !IsTerminalReservationState(reservation.State) {
			key := fmt.Sprintf("%s\x00%d\x00%s", reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
			if _, exists := inventories[key]; !exists {
				return fmt.Errorf("reservation %s references missing inventory", reservation.ReservationId)
			}
		}
		if reservation.State == ReservationStateActive || reservation.State == ReservationStateConsumed || reservation.State == ReservationStateQuarantined {
			if reservation.ConsumerType == "" || reservation.ConsumerId == "" {
				return fmt.Errorf("reservation %s has no executable consumer lineage", reservation.ReservationId)
			}
			key := reservation.ConsumerType + "\x00" + reservation.ConsumerId
			if owner, exists := consumers[key]; exists && owner != reservation.ReservationId {
				return fmt.Errorf("consumer %s has reservations %s and %s", key, owner, reservation.ReservationId)
			}
			consumers[key] = reservation.ReservationId
		}
		for _, lineage := range [][2]string{{"order", reservation.MarketOrderId}, {"bid", reservation.MarketBidId}, {"lease", reservation.MarketLeaseId}, {"job", reservation.HpcJobId}} {
			kind, id := lineage[0], lineage[1]
			if id == "" {
				continue
			}
			key := kind + "\x00" + id
			if owner, exists := lineages[key]; exists && owner != reservation.ReservationId {
				return fmt.Errorf("lineage %s has reservations %s and %s", key, owner, reservation.ReservationId)
			}
			lineages[key] = reservation.ReservationId
		}
		if reservation.State == ReservationStatePending || reservation.State == ReservationStateActive || reservation.State == ReservationStateConsumed || reservation.State == ReservationStateQuarantined {
			key := fmt.Sprintf("%s\x00%d\x00%s", reservation.ProviderAddress, reservation.ResourceClass, reservation.InventoryId)
			var err error
			reserved[key], err = addGenesisCapacity(reserved[key], reservation.Capacity)
			if err != nil {
				return fmt.Errorf("reservation %s: %w", reservation.ReservationId, err)
			}
		}
		seen[reservation.ReservationId] = struct{}{}
	}
	keys := make([]string, 0, len(inventories))
	for key := range inventories {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		inventory := inventories[key]
		total, err := addGenesisCapacity(inventory.Available, reserved[key])
		if err != nil || !genesisCapacityEqual(total, inventory.Total) {
			return fmt.Errorf("inventory %s does not conserve capacity", inventory.InventoryId)
		}
	}
	reservedKeys := make([]string, 0, len(reserved))
	for key := range reserved {
		reservedKeys = append(reservedKeys, key)
	}
	sort.Strings(reservedKeys)
	for _, key := range reservedKeys {
		if _, found := inventories[key]; !found && !genesisCapacityZero(reserved[key]) {
			return fmt.Errorf("nonterminal reservation references missing inventory %q", key)
		}
	}
	events := make(map[string]struct{}, len(gs.ReservationEvents))
	for _, event := range gs.ReservationEvents {
		if event.ReservationId == "" || event.Sequence == 0 {
			return fmt.Errorf("reservation event identity is incomplete")
		}
		if _, found := seen[event.ReservationId]; !found {
			return fmt.Errorf("reservation event references missing reservation %s", event.ReservationId)
		}
		if event.ToState < ReservationStatePending || event.ToState > ReservationStateSlashed || event.FromState < ReservationStateUnspecified || event.FromState > ReservationStateSlashed {
			return fmt.Errorf("reservation event %s/%d has invalid state", event.ReservationId, event.Sequence)
		}
		key := fmt.Sprintf("%s\x00%020d", event.ReservationId, event.Sequence)
		if _, duplicate := events[key]; duplicate {
			return fmt.Errorf("duplicate reservation event %s/%d", event.ReservationId, event.Sequence)
		}
		events[key] = struct{}{}
	}
	return nil
}

func validateGenesisCapacity(capacity ResourceCapacity, allowZero bool) error {
	values := []int64{capacity.CpuCores, capacity.MemoryGb, capacity.StorageGb, capacity.NetworkMbps, capacity.Gpus}
	allZero := true
	for _, value := range values {
		if value < 0 {
			return fmt.Errorf("negative capacity")
		}
		allZero = allZero && value == 0
	}
	if !allowZero && allZero {
		return fmt.Errorf("empty capacity")
	}
	if capacity.Gpus > 0 && capacity.GpuType == "" || capacity.Gpus == 0 && capacity.GpuType != "" {
		return fmt.Errorf("GPU count/type mismatch")
	}
	return nil
}

func genesisCapacitySatisfies(available, required ResourceCapacity) bool {
	return available.CpuCores >= required.CpuCores && available.MemoryGb >= required.MemoryGb && available.StorageGb >= required.StorageGb && available.NetworkMbps >= required.NetworkMbps && available.Gpus >= required.Gpus && (required.GpuType == "" || available.GpuType == required.GpuType)
}

func addGenesisCapacity(a, b ResourceCapacity) (ResourceCapacity, error) {
	left := []*int64{&a.CpuCores, &a.MemoryGb, &a.StorageGb, &a.NetworkMbps, &a.Gpus}
	right := []int64{b.CpuCores, b.MemoryGb, b.StorageGb, b.NetworkMbps, b.Gpus}
	for i, value := range right {
		if value < 0 || *left[i] > math.MaxInt64-value {
			return ResourceCapacity{}, fmt.Errorf("capacity overflow")
		}
		*left[i] += value
	}
	if a.GpuType == "" {
		a.GpuType = b.GpuType
	}
	if a.GpuType != "" && b.GpuType != "" && a.GpuType != b.GpuType {
		return ResourceCapacity{}, fmt.Errorf("GPU type mismatch")
	}
	return a, nil
}

func genesisCapacityEqual(a, b ResourceCapacity) bool {
	return a.CpuCores == b.CpuCores && a.MemoryGb == b.MemoryGb && a.StorageGb == b.StorageGb && a.NetworkMbps == b.NetworkMbps && a.Gpus == b.Gpus && a.GpuType == b.GpuType
}

func genesisCapacityZero(capacity ResourceCapacity) bool {
	return capacity.CpuCores == 0 && capacity.MemoryGb == 0 && capacity.StorageGb == 0 && capacity.NetworkMbps == 0 && capacity.Gpus == 0
}
