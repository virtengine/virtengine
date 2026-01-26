// Package types contains types for the Fraud module.
//
// VE-912: Fraud reporting flow - Genesis state
package types

import (
	"fmt"
)

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return Params{
		MinDescriptionLength:       MinDescriptionLength,
		MaxDescriptionLength:       MaxDescriptionLength,
		MaxEvidenceCount:           10,
		MaxEvidenceSizeBytes:       10 * 1024 * 1024, // 10MB
		AutoAssignEnabled:          true,
		EscalationThresholdDays:    7,
		ReportRetentionDays:        365,
		AuditLogRetentionDays:      730, // 2 years
	}
}

// Params defines the parameters for the fraud module
type Params struct {
	// MinDescriptionLength is the minimum description length
	MinDescriptionLength int `json:"min_description_length"`

	// MaxDescriptionLength is the maximum description length
	MaxDescriptionLength int `json:"max_description_length"`

	// MaxEvidenceCount is the maximum number of evidence items per report
	MaxEvidenceCount int `json:"max_evidence_count"`

	// MaxEvidenceSizeBytes is the maximum size per evidence item
	MaxEvidenceSizeBytes int64 `json:"max_evidence_size_bytes"`

	// AutoAssignEnabled enables automatic moderator assignment
	AutoAssignEnabled bool `json:"auto_assign_enabled"`

	// EscalationThresholdDays is days before auto-escalation
	EscalationThresholdDays int `json:"escalation_threshold_days"`

	// ReportRetentionDays is how long to retain resolved reports
	ReportRetentionDays int `json:"report_retention_days"`

	// AuditLogRetentionDays is how long to retain audit logs
	AuditLogRetentionDays int `json:"audit_log_retention_days"`
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.MinDescriptionLength < 10 {
		return fmt.Errorf("min_description_length must be at least 10")
	}
	if p.MaxDescriptionLength < p.MinDescriptionLength {
		return fmt.Errorf("max_description_length must be greater than min_description_length")
	}
	if p.MaxEvidenceCount < 1 {
		return fmt.Errorf("max_evidence_count must be at least 1")
	}
	if p.MaxEvidenceSizeBytes < 1024 {
		return fmt.Errorf("max_evidence_size_bytes must be at least 1KB")
	}
	if p.EscalationThresholdDays < 1 {
		return fmt.Errorf("escalation_threshold_days must be at least 1")
	}
	if p.ReportRetentionDays < 30 {
		return fmt.Errorf("report_retention_days must be at least 30")
	}
	if p.AuditLogRetentionDays < p.ReportRetentionDays {
		return fmt.Errorf("audit_log_retention_days must be at least report_retention_days")
	}
	return nil
}

// DefaultGenesisState returns default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                   DefaultParams(),
		FraudReports:             []FraudReport{},
		AuditLogs:                []FraudAuditLog{},
		ModeratorQueue:           []ModeratorQueueEntry{},
		NextFraudReportSequence:  1,
		NextAuditLogSequence:     1,
	}
}

// GenesisState defines the genesis state for the fraud module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// FraudReports are all fraud reports
	FraudReports []FraudReport `json:"fraud_reports"`

	// AuditLogs are all audit log entries
	AuditLogs []FraudAuditLog `json:"audit_logs"`

	// ModeratorQueue are the pending queue entries
	ModeratorQueue []ModeratorQueueEntry `json:"moderator_queue"`

	// NextFraudReportSequence is the next report sequence number
	NextFraudReportSequence uint64 `json:"next_fraud_report_sequence"`

	// NextAuditLogSequence is the next audit log sequence number
	NextAuditLogSequence uint64 `json:"next_audit_log_sequence"`
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	reportIDs := make(map[string]bool)
	for i, report := range gs.FraudReports {
		if err := report.Validate(); err != nil {
			return fmt.Errorf("invalid fraud report %d: %w", i, err)
		}
		if reportIDs[report.ID] {
			return fmt.Errorf("duplicate fraud report ID: %s", report.ID)
		}
		reportIDs[report.ID] = true
	}

	logIDs := make(map[string]bool)
	for i, log := range gs.AuditLogs {
		if err := log.Validate(); err != nil {
			return fmt.Errorf("invalid audit log %d: %w", i, err)
		}
		if logIDs[log.ID] {
			return fmt.Errorf("duplicate audit log ID: %s", log.ID)
		}
		logIDs[log.ID] = true
		// Verify referenced report exists
		if !reportIDs[log.ReportID] {
			return fmt.Errorf("audit log %s references non-existent report: %s", log.ID, log.ReportID)
		}
	}

	for i, entry := range gs.ModeratorQueue {
		if entry.ReportID == "" {
			return fmt.Errorf("invalid moderator queue entry %d: empty report ID", i)
		}
		if !reportIDs[entry.ReportID] {
			return fmt.Errorf("moderator queue entry %d references non-existent report: %s", i, entry.ReportID)
		}
	}

	if gs.NextFraudReportSequence == 0 {
		return fmt.Errorf("next_fraud_report_sequence must be positive")
	}
	if gs.NextAuditLogSequence == 0 {
		return fmt.Errorf("next_audit_log_sequence must be positive")
	}

	return nil
}
