// Package govdata provides government data source integration for identity verification.
//
// SECURITY-004: Cost tracking and management for government API calls
package govdata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Cost Management Errors
// ============================================================================

var (
	// ErrBudgetExceeded is returned when a budget limit is exceeded
	ErrBudgetExceeded = errors.New("budget limit exceeded")

	// ErrCostNotConfigured is returned when cost is not configured for adapter
	ErrCostNotConfigured = errors.New("cost not configured for adapter")

	// ErrInvalidBudgetPeriod is returned for invalid budget period
	ErrInvalidBudgetPeriod = errors.New("invalid budget period")
)

// ============================================================================
// Cost Configuration
// ============================================================================

// CostConfig contains cost management configuration
type CostConfig struct {
	// Enabled indicates if cost tracking is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// TrackRealtime enables real-time cost tracking
	TrackRealtime bool `json:"track_realtime" yaml:"track_realtime"`

	// AlertEnabled enables cost alerts
	AlertEnabled bool `json:"alert_enabled" yaml:"alert_enabled"`

	// AlertThresholdPercent is the percentage threshold for alerts
	AlertThresholdPercent float64 `json:"alert_threshold_percent" yaml:"alert_threshold_percent"`

	// DefaultBudget is the default monthly budget in cents
	DefaultBudgetCents int64 `json:"default_budget_cents" yaml:"default_budget_cents"`

	// AdapterCosts defines per-adapter cost configuration
	AdapterCosts map[string]*AdapterCostConfig `json:"adapter_costs" yaml:"adapter_costs"`
}

// AdapterCostConfig contains cost configuration for an adapter
type AdapterCostConfig struct {
	// CostPerCallCents is the cost per API call in cents
	CostPerCallCents int64 `json:"cost_per_call_cents" yaml:"cost_per_call_cents"`

	// CostPerSuccessfulVerificationCents is cost for successful verification
	CostPerSuccessfulVerificationCents int64 `json:"cost_per_successful_verification_cents" yaml:"cost_per_successful_verification_cents"`

	// MonthlyMinimumCents is the monthly minimum cost in cents
	MonthlyMinimumCents int64 `json:"monthly_minimum_cents" yaml:"monthly_minimum_cents"`

	// MonthlyBudgetCents is the monthly budget limit
	MonthlyBudgetCents int64 `json:"monthly_budget_cents" yaml:"monthly_budget_cents"`

	// DailyBudgetCents is the daily budget limit
	DailyBudgetCents int64 `json:"daily_budget_cents" yaml:"daily_budget_cents"`

	// Currency is the currency code (default: USD)
	Currency string `json:"currency" yaml:"currency"`

	// BillingModel is the billing model (per_call, per_success, tiered)
	BillingModel string `json:"billing_model" yaml:"billing_model"`
}

// DefaultCostConfig returns default cost configuration
func DefaultCostConfig() CostConfig {
	return CostConfig{
		Enabled:               true,
		TrackRealtime:         true,
		AlertEnabled:          true,
		AlertThresholdPercent: 80.0,
		DefaultBudgetCents:    100000, // $1000/month default
		AdapterCosts: map[string]*AdapterCostConfig{
			"aamva": {
				CostPerCallCents:                   50, // $0.50 per call
				CostPerSuccessfulVerificationCents: 75, // $0.75 for success
				MonthlyMinimumCents:                5000,
				MonthlyBudgetCents:                 50000, // $500/month
				Currency:                           "USD",
				BillingModel:                       "per_call",
			},
			"dvs": {
				CostPerCallCents:                   30, // AUD $0.30 per call
				CostPerSuccessfulVerificationCents: 30,
				MonthlyMinimumCents:                0,
				MonthlyBudgetCents:                 30000, // $300/month
				Currency:                           "AUD",
				BillingModel:                       "per_call",
			},
			"eidas": {
				CostPerCallCents:                   100, // €1.00 per call
				CostPerSuccessfulVerificationCents: 100,
				MonthlyMinimumCents:                10000,
				MonthlyBudgetCents:                 100000, // €1000/month
				Currency:                           "EUR",
				BillingModel:                       "per_call",
			},
			"govuk": {
				CostPerCallCents:                   25, // £0.25 per call
				CostPerSuccessfulVerificationCents: 25,
				MonthlyMinimumCents:                0,
				MonthlyBudgetCents:                 25000, // £250/month
				Currency:                           "GBP",
				BillingModel:                       "per_call",
			},
			"pctf": {
				CostPerCallCents:                   40, // CAD $0.40 per call
				CostPerSuccessfulVerificationCents: 40,
				MonthlyMinimumCents:                2000,
				MonthlyBudgetCents:                 40000, // $400/month
				Currency:                           "CAD",
				BillingModel:                       "per_call",
			},
		},
	}
}

