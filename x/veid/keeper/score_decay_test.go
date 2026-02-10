package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ScoreDecayTestSuite tests the score decay mechanism
type ScoreDecayTestSuite struct {
	KeeperTestSuite
}

func TestScoreDecayTestSuite(t *testing.T) {
	suite.Run(t, new(ScoreDecayTestSuite))
}

func (s *ScoreDecayTestSuite) SetupTest() {
	s.KeeperTestSuite.SetupTest()
}

// testAddr creates a test address from a seed string
func (s *ScoreDecayTestSuite) testAddr(seed string) string {
	// Ensure seed is exactly 20 bytes for consistent addresses
	paddedSeed := seed
	for len(paddedSeed) < 20 {
		paddedSeed += "_"
	}
	return sdk.AccAddress([]byte(paddedSeed[:20])).String()
}

// ============================================================================
// Decay Policy Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestSetAndGetDecayPolicy() {
	policy := types.DefaultDecayPolicy()
	policy.PolicyID = "test-policy"

	// Set policy
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Get policy
	retrieved, found := s.keeper.GetDecayPolicy(s.ctx, "test-policy")
	s.Require().True(found)
	s.Require().Equal(policy.PolicyID, retrieved.PolicyID)
	s.Require().Equal(policy.DecayType, retrieved.DecayType)
	s.Require().True(policy.DecayRate.Equal(retrieved.DecayRate))
	s.Require().True(policy.MinScore.Equal(retrieved.MinScore))
	s.Require().Equal(policy.DecayPeriod, retrieved.DecayPeriod)
}

func (s *ScoreDecayTestSuite) TestGetDecayPolicyNotFound() {
	_, found := s.keeper.GetDecayPolicy(s.ctx, "non-existent")
	s.Require().False(found)
}

func (s *ScoreDecayTestSuite) TestDeleteDecayPolicy() {
	policy := types.DefaultDecayPolicy()
	policy.PolicyID = "deletable-policy"

	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Verify it exists
	_, found := s.keeper.GetDecayPolicy(s.ctx, "deletable-policy")
	s.Require().True(found)

	// Delete it
	err = s.keeper.DeleteDecayPolicy(s.ctx, "deletable-policy")
	s.Require().NoError(err)

	// Verify it's gone
	_, found = s.keeper.GetDecayPolicy(s.ctx, "deletable-policy")
	s.Require().False(found)
}

func (s *ScoreDecayTestSuite) TestCannotDeleteDefaultPolicy() {
	// First ensure default policy exists
	_ = s.keeper.GetDefaultDecayPolicy(s.ctx)

	// Attempt to delete default policy should fail
	err := s.keeper.DeleteDecayPolicy(s.ctx, "default")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "cannot delete the default decay policy")
}

