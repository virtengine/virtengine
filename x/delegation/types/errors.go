// Package types contains types for the delegation module.
//
// VE-922: Delegation module errors
package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// Delegation module sentinel errors
var (
	// ErrInvalidDelegator is returned when delegator address is invalid
	ErrInvalidDelegator = sdkerrors.Register(ModuleName, 2, "invalid delegator")

	// ErrInvalidValidator is returned when validator address is invalid
	ErrInvalidValidator = sdkerrors.Register(ModuleName, 3, "invalid validator")

	// ErrDelegationNotFound is returned when delegation is not found
	ErrDelegationNotFound = sdkerrors.Register(ModuleName, 4, "delegation not found")

	// ErrUnbondingNotFound is returned when unbonding delegation is not found
	ErrUnbondingNotFound = sdkerrors.Register(ModuleName, 5, "unbonding delegation not found")

	// ErrRedelegationNotFound is returned when redelegation is not found
	ErrRedelegationNotFound = sdkerrors.Register(ModuleName, 6, "redelegation not found")

	// ErrInsufficientShares is returned when delegation has insufficient shares
	ErrInsufficientShares = sdkerrors.Register(ModuleName, 7, "insufficient delegation shares")

	// ErrInsufficientBalance is returned when account has insufficient balance
	ErrInsufficientBalance = sdkerrors.Register(ModuleName, 8, "insufficient balance")

	// ErrMinDelegationAmount is returned when delegation amount is below minimum
	ErrMinDelegationAmount = sdkerrors.Register(ModuleName, 9, "delegation amount below minimum")

	// ErrMaxValidators is returned when delegator exceeds max validators
	ErrMaxValidators = sdkerrors.Register(ModuleName, 10, "max validators per delegator exceeded")

	// ErrMaxRedelegations is returned when max redelegations exceeded
	ErrMaxRedelegations = sdkerrors.Register(ModuleName, 11, "max redelegations exceeded")

	// ErrSelfRedelegation is returned when redelegating to the same validator
	ErrSelfRedelegation = sdkerrors.Register(ModuleName, 12, "cannot redelegate to same validator")

	// ErrTransitiveRedelegation is returned when attempting transitive redelegation
	ErrTransitiveRedelegation = sdkerrors.Register(ModuleName, 13, "transitive redelegation not allowed")

	// ErrInvalidParams is returned when parameters are invalid
	ErrInvalidParams = sdkerrors.Register(ModuleName, 14, "invalid parameters")

	// ErrValidatorNotFound is returned when validator is not found
	ErrValidatorNotFound = sdkerrors.Register(ModuleName, 15, "validator not found")

	// ErrRewardNotFound is returned when delegator reward is not found
	ErrRewardNotFound = sdkerrors.Register(ModuleName, 16, "delegator reward not found")

	// ErrRewardAlreadyClaimed is returned when reward was already claimed
	ErrRewardAlreadyClaimed = sdkerrors.Register(ModuleName, 17, "reward already claimed")

	// ErrNoDelegations is returned when there are no delegations
	ErrNoDelegations = sdkerrors.Register(ModuleName, 18, "no delegations found")

	// ErrInvalidAmount is returned when amount is invalid
	ErrInvalidAmount = sdkerrors.Register(ModuleName, 19, "invalid amount")

	// ErrUnbondingPending is returned when unbonding is still pending
	ErrUnbondingPending = sdkerrors.Register(ModuleName, 20, "unbonding still pending")

	// ErrValidatorJailed is returned when validator is jailed
	ErrValidatorJailed = sdkerrors.Register(ModuleName, 21, "validator is jailed")

	// ErrInvalidShares is returned when shares value is invalid
	ErrInvalidShares = sdkerrors.Register(ModuleName, 22, "invalid shares value")
)
