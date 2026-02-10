// Package v1 provides errors and constants for marketplace types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "marketplace"
)

// Module error codes for SDK-level validation
var (
	ErrInvalidAddress    = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidCallback   = sdkerrors.Register(ModuleName, 2, "invalid callback")
	ErrInvalidRequest    = sdkerrors.Register(ModuleName, 3, "invalid request")
	ErrInvalidOffering   = sdkerrors.Register(ModuleName, 4, "invalid offering")
	ErrInvalidOrder      = sdkerrors.Register(ModuleName, 5, "invalid order")
	ErrInvalidBid        = sdkerrors.Register(ModuleName, 6, "invalid bid")
	ErrInvalidAllocation = sdkerrors.Register(ModuleName, 7, "invalid allocation")
)
