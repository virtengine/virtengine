// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-5B: Routing Enforcer - enforces on-chain scheduling decisions for job placement
package provider_daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCSchedulingQuerier queries scheduling decisions from the blockchain
type HPCSchedulingQuerier interface {
	// GetSchedulingDecision retrieves a scheduling decision by ID
	GetSchedulingDecision(ctx context.Context, decisionID string) (*hpctypes.SchedulingDecision, error)

	// GetSchedulingDecisionForJob gets the scheduling decision for a specific job
	GetSchedulingDecisionForJob(ctx context.Context, jobID string) (*hpctypes.SchedulingDecision, error)

	// RequestNewSchedulingDecision requests a new scheduling decision for a job
	RequestNewSchedulingDecision(ctx context.Context, job *hpctypes.HPCJob) (*hpctypes.SchedulingDecision, error)

	// GetClusterStatus gets the current status of a cluster
	GetClusterStatus(ctx context.Context, clusterID string) (*HPCClusterStatus, error)

	// GetCurrentBlockHeight gets the current block height
	GetCurrentBlockHeight(ctx context.Context) (int64, error)
}

// HPCClusterStatus represents the current status of an HPC cluster
type HPCClusterStatus struct {
	ClusterID      string                `json:"cluster_id"`
	State          hpctypes.ClusterState `json:"state"`
	TotalNodes     int32                 `json:"total_nodes"`
	AvailableNodes int32                 `json:"available_nodes"`
	ActiveJobs     int32                 `json:"active_jobs"`
	LastUpdated    time.Time             `json:"last_updated"`
}

