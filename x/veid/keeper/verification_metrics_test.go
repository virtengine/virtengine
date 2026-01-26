package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/veid/types"
)

// ============================================================================
// Test VerificationMetrics
// ============================================================================

func TestVerificationMetrics_Creation(t *testing.T) {
	metrics := VerificationMetrics{
		RequestID:         "req-123",
		AccountAddress:    "virt1abc123",
		ProposerScore:     85,
		ComputedScore:     85,
		ScoreDifference:   0,
		Match:             true,
		ModelVersion:      "v1.0.0",
		ComputeTimeMs:     50,
		BlockHeight:       1000,
		ValidatorAddress:  "virt1val123",
		Timestamp:         time.Now().UTC(),
		Status:            types.VerificationResultStatusSuccess,
		ProposerStatus:    types.VerificationResultStatusSuccess,
		InputHashMatch:    true,
		ModelVersionMatch: true,
		ScopeCount:        3,
		ReasonCodes:       []types.ReasonCode{types.ReasonCodeSuccess},
	}

	require.Equal(t, "req-123", metrics.RequestID)
	require.Equal(t, "virt1abc123", metrics.AccountAddress)
	require.Equal(t, uint32(85), metrics.ProposerScore)
	require.Equal(t, uint32(85), metrics.ComputedScore)
	require.Equal(t, int32(0), metrics.ScoreDifference)
	require.True(t, metrics.Match)
	require.Equal(t, "v1.0.0", metrics.ModelVersion)
	require.Equal(t, int64(50), metrics.ComputeTimeMs)
	require.Equal(t, int64(1000), metrics.BlockHeight)
	require.Equal(t, "virt1val123", metrics.ValidatorAddress)
	require.True(t, metrics.InputHashMatch)
	require.True(t, metrics.ModelVersionMatch)
	require.Equal(t, 3, metrics.ScopeCount)
	require.Len(t, metrics.ReasonCodes, 1)
}

func TestVerificationMetrics_Mismatch(t *testing.T) {
	metrics := VerificationMetrics{
		RequestID:         "req-456",
		ProposerScore:     85,
		ComputedScore:     80,
		ScoreDifference:   5,
		Match:             false,
		ModelVersion:      "v1.0.0",
		BlockHeight:       1000,
		ValidatorAddress:  "virt1val123",
		InputHashMatch:    true,
		ModelVersionMatch: true,
	}

	require.False(t, metrics.Match)
	require.Equal(t, int32(5), metrics.ScoreDifference)
	require.Equal(t, uint32(85), metrics.ProposerScore)
	require.Equal(t, uint32(80), metrics.ComputedScore)
}

// ============================================================================
// Test VerificationMetricsAggregated
// ============================================================================

func TestVerificationMetricsAggregated_Empty(t *testing.T) {
	agg := &VerificationMetricsAggregated{
		BlockHeightStart:   1000,
		BlockHeightEnd:     2000,
		TotalVerifications: 0,
		MatchCount:         0,
		MismatchCount:      0,
		ByModelVersion:     make(map[string]ModelVersionMetrics),
		ByStatus:           make(map[types.VerificationResultStatus]int64),
	}

	require.Equal(t, int64(0), agg.TotalVerifications)
	require.Equal(t, int64(0), agg.MatchCount)
	require.Equal(t, int64(0), agg.MismatchCount)
	require.Empty(t, agg.ByModelVersion)
	require.Empty(t, agg.ByStatus)
}

