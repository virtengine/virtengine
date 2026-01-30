// Package types contains types for the HPC module.
//
// VE-5B: Routing enforcement types for on-chain scheduling
package types

import (
	"time"
)

// RoutingDecisionStatus represents the outcome of routing enforcement
type RoutingDecisionStatus string

const (
	// RoutingDecisionStatusApproved indicates the job was routed as per scheduling decision
	RoutingDecisionStatusApproved RoutingDecisionStatus = "approved"

	// RoutingDecisionStatusFallback indicates the job used fallback routing
	RoutingDecisionStatusFallback RoutingDecisionStatus = "fallback"

	// RoutingDecisionStatusRejected indicates the job was rejected (fail-closed)
	RoutingDecisionStatusRejected RoutingDecisionStatus = "rejected"

	// RoutingDecisionStatusRescheduled indicates the job was re-scheduled
	RoutingDecisionStatusRescheduled RoutingDecisionStatus = "rescheduled"
)

// RoutingViolationType represents the type of routing violation
type RoutingViolationType string

const (
	// RoutingViolationClusterMismatch indicates job placed on wrong cluster
	RoutingViolationClusterMismatch RoutingViolationType = "cluster_mismatch"

	// RoutingViolationStaleDecision indicates decision was too old
	RoutingViolationStaleDecision RoutingViolationType = "stale_decision"

	// RoutingViolationMissingDecision indicates no decision was referenced
	RoutingViolationMissingDecision RoutingViolationType = "missing_decision"

	// RoutingViolationCapacityExceeded indicates cluster capacity was exceeded
	RoutingViolationCapacityExceeded RoutingViolationType = "capacity_exceeded"

	// RoutingViolationClusterUnavailable indicates cluster was not available
	RoutingViolationClusterUnavailable RoutingViolationType = "cluster_unavailable"

	// RoutingViolationUnauthorizedFallback indicates fallback was not authorized
	RoutingViolationUnauthorizedFallback RoutingViolationType = "unauthorized_fallback"
)

// RoutingAuditRecord records the routing decision and enforcement for a job
type RoutingAuditRecord struct {
	// RecordID is the unique identifier for this audit record
	RecordID string `json:"record_id"`

	// JobID is the job being routed
	JobID string `json:"job_id"`

	// SchedulingDecisionID is the referenced scheduling decision
	SchedulingDecisionID string `json:"scheduling_decision_id"`

	// ExpectedClusterID is the cluster from the scheduling decision
	ExpectedClusterID string `json:"expected_cluster_id"`

	// ActualClusterID is the cluster where the job was placed
	ActualClusterID string `json:"actual_cluster_id"`

	// Status is the routing decision status
	Status RoutingDecisionStatus `json:"status"`

	// Reason explains the routing outcome
	Reason string `json:"reason"`

	// IsFallback indicates if fallback routing was used
	IsFallback bool `json:"is_fallback"`

	// FallbackReason explains why fallback was used
	FallbackReason string `json:"fallback_reason,omitempty"`

	// FallbackAuthorized indicates if fallback was explicitly authorized
	FallbackAuthorized bool `json:"fallback_authorized"`

	// ViolationType is set if there was a routing violation
	ViolationType RoutingViolationType `json:"violation_type,omitempty"`

	// ViolationDetails provides additional context for violations
	ViolationDetails string `json:"violation_details,omitempty"`

	// DecisionAge is how old the scheduling decision was when enforced
	DecisionAgeBlocks int64 `json:"decision_age_blocks"`

	// DecisionAgeSeconds is the age in seconds
	DecisionAgeSeconds int64 `json:"decision_age_seconds"`

	// ClusterCapacityAtRouting is the cluster capacity when routing occurred
	ClusterCapacityAtRouting *ClusterCapacitySnapshot `json:"cluster_capacity_at_routing,omitempty"`

	// ProviderAddress is the provider handling the job
	ProviderAddress string `json:"provider_address"`

	// CreatedAt is when this record was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is when this record was created
	BlockHeight int64 `json:"block_height"`
}

// ClusterCapacitySnapshot captures cluster capacity at a point in time
type ClusterCapacitySnapshot struct {
	// ClusterID is the cluster identifier
	ClusterID string `json:"cluster_id"`

	// TotalNodes is the total number of nodes
	TotalNodes int32 `json:"total_nodes"`

	// AvailableNodes is the number of available nodes
	AvailableNodes int32 `json:"available_nodes"`

	// ActiveJobs is the number of active jobs
	ActiveJobs int32 `json:"active_jobs"`

	// State is the cluster state at snapshot time
	State ClusterState `json:"state"`

	// SnapshotTime is when this snapshot was taken
	SnapshotTime time.Time `json:"snapshot_time"`
}

