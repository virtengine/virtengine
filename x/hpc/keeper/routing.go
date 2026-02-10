// Package keeper implements the HPC module keeper.
//
// VE-5B: Routing audit and violation management
package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// ============================================================================
// Routing Audit Management
// ============================================================================

// CreateRoutingAuditRecord creates a new routing audit record
func (k Keeper) CreateRoutingAuditRecord(ctx sdk.Context, record *types.RoutingAuditRecord) error {
	// Generate record ID if not set
	if record.RecordID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyRoutingAudit)
		record.RecordID = fmt.Sprintf("routing-audit-%d", seq)
	}

	if err := record.Validate(); err != nil {
		return err
	}

	record.CreatedAt = ctx.BlockTime()
	record.BlockHeight = ctx.BlockHeight()

	return k.SetRoutingAuditRecord(ctx, *record)
}

// GetRoutingAuditRecord retrieves a routing audit record by ID
func (k Keeper) GetRoutingAuditRecord(ctx sdk.Context, recordID string) (types.RoutingAuditRecord, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetRoutingAuditKey(recordID))
	if bz == nil {
		return types.RoutingAuditRecord{}, false
	}

	var record types.RoutingAuditRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.RoutingAuditRecord{}, false
	}
	return record, true
}

// GetRoutingAuditRecordsByJob retrieves all routing audit records for a job
func (k Keeper) GetRoutingAuditRecordsByJob(ctx sdk.Context, jobID string) []types.RoutingAuditRecord {
	var records []types.RoutingAuditRecord
	k.WithRoutingAuditRecords(ctx, func(record types.RoutingAuditRecord) bool {
		if record.JobID == jobID {
			records = append(records, record)
		}
		return false
	})
	return records
}

// SetRoutingAuditRecord stores a routing audit record
func (k Keeper) SetRoutingAuditRecord(ctx sdk.Context, record types.RoutingAuditRecord) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(types.GetRoutingAuditKey(record.RecordID), bz)
	return nil
}

// WithRoutingAuditRecords iterates over all routing audit records
func (k Keeper) WithRoutingAuditRecords(ctx sdk.Context, fn func(types.RoutingAuditRecord) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.RoutingAuditPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var record types.RoutingAuditRecord
		if err := json.Unmarshal(iter.Value(), &record); err != nil {
			continue
		}
		if fn(record) {
			break
		}
	}
}

// ============================================================================
// Routing Violation Management
// ============================================================================

// CreateRoutingViolation creates a new routing violation record
func (k Keeper) CreateRoutingViolation(ctx sdk.Context, violation *types.RoutingViolation) error {
	// Generate violation ID if not set
	if violation.ViolationID == "" {
		seq := k.incrementSequence(ctx, types.SequenceKeyRoutingViolation)
		violation.ViolationID = fmt.Sprintf("routing-violation-%d", seq)
	}

	if err := violation.Validate(); err != nil {
		return err
	}

	violation.CreatedAt = ctx.BlockTime()
	violation.BlockHeight = ctx.BlockHeight()

	// Emit violation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_routing_violation",
			sdk.NewAttribute("violation_id", violation.ViolationID),
			sdk.NewAttribute("job_id", violation.JobID),
			sdk.NewAttribute("violation_type", string(violation.ViolationType)),
			sdk.NewAttribute("severity", fmt.Sprintf("%d", violation.Severity)),
			sdk.NewAttribute("provider_address", violation.ProviderAddress),
		),
	)

	return k.SetRoutingViolation(ctx, *violation)
}

// GetRoutingViolation retrieves a routing violation by ID
func (k Keeper) GetRoutingViolation(ctx sdk.Context, violationID string) (types.RoutingViolation, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetRoutingViolationKey(violationID))
	if bz == nil {
		return types.RoutingViolation{}, false
	}

	var violation types.RoutingViolation
	if err := json.Unmarshal(bz, &violation); err != nil {
		return types.RoutingViolation{}, false
	}
	return violation, true
}

