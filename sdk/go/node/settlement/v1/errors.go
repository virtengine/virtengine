// Package v1 provides errors and constants for settlement types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "settlement"
)

// Module error codes for SDK-level validation
var (
	ErrInvalidAddress     = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidEscrow      = sdkerrors.Register(ModuleName, 2, "invalid escrow")
	ErrInvalidAmount      = sdkerrors.Register(ModuleName, 3, "invalid amount")
	ErrInvalidSettlement  = sdkerrors.Register(ModuleName, 4, "invalid settlement")
	ErrInvalidUsageRecord = sdkerrors.Register(ModuleName, 5, "invalid usage record")
	ErrInvalidSignature   = sdkerrors.Register(ModuleName, 6, "invalid signature")
	ErrInvalidReward      = sdkerrors.Register(ModuleName, 7, "invalid reward")
)

