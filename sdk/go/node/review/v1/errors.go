// Package v1 provides errors and constants for review types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "review"

	// Rating constraints
	MinRating = 1
	MaxRating = 5

	// Review text constraints
	MinReviewTextLength = 10
	MaxReviewTextLength = 2000
)

// Module error codes
var (
	ErrInvalidAddress     = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidOrderID     = sdkerrors.Register(ModuleName, 2, "invalid order ID")
	ErrInvalidRating      = sdkerrors.Register(ModuleName, 3, "invalid rating")
	ErrReviewTextTooShort = sdkerrors.Register(ModuleName, 4, "review text too short")
	ErrReviewTextTooLong  = sdkerrors.Register(ModuleName, 5, "review text too long")
	ErrInvalidReviewID    = sdkerrors.Register(ModuleName, 6, "invalid review ID")
	ErrInvalidReason      = sdkerrors.Register(ModuleName, 7, "invalid reason")
)
