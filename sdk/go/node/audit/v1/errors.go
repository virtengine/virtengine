package v1

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errProviderNotFound uint32 = iota + 1
	errInvalidAddress
	errAttributeNotFound
)

var (
	// ErrProviderNotFound provider not found
	ErrProviderNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errProviderNotFound, codes.NotFound, "invalid provider: address not found")

	// ErrInvalidAddress invalid address
	ErrInvalidAddress = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAddress, codes.InvalidArgument, "invalid address")

	// ErrAttributeNotFound attribute not found
	ErrAttributeNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errAttributeNotFound, codes.NotFound, "attribute not found")
)
