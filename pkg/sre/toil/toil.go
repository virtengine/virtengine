// Package toil provides toil tracking and analysis for SRE operations.
//
// Toil is manual, repetitive, automatable work that scales linearly with service growth.
// This package helps identify, measure, and prioritize toil reduction efforts.
package toil

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/virtengine/virtengine/pkg/observability"
)

// Category represents a category of toil work
type Category string

const (
	CategoryDeployment   Category = "deployment"
	CategoryConfig       Category = "configuration"
	CategoryCertificates Category = "certificates"
	CategoryProvisioning Category = "provisioning"
	CategoryMonitoring   Category = "monitoring"
	CategoryDatabase     Category = "database"
	CategoryLogs         Category = "logs"
	CategoryBackup       Category = "backup"
	CategoryIncident     Category = "incident"
	CategoryOther        Category = "other"
)

// Priority represents automation priority
type Priority int

const (
	PriorityLow Priority = iota + 1
	PriorityMedium
	PriorityHigh
	PriorityCritical
	PriorityImmediate
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	case PriorityImmediate:
		return "immediate"
	default:
		return "unknown"
	}
}

// Entry represents a single toil tracking entry
type Entry struct {
	ID          string        `json:"id"`
	Date        time.Time     `json:"date"`
	Engineer    string        `json:"engineer"`
	Category    Category      `json:"category"`
	Task        string        `json:"task"`
	TimeSpent   time.Duration `json:"time_spent"`
	Frequency   int           `json:"frequency"` // times per month
	Automatable bool          `json:"automatable"`
	Priority    Priority      `json:"priority"`
	Notes       string        `json:"notes"`
	Automated   bool          `json:"automated"` // has this been automated?
	AutomatedAt *time.Time    `json:"automated_at,omitempty"`
}

// Tracker manages toil tracking and analysis
type Tracker struct {
	entries []Entry
	mu      sync.RWMutex
	obs     observability.Observability
}

// NewTracker creates a new toil tracker
func NewTracker(obs observability.Observability) *Tracker {
	return &Tracker{
		entries: make([]Entry, 0),
		obs:     obs,
	}
}

// RecordToil records a toil entry
func (t *Tracker) RecordToil(ctx context.Context, entry Entry) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if entry.Date.IsZero() {
		entry.Date = time.Now()
	}

	// Calculate priority if not set
	if entry.Priority == 0 && entry.Automatable {
		entry.Priority = t.calculatePriority(entry)
	}

	t.entries = append(t.entries, entry)

	t.obs.LogInfo(ctx, "Toil recorded",
		"id", entry.ID,
		"engineer", entry.Engineer,
		"category", entry.Category,
		"task", entry.Task,
		"time_spent", entry.TimeSpent,
		"automatable", entry.Automatable,
		"priority", entry.Priority)

	// Export metrics
	t.exportMetrics(entry)

	return entry.ID, nil
}

// MarkAutomated marks a toil task as automated
func (t *Tracker) MarkAutomated(ctx context.Context, taskName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	updated := false

	for i := range t.entries {
		if t.entries[i].Task == taskName && !t.entries[i].Automated {
			t.entries[i].Automated = true
			t.entries[i].AutomatedAt = &now
			updated = true
		}
	}

	if !updated {
		return fmt.Errorf("task not found or already automated: %s", taskName)
	}

	t.obs.LogInfo(ctx, "Toil task marked as automated",
		"task", taskName,
		"automated_at", now)

	return nil
}

// GetToilPercentage calculates toil percentage for an engineer over a period
func (t *Tracker) GetToilPercentage(engineer string, period time.Duration) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var toilHours float64
	cutoff := time.Now().Add(-period)

	for _, entry := range t.entries {
		if (engineer == "" || entry.Engineer == engineer) && entry.Date.After(cutoff) {
			toilHours += entry.TimeSpent.Hours()
		}
	}

	// Assume 40-hour work week
	weeks := period.Hours() / (7 * 24)
	totalHours := 40.0 * weeks

	if totalHours == 0 {
		return 0
	}

	return (toilHours / totalHours) * 100
}

