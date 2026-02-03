package errorbudget

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/observability"
)

func TestCalculateBudgetMinutes(t *testing.T) {
	tests := []struct {
		name      string
		targetSLO float64
		period    time.Duration
		expected  float64
	}{
		{
			name:      "99.90% over 28 days",
			targetSLO: 0.9990,
			period:    28 * 24 * time.Hour,
			expected:  40.32, // (1 - 0.9990) * 28 * 24 * 60
		},
		{
			name:      "99.50% over 30 days",
			targetSLO: 0.9950,
			period:    30 * 24 * time.Hour,
			expected:  216.0, // (1 - 0.9950) * 30 * 24 * 60
		},
		{
			name:      "99.95% over 28 days",
			targetSLO: 0.9995,
			period:    28 * 24 * time.Hour,
			expected:  20.16, // (1 - 0.9995) * 28 * 24 * 60
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := calculateBudgetMinutes(tt.targetSLO, tt.period)
			assert.InDelta(t, tt.expected, actual, 0.01)
		})
	}
}

func TestCalculateStatus(t *testing.T) {
	tests := []struct {
		remainingPct float64
		expected     BudgetStatus
	}{
		{100.0, StatusHealthy},
		{75.0, StatusHealthy},
		{51.0, StatusHealthy},
		{50.0, StatusWarning},
		{30.0, StatusWarning},
		{25.1, StatusWarning},
		{25.0, StatusCritical},
		{10.0, StatusCritical},
		{5.1, StatusCritical},
		{5.0, StatusDepleted},
		{1.0, StatusDepleted},
		{0.0, StatusDepleted},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			actual := calculateStatus(tt.remainingPct)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestManagerRegisterBudget(t *testing.T) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	cfg := BudgetConfig{
		ServiceID:    "virtengine-node",
		SLOID:        "SLO-NODE-001",
		ServiceTier:  TierCritical,
		BudgetPeriod: 28 * 24 * time.Hour,
		TargetSLO:    0.9990,
	}

	err := manager.RegisterBudget(cfg)
	require.NoError(t, err)

	// Verify budget was created
	budget, err := manager.GetBudget("virtengine-node", "SLO-NODE-001")
	require.NoError(t, err)
	assert.Equal(t, "virtengine-node", budget.ServiceID)
	assert.Equal(t, "SLO-NODE-001", budget.SLOID)
	assert.Equal(t, TierCritical, budget.ServiceTier)
	assert.InDelta(t, 40.32, budget.BudgetMinutes, 0.01)
	assert.Equal(t, 0.0, budget.ConsumedMinutes)
	assert.Equal(t, 100.0, budget.RemainingPct)
	assert.Equal(t, StatusHealthy, budget.Status)

	// Test duplicate registration
	err = manager.RegisterBudget(cfg)
	assert.Error(t, err)
}

func TestManagerRecordDowntime(t *testing.T) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	cfg := BudgetConfig{
		ServiceID:    "provider-daemon",
		SLOID:        "SLO-PROV-001",
		ServiceTier:  TierHigh,
		BudgetPeriod: 28 * 24 * time.Hour,
		TargetSLO:    0.9990,
	}

	err := manager.RegisterBudget(cfg)
	require.NoError(t, err)

	// Record 10 minutes of downtime
	err = manager.RecordDowntime("provider-daemon", "SLO-PROV-001", 10*time.Minute)
	require.NoError(t, err)

	budget, err := manager.GetBudget("provider-daemon", "SLO-PROV-001")
	require.NoError(t, err)
	assert.Equal(t, 10.0, budget.ConsumedMinutes)
	assert.InDelta(t, 75.2, budget.RemainingPct, 0.5) // (40.32 - 10) / 40.32 * 100
	assert.Equal(t, StatusHealthy, budget.Status)

	// Record another 20 minutes (total 30, should be in warning)
	err = manager.RecordDowntime("provider-daemon", "SLO-PROV-001", 20*time.Minute)
	require.NoError(t, err)

	budget, err = manager.GetBudget("provider-daemon", "SLO-PROV-001")
	require.NoError(t, err)
	assert.Equal(t, 30.0, budget.ConsumedMinutes)
	assert.InDelta(t, 25.6, budget.RemainingPct, 0.5) // (40.32 - 30) / 40.32 * 100
	assert.Equal(t, StatusWarning, budget.Status)

	// Record another 10 minutes (total 40, should be depleted)
	err = manager.RecordDowntime("provider-daemon", "SLO-PROV-001", 10*time.Minute)
	require.NoError(t, err)

	budget, err = manager.GetBudget("provider-daemon", "SLO-PROV-001")
	require.NoError(t, err)
	assert.Equal(t, 40.0, budget.ConsumedMinutes)
	assert.InDelta(t, 0.8, budget.RemainingPct, 0.5)
	assert.Equal(t, StatusDepleted, budget.Status)
}

