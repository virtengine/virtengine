package v1

import (
	"errors"

	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errCertificateNotFound uint32 = iota + 1
	errInvalidAddress
	errCertificateExists
	errCertificateAlreadyRevoked
	errInvalidSerialNumber
	errInvalidCertificateValue
	errInvalidPubkeyValue
	errInvalidState
	errInvalidKeySize
)

var (
	ErrCertificate = errors.New("certificate error")
)

var (
	// ErrCertificateNotFound certificate not found
	ErrCertificateNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errCertificateNotFound, codes.NotFound, "certificate not found")

	// ErrInvalidAddress invalid trusted auditor address
	ErrInvalidAddress = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAddress, codes.InvalidArgument, "invalid address")

	// ErrCertificateExists certificate already exists
	ErrCertificateExists = sdkerrors.RegisterWithGRPCCode(ModuleName, errCertificateExists, codes.AlreadyExists, "certificate exists")

	// ErrCertificateAlreadyRevoked certificate already revoked
	ErrCertificateAlreadyRevoked = sdkerrors.RegisterWithGRPCCode(ModuleName, errCertificateAlreadyRevoked, codes.FailedPrecondition, "certificate already revoked")

	// ErrInvalidSerialNumber invalid serial number
	ErrInvalidSerialNumber = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidSerialNumber, codes.InvalidArgument, "invalid serial number")

	// ErrInvalidCertificateValue certificate content is not valid
	ErrInvalidCertificateValue = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidCertificateValue, codes.InvalidArgument, "invalid certificate value")

	// ErrInvalidPubkeyValue public key is not valid
	ErrInvalidPubkeyValue = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidPubkeyValue, codes.InvalidArgument, "invalid pubkey value")

	// ErrInvalidState invalid certificate state
	ErrInvalidState = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidState, codes.InvalidArgument, "invalid state")

	// ErrInvalidKeySize invalid key size
	ErrInvalidKeySize = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidKeySize, codes.InvalidArgument, "invalid key size")
)
