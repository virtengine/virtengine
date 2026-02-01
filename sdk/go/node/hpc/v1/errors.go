// Package v1 provides errors and constants for HPC types.
package v1

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module constants
const (
	ModuleName = "hpc"
)

// Module error codes for SDK-level validation
var (
	ErrInvalidAddress  = sdkerrors.Register(ModuleName, 1, "invalid address")
	ErrInvalidCluster  = sdkerrors.Register(ModuleName, 2, "invalid cluster")
	ErrInvalidOffering = sdkerrors.Register(ModuleName, 3, "invalid offering")
	ErrInvalidJob      = sdkerrors.Register(ModuleName, 4, "invalid job")
	ErrInvalidNode     = sdkerrors.Register(ModuleName, 5, "invalid node")
	ErrInvalidDispute  = sdkerrors.Register(ModuleName, 6, "invalid dispute")
	ErrInvalidStatus   = sdkerrors.Register(ModuleName, 7, "invalid status")
)

