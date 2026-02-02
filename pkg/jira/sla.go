// Package jira provides Jira Service Desk integration for VirtEngine.
//
// VE-919: Jira Service Desk using Waldur
// This file implements SLA tracking for support tickets.
//
// CRITICAL: Never log API tokens or sensitive ticket content.
package jira

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SLAConfig holds SLA configuration
type SLAConfig struct {
	// ResponseTimeTargets maps priority to response time in minutes
	ResponseTimeTargets map[string]int64

	// ResolutionTimeTargets maps priority to resolution time in minutes
	ResolutionTimeTargets map[string]int64

	// BusinessHours defines business hours for SLA calculation
	BusinessHours *BusinessHours

	// ExcludeWeekends excludes weekends from SLA calculation
	ExcludeWeekends bool
}

// BusinessHours defines business hours for SLA calculation
type BusinessHours struct {
	// StartHour is the start hour (0-23)
	StartHour int

	// EndHour is the end hour (0-23)
	EndHour int

	// Timezone is the timezone for business hours
	Timezone string
}

// DefaultSLAConfig returns default SLA configuration
func DefaultSLAConfig() SLAConfig {
	return SLAConfig{
		ResponseTimeTargets: map[string]int64{
			"urgent": 15,   // 15 minutes
			"high":   60,   // 1 hour
			"medium": 240,  // 4 hours
			"low":    1440, // 24 hours
		},
		ResolutionTimeTargets: map[string]int64{
			"urgent": 240,   // 4 hours
			"high":   480,   // 8 hours
			"medium": 2880,  // 2 days
			"low":    10080, // 7 days
		},
		BusinessHours: &BusinessHours{
			StartHour: 9,
			EndHour:   17,
			Timezone:  "UTC",
		},
		ExcludeWeekends: true,
	}
}

// SLATracker tracks SLA metrics for tickets
type SLATracker struct {
	config  SLAConfig
	tickets map[string]*TicketSLA
	mu      sync.RWMutex
}

// TicketSLA holds SLA tracking data for a single ticket
type TicketSLA struct {
	// TicketID is the VirtEngine ticket ID
	TicketID string

	// JiraKey is the Jira issue key
	JiraKey string

	// Priority is the ticket priority
	Priority string

	// CreatedAt is when the ticket was created
	CreatedAt time.Time

	// FirstResponseAt is when the first response was made
	FirstResponseAt *time.Time

	// ResolvedAt is when the ticket was resolved
	ResolvedAt *time.Time

	// ResponseSLABreached indicates if response SLA was breached
	ResponseSLABreached bool

	// ResolutionSLABreached indicates if resolution SLA was breached
	ResolutionSLABreached bool

	// IsPaused indicates if SLA clock is paused (waiting for customer)
	IsPaused bool

	// PausedAt is when SLA was paused
	PausedAt *time.Time

	// TotalPausedMinutes is total paused time in minutes
	TotalPausedMinutes int64

	// Status is the current ticket status
	Status string
}

// ISLATracker defines the SLA tracker interface
type ISLATracker interface {
	// StartTracking starts SLA tracking for a ticket
	StartTracking(ticketID, jiraKey, priority string, createdAt time.Time) error

	// RecordFirstResponse records the first response time
	RecordFirstResponse(ticketID string, responseAt time.Time) error

	// RecordResolution records the resolution time
	RecordResolution(ticketID string, resolvedAt time.Time) error

	// PauseSLA pauses SLA tracking (e.g., waiting for customer)
	PauseSLA(ticketID string) error

	// ResumeSLA resumes SLA tracking
	ResumeSLA(ticketID string) error

	// GetSLAInfo retrieves SLA information for a ticket
	GetSLAInfo(ticketID string) (*SLAInfo, error)

	// GetAllSLAInfo retrieves all SLA information
	GetAllSLAInfo() []*SLAInfo

	// CheckSLABreaches checks for SLA breaches and returns breached tickets
	CheckSLABreaches(ctx context.Context) ([]*TicketSLA, error)

	// GetMetrics returns SLA metrics
	GetMetrics() *SLAMetrics
}