// GetTeamToilPercentage calculates total team toil percentage
func (t *Tracker) GetTeamToilPercentage(period time.Duration) float64 {
	return t.GetToilPercentage("", period)
}

// GetToilByCategory returns total toil time by category
func (t *Tracker) GetToilByCategory(period time.Duration) map[Category]time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	categoryTime := make(map[Category]time.Duration)
	cutoff := time.Now().Add(-period)

	for _, entry := range t.entries {
		if entry.Date.After(cutoff) {
			categoryTime[entry.Category] += entry.TimeSpent
		}
	}

	return categoryTime
}

// GetTopToilTasks returns the top N toil tasks by total time spent
func (t *Tracker) GetTopToilTasks(period time.Duration, limit int) []ToilSummary {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Aggregate by task
	taskMap := make(map[string]*ToilSummary)
	cutoff := time.Now().Add(-period)

	for _, entry := range t.entries {
		if entry.Date.After(cutoff) {
			if summary, exists := taskMap[entry.Task]; exists {
				summary.TotalTime += entry.TimeSpent
				summary.Occurrences++
			} else {
				taskMap[entry.Task] = &ToilSummary{
					Task:        entry.Task,
					Category:    entry.Category,
					TotalTime:   entry.TimeSpent,
					Occurrences: 1,
					Automatable: entry.Automatable,
					Priority:    entry.Priority,
					Automated:   entry.Automated,
				}
			}
		}
	}

	// Convert to slice
	summaries := make([]ToilSummary, 0, len(taskMap))
	for _, summary := range taskMap {
		summaries = append(summaries, *summary)
	}

	// Sort by total time descending
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TotalTime > summaries[j].TotalTime
	})

	// Return top N
	if limit > 0 && limit < len(summaries) {
		return summaries[:limit]
	}
	return summaries
}

// GetAutomationOpportunities returns toil tasks that should be automated
func (t *Tracker) GetAutomationOpportunities(period time.Duration) []AutomationOpportunity {
	t.mu.RLock()
	defer t.mu.RUnlock()

	opportunities := make([]AutomationOpportunity, 0)
	taskMap := make(map[string]*AutomationOpportunity)
	cutoff := time.Now().Add(-period)

	for _, entry := range t.entries {
		if entry.Date.After(cutoff) && entry.Automatable && !entry.Automated {
			if opp, exists := taskMap[entry.Task]; exists {
				opp.TotalTime += entry.TimeSpent
				opp.Occurrences++
			} else {
				taskMap[entry.Task] = &AutomationOpportunity{
					Task:         entry.Task,
					Category:     entry.Category,
					TotalTime:    entry.TimeSpent,
					Occurrences:  1,
					Priority:     entry.Priority,
					EstimatedROI: calculateROI(entry.TimeSpent, entry.Frequency),
				}
			}
		}
	}

	// Convert to slice
	for _, opp := range taskMap {
		opportunities = append(opportunities, *opp)
	}

	// Sort by priority and ROI
	sort.Slice(opportunities, func(i, j int) bool {
		if opportunities[i].Priority == opportunities[j].Priority {
			return opportunities[i].EstimatedROI > opportunities[j].EstimatedROI
		}
		return opportunities[i].Priority > opportunities[j].Priority
	})

	return opportunities
}

// GetToilTrend returns toil percentage trend over time
func (t *Tracker) GetToilTrend(engineer string, weeks int) []TrendPoint {
	t.mu.RLock()
	defer t.mu.RUnlock()

	trend := make([]TrendPoint, weeks)
	now := time.Now()

	for i := 0; i < weeks; i++ {
		weekStart := now.AddDate(0, 0, -(weeks-i)*7)
		weekEnd := weekStart.AddDate(0, 0, 7)

		var toilHours float64
		for _, entry := range t.entries {
			if (engineer == "" || entry.Engineer == engineer) &&
				entry.Date.After(weekStart) && entry.Date.Before(weekEnd) {
				toilHours += entry.TimeSpent.Hours()
			}
		}

		toilPercentage := (toilHours / 40.0) * 100 // 40-hour week

		trend[i] = TrendPoint{
			Week:           weekStart,
			ToilPercentage: toilPercentage,
			ToilHours:      toilHours,
		}
	}

	return trend
}