// GetViolationsByJob retrieves all routing violations for a job
func (k Keeper) GetViolationsByJob(ctx sdk.Context, jobID string) []types.RoutingViolation {
	var violations []types.RoutingViolation
	k.WithRoutingViolations(ctx, func(v types.RoutingViolation) bool {
		if v.JobID == jobID {
			violations = append(violations, v)
		}
		return false
	})
	return violations
}

// GetViolationsByProvider retrieves all routing violations for a provider
func (k Keeper) GetViolationsByProvider(ctx sdk.Context, providerAddr string) []types.RoutingViolation {
	var violations []types.RoutingViolation
	k.WithRoutingViolations(ctx, func(v types.RoutingViolation) bool {
		if v.ProviderAddress == providerAddr {
			violations = append(violations, v)
		}
		return false
	})
	return violations
}

// GetUnresolvedViolations retrieves all unresolved routing violations
func (k Keeper) GetUnresolvedViolations(ctx sdk.Context) []types.RoutingViolation {
	var violations []types.RoutingViolation
	k.WithRoutingViolations(ctx, func(v types.RoutingViolation) bool {
		if !v.Resolved {
			violations = append(violations, v)
		}
		return false
	})
	return violations
}

// ResolveRoutingViolation marks a violation as resolved
func (k Keeper) ResolveRoutingViolation(ctx sdk.Context, violationID string, resolution string) error {
	violation, exists := k.GetRoutingViolation(ctx, violationID)
	if !exists {
		return types.ErrInvalidRoutingViolation.Wrap("violation not found")
	}

	violation.Resolved = true
	violation.ResolutionDetails = resolution
	now := ctx.BlockTime()
	violation.ResolvedAt = &now

	// Emit resolution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_routing_violation_resolved",
			sdk.NewAttribute("violation_id", violation.ViolationID),
			sdk.NewAttribute("job_id", violation.JobID),
			sdk.NewAttribute("resolution", resolution),
		),
	)

	return k.SetRoutingViolation(ctx, violation)
}

// SetRoutingViolation stores a routing violation
func (k Keeper) SetRoutingViolation(ctx sdk.Context, violation types.RoutingViolation) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(violation)
	if err != nil {
		return err
	}
	store.Set(types.GetRoutingViolationKey(violation.ViolationID), bz)
	return nil
}

// WithRoutingViolations iterates over all routing violations
func (k Keeper) WithRoutingViolations(ctx sdk.Context, fn func(types.RoutingViolation) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.RoutingViolationPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var violation types.RoutingViolation
		if err := json.Unmarshal(iter.Value(), &violation); err != nil {
			continue
		}
		if fn(violation) {
			break
		}
	}
}

// ============================================================================
// Routing Enforcement Helpers
// ============================================================================

