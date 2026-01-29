package review_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/review"
	"github.com/virtengine/virtengine/x/review/types"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

// Test: DefaultGenesisState returns valid state
func (s *GenesisTestSuite) TestDefaultGenesisState() {
	genesis := types.DefaultGenesisState()

	s.Require().NotNil(genesis)
	s.Require().NotNil(genesis.Params)
	s.Require().Empty(genesis.Reviews)
	s.Require().Empty(genesis.Aggregations)
	s.Require().Equal(uint64(1), genesis.NextReviewSequence)
}

// Test: ValidateGenesis with default state
func (s *GenesisTestSuite) TestValidateGenesis_Default() {
	genesis := types.DefaultGenesisState()
	err := genesis.Validate()
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid reviews
func (s *GenesisTestSuite) TestValidateGenesis_ValidReviews() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID:    "review-1",
				OrderID:     "order-1",
				Reviewer:    "cosmos1reviewer",
				Provider:    "cosmos1provider",
				Rating:      5,
				Comment:     "Excellent service!",
				CreatedAt:   now,
			},
		},
		NextReviewSequence: 2,
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}

// Test: ValidateGenesis with valid aggregations
func (s *GenesisTestSuite) TestValidateGenesis_ValidAggregations() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Aggregations: []types.ProviderAggregation{
			{
				Provider:       "cosmos1provider",
				TotalReviews:   100,
				AverageRating:  4.5,
				TotalRating:    450,
				FiveStarCount:  60,
				FourStarCount:  25,
				ThreeStarCount: 10,
				TwoStarCount:   3,
				OneStarCount:   2,
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}

// Test: ValidateGenesis with invalid review - empty ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidReview_EmptyID() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID: "", // Invalid
				OrderID:  "order-1",
				Reviewer: "cosmos1reviewer",
				Provider: "cosmos1provider",
				Rating:   5,
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid review - empty order ID
func (s *GenesisTestSuite) TestValidateGenesis_InvalidReview_EmptyOrderID() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID: "review-1",
				OrderID:  "", // Invalid
				Reviewer: "cosmos1reviewer",
				Provider: "cosmos1provider",
				Rating:   5,
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid review - rating out of range
func (s *GenesisTestSuite) TestValidateGenesis_InvalidReview_RatingOutOfRange() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID: "review-1",
				OrderID:  "order-1",
				Reviewer: "cosmos1reviewer",
				Provider: "cosmos1provider",
				Rating:   6, // Invalid: max is 5
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid review - zero rating
func (s *GenesisTestSuite) TestValidateGenesis_InvalidReview_ZeroRating() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID: "review-1",
				OrderID:  "order-1",
				Reviewer: "cosmos1reviewer",
				Provider: "cosmos1provider",
				Rating:   0, // Invalid: min is 1
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: ValidateGenesis with duplicate review IDs
func (s *GenesisTestSuite) TestValidateGenesis_DuplicateReviews() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID:  "review-1",
				OrderID:   "order-1",
				Reviewer:  "cosmos1reviewer1",
				Provider:  "cosmos1provider",
				Rating:    5,
				CreatedAt: now,
			},
			{
				ReviewID:  "review-1", // Duplicate
				OrderID:   "order-2",
				Reviewer:  "cosmos1reviewer2",
				Provider:  "cosmos1provider",
				Rating:    4,
				CreatedAt: now,
			},
		},
		NextReviewSequence: 2,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid aggregation - empty provider
