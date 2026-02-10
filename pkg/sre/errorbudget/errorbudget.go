// Package errorbudget provides error budget tracking and management for SLO compliance.
//
// Error budgets are calculated as:
//
//	ErrorBudget = (1 - SLO_Target) × Total_Time_Period
//
// For example, 99.90% SLO over 28 days:
//
//	ErrorBudget = (1 - 0.9990) × 28 days = 40.32 minutes
package errorbudget

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	verrors "github.com/virtengine/virtengine/pkg/errors"
	"github.com/virtengine/virtengine/pkg/observability"
)

// BudgetStatus represents the health status of an error budget
type BudgetStatus string

const (
	// StatusHealthy indicates > 50% budget remaining
	StatusHealthy BudgetStatus = "healthy"
	// StatusWarning indicates 25-50% budget remaining
	StatusWarning BudgetStatus = "warning"
	// StatusCritical indicates 5-25% budget remaining
	StatusCritical BudgetStatus = "critical"
	// StatusDepleted indicates < 5% budget remaining
	StatusDepleted BudgetStatus = "depleted"
)

// ServiceTier represents the criticality tier of a service
type ServiceTier string

const (
	// TierCritical (Tier 0) - Core consensus and transaction processing
	TierCritical ServiceTier = "tier0"
	// TierHigh (Tier 1) - Provider daemon, API services
	TierHigh ServiceTier = "tier1"
	// TierStandard (Tier 2) - Benchmark services, auxiliary features
	TierStandard ServiceTier = "tier2"
)

// Budget represents an error budget for a specific SLO
type Budget struct {
	// ServiceID identifies the service (e.g., "virtengine-node")
	ServiceID string `json:"service_id"`

	// SLOID identifies the specific SLO (e.g., "SLO-NODE-001")
	SLOID string `json:"slo_id"`

	// ServiceTier indicates criticality level
	ServiceTier ServiceTier `json:"service_tier"`

	// BudgetPeriod is the time window for budget calculation (typically 28 days)
	BudgetPeriod time.Duration `json:"budget_period"`

	// TargetSLO is the target percentage (e.g., 0.9990 for 99.90%)
	TargetSLO float64 `json:"target_slo"`

	// BudgetMinutes is the total allowed downtime in minutes
	BudgetMinutes float64 `json:"budget_minutes"`

	// ConsumedMinutes is the current consumption in minutes
	ConsumedMinutes float64 `json:"consumed_minutes"`

	// RemainingPct is the percentage of budget remaining
	RemainingPct float64 `json:"remaining_pct"`

	// LastReset is when the budget period started
	LastReset time.Time `json:"last_reset"`

	// Status is the current health status
	Status BudgetStatus `json:"status"`

	// BurnRate is the current rate of budget consumption (1.0 = sustainable)
	BurnRate float64 `json:"burn_rate"`

	// ProjectedDepletion is the estimated time when budget will be fully consumed
	ProjectedDepletion *time.Time `json:"projected_depletion,omitempty"`
}

// BudgetConfig defines the configuration for an error budget
type BudgetConfig struct {
	ServiceID    string
	SLOID        string
	ServiceTier  ServiceTier
	BudgetPeriod time.Duration
	TargetSLO    float64
}

// Manager manages error budgets for multiple services
type Manager struct {
	budgets map[string]*Budget
	mu      sync.RWMutex
	obs     observability.Observability
}

// NewManager creates a new error budget manager
func NewManager(obs observability.Observability) *Manager {
	return &Manager{
		budgets: make(map[string]*Budget),
		obs:     obs,
	}
}

