// Package daemon provides on-chain types for provider daemon operations.
//
// VE-403: Kubernetes orchestration adapter status updates
package daemon

import (
	"errors"
	"fmt"
	"time"
)

// WorkloadStatus represents the status of a workload managed by the daemon
type WorkloadStatus uint8

const (
	// WorkloadStatusUnspecified is an unspecified status
	WorkloadStatusUnspecified WorkloadStatus = 0

	// WorkloadStatusPending indicates the workload is pending creation
	WorkloadStatusPending WorkloadStatus = 1

	// WorkloadStatusProvisioning indicates the workload is being provisioned
	WorkloadStatusProvisioning WorkloadStatus = 2

	// WorkloadStatusRunning indicates the workload is running
	WorkloadStatusRunning WorkloadStatus = 3

	// WorkloadStatusSuspended indicates the workload is suspended
	WorkloadStatusSuspended WorkloadStatus = 4

	// WorkloadStatusTerminating indicates the workload is terminating
	WorkloadStatusTerminating WorkloadStatus = 5

	// WorkloadStatusTerminated indicates the workload has terminated
	WorkloadStatusTerminated WorkloadStatus = 6

	// WorkloadStatusFailed indicates the workload has failed
	WorkloadStatusFailed WorkloadStatus = 7
)

// WorkloadStatusNames maps workload status to names
var WorkloadStatusNames = map[WorkloadStatus]string{
	WorkloadStatusUnspecified:  "unspecified",
	WorkloadStatusPending:      "pending",
	WorkloadStatusProvisioning: "provisioning",
	WorkloadStatusRunning:      "running",
	WorkloadStatusSuspended:    "suspended",
	WorkloadStatusTerminating:  "terminating",
	WorkloadStatusTerminated:   "terminated",
	WorkloadStatusFailed:       "failed",
}

