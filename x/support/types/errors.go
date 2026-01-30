package types

import (
	"cosmossdk.io/errors"
)

// Support module sentinel errors
var (
	ErrTicketNotFound       = errors.Register(ModuleName, 1, "ticket not found")
	ErrTicketAlreadyExists  = errors.Register(ModuleName, 2, "ticket already exists")
	ErrInvalidTicketID      = errors.Register(ModuleName, 3, "invalid ticket ID")
	ErrInvalidTicketStatus  = errors.Register(ModuleName, 4, "invalid ticket status")
	ErrInvalidTicketPriority = errors.Register(ModuleName, 5, "invalid ticket priority")
	ErrInvalidAddress       = errors.Register(ModuleName, 6, "invalid address")
	ErrUnauthorized         = errors.Register(ModuleName, 7, "unauthorized")
	ErrTicketClosed         = errors.Register(ModuleName, 8, "ticket is closed")
	ErrTicketNotAssigned    = errors.Register(ModuleName, 9, "ticket not assigned")
	ErrTicketAlreadyAssigned = errors.Register(ModuleName, 10, "ticket already assigned")
	ErrInvalidEncryptedPayload = errors.Register(ModuleName, 11, "invalid encrypted payload")
	ErrRateLimitExceeded    = errors.Register(ModuleName, 12, "rate limit exceeded")
	ErrMaxResponsesExceeded = errors.Register(ModuleName, 13, "maximum responses exceeded")
	ErrInvalidParams        = errors.Register(ModuleName, 14, "invalid module parameters")
	ErrTicketResolved       = errors.Register(ModuleName, 15, "ticket is already resolved")
	ErrCannotReopen         = errors.Register(ModuleName, 16, "cannot reopen ticket")
	ErrInvalidResourceRef   = errors.Register(ModuleName, 17, "invalid resource reference")
	ErrResponseNotFound     = errors.Register(ModuleName, 18, "response not found")
	ErrInvalidCategory      = errors.Register(ModuleName, 19, "invalid ticket category")
	ErrTicketNotResolved    = errors.Register(ModuleName, 20, "ticket is not resolved")
	ErrSelfAssignment       = errors.Register(ModuleName, 21, "cannot assign ticket to self")
)