// RegisterBudget registers a new error budget for tracking
func (m *Manager) RegisterBudget(cfg BudgetConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", cfg.ServiceID, cfg.SLOID)

	if _, exists := m.budgets[key]; exists {
		return fmt.Errorf("budget already registered: %s", key)
	}

	budget := &Budget{
		ServiceID:       cfg.ServiceID,
		SLOID:           cfg.SLOID,
		ServiceTier:     cfg.ServiceTier,
		BudgetPeriod:    cfg.BudgetPeriod,
		TargetSLO:       cfg.TargetSLO,
		BudgetMinutes:   calculateBudgetMinutes(cfg.TargetSLO, cfg.BudgetPeriod),
		ConsumedMinutes: 0,
		RemainingPct:    100.0,
		LastReset:       time.Now(),
		Status:          StatusHealthy,
		BurnRate:        0,
	}

	m.budgets[key] = budget

	m.obs.LogInfo(context.Background(), "Error budget registered",
		"service_id", cfg.ServiceID,
		"slo_id", cfg.SLOID,
		"budget_minutes", budget.BudgetMinutes)

	// Export initial metric
	m.exportMetrics(budget)

	return nil
}

// RecordDowntime records a period of downtime and updates the budget
func (m *Manager) RecordDowntime(serviceID, sloID string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", serviceID, sloID)
	budget, exists := m.budgets[key]
	if !exists {
		return fmt.Errorf("budget not found: %s", key)
	}

	// Add downtime to consumed budget
	downtimeMinutes := duration.Minutes()
	budget.ConsumedMinutes += downtimeMinutes

	// Recalculate remaining percentage
	budget.RemainingPct = ((budget.BudgetMinutes - budget.ConsumedMinutes) / budget.BudgetMinutes) * 100

	// Clamp to 0-100 range
	if budget.RemainingPct < 0 {
		budget.RemainingPct = 0
	}

	// Update status based on remaining budget
	budget.Status = calculateStatus(budget.RemainingPct)

	// Calculate burn rate (consumption rate relative to sustainable rate)
	budget.BurnRate = calculateBurnRate(budget)

	// Calculate projected depletion time
	if budget.BurnRate > 0 {
		timeToDepletion := time.Duration(budget.RemainingPct/budget.BurnRate*float64(budget.BudgetPeriod)) * time.Minute
		projectedTime := time.Now().Add(timeToDepletion)
		budget.ProjectedDepletion = &projectedTime
	} else {
		budget.ProjectedDepletion = nil
	}

	m.obs.LogInfo(context.Background(), "Downtime recorded",
		"service_id", serviceID,
		"slo_id", sloID,
		"downtime_minutes", downtimeMinutes,
		"consumed_minutes", budget.ConsumedMinutes,
		"remaining_pct", budget.RemainingPct,
		"status", budget.Status,
		"burn_rate", budget.BurnRate)

	// Export updated metrics
	m.exportMetrics(budget)

	// Check if alert should be triggered
	m.checkAlerts(budget)

	return nil
}

// RecordFailure records a single failure event (converts to time-based consumption)
func (m *Manager) RecordFailure(serviceID, sloID string, count int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", serviceID, sloID)
	budget, exists := m.budgets[key]
	if !exists {
		return fmt.Errorf("budget not found: %s", key)
	}

	// Convert failure count to time-based consumption
	// Assume each failure consumes a proportional amount of budget
	// This is a simplified model; real implementation may vary
	failureImpactMinutes := float64(count) * (budget.BudgetMinutes / 1000.0)
	budget.ConsumedMinutes += failureImpactMinutes

	// Recalculate remaining percentage
	budget.RemainingPct = ((budget.BudgetMinutes - budget.ConsumedMinutes) / budget.BudgetMinutes) * 100

	// Clamp to 0-100 range
	if budget.RemainingPct < 0 {
		budget.RemainingPct = 0
	}

	// Update status
	budget.Status = calculateStatus(budget.RemainingPct)

	// Calculate burn rate
	budget.BurnRate = calculateBurnRate(budget)

	m.obs.LogInfo(context.Background(), "Failure recorded",
		"service_id", serviceID,
		"slo_id", sloID,
		"failure_count", count,
		"impact_minutes", failureImpactMinutes,
		"remaining_pct", budget.RemainingPct)

	// Export updated metrics
	m.exportMetrics(budget)

	// Check alerts
	m.checkAlerts(budget)

	return nil
}

