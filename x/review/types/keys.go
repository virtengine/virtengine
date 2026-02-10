// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Store keys and prefixes
package types

const (
	// ModuleName is the name of the review module
	ModuleName = "review"

	// StoreKey is the store key for the review module
	StoreKey = ModuleName

	// RouterKey is the router key for the review module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the review module
	QuerierRoute = ModuleName
)

// Key prefixes for review store
var (
	// ReviewPrefix is the prefix for review storage
	ReviewPrefix = []byte{0x01}

	// ProviderAggregationPrefix is the prefix for provider rating aggregations
	ProviderAggregationPrefix = []byte{0x02}

	// ReviewerIndexPrefix is the prefix for reviewer-to-reviews index
	ReviewerIndexPrefix = []byte{0x03}

	// ProviderIndexPrefix is the prefix for provider-to-reviews index
	ProviderIndexPrefix = []byte{0x04}

	// OrderIndexPrefix is the prefix for order-to-review index
	OrderIndexPrefix = []byte{0x05}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x10}

	// SequenceKeyReview is the sequence key for reviews
	SequenceKeyReview = []byte{0x20}
)

// GetReviewKey returns the key for a review
func GetReviewKey(reviewID string) []byte {
	return append(ReviewPrefix, []byte(reviewID)...)
}

// GetProviderAggregationKey returns the key for a provider's aggregation
func GetProviderAggregationKey(providerAddr string) []byte {
	return append(ProviderAggregationPrefix, []byte(providerAddr)...)
}

// GetReviewerIndexKey returns the index key for a reviewer
func GetReviewerIndexKey(reviewerAddr string) []byte {
	return append(ReviewerIndexPrefix, []byte(reviewerAddr)...)
}

// GetProviderIndexKey returns the index key for a provider
func GetProviderIndexKey(providerAddr string) []byte {
	return append(ProviderIndexPrefix, []byte(providerAddr)...)
}

// GetOrderIndexKey returns the index key for an order
func GetOrderIndexKey(orderID string) []byte {
	return append(OrderIndexPrefix, []byte(orderID)...)
}