// ============================================================================
// Cost Tracking Types
// ============================================================================

// CostRecord represents a single cost record
type CostRecord struct {
	// ID is the record ID
	ID string `json:"id"`

	// AdapterName is the adapter that incurred the cost
	AdapterName string `json:"adapter_name"`

	// Timestamp is when the cost was incurred
	Timestamp time.Time `json:"timestamp"`

	// AmountCents is the cost amount in cents
	AmountCents int64 `json:"amount_cents"`

	// Currency is the currency code
	Currency string `json:"currency"`

	// Operation is the operation type
	Operation string `json:"operation"`

	// Success indicates if the operation was successful
	Success bool `json:"success"`

	// RequestID is the associated request ID
	RequestID string `json:"request_id,omitempty"`

	// Jurisdiction is the verification jurisdiction
	Jurisdiction string `json:"jurisdiction,omitempty"`

	// DocumentType is the document type verified
	DocumentType string `json:"document_type,omitempty"`
}

// CostSummary summarizes costs for a period
type CostSummary struct {
	// Period is the summary period
	Period string `json:"period"`

	// StartTime is the period start
	StartTime time.Time `json:"start_time"`

	// EndTime is the period end
	EndTime time.Time `json:"end_time"`

	// TotalCostCents is the total cost in cents
	TotalCostCents int64 `json:"total_cost_cents"`

	// BudgetCents is the budget for the period
	BudgetCents int64 `json:"budget_cents"`

	// BudgetRemainingCents is remaining budget
	BudgetRemainingCents int64 `json:"budget_remaining_cents"`

	// BudgetUtilizationPercent is budget utilization
	BudgetUtilizationPercent float64 `json:"budget_utilization_percent"`

	// CallCount is the number of API calls
	CallCount int64 `json:"call_count"`

	// SuccessCount is the number of successful calls
	SuccessCount int64 `json:"success_count"`

	// ByAdapter breaks down costs by adapter
	ByAdapter map[string]*AdapterCostSummary `json:"by_adapter"`

	// ByJurisdiction breaks down costs by jurisdiction
	ByJurisdiction map[string]int64 `json:"by_jurisdiction"`
}

// AdapterCostSummary summarizes costs for an adapter
type AdapterCostSummary struct {
	// TotalCostCents is total cost for adapter
	TotalCostCents int64 `json:"total_cost_cents"`

	// CallCount is number of calls
	CallCount int64 `json:"call_count"`

	// SuccessCount is number of successful calls
	SuccessCount int64 `json:"success_count"`

	// AverageCostCents is average cost per call
	AverageCostCents int64 `json:"average_cost_cents"`

	// BudgetCents is the adapter budget
	BudgetCents int64 `json:"budget_cents"`

	// BudgetUtilizationPercent is budget utilization
	BudgetUtilizationPercent float64 `json:"budget_utilization_percent"`
}

// CostAlert represents a cost alert
type CostAlert struct {
	// ID is the alert ID
	ID string `json:"id"`

	// Timestamp is when the alert was generated
	Timestamp time.Time `json:"timestamp"`

	// Level is the alert level
	Level AlertLevel `json:"level"`

	// Type is the alert type
	Type AlertType `json:"type"`

	// Message is the alert message
	Message string `json:"message"`

	// AdapterName is the affected adapter (if applicable)
	AdapterName string `json:"adapter_name,omitempty"`

	// CurrentCostCents is the current cost
	CurrentCostCents int64 `json:"current_cost_cents"`

	// ThresholdCents is the threshold that was exceeded
	ThresholdCents int64 `json:"threshold_cents"`
}

// AlertLevel represents alert severity
type AlertLevel string

const (
	// AlertLevelInfo is informational
	AlertLevelInfo AlertLevel = "info"

	// AlertLevelWarning is a warning
	AlertLevelWarning AlertLevel = "warning"

	// AlertLevelCritical is critical
	AlertLevelCritical AlertLevel = "critical"
)

// AlertType represents the type of cost alert
type AlertType string

const (
	// AlertTypeBudgetWarning is budget threshold warning
	AlertTypeBudgetWarning AlertType = "budget_warning"

	// AlertTypeBudgetExceeded is budget exceeded
	AlertTypeBudgetExceeded AlertType = "budget_exceeded"

	// AlertTypeDailyLimit is daily limit reached
	AlertTypeDailyLimit AlertType = "daily_limit"

	// AlertTypeCostSpike is unusual cost spike
	AlertTypeCostSpike AlertType = "cost_spike"
)