// GetBudget retrieves the current state of an error budget
func (m *Manager) GetBudget(serviceID, sloID string) (*Budget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", serviceID, sloID)
	budget, exists := m.budgets[key]
	if !exists {
		return nil, fmt.Errorf("budget not found: %s", key)
	}

	// Return a copy to prevent external modification
	budgetCopy := *budget
	return &budgetCopy, nil
}

// GetAllBudgets retrieves all error budgets
func (m *Manager) GetAllBudgets() []*Budget {
	m.mu.RLock()
	defer m.mu.RUnlock()

	budgets := make([]*Budget, 0, len(m.budgets))
	for _, budget := range m.budgets {
		budgetCopy := *budget
		budgets = append(budgets, &budgetCopy)
	}

	return budgets
}

// ResetBudget resets an error budget (typically at the end of the budget period)
func (m *Manager) ResetBudget(serviceID, sloID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", serviceID, sloID)
	budget, exists := m.budgets[key]
	if !exists {
		return fmt.Errorf("budget not found: %s", key)
	}

	budget.ConsumedMinutes = 0
	budget.RemainingPct = 100.0
	budget.LastReset = time.Now()
	budget.Status = StatusHealthy
	budget.BurnRate = 0
	budget.ProjectedDepletion = nil

	m.obs.LogInfo(context.Background(), "Error budget reset",
		"service_id", serviceID,
		"slo_id", sloID)

	// Export reset metrics
	m.exportMetrics(budget)

	return nil
}

// AutoReset automatically resets budgets that have exceeded their period
func (m *Manager) AutoReset(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, budget := range m.budgets {
		if now.Sub(budget.LastReset) >= budget.BudgetPeriod {
			m.obs.LogInfo(ctx, "Auto-resetting expired budget",
				"key", key,
				"period", budget.BudgetPeriod)

			budget.ConsumedMinutes = 0
			budget.RemainingPct = 100.0
			budget.LastReset = now
			budget.Status = StatusHealthy
			budget.BurnRate = 0
			budget.ProjectedDepletion = nil

			m.exportMetrics(budget)
		}
	}
}

// StartAutoResetDaemon starts a background goroutine that auto-resets budgets
func (m *Manager) StartAutoResetDaemon(ctx context.Context, checkInterval time.Duration) {
	verrors.SafeGo("error-budget-reset", func() {
		defer func() {}() // WG Done if needed
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				m.obs.LogInfo(ctx, "Error budget auto-reset daemon stopped")
				return
			case <-ticker.C:
				m.AutoReset(ctx)
			}
		}
	})

	m.obs.LogInfo(ctx, "Error budget auto-reset daemon started", "check_interval", checkInterval)
}

// calculateBudgetMinutes calculates the total allowed downtime in minutes
func calculateBudgetMinutes(targetSLO float64, period time.Duration) float64 {
	return (1.0 - targetSLO) * period.Minutes()
}

// calculateStatus determines the budget status based on remaining percentage
func calculateStatus(remainingPct float64) BudgetStatus {
	switch {
	case remainingPct > 50:
		return StatusHealthy
	case remainingPct > 25:
		return StatusWarning
	case remainingPct > 5:
		return StatusCritical
	default:
		return StatusDepleted
	}
}

// calculateBurnRate calculates the current burn rate
// Burn rate = (consumed per hour) / (sustainable rate per hour)
func calculateBurnRate(budget *Budget) float64 {
	timeSinceReset := time.Since(budget.LastReset)
	if timeSinceReset.Hours() < 1 {
		return 0 // Not enough data
	}

	// Actual consumption rate (minutes per hour)
	actualRate := budget.ConsumedMinutes / timeSinceReset.Hours()

	// Sustainable rate (total budget / total hours in period)
	sustainableRate := budget.BudgetMinutes / budget.BudgetPeriod.Hours()

	if sustainableRate == 0 {
		return 0
	}

	return actualRate / sustainableRate
}

