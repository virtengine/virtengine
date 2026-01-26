// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Genesis state and parameters
package types

import (
	"fmt"
)

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

// Params contains the review module parameters
type Params struct {
	// MinReviewTextLength is the minimum review text length
	MinReviewTextLength int64 `json:"min_review_text_length"`

	// MaxReviewTextLength is the maximum review text length
	MaxReviewTextLength int64 `json:"max_review_text_length"`

	// ReviewCooldownSeconds is the cooldown between reviews from same reviewer to same provider
	ReviewCooldownSeconds int64 `json:"review_cooldown_seconds"`

	// MaxReviewsPerProvider is the maximum reviews to retain per provider
	MaxReviewsPerProvider int64 `json:"max_reviews_per_provider"`

	// RequireCompletedOrder indicates if reviews must be linked to completed orders
	RequireCompletedOrder bool `json:"require_completed_order"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinReviewTextLength:   DefaultMinReviewTextLength,
		MaxReviewTextLength:   DefaultMaxReviewTextLength,
		ReviewCooldownSeconds: DefaultReviewCooldownSeconds,
		MaxReviewsPerProvider: DefaultMaxReviewsPerProvider,
		RequireCompletedOrder: true, // Reviews must be linked to verified completed orders
	}
}

// Validate validates the parameters
func (p *Params) Validate() error {
	if p.MinReviewTextLength <= 0 {
		return fmt.Errorf("min_review_text_length must be positive")
	}

	if p.MaxReviewTextLength <= 0 {
		return fmt.Errorf("max_review_text_length must be positive")
	}

	if p.MinReviewTextLength > p.MaxReviewTextLength {
		return fmt.Errorf("min_review_text_length cannot exceed max_review_text_length")
	}

	if p.ReviewCooldownSeconds < 0 {
		return fmt.Errorf("review_cooldown_seconds cannot be negative")
	}

	if p.MaxReviewsPerProvider <= 0 {
		return fmt.Errorf("max_reviews_per_provider must be positive")
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
	if err := gs.Params.Validate(); err != nil {
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
