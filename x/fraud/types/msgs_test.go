// Package types contains tests for the Fraud module message types.
//
// VE-912: Fraud reporting flow - Message tests
// VE-3053: Updated to use proto types correctly
package types

import (
	"strings"
	"testing"
)

func TestMsgSubmitFraudReport_ValidateBasic(t *testing.T) {
	// Use proto type with correct field names (AlgorithmId, RecipientKeyIds, etc.)
	validEvidence := []EncryptedEvidencePB{{
		AlgorithmId:     "X25519",
		RecipientKeyIds: []string{"key1"},
		Nonce:           []byte("nonce"),
		Ciphertext:      []byte("data"),
		SenderPubKey:    []byte("pubkey"),
		EvidenceHash:    "hash",
	}}
	validDescription := "This is a description that exceeds the minimum length requirement."

	tests := []struct {
		name    string
		msg     *MsgSubmitFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: false,
		},
		{
			name: "invalid reporter address",
			msg: &MsgSubmitFraudReport{
				Reporter:      "invalid",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "invalid reported party address",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "invalid",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "self report",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "invalid category",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
				Category:      FraudCategoryPBUnspecified,
				Description:   validDescription,
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "description too short",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   "short",
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "description too long",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   strings.Repeat("a", MaxDescriptionLength+1),
				Evidence:      validEvidence,
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			msg: &MsgSubmitFraudReport{
				Reporter:      "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportedParty: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
				Category:      FraudCategoryPBFakeIdentity,
				Description:   validDescription,
				Evidence:      []EncryptedEvidencePB{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgSubmitFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgSubmitFraudReport_GetSigners(t *testing.T) {
	msg := &MsgSubmitFraudReport{
		Reporter: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
	}

	signers := msg.GetSigners()
	if len(signers) != 1 {
		t.Errorf("GetSigners() returned %d signers, want 1", len(signers))
	}
}

func TestMsgSubmitFraudReport_Route(t *testing.T) {
	msg := &MsgSubmitFraudReport{}
	if got := msg.Route(); got != RouterKey {
		t.Errorf("Route() = %v, want %v", got, RouterKey)
	}
}

func TestMsgSubmitFraudReport_Type(t *testing.T) {
	msg := &MsgSubmitFraudReport{}
	if got := msg.Type(); got != TypeMsgSubmitFraudReport {
		t.Errorf("Type() = %v, want %v", got, TypeMsgSubmitFraudReport)
	}
}

func TestMsgAssignModerator_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgAssignModerator
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgAssignModerator{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				AssignTo:  "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgAssignModerator{
				Moderator: "invalid",
				ReportId:  "fraud-report-1",
				AssignTo:  "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgAssignModerator{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "",
				AssignTo:  "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgp0ctjdj",
			},
			wantErr: true,
		},
		{
			name: "invalid assign_to address",
			msg: &MsgAssignModerator{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				AssignTo:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgAssignModerator.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgUpdateReportStatus_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgUpdateReportStatus
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgUpdateReportStatus{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				NewStatus: FraudReportStatusPBReviewing,
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgUpdateReportStatus{
				Moderator: "invalid",
				ReportId:  "fraud-report-1",
				NewStatus: FraudReportStatusPBReviewing,
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgUpdateReportStatus{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "",
				NewStatus: FraudReportStatusPBReviewing,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			msg: &MsgUpdateReportStatus{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				NewStatus: FraudReportStatusPBUnspecified,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgUpdateReportStatus.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgResolveFraudReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgResolveFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:   "fraud-report-1",
				Resolution: ResolutionTypePBWarning,
				Notes:      "User has been warned",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgResolveFraudReport{
				Moderator:  "invalid",
				ReportId:   "fraud-report-1",
				Resolution: ResolutionTypePBWarning,
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:   "",
				Resolution: ResolutionTypePBWarning,
			},
			wantErr: true,
		},
		{
			name: "invalid resolution",
			msg: &MsgResolveFraudReport{
				Moderator:  "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:   "fraud-report-1",
				Resolution: ResolutionTypePBUnspecified,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgResolveFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgRejectFraudReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRejectFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgRejectFraudReport{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				Notes:     "No evidence of fraud",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgRejectFraudReport{
				Moderator: "invalid",
				ReportId:  "fraud-report-1",
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgRejectFraudReport{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgRejectFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgEscalateFraudReport_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgEscalateFraudReport
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgEscalateFraudReport{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				Reason:    "Requires admin review",
			},
			wantErr: false,
		},
		{
			name: "invalid moderator address",
			msg: &MsgEscalateFraudReport{
				Moderator: "invalid",
				ReportId:  "fraud-report-1",
				Reason:    "Requires admin review",
			},
			wantErr: true,
		},
		{
			name: "empty report ID",
			msg: &MsgEscalateFraudReport{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "",
				Reason:    "Requires admin review",
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			msg: &MsgEscalateFraudReport{
				Moderator: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				ReportId:  "fraud-report-1",
				Reason:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgEscalateFraudReport.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	// Use proto Params type
	validParams := *ParamsToProto(&Params{
		MinDescriptionLength:    MinDescriptionLength,
		MaxDescriptionLength:    MaxDescriptionLength,
		MaxEvidenceCount:        10,
		MaxEvidenceSizeBytes:    10 * 1024 * 1024,
		AutoAssignEnabled:       true,
		EscalationThresholdDays: 7,
		ReportRetentionDays:     365,
		AuditLogRetentionDays:   730,
	})

	tests := []struct {
		name    string
		msg     *MsgUpdateParams
		wantErr bool
	}{
		{
			name: "valid message",
			msg: &MsgUpdateParams{
				Authority: "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq",
				Params:    validParams,
			},
			wantErr: false,
		},
		{
			name: "invalid authority address",
			msg: &MsgUpdateParams{
				Authority: "invalid",
				Params:    validParams,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("MsgUpdateParams.ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
