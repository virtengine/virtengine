// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Event definitions
package types

// Event types for the review module
const (
	// EventTypeReviewSubmitted is emitted when a new review is submitted
	EventTypeReviewSubmitted = "review_submitted"

	// EventTypeReviewUpdated is emitted when a review is updated
	EventTypeReviewUpdated = "review_updated"

	// EventTypeReviewDeleted is emitted when a review is deleted (by moderator)
	EventTypeReviewDeleted = "review_deleted"

	// EventTypeAggregationUpdated is emitted when provider aggregation is updated
	EventTypeAggregationUpdated = "aggregation_updated"
)

// Attribute keys for review events
const (
	// AttributeKeyReviewID is the review ID attribute
	AttributeKeyReviewID = "review_id"

	// AttributeKeyOrderID is the order ID attribute
	AttributeKeyOrderID = "order_id"

	// AttributeKeyProviderAddress is the provider address attribute
	AttributeKeyProviderAddress = "provider_address"

	// AttributeKeyReviewerAddress is the reviewer address attribute
	AttributeKeyReviewerAddress = "reviewer_address"

	// AttributeKeyRating is the rating attribute
	AttributeKeyRating = "rating"

	// AttributeKeyContentHash is the content hash attribute
	AttributeKeyContentHash = "content_hash"

	// AttributeKeyAverageRating is the average rating attribute
	AttributeKeyAverageRating = "average_rating"

	// AttributeKeyTotalReviews is the total reviews count attribute
	AttributeKeyTotalReviews = "total_reviews"

	// AttributeKeyModeratorAddress is the moderator address for delete operations
	AttributeKeyModeratorAddress = "moderator_address"

	// AttributeKeyDeleteReason is the reason for review deletion
	AttributeKeyDeleteReason = "delete_reason"
)
