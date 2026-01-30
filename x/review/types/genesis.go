// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Genesis state and parameters
package types

import (
	"fmt"

	reviewv1 "github.com/virtengine/virtengine/sdk/go/node/review/v1"
)

// Type alias for Params from generated proto
type Params = reviewv1.Params

// Default parameter values
const (
	// DefaultMinReviewTextLength is the default minimum review text length
	DefaultMinReviewTextLength = 10

	// DefaultMaxReviewTextLength is the default maximum review text length
	DefaultMaxReviewTextLength = 2000

	// DefaultReviewCooldownSeconds is the default cooldown between reviews from same reviewer
	DefaultReviewCooldownSeconds = 86400 // 24 hours

	// DefaultMaxReviewsPerProvider is the default max reviews to store per provider
	DefaultMaxReviewsPerProvider = 1000
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinReviewInterval:     uint64(DefaultReviewCooldownSeconds),
		MaxCommentLength:      uint64(DefaultMaxReviewTextLength),
		RequireCompletedOrder: true,
		ReviewWindow:          uint64(DefaultReviewCooldownSeconds * 7), // 7 days
		MinRating:             MinRating,
		MaxRating:             MaxRating,
	}
}

// ValidateParams validates the parameters
func ValidateParams(p *Params) error {
	if p.MaxCommentLength <= 0 {
		return fmt.Errorf("max_comment_length must be positive")
	}

	if p.MinRating > p.MaxRating {
		return fmt.Errorf("min_rating cannot exceed max_rating")
	}

	return nil
}

// GenesisState is the genesis state for the review module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// Reviews are the existing reviews
	Reviews []Review `json:"reviews,omitempty"`

	// Aggregations are the provider rating aggregations
	Aggregations []ProviderAggregation `json:"aggregations,omitempty"`

	// NextReviewSequence is the next review sequence number
	NextReviewSequence uint64 `json:"next_review_sequence"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:             DefaultParams(),
		Reviews:            []Review{},
		Aggregations:       []ProviderAggregation{},
		NextReviewSequence: 1,
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	if err := ValidateParams(&gs.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate all reviews
	reviewIDs := make(map[string]bool)
	for _, review := range gs.Reviews {
		if err := review.Validate(); err != nil {
			return fmt.Errorf("invalid review %s: %w", review.ID.String(), err)
		}

		// Check for duplicate IDs
		idStr := review.ID.String()
		if reviewIDs[idStr] {
			return fmt.Errorf("duplicate review ID: %s", idStr)
		}
		reviewIDs[idStr] = true
	}

	// Validate all aggregations
	providerAddrs := make(map[string]bool)
	for _, agg := range gs.Aggregations {
		if err := agg.Validate(); err != nil {
			return fmt.Errorf("invalid aggregation for provider %s: %w", agg.ProviderAddress, err)
		}

		// Check for duplicate providers
		if providerAddrs[agg.ProviderAddress] {
			return fmt.Errorf("duplicate aggregation for provider: %s", agg.ProviderAddress)
		}
		providerAddrs[agg.ProviderAddress] = true
	}

	return nil
}
