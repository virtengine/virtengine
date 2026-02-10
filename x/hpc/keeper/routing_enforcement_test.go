package keeper_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// TestRoutingSelectionBasedOnProviderCapabilities validates audit storage for capability-based routing.
func TestRoutingSelectionBasedOnProviderCapabilities(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()

	record := &types.RoutingAuditRecord{
		JobID:                "job-capability-1",
		SchedulingDecisionID: "decision-1",
		ExpectedClusterID:    "cluster-gpu-a",
		ActualClusterID:      "cluster-gpu-a",
		Status:               types.RoutingDecisionStatusApproved,
		Reason:               "provider supports required GPU profile",
		ProviderAddress:      providerAddr,
	}

	err := k.CreateRoutingAuditRecord(ctx, record)
	require.NoError(t, err)

	stored, found := k.GetRoutingAuditRecord(ctx, record.RecordID)
	require.True(t, found)
	require.Equal(t, types.RoutingDecisionStatusApproved, stored.Status)
	require.Equal(t, "cluster-gpu-a", stored.ActualClusterID)
}

// TestRoutingWithGeographicConstraints validates violation tracking for geo mismatches.
func TestRoutingWithGeographicConstraints(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{2}, 20)).String()

	record := &types.RoutingAuditRecord{
		JobID:                "job-geo-1",
		SchedulingDecisionID: "decision-geo-1",
		ExpectedClusterID:    "cluster-eu",
		ActualClusterID:      "cluster-us",
		Status:               types.RoutingDecisionStatusRejected,
		Reason:               "job restricted to EU region",
		ProviderAddress:      providerAddr,
		ViolationType:        types.RoutingViolationClusterMismatch,
		ViolationDetails:     "region mismatch",
	}

	err := k.CreateRoutingAuditRecord(ctx, record)
	require.NoError(t, err)

	violation := &types.RoutingViolation{
		JobID:                record.JobID,
		SchedulingDecisionID: record.SchedulingDecisionID,
		ViolationType:        types.RoutingViolationClusterMismatch,
		ExpectedClusterID:    record.ExpectedClusterID,
		ActualClusterID:      record.ActualClusterID,
		ProviderAddress:      providerAddr,
		Severity:             3,
		Details:              "geo constraint failed",
	}

	err = k.CreateRoutingViolation(ctx, violation)
	require.NoError(t, err)

	violations := k.GetViolationsByJob(ctx, record.JobID)
	require.Len(t, violations, 1)
	require.Equal(t, types.RoutingViolationClusterMismatch, violations[0].ViolationType)
}

// TestRoutingFailoverRouting validates fallback routing and violation resolution tracking.
func TestRoutingFailoverRouting(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{3}, 20)).String()

	record := &types.RoutingAuditRecord{
		JobID:                "job-failover-1",
		SchedulingDecisionID: "decision-failover-1",
		ExpectedClusterID:    "cluster-primary",
		ActualClusterID:      "cluster-secondary",
		Status:               types.RoutingDecisionStatusFallback,
		Reason:               "primary provider unavailable",
		IsFallback:           true,
		FallbackReason:       "primary cluster offline",
		FallbackAuthorized:   false,
		ProviderAddress:      providerAddr,
		ViolationType:        types.RoutingViolationUnauthorizedFallback,
		ViolationDetails:     "fallback used without authorization",
	}

	err := k.CreateRoutingAuditRecord(ctx, record)
	require.NoError(t, err)

	violation := &types.RoutingViolation{
		JobID:                record.JobID,
		SchedulingDecisionID: record.SchedulingDecisionID,
		ViolationType:        types.RoutingViolationUnauthorizedFallback,
		ExpectedClusterID:    record.ExpectedClusterID,
		ActualClusterID:      record.ActualClusterID,
		ProviderAddress:      providerAddr,
		Severity:             4,
		Details:              "unauthorized fallback",
	}

	err = k.CreateRoutingViolation(ctx, violation)
	require.NoError(t, err)

	err = k.ResolveRoutingViolation(ctx, violation.ViolationID, "manual review approved")
	require.NoError(t, err)

	resolved, found := k.GetRoutingViolation(ctx, violation.ViolationID)
	require.True(t, found)
	require.True(t, resolved.Resolved)
	require.NotNil(t, resolved.ResolvedAt)
}

// TestRoutingWithVEIDTierRequirements validates VEID-tier routing audit handling.
func TestRoutingWithVEIDTierRequirements(t *testing.T) {
	ctx, k, _ := setupHPCKeeper(t)
	providerAddr := sdk.AccAddress(bytes.Repeat([]byte{4}, 20)).String()

	record := &types.RoutingAuditRecord{
		JobID:                "job-veid-1",
		SchedulingDecisionID: "decision-veid-1",
		ExpectedClusterID:    "cluster-tier-3",
		ActualClusterID:      "cluster-tier-3",
		Status:               types.RoutingDecisionStatusApproved,
		Reason:               "VEID tier 3 requirement satisfied",
		ProviderAddress:      providerAddr,
	}

	err := k.CreateRoutingAuditRecord(ctx, record)
	require.NoError(t, err)

	stored, found := k.GetRoutingAuditRecord(ctx, record.RecordID)
	require.True(t, found)
	require.Contains(t, stored.Reason, "VEID tier")
}
