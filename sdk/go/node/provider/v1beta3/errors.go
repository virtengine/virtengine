package v1beta3

import (
	"errors"
)

var (
	// ErrInvalidProviderURI register error code for invalid provider uri
	ErrInvalidProviderURI = errors.New("invalid provider: invalid host uri")

	// ErrNotAbsProviderURI register error code for not absolute provider uri
	ErrNotAbsProviderURI = errors.New("invalid provider: not absolute host uri")

	// ErrProviderNotFound provider not found
	ErrProviderNotFound = errors.New("invalid provider: address not found")

	// ErrProviderExists provider already exists
	ErrProviderExists = errors.New("invalid provider: already exists")

	// ErrInvalidAddress invalid provider address
	ErrInvalidAddress = errors.New("invalid address")

	// ErrAttributes error code for provider attribute problems
	ErrAttributes = errors.New("attribute specification error")

	// ErrIncompatibleAttributes error code for attributes update
	ErrIncompatibleAttributes = errors.New("attributes cannot be changed")

	// ErrInvalidInfoWebsite register error code for invalid info website
	ErrInvalidInfoWebsite = errors.New("invalid provider: invalid info website")
)