// SLAMetrics holds aggregated SLA metrics
type SLAMetrics struct {
	// TotalTickets is the total number of tracked tickets
	TotalTickets int

	// ActiveTickets is the number of active (unresolved) tickets
	ActiveTickets int

	// ResponseSLAMet is the count of tickets meeting response SLA
	ResponseSLAMet int

	// ResponseSLABreached is the count of tickets breaching response SLA
	ResponseSLABreached int

	// ResolutionSLAMet is the count of tickets meeting resolution SLA
	ResolutionSLAMet int

	// ResolutionSLABreached is the count of tickets breaching resolution SLA
	ResolutionSLABreached int

	// AverageResponseTimeMinutes is the average response time in minutes
	AverageResponseTimeMinutes float64

	// AverageResolutionTimeMinutes is the average resolution time in minutes
	AverageResolutionTimeMinutes float64

	// ResponseSLAComplianceRate is the response SLA compliance percentage
	ResponseSLAComplianceRate float64

	// ResolutionSLAComplianceRate is the resolution SLA compliance percentage
	ResolutionSLAComplianceRate float64
}

// NewSLATracker creates a new SLA tracker
func NewSLATracker(config SLAConfig) *SLATracker {
	return &SLATracker{
		config:  config,
		tickets: make(map[string]*TicketSLA),
	}
}

// Ensure SLATracker implements ISLATracker
var _ ISLATracker = (*SLATracker)(nil)

// StartTracking starts SLA tracking for a ticket
func (t *SLATracker) StartTracking(ticketID, jiraKey, priority string, createdAt time.Time) error {
	if ticketID == "" {
		return fmt.Errorf("sla tracker: ticket ID is required")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.tickets[ticketID]; exists {
		return fmt.Errorf("sla tracker: ticket %s is already being tracked", ticketID)
	}

	t.tickets[ticketID] = &TicketSLA{
		TicketID:  ticketID,
		JiraKey:   jiraKey,
		Priority:  priority,
		CreatedAt: createdAt,
		Status:    "open",
	}

	return nil
}

// RecordFirstResponse records the first response time
func (t *SLATracker) RecordFirstResponse(ticketID string, responseAt time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	ticket, exists := t.tickets[ticketID]
	if !exists {
		return fmt.Errorf("sla tracker: ticket %s not found", ticketID)
	}

	if ticket.FirstResponseAt != nil {
		// First response already recorded
		return nil
	}

	ticket.FirstResponseAt = &responseAt

	// Check if response SLA was breached
	targetMinutes := t.config.ResponseTimeTargets[ticket.Priority]
	if targetMinutes == 0 {
		targetMinutes = t.config.ResponseTimeTargets["medium"]
	}

	elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, responseAt, ticket.TotalPausedMinutes)
	ticket.ResponseSLABreached = elapsedMinutes > targetMinutes

	return nil
}

// RecordResolution records the resolution time
func (t *SLATracker) RecordResolution(ticketID string, resolvedAt time.Time) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	ticket, exists := t.tickets[ticketID]
	if !exists {
		return fmt.Errorf("sla tracker: ticket %s not found", ticketID)
	}

	if ticket.ResolvedAt != nil {
		// Already resolved
		return nil
	}

	// Resume SLA if paused
	if ticket.IsPaused && ticket.PausedAt != nil {
		pausedDuration := resolvedAt.Sub(*ticket.PausedAt)
		ticket.TotalPausedMinutes += int64(pausedDuration.Minutes())
		ticket.IsPaused = false
		ticket.PausedAt = nil
	}

	ticket.ResolvedAt = &resolvedAt
	ticket.Status = "resolved"

	// Check if resolution SLA was breached
	targetMinutes := t.config.ResolutionTimeTargets[ticket.Priority]
	if targetMinutes == 0 {
		targetMinutes = t.config.ResolutionTimeTargets["medium"]
	}

	elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, resolvedAt, ticket.TotalPausedMinutes)
	ticket.ResolutionSLABreached = elapsedMinutes > targetMinutes

	return nil
}

// PauseSLA pauses SLA tracking
func (t *SLATracker) PauseSLA(ticketID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	ticket, exists := t.tickets[ticketID]
	if !exists {
		return fmt.Errorf("sla tracker: ticket %s not found", ticketID)
	}

	if ticket.IsPaused {
		return nil // Already paused
	}

	now := time.Now()
	ticket.IsPaused = true
	ticket.PausedAt = &now
	ticket.Status = "waiting_customer"

	return nil
}

// ResumeSLA resumes SLA tracking
func (t *SLATracker) ResumeSLA(ticketID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	ticket, exists := t.tickets[ticketID]
	if !exists {
		return fmt.Errorf("sla tracker: ticket %s not found", ticketID)
	}

	if !ticket.IsPaused {
		return nil // Not paused
	}

	if ticket.PausedAt != nil {
		pausedDuration := time.Since(*ticket.PausedAt)
		ticket.TotalPausedMinutes += int64(pausedDuration.Minutes())
	}

	ticket.IsPaused = false
	ticket.PausedAt = nil
	ticket.Status = "in_progress"

	return nil
}