func TestManagerRecordFailure(t *testing.T) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	cfg := BudgetConfig{
		ServiceID:    "api-service",
		SLOID:        "SLO-API-001",
		ServiceTier:  TierHigh,
		BudgetPeriod: 28 * 24 * time.Hour,
		TargetSLO:    0.9990,
	}

	err := manager.RegisterBudget(cfg)
	require.NoError(t, err)

	// Record 100 failures
	err = manager.RecordFailure("api-service", "SLO-API-001", 100)
	require.NoError(t, err)

	budget, err := manager.GetBudget("api-service", "SLO-API-001")
	require.NoError(t, err)
	assert.Greater(t, budget.ConsumedMinutes, 0.0)
	assert.Less(t, budget.RemainingPct, 100.0)
}

func TestManagerResetBudget(t *testing.T) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	cfg := BudgetConfig{
		ServiceID:    "benchmark-daemon",
		SLOID:        "SLO-BENCH-001",
		ServiceTier:  TierStandard,
		BudgetPeriod: 28 * 24 * time.Hour,
		TargetSLO:    0.9950,
	}

	err := manager.RegisterBudget(cfg)
	require.NoError(t, err)

	// Consume some budget
	err = manager.RecordDowntime("benchmark-daemon", "SLO-BENCH-001", 50*time.Minute)
	require.NoError(t, err)

	budget, err := manager.GetBudget("benchmark-daemon", "SLO-BENCH-001")
	require.NoError(t, err)
	assert.Greater(t, budget.ConsumedMinutes, 0.0)

	// Reset budget
	err = manager.ResetBudget("benchmark-daemon", "SLO-BENCH-001")
	require.NoError(t, err)

	budget, err = manager.GetBudget("benchmark-daemon", "SLO-BENCH-001")
	require.NoError(t, err)
	assert.Equal(t, 0.0, budget.ConsumedMinutes)
	assert.Equal(t, 100.0, budget.RemainingPct)
	assert.Equal(t, StatusHealthy, budget.Status)
}

func TestManagerGetAllBudgets(t *testing.T) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	configs := []BudgetConfig{
		{
			ServiceID:    "service1",
			SLOID:        "SLO-001",
			ServiceTier:  TierCritical,
			BudgetPeriod: 28 * 24 * time.Hour,
			TargetSLO:    0.9990,
		},
		{
			ServiceID:    "service2",
			SLOID:        "SLO-002",
			ServiceTier:  TierHigh,
			BudgetPeriod: 28 * 24 * time.Hour,
			TargetSLO:    0.9950,
		},
	}

	for _, cfg := range configs {
		err := manager.RegisterBudget(cfg)
		require.NoError(t, err)
	}

	budgets := manager.GetAllBudgets()
	assert.Len(t, budgets, 2)
}

