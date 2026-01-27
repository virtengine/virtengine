package v1beta4

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errInvalidProviderURI uint32 = iota + 1
	errNotAbsProviderURI
	errProviderNotFound
	errProviderExists
	errInvalidAddress
	errAttributes
	errIncompatibleAttributes
	errInvalidInfoWebsite
	errInternal
)

var (
	// ErrInvalidProviderURI register error code for invalid provider uri
	ErrInvalidProviderURI = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidProviderURI, codes.InvalidArgument, "invalid provider: invalid host uri")

	// ErrNotAbsProviderURI register error code for not absolute provider uri
	ErrNotAbsProviderURI = sdkerrors.RegisterWithGRPCCode(ModuleName, errNotAbsProviderURI, codes.InvalidArgument, "invalid provider: not absolute host uri")

	// ErrProviderNotFound provider not found
	ErrProviderNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errProviderNotFound, codes.NotFound, "invalid provider: address not found")

	// ErrProviderExists provider already exists
	ErrProviderExists = sdkerrors.RegisterWithGRPCCode(ModuleName, errProviderExists, codes.AlreadyExists, "invalid provider: already exists")

	// ErrInvalidAddress invalid provider address
	ErrInvalidAddress = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAddress, codes.InvalidArgument, "invalid address")

	// ErrAttributes error code for provider attribute problems
	ErrAttributes = sdkerrors.RegisterWithGRPCCode(ModuleName, errAttributes, codes.InvalidArgument, "attribute specification error")

	// ErrIncompatibleAttributes error code for attributes update
	ErrIncompatibleAttributes = sdkerrors.RegisterWithGRPCCode(ModuleName, errIncompatibleAttributes, codes.FailedPrecondition, "attributes cannot be changed")

	// ErrInvalidInfoWebsite register error code for invalid info website
	ErrInvalidInfoWebsite = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidInfoWebsite, codes.InvalidArgument, "invalid provider: invalid info website")

	// ErrInternal internal error
	ErrInternal = sdkerrors.RegisterWithGRPCCode(ModuleName, errInternal, codes.Internal, "internal error")
)
