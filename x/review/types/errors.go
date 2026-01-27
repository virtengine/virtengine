// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Error definitions
package types

import (
	"cosmossdk.io/errors"
)

// Review module error codes
var (
	// ErrInvalidRating is returned when a rating is outside the 1-5 range
	ErrInvalidRating = errors.Register(ModuleName, 2300, "invalid rating: must be between 1 and 5")

	// ErrInvalidReviewText is returned when review text is invalid
	ErrInvalidReviewText = errors.Register(ModuleName, 2301, "invalid review text")

	// ErrOrderNotFound is returned when the order is not found
	ErrOrderNotFound = errors.Register(ModuleName, 2302, "order not found")

	// ErrOrderNotCompleted is returned when trying to review an incomplete order
	ErrOrderNotCompleted = errors.Register(ModuleName, 2303, "order not completed: only completed orders can be reviewed")

	// ErrUnauthorizedReviewer is returned when reviewer is not the order customer
	ErrUnauthorizedReviewer = errors.Register(ModuleName, 2304, "unauthorized: reviewer must be the order customer")

	// ErrDuplicateReview is returned when an order has already been reviewed
	ErrDuplicateReview = errors.Register(ModuleName, 2305, "duplicate review: order has already been reviewed")

	// ErrReviewNotFound is returned when a review is not found
	ErrReviewNotFound = errors.Register(ModuleName, 2306, "review not found")

	// ErrProviderNotFound is returned when the provider is not found
	ErrProviderNotFound = errors.Register(ModuleName, 2307, "provider not found")

	// ErrInvalidReviewID is returned when a review ID is invalid
	ErrInvalidReviewID = errors.Register(ModuleName, 2308, "invalid review ID")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errors.Register(ModuleName, 2309, "invalid address")

	// ErrContentHashMismatch is returned when content hash verification fails
	ErrContentHashMismatch = errors.Register(ModuleName, 2310, "content hash mismatch: review content integrity check failed")

	// ErrReviewTextTooLong is returned when review text exceeds maximum length
	ErrReviewTextTooLong = errors.Register(ModuleName, 2311, "review text too long")

	// ErrReviewTextTooShort is returned when review text is too short
	ErrReviewTextTooShort = errors.Register(ModuleName, 2312, "review text too short")

	// ErrInvalidOrderID is returned when the order ID is invalid
	ErrInvalidOrderID = errors.Register(ModuleName, 2313, "invalid order ID")

	// ErrAggregationNotFound is returned when provider aggregation is not found
	ErrAggregationNotFound = errors.Register(ModuleName, 2314, "provider rating aggregation not found")
)