// ExportJSON exports toil entries as JSON
func (t *Tracker) ExportJSON() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return json.MarshalIndent(t.entries, "", "  ")
}

// ImportJSON imports toil entries from JSON
func (t *Tracker) ImportJSON(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal toil entries: %w", err)
	}

	t.entries = append(t.entries, entries...)
	return nil
}

// calculatePriority calculates automation priority based on impact
func (t *Tracker) calculatePriority(entry Entry) Priority {
	// Impact = Frequency × Time Per Occurrence
	monthlyHours := float64(entry.Frequency) * entry.TimeSpent.Hours()

	switch {
	case monthlyHours >= 40: // Full week per month
		return PriorityImmediate
	case monthlyHours >= 20: // Half week per month
		return PriorityCritical
	case monthlyHours >= 8: // Full day per month
		return PriorityHigh
	case monthlyHours >= 2: // Quarter day per month
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// exportMetrics exports toil metrics to observability system
func (t *Tracker) exportMetrics(entry Entry) {
	labels := map[string]string{
		"engineer": entry.Engineer,
		"category": string(entry.Category),
		"task":     entry.Task,
	}

	// Record time spent
	t.obs.RecordHistogram(context.Background(), "toil_time_spent_hours",
		entry.TimeSpent.Hours(), labels)

	// Record automatable flag
	automatableValue := 0.0
	if entry.Automatable {
		automatableValue = 1.0
	}
	t.obs.RecordGauge(context.Background(), "toil_automatable", automatableValue, labels)

	// Record priority
	t.obs.RecordGauge(context.Background(), "toil_priority", float64(entry.Priority), labels)
}

// calculateROI estimates ROI for automating a task
func calculateROI(timePerOccurrence time.Duration, frequencyPerMonth int) float64 {
	// Assume 1 week (40 hours) to automate
	automationCost := 40.0
	timeSavedPerYear := timePerOccurrence.Hours() * float64(frequencyPerMonth) * 12

	if automationCost == 0 {
		return 0
	}

	return timeSavedPerYear / automationCost
}

// ToilSummary represents aggregated toil data for a task
type ToilSummary struct {
	Task        string
	Category    Category
	TotalTime   time.Duration
	Occurrences int
	Automatable bool
	Priority    Priority
	Automated   bool
}

// AutomationOpportunity represents a task that should be automated
type AutomationOpportunity struct {
	Task         string
	Category     Category
	TotalTime    time.Duration
	Occurrences  int
	Priority     Priority
	EstimatedROI float64
}

// TrendPoint represents a point in the toil trend
type TrendPoint struct {
	Week           time.Time
	ToilPercentage float64
	ToilHours      float64
}

// GetSummary returns a summary string for toil summary
func (ts *ToilSummary) GetSummary() string {
	status := "not automated"
	if ts.Automated {
		status = "✅ automated"
	} else if ts.Automatable {
		status = fmt.Sprintf("⚠️ automatable (priority: %s)", ts.Priority)
	}

	return fmt.Sprintf(
		"Task: %s, Category: %s, Total Time: %s, Occurrences: %d, Status: %s",
		ts.Task, ts.Category, ts.TotalTime, ts.Occurrences, status,
	)
}

// GetRecommendation returns a recommendation for automation opportunity
func (ao *AutomationOpportunity) GetRecommendation() string {
	urgency := "Consider automating"
	if ao.Priority >= PriorityCritical {
		urgency = "⚠️ Automate immediately"
	} else if ao.Priority >= PriorityHigh {
		urgency = "Automate this quarter"
	}

	return fmt.Sprintf(
		"%s: %s (Category: %s, Time: %s, ROI: %.1fx)",
		urgency, ao.Task, ao.Category, ao.TotalTime, ao.EstimatedROI,
	)
}
