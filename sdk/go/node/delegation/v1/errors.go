// Package v1 provides errors and constants for delegation types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "delegation"
)

// Module error codes
var (
	ErrInvalidAddress    = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidAmount     = sdkerrors.Register(ModuleName, 2, "invalid amount")
	ErrSelfRedelegation  = sdkerrors.Register(ModuleName, 3, "cannot redelegate to same validator")
)