// ============================================================================
// Cost Manager Interface
// ============================================================================

// CostManager manages API cost tracking and budgets
type CostManager interface {
	// RecordCost records a cost for an API call
	RecordCost(ctx context.Context, record *CostRecord) error

	// GetSummary returns cost summary for a period
	GetSummary(ctx context.Context, period string, startTime, endTime time.Time) (*CostSummary, error)

	// CheckBudget checks if an operation is within budget
	CheckBudget(ctx context.Context, adapterName string) error

	// GetAlerts returns recent cost alerts
	GetAlerts(ctx context.Context, since time.Time) ([]*CostAlert, error)

	// SetBudget sets the budget for an adapter
	SetBudget(ctx context.Context, adapterName string, monthlyCents, dailyCents int64) error

	// GetAdapterCost returns cost configuration for an adapter
	GetAdapterCost(adapterName string) (*AdapterCostConfig, error)
}

// ============================================================================
// Cost Manager Implementation
// ============================================================================

// costManager implements CostManager
type costManager struct {
	config       CostConfig
	records      []*CostRecord
	alerts       []*CostAlert
	dailyCosts   map[string]map[string]int64 // adapter -> date -> cost
	monthlyCosts map[string]map[string]int64 // adapter -> month -> cost
	mu           sync.RWMutex
}

// newCostManager creates a new cost manager
func newCostManager(config CostConfig) *costManager {
	return &costManager{
		config:       config,
		records:      make([]*CostRecord, 0),
		alerts:       make([]*CostAlert, 0),
		dailyCosts:   make(map[string]map[string]int64),
		monthlyCosts: make(map[string]map[string]int64),
	}
}

// RecordCost records a cost for an API call
func (c *costManager) RecordCost(ctx context.Context, record *CostRecord) error {
	if !c.config.Enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Store record
	c.records = append(c.records, record)

	// Update aggregates
	dateKey := record.Timestamp.Format("2006-01-02")
	monthKey := record.Timestamp.Format("2006-01")

	if c.dailyCosts[record.AdapterName] == nil {
		c.dailyCosts[record.AdapterName] = make(map[string]int64)
	}
	if c.monthlyCosts[record.AdapterName] == nil {
		c.monthlyCosts[record.AdapterName] = make(map[string]int64)
	}

	c.dailyCosts[record.AdapterName][dateKey] += record.AmountCents
	c.monthlyCosts[record.AdapterName][monthKey] += record.AmountCents

	// Check for alerts
	c.checkAlerts(record.AdapterName, dateKey, monthKey)

	return nil
}

// checkAlerts checks for cost alerts
func (c *costManager) checkAlerts(adapterName, dateKey, monthKey string) {
	if !c.config.AlertEnabled {
		return
	}

	adapterConfig, ok := c.config.AdapterCosts[adapterName]
	if !ok {
		return
	}

	// Check daily budget
	dailyCost := c.dailyCosts[adapterName][dateKey]
	if adapterConfig.DailyBudgetCents > 0 {
		utilization := float64(dailyCost) / float64(adapterConfig.DailyBudgetCents) * 100

		if utilization >= 100 {
			c.alerts = append(c.alerts, &CostAlert{
				ID:               fmt.Sprintf("alert-%d", len(c.alerts)+1),
				Timestamp:        time.Now(),
				Level:            AlertLevelCritical,
				Type:             AlertTypeDailyLimit,
				Message:          fmt.Sprintf("Daily budget exceeded for %s: $%.2f / $%.2f", adapterName, float64(dailyCost)/100, float64(adapterConfig.DailyBudgetCents)/100),
				AdapterName:      adapterName,
				CurrentCostCents: dailyCost,
				ThresholdCents:   adapterConfig.DailyBudgetCents,
			})
		} else if utilization >= c.config.AlertThresholdPercent {
			c.alerts = append(c.alerts, &CostAlert{
				ID:               fmt.Sprintf("alert-%d", len(c.alerts)+1),
				Timestamp:        time.Now(),
				Level:            AlertLevelWarning,
				Type:             AlertTypeBudgetWarning,
				Message:          fmt.Sprintf("Daily budget %.0f%% utilized for %s", utilization, adapterName),
				AdapterName:      adapterName,
				CurrentCostCents: dailyCost,
				ThresholdCents:   int64(float64(adapterConfig.DailyBudgetCents) * c.config.AlertThresholdPercent / 100),
			})
		}
	}

	// Check monthly budget
	monthlyCost := c.monthlyCosts[adapterName][monthKey]
	if adapterConfig.MonthlyBudgetCents > 0 {
		utilization := float64(monthlyCost) / float64(adapterConfig.MonthlyBudgetCents) * 100

		if utilization >= 100 {
			c.alerts = append(c.alerts, &CostAlert{
				ID:               fmt.Sprintf("alert-%d", len(c.alerts)+1),
				Timestamp:        time.Now(),
				Level:            AlertLevelCritical,
				Type:             AlertTypeBudgetExceeded,
				Message:          fmt.Sprintf("Monthly budget exceeded for %s: $%.2f / $%.2f", adapterName, float64(monthlyCost)/100, float64(adapterConfig.MonthlyBudgetCents)/100),
				AdapterName:      adapterName,
				CurrentCostCents: monthlyCost,
				ThresholdCents:   adapterConfig.MonthlyBudgetCents,
			})
		} else if utilization >= c.config.AlertThresholdPercent {
			c.alerts = append(c.alerts, &CostAlert{
				ID:               fmt.Sprintf("alert-%d", len(c.alerts)+1),
				Timestamp:        time.Now(),
				Level:            AlertLevelWarning,
				Type:             AlertTypeBudgetWarning,
				Message:          fmt.Sprintf("Monthly budget %.0f%% utilized for %s", utilization, adapterName),
				AdapterName:      adapterName,
				CurrentCostCents: monthlyCost,
				ThresholdCents:   int64(float64(adapterConfig.MonthlyBudgetCents) * c.config.AlertThresholdPercent / 100),
			})
		}
	}
}

