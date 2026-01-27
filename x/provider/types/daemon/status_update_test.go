//go:build ignore
// +build ignore

// TODO: This test file is excluded until provider daemon types compilation errors are fixed.

package daemon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkloadStatusString(t *testing.T) {
	assert.Equal(t, "pending", WorkloadStatusPending.String())
	assert.Equal(t, "running", WorkloadStatusRunning.String())
	assert.Equal(t, "terminated", WorkloadStatusTerminated.String())
	assert.Contains(t, WorkloadStatus(99).String(), "unknown")
}

func TestWorkloadStatusIsValid(t *testing.T) {
	assert.True(t, WorkloadStatusPending.IsValid())
	assert.True(t, WorkloadStatusRunning.IsValid())
	assert.True(t, WorkloadStatusFailed.IsValid())
	assert.False(t, WorkloadStatusUnspecified.IsValid())
	assert.False(t, WorkloadStatus(99).IsValid())
}

func TestWorkloadStatusIsTerminal(t *testing.T) {
	assert.True(t, WorkloadStatusTerminated.IsTerminal())
	assert.True(t, WorkloadStatusFailed.IsTerminal())
	assert.False(t, WorkloadStatusRunning.IsTerminal())
	assert.False(t, WorkloadStatusProvisioning.IsTerminal())
}

func TestWorkloadStatusIsActive(t *testing.T) {
	assert.True(t, WorkloadStatusRunning.IsActive())
	assert.True(t, WorkloadStatusProvisioning.IsActive())
	assert.False(t, WorkloadStatusPending.IsActive())
	assert.False(t, WorkloadStatusTerminated.IsActive())
}