func TestBudgetIsActionAllowed(t *testing.T) {
	tests := []struct {
		name     string
		status   BudgetStatus
		action   string
		expected bool
	}{
		// Healthy - all allowed
		{"healthy feature", StatusHealthy, "feature_release", true},
		{"healthy experimental", StatusHealthy, "experimental_feature", true},
		{"healthy bug fix", StatusHealthy, "bug_fix", true},

		// Warning - experimental blocked
		{"warning feature", StatusWarning, "feature_release", true},
		{"warning experimental", StatusWarning, "experimental_feature", false},
		{"warning bug fix", StatusWarning, "bug_fix", true},

		// Critical - only fixes
		{"critical feature", StatusCritical, "feature_release", false},
		{"critical bug fix", StatusCritical, "bug_fix", true},
		{"critical stability", StatusCritical, "stability_improvement", true},

		// Depleted - emergency only
		{"depleted feature", StatusDepleted, "feature_release", false},
		{"depleted bug fix", StatusDepleted, "bug_fix", false},
		{"depleted emergency", StatusDepleted, "emergency_fix", true},
		{"depleted security", StatusDepleted, "security_patch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			budget := &Budget{Status: tt.status}
			actual := budget.IsActionAllowed(tt.action)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBudgetGetSummary(t *testing.T) {
	budget := &Budget{
		ServiceID:       "test-service",
		SLOID:           "SLO-TEST-001",
		Status:          StatusWarning,
		RemainingPct:    35.5,
		ConsumedMinutes: 25.0,
		BudgetMinutes:   40.0,
		BurnRate:        1.5,
	}

	summary := budget.GetBudgetSummary()
	assert.Contains(t, summary, "test-service")
	assert.Contains(t, summary, "SLO-TEST-001")
	assert.Contains(t, summary, "warning")
	assert.Contains(t, summary, "35.50%")
}

func TestAutoReset(t *testing.T) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	// Create a budget with a very short period (1 second for testing)
	cfg := BudgetConfig{
		ServiceID:    "test-service",
		SLOID:        "SLO-TEST-001",
		ServiceTier:  TierStandard,
		BudgetPeriod: 1 * time.Second,
		TargetSLO:    0.9990,
	}

	err := manager.RegisterBudget(cfg)
	require.NoError(t, err)

	// Consume some budget
	err = manager.RecordDowntime("test-service", "SLO-TEST-001", 1*time.Minute)
	require.NoError(t, err)

	budget, err := manager.GetBudget("test-service", "SLO-TEST-001")
	require.NoError(t, err)
	assert.Greater(t, budget.ConsumedMinutes, 0.0)

	// Wait for period to expire
	time.Sleep(2 * time.Second)

	// Run auto-reset
	manager.AutoReset(context.Background())

	// Budget should be reset
	budget, err = manager.GetBudget("test-service", "SLO-TEST-001")
	require.NoError(t, err)
	assert.Equal(t, 0.0, budget.ConsumedMinutes)
	assert.Equal(t, 100.0, budget.RemainingPct)
}

func TestCalculateBurnRate(t *testing.T) {
	// Create a budget that has been running for 10 hours
	budget := &Budget{
		BudgetMinutes:   40.32, // 28-day period at 99.90%
		BudgetPeriod:    28 * 24 * time.Hour,
		LastReset:       time.Now().Add(-10 * time.Hour),
		ConsumedMinutes: 2.0, // Consumed 2 minutes in 10 hours
	}

	burnRate := calculateBurnRate(budget)

	// Sustainable rate = 40.32 minutes / (28 * 24 hours) = 0.06 minutes/hour
	// Actual rate = 2 minutes / 10 hours = 0.2 minutes/hour
	// Burn rate = 0.2 / 0.06 ≈ 3.33x
	assert.InDelta(t, 3.33, burnRate, 0.5)
}

func BenchmarkRecordDowntime(b *testing.B) {
	obs := observability.NewNoopObservability()
	manager := NewManager(obs)

	cfg := BudgetConfig{
		ServiceID:    "bench-service",
		SLOID:        "SLO-BENCH-001",
		ServiceTier:  TierStandard,
		BudgetPeriod: 28 * 24 * time.Hour,
		TargetSLO:    0.9990,
	}

	_ = manager.RegisterBudget(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.RecordDowntime("bench-service", "SLO-BENCH-001", 1*time.Second)
	}
}
