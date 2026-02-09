package v1

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errInvalidParam uint32 = iota + 1
)

var (
	// ErrInvalidParam indicates an invalid chain parameter
	ErrInvalidParam = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidParam, codes.InvalidArgument, "parameter invalid")
)