// RoutingViolation represents a routing violation event
type RoutingViolation struct {
	// ViolationID is the unique identifier
	ViolationID string `json:"violation_id"`

	// JobID is the job that violated routing
	JobID string `json:"job_id"`

	// SchedulingDecisionID is the referenced decision
	SchedulingDecisionID string `json:"scheduling_decision_id"`

	// ViolationType is the type of violation
	ViolationType RoutingViolationType `json:"violation_type"`

	// ExpectedClusterID is where the job should have been placed
	ExpectedClusterID string `json:"expected_cluster_id"`

	// ActualClusterID is where the job was placed (if any)
	ActualClusterID string `json:"actual_cluster_id,omitempty"`

	// ProviderAddress is the provider involved
	ProviderAddress string `json:"provider_address"`

	// Severity indicates the violation severity (1-5)
	Severity int32 `json:"severity"`

	// Details provides additional context
	Details string `json:"details"`

	// Resolved indicates if the violation has been addressed
	Resolved bool `json:"resolved"`

	// ResolutionDetails explains how the violation was resolved
	ResolutionDetails string `json:"resolution_details,omitempty"`

	// CreatedAt is when the violation was detected
	CreatedAt time.Time `json:"created_at"`

	// ResolvedAt is when the violation was resolved
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// BlockHeight is when this was recorded
	BlockHeight int64 `json:"block_height"`
}

// RoutingPolicy defines the routing enforcement policy
type RoutingPolicy struct {
	// EnforcementMode is the enforcement mode (strict, permissive, audit-only)
	EnforcementMode RoutingEnforcementMode `json:"enforcement_mode"`

	// MaxDecisionAgeBlocks is the maximum age of a scheduling decision in blocks
	MaxDecisionAgeBlocks int64 `json:"max_decision_age_blocks"`

	// MaxDecisionAgeSeconds is the maximum age in seconds
	MaxDecisionAgeSeconds int64 `json:"max_decision_age_seconds"`

	// AllowAutomaticFallback indicates if automatic fallback is permitted
	AllowAutomaticFallback bool `json:"allow_automatic_fallback"`

	// RequireDecisionForSubmission requires a scheduling decision for job submission
	RequireDecisionForSubmission bool `json:"require_decision_for_submission"`

	// ViolationThresholdForAlert is the number of violations before alerting
	ViolationThresholdForAlert int32 `json:"violation_threshold_for_alert"`
}

// RoutingEnforcementMode defines how strictly routing is enforced
type RoutingEnforcementMode string

const (
	// RoutingEnforcementModeStrict rejects jobs that cannot be routed as per decision
	RoutingEnforcementModeStrict RoutingEnforcementMode = "strict"

	// RoutingEnforcementModePermissive allows fallback with logging
	RoutingEnforcementModePermissive RoutingEnforcementMode = "permissive"

	// RoutingEnforcementModeAuditOnly only logs violations without enforcement
	RoutingEnforcementModeAuditOnly RoutingEnforcementMode = "audit_only"
)

// DefaultRoutingPolicy returns the default routing policy
func DefaultRoutingPolicy() RoutingPolicy {
	return RoutingPolicy{
		EnforcementMode:              RoutingEnforcementModeStrict,
		MaxDecisionAgeBlocks:         100, // ~10 minutes at 6s blocks
		MaxDecisionAgeSeconds:        600, // 10 minutes
		AllowAutomaticFallback:       true,
		RequireDecisionForSubmission: true,
		ViolationThresholdForAlert:   5,
	}
}

// Validate validates a routing audit record
func (r *RoutingAuditRecord) Validate() error {
	if r.RecordID == "" {
		return ErrInvalidRoutingAudit.Wrap("record_id cannot be empty")
	}

	if r.JobID == "" {
		return ErrInvalidRoutingAudit.Wrap("job_id cannot be empty")
	}

	if r.SchedulingDecisionID == "" {
		return ErrInvalidRoutingAudit.Wrap("scheduling_decision_id cannot be empty")
	}

	if r.ExpectedClusterID == "" {
		return ErrInvalidRoutingAudit.Wrap("expected_cluster_id cannot be empty")
	}

	return nil
}

// Validate validates a routing violation
func (v *RoutingViolation) Validate() error {
	if v.ViolationID == "" {
		return ErrInvalidRoutingViolation.Wrap("violation_id cannot be empty")
	}

	if v.JobID == "" {
		return ErrInvalidRoutingViolation.Wrap("job_id cannot be empty")
	}

	if v.ViolationType == "" {
		return ErrInvalidRoutingViolation.Wrap("violation_type cannot be empty")
	}

	if v.Severity < 1 || v.Severity > 5 {
		return ErrInvalidRoutingViolation.Wrap("severity must be between 1 and 5")
	}

	return nil
}

// IsStale checks if a scheduling decision is stale based on policy
func (d *SchedulingDecision) IsStale(currentBlockHeight int64, currentTime time.Time, policy RoutingPolicy) bool {
	// Check block age
	if policy.MaxDecisionAgeBlocks > 0 {
		blockAge := currentBlockHeight - d.BlockHeight
		if blockAge > policy.MaxDecisionAgeBlocks {
			return true
		}
	}

	// Check time age
	if policy.MaxDecisionAgeSeconds > 0 {
		timeAge := currentTime.Sub(d.CreatedAt)
		if timeAge.Seconds() > float64(policy.MaxDecisionAgeSeconds) {
			return true
		}
	}

	return false
}