// String returns the string representation
func (s WorkloadStatus) String() string {
	if name, ok := WorkloadStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid returns true if the status is valid
func (s WorkloadStatus) IsValid() bool {
	return s >= WorkloadStatusPending && s <= WorkloadStatusFailed
}

// IsTerminal returns true if the status is terminal
func (s WorkloadStatus) IsTerminal() bool {
	return s == WorkloadStatusTerminated || s == WorkloadStatusFailed
}

// IsActive returns true if the workload is active
func (s WorkloadStatus) IsActive() bool {
	return s == WorkloadStatusRunning || s == WorkloadStatusProvisioning
}

// StatusUpdate represents a signed status update from the provider daemon
type StatusUpdate struct {
	// UpdateID is the unique identifier for this update
	UpdateID string `json:"update_id"`

	// OrderID is the order this status applies to
	OrderID string `json:"order_id"`

	// AllocationID is the allocation this status applies to
	AllocationID string `json:"allocation_id"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// PreviousStatus is the previous workload status
	PreviousStatus WorkloadStatus `json:"previous_status"`

	// CurrentStatus is the current workload status
	CurrentStatus WorkloadStatus `json:"current_status"`

	// StatusMessage is an optional human-readable message
	StatusMessage string `json:"status_message,omitempty"`

	// ErrorCode is set if the status is failed
	ErrorCode string `json:"error_code,omitempty"`

	// WorkloadRef contains orchestrator-specific workload reference
	WorkloadRef WorkloadRef `json:"workload_ref"`

	// Signature is the daemon's cryptographic signature
	Signature DaemonSignature `json:"signature"`

	// CreatedAt is when the update was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is the block at which this was recorded
	BlockHeight int64 `json:"block_height"`
}

// WorkloadRef contains the orchestrator-specific workload reference
type WorkloadRef struct {
	// Orchestrator is the orchestrator type (kubernetes, slurm)
	Orchestrator string `json:"orchestrator"`

	// Namespace is the namespace for the workload (Kubernetes)
	Namespace string `json:"namespace,omitempty"`

	// Name is the workload name
	Name string `json:"name"`

	// UID is the orchestrator-specific unique ID
	UID string `json:"uid,omitempty"`

	// JobID is the job ID for batch orchestrators (SLURM)
	JobID string `json:"job_id,omitempty"`
}

// Validate validates the WorkloadRef
func (wr WorkloadRef) Validate() error {
	if wr.Orchestrator == "" {
		return errors.New("orchestrator is required")
	}
	if wr.Orchestrator != "kubernetes" && wr.Orchestrator != "slurm" {
		return fmt.Errorf("unsupported orchestrator: %s", wr.Orchestrator)
	}
	if wr.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// Validate validates the StatusUpdate
func (su StatusUpdate) Validate() error {
	if su.UpdateID == "" {
		return errors.New("update_id is required")
	}
	if su.OrderID == "" {
		return errors.New("order_id is required")
	}
	if su.AllocationID == "" {
		return errors.New("allocation_id is required")
	}
	if su.ProviderAddress == "" {
		return errors.New("provider_address is required")
	}
	if !su.CurrentStatus.IsValid() {
		return fmt.Errorf("invalid current_status: %d", su.CurrentStatus)
	}
	if err := su.WorkloadRef.Validate(); err != nil {
		return fmt.Errorf("workload_ref: %w", err)
	}
	if err := su.Signature.Validate(); err != nil {
		return fmt.Errorf("signature: %w", err)
	}
	if su.CreatedAt.IsZero() {
		return errors.New("created_at is required")
	}
	return nil
}

// ValidTransition checks if the status transition is valid
var validStatusTransitions = map[WorkloadStatus][]WorkloadStatus{
	WorkloadStatusPending:      {WorkloadStatusProvisioning, WorkloadStatusFailed},
	WorkloadStatusProvisioning: {WorkloadStatusRunning, WorkloadStatusFailed},
	WorkloadStatusRunning:      {WorkloadStatusSuspended, WorkloadStatusTerminating},
	WorkloadStatusSuspended:    {WorkloadStatusRunning, WorkloadStatusTerminating},
	WorkloadStatusTerminating:  {WorkloadStatusTerminated, WorkloadStatusFailed},
}

// IsValidTransition checks if the status transition is valid
func (su StatusUpdate) IsValidTransition() bool {
	allowed, ok := validStatusTransitions[su.PreviousStatus]
	if !ok {
		// If previous status is terminal, no transitions allowed
		return su.PreviousStatus.IsTerminal()
	}
	for _, s := range allowed {
		if s == su.CurrentStatus {
			return true
		}
	}
	return false
}

// FraudCheck represents a fraud check result
type FraudCheck struct {
	// CheckType is the type of fraud check performed
	CheckType string `json:"check_type"`

	// Passed indicates if the check passed
	Passed bool `json:"passed"`

	// Reason explains why the check failed (if applicable)
	Reason string `json:"reason,omitempty"`

	// Severity indicates the severity (info, warning, critical)
	Severity string `json:"severity"`

	// Details contains additional check details
	Details map[string]string `json:"details,omitempty"`

	// CheckedAt is when the check was performed
	CheckedAt time.Time `json:"checked_at"`
}

// UsageFraudFlags contains fraud check results for usage records
type UsageFraudFlags struct {
	// RecordID is the usage record being checked
	RecordID string `json:"record_id"`

	// Checks are the individual fraud checks performed
	Checks []FraudCheck `json:"checks"`

	// OverallPassed indicates if all checks passed
	OverallPassed bool `json:"overall_passed"`

	// FlaggedForReview indicates if manual review is needed
	FlaggedForReview bool `json:"flagged_for_review"`

	// CheckedAt is when the fraud checks were performed
	CheckedAt time.Time `json:"checked_at"`
}

// PerformUsageFraudChecks performs fraud checks on a usage record
func PerformUsageFraudChecks(record UsageRecord, previousRecord *UsageRecord) UsageFraudFlags {
	flags := UsageFraudFlags{
		RecordID:      record.RecordID,
		Checks:        make([]FraudCheck, 0),
		OverallPassed: true,
		CheckedAt:     time.Now().UTC(),
	}

	// Check 1: Impossible CPU usage (more than 1000 cores)
	if record.ResourceUsage.CPUMillicores > 1000*1000 {
		check := FraudCheck{
			CheckType: "impossible_cpu",
			Passed:    false,
			Reason:    fmt.Sprintf("CPU usage %d millicores exceeds maximum reasonable limit", record.ResourceUsage.CPUMillicores),
			Severity:  "critical",
			CheckedAt: time.Now().UTC(),
		}
		flags.Checks = append(flags.Checks, check)
		flags.OverallPassed = false
	}

	// Check 2: Impossible memory usage (more than 10TB)
	maxMemory := int64(10 * 1024 * 1024 * 1024 * 1024) // 10TB
	if record.ResourceUsage.MemoryBytesMax > maxMemory {
		check := FraudCheck{
			CheckType: "impossible_memory",
			Passed:    false,
			Reason:    fmt.Sprintf("Memory usage %d bytes exceeds maximum reasonable limit", record.ResourceUsage.MemoryBytesMax),
			Severity:  "critical",
			CheckedAt: time.Now().UTC(),
		}
		flags.Checks = append(flags.Checks, check)
		flags.OverallPassed = false
	}

	// Check 3: Usage delta anomaly (if we have a previous record)
	if previousRecord != nil {
		// Check for impossible CPU increase (more than 100x previous in one period)
		if previousRecord.ResourceUsage.CPUMillicores > 0 {
			ratio := float64(record.ResourceUsage.CPUMillicores) / float64(previousRecord.ResourceUsage.CPUMillicores)
			if ratio > 100 {
				check := FraudCheck{
					CheckType: "anomalous_cpu_increase",
					Passed:    false,
					Reason:    fmt.Sprintf("CPU increased %.2fx from previous record, which is anomalous", ratio),
					Severity:  "warning",
					Details: map[string]string{
						"previous_cpu": fmt.Sprintf("%d", previousRecord.ResourceUsage.CPUMillicores),
						"current_cpu":  fmt.Sprintf("%d", record.ResourceUsage.CPUMillicores),
					},
					CheckedAt: time.Now().UTC(),
				}
				flags.Checks = append(flags.Checks, check)
				flags.FlaggedForReview = true
			}
		}

		// Check for time window consistency
		if !previousRecord.TimeWindow.EndTime.Equal(record.TimeWindow.StartTime) {
			gap := record.TimeWindow.StartTime.Sub(previousRecord.TimeWindow.EndTime)
			if gap < 0 {
				check := FraudCheck{
					CheckType: "overlapping_time_windows",
					Passed:    false,
					Reason:    "Time windows overlap with previous record",
					Severity:  "critical",
					CheckedAt: time.Now().UTC(),
				}
				flags.Checks = append(flags.Checks, check)
				flags.OverallPassed = false
			} else if gap > time.Hour {
				check := FraudCheck{
					CheckType: "time_window_gap",
					Passed:    true,
					Reason:    fmt.Sprintf("Gap of %v between time windows", gap),
					Severity:  "info",
					CheckedAt: time.Now().UTC(),
				}
				flags.Checks = append(flags.Checks, check)
				flags.FlaggedForReview = true
			}
		}
	}

	// Check 4: Duration vs actual time mismatch
	actualDuration := int64(record.TimeWindow.EndTime.Sub(record.TimeWindow.StartTime).Seconds())
	if record.TimeWindow.DurationSeconds != actualDuration {
		check := FraudCheck{
			CheckType: "duration_mismatch",
			Passed:    false,
			Reason:    fmt.Sprintf("Duration %d does not match time window %d", record.TimeWindow.DurationSeconds, actualDuration),
			Severity:  "critical",
			CheckedAt: time.Now().UTC(),
		}
		flags.Checks = append(flags.Checks, check)
		flags.OverallPassed = false
	}

	return flags
}