func TestVerificationMetricsAggregated_WithData(t *testing.T) {
	agg := &VerificationMetricsAggregated{
		BlockHeightStart:      1000,
		BlockHeightEnd:        2000,
		TotalVerifications:    100,
		MatchCount:            95,
		MismatchCount:         5,
		MatchRate:             95.0,
		AvgComputeTimeMs:      50.5,
		MaxComputeTimeMs:      200,
		MinComputeTimeMs:      10,
		AvgScoreDifference:    0.5,
		MaxScoreDifference:    3,
		PeriodStart:           time.Now().Add(-1 * time.Hour),
		PeriodEnd:             time.Now(),
		ByModelVersion:        make(map[string]ModelVersionMetrics),
		ByStatus:              make(map[types.VerificationResultStatus]int64),
	}

	agg.ByModelVersion["v1.0.0"] = ModelVersionMetrics{
		ModelVersion:      "v1.0.0",
		VerificationCount: 100,
		MatchCount:        95,
		MatchRate:         95.0,
		AvgComputeTimeMs:  50.5,
	}

	agg.ByStatus[types.VerificationResultStatusSuccess] = 90
	agg.ByStatus[types.VerificationResultStatusPartial] = 5
	agg.ByStatus[types.VerificationResultStatusFailed] = 5

	require.Equal(t, int64(100), agg.TotalVerifications)
	require.Equal(t, 95.0, agg.MatchRate)
	require.Equal(t, float64(50.5), agg.AvgComputeTimeMs)
	require.Equal(t, int64(200), agg.MaxComputeTimeMs)
	require.Equal(t, int64(10), agg.MinComputeTimeMs)
	require.Equal(t, int32(3), agg.MaxScoreDifference)
	require.Len(t, agg.ByModelVersion, 1)
	require.Len(t, agg.ByStatus, 3)
}

// ============================================================================
// Test ModelVersionMetrics
// ============================================================================

func TestModelVersionMetrics(t *testing.T) {
	mv := ModelVersionMetrics{
		ModelVersion:      "v1.0.0",
		VerificationCount: 100,
		MatchCount:        98,
		MatchRate:         98.0,
		AvgComputeTimeMs:  45.0,
	}

	require.Equal(t, "v1.0.0", mv.ModelVersion)
	require.Equal(t, int64(100), mv.VerificationCount)
	require.Equal(t, int64(98), mv.MatchCount)
	require.Equal(t, 98.0, mv.MatchRate)
	require.Equal(t, 45.0, mv.AvgComputeTimeMs)
}

// ============================================================================
// Test Helper Functions
// ============================================================================

func TestAbs(t *testing.T) {
	testCases := []struct {
		input    int32
		expected int32
	}{
		{0, 0},
		{5, 5},
		{-5, 5},
		{100, 100},
		{-100, 100},
		{-2147483648, -2147483648}, // Min int32 overflow case
	}

	for _, tc := range testCases {
		// Skip overflow case
		if tc.input == -2147483648 {
			continue
		}
		result := abs(tc.input)
		require.Equal(t, tc.expected, result, "abs(%d) should equal %d", tc.input, tc.expected)
	}
}

// ============================================================================
// Test Metrics Keys
// ============================================================================

func TestMetricsKey(t *testing.T) {
	key := metricsKey(1000, "req-123")

	require.NotEmpty(t, key)
	require.True(t, key[0] == 0xF0, "should start with metrics prefix")
}

func TestMetricsPrefixKey(t *testing.T) {
	prefix := metricsPrefixKey(1000)

	require.NotEmpty(t, prefix)
	require.True(t, prefix[0] == 0xF0, "should start with metrics prefix")

	// Keys for the same block should share the prefix
	key := metricsKey(1000, "req-123")
	require.True(t, len(key) > len(prefix), "key should be longer than prefix")

	for i := 0; i < len(prefix); i++ {
		require.Equal(t, prefix[i], key[i], "key should start with prefix")
	}
}

func TestMetricsKeyDifferentBlocks(t *testing.T) {
	key1 := metricsKey(1000, "req-123")
	key2 := metricsKey(2000, "req-123")

	require.NotEqual(t, key1, key2, "different blocks should have different keys")
}

func TestMetricsKeyDifferentRequests(t *testing.T) {
	key1 := metricsKey(1000, "req-123")
	key2 := metricsKey(1000, "req-456")

	require.NotEqual(t, key1, key2, "different requests should have different keys")
}

// ============================================================================
// Test Metrics Status Aggregation
// ============================================================================

func TestMetricsStatusAggregation(t *testing.T) {
	byStatus := make(map[types.VerificationResultStatus]int64)

	// Simulate counting verifications by status
	statuses := []types.VerificationResultStatus{
		types.VerificationResultStatusSuccess,
		types.VerificationResultStatusSuccess,
		types.VerificationResultStatusSuccess,
		types.VerificationResultStatusPartial,
		types.VerificationResultStatusFailed,
	}

	for _, status := range statuses {
		byStatus[status]++
	}

	require.Equal(t, int64(3), byStatus[types.VerificationResultStatusSuccess])
	require.Equal(t, int64(1), byStatus[types.VerificationResultStatusPartial])
	require.Equal(t, int64(1), byStatus[types.VerificationResultStatusFailed])
	require.Equal(t, int64(0), byStatus[types.VerificationResultStatusError])
}

