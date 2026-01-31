// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-5B: Routing Enforcer Tests - tests for on-chain scheduling enforcement
package provider_daemon

import (
	"context"
	"fmt"
	"testing"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// mockSchedulingQuerier implements HPCSchedulingQuerier for testing
type mockSchedulingQuerier struct {
	decisions       map[string]*hpctypes.SchedulingDecision
	clusterStatuses map[string]*HPCClusterStatus
	currentBlock    int64
	failGetDecision bool
	failGetCluster  bool
}

func newMockSchedulingQuerier() *mockSchedulingQuerier {
	return &mockSchedulingQuerier{
		decisions:       make(map[string]*hpctypes.SchedulingDecision),
		clusterStatuses: make(map[string]*HPCClusterStatus),
		currentBlock:    100,
	}
}

func (m *mockSchedulingQuerier) GetSchedulingDecision(ctx context.Context, decisionID string) (*hpctypes.SchedulingDecision, error) {
	if m.failGetDecision {
		return nil, fmt.Errorf("failed to get decision")
	}
	if d, ok := m.decisions[decisionID]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("decision not found: %s", decisionID)
}

func (m *mockSchedulingQuerier) GetSchedulingDecisionForJob(ctx context.Context, jobID string) (*hpctypes.SchedulingDecision, error) {
	for _, d := range m.decisions {
		if d.JobID == jobID {
			return d, nil
		}
	}
	return nil, fmt.Errorf("decision not found for job: %s", jobID)
}

func (m *mockSchedulingQuerier) RequestNewSchedulingDecision(ctx context.Context, job *hpctypes.HPCJob) (*hpctypes.SchedulingDecision, error) {
	if m.failGetDecision {
		return nil, fmt.Errorf("failed to request new decision")
	}
	// Create a new decision
	decision := &hpctypes.SchedulingDecision{
		DecisionID:        fmt.Sprintf("new-decision-%s", job.JobID),
		JobID:             job.JobID,
		SelectedClusterID: "cluster-1",
		DecisionReason:    "Test decision",
		CreatedAt:         time.Now(),
		BlockHeight:       m.currentBlock,
	}
	m.decisions[decision.DecisionID] = decision
	return decision, nil
}

func (m *mockSchedulingQuerier) GetClusterStatus(ctx context.Context, clusterID string) (*HPCClusterStatus, error) {
	if m.failGetCluster {
		return nil, fmt.Errorf("failed to get cluster status")
	}
	if s, ok := m.clusterStatuses[clusterID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("cluster not found: %s", clusterID)
}

func (m *mockSchedulingQuerier) GetCurrentBlockHeight(ctx context.Context) (int64, error) {
	return m.currentBlock, nil
}

func (m *mockSchedulingQuerier) addDecision(decision *hpctypes.SchedulingDecision) {
	m.decisions[decision.DecisionID] = decision
}

func (m *mockSchedulingQuerier) addCluster(cluster *HPCClusterStatus) {
	m.clusterStatuses[cluster.ClusterID] = cluster
}

// mockAuditLogger implements HPCAuditLogger for testing
type mockAuditLogger struct {
	events []HPCAuditEvent
}

func (m *mockAuditLogger) LogJobEvent(event HPCAuditEvent) {
	m.events = append(m.events, event)
}

func (m *mockAuditLogger) LogSecurityEvent(event HPCAuditEvent) {
	m.events = append(m.events, event)
}

func (m *mockAuditLogger) LogUsageReport(event HPCAuditEvent) {
	m.events = append(m.events, event)
}

// mockViolationHandler implements RoutingViolationHandler for testing
type mockViolationHandler struct {
	violations []*RoutingViolationInfo
}

func (m *mockViolationHandler) HandleViolation(ctx context.Context, violation *RoutingViolationInfo, job *hpctypes.HPCJob) error {
	m.violations = append(m.violations, violation)
	return nil
}

func TestRoutingEnforcer_EnforceRouting_Success(t *testing.T) {
	querier := newMockSchedulingQuerier()
	auditor := &mockAuditLogger{}

	// Setup test data
	decision := &hpctypes.SchedulingDecision{
		DecisionID:        "decision-1",
		JobID:             "job-1",
		SelectedClusterID: "cluster-1",
		DecisionReason:    "Best capacity match",
		CreatedAt:         time.Now(),
		BlockHeight:       95, // 5 blocks old
	}
	querier.addDecision(decision)

	clusterStatus := &HPCClusterStatus{
		ClusterID:      "cluster-1",
		State:          hpctypes.ClusterStateActive,
		TotalNodes:     10,
		AvailableNodes: 5,
		LastUpdated:    time.Now(),
	}
	querier.addCluster(clusterStatus)

	config := DefaultRoutingEnforcerConfig()
	config.ClusterID = "cluster-1"

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)

	job := &hpctypes.HPCJob{
		JobID:                "job-1",
		ClusterID:            "cluster-1",
		SchedulingDecisionID: "decision-1",
		ProviderAddress:      "virtengine1provider",
		Resources: hpctypes.JobResources{
			Nodes: 2,
		},
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	if err != nil {
		t.Fatalf("EnforceRouting failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected routing to be allowed")
	}

	if result.TargetClusterID != "cluster-1" {
		t.Errorf("Expected target cluster 'cluster-1', got '%s'", result.TargetClusterID)
	}

	if result.IsFallback {
		t.Error("Expected no fallback")
	}

	if result.Decision == nil {
		t.Error("Expected decision to be set")
	}
}

func TestRoutingEnforcer_EnforceRouting_MissingDecision_Strict(t *testing.T) {
	querier := newMockSchedulingQuerier()
	querier.failGetDecision = true
	auditor := &mockAuditLogger{}
	violationHandler := &mockViolationHandler{}

	config := DefaultRoutingEnforcerConfig()
	config.EnforcementMode = hpctypes.RoutingEnforcementModeStrict
	config.RequireDecisionForSubmission = true

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)
	enforcer.SetViolationHandler(violationHandler)

	job := &hpctypes.HPCJob{
		JobID:           "job-2",
		ProviderAddress: "virtengine1provider",
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	if err == nil {
		t.Fatal("Expected error for missing decision in strict mode")
	}

	if result.Allowed {
		t.Error("Expected routing to be rejected")
	}

	if len(violationHandler.violations) == 0 {
		t.Error("Expected violation to be recorded")
	}

	if violationHandler.violations[0].Type != hpctypes.RoutingViolationMissingDecision {
		t.Errorf("Expected violation type 'missing_decision', got '%s'", violationHandler.violations[0].Type)
	}
}

func TestRoutingEnforcer_EnforceRouting_StaleDecision_AutoRefresh(t *testing.T) {
	querier := newMockSchedulingQuerier()
	auditor := &mockAuditLogger{}

	// Create a stale decision (200 blocks old)
	staleDecision := &hpctypes.SchedulingDecision{
		DecisionID:        "stale-decision",
		JobID:             "job-3",
		SelectedClusterID: "cluster-1",
		DecisionReason:    "Old decision",
		CreatedAt:         time.Now().Add(-20 * time.Minute),
		BlockHeight:       1, // Very old
	}
	querier.addDecision(staleDecision)
	querier.currentBlock = 200

	// Add available cluster
	clusterStatus := &HPCClusterStatus{
		ClusterID:      "cluster-1",
		State:          hpctypes.ClusterStateActive,
		TotalNodes:     10,
		AvailableNodes: 5,
		LastUpdated:    time.Now(),
	}
	querier.addCluster(clusterStatus)

	config := DefaultRoutingEnforcerConfig()
	config.MaxDecisionAgeBlocks = 100
	config.AutoRefreshStaleDecisions = true

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)

	job := &hpctypes.HPCJob{
		JobID:                "job-3",
		SchedulingDecisionID: "stale-decision",
		ProviderAddress:      "virtengine1provider",
		Resources: hpctypes.JobResources{
			Nodes: 2,
		},
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	if err != nil {
		t.Fatalf("EnforceRouting failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected routing to be allowed after refresh")
	}

	if !result.WasRescheduled {
		t.Error("Expected job to be rescheduled")
	}

	// Job should have new decision ID
	if job.SchedulingDecisionID == "stale-decision" {
		t.Error("Expected decision ID to be updated")
	}
}

func TestRoutingEnforcer_EnforceRouting_ClusterUnavailable_Fallback(t *testing.T) {
	querier := newMockSchedulingQuerier()
	auditor := &mockAuditLogger{}

	decision := &hpctypes.SchedulingDecision{
		DecisionID:        "decision-4",
		JobID:             "job-4",
		SelectedClusterID: "cluster-offline",
		DecisionReason:    "Best capacity match",
		CreatedAt:         time.Now(),
		BlockHeight:       95,
	}
	querier.addDecision(decision)

	// Cluster is offline
	offlineCluster := &HPCClusterStatus{
		ClusterID:      "cluster-offline",
		State:          hpctypes.ClusterStateOffline,
		TotalNodes:     10,
		AvailableNodes: 0,
		LastUpdated:    time.Now(),
	}
	querier.addCluster(offlineCluster)

	// Add a fallback cluster
	fallbackCluster := &HPCClusterStatus{
		ClusterID:      "cluster-1",
		State:          hpctypes.ClusterStateActive,
		TotalNodes:     10,
		AvailableNodes: 5,
		LastUpdated:    time.Now(),
	}
	querier.addCluster(fallbackCluster)

	config := DefaultRoutingEnforcerConfig()
	config.EnforcementMode = hpctypes.RoutingEnforcementModePermissive
	config.AllowAutomaticFallback = true

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)

	job := &hpctypes.HPCJob{
		JobID:                "job-4",
		SchedulingDecisionID: "decision-4",
		ProviderAddress:      "virtengine1provider",
		Resources: hpctypes.JobResources{
			Nodes: 2,
		},
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	if err != nil {
		t.Fatalf("EnforceRouting failed: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected routing to be allowed with fallback")
	}

	if !result.IsFallback {
		t.Error("Expected fallback to be used")
	}

	if result.FallbackReason == "" {
		t.Error("Expected fallback reason to be set")
	}
}

func TestRoutingEnforcer_ValidateJobPlacement_ClusterMismatch_Strict(t *testing.T) {
	querier := newMockSchedulingQuerier()
	auditor := &mockAuditLogger{}
	violationHandler := &mockViolationHandler{}

	decision := &hpctypes.SchedulingDecision{
		DecisionID:        "decision-5",
		JobID:             "job-5",
		SelectedClusterID: "cluster-1",
		CreatedAt:         time.Now(),
		BlockHeight:       95,
	}
	querier.addDecision(decision)

	config := DefaultRoutingEnforcerConfig()
	config.EnforcementMode = hpctypes.RoutingEnforcementModeStrict

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)
	enforcer.SetViolationHandler(violationHandler)

	job := &hpctypes.HPCJob{
		JobID:                "job-5",
		SchedulingDecisionID: "decision-5",
		ProviderAddress:      "virtengine1provider",
	}

	ctx := context.Background()
	err := enforcer.ValidateJobPlacement(ctx, job, "cluster-2") // Wrong cluster!

	if err == nil {
		t.Fatal("Expected error for cluster mismatch in strict mode")
	}

	if len(violationHandler.violations) == 0 {
		t.Error("Expected violation to be recorded")
	}

	if violationHandler.violations[0].Type != hpctypes.RoutingViolationClusterMismatch {
		t.Errorf("Expected violation type 'cluster_mismatch', got '%s'", violationHandler.violations[0].Type)
	}

	if violationHandler.violations[0].Severity != 5 {
		t.Errorf("Expected severity 5 for cluster mismatch, got %d", violationHandler.violations[0].Severity)
	}
}

func TestRoutingEnforcer_ViolationThreshold(t *testing.T) {
	querier := newMockSchedulingQuerier()
	querier.failGetDecision = true
	auditor := &mockAuditLogger{}
	violationHandler := &mockViolationHandler{}

	config := DefaultRoutingEnforcerConfig()
	config.ViolationAlertThreshold = 3
	config.RequireDecisionForSubmission = true

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)
	enforcer.SetViolationHandler(violationHandler)

	// Generate violations
	for i := 0; i < 5; i++ {
		job := &hpctypes.HPCJob{
			JobID:           fmt.Sprintf("job-%d", i),
			ProviderAddress: "virtengine1provider",
		}

		ctx := context.Background()
		_, _ = enforcer.EnforceRouting(ctx, job)
	}

	// Check violation count
	count := enforcer.GetViolationCount("virtengine1provider")
	if count != 5 {
		t.Errorf("Expected violation count 5, got %d", count)
	}

	// Check that threshold exceeded event was logged
	foundThresholdEvent := false
	for _, event := range auditor.events {
		if event.EventType == "routing_violation_threshold_exceeded" {
			foundThresholdEvent = true
			break
		}
	}

	if !foundThresholdEvent {
		t.Error("Expected threshold exceeded event to be logged")
	}
}

