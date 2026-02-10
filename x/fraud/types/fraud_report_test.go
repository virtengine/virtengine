// Package types contains tests for the Fraud module types.
//
// VE-912: Fraud reporting flow - Type tests
package types

import (
	"testing"
	"time"
)

func TestFraudReportStatus_String(t *testing.T) {
	tests := []struct {
		status   FraudReportStatus
		expected string
	}{
		{FraudReportStatusUnspecified, "unspecified"},
		{FraudReportStatusSubmitted, "submitted"},
		{FraudReportStatusReviewing, "reviewing"},
		{FraudReportStatusResolved, "resolved"},
		{FraudReportStatusRejected, "rejected"},
		{FraudReportStatusEscalated, "escalated"},
		{FraudReportStatus(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("FraudReportStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFraudReportStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   FraudReportStatus
		expected bool
	}{
		{FraudReportStatusUnspecified, false},
		{FraudReportStatusSubmitted, true},
		{FraudReportStatusReviewing, true},
		{FraudReportStatusResolved, true},
		{FraudReportStatusRejected, true},
		{FraudReportStatusEscalated, true},
		{FraudReportStatus(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.expected {
				t.Errorf("FraudReportStatus.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFraudReportStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   FraudReportStatus
		expected bool
	}{
		{FraudReportStatusSubmitted, false},
		{FraudReportStatusReviewing, false},
		{FraudReportStatusResolved, true},
		{FraudReportStatusRejected, true},
		{FraudReportStatusEscalated, false},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.expected {
				t.Errorf("FraudReportStatus.IsTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFraudReportStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from     FraudReportStatus
		to       FraudReportStatus
		expected bool
	}{
		// From Submitted
		{FraudReportStatusSubmitted, FraudReportStatusReviewing, true},
		{FraudReportStatusSubmitted, FraudReportStatusRejected, true},
		{FraudReportStatusSubmitted, FraudReportStatusResolved, false},
		{FraudReportStatusSubmitted, FraudReportStatusEscalated, false},

		// From Reviewing
		{FraudReportStatusReviewing, FraudReportStatusResolved, true},
		{FraudReportStatusReviewing, FraudReportStatusRejected, true},
		{FraudReportStatusReviewing, FraudReportStatusEscalated, true},
		{FraudReportStatusReviewing, FraudReportStatusSubmitted, false},

		// From Escalated
		{FraudReportStatusEscalated, FraudReportStatusResolved, true},
		{FraudReportStatusEscalated, FraudReportStatusRejected, true},
		{FraudReportStatusEscalated, FraudReportStatusReviewing, false},

		// From Terminal states
		{FraudReportStatusResolved, FraudReportStatusReviewing, false},
		{FraudReportStatusResolved, FraudReportStatusRejected, false},
		{FraudReportStatusRejected, FraudReportStatusResolved, false},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"->"+tt.to.String(), func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFraudCategory_String(t *testing.T) {
	tests := []struct {
		category FraudCategory
		expected string
	}{
		{FraudCategoryUnspecified, "unspecified"},
		{FraudCategoryFakeIdentity, "fake_identity"},
		{FraudCategoryPaymentFraud, "payment_fraud"},
		{FraudCategoryServiceMisrepresentation, "service_misrepresentation"},
		{FraudCategoryResourceAbuse, "resource_abuse"},
		{FraudCategorySybilAttack, "sybil_attack"},
		{FraudCategoryMaliciousContent, "malicious_content"},
		{FraudCategoryTermsViolation, "terms_violation"},
		{FraudCategoryOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.category.String(); got != tt.expected {
				t.Errorf("FraudCategory.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFraudCategory_IsValid(t *testing.T) {
	tests := []struct {
		category FraudCategory
		expected bool
	}{
		{FraudCategoryUnspecified, false},
		{FraudCategoryFakeIdentity, true},
		{FraudCategoryPaymentFraud, true},
		{FraudCategoryOther, true},
		{FraudCategory(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			if got := tt.category.IsValid(); got != tt.expected {
				t.Errorf("FraudCategory.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFraudCategoryFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected FraudCategory
		wantErr  bool
	}{
		{"fake_identity", FraudCategoryFakeIdentity, false},
		{"payment_fraud", FraudCategoryPaymentFraud, false},
		{"other", FraudCategoryOther, false},
		{"invalid", FraudCategoryUnspecified, true},
		{"", FraudCategoryUnspecified, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := FraudCategoryFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FraudCategoryFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("FraudCategoryFromString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResolutionType_String(t *testing.T) {
	tests := []struct {
		resolution ResolutionType
		expected   string
	}{
		{ResolutionTypeUnspecified, "unspecified"},
		{ResolutionTypeWarning, "warning"},
		{ResolutionTypeSuspension, "suspension"},
		{ResolutionTypeTermination, "termination"},
		{ResolutionTypeRefund, "refund"},
		{ResolutionTypeNoAction, "no_action"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.resolution.String(); got != tt.expected {
				t.Errorf("ResolutionType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResolutionType_IsValid(t *testing.T) {
	tests := []struct {
		resolution ResolutionType
		expected   bool
	}{
		{ResolutionTypeUnspecified, false},
		{ResolutionTypeWarning, true},
		{ResolutionTypeSuspension, true},
		{ResolutionTypeNoAction, true},
		{ResolutionType(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.resolution.String(), func(t *testing.T) {
			if got := tt.resolution.IsValid(); got != tt.expected {
				t.Errorf("ResolutionType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEncryptedEvidence_Validate(t *testing.T) {
	validEvidence := &EncryptedEvidence{
		AlgorithmID:     "X25519-XSALSA20-POLY1305",
		RecipientKeyIDs: []string{"key1", "key2"},
		Nonce:           []byte("nonce123"),
		Ciphertext:      []byte("encrypted_data"),
		SenderPubKey:    []byte("pubkey"),
		EvidenceHash:    "abc123",
	}

	tests := []struct {
		name     string
		evidence *EncryptedEvidence
		wantErr  bool
	}{
		{"valid evidence", validEvidence, false},
		{"nil evidence", nil, true},
		{"missing algorithm", &EncryptedEvidence{
			RecipientKeyIDs: []string{"key1"},
			Nonce:           []byte("nonce"),
			Ciphertext:      []byte("data"),
			SenderPubKey:    []byte("pubkey"),
			EvidenceHash:    "hash",
		}, true},
		{"missing recipients", &EncryptedEvidence{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{},
			Nonce:           []byte("nonce"),
			Ciphertext:      []byte("data"),
			SenderPubKey:    []byte("pubkey"),
			EvidenceHash:    "hash",
		}, true},
		{"missing nonce", &EncryptedEvidence{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			Nonce:           []byte{},
			Ciphertext:      []byte("data"),
			SenderPubKey:    []byte("pubkey"),
			EvidenceHash:    "hash",
		}, true},
		{"missing ciphertext", &EncryptedEvidence{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			Nonce:           []byte("nonce"),
			Ciphertext:      []byte{},
			SenderPubKey:    []byte("pubkey"),
			EvidenceHash:    "hash",
		}, true},
		{"missing sender pubkey", &EncryptedEvidence{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			Nonce:           []byte("nonce"),
			Ciphertext:      []byte("data"),
			SenderPubKey:    []byte{},
			EvidenceHash:    "hash",
		}, true},
		{"missing evidence hash", &EncryptedEvidence{
			AlgorithmID:     "X25519",
			RecipientKeyIDs: []string{"key1"},
			Nonce:           []byte("nonce"),
			Ciphertext:      []byte("data"),
			SenderPubKey:    []byte("pubkey"),
			EvidenceHash:    "",
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.evidence.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptedEvidence.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptedEvidence_ComputeCiphertextHash(t *testing.T) {
	evidence := &EncryptedEvidence{
		Ciphertext: []byte("test_data"),
	}

	hash := evidence.ComputeCiphertextHash()
	if hash == "" {
		t.Error("ComputeCiphertextHash() returned empty string")
	}

	// Same data should produce same hash
	evidence2 := &EncryptedEvidence{
		Ciphertext: []byte("test_data"),
	}
	if evidence.ComputeCiphertextHash() != evidence2.ComputeCiphertextHash() {
		t.Error("ComputeCiphertextHash() should be deterministic")
	}

	// Different data should produce different hash
	evidence3 := &EncryptedEvidence{
		Ciphertext: []byte("different_data"),
	}
	if evidence.ComputeCiphertextHash() == evidence3.ComputeCiphertextHash() {
		t.Error("ComputeCiphertextHash() should produce different hashes for different data")
	}
}

func TestFraudReportID_String(t *testing.T) {
	id := FraudReportID{Sequence: 123}
	expected := "fraud-report-123"
	if got := id.String(); got != expected {
		t.Errorf("FraudReportID.String() = %v, want %v", got, expected)
	}
}

func TestFraudReportID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      FraudReportID
		wantErr bool
	}{
		{"valid id", FraudReportID{Sequence: 1}, false},
		{"zero sequence", FraudReportID{Sequence: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FraudReportID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func createValidEvidence() []EncryptedEvidence {
	return []EncryptedEvidence{{
		AlgorithmID:     "X25519-XSALSA20-POLY1305",
		RecipientKeyIDs: []string{"moderator-key-1"},
		Nonce:           []byte("unique_nonce_123"),
		Ciphertext:      []byte("encrypted_evidence_data"),
		SenderPubKey:    []byte("sender_public_key"),
		EvidenceHash:    "sha256_hash_of_original",
	}}
}

func TestNewFraudReport(t *testing.T) {
	now := time.Now()
	evidence := createValidEvidence()

	report := NewFraudReport(
		"fraud-report-1",
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"This is a detailed description of the fraud that exceeds the minimum length requirement.",
		evidence,
		100,
		now,
	)

	if report.ID != "fraud-report-1" {
		t.Errorf("NewFraudReport() ID = %v, want %v", report.ID, "fraud-report-1")
	}
	if report.Status != FraudReportStatusSubmitted {
		t.Errorf("NewFraudReport() Status = %v, want %v", report.Status, FraudReportStatusSubmitted)
	}
	if report.ContentHash == "" {
		t.Error("NewFraudReport() ContentHash should not be empty")
	}
}

func TestFraudReport_Validate(t *testing.T) {
	now := time.Now()
	evidence := createValidEvidence()
	longDescription := "This is a detailed description that meets the minimum length requirement of 20 characters."

	tests := []struct {
		name    string
		report  *FraudReport
		wantErr bool
	}{
		{
			name: "valid report",
			report: &FraudReport{
				ID:            "fraud-report-1",
				Reporter:      "cosmos1reporter",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryFakeIdentity,
				Description:   longDescription,
				Evidence:      evidence,
				Status:        FraudReportStatusSubmitted,
				SubmittedAt:   now,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			report: &FraudReport{
				Reporter:      "cosmos1reporter",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryFakeIdentity,
				Description:   longDescription,
				Evidence:      evidence,
				SubmittedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing reporter",
			report: &FraudReport{
				ID:            "fraud-report-1",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryFakeIdentity,
				Description:   longDescription,
				Evidence:      evidence,
				SubmittedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "self report",
			report: &FraudReport{
				ID:            "fraud-report-1",
				Reporter:      "cosmos1same",
				ReportedParty: "cosmos1same",
				Category:      FraudCategoryFakeIdentity,
				Description:   longDescription,
				Evidence:      evidence,
				SubmittedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid category",
			report: &FraudReport{
				ID:            "fraud-report-1",
				Reporter:      "cosmos1reporter",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryUnspecified,
				Description:   longDescription,
				Evidence:      evidence,
				SubmittedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "description too short",
			report: &FraudReport{
				ID:            "fraud-report-1",
				Reporter:      "cosmos1reporter",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryFakeIdentity,
				Description:   "short",
				Evidence:      evidence,
				SubmittedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			report: &FraudReport{
				ID:            "fraud-report-1",
				Reporter:      "cosmos1reporter",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryFakeIdentity,
				Description:   longDescription,
				Evidence:      []EncryptedEvidence{},
				SubmittedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "missing submitted_at",
			report: &FraudReport{
				ID:            "fraud-report-1",
				Reporter:      "cosmos1reporter",
				ReportedParty: "cosmos1reported",
				Category:      FraudCategoryFakeIdentity,
				Description:   longDescription,
				Evidence:      evidence,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FraudReport.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFraudReport_UpdateStatus(t *testing.T) {
	now := time.Now()
	evidence := createValidEvidence()

	report := NewFraudReport(
		"fraud-report-1",
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"Detailed description exceeding minimum length requirement.",
		evidence,
		100,
		now,
	)

	// Valid transition: Submitted -> Reviewing
	err := report.UpdateStatus(FraudReportStatusReviewing, now.Add(time.Hour))
	if err != nil {
		t.Errorf("UpdateStatus() valid transition error = %v", err)
	}
	if report.Status != FraudReportStatusReviewing {
		t.Errorf("UpdateStatus() Status = %v, want %v", report.Status, FraudReportStatusReviewing)
	}

	// Invalid transition: Reviewing -> Submitted
	err = report.UpdateStatus(FraudReportStatusSubmitted, now.Add(time.Hour*2))
	if err == nil {
		t.Error("UpdateStatus() invalid transition should return error")
	}
}

func TestFraudReport_Resolve(t *testing.T) {
	now := time.Now()
	evidence := createValidEvidence()

	report := NewFraudReport(
		"fraud-report-1",
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"Detailed description exceeding minimum length requirement.",
		evidence,
		100,
		now,
	)

	// Move to reviewing first
	report.Status = FraudReportStatusReviewing

	resolvedAt := now.Add(time.Hour)
	err := report.Resolve(ResolutionTypeWarning, "User warned", resolvedAt)
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if report.Status != FraudReportStatusResolved {
		t.Errorf("Resolve() Status = %v, want %v", report.Status, FraudReportStatusResolved)
	}
	if report.Resolution != ResolutionTypeWarning {
		t.Errorf("Resolve() Resolution = %v, want %v", report.Resolution, ResolutionTypeWarning)
	}
	if report.ResolvedAt == nil {
		t.Error("Resolve() ResolvedAt should not be nil")
	}

	// Try to resolve again - should fail
	err = report.Resolve(ResolutionTypeSuspension, "Second resolution", now.Add(time.Hour*2))
	if err == nil {
		t.Error("Resolve() should fail on already resolved report")
	}
}

func TestFraudReport_Reject(t *testing.T) {
	now := time.Now()
	evidence := createValidEvidence()

	report := NewFraudReport(
		"fraud-report-1",
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"Detailed description exceeding minimum length requirement.",
		evidence,
		100,
		now,
	)

	// Move to reviewing first
	report.Status = FraudReportStatusReviewing

	rejectedAt := now.Add(time.Hour)
	err := report.Reject("No evidence of fraud", rejectedAt)
	if err != nil {
		t.Errorf("Reject() error = %v", err)
	}

	if report.Status != FraudReportStatusRejected {
		t.Errorf("Reject() Status = %v, want %v", report.Status, FraudReportStatusRejected)
	}
	if report.Resolution != ResolutionTypeNoAction {
		t.Errorf("Reject() Resolution = %v, want %v", report.Resolution, ResolutionTypeNoAction)
	}

	// Try to reject again - should fail
	err = report.Reject("Second rejection", now.Add(time.Hour*2))
	if err == nil {
		t.Error("Reject() should fail on already rejected report")
	}
}

func TestFraudReport_ComputeContentHash(t *testing.T) {
	now := time.Now()
	evidence := createValidEvidence()

	report1 := NewFraudReport(
		"fraud-report-1",
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"Detailed description exceeding minimum length requirement.",
		evidence,
		100,
		now,
	)

	report2 := NewFraudReport(
		"fraud-report-1",
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"Detailed description exceeding minimum length requirement.",
		evidence,
		100,
		now,
	)

	// Same content should produce same hash
	if report1.ComputeContentHash() != report2.ComputeContentHash() {
		t.Error("ComputeContentHash() should be deterministic for same content")
	}

	// Different content should produce different hash
	report3 := NewFraudReport(
		"fraud-report-2", // Different ID
		"cosmos1reporter",
		"cosmos1reported",
		FraudCategoryFakeIdentity,
		"Detailed description exceeding minimum length requirement.",
		evidence,
		100,
		now,
	)
	if report1.ComputeContentHash() == report3.ComputeContentHash() {
		t.Error("ComputeContentHash() should produce different hashes for different content")
	}
}

func TestAuditAction_String(t *testing.T) {
	tests := []struct {
		action   AuditAction
		expected string
	}{
		{AuditActionUnspecified, "unspecified"},
		{AuditActionSubmitted, "submitted"},
		{AuditActionAssigned, "assigned"},
		{AuditActionStatusChanged, "status_changed"},
		{AuditActionEvidenceViewed, "evidence_viewed"},
		{AuditActionResolved, "resolved"},
		{AuditActionRejected, "rejected"},
		{AuditActionEscalated, "escalated"},
		{AuditActionCommentAdded, "comment_added"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.action.String(); got != tt.expected {
				t.Errorf("AuditAction.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAuditAction_IsValid(t *testing.T) {
	tests := []struct {
		action   AuditAction
		expected bool
	}{
		{AuditActionUnspecified, false},
		{AuditActionSubmitted, true},
		{AuditActionCommentAdded, true},
		{AuditAction(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.action.String(), func(t *testing.T) {
			if got := tt.action.IsValid(); got != tt.expected {
				t.Errorf("AuditAction.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewFraudAuditLog(t *testing.T) {
	now := time.Now()

	log := NewFraudAuditLog(
		"fraud-report-1/audit-1",
		"fraud-report-1",
		AuditActionSubmitted,
		"cosmos1actor",
		FraudReportStatusUnspecified,
		FraudReportStatusSubmitted,
		"Initial submission",
		now,
		100,
	)

	if log.ID != "fraud-report-1/audit-1" {
		t.Errorf("NewFraudAuditLog() ID = %v, want %v", log.ID, "fraud-report-1/audit-1")
	}
	if log.Action != AuditActionSubmitted {
		t.Errorf("NewFraudAuditLog() Action = %v, want %v", log.Action, AuditActionSubmitted)
	}
	if log.BlockHeight != 100 {
		t.Errorf("NewFraudAuditLog() BlockHeight = %v, want %v", log.BlockHeight, 100)
	}
}

func TestFraudAuditLog_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		log     *FraudAuditLog
		wantErr bool
	}{
		{
			name: "valid log",
			log: &FraudAuditLog{
				ID:        "log-1",
				ReportID:  "report-1",
				Action:    AuditActionSubmitted,
				Actor:     "cosmos1actor",
				Timestamp: now,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			log: &FraudAuditLog{
				ReportID:  "report-1",
				Action:    AuditActionSubmitted,
				Actor:     "cosmos1actor",
				Timestamp: now,
			},
			wantErr: true,
		},
		{
			name: "missing report ID",
			log: &FraudAuditLog{
				ID:        "log-1",
				Action:    AuditActionSubmitted,
				Actor:     "cosmos1actor",
				Timestamp: now,
			},
			wantErr: true,
		},
		{
			name: "invalid action",
			log: &FraudAuditLog{
				ID:        "log-1",
				ReportID:  "report-1",
				Action:    AuditActionUnspecified,
				Actor:     "cosmos1actor",
				Timestamp: now,
			},
			wantErr: true,
		},
		{
			name: "missing actor",
			log: &FraudAuditLog{
				ID:        "log-1",
				ReportID:  "report-1",
				Action:    AuditActionSubmitted,
				Timestamp: now,
			},
			wantErr: true,
		},
		{
			name: "missing timestamp",
			log: &FraudAuditLog{
				ID:       "log-1",
				ReportID: "report-1",
				Action:   AuditActionSubmitted,
				Actor:    "cosmos1actor",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FraudAuditLog.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewModeratorQueueEntry(t *testing.T) {
	now := time.Now()

	entry := NewModeratorQueueEntry(
		"fraud-report-1",
		FraudCategoryFakeIdentity,
		10,
		now,
	)

	if entry.ReportID != "fraud-report-1" {
		t.Errorf("NewModeratorQueueEntry() ReportID = %v, want %v", entry.ReportID, "fraud-report-1")
	}
	if entry.Priority != 10 {
		t.Errorf("NewModeratorQueueEntry() Priority = %v, want %v", entry.Priority, 10)
	}
	if entry.Category != FraudCategoryFakeIdentity {
		t.Errorf("NewModeratorQueueEntry() Category = %v, want %v", entry.Category, FraudCategoryFakeIdentity)
	}
	if entry.AssignedTo != "" {
		t.Errorf("NewModeratorQueueEntry() AssignedTo should be empty, got %v", entry.AssignedTo)
	}
}
