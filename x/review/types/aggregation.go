// Package types contains types for the Review module.
//
// VE-911: Provider public reviews - Rating aggregation types
// This file defines types for aggregating provider ratings and computing averages.
package types

import (
	"fmt"
	"time"
)

// RatingDistribution tracks the count of reviews for each rating level
type RatingDistribution struct {
	// OneStar is the count of 1-star reviews
	OneStar uint64 `json:"one_star"`

	// TwoStar is the count of 2-star reviews
	TwoStar uint64 `json:"two_star"`

	// ThreeStar is the count of 3-star reviews
	ThreeStar uint64 `json:"three_star"`

	// FourStar is the count of 4-star reviews
	FourStar uint64 `json:"four_star"`

	// FiveStar is the count of 5-star reviews
	FiveStar uint64 `json:"five_star"`
}

// Total returns the total number of reviews
func (d *RatingDistribution) Total() uint64 {
	return d.OneStar + d.TwoStar + d.ThreeStar + d.FourStar + d.FiveStar
}

// WeightedSum returns the weighted sum of all ratings
func (d *RatingDistribution) WeightedSum() uint64 {
	return d.OneStar*1 + d.TwoStar*2 + d.ThreeStar*3 + d.FourStar*4 + d.FiveStar*5
}

// Add adds a rating to the distribution
func (d *RatingDistribution) Add(rating uint8) error {
	switch rating {
	case 1:
		d.OneStar++
	case 2:
		d.TwoStar++
	case 3:
		d.ThreeStar++
	case 4:
		d.FourStar++
	case 5:
		d.FiveStar++
	default:
		return fmt.Errorf("invalid rating: %d", rating)
	}
	return nil
}

// Remove removes a rating from the distribution
func (d *RatingDistribution) Remove(rating uint8) error {
	switch rating {
	case 1:
		if d.OneStar == 0 {
			return fmt.Errorf("cannot remove: no 1-star ratings")
		}
		d.OneStar--
	case 2:
		if d.TwoStar == 0 {
			return fmt.Errorf("cannot remove: no 2-star ratings")
		}
		d.TwoStar--
	case 3:
		if d.ThreeStar == 0 {
			return fmt.Errorf("cannot remove: no 3-star ratings")
		}
		d.ThreeStar--
	case 4:
		if d.FourStar == 0 {
			return fmt.Errorf("cannot remove: no 4-star ratings")
		}
		d.FourStar--
	case 5:
		if d.FiveStar == 0 {
			return fmt.Errorf("cannot remove: no 5-star ratings")
		}
		d.FiveStar--
	default:
		return fmt.Errorf("invalid rating: %d", rating)
	}
	return nil
}

// ProviderAggregation contains aggregated rating data for a provider
type ProviderAggregation struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// TotalReviews is the total number of reviews
	TotalReviews uint64 `json:"total_reviews"`

	// AverageRating is the average rating (fixed-point: value * 100 for precision)
	// e.g., 450 represents 4.50 stars
	AverageRating uint64 `json:"average_rating"`

	// Distribution is the breakdown of ratings by star level
	Distribution RatingDistribution `json:"distribution"`

	// LastReviewAt is when the last review was submitted
	LastReviewAt time.Time `json:"last_review_at"`

	// UpdatedAt is when the aggregation was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// BlockHeight is the block height when last updated
	BlockHeight int64 `json:"block_height"`
}

// NewProviderAggregation creates a new provider aggregation
// BUGFIX-001: now parameter is passed in to ensure consensus safety (use ctx.BlockTime())
func NewProviderAggregation(providerAddress string) *ProviderAggregation {
	return &ProviderAggregation{
		ProviderAddress: providerAddress,
		TotalReviews:    0,
		AverageRating:   0,
		Distribution:    RatingDistribution{},
		LastReviewAt:    time.Time{},
		UpdatedAt:       time.Time{}, // Set by caller using blockchain time
	}
}