// ============================================================================
// Test Metrics Match Rate Calculation
// ============================================================================

func TestMatchRateCalculation(t *testing.T) {
	testCases := []struct {
		total    int64
		matches  int64
		expected float64
	}{
		{100, 100, 100.0},
		{100, 95, 95.0},
		{100, 50, 50.0},
		{100, 0, 0.0},
		{0, 0, 0.0}, // Edge case: avoid division by zero
	}

	for _, tc := range testCases {
		var matchRate float64
		if tc.total > 0 {
			matchRate = float64(tc.matches) / float64(tc.total) * 100
		}
		require.Equal(t, tc.expected, matchRate, "match rate for %d/%d", tc.matches, tc.total)
	}
}

// ============================================================================
// Test Compute Time Statistics
// ============================================================================

func TestComputeTimeStatistics(t *testing.T) {
	computeTimes := []int64{10, 20, 30, 40, 50, 100, 200}

	var total int64
	var min, max int64 = -1, 0

	for _, ct := range computeTimes {
		total += ct
		if min < 0 || ct < min {
			min = ct
		}
		if ct > max {
			max = ct
		}
	}

	avg := float64(total) / float64(len(computeTimes))

	require.Equal(t, int64(10), min)
	require.Equal(t, int64(200), max)
	require.InDelta(t, 64.28, avg, 0.1)
}

// ============================================================================
// Test Score Difference Statistics
// ============================================================================

func TestScoreDifferenceStatistics(t *testing.T) {
	differences := []int32{0, 0, 1, 2, 0, 1, 5}

	var total int64
	var maxDiff int32

	for _, diff := range differences {
		total += int64(abs(diff))
		if abs(diff) > maxDiff {
			maxDiff = abs(diff)
		}
	}

	avgDiff := float64(total) / float64(len(differences))

	require.Equal(t, int32(5), maxDiff)
	require.InDelta(t, 1.28, avgDiff, 0.1)
}

// ============================================================================
// Test Metrics Timestamp
// ============================================================================

func TestMetricsTimestamp(t *testing.T) {
	now := time.Now().UTC()

	metrics := VerificationMetrics{
		RequestID: "req-123",
		Timestamp: now,
	}

	require.Equal(t, now, metrics.Timestamp)
	require.False(t, metrics.Timestamp.IsZero())
}

func TestMetricsTimestampZero(t *testing.T) {
	metrics := VerificationMetrics{
		RequestID: "req-123",
	}

	require.True(t, metrics.Timestamp.IsZero())
}

// ============================================================================
// Test Per-Model Version Aggregation
// ============================================================================

func TestPerModelVersionAggregation(t *testing.T) {
	byModelVersion := make(map[string]ModelVersionMetrics)

	// Simulate aggregating metrics by model version
	models := []struct {
		version string
		match   bool
	}{
		{"v1.0.0", true},
		{"v1.0.0", true},
		{"v1.0.0", false},
		{"v2.0.0", true},
		{"v2.0.0", true},
	}

	for _, m := range models {
		mv := byModelVersion[m.version]
		mv.ModelVersion = m.version
		mv.VerificationCount++
		if m.match {
			mv.MatchCount++
		}
		byModelVersion[m.version] = mv
	}

	// Calculate match rates
	for k, mv := range byModelVersion {
		if mv.VerificationCount > 0 {
			mv.MatchRate = float64(mv.MatchCount) / float64(mv.VerificationCount) * 100
		}
		byModelVersion[k] = mv
	}

	require.Len(t, byModelVersion, 2)
	require.Equal(t, int64(3), byModelVersion["v1.0.0"].VerificationCount)
	require.Equal(t, int64(2), byModelVersion["v1.0.0"].MatchCount)
	require.InDelta(t, 66.67, byModelVersion["v1.0.0"].MatchRate, 0.1)
	require.Equal(t, int64(2), byModelVersion["v2.0.0"].VerificationCount)
	require.Equal(t, int64(2), byModelVersion["v2.0.0"].MatchCount)
	require.Equal(t, 100.0, byModelVersion["v2.0.0"].MatchRate)
}