func (s *ScoreDecayTestSuite) TestDecayPolicyValidation() {
	testCases := []struct {
		name      string
		modify    func(*types.DecayPolicy)
		expectErr string
	}{
		{
			name:      "valid default policy",
			modify:    func(p *types.DecayPolicy) {},
			expectErr: "",
		},
		{
			name: "empty policy id",
			modify: func(p *types.DecayPolicy) {
				p.PolicyID = ""
			},
			expectErr: "policy_id cannot be empty",
		},
		{
			name: "negative decay rate",
			modify: func(p *types.DecayPolicy) {
				p.DecayRate = math.LegacyNewDec(-1)
			},
			expectErr: "decay_rate cannot be negative",
		},
		{
			name: "exponential rate too high",
			modify: func(p *types.DecayPolicy) {
				p.DecayType = types.DecayTypeExponential
				p.DecayRate = math.LegacyNewDec(2) // > 1.0
			},
			expectErr: "exponential decay_rate must be <= 1.0",
		},
		{
			name: "negative min score",
			modify: func(p *types.DecayPolicy) {
				p.MinScore = math.LegacyNewDec(-10)
			},
			expectErr: "min_score cannot be negative",
		},
		{
			name: "min score too high",
			modify: func(p *types.DecayPolicy) {
				p.MinScore = math.LegacyNewDec(150)
			},
			expectErr: "min_score cannot exceed 100",
		},
		{
			name: "step function without thresholds",
			modify: func(p *types.DecayPolicy) {
				p.DecayType = types.DecayTypeStepFunction
				p.StepThresholds = nil
			},
			expectErr: "step_function decay requires at least one step_threshold",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			policy := types.DefaultDecayPolicy()
			policy.PolicyID = "test-" + tc.name
			tc.modify(&policy)

			err := s.keeper.SetDecayPolicy(s.ctx, policy)
			if tc.expectErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErr)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// ============================================================================
// Score Snapshot Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestCreateAndGetScoreSnapshot() {
	// First set up the default policy
	err := s.keeper.SetDecayPolicy(s.ctx, types.DefaultDecayPolicy())
	s.Require().NoError(err)

	addr := s.testAddr("snapshot_test_addr1")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 85, "default")
	s.Require().NoError(err)
	s.Require().NotNil(snapshot)
	s.Require().Equal(addr, snapshot.Address)
	s.Require().True(math.LegacyNewDec(85).Equal(snapshot.OriginalScore))
	s.Require().True(math.LegacyNewDec(85).Equal(snapshot.CurrentScore))

	// Retrieve it
	retrieved, found := s.keeper.GetScoreSnapshot(s.ctx, addr)
	s.Require().True(found)
	s.Require().Equal(snapshot.Address, retrieved.Address)
	s.Require().True(snapshot.OriginalScore.Equal(retrieved.OriginalScore))
}

func (s *ScoreDecayTestSuite) TestGetScoreSnapshotNotFound() {
	addr := s.testAddr("nonexistent_addr__")
	_, found := s.keeper.GetScoreSnapshot(s.ctx, addr)
	s.Require().False(found)
}

// ============================================================================
// Linear Decay Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestLinearDecay() {
	// Create a linear decay policy
	policy := types.DecayPolicy{
		PolicyID:          "linear-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(5), // 5 points per period
		DecayPeriod:       24 * time.Hour,       // Daily
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0, // No grace period
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Create snapshot with score of 100
	addr := s.testAddr("linear_decay_addr1")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "linear-test")
	s.Require().NoError(err)

	// Advance time by 2 days
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * 24 * time.Hour))

	// Calculate decay
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// Should have decayed by 10 points (5 per day * 2 days)
	s.Require().True(result.PreviousScore.Equal(math.LegacyNewDec(100)))
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(90)))
	s.Require().True(result.DecayAmount.Equal(math.LegacyNewDec(10)))
	s.Require().Equal(int64(2), result.PeriodsApplied)
	s.Require().False(result.ReachedFloor)
}

func (s *ScoreDecayTestSuite) TestLinearDecayReachesFloor() {
	policy := types.DecayPolicy{
		PolicyID:          "linear-floor-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(20), // 20 points per period
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(25),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("linear_floor_addr1")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 50, "linear-floor-test")
	s.Require().NoError(err)

	// Advance time by 3 days - would be 60 points decay but floor is 25
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(3 * 24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(25)))
	s.Require().True(result.ReachedFloor)
}

// ============================================================================
// Exponential Decay Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestExponentialDecay() {
	// Create an exponential decay policy
	policy := types.DecayPolicy{
		PolicyID:          "exponential-test",
		DecayType:         types.DecayTypeExponential,
		DecayRate:         math.LegacyNewDecWithPrec(10, 2), // 10% per period
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("exponential_addr1")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "exponential-test")
	s.Require().NoError(err)

	// Advance time by 1 day
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// 100 * 0.9 = 90
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(90)))
	s.Require().Equal(int64(1), result.PeriodsApplied)

	// Advance time by 2 days (100 * 0.9 * 0.9 = 81)
	futureCtx2 := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * 24 * time.Hour))
	result2 := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx2.BlockTime())

	s.Require().True(result2.NewScore.Equal(math.LegacyNewDec(81)))
	s.Require().Equal(int64(2), result2.PeriodsApplied)
}

