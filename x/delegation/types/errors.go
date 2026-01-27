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
	ErrInvalidDelegator = sdkerrors.Register(ModuleName, 1800, "invalid delegator")

	// ErrInvalidValidator is returned when validator address is invalid
	ErrInvalidValidator = sdkerrors.Register(ModuleName, 1801, "invalid validator")

	// ErrDelegationNotFound is returned when delegation is not found
	ErrDelegationNotFound = sdkerrors.Register(ModuleName, 1802, "delegation not found")

	// ErrUnbondingNotFound is returned when unbonding delegation is not found
	ErrUnbondingNotFound = sdkerrors.Register(ModuleName, 1803, "unbonding delegation not found")

	// ErrRedelegationNotFound is returned when redelegation is not found
	ErrRedelegationNotFound = sdkerrors.Register(ModuleName, 1804, "redelegation not found")

	// ErrInsufficientShares is returned when delegation has insufficient shares
	ErrInsufficientShares = sdkerrors.Register(ModuleName, 1805, "insufficient delegation shares")

	// ErrInsufficientBalance is returned when account has insufficient balance
	ErrInsufficientBalance = sdkerrors.Register(ModuleName, 1806, "insufficient balance")

	// ErrMinDelegationAmount is returned when delegation amount is below minimum
	ErrMinDelegationAmount = sdkerrors.Register(ModuleName, 1807, "delegation amount below minimum")

	// ErrMaxValidators is returned when delegator exceeds max validators
	ErrMaxValidators = sdkerrors.Register(ModuleName, 1808, "max validators per delegator exceeded")

	// ErrMaxRedelegations is returned when max redelegations exceeded
	ErrMaxRedelegations = sdkerrors.Register(ModuleName, 1809, "max redelegations exceeded")

	// ErrSelfRedelegation is returned when redelegating to the same validator
	ErrSelfRedelegation = sdkerrors.Register(ModuleName, 1810, "cannot redelegate to same validator")

	// ErrTransitiveRedelegation is returned when attempting transitive redelegation
	ErrTransitiveRedelegation = sdkerrors.Register(ModuleName, 1811, "transitive redelegation not allowed")

	// ErrInvalidParams is returned when parameters are invalid
	ErrInvalidParams = sdkerrors.Register(ModuleName, 1812, "invalid parameters")

	// ErrValidatorNotFound is returned when validator is not found
	ErrValidatorNotFound = sdkerrors.Register(ModuleName, 1813, "validator not found")

	// ErrRewardNotFound is returned when delegator reward is not found
	ErrRewardNotFound = sdkerrors.Register(ModuleName, 1814, "delegator reward not found")

	// ErrRewardAlreadyClaimed is returned when reward was already claimed
	ErrRewardAlreadyClaimed = sdkerrors.Register(ModuleName, 1815, "reward already claimed")

	// ErrNoDelegations is returned when there are no delegations
	ErrNoDelegations = sdkerrors.Register(ModuleName, 1816, "no delegations found")

	// ErrInvalidAmount is returned when amount is invalid
	ErrInvalidAmount = sdkerrors.Register(ModuleName, 1817, "invalid amount")

	// ErrUnbondingPending is returned when unbonding is still pending
	ErrUnbondingPending = sdkerrors.Register(ModuleName, 1818, "unbonding still pending")

	// ErrValidatorJailed is returned when validator is jailed
	ErrValidatorJailed = sdkerrors.Register(ModuleName, 1819, "validator is jailed")

	// ErrInvalidShares is returned when shares value is invalid
	ErrInvalidShares = sdkerrors.Register(ModuleName, 1820, "invalid shares value")
)