// GetSLAInfo retrieves SLA information for a ticket
func (t *SLATracker) GetSLAInfo(ticketID string) (*SLAInfo, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ticket, exists := t.tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("sla tracker: ticket %s not found", ticketID)
	}

	return t.buildSLAInfo(ticket), nil
}

// GetAllSLAInfo retrieves all SLA information
func (t *SLATracker) GetAllSLAInfo() []*SLAInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*SLAInfo, 0, len(t.tickets))
	for _, ticket := range t.tickets {
		result = append(result, t.buildSLAInfo(ticket))
	}

	return result
}

// CheckSLABreaches checks for SLA breaches and returns breached tickets
func (t *SLATracker) CheckSLABreaches(ctx context.Context) ([]*TicketSLA, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	var breached []*TicketSLA

	for _, ticket := range t.tickets {
		// Skip resolved tickets
		if ticket.ResolvedAt != nil {
			continue
		}

		// Check response SLA for tickets without first response
		if ticket.FirstResponseAt == nil && !ticket.ResponseSLABreached {
			targetMinutes := t.config.ResponseTimeTargets[ticket.Priority]
			if targetMinutes == 0 {
				targetMinutes = t.config.ResponseTimeTargets["medium"]
			}

			pausedMinutes := ticket.TotalPausedMinutes
			if ticket.IsPaused && ticket.PausedAt != nil {
				pausedMinutes += int64(now.Sub(*ticket.PausedAt).Minutes())
			}

			elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, now, pausedMinutes)
			if elapsedMinutes > targetMinutes {
				ticket.ResponseSLABreached = true
				breached = append(breached, ticket)
			}
		}

		// Check resolution SLA for unresolved tickets
		if !ticket.ResolutionSLABreached {
			targetMinutes := t.config.ResolutionTimeTargets[ticket.Priority]
			if targetMinutes == 0 {
				targetMinutes = t.config.ResolutionTimeTargets["medium"]
			}

			pausedMinutes := ticket.TotalPausedMinutes
			if ticket.IsPaused && ticket.PausedAt != nil {
				pausedMinutes += int64(now.Sub(*ticket.PausedAt).Minutes())
			}

			elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, now, pausedMinutes)
			if elapsedMinutes > targetMinutes {
				ticket.ResolutionSLABreached = true
				// Add if not already added for response breach
				found := false
				for _, b := range breached {
					if b.TicketID == ticket.TicketID {
						found = true
						break
					}
				}
				if !found {
					breached = append(breached, ticket)
				}
			}
		}
	}

	return breached, nil
}

// GetMetrics returns SLA metrics
func (t *SLATracker) GetMetrics() *SLAMetrics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	metrics := &SLAMetrics{}

	var totalResponseTime int64
	var totalResolutionTime int64
	var respondedCount int
	var resolvedCount int

	for _, ticket := range t.tickets {
		metrics.TotalTickets++

		if ticket.ResolvedAt == nil {
			metrics.ActiveTickets++
		}

		// Response SLA
		if ticket.FirstResponseAt != nil {
			respondedCount++
			responseTime := t.calculateElapsedMinutes(ticket.CreatedAt, *ticket.FirstResponseAt, ticket.TotalPausedMinutes)
			totalResponseTime += responseTime

			if ticket.ResponseSLABreached {
				metrics.ResponseSLABreached++
			} else {
				metrics.ResponseSLAMet++
			}
		}

		// Resolution SLA
		if ticket.ResolvedAt != nil {
			resolvedCount++
			resolutionTime := t.calculateElapsedMinutes(ticket.CreatedAt, *ticket.ResolvedAt, ticket.TotalPausedMinutes)
			totalResolutionTime += resolutionTime

			if ticket.ResolutionSLABreached {
				metrics.ResolutionSLABreached++
			} else {
				metrics.ResolutionSLAMet++
			}
		}
	}

	// Calculate averages
	if respondedCount > 0 {
		metrics.AverageResponseTimeMinutes = float64(totalResponseTime) / float64(respondedCount)
	}
	if resolvedCount > 0 {
		metrics.AverageResolutionTimeMinutes = float64(totalResolutionTime) / float64(resolvedCount)
	}

	// Calculate compliance rates
	totalResponded := metrics.ResponseSLAMet + metrics.ResponseSLABreached
	if totalResponded > 0 {
		metrics.ResponseSLAComplianceRate = float64(metrics.ResponseSLAMet) / float64(totalResponded) * 100
	}

	totalResolved := metrics.ResolutionSLAMet + metrics.ResolutionSLABreached
	if totalResolved > 0 {
		metrics.ResolutionSLAComplianceRate = float64(metrics.ResolutionSLAMet) / float64(totalResolved) * 100
	}

	return metrics
}

