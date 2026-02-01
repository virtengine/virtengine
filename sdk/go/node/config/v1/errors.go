// Package v1 provides errors and constants for config types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "config"
)

// Module error codes
var (
	ErrInvalidAddress           = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidClientID          = sdkerrors.Register(ModuleName, 2, "invalid client ID")
	ErrInvalidPublicKey         = sdkerrors.Register(ModuleName, 3, "invalid public key")
	ErrInvalidVersionConstraint = sdkerrors.Register(ModuleName, 4, "invalid version constraint")
	ErrInvalidReason            = sdkerrors.Register(ModuleName, 5, "invalid reason")
)