// ValidateJobRouting validates that a job is being routed according to its scheduling decision
func (k Keeper) ValidateJobRouting(ctx sdk.Context, job *types.HPCJob, targetClusterID string) (*types.RoutingAuditRecord, error) {
	params := k.GetParams(ctx)
	policy := types.DefaultRoutingPolicy()

	// Check if scheduling decision is required
	if policy.RequireDecisionForSubmission && job.SchedulingDecisionID == "" {
		return nil, types.ErrMissingSchedulingDecision
	}

	// Get the scheduling decision
	decision, exists := k.GetSchedulingDecision(ctx, job.SchedulingDecisionID)
	if !exists {
		return nil, types.ErrRoutingDecisionNotFound
	}

	// Check if decision is stale
	if decision.IsStale(ctx.BlockHeight(), ctx.BlockTime(), policy) {
		return nil, types.ErrRoutingDecisionStale
	}

	// Create audit record
	auditRecord := &types.RoutingAuditRecord{
		JobID:                job.JobID,
		SchedulingDecisionID: decision.DecisionID,
		ExpectedClusterID:    decision.SelectedClusterID,
		ActualClusterID:      targetClusterID,
		DecisionAgeBlocks:    ctx.BlockHeight() - decision.BlockHeight,
		DecisionAgeSeconds:   int64(ctx.BlockTime().Sub(decision.CreatedAt).Seconds()),
		ProviderAddress:      job.ProviderAddress,
	}

	// Capture cluster capacity snapshot
	if cluster, clusterExists := k.GetCluster(ctx, targetClusterID); clusterExists {
		// Count active jobs for this cluster
		var activeJobCount int32
		k.WithJobs(ctx, func(job types.HPCJob) bool {
			if job.ClusterID == targetClusterID && !types.IsTerminalJobState(job.State) {
				activeJobCount++
			}
			return false
		})

		auditRecord.ClusterCapacityAtRouting = &types.ClusterCapacitySnapshot{
			ClusterID:      cluster.ClusterID,
			TotalNodes:     cluster.TotalNodes,
			AvailableNodes: cluster.AvailableNodes,
			ActiveJobs:     activeJobCount,
			State:          cluster.State,
			SnapshotTime:   ctx.BlockTime(),
		}
	}

	// Validate cluster match
	if targetClusterID != decision.SelectedClusterID {
		// This is a mismatch - check if fallback is allowed
		if !policy.AllowAutomaticFallback || params.RoutingEnforcementMode == string(types.RoutingEnforcementModeStrict) {
			auditRecord.Status = types.RoutingDecisionStatusRejected
			auditRecord.Reason = "Cluster mismatch and fallback not authorized"
			auditRecord.ViolationType = types.RoutingViolationClusterMismatch
			auditRecord.ViolationDetails = fmt.Sprintf("Expected cluster %s, got %s", decision.SelectedClusterID, targetClusterID)

			if err := k.CreateRoutingAuditRecord(ctx, auditRecord); err != nil {
				k.Logger(ctx).Error("failed to create routing audit record", "error", err)
			}

			return auditRecord, types.ErrRoutingClusterMismatch
		}

		// Fallback is allowed - record it
		auditRecord.Status = types.RoutingDecisionStatusFallback
		auditRecord.IsFallback = true
		auditRecord.FallbackReason = fmt.Sprintf("Original cluster %s unavailable, using %s", decision.SelectedClusterID, targetClusterID)
		auditRecord.FallbackAuthorized = true
		auditRecord.Reason = "Fallback routing authorized by policy"
	} else {
		// Normal routing
		auditRecord.Status = types.RoutingDecisionStatusApproved
		auditRecord.Reason = "Job routed to scheduled cluster"
	}

	if err := k.CreateRoutingAuditRecord(ctx, auditRecord); err != nil {
		k.Logger(ctx).Error("failed to create routing audit record", "error", err)
	}

	return auditRecord, nil
}

// RefreshSchedulingDecision creates a new scheduling decision for a job with stale decision
func (k Keeper) RefreshSchedulingDecision(ctx sdk.Context, job *types.HPCJob) (*types.SchedulingDecision, error) {
	// Create new scheduling decision
	decision, err := k.ScheduleJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh scheduling decision: %w", err)
	}

	// Create audit record for re-scheduling
	auditRecord := &types.RoutingAuditRecord{
		JobID:                job.JobID,
		SchedulingDecisionID: decision.DecisionID,
		ExpectedClusterID:    decision.SelectedClusterID,
		ActualClusterID:      decision.SelectedClusterID,
		Status:               types.RoutingDecisionStatusRescheduled,
		Reason:               "Scheduling decision refreshed due to stale original decision",
		ProviderAddress:      job.ProviderAddress,
	}

	if err := k.CreateRoutingAuditRecord(ctx, auditRecord); err != nil {
		k.Logger(ctx).Error("failed to create routing audit record for reschedule", "error", err)
	}

	// Emit re-scheduling event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_job_rescheduled",
			sdk.NewAttribute("job_id", job.JobID),
			sdk.NewAttribute("old_decision_id", job.SchedulingDecisionID),
			sdk.NewAttribute("new_decision_id", decision.DecisionID),
			sdk.NewAttribute("new_cluster_id", decision.SelectedClusterID),
		),
	)

	return decision, nil
}