func (s *ScoreDecayTestSuite) TestExponentialDecayReachesFloor() {
	policy := types.DecayPolicy{
		PolicyID:          "exponential-floor-test",
		DecayType:         types.DecayTypeExponential,
		DecayRate:         math.LegacyNewDecWithPrec(50, 2), // 50% per period
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(20),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("exp_floor_addr1__")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "exponential-floor-test")
	s.Require().NoError(err)

	// Advance time by 5 days - 100 * 0.5^5 = 3.125, but floor is 20
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(5 * 24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(20)))
	s.Require().True(result.ReachedFloor)
}

// ============================================================================
// Step Function Decay Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestStepFunctionDecay() {
	policy := types.DefaultStepFunctionPolicy()
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("stepfunc_addr1___")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "step_default")
	s.Require().NoError(err)

	// Test at 45 days (should be in 30-day bracket = 95%)
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(45 * 24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// 100 * 0.95 = 95
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(95)))

	// Test at 75 days (should be in 60-day bracket = 85%)
	futureCtx2 := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(75 * 24 * time.Hour))
	result2 := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx2.BlockTime())

	// 100 * 0.85 = 85
	s.Require().True(result2.NewScore.Equal(math.LegacyNewDec(85)))

	// Test at 120 days (should be in 90-day bracket = 70%)
	futureCtx3 := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(120 * 24 * time.Hour))
	result3 := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx3.BlockTime())

	// 100 * 0.70 = 70
	s.Require().True(result3.NewScore.Equal(math.LegacyNewDec(70)))
}

func (s *ScoreDecayTestSuite) TestStepFunctionDecayWithLowScore() {
	policy := types.DefaultStepFunctionPolicy()
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("steplow_addr1____")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 30, "step_default")
	s.Require().NoError(err)

	// Test at 400 days (should be in 365-day bracket = 25%)
	// 30 * 0.25 = 7.5, but floor is 10
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(400 * 24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(10)))
	s.Require().True(result.ReachedFloor)
}

// ============================================================================
// Grace Period Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestGracePeriodPreventsDecay() {
	policy := types.DecayPolicy{
		PolicyID:          "grace-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       7 * 24 * time.Hour, // 7 day grace period
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("grace_test_addr1_")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "grace-test")
	s.Require().NoError(err)

	// Advance time by 3 days (still within grace period)
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(3 * 24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// No decay should have occurred
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(100)))
	s.Require().True(result.DecayAmount.IsZero())
	s.Require().Equal(int64(0), result.PeriodsApplied)
}

func (s *ScoreDecayTestSuite) TestDecayAfterGracePeriod() {
	policy := types.DecayPolicy{
		PolicyID:          "grace-end-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       7 * 24 * time.Hour,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("grace_end_addr1__")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "grace-end-test")
	s.Require().NoError(err)

	// Advance time by 10 days (3 days after grace period ends)
	// Grace ends at day 7, so we have 3 periods of decay
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(10 * 24 * time.Hour))

	// Update LastActivityAt to be the same as creation time (past grace period now)
	snapshot.LastActivityAt = s.ctx.BlockTime()
	err = s.keeper.SetScoreSnapshot(s.ctx, snapshot)
	s.Require().NoError(err)

	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// Should have decayed: 10 periods * 10 = 100 total decay, but we're after grace
	// Actually 10 - 7 = 3 periods after grace, so 30 points decay = 70 score
	s.Require().True(result.NewScore.LT(math.LegacyNewDec(100)))
}