// GetSummary returns cost summary for a period
func (c *costManager) GetSummary(ctx context.Context, period string, startTime, endTime time.Time) (*CostSummary, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := &CostSummary{
		Period:         period,
		StartTime:      startTime,
		EndTime:        endTime,
		ByAdapter:      make(map[string]*AdapterCostSummary),
		ByJurisdiction: make(map[string]int64),
	}

	// Calculate totals from records
	for _, record := range c.records {
		if record.Timestamp.Before(startTime) || record.Timestamp.After(endTime) {
			continue
		}

		summary.TotalCostCents += record.AmountCents
		summary.CallCount++
		if record.Success {
			summary.SuccessCount++
		}

		// By adapter
		if summary.ByAdapter[record.AdapterName] == nil {
			summary.ByAdapter[record.AdapterName] = &AdapterCostSummary{}
		}
		summary.ByAdapter[record.AdapterName].TotalCostCents += record.AmountCents
		summary.ByAdapter[record.AdapterName].CallCount++
		if record.Success {
			summary.ByAdapter[record.AdapterName].SuccessCount++
		}

		// By jurisdiction
		if record.Jurisdiction != "" {
			summary.ByJurisdiction[record.Jurisdiction] += record.AmountCents
		}
	}

	// Calculate averages and utilization
	for adapterName, adapterSummary := range summary.ByAdapter {
		if adapterSummary.CallCount > 0 {
			adapterSummary.AverageCostCents = adapterSummary.TotalCostCents / adapterSummary.CallCount
		}
		if config, ok := c.config.AdapterCosts[adapterName]; ok {
			adapterSummary.BudgetCents = config.MonthlyBudgetCents
			if config.MonthlyBudgetCents > 0 {
				adapterSummary.BudgetUtilizationPercent = float64(adapterSummary.TotalCostCents) / float64(config.MonthlyBudgetCents) * 100
			}
		}
	}

	// Overall budget
	summary.BudgetCents = c.config.DefaultBudgetCents
	summary.BudgetRemainingCents = summary.BudgetCents - summary.TotalCostCents
	if summary.BudgetCents > 0 {
		summary.BudgetUtilizationPercent = float64(summary.TotalCostCents) / float64(summary.BudgetCents) * 100
	}

	return summary, nil
}