func TestRoutingEnforcer_AuditRecordCreation(t *testing.T) {
	querier := newMockSchedulingQuerier()
	auditor := &mockAuditLogger{}

	decision := &hpctypes.SchedulingDecision{
		DecisionID:        "decision-audit",
		JobID:             "job-audit",
		SelectedClusterID: "cluster-1",
		DecisionReason:    "Test decision",
		CreatedAt:         time.Now(),
		BlockHeight:       95,
	}
	querier.addDecision(decision)

	clusterStatus := &HPCClusterStatus{
		ClusterID:      "cluster-1",
		State:          hpctypes.ClusterStateActive,
		TotalNodes:     10,
		AvailableNodes: 5,
		LastUpdated:    time.Now(),
	}
	querier.addCluster(clusterStatus)

	config := DefaultRoutingEnforcerConfig()
	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)

	job := &hpctypes.HPCJob{
		JobID:                "job-audit",
		SchedulingDecisionID: "decision-audit",
		ProviderAddress:      "virtengine1provider",
		Resources: hpctypes.JobResources{
			Nodes: 2,
		},
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	if err != nil {
		t.Fatalf("EnforceRouting failed: %v", err)
	}

	if result.AuditRecord == nil {
		t.Fatal("Expected audit record to be created")
	}

	auditRecord := result.AuditRecord
	if auditRecord.JobID != "job-audit" {
		t.Errorf("Expected job ID 'job-audit', got '%s'", auditRecord.JobID)
	}

	if auditRecord.SchedulingDecisionID != "decision-audit" {
		t.Errorf("Expected decision ID 'decision-audit', got '%s'", auditRecord.SchedulingDecisionID)
	}

	if auditRecord.Status != hpctypes.RoutingDecisionStatusApproved {
		t.Errorf("Expected status 'approved', got '%s'", auditRecord.Status)
	}
}

