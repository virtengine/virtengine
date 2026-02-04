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
	ErrOfferingNotFound = errors.Register("marketplace", 2200, "offering not found")

	// ErrOfferingExists indicates the offering already exists
	ErrOfferingExists = errors.Register("marketplace", 2201, "offering already exists")

	// ErrOfferingNotActive indicates the offering is not active
	ErrOfferingNotActive = errors.Register("marketplace", 2202, "offering is not active")

	// ErrOfferingMaxOrders indicates the offering has reached max orders
	ErrOfferingMaxOrders = errors.Register("marketplace", 2203, "offering has reached maximum orders")

	// ErrOrderNotFound indicates the order was not found
	ErrOrderNotFound = errors.Register("marketplace", 2204, "order not found")

	// ErrOrderExists indicates the order already exists
	ErrOrderExists = errors.Register("marketplace", 2205, "order already exists")

	// ErrOrderNotOpen indicates the order is not open for bids
	ErrOrderNotOpen = errors.Register("marketplace", 2206, "order is not open for bids")

	// ErrOrderExpired indicates the order has expired
	ErrOrderExpired = errors.Register("marketplace", 2207, "order has expired")

	// ErrInvalidOrderState indicates invalid order state
	ErrInvalidOrderState = errors.Register("marketplace", 2208, "invalid order state")

	// ErrInvalidStateTransition indicates an invalid state transition
	ErrInvalidStateTransition = errors.Register("marketplace", 2209, "invalid state transition")

	// ErrAllocationNotFound indicates the allocation was not found
	ErrAllocationNotFound = errors.Register("marketplace", 2210, "allocation not found")

	// ErrAllocationExists indicates the allocation already exists
	ErrAllocationExists = errors.Register("marketplace", 2211, "allocation already exists")

	// ErrBidNotFound indicates the bid was not found
	ErrBidNotFound = errors.Register("marketplace", 2212, "bid not found")

	// ErrBidExists indicates the bid already exists
	ErrBidExists = errors.Register("marketplace", 2213, "bid already exists")

	// ErrBidPriceTooHigh indicates bid price exceeds order max
	ErrBidPriceTooHigh = errors.Register("marketplace", 2214, "bid price exceeds order maximum")

	// ErrBidNotOpen indicates bid is not open
	ErrBidNotOpen = errors.Register("marketplace", 2215, "bid is not open")

	// ErrIdentityGatingFailed indicates identity gating checks failed
	ErrIdentityGatingFailed = errors.Register("marketplace", 2216, "identity gating failed")

	// ErrInsufficientIdentityScore indicates identity score is too low
	ErrInsufficientIdentityScore = errors.Register("marketplace", 2217, "insufficient identity score")

	// ErrIdentityNotVerified indicates identity is not verified
	ErrIdentityNotVerified = errors.Register("marketplace", 2218, "identity not verified")

	// ErrEmailNotVerified indicates email is not verified
	ErrEmailNotVerified = errors.Register("marketplace", 2219, "email not verified")

	// ErrDomainNotVerified indicates domain is not verified
	ErrDomainNotVerified = errors.Register("marketplace", 2220, "domain not verified")

	// ErrMFARequired indicates MFA is required
	ErrMFARequired = errors.Register("marketplace", 2221, "MFA verification required")

	// ErrMFANotSatisfied indicates MFA requirements not satisfied
	ErrMFANotSatisfied = errors.Register("marketplace", 2222, "MFA requirements not satisfied")

	// ErrMFAChallengeFailed indicates MFA challenge failed
	ErrMFAChallengeFailed = errors.Register("marketplace", 2223, "MFA challenge failed")

	// ErrMFAChallengeExpired indicates MFA challenge expired
	ErrMFAChallengeExpired = errors.Register("marketplace", 2224, "MFA challenge expired")

	// ErrWaldurSyncFailed indicates Waldur sync failed
	ErrWaldurSyncFailed = errors.Register("marketplace", 2225, "Waldur synchronization failed")

	// ErrWaldurCallbackInvalid indicates invalid Waldur callback
	ErrWaldurCallbackInvalid = errors.Register("marketplace", 2226, "invalid Waldur callback")

	// ErrWaldurCallbackExpired indicates Waldur callback expired
	ErrWaldurCallbackExpired = errors.Register("marketplace", 2227, "Waldur callback expired")

	// ErrWaldurNonceReplayed indicates nonce was replayed
	ErrWaldurNonceReplayed = errors.Register("marketplace", 2228, "Waldur callback nonce already processed")

	// ErrWaldurSignatureInvalid indicates invalid Waldur signature
	ErrWaldurSignatureInvalid = errors.Register("marketplace", 2229, "invalid Waldur callback signature")

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = errors.Register("marketplace", 2230, "unauthorized")

	// ErrNotProvider indicates account is not a provider
	ErrNotProvider = errors.Register("marketplace", 2231, "account is not a provider")

	// ErrNotCustomer indicates account is not the customer
	ErrNotCustomer = errors.Register("marketplace", 2232, "account is not the customer")

	// ErrInvalidEncryptedPayload indicates invalid encrypted payload
	ErrInvalidEncryptedPayload = errors.Register("marketplace", 2233, "invalid encrypted payload")

	// ErrDecryptionFailed indicates decryption failed
	ErrDecryptionFailed = errors.Register("marketplace", 2234, "decryption failed")

	// ErrEncryptionRequired indicates encryption is required
	ErrEncryptionRequired = errors.Register("marketplace", 2235, "encryption required for sensitive data")

	// ErrEventSubscriptionNotFound indicates subscription not found
	ErrEventSubscriptionNotFound = errors.Register("marketplace", 2236, "event subscription not found")

	// ErrCheckpointNotFound indicates checkpoint not found
	ErrCheckpointNotFound = errors.Register("marketplace", 2237, "event checkpoint not found")

	// ErrInvalidEventSequence indicates invalid event sequence
	ErrInvalidEventSequence = errors.Register("marketplace", 2238, "invalid event sequence")

	// ErrPricingInvalid indicates pricing validation failed
	ErrPricingInvalid = errors.Register("marketplace", 2239, "pricing validation failed")

	// ErrInvalidRequest indicates the request is invalid
	ErrInvalidRequest = errors.Register("marketplace", 2240, "invalid request")
)

// WrapIdentityGatingError wraps an identity gating error with context
func WrapIdentityGatingError(err *IdentityGatingError) error {
	if err == nil || !err.HasErrors() {
		return nil
	}
	return ErrIdentityGatingFailed.Wrap(err.Error())
}