// exportMetrics exports error budget metrics to observability system
func (m *Manager) exportMetrics(budget *Budget) {
	labels := map[string]string{
		"service_id":   budget.ServiceID,
		"slo_id":       budget.SLOID,
		"service_tier": string(budget.ServiceTier),
		"status":       string(budget.Status),
	}

	// Budget remaining percentage
	m.obs.RecordGauge(context.Background(), "error_budget_remaining_pct", budget.RemainingPct, labels)

	// Budget consumed minutes
	m.obs.RecordGauge(context.Background(), "error_budget_consumed_minutes", budget.ConsumedMinutes, labels)

	// Budget total minutes
	m.obs.RecordGauge(context.Background(), "error_budget_total_minutes", budget.BudgetMinutes, labels)

	// Burn rate
	m.obs.RecordGauge(context.Background(), "error_budget_burn_rate", budget.BurnRate, labels)

	// Status as numeric (for alerting)
	statusValue := 0.0
	switch budget.Status {
	case StatusHealthy:
		statusValue = 0
	case StatusWarning:
		statusValue = 1
	case StatusCritical:
		statusValue = 2
	case StatusDepleted:
		statusValue = 3
	}
	m.obs.RecordGauge(context.Background(), "error_budget_status", statusValue, labels)
}

// checkAlerts checks if budget status requires alerting
func (m *Manager) checkAlerts(budget *Budget) {
	ctx := context.Background()

	// Alert on status changes
	switch budget.Status {
	case StatusWarning:
		m.obs.LogWarn(ctx, "Error budget in warning state",
			"service_id", budget.ServiceID,
			"slo_id", budget.SLOID,
			"remaining_pct", budget.RemainingPct)

	case StatusCritical:
		m.obs.LogError(ctx, "Error budget in critical state",
			"service_id", budget.ServiceID,
			"slo_id", budget.SLOID,
			"remaining_pct", budget.RemainingPct,
			"burn_rate", budget.BurnRate)

	case StatusDepleted:
		m.obs.LogError(ctx, "Error budget depleted",
			"service_id", budget.ServiceID,
			"slo_id", budget.SLOID,
			"consumed_minutes", budget.ConsumedMinutes,
			"budget_minutes", budget.BudgetMinutes)
	}

	// Alert on fast burn rate
	if budget.BurnRate >= 5.0 {
		m.obs.LogError(ctx, "Critical burn rate detected",
			"service_id", budget.ServiceID,
			"slo_id", budget.SLOID,
			"burn_rate", budget.BurnRate,
			"projected_depletion", budget.ProjectedDepletion)
	} else if budget.BurnRate >= 2.0 {
		m.obs.LogWarn(ctx, "Elevated burn rate detected",
			"service_id", budget.ServiceID,
			"slo_id", budget.SLOID,
			"burn_rate", budget.BurnRate)
	}
}

// MarshalJSON implements custom JSON marshaling for Budget
func (b *Budget) MarshalJSON() ([]byte, error) {
	type Alias Budget
	return json.Marshal(&struct {
		*Alias
		BudgetPeriod string `json:"budget_period"`
	}{
		Alias:        (*Alias)(b),
		BudgetPeriod: b.BudgetPeriod.String(),
	})
}

// GetBudgetSummary returns a human-readable summary of the budget
func (b *Budget) GetBudgetSummary() string {
	return fmt.Sprintf(
		"Service: %s, SLO: %s, Status: %s, Remaining: %.2f%%, Consumed: %.2f/%.2f minutes, Burn Rate: %.2fx",
		b.ServiceID, b.SLOID, b.Status, b.RemainingPct, b.ConsumedMinutes, b.BudgetMinutes, b.BurnRate,
	)
}

// IsActionAllowed checks if a specific action is allowed based on budget status
func (b *Budget) IsActionAllowed(action string) bool {
	switch b.Status {
	case StatusHealthy:
		return true // All actions allowed

	case StatusWarning:
		// Restrict experimental features
		return action != "experimental_feature"

	case StatusCritical:
		// Only bug fixes allowed
		return action == "bug_fix" || action == "stability_improvement"

	case StatusDepleted:
		// Only emergency fixes allowed
		return action == "emergency_fix" || action == "security_patch"

	default:
		return false
	}
}
