// Package v1 provides errors and constants for benchmark types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "benchmark"

	// Validation constraints
	MaxReasonLength     = 500
	MaxResolutionLength = 1000
)

// Module error codes
var (
	ErrInvalidProvider      = sdkerrors.Register(ModuleName, 1, "invalid provider address")
	ErrInvalidRequester     = sdkerrors.Register(ModuleName, 2, "invalid requester address")
	ErrInvalidReporter      = sdkerrors.Register(ModuleName, 3, "invalid reporter address")
	ErrInvalidAuthority     = sdkerrors.Register(ModuleName, 4, "invalid authority address")
	ErrInvalidChallengeID   = sdkerrors.Register(ModuleName, 5, "invalid challenge ID")
	ErrInvalidReason        = sdkerrors.Register(ModuleName, 6, "invalid reason")
	ErrInvalidResolution    = sdkerrors.Register(ModuleName, 7, "invalid resolution")
	ErrInvalidBenchmarkType = sdkerrors.Register(ModuleName, 8, "invalid benchmark type")
	ErrMissingResults       = sdkerrors.Register(ModuleName, 9, "missing benchmark results")
	ErrReasonTooLong        = sdkerrors.Register(ModuleName, 10, "reason too long")
	ErrResolutionTooLong    = sdkerrors.Register(ModuleName, 11, "resolution too long")
)