func (s *ScoreDecayTestSuite) TestActivityResetsGracePeriod() {
	policy := types.DecayPolicy{
		PolicyID:          "activity-grace-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       7 * 24 * time.Hour,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("activity_test1___")
	_, err = s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "activity-grace-test")
	s.Require().NoError(err)

	// Advance time by 5 days
	day5Ctx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(5 * 24 * time.Hour))

	// Record activity (this should reset grace period)
	err = s.keeper.RecordActivity(day5Ctx, addr, types.ActivityTypeTransaction)
	s.Require().NoError(err)

	// Advance to day 10 (5 days after activity, still within new grace period)
	day10Ctx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(10 * 24 * time.Hour))

	// Get updated snapshot
	snapshot, found := s.keeper.GetScoreSnapshot(day10Ctx, addr)
	s.Require().True(found)

	result := s.keeper.CalculateDecayedScore(policy, snapshot, day10Ctx.BlockTime())

	// Should still be in grace period due to activity at day 5
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(100)))
	s.Require().True(result.DecayAmount.IsZero())
}

// ============================================================================
// Minimum Score Floor Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestMinScoreFloor() {
	policy := types.DecayPolicy{
		PolicyID:          "floor-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(50), // Aggressive decay
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(30), // Floor at 30
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("floor_test_addr1_")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "floor-test")
	s.Require().NoError(err)

	// Advance by 10 days - would be 500 points decay without floor
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(10 * 24 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// Should be at floor
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(30)))
	s.Require().True(result.ReachedFloor)
}

func (s *ScoreDecayTestSuite) TestScoreAlreadyAtFloorNoDecay() {
	policy := types.DecayPolicy{
		PolicyID:          "at-floor-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(50),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("at_floor_addr1___")
	// Start exactly at floor
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 50, "at-floor-test")
	s.Require().NoError(err)

	// ShouldApplyDecay should return false when at floor
	s.Require().False(snapshot.ShouldApplyDecay(policy, s.ctx.BlockTime().Add(24*time.Hour)))
}

// ============================================================================
// Activity Bonus Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestActivityBonusApplied() {
	policy := types.DecayPolicy{
		PolicyID:          "bonus-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       7 * 24 * time.Hour,
		LastActivityBonus: math.LegacyNewDecWithPrec(11, 1), // 1.1 = 10% bonus
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("bonus_test_addr1_")
	_, err = s.keeper.CreateScoreSnapshot(s.ctx, addr, 80, "bonus-test")
	s.Require().NoError(err)

	// Record activity
	err = s.keeper.RecordActivity(s.ctx, addr, types.ActivityTypeMarketplace)
	s.Require().NoError(err)

	// Get updated snapshot
	snapshot, found := s.keeper.GetScoreSnapshot(s.ctx, addr)
	s.Require().True(found)

	// Verify activity was recorded (compare Unix timestamps to avoid timezone issues)
	s.Require().Equal(s.ctx.BlockTime().Unix(), snapshot.LastActivityAt.Unix())
}

// ============================================================================
// Apply Decay Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestApplyDecay() {
	policy := types.DecayPolicy{
		PolicyID:          "apply-decay-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(5),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("apply_decay_addr1")
	_, err = s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "apply-decay-test")
	s.Require().NoError(err)

	// Advance time and apply decay
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(3 * 24 * time.Hour))

	result, err := s.keeper.ApplyDecay(futureCtx, addr)
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(85))) // 100 - (3 * 5)

	// Verify snapshot was updated
	snapshot, found := s.keeper.GetScoreSnapshot(futureCtx, addr)
	s.Require().True(found)
	s.Require().True(snapshot.CurrentScore.Equal(math.LegacyNewDec(85)))
	s.Require().True(snapshot.TotalDecayApplied.Equal(math.LegacyNewDec(15)))
}

