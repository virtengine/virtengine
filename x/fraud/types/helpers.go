// Package types contains types for the Fraud module.
package types

import (
	fraudv1 "github.com/virtengine/virtengine/sdk/go/node/fraud/v1"
)

// IsValidFraudCategory checks if a fraud category is valid
func IsValidFraudCategory(cat fraudv1.FraudCategory) bool {
	return cat >= fraudv1.FraudCategoryFakeIdentity &&
		cat <= fraudv1.FraudCategoryOther
}

// IsValidFraudReportStatus checks if a fraud report status is valid
func IsValidFraudReportStatus(status fraudv1.FraudReportStatus) bool {
	return status >= fraudv1.FraudReportStatusSubmitted &&
		status <= fraudv1.FraudReportStatusEscalated
}

// IsValidResolutionType checks if a resolution type is valid
func IsValidResolutionType(res fraudv1.ResolutionType) bool {
	return res >= fraudv1.ResolutionTypeWarning &&
		res <= fraudv1.ResolutionTypeNoAction
}

// ValidateEncryptedEvidence validates an encrypted evidence structure
func ValidateEncryptedEvidence(e *fraudv1.EncryptedEvidence) error {
	if e == nil {
		return ErrMissingEvidence
	}
	return ValidateEncryptedEvidenceValue(*e)
}

// ValidateEncryptedEvidenceValue validates an encrypted evidence structure by value
func ValidateEncryptedEvidenceValue(e fraudv1.EncryptedEvidence) error {
	if e.AlgorithmId == "" {
		return ErrInvalidEvidence.Wrap("algorithm ID is required")
	}
	if len(e.RecipientKeyIds) == 0 {
		return ErrInvalidEvidence.Wrap("at least one moderator recipient required")
	}
	if len(e.Nonce) == 0 {
		return ErrInvalidEvidence.Wrap("nonce is required")
	}
	if len(e.Ciphertext) == 0 {
		return ErrInvalidEvidence.Wrap("ciphertext is required")
	}
	if len(e.SenderPubKey) == 0 {
		return ErrInvalidEvidence.Wrap("sender public key is required")
	}
	if e.EvidenceHash == "" {
		return ErrInvalidEvidence.Wrap("evidence hash is required")
	}
	return nil
}

// ValidateParams validates module parameters
func ValidateParams(p *fraudv1.Params) error {
	if p == nil {
		return ErrInvalidDescription.Wrap("params cannot be nil")
	}
	if p.MinDescriptionLength <= 0 || p.MinDescriptionLength > p.MaxDescriptionLength {
		return ErrInvalidDescription.Wrap("invalid min_description_length")
	}
	if p.MaxDescriptionLength <= 0 || p.MaxDescriptionLength < p.MinDescriptionLength {
		return ErrInvalidDescription.Wrap("invalid max_description_length")
	}
	if p.MaxEvidenceCount <= 0 {
		return ErrInvalidEvidence.Wrap("max_evidence_count must be positive")
	}
	if p.MaxEvidenceSizeBytes <= 0 {
		return ErrInvalidEvidence.Wrap("max_evidence_size_bytes must be positive")
	}
	if p.EscalationThresholdDays <= 0 {
		return ErrInvalidStatus.Wrap("escalation_threshold_days must be positive")
	}
	if p.ReportRetentionDays <= 0 {
		return ErrInvalidStatus.Wrap("report_retention_days must be positive")
	}
	if p.AuditLogRetentionDays <= 0 {
		return ErrInvalidStatus.Wrap("audit_log_retention_days must be positive")
	}
	return nil
}
