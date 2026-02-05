package types

import (
	"cosmossdk.io/errors"
)

// Support module sentinel errors
var (
	// Reference errors
	ErrRefNotFound             = errors.Register(ModuleName, 1, "external ticket reference not found")
	ErrRefAlreadyExists        = errors.Register(ModuleName, 2, "external ticket reference already exists")
	ErrInvalidResourceRef      = errors.Register(ModuleName, 3, "invalid resource reference")
	ErrInvalidExternalSystem   = errors.Register(ModuleName, 4, "invalid external system")
	ErrInvalidExternalTicketID = errors.Register(ModuleName, 5, "invalid external ticket ID")
	ErrInvalidExternalURL      = errors.Register(ModuleName, 6, "invalid external URL")
	ErrInvalidAddress          = errors.Register(ModuleName, 7, "invalid address")
	ErrUnauthorized            = errors.Register(ModuleName, 8, "unauthorized")
	ErrInvalidParams           = errors.Register(ModuleName, 9, "invalid module parameters")
	ErrInvalidResourceType     = errors.Register(ModuleName, 10, "invalid resource type")
	ErrInvalidSupportRequest   = errors.Register(ModuleName, 11, "invalid support request")
	ErrInvalidSupportResponse  = errors.Register(ModuleName, 12, "invalid support response")
	ErrSupportRequestNotFound  = errors.Register(ModuleName, 13, "support request not found")
	ErrSupportResponseNotFound = errors.Register(ModuleName, 14, "support response not found")
	ErrInvalidStatusTransition = errors.Register(ModuleName, 15, "invalid status transition")
	ErrInvalidRetentionPolicy  = errors.Register(ModuleName, 16, "invalid retention policy")
	ErrInvalidPayload          = errors.Register(ModuleName, 17, "invalid payload")
	ErrMaxResponsesExceeded    = errors.Register(ModuleName, 18, "max responses exceeded")
	ErrRequestArchived         = errors.Register(ModuleName, 19, "support request archived")
)
