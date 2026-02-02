package v1

import (
	sdkerrors "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

const (
	errCodeEmptyProvider uint32 = iota + 1
	errCodeSameAccount
	errCodeInternal
	errCodeOverOrder
	errCodeAttributeMismatch
	errCodeUnknownBid
	errCodeUnknownLease
	errCodeUnknownLeaseForBid
	errCodeUnknownOrderForBid
	errCodeLeaseNotActive
	errCodeBidNotActive
	errCodeBidNotOpen
	errCodeOrderNotOpen
	errCodeNoLeaseForOrder
	errCodeOrderNotFound
	errCodeGroupNotFound
	errCodeGroupNotOpen
	errCodeBidNotFound
	errCodeBidZeroPrice
	errCodeLeaseNotFound
	errCodeBidExists
	errCodeInvalidPrice
	errCodeOrderActive
	errCodeOrderClosed
	errCodeOrderExists
	errCodeOrderDurationExceeded
	errCodeOrderTooEarly
	errInvalidDeposit
	errInvalidParam
	errUnknownProvider
	errInvalidBid
	errCodeCapabilitiesMismatch
	errInvalidLeaseClosedReason
	errInvalidEscrowID
)

var (
	// ErrEmptyProvider is the error when provider is empty
	ErrEmptyProvider = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeEmptyProvider, codes.InvalidArgument, "empty provider")
	// ErrSameAccount is the error when owner and provider are the same account
	ErrSameAccount = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeSameAccount, codes.InvalidArgument, "owner and provider are the same account")
	// ErrInternal is the error for internal error
	ErrInternal = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeInternal, codes.Internal, "internal error")
	// ErrBidOverOrder is the error when bid price is above max order price
	ErrBidOverOrder = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOverOrder, codes.InvalidArgument, "bid price above max order price")
	// ErrAttributeMismatch is the error for attribute mismatch
	ErrAttributeMismatch = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeAttributeMismatch, codes.InvalidArgument, "attribute mismatch")
	// ErrCapabilitiesMismatch is the error for capabilities mismatch
	ErrCapabilitiesMismatch = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeCapabilitiesMismatch, codes.InvalidArgument, "capabilities mismatch")
	// ErrUnknownBid is the error for unknown bid
	ErrUnknownBid = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeUnknownBid, codes.NotFound, "unknown bid")
	// ErrUnknownLease is the error for unknown lease
	ErrUnknownLease = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeUnknownLease, codes.NotFound, "unknown lease")
	// ErrUnknownLeaseForBid is the error when lease is unknown for bid
	ErrUnknownLeaseForBid = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeUnknownLeaseForBid, codes.NotFound, "unknown lease for bid")
	// ErrUnknownOrderForBid is the error when order is unknown for bid
	ErrUnknownOrderForBid = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeUnknownOrderForBid, codes.NotFound, "unknown order for bid")
	// ErrLeaseNotActive is the error when lease is not active
	ErrLeaseNotActive = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeLeaseNotActive, codes.FailedPrecondition, "lease not active")
	// ErrBidNotActive is the error when bid is not active
	ErrBidNotActive = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeBidNotActive, codes.FailedPrecondition, "bid not active")
	// ErrBidNotOpen is the error when bid is not open
	ErrBidNotOpen = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeBidNotOpen, codes.FailedPrecondition, "bid not open")
	// ErrNoLeaseForOrder is the error when there is no lease for order
	ErrNoLeaseForOrder = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeNoLeaseForOrder, codes.NotFound, "no lease for order")
	// ErrOrderNotFound order not found
	ErrOrderNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderNotFound, codes.NotFound, "invalid order: order not found")
	// ErrGroupNotFound group not found
	ErrGroupNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeGroupNotFound, codes.NotFound, "group not found")
	// ErrGroupNotOpen group not open
	ErrGroupNotOpen = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeGroupNotOpen, codes.FailedPrecondition, "group not open")
	// ErrOrderNotOpen order not open
	ErrOrderNotOpen = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderNotOpen, codes.FailedPrecondition, "bid: order not open")
	// ErrBidNotFound bid not found
	ErrBidNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeBidNotFound, codes.NotFound, "invalid bid: bid not found")
	// ErrBidZeroPrice zero price
	ErrBidZeroPrice = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeBidZeroPrice, codes.InvalidArgument, "invalid bid: zero price")
	// ErrLeaseNotFound lease not found
	ErrLeaseNotFound = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeLeaseNotFound, codes.NotFound, "invalid lease: lease not found")
	// ErrBidExists bid exists
	ErrBidExists = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeBidExists, codes.AlreadyExists, "invalid bid: bid exists from provider")
	// ErrBidInvalidPrice bid invalid price
	ErrBidInvalidPrice = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeInvalidPrice, codes.InvalidArgument, "bid price is invalid")
	// ErrOrderActive order active
	ErrOrderActive = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderActive, codes.FailedPrecondition, "order active")
	// ErrOrderClosed order closed
	ErrOrderClosed = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderClosed, codes.FailedPrecondition, "order closed")
	// ErrOrderExists indicates a new order was proposed overwrite the existing store key
	ErrOrderExists = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderExists, codes.AlreadyExists, "order already exists in store")
	// ErrOrderTooEarly to match bid
	ErrOrderTooEarly = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderTooEarly, codes.FailedPrecondition, "order: chain height to low for bidding")
	// ErrOrderDurationExceeded order should be closed
	ErrOrderDurationExceeded = sdkerrors.RegisterWithGRPCCode(ModuleName, errCodeOrderDurationExceeded, codes.FailedPrecondition, "order duration has exceeded the bidding duration")
	// ErrInvalidDeposit indicates an invalid deposit
	ErrInvalidDeposit = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidDeposit, codes.InvalidArgument, "deposit invalid")
	// ErrInvalidParam indicates an invalid chain parameter
	ErrInvalidParam = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidParam, codes.InvalidArgument, "parameter invalid")
	// ErrUnknownProvider is the error for unknown provider
	ErrUnknownProvider = sdkerrors.RegisterWithGRPCCode(ModuleName, errUnknownProvider, codes.NotFound, "unknown provider")
	// ErrInvalidBid is the error for invalid bid
	ErrInvalidBid = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidBid, codes.InvalidArgument, "invalid bid")
	// ErrInvalidLeaseClosedReason indicates reason for lease close does not match context
	ErrInvalidLeaseClosedReason = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidLeaseClosedReason, codes.InvalidArgument, "invalid lease closed reason")
	ErrInvalidEscrowID          = sdkerrors.RegisterWithGRPCCode(ModuleName, errInvalidEscrowID, codes.InvalidArgument, "invalid escrow id")
)
