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
	ErrInvalidAddress  = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidCallback = sdkerrors.Register(ModuleName, 2, "invalid callback")
)
