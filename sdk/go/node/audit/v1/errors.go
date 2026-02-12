package v1

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errProviderNotFound uint32 = iota + 1
	errInvalidAddress
	errAttributeNotFound
	// Audit log error codes
	errAuditLogDisabled
	errInvalidExportFormat
	errExportJobNotFound
	errInvalidRequest
	errNotFound
	errUnauthorized
)

var (
	// ErrProviderNotFound provider not found
	ErrProviderNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errProviderNotFound, codes.NotFound, "invalid provider: address not found")

	// ErrInvalidAddress invalid address
	ErrInvalidAddress = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAddress, codes.InvalidArgument, "invalid address")

	// ErrAttributeNotFound attribute not found
	ErrAttributeNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errAttributeNotFound, codes.NotFound, "attribute not found")

	// Audit log errors
	ErrAuditLogDisabled    = sdkerrors.RegisterWithGRPCCode(ModuleName, errAuditLogDisabled, codes.FailedPrecondition, "audit logging is disabled")
	ErrInvalidExportFormat = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidExportFormat, codes.InvalidArgument, "invalid export format")
	ErrExportJobNotFound   = sdkerrors.RegisterWithGRPCCode(ModuleName, errExportJobNotFound, codes.NotFound, "export job not found")
	ErrInvalidRequest      = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidRequest, codes.InvalidArgument, "invalid request")
	ErrNotFound            = sdkerrors.RegisterWithGRPCCode(ModuleName, errNotFound, codes.NotFound, "not found")
	ErrUnauthorized        = sdkerrors.RegisterWithGRPCCode(ModuleName, errUnauthorized, codes.PermissionDenied, "unauthorized")
)
