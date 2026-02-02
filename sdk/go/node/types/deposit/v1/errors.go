package v1

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"

	attr "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1"
)

const (
	errInvalidDepositor = iota + attr.ErrLast
	errInvalidDepositSource
)

var (
	// ErrInvalidDepositor invalid depositor
	ErrInvalidDepositor = sdkerrors.RegisterWithGRPCCode(attr.ModuleName, errInvalidDepositor, codes.InvalidArgument, "invalid depositor")
	// ErrInvalidDepositSource indicates invalid deposit source for the deployment
	ErrInvalidDepositSource = sdkerrors.RegisterWithGRPCCode(attr.ModuleName, errInvalidDepositSource, codes.InvalidArgument, "invalid deposit source")
)
