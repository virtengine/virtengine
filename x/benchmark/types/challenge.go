// Package types contains types for the Benchmark module.
//
// VE-603: Challenge protocol types
package types

import (
	"fmt"
	"time"
)

// ChallengeState represents the state of a benchmark challenge
type ChallengeState string

const (
	// ChallengeStatePending indicates the challenge is awaiting response
	ChallengeStatePending ChallengeState = "pending"

	// ChallengeStateCompleted indicates the challenge was completed
	ChallengeStateCompleted ChallengeState = "completed"

	// ChallengeStateExpired indicates the challenge expired without response
	ChallengeStateExpired ChallengeState = "expired"

	// ChallengeStateFlagged indicates the response was flagged for anomalies
	ChallengeStateFlagged ChallengeState = "flagged"
)

// IsValidChallengeState checks if the state is valid
func IsValidChallengeState(s ChallengeState) bool {
	switch s {
	case ChallengeStatePending, ChallengeStateCompleted, ChallengeStateExpired, ChallengeStateFlagged:
		return true
	default:
		return false
	}
}

// BenchmarkChallenge represents a challenge to a provider
type BenchmarkChallenge struct {
	// ChallengeID is the unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// ProviderAddress is the challenged provider's address
	ProviderAddress string `json:"provider_address"`

	// ClusterID is the cluster to benchmark
	ClusterID string `json:"cluster_id"`

	// OfferingID is the optional offering to benchmark
	OfferingID string `json:"offering_id,omitempty"`

	// RequiredSuiteVersion is the required benchmark suite version
	RequiredSuiteVersion string `json:"required_suite_version"`

	// SuiteHash is the expected suite hash
	SuiteHash string `json:"suite_hash"`

	// Deadline is when the challenge expires
	Deadline time.Time `json:"deadline"`

	// State is the current challenge state
	State ChallengeState `json:"state"`

	// CreatedAt is when the challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ResponseReportID is the report ID of the response (if any)
	ResponseReportID string `json:"response_report_id,omitempty"`

	// BlockHeight is when the challenge was created on-chain
	BlockHeight int64 `json:"block_height"`

	// Requester is who requested the challenge (empty for random)
	Requester string `json:"requester,omitempty"`
}

// Validate validates a challenge
func (c *BenchmarkChallenge) Validate() error {
	if c.ChallengeID == "" {
		return fmt.Errorf("challenge_id cannot be empty")
	}

	if c.ProviderAddress == "" {
		return fmt.Errorf("provider_address cannot be empty")
	}

	if c.ClusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}

	if c.RequiredSuiteVersion == "" {
		return fmt.Errorf("required_suite_version cannot be empty")
	}

	if c.Deadline.IsZero() {
		return fmt.Errorf("deadline cannot be zero")
	}

	return nil
}

// IsExpired checks if the challenge has expired
func (c *BenchmarkChallenge) IsExpired(now time.Time) bool {
	return now.After(c.Deadline)
}

// AnomalyType represents the type of anomaly detected
type AnomalyType string

const (
	// AnomalyTypeSuddenJump is a sudden unrealistic performance jump
	AnomalyTypeSuddenJump AnomalyType = "sudden_jump"

	// AnomalyTypeInconsistentRatio is inconsistent CPU/memory ratio
	AnomalyTypeInconsistentRatio AnomalyType = "inconsistent_ratio"

	// AnomalyTypeRepeatedOutput is repeated identical benchmark output
	AnomalyTypeRepeatedOutput AnomalyType = "repeated_output"

	// AnomalyTypeTimestampAnomaly is suspicious timing patterns
	AnomalyTypeTimestampAnomaly AnomalyType = "timestamp_anomaly"

	// AnomalyTypeSuiteMismatch is suite version/hash mismatch
	AnomalyTypeSuiteMismatch AnomalyType = "suite_mismatch"

	// AnomalyTypeSignatureIssue is signature-related issues
	AnomalyTypeSignatureIssue AnomalyType = "signature_issue"
)

// AnomalySeverity represents the severity of an anomaly
type AnomalySeverity string

