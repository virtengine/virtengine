package v1

import (
	cerrors "cosmossdk.io/errors"
)

const (
	errPriceEntryExists uint32 = iota + 1
	errInvalidTimestamp
	errUnauthorizedWriterAddress
	errPriceStalled
	errInvalidFeedContractParams
	errInvalidFeedContractConfig
	errTWAPZeroWeight
)

var (
	// ErrPriceEntryExists is the error when price entry already exists
	ErrPriceEntryExists = cerrors.Register(ModuleName, errPriceEntryExists, "price entry exist")
	// ErrInvalidTimestamp is the error indicating invalid timestamp
	ErrInvalidTimestamp = cerrors.Register(ModuleName, errInvalidTimestamp, "invalid timestamp")
	// ErrUnauthorizedWriterAddress is the error indicating signer is not allowed to add price records
	ErrUnauthorizedWriterAddress = cerrors.Register(ModuleName, errUnauthorizedWriterAddress, "unauthorized writer address")
	// ErrPriceStalled is the error when price data is stale
	ErrPriceStalled = cerrors.Register(ModuleName, errPriceStalled, "price stalled")
	// ErrInvalidFeedContractParams is the error when feed contract params are invalid
	ErrInvalidFeedContractParams = cerrors.Register(ModuleName, errInvalidFeedContractParams, "invalid feed contract params")
	// ErrInvalidFeedContractConfig is the error when feed contract config is invalid
	ErrInvalidFeedContractConfig = cerrors.Register(ModuleName, errInvalidFeedContractConfig, "invalid feed contract config")
	ErrTWAPZeroWeight            = cerrors.Register(ModuleName, errTWAPZeroWeight, "invalid TWAP calculation: zero weight")
)
