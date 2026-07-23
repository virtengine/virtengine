package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidRequest               = errorsmod.Register(ModuleName, 1, "invalid resource request")
	ErrInventoryNotFound            = errorsmod.Register(ModuleName, 2, "inventory not found")
	ErrNoEligibleInventory          = errorsmod.Register(ModuleName, 3, "no eligible inventory")
	ErrAllocationNotFound           = errorsmod.Register(ModuleName, 4, "allocation not found")
	ErrInvalidState                 = errorsmod.Register(ModuleName, 5, "invalid allocation state")
	ErrUnauthorized                 = errorsmod.Register(ModuleName, 6, "unauthorized")
	ErrStaleHeartbeat               = errorsmod.Register(ModuleName, 7, "stale heartbeat")
	ErrInvalidParams                = errorsmod.Register(ModuleName, 8, "invalid params")
	ErrReservationNotFound          = errorsmod.Register(ModuleName, 9, "reservation not found")
	ErrReservationConflict          = errorsmod.Register(ModuleName, 10, "reservation idempotency conflict")
	ErrInvalidReservationTransition = errorsmod.Register(ModuleName, 11, "invalid reservation state transition")
	ErrCapacityInvariant            = errorsmod.Register(ModuleName, 12, "capacity conservation invariant violated")
	ErrEligibilityUnavailable       = errorsmod.Register(ModuleName, 13, "mandatory reservation eligibility dependency unavailable")
	ErrProviderIneligible           = errorsmod.Register(ModuleName, 14, "provider is not eligible for reservation")
	ErrLineageConflict              = errorsmod.Register(ModuleName, 15, "reservation lineage conflict")
	ErrCapacityOverflow             = errorsmod.Register(ModuleName, 16, "resource capacity overflow")
	ErrLegacyAllocationDeprecated   = errorsmod.Register(ModuleName, 17, "legacy resource allocation writes deprecated; use canonical reservations")
)