func TestRoutingEnforcer_InsufficientCapacity(t *testing.T) {
	querier := newMockSchedulingQuerier()
	auditor := &mockAuditLogger{}

	decision := &hpctypes.SchedulingDecision{
		DecisionID:        "decision-capacity",
		JobID:             "job-capacity",
		SelectedClusterID: "cluster-1",
		CreatedAt:         time.Now(),
		BlockHeight:       95,
	}
	querier.addDecision(decision)

	// Cluster has insufficient nodes
	clusterStatus := &HPCClusterStatus{
		ClusterID:      "cluster-1",
		State:          hpctypes.ClusterStateActive,
		TotalNodes:     10,
		AvailableNodes: 2, // Only 2 available
		LastUpdated:    time.Now(),
	}
	querier.addCluster(clusterStatus)

	config := DefaultRoutingEnforcerConfig()
	config.EnforcementMode = hpctypes.RoutingEnforcementModeStrict
	config.AllowAutomaticFallback = false

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)

	job := &hpctypes.HPCJob{
		JobID:                "job-capacity",
		SchedulingDecisionID: "decision-capacity",
		ProviderAddress:      "virtengine1provider",
		Resources: hpctypes.JobResources{
			Nodes: 5, // Requesting 5 nodes
		},
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	if err == nil {
		t.Fatal("Expected error for insufficient capacity")
	}

	if result.Allowed {
		t.Error("Expected routing to be rejected")
	}

	if result.Violation == nil {
		t.Fatal("Expected violation to be recorded")
	}

	if result.Violation.Type != hpctypes.RoutingViolationClusterUnavailable {
		t.Errorf("Expected violation type 'cluster_unavailable', got '%s'", result.Violation.Type)
	}
}

