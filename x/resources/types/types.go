package types

import (
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

// Type aliases to generated protobuf types.
type (
	ResourceInventory  = resourcesv1.ResourceInventory
	ResourceRequest    = resourcesv1.ResourceRequest
	ResourceAllocation = resourcesv1.ResourceAllocation
	ResourceCapacity   = resourcesv1.ResourceCapacity
	Locality           = resourcesv1.Locality
	AllocationEvent    = resourcesv1.AllocationEvent
	SlashingEvent      = resourcesv1.SlashingEvent
	Reservation        = resourcesv1.Reservation
	ReservationRequest = resourcesv1.ReservationRequest
	ReservationLink    = resourcesv1.ReservationLink
	ReservationEvent   = resourcesv1.ReservationEvent
	ReservationState   = resourcesv1.ReservationState
	ResourceClass      = resourcesv1.ResourceClass
	AllocationState    = resourcesv1.AllocationState
)

// ResourceClass enum constants.
const (
	ResourceClassUnspecified = resourcesv1.ResourceClass_RESOURCE_CLASS_UNSPECIFIED
	ResourceClassCompute     = resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE
	ResourceClassStorage     = resourcesv1.ResourceClass_RESOURCE_CLASS_STORAGE
	ResourceClassNetwork     = resourcesv1.ResourceClass_RESOURCE_CLASS_NETWORK
)

// AllocationState enum constants.
const (
	AllocationStateUnspecified = resourcesv1.AllocationState_ALLOCATION_STATE_UNSPECIFIED
	AllocationStatePending     = resourcesv1.AllocationState_ALLOCATION_STATE_PENDING
	AllocationStateActive      = resourcesv1.AllocationState_ALLOCATION_STATE_ACTIVE
	AllocationStateExpired     = resourcesv1.AllocationState_ALLOCATION_STATE_EXPIRED
	AllocationStateReleased    = resourcesv1.AllocationState_ALLOCATION_STATE_RELEASED
)

const (
	ReservationStateUnspecified = resourcesv1.ReservationState_RESERVATION_STATE_UNSPECIFIED
	ReservationStatePending     = resourcesv1.ReservationState_RESERVATION_STATE_PENDING
	ReservationStateActive      = resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE
	ReservationStateConsumed    = resourcesv1.ReservationState_RESERVATION_STATE_CONSUMED
	ReservationStateReleased    = resourcesv1.ReservationState_RESERVATION_STATE_RELEASED
	ReservationStateExpired     = resourcesv1.ReservationState_RESERVATION_STATE_EXPIRED
	ReservationStateQuarantined = resourcesv1.ReservationState_RESERVATION_STATE_QUARANTINED
	ReservationStateSlashed     = resourcesv1.ReservationState_RESERVATION_STATE_SLASHED
)

// CanTransitionReservation is the single explicit reservation transition table.
func CanTransitionReservation(from, to ReservationState) bool {
	switch from {
	case ReservationStatePending:
		return to == ReservationStateActive || to == ReservationStateReleased || to == ReservationStateExpired || to == ReservationStateQuarantined || to == ReservationStateSlashed
	case ReservationStateActive:
		return to == ReservationStateConsumed || to == ReservationStateReleased || to == ReservationStateQuarantined || to == ReservationStateSlashed
	case ReservationStateConsumed:
		return to == ReservationStateReleased || to == ReservationStateQuarantined || to == ReservationStateSlashed
	case ReservationStateQuarantined:
		return to == ReservationStateSlashed
	default:
		return false
	}
}

func IsTerminalReservationState(state ReservationState) bool {
	return state == ReservationStateReleased || state == ReservationStateExpired || state == ReservationStateSlashed
}
