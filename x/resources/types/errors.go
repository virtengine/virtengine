package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidRequest      = errorsmod.Register(ModuleName, 1, "invalid resource request")
	ErrInventoryNotFound   = errorsmod.Register(ModuleName, 2, "inventory not found")
	ErrNoEligibleInventory = errorsmod.Register(ModuleName, 3, "no eligible inventory")
	ErrAllocationNotFound  = errorsmod.Register(ModuleName, 4, "allocation not found")
	ErrInvalidState        = errorsmod.Register(ModuleName, 5, "invalid allocation state")
	ErrUnauthorized        = errorsmod.Register(ModuleName, 6, "unauthorized")
	ErrStaleHeartbeat      = errorsmod.Register(ModuleName, 7, "stale heartbeat")
	ErrInvalidParams       = errorsmod.Register(ModuleName, 8, "invalid params")
)