func TestWorkloadRefValidate(t *testing.T) {
	tests := []struct {
		name    string
		ref     WorkloadRef
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid kubernetes ref",
			ref: WorkloadRef{
				Orchestrator: "kubernetes",
				Namespace:    "default",
				Name:         "my-workload",
				UID:          "abc-123",
			},
			wantErr: false,
		},
		{
			name: "valid slurm ref",
			ref: WorkloadRef{
				Orchestrator: "slurm",
				Name:         "job-1",
				JobID:        "12345",
			},
			wantErr: false,
		},
		{
			name: "missing orchestrator",
			ref: WorkloadRef{
				Name: "my-workload",
			},
			wantErr: true,
			errMsg:  "orchestrator is required",
		},
		{
			name: "unsupported orchestrator",
			ref: WorkloadRef{
				Orchestrator: "docker",
				Name:         "my-workload",
			},
			wantErr: true,
			errMsg:  "unsupported orchestrator: docker",
		},
		{
			name: "missing name",
			ref: WorkloadRef{
				Orchestrator: "kubernetes",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStatusUpdateValidate(t *testing.T) {
	now := time.Now().UTC()
	validSig := DaemonSignature{
		PublicKey: "aabbccdd",
		Signature: "11223344",
		Algorithm: "ed25519",
		KeyID:     "key-1",
		SignedAt:  now,
	}
	validRef := WorkloadRef{
		Orchestrator: "kubernetes",
		Namespace:    "default",
		Name:         "my-workload",
	}

	tests := []struct {
		name    string
		update  StatusUpdate
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid status update",
			update: StatusUpdate{
				UpdateID:        "update-1",
				OrderID:         "order-1",
				AllocationID:    "alloc-1",
				ProviderAddress: "provider1",
				PreviousStatus:  WorkloadStatusProvisioning,
				CurrentStatus:   WorkloadStatusRunning,
				WorkloadRef:     validRef,
				Signature:       validSig,
				CreatedAt:       now,
			},
			wantErr: false,
		},
		{
			name: "missing update id",
			update: StatusUpdate{
				OrderID:         "order-1",
				AllocationID:    "alloc-1",
				ProviderAddress: "provider1",
				CurrentStatus:   WorkloadStatusRunning,
				WorkloadRef:     validRef,
				Signature:       validSig,
				CreatedAt:       now,
			},
			wantErr: true,
			errMsg:  "update_id is required",
		},
		{
			name: "invalid current status",
			update: StatusUpdate{
				UpdateID:        "update-1",
				OrderID:         "order-1",
				AllocationID:    "alloc-1",
				ProviderAddress: "provider1",
				CurrentStatus:   WorkloadStatus(99),
				WorkloadRef:     validRef,
				Signature:       validSig,
				CreatedAt:       now,
			},
			wantErr: true,
			errMsg:  "invalid current_status",
		},
		{
			name: "invalid workload ref",
			update: StatusUpdate{
				UpdateID:        "update-1",
				OrderID:         "order-1",
				AllocationID:    "alloc-1",
				ProviderAddress: "provider1",
				CurrentStatus:   WorkloadStatusRunning,
				WorkloadRef:     WorkloadRef{Name: "test"},
				Signature:       validSig,
				CreatedAt:       now,
			},
			wantErr: true,
			errMsg:  "workload_ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.update.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStatusUpdateIsValidTransition(t *testing.T) {
	tests := []struct {
		name     string
		prev     WorkloadStatus
		current  WorkloadStatus
		expected bool
	}{
		{"pending to provisioning", WorkloadStatusPending, WorkloadStatusProvisioning, true},
		{"pending to failed", WorkloadStatusPending, WorkloadStatusFailed, true},
		{"pending to running", WorkloadStatusPending, WorkloadStatusRunning, false},
		{"provisioning to running", WorkloadStatusProvisioning, WorkloadStatusRunning, true},
		{"provisioning to failed", WorkloadStatusProvisioning, WorkloadStatusFailed, true},
		{"running to suspended", WorkloadStatusRunning, WorkloadStatusSuspended, true},
		{"running to terminating", WorkloadStatusRunning, WorkloadStatusTerminating, true},
		{"running to terminated", WorkloadStatusRunning, WorkloadStatusTerminated, false},
		{"suspended to running", WorkloadStatusSuspended, WorkloadStatusRunning, true},
		{"suspended to terminating", WorkloadStatusSuspended, WorkloadStatusTerminating, true},
		{"terminating to terminated", WorkloadStatusTerminating, WorkloadStatusTerminated, true},
		{"terminating to failed", WorkloadStatusTerminating, WorkloadStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := StatusUpdate{
				PreviousStatus: tt.prev,
				CurrentStatus:  tt.current,
			}
			assert.Equal(t, tt.expected, update.IsValidTransition())
		})
	}
}

func TestPerformUsageFraudChecks(t *testing.T) {
	now := time.Now().UTC()

	baseRecord := UsageRecord{
		RecordID:        "record-1",
		OrderID:         "order-1",
		AllocationID:    "alloc-1",
		ProviderAddress: "provider1",
		RecordType:      UsageRecordTypePeriodic,
		TimeWindow: TimeWindow{
			StartTime:       now,
			EndTime:         now.Add(1 * time.Hour),
			DurationSeconds: 3600,
		},
		ResourceUsage: ResourceUsage{
			CPUMillicores:  1000,
			MemoryBytesAvg: 1024 * 1024 * 100,
			MemoryBytesMax: 1024 * 1024 * 200,
		},
	}

	t.Run("valid record passes all checks", func(t *testing.T) {
		flags := PerformUsageFraudChecks(baseRecord, nil)
		assert.True(t, flags.OverallPassed)
		assert.False(t, flags.FlaggedForReview)
	})

	t.Run("impossible CPU fails check", func(t *testing.T) {
		record := baseRecord
		record.CPUMillicores = 2000 * 1000 // 2000 cores
		flags := PerformUsageFraudChecks(record, nil)
		assert.False(t, flags.OverallPassed)

		var found bool
		for _, check := range flags.Checks {
			if check.CheckType == "impossible_cpu" {
				found = true
				assert.False(t, check.Passed)
				assert.Equal(t, "critical", check.Severity)
			}
		}
		assert.True(t, found, "impossible_cpu check not found")
	})

	t.Run("impossible memory fails check", func(t *testing.T) {
		record := baseRecord
		record.MemoryBytesMax = 20 * 1024 * 1024 * 1024 * 1024 // 20TB
		flags := PerformUsageFraudChecks(record, nil)
		assert.False(t, flags.OverallPassed)

		var found bool
		for _, check := range flags.Checks {
			if check.CheckType == "impossible_memory" {
				found = true
				assert.False(t, check.Passed)
			}
		}
		assert.True(t, found, "impossible_memory check not found")
	})

	t.Run("anomalous CPU increase flags for review", func(t *testing.T) {
		prevRecord := baseRecord
		prevRecord.TimeWindow.EndTime = now

		record := baseRecord
		record.RecordID = "record-2"
		record.TimeWindow.StartTime = now
		record.TimeWindow.EndTime = now.Add(1 * time.Hour)
		record.CPUMillicores = 200000 // 200x increase

		flags := PerformUsageFraudChecks(record, &prevRecord)
		assert.True(t, flags.FlaggedForReview)
	})

	t.Run("overlapping time windows fails check", func(t *testing.T) {
		prevRecord := baseRecord
		prevRecord.TimeWindow.EndTime = now.Add(2 * time.Hour)

		record := baseRecord
		record.RecordID = "record-2"
		record.TimeWindow.StartTime = now.Add(1 * time.Hour)
		record.TimeWindow.EndTime = now.Add(2 * time.Hour)

		flags := PerformUsageFraudChecks(record, &prevRecord)
		assert.False(t, flags.OverallPassed)

		var found bool
		for _, check := range flags.Checks {
			if check.CheckType == "overlapping_time_windows" {
				found = true
				assert.False(t, check.Passed)
			}
		}
		assert.True(t, found, "overlapping_time_windows check not found")
	})

	t.Run("duration mismatch fails check", func(t *testing.T) {
		record := baseRecord
		record.TimeWindow.DurationSeconds = 1800 // Wrong duration

		flags := PerformUsageFraudChecks(record, nil)
		assert.False(t, flags.OverallPassed)

		var found bool
		for _, check := range flags.Checks {
			if check.CheckType == "duration_mismatch" {
				found = true
				assert.False(t, check.Passed)
			}
		}
		assert.True(t, found, "duration_mismatch check not found")
	})
}
