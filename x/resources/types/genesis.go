package types

import "fmt"

// GenesisState defines the module genesis state.
type GenesisState struct {
	Params       Params
	Inventories  []ResourceInventory
	Allocations  []ResourceAllocation
	Reservations []Reservation
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:       DefaultParams(),
		Inventories:  []ResourceInventory{},
		Allocations:  []ResourceAllocation{},
		Reservations: []Reservation{},
	}
}

// Validate validates genesis state.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}
	seen := make(map[string]struct{}, len(gs.Reservations))
	consumers := make(map[string]string, len(gs.Reservations))
	for _, reservation := range gs.Reservations {
		if reservation.ReservationId == "" {
			return fmt.Errorf("reservation ID cannot be empty")
		}
		if _, exists := seen[reservation.ReservationId]; exists {
			return fmt.Errorf("duplicate reservation ID %s", reservation.ReservationId)
		}
		if reservation.State == ReservationStateUnspecified {
			return fmt.Errorf("reservation %s has unspecified state", reservation.ReservationId)
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
		seen[reservation.ReservationId] = struct{}{}
	}
	return nil
}