// CheckBudget checks if an operation is within budget
func (c *costManager) CheckBudget(ctx context.Context, adapterName string) error {
	if !c.config.Enabled {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	adapterConfig, ok := c.config.AdapterCosts[adapterName]
	if !ok {
		return nil // No budget configured, allow
	}

	now := time.Now()
	dateKey := now.Format("2006-01-02")
	monthKey := now.Format("2006-01")

	// Check daily budget
	if adapterConfig.DailyBudgetCents > 0 {
		dailyCost := c.dailyCosts[adapterName][dateKey]
		if dailyCost >= adapterConfig.DailyBudgetCents {
			return fmt.Errorf("%w: daily limit reached for %s ($%.2f / $%.2f)",
				ErrBudgetExceeded, adapterName,
				float64(dailyCost)/100, float64(adapterConfig.DailyBudgetCents)/100)
		}
	}

	// Check monthly budget
	if adapterConfig.MonthlyBudgetCents > 0 {
		monthlyCost := c.monthlyCosts[adapterName][monthKey]
		if monthlyCost >= adapterConfig.MonthlyBudgetCents {
			return fmt.Errorf("%w: monthly limit reached for %s ($%.2f / $%.2f)",
				ErrBudgetExceeded, adapterName,
				float64(monthlyCost)/100, float64(adapterConfig.MonthlyBudgetCents)/100)
		}
	}

	return nil
}

// GetAlerts returns recent cost alerts
func (c *costManager) GetAlerts(ctx context.Context, since time.Time) ([]*CostAlert, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*CostAlert
	for _, alert := range c.alerts {
		if alert.Timestamp.After(since) {
			result = append(result, alert)
		}
	}
	return result, nil
}

// SetBudget sets the budget for an adapter
func (c *costManager) SetBudget(ctx context.Context, adapterName string, monthlyCents, dailyCents int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config.AdapterCosts[adapterName] == nil {
		c.config.AdapterCosts[adapterName] = &AdapterCostConfig{
			Currency: "USD",
		}
	}

	c.config.AdapterCosts[adapterName].MonthlyBudgetCents = monthlyCents
	c.config.AdapterCosts[adapterName].DailyBudgetCents = dailyCents

	return nil
}

// GetAdapterCost returns cost configuration for an adapter
func (c *costManager) GetAdapterCost(adapterName string) (*AdapterCostConfig, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if config, ok := c.config.AdapterCosts[adapterName]; ok {
		return config, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrCostNotConfigured, adapterName)
}

// CalculateCost calculates the cost for a verification operation
func (c *costManager) CalculateCost(adapterName string, success bool) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	config, ok := c.config.AdapterCosts[adapterName]
	if !ok {
		return 0
	}

	switch config.BillingModel {
	case "per_success":
		if success {
			return config.CostPerSuccessfulVerificationCents
		}
		return 0
	case "per_call":
		fallthrough
	default:
		cost := config.CostPerCallCents
		if success && config.CostPerSuccessfulVerificationCents > config.CostPerCallCents {
			cost = config.CostPerSuccessfulVerificationCents
		}
		return cost
	}
}

// ============================================================================
// Cost Estimation
// ============================================================================

// CostEstimate provides cost estimation for verification operations
type CostEstimate struct {
	// AdapterName is the adapter for the estimate
	AdapterName string `json:"adapter_name"`

	// EstimatedCostCents is the estimated cost
	EstimatedCostCents int64 `json:"estimated_cost_cents"`

	// Currency is the currency
	Currency string `json:"currency"`

	// RemainingBudgetCents is remaining budget after operation
	RemainingBudgetCents int64 `json:"remaining_budget_cents"`

	// WillExceedBudget indicates if operation will exceed budget
	WillExceedBudget bool `json:"will_exceed_budget"`
}

// EstimateCost estimates the cost for a verification operation
func (c *costManager) EstimateCost(adapterName string) (*CostEstimate, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	config, ok := c.config.AdapterCosts[adapterName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrCostNotConfigured, adapterName)
	}

	monthKey := time.Now().Format("2006-01")
	currentMonthly := c.monthlyCosts[adapterName][monthKey]

	estimate := &CostEstimate{
		AdapterName:          adapterName,
		EstimatedCostCents:   config.CostPerCallCents,
		Currency:             config.Currency,
		RemainingBudgetCents: config.MonthlyBudgetCents - currentMonthly - config.CostPerCallCents,
		WillExceedBudget:     currentMonthly+config.CostPerCallCents > config.MonthlyBudgetCents,
	}

	return estimate, nil
}

// GetMonthlyTotal returns total monthly costs
func (c *costManager) GetMonthlyTotal() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	monthKey := time.Now().Format("2006-01")
	var total int64
	for _, costs := range c.monthlyCosts {
		total += costs[monthKey]
	}
	return total
}

// GetDailyTotal returns total daily costs
func (c *costManager) GetDailyTotal() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dateKey := time.Now().Format("2006-01-02")
	var total int64
	for _, costs := range c.dailyCosts {
		total += costs[dateKey]
	}
	return total
}
