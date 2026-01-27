package v1beta4

import (
	"errors"
)

var (
	// ErrEmptyProvider is the error when provider is empty
	ErrEmptyProvider = errors.New("empty provider")
	// ErrSameAccount is the error when owner and provider are the same account
	ErrSameAccount = errors.New("owner and provider are the same account")
	// ErrInternal is the error for internal error
	ErrInternal = errors.New("internal error")
	// ErrBidOverOrder is the error when bid price is above max order price
	ErrBidOverOrder = errors.New("bid price above max order price")
	// ErrAttributeMismatch is the error for attribute mismatch
	ErrAttributeMismatch = errors.New("attribute mismatch")
	// ErrCapabilitiesMismatch is the error for capabilities mismatch
	ErrCapabilitiesMismatch = errors.New("capabilities mismatch")
	// ErrUnknownBid is the error for unknown bid
	ErrUnknownBid = errors.New("unknown bid")
	// ErrUnknownLease is the error for unknown bid
	ErrUnknownLease = errors.New("unknown lease")
	// ErrUnknownLeaseForBid is the error when lease is unknown for bid
	ErrUnknownLeaseForBid = errors.New("unknown lease for bid")
	// ErrUnknownOrderForBid is the error when order is unknown for bid
	ErrUnknownOrderForBid = errors.New("unknown order for bid")
	// ErrLeaseNotActive is the error when lease is not active
	ErrLeaseNotActive = errors.New("lease not active")
	// ErrBidNotActive is the error when bid is not matched
	ErrBidNotActive = errors.New("bid not active")
	// ErrBidNotOpen is the error when bid is not matched
	ErrBidNotOpen = errors.New("bid not open")
	// ErrNoLeaseForOrder is the error when there is no lease for order
	ErrNoLeaseForOrder = errors.New("no lease for order")
	// ErrOrderNotFound order not found
	ErrOrderNotFound = errors.New("invalid order: order not found")
	// ErrGroupNotFound order not found
	ErrGroupNotFound = errors.New("order not found")
	// ErrGroupNotOpen order not found
	ErrGroupNotOpen = errors.New("order not open")
	// ErrOrderNotOpen order not found
	ErrOrderNotOpen = errors.New("bid: order not open")
	// ErrBidNotFound bid not found
	ErrBidNotFound = errors.New("invalid bid: bid not found")
	// ErrBidZeroPrice zero price
	ErrBidZeroPrice = errors.New("invalid bid: zero price")
	// ErrLeaseNotFound lease not found
	ErrLeaseNotFound = errors.New("invalid lease: lease not found")
	// ErrBidExists bid exists
	ErrBidExists = errors.New("invalid bid: bid exists from provider")
	// ErrBidInvalidPrice bid invalid price
	ErrBidInvalidPrice = errors.New("bid price is invalid")
	// ErrOrderActive order active
	ErrOrderActive = errors.New("order active")
	// ErrOrderClosed order closed
	ErrOrderClosed = errors.New("order closed")
	// ErrOrderExists indicates a new order was proposed overwrite the existing store key
	ErrOrderExists = errors.New("order already exists in store")
	// ErrOrderTooEarly to match bid
	ErrOrderTooEarly = errors.New("order: chain height to low for bidding")
	// ErrOrderDurationExceeded order should be closed
	ErrOrderDurationExceeded = errors.New("order duration has exceeded the bidding duration")
	// ErrInvalidDeposit indicates an invalid deposit
	ErrInvalidDeposit = errors.New("Deposit invalid")
	// ErrInvalidParam indicates an invalid chain parameter
	ErrInvalidParam = errors.New("parameter invalid")
	// ErrUnknownProvider indicates an invalid chain parameter
	ErrUnknownProvider = errors.New("unknown provider")
	// ErrInvalidBid indicates an invalid chain parameter
	ErrInvalidBid = errors.New("unknown provider")
)
