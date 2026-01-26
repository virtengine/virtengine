// Package marketplace provides types for the marketplace on-chain module.
//
// VE-300 to VE-304: Marketplace on-chain module
// This file defines error types for the marketplace module.
package marketplace

import (
	"cosmossdk.io/errors"
)

// Module sentinel errors
var (
	// ErrOfferingNotFound indicates the offering was not found
	ErrOfferingNotFound = errors.Register("marketplace", 1, "offering not found")

	// ErrOfferingExists indicates the offering already exists
	ErrOfferingExists = errors.Register("marketplace", 2, "offering already exists")

	// ErrOfferingNotActive indicates the offering is not active
	ErrOfferingNotActive = errors.Register("marketplace", 3, "offering is not active")

	// ErrOfferingMaxOrders indicates the offering has reached max orders
	ErrOfferingMaxOrders = errors.Register("marketplace", 4, "offering has reached maximum orders")

	// ErrOrderNotFound indicates the order was not found
	ErrOrderNotFound = errors.Register("marketplace", 10, "order not found")

	// ErrOrderExists indicates the order already exists
	ErrOrderExists = errors.Register("marketplace", 11, "order already exists")

	// ErrOrderNotOpen indicates the order is not open for bids
	ErrOrderNotOpen = errors.Register("marketplace", 12, "order is not open for bids")

	// ErrOrderExpired indicates the order has expired
	ErrOrderExpired = errors.Register("marketplace", 13, "order has expired")

	// ErrInvalidOrderState indicates invalid order state
	ErrInvalidOrderState = errors.Register("marketplace", 14, "invalid order state")

	// ErrInvalidStateTransition indicates an invalid state transition
	ErrInvalidStateTransition = errors.Register("marketplace", 15, "invalid state transition")

	// ErrAllocationNotFound indicates the allocation was not found
	ErrAllocationNotFound = errors.Register("marketplace", 20, "allocation not found")

	// ErrAllocationExists indicates the allocation already exists
	ErrAllocationExists = errors.Register("marketplace", 21, "allocation already exists")

	// ErrBidNotFound indicates the bid was not found
	ErrBidNotFound = errors.Register("marketplace", 30, "bid not found")

	// ErrBidExists indicates the bid already exists
	ErrBidExists = errors.Register("marketplace", 31, "bid already exists")

	// ErrBidPriceTooHigh indicates bid price exceeds order max
	ErrBidPriceTooHigh = errors.Register("marketplace", 32, "bid price exceeds order maximum")

	// ErrBidNotOpen indicates bid is not open
	ErrBidNotOpen = errors.Register("marketplace", 33, "bid is not open")

	// ErrIdentityGatingFailed indicates identity gating checks failed
	ErrIdentityGatingFailed = errors.Register("marketplace", 40, "identity gating failed")

	// ErrInsufficientIdentityScore indicates identity score is too low
	ErrInsufficientIdentityScore = errors.Register("marketplace", 41, "insufficient identity score")

	// ErrIdentityNotVerified indicates identity is not verified
	ErrIdentityNotVerified = errors.Register("marketplace", 42, "identity not verified")

	// ErrEmailNotVerified indicates email is not verified
	ErrEmailNotVerified = errors.Register("marketplace", 43, "email not verified")

	// ErrDomainNotVerified indicates domain is not verified
	ErrDomainNotVerified = errors.Register("marketplace", 44, "domain not verified")

	// ErrMFARequired indicates MFA is required
	ErrMFARequired = errors.Register("marketplace", 50, "MFA verification required")

	// ErrMFANotSatisfied indicates MFA requirements not satisfied
	ErrMFANotSatisfied = errors.Register("marketplace", 51, "MFA requirements not satisfied")

	// ErrMFAChallengeFailed indicates MFA challenge failed
	ErrMFAChallengeFailed = errors.Register("marketplace", 52, "MFA challenge failed")

	// ErrMFAChallengeExpired indicates MFA challenge expired
	ErrMFAChallengeExpired = errors.Register("marketplace", 53, "MFA challenge expired")

	// ErrWaldurSyncFailed indicates Waldur sync failed
	ErrWaldurSyncFailed = errors.Register("marketplace", 60, "Waldur synchronization failed")

	// ErrWaldurCallbackInvalid indicates invalid Waldur callback
	ErrWaldurCallbackInvalid = errors.Register("marketplace", 61, "invalid Waldur callback")

	// ErrWaldurCallbackExpired indicates Waldur callback expired
	ErrWaldurCallbackExpired = errors.Register("marketplace", 62, "Waldur callback expired")

	// ErrWaldurNonceReplayed indicates nonce was replayed
	ErrWaldurNonceReplayed = errors.Register("marketplace", 63, "Waldur callback nonce already processed")

	// ErrWaldurSignatureInvalid indicates invalid Waldur signature
	ErrWaldurSignatureInvalid = errors.Register("marketplace", 64, "invalid Waldur callback signature")

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = errors.Register("marketplace", 70, "unauthorized")

	// ErrNotProvider indicates account is not a provider
	ErrNotProvider = errors.Register("marketplace", 71, "account is not a provider")

	// ErrNotCustomer indicates account is not the customer
	ErrNotCustomer = errors.Register("marketplace", 72, "account is not the customer")

	// ErrInvalidEncryptedPayload indicates invalid encrypted payload
	ErrInvalidEncryptedPayload = errors.Register("marketplace", 80, "invalid encrypted payload")

	// ErrDecryptionFailed indicates decryption failed
	ErrDecryptionFailed = errors.Register("marketplace", 81, "decryption failed")

	// ErrEncryptionRequired indicates encryption is required
	ErrEncryptionRequired = errors.Register("marketplace", 82, "encryption required for sensitive data")

	// ErrEventSubscriptionNotFound indicates subscription not found
	ErrEventSubscriptionNotFound = errors.Register("marketplace", 90, "event subscription not found")

	// ErrCheckpointNotFound indicates checkpoint not found
	ErrCheckpointNotFound = errors.Register("marketplace", 91, "event checkpoint not found")

	// ErrInvalidEventSequence indicates invalid event sequence
	ErrInvalidEventSequence = errors.Register("marketplace", 92, "invalid event sequence")
)

// WrapIdentityGatingError wraps an identity gating error with context
func WrapIdentityGatingError(err *IdentityGatingError) error {
	if err == nil || !err.HasErrors() {
		return nil
	}
	return ErrIdentityGatingFailed.Wrap(err.Error())
}
