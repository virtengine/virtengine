package review_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

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
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: now},
				Rating:          5,
				Text:            "Excellent service provided!",
				State:           types.ReviewStateActive,
				CreatedAt:       now,
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
				ProviderAddress: "cosmos1provider",
				TotalReviews:    100,
				AverageRating:   450, // 4.50 stars (fixed-point: value * 100)
				Distribution: types.RatingDistribution{
					OneStar:   2,
					TwoStar:   3,
					ThreeStar: 10,
					FourStar:  25,
					FiveStar:  60,
				},
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
				ID:              types.ReviewID{ProviderAddress: "", Sequence: 0}, // Invalid - empty provider and zero sequence
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: time.Now()},
				Rating:          5,
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
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: time.Now()}, // Invalid - empty OrderID
				Rating:          5,
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
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: time.Now()},
				Rating:          6, // Invalid: max is 5
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
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: time.Now()},
				Rating:          0, // Invalid: min is 1
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
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer1",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer1", ProviderAddress: "cosmos1provider", CompletedAt: now},
				Rating:          5,
				CreatedAt:       now,
			},
			{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1}, // Duplicate
				ReviewerAddress: "cosmos1reviewer2",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-2", CustomerAddress: "cosmos1reviewer2", ProviderAddress: "cosmos1provider", CompletedAt: now},
				Rating:          4,
				CreatedAt:       now,
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
				ProviderAddress: "", // Invalid
				TotalReviews:    10,
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
				ProviderAddress: "cosmos1provider",
				TotalReviews:    10,
				AverageRating:   550, // Invalid: max is 500 (5.0 * 100)
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
	s.Require().GreaterOrEqual(params.MaxRating, params.MinRating)
	s.Require().Greater(params.MaxCommentLength, uint64(0))
}

// Test: Params validation - valid
func (s *GenesisTestSuite) TestParamsValidation_Valid() {
	params := types.DefaultParams()
	err := types.ValidateParams(&params)
	s.Require().NoError(err)
}

// Test: Params validation - min rating greater than max
func (s *GenesisTestSuite) TestParamsValidation_MinGreaterThanMax() {
	params := types.DefaultParams()
	params.MinRating = 5
	params.MaxRating = 1

	err := types.ValidateParams(&params)
	s.Require().Error(err)
}

// Test: Params validation - zero max comment length
func (s *GenesisTestSuite) TestParamsValidation_ZeroMaxCommentLength() {
	params := types.DefaultParams()
	params.MaxCommentLength = 0

	err := types.ValidateParams(&params)
	s.Require().Error(err)
}

// Table-driven tests for review validation
func TestReviewValidationTable(t *testing.T) {
	now := time.Now().UTC()
	orderRef := types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: now}
	tests := []struct {
		name        string
		review      types.Review
		expectError bool
	}{
		{
			name: "valid review",
			review: types.Review{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        orderRef,
				Rating:          5,
				Text:            "Great service provided!",
				State:           types.ReviewStateActive,
				CreatedAt:       now,
			},
			expectError: false,
		},
		{
			name: "empty review ID",
			review: types.Review{
				ID:              types.ReviewID{ProviderAddress: "", Sequence: 0}, // Invalid
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        orderRef,
				Rating:          5,
				Text:            "Great service provided!",
				State:           types.ReviewStateActive,
			},
			expectError: true,
		},
		{
			name: "empty reviewer",
			review: types.Review{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "", // Invalid
				ProviderAddress: "cosmos1provider",
				OrderRef:        orderRef,
				Rating:          5,
				Text:            "Great service provided!",
				State:           types.ReviewStateActive,
			},
			expectError: true,
		},
		{
			name: "empty provider",
			review: types.Review{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "", // Invalid
				OrderRef:        orderRef,
				Rating:          5,
				Text:            "Great service provided!",
				State:           types.ReviewStateActive,
			},
			expectError: true,
		},
		{
			name: "rating 1 (min valid)",
			review: types.Review{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        orderRef,
				Rating:          1,
				Text:            "Poor service provided!",
				State:           types.ReviewStateActive,
				CreatedAt:       now,
			},
			expectError: false,
		},
		{
			name: "rating 5 (max valid)",
			review: types.Review{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        orderRef,
				Rating:          5,
				Text:            "Excellent service!",
				State:           types.ReviewStateActive,
				CreatedAt:       now,
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
				ProviderAddress: "cosmos1provider",
				TotalReviews:    100,
				AverageRating:   450, // 4.5 stars (fixed-point: value * 100)
				Distribution: types.RatingDistribution{
					OneStar:   5,
					TwoStar:   10,
					ThreeStar: 15,
					FourStar:  30,
					FiveStar:  40,
				},
			},
			expectError: false,
		},
		{
			name: "empty provider",
			agg: types.ProviderAggregation{
				ProviderAddress: "",
				TotalReviews:    10,
			},
			expectError: true,
		},
		{
			name: "average rating too high",
			agg: types.ProviderAggregation{
				ProviderAddress: "cosmos1provider",
				TotalReviews:    10,
				AverageRating:   600, // 6.0 stars - too high (max is 500)
			},
			expectError: true,
		},
		{
			name: "distribution total mismatch",
			agg: types.ProviderAggregation{
				ProviderAddress: "cosmos1provider",
				TotalReviews:    5,
				Distribution: types.RatingDistribution{
					FiveStar: 10, // More than total
				},
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
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
				ReviewerAddress: "cosmos1reviewer",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-1", CustomerAddress: "cosmos1reviewer", ProviderAddress: "cosmos1provider", CompletedAt: now},
				Rating:          5,
				Text:            "Excellent service provided!",
				State:           types.ReviewStateActive,
				CreatedAt:       now,
			},
			{
				ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 2},
				ReviewerAddress: "cosmos1reviewer2",
				ProviderAddress: "cosmos1provider",
				OrderRef:        types.OrderReference{OrderID: "order-2", CustomerAddress: "cosmos1reviewer2", ProviderAddress: "cosmos1provider", CompletedAt: now},
				Rating:          4,
				Text:            "Good service provided!",
				State:           types.ReviewStateActive,
				CreatedAt:       now,
			},
		},
		Aggregations: []types.ProviderAggregation{
			{
				ProviderAddress: "cosmos1provider",
				TotalReviews:    2,
				AverageRating:   450, // 4.5 stars (fixed-point: value * 100)
				Distribution: types.RatingDistribution{
					FourStar: 1,
					FiveStar: 1,
				},
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
				ProviderAddress: "cosmos1provider",
				TotalReviews:    10,
				AverageRating:   400, // 4.0 stars (fixed-point: value * 100)
				Distribution: types.RatingDistribution{
					OneStar:   0,
					TwoStar:   1,
					ThreeStar: 2,
					FourStar:  4,
					FiveStar:  3,
				},
				// Sum: 0+1+2+4+3 = 10 = TotalReviews (valid)
			},
		},
		NextReviewSequence: 1,
	}

	err := genesis.Validate()
	s.Require().NoError(err)
}
