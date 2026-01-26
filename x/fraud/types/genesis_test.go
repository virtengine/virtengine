// Package types contains tests for the Fraud module types.
//
// VE-912: Fraud reporting flow - Genesis tests
package types

import (
	"testing"
	"time"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params.MinDescriptionLength != MinDescriptionLength {
		t.Errorf("DefaultParams() MinDescriptionLength = %v, want %v",
			params.MinDescriptionLength, MinDescriptionLength)
	}
	if params.MaxDescriptionLength != MaxDescriptionLength {
		t.Errorf("DefaultParams() MaxDescriptionLength = %v, want %v",
			params.MaxDescriptionLength, MaxDescriptionLength)
	}
	if params.MaxEvidenceCount != 10 {
		t.Errorf("DefaultParams() MaxEvidenceCount = %v, want %v",
			params.MaxEvidenceCount, 10)
	}
	if !params.AutoAssignEnabled {
		t.Error("DefaultParams() AutoAssignEnabled should be true")
	}
	if params.EscalationThresholdDays != 7 {
		t.Errorf("DefaultParams() EscalationThresholdDays = %v, want %v",
			params.EscalationThresholdDays, 7)
	}
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			name:    "valid default params",
			params:  DefaultParams(),
			wantErr: false,
		},
		{
			name: "min description too short",
			params: Params{
				MinDescriptionLength:    5,
				MaxDescriptionLength:    5000,
				MaxEvidenceCount:        10,
				MaxEvidenceSizeBytes:    10 * 1024 * 1024,
				EscalationThresholdDays: 7,
				ReportRetentionDays:     365,
				AuditLogRetentionDays:   730,
			},
			wantErr: true,
		},
		{
			name: "max less than min",
			params: Params{
				MinDescriptionLength:    100,
				MaxDescriptionLength:    50,
				MaxEvidenceCount:        10,
				MaxEvidenceSizeBytes:    10 * 1024 * 1024,
				EscalationThresholdDays: 7,
				ReportRetentionDays:     365,
				AuditLogRetentionDays:   730,
			},
			wantErr: true,
		},
		{
			name: "zero evidence count",
			params: Params{
				MinDescriptionLength:    20,
				MaxDescriptionLength:    5000,
				MaxEvidenceCount:        0,
				MaxEvidenceSizeBytes:    10 * 1024 * 1024,
				EscalationThresholdDays: 7,
				ReportRetentionDays:     365,
				AuditLogRetentionDays:   730,
			},
			wantErr: true,
		},
		{
			name: "evidence size too small",
			params: Params{
				MinDescriptionLength:    20,
				MaxDescriptionLength:    5000,
				MaxEvidenceCount:        10,
				MaxEvidenceSizeBytes:    100,
				EscalationThresholdDays: 7,
				ReportRetentionDays:     365,
				AuditLogRetentionDays:   730,
			},
			wantErr: true,
		},
		{
			name: "zero escalation days",
			params: Params{
				MinDescriptionLength:    20,
				MaxDescriptionLength:    5000,
				MaxEvidenceCount:        10,
				MaxEvidenceSizeBytes:    10 * 1024 * 1024,
				EscalationThresholdDays: 0,
				ReportRetentionDays:     365,
				AuditLogRetentionDays:   730,
			},
			wantErr: true,
		},
		{
			name: "retention days too low",
			params: Params{
				MinDescriptionLength:    20,
				MaxDescriptionLength:    5000,
				MaxEvidenceCount:        10,
				MaxEvidenceSizeBytes:    10 * 1024 * 1024,
				EscalationThresholdDays: 7,
				ReportRetentionDays:     10,
				AuditLogRetentionDays:   730,
			},
			wantErr: true,
		},
		{
			name: "audit log retention less than report retention",
			params: Params{
				MinDescriptionLength:    20,
				MaxDescriptionLength:    5000,
				MaxEvidenceCount:        10,
				MaxEvidenceSizeBytes:    10 * 1024 * 1024,
				EscalationThresholdDays: 7,
				ReportRetentionDays:     365,
				AuditLogRetentionDays:   100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Params.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultGenesisState(t *testing.T) {
	gs := DefaultGenesisState()

	if gs.NextFraudReportSequence != 1 {
		t.Errorf("DefaultGenesisState() NextFraudReportSequence = %v, want %v",
			gs.NextFraudReportSequence, 1)
	}
	if gs.NextAuditLogSequence != 1 {
		t.Errorf("DefaultGenesisState() NextAuditLogSequence = %v, want %v",
			gs.NextAuditLogSequence, 1)
	}
	if len(gs.FraudReports) != 0 {
		t.Errorf("DefaultGenesisState() FraudReports should be empty")
	}
	if len(gs.AuditLogs) != 0 {
		t.Errorf("DefaultGenesisState() AuditLogs should be empty")
	}
	if len(gs.ModeratorQueue) != 0 {
		t.Errorf("DefaultGenesisState() ModeratorQueue should be empty")
	}
}

func TestGenesisState_Validate(t *testing.T) {
	now := time.Now()
	validEvidence := []EncryptedEvidence{{
		AlgorithmID:     "X25519",
		RecipientKeyIDs: []string{"key1"},
		Nonce:           []byte("nonce"),
		Ciphertext:      []byte("data"),
		SenderPubKey:    []byte("pubkey"),
		EvidenceHash:    "hash",
	}}
	validDescription := "This is a description that exceeds the minimum length requirement."

	validReport := FraudReport{
		ID:            "fraud-report-1",
		Reporter:      "cosmos1reporter",
		ReportedParty: "cosmos1reported",
		Category:      FraudCategoryFakeIdentity,
		Description:   validDescription,
		Evidence:      validEvidence,
		Status:        FraudReportStatusSubmitted,
		SubmittedAt:   now,
		UpdatedAt:     now,
	}

	validLog := FraudAuditLog{
		ID:        "fraud-report-1/audit-1",
		ReportID:  "fraud-report-1",
		Action:    AuditActionSubmitted,
		Actor:     "cosmos1actor",
		Timestamp: now,
	}

	validQueueEntry := ModeratorQueueEntry{
		ReportID: "fraud-report-1",
		Priority: 10,
		QueuedAt: now,
		Category: FraudCategoryFakeIdentity,
	}

	tests := []struct {
		name    string
		gs      *GenesisState
		wantErr bool
	}{
		{
			name:    "valid default genesis",
			gs:      DefaultGenesisState(),
			wantErr: false,
		},
		{
			name: "valid genesis with data",
			gs: &GenesisState{
				Params:                  DefaultParams(),
				FraudReports:            []FraudReport{validReport},
				AuditLogs:               []FraudAuditLog{validLog},
				ModeratorQueue:          []ModeratorQueueEntry{validQueueEntry},
				NextFraudReportSequence: 2,
				NextAuditLogSequence:    2,
			},
			wantErr: false,
		},
		{
			name: "invalid params",
			gs: &GenesisState{
				Params: Params{
					MinDescriptionLength: 5, // Invalid
				},
				NextFraudReportSequence: 1,
				NextAuditLogSequence:    1,
			},
			wantErr: true,
		},
		{
			name: "duplicate report IDs",
			gs: &GenesisState{
				Params:                  DefaultParams(),
				FraudReports:            []FraudReport{validReport, validReport},
				NextFraudReportSequence: 2,
				NextAuditLogSequence:    1,
			},
			wantErr: true,
		},
		{
			name: "audit log references non-existent report",
			gs: &GenesisState{
				Params: DefaultParams(),
				AuditLogs: []FraudAuditLog{{
					ID:        "log-1",
					ReportID:  "non-existent-report",
					Action:    AuditActionSubmitted,
					Actor:     "cosmos1actor",
					Timestamp: now,
				}},
				NextFraudReportSequence: 1,
				NextAuditLogSequence:    2,
			},
			wantErr: true,
		},
		{
			name: "queue entry references non-existent report",
			gs: &GenesisState{
				Params: DefaultParams(),
				ModeratorQueue: []ModeratorQueueEntry{{
					ReportID: "non-existent-report",
					Priority: 10,
					QueuedAt: now,
					Category: FraudCategoryFakeIdentity,
				}},
				NextFraudReportSequence: 1,
				NextAuditLogSequence:    1,
			},
			wantErr: true,
		},
		{
			name: "zero fraud report sequence",
			gs: &GenesisState{
				Params:                  DefaultParams(),
				NextFraudReportSequence: 0,
				NextAuditLogSequence:    1,
			},
			wantErr: true,
		},
		{
			name: "zero audit log sequence",
			gs: &GenesisState{
				Params:                  DefaultParams(),
				NextFraudReportSequence: 1,
				NextAuditLogSequence:    0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gs.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenesisState.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
