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
)