const (
	// AnomalySeverityLow is a low severity anomaly
	AnomalySeverityLow AnomalySeverity = "low"

	// AnomalySeverityMedium is a medium severity anomaly
	AnomalySeverityMedium AnomalySeverity = "medium"

	// AnomalySeverityHigh is a high severity anomaly
	AnomalySeverityHigh AnomalySeverity = "high"

	// AnomalySeverityCritical is a critical severity anomaly
	AnomalySeverityCritical AnomalySeverity = "critical"
)

// AnomalyFlag represents an anomaly flag on a benchmark report
type AnomalyFlag struct {
	// FlagID is the unique identifier for this flag
	FlagID string `json:"flag_id"`

	// ReportID is the flagged benchmark report
	ReportID string `json:"report_id"`

	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// Type is the type of anomaly
	Type AnomalyType `json:"type"`

	// Severity is the severity level
	Severity AnomalySeverity `json:"severity"`

	// Description describes the anomaly
	Description string `json:"description"`

	// Evidence contains evidence data (JSON)
	Evidence string `json:"evidence,omitempty"`

	// CreatedAt is when the flag was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is when the flag was created on-chain
	BlockHeight int64 `json:"block_height"`

	// Resolved indicates if the flag has been reviewed
	Resolved bool `json:"resolved"`

	// Resolution contains the resolution details
	Resolution string `json:"resolution,omitempty"`

	// ResolvedAt is when the flag was resolved
	ResolvedAt time.Time `json:"resolved_at,omitempty"`

	// ResolvedBy is who resolved the flag
	ResolvedBy string `json:"resolved_by,omitempty"`
}

// Validate validates an anomaly flag
func (f *AnomalyFlag) Validate() error {
	if f.FlagID == "" {
		return fmt.Errorf("flag_id cannot be empty")
	}

	if f.ReportID == "" {
		return fmt.Errorf("report_id cannot be empty")
	}

	if f.ProviderAddress == "" {
		return fmt.Errorf("provider_address cannot be empty")
	}

	if f.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}

	if f.Severity == "" {
		return fmt.Errorf("severity cannot be empty")
	}

	return nil
}

// ProviderFlag represents a moderation flag on a provider
type ProviderFlag struct {
	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// Active indicates if the flag is active
	Active bool `json:"active"`

	// Reason is the reason for flagging
	Reason string `json:"reason"`

	// FlaggedAt is when the flag was set
	FlaggedAt time.Time `json:"flagged_at"`

	// FlaggedBy is who set the flag
	FlaggedBy string `json:"flagged_by"`

	// BlockHeight is when the flag was set on-chain
	BlockHeight int64 `json:"block_height"`

	// ExpiresAt is when the flag expires (zero means permanent)
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// Validate validates a provider flag
func (f *ProviderFlag) Validate() error {
	if f.ProviderAddress == "" {
		return fmt.Errorf("provider_address cannot be empty")
	}

	if f.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}

	if f.FlaggedBy == "" {
		return fmt.Errorf("flagged_by cannot be empty")
	}

	return nil
}

// IsActive checks if the flag is currently active
func (f *ProviderFlag) IsActive(now time.Time) bool {
	if !f.Active {
		return false
	}

	if !f.ExpiresAt.IsZero() && now.After(f.ExpiresAt) {
		return false
	}

	return true
}

// ChallengeResponse represents an explanation artifact from a provider
type ChallengeResponse struct {
	// ChallengeID is the challenge being responded to
	ChallengeID string `json:"challenge_id"`

	// ProviderAddress is the provider's address
	ProviderAddress string `json:"provider_address"`

	// ReportID is the benchmark report submitted
	ReportID string `json:"report_id"`

	// ExplanationRef is an optional encrypted explanation artifact reference
	ExplanationRef string `json:"explanation_ref,omitempty"`

	// SubmittedAt is when the response was submitted
	SubmittedAt time.Time `json:"submitted_at"`

	// BlockHeight is when the response was submitted on-chain
	BlockHeight int64 `json:"block_height"`
}