// RoutingEnforcerConfig configures the routing enforcer
type RoutingEnforcerConfig struct {
	// EnforcementMode is the enforcement mode (strict, permissive, audit_only)
	EnforcementMode hpctypes.RoutingEnforcementMode `json:"enforcement_mode"`

	// MaxDecisionAgeBlocks is the maximum age of a scheduling decision in blocks
	MaxDecisionAgeBlocks int64 `json:"max_decision_age_blocks"`

	// MaxDecisionAgeSeconds is the maximum age in seconds
	MaxDecisionAgeSeconds int64 `json:"max_decision_age_seconds"`

	// AllowAutomaticFallback indicates if automatic fallback is permitted
	AllowAutomaticFallback bool `json:"allow_automatic_fallback"`

	// RequireDecisionForSubmission requires a scheduling decision for job submission
	RequireDecisionForSubmission bool `json:"require_decision_for_submission"`

	// AutoRefreshStaleDecisions automatically refreshes stale decisions
	AutoRefreshStaleDecisions bool `json:"auto_refresh_stale_decisions"`

	// ViolationAlertThreshold is the number of violations before alerting
	ViolationAlertThreshold int32 `json:"violation_alert_threshold"`

	// ClusterID is the cluster this enforcer is managing
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`
}

// DefaultRoutingEnforcerConfig returns the default routing enforcer config
func DefaultRoutingEnforcerConfig() RoutingEnforcerConfig {
	return RoutingEnforcerConfig{
		EnforcementMode:              hpctypes.RoutingEnforcementModeStrict,
		MaxDecisionAgeBlocks:         100,
		MaxDecisionAgeSeconds:        600,
		AllowAutomaticFallback:       true,
		RequireDecisionForSubmission: true,
		AutoRefreshStaleDecisions:    true,
		ViolationAlertThreshold:      5,
	}
}

// RoutingEnforcementResult represents the result of routing enforcement
type RoutingEnforcementResult struct {
	// Allowed indicates if the job placement is allowed
	Allowed bool `json:"allowed"`

	// Decision is the scheduling decision used
	Decision *hpctypes.SchedulingDecision `json:"decision,omitempty"`

	// TargetClusterID is the cluster where the job should be placed
	TargetClusterID string `json:"target_cluster_id"`

	// IsFallback indicates if fallback routing was used
	IsFallback bool `json:"is_fallback"`

	// FallbackReason explains why fallback was used
	FallbackReason string `json:"fallback_reason,omitempty"`

	// WasRescheduled indicates if the job was re-scheduled
	WasRescheduled bool `json:"was_rescheduled"`

	// Violation is set if there was a routing violation
	Violation *RoutingViolationInfo `json:"violation,omitempty"`

	// AuditRecord is the audit record for this enforcement
	AuditRecord *hpctypes.RoutingAuditRecord `json:"audit_record,omitempty"`

	// Error is set if enforcement failed
	Error error `json:"-"`
}

// RoutingViolationInfo contains information about a routing violation
type RoutingViolationInfo struct {
	Type             hpctypes.RoutingViolationType `json:"type"`
	ExpectedCluster  string                        `json:"expected_cluster"`
	AttemptedCluster string                        `json:"attempted_cluster,omitempty"`
	Details          string                        `json:"details"`
	Severity         int32                         `json:"severity"`
}

// RoutingViolationHandler handles routing violations
type RoutingViolationHandler interface {
	// HandleViolation processes a routing violation
	HandleViolation(ctx context.Context, violation *RoutingViolationInfo, job *hpctypes.HPCJob) error
}

// RoutingEnforcer enforces on-chain scheduling decisions for job placement
type RoutingEnforcer struct {
	config   RoutingEnforcerConfig
	querier  HPCSchedulingQuerier
	reporter HPCOnChainReporter
	auditor  HPCAuditLogger

	mu               sync.RWMutex
	violationHandler RoutingViolationHandler
	violationCount   map[string]int32 // provider -> count
	decisionCache    map[string]*cachedDecision
}

type cachedDecision struct {
	decision    *hpctypes.SchedulingDecision
	fetchedAt   time.Time
	blockHeight int64
}

// NewRoutingEnforcer creates a new routing enforcer
func NewRoutingEnforcer(
	config RoutingEnforcerConfig,
	querier HPCSchedulingQuerier,
	reporter HPCOnChainReporter,
	auditor HPCAuditLogger,
) *RoutingEnforcer {
	return &RoutingEnforcer{
		config:         config,
		querier:        querier,
		reporter:       reporter,
		auditor:        auditor,
		violationCount: make(map[string]int32),
		decisionCache:  make(map[string]*cachedDecision),
	}
}

// SetViolationHandler sets the violation handler
func (e *RoutingEnforcer) SetViolationHandler(handler RoutingViolationHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.violationHandler = handler
}

// EnforceRouting enforces routing for a job before submission
func (e *RoutingEnforcer) EnforceRouting(ctx context.Context, job *hpctypes.HPCJob) (*RoutingEnforcementResult, error) {
	result := &RoutingEnforcementResult{
		Allowed: false,
	}

	// Step A: Fetch scheduling decision
	decision, err := e.fetchSchedulingDecision(ctx, job)
	if err != nil {
		if e.config.RequireDecisionForSubmission {
			result.Error = fmt.Errorf("failed to fetch scheduling decision: %w", err)
			result.Violation = &RoutingViolationInfo{
				Type:     hpctypes.RoutingViolationMissingDecision,
				Details:  err.Error(),
				Severity: 4,
			}
			e.handleViolation(ctx, result.Violation, job)
			return result, result.Error
		}
		// In permissive mode, we can proceed without decision
		e.logAudit("routing_decision_missing", job.JobID, map[string]interface{}{
			"error": err.Error(),
		}, false)
	}
	result.Decision = decision

	// Step B: Validate cluster availability and node capacity
	if decision != nil {
		clusterStatus, err := e.validateClusterAvailability(ctx, decision.SelectedClusterID, job)
		if err != nil {
			return e.handleClusterUnavailable(ctx, job, decision, err)
		}
		result.TargetClusterID = clusterStatus.ClusterID
	}

	// Step C: Check for stale decision
	if decision != nil {
		isStale, staleErr := e.checkDecisionStaleness(ctx, decision)
		if isStale {
			return e.handleStaleDecision(ctx, job, decision, staleErr)
		}
	}

	// Routing is approved
	result.Allowed = true
	if decision != nil {
		result.TargetClusterID = decision.SelectedClusterID
	}

	// Create audit record
	result.AuditRecord = e.createAuditRecord(job, decision, result)
	decisionID := ""
	if decision != nil {
		decisionID = decision.DecisionID
	}
	e.logAudit("routing_enforced", job.JobID, map[string]interface{}{
		"decision_id": decisionID,
		"cluster_id":  result.TargetClusterID,
		"is_fallback": result.IsFallback,
	}, true)

	return result, nil
}

// fetchSchedulingDecision fetches the scheduling decision for a job
func (e *RoutingEnforcer) fetchSchedulingDecision(ctx context.Context, job *hpctypes.HPCJob) (*hpctypes.SchedulingDecision, error) {
	// Check cache first
	e.mu.RLock()
	if cached, ok := e.decisionCache[job.SchedulingDecisionID]; ok {
		if time.Since(cached.fetchedAt) < 30*time.Second {
			e.mu.RUnlock()
			return cached.decision, nil
		}
	}
	e.mu.RUnlock()

	var decision *hpctypes.SchedulingDecision
	var err error

	if job.SchedulingDecisionID != "" {
		// Job already has a decision ID, fetch it
		decision, err = e.querier.GetSchedulingDecision(ctx, job.SchedulingDecisionID)
	} else {
		// Request a new scheduling decision
		decision, err = e.querier.RequestNewSchedulingDecision(ctx, job)
		if err == nil && decision != nil {
			job.SchedulingDecisionID = decision.DecisionID
		}
	}

	if err != nil {
		return nil, err
	}

	// Cache the decision
	if decision != nil {
		e.mu.Lock()
		blockHeight, _ := e.querier.GetCurrentBlockHeight(ctx)
		e.decisionCache[decision.DecisionID] = &cachedDecision{
			decision:    decision,
			fetchedAt:   time.Now(),
			blockHeight: blockHeight,
		}
		e.mu.Unlock()
	}

	return decision, nil
}

// validateClusterAvailability validates that the target cluster is available
func (e *RoutingEnforcer) validateClusterAvailability(ctx context.Context, clusterID string, job *hpctypes.HPCJob) (*HPCClusterStatus, error) {
	status, err := e.querier.GetClusterStatus(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster status: %w", err)
	}

	// Check cluster state
	if status.State != hpctypes.ClusterStateActive {
		return nil, fmt.Errorf("cluster %s is not active (state: %s)", clusterID, status.State)
	}

	// Check node capacity
	if status.AvailableNodes < job.Resources.Nodes {
		return nil, fmt.Errorf("cluster %s has insufficient nodes: need %d, available %d",
			clusterID, job.Resources.Nodes, status.AvailableNodes)
	}

	return status, nil
}

// checkDecisionStaleness checks if a decision is stale
func (e *RoutingEnforcer) checkDecisionStaleness(ctx context.Context, decision *hpctypes.SchedulingDecision) (bool, error) {
	currentBlockHeight, err := e.querier.GetCurrentBlockHeight(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get current block height: %w", err)
	}

	// Check block age
	blockAge := currentBlockHeight - decision.BlockHeight
	if e.config.MaxDecisionAgeBlocks > 0 && blockAge > e.config.MaxDecisionAgeBlocks {
		return true, fmt.Errorf("decision is %d blocks old (max: %d)", blockAge, e.config.MaxDecisionAgeBlocks)
	}

	// Check time age
	timeAge := time.Since(decision.CreatedAt)
	if e.config.MaxDecisionAgeSeconds > 0 && timeAge.Seconds() > float64(e.config.MaxDecisionAgeSeconds) {
		return true, fmt.Errorf("decision is %.0f seconds old (max: %d)", timeAge.Seconds(), e.config.MaxDecisionAgeSeconds)
	}

	return false, nil
}

// handleClusterUnavailable handles the case when the scheduled cluster is unavailable
func (e *RoutingEnforcer) handleClusterUnavailable(ctx context.Context, job *hpctypes.HPCJob, decision *hpctypes.SchedulingDecision, originalErr error) (*RoutingEnforcementResult, error) {
	result := &RoutingEnforcementResult{
		Allowed:  false,
		Decision: decision,
		Violation: &RoutingViolationInfo{
			Type:            hpctypes.RoutingViolationClusterUnavailable,
			ExpectedCluster: decision.SelectedClusterID,
			Details:         originalErr.Error(),
			Severity:        3,
		},
	}

	// Check if fallback is allowed
	if !e.config.AllowAutomaticFallback {
		result.Error = fmt.Errorf("cluster unavailable and fallback not allowed: %w", originalErr)
		e.handleViolation(ctx, result.Violation, job)
		return result, result.Error
	}

	// In strict mode, reject the job
	if e.config.EnforcementMode == hpctypes.RoutingEnforcementModeStrict {
		result.Error = fmt.Errorf("cluster unavailable in strict mode: %w", originalErr)
		e.handleViolation(ctx, result.Violation, job)
		return result, result.Error
	}

	// Try to find a fallback cluster by requesting a new decision
	newDecision, err := e.querier.RequestNewSchedulingDecision(ctx, job)
	if err != nil {
		result.Error = fmt.Errorf("failed to get fallback cluster: %w", err)
		e.handleViolation(ctx, result.Violation, job)
		return result, result.Error
	}

	// Use fallback cluster
	result.Allowed = true
	result.Decision = newDecision
	result.TargetClusterID = newDecision.SelectedClusterID
	result.IsFallback = true
	result.FallbackReason = fmt.Sprintf("Original cluster %s unavailable: %s", decision.SelectedClusterID, originalErr.Error())
	result.WasRescheduled = true

	// Update job with new decision
	job.SchedulingDecisionID = newDecision.DecisionID
	job.ClusterID = newDecision.SelectedClusterID

	e.logAudit("routing_fallback_used", job.JobID, map[string]interface{}{
		"original_decision_id": decision.DecisionID,
		"original_cluster_id":  decision.SelectedClusterID,
		"new_decision_id":      newDecision.DecisionID,
		"new_cluster_id":       newDecision.SelectedClusterID,
		"fallback_reason":      result.FallbackReason,
	}, true)

	return result, nil
}

// handleStaleDecision handles stale scheduling decisions
func (e *RoutingEnforcer) handleStaleDecision(ctx context.Context, job *hpctypes.HPCJob, decision *hpctypes.SchedulingDecision, staleErr error) (*RoutingEnforcementResult, error) {
	result := &RoutingEnforcementResult{
		Allowed:  false,
		Decision: decision,
		Violation: &RoutingViolationInfo{
			Type:            hpctypes.RoutingViolationStaleDecision,
			ExpectedCluster: decision.SelectedClusterID,
			Details:         staleErr.Error(),
			Severity:        2,
		},
	}

	// In strict mode without auto-refresh, reject
	if e.config.EnforcementMode == hpctypes.RoutingEnforcementModeStrict && !e.config.AutoRefreshStaleDecisions {
		result.Error = fmt.Errorf("scheduling decision is stale: %w", staleErr)
		e.handleViolation(ctx, result.Violation, job)
		return result, result.Error
	}

	// Try to refresh the decision
	if e.config.AutoRefreshStaleDecisions {
		newDecision, err := e.querier.RequestNewSchedulingDecision(ctx, job)
		if err != nil {
			result.Error = fmt.Errorf("failed to refresh stale decision: %w", err)
			e.handleViolation(ctx, result.Violation, job)
			return result, result.Error
		}

		// Use refreshed decision
		result.Allowed = true
		result.Decision = newDecision
		result.TargetClusterID = newDecision.SelectedClusterID
		result.WasRescheduled = true

		// Update job with new decision
		job.SchedulingDecisionID = newDecision.DecisionID
		job.ClusterID = newDecision.SelectedClusterID

		e.logAudit("routing_decision_refreshed", job.JobID, map[string]interface{}{
			"old_decision_id": decision.DecisionID,
			"new_decision_id": newDecision.DecisionID,
			"new_cluster_id":  newDecision.SelectedClusterID,
		}, true)

		return result, nil
	}

	// In permissive mode, allow with stale decision
	if e.config.EnforcementMode == hpctypes.RoutingEnforcementModePermissive ||
		e.config.EnforcementMode == hpctypes.RoutingEnforcementModeAuditOnly {
		result.Allowed = true
		result.TargetClusterID = decision.SelectedClusterID
		result.Violation.Severity = 1 // Lower severity for audit-only

		e.logAudit("routing_stale_decision_allowed", job.JobID, map[string]interface{}{
			"decision_id":      decision.DecisionID,
			"cluster_id":       decision.SelectedClusterID,
			"enforcement_mode": e.config.EnforcementMode,
		}, false)

		return result, nil
	}

	result.Error = fmt.Errorf("scheduling decision is stale: %w", staleErr)
	return result, result.Error
}

// ValidateJobPlacement validates that a job is being placed on the correct cluster
func (e *RoutingEnforcer) ValidateJobPlacement(ctx context.Context, job *hpctypes.HPCJob, targetClusterID string) error {
	if job.SchedulingDecisionID == "" {
		if e.config.RequireDecisionForSubmission {
			violation := &RoutingViolationInfo{
				Type:             hpctypes.RoutingViolationMissingDecision,
				AttemptedCluster: targetClusterID,
				Details:          "Job submitted without scheduling decision",
				Severity:         4,
			}
			e.handleViolation(ctx, violation, job)
			return hpctypes.ErrMissingSchedulingDecision
		}
		return nil
	}

	decision, err := e.querier.GetSchedulingDecision(ctx, job.SchedulingDecisionID)
	if err != nil {
		return fmt.Errorf("failed to get scheduling decision: %w", err)
	}

	// Check cluster match
	if targetClusterID != decision.SelectedClusterID {
		violation := &RoutingViolationInfo{
			Type:             hpctypes.RoutingViolationClusterMismatch,
			ExpectedCluster:  decision.SelectedClusterID,
			AttemptedCluster: targetClusterID,
			Details:          fmt.Sprintf("Job placed on %s but scheduled for %s", targetClusterID, decision.SelectedClusterID),
			Severity:         5,
		}

		// In strict mode, reject the placement
		if e.config.EnforcementMode == hpctypes.RoutingEnforcementModeStrict {
			e.handleViolation(ctx, violation, job)
			return hpctypes.ErrRoutingClusterMismatch
		}

		// In other modes, log the violation but allow
		e.handleViolation(ctx, violation, job)
	}

	return nil
}

// handleViolation processes a routing violation
func (e *RoutingEnforcer) handleViolation(ctx context.Context, violation *RoutingViolationInfo, job *hpctypes.HPCJob) {
	e.mu.Lock()
	e.violationCount[job.ProviderAddress]++
	count := e.violationCount[job.ProviderAddress]
	e.mu.Unlock()

	e.logAudit("routing_violation", job.JobID, map[string]interface{}{
		"violation_type":    string(violation.Type),
		"expected_cluster":  violation.ExpectedCluster,
		"attempted_cluster": violation.AttemptedCluster,
		"details":           violation.Details,
		"severity":          violation.Severity,
		"provider_address":  job.ProviderAddress,
		"violation_count":   count,
	}, false)

	// Check if we've exceeded the violation threshold
	if count >= e.config.ViolationAlertThreshold {
		e.logAudit("routing_violation_threshold_exceeded", job.JobID, map[string]interface{}{
			"provider_address": job.ProviderAddress,
			"violation_count":  count,
			"threshold":        e.config.ViolationAlertThreshold,
		}, false)
	}

	// Call violation handler if set
	e.mu.RLock()
	handler := e.violationHandler
	e.mu.RUnlock()

	if handler != nil {
		if err := handler.HandleViolation(ctx, violation, job); err != nil {
			e.logAudit("violation_handler_failed", job.JobID, map[string]interface{}{
				"error": err.Error(),
			}, false)
		}
	}
}

// createAuditRecord creates a routing audit record
func (e *RoutingEnforcer) createAuditRecord(job *hpctypes.HPCJob, decision *hpctypes.SchedulingDecision, result *RoutingEnforcementResult) *hpctypes.RoutingAuditRecord {
	record := &hpctypes.RoutingAuditRecord{
		JobID:              job.JobID,
		ActualClusterID:    result.TargetClusterID,
		IsFallback:         result.IsFallback,
		FallbackReason:     result.FallbackReason,
		FallbackAuthorized: result.IsFallback && result.Allowed,
		ProviderAddress:    job.ProviderAddress,
		CreatedAt:          time.Now(),
	}

	if decision != nil {
		record.SchedulingDecisionID = decision.DecisionID
		record.ExpectedClusterID = decision.SelectedClusterID
		record.DecisionAgeSeconds = int64(time.Since(decision.CreatedAt).Seconds())
	}

	if result.Allowed {
		if result.WasRescheduled {
			record.Status = hpctypes.RoutingDecisionStatusRescheduled
			record.Reason = "Job was re-scheduled due to stale or unavailable original decision"
		} else if result.IsFallback {
			record.Status = hpctypes.RoutingDecisionStatusFallback
			record.Reason = result.FallbackReason
		} else {
			record.Status = hpctypes.RoutingDecisionStatusApproved
			record.Reason = "Job routed to scheduled cluster"
		}
	} else {
		record.Status = hpctypes.RoutingDecisionStatusRejected
		if result.Error != nil {
			record.Reason = result.Error.Error()
		}
	}

	if result.Violation != nil {
		record.ViolationType = result.Violation.Type
		record.ViolationDetails = result.Violation.Details
	}

	return record
}

// logAudit logs an audit event
func (e *RoutingEnforcer) logAudit(eventType, jobID string, details map[string]interface{}, success bool) {
	if e.auditor == nil {
		return
	}

	event := HPCAuditEvent{
		Timestamp: time.Now(),
		EventType: eventType,
		JobID:     jobID,
		ClusterID: e.config.ClusterID,
		Details:   details,
		Success:   success,
	}

	e.auditor.LogSecurityEvent(event)
}

// GetViolationCount returns the violation count for a provider
func (e *RoutingEnforcer) GetViolationCount(providerAddress string) int32 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.violationCount[providerAddress]
}

// ResetViolationCount resets the violation count for a provider
func (e *RoutingEnforcer) ResetViolationCount(providerAddress string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.violationCount[providerAddress] = 0
}

// ClearDecisionCache clears the decision cache
func (e *RoutingEnforcer) ClearDecisionCache() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.decisionCache = make(map[string]*cachedDecision)
}