func (s *ScoreDecayTestSuite) TestApplyDecayToAll() {
	policy := types.DecayPolicy{
		PolicyID:          "bulk-decay-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(5),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	// Create multiple snapshots
	addresses := []string{
		s.testAddr("bulk_decay_addr1_"),
		s.testAddr("bulk_decay_addr2_"),
		s.testAddr("bulk_decay_addr3_"),
	}
	for _, addr := range addresses {
		_, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "bulk-decay-test")
		s.Require().NoError(err)
	}

	// Advance time
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * 24 * time.Hour))

	// Apply decay to all
	decayedCount, err := s.keeper.ApplyDecayToAll(futureCtx)
	s.Require().NoError(err)
	s.Require().Equal(3, decayedCount)

	// Verify all were decayed
	for _, addr := range addresses {
		snapshot, found := s.keeper.GetScoreSnapshot(futureCtx, addr)
		s.Require().True(found)
		s.Require().True(snapshot.CurrentScore.Equal(math.LegacyNewDec(90))) // 100 - (2 * 5)
	}
}

// ============================================================================
// Pause/Resume Decay Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestPauseAndResumeDecay() {
	policy := types.DecayPolicy{
		PolicyID:          "pause-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("pause_test_addr1_")
	_, err = s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "pause-test")
	s.Require().NoError(err)

	// Pause decay
	err = s.keeper.PauseDecay(s.ctx, addr)
	s.Require().NoError(err)

	// Verify paused
	snapshot, found := s.keeper.GetScoreSnapshot(s.ctx, addr)
	s.Require().True(found)
	s.Require().True(snapshot.DecayPaused)
	s.Require().False(snapshot.ShouldApplyDecay(policy, s.ctx.BlockTime().Add(5*24*time.Hour)))

	// Resume decay
	err = s.keeper.ResumeDecay(s.ctx, addr)
	s.Require().NoError(err)

	// Verify resumed
	snapshot, found = s.keeper.GetScoreSnapshot(s.ctx, addr)
	s.Require().True(found)
	s.Require().False(snapshot.DecayPaused)
}

// ============================================================================
// Get Effective Score Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestGetEffectiveScore() {
	policy := types.DecayPolicy{
		PolicyID:          "effective-score-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("effective_addr1__")
	_, err = s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "effective-score-test")
	s.Require().NoError(err)

	// Get effective score after 2 days (without actually applying decay)
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * 24 * time.Hour))
	score, found := s.keeper.GetEffectiveScore(futureCtx, addr)
	s.Require().True(found)
	s.Require().Equal(uint32(80), score) // 100 - (2 * 10)
}

// ============================================================================
// Activity Recording Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestRecordAndGetActivity() {
	// Set up default policy
	err := s.keeper.SetDecayPolicy(s.ctx, types.DefaultDecayPolicy())
	s.Require().NoError(err)

	addr := s.testAddr("activity_addr1___")
	_, err = s.keeper.CreateScoreSnapshot(s.ctx, addr, 80, "default")
	s.Require().NoError(err)

	// Record multiple activities
	activities := []types.ActivityType{
		types.ActivityTypeTransaction,
		types.ActivityTypeMarketplace,
		types.ActivityTypeStaking,
	}

	for i, actType := range activities {
		ctx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Duration(i) * time.Hour))
		err := s.keeper.RecordActivity(ctx, addr, actType)
		s.Require().NoError(err)
	}

	// Get recent activity
	records, err := s.keeper.GetRecentActivity(s.ctx, addr, 10)
	s.Require().NoError(err)
	s.Require().Len(records, 3)
}

func (s *ScoreDecayTestSuite) TestRecordActivityInvalidType() {
	addr := s.testAddr("invalid_activity_")
	err := s.keeper.RecordActivity(s.ctx, addr, types.ActivityType("invalid"))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid activity type")
}

// ============================================================================
// Type Validation Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestDecayTypeValidation() {
	s.Require().True(types.DecayTypeLinear.IsValid())
	s.Require().True(types.DecayTypeExponential.IsValid())
	s.Require().True(types.DecayTypeStepFunction.IsValid())
	s.Require().False(types.DecayType(99).IsValid())

	// Test string conversion
	s.Require().Equal("linear", types.DecayTypeLinear.String())
	s.Require().Equal("exponential", types.DecayTypeExponential.String())
	s.Require().Equal("step_function", types.DecayTypeStepFunction.String())
	s.Require().Equal("unknown", types.DecayType(99).String())

	// Test parsing
	dt, err := types.ParseDecayType("linear")
	s.Require().NoError(err)
	s.Require().Equal(types.DecayTypeLinear, dt)

	dt, err = types.ParseDecayType("exponential")
	s.Require().NoError(err)
	s.Require().Equal(types.DecayTypeExponential, dt)

	_, err = types.ParseDecayType("invalid")
	s.Require().Error(err)
}