func TestRoutingEnforcer_PermissiveMode(t *testing.T) {
	querier := newMockSchedulingQuerier()
	querier.failGetDecision = true
	auditor := &mockAuditLogger{}

	config := DefaultRoutingEnforcerConfig()
	config.EnforcementMode = hpctypes.RoutingEnforcementModePermissive
	config.RequireDecisionForSubmission = false

	enforcer := NewRoutingEnforcer(config, querier, nil, auditor)

	job := &hpctypes.HPCJob{
		JobID:           "job-permissive",
		ProviderAddress: "virtengine1provider",
	}

	ctx := context.Background()
	result, err := enforcer.EnforceRouting(ctx, job)

	// In permissive mode without required decision, should succeed
	if err != nil {
		t.Fatalf("Expected no error in permissive mode: %v", err)
	}

	if !result.Allowed {
		t.Error("Expected routing to be allowed in permissive mode")
	}
}

func TestRoutingEnforcer_ClearDecisionCache(t *testing.T) {
	querier := newMockSchedulingQuerier()
	config := DefaultRoutingEnforcerConfig()
	enforcer := NewRoutingEnforcer(config, querier, nil, nil)

	// Add a decision to cache manually
	decision := &hpctypes.SchedulingDecision{
		DecisionID:        "cached-decision",
		SelectedClusterID: "cluster-1",
		CreatedAt:         time.Now(),
		BlockHeight:       100,
	}
	querier.addDecision(decision)

	// Fetch to populate cache
	ctx := context.Background()
	job := &hpctypes.HPCJob{
		JobID:                "test-job",
		SchedulingDecisionID: "cached-decision",
		ProviderAddress:      "virtengine1provider",
	}
	_, _ = enforcer.fetchSchedulingDecision(ctx, job)

	// Clear cache
	enforcer.ClearDecisionCache()

	// Verify cache is empty
	enforcer.mu.RLock()
	cacheLen := len(enforcer.decisionCache)
	enforcer.mu.RUnlock()

	if cacheLen != 0 {
		t.Errorf("Expected cache to be empty, got %d entries", cacheLen)
	}
}

func TestRoutingEnforcer_ResetViolationCount(t *testing.T) {
	querier := newMockSchedulingQuerier()
	querier.failGetDecision = true
	config := DefaultRoutingEnforcerConfig()
	config.RequireDecisionForSubmission = true

	enforcer := NewRoutingEnforcer(config, querier, nil, nil)

	// Generate violations
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		job := &hpctypes.HPCJob{
			JobID:           fmt.Sprintf("job-%d", i),
			ProviderAddress: "virtengine1provider",
		}
		_, _ = enforcer.EnforceRouting(ctx, job)
	}

	// Verify violations counted
	if enforcer.GetViolationCount("virtengine1provider") != 3 {
		t.Error("Expected 3 violations")
	}

	// Reset count
	enforcer.ResetViolationCount("virtengine1provider")

	// Verify reset
	if enforcer.GetViolationCount("virtengine1provider") != 0 {
		t.Error("Expected 0 violations after reset")
	}
}