// buildSLAInfo builds SLA info for a ticket
func (t *SLATracker) buildSLAInfo(ticket *TicketSLA) *SLAInfo {
	now := time.Now()

	info := &SLAInfo{
		TicketKey: ticket.JiraKey,
	}

	// Response SLA
	responseTarget := t.config.ResponseTimeTargets[ticket.Priority]
	if responseTarget == 0 {
		responseTarget = t.config.ResponseTimeTargets["medium"]
	}

	if ticket.FirstResponseAt != nil {
		elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, *ticket.FirstResponseAt, ticket.TotalPausedMinutes)
		info.ResponseSLA = &SLAMetric{
			Name:              "First Response",
			TargetDuration:    responseTarget,
			ElapsedDuration:   elapsedMinutes,
			RemainingDuration: responseTarget - elapsedMinutes,
			Breached:          ticket.ResponseSLABreached,
			Paused:            false,
			CompletedAt:       ticket.FirstResponseAt,
		}
	} else {
		pausedMinutes := ticket.TotalPausedMinutes
		if ticket.IsPaused && ticket.PausedAt != nil {
			pausedMinutes += int64(now.Sub(*ticket.PausedAt).Minutes())
		}
		elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, now, pausedMinutes)
		info.ResponseSLA = &SLAMetric{
			Name:              "First Response",
			TargetDuration:    responseTarget,
			ElapsedDuration:   elapsedMinutes,
			RemainingDuration: responseTarget - elapsedMinutes,
			Breached:          ticket.ResponseSLABreached,
			Paused:            ticket.IsPaused,
		}
	}

	// Resolution SLA
	resolutionTarget := t.config.ResolutionTimeTargets[ticket.Priority]
	if resolutionTarget == 0 {
		resolutionTarget = t.config.ResolutionTimeTargets["medium"]
	}

	if ticket.ResolvedAt != nil {
		elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, *ticket.ResolvedAt, ticket.TotalPausedMinutes)
		info.ResolutionSLA = &SLAMetric{
			Name:              "Resolution",
			TargetDuration:    resolutionTarget,
			ElapsedDuration:   elapsedMinutes,
			RemainingDuration: resolutionTarget - elapsedMinutes,
			Breached:          ticket.ResolutionSLABreached,
			Paused:            false,
			CompletedAt:       ticket.ResolvedAt,
		}
	} else {
		pausedMinutes := ticket.TotalPausedMinutes
		if ticket.IsPaused && ticket.PausedAt != nil {
			pausedMinutes += int64(now.Sub(*ticket.PausedAt).Minutes())
		}
		elapsedMinutes := t.calculateElapsedMinutes(ticket.CreatedAt, now, pausedMinutes)
		info.ResolutionSLA = &SLAMetric{
			Name:              "Resolution",
			TargetDuration:    resolutionTarget,
			ElapsedDuration:   elapsedMinutes,
			RemainingDuration: resolutionTarget - elapsedMinutes,
			Breached:          ticket.ResolutionSLABreached,
			Paused:            ticket.IsPaused,
		}
	}

	return info
}

// calculateElapsedMinutes calculates elapsed time in minutes
func (t *SLATracker) calculateElapsedMinutes(start, end time.Time, pausedMinutes int64) int64 {
	totalMinutes := int64(end.Sub(start).Minutes())

	// Subtract paused time
	totalMinutes -= pausedMinutes

	// Apply business hours if configured
	if t.config.BusinessHours != nil {
		totalMinutes = t.calculateBusinessMinutes(start, end, pausedMinutes)
	}

	if totalMinutes < 0 {
		return 0
	}

	return totalMinutes
}

// calculateBusinessMinutes calculates elapsed business minutes
func (t *SLATracker) calculateBusinessMinutes(start, end time.Time, pausedMinutes int64) int64 {
	bh := t.config.BusinessHours
	if bh == nil {
		return int64(end.Sub(start).Minutes()) - pausedMinutes
	}

	loc, err := time.LoadLocation(bh.Timezone)
	if err != nil {
		loc = time.UTC
	}

	start = start.In(loc)
	end = end.In(loc)

	var businessMinutes int64
	current := start

	for current.Before(end) {
		// Skip weekends if configured
		if t.config.ExcludeWeekends {
			if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
				current = current.Add(time.Minute)
				continue
			}
		}

		// Check if within business hours
		hour := current.Hour()
		if hour >= bh.StartHour && hour < bh.EndHour {
			businessMinutes++
		}

		current = current.Add(time.Minute)
	}

	return businessMinutes - pausedMinutes
}