func (s *GenesisTestSuite) TestValidateGenesis_InvalidAggregation_EmptyProvider() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Aggregations: []types.ProviderAggregation{
			{
				Provider:     "", // Invalid
				TotalReviews: 10,
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: ValidateGenesis with invalid aggregation - average rating out of range
func (s *GenesisTestSuite) TestValidateGenesis_InvalidAggregation_AvgRatingOutOfRange() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Aggregations: []types.ProviderAggregation{
			{
				Provider:      "cosmos1provider",
				TotalReviews:  10,
				AverageRating: 5.5, // Invalid: max is 5.0
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().Error(err)
}

// Test: DefaultParams
func (s *GenesisTestSuite) TestDefaultParams() {
	params := types.DefaultParams()

	s.Require().NotNil(params)
	s.Require().Greater(params.MinRating, uint32(0))
	s.Require().LessOrEqual(params.MinRating, params.MaxRating)
	s.Require().Greater(params.MaxCommentLength, uint32(0))
}

// Test: Params validation - valid
func (s *GenesisTestSuite) TestParamsValidation_Valid() {
	params := types.DefaultParams()
	err := params.Validate()
	s.Require().NoError(err)
}

// Test: Params validation - min rating greater than max
func (s *GenesisTestSuite) TestParamsValidation_MinGreaterThanMax() {
	params := types.DefaultParams()
	params.MinRating = 5
	params.MaxRating = 1

	err := params.Validate()
	s.Require().Error(err)
}

// Test: Params validation - zero max comment length
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxCommentLength() {
	params := types.DefaultParams()
	params.MaxCommentLength = 0

	err := params.Validate()
	s.Require().Error(err)
}

// Table-driven tests for review validation
func TestReviewValidationTable(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name        string
		review      types.Review
		expectError bool
	}{
		{
			name: "valid review",
			review: types.Review{
				ReviewID:  "review-1",
				OrderID:   "order-1",
				Reviewer:  "cosmos1reviewer",
				Provider:  "cosmos1provider",
				Rating:    5,
				Comment:   "Great service!",
				CreatedAt: now,
			},
			expectError: false,
		},
		{
			name: "empty review ID",
			review: types.Review{
				ReviewID: "",
				OrderID:  "order-1",
				Reviewer: "cosmos1reviewer",
				Provider: "cosmos1provider",
				Rating:   5,
			},
			expectError: true,
		},
		{
			name: "empty reviewer",
			review: types.Review{
				ReviewID: "review-1",
				OrderID:  "order-1",
				Reviewer: "",
				Provider: "cosmos1provider",
				Rating:   5,
			},
			expectError: true,
		},
		{
			name: "empty provider",
			review: types.Review{
				ReviewID: "review-1",
				OrderID:  "order-1",
				Reviewer: "cosmos1reviewer",
				Provider: "",
				Rating:   5,
			},
			expectError: true,
		},
		{
			name: "rating 1 (min valid)",
			review: types.Review{
				ReviewID:  "review-1",
				OrderID:   "order-1",
				Reviewer:  "cosmos1reviewer",
				Provider:  "cosmos1provider",
				Rating:    1,
				CreatedAt: now,
			},
			expectError: false,
		},
		{
			name: "rating 5 (max valid)",
			review: types.Review{
				ReviewID:  "review-1",
				OrderID:   "order-1",
				Reviewer:  "cosmos1reviewer",
				Provider:  "cosmos1provider",
				Rating:    5,
				CreatedAt: now,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.review.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Table-driven tests for aggregation validation
func TestAggregationValidationTable(t *testing.T) {
	tests := []struct {
		name        string
		agg         types.ProviderAggregation
		expectError bool
	}{
		{
			name: "valid aggregation",
			agg: types.ProviderAggregation{
				Provider:      "cosmos1provider",
				TotalReviews:  100,
				AverageRating: 4.5,
				TotalRating:   450,
			},
			expectError: false,
		},
		{
			name: "empty provider",
			agg: types.ProviderAggregation{
				Provider:     "",
				TotalReviews: 10,
			},
			expectError: true,
		},
		{
			name: "negative average rating",
			agg: types.ProviderAggregation{
				Provider:      "cosmos1provider",
				TotalReviews:  10,
				AverageRating: -1.0,
			},
			expectError: true,
		},
		{
			name: "average rating too high",
			agg: types.ProviderAggregation{
				Provider:      "cosmos1provider",
				TotalReviews:  10,
				AverageRating: 6.0,
			},
			expectError: true,
		},
		{
			name: "star counts exceed total",
			agg: types.ProviderAggregation{
				Provider:       "cosmos1provider",
				TotalReviews:   5,
				FiveStarCount:  10, // More than total
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.agg.Validate()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test: Complete genesis state with all entities
func (s *GenesisTestSuite) TestValidateGenesis_CompleteState() {
	now := time.Now().UTC()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Reviews: []types.Review{
			{
				ReviewID:  "review-1",
				OrderID:   "order-1",
				Reviewer:  "cosmos1reviewer",
				Provider:  "cosmos1provider",
				Rating:    5,
				Comment:   "Excellent!",
				CreatedAt: now,
			},
			{
				ReviewID:  "review-2",
				OrderID:   "order-2",
				Reviewer:  "cosmos1reviewer2",
				Provider:  "cosmos1provider",
				Rating:    4,
				Comment:   "Good service",
				CreatedAt: now,
			},
		},
		Aggregations: []types.ProviderAggregation{
			{
				Provider:       "cosmos1provider",
				TotalReviews:   2,
				AverageRating:  4.5,
				TotalRating:    9,
				FiveStarCount:  1,
				FourStarCount:  1,
			},
		},
		NextReviewSequence: 3,
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}

// Test: Consistency check - star counts should sum to total reviews
func (s *GenesisTestSuite) TestValidateGenesis_AggregationConsistency() {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		Aggregations: []types.ProviderAggregation{
			{
				Provider:       "cosmos1provider",
				TotalReviews:   10,
				AverageRating:  4.0,
				TotalRating:    40,
				FiveStarCount:  3,
				FourStarCount:  4,
				ThreeStarCount: 2,
				TwoStarCount:   1,
				OneStarCount:   0,
				// Sum: 3+4+2+1+0 = 10 = TotalReviews (valid)
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}
