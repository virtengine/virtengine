package v1beta3

import (
	"errors"
)

var (
	// ErrClientNotFound is a new error with message "Client not found"
	ErrClientNotFound = errors.New("client not found")
	ErrNodeNotSynced  = errors.New("rpc node is not catching up")
)
