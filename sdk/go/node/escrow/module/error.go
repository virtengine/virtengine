package module

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errAccountExists uint32 = iota + 1
	errAccountClosed
	errAccountNotFound
	errAccountOverdrawn
	errInvalidDenomination
	errPaymentExists
	errPaymentClosed
	errPaymentNotFound
	errPaymentRateZero
	errInvalidPayment
	errInvalidSettlement
	errInvalidID
	errInvalidAccount
	errInvalidAccountDepositor
	errUnauthorizedDepositScope
	errInvalidDeposit
	errInvalidAuthzScope
)

var (
	ErrAccountExists            = sdkerrors.RegisterWithGRPCCode(ModuleName, errAccountExists, codes.AlreadyExists, "account exists")
	ErrAccountClosed            = sdkerrors.RegisterWithGRPCCode(ModuleName, errAccountClosed, codes.FailedPrecondition, "account closed")
	ErrAccountNotFound          = sdkerrors.RegisterWithGRPCCode(ModuleName, errAccountNotFound, codes.NotFound, "account not found")
	ErrAccountOverdrawn         = sdkerrors.RegisterWithGRPCCode(ModuleName, errAccountOverdrawn, codes.FailedPrecondition, "account overdrawn")
	ErrInvalidDenomination      = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidDenomination, codes.InvalidArgument, "invalid denomination")
	ErrPaymentExists            = sdkerrors.RegisterWithGRPCCode(ModuleName, errPaymentExists, codes.AlreadyExists, "payment exists")
	ErrPaymentClosed            = sdkerrors.RegisterWithGRPCCode(ModuleName, errPaymentClosed, codes.FailedPrecondition, "payment closed")
	ErrPaymentNotFound          = sdkerrors.RegisterWithGRPCCode(ModuleName, errPaymentNotFound, codes.NotFound, "payment not found")
	ErrPaymentRateZero          = sdkerrors.RegisterWithGRPCCode(ModuleName, errPaymentRateZero, codes.InvalidArgument, "payment rate zero")
	ErrInvalidPayment           = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidPayment, codes.InvalidArgument, "invalid payment")
	ErrInvalidSettlement        = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidSettlement, codes.InvalidArgument, "invalid settlement")
	ErrInvalidID                = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidID, codes.InvalidArgument, "invalid ID")
	ErrInvalidAccount           = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAccount, codes.InvalidArgument, "invalid account")
	ErrInvalidAccountDepositor  = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAccountDepositor, codes.InvalidArgument, "invalid account depositor")
	ErrUnauthorizedDepositScope = sdkerrors.RegisterWithGRPCCode(ModuleName, errUnauthorizedDepositScope, codes.PermissionDenied, "unauthorized deposit scope")
	ErrInvalidDeposit           = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidDeposit, codes.InvalidArgument, "invalid deposit")
	ErrInvalidAuthzScope        = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidAuthzScope, codes.InvalidArgument, "invalid authz scope")
)
