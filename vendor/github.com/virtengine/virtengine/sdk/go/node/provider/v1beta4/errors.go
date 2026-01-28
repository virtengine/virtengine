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

// Public key error codes start at 20 to avoid conflicts with existing error registrations
const (
	errInvalidPublicKey         uint32 = 20
	errInvalidPublicKeyType     uint32 = 21
	errPublicKeyNotFound        uint32 = 22
	errInvalidRotationSignature uint32 = 23
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

	// ErrInvalidPublicKey invalid public key format or length
	ErrInvalidPublicKey = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidPublicKey, codes.InvalidArgument, "invalid public key")

	// ErrInvalidPublicKeyType unsupported public key type
	ErrInvalidPublicKeyType = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidPublicKeyType, codes.InvalidArgument, "invalid public key type")

	// ErrPublicKeyNotFound public key not found for provider
	ErrPublicKeyNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errPublicKeyNotFound, codes.NotFound, "public key not found")

	// ErrInvalidRotationSignature invalid signature for key rotation
	ErrInvalidRotationSignature = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidRotationSignature, codes.InvalidArgument, "invalid rotation signature")
)