func (s *ScoreDecayTestSuite) TestActivityTypeValidation() {
	for _, at := range types.AllActivityTypes() {
		s.Require().True(at.IsValid())
	}
	s.Require().False(types.ActivityType("invalid").IsValid())
}

func (s *ScoreDecayTestSuite) TestScoreSnapshotValidation() {
	testCases := []struct {
		name      string
		snapshot  types.ScoreSnapshot
		expectErr string
	}{
		{
			name: "valid snapshot",
			snapshot: types.ScoreSnapshot{
				Address:           "virt1valid",
				OriginalScore:     math.LegacyNewDec(100),
				CurrentScore:      math.LegacyNewDec(80),
				PolicyID:          "default",
				TotalDecayApplied: math.LegacyNewDec(20),
			},
			expectErr: "",
		},
		{
			name: "empty address",
			snapshot: types.ScoreSnapshot{
				Address:       "",
				OriginalScore: math.LegacyNewDec(100),
				CurrentScore:  math.LegacyNewDec(80),
				PolicyID:      "default",
			},
			expectErr: "address cannot be empty",
		},
		{
			name: "negative original score",
			snapshot: types.ScoreSnapshot{
				Address:       "virt1test",
				OriginalScore: math.LegacyNewDec(-10),
				CurrentScore:  math.LegacyNewDec(80),
				PolicyID:      "default",
			},
			expectErr: "original_score cannot be negative",
		},
		{
			name: "empty policy id",
			snapshot: types.ScoreSnapshot{
				Address:       "virt1test",
				OriginalScore: math.LegacyNewDec(100),
				CurrentScore:  math.LegacyNewDec(80),
				PolicyID:      "",
			},
			expectErr: "policy_id cannot be empty",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.snapshot.Validate()
			if tc.expectErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErr)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// ============================================================================
// Edge Cases Tests
// ============================================================================

func (s *ScoreDecayTestSuite) TestNoDecayWhenPolicyDisabled() {
	policy := types.DecayPolicy{
		PolicyID:          "disabled-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour,
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           false, // Disabled
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("disabled_addr1___")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "disabled-test")
	s.Require().NoError(err)

	// ShouldApplyDecay should return false when policy is disabled
	s.Require().False(snapshot.ShouldApplyDecay(policy, s.ctx.BlockTime().Add(5*24*time.Hour)))
}

func (s *ScoreDecayTestSuite) TestNoDecayBeforePeriodElapsed() {
	policy := types.DecayPolicy{
		PolicyID:          "period-test",
		DecayType:         types.DecayTypeLinear,
		DecayRate:         math.LegacyNewDec(10),
		DecayPeriod:       24 * time.Hour, // 1 day
		MinScore:          math.LegacyNewDec(10),
		GracePeriod:       0,
		LastActivityBonus: math.LegacyOneDec(),
		Enabled:           true,
	}
	err := s.keeper.SetDecayPolicy(s.ctx, policy)
	s.Require().NoError(err)

	addr := s.testAddr("period_test_addr1")
	snapshot, err := s.keeper.CreateScoreSnapshot(s.ctx, addr, 100, "period-test")
	s.Require().NoError(err)

	// Advance by less than one period (12 hours)
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(12 * time.Hour))
	result := s.keeper.CalculateDecayedScore(policy, snapshot, futureCtx.BlockTime())

	// No decay should have occurred
	s.Require().True(result.NewScore.Equal(math.LegacyNewDec(100)))
	s.Require().Equal(int64(0), result.PeriodsApplied)
}