// Validate validates the provider aggregation
func (pa *ProviderAggregation) Validate() error {
	if pa.ProviderAddress == "" {
		return ErrInvalidAddress.Wrap("provider address is required")
	}

	// Verify total matches distribution
	distributionTotal := pa.Distribution.Total()
	if pa.TotalReviews != distributionTotal {
		return fmt.Errorf("total mismatch: total_reviews=%d, distribution_total=%d",
			pa.TotalReviews, distributionTotal)
	}

	return nil
}

// AddReview adds a review to the aggregation and recalculates the average
// BUGFIX-001: reviewTime is used for both LastReviewAt and UpdatedAt for consensus safety
func (pa *ProviderAggregation) AddReview(rating uint8, reviewTime time.Time) error {
	if err := pa.Distribution.Add(rating); err != nil {
		return err
	}

	pa.TotalReviews++
	pa.recalculateAverage()
	pa.LastReviewAt = reviewTime
	pa.UpdatedAt = reviewTime // Use provided time for consensus safety

	return nil
}

// RemoveReview removes a review from the aggregation and recalculates the average
// BUGFIX-001: updatedAt parameter ensures consensus safety (use ctx.BlockTime())
func (pa *ProviderAggregation) RemoveReview(rating uint8, updatedAt time.Time) error {
	if pa.TotalReviews == 0 {
		return fmt.Errorf("cannot remove: no reviews exist")
	}

	if err := pa.Distribution.Remove(rating); err != nil {
		return err
	}

	pa.TotalReviews--
	pa.recalculateAverage()
	pa.UpdatedAt = updatedAt

	return nil
}

// recalculateAverage recalculates the average rating
func (pa *ProviderAggregation) recalculateAverage() {
	if pa.TotalReviews == 0 {
		pa.AverageRating = 0
		return
	}

	// Calculate average with 2 decimal precision (multiply by 100)
	weightedSum := pa.Distribution.WeightedSum()
	pa.AverageRating = (weightedSum * 100) / pa.TotalReviews
}

// GetAverageRatingFloat returns the average rating as a float64
func (pa *ProviderAggregation) GetAverageRatingFloat() float64 {
	return float64(pa.AverageRating) / 100.0
}

// GetAverageRatingDisplay returns a display string for the average rating
func (pa *ProviderAggregation) GetAverageRatingDisplay() string {
	return fmt.Sprintf("%.2f", pa.GetAverageRatingFloat())
}

// ProviderAggregations is a slice of ProviderAggregation
type ProviderAggregations []ProviderAggregation

// ByTotalReviews returns aggregations sorted by total reviews (descending)
// Note: This returns a copy, does not modify the original
func (aggs ProviderAggregations) TopByReviewCount(limit int) ProviderAggregations {
	if limit <= 0 || limit > len(aggs) {
		limit = len(aggs)
	}

	// Simple selection of top N by review count
	result := make(ProviderAggregations, 0, limit)
	used := make(map[int]bool)

	for i := 0; i < limit; i++ {
		maxIdx := -1
		var maxCount uint64 = 0

		for j, agg := range aggs {
			if !used[j] && agg.TotalReviews > maxCount {
				maxCount = agg.TotalReviews
				maxIdx = j
			}
		}

		if maxIdx >= 0 {
			result = append(result, aggs[maxIdx])
			used[maxIdx] = true
		}
	}

	return result
}

// TopByRating returns aggregations sorted by average rating (descending)
func (aggs ProviderAggregations) TopByRating(limit int) ProviderAggregations {
	if limit <= 0 || limit > len(aggs) {
		limit = len(aggs)
	}

	result := make(ProviderAggregations, 0, limit)
	used := make(map[int]bool)

	for i := 0; i < limit; i++ {
		maxIdx := -1
		var maxRating uint64 = 0

		for j, agg := range aggs {
			if !used[j] && agg.AverageRating > maxRating {
				maxRating = agg.AverageRating
				maxIdx = j
			}
		}

		if maxIdx >= 0 {
			result = append(result, aggs[maxIdx])
			used[maxIdx] = true
		}
	}

	return result
}
